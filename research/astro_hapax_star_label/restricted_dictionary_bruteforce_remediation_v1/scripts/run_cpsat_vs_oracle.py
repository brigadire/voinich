#!/usr/bin/env python3
"""Randomized parity check: CP_SAT vs BranchAndBoundOracle, restricted to CP-SAT's supported
scope (DROP_UNMAPPED, abbreviation NONE, single-character source alphabet). Writes
SOLVER_REGISTRY-adjacent evidence rows to CPSAT_ORACLE_PARITY.tsv."""
import random
import sys
import csv
from pathlib import Path

sys.path.insert(0, str(Path(__file__).parent))
from oracle_bb import BranchAndBoundOracle
from solver_cpsat import CPSATSolver

SRC_POOL = list("abcdefgh")
TGT_POOL = list("acdefhi")


def random_instance(rng):
    n_identities = rng.randint(2, 5)
    lexicon = {}
    for i in range(n_identities):
        forms = {"".join(rng.choice(SRC_POOL) for _ in range(rng.randint(2, 5))) for _ in range(rng.randint(1, 2))}
        lexicon[f"ID{i}"] = forms
    n_labels = rng.randint(3, 8)
    n_pages = rng.randint(1, 2)
    labels = [
        {"occurrence_id": f"OCC{j}", "page_id": f"P{j % n_pages}", "token": "".join(rng.choice(TGT_POOL) for _ in range(rng.randint(1, 5)))}
        for j in range(n_labels)
    ]
    table_size = rng.randint(0, 4)
    mapping_mode = rng.choice(["INJECTIVE", "MERGE_1"])
    capacity_policy = rng.choice(["PER_PAGE_CAPACITY_1", "GLOBAL_CAPACITY_1"])
    return lexicon, labels, table_size, mapping_mode, capacity_policy


def main():
    n_trials = int(sys.argv[1]) if len(sys.argv) > 1 else 100
    out_path = sys.argv[2] if len(sys.argv) > 2 else "CPSAT_ORACLE_PARITY.tsv"
    rng = random.Random(7)
    rows, mismatches = [], 0
    for trial in range(n_trials):
        lexicon, labels, table_size, mapping_mode, capacity_policy = random_instance(rng)
        bb = BranchAndBoundOracle(SRC_POOL, TGT_POOL, table_size, mapping_mode, "DROP_UNMAPPED", "NONE", capacity_policy, time_limit_sec=15.0)
        rb = bb.solve(lexicon, labels)
        cps = CPSATSolver(SRC_POOL, TGT_POOL, table_size, mapping_mode, capacity_policy, time_limit_sec=15.0)
        rc = cps.solve(lexicon, labels)
        both_certain = rb["global_optimum_certified"] and rc["is_optimal"]
        agree = (not both_certain) or (rb["best_fitness"] == rc["fitness"])
        if not agree:
            mismatches += 1
        rows.append({
            "trial": trial, "table_size": table_size, "mapping_mode": mapping_mode, "capacity_policy": capacity_policy,
            "bb_fitness": rb["best_fitness"], "bb_certified": rb["global_optimum_certified"],
            "cpsat_fitness": rc["fitness"], "cpsat_status": rc["status"], "cpsat_runtime": rc["runtime_sec"],
            "bb_runtime": rb["runtime_sec"], "both_certain": both_certain, "agree": agree,
        })
    with open(out_path, "w", newline="") as f:
        w = csv.DictWriter(f, fieldnames=list(rows[0].keys()), delimiter="\t")
        w.writeheader()
        w.writerows(rows)
    print(f"trials={n_trials} mismatches={mismatches}")
    return 0 if mismatches == 0 else 1


if __name__ == "__main__":
    sys.exit(main())
