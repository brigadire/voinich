#!/usr/bin/env python3
"""
New discriminative benchmark (clean_room.md Section 12). Replaces v3_design's
EXACT_SMALL_INSTANCE_RESULTS.tsv, which the audit found non-discriminative because every
instance had 2-10 tied co-optimal tables and a random baseline reached the "global optimum"
on 25/25 instances (Finding F004; COMPONENT_REUSE_REGISTRY.tsv: REJECTED).

Each case here is hand-designed (not randomly sampled) so it has either a UNIQUE optimum or an
explicitly enumerated equivalence class, and is deliberately built to require the property named
in its `targets` field. Every case is small enough (table_size <= 6, source alphabet <= 8, target
alphabet <= 7) that full brute-force enumeration of the table space is tractable, which is what
makes `truth_rank` and `certified_optimum` genuinely computable here rather than asserted.
"""
import hashlib
import itertools
import random
import sys
import csv
from pathlib import Path
from typing import Dict, List, Set, Tuple

sys.path.insert(0, str(Path(__file__).parent))
from scorer import Scorer, is_valid_table
from oracle_bb import BranchAndBoundOracle
from solver_cpsat import CPSATSolver, SUPPORTED_DELETION_MODES, SUPPORTED_ABBREVIATION
from solver_heuristic import HeuristicSolver
from hybrid import HybridSolver


def stable_seed(s: str) -> int:
    """Deterministic across processes/runs, unlike Python's built-in hash() on strings, which is
    salted per-process by PYTHONHASHSEED (a security feature, not a bug) -- using hash()/`.
    __hash__()` for a reproducible seed was a real bug caught by SCIENTIFIC_SAFETY_TESTS.md's
    reproducibility check: re-running this script produced different RANDOM_BASELINE and (silently)
    different HEURISTIC/HYBRID master seeds on every invocation."""
    return int(hashlib.sha256(s.encode()).hexdigest(), 16) & 0xFFFF


def enumerate_all_tables(source_alphabet, target_alphabet, table_size, mapping_mode):
    for combo in itertools.combinations(sorted(source_alphabet), table_size):
        for assignment in itertools.product(sorted(target_alphabet), repeat=table_size):
            table = dict(zip(combo, assignment))
            if is_valid_table(table, mapping_mode):
                yield table


def full_ranking(lexicon, labels, source_alphabet, target_alphabet, table_size, mapping_mode, deletion_mode, abbreviation, capacity_policy):
    scorer = Scorer(mapping_mode, deletion_mode, abbreviation, capacity_policy)
    scored = []
    for table in enumerate_all_tables(source_alphabet, target_alphabet, table_size, mapping_mode):
        fit = scorer.evaluate(table, lexicon, labels)["fitness"]
        scored.append((fit, table))
    scored.sort(key=lambda x: -x[0])
    return scored


def rank_of_table(scored, table):
    """1-indexed rank of `table`'s score among all enumerated scores (ties share the best rank),
    or None if table not found (shouldn't happen for tables drawn from the same space)."""
    key = tuple(sorted(table.items()))
    target_fit = None
    for fit, t in scored:
        if tuple(sorted(t.items())) == key:
            target_fit = fit
            break
    if target_fit is None:
        return None, None
    better = sum(1 for fit, _ in scored if fit > target_fit)
    return better + 1, target_fit


def equivalence_class_size(scored, best_fit):
    return sum(1 for fit, _ in scored if fit == best_fit)


CASES = []


def register(case_id, targets, source_alphabet, target_alphabet, lexicon, labels, true_table,
             table_size, mapping_mode="INJECTIVE", deletion_mode="DROP_UNMAPPED", abbreviation="NONE",
             capacity_policy="PER_PAGE_CAPACITY_1", equivalence_note=""):
    CASES.append(dict(
        case_id=case_id, targets=targets, source_alphabet=source_alphabet, target_alphabet=target_alphabet,
        lexicon=lexicon, labels=labels, true_table=true_table, table_size=table_size, mapping_mode=mapping_mode,
        deletion_mode=deletion_mode, abbreviation=abbreviation, capacity_policy=capacity_policy,
        equivalence_note=equivalence_note,
    ))


# --- Case 1: unique-optimum baseline (positive control) ---
register(
    "C01_UNIQUE_OPTIMUM_BASELINE", ["unique_optimum"],
    list("abcd"), list("acdf"),
    {"ID0": {"ab"}, "ID1": {"cd"}, "ID2": {"ba"}, "ID3": {"dc"}},
    [
        {"occurrence_id": "O0", "page_id": "P0", "token": "ac"},
        {"occurrence_id": "O1", "page_id": "P0", "token": "df"},
        {"occurrence_id": "O2", "page_id": "P1", "token": "ca"},
        {"occurrence_id": "O3", "page_id": "P1", "token": "fd"},
    ],
    {"a": "a", "b": "c", "c": "d", "d": "f"}, table_size=4,
)

# --- Case 2: false frequency correspondence trap (truth is NOT the score optimum) ---
# 'e' is by far the most frequent source character (appears in every distractor identity's
# form), so mapping e->o (a very common EVA-manifest symbol) coincidentally covers many
# distractor labels even though it plays no role in the true generative table.
register(
    "C02_FALSE_FREQUENCY_TRAP", ["false_frequency_correspondence", "truth_not_optimum"],
    list("abcdefgh"), list("acdefho"),
    {
        "ID0": {"ab"}, "ID1": {"cd"},  # true, low-frequency-source identities
        "DECOY0": {"ee"}, "DECOY1": {"eee"}, "DECOY2": {"ee"}, "DECOY3": {"eee"},
        "DECOY4": {"ee"}, "DECOY5": {"eee"},
    },
    [
        {"occurrence_id": "O0", "page_id": "P0", "token": "ac"},
        {"occurrence_id": "O1", "page_id": "P0", "token": "cd"},
        {"occurrence_id": "O2", "page_id": "P0", "token": "oo"},
        {"occurrence_id": "O3", "page_id": "P0", "token": "ooo"},
        {"occurrence_id": "O4", "page_id": "P1", "token": "oo"},
        {"occurrence_id": "O5", "page_id": "P1", "token": "ooo"},
        {"occurrence_id": "O6", "page_id": "P1", "token": "oo"},
        {"occurrence_id": "O7", "page_id": "P1", "token": "ooo"},
    ],
    {"a": "a", "b": "c", "c": "d", "d": "f"}, table_size=4,
)

# --- Case 3: bounded merge ambiguity ---
register(
    "C03_MERGE_AMBIGUITY", ["merge_ambiguity"],
    list("abcd"), list("acd"),
    {"ID0": {"ab"}, "ID1": {"cb"}, "ID2": {"ad"}},
    [
        {"occurrence_id": "O0", "page_id": "P0", "token": "ac"},
        {"occurrence_id": "O1", "page_id": "P0", "token": "cc"},
        {"occurrence_id": "O2", "page_id": "P1", "token": "ad"},
    ],
    {"a": "a", "b": "c", "c": "c", "d": "d"}, table_size=4, mapping_mode="MERGE_1",
)

# --- Case 4: selective deletion (vowels dropped, consonants kept) ---
register(
    "C04_SELECTIVE_DELETION", ["selective_deletion"],
    list("abcdeiou"), list("acdf"),
    {"ID0": {"aebic"}, "ID1": {"oud"}},
    [
        {"occurrence_id": "O0", "page_id": "P0", "token": "ac"},
        {"occurrence_id": "O1", "page_id": "P0", "token": "d"},
    ],
    {"a": "a", "b": "c", "c": "c", "d": "d"}, table_size=4, deletion_mode="SELECTIVE_VOWEL_DROP",
)

# --- Case 5: abbreviation (suffix suspension) ---
register(
    "C05_ABBREVIATION_SUSPENSION", ["abbreviation"],
    list("abcde"), list("acdfh"),
    {"ID0": {"abcd"}, "ID1": {"bcde"}},
    [
        {"occurrence_id": "O0", "page_id": "P0", "token": "aca"},
        {"occurrence_id": "O1", "page_id": "P0", "token": "cad"},
    ],
    {"a": "a", "b": "c", "c": "a", "d": "d", "e": "f"}, table_size=5, abbreviation="SUSPENSION_1",
)

# --- Case 6: full composition (merge + selective deletion + abbreviation) ---
register(
    "C06_FULL_COMPOSITION", ["merge_ambiguity", "selective_deletion", "abbreviation", "composition"],
    list("abcdeiou"), list("acd"),
    {"ID0": {"aebicd"}, "ID1": {"oudac"}},
    [
        {"occurrence_id": "O0", "page_id": "P0", "token": "ac"},
        {"occurrence_id": "O1", "page_id": "P0", "token": "da"},
    ],
    {"a": "a", "b": "c", "c": "c", "d": "d"}, table_size=4, mapping_mode="MERGE_1",
    deletion_mode="SELECTIVE_VOWEL_DROP", abbreviation="SUSPENSION_1",
)

# --- Case 7: high unmatched rate / distractor-dense (low signal) ---
_rng7 = random.Random(701)
_labels7 = [{"occurrence_id": "O0", "page_id": "P0", "token": "ac"}, {"occurrence_id": "O1", "page_id": "P1", "token": "cd"}]
for i in range(10):
    _labels7.append({"occurrence_id": f"D{i}", "page_id": "P0" if i % 2 == 0 else "P1", "token": "".join(_rng7.choice("efho") for _ in range(_rng7.randint(1, 4)))})
register(
    "C07_HIGH_UNMATCHED_DISTRACTOR_DENSE", ["high_unmatched_rate", "distractor_density_high"],
    list("abcd"), list("acdefho"),
    {"ID0": {"ab"}, "ID1": {"cd"}},
    _labels7,
    {"a": "a", "b": "c", "c": "c", "d": "d"}, table_size=4,
)

# --- Case 8: low distractor density (contrast with C07) ---
register(
    "C08_LOW_DISTRACTOR_DENSITY", ["distractor_density_low"],
    list("abcd"), list("acdf"),
    {"ID0": {"ab"}, "ID1": {"cd"}},
    [
        {"occurrence_id": "O0", "page_id": "P0", "token": "ac"},
        {"occurrence_id": "O1", "page_id": "P1", "token": "cd"},
        {"occurrence_id": "D0", "page_id": "P0", "token": "dd"},
    ],
    {"a": "a", "b": "c", "c": "c", "d": "d"}, table_size=4,
)

# --- Case 9: near-equal-but-not-equal competing solutions ---
register(
    "C09_NEAR_TIE_NOT_EQUAL", ["near_equal_not_equal"],
    list("abcdef"), list("acdf"),
    {"ID0": {"ab"}, "ID1": {"cd"}, "ID2": {"ef"}, "ID3": {"ba"}},
    [
        {"occurrence_id": "O0", "page_id": "P0", "token": "ac"},
        {"occurrence_id": "O1", "page_id": "P0", "token": "cd"},
        {"occurrence_id": "O2", "page_id": "P1", "token": "df"},
    ],
    {"a": "a", "b": "c", "c": "c", "d": "d", "e": "d", "f": "f"}, table_size=5, mapping_mode="MERGE_1",
)

# --- Case 10: explicit equivalence class (multiple distinct tables, identical optimal score) ---
register(
    "C10_EXPLICIT_EQUIVALENCE_CLASS", ["equivalence_class"],
    list("abcd"), list("acdfhi"),
    {"ID0": {"ab"}, "ID1": {"cd"}},
    [
        {"occurrence_id": "O0", "page_id": "P0", "token": "ac"},
        {"occurrence_id": "O1", "page_id": "P1", "token": "df"},
    ],
    {"a": "a", "b": "c", "c": "d", "d": "f"}, table_size=4,
    equivalence_note="Unused source chars c,d and unused target symbols h,i are interchangeable "
                      "under any table achieving the same 2 matches at the same complexity; the "
                      "true generative table is one canonical representative of a >1-member "
                      "equivalence class by construction (any relabeling of the *unused* mapping "
                      "slots ties it).",
)


def run_case(case: dict) -> List[dict]:
    lexicon = {k: set(v) for k, v in case["lexicon"].items()}
    labels = case["labels"]
    source_alphabet, target_alphabet = case["source_alphabet"], case["target_alphabet"]
    table_size, mapping_mode = case["table_size"], case["mapping_mode"]
    deletion_mode, abbreviation, capacity_policy = case["deletion_mode"], case["abbreviation"], case["capacity_policy"]

    scored = full_ranking(lexicon, labels, source_alphabet, target_alphabet, table_size, mapping_mode, deletion_mode, abbreviation, capacity_policy)
    best_fit = scored[0][0]
    eq_size = equivalence_class_size(scored, best_fit)
    truth_rank, truth_fit = rank_of_table(scored, case["true_table"])
    truth_is_optimum = (truth_rank == 1)

    rows = []

    def record(algorithm, fitness, matched, certified, timed_out, gap, extra=""):
        rank, _ = (1, fitness) if fitness == best_fit else (None, None)
        rows.append({
            "case_id": case["case_id"], "targets": ";".join(case["targets"]), "algorithm": algorithm,
            "table_size": table_size, "mapping_mode": mapping_mode, "deletion_mode": deletion_mode,
            "abbreviation": abbreviation, "capacity_policy": capacity_policy,
            "certified_optimum_score": best_fit, "equivalence_class_size": eq_size,
            "truth_score": truth_fit, "truth_rank": truth_rank, "truth_is_optimum": truth_is_optimum,
            "algo_fitness": fitness, "algo_matched": matched,
            "reached_optimum_score": fitness == best_fit,
            "certified": certified, "timed_out": timed_out, "optimality_gap": gap,
            "notes": extra,
        })

    bb = BranchAndBoundOracle(source_alphabet, target_alphabet, table_size, mapping_mode, deletion_mode, abbreviation, capacity_policy, time_limit_sec=20.0)
    rb = bb.solve(lexicon, labels)
    record("BRANCH_AND_BOUND", rb["best_fitness"], rb["best_solution"]["matched"], rb["global_optimum_certified"], rb["timed_out"], 0 if rb["global_optimum_certified"] else None)

    cpsat_ok = all(len(c) == 1 for c in source_alphabet) and deletion_mode in SUPPORTED_DELETION_MODES and abbreviation in SUPPORTED_ABBREVIATION
    if cpsat_ok:
        cps = CPSATSolver(source_alphabet, target_alphabet, table_size, mapping_mode, capacity_policy, time_limit_sec=20.0)
        rc = cps.solve(lexicon, labels, deletion_mode, abbreviation)
        gap = rc["optimality_gap"] if rc["is_feasible_only"] else (0 if rc["is_optimal"] else None)
        record("CP_SAT", rc["fitness"], rc["matched"], rc["is_optimal"], rc["timed_out"], gap, rc["status"])
    else:
        record("CP_SAT", None, None, False, False, None, "OUT_OF_SCOPE_deletion_mode_or_alphabet")

    heur = HeuristicSolver(source_alphabet, target_alphabet, table_size, mapping_mode, deletion_mode, abbreviation, capacity_policy, n_starts=8, max_iters=300, master_seed=stable_seed(case["case_id"]))
    rh = heur.solve(lexicon, labels)
    record("DIRECTED_HEURISTIC", rh["fitness"], rh["matched"], False, False, None)

    hy = HybridSolver(source_alphabet, target_alphabet, table_size, mapping_mode, deletion_mode, abbreviation, capacity_policy, master_seed=stable_seed(case["case_id"]))
    ry = hy.solve(lexicon, labels)
    record("HYBRID", ry["final_fitness"], None, ry["classification"] == "GLOBAL_OPTIMUM_CERTIFIED", ry["classification"] == "VERIFICATION_TIMEOUT", ry.get("optimality_gap"), ry["classification"])

    # Random baseline = distribution over 30 INDEPENDENT SINGLE draws (mean/hit-rate), not the
    # best of 30 -- "best of N" is a search procedure in disguise and biases the baseline toward
    # the ceiling on small search spaces (the same nearest-of-N-vs-single-draw bias documented in
    # this project's task65 null-model work). A discriminative benchmark that lets the "random"
    # baseline search must not then call it random.
    rng = random.Random(1000 + stable_seed(case["case_id"]))
    scorer = Scorer(mapping_mode, deletion_mode, abbreviation, capacity_policy)
    random_fits = []
    for _ in range(30):
        srcs = rng.sample(source_alphabet, table_size)
        table = {}
        for s in srcs:
            cands = list(target_alphabet)
            rng.shuffle(cands)
            for t in cands:
                trial = dict(table)
                trial[s] = t
                if is_valid_table(trial, mapping_mode):
                    table = trial
                    break
        random_fits.append(scorer.evaluate(table, lexicon, labels)["fitness"])
    random_mean = sum(random_fits) / len(random_fits)
    random_hit_rate = sum(1 for f in random_fits if f == best_fit) / len(random_fits)
    record(
        "RANDOM_BASELINE_SINGLE_DRAW_MEAN_OF_30", random_mean, None, random_hit_rate >= 0.5, False, None,
        f"hit_rate_over_30_independent_single_draws={random_hit_rate:.3f}",
    )

    return rows


def main():
    out_path = sys.argv[1] if len(sys.argv) > 1 else "DISCRIMINATIVE_BENCHMARK_RESULTS.tsv"
    all_rows = []
    for case in CASES:
        all_rows.extend(run_case(case))
    with open(out_path, "w", newline="") as f:
        w = csv.DictWriter(f, fieldnames=list(all_rows[0].keys()), delimiter="\t")
        w.writeheader()
        w.writerows(all_rows)

    n_cases = len(CASES)
    random_hits = sum(1 for r in all_rows if r["algorithm"] == "RANDOM_BASELINE_SINGLE_DRAW_MEAN_OF_30" and r["reached_optimum_score"])
    bb_hits = sum(1 for r in all_rows if r["algorithm"] == "BRANCH_AND_BOUND" and r["reached_optimum_score"])
    print(f"cases={n_cases} random_baseline_optimum_rate={random_hits/n_cases:.3f} bb_optimum_rate={bb_hits/n_cases:.3f}")


if __name__ == "__main__":
    main()
