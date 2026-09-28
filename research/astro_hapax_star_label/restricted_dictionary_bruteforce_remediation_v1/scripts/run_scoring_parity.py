#!/usr/bin/env python3
"""
Generates randomized small scoring instances and checks scorer.Scorer against
independent_scorer.evaluate_v2 (different matching library, different encode
implementation). Writes SCORING_PARITY_RESULTS.tsv.

Usage: python3 run_scoring_parity.py <n_trials> <out_tsv>
"""
import random
import sys
import csv
from pathlib import Path

sys.path.insert(0, str(Path(__file__).parent))
from scorer import Scorer
from independent_scorer import evaluate_v2

SRC_CHARS = list("abcdefg")
TGT_CHARS = list("acdefhi")


def random_instance(rng: random.Random):
    n_identities = rng.randint(2, 6)
    lexicon = {}
    for i in range(n_identities):
        n_forms = rng.randint(1, 2)
        forms = ["".join(rng.choice(SRC_CHARS) for _ in range(rng.randint(2, 5))) for _ in range(n_forms)]
        lexicon[f"ID{i}"] = set(forms)

    n_labels = rng.randint(3, 10)
    n_pages = rng.randint(1, 2)
    labels = []
    for j in range(n_labels):
        tok_len = rng.randint(1, 5)
        token = "".join(rng.choice(TGT_CHARS) for _ in range(tok_len))
        labels.append({
            "occurrence_id": f"OCC{j}",
            "page_id": f"P{j % n_pages}",
            "token": token,
        })

    table_size = rng.randint(0, len(SRC_CHARS))
    src_sample = rng.sample(SRC_CHARS, table_size)
    mapping_mode = rng.choice(["INJECTIVE", "MERGE_1"])
    if mapping_mode == "INJECTIVE" or table_size == 0:
        tgt_sample = rng.sample(TGT_CHARS, min(table_size, len(TGT_CHARS))) if table_size <= len(TGT_CHARS) else None
        if tgt_sample is None:
            tgt_sample = [rng.choice(TGT_CHARS) for _ in range(table_size)]
    else:
        tgt_pool = TGT_CHARS[:]
        tgt_sample = []
        dup_used = False
        for _ in range(table_size):
            if not dup_used and tgt_sample and rng.random() < 0.3:
                tgt_sample.append(rng.choice(tgt_sample))
                dup_used = True
            else:
                choice = rng.choice(tgt_pool)
                tgt_sample.append(choice)
    table = dict(zip(src_sample, tgt_sample))

    deletion_mode = rng.choice(["DROP_UNMAPPED", "SELECTIVE_VOWEL_DROP", "NONE"])
    abbreviation = rng.choice(["NONE", "SUSPENSION_1", "SUSPENSION_2", "PREFIX_4"])
    capacity_policy = rng.choice(["PER_PAGE_CAPACITY_1", "GLOBAL_CAPACITY_1"])

    return lexicon, labels, table, mapping_mode, deletion_mode, abbreviation, capacity_policy


def main():
    n_trials = int(sys.argv[1]) if len(sys.argv) > 1 else 500
    out_path = sys.argv[2] if len(sys.argv) > 2 else "SCORING_PARITY_RESULTS.tsv"
    rng = random.Random(20260917)

    rows = []
    mismatches = 0
    for trial in range(n_trials):
        lexicon, labels, table, mapping_mode, deletion_mode, abbreviation, capacity_policy = random_instance(rng)
        scorer = Scorer(mapping_mode, deletion_mode, abbreviation, capacity_policy)
        r1 = scorer.evaluate(table, lexicon, labels)
        r2 = evaluate_v2(table, lexicon, labels, mapping_mode, deletion_mode, abbreviation, capacity_policy)

        agree = (
            r1["valid_table"] == r2["valid_table"]
            and r1["matched"] == r2["matched"]
            and r1["complexity"] == r2["complexity"]
            and r1["fitness"] == r2["fitness"]
        )
        if not agree:
            mismatches += 1
        rows.append({
            "trial": trial,
            "n_identities": len(lexicon),
            "n_labels": len(labels),
            "table_size": len(table),
            "mapping_mode": mapping_mode,
            "deletion_mode": deletion_mode,
            "abbreviation": abbreviation,
            "capacity_policy": capacity_policy,
            "scorer_matched": r1["matched"],
            "independent_matched": r2["matched"],
            "scorer_fitness": r1["fitness"],
            "independent_fitness": r2["fitness"],
            "scorer_valid": r1["valid_table"],
            "independent_valid": r2["valid_table"],
            "agree": agree,
        })

    fieldnames = list(rows[0].keys())
    with open(out_path, "w", newline="") as f:
        w = csv.DictWriter(f, fieldnames=fieldnames, delimiter="\t")
        w.writeheader()
        w.writerows(rows)

    print(f"trials={n_trials} mismatches={mismatches} parity_rate={(n_trials - mismatches) / n_trials:.4f}")
    return 0 if mismatches == 0 else 1


if __name__ == "__main__":
    sys.exit(main())
