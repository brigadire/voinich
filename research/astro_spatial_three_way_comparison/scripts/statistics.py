"""NumPy plus exact combinatorics; no network or optional scientific libraries."""
from __future__ import annotations
import hashlib
import math
import numpy as np
from common import SEED, BOOTSTRAPS

def rng(key):
    return np.random.default_rng(SEED + int(hashlib.sha256(key.encode()).hexdigest()[:12], 16))

def wilson(k, n):
    if not n: return [None, None]
    z = 1.959963984540054
    p = k/n
    den = 1 + z*z/n
    center = (p + z*z/(2*n))/den
    half = z*math.sqrt(p*(1-p)/n + z*z/(4*n*n))/den
    return [max(0., center-half), min(1., center+half)]

def fraction(k, n, values=None, panels=None, key=''):
    result = {'numerator': int(k), 'denominator': int(n), 'estimate': k/n if n else None,
              'ci95': wilson(k, n), 'ci_method': 'Wilson_conditional_candidate_interval'}
    if values is not None:
        boot, method = bootstrap(values, panels, key, 'mean')
        result.update(cluster_ci95=interval(boot), cluster_ci_method=method, valid_bootstraps=len(boot))
    return result

def bootstrap(values, panels, key, statistic='median', replicates=BOOTSTRAPS):
    values = np.asarray(values, dtype=float)
    panels = np.asarray(panels)
    if not len(values): return np.array([]), 'NO_OBSERVATIONS'
    groups = sorted(set(panels))
    generator = rng(key)
    fn = np.median if statistic == 'median' else np.mean
    if len(groups) == 1:
        draws = generator.integers(len(values), size=(replicates, len(values)))
        return fn(values[draws], axis=1), 'candidate_bootstrap_single_panel_no_between_panel_inference'
    draws = generator.integers(len(groups), size=(replicates, len(groups)))
    # Exact cluster resampling represented as integer multiplicities. Weighted
    # order statistics avoid thousands of Python concatenate/sort operations.
    counts = np.stack([np.sum(draws == i,axis=1) for i in range(len(groups))],axis=1)
    order = np.argsort(values,kind='stable')
    sorted_values = values[order]
    panel_index = np.array([groups.index(p) for p in panels[order]])
    weights = counts[:,panel_index]
    totals = weights.sum(axis=1)
    if statistic == 'mean':
        result = (weights*sorted_values).sum(axis=1)/totals
    else:
        cumulative = weights.cumsum(axis=1)
        lo = (cumulative > ((totals-1)//2)[:,None]).argmax(axis=1)
        hi = (cumulative > (totals//2)[:,None]).argmax(axis=1)
        result = (sorted_values[lo]+sorted_values[hi])/2
    return result, 'panel_cluster_percentile'

def interval(values):
    return [float(v) for v in np.quantile(values, [.025, .975])] if len(values) else [None, None]

def summarize(values, panels, key):
    values = np.asarray(values, dtype=float)
    if not len(values):
        return {'n': 0, 'median': None, 'q1': None, 'q3': None, 'iqr': None, 'ci95': [None, None], 'ci_method': 'NO_OBSERVATIONS'}
    q1, median, q3 = np.quantile(values, [.25,.5,.75])
    boot, method = bootstrap(values, panels, key)
    return {'n': len(values), 'median': float(median), 'q1': float(q1), 'q3': float(q3),
            'iqr': float(q3-q1), 'mean': float(np.mean(values)), 'ci95': interval(boot),
            'ci_method': method, 'panel_clusters': len(set(panels)), 'bootstrap_replicates': len(boot)}

def contrast(left, lp, right, rp, key, statistic='mean'):
    """Resample shared panels jointly; discard undefined group denominators."""
    if not len(left) or not len(right): return None, [None, None], None, 0, 'INSUFFICIENT_GROUP'
    left, right = np.asarray(left, dtype=float), np.asarray(right, dtype=float)
    lp, rp = np.asarray(lp), np.asarray(rp)
    fn = np.mean if statistic == 'mean' else np.median
    effect = float(fn(left) - fn(right))
    panels = sorted(set(lp) | set(rp))
    generator = rng(key)
    boot = []
    if len(panels) == 1:
        lb, _ = bootstrap(left, lp, key+'-left', statistic)
        rb, _ = bootstrap(right, rp, key+'-right', statistic)
        boot = lb-rb
        method = 'independent_candidate_bootstrap_single_panel'
    else:
        for draw in generator.integers(len(panels), size=(BOOTSTRAPS,len(panels))):
            lv = np.concatenate([left[lp == panels[i]] for i in draw])
            rv = np.concatenate([right[rp == panels[i]] for i in draw])
            if len(lv) and len(rv): boot.append(float(fn(lv)-fn(rv)))
        method = 'joint_panel_cluster_percentile'
    boot = np.asarray(boot)
    # Supplementary sign-tail evidence, not a randomization/exact p-value.
    p = min(1., 2*min((np.sum(boot <= 0)+1)/(len(boot)+1), (np.sum(boot >= 0)+1)/(len(boot)+1))) if len(boot) else None
    return effect, interval(boot), p, len(boot), method

def fisher(a, b, c, d):
    n, row1, col1 = a+b+c+d, a+b, a+c
    def logchoose(n, k):
        return math.lgamma(n+1)-math.lgamma(k+1)-math.lgamma(n-k+1)
    def probability(x):
        return math.exp(logchoose(col1,x)+logchoose(n-col1,row1-x)-logchoose(n,row1))
    observed = probability(a)
    return min(1., sum(probability(x) for x in range(max(0,row1-(n-col1)),min(row1,col1)+1)
                       if probability(x) <= observed*(1+1e-10)))

def holm(rows):
    families = sorted({r['family'] for r in rows})
    for family in families:
        selected = sorted([r for r in rows if r['family'] == family and r['p_raw'] is not None], key=lambda r: (r['p_raw'],r['test_id']))
        last = 0.
        for i, row in enumerate(selected):
            last = max(last, min(1., row['p_raw']*(len(selected)-i)))
            row['p_holm'] = last
        for row in rows:
            if row['family'] == family:
                row['family_test_count'] = len(selected)

def categorical(pairs, classes):
    matrix = {a: {b: 0 for b in classes} for a in classes}
    for a, b in pairs: matrix[a][b] += 1
    n = len(pairs)
    agreement = sum(matrix[c][c] for c in classes)/n if n else None
    chance = sum(sum(matrix[c].values())*sum(matrix[a][c] for a in classes) for c in classes)/(n*n) if n else None
    kappa = (agreement-chance)/(1-chance) if n and chance < 1 else None
    return {'n':n, 'matrix':matrix, 'raw_agreement':agreement,
            'raw_agreement_fraction':fraction(sum(matrix[c][c] for c in classes),n), 'cohen_kappa':kappa,
            'kappa_status':'ESTIMABLE' if kappa is not None else 'NO_VARIATION_OR_NO_OBSERVATIONS'}
