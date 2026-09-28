"""Oracle Decomposition Audit for M2R Identifiability Diagnostic v1.
Section 2: Executes modes A, B, C, D, E, F on synthetic datasets to isolate failure modes.
"""
from __future__ import annotations
import csv
import json
from pathlib import Path
import audit_engine


def rule_metrics(inferred, truth):
    a = set(map(tuple, inferred))
    b = set(map(tuple, truth))
    hits = len(a & b)
    x = {r[1:] for r in a}
    y = {r[1:] for r in b}
    return {
        'precision': hits / len(a) if a else 0.0,
        'recall': hits / len(b) if b else 0.0,
        'role_aware_precision': hits / len(a) if a else 0.0,
        'role_aware_recall': hits / len(b) if b else 0.0,
        'role_agnostic_precision': len(x & y) / len(x) if x else 0.0,
        'role_agnostic_recall': len(x & y) / len(y) if y else 0.0
    }


def assignment_accuracy(model_assignments, true_assignments, alternative_optima=None):
    expected = set(map(tuple, true_assignments))
    actual = {(r['term'], r['label']) for r in model_assignments}
    best = set(actual)
    if alternative_optima:
        for alt in alternative_optima:
            best.update(map(tuple, alt))
    acc = len(actual & expected) / len(expected) if expected else 0.0
    top_k = len(best & expected) / len(expected) if expected else 0.0
    return acc, top_k


def run_oracle_decomposition(registry_path: Path, output_tsv_path: Path):
    reg = json.loads(registry_path.read_text())
    target_families = {'hidden', 'development', 'calibration', 'out_of_family'}
    cases = [c for c in reg if c['family'] in target_families]
    
    cfg = audit_engine.Config()
    matrix_rows = []

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
        
        # Mode definitions:
        # A: Normal (unknown assignment, unknown rules, normal pool)
        # B: Oracle assignment (true assignment, unknown rules, normal pool)
        # C: Oracle rules (unknown assignment, true rules, oracle pool)
        # D: Oracle candidate pool (unknown assignment, unknown rules, true rules in pool)
        # E: Oracle assignment + pool (true assignment, unknown rules, true rules in pool)
        # F: Full oracle ceiling (true assignment, true rules, oracle pool)
        
        modes = ['A_Normal', 'B_OracleAssignment', 'C_OracleRules',
                 'D_OracleCandidatePool', 'E_OracleAssignmentAndPool', 'F_FullOracleCeiling']
                 
        for mode in modes:
            status = 'COMPLETE'
            try:
                if mode == 'A_Normal':
                    res = audit_engine.search(term_rows, label_rows, cfg)
                    m = res['model']
                    inferred_rules = m['rules']
                    assignments = m['assignments']
                    alts = m.get('alternative_optima', [])
                elif mode == 'B_OracleAssignment':
                    res = audit_engine.search(term_rows, label_rows, cfg, oracle_assignments=true_pairs)
                    m = res['model']
                    inferred_rules = m['rules']
                    assignments = m['assignments']
                    alts = m.get('alternative_optima', [])
                elif mode == 'C_OracleRules':
                    m = audit_engine.evaluate(ts, ls, true_rules, cfg, training=False, alternatives=True)
                    inferred_rules = true_rules
                    assignments = m['assignments']
                    alts = m.get('alternative_optima', [])
                elif mode == 'D_OracleCandidatePool':
                    res = audit_engine.search(term_rows, label_rows, cfg, force_rules=true_rules)
                    m = res['model']
                    inferred_rules = m['rules']
                    assignments = m['assignments']
                    alts = m.get('alternative_optima', [])
                elif mode == 'E_OracleAssignmentAndPool':
                    res = audit_engine.search(term_rows, label_rows, cfg, force_rules=true_rules, oracle_assignments=true_pairs)
                    m = res['model']
                    inferred_rules = m['rules']
                    assignments = m['assignments']
                    alts = m.get('alternative_optima', [])
                elif mode == 'F_FullOracleCeiling':
                    m = audit_engine.evaluate(ts, ls, true_rules, cfg, training=False, alternatives=False, oracle_assignments=true_pairs)
                    inferred_rules = true_rules
                    assignments = m['assignments']
                    alts = []
                    
                rm = rule_metrics(inferred_rules, true_rules)
                acc, top_k = assignment_accuracy(assignments, true_pairs, alts)
                margins = [r['margin'] for r in m.get('margins', [])]
                min_margin = min(margins) if margins else (0.0 if alts else 1.0)
                unsup = sum(min(r['pair_support'], r['distinct_term_support'], r['distinct_label_support']) < 3
                            for r in m.get('support', []))
                            
                matrix_rows.append({
                    'dataset': case['dataset'],
                    'family': case['family'],
                    'kind': case['kind'],
                    'size': case['size'],
                    'mode': mode,
                    'status': status,
                    'rules_count': len(inferred_rules),
                    'rule_precision': round(rm['precision'], 4),
                    'rule_recall': round(rm['recall'], 4),
                    'role_agnostic_precision': round(rm['role_agnostic_precision'], 4),
                    'role_agnostic_recall': round(rm['role_agnostic_recall'], 4),
                    'assignment_accuracy': round(acc, 4),
                    'top_k_accuracy': round(top_k, 4),
                    'train_gain': round(m['gain'], 4),
                    'train_coverage': round(m['coverage'], 4),
                    'valid_support': 'YES' if m.get('valid_support') else 'NO',
                    'unsupported_rules': unsup,
                    'alternative_optima_count': len(alts),
                    'min_margin': round(min_margin, 4)
                })
            except Exception as e:
                matrix_rows.append({
                    'dataset': case['dataset'],
                    'family': case['family'],
                    'kind': case['kind'],
                    'size': case['size'],
                    'mode': mode,
                    'status': f'ERROR: {e}',
                    'rules_count': 0,
                    'rule_precision': 0.0,
                    'rule_recall': 0.0,
                    'role_agnostic_precision': 0.0,
                    'role_agnostic_recall': 0.0,
                    'assignment_accuracy': 0.0,
                    'top_k_accuracy': 0.0,
                    'train_gain': 0.0,
                    'train_coverage': 0.0,
                    'valid_support': 'NO',
                    'unsupported_rules': 0,
                    'alternative_optima_count': 0,
                    'min_margin': 0.0
                })

    fieldnames = [
        'dataset', 'family', 'kind', 'size', 'mode', 'status', 'rules_count',
        'rule_precision', 'rule_recall', 'role_agnostic_precision', 'role_agnostic_recall',
        'assignment_accuracy', 'top_k_accuracy', 'train_gain', 'train_coverage',
        'valid_support', 'unsupported_rules', 'alternative_optima_count', 'min_margin'
    ]
    with output_tsv_path.open('w', newline='', encoding='utf-8') as f:
        writer = csv.DictWriter(f, fieldnames=fieldnames, delimiter='\t')
        writer.writeheader()
        writer.writerows(matrix_rows)
        
    return matrix_rows
