#!/usr/bin/env python3
"""Runs the pre-registered dev grid (SEALED_BENCHMARK_PROTOCOL.md) with dev seeds 20260001+.
Dev results are used only for threshold sanity-checking before the code freeze; they are never
mixed with hidden results."""
import csv
import sys
from pathlib import Path

sys.path.insert(0, str(Path(__file__).parent))
from sealed_harness import build_cell_matrix, run_one_instance

N_SEEDS = 5
DEV_SEED_BASE = 20260001


def main():
    out_path = sys.argv[1] if len(sys.argv) > 1 else "DEV_GRID_RESULTS.tsv"
    cells = build_cell_matrix()
    rows = []
    total = len(cells) * N_SEEDS
    done = 0
    for cell in cells:
        for i in range(N_SEEDS):
            seed = DEV_SEED_BASE + i
            row = run_one_instance(cell, seed)
            rows.append(row)
            done += 1
            if done % 20 == 0:
                print(f"{done}/{total}", flush=True)
    with open(out_path, "w", newline="") as f:
        w = csv.DictWriter(f, fieldnames=list(rows[0].keys()), delimiter="\t")
        w.writeheader()
        w.writerows(rows)
    print(f"wrote {len(rows)} rows to {out_path}")


if __name__ == "__main__":
    main()
