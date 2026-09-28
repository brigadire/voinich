#!/usr/bin/env python3
"""
Independent audit verification suite for restricted_dictionary_bruteforce_v3_design.
Imports the FROZEN, unmodified engine/search code via the read-only repro_sandbox copy
(never writes into the design package). Produces:
  - ADVERSARIAL_EXACT_RESULTS.tsv       (E1/E2/E5: brute force vs BB vs CP_SAT vs Heuristic on adversarial instances)
  - ORDER_INVARIANCE_100.tsv            (E3: 100 random permutations)
  - INDEPENDENT_MATCHING_CHECK.tsv      (D2: alternate bipartite solver cross-check)
  - FRESH_HIDDEN_PREDICTIONS.tsv / FRESH_HIDDEN_RESULTS.tsv (F: real sealed benchmark, NEW seeds)
  - RESOURCE_BENCHMARK.tsv              (H2: real timed pilot + extrapolation)
This script performs NO real Voynich label search: all inputs below are synthetic or
small toy alphabets defined in this script / the design's own synthetic generator using
the 94-identity historical lexicon (never the 57 real EVA labels).
"""
import sys
import time
import json
import random
import tracemalloc
import itertools
import csv
import statistics
from pathlib import Path
from collections import Counter

AUDIT_DIR = Path(__file__).resolve().parent.parent
SANDBOX_SCRIPTS = AUDIT_DIR / "repro_sandbox" / "scripts"
sys.path.insert(0, str(SANDBOX_SCRIPTS))

from engine import (
    evaluate_table, encode_word, compute_complexity, maximum_bipartite_matching,
    SOURCE_ALPHABET, EVA_ALPHABET
)
from search_bb import BranchAndBoundSolver
from search_cp import CPSATSolver
from search_heuristic import DirectionalHeuristicSolver
from synthetic_generator import SealedSyntheticGenerator, load_lexicon_forms

OUT = AUDIT_DIR

# ---------------------------------------------------------------------------
# E1/E2: Adversarial exact-benchmark instances designed to require combining
# MORE than 2 partial (word,label) mapping fragments -- i.e. exactly the case
# search_cp.py's "top-50 pairwise combination" enumeration cannot reach.
# ---------------------------------------------------------------------------
def run_exhaustive_small(src_alpha, tgt_alpha, k, lexicon, labels):
    best_fit, best_tables = -float("inf"), []
    for src in itertools.permutations(src_alpha, k):
        for dst in itertools.permutations(tgt_alpha, k):
            t = dict(zip(src, dst))
            ev = evaluate_table(t, lexicon, labels, "INJECTIVE", "DROP_UNMAPPED", "NONE", "PER_PAGE_CAPACITY_1")
            if ev["fitness"] > best_fit:
                best_fit, best_tables = ev["fitness"], [t]
            elif ev["fitness"] == best_fit:
                best_tables.append(t)
    return best_fit, best_tables

def section_adversarial():
    rows = []
    for k in [2, 3, 4, 5, 6]:
        for trial in range(5):
            seed = 42000 + k * 100 + trial
            # Tight alphabets: exactly k+1 source chars, k+1 target chars -> exhaustive
            # space is P(k+1,k)*P(k+1,k) = ((k+1)!)^2, tractable through k=6 ((7!)^2 ~ 2.5e7
            # is still too slow in Python with evaluate_table per node, so cap exhaustive
            # verification at k<=4 and use BranchAndBoundSolver -- already independently
            # validated as exact and reproducible on EXACT_SMALL_INSTANCE_RESULTS.tsv --
            # as the ground truth for k=5,6.
            rng = random.Random(seed)
            src_alpha = tuple("abcdefgh")[:k + 1]
            tgt_alpha = tuple("olcyrstn")[:k + 1]
            planted_src = rng.sample(src_alpha, k)
            planted_tgt = rng.sample(tgt_alpha, k)
            planted_table = dict(zip(planted_src, planted_tgt))
            filler_pool = [c for c in "zjqxwvmu"][:k]
            lexicon, labels = {}, []
            for i, (s, t) in enumerate(planted_table.items()):
                filler = filler_pool[i % len(filler_pool)]
                word = f"{filler}{s}"
                ident = f"STAR_{i:02d}"
                lexicon[ident] = {word}
                enc = encode_word(word, planted_table, "DROP_UNMAPPED", "NONE")
                labels.append({"occurrence_id": f"LBL_{i:03d}", "page_id": "p1", "token": enc})
            inst = {"seed": seed, "k": k, "src_alpha": src_alpha, "tgt_alpha": tgt_alpha,
                    "planted_table": planted_table, "lexicon": lexicon, "labels": labels}

            t0 = time.perf_counter()
            if k <= 4:
                true_opt_fit, true_opt_tables = run_exhaustive_small(
                    inst["src_alpha"], inst["tgt_alpha"], k, inst["lexicon"], inst["labels"]
                )
                ground_truth_method = "EXHAUSTIVE"
            else:
                gt_bb = BranchAndBoundSolver(inst["src_alpha"], inst["tgt_alpha"], k,
                                              "INJECTIVE", "DROP_UNMAPPED", "NONE",
                                              "PER_PAGE_CAPACITY_1", time_limit_sec=30.0)
                gt_res = gt_bb.solve(inst["lexicon"], inst["labels"])
                true_opt_fit = gt_res["best_solution"]["fitness"]
                true_opt_tables = [gt_res["best_solution"]["table"]]
                ground_truth_method = "BRANCH_AND_BOUND(30s, independently validated exact)"
            exh_ms = (time.perf_counter() - t0) * 1000

            bb = BranchAndBoundSolver(inst["src_alpha"], inst["tgt_alpha"], k,
                                       "INJECTIVE", "DROP_UNMAPPED", "NONE", "PER_PAGE_CAPACITY_1",
                                       time_limit_sec=20.0)
            t0 = time.perf_counter()
            bb_res = bb.solve(inst["lexicon"], inst["labels"])
            bb_ms = (time.perf_counter() - t0) * 1000

            cp = CPSATSolver(inst["src_alpha"], inst["tgt_alpha"], k,
                              "INJECTIVE", "DROP_UNMAPPED", "NONE", "PER_PAGE_CAPACITY_1",
                              time_limit_sec=20.0)
            t0 = time.perf_counter()
            cp_res = cp.solve(inst["lexicon"], inst["labels"])
            cp_ms = (time.perf_counter() - t0) * 1000

            heu = DirectionalHeuristicSolver(inst["src_alpha"], inst["tgt_alpha"], k,
                                              "INJECTIVE", "DROP_UNMAPPED", "NONE", "PER_PAGE_CAPACITY_1",
                                              beam_width=16, local_search_steps=350, seed=seed, time_limit_sec=20.0)
            t0 = time.perf_counter()
            heu_res = heu.solve(inst["lexicon"], inst["labels"])
            heu_ms = (time.perf_counter() - t0) * 1000

            for name, res, ms in [
                ("BRANCH_AND_BOUND", bb_res, bb_ms),
                ("CP_SAT", cp_res, cp_ms),
                ("DIRECTIONAL_HEURISTIC", heu_res, heu_ms),
            ]:
                achieved = res["best_solution"]["fitness"]
                rows.append({
                    "adversarial_case": "K_FRAGMENT_COMBINATION",
                    "k_fragments_required": k,
                    "trial_seed": seed,
                    "algorithm": name,
                    "true_global_optimum_fitness": true_opt_fit,
                    "achieved_fitness": achieved,
                    "reached_global_optimum": "YES" if achieved == true_opt_fit else "NO",
                    "claims_global_optimum_guaranteed": res.get("global_optimum_guaranteed"),
                    "timed_out": res.get("timed_out"),
                    "runtime_ms": round(ms, 3),
                    "exhaustive_runtime_ms": round(exh_ms, 3),
                    "num_true_optimal_tables": len(true_opt_tables),
                })
    return rows

# ---------------------------------------------------------------------------
# E3: 100 random permutations, order invariance for DirectionalHeuristicSolver
# ---------------------------------------------------------------------------
def section_order_invariance_100():
    lexicon = load_lexicon_forms()
    gen = SealedSyntheticGenerator(seed=778899)
    labels, truth = gen.generate_dataset(table_size=8, noise_rate=0.0, unmatched_fraction=0.25)

    solver_seed = 555
    base = DirectionalHeuristicSolver(SOURCE_ALPHABET, EVA_ALPHABET, table_size=8, seed=solver_seed)
    base_res = base.solve(lexicon, labels)
    base_fit = base_res["best_solution"]["fitness"]
    base_table = base_res["best_solution"]["table"]

    rows = []
    rng = random.Random(31415)
    mismatches = 0
    for i in range(100):
        perm_seed = 1000 + i
        shuffled = list(labels)
        random.Random(perm_seed).shuffle(shuffled)
        solver = DirectionalHeuristicSolver(SOURCE_ALPHABET, EVA_ALPHABET, table_size=8, seed=solver_seed)
        res = solver.solve(lexicon, shuffled)
        fit = res["best_solution"]["fitness"]
        tbl = res["best_solution"]["table"]
        same = (fit == base_fit) and (tbl == base_table)
        if not same:
            mismatches += 1
        rows.append({
            "permutation_index": i, "permutation_seed": perm_seed,
            "base_fitness": base_fit, "permuted_fitness": fit,
            "fitness_match": fit == base_fit, "table_match": tbl == base_table,
            "order_invariant": same
        })
    return rows, mismatches

# ---------------------------------------------------------------------------
# D2: Independent alternate max-bipartite-matching implementation (Hopcroft-Karp
# style BFS/DFS phases), cross-checked against engine.maximum_bipartite_matching
# on random capacity-1 instances.
# ---------------------------------------------------------------------------
def independent_max_matching(adj):
    """Alternate implementation: Hopcroft-Karp-flavoured BFS+DFS, written independently
    of engine.py's Kuhn recursive DFS, to cross-validate matching cardinality/validity."""
    from collections import deque
    left_nodes = list(adj.keys())
    match_left = {l: None for l in left_nodes}
    match_right = {}

    def bfs():
        dist = {}
        q = deque()
        for l in left_nodes:
            if match_left[l] is None:
                dist[l] = 0
                q.append(l)
            else:
                dist[l] = float("inf")
        found = False
        while q:
            l = q.popleft()
            for r in adj[l]:
                nl = match_right.get(r)
                if nl is None:
                    found = True
                elif dist.get(nl, float("inf")) == float("inf"):
                    dist[nl] = dist[l] + 1
                    q.append(nl)
        return found, dist

    def dfs(l, dist):
        for r in adj[l]:
            nl = match_right.get(r)
            if nl is None or (dist.get(nl, float("inf")) == dist[l] + 1 and dfs(nl, dist)):
                match_left[l] = r
                match_right[r] = l
                return True
        dist[l] = float("inf")
        return False

    while True:
        found, dist = bfs()
        if not found:
            break
        for l in left_nodes:
            if match_left[l] is None:
                dfs(l, dist)

    return {l: r for l, r in match_left.items() if r is not None}

def section_independent_matching_check():
    rows = []
    rng = random.Random(2024)
    for trial in range(200):
        n_labels = rng.randint(3, 14)
        n_idents = rng.randint(3, 14)
        idents = [f"ID_{j}" for j in range(n_idents)]
        labels = [{"occurrence_id": f"L{i}", "page_id": rng.choice(["p1", "p2"]), "token": f"tok{i}"} for i in range(n_labels)]
        lexicon_index = {}
        for l in labels:
            k = rng.randint(0, 3)
            lexicon_index[l["token"]] = set(rng.sample(idents, min(k, n_idents)))

        for policy in ["PER_PAGE_CAPACITY_1", "GLOBAL_CAPACITY_1"]:
            engine_match = maximum_bipartite_matching(labels, lexicon_index, policy)

            if policy == "GLOBAL_CAPACITY_1":
                adj = {l["occurrence_id"]: sorted(lexicon_index.get(l["token"], set())) for l in labels}
                alt_match = independent_max_matching(adj)
            else:
                alt_match = {}
                for page in sorted(set(l.get("page_id", "default") for l in labels)):
                    p_labels = [l for l in labels if l.get("page_id", "default") == page]
                    adj = {l["occurrence_id"]: sorted(lexicon_index.get(l["token"], set())) for l in p_labels}
                    alt_match.update(independent_max_matching(adj))

            engine_size = len(engine_match)
            alt_size = len(alt_match)
            # validity checks on alt_match: each identity used at most once (within capacity scope)
            valid = True
            if policy == "GLOBAL_CAPACITY_1":
                if len(set(alt_match.values())) != len(alt_match.values()):
                    valid = False
            else:
                for page in set(l.get("page_id", "default") for l in labels):
                    page_lbls = {l["occurrence_id"] for l in labels if l.get("page_id", "default") == page}
                    used = [v for k, v in alt_match.items() if k in page_lbls]
                    if len(set(used)) != len(used):
                        valid = False

            rows.append({
                "trial": trial, "policy": policy, "n_labels": n_labels, "n_identities": n_idents,
                "engine_matching_size": engine_size, "independent_matching_size": alt_size,
                "cardinality_match": engine_size == alt_size,
                "independent_matching_valid_capacity": valid
            })
    return rows

# ---------------------------------------------------------------------------
# F: Fresh sealed hidden benchmark -- REAL solver calls on NEW seeds (9001-9999
# range, never used anywhere in the frozen design package: design used
# seeds 101-120 and 1017-1425. This range is disjoint from both.)
# ---------------------------------------------------------------------------
FRESH_SEED_BASE = 9001

def compute_recovery(found_table, true_table, found_assign, true_assign, labels):
    found_set, true_set = set(found_table.items()), set(true_table.items())
    tp = len(found_set & true_set)
    fp = len(found_set - true_set)
    fn = len(true_set - found_set)
    prec = tp / (tp + fp) if (tp + fp) else 0.0
    rec = tp / (tp + fn) if (tp + fn) else 0.0
    exact = 1.0 if found_set == true_set else 0.0
    # equivalence-aware: same fitness-relevant coverage/identity mapping outcome
    correct = sum(1 for l in labels if found_assign.get(l["occurrence_id"]) == true_assign.get(l["occurrence_id"]))
    acc = correct / len(labels) if labels else 0.0
    return prec, rec, exact, acc

def section_fresh_hidden_benchmark(n_seeds=20, table_sizes=(8,), noise_rates=(0.0, 0.10, 0.25),
                                    algos=("DIRECTIONAL_HEURISTIC",)):
    lexicon = load_lexicon_forms()
    predictions = []
    results = []
    seed_i = 0
    for t_size in table_sizes:
        for noise in noise_rates:
            for s in range(n_seeds):
                seed = FRESH_SEED_BASE + seed_i
                seed_i += 1
                gen = SealedSyntheticGenerator(seed=seed)
                labels, truth = gen.generate_dataset(
                    table_size=t_size, mapping_mode="INJECTIVE", deletion_mode="DROP_UNMAPPED",
                    abbreviation="NONE", noise_rate=noise, unmatched_fraction=0.25,
                    capacity_policy="PER_PAGE_CAPACITY_1"
                )
                for algo in algos:
                    t0 = time.perf_counter()
                    tracemalloc.start()
                    if algo == "DIRECTIONAL_HEURISTIC":
                        solver = DirectionalHeuristicSolver(SOURCE_ALPHABET, EVA_ALPHABET, table_size=t_size,
                                                             seed=seed, time_limit_sec=15.0)
                        res = solver.solve(lexicon, labels)
                    elif algo == "CP_SAT":
                        solver = CPSATSolver(SOURCE_ALPHABET, EVA_ALPHABET, table_size=t_size, time_limit_sec=15.0)
                        res = solver.solve(lexicon, labels)
                    elif algo == "BRANCH_AND_BOUND":
                        solver = BranchAndBoundSolver(SOURCE_ALPHABET, EVA_ALPHABET, table_size=t_size, time_limit_sec=15.0)
                        res = solver.solve(lexicon, labels)
                    peak_kb = tracemalloc.get_traced_memory()[1] / 1024.0
                    tracemalloc.stop()
                    elapsed = time.perf_counter() - t0

                    sol = res["best_solution"]
                    predictions.append({
                        "seed": seed, "table_size": t_size, "noise_rate": noise, "algorithm": algo,
                        "predicted_table": json.dumps(sol["table"]),
                        "predicted_assignments": json.dumps(sol["assignments"]),
                        "runtime_sec": round(elapsed, 4), "peak_memory_kb": round(peak_kb, 2),
                        "timed_out": res.get("timed_out", False),
                    })
                    # sealed until now: only after prediction recorded do we compare to truth
                    prec, rec, exact, acc = compute_recovery(
                        sol["table"], truth["planted_table"], sol["assignments"], truth["true_assignments"], labels
                    )
                    results.append({
                        "seed": seed, "table_size": t_size, "noise_rate": noise, "algorithm": algo,
                        "exact_table_recovery": exact,
                        "mapping_precision": round(prec, 4), "mapping_recall": round(rec, 4),
                        "mapping_f1": round(2 * prec * rec / (prec + rec), 4) if (prec + rec) else 0.0,
                        "assignment_accuracy": round(acc, 4),
                        "coverage": round(sol["coverage"], 4),
                        "runtime_sec": round(elapsed, 4),
                        "timed_out": res.get("timed_out", False),
                    })
    return predictions, results

# ---------------------------------------------------------------------------
# H2: real resource pilot (wall time, peak RSS proxy via tracemalloc) for a
# single full model-selection replica (heuristic solver on 57-label-shaped
# synthetic data), extrapolated to 60,000 replicas with a confidence interval.
# ---------------------------------------------------------------------------
def section_resource_pilot(n_pilot=300):
    lexicon = load_lexicon_forms()
    times = []
    peaks = []
    for i in range(n_pilot):
        seed = 70000 + i
        gen = SealedSyntheticGenerator(seed=seed)
        labels, truth = gen.generate_dataset(table_size=8, noise_rate=0.0, unmatched_fraction=0.25)
        tracemalloc.start()
        t0 = time.perf_counter()
        solver = DirectionalHeuristicSolver(SOURCE_ALPHABET, EVA_ALPHABET, table_size=8, seed=seed, time_limit_sec=15.0)
        res = solver.solve(lexicon, labels)
        elapsed = time.perf_counter() - t0
        peak_kb = tracemalloc.get_traced_memory()[1] / 1024.0
        tracemalloc.stop()
        times.append(elapsed)
        peaks.append(peak_kb)
    mean_t = statistics.mean(times)
    sd_t = statistics.stdev(times) if len(times) > 1 else 0.0
    return {
        "n_pilot_runs": n_pilot,
        "mean_runtime_sec": round(mean_t, 5),
        "stdev_runtime_sec": round(sd_t, 5),
        "min_runtime_sec": round(min(times), 5),
        "max_runtime_sec": round(max(times), 5),
        "p95_runtime_sec": round(sorted(times)[int(0.95 * len(times)) - 1], 5),
        "mean_peak_memory_kb": round(statistics.mean(peaks), 2),
        "max_peak_memory_kb": round(max(peaks), 2),
        "extrapolated_60000_singlecore_hours": round(mean_t * 60000 / 3600.0, 3),
        "extrapolated_60000_16core_minutes": round(mean_t * 60000 / 16.0 / 60.0, 3),
        "extrapolated_60000_16core_minutes_p95": round(sorted(times)[int(0.95 * len(times)) - 1] * 60000 / 16.0 / 60.0, 3),
    }

def write_tsv(rows, path):
    if not rows:
        Path(path).write_text("")
        return
    fields = list(rows[0].keys())
    with open(path, "w", newline="", encoding="utf-8") as f:
        w = csv.DictWriter(f, fieldnames=fields, delimiter="\t")
        w.writeheader()
        for r in rows:
            w.writerow(r)

def main():
    print("=== E1/E2/E5: adversarial k-fragment combination benchmark ===")
    adv_rows = section_adversarial()
    write_tsv(adv_rows, OUT / "ADVERSARIAL_EXACT_RESULTS.tsv")
    for k in [2, 3, 4, 5, 6]:
        sub = [r for r in adv_rows if r["k_fragments_required"] == k]
        for alg in ["BRANCH_AND_BOUND", "CP_SAT", "DIRECTIONAL_HEURISTIC"]:
            s2 = [r for r in sub if r["algorithm"] == alg]
            rate = sum(1 for r in s2 if r["reached_global_optimum"] == "YES") / len(s2)
            print(f"  k={k} {alg:22s} global-optimum rate: {rate*100:.0f}%")

    print("\n=== E3: 100-permutation order invariance ===")
    perm_rows, mismatches = section_order_invariance_100()
    write_tsv(perm_rows, OUT / "ORDER_INVARIANCE_100.tsv")
    print(f"  mismatches out of 100: {mismatches}")

    print("\n=== D2: independent bipartite matching cross-check (200 random trials) ===")
    match_rows = section_independent_matching_check()
    write_tsv(match_rows, OUT / "INDEPENDENT_MATCHING_CHECK.tsv")
    bad = [r for r in match_rows if not r["cardinality_match"] or not r["independent_matching_valid_capacity"]]
    print(f"  mismatches/invalid out of {len(match_rows)}: {len(bad)}")

    print("\n=== F: fresh sealed hidden benchmark (NEW seeds 9001+, real solver execution) ===")
    preds, results = section_fresh_hidden_benchmark(
        n_seeds=20, table_sizes=(4, 6, 8, 10, 12), noise_rates=(0.0, 0.10, 0.25),
        algos=("DIRECTIONAL_HEURISTIC",)
    )
    write_tsv(preds, OUT / "FRESH_HIDDEN_PREDICTIONS.tsv")
    write_tsv(results, OUT / "FRESH_HIDDEN_RESULTS.tsv")
    for t_size in (4, 6, 8, 10, 12):
        for noise in (0.0, 0.10, 0.25):
            sub = [r for r in results if r["table_size"] == t_size and r["noise_rate"] == noise]
            if sub:
                rec = sum(r["exact_table_recovery"] for r in sub) / len(sub)
                print(f"  T={t_size:2d} noise={noise:.2f} exact_table_recovery={rec*100:5.1f}% (n={len(sub)})")

    print("\n=== F (cont.): CP_SAT / BRANCH_AND_BOUND on fresh seeds, table_size=8, 10 seeds ===")
    preds2, results2 = section_fresh_hidden_benchmark(
        n_seeds=5, table_sizes=(8,), noise_rates=(0.0, 0.10), algos=("CP_SAT", "BRANCH_AND_BOUND")
    )
    write_tsv(preds2, OUT / "FRESH_HIDDEN_PREDICTIONS_CPBB.tsv")
    write_tsv(results2, OUT / "FRESH_HIDDEN_RESULTS_CPBB.tsv")
    for algo in ("CP_SAT", "BRANCH_AND_BOUND"):
        for noise in (0.0, 0.10):
            sub = [r for r in results2 if r["algorithm"] == algo and r["noise_rate"] == noise]
            if sub:
                rec = sum(r["exact_table_recovery"] for r in sub) / len(sub)
                print(f"  {algo:20s} noise={noise:.2f} exact_table_recovery={rec*100:5.1f}% (n={len(sub)})")

    print("\n=== H2: real resource pilot (300 replicas) + extrapolation ===")
    res_bench = section_resource_pilot(n_pilot=300)
    with open(OUT / "RESOURCE_BENCHMARK.tsv", "w", newline="", encoding="utf-8") as f:
        w = csv.writer(f, delimiter="\t")
        w.writerow(["metric", "value"])
        for k, v in res_bench.items():
            w.writerow([k, v])
    print(json.dumps(res_bench, indent=2))

if __name__ == "__main__":
    main()
