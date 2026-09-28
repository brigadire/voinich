#!/usr/bin/env python3
"""
Completes the two sections of audit_verification.py that were cut short by a wall-clock
budget cap (CP_SAT/BB on the real lexicon has no bound on its candidate-generation phase --
this is itself Finding F014). Reduced to 2 seeds / 8s time limit for tractability; the
adversarial synthetic k-fragment test (already completed, ADVERSARIAL_EXACT_RESULTS.tsv)
remains the primary, conclusive evidence for CP_SAT incompleteness (F003).
"""
import sys, time, csv, json, tracemalloc, statistics
from pathlib import Path

AUDIT_DIR = Path(__file__).resolve().parent.parent
sys.path.insert(0, str(AUDIT_DIR / "repro_sandbox" / "scripts"))
from engine import SOURCE_ALPHABET, EVA_ALPHABET
from search_cp import CPSATSolver
from search_bb import BranchAndBoundSolver
from search_heuristic import DirectionalHeuristicSolver
from synthetic_generator import SealedSyntheticGenerator, load_lexicon_forms

def compute_recovery(found_table, true_table, found_assign, true_assign, labels):
    found_set, true_set = set(found_table.items()), set(true_table.items())
    tp = len(found_set & true_set); fp = len(found_set - true_set); fn = len(true_set - found_set)
    prec = tp / (tp + fp) if (tp + fp) else 0.0
    rec = tp / (tp + fn) if (tp + fn) else 0.0
    exact = 1.0 if found_set == true_set else 0.0
    correct = sum(1 for l in labels if found_assign.get(l["occurrence_id"]) == true_assign.get(l["occurrence_id"]))
    acc = correct / len(labels) if labels else 0.0
    return prec, rec, exact, acc

def write_tsv(rows, path):
    if not rows:
        return
    with open(path, "w", newline="", encoding="utf-8") as f:
        w = csv.DictWriter(f, fieldnames=list(rows[0].keys()), delimiter="\t")
        w.writeheader()
        for r in rows: w.writerow(r)

def cpbb_section():
    lexicon = load_lexicon_forms()
    preds, results = [], []
    seed_i = 0
    FRESH_SEED_BASE = 9500
    for noise in (0.0, 0.10):
        for s in range(2):
            seed = FRESH_SEED_BASE + seed_i; seed_i += 1
            gen = SealedSyntheticGenerator(seed=seed)
            labels, truth = gen.generate_dataset(table_size=8, mapping_mode="INJECTIVE",
                                                  deletion_mode="DROP_UNMAPPED", abbreviation="NONE",
                                                  noise_rate=noise, unmatched_fraction=0.25,
                                                  capacity_policy="PER_PAGE_CAPACITY_1")
            for algo in ("CP_SAT", "BRANCH_AND_BOUND"):
                t0 = time.perf_counter(); tracemalloc.start()
                if algo == "CP_SAT":
                    solver = CPSATSolver(SOURCE_ALPHABET, EVA_ALPHABET, table_size=8, time_limit_sec=8.0)
                else:
                    solver = BranchAndBoundSolver(SOURCE_ALPHABET, EVA_ALPHABET, table_size=8, time_limit_sec=8.0)
                res = solver.solve(lexicon, labels)
                peak_kb = tracemalloc.get_traced_memory()[1] / 1024.0
                tracemalloc.stop()
                elapsed = time.perf_counter() - t0
                sol = res["best_solution"]
                preds.append({"seed": seed, "table_size": 8, "noise_rate": noise, "algorithm": algo,
                               "predicted_table": json.dumps(sol["table"]),
                               "predicted_assignments": json.dumps(sol["assignments"]),
                               "runtime_sec": round(elapsed, 4), "peak_memory_kb": round(peak_kb, 2),
                               "timed_out": res.get("timed_out", False)})
                prec, rec, exact, acc = compute_recovery(sol["table"], truth["planted_table"], sol["assignments"], truth["true_assignments"], labels)
                results.append({"seed": seed, "table_size": 8, "noise_rate": noise, "algorithm": algo,
                                 "exact_table_recovery": exact, "mapping_precision": round(prec, 4),
                                 "mapping_recall": round(rec, 4), "assignment_accuracy": round(acc, 4),
                                 "coverage": round(sol["coverage"], 4), "runtime_sec": round(elapsed, 4),
                                 "timed_out": res.get("timed_out", False)})
                print(f"  seed={seed} noise={noise} {algo:20s} elapsed={elapsed:.2f}s timed_out={res.get('timed_out')} exact_recovery={exact}", flush=True)
    write_tsv(preds, AUDIT_DIR / "FRESH_HIDDEN_PREDICTIONS_CPBB.tsv")
    write_tsv(results, AUDIT_DIR / "FRESH_HIDDEN_RESULTS_CPBB.tsv")

def resource_pilot(n_pilot=300):
    lexicon = load_lexicon_forms()
    times, peaks = [], []
    for i in range(n_pilot):
        seed = 70000 + i
        gen = SealedSyntheticGenerator(seed=seed)
        labels, truth = gen.generate_dataset(table_size=8, noise_rate=0.0, unmatched_fraction=0.25)
        tracemalloc.start(); t0 = time.perf_counter()
        solver = DirectionalHeuristicSolver(SOURCE_ALPHABET, EVA_ALPHABET, table_size=8, seed=seed, time_limit_sec=15.0)
        solver.solve(lexicon, labels)
        elapsed = time.perf_counter() - t0
        peak_kb = tracemalloc.get_traced_memory()[1] / 1024.0
        tracemalloc.stop()
        times.append(elapsed); peaks.append(peak_kb)
        if (i + 1) % 50 == 0:
            print(f"  pilot progress: {i+1}/{n_pilot}", flush=True)
    mean_t = statistics.mean(times); sd_t = statistics.stdev(times) if len(times) > 1 else 0.0
    res = {
        "n_pilot_runs": n_pilot, "mean_runtime_sec": round(mean_t, 5), "stdev_runtime_sec": round(sd_t, 5),
        "min_runtime_sec": round(min(times), 5), "max_runtime_sec": round(max(times), 5),
        "p95_runtime_sec": round(sorted(times)[int(0.95 * len(times)) - 1], 5),
        "mean_peak_memory_kb": round(statistics.mean(peaks), 2), "max_peak_memory_kb": round(max(peaks), 2),
        "extrapolated_60000_singlecore_hours": round(mean_t * 60000 / 3600.0, 3),
        "extrapolated_60000_16core_minutes": round(mean_t * 60000 / 16.0 / 60.0, 3),
        "extrapolated_60000_16core_minutes_p95": round(sorted(times)[int(0.95 * len(times)) - 1] * 60000 / 16.0 / 60.0, 3),
    }
    with open(AUDIT_DIR / "RESOURCE_BENCHMARK.tsv", "w", newline="", encoding="utf-8") as f:
        w = csv.writer(f, delimiter="\t"); w.writerow(["metric", "value"])
        for k, v in res.items(): w.writerow([k, v])
    print(json.dumps(res, indent=2))

if __name__ == "__main__":
    print("=== CP_SAT / BRANCH_AND_BOUND on fresh seeds (reduced: 2 seeds x 2 noise, 8s cap) ===", flush=True)
    cpbb_section()
    print("=== resource pilot (300 replicas) ===", flush=True)
    resource_pilot(300)
