"""Assignment Identifiability Audit for M2R Identifiability Diagnostic v1.
Section 6: Analyzes bipartite assignment identifiability when true rules are known.
"""
from __future__ import annotations
import copy
import csv
import json
import math
from pathlib import Path
import random
import audit_engine


def run_assignment_identifiability_audit(registry_path: Path, output_tsv_path: Path):
    reg = json.loads(registry_path.read_text())
    target_families = {'hidden', 'development', 'calibration', 'out_of_family'}
    cases = [c for c in reg if c['family'] in target_families]

    cfg = audit_engine.Config()
    ident_rows = []

    for case in cases:
        surface_path = registry_path.parent / case['surface']
        truth_path = registry_path.parent / case['truth']

        surface = json.loads(surface_path.read_text())
        truth = json.loads(truth_path.read_text())

        term_rows = surface['train']['terms']
        label_rows = surface['train']['labels']
        ts, _ = audit_engine.load_rows(term_rows, cfg)
        ls, _ = audit_engine.load_rows(label_rows, cfg)

        true_rules = truth.get('rules', [])
        true_pairs = [tuple(p) for p in truth['splits']['train']['assignments']]
        expected_set = set(true_pairs)
        truth_dict = dict(true_pairs)

        # 1. Evaluate with true rules and alternatives enabled
        eval_res = audit_engine.evaluate(ts, ls, true_rules, cfg, training=False, alternatives=True)
        assigned = eval_res['assignments']
        alts = eval_res['alternative_optima']

        actual_pairs = set((a['term'], a['label']) for a in assigned)
        bipartite_acc = len(actual_pairs & expected_set) / len(expected_set) if expected_set else 0.0

        # Build pairwise weight matrix from pairwise explanations
        weights = {}
        for row in eval_res['pairwise']:
            weights[(row['term'], row['label'])] = row['weight']

        # 2. Row-wise (term) Top-1, Top-k, margin, entropy, ambiguity
        margins = []
        entropies = []
        top_1_hits = 0
        top_3_hits = 0
        ambiguous_rows = 0

        for t in ts:
            row_weights = [(weights.get((t, l), 0.0), l) for l in ls]
            row_weights.sort(key=lambda x: x[0], reverse=True)
            w_sorted = [x[0] for x in row_weights]
            l_sorted = [x[1] for x in row_weights]

            true_label = truth_dict.get(t)
            if l_sorted and l_sorted[0] == true_label:
                top_1_hits += 1
            if true_label in l_sorted[:3]:
                top_3_hits += 1

            if len(w_sorted) >= 2:
                m = w_sorted[0] - w_sorted[1]
                margins.append(m)
                if m < 1e-6:
                    ambiguous_rows += 1
            else:
                margins.append(w_sorted[0] if w_sorted else 0.0)

            # Softmax entropy
            if any(w > 0 for w in w_sorted):
                max_w = max(w_sorted)
                # scale for numerical stability
                exp_w = [math.exp(min(50.0, max(-50.0, (w - max_w) / 4.0))) for w in w_sorted]
                sum_exp = sum(exp_w)
                probs = [p / sum_exp for p in exp_w]
                ent = -sum(p * math.log2(p + 1e-12) for p in probs)
                entropies.append(ent)
            else:
                entropies.append(0.0)

        # 3. Column-wise (label) ambiguity
        ambiguous_cols = 0
        for l in ls:
            col_weights = [weights.get((t, l), 0.0) for t in ts]
            col_weights.sort(reverse=True)
            if len(col_weights) >= 2 and (col_weights[0] - col_weights[1]) < 1e-6:
                ambiguous_cols += 1

        top_1_acc = top_1_hits / len(ts) if ts else 0.0
        top_k_acc = top_3_hits / len(ts) if ts else 0.0
        mean_margin = sum(margins) / len(margins) if margins else 0.0
        min_margin = min(margins) if margins else 0.0
        mean_entropy = sum(entropies) / len(entropies) if entropies else 0.0
        ambiguous_rows_frac = ambiguous_rows / len(ts) if ts else 0.0
        ambiguous_cols_frac = ambiguous_cols / len(ls) if ls else 0.0

        # 4. Stability under input permutation
        rev_ts = list(reversed(ts))
        rev_ls = list(reversed(ls))
        eval_rev = audit_engine.evaluate(rev_ts, rev_ls, true_rules, cfg, training=False)
        rev_pairs = set((a['term'], a['label']) for a in eval_rev['assignments'])
        order_stable = 'PASS' if rev_pairs == actual_pairs else 'FAIL'

        # 5. Stability under small surface perturbation (noise)
        # Check if Hungarian matching changes under tiny +/- 0.01 bit weights
        perturbed_weights = []
        rng = random.Random(42)
        for i, t in enumerate(ts):
            row = []
            for j, l in enumerate(ls):
                w = weights.get((t, l), 0.0)
                if w > 0:
                    w += rng.uniform(-0.01, 0.01)
                row.append(max(0.0, w))
            perturbed_weights.append(row)
        p_pairs, _ = audit_engine.hungarian(perturbed_weights)
        perturbed_actual = set((ts[i], ls[j]) for i, j in p_pairs)
        noise_stable = 'PASS' if len(perturbed_actual & actual_pairs) / max(1, len(actual_pairs)) >= 0.95 else 'FAIL'

        ident_rows.append({
            'dataset': case['dataset'],
            'family': case['family'],
            'kind': case['kind'],
            'size': case['size'],
            'bipartite_accuracy': round(bipartite_acc, 4),
            'top_1_accuracy': round(top_1_acc, 4),
            'top_k_accuracy': round(top_k_acc, 4),
            'mean_margin': round(mean_margin, 4),
            'min_margin': round(min_margin, 4),
            'mean_entropy': round(mean_entropy, 4),
            'ambiguous_rows_fraction': round(ambiguous_rows_frac, 4),
            'ambiguous_cols_fraction': round(ambiguous_cols_frac, 4),
            'num_alternative_optima': len(alts),
            'order_permutation_stable': order_stable,
            'noise_stable': noise_stable
        })

    fields = [
        'dataset', 'family', 'kind', 'size', 'bipartite_accuracy', 'top_1_accuracy',
        'top_k_accuracy', 'mean_margin', 'min_margin', 'mean_entropy',
        'ambiguous_rows_fraction', 'ambiguous_cols_fraction', 'num_alternative_optima',
        'order_permutation_stable', 'noise_stable'
    ]
    with output_tsv_path.open('w', newline='', encoding='utf-8') as f:
        w = csv.DictWriter(f, fieldnames=fields, delimiter='\t')
        w.writeheader()
        w.writerows(ident_rows)

    return ident_rows
