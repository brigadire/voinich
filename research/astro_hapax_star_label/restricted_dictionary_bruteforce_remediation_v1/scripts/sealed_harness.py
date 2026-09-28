#!/usr/bin/env python3
"""
Shared harness for the sealed benchmark (dev and hidden phases both call this; see
SEALED_BENCHMARK_PROTOCOL.md for the process-separation discipline around which script is
allowed to see truth when).
"""
import random
from typing import Dict, List, Tuple
from generator import GeneratorConfig, generate, split_public_private
from scorer import Scorer
from hybrid import HybridSolver
from oracle_bb import BranchAndBoundOracle
from solver_cpsat import CPSATSolver, SUPPORTED_DELETION_MODES, SUPPORTED_ABBREVIATION

SOURCE_ALPHABET = None  # derived per-instance from generator.SOURCE_POOL length


def build_cell_matrix() -> List[dict]:
    """Returns the 36 cell specs (27 main grid + 8 mode variants + 1 capacity variant) per
    SEALED_BENCHMARK_PROTOCOL.md. Each cell gets N_SEEDS runs."""
    cells = []
    for table_size in (4, 8, 12):
        for noise in (0.0, 0.10, 0.25):
            for unmatched in (0.0, 0.25, 0.50):
                cells.append(dict(
                    cell_id=f"MAIN_ts{table_size}_n{int(noise*100)}_u{int(unmatched*100)}",
                    table_size=table_size, noise_rate=noise, unmatched_rate=unmatched,
                    mapping_mode="INJECTIVE", deletion_mode="DROP_UNMAPPED", abbreviation="NONE",
                    capacity_policy="PER_PAGE_CAPACITY_1",
                ))
    ref = dict(table_size=8, noise_rate=0.10, unmatched_rate=0.25)
    variants = [
        ("VARIANT_MERGE1", dict(mapping_mode="MERGE_1")),
        ("VARIANT_SELECTIVE_DELETION", dict(deletion_mode="SELECTIVE_VOWEL_DROP")),
        ("VARIANT_ABBREVIATION", dict(abbreviation="SUSPENSION_1")),
        ("VARIANT_SUBST_PLUS_ABBREV", dict(mapping_mode="MERGE_1", abbreviation="SUSPENSION_1")),
        ("VARIANT_SUBST_PLUS_DELETION", dict(mapping_mode="MERGE_1", deletion_mode="SELECTIVE_VOWEL_DROP")),
        ("VARIANT_FULL_COMPOSITION", dict(mapping_mode="MERGE_1", deletion_mode="SELECTIVE_VOWEL_DROP", abbreviation="SUSPENSION_1")),
        ("VARIANT_LOW_DISTRACTOR_DENSITY", dict(unmatched_rate=0.10)),
        ("VARIANT_HIGH_DISTRACTOR_DENSITY", dict(unmatched_rate=0.60)),
    ]
    for name, overrides in variants:
        cell = dict(cell_id=name, mapping_mode="INJECTIVE", deletion_mode="DROP_UNMAPPED", abbreviation="NONE", capacity_policy="PER_PAGE_CAPACITY_1")
        cell.update(ref)
        cell.update(overrides)
        cells.append(cell)

    cell = dict(cell_id="VARIANT_GLOBAL_CAPACITY", mapping_mode="INJECTIVE", deletion_mode="DROP_UNMAPPED", abbreviation="NONE", capacity_policy="GLOBAL_CAPACITY_1")
    cell.update(ref)
    cells.append(cell)
    return cells


def held_out_split(labels: List[dict], seed: int) -> Tuple[List[dict], List[dict]]:
    rng = random.Random(seed * 31 + 7)
    by_page: Dict[str, List[dict]] = {}
    for l in labels:
        by_page.setdefault(l["page_id"], []).append(l)
    train, held = [], []
    for page, page_labels in sorted(by_page.items()):
        shuffled = page_labels[:]
        rng.shuffle(shuffled)
        n_held = max(1, round(0.2 * len(shuffled)))
        held.extend(shuffled[:n_held])
        train.extend(shuffled[n_held:])
    return train, held


#: n_identities=16 matches the larger synthetic page (16) so that PER_PAGE_CAPACITY_1 at
#: unmatched_rate=0.0 is actually capacity-feasible (see generator.py's capacity-downgrade fix --
#: an earlier n_identities=8/12 over 16-30-slot pages made "0% unmatched" structurally impossible
#: to realize without violating capacity, silently inflating the recall/precision denominator with
#: unsatisfiable ground truth; REMEDIATION_REPORT.md has the full account).
N_IDENTITIES = 16
N_OCCURRENCES = 30
PAGE_SPLIT = (16, 14)


def build_generator_config(cell: dict, seed: int) -> GeneratorConfig:
    """The ONE place a GeneratorConfig is built for the sealed dev/hidden pipeline -- both
    run_dev_grid.py (via run_one_instance) and generate_hidden.py call this, so dev and hidden
    instances can never again silently diverge in scale (they previously did: generate_hidden.py
    had its own separate GeneratorConfig call using n_identities=12 and the default
    n_occurrences=57/page_split=(30,27), while this module used 8/30/(16,14).)"""
    return GeneratorConfig(
        seed=seed, n_identities=N_IDENTITIES, table_size=cell["table_size"], mapping_mode=cell["mapping_mode"],
        deletion_mode=cell["deletion_mode"], abbreviation=cell["abbreviation"], capacity_policy=cell["capacity_policy"],
        noise_rate=cell["noise_rate"], unmatched_rate=cell["unmatched_rate"],
        n_occurrences=N_OCCURRENCES, page_split=PAGE_SPLIT,
    )


def run_one_instance(cell: dict, seed: int) -> dict:
    cfg = build_generator_config(cell, seed)
    instance = generate(cfg)
    public, private = split_public_private(instance)
    lexicon = {k: set(v) for k, v in public["lexicon"].items()}
    all_labels_public = public["labels"]

    train_labels, held_labels_public = held_out_split(all_labels_public, seed)
    truth_by_occ = private["true_identity_by_occurrence"]

    source_alphabet = sorted(set(c for forms in lexicon.values() for f in forms for c in f))
    target_alphabet = public["target_alphabet"]

    ts = cell["table_size"]
    # Budget fixed at 60s uniformly (not scaled by table_size) based on BUDGET_SCALING_AUDIT.md's
    # dev-only finding: certification rate and unconditional recovery both plateau by 60s for the
    # exact-CP-SAT-scope portion of this model class, with no further gain at 300s/1200s (CP-SAT
    # exits early on proof of optimality; it does not "use" extra allotted time once actually
    # done). The original 4s/8s/12s budgets were shown, on dev data, to be a genuine bottleneck
    # (BUDGET_SCALING_AUDIT.md), so this is dev-informed tuning fixed BEFORE the hidden run per
    # SEALED_BENCHMARK_PROTOCOL.md steps 3-4 -- not a threshold chosen after seeing hidden results.
    verification_time_limit_sec = 60.0
    hy = HybridSolver(
        source_alphabet, target_alphabet, cell["table_size"], cell["mapping_mode"], cell["deletion_mode"],
        cell["abbreviation"], cell["capacity_policy"], master_seed=seed,
        neighborhood_radius=1, neighborhood_max_rounds=1, verification_time_limit_sec=verification_time_limit_sec,
        heuristic_n_starts=5, heuristic_max_iters=200,
    )
    result = hy.solve(lexicon, train_labels)
    predicted_table = result["final_table"]

    scorer = Scorer(cell["mapping_mode"], cell["deletion_mode"], cell["abbreviation"], cell["capacity_policy"])
    held_labels_with_truth = [dict(l, true_identity=truth_by_occ[l["occurrence_id"]]) for l in held_labels_public]
    full_eval = scorer.evaluate(predicted_table, lexicon, all_labels_public, held_out_labels=held_labels_with_truth)

    true_table = private["true_table"]
    exact_table_recovery = predicted_table == true_table
    true_eval_full = scorer.evaluate(true_table, lexicon, all_labels_public)
    equivalence_aware_recovery = full_eval["fitness"] >= true_eval_full["fitness"]

    all_labels_with_truth = [dict(l, true_identity=truth_by_occ[l["occurrence_id"]]) for l in all_labels_public]
    pred_assign_full = scorer.evaluate(predicted_table, lexicon, all_labels_with_truth)["assignments"]
    tp = sum(1 for occ, ident in pred_assign_full.items() if truth_by_occ.get(occ) == ident)
    fp = sum(1 for occ, ident in pred_assign_full.items() if truth_by_occ.get(occ) != ident)
    n_true_matched = sum(1 for v in truth_by_occ.values() if v is not None)
    fn = n_true_matched - tp
    precision = tp / (tp + fp) if (tp + fp) else 0.0
    recall = tp / (tp + fn) if (tp + fn) else 0.0

    correct_assign = sum(1 for occ, ident in pred_assign_full.items() if truth_by_occ.get(occ) == ident)
    assignment_accuracy = correct_assign / len(all_labels_public)

    return {
        "cell_id": cell["cell_id"], "seed": seed, "table_size": cell["table_size"],
        "noise_rate": cell["noise_rate"], "unmatched_rate": cell["unmatched_rate"],
        "mapping_mode": cell["mapping_mode"], "deletion_mode": cell["deletion_mode"],
        "abbreviation": cell["abbreviation"], "capacity_policy": cell["capacity_policy"],
        "classification": result["classification"], "verifier_mode": result["verifier_mode"],
        "exact_table_recovery": exact_table_recovery,
        "equivalence_aware_table_recovery": equivalence_aware_recovery,
        "mapping_precision": precision, "mapping_recall": recall,
        "assignment_accuracy": assignment_accuracy,
        "held_out_assignment_accuracy": full_eval.get("held_out_assignment_accuracy"),
        "coverage": full_eval["coverage"], "matched": full_eval["matched"], "total": full_eval["total"],
        "fitness": full_eval["fitness"], "true_fitness": true_eval_full["fitness"],
    }
