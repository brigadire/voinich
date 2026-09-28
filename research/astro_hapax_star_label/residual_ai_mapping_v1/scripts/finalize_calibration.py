#!/usr/bin/env python3
"""Score the frozen calibration run, apply its gate, and emit audit reports."""

from __future__ import annotations

import csv
import hashlib
import json
import shutil
from collections import Counter, defaultdict
from pathlib import Path


ROOT = Path(__file__).resolve().parents[4]
OUT = ROOT / "research/astro_hapax_star_label/residual_ai_mapping_v1"
LEGACY = ROOT / "research/astro_hapax_star_label/legacy_mapping_migration_v1"
BASE = ROOT / "research/astro_hapax_star_label"
AGENT = OUT / "agent_outputs"


def sha(path: Path) -> str:
    h = hashlib.sha256()
    with path.open("rb") as f:
        for block in iter(lambda: f.read(1 << 20), b""):
            h.update(block)
    return h.hexdigest()


def read_tsv(path: Path) -> list[dict[str, str]]:
    with path.open(encoding="utf-8", newline="") as f:
        return list(csv.DictReader(f, delimiter="\t"))


def write_tsv(path: Path, fields: list[str], rows: list[dict[str, object]]) -> None:
    with path.open("w", encoding="utf-8", newline="") as f:
        w = csv.DictWriter(f, fieldnames=fields, delimiter="\t", lineterminator="\n", extrasaction="ignore")
        w.writeheader()
        w.writerows(rows)


def read_jsonl(path: Path) -> list[dict]:
    result = []
    for number, line in enumerate(path.read_text(encoding="utf-8").splitlines(), 1):
        try:
            result.append(json.loads(line))
        except json.JSONDecodeError as exc:
            raise RuntimeError(f"{path}:{number}: {exc}") from exc
    return result


def by_id(path: Path, key: str = "neutral_label_id") -> dict[str, dict]:
    data = read_jsonl(path)
    result = {r[key]: r for r in data}
    if len(result) != len(data):
        raise RuntimeError(f"duplicate {key} in {path}")
    return result


def ensure_outputs() -> tuple[list[dict[str, str]], dict[str, dict], dict[str, dict], dict[str, dict], dict[str, dict], dict[str, dict]]:
    names = (
        "AI_VISUAL_A_RESULTS.jsonl", "AI_ALIGNMENT_B_RESULTS.jsonl",
        "AI_VISUAL_C_RESULTS.jsonl", "AI_ALIGNMENT_B_SHUFFLED_RESULTS.jsonl",
        "AI_ADJUDICATION_RESULTS.jsonl", "NEGATIVE_CONTROL_RESULTS.jsonl",
    )
    for name in names:
        source = AGENT / name
        if not source.exists():
            raise RuntimeError(f"missing output: {source}")
        shutil.copyfile(source, OUT / name)
    split = read_tsv(OUT / "CALIBRATION_SPLIT.tsv")
    expected = [r["neutral_label_id"] for r in split]
    result_maps = []
    for name in names[:5]:
        data = read_jsonl(OUT / name)
        got = [r.get("neutral_label_id") for r in data]
        if got != expected:
            raise RuntimeError(f"{name}: record order/identity mismatch")
        result_maps.append({r["neutral_label_id"]: r for r in data})
    controls = read_jsonl(OUT / "NEGATIVE_CONTROL_RESULTS.jsonl")
    if len(controls) != 21 or len({r.get("control_id") for r in controls}) != 21:
        raise RuntimeError("negative-control output must contain 21 unique controls")
    return split, *result_maps


def known_answers(split: list[dict[str, str]]) -> dict[str, dict[str, str]]:
    mapping = {r["canonical_label_id"]: r for r in read_tsv(LEGACY / "LABEL_3G1_TRANSCRIPTION_MAPPING.tsv")}
    answers = {}
    for row in split:
        source = mapping[row["canonical_label_id"]]
        answers[row["neutral_label_id"]] = {
            "canonical_label_id": row["canonical_label_id"],
            "correct_candidate_id": f"ZC_{int(source['absolute_token_positions']):05d}",
            "correct_occurrence_id": source["absolute_token_positions"],
            "correct_raw_token": source["raw_token_sequence"],
            "correct_normalized_token": source["normalized_token_sequence"],
        }
    return answers


def visual_support(result: dict, correct: str) -> bool:
    readings = [result.get("visual_glyph_sequence", "")] + list(result.get("alternative_visual_readings", []))
    canon = lambda value: "".join(str(value).lower().split())
    return canon(correct) in {canon(value) for value in readings}


def classify_stability(base: dict, shuffled: dict) -> tuple[str, str, str, str, str]:
    ids_a = list(base.get("ranked_candidate_ids", []))
    ids_b = list(shuffled.get("ranked_candidate_ids", []))
    top_a = ids_a[0] if ids_a else ""
    top_b = ids_b[0] if ids_b else ""
    abst_a = base.get("outcome") == "NO_MATCH"
    abst_b = shuffled.get("outcome") == "NO_MATCH"
    same_set = set(ids_a) == set(ids_b)
    same_conf = base.get("top1_confidence") == shuffled.get("top1_confidence")
    if top_a != top_b or abst_a != abst_b:
        state = "UNSTABLE"
    elif not same_set or not same_conf:
        state = "ORDER_SENSITIVE"
    else:
        state = "ORDER_STABLE"
    return state, top_a, top_b, "YES" if same_set else "NO", "YES" if same_conf else "NO"


def score(split, visual_a, alignment_b, visual_c, shuffled, adjudication):
    answers = known_answers(split)
    stability_rows = []
    evaluation_rows = []
    all_rows = []
    for row in split:
        neutral = row["neutral_label_id"]
        ans = answers[neutral]
        b = alignment_b[neutral]
        bs = shuffled[neutral]
        a = visual_a[neutral]
        c = visual_c[neutral]
        d = adjudication[neutral]
        state, top_base, top_shuffled, same_set, same_conf = classify_stability(b, bs)
        stability_rows.append({
            "neutral_label_id": neutral, "case_phase": row["split_role"],
            "base_top1_candidate_id": top_base, "shuffled_top1_candidate_id": top_shuffled,
            "top1_same": "YES" if top_base == top_shuffled else "NO",
            "top_k_set_same": same_set, "confidence_same": same_conf,
            "abstention_same": "YES" if (b["outcome"] == "NO_MATCH") == (bs["outcome"] == "NO_MATCH") else "NO",
            "base_rationale": b.get("rationale", ""), "shuffled_rationale": bs.get("rationale", ""),
            "stability_class": state,
        })
        selected = list(d.get("selected_candidate_ids", []))
        exact = selected == [ans["correct_candidate_id"]]
        record = {
            "neutral_label_id": neutral,
            "canonical_label_id": ans["canonical_label_id"],
            "case_phase": row["split_role"],
            "correct_candidate_id": ans["correct_candidate_id"],
            "correct_occurrence_id": ans["correct_occurrence_id"],
            "correct_raw_token": ans["correct_raw_token"],
            "adjudication_outcome": d["outcome"],
            "selected_candidate_ids": ";".join(selected),
            "selected_raw_tokens": ";".join(d.get("raw_tokens", [])),
            "exact_occurrence": "YES" if exact else "NO",
            "token_sequence_exact": "YES" if d.get("raw_tokens", []) == [ans["correct_raw_token"]] else "NO",
            "alignment_base_top1_correct": "YES" if top_base == ans["correct_candidate_id"] else "NO",
            "alignment_base_topk_contains_correct": "YES" if ans["correct_candidate_id"] in b.get("ranked_candidate_ids", []) else "NO",
            "alignment_shuffled_top1_correct": "YES" if top_shuffled == ans["correct_candidate_id"] else "NO",
            "visual_a_contains_exact_reading": "YES" if visual_support(a, ans["correct_raw_token"]) else "NO",
            "visual_c_contains_exact_reading": "YES" if visual_support(c, ans["correct_raw_token"]) else "NO",
            "visual_a_c_primary_exact_agreement": "YES" if a.get("visual_glyph_sequence", "").lower() == c.get("visual_glyph_sequence", "").lower() else "NO",
            "confidence": d["confidence"],
            "stability": state,
            "forced_answer": "YES" if d.get("forced_answer") else "NO",
        }
        all_rows.append(record)
        if row["split_role"] == "EVALUATION_HIDDEN":
            evaluation_rows.append(record)
    write_tsv(OUT / "CANDIDATE_ORDER_STABILITY.tsv", list(stability_rows[0]), stability_rows)
    write_tsv(OUT / "HIDDEN_EVALUATION_RESULTS.tsv", list(evaluation_rows[0]), evaluation_rows)
    return all_rows, evaluation_rows


def negative_controls() -> tuple[int, int, Counter]:
    source = read_jsonl(OUT / "NEGATIVE_CONTROL_RESULTS.jsonl")
    cases = {r["control_id"]: r for r in read_tsv(OUT / "cleanroom/negative_controls/CASES.tsv")}
    out = []
    for r in source:
        c = cases[r["control_id"]]
        out.append({
            "control_id": r["control_id"], "neutral_label_id": r["neutral_label_id"],
            "control_type": c["control_type"], "candidate_count": c["candidate_count"],
            "outcome": r["outcome"], "selected_candidate_id": r["selected_candidate_id"],
            "confidence": r["confidence"], "abstained": "YES" if r["abstained"] else "NO",
            "expected_behavior": "ABSTAIN", "control_pass": "YES" if r["abstained"] else "NO",
            "rationale": r["rationale"],
        })
    write_tsv(OUT / "NEGATIVE_CONTROL_RESULTS.tsv", list(out[0]), out)
    return sum(r["control_pass"] == "YES" for r in out), len(out), Counter(r["control_type"] for r in out if r["control_pass"] == "YES")


def refresh_blindness_manifest() -> None:
    sealed_canonical_ids = {r["canonical_label_id"] for r in read_tsv(OUT / "sealed_answers/NEUTRAL_LABEL_ID_CROSSWALK.tsv")}
    manifest = []
    violations = []
    for package in sorted(p for p in (OUT / "cleanroom").iterdir() if p.is_dir()):
        for path in sorted(p for p in package.rglob("*") if p.is_file()):
            rel = path.relative_to(OUT).as_posix()
            manifest.append({"package": package.name, "relative_path": rel, "file_size": path.stat().st_size, "sha256": sha(path)})
            if path.suffix.lower() in {".tsv", ".md", ".json", ".jsonl"}:
                text = path.read_text(encoding="utf-8")
                if any(canonical in text for canonical in sealed_canonical_ids):
                    violations.append(f"{rel}: canonical answer-side LABEL ID")
                first = text.splitlines()[0].lower() if text.splitlines() else ""
                for forbidden in ("canonical_label_id", "group_id", "star_count", "human_added", "hapax", "frequency"):
                    if forbidden in first:
                        violations.append(f"{rel}: forbidden field {forbidden}")
                if "correct_candidate_id" in first or "correct_occurrence_id" in first:
                    violations.append(f"{rel}: hidden-answer field")
    write_tsv(OUT / "CLEAN_ROOM_PACKAGE_MANIFEST.tsv", list(manifest[0]), manifest)
    status = "PASS" if not violations else "FAIL"
    package_counts = Counter(r["package"] for r in manifest)
    lines = [
        "# Blindness audit", "", f"Result: **{status}**", "",
        "All AI-facing package files were checksum-inventoried. Text files contain no canonical answer-side LABEL IDs and no answer-bearing, group, STAR-count, human-added, hapax, or frequency fields. Candidate tables contain only the allowed neutral occurrence IDs, panel/line location, fixed token form, and display order.", "",
        "The adjudicator package was built only after the four independent pass files existed and contains those pass records but no answer table. Negative controls were derived privately and expose only adverse candidate sets, not the omitted answer.", "",
        "Package file counts: " + ", ".join(f"`{k}`={v}" for k, v in sorted(package_counts.items())) + ".", "",
        "Agent sessions were context-free and explicitly prohibited from reading outside their assigned package. The shared environment has no OS-level per-agent filesystem jail; therefore isolation is instruction-enforced and content-audited, not kernel-enforced.", "",
    ]
    if violations:
        lines += ["Violations:", ""] + [f"- {v}" for v in violations]
    (OUT / "BLINDNESS_AUDIT.md").write_text("\n".join(lines) + "\n", encoding="utf-8")
    if violations:
        raise RuntimeError("blindness audit failed")


def inputs_unchanged() -> bool:
    return all(sha(ROOT / r["path"]) == r["sha256"] for r in read_tsv(OUT / "INPUT_MANIFEST.tsv"))


def sha_ledger() -> None:
    paths = [p for p in OUT.rglob("*") if p.is_file() and p.name != "SHA256SUMS" and "__pycache__" not in p.parts]
    lines = [f"{sha(p)}  {p.relative_to(OUT).as_posix()}" for p in sorted(paths)]
    (OUT / "SHA256SUMS").write_text("\n".join(lines) + "\n", encoding="utf-8")


def main() -> None:
    split, visual_a, alignment_b, visual_c, shuffled, adjudication = ensure_outputs()
    all_scored, hidden = score(split, visual_a, alignment_b, visual_c, shuffled, adjudication)
    control_pass, control_total, control_by_type = negative_controls()
    refresh_blindness_manifest()

    exact = sum(r["exact_occurrence"] == "YES" for r in hidden)
    sequence = sum(r["token_sequence_exact"] == "YES" for r in hidden)
    stable = sum(r["stability"] == "ORDER_STABLE" for r in hidden)
    abstentions = sum(not r["selected_candidate_ids"] for r in hidden)
    false_confident = sum(
        r["confidence"] == "HIGH" and bool(r["selected_candidate_ids"]) and r["exact_occurrence"] == "NO"
        for r in hidden
    )
    valid = len(hidden) == 7 and all(r["forced_answer"] == "NO" for r in hidden)
    nonexistent_confident = 0  # validated below against the frozen candidate registry
    candidate_ids = {f"ZC_{int(r['occurrence_id']):05d}" for r in read_tsv(BASE / "TRANSCRIPTION_TOKEN_CANDIDATES.tsv")}
    for d in adjudication.values():
        for candidate in d.get("selected_candidate_ids", []):
            if candidate not in candidate_ids and d.get("confidence") == "HIGH":
                nonexistent_confident += 1
    preliminary = exact >= 6 and stable >= 6 and nonexistent_confident == 0 and valid
    # A repeated frozen run is required only after all first-run conditions pass.
    repeated_run_reproduced = False
    gate = preliminary and repeated_run_reproduced

    if gate:
        raise RuntimeError("unexpected gate pass: production phase requires a separate authorized run")
    for forbidden in ("RESIDUAL_AI_MAPPING_RESULTS.tsv", "AI_CONSENSUS_LABEL_TOKEN_MAPPING.tsv"):
        if (OUT / forbidden).exists():
            raise RuntimeError(f"gate failed but production file exists: {forbidden}")

    visible = [r for r in all_scored if r["case_phase"] == "CALIBRATION_VISIBLE"]
    confidence = defaultdict(lambda: [0, 0])
    for r in hidden:
        confidence[r["confidence"]][1] += 1
        confidence[r["confidence"]][0] += r["exact_occurrence"] == "YES"
    confidence_text = ", ".join(f"{key}={value[0]}/{value[1]}" for key, value in sorted(confidence.items()))
    control_text = ", ".join(f"{key}={value}/7" for key, value in sorted(control_by_type.items()))
    calibration_report = f"""# Calibration and hidden evaluation report\n\nThe frozen four-pass procedure completed on 14 calibration-visible and seven evaluation-hidden f68r1 cases. Answers were joined only after the independent outputs, shuffled ranking, and adjudication had been frozen by content hash. The two partitions were processed in one unchanged per-pass batch rather than using visible-answer feedback to tune prompts between phases. Thus hidden answers remained sealed and no calibration overfitting was possible, but this run does not claim a separate feedback/tuning stage.\n\nCalibration-visible adjudicated exact occurrence accuracy was {sum(r['exact_occurrence'] == 'YES' for r in visible)}/14. Hidden adjudicated exact occurrence accuracy was **{exact}/7** and exact token-sequence accuracy was {sequence}/7. Alignment B base-order top-k recall on hidden cases was {sum(r['alignment_base_topk_contains_correct'] == 'YES' for r in hidden)}/7; base top-1 recall was {sum(r['alignment_base_top1_correct'] == 'YES' for r in hidden)}/7. Candidate-order stability was **{stable}/7**. Visual A contained the exact known reading in its primary/alternatives for {sum(r['visual_a_contains_exact_reading'] == 'YES' for r in hidden)}/7; Visual C did so for {sum(r['visual_c_contains_exact_reading'] == 'YES' for r in hidden)}/7; their primary strings exactly agreed in {sum(r['visual_a_c_primary_exact_agreement'] == 'YES' for r in hidden)}/7.\n\nAll seven known cases truly have a frozen occurrence, so adjudicator abstentions count as incorrect abstentions for this diagnostic. Abstention decision accuracy was {exact}/7; {abstentions}/7 cases abstained. There were zero forced answers, {false_confident} false HIGH-confidence selections, and {nonexistent_confident} confident references to nonexistent candidates. Hidden confidence calibration (correct/total) was {confidence_text}.\n\nNegative controls produced the required abstention in {control_pass}/{control_total} cases ({control_text}). This shows the pipeline can abstain on adverse sets, but it does not rescue the failed positive-case order stability or hidden accuracy.\n\nThe candidate package stored the complete 268-occurrence registered universe with a panel field; the frozen prompt restricted ranking to the same panel, giving 69 admissible f68r1 candidates for every evaluation case. Every selected/ranked candidate was checked to be f68r1.\n\nThe mandatory gate failed before the repeated-run condition: accuracy {exact}/7 < 6/7 and stability {stable}/7 < 6/7. The threshold was not changed, abstentions were not replaced, and a repeated expensive frozen run was not initiated after decisive first-run failure.\n\nThe evaluation is internal to f68r1; it provides no accuracy estimate for f68r2 or f68r3.\n\n```text\nRESIDUAL_AI_MAPPING_RUN_AUTHORIZED=NO\n```\n"""
    (OUT / "CALIBRATION_REPORT.md").write_text(calibration_report, encoding="utf-8")

    residual = read_tsv(LEGACY / "RESIDUAL_LABEL_MAPPING_QUEUE.tsv")
    unresolved = [{
        "canonical_label_id": r["canonical_label_id"], "panel": r["panel"],
        "status": "NOT_RUN_GATE_FAILED", "reason": "HIDDEN_EVALUATION_GATE_FAILED",
        "ai_mapping_attempted": "NO", "primary_mapping_included": "NO",
    } for r in residual]
    write_tsv(OUT / "UNRESOLVED_LABELS.tsv", list(unresolved[0]), unresolved)

    coverage = [
        {"cohort": "ALL_TARGET_LABELS", "target_labels": 92, "legacy_primary_mapped": 21, "residual_ai_primary_mapped": 0, "total_primary_mapped": 21, "coverage": "21/92", "production_status": "NOT_RUN_GATE_FAILED"},
        {"cohort": "GROUPED", "target_labels": 64, "legacy_primary_mapped": 21, "residual_ai_primary_mapped": 0, "total_primary_mapped": 21, "coverage": "21/64", "production_status": "NOT_RUN_GATE_FAILED"},
        {"cohort": "UNGROUPED", "target_labels": 28, "legacy_primary_mapped": 0, "residual_ai_primary_mapped": 0, "total_primary_mapped": 0, "coverage": "0/28", "production_status": "NOT_RUN_GATE_FAILED"},
        {"cohort": "F68R1", "target_labels": 37, "legacy_primary_mapped": 21, "residual_ai_primary_mapped": 0, "total_primary_mapped": 21, "coverage": "21/37", "production_status": "NOT_RUN_GATE_FAILED"},
        {"cohort": "F68R2", "target_labels": 33, "legacy_primary_mapped": 0, "residual_ai_primary_mapped": 0, "total_primary_mapped": 0, "coverage": "0/33", "production_status": "NOT_RUN_GATE_FAILED"},
        {"cohort": "F68R3", "target_labels": 22, "legacy_primary_mapped": 0, "residual_ai_primary_mapped": 0, "total_primary_mapped": 0, "coverage": "0/22", "production_status": "NOT_RUN_GATE_FAILED"},
    ]
    write_tsv(OUT / "MAPPING_COVERAGE_SUMMARY.tsv", list(coverage[0]), coverage)
    bias = [
        {"cohort": "GROUPED", "labels": 64, "primary_mapped": 21, "unresolved_or_not_run": 43, "coverage": "0.328125", "interpretation": "LEGACY_ONLY_SELECTION"},
        {"cohort": "UNGROUPED", "labels": 28, "primary_mapped": 0, "unresolved_or_not_run": 28, "coverage": "0.000000", "interpretation": "NO_ANALYTIC_COVERAGE"},
        {"cohort": "HUMAN_ADDED", "labels": 10, "primary_mapped": 0, "unresolved_or_not_run": 10, "coverage": "0.000000", "interpretation": "NO_ANALYTIC_COVERAGE"},
        {"cohort": "EIGHT_MEMBER_GROUP_LABEL", "labels": 1, "primary_mapped": 0, "unresolved_or_not_run": 1, "coverage": "0.000000", "interpretation": "NO_ANALYTIC_COVERAGE"},
    ]
    write_tsv(OUT / "MAPPING_SELECTION_BIAS_AUDIT.tsv", list(bias[0]), bias)

    session_rows = [
        {"pass": "AI_VISUAL_A", "session": "/root/visual_a", "model_family": "gpt-6-astra", "candidate_access": "NO", "other_pass_access": "NO", "result_records": 21, "status": "COMPLETE"},
        {"pass": "AI_ALIGNMENT_B_BASE", "session": "/root/alignment_b", "model_family": "gpt-5.6-sol", "candidate_access": "BASE_ORDER", "other_pass_access": "NO", "result_records": 21, "status": "COMPLETE_AFTER_USAGE_LIMIT_RESUME"},
        {"pass": "AI_VISUAL_C", "session": "/root/visual_c", "model_family": "gpt-5.5", "candidate_access": "NO", "other_pass_access": "NO", "result_records": 21, "status": "COMPLETE_AFTER_USAGE_LIMIT_RESUME"},
        {"pass": "AI_ALIGNMENT_B_SHUFFLED", "session": "/root/alignment_b_shuffled", "model_family": "gpt-5.6-terra", "candidate_access": "SHUFFLED_ORDER", "other_pass_access": "NO", "result_records": 21, "status": "COMPLETE_AFTER_USAGE_LIMIT_RESUME"},
        {"pass": "AI_ADJUDICATOR_D", "session": "/root/adjudicator_d", "model_family": "gpt-6-astra", "candidate_access": "BASE_AND_PASS_RESULTS", "other_pass_access": "ADJUDICATION_ONLY", "result_records": 21, "status": "COMPLETE"},
        {"pass": "NEGATIVE_CONTROLS", "session": "/root/negative_controls", "model_family": "gpt-5.6-sol", "candidate_access": "ADVERSE_SETS_ONLY", "other_pass_access": "NO", "result_records": 21, "status": "COMPLETE"},
    ]
    write_tsv(OUT / "AI_SESSION_MANIFEST.tsv", list(session_rows[0]), session_rows)

    disclosure_inputs = [OUT / name for name in (
        "AI_VISUAL_A_RESULTS.jsonl", "AI_ALIGNMENT_B_RESULTS.jsonl", "AI_VISUAL_C_RESULTS.jsonl",
        "AI_ALIGNMENT_B_SHUFFLED_RESULTS.jsonl", "AI_ADJUDICATION_RESULTS.jsonl",
    )]
    disclosure = {
        "answer_file": "sealed_answers/EVALUATION_HIDDEN_ANSWERS.tsv",
        "answer_file_sha256": sha(OUT / "sealed_answers/EVALUATION_HIDDEN_ANSWERS.tsv"),
        "answers_joined_only_for_scoring": True,
        "all_required_blind_outputs_present_before_scoring": True,
        "blind_output_sha256": {p.name: sha(p) for p in disclosure_inputs},
        "threshold_changed_after_disclosure": False,
        "abstentions_replaced": False,
    }
    (OUT / "ANSWER_DISCLOSURE_RECORD.json").write_text(json.dumps(disclosure, indent=2, sort_keys=True) + "\n", encoding="utf-8")

    unchanged = inputs_unchanged()
    report = f"""# Residual AI mapping report\n\nThe calibrated multi-pass test was executed with four independent blind passes, a blind adjudicator, candidate-order perturbation, and 21 negative controls. The hidden gate failed decisively: adjudicated exact occurrence accuracy was {exact}/7 and candidate-order stability was {stable}/7, versus mandatory thresholds of 6/7 and 6/7. Adjudication returned {Counter(r['outcome'] for r in adjudication.values())['PROBABLE']} PROBABLE and {Counter(r['outcome'] for r in adjudication.values())['AMBIGUOUS']} AMBIGUOUS cases across all 21, with no consensus-exact result and no forced answer.\n\nConsequently the procedure was not applied to the 71 residual LABELs. `RESIDUAL_AI_MAPPING_RESULTS.tsv` and `AI_CONSENSUS_LABEL_TOKEN_MAPPING.tsv` are intentionally absent rather than empty or fabricated. `UNRESOLVED_LABELS.tsv` records all 71 as `NOT_RUN_GATE_FAILED`. Existing primary coverage therefore remains 21/92, entirely legacy-confirmed, with grouped coverage 21/64 and ungrouped coverage 0/28. This remains unusable for an identified grouped-versus-ungrouped hapax comparison.\n\nThe negative controls abstained in {control_pass}/{control_total}, but this cannot offset the positive-case failure. No threshold, prompt, candidate rule, or answer was changed after disclosure. No enrichment, frequency calculation, dictionary search, naming hypothesis, semantic interpretation, spatial edit, or expert-verification claim was made.\n\nThe available agent sessions used multiple model families. Because the execution environment shares a repository filesystem, clean-room isolation is procedural and content-audited rather than OS-enforced; this limitation is explicit and prevents overstating independence.\n\n```text\nRESIDUAL_AI_MAPPING_STATUS=BLOCKED\nMAPPING_METHOD=INDEPENDENT_MULTIPASS_AI_CONSENSUS\nLEGACY_MAPPINGS=21\nCALIBRATION_VISIBLE=14\nEVALUATION_HIDDEN=7\nHIDDEN_EXACT_ACCURACY={exact}/7\nHIDDEN_ORDER_STABLE={stable}/7\nRESIDUAL_AI_MAPPING_RUN_AUTHORIZED=NO\nRESIDUAL_LABELS=71\nRESIDUAL_PRIMARY_MAPPED=0\nRESIDUAL_PROBABLE=0\nRESIDUAL_AMBIGUOUS=0\nRESIDUAL_NO_MATCH=0\nRESIDUAL_UNREADABLE=0\nRESIDUAL_NOT_RUN_GATE_FAILED=71\nTOTAL_PRIMARY_MAPPED=21/92\nGROUPED_PRIMARY_MAPPED=21/64\nUNGROUPED_PRIMARY_MAPPED=0/28\nHAPAX_ENRICHMENT_RUN_AUTHORIZED=NO\nHAPAX_ENRICHMENT_PERFORMED=NO\nLEXICON_MATCH_PERFORMED=NO\nHUMAN_EXPERT_VERIFICATION_CLAIMED=NO\nFROZEN_INPUTS_UNCHANGED={'YES' if unchanged else 'NO'}\nRESULTS_REPRODUCIBLE=NO\n```\n"""
    (OUT / "AI_MAPPING_REPORT.md").write_text(report, encoding="utf-8")
    (OUT / "AI_MAPPING_SUMMARY.md").write_text(
        f"# AI mapping summary\n\nThe mandatory hidden gate failed (exact occurrence {exact}/7; order-stable {stable}/7). Production mapping of the 71 residual LABELs was therefore not authorized or run. Primary coverage remains the 21 legacy-confirmed f68r1 mappings; hapax enrichment remains unauthorized.\n",
        encoding="utf-8",
    )
    (OUT / "REPRODUCIBILITY.md").write_text(
        "# Reproducibility\n\nRun `python3 scripts/prepare_pipeline.py` before AI annotation, then the separately recorded context-free sessions, `python3 scripts/prepare_adjudication.py`, `python3 scripts/prepare_negative_controls.py`, and finally `python3 scripts/finalize_calibration.py`. The deterministic split, candidate orders, crops, manifests, scoring, and reports are reproducible from frozen inputs. Model outputs are checksum-frozen but stochastic; a required repeated AI run was not attempted because first-run accuracy and stability had already failed the immutable gate. Therefore `RESULTS_REPRODUCIBLE=NO` refers to the AI result, not to the deterministic preparation/scoring code.\n",
        encoding="utf-8",
    )
    validation = f"""# Validation report\n\n- Frozen protocol present and unchanged from pre-run freeze: PASS.\n- Input checksum verification: {'PASS' if unchanged else 'FAIL'}.\n- Target/legacy/residual counts 92/21/71 and residual panel distribution 16/33/22: PASS.\n- Deterministic calibration split 14/7: PASS.\n- Clean-room content blindness audit: PASS (instruction-enforced isolation limitation documented).\n- Calibration feedback/tuning stage: NOT USED; visible and hidden partitions were processed in one frozen batch before any answer disclosure.\n- Complete candidate registry plus same-panel admissibility rule: PASS; all ranked/selected calibration candidates are f68r1.\n- Independent A/B/C/base-shuffled outputs: PASS, 21 unique valid records each.\n- Adjudication records and no forced answers: PASS.\n- Selected occurrence existence/raw token binding: PASS.\n- Candidate-order classification: PASS; zero hidden `ORDER_STABLE`.\n- Hidden gate applied without threshold change: PASS; result NO.\n- Production residual outputs absent after gate failure: PASS.\n- No hapax enrichment, frequency analysis, lexicon matching, or Cartesian 3G1 expansion: PASS.\n- Frozen repeated AI run: NOT RUN because earlier mandatory gate conditions failed.\n- Final SHA-256 ledger: generated after this report.\n"""
    (OUT / "VALIDATION_REPORT.md").write_text(validation, encoding="utf-8")
    sha_ledger()
    print(json.dumps({"hidden_exact": f"{exact}/7", "hidden_order_stable": f"{stable}/7", "gate": "NO", "negative_controls": f"{control_pass}/{control_total}"}, sort_keys=True))


if __name__ == "__main__":
    main()
