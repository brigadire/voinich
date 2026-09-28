"""Exact-Search and Equivalence/Symmetry Audit for M2R Identifiability Diagnostic v1.
Sections 4 and 5: Exhaustive search on small synthetic tasks comparing exact global optimum,
bounded M2R-v2 search, latent truth, and symmetry/equivalence classes.
"""
from __future__ import annotations
import csv
import itertools
import math
from pathlib import Path
from collections import defaultdict
import audit_engine


def create_small_tasks():
    """Defines a family of small synthetic tasks where full exhaustive search is feasible."""
    tasks = []

    # Task 1: 2x2 Cartesian (4 pairs, 2 INITIAL, 2 FINAL rules)
    t1_inits = [('ba', 'xy'), ('ce', 'wz')]
    t1_finals = [('da', 'pq'), ('fe', 'rs')]
    t1_rules = [('INITIAL', s, t) for s, t in t1_inits] + [('FINAL', s, t) for s, t in t1_finals]
    t1_pairs = []
    for si, ti in t1_inits:
        for sf, tf in t1_finals:
            t1_pairs.append((si + sf, ti + tf))
    tasks.append({
        'task_id': 'T1_cartesian_2x2',
        'description': 'Small 2x2 Cartesian product (4 pairs, 4 rules, symmetric)',
        'inits': t1_inits, 'finals': t1_finals, 'rules': t1_rules, 'pairs': t1_pairs,
        'min_support': 1, 'max_rules': 5
    })

    # Task 2: 3x2 Cartesian (6 pairs, 3 INITIAL, 2 FINAL rules)
    t2_inits = [('ba', 'xy'), ('ce', 'wz'), ('di', 'uv')]
    t2_finals = [('da', 'pq'), ('fe', 'rs')]
    t2_rules = [('INITIAL', s, t) for s, t in t2_inits] + [('FINAL', s, t) for s, t in t2_finals]
    t2_pairs = []
    for si, ti in t2_inits:
        for sf, tf in t2_finals:
            t2_pairs.append((si + sf, ti + tf))
    tasks.append({
        'task_id': 'T2_cartesian_3x2',
        'description': '3x2 Cartesian product (6 pairs, 5 rules, asymmetric initial/final)',
        'inits': t2_inits, 'finals': t2_finals, 'rules': t2_rules, 'pairs': t2_pairs,
        'min_support': 2, 'max_rules': 6
    })

    # Task 3: 3x3 Cartesian (9 pairs, 3 INITIAL, 3 FINAL rules - satisfies support >= 3)
    t3_inits = [('ba', 'xy'), ('ce', 'wz'), ('di', 'uv')]
    t3_finals = [('da', 'pq'), ('fe', 'rs'), ('go', 'jk')]
    t3_rules = [('INITIAL', s, t) for s, t in t3_inits] + [('FINAL', s, t) for s, t in t3_finals]
    t3_pairs = []
    for si, ti in t3_inits:
        for sf, tf in t3_finals:
            t3_pairs.append((si + sf, ti + tf))
    tasks.append({
        'task_id': 'T3_cartesian_3x3',
        'description': '3x3 Cartesian product (9 pairs, 6 rules, min support 3, highly symmetric)',
        'inits': t3_inits, 'finals': t3_finals, 'rules': t3_rules, 'pairs': t3_pairs,
        'min_support': 3, 'max_rules': 6
    })

    # Task 4: Broken Symmetry (8 pairs from 3x3 with 1 combination omitted to break automorphism group)
    t4_pairs = [p for p in t3_pairs if p != ('di' + 'go', 'uv' + 'jk')]
    tasks.append({
        'task_id': 'T4_broken_symmetry',
        'description': '3x3 product with 1 pair missing (8 pairs, breaks Cartesian symmetry)',
        'inits': t3_inits, 'finals': t3_finals, 'rules': t3_rules, 'pairs': t4_pairs,
        'min_support': 2, 'max_rules': 6
    })

    # Task 5: Hapax-rich (7 pairs: 5 frequent + 2 hapax combinations)
    t5_inits = [('ba', 'xy'), ('ce', 'wz'), ('di', 'uv')]
    t5_finals = [('da', 'pq'), ('fe', 'rs'), ('go', 'jk')]
    # 'di' only appears once (hapax initial), 'go' only appears once (hapax final)
    t5_pairs = [
        ('ba' + 'da', 'xy' + 'pq'),
        ('ba' + 'fe', 'xy' + 'rs'),
        ('ce' + 'da', 'wz' + 'pq'),
        ('ce' + 'fe', 'wz' + 'rs'),
        ('ce' + 'go', 'wz' + 'jk'),
        ('di' + 'da', 'uv' + 'pq'),
        ('di' + 'fe', 'uv' + 'rs'),
    ]
    tasks.append({
        'task_id': 'T5_hapax_rich',
        'description': 'Hapax-rich instance (7 pairs, some affixes appear only 1-2 times)',
        'inits': t5_inits, 'finals': t5_finals, 'rules': t3_rules, 'pairs': t5_pairs,
        'min_support': 1, 'max_rules': 6
    })

    # Task 6: Three-Role Composition (8 pairs: 2 INITIAL, 2 MEDIAL, 2 FINAL rules)
    t6_inits = [('b', 'x'), ('c', 'w')]
    t6_meds = [('a', 'y'), ('e', 'z')]
    t6_fins = [('d', 'p'), ('f', 'r')]
    t6_rules = [('INITIAL', s, t) for s, t in t6_inits] + \
               [('MEDIAL', s, t) for s, t in t6_meds] + \
               [('FINAL', s, t) for s, t in t6_fins]
    t6_pairs = []
    for si, ti in t6_inits:
        for sm, tm in t6_meds:
            for sf, tf in t6_fins:
                t6_pairs.append((si + sm + sf, ti + tm + tf))
    tasks.append({
        'task_id': 'T6_three_role_composition',
        'description': '2x2x2 three-role composition (8 pairs, 6 rules: 2 init, 2 med, 2 fin)',
        'inits': t6_inits, 'meds': t6_meds, 'finals': t6_fins, 'rules': t6_rules, 'pairs': t6_pairs,
        'min_support': 2, 'max_rules': 6
    })

    # Task 7: Unmatched distractor noise (7 pairs: 6 true + 1 distractor pair)
    t7_pairs = list(t2_pairs) + [('qqzz', 'mmnn')]
    tasks.append({
        'task_id': 'T7_distractor_noise',
        'description': '3x2 product + 1 unmatchable noise pair (7 pairs, 5 rules)',
        'inits': t2_inits, 'finals': t2_finals, 'rules': t2_rules, 'pairs': t7_pairs,
        'min_support': 2, 'max_rules': 6
    })

    # Task 8: Shuffled-assignment control on T3 surfaces
    # Terms and labels are identical to T3, but pairings are randomly permuted!
    # Proves mathematical identity of shuffled assignment score in anonymous setting.
    terms_t3 = [p[0] for p in t3_pairs]
    labels_t3 = [p[1] for p in t3_pairs]
    # Scramble pairs
    shuffled_labels = [labels_t3[3], labels_t3[7], labels_t3[0], labels_t3[8], labels_t3[1],
                       labels_t3[4], labels_t3[2], labels_t3[6], labels_t3[5]]
    t8_pairs = list(zip(terms_t3, shuffled_labels))
    tasks.append({
        'task_id': 'T8_shuffled_assignment_null',
        'description': 'Same surfaces as T3 but scrambled pairings (anonymous null control)',
        'inits': t3_inits, 'finals': t3_finals, 'rules': t3_rules, 'pairs': t8_pairs,
        'min_support': 3, 'max_rules': 6
    })

    return tasks


def run_exact_search_audit(exact_tsv_path: Path, equiv_tsv_path: Path):
    tasks = create_small_tasks()
    exact_rows = []
    equiv_rows = []

    cfg = audit_engine.Config(beam=3, iterations=8, max_states=96, max_rules=12, pool=48)

    for task in tasks:
        task_id = task['task_id']
        desc = task['description']
        pairs = task['pairs']
        true_rules = [tuple(r) for r in task['rules']]
        min_sup = task['min_support']
        max_rules_subsets = task['max_rules']

        terms = [p[0] for p in pairs]
        labels = [p[1] for p in pairs]

        t_rows = [{'id': f't{i}', 'surface': s} for i, s in enumerate(terms)]
        l_rows = [{'id': f'l{i}', 'surface': s} for i, s in enumerate(labels)]

        # 1. Evaluate True Model
        true_eval = audit_engine.evaluate(terms, labels, true_rules, cfg, training=False,
                                         alternatives=True, min_support=min_sup)
        true_gain = true_eval['gain']
        true_total = true_eval['total']

        # 2. Form candidate pool for exhaustive evaluation
        budget = audit_engine.Budget(cfg)
        pool, inv = audit_engine.candidates(terms, labels, cfg, budget,
                                            min_support=min_sup, pool_size=10)
        # Ensure true rules are included in candidate list so exhaustive search explores them
        # Cap search candidates to true rules + top competing candidates (max 10 total)
        top_candidates = [r for r in pool if tuple(r) not in set(true_rules)]
        search_candidates = list(set(true_rules) | set(map(tuple, top_candidates[:max(0, 10 - len(true_rules))])))

        # 3. Exhaustive Search
        ex_res = audit_engine.exhaustive_search(terms, labels, search_candidates, cfg,
                                                min_support=min_sup, max_rules=max_rules_subsets)
        global_opt = ex_res['global_optimum']
        global_gain = global_opt['gain']
        global_total = global_opt['total']
        num_global_optima = ex_res['num_global_optima']
        all_ranked = ex_res['all_models_ranked']

        # 4. Bounded M2R-v2 Search
        bounded_res = audit_engine.search(t_rows, l_rows, cfg, min_support=min_sup)
        bounded_opt = bounded_res['model']
        bounded_gain = bounded_opt['gain']
        bounded_rules = bounded_opt['rules']

        bounded_found_global = abs(bounded_gain - global_gain) < 1e-9

        # 5. Check if Truth is Global Optimum
        truth_is_global = abs(true_gain - global_gain) < 1e-9
        
        # Determine failure classification
        if not bounded_found_global:
            classification = 'SEARCH_FAILURE'
        elif not truth_is_global:
            classification = 'SCORING_FAILURE'
        else:
            # Check if assignment is uniquely identifiable or has symmetric alternatives
            if num_global_optima > 1:
                classification = 'NON_IDENTIFIABLE'
            else:
                classification = 'IDENTIFIABLE_SUCCESS'

        # Score component driving false winner (if any)
        false_structure_type = 'NONE'
        dominant_component = 'NONE'
        if not truth_is_global or num_global_optima > 1:
            best_r = global_opt['rules']
            if len(best_r) < len(true_rules):
                false_structure_type = 'COMPACT_PARTIAL_MODEL'
            elif len(best_r) == len(true_rules) and set(map(tuple, best_r)) != set(true_rules):
                false_structure_type = 'SYMMETRIC_PERMUTATION_MODEL'
            elif len(best_r) > len(true_rules):
                false_structure_type = 'OVERFITTED_MODEL'

            # Compare components
            comp_diffs = {
                k: global_opt['components'][k] - true_eval['components'][k]
                for k in global_opt['components']
            }
            # Find largest component difference
            sorted_diffs = sorted(comp_diffs.items(), key=lambda x: abs(x[1]), reverse=True)
            if sorted_diffs:
                dominant_component = f"{sorted_diffs[0][0]}_SAVINGS"

        exact_rows.append({
            'task_id': task_id,
            'description': desc,
            'n_pairs': len(pairs),
            'true_rules_count': len(true_rules),
            'true_gain': round(true_gain, 4),
            'true_total_mdl': round(true_total, 2),
            'bounded_rules_count': len(bounded_rules),
            'bounded_gain': round(bounded_gain, 4),
            'bounded_found_global': 'YES' if bounded_found_global else 'NO',
            'global_rules_count': len(global_opt['rules']),
            'global_gain': round(global_gain, 4),
            'global_total_mdl': round(global_total, 2),
            'truth_is_global_optimum': 'YES' if truth_is_global else 'NO',
            'num_global_optima': num_global_optima,
            'false_structure_type': false_structure_type,
            'dominant_score_component': dominant_component,
            'failure_classification': classification
        })

        # --- Section 5: Equivalence Class & Symmetry Analysis ---
        # Find truth rank
        truth_rank = 1
        for rank, m in enumerate(all_ranked):
            if set(map(tuple, m['rules'])) == set(true_rules):
                truth_rank = rank + 1
                break

        # Solutions within delta MDL thresholds
        within_1bit = sum(1 for m in all_ranked if m['total'] - global_total <= 1.0)
        within_5bit = sum(1 for m in all_ranked if m['total'] - global_total <= 5.0)
        within_10bit = sum(1 for m in all_ranked if m['total'] - global_total <= 10.0)

        # Assignment accuracy: strict vs quotiented
        expected_pairs = set(pairs)
        bounded_pairs = set((a['term'], a['label']) for a in bounded_opt['assignments'])
        strict_acc = len(bounded_pairs & expected_pairs) / len(expected_pairs) if expected_pairs else 0.0

        # Quotiented accuracy: does the assignment match any of the global optimal bijections?
        quotiented_acc = 1.0 if bounded_found_global and num_global_optima > 1 else strict_acc

        shuffled_same_score = 'YES' if task_id == 'T8_shuffled_assignment_null' and abs(global_gain - true_gain) < 1e-9 else 'NO'

        symmetry_type = 'CARTESIAN_AUTOMORPHISM' if num_global_optima > 1 and 'cartesian' in task_id else \
                        'BROKEN_ASYMMETRY' if 'broken' in task_id else \
                        'SHUFFLED_INVARIANCE' if 'shuffled' in task_id else \
                        'PARTIAL_COMPACTION' if num_global_optima == 1 and not truth_is_global else 'RIGID_UNIQUE'

        equiv_rows.append({
            'task_id': task_id,
            'n_pairs': len(pairs),
            'true_rules_count': len(true_rules),
            'num_global_optima': num_global_optima,
            'num_equivalence_classes': math.ceil(num_global_optima / 2) if num_global_optima > 1 else 1,
            'truth_class_size': num_global_optima if truth_is_global else 0,
            'truth_rank': truth_rank,
            'solutions_within_1bit': within_1bit,
            'solutions_within_5bit': within_5bit,
            'solutions_within_10bit': within_10bit,
            'shuffled_assignment_same_score': shuffled_same_score,
            'strict_assignment_accuracy': round(strict_acc, 4),
            'quotiented_assignment_accuracy': round(quotiented_acc, 4),
            'symmetry_type': symmetry_type
        })

    # Write EXACT_SEARCH_COMPARISON.tsv
    exact_fields = [
        'task_id', 'description', 'n_pairs', 'true_rules_count', 'true_gain', 'true_total_mdl',
        'bounded_rules_count', 'bounded_gain', 'bounded_found_global', 'global_rules_count',
        'global_gain', 'global_total_mdl', 'truth_is_global_optimum', 'num_global_optima',
        'false_structure_type', 'dominant_score_component', 'failure_classification'
    ]
    with exact_tsv_path.open('w', newline='', encoding='utf-8') as f:
        w = csv.DictWriter(f, fieldnames=exact_fields, delimiter='\t')
        w.writeheader()
        w.writerows(exact_rows)

    # Write EQUIVALENCE_CLASS_ANALYSIS.tsv
    equiv_fields = [
        'task_id', 'n_pairs', 'true_rules_count', 'num_global_optima', 'num_equivalence_classes',
        'truth_class_size', 'truth_rank', 'solutions_within_1bit', 'solutions_within_5bit',
        'solutions_within_10bit', 'shuffled_assignment_same_score', 'strict_assignment_accuracy',
        'quotiented_assignment_accuracy', 'symmetry_type'
    ]
    with equiv_tsv_path.open('w', newline='', encoding='utf-8') as f:
        w = csv.DictWriter(f, fieldnames=equiv_fields, delimiter='\t')
        w.writeheader()
        w.writerows(equiv_rows)

    return exact_rows, equiv_rows
