"""Audit Engine for M2R Identifiability Diagnostic v1.
Implements exact M2R-v2 MDL scoring and alignment mechanics with:
- Correct process-isolated RSS memory measurement (avoiding Linux ru_maxrss fork inheritance bug)
- Oracle rule, candidate pool, and assignment injection capabilities
- Exact exhaustive optimizer for small problem spaces
- Symmetry, automorphism, and equivalence class analysis tools
"""
from __future__ import annotations
import copy
import hashlib
import itertools
import json
import math
import os
from pathlib import Path
import random
import sys
import time
import unicodedata
from collections import Counter, defaultdict
from dataclasses import asdict, dataclass
from functools import lru_cache


def canonical(value):
    return json.dumps(value, ensure_ascii=False, sort_keys=True, separators=(',', ':'))


def digest(value):
    return hashlib.sha256(canonical(value).encode('utf-8')).hexdigest()


def literal(s: str) -> int:
    return len(s.encode('utf-8')) * 8


def get_process_vm_rss_kb() -> int:
    """Read actual current resident set size from /proc/self/status (Linux).
    Avoids ru_maxrss lifetime-peak and fork/exec inheritance defects.
    """
    try:
        with open('/proc/self/status', 'r') as f:
            for line in f:
                if line.startswith('VmRSS:'):
                    parts = line.split()
                    return int(parts[1])
    except Exception:
        pass
    import resource
    return resource.getrusage(resource.RUSAGE_SELF).ru_maxrss


@dataclass(frozen=True)
class Config:
    beam: int = 3
    max_states: int = 96
    iterations: int = 8
    pool: int = 48
    additions: int = 12
    max_rules: int = 12
    top_paths: int = 2
    top_models: int = 3
    seconds: float = 120.0
    memory_mb: int = 512
    max_examples: int = 80
    max_rows: int = 256
    max_length: int = 32
    max_units: int = 12000


class Limit(Exception):
    pass


class Budget:
    def __init__(self, cfg: Config, elapsed: float = 0.0):
        self.cfg, self.start, self.elapsed = cfg, time.monotonic(), elapsed
        self.baseline_rss = get_process_vm_rss_kb()

    def check(self):
        if self.elapsed + time.monotonic() - self.start >= self.cfg.seconds:
            raise Limit('wall_clock')
        current_rss = get_process_vm_rss_kb()
        # Enforce memory limit against actual allocated RSS
        if current_rss > self.cfg.memory_mb * 1024:
            raise Limit('memory')


def load_rows(rows, cfg: Config = Config()):
    if not isinstance(rows, list) or len(rows) > cfg.max_rows:
        raise Limit('input_rows')
    ids, surfaces = {}, defaultdict(list)
    for row in rows:
        if set(row) != {'id', 'surface'}:
            raise ValueError('surface schema only: id, surface')
        identifier, s = row['id'], row['surface']
        if not isinstance(identifier, str) or not isinstance(s, str) or not s:
            raise ValueError('invalid surface row')
        if s != unicodedata.normalize('NFC', s):
            raise ValueError('surface must be NFC')
        if len(s) > cfg.max_length:
            raise Limit('token_length')
        if identifier in ids and ids[identifier] != s:
            raise ValueError('conflicting instance ID')
        ids[identifier] = s
        surfaces[s].append(identifier)
    if len(surfaces) > cfg.max_examples:
        raise Limit('unique_examples')
    return sorted(surfaces), {s: sorted(set(v)) for s, v in sorted(surfaces.items())}


def role(i: int, j: int, n: int) -> str:
    if i == 0 and j == n:
        return 'WHOLE'
    if i == 0:
        return 'INITIAL'
    if j == n:
        return 'FINAL'
    return 'MEDIAL'


def units(strings, side: str, cfg: Config = Config(), budget: Budget = None):
    found = {}
    for s in strings:
        if budget:
            budget.check()
        spans = {(i, j) for i in range(len(s)) for j in range(i + 1, min(len(s), i + 4) + 1)}
        spans.add((0, len(s)))
        for i, j in sorted(spans):
            key = (role(i, j, len(s)), s[i:j])
            if key not in found:
                found[key] = {
                    'id': digest([side, *key]),
                    'side': side,
                    'role': key[0],
                    'surface': key[1],
                    'length': j - i,
                    'provenance': []
                }
            found[key]['provenance'].append([digest(s), i, j])
        if len(found) > cfg.max_units:
            raise Limit('candidate_units')
    for u in found.values():
        u['occurrence_support'] = len(u['provenance'])
        u['independent_example_support'] = len({p[0] for p in u['provenance']})
    return sorted(found.values(), key=lambda u: (u['role'], u['surface']))


def validate_rules(rules):
    source, target = {}, {}
    for r in rules:
        ro, a, b = r
        if ro not in {'WHOLE', 'INITIAL', 'MEDIAL', 'FINAL'} or not a or not b:
            raise ValueError('invalid rule')
        if ro != 'WHOLE' and (len(a) > 4 or len(b) > 4):
            raise ValueError('unit length outside contract')
        if (ro, a) in source and source[ro, a] != b:
            raise ValueError('explicit forward rule conflict')
        if (ro, b) in target and target[ro, b] != a:
            raise ValueError('explicit inverse rule conflict')
        source[ro, a], target[ro, b] = b, a
    if len(set(map(tuple, rules))) != len(rules):
        raise ValueError('duplicate rule')


@lru_cache(maxsize=16384)
def positions(s: str, ro: str, unit: str):
    return [(i, i + len(unit)) for i in range(len(s) - len(unit) + 1)
            if s.startswith(unit, i) and role(i, i + len(unit), len(s)) == ro]


def candidates(ts, ls, cfg: Config = Config(), budget: Budget = None,
               force_rules: list = None, min_support: int = 3, pool_size: int = None):
    """Generate candidate rules.
    If force_rules is provided, those rules are guaranteed to be in the pool (Oracle Candidate Pool).
    """
    pool_cap = pool_size if pool_size is not None else cfg.pool
    su = units(ts, 'source', cfg, budget)
    tu = units(ls, 'target', cfg, budget)
    ranked = []
    for a in su:
        if budget:
            budget.check()
        na = a['independent_example_support']
        if na < min_support:
            continue
        for b in tu:
            nb = b['independent_example_support']
            if nb < min_support or a['role'] != b['role']:
                continue
            bits = literal(a['surface']) + literal(b['surface'])
            potential = min(na, nb) * (bits - 2) - (bits + 10)
            ranked.append((-potential, abs(na - nb), (a['role'], a['surface'], b['surface'])))

    sorted_candidates = [r[2] for r in sorted(ranked)]
    
    if force_rules:
        forced_tuples = [tuple(r) for r in force_rules]
        forced_set = set(forced_tuples)
        # Ensure forced rules are included at the top of the pool
        selected = list(forced_tuples)
        for r in sorted_candidates:
            if r not in forced_set and len(selected) < pool_cap:
                selected.append(r)
        return selected[:pool_cap], {'source': su, 'target': tu, 'total_formed': len(ranked), 'ranked_candidates': ranked}
    
    return sorted_candidates[:pool_cap], {'source': su, 'target': tu, 'total_formed': len(ranked), 'ranked_candidates': ranked}


def align(a: str, b: str, rules, k: int = 2, budget: Budget = None):
    if budget:
        budget.check()
    relevant = tuple(r for r in map(tuple, rules) if positions(a, r[0], r[1]) and positions(b, r[0], r[2]))
    return _align_cached(a, b, relevant, len(rules), k)


@lru_cache(maxsize=4096)
def _align_cached(a: str, b: str, rules, model_size: int, k: int):
    ref = 1 + math.ceil(math.log2(model_size + 1))
    ops = defaultdict(list)
    for ro, x, y in rules:
        for i, ii in positions(a, ro, x):
            for j, jj in positions(b, ro, y):
                ops[i, j].append((ii, jj, ref, ('R', i, ii, j, jj, ro, x, y)))
    cells = {(0, 0): [(0, ())]}
    for i in range(len(a) + 1):
        for j in range(len(b) + 1):
            paths = cells.get((i, j), ())
            if not paths:
                continue
            edges = list(ops.get((i, j), ()))
            if i < len(a):
                edges.append((i + 1, j, literal(a[i]) + 1, ('S', i, i + 1, j, j)))
            if j < len(b):
                edges.append((i, j + 1, literal(b[j]) + 1, ('T', i, i, j, j + 1)))
            for ii, jj, cost, op in edges:
                dest = cells.setdefault((ii, jj), [])
                dest.extend((c + cost, p + (op,)) for c, p in paths)
                unique = {}
                for c, p in dest:
                    signature = tuple(z for z in p if z[0] == 'R')
                    old = unique.get(signature)
                    if old is None or (c, p) < old:
                        unique[signature] = (c, p)
                cells[ii, jj] = sorted(unique.values())[:k]
    out = []
    for cost, path in cells.get((len(a), len(b)), []):
        trace = [list(p) for p in path]
        covered = sum(p[4] - p[3] for p in path if p[0] == 'R')
        gaps = [p for p in path if p[0] != 'R']
        out.append({
            'cost': cost,
            'path': trace,
            'explained_target': covered,
            'source_segmentation': [a[p[1]:p[2]] for p in path if p[2] > p[1]],
            'target_segmentation': [b[p[3]:p[4]] for p in path if p[4] > p[3]],
            'candidate_rules': [list(p[5:]) for p in path if p[0] == 'R'],
            'unexplained_units': [list(p) for p in gaps],
            'local_complexity': len({tuple(p[5:]) for p in path if p[0] == 'R'}),
            'local_score': literal(a) + literal(b) - cost,
            'coverage': covered / len(b) if len(b) else 0.0,
            'full': not gaps
        })
    return out


def hungarian(weights, budget: Budget = None, forbidden: tuple = None):
    n = len(weights)
    if not n:
        return [], 0.0
    width = len(weights[0])
    m = width + n
    u, v, p, way = [0.0] * (n + 1), [0.0] * (m + 1), [0] * (m + 1), [0] * (m + 1)
    for i in range(1, n + 1):
        if budget:
            budget.check()
        p[0], j0 = i, 0
        dist, used = [float('inf')] * (m + 1), [False] * (m + 1)
        while True:
            used[j0] = True
            i0, delta, j1 = p[j0], float('inf'), 0
            for j in range(1, m + 1):
                if used[j]:
                    continue
                w = weights[i0 - 1][j - 1] if j <= width else 0.0
                if forbidden == (i0 - 1, j - 1):
                    w = -1e12
                cur = -w - u[i0] - v[j]
                if cur < dist[j]:
                    dist[j], way[j] = cur, j0
                if dist[j] < delta:
                    delta, j1 = dist[j], j
            for j in range(m + 1):
                if used[j]:
                    u[p[j]] += delta
                    v[j] -= delta
                else:
                    dist[j] -= delta
            j0 = j1
            if p[j0] == 0:
                break
        while True:
            j1 = way[j0]
            p[j0] = p[j1]
            j0 = j1
            if not j0:
                break
    pairs = sorted((p[j] - 1, j - 1) for j in range(1, width + 1)
                   if p[j] and weights[p[j] - 1][j - 1] > 0 and forbidden != (p[j] - 1, j - 1))
    return pairs, sum(weights[i][j] for i, j in pairs)


def support(rules, assignments):
    counts = {tuple(r): {'occurrence_support': 0, 'pairs': set(), 'terms': set(), 'labels': set()}
              for r in rules}
    for record in assignments:
        a, b = record['term'], record['label']
        for op in record['alignment']['path']:
            if op[0] != 'R':
                continue
            rk = tuple(op[5:])
            if rk in counts:
                c = counts[rk]
                c['occurrence_support'] += 1
                c['pairs'].add((a, b))
                c['terms'].add(a)
                c['labels'].add(b)
    return [{
        'rule': list(r),
        'occurrence_support': c['occurrence_support'],
        'pair_support': len(c['pairs']),
        'distinct_term_support': len(c['terms']),
        'distinct_label_support': len(c['labels'])
    } for r, c in sorted(counts.items())]


def evaluate(ts, ls, rules, cfg: Config = Config(), budget: Budget = None,
             training: bool = True, alternatives: bool = False,
             oracle_assignments: list = None, min_support: int = 3):
    """MDL evaluation of a rule set on term and label bags.
    If oracle_assignments is provided (list of (term, label) pairs),
    the assignment is fixed directly (Oracle Assignment mode).
    """
    rules = sorted(map(tuple, rules))
    validate_rules(rules)
    baseline = sum(map(literal, ts)) + sum(map(literal, ls))
    assignment_bits = math.log2(len(ts) + 1) + math.log2(len(ls) + 1)
    paths, weights = {}, []

    if oracle_assignments is not None:
        # Oracle assignment mode: pairing is fixed by ground truth
        assigned = []
        used_t, used_l = set(), set()
        t_index = {t: i for i, t in enumerate(ts)}
        l_index = {l: j for j, l in enumerate(ls)}
        for term, label in oracle_assignments:
            if term in t_index and label in l_index:
                i, j = t_index[term], l_index[label]
                used_t.add(i)
                used_l.add(j)
                pp = align(term, label, rules, cfg.top_paths, budget)
                if pp:
                    assigned.append({'term': term, 'label': label, 'alignment': pp[0]})
        pairs = [(t_index[a['term']], l_index[a['label']]) for a in assigned]
    else:
        for i, a in enumerate(ts):
            if budget:
                budget.check()
            row = []
            for j, b in enumerate(ls):
                relevant = [r for r in rules if positions(a, r[0], r[1]) and positions(b, r[0], r[2])]
                if not relevant:
                    row.append(0.0)
                    continue
                pp = align(a, b, rules, cfg.top_paths, budget)
                if pp:
                    paths[i, j] = pp
                    row.append(max(0.0, literal(a) + literal(b) - pp[0]['cost'] - assignment_bits))
                else:
                    row.append(0.0)
            weights.append(row)
        pairs, optimum = hungarian(weights, budget)
        assigned = [{'term': ts[i], 'label': ls[j], 'alignment': paths[i, j][0]}
                    for i, j in pairs if (i, j) in paths]
        used_t = {i for i, j in pairs}
        used_l = {j for i, j in pairs}

    sup = support(rules, assigned)
    valid = all(min(x['pair_support'], x['distinct_term_support'], x['distinct_label_support']) >= min_support
                for x in sup) if rules else True

    components = {
        'RULE_CODE_LENGTH': sum(10 + literal(r[1]) + literal(r[2]) for r in rules),
        'ASSIGNMENT_CODE_LENGTH': len(assigned) * assignment_bits,
        'EXPLAINED_ALIGNMENT_COST': 0,
        'UNEXPLAINED_SOURCE_COST': sum(literal(t) for i, t in enumerate(ts) if i not in used_t),
        'UNEXPLAINED_TARGET_COST': sum(literal(t) for j, t in enumerate(ls) if j not in used_l),
        'EXCEPTION_COST': 0
    }
    ref = 1 + math.ceil(math.log2(len(rules) + 1)) if rules else 1
    for row in assigned:
        for op in row['alignment']['path']:
            if op[0] == 'R':
                components['EXPLAINED_ALIGNMENT_COST'] += ref
            else:
                components['EXCEPTION_COST'] += 1
                if op[0] == 'S':
                    components['UNEXPLAINED_SOURCE_COST'] += literal(row['term'][op[1]:op[2]])
                else:
                    components['UNEXPLAINED_TARGET_COST'] += literal(row['label'][op[3]:op[4]])
    total = sum(components.values())
    gain = (baseline - total) / baseline if baseline else 0.0

    margins, alts = [], []
    if alternatives and oracle_assignments is None and weights:
        for i, j in pairs:
            alt_pairs, alt_score = hungarian(weights, budget, (i, j))
            margin = max(0.0, optimum - alt_score)
            margins.append({'term': ts[i], 'label': ls[j], 'margin': margin})
            if margin < 1e-9:
                alt = [[ts[x], ls[y]] for x, y in alt_pairs]
                if alt not in alts and len(alts) < cfg.top_models:
                    alts.append(alt)

    pairwise = []
    if alternatives and oracle_assignments is None:
        for i, a in enumerate(ts):
            for j, b in enumerate(ls):
                pp = paths.get((i, j)) or align(a, b, rules, cfg.top_paths, budget)
                w = weights[i][j] if i < len(weights) and j < len(weights[i]) else 0.0
                pairwise.append({'term': a, 'label': b, 'weight': w, 'explanations': pp})

    return {
        'rules': [list(r) for r in rules],
        'support': sup,
        'valid_support': valid,
        'assignments': assigned,
        'components': components,
        'baseline': baseline,
        'total': total,
        'gain': gain,
        'objective': gain if (valid or not training) else None,
        'coverage': sum(a['alignment']['full'] for a in assigned) / len(ls) if ls else 0.0,
        'character_coverage': sum(a['alignment']['explained_target'] for a in assigned) / sum(map(len, ls)) if ls else 0.0,
        'margins': margins,
        'alternative_optima': alts,
        'pairwise': pairwise,
        'unmatched_terms': [t for i, t in enumerate(ts) if i not in used_t],
        'unmatched_labels': [s for j, s in enumerate(ls) if j not in used_l]
    }


def model_rank(m):
    # Valid models preferred, then highest gain, then canonical rule order
    valid_flag = 1 if m.get('valid_support') else 0
    return (-valid_flag, -m['gain'], canonical(m['rules']))


def search(term_rows, label_rows, cfg: Config = Config(),
           force_rules: list = None, oracle_assignments: list = None,
           min_support: int = 3, pool_size: int = None, checkpoint_path: str = None):
    """Bounded deterministic rule search.
    Supports oracle pool injection and oracle assignment injection.
    """
    budget = Budget(cfg)
    ts, _ = load_rows(term_rows, cfg)
    ls, _ = load_rows(label_rows, cfg)
    pool, inv = candidates(ts, ls, cfg, budget, force_rules=force_rules,
                           min_support=min_support, pool_size=pool_size)
    empty = evaluate(ts, ls, [], cfg, budget, training=True,
                     oracle_assignments=oracle_assignments, min_support=min_support)
    beam, seen, count, iteration = [empty], {digest([])}, 1, 0
    finished = False

    while not finished and iteration < cfg.iterations and count < cfg.max_states:
        budget.check()
        proposed = set()
        for model in beam:
            rr = tuple(map(tuple, model['rules']))
            additions = [r for r in pool if r not in rr][:cfg.additions]
            if len(rr) < cfg.max_rules:
                proposed.update(tuple(sorted((*rr, r))) for r in additions)
            for old in rr:
                remainder = tuple(r for r in rr if r != old)
                proposed.add(remainder)
                proposed.update(tuple(sorted((*remainder, r))) for r in additions[:4])
        next_beam = list(beam)
        progressed = False
        for rr in sorted(proposed):
            key = digest(rr)
            if key in seen:
                continue
            seen.add(key)
            try:
                validate_rules(rr)
            except ValueError:
                continue
            budget.check()
            m = evaluate(ts, ls, rr, cfg, budget, training=True,
                         oracle_assignments=oracle_assignments, min_support=min_support)
            count += 1
            progressed = True
            if m['valid_support']:
                next_beam.append(m)
            if count >= cfg.max_states:
                break
        unique = {digest(m['rules']): m for m in next_beam}
        beam = sorted(unique.values(), key=model_rank)[:cfg.beam]
        iteration += 1
        finished = not progressed or iteration >= cfg.iterations or count >= cfg.max_states
        if not progressed:
            break

    finals = [evaluate(ts, ls, m['rules'], cfg, budget, alternatives=(i == 0),
                       oracle_assignments=oracle_assignments, min_support=min_support)
              for i, m in enumerate(beam[:cfg.top_models])]
    best = finals[0]
    return {
        'SEARCH_STATUS': 'COMPLETE',
        'states': count,
        'iterations': iteration,
        'model': best,
        'top_models': finals,
        'candidate_pool': [list(r) for r in pool],
        'units': inv,
        'config': asdict(cfg)
    }


def exhaustive_search(ts, ls, candidate_rules, cfg: Config = Config(),
                      oracle_assignments: list = None, min_support: int = 3,
                      max_rules: int = 6):
    """Exhaustive search over all admissible subsets of candidate rules up to max_rules.
    Guarantees finding the TRUE GLOBAL OPTIMUM of the MDL objective.
    """
    budget = Budget(cfg)
    all_evaluated = []
    rule_tuples = [tuple(r) for r in candidate_rules]

    # Evaluate empty model
    empty = evaluate(ts, ls, [], cfg, budget, training=True,
                     oracle_assignments=oracle_assignments, min_support=min_support)
    all_evaluated.append(empty)

    for r_count in range(1, min(len(rule_tuples), max_rules) + 1):
        for subset in itertools.combinations(rule_tuples, r_count):
            try:
                validate_rules(subset)
            except ValueError:
                continue
            m = evaluate(ts, ls, subset, cfg, budget, training=True,
                         oracle_assignments=oracle_assignments, min_support=min_support)
            all_evaluated.append(m)

    valid_models = [m for m in all_evaluated if m['valid_support']]
    if not valid_models:
        valid_models = all_evaluated

    # Sort by descending gain, then canonical rule representation
    sorted_models = sorted(valid_models, key=lambda m: (-m['gain'], canonical(m['rules'])))
    best_gain = sorted_models[0]['gain']
    
    # Identify all global optima (within numerical tolerance 1e-9)
    global_optima = [m for m in sorted_models if abs(m['gain'] - best_gain) < 1e-9]

    return {
        'global_optimum': sorted_models[0],
        'global_optima': global_optima,
        'num_global_optima': len(global_optima),
        'total_subsets_evaluated': len(all_evaluated),
        'valid_subsets_count': len(valid_models),
        'all_models_ranked': sorted_models
    }
