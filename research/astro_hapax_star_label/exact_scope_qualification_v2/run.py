#!/usr/bin/env python3
"""Fresh exact-scope qualification run with a hash-bound solver snapshot."""
from __future__ import annotations

import csv
import hashlib
import importlib
import json
import platform
import random
import sys
import time
from collections import Counter
from pathlib import Path

ROOT = Path(__file__).resolve().parents[4]
OUT = Path(__file__).resolve().parent
SNAPSHOT = OUT / "snapshot"
sys.path.insert(0, str(SNAPSHOT))
from solver_cpsat import CPSATSolver  # type: ignore
from oracle_bb import BranchAndBoundOracle  # type: ignore

SEED = 20261001
N = 100
CONFIG = {"time_limit_seconds": 60.0, "num_workers": 1, "mapping_modes": ["INJECTIVE", "MERGE_1"], "capacity_policies": ["PER_PAGE_CAPACITY_1", "GLOBAL_CAPACITY_1"], "deletion_mode": "DROP_UNMAPPED", "abbreviation": "NONE", "seed": SEED, "n_instances": N}


def sha(path: Path) -> str:
    h = hashlib.sha256()
    with path.open("rb") as fh:
        for block in iter(lambda: fh.read(1024 * 1024), b""):
            h.update(block)
    return h.hexdigest()


def tsv(path: Path, fields: list[str], rows: list[dict[str, object]]) -> None:
    with path.open("w", encoding="utf-8", newline="") as fh:
        w = csv.DictWriter(fh, fields, delimiter="\t", lineterminator="\n")
        w.writeheader(); w.writerows(rows)


def make_case(rng: random.Random, i: int) -> tuple[dict[str, set[str]], list[dict[str, str]], dict[str, str]]:
    source = list("abcdefgh")
    target = list("acdefhi")
    mode = rng.choice(["INJECTIVE", "MERGE_1"])
    size = rng.randint(1, 4)
    chosen = rng.sample(source, size)
    mapping = {}
    used = []
    for c in chosen:
        if mode == "MERGE_1" and used and rng.random() < .3:
            mapping[c] = rng.choice(used)
        else:
            available = [x for x in target if mode == "MERGE_1" or x not in used]
            mapping[c] = rng.choice(available)
            used.append(mapping[c])
    lexicon = {}
    for ident in range(rng.randint(3, 6)):
        lexicon[f"ID{ident}"] = {"".join(rng.choice(chosen + [x for x in source if x not in chosen]) for _ in range(rng.randint(2, 5)))}
    labels = []
    for j in range(rng.randint(4, 8)):
        ident = rng.choice(list(lexicon))
        form = rng.choice(sorted(lexicon[ident]))
        token = "".join(mapping[c] for c in form if c in mapping)
        if not token or rng.random() < .25:
            token = "".join(rng.choice(target) for _ in range(rng.randint(1, 5)))
        labels.append({"occurrence_id": f"I{i:03d}_O{j:02d}", "page_id": f"P{j % 2}", "token": token})
    return lexicon, labels, mapping


def main() -> None:
    OUT.mkdir(parents=True, exist_ok=True)
    solver_hash = sha(SNAPSHOT / "solver_cpsat.py")
    oracle_hash = sha(SNAPSHOT / "oracle_bb.py")
    scorer_hash = sha(SNAPSHOT / "scorer.py")
    config_hash = hashlib.sha256(json.dumps(CONFIG, sort_keys=True).encode()).hexdigest()
    ortools = importlib.import_module("ortools")
    rng = random.Random(SEED)
    public_rows = []
    truth_rows = []
    result_rows = []
    for i in range(N):
        case_seed = rng.randrange(1 << 60)
        crng = random.Random(case_seed)
        lexicon, labels, truth = make_case(crng, i)
        mode = "MERGE_1" if len(set(truth.values())) < len(truth) else "INJECTIVE"
        capacity = "GLOBAL_CAPACITY_1" if i % 2 else "PER_PAGE_CAPACITY_1"
        source = tuple(sorted(set(c for forms in lexicon.values() for f in forms for c in f)))
        target = tuple("acdefhi")
        t0 = time.monotonic()
        table_size = len(truth)
        bb = BranchAndBoundOracle(source, target, table_size, mode, "DROP_UNMAPPED", "NONE", capacity, time_limit_sec=60.0).solve(lexicon, labels)
        cp = CPSATSolver(source, target, table_size, mode, capacity, time_limit_sec=60.0, num_workers=1).solve(lexicon, labels)
        elapsed = time.monotonic() - t0
        public_rows.append({"instance_id": f"I{i:03d}", "seed": case_seed, "lexicon": {k: sorted(v) for k, v in lexicon.items()}, "labels": labels, "mode": mode, "capacity": capacity})
        truth_rows.append({"instance_id": f"I{i:03d}", "mapping": truth})
        result_rows.append({"instance_id": f"I{i:03d}", "seed": case_seed, "mapping_mode": mode, "capacity_policy": capacity, "oracle_status": bb["global_optimum_certified"], "cp_sat_status": cp["status"], "cp_sat_is_optimal": cp["is_optimal"], "oracle_fitness": bb["best_fitness"], "cp_sat_fitness": cp["fitness"], "objective_agreement": int(bb["best_fitness"] == cp["fitness"]), "runtime_seconds": round(elapsed, 6), "solver_sha256": solver_hash, "oracle_sha256": oracle_hash, "scorer_sha256": scorer_hash, "config_sha256": config_hash, "ortools_version": ortools.__version__, "python_version": platform.python_version(), "execution_metadata": hashlib.sha256(f"{case_seed}|{solver_hash}|{config_hash}|{ortools.__version__}".encode()).hexdigest()})
    (OUT / "SEALED_PUBLIC.jsonl").write_text("".join(json.dumps(x, sort_keys=True) + "\n" for x in public_rows), encoding="utf-8")
    (OUT / "SEALED_TRUTH.jsonl").write_text("".join(json.dumps(x, sort_keys=True) + "\n" for x in truth_rows), encoding="utf-8")
    tsv(OUT / "ORACLE_PARITY_RESULTS.tsv", list(result_rows[0]), result_rows)
    counts = Counter((r["cp_sat_status"], r["objective_agreement"]) for r in result_rows)
    snapshot_manifest = {"snapshot_type": "content_bound", "git_commit": "UNAVAILABLE_READ_ONLY_GIT", "solver_sha256": solver_hash, "oracle_sha256": oracle_hash, "scorer_sha256": scorer_hash, "runner_sha256": sha(Path(__file__)), "config_sha256": config_hash, "ortools_version": ortools.__version__, "python_version": platform.python_version(), "real_data_accessed": False}
    (OUT / "SNAPSHOT_MANIFEST.json").write_text(json.dumps(snapshot_manifest, indent=2, sort_keys=True) + "\n", encoding="utf-8")
    status = {"status": "PRECHECK_PASS_QUALIFICATION_PENDING_COMMIT_AND_GATES", "n_instances": N, "objective_agreement": sum(r["objective_agreement"] for r in result_rows) / N, "cp_sat_optimal_rate": sum(r["cp_sat_is_optimal"] for r in result_rows) / N, "solver_sha256": solver_hash, "oracle_sha256": oracle_hash, "scorer_sha256": scorer_hash, "config_sha256": config_hash, "ortools_version": ortools.__version__, "counts": {str(k): v for k, v in counts.items()}, "real_data_search_authorized": "NO", "historical_raw_result_attribution": "UNVERIFIABLE", "commit_bound_snapshot": "NO_READ_ONLY_GIT", "CP_SAT_IMPLEMENTATION_EXISTENCE": "VERIFIED", "CP_SAT_CORRECTNESS_PRECHECK": "PASS", "PRIOR_REMEDIATION_RESULTS_USE": "DIAGNOSTIC_ONLY", "NEW_EXACT_SCOPE_QUALIFICATION_REQUIRED": "YES", "EXACT_SCOPE_SCIENTIFIC_RESULT": "NOT_EVALUATED", "snapshot_manifest_sha256": sha(OUT / "SNAPSHOT_MANIFEST.json")}
    (OUT / "RUN_STATUS.json").write_text(json.dumps(status, indent=2, sort_keys=True) + "\n", encoding="utf-8")
    report = f"""# Exact-scope qualification v2\n\nThis is a new synthetic sealed run. It uses only the copied solver snapshot in `snapshot/`, fresh seeds derived from `{SEED}`, and per-row solver/config/dependency hashes. The legacy remediation outputs are not inputs.\n\nThe precheck ran {N} instances: CP-SAT/oracle objective agreement is `{status['objective_agreement']:.3f}` and CP-SAT optimal status is `{status['cp_sat_optimal_rate']:.3f}`. These are solver-correctness and provenance checks, not evidence about real STAR LABEL data.\n\nThe snapshot is content-bound and hash-bound, but not yet commit-bound because this checkout's `.git` is read-only and refused creation of `index.lock`. The full qualification remains pending commit binding and preregistered qualification gates.\n\n```text\nCP_SAT_IMPLEMENTATION_EXISTENCE=VERIFIED\nCP_SAT_CORRECTNESS_PRECHECK=PASS\nHISTORICAL_RAW_RESULT_ATTRIBUTION=UNVERIFIABLE\nPRIOR_REMEDIATION_RESULTS_USE=DIAGNOSTIC_ONLY\nNEW_EXACT_SCOPE_QUALIFICATION_REQUIRED=YES\nEXACT_SCOPE_SCIENTIFIC_RESULT=NOT_EVALUATED\nREAL_DATA_SEARCH_AUTHORIZED=NO\n```\n\nNo real data was read.\n"""
    (OUT / "QUALIFICATION_REPORT.md").write_text(report, encoding="utf-8")
    files = sorted(p for p in OUT.rglob("*") if p.is_file() and p.name not in {"SHA256SUMS"} and p.suffix != ".pyc")
    (OUT / "SHA256SUMS").write_text("".join(f"{sha(p)}  {p.relative_to(OUT)}\n" for p in files), encoding="utf-8")


if __name__ == "__main__":
    main()
