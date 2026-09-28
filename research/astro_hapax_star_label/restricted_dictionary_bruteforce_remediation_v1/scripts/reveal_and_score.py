#!/usr/bin/env python3
"""
Steps 9-10 of SEALED_BENCHMARK_PROTOCOL.md. This is the FIRST script in the pipeline allowed to
open hidden/TRUTH_SEALED.jsonl -- it does so only after HIDDEN_PREDICTIONS.tsv already exists
(predictions were committed before truth was read). Computes HIDDEN_RESULTS.tsv mechanically from
the two files; no numbers in this file are hand-entered.
"""
import csv
import json
import sys
from pathlib import Path

sys.path.insert(0, str(Path(__file__).parent))
from scorer import Scorer


def main():
    truth_path = sys.argv[1] if len(sys.argv) > 1 else "../hidden/TRUTH_SEALED.jsonl"
    predictions_path = sys.argv[2] if len(sys.argv) > 2 else "../HIDDEN_PREDICTIONS.tsv"
    out_path = sys.argv[3] if len(sys.argv) > 3 else "../HIDDEN_RESULTS.tsv"
    public_path = sys.argv[4] if len(sys.argv) > 4 else "../hidden/PUBLIC_SEALED.jsonl"

    truth_by_key = {}
    with open(truth_path) as f:
        for line in f:
            rec = json.loads(line)
            truth_by_key[(rec["cell_id"], rec["seed"])] = rec["private"]

    public_by_key = {}
    with open(public_path) as f:
        for line in f:
            rec = json.loads(line)
            public_by_key[(rec["cell"]["cell_id"], rec["seed"])] = rec

    rows = []
    with open(predictions_path) as f:
        reader = csv.DictReader(f, delimiter="\t")
        for pred in reader:
            key = (pred["cell_id"], int(pred["seed"]))
            truth = truth_by_key[key]
            true_table = truth["true_table"]
            true_identity_by_occ = truth["true_identity_by_occurrence"]
            predicted_table = json.loads(pred["predicted_table_json"])
            assignments = json.loads(pred["assignments_json"])
            held_occ_ids = set(json.loads(pred["held_out_occurrence_ids_json"]))

            exact_table_recovery = predicted_table == true_table

            pub_rec = public_by_key[key]
            cell = pub_rec["cell"]
            lexicon = {k: set(v) for k, v in pub_rec["public"]["lexicon"].items()}
            all_labels = pub_rec["public"]["labels"]
            scorer = Scorer(cell["mapping_mode"], cell["deletion_mode"], cell["abbreviation"], cell["capacity_policy"])
            true_fitness = scorer.evaluate(true_table, lexicon, all_labels)["fitness"]
            equivalence_aware_table_recovery = float(pred["fitness"]) >= true_fitness

            tp = sum(1 for occ, ident in assignments.items() if true_identity_by_occ.get(occ) == ident)
            fp = sum(1 for occ, ident in assignments.items() if true_identity_by_occ.get(occ) != ident)
            n_true_matched = sum(1 for v in true_identity_by_occ.values() if v is not None)
            fn = n_true_matched - tp
            precision = tp / (tp + fp) if (tp + fp) else 0.0
            recall = tp / (tp + fn) if (tp + fn) else 0.0

            total_occ = len(true_identity_by_occ)
            correct_assign = sum(1 for occ, ident in assignments.items() if true_identity_by_occ.get(occ) == ident)
            assignment_accuracy = correct_assign / total_occ if total_occ else 0.0

            held_correct = sum(1 for occ in held_occ_ids if assignments.get(occ) == true_identity_by_occ.get(occ) and true_identity_by_occ.get(occ) is not None)
            held_out_assignment_accuracy = held_correct / len(held_occ_ids) if held_occ_ids else 0.0

            rows.append({
                "cell_id": pred["cell_id"], "seed": pred["seed"],
                "exact_table_recovery": exact_table_recovery,
                "equivalence_aware_table_recovery": equivalence_aware_table_recovery,
                "mapping_precision": precision, "mapping_recall": recall,
                "assignment_accuracy": assignment_accuracy,
                "held_out_assignment_accuracy": held_out_assignment_accuracy,
                "matched": pred["matched"], "total": pred["total"], "coverage": pred["coverage"],
                "fitness": pred["fitness"], "classification": pred["classification"],
            })

    with open(out_path, "w", newline="") as f:
        w = csv.DictWriter(f, fieldnames=list(rows[0].keys()), delimiter="\t")
        w.writeheader()
        w.writerows(rows)
    print(f"wrote {len(rows)} rows to {out_path}")


if __name__ == "__main__":
    main()
