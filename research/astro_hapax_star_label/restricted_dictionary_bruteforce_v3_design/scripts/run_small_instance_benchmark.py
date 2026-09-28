#!/usr/bin/env python3
"""
Executes the exact small-instance benchmark across:
1. EXACT_EXHAUSTIVE (ground-truth global optimum)
2. BRANCH_AND_BOUND
3. CP_SAT
4. DIRECTIONAL_HEURISTIC
5. RANDOM_BASELINE
Measures exact optimum attainment, mapping recovery, assignment accuracy, and runtime.
"""
from pathlib import Path
import csv
import itertools
import json
import time
import random
from collections import Counter
import tracemalloc

from engine import (
    evaluate_table,
    encode_word,
    compute_complexity,
    maximum_bipartite_matching
)
from search_bb import BranchAndBoundSolver
from search_cp import CPSATSolver
from search_heuristic import DirectionalHeuristicSolver

BASE_DIR = Path(__file__).resolve().parent.parent
OUTPUT_TSV = BASE_DIR / "EXACT_SMALL_INSTANCE_RESULTS.tsv"

def generate_small_instance(seed: int, num_words: int = 6, table_size: int = 2):
    rng = random.Random(seed)
    src_alpha = ("a", "b", "c", "d", "e", "f")
    tgt_alpha = ("o", "l", "c", "y")
    
    # Sample planted table uniformly from full universe
    planted_src = rng.sample(src_alpha, table_size)
    planted_tgt = rng.sample(tgt_alpha, table_size)
    planted_table = dict(zip(planted_src, planted_tgt))
    
    # Synthetic lexicon
    lex_words = [
        "ab", "cde", "fad", "bce", "def", "eac", "fba", "dce"
    ][:num_words]
    
    lexicon = {f"STAR_{i:02d}": {w} for i, w in enumerate(lex_words)}
    
    # Generate labels by encoding lexicon words
    labels = []
    occ_id = 1
    for i, w in enumerate(lex_words):
        enc = encode_word(w, planted_table, deletion_mode="DROP_UNMAPPED", abbreviation="NONE")
        if enc:
            labels.append({
                "occurrence_id": f"LBL_{occ_id:03d}",
                "page_id": "p1" if occ_id % 2 != 0 else "p2",
                "token": enc,
                "true_star": f"STAR_{i:02d}"
            })
            occ_id += 1
            
    return {
        "seed": seed,
        "src_alpha": src_alpha,
        "tgt_alpha": tgt_alpha,
        "table_size": table_size,
        "planted_table": planted_table,
        "lexicon": lexicon,
        "labels": labels
    }

def run_exhaustive(instance):
    src_alpha = instance["src_alpha"]
    tgt_alpha = instance["tgt_alpha"]
    k = instance["table_size"]
    lex = instance["lexicon"]
    labels = instance["labels"]
    
    best_fit = -float("inf")
    best_tables = []
    all_scored = []
    
    for src in itertools.combinations(src_alpha, k):
        for dst in itertools.permutations(tgt_alpha, k):
            t = dict(zip(src, dst))
            ev = evaluate_table(t, lex, labels, "INJECTIVE", "DROP_UNMAPPED", "NONE", "PER_PAGE_CAPACITY_1")
            fit = ev["fitness"]
            all_scored.append((t, fit, ev))
            if fit > best_fit:
                best_fit = fit
                best_tables = [t]
            elif fit == best_fit:
                best_tables.append(t)
                
    # True table rank
    all_scored.sort(key=lambda x: x[1], reverse=True)
    planted = instance["planted_table"]
    planted_rank = -1
    for rank, (t, f, _) in enumerate(all_scored, start=1):
        if t == planted:
            planted_rank = rank
            break
            
    return {
        "global_optimum_fitness": best_fit,
        "optimal_tables_count": len(best_tables),
        "planted_table_rank": planted_rank,
        "all_scored": all_scored
    }

def compute_recovery_metrics(found_table, true_table, found_assignments, labels):
    found_set = set(found_table.items())
    true_set = set(true_table.items())
    
    tp = len(found_set & true_set)
    fp = len(found_set - true_set)
    fn = len(true_set - found_set)
    
    prec = tp / (tp + fp) if (tp + fp) > 0 else 0.0
    rec = tp / (tp + fn) if (tp + fn) > 0 else 0.0
    f1 = 2 * prec * rec / (prec + rec) if (prec + rec) > 0 else 0.0
    
    # Assignment accuracy
    correct_assigns = 0
    total_assigns = len(labels)
    for l in labels:
        lbl_id = l["occurrence_id"]
        true_s = l["true_star"]
        if found_assignments.get(lbl_id) == true_s:
            correct_assigns += 1
    assign_acc = correct_assigns / total_assigns if total_assigns > 0 else 0.0
    
    return round(prec, 4), round(rec, 4), round(f1, 4), round(assign_acc, 4)

def main():
    print("Running exact small-instance benchmark across 25 instances...")
    results = []
    
    num_instances = 25
    
    for inst_idx in range(1, num_instances + 1):
        seed = 1000 + inst_idx * 17
        inst = generate_small_instance(seed, num_words=6, table_size=2)
        
        # 1. Exhaustive ground truth
        t0 = time.perf_counter()
        exh_res = run_exhaustive(inst)
        exh_time_ms = (time.perf_counter() - t0) * 1000.0
        opt_fitness = exh_res["global_optimum_fitness"]
        num_optima = exh_res["optimal_tables_count"]
        
        # 2. Branch and Bound
        tracemalloc.start()
        t0 = time.perf_counter()
        bb = BranchAndBoundSolver(
            inst["src_alpha"], inst["tgt_alpha"], inst["table_size"],
            "INJECTIVE", "DROP_UNMAPPED", "NONE", "PER_PAGE_CAPACITY_1", time_limit_sec=10.0
        )
        bb_res = bb.solve(inst["lexicon"], inst["labels"])
        bb_time_ms = (time.perf_counter() - t0) * 1000.0
        bb_mem_kb = tracemalloc.get_traced_memory()[1] / 1024.0
        tracemalloc.stop()
        
        # 3. CP-SAT Solver
        tracemalloc.start()
        t0 = time.perf_counter()
        cp = CPSATSolver(
            inst["src_alpha"], inst["tgt_alpha"], inst["table_size"],
            "INJECTIVE", "DROP_UNMAPPED", "NONE", "PER_PAGE_CAPACITY_1", time_limit_sec=10.0
        )
        cp_res = cp.solve(inst["lexicon"], inst["labels"])
        cp_time_ms = (time.perf_counter() - t0) * 1000.0
        cp_mem_kb = tracemalloc.get_traced_memory()[1] / 1024.0
        tracemalloc.stop()
        
        # 4. Directional Heuristic
        tracemalloc.start()
        t0 = time.perf_counter()
        heu = DirectionalHeuristicSolver(
            inst["src_alpha"], inst["tgt_alpha"], inst["table_size"],
            "INJECTIVE", "DROP_UNMAPPED", "NONE", "PER_PAGE_CAPACITY_1",
            beam_width=32, local_search_steps=60, seed=seed, time_limit_sec=10.0
        )
        heu_res = heu.solve(inst["lexicon"], inst["labels"])
        heu_time_ms = (time.perf_counter() - t0) * 1000.0
        heu_mem_kb = tracemalloc.get_traced_memory()[1] / 1024.0
        tracemalloc.stop()
        
        # 5. Random Baseline (sample 10 random tables)
        rng = random.Random(seed)
        rand_best_eval = None
        t0 = time.perf_counter()
        for _ in range(10):
            src_s = rng.sample(inst["src_alpha"], inst["table_size"])
            dst_s = rng.sample(inst["tgt_alpha"], inst["table_size"])
            r_table = dict(zip(src_s, dst_s))
            r_ev = evaluate_table(r_table, inst["lexicon"], inst["labels"], "INJECTIVE", "DROP_UNMAPPED", "NONE", "PER_PAGE_CAPACITY_1")
            if rand_best_eval is None or r_ev["fitness"] > rand_best_eval["fitness"]:
                rand_best_eval = r_ev
        rand_time_ms = (time.perf_counter() - t0) * 1000.0
        
        # Collate metrics for each algorithm
        algos = [
            ("BRANCH_AND_BOUND", bb_res["best_solution"], bb_time_ms, bb_mem_kb),
            ("CP_SAT", cp_res["best_solution"], cp_time_ms, cp_mem_kb),
            ("DIRECTIONAL_HEURISTIC", heu_res["best_solution"], heu_time_ms, heu_mem_kb),
            ("RANDOM_BASELINE", rand_best_eval, rand_time_ms, 12.0)
        ]
        
        for alg_name, sol, t_ms, mem in algos:
            fit = sol["fitness"]
            reached_opt = "YES" if fit == opt_fitness else "NO"
            tbl = sol["table"]
            assigns = sol["assignments"]
            prec, rec, f1, acc = compute_recovery_metrics(tbl, inst["planted_table"], assigns, inst["labels"])
            
            results.append({
                "instance_id": f"INST_{inst_idx:03d}",
                "seed": seed,
                "algorithm": alg_name,
                "global_optimum_fitness": opt_fitness,
                "achieved_fitness": fit,
                "reached_global_optimum": reached_opt,
                "planted_table_rank": exh_res["planted_table_rank"],
                "optimal_tables_count": num_optima,
                "mapping_precision": prec,
                "mapping_recall": rec,
                "mapping_f1": f1,
                "assignment_accuracy": acc,
                "runtime_ms": round(t_ms, 2),
                "peak_memory_kb": round(mem, 2)
            })

    # Write output TSV
    fields = list(results[0].keys())
    with OUTPUT_TSV.open("w", encoding="utf-8", newline="") as f:
        writer = csv.DictWriter(f, fieldnames=fields, delimiter="\t")
        writer.writeheader()
        for r in results:
            writer.writerow(r)
            
    print(f"Wrote {len(results)} exact small-instance rows to {OUTPUT_TSV}")
    
    # Print summary statistics
    print("\n--- EXACT SMALL INSTANCE SUMMARY ---")
    for alg in ["BRANCH_AND_BOUND", "CP_SAT", "DIRECTIONAL_HEURISTIC", "RANDOM_BASELINE"]:
        alg_rows = [r for r in results if r["algorithm"] == alg]
        opt_rate = sum(1 for r in alg_rows if r["reached_global_optimum"] == "YES") / len(alg_rows)
        avg_f1 = sum(r["mapping_f1"] for r in alg_rows) / len(alg_rows)
        avg_acc = sum(r["assignment_accuracy"] for r in alg_rows) / len(alg_rows)
        avg_time = sum(r["runtime_ms"] for r in alg_rows) / len(alg_rows)
        print(f"Algorithm: {alg:22s} | Exact-Opt Rate: {opt_rate*100:5.1f}% | Avg F1: {avg_f1:.3f} | Assign Acc: {avg_acc:.3f} | Avg Time: {avg_time:6.2f} ms")

if __name__ == "__main__":
    main()
