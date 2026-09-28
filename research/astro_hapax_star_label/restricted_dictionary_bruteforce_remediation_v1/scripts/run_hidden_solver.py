#!/usr/bin/env python3
"""
Step 8 of SEALED_BENCHMARK_PROTOCOL.md. Reads ONLY hidden/PUBLIC_SEALED.jsonl. Never opens
TRUTH_SEALED.jsonl -- this is checked mechanically by SCIENTIFIC_SAFETY_TESTS.md (grep for
"TRUTH_SEALED" and "true_identity" in this file: must find neither). Writes HIDDEN_PREDICTIONS.tsv:
the predicted table and the resulting assignments over the FULL public label set (train+held-out),
computed purely from public tokens/lexicon -- no ground truth is needed to know which identity a
predicted table assigns a label to.
"""
import csv
import json
import sys
from pathlib import Path

sys.path.insert(0, str(Path(__file__).parent))
from hybrid import HybridSolver
from scorer import Scorer


def held_out_split_public(labels, seed):
    import random
    rng = random.Random(seed * 31 + 7)
    by_page = {}
    for l in labels:
        by_page.setdefault(l["page_id"], []).append(l)
    train, held = [], []
    for page, page_labels in sorted(by_page.items()):
        shuffled = page_labels[:]
        rng.shuffle(shuffled)
        n_held = max(1, round(0.2 * len(shuffled)))
        held.extend(shuffled[:n_held])
        train.extend(shuffled[n_held:])
    return train, held


def main():
    public_path = sys.argv[1] if len(sys.argv) > 1 else "../hidden/PUBLIC_SEALED.jsonl"
    out_path = sys.argv[2] if len(sys.argv) > 2 else "../HIDDEN_PREDICTIONS.tsv"

    rows = []
    with open(public_path) as f:
        lines = f.readlines()

    for i, line in enumerate(lines):
        rec = json.loads(line)
        cell, seed, public = rec["cell"], rec["seed"], rec["public"]
        lexicon = {k: set(v) for k, v in public["lexicon"].items()}
        labels = public["labels"]
        target_alphabet = public["target_alphabet"]
        source_alphabet = sorted(set(c for forms in lexicon.values() for f in forms for c in f))

        train, held = held_out_split_public(labels, seed)
        ts = cell["table_size"]
        # Must exactly match sealed_harness.py's run_one_instance budget/heuristic parameters --
        # dev and hidden may differ only in seed. See that file's comment for why 60.0.
        vt = 60.0
        hy = HybridSolver(
            source_alphabet, target_alphabet, ts, cell["mapping_mode"], cell["deletion_mode"],
            cell["abbreviation"], cell["capacity_policy"], master_seed=seed,
            neighborhood_radius=1, neighborhood_max_rounds=1, verification_time_limit_sec=vt,
            heuristic_n_starts=6, heuristic_max_iters=300,
        )
        result = hy.solve(lexicon, train)
        predicted_table = result["final_table"]

        scorer = Scorer(cell["mapping_mode"], cell["deletion_mode"], cell["abbreviation"], cell["capacity_policy"])
        full_eval = scorer.evaluate(predicted_table, lexicon, labels)

        held_occ_ids = sorted(l["occurrence_id"] for l in held)
        rows.append({
            "cell_id": cell["cell_id"], "seed": seed,
            "predicted_table_json": json.dumps(predicted_table, sort_keys=True),
            "classification": result["classification"], "verifier_mode": result["verifier_mode"],
            "matched": full_eval["matched"], "total": full_eval["total"], "fitness": full_eval["fitness"],
            "coverage": full_eval["coverage"],
            "assignments_json": json.dumps(full_eval["assignments"], sort_keys=True),
            "held_out_occurrence_ids_json": json.dumps(held_occ_ids),
        })
        if (i + 1) % 20 == 0:
            print(f"{i + 1}/{len(lines)}", flush=True)

    with open(out_path, "w", newline="") as f:
        w = csv.DictWriter(f, fieldnames=list(rows[0].keys()), delimiter="\t")
        w.writeheader()
        w.writerows(rows)
    print(f"wrote {len(rows)} predictions to {out_path}")


if __name__ == "__main__":
    main()
