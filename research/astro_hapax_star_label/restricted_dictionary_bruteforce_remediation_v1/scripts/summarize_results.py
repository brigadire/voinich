#!/usr/bin/env python3
"""
clean_room.md Section 15/16: aggregates a results TSV (DEV_GRID_RESULTS.tsv or HIDDEN_RESULTS.tsv
joined with cell metadata) into macro/micro/median/CI summaries per required mode, and checks the
pre-registered gates from SEALED_BENCHMARK_PROTOCOL.md. The mean is never reported alone -- worst
required mode is always shown alongside it, per the task's explicit instruction that an average
must not hide a mandatory mode's failure.
"""
import csv
import math
import sys
from collections import defaultdict


def mean(xs):
    xs = list(xs)
    return sum(xs) / len(xs) if xs else 0.0


def median(xs):
    xs = sorted(xs)
    n = len(xs)
    if n == 0:
        return 0.0
    mid = n // 2
    return xs[mid] if n % 2 else (xs[mid - 1] + xs[mid]) / 2


def ci95(xs):
    xs = list(xs)
    n = len(xs)
    if n < 2:
        return (0.0, 0.0)
    m = mean(xs)
    var = sum((x - m) ** 2 for x in xs) / (n - 1)
    se = math.sqrt(var / n)
    return (round(m - 1.96 * se, 4), round(m + 1.96 * se, 4))


def to_bool(v):
    return str(v).strip() in ("True", "true", "1")


def main():
    path = sys.argv[1] if len(sys.argv) > 1 else "DEV_GRID_RESULTS.tsv"
    with open(path) as f:
        rows = list(csv.DictReader(f, delimiter="\t"))

    by_cell = defaultdict(list)
    for r in rows:
        by_cell[r["cell_id"]].append(r)

    print(f"{'cell_id':45s} {'n':>3s} {'eq_recovery':>11s} {'prec':>7s} {'recall':>7s} {'held_acc':>9s}")
    macro_eq, macro_prec, macro_recall, macro_held = [], [], [], []
    all_eq, all_prec, all_recall, all_held = [], [], [], []
    worst = {"eq_recovery": (1.0, None), "prec": (1.0, None), "recall": (1.0, None), "held_acc": (1.0, None)}

    for cell_id, cell_rows in sorted(by_cell.items()):
        eq = [1.0 if to_bool(r.get("equivalence_aware_table_recovery", "False")) else 0.0 for r in cell_rows]
        prec = [float(r["mapping_precision"]) for r in cell_rows]
        recall = [float(r["mapping_recall"]) for r in cell_rows]
        held = [float(r["held_out_assignment_accuracy"]) for r in cell_rows if r.get("held_out_assignment_accuracy") not in (None, "", "None")]

        m_eq, m_prec, m_recall, m_held = mean(eq), mean(prec), mean(recall), mean(held) if held else 0.0
        print(f"{cell_id:45s} {len(cell_rows):3d} {m_eq:11.3f} {m_prec:7.3f} {m_recall:7.3f} {m_held:9.3f}")

        macro_eq.append(m_eq); macro_prec.append(m_prec); macro_recall.append(m_recall); macro_held.append(m_held)
        all_eq.extend(eq); all_prec.extend(prec); all_recall.extend(recall); all_held.extend(held)
        for key, val in (("eq_recovery", m_eq), ("prec", m_prec), ("recall", m_recall), ("held_acc", m_held)):
            if val < worst[key][0]:
                worst[key] = (val, cell_id)

    print("\n--- Aggregates ---")
    print(f"MACRO (mean of per-cell means): eq_recovery={mean(macro_eq):.3f} precision={mean(macro_prec):.3f} recall={mean(macro_recall):.3f} held_out_acc={mean(macro_held):.3f}")
    print(f"MICRO (pooled over all runs):   eq_recovery={mean(all_eq):.3f} precision={mean(all_prec):.3f} recall={mean(all_recall):.3f} held_out_acc={mean(all_held):.3f}")
    print(f"MEDIAN (per-run):               eq_recovery={median(all_eq):.3f} precision={median(all_prec):.3f} recall={median(all_recall):.3f} held_out_acc={median(all_held):.3f}")
    print(f"95% CI (per-run, normal approx): eq_recovery={ci95(all_eq)} precision={ci95(all_prec)} recall={ci95(all_recall)} held_out_acc={ci95(all_held)}")
    print(f"\nWORST REQUIRED MODE (macro): eq_recovery {worst['eq_recovery']}, precision {worst['prec']}, recall {worst['recall']}, held_out_acc {worst['held_acc']}")


if __name__ == "__main__":
    main()
