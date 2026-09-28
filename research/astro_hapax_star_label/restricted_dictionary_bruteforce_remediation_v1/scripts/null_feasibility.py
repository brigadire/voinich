#!/usr/bin/env python3
"""
clean_room.md Section 19. Runs >=1000 FULL synthetic null replicas -- each one exercises the
complete pipeline (fresh generation, train/held-out split, full multi-start heuristic retraining,
hybrid verification, checkpointing, aggregation) -- at a fixed, disclosed reference operating
point (table_size=4, matching the fastest main-grid cell; a reduced-scale timed pilot is
established practice in this project's own history, see restricted_dictionary_bruteforce_v3_audit's
300-replica pilot, Finding F016). Production nulls (on real data) are never run -- this measures
resource cost and null-condition behavior on synthetic data only.

The null condition itself: unmatched_rate=1.0 (every label is a pure distractor, no lexicon
identity has any true correspondence at all) -- this is the SHUFFLED_LABELS/no-real-correspondence
control the task's own audit found completely missing from v3 (Finding F017).

Checkpointing: every CHECKPOINT_EVERY replicas, partial results + timing are flushed to
checkpoints/NULL_CHECKPOINT_<n>.json so a killed/resumed run does not lose completed work.
"""
import json
import random
import resource
import sys
import time
from pathlib import Path

sys.path.insert(0, str(Path(__file__).parent))
from generator import GeneratorConfig, generate, split_public_private
from scorer import Scorer
from hybrid import HybridSolver
from sealed_harness import held_out_split

N_RUNS = 1000
CHECKPOINT_EVERY = 100
NULL_SEED_BASE = 90000001
CELL = dict(
    table_size=4, mapping_mode="INJECTIVE", deletion_mode="DROP_UNMAPPED", abbreviation="NONE",
    capacity_policy="PER_PAGE_CAPACITY_1",
)


def run_one_null(seed: int) -> dict:
    t0 = time.time()
    cfg = GeneratorConfig(
        seed=seed, n_identities=8, table_size=CELL["table_size"], mapping_mode=CELL["mapping_mode"],
        deletion_mode=CELL["deletion_mode"], abbreviation=CELL["abbreviation"], capacity_policy=CELL["capacity_policy"],
        noise_rate=0.0, unmatched_rate=1.0, n_occurrences=30, page_split=(16, 14),
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
        neighborhood_radius=1, neighborhood_max_rounds=1, verification_time_limit_sec=4.0,
        heuristic_n_starts=5, heuristic_max_iters=200,
    )
    result = hy.solve(lexicon, train)

    scorer = Scorer(CELL["mapping_mode"], CELL["deletion_mode"], CELL["abbreviation"], CELL["capacity_policy"])
    truth_by_occ = private["true_identity_by_occurrence"]
    held_with_truth = [dict(l, true_identity=truth_by_occ[l["occurrence_id"]]) for l in held]
    full_eval = scorer.evaluate(result["final_table"], lexicon, labels, held_out_labels=held_with_truth)

    elapsed = time.time() - t0
    peak_rss_kb = resource.getrusage(resource.RUSAGE_SELF).ru_maxrss

    return {
        "seed": seed, "wall_time_sec": round(elapsed, 4), "peak_rss_kb": peak_rss_kb,
        "classification": result["classification"], "matched": full_eval["matched"],
        "total": full_eval["total"], "coverage": round(full_eval["coverage"], 4),
        "fitness": full_eval["fitness"],
        "held_out_assignment_accuracy": full_eval.get("held_out_assignment_accuracy"),
        "spurious_high_coverage": full_eval["coverage"] > 0.5,  # coverage should stay LOW under a genuine null
    }


def main():
    out_path = Path(sys.argv[1]) if len(sys.argv) > 1 else Path("NULL_FEASIBILITY_RUNS.tsv")
    ckpt_dir = Path(sys.argv[2]) if len(sys.argv) > 2 else Path("checkpoints")
    ckpt_dir.mkdir(parents=True, exist_ok=True)

    results = []
    t_start = time.time()
    for i in range(N_RUNS):
        seed = NULL_SEED_BASE + i
        r = run_one_null(seed)
        results.append(r)
        if (i + 1) % CHECKPOINT_EVERY == 0:
            ckpt_path = ckpt_dir / f"NULL_CHECKPOINT_{i + 1:05d}.json"
            with open(ckpt_path, "w") as f:
                json.dump({"n_completed": i + 1, "elapsed_sec": round(time.time() - t_start, 2), "results": results}, f)
            print(f"checkpoint {i + 1}/{N_RUNS} written to {ckpt_path} (elapsed {time.time() - t_start:.1f}s)", flush=True)

    import csv
    with open(out_path, "w", newline="") as f:
        w = csv.DictWriter(f, fieldnames=list(results[0].keys()), delimiter="\t")
        w.writeheader()
        w.writerows(results)
    print(f"wrote {len(results)} null runs to {out_path}, total wall time {time.time() - t_start:.1f}s")


if __name__ == "__main__":
    main()
