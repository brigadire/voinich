#!/usr/bin/env python3
"""
Budget-scaling diagnostic audit, run in response to a methodological critique of the sealed
hidden benchmark's original FAILURE_DIAGNOSIS.md conclusion (see REMEDIATION_REPORT.md
"Budget-scaling audit" section for the full account). This is NOT a re-qualification pass and
does not by itself authorize anything -- it exists only to answer one narrow question: if the
CP-SAT verifier is given more time, does recovery improve UNCONDITIONALLY (averaged over all
instances, not just the ones that happened to certify), or does only the certification RATE of
already-easy instances improve while hard/noisy/composition instances remain wrong or unresolved?

Design (a disclosed reduction from the critique's full proposal, for wall-clock feasibility):
- Fresh, sealed seeds (33330001+), disjoint from both the dev (20260001+) and hidden (27170001+)
  ranges used elsewhere in this package.
- 5 cells: two "easy" exact-CP-SAT-scope cells (small/medium table_size, no noise/unmatched),
  two "hard" exact-scope cells (25% noise, 50% unmatched, medium/large table_size), and one
  neighborhood-only cell (SELECTIVE_VOWEL_DROP, permanently outside CP-SAT's scope) as a control
  that budget scaling should NOT help, since that path never even attempts exact verification.
- 2 seeds per cell (not 20+) -- another disclosed reduction; this audit is diagnostic, not a
  fresh qualification claim, so it does not need the same seed count as a qualification gate.
- Budgets: 5, 15, 60, 300, 1200 seconds, exactly as specified in the critique.
- THE SAME generated instance (lexicon + labels + train/held split) is reused across all 5
  budgets for a given (cell, seed) -- only the CP-SAT verifier's time_limit_sec changes. This is
  the load-bearing design choice: it isolates the effect of budget from instance-to-instance
  variance, which the original sealed benchmark (different seeds per cell, no budget sweep)
  could not do.
- For the neighborhood-only cell, `verification_time_limit_sec` is passed through unchanged but
  has NO EFFECT on the result (the neighborhood verifier does not consult it at all) -- this is
  itself the finding, not a bug, and is recorded explicitly.
"""
import csv
import sys
import time
from pathlib import Path

sys.path.insert(0, str(Path(__file__).parent))
from generator import generate, split_public_private
from scorer import Scorer
from hybrid import HybridSolver
from sealed_harness import held_out_split, N_IDENTITIES, N_OCCURRENCES, PAGE_SPLIT
from generator import GeneratorConfig

BUDGETS = [5, 15, 60, 300, 1200]
SEED_BASE = 33330001
N_SEEDS = 2

AUDIT_CELLS = [
    dict(cell_id="AUDIT_EASY_TS4", table_size=4, noise_rate=0.0, unmatched_rate=0.0,
         mapping_mode="INJECTIVE", deletion_mode="DROP_UNMAPPED", abbreviation="NONE", capacity_policy="PER_PAGE_CAPACITY_1"),
    dict(cell_id="AUDIT_EASY_TS8", table_size=8, noise_rate=0.0, unmatched_rate=0.0,
         mapping_mode="INJECTIVE", deletion_mode="DROP_UNMAPPED", abbreviation="NONE", capacity_policy="PER_PAGE_CAPACITY_1"),
    dict(cell_id="AUDIT_HARD_TS8", table_size=8, noise_rate=0.25, unmatched_rate=0.50,
         mapping_mode="INJECTIVE", deletion_mode="DROP_UNMAPPED", abbreviation="NONE", capacity_policy="PER_PAGE_CAPACITY_1"),
    dict(cell_id="AUDIT_HARD_TS12", table_size=12, noise_rate=0.25, unmatched_rate=0.50,
         mapping_mode="INJECTIVE", deletion_mode="DROP_UNMAPPED", abbreviation="NONE", capacity_policy="PER_PAGE_CAPACITY_1"),
    dict(cell_id="AUDIT_NEIGHBORHOOD_ONLY_TS8", table_size=8, noise_rate=0.10, unmatched_rate=0.25,
         mapping_mode="INJECTIVE", deletion_mode="SELECTIVE_VOWEL_DROP", abbreviation="NONE", capacity_policy="PER_PAGE_CAPACITY_1"),
]


def main():
    out_path = sys.argv[1] if len(sys.argv) > 1 else "BUDGET_SCALING_AUDIT_RESULTS.tsv"
    rows = []

    for cell in AUDIT_CELLS:
        for i in range(N_SEEDS):
            seed = SEED_BASE + i * 17 + AUDIT_CELLS.index(cell) * 101
            cfg = GeneratorConfig(
                seed=seed, n_identities=N_IDENTITIES, table_size=cell["table_size"], mapping_mode=cell["mapping_mode"],
                deletion_mode=cell["deletion_mode"], abbreviation=cell["abbreviation"], capacity_policy=cell["capacity_policy"],
                noise_rate=cell["noise_rate"], unmatched_rate=cell["unmatched_rate"],
                n_occurrences=N_OCCURRENCES, page_split=PAGE_SPLIT,
            )
            instance = generate(cfg)
            public, private = split_public_private(instance)
            lexicon = {k: set(v) for k, v in public["lexicon"].items()}
            all_labels = public["labels"]
            source_alphabet = sorted(set(c for forms in lexicon.values() for f in forms for c in f))
            target_alphabet = public["target_alphabet"]
            true_table = private["true_table"]
            truth_by_occ = private["true_identity_by_occurrence"]

            train, held = held_out_split(all_labels, seed)
            scorer = Scorer(cell["mapping_mode"], cell["deletion_mode"], cell["abbreviation"], cell["capacity_policy"])
            true_fitness_full = scorer.evaluate(true_table, lexicon, all_labels)["fitness"]

            for budget in BUDGETS:
                t0 = time.time()
                hy = HybridSolver(
                    source_alphabet, target_alphabet, cell["table_size"], cell["mapping_mode"], cell["deletion_mode"],
                    cell["abbreviation"], cell["capacity_policy"], master_seed=seed,
                    neighborhood_radius=1, neighborhood_max_rounds=1, verification_time_limit_sec=float(budget),
                    heuristic_n_starts=6, heuristic_max_iters=300,
                )
                result = hy.solve(lexicon, train)
                elapsed = time.time() - t0
                predicted_table = result["final_table"]

                held_with_truth = [dict(l, true_identity=truth_by_occ[l["occurrence_id"]]) for l in held]
                full_eval = scorer.evaluate(predicted_table, lexicon, all_labels, held_out_labels=held_with_truth)

                all_with_truth = [dict(l, true_identity=truth_by_occ[l["occurrence_id"]]) for l in all_labels]
                pred_assign = scorer.evaluate(predicted_table, lexicon, all_with_truth)["assignments"]
                tp = sum(1 for occ, ident in pred_assign.items() if truth_by_occ.get(occ) == ident)
                fp = sum(1 for occ, ident in pred_assign.items() if truth_by_occ.get(occ) != ident)
                n_true_matched = sum(1 for v in truth_by_occ.values() if v is not None)
                fn = n_true_matched - tp
                precision = tp / (tp + fp) if (tp + fp) else 0.0
                recall = tp / (tp + fn) if (tp + fn) else 0.0

                rows.append({
                    "cell_id": cell["cell_id"], "seed": seed, "budget_sec": budget,
                    "table_size": cell["table_size"], "noise_rate": cell["noise_rate"], "unmatched_rate": cell["unmatched_rate"],
                    "deletion_mode": cell["deletion_mode"], "abbreviation": cell["abbreviation"],
                    "classification": result["classification"], "verifier_mode": result["verifier_mode"],
                    "wall_time_sec": round(elapsed, 3),
                    "exact_table_recovery": predicted_table == true_table,
                    "equivalence_aware_table_recovery": full_eval["fitness"] >= true_fitness_full,
                    "mapping_precision": round(precision, 4), "mapping_recall": round(recall, 4),
                    "held_out_assignment_accuracy": full_eval.get("held_out_assignment_accuracy"),
                    "achieved_fitness": full_eval["fitness"], "true_fitness_full": true_fitness_full,
                    "optimality_gap": result.get("optimality_gap"),
                })
                print(f"{cell['cell_id']} seed={seed} budget={budget}s -> {result['classification']} "
                      f"eq_recovery={full_eval['fitness'] >= true_fitness_full} recall={recall:.3f} wall={elapsed:.2f}s", flush=True)

    with open(out_path, "w", newline="") as f:
        w = csv.DictWriter(f, fieldnames=list(rows[0].keys()), delimiter="\t")
        w.writeheader()
        w.writerows(rows)
    print(f"\nwrote {len(rows)} rows to {out_path}")


if __name__ == "__main__":
    main()
