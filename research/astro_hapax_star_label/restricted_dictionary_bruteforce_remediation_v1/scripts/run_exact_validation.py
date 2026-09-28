#!/usr/bin/env python3
"""
Validates oracle_bb.BranchAndBoundOracle against literal brute-force enumeration of every
valid table on small randomized instances (>=500 trials, Section 9 of clean_room.md).
Writes EXACT_SOLVER_VALIDATION.tsv.

Brute force enumerates every table of exactly `table_size` over a small source-character
subset and every target-alphabet assignment (permutations for INJECTIVE, permutations plus
one bounded duplicate for MERGE_1), scores each with the same Scorer, and takes the true max.
This is only tractable because trial instances are kept deliberately small
(table_size <= 3, target alphabet <= 5) -- oracle_bb itself is unrestricted in size elsewhere.
"""
import random
import sys
import csv
import itertools
from pathlib import Path

sys.path.insert(0, str(Path(__file__).parent))
from scorer import Scorer
from oracle_bb import BranchAndBoundOracle

SRC_POOL = list("abcdefgh")
TGT_POOL = list("acdefh")


def brute_force_best(lexicon, labels, source_subset, target_subset, table_size, mapping_mode, deletion_mode, abbreviation, capacity_policy):
    scorer = Scorer(mapping_mode, deletion_mode, abbreviation, capacity_policy)
    # table_size is an exact constraint (see oracle_bb.py's fix note) -- table_size==0 is still
    # covered correctly because itertools.combinations(..., 0) yields exactly one empty combo.
    best_fitness = -float("inf")
    for src_combo in itertools.combinations(source_subset, table_size):
        if mapping_mode == "INJECTIVE":
            for tgt_perm in itertools.permutations(target_subset, table_size):
                table = dict(zip(src_combo, tgt_perm))
                fit = scorer.evaluate(table, lexicon, labels)["fitness"]
                if fit > best_fitness:
                    best_fitness = fit
        else:  # MERGE_1: every assignment of target_subset values (with repetition) that has
               # at most one target used exactly twice and none used 3+ times
            for tgt_combo in itertools.product(target_subset, repeat=table_size):
                counts = {}
                for t in tgt_combo:
                    counts[t] = counts.get(t, 0) + 1
                if max(counts.values(), default=0) > 2:
                    continue
                if sum(1 for c in counts.values() if c == 2) > 1:
                    continue
                table = dict(zip(src_combo, tgt_combo))
                fit = scorer.evaluate(table, lexicon, labels)["fitness"]
                if fit > best_fitness:
                    best_fitness = fit
    return best_fitness


def random_instance(rng):
    n_identities = rng.randint(2, 4)
    lexicon = {}
    for i in range(n_identities):
        forms = {"".join(rng.choice(SRC_POOL[:5]) for _ in range(rng.randint(2, 4))) for _ in range(rng.randint(1, 2))}
        lexicon[f"ID{i}"] = forms
    n_labels = rng.randint(2, 6)
    n_pages = rng.randint(1, 2)
    labels = [
        {"occurrence_id": f"OCC{j}", "page_id": f"P{j % n_pages}", "token": "".join(rng.choice(TGT_POOL) for _ in range(rng.randint(1, 4)))}
        for j in range(n_labels)
    ]
    table_size = rng.randint(0, 3)
    source_subset = rng.sample(SRC_POOL[:5], min(5, len(SRC_POOL[:5])))
    target_subset = TGT_POOL[:]
    mapping_mode = rng.choice(["INJECTIVE", "MERGE_1"])
    deletion_mode = rng.choice(["DROP_UNMAPPED", "SELECTIVE_VOWEL_DROP", "NONE"])
    abbreviation = rng.choice(["NONE", "SUSPENSION_1", "PREFIX_4"])
    capacity_policy = rng.choice(["PER_PAGE_CAPACITY_1", "GLOBAL_CAPACITY_1"])
    return lexicon, labels, source_subset, target_subset, table_size, mapping_mode, deletion_mode, abbreviation, capacity_policy


def main():
    n_trials = int(sys.argv[1]) if len(sys.argv) > 1 else 500
    out_path = sys.argv[2] if len(sys.argv) > 2 else "EXACT_SOLVER_VALIDATION.tsv"
    rng = random.Random(20260917)

    rows = []
    mismatches = 0
    for trial in range(n_trials):
        lexicon, labels, source_subset, target_subset, table_size, mapping_mode, deletion_mode, abbreviation, capacity_policy = random_instance(rng)

        oracle = BranchAndBoundOracle(
            source_subset, target_subset, table_size, mapping_mode, deletion_mode, abbreviation, capacity_policy,
            time_limit_sec=10.0,
        )
        result = oracle.solve(lexicon, labels)
        bf_best = brute_force_best(lexicon, labels, source_subset, target_subset, table_size, mapping_mode, deletion_mode, abbreviation, capacity_policy)

        agree = result["global_optimum_certified"] and (result["best_fitness"] == bf_best)
        if not agree:
            mismatches += 1

        rows.append({
            "trial": trial,
            "n_identities": len(lexicon),
            "n_labels": len(labels),
            "table_size": table_size,
            "mapping_mode": mapping_mode,
            "deletion_mode": deletion_mode,
            "abbreviation": abbreviation,
            "capacity_policy": capacity_policy,
            "oracle_fitness": result["best_fitness"],
            "brute_force_fitness": bf_best,
            "oracle_certified": result["global_optimum_certified"],
            "oracle_timed_out": result["timed_out"],
            "nodes_explored": result["nodes_explored"],
            "branches_pruned": result["branches_pruned"],
            "runtime_sec": result["runtime_sec"],
            "agree": agree,
        })

    with open(out_path, "w", newline="") as f:
        w = csv.DictWriter(f, fieldnames=list(rows[0].keys()), delimiter="\t")
        w.writeheader()
        w.writerows(rows)

    print(f"trials={n_trials} mismatches={mismatches} exactness_rate={(n_trials - mismatches) / n_trials:.4f}")
    return 0 if mismatches == 0 else 1


if __name__ == "__main__":
    sys.exit(main())
