"""Truth-aware metrics; imported by evaluation only after inference has committed."""
import math


def rule_metrics(inferred, truth):
    a, b = set(map(tuple, inferred)), set(map(tuple, truth))
    hits = len(a & b)
    # Exact identity is role-aware; the second metric explicitly ignores roles.
    x, y = {r[1:] for r in a}, {r[1:] for r in b}
    return {'precision': hits/len(a) if a else 0., 'recall': hits/len(b) if b else 0.,
            'role_aware_precision': hits/len(a) if a else 0.,
            'role_aware_recall': hits/len(b) if b else 0.,
            'role_agnostic_precision': len(x & y)/len(x) if x else 0.,
            'role_agnostic_recall': len(x & y)/len(y) if y else 0.}


def evaluate(output, truth):
    if any(output[s]['SEARCH_STATUS'] != 'COMPLETE' for s in ('train', 'heldout')):
        return {'status': 'INCOMPLETE', 'accepted': False}
    a, b = output['train']['model'], output['heldout']['model']
    result = {'status': 'COMPLETE', 'accepted': output['accepted'], **rule_metrics(a['rules'], truth['rules'])}
    for split, model in [('train', a), ('heldout', b)]:
        expected = set(map(tuple, truth['splits'][split]['assignments']))
        actual = {(r['term'], r['label']) for r in model['assignments']}
        best = set(actual)
        for alt in model['alternative_optima']:
            best.update(map(tuple, alt))
        result[split+'_assignment_accuracy'] = len(actual & expected)/len(expected) if expected else 0.
        result[split+'_top_k_accuracy'] = len(best & expected)/len(expected) if expected else 0.
        result[split+'_coverage'] = model['coverage']
        result[split+'_gain'] = model['gain']
        margins = [r['margin'] for r in model['margins']]
        result[split+'_margin'] = min(margins) if margins else 0.
        result[split+'_ambiguous'] = bool(model['alternative_optima'])
        predicted = {(r['term'], r['label']): r['alignment'] for r in model['assignments']}
        path_hits, boundary_hits = 0, 0
        true_paths = truth['splits'][split]['paths']
        for p in true_paths:
            match = predicted.get((p['term'], p['label']))
            if not match:
                continue
            inferred_rules = [op[5:] for op in match['path'] if op[0] == 'R']
            path_hits += inferred_rules == p['rules'] and match['full']
            def boundaries(rr):
                x = y = 0
                out = []
                for r in rr:
                    x, y = x+len(r[1]), y+len(r[2])
                    out.append((x, y))
                return out
            boundary_hits += boundaries(inferred_rules) == boundaries(p['rules']) and match['full']
        result[split+'_path_accuracy'] = path_hits/len(true_paths) if true_paths else None
        result[split+'_boundary_accuracy'] = boundary_hits/len(true_paths) if true_paths else None
    supports = [r['pair_support'] for r in a['support']]
    result['mean_support'] = sum(supports)/len(supports) if supports else 0.
    result['singleton_fraction'] = sum(x == 1 for x in supports)/len(supports) if supports else 0.
    result['unsupported_rules'] = sum(min(r['pair_support'], r['distinct_term_support'], r['distinct_label_support'])<3 for r in a['support'])
    result['transfer_retention'] = b['coverage']/max(a['coverage'], 1e-12)
    return result


def quantile(values, q):
    return sorted(values)[max(0, math.ceil(q*len(values))-1)] if values else None
