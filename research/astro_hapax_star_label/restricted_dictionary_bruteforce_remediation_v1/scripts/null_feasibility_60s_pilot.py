#!/usr/bin/env python3
"""
Small supplementary null-feasibility pilot at the budget-scaling-audit-validated 60s CP-SAT
budget and table_size=8 (the reference operating point the sealed benchmark actually qualifies
at), run in response to a methodological critique of NULL_FEASIBILITY_RESULTS.tsv: that pilot's
100% VERIFICATION_TIMEOUT rate at a 4s budget proves only that a TIME-BOUNDED (uncertified) null
pipeline is feasible, not that a full EXACT/certified null pipeline is feasible at a budget that
actually clears the synthetic gate. This pilot is intentionally small (30 replicas, not >=1000) --
it exists only to check whether null (no true correspondence at all) instances ever certify at the
now-validated budget, not to re-establish resource feasibility at full pilot scale.
"""
import csv
import resource
import sys
import time
from pathlib import Path

sys.path.insert(0, str(Path(__file__).parent))
from generator import GeneratorConfig, generate, split_public_private
from scorer import Scorer
from hybrid import HybridSolver
from sealed_harness import held_out_split, N_IDENTITIES, N_OCCURRENCES, PAGE_SPLIT

N_RUNS = 30
SEED_BASE = 90050001
CELL = dict(table_size=8, mapping_mode="INJECTIVE", deletion_mode="DROP_UNMAPPED", abbreviation="NONE", capacity_policy="PER_PAGE_CAPACITY_1")


def run_one(seed: int) -> dict:
    t0 = time.time()
    cfg = GeneratorConfig(
        seed=seed, n_identities=N_IDENTITIES, table_size=CELL["table_size"], mapping_mode=CELL["mapping_mode"],
        deletion_mode=CELL["deletion_mode"], abbreviation=CELL["abbreviation"], capacity_policy=CELL["capacity_policy"],
        noise_rate=0.0, unmatched_rate=1.0, n_occurrences=N_OCCURRENCES, page_split=PAGE_SPLIT,
    )
    instance = generate(cfg)
    public, private = split_public_private(instance)
    lexicon = {k: set(v) for k, v in public["lexicon"].items()}
    labels = public["labels"]
    source_alphabet = sorted(set(c for forms in lexicon.values() for f in forms for c in f))
    target_alphabet = public["target_alphabet"]

    train, held = held_out_split(labels, seed)
    hy = HybridSolver(
        source_alphabet, target_alphabet, CELL["table_size"], CELL["mapping_mode"], CELL["deletion_mode"],
        CELL["abbreviation"], CELL["capacity_policy"], master_seed=seed,
        neighborhood_radius=1, neighborhood_max_rounds=1, verification_time_limit_sec=60.0,
        heuristic_n_starts=6, heuristic_max_iters=300,
    )
    result = hy.solve(lexicon, train)
    scorer = Scorer(CELL["mapping_mode"], CELL["deletion_mode"], CELL["abbreviation"], CELL["capacity_policy"])
    full_eval = scorer.evaluate(result["final_table"], lexicon, labels)
    elapsed = time.time() - t0
    return {
        "seed": seed, "wall_time_sec": round(elapsed, 4),
        "peak_rss_kb": resource.getrusage(resource.RUSAGE_SELF).ru_maxrss,
        "classification": result["classification"], "coverage": round(full_eval["coverage"], 4),
        "spurious_high_coverage": full_eval["coverage"] > 0.5,
    }


def main():
    out_path = sys.argv[1] if len(sys.argv) > 1 else "NULL_FEASIBILITY_60S_PILOT.tsv"
    rows = [run_one(SEED_BASE + i) for i in range(N_RUNS)]
    with open(out_path, "w", newline="") as f:
        w = csv.DictWriter(f, fieldnames=list(rows[0].keys()), delimiter="\t")
        w.writeheader()
        w.writerows(rows)
    from collections import Counter
    print(Counter(r["classification"] for r in rows))
    print("mean wall_time_sec:", round(sum(r["wall_time_sec"] for r in rows) / len(rows), 3))
    print("mean coverage:", round(sum(r["coverage"] for r in rows) / len(rows), 4))
    print(f"wrote {len(rows)} rows to {out_path}")


if __name__ == "__main__":
    main()
