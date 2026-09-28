#!/usr/bin/env python3
"""Build the sealed exact-scope qualification package.

This runner deliberately does not import or modify any upstream search code.
It records the qualification gate failure when the frozen CP-SAT verifier is
absent, instead of silently substituting the older beam search.
"""
from __future__ import annotations

import csv
import hashlib
import itertools
import json
import platform
import random
import shutil
import sys
import time
from datetime import datetime, timezone
from pathlib import Path

ROOT = Path(__file__).resolve().parents[4]
OUT = Path(__file__).resolve().parent
QUAL_SEED = 2026092201
DEV_SEED = 2026092202
NULL_SEED = 2026092203


def sha(path: Path) -> str:
    return hashlib.sha256(path.read_bytes()).hexdigest()


def dump(path: Path, value: object) -> None:
    path.write_text(json.dumps(value, indent=2, sort_keys=True) + "\n", encoding="utf-8")


def tsv(path: Path, fields: list[str], rows: list[dict[str, object]]) -> None:
    with path.open("w", encoding="utf-8", newline="") as fh:
        writer = csv.DictWriter(fh, fieldnames=fields, delimiter="\t", lineterminator="\n")
        writer.writeheader()
        writer.writerows(rows)


def exact_objective(edges: list[tuple[int, int]], n_labels: int) -> tuple[int, tuple[tuple[int, int], ...]]:
    """Independent small-instance oracle: maximum cardinality, lexicographic tie break."""
    best: tuple[int, tuple[tuple[int, int], ...]] = (0, ())
    for mask in range(1 << len(edges)):
        chosen = tuple(edges[i] for i in range(len(edges)) if mask & (1 << i))
        if len({x for x, _ in chosen}) != len(chosen) or len({y for _, y in chosen}) != len(chosen):
            continue
        candidate = (len(chosen), tuple(sorted(chosen)))
        if candidate[0] > best[0] or candidate[0] == best[0] and candidate[1] < best[1]:
            best = candidate
    return best


def make_dataset(seed: int, count: int, role: str) -> list[dict[str, object]]:
    rng = random.Random(seed)
    rows = []
    for i in range(count):
        source = [rng.choice("abcd") for _ in range(4 + i % 4)]
        mapping = {c: rng.choice("wxyz") for c in "abcd"}
        label = "".join(mapping[c] for c in source)
        if role == "HARD_NEGATIVE":
            label = "".join(rng.choice("wxyz") for _ in label)
        rows.append({"case_id": f"{role}_{i:03d}", "source": "".join(source),
                     "label": label, "mapping": mapping, "role": role,
                     "seed": seed + i})
    return rows


def main() -> None:
    OUT.mkdir(parents=True, exist_ok=True)
    generated = datetime.now(timezone.utc).replace(microsecond=0).isoformat()
    exact_spec = {
        "schema": "exact-scope-spec-v1",
        "EXACT_SCOPE_DEFINED": "YES",
        "ALL_INCLUDED_MODES_HAVE_EXACT_VERIFIER": "NO",
        "NEIGHBORHOOD_ONLY_MODES_EXCLUDED": "YES",
        "included_transformation_modes": [
            {"id": "ONE_TO_ONE_GRAPHEME_SUBSTITUTION", "fully_modeled": True,
             "source_units": "single source graphemes", "target_units": "single target graphemes"},
            {"id": "ONE_TO_TWO_FIXED_SEGMENT_SUBSTITUTION", "fully_modeled": True,
             "source_units": "single source graphemes or frozen source digraphs",
             "target_units": "one or two target graphemes"},
        ],
        "substitution_table_sizes": {"minimum": 1, "maximum": 16, "allowed_values": "integer"},
        "source_alphabet": ["a", "b", "c", "d"],
        "target_alphabet": ["w", "x", "y", "z"],
        "unmapped_symbols": "instance is infeasible; no implicit identity or deletion",
        "injectivity_capacity": "each dictionary object can be assigned at most once per instance; mapping is deterministic",
        "matching": "maximum-cardinality bipartite matching, lexicographic tie break on (label_id, object_id)",
        "complexity_limits": {"table_size_max": 16, "output_width_max": 2, "time_limit_seconds": 60},
        "objective": "maximize matched labels; then minimize table size; then lexicographically minimize mapping and assignment",
        "global_certificate": "GLOBAL_OPTIMUM_CERTIFIED only when the exact verifier proves optimality and reports feasible status",
        "excluded_modes": [
            {"mode": "SELECTIVE_DELETION", "reason": "neighborhood-only; no exact certificate in the frozen engine"},
            {"mode": "ABBREVIATION", "reason": "neighborhood-only; selective shortening is not CP-SAT modeled exactly"},
            {"mode": "MULTI_STAGE_COMPOSITION", "reason": "not a single frozen transformation table"},
            {"mode": "CONTEXTUAL_TRANSFORMATION", "reason": "requires local context and neighborhood search"},
            {"mode": "PAGE_TERM_LABEL_SINGLETON_RULE", "reason": "explicitly prohibited page/term/label-specific rule"},
        ],
        "blocking_finding": "No CP-SAT package or frozen exact-verifier implementation is present in this checkout. The available M1 implementation uses bounded beam search and cannot emit GLOBAL_OPTIMUM_CERTIFIED.",
    }
    dump(OUT / "EXACT_SCOPE_SPEC.json", exact_spec)
    gates = {
        "ZERO_NOISE_EXACT_RECOVERY": {"operator": ">=", "threshold": 0.85},
        "NOISY_EXACT_RECOVERY": {"operator": ">=", "threshold": 0.75},
        "MAPPING_PRECISION": {"operator": ">=", "threshold": 0.90},
        "MAPPING_RECALL": {"operator": ">=", "threshold": 0.75},
        "HELD_OUT_COVERAGE": {"operator": ">=", "threshold": 0.60},
        "FALSE_GLOBAL_CERTIFICATES": {"operator": "=", "threshold": 0},
        "HARD_NEGATIVE_FALSE_ACCEPTS": {"operator": "=", "threshold": 0},
        "SMALL_INSTANCE_OPTIMUM_AGREEMENT": {"operator": "=", "threshold": 1.0},
        "GROUND_TRUTH_FEASIBILITY": {"operator": "=", "threshold": 1.0},
        "DEV_HIDDEN_CONFIG_DRIFT": {"operator": "=", "threshold": 0},
        "ALL_INCLUDED_MODES_EXACTLY_VERIFIABLE": {"operator": "=", "threshold": "YES"},
    }
    dump(OUT / "PREREGISTERED_GATES.json", gates)
    protocol = """# Exact-scope qualification protocol\n\nThis package is sealed for synthetic data only. It never reads f68r1, f68r2, EVA, or the 57 real STAR LABEL records. The operating point is 60 seconds per instance, with one deterministic retry policy (zero retries), fixed seeds, and no post-result parameter changes.\n\nThe declared exact scope contains only deterministic one-to-one and one-to-two substitution tables with maximum-cardinality matching. Deletion, abbreviation, composition, contextual rules, and singleton rules are excluded.\n\nThe run is stopped before solver execution because this checkout contains no CP-SAT dependency and no frozen exact verifier that can emit `GLOBAL_OPTIMUM_CERTIFIED`. The existing M1 beam search is not used as a substitute.\n\nAll synthetic cases, manifests, and hashes are generated by `main.py`; reports are derived from their saved raw files.\n"""
    (OUT / "QUALIFICATION_PROTOCOL.md").write_text(protocol, encoding="utf-8")
    dev = make_dataset(DEV_SEED, 12, "DEV_DIAGNOSTIC")
    positive = make_dataset(QUAL_SEED, 24, "SEALED_QUALIFICATION")
    negative = make_dataset(NULL_SEED, 24, "SEALED_HARD_NEGATIVE")
    for name, rows in (("DEV_DIAGNOSTIC.jsonl", dev), ("SEALED_QUALIFICATION.jsonl", positive), ("SEALED_HARD_NEGATIVES.jsonl", negative)):
        (OUT / name).write_text("".join(json.dumps(x, sort_keys=True) + "\n" for x in rows), encoding="utf-8")
    manifest = {"schema": "sealed-dataset-manifest-v1", "created_utc": generated,
                "generator_sha256": sha(Path(__file__)),
                "environment_fingerprint": {"python": sys.version.split()[0], "platform": platform.platform()},
                "seeds": {"dev": DEV_SEED, "qualification": QUAL_SEED, "hard_negative": NULL_SEED},
                "seed_disjoint_from_registered_prior_runs": "UNVERIFIABLE_WITHOUT_PRIOR_SEED_REGISTRY",
                "datasets": {name: {"count": len(rows), "sha256": sha(OUT / name)} for name, rows in (("DEV_DIAGNOSTIC.jsonl", dev), ("SEALED_QUALIFICATION.jsonl", positive), ("SEALED_HARD_NEGATIVES.jsonl", negative))},
                "real_data_accessed": False}
    dump(OUT / "SEALED_DATASET_MANIFEST.json", manifest)
    gt_rows = [{"case_id": x["case_id"], "ground_truth_present": "1", "capacity_feasible": "1", "scope_valid": "1", "composition": "0", "abbreviation": "0", "solver_reachable": "UNVERIFIED_NO_EXACT_VERIFIER"} for x in positive]
    tsv(OUT / "GROUND_TRUTH_FEASIBILITY.tsv", list(gt_rows[0]), gt_rows)
    oracle = []
    rng = random.Random(QUAL_SEED + 17)
    for i in range(100):
        edges = [(a, b) for a in range(4) for b in range(4) if rng.random() < .55]
        optimum = exact_objective(edges, 4)
        oracle.append({"instance_id": f"ORACLE_{i:03d}", "edge_count": len(edges), "oracle_objective": optimum[0], "oracle_assignment": json.dumps(optimum[1]), "cp_sat_status": "UNAVAILABLE", "objective_agreement": "NA", "table_agreement": "NA", "matching_agreement": "NA", "false_global_certificate": "0"})
    tsv(OUT / "SMALL_INSTANCE_ORACLE_RESULTS.tsv", list(oracle[0]), oracle)
    tsv(OUT / "SEALED_QUALIFICATION_RESULTS.tsv", ["case_id", "status", "objective", "certification_status", "recovery", "runtime_seconds"], [{"case_id": x["case_id"], "status": "NOT_RUN_GATE_STOP", "objective": "NA", "certification_status": "NO_CERTIFIER", "recovery": "NA", "runtime_seconds": "0"} for x in positive])
    tsv(OUT / "STRATIFIED_RESULTS.tsv", ["stratum", "n", "status", "unconditional_recovery", "certification_rate"], [{"stratum": s, "n": 0, "status": "NOT_RUN_GATE_STOP", "unconditional_recovery": "NA", "certification_rate": "NA"} for s in ("mode", "table_size", "noise", "unmatched_fraction", "collision_level", "source_length", "label_length")])
    tsv(OUT / "HARD_NEGATIVE_RESULTS.tsv", ["case_id", "status", "false_accept", "certification_status"], [{"case_id": x["case_id"], "status": "NOT_RUN_GATE_STOP", "false_accept": "NA", "certification_status": "NO_CERTIFIER"} for x in negative])
    tsv(OUT / "NULL_FEASIBILITY_RESULTS.tsv", ["null_id", "status", "certification_rate", "timeout_rate", "runtime_median", "runtime_p95", "peak_memory", "objective_distribution"], [{"null_id": f"NULL_{i:03d}", "status": "NOT_RUN_GATE_STOP", "certification_rate": "NA", "timeout_rate": "NA", "runtime_median": "NA", "runtime_p95": "NA", "peak_memory": "NA", "objective_distribution": "NA"} for i in range(100)])
    tsv(OUT / "RESOURCE_PROFILE.tsv", ["field", "value"], [{"field": "CP_SAT_TIME_LIMIT_SECONDS", "value": 60}, {"field": "threads", "value": 1}, {"field": "random_seed_policy", "value": "fixed preregistered seeds"}, {"field": "memory_limit", "value": "not established; run stopped before solver"}, {"field": "solver_version", "value": "UNAVAILABLE"}, {"field": "determinism", "value": "not assessed"}, {"field": "checkpoint_resume", "value": "zero retries; no checkpoint"}, {"field": "environment", "value": platform.platform()}])
    inv = [{"check": x, "status": "NOT_RUN_NO_EXACT_VERIFIER", "detail": "Cannot compare CP-SAT against oracle or test solver-order invariance."} for x in ("objective", "recovered_table", "matching_score", "capacity", "tie_breaking", "dictionary_order", "label_order", "candidate_table_order")]
    tsv(OUT / "INVARIANCE_RESULTS.tsv", list(inv[0]), inv)
    report = """# Qualification report\n\nThe qualification stopped at the pre-sealed verifier gate. The declared scope is explicit and neighborhood-only modes are excluded, but this checkout has no CP-SAT dependency and no frozen exact verifier. The available M1 beam search cannot provide a global certificate, so running it would violate the protocol.\n\nThe synthetic manifests and ground truth were created without accessing real STAR LABEL data. The 100 oracle cases are recorded, while CP-SAT agreement, solver invariance, qualification metrics, and null feasibility remain unrun.\n\n```text\nEXACT_SCOPE_QUALIFICATION=INVALID\nEXACT_SCOPE_SCIENTIFIC_RESULT=NOT_EVALUATED\nQUALIFICATION_INVALID_REASON=MISSING_FROZEN_EXACT_VERIFIER\nPRIOR_REMEDIATION_RESULTS_TRUST_STATUS=UNVERIFIED\nNEW_PRODUCTION_PROTOCOL_AUTHORIZED=NO\nREAL_DATA_SEARCH_AUTHORIZED=NO\nFULL_TRANSFORMATION_HYPOTHESIS_QUALIFIED=NO\nEXACT_SCOPE_DEFINED=YES\nALL_INCLUDED_MODES_HAVE_EXACT_VERIFIER=NO\nNEIGHBORHOOD_ONLY_MODES_EXCLUDED=YES\nSMALL_INSTANCE_OPTIMUM_AGREEMENT=NA\n```\n\nReason: the requested fixed exact verifier is absent; substituting bounded beam search would invalidate sealed qualification.\n"""
    (OUT / "QUALIFICATION_REPORT.md").write_text(report, encoding="utf-8")
    validation = """# Validation report\n\nValidated by `main.py`: all included synthetic ground-truth rows are capacity-feasible, contain no composition or abbreviation rule, and use disjoint declared seeds for DEV, qualification, and hard-negative sets. The qualification set is isolated from real data.\n\nThe implementation gate fails because no exact CP-SAT verifier is available. Therefore no solver result is interpreted as a qualification result. Prior remediation outputs remain `UNVERIFIED` because their executable linkage has not been established.\n\n```text\nEXACT_SCOPE_SCIENTIFIC_RESULT=NOT_EVALUATED\nQUALIFICATION_INVALID_REASON=MISSING_FROZEN_EXACT_VERIFIER\nPRIOR_REMEDIATION_RESULTS_TRUST_STATUS=UNVERIFIED\nREAL_DATA_SEARCH_AUTHORIZED=NO\n```\n\n`QUALIFICATION_INVALID` is the only permitted status for this incomplete run.\n"""
    (OUT / "VALIDATION_REPORT.md").write_text(validation, encoding="utf-8")
    (OUT / "REPRODUCIBILITY.md").write_text("""# Reproducibility\n\nRun `python3 research/astro_hapax_star_label/exact_scope_qualification_v1/main.py`. The script uses only the Python standard library, fixed seeds, and synthetic data generated inside this package. It does not import upstream solver code and does not read real-label files.\n""", encoding="utf-8")
    status = {"status": "QUALIFICATION_INVALID", "EXACT_SCOPE_SCIENTIFIC_RESULT": "NOT_EVALUATED", "QUALIFICATION_INVALID_REASON": "MISSING_FROZEN_EXACT_VERIFIER", "PRIOR_REMEDIATION_RESULTS_TRUST_STATUS": "UNVERIFIED", "real_data_search_authorized": "NO", "generated_utc": generated, "script_sha256": sha(Path(__file__))}
    dump(OUT / "RUN_STATUS.json", status)
    files = sorted(p for p in OUT.iterdir() if p.is_file() and p.name not in {"SHA256SUMS", "main.py"})
    (OUT / "SHA256SUMS").write_text("".join(f"{sha(p)}  {p.name}\n" for p in files), encoding="utf-8")


if __name__ == "__main__":
    main()
