#!/usr/bin/env python3
"""
ORDER_INVARIANCE, ALPHABET_RENAMING_INVARIANCE, and CHECKPOINT_IDENTITY checks required by
SEALED_BENCHMARK_PROTOCOL.md's pre-registered gates. Each is a real assertion over the actual
scorer/solver code, run at multiple seeds, not a one-off manual spot check.
"""
import csv
import json
import random
import sys
from pathlib import Path

sys.path.insert(0, str(Path(__file__).parent))
from scorer import Scorer
from oracle_bb import BranchAndBoundOracle

results = []


def check(name, ok, detail=""):
    results.append((name, ok, detail))
    print(f"[{'PASS' if ok else 'FAIL'}] {name}: {detail}")


def order_invariance(n_trials=50):
    rng = random.Random(4242)
    fails = 0
    for t in range(n_trials):
        n_id = rng.randint(2, 6)
        lexicon = {f"ID{i}": {"".join(rng.choice("abcdef") for _ in range(rng.randint(2, 5))) for _ in range(rng.randint(1, 2))} for i in range(n_id)}
        n_labels = rng.randint(3, 10)
        labels = [{"occurrence_id": f"O{j}", "page_id": f"P{j % 2}", "token": "".join(rng.choice("acdefh") for _ in range(rng.randint(1, 5)))} for j in range(n_labels)]
        table = {}
        srcs = rng.sample(list("abcdef"), rng.randint(0, 4))
        for s in srcs:
            table[s] = rng.choice("acdefh")

        scorer = Scorer("INJECTIVE", "DROP_UNMAPPED", "NONE", "PER_PAGE_CAPACITY_1")
        base = scorer.evaluate(table, lexicon, labels)

        shuffled_labels = labels[:]
        rng.shuffle(shuffled_labels)
        shuffled_lexicon = dict(reversed(list(lexicon.items())))
        r2 = scorer.evaluate(table, shuffled_lexicon, shuffled_labels)

        if not (base["matched"] == r2["matched"] and base["fitness"] == r2["fitness"] and base["coverage"] == r2["coverage"]):
            fails += 1
    check("ORDER_INVARIANCE", fails == 0, f"{n_trials - fails}/{n_trials} trials: identical matched/fitness/coverage under shuffled label order and reversed lexicon dict order")


def alphabet_renaming_invariance(n_trials=30):
    rng = random.Random(777)
    fails = 0
    target_alphabet = list("acdefhi")
    for t in range(n_trials):
        n_id = rng.randint(2, 5)
        lexicon = {f"ID{i}": {"".join(rng.choice("abcd") for _ in range(rng.randint(2, 4))) for _ in range(rng.randint(1, 2))} for i in range(n_id)}
        n_labels = rng.randint(3, 8)
        labels = [{"occurrence_id": f"O{j}", "page_id": "P0", "token": "".join(rng.choice(target_alphabet) for _ in range(rng.randint(1, 4)))} for j in range(n_labels)]
        table_size = rng.randint(1, 3)

        oracle1 = BranchAndBoundOracle(list("abcd"), target_alphabet, table_size, "INJECTIVE", "DROP_UNMAPPED", "NONE", "PER_PAGE_CAPACITY_1", time_limit_sec=5.0)
        r1 = oracle1.solve(lexicon, labels)

        perm = target_alphabet[:]
        rng.shuffle(perm)
        rename = dict(zip(target_alphabet, perm))
        renamed_lexicon = lexicon  # source alphabet untouched; only target/label alphabet renamed
        renamed_labels = [dict(l, token="".join(rename[c] for c in l["token"])) for l in labels]

        oracle2 = BranchAndBoundOracle(list("abcd"), perm, table_size, "INJECTIVE", "DROP_UNMAPPED", "NONE", "PER_PAGE_CAPACITY_1", time_limit_sec=5.0)
        r2 = oracle2.solve(renamed_lexicon, renamed_labels)

        if not (r1["global_optimum_certified"] and r2["global_optimum_certified"] and r1["best_fitness"] == r2["best_fitness"]):
            fails += 1
    check("ALPHABET_RENAMING_INVARIANCE", fails == 0, f"{n_trials - fails}/{n_trials} trials: identical achievable optimum fitness under a consistent target-alphabet renaming")


def checkpoint_identity(checkpoint_dir="../checkpoints", final_tsv="../NULL_FEASIBILITY_RUNS.tsv"):
    ckpt_path = Path(__file__).parent / checkpoint_dir / "NULL_CHECKPOINT_00100.json"
    final_path = Path(__file__).parent / final_tsv
    if not ckpt_path.exists() or not final_path.exists():
        check("CHECKPOINT_IDENTITY", False, f"missing {ckpt_path} or {final_path}")
        return
    with open(ckpt_path) as f:
        ckpt = json.load(f)
    with open(final_path) as f:
        final_rows = list(csv.DictReader(f, delimiter="\t"))

    ok = True
    detail = ""
    if len(ckpt["results"]) != 100:
        ok = False
        detail = f"checkpoint has {len(ckpt['results'])} results, expected 100"
    else:
        for i in range(100):
            c, r = ckpt["results"][i], final_rows[i]
            if str(c["seed"]) != r["seed"] or str(c["fitness"]) != r["fitness"] or str(c["matched"]) != r["matched"]:
                ok = False
                detail = f"row {i} mismatch: checkpoint={c} final={r}"
                break
        else:
            detail = "first 100 rows of the checkpoint written mid-run exactly match the corresponding first 100 rows of the completed final TSV"
    check("CHECKPOINT_IDENTITY", ok, detail)


def main():
    order_invariance()
    alphabet_renaming_invariance()
    checkpoint_identity()
    n_fail = sum(1 for _, ok, _ in results if not ok)
    print(f"\n{len(results) - n_fail}/{len(results)} invariance checks passed")
    return 1 if n_fail else 0


if __name__ == "__main__":
    sys.exit(main())
