"""Candidate Pool Audit for M2R Identifiability Diagnostic v1.
Section 3: Analyzes candidate generation across all true latent rules.
"""
from __future__ import annotations
import csv
import json
from pathlib import Path
from collections import defaultdict
import audit_engine


def run_candidate_audit(registry_path: Path, output_tsv_path: Path):
    reg = json.loads(registry_path.read_text())
    # Audit hidden positive datasets, development, calibration, out_of_family
    target_families = {'hidden', 'development', 'calibration', 'out_of_family'}
    cases = [c for c in reg if c['family'] in target_families]
    
    cfg = audit_engine.Config()
    audit_rows = []
    
    # Aggregators for summary statistics
    total_true_rules = 0
    total_in_pool = 0
    total_equiv_in_pool = 0
    
    by_role_true = defaultdict(int)
    by_role_pool = defaultdict(int)
    
    by_support_true = defaultdict(int)
    by_support_pool = defaultdict(int)
    
    by_len_true = defaultdict(int)
    by_len_pool = defaultdict(int)

    for case in cases:
        surface_path = registry_path.parent / case['surface']
        truth_path = registry_path.parent / case['truth']
        
        surface = json.loads(surface_path.read_text())
        truth = json.loads(truth_path.read_text())
        
        ts, _ = audit_engine.load_rows(surface['train']['terms'], cfg)
        ls, _ = audit_engine.load_rows(surface['train']['labels'], cfg)
        
        budget = audit_engine.Budget(cfg)
        pool, inv = audit_engine.candidates(ts, ls, cfg, budget)
        
        pool_set = set(map(tuple, pool))
        all_ranked = inv.get('ranked_candidates', [])
        ranked_lookup = {r[2]: (rank + 1, -r[0], r[1]) for rank, r in enumerate(all_ranked)}
        
        source_units = { (u['role'], u['surface']): u for u in inv['source'] }
        target_units = { (u['role'], u['surface']): u for u in inv['target'] }
        
        true_rules = truth.get('rules', [])
        
        for r in true_rules:
            role_val, src, tgt = r[0], r[1], r[2]
            rule_tuple = (role_val, src, tgt)
            total_true_rules += 1
            by_role_true[role_val] += 1
            
            su = source_units.get((role_val, src))
            tu = target_units.get((role_val, tgt))
            
            s_sup = su['independent_example_support'] if su else 0
            t_sup = tu['independent_example_support'] if tu else 0
            min_sup = min(s_sup, t_sup)
            
            sup_bucket = '1' if min_sup == 1 else '2' if min_sup == 2 else '3+' if min_sup >= 3 else '0'
            by_support_true[sup_bucket] += 1
            
            max_len = max(len(src), len(tgt))
            len_bucket = str(min(max_len, 4))
            by_len_true[len_bucket] += 1
            
            in_pool = rule_tuple in pool_set
            
            # Check for equivalence: functionally identical substring mapping
            is_equiv = in_pool
            if not is_equiv:
                # Check if any candidate has same role and either exact prefix/suffix identity
                for p in pool:
                    if p[0] == role_val and (p[1] == src or p[2] == tgt):
                        is_equiv = True
                        break
            
            if in_pool:
                total_in_pool += 1
                by_role_pool[role_val] += 1
                by_support_pool[sup_bucket] += 1
                by_len_pool[len_bucket] += 1
            
            if is_equiv:
                total_equiv_in_pool += 1
            
            # Determine loss stage and reason
            if in_pool:
                exclusion_reason = 'NONE_IN_POOL'
                pool_rank = pool.index(rule_tuple) + 1
                pool_potential = ranked_lookup[rule_tuple][1] if rule_tuple in ranked_lookup else 0
            elif min_sup < 3:
                exclusion_reason = 'INSUFFICIENT_SUPPORT_LT_3'
                pool_rank = -1
                pool_potential = 0
            elif rule_tuple in ranked_lookup:
                rank, pot, diff = ranked_lookup[rule_tuple]
                exclusion_reason = 'PRUNED_BY_POOL_CAP_48'
                pool_rank = rank
                pool_potential = pot
            elif not su and not tu:
                exclusion_reason = 'BOTH_UNITS_MISSING'
                pool_rank = -1
                pool_potential = 0
            elif not su:
                exclusion_reason = 'SOURCE_UNIT_SEGMENTATION_FAILURE'
                pool_rank = -1
                pool_potential = 0
            elif not tu:
                exclusion_reason = 'TARGET_UNIT_SEGMENTATION_FAILURE'
                pool_rank = -1
                pool_potential = 0
            else:
                exclusion_reason = 'ROLE_OR_PAIRING_MISMATCH'
                pool_rank = -1
                pool_potential = 0
                
            audit_rows.append({
                'dataset': case['dataset'],
                'family': case['family'],
                'kind': case['kind'],
                'size': case['size'],
                'rule_role': role_val,
                'source_unit': src,
                'target_unit': tgt,
                'source_len': len(src),
                'target_len': len(tgt),
                'source_support': s_sup,
                'target_support': t_sup,
                'min_support': min_sup,
                'in_pool': 'YES' if in_pool else 'NO',
                'pool_rank': pool_rank,
                'pool_potential': pool_potential,
                'exclusion_reason': exclusion_reason,
                'is_equivalent': 'YES' if is_equiv else 'NO'
            })
            
    # Write TSV output
    fieldnames = [
        'dataset', 'family', 'kind', 'size', 'rule_role', 'source_unit', 'target_unit',
        'source_len', 'target_len', 'source_support', 'target_support', 'min_support',
        'in_pool', 'pool_rank', 'pool_potential', 'exclusion_reason', 'is_equivalent'
    ]
    with output_tsv_path.open('w', newline='', encoding='utf-8') as f:
        writer = csv.DictWriter(f, fieldnames=fieldnames, delimiter='\t')
        writer.writeheader()
        writer.writerows(audit_rows)
        
    summary = {
        'total_true_rules': total_true_rules,
        'true_rule_candidate_recall': total_in_pool / total_true_rules if total_true_rules else 0.0,
        'equivalence_aware_candidate_recall': total_equiv_in_pool / total_true_rules if total_true_rules else 0.0,
        'recall_by_role': {r: by_role_pool[r] / by_role_true[r] if by_role_true[r] else 0.0 for r in by_role_true},
        'recall_by_support': {s: by_support_pool[s] / by_support_true[s] if by_support_true[s] else 0.0 for s in by_support_true},
        'recall_by_len': {l: by_len_pool[l] / by_len_true[l] if by_len_true[l] else 0.0 for l in by_len_true},
        'upper_bound_latent_rule_recall': total_in_pool / total_true_rules if total_true_rules else 0.0
    }
    return audit_rows, summary
