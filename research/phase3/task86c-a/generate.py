#!/usr/bin/env python3
"""Deterministic, read-only post-hoc audit of frozen Task86C artifacts.

This program deliberately does not load corpus contents or invoke model code.
It reconstructs only facts retained by the frozen manifest, ledger and result
JSON files.  Missing per-metric values are emitted as NOT_RECORDED.
"""

from __future__ import annotations

import csv
import hashlib
import json
from collections import Counter, defaultdict
from pathlib import Path

ROOT = Path(__file__).resolve().parents[3]
SRC = ROOT / "research/phase3/task86c"
OUT = ROOT / "research/phase3/task86c-a"
MODELS = tuple(f"M{i}" for i in range(6))
PMS = ("PM1", "PM2", "PM4", "PM5", "PM6")


def read_tsv(name: str) -> list[dict[str, str]]:
    with (SRC / name).open(newline="", encoding="utf-8") as f:
        return list(csv.DictReader(f, delimiter="\t"))


def write_tsv(name: str, header: list[str], rows) -> None:
    with (OUT / name).open("w", newline="", encoding="utf-8") as f:
        w = csv.writer(f, delimiter="\t", lineterminator="\n")
        w.writerow(header)
        w.writerows(rows)


def sha(path: Path) -> str:
    h = hashlib.sha256()
    with path.open("rb") as f:
        for block in iter(lambda: f.read(1024 * 1024), b""):
            h.update(block)
    return h.hexdigest()


def scientific_result_sha(r: dict) -> str:
    """Reproduce task86cScientificHash (Go's compact JSON is canonical here)."""
    copy = dict(r)  # json.load preserves the frozen struct field order
    copy["Node"] = ""
    copy["StartedUTC"] = ""
    copy["FinishedUTC"] = ""
    copy["RuntimeSeconds"] = 0
    body = json.dumps(copy, separators=(",", ":"), ensure_ascii=False).encode()
    return hashlib.sha256(body).hexdigest()


def yn(v: bool) -> str:
    return "PASS" if v else "FAIL"


def failures_for(r: dict, model: str) -> list[str]:
    pfx = model + ":"
    return sorted(x[len(pfx):] for x in r["Failures"] if x.startswith(pfx))


def attribution(reason: str) -> str:
    return {
        "TRAINING_FAILED": "COMPUTATIONAL_EVIDENCE",
        "NEGATIVE_EXHAUSTED": "PROTOCOL_EVIDENCE",
        "NUMERICALLY_UNSTABLE": "COMPUTATIONAL_EVIDENCE",
        "COMPLEXITY_UNBOUNDED": "MODEL_EVIDENCE",
        "MEMORIZATION_DOMINATED": "MODEL_EVIDENCE",
    }.get(reason, "UNRESOLVED")


def category(reason: str) -> str:
    return {
        "TRAINING_FAILED": "TRAINING_FAILURE",
        "NEGATIVE_EXHAUSTED": "NEGATIVE_SET_EXHAUSTION",
        "NUMERICALLY_UNSTABLE": "GENERATION_FAILURE",
        "COMPLEXITY_UNBOUNDED": "MODEL_MISMATCH",
        "MEMORIZATION_DOMINATED": "MODEL_MISMATCH",
    }.get(reason, "NOT_RECORDED")


def fit_status(r: dict, m: str) -> str:
    return "TRAINING_FAILED" if "TRAINING_FAILED" in failures_for(r, m) else "PASS"


def pm_status(r: dict, m: str, pm: str) -> str:
    if fit_status(r, m) != "PASS":
        return "NOT_REACHED"
    if pm != "PM6":
        return "NOT_RECORDED"
    if r["PM6ByClass"][m]:
        return "AVAILABLE_GATE_NOT_RECORDED"
    if "NEGATIVE_EXHAUSTED" in failures_for(r, m):
        return "UNAVAILABLE_NEGATIVE_SET_EXHAUSTION"
    return "UNAVAILABLE_REASON_NOT_RECORDED"


def first_failure(r: dict, m: str) -> str:
    if fit_status(r, m) != "PASS":
        return "TRAINING"
    # Individual metric gates were discarded.  Aggregate predictive=false is
    # retained, so PREDICTIVE_GATE is the earliest defensible localization.
    if not r["PredictivePassByClass"][m]:
        return "PREDICTIVE_GATE"
    if "NUMERICALLY_UNSTABLE" in failures_for(r, m):
        return "GENERATION"
    if not r["StructuralPassByClass"][m]:
        return "STRUCTURAL_GATE"
    if failures_for(r, m):
        return "COMPLEXITY"
    return "NONE"


def structural_status(r: dict, m: str) -> str:
    if fit_status(r, m) != "PASS":
        return "NOT_REACHED"
    return yn(r["StructuralPassByClass"][m])


def main() -> None:
    OUT.mkdir(parents=True, exist_ok=True)
    jobs = read_tsv("TASK86C_JOB_MANIFEST.tsv")
    ledger = read_tsv("TASK86C_EXECUTION_LEDGER.tsv")
    truth = {x["opaque_id"]: x for x in read_tsv("SYNTHETIC_GROUND_TRUTH.tsv")}
    provenance = {x["opaque_id"]: x["language"] for x in read_tsv("NATURAL_LANGUAGE_CORPUS_PROVENANCE.tsv")}
    frozen_manifest = json.loads((SRC / "TASK86C_RESULTS_MANIFEST.json").read_text())

    integrity: list[str] = []
    if not (SRC / "TASK86C_G1_CONTROL_VALIDATION_FROZEN").read_text().startswith("TASK86C_G1_CONTROL_VALIDATION_FROZEN\n"):
        integrity.append("terminal marker mismatch")
    if frozen_manifest.get("terminal_marker") != "TASK86C_G1_CONTROL_VALIDATION_FROZEN":
        integrity.append("results manifest terminal marker mismatch")
    if frozen_manifest.get("jobs") != 672 or len(jobs) != 672 or len(ledger) != 672:
        integrity.append("672-job accounting mismatch")
    for name, expected in sorted(frozen_manifest["artifacts"].items()):
        p = SRC / name
        if not p.is_file() or sha(p) != expected:
            integrity.append(f"artifact hash mismatch:{name}")

    job_by_id = {x["job_id"]: x for x in jobs}
    led_by_id = {x["job_id"]: x for x in ledger}
    if len(job_by_id) != 672 or set(job_by_id) != set(led_by_id):
        integrity.append("manifest/ledger identity mismatch")
    results: dict[str, dict] = {}
    for jid in sorted(job_by_id):
        p = SRC / "results" / jid / "result.json"
        if not p.is_file():
            integrity.append(f"missing result:{jid}")
            continue
        r = json.loads(p.read_text())
        if scientific_result_sha(r) != led_by_id[jid]["result_sha256"]:
            integrity.append(f"result hash mismatch:{jid}")
        j = job_by_id[jid]
        if r["JobID"] != jid or r["InputSHA"] != j["input_sha256"] or r["ConfigSHA"] != j["config_sha256"]:
            integrity.append(f"result identity mismatch:{jid}")
        # No input_path is opened by this audit.  Identity is restricted to
        # the 18 synthetic and three natural opaque control IDs in registries.
        if j["corpus_id"] not in truth and j["corpus_id"] not in provenance:
            integrity.append(f"unregistered/non-control corpus:{jid}")
        results[jid] = r
    if len(results) != 672:
        integrity.append("result count mismatch")

    write_tsv("G1_DECISION_PATH_REGISTRY.tsv",
              ["ordinal", "stage", "retained_evidence", "allowed_statuses", "audit_rule"], [
        [1, "INPUT", "manifest identity/hash", "PASS/FAIL", "verify only; corpus contents not opened"],
        [2, "TRAINING", "TRAINING_FAILED failure class", "PASS/TRAINING_FAILED", "absence means StageD received a fitted model"],
        [3, "PM1", "none", "NOT_RECORDED/NOT_REACHED", "never infer from predictive aggregate"],
        [4, "PM2", "none", "NOT_RECORDED/NOT_REACHED", "never infer from candidate selection"],
        [5, "PM4", "none", "NOT_RECORDED/NOT_REACHED", "never infer from predictive aggregate"],
        [6, "PM5", "none", "NOT_RECORDED/NOT_REACHED", "never infer from predictive aggregate"],
        [7, "PM6", "validity only", "AVAILABLE_GATE_NOT_RECORDED/UNAVAILABLE/NOT_REACHED", "validity is not gate pass"],
        [8, "PREDICTIVE_GATE", "PredictivePassByClass", "PASS/FAIL", "recorded aggregate"],
        [9, "GENERATION", "NUMERICALLY_UNSTABLE", "PASS/GENERATION_FAILURE/NOT_REACHED", "failure may coexist with gate failure"],
        [10, "F2", "none", "NOT_RECORDED/NOT_REACHED", "individual metrics discarded"],
        [11, "STRUCTURAL_GATE", "StructuralPassByClass", "PASS/FAIL/NOT_REACHED", "recorded aggregate; pipeline ran it independently"],
        [12, "COMPLEXITY", "named failure classes", "PASS/FAIL/NOT_REACHED", "only explicit classes attributable"],
        [13, "FINAL", "SufficientByClass", "PASS/FAIL", "recorded aggregate"],
    ])

    decomp, first, pm_rows, struct_rows = [], [], [], []
    model_records = []
    for jid, r in sorted(results.items()):
        j = job_by_id[jid]
        branch_name = "SYNTHETIC" if j["branch"] == "S" else "NATURAL"
        corpus_name = truth[j["corpus_id"]]["ground_truth_class"] if j["branch"] == "S" else provenance[j["corpus_id"]]
        for m in MODELS:
            fs = failures_for(r, m)
            fit = fit_status(r, m)
            ff = first_failure(r, m)
            all_reasons = ";".join(fs) if fs else "NONE"
            attrs = ";".join(sorted(set(attribution(x) for x in fs))) if fs else "NONE"
            cats = ";".join(sorted(set(category(x) for x in fs))) if fs else "NONE"
            pm6 = pm_status(r, m, "PM6")
            pred = yn(r["PredictivePassByClass"][m]) if fit == "PASS" else "NOT_REACHED"
            gen = "NOT_REACHED" if fit != "PASS" else ("GENERATION_FAILURE" if "NUMERICALLY_UNSTABLE" in fs else "PASS")
            structural = structural_status(r, m)
            complexity = "NOT_REACHED" if fit != "PASS" else ("FAIL" if any(x in fs for x in ("COMPLEXITY_UNBOUNDED", "MEMORIZATION_DOMINATED")) else "PASS")
            final = yn(r["SufficientByClass"][m])
            base = [jid, branch_name, j["corpus_id"], corpus_name, j["generator_variant"], j["scale"], j["replicate"], j["protocol_layer"], m, r["CandidateByClass"][m]]
            decomp.append(base + [fit] + [pm_status(r, m, p) for p in PMS] + [pred, gen, "NOT_RECORDED" if fit == "PASS" else "NOT_REACHED", structural, complexity, final, ff, all_reasons, cats, attrs])
            first.append(base[:9] + [ff, all_reasons, attrs])
            for p in PMS:
                st = pm_status(r, m, p)
                pm_rows.append(base[:9] + [p, "NOT_RECORDED", st, "NOT_RECORDED" if st != "NOT_REACHED" else "NOT_REACHED", "FROZEN_RESULT_OMISSION"])
            for fam in ("EDIT", "LEXICAL", "INDIVIDUAL_F2", "AGGREGATE"):
                if fam == "AGGREGATE":
                    st = structural
                    src = "StructuralPassByClass"
                else:
                    st = "NOT_REACHED" if fit != "PASS" else "NOT_RECORDED"
                    src = "FROZEN_RESULT_OMISSION"
                struct_rows.append(base[:9] + [fam, "NOT_RECORDED", "NOT_RECORDED", st, src])
            model_records.append((j, r, m, corpus_name, fit, ff, fs, structural, pm6))

    write_tsv("JOB_FAILURE_DECOMPOSITION.tsv", ["job_id", "branch", "corpus_id", "corpus_label", "generator_variant", "scale", "replicate", "protocol_layer", "model", "candidate", "fit", "PM1", "PM2", "PM4", "PM5", "PM6", "predictive_gate", "generation", "F2", "structural_gate", "complexity", "final", "first_failure", "all_failure_reasons", "failure_categories", "evidence_attribution"], decomp)
    write_tsv("FIRST_FAILURE_STAGE.tsv", ["job_id", "branch", "corpus_id", "corpus_label", "generator_variant", "scale", "replicate", "protocol_layer", "model", "first_failure_stage", "all_failure_reasons", "evidence_attribution"], first)
    write_tsv("PM_DECOMPOSITION.tsv", ["job_id", "branch", "corpus_id", "corpus_label", "generator_variant", "scale", "replicate", "protocol_layer", "model", "metric", "value", "availability_status", "gate_status", "provenance"], pm_rows)
    write_tsv("STRUCTURAL_DECOMPOSITION.tsv", ["job_id", "branch", "corpus_id", "corpus_label", "generator_variant", "scale", "replicate", "protocol_layer", "model", "family", "metric", "value", "status", "provenance"], struct_rows)

    # PM6 ablation and counterfactuals are explicitly non-computable because
    # neither the PM1/2/4/5 gates nor PM6 gate pass were retained.
    ablation, cfs = [], []
    for j, r, m, label, fit, ff, fs, structural, pm6 in model_records:
        original = yn(r["PredictivePassByClass"][m]) if fit == "PASS" else "NOT_REACHED"
        interp = "NOT_ASSESSABLE: individual PM gate outcomes not recorded"
        ablation.append([j["job_id"], label, m, j["scale"], j["replicate"], original, pm6, "NOT_ASSESSABLE", interp])
        cfs.append([j["job_id"], label, m, "CF0", original, "FROZEN_HISTORICAL_RESULT"])
        cfs.append([j["job_id"], label, m, "CF1", "NOT_ASSESSABLE", "PM1/PM2/PM4/PM5 and PM6 gate pass not recorded"])
        cfs.append([j["job_id"], label, m, "CF2", "NOT_ASSESSABLE", "metric/structural failures cannot be fully attributed"])
    write_tsv("PM6_ABLATION.tsv", ["job_id", "corpus", "model", "scale", "replicate", "original_predictive_adequacy", "PM6_status", "predictive_without_PM6", "interpretation"], ablation)
    write_tsv("PREDICTIVE_COUNTERFACTUALS.tsv", ["job_id", "corpus", "model", "counterfactual", "status", "basis"], cfs)

    # Expected-class synthetic and all natural-language matrices.
    syn_surv, syn_latent, nat_matrix, nat_latent = [], [], [], []
    for j, r, m, label, fit, ff, fs, structural, pm6 in model_records:
        if j["branch"] == "S" and m == truth[j["corpus_id"]]["ground_truth_class"]:
            row = [j["job_id"], label, j["corpus_id"], j["generator_variant"], j["scale"], j["replicate"], fit]
            row += [pm_status(r, m, p) for p in PMS]
            row += ["NOT_ASSESSABLE", structural, yn(r["SufficientByClass"][m]), ff]
            syn_surv.append(row)
            independent_model_failure = any(attribution(x) == "MODEL_EVIDENCE" for x in fs)
            latent = "NOT_ASSESSABLE" if fit == "PASS" and not independent_model_failure else "NOT_SUPPORTED"
            syn_latent.append([j["job_id"], label, j["generator_variant"], j["scale"], j["replicate"], latent, "individual predictive gates not recorded" if latent == "NOT_ASSESSABLE" else "training/model-evidence failure"])
        if j["branch"] == "N":
            row = [j["job_id"], label, j["scale"], j["replicate"], m, r["CandidateByClass"][m], fit]
            row += [pm_status(r, m, p) for p in PMS]
            row += ["NOT_ASSESSABLE", structural, ff, yn(r["SufficientByClass"][m])]
            nat_matrix.append(row)
            independent_model_failure = any(attribution(x) == "MODEL_EVIDENCE" for x in fs)
            latent = "NOT_ASSESSABLE" if fit == "PASS" and not independent_model_failure else "NOT_SUPPORTED"
            nat_latent.append([j["job_id"], label, j["scale"], j["replicate"], m, latent, "individual predictive/structural families not recorded" if latent == "NOT_ASSESSABLE" else "training/model-evidence failure"])
    survival_head = ["job_id", "ground_truth", "corpus_id", "generator_variant", "scale", "replicate", "fit", "PM1", "PM2", "PM4", "PM5", "PM6", "predictive_no_PM6", "structural", "final", "expected_model_first_failure"]
    write_tsv("SYNTHETIC_EXPECTED_CLASS_SURVIVAL.tsv", survival_head, syn_surv)
    write_tsv("SYNTHETIC_LATENT_RECOVERY.tsv", ["job_id", "ground_truth", "generator_variant", "scale", "replicate", "latent_recovery", "basis"], syn_latent)
    write_tsv("NATURAL_LANGUAGE_MODEL_MATRIX.tsv", ["job_id", "language", "scale", "replicate", "model", "candidate", "fit", "PM1", "PM2", "PM4", "PM5", "PM6", "predictive_no_PM6", "structural", "first_failure", "final"], nat_matrix)
    write_tsv("NATURAL_LANGUAGE_LATENT_ADEQUACY.tsv", ["job_id", "language", "scale", "replicate", "model", "latent_adequacy", "basis"], nat_latent)

    # Replicate stability: distribution of defensible first-failure stages.
    grouped = defaultdict(list)
    for j, r, m, label, fit, ff, fs, structural, pm6 in model_records:
        grouped[("SYNTHETIC" if j["branch"] == "S" else "NATURAL", label, j["generator_variant"], j["scale"], m)].append(ff)
    stab = []
    for key, vals in sorted(grouped.items()):
        counts = Counter(vals)
        modal, n = sorted(counts.items(), key=lambda x: (-x[1], x[0]))[0]
        stab.append(list(key) + [len(vals), modal, n, f"{n/len(vals):.6f}", ";".join(f"{k}:{v}" for k, v in sorted(counts.items()))])
    write_tsv("FAILURE_STAGE_STABILITY.tsv", ["branch", "corpus_label", "generator_variant", "scale", "model", "replicates", "modal_first_failure", "modal_count", "agreement_rate", "distribution"], stab)

    # Cascades count model executions, not corpora. Sequential PM survival is
    # unknowable after fitting, and is therefore NOT_RECORDED rather than 0.
    def cascade_rows(branch: str, label: str, recs: list) -> list[list]:
        total = len(recs)
        fitted = sum(x[4] == "PASS" for x in recs)
        pred = sum(x[1]["PredictivePassByClass"][x[2]] for x in recs)
        seval = fitted
        spass = sum(x[1]["StructuralPassByClass"][x[2]] for x in recs if x[4] == "PASS")
        final = sum(x[1]["SufficientByClass"][x[2]] for x in recs)
        vals = [("TOTAL_MODEL_EXECUTIONS", str(total)), ("TRAINING_SUCCESS", str(fitted))]
        vals += [(f"{p}_AVAILABLE_PASS", "NOT_RECORDED") for p in PMS]
        vals += [("PREDICTIVE_ADEQUATE", str(pred)), ("STRUCTURAL_EVALUABLE", str(seval)), ("STRUCTURAL_ADEQUATE", str(spass)), ("FINAL_ADEQUATE", str(final))]
        return [[branch, label, stage, value, "DERIVED_RECOMPUTATION"] for stage, value in vals]

    syn_cascade, nat_cascade = [], []
    syn_recs = [x for x in model_records if x[0]["branch"] == "S"]
    nat_recs = [x for x in model_records if x[0]["branch"] == "N"]
    syn_cascade += cascade_rows("SYNTHETIC", "ALL", syn_recs)
    for gt in MODELS:
        syn_cascade += cascade_rows("SYNTHETIC", gt, [x for x in syn_recs if x[3] == gt])
    nat_cascade += cascade_rows("NATURAL", "ALL", nat_recs)
    for lang in ("English", "Latin", "Sanskrit"):
        nat_cascade += cascade_rows("NATURAL", lang, [x for x in nat_recs if x[3] == lang])
    ch = ["branch", "subset", "stage", "count_or_status", "provenance"]
    write_tsv("SYNTHETIC_FAILURE_CASCADE.tsv", ch, syn_cascade)
    write_tsv("NATURAL_LANGUAGE_FAILURE_CASCADE.tsv", ch, nat_cascade)

    # Model diagnostics use only retained events.
    diag = []
    for m in MODELS:
        recs = [x for x in model_records if x[2] == m]
        fitted = sum(x[4] == "PASS" for x in recs)
        fc = Counter(y for x in recs for y in x[6])
        diag.append([m, len(recs), fitted, len(recs)-fitted, sum(x[1]["PM6ByClass"][m] for x in recs), sum(x[1]["StructuralPassByClass"][m] for x in recs), ";".join(f"{k}:{v}" for k, v in sorted(fc.items())), "NOT_ASSESSABLE", "per-PM and per-family evidence not recorded"])
    write_tsv("MODEL_CLASS_DIAGNOSTICS.tsv", ["model", "executions", "fit_success", "training_failed", "PM6_valid", "structural_pass", "failure_counts", "latent_diagnostic", "limitation"], diag)

    verdicts = [
        ["PM6_VETO_ROLE", "INCONCLUSIVE", "PM6 validity retained, PM6 gate pass and other PM gates not retained"],
        ["PRE_PM6_PREDICTIVE_INFORMATION", "NON_INFORMATIVE", "no individual PM1/PM2/PM4/PM5 result survived freeze"],
        ["SYNTHETIC_LATENT_RECOVERY", "NOT_ASSESSABLE", "expected-class predictive gates cannot be reconstructed"],
        ["NATURAL_LANGUAGE_LATENT_ADEQUACY", "NOT_ASSESSABLE", "same omission plus missing structural families"],
        ["ENGLISH_LATENT_MODEL_SET", "NOT_ASSESSABLE", "CF1 cannot be evaluated"],
        ["LATIN_LATENT_MODEL_SET", "NOT_ASSESSABLE", "CF1 cannot be evaluated"],
        ["SANSKRIT_LATENT_MODEL_SET", "NOT_ASSESSABLE", "CF1 cannot be evaluated"],
        ["M0_RECOVERY_DIAGNOSTIC", "PARTIAL", "fitting recorded; required predictive details absent"],
        ["M1_RECOVERY_DIAGNOSTIC", "PARTIAL", "fitting recorded; required predictive details absent"],
        ["M2_RECOVERY_DIAGNOSTIC", "PARTIAL", "fitting recorded; required predictive details absent"],
        ["M3_RECOVERY_LIMITATION", "MIXED", "training failures and protocol omissions coexist"],
        ["M4_RECOVERY_LIMITATION", "MIXED", "training failures and protocol omissions coexist"],
        ["M5_RECOVERY_LIMITATION", "MIXED", "negative exhaustion/generation failures plus omitted gates"],
        ["G1_V1_PRIMARY_FAILURE_SOURCE", "UNRESOLVED", "universal predictive failure cannot be decomposed from retained evidence"],
        ["TASK85B_READY", "SUPPORTED", "audit identifies mandatory observability and validation requirements"],
        ["TASK86C_A_VALID", "NOT_SUPPORTED", "meaningful required PM/F2 decomposition is impossible from frozen artifacts"],
    ]
    write_tsv("G1_FAILURE_SUMMARY.tsv", ["verdict", "value", "basis"], verdicts)

    reqs = [
        ["G1V2-OBS-001", "PM1/PM2/PM4/PM5", "M0-M5", "all 576 synthetic jobs lack retained values/gates", "all 96 natural jobs lack retained values/gates", "yes", "persist candidate, baselines, thresholds, finite flags, values and gate outcomes", "round-trip audit reconstructs every gate without fitting"],
        ["G1V2-OBS-002", "PM6", "M0-M5", "validity retained but value/gate omitted", "same", "yes", "separate sampler availability, finite score and acceptance gate", "known available/unavailable/pass/fail fixtures"],
        ["G1V2-OBS-003", "F2/structural", "M0-M5", "only aggregate boolean retained", "same", "yes", "persist individual metrics, families, thresholds and reachability", "family-specific synthetic failure fixtures"],
        ["G1V2-COMP-001", "training", "M3;M4", "training failed in frozen controls", "training failed in frozen controls", "no", "report convergence/cap/state/operation diagnostics separately from class adequacy", "known-correct M3/M4 controls at every scale"],
        ["G1V2-NEG-001", "PM6 negative construction", "M0-M5", "negative exhaustion is widespread", "negative exhaustion is widespread", "yes", "prevalidate constructibility and define non-veto unavailable semantics", "saturated alphabet/vocabulary controls"],
        ["G1V2-GEN-001", "generation", "M0-M5", "numeric instability retained", "numeric instability retained", "no", "persist convergence, CV and per-scale failure separately", "known-correct generators with deterministic limits"],
        ["G1V2-PROV-001", "results freeze", "M0-M5", "post-hoc audit cannot recover intermediate evidence", "same", "yes", "manifest every per-job intermediate artifact by hash", "delete-free reproducibility audit before freeze"],
    ]
    write_tsv("G1_V2_REQUIREMENTS.tsv", ["defect_id", "affected_stage", "affected_models", "synthetic_evidence", "natural_language_evidence", "universal", "required_property_of_G1_v2", "proposed_validation_test"], reqs)

    total_exec = len(model_records)
    fit_ok = sum(x[4] == "PASS" for x in model_records)
    train_fail = total_exec - fit_ok
    jobs_train_fail = len({x[0]["job_id"] for x in model_records if x[4] != "PASS"})
    pm6_valid = sum(x[1]["PM6ByClass"][x[2]] for x in model_records)
    struct_pass = sum(x[1]["StructuralPassByClass"][x[2]] for x in model_records)
    fail_counts = Counter(y for x in model_records for y in x[6])
    class_fit = {m: sum(x[4] == "PASS" for x in model_records if x[2] == m) for m in MODELS}

    design = """# Task86C-a design\n\nTask86C-a is a read-only post-hoc diagnostic audit of frozen Task86C.  The unit is job × model class; synthetic rows retain generator, scale and replicate.  The generator reads only Task86C manifests, ledger, frozen aggregate manifest and result JSON.  It never opens a corpus, invokes fitting/generation, changes a threshold, or accesses Voynich data.\n\nStatuses are not collapsed: `TRAINING_FAILED`, `UNAVAILABLE_NEGATIVE_SET_EXHAUSTION`, `NOT_REACHED`, and `NOT_RECORDED` remain distinct. Absence of `TRAINING_FAILED` establishes that Stage D received a fitted model; it does not establish adequacy. `PM6ByClass` establishes score validity only, not PM6 gate passage. `PredictivePassByClass` and `StructuralPassByClass` are retained aggregate gates and are not used to invent individual metric outcomes.\n\nAll aggregation here is `DERIVED_RECOMPUTATION`. No `NEW_DIAGNOSTIC_COMPUTATION` was run. Deterministic regeneration command:\n\n```sh\npython3 research/phase3/task86c-a/generate.py\n```\n\nThe audit terminal state is evidence-insufficient because the required PM-by-PM and F2 decompositions and PM6 ablation cannot be obtained without prohibited rerunning.\n"""
    (OUT / "TASK86C_A_DESIGN.md").write_text(design)

    requirements_md = """# Task85b / G1-v2 requirements\n\nTask85b is ready to be specified, but not because Task86C-a recovered a latent winner. It is ready because the audit localized an observability failure and retained several independent computational/protocol failure classes.\n\nRequired properties:\n\n1. Persist every PM value, baseline, frozen threshold, finite/availability flag, and per-baseline gate outcome for every candidate selected into HELDOUT.\n2. Treat PM6 construction availability, PM6 score validity, and PM6 acceptance as separate fields; validate saturated controls before confirmation.\n3. Persist each F2 metric/family result, generation-scale result and `NOT_REACHED` reason.\n4. Separate model evidence from induction caps, convergence failures, generation failures and protocol vetoes.\n5. Require known-correct M0–M5 controls to survive an end-to-end pre-freeze audit, including scale and replicate stability.\n6. Hash all per-job intermediate artifacts in the frozen result manifest and prove that decision paths can be regenerated without model execution.\n7. Do not tune thresholds from these controls and do not treat CF1/CF2 as scientific classifications.\n\nConcrete defects and validation tests are in `G1_V2_REQUIREMENTS.tsv`.\n"""
    (OUT / "TASK85B_REQUIREMENTS.md").write_text(requirements_md)

    report = f"""# Task86C-a — G1 failure decomposition and latent-model adequacy audit\n\n## Outcome\n\nTask86C's `NONE` means that none of {total_exec} job × model executions satisfied the conjunction of the frozen aggregate predictive, structural, complexity and failure gates. It does **not** establish that M0–M5 fail to describe the controls. The frozen results omitted the individual evidence needed to make that inference.\n\nIntegrity checks account for {len(results)} jobs and their ledger hashes; frozen aggregate artifact hashes and the Task86C terminal marker were verified. Integrity issues: {"; ".join(integrity) if integrity else "NONE"}. No corpus file was opened, no model was fitted/generated, and no Voynich path was accessed.\n\n## Failure accounting\n\nThere were {total_exec} model executions. {fit_ok} reached Stage D with a fitted model and {train_fail} recorded `TRAINING_FAILED`, affecting {jobs_train_fail} jobs. Fit-success counts are {", ".join(f"{m}={class_fit[m]}" for m in MODELS)}. All {total_exec} aggregate predictive gates failed; {struct_pass} aggregate structural gates passed; no final model was sufficient. Retained independent failure events are {", ".join(f"{k}={v}" for k, v in sorted(fail_counts.items()))}. Training failure is computational evidence, not model-class inadequacy.\n\nPM6 produced a valid score in {pm6_valid}/{total_exec} executions. The remaining cells are unavailable (mostly explicitly `NEGATIVE_EXHAUSTED`) or not reached after training failure. Crucially, Task86C did not retain whether any valid PM6 value passed its threshold.\n\n## PM-by-PM answers\n\nFor PM1, PM2, PM4 and PM5, values, finiteness and gate outcomes are `NOT_RECORDED` for every fitted execution; rates therefore are not estimable. For PM6, validity/availability is recorded, but value and pass/fail are not. Consequently `PREDICTIVE_ADEQUACY_WITHOUT_PM6`, CF1 and CF2 are `NOT_ASSESSABLE`, and the number of FAIL→PASS* changes is unknown.\n\n`PM6_VETO_ROLE = INCONCLUSIVE`. It cannot be called the sole universal veto: PM6 gate passage is absent, while all predictive aggregates fail even in the {pm6_valid} executions with a valid PM6 score. Nor can an independent universal PM be named because their individual outcomes were discarded. `PRE_PM6_PREDICTIVE_INFORMATION = NON_INFORMATIVE` in the frozen artifact layer.\n\n## Synthetic controls\n\nExpected-class fitting and aggregate paths are listed by generator, scale and replicate. M0/M1/M2/M5 have no recorded training failures; M3/M4 have substantial induction failures. However the correct models' PM1/2/4/5 values, PM6 gate and structural families were not frozen. Thus latent recovery rates, context/depth recovery, fitted-vs-generating M0 distribution, M3/M4 state/operation diagnostics and M5 coverage/rule diagnostics are not computable without new experiments.\n\n`SYNTHETIC_LATENT_RECOVERY = NOT_ASSESSABLE`; observed historical recovery remains 0%. M0/M1/M2 recovery diagnostics are `PARTIAL` only in the narrow sense that fitting survival is known while adequacy is not. M3/M4/M5 limitations are `MIXED`: retained induction/protocol/generation failures coexist with missing measurement evidence. Lower/higher-class discrimination is likewise not assessable.\n\n## Natural-language controls\n\nEnglish, Latin and Sanskrit each have a complete job × M0–M5 bookkeeping matrix. It establishes fitting survival and aggregate rejection, but cannot show which models passed PM1/2/4/5 or which structural family failed. Therefore:\n\n- `ENGLISH_LATENT_MODEL_SET = NOT_ASSESSABLE`\n- `LATIN_LATENT_MODEL_SET = NOT_ASSESSABLE`\n- `SANSKRIT_LATENT_MODEL_SET = NOT_ASSESSABLE`\n- `NATURAL_LANGUAGE_LATENT_ADEQUACY = NOT_ASSESSABLE`\n\nThe evidence supports terminal interpretation C: existing artifacts are insufficient. Repeating `English = NONE` would not prove that M0–M5 do not model English; the same applies to Latin and Sanskrit.\n\n## Structural, scale and replicate evidence\n\nStructural evaluation was executed independently of predictive passage for fitted models, so aggregate false is retained as `FAIL`, not `NOT_REACHED`; training failures are `NOT_REACHED`. Individual F2 values/families/thresholds are `NOT_RECORDED`. Hence one-family failures and closeness cannot be identified.\n\nScale and replicate tables report stable distributions of the defensible first failure (`TRAINING` or aggregate `PREDICTIVE_GATE`). They cannot reveal whether a specific PM improves with scale, whether PM6 worsens, or whether dependency depth recovers. No strong metric-specific scale conclusion is possible.\n\n## Required verdicts and next task\n\nThe primary G1-v1 failure source is `UNRESOLVED`: measurement observability is certainly defective, but retained independent induction, negative-construction and generation failures prevent attribution solely to protocol. `TASK85B_READY = SUPPORTED` because concrete observability and known-correct-control requirements can be stated. `TASK86C_A_VALID = NOT_SUPPORTED` because the definition of done requires decompositions that the authoritative frozen source does not contain.\n\nThe first universal recorded rejection point is the aggregate `PREDICTIVE_GATE`; the responsible PM is not recorded. Exact/minimal recovery are zero because every aggregate predictive gate failed and sufficiency requires predictive passage, structural passage, multi-scale sufficiency and no failure class.\n\n## Direct answers to the 32 report questions\n\n1. `NONE` is failure of the conjunction, not proof of model-space inadequacy. 2. No job was absent before fitting began; {jobs_train_fail} jobs contain at least one class training failure. 3. {fit_ok} model executions fitted. 4–8. PM1/2/4/5 rates are `NOT_RECORDED`; PM6 validity is {pm6_valid}/{total_exec}, but PM6 pass rate is `NOT_RECORDED`. 9. PM6 role is inconclusive. 10. Training, negative construction, generation and complexity failures exist independently as named events, though their temporal relation to hidden PM gates is limited. 11–16. Expected M0–M5 paths are in the survival table; individual PM causes are missing. 17. Synthetic latent recovery count is not assessable. 18. All predictive aggregates failed. 19. Natural fit outcomes are in the matrix. 20. PM1/2/4/5 passage is not recorded. 21. Only aggregate structural failures are known. 22–24. All latent model sets are not assessable. 25. No. 26. Measurement observability is a major defect, but primary causation remains unresolved. 27. Only recorded aggregate/failure scale effects can be tabulated. 28. Their replicate agreement is tabulated. 29. Aggregate predictive gate. 30. M3/M4 training, widespread negative exhaustion, numeric generation instability and some complexity-unbounded events. 31. G1-v2 properties are enumerated. 32. Yes, for Task85b requirements, not for a new adequacy verdict.\n\nTASK86C_A_EVIDENCE_INSUFFICIENT\n"""
    (OUT / "TASK86C_A_REPORT.md").write_text(report)

    (OUT / "TASK86C_A_EVIDENCE_INSUFFICIENT").write_text(
        "TASK86C_A_EVIDENCE_INSUFFICIENT\n"
        "source_terminal_marker=TASK86C_G1_CONTROL_VALIDATION_FROZEN\n"
        "source_jobs=672\n"
        "new_fitting_or_generation=false\n"
    )

    # Manifest is last and excludes itself to avoid a recursive hash.
    artifact_names = sorted(p.name for p in OUT.iterdir() if p.is_file() and p.name != "TASK86C_A_RESULTS_MANIFEST.json")
    manifest = {
        "version": "task86c-a-posthoc-v1",
        "analysis_class": "DERIVED_RECOMPUTATION",
        "source": "TASK86C_G1_CONTROL_VALIDATION_FROZEN",
        "source_jobs": len(results),
        "model_executions": total_exec,
        "integrity_status": "PASS" if not integrity else "FAIL",
        "integrity_issues": integrity,
        "voynich_corpus_access": False,
        "new_fitting_or_generation": False,
        "terminal_marker": "TASK86C_A_EVIDENCE_INSUFFICIENT",
        "artifacts": {name: sha(OUT / name) for name in artifact_names},
    }
    (OUT / "TASK86C_A_RESULTS_MANIFEST.json").write_text(json.dumps(manifest, indent=2, sort_keys=True) + "\n")


if __name__ == "__main__":
    main()
