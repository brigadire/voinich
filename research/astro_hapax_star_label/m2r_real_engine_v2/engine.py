"""Independent surface-only M2R-v2 inference; no generator or truth dependency."""
from __future__ import annotations
import hashlib
import json
import math
import os
import resource
import time
import unicodedata
from collections import Counter, defaultdict
from dataclasses import asdict, dataclass
from functools import lru_cache
from pathlib import Path


def canonical(value):
    return json.dumps(value, ensure_ascii=False, sort_keys=True, separators=(',', ':'))


def digest(value):
    return hashlib.sha256(canonical(value).encode()).hexdigest()


def literal(s):
    return len(s.encode('utf-8')) * 8


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
    seconds: float = 120
    memory_mb: int = 512
    max_examples: int = 80
    max_rows: int = 256
    max_length: int = 32
    max_units: int = 12000

    def __post_init__(self):
        if any(v <= 0 for v in asdict(self).values()):
            raise ValueError('resource bounds must be positive')


class Limit(Exception):
    pass


class Budget:
    def __init__(self, cfg, elapsed=0):
        self.cfg, self.start, self.elapsed = cfg, time.monotonic(), elapsed

    def check(self):
        if self.elapsed + time.monotonic() - self.start >= self.cfg.seconds:
            raise Limit('wall_clock')
        if resource.getrusage(resource.RUSAGE_SELF).ru_maxrss > self.cfg.memory_mb * 1024:
            raise Limit('memory')


def load_rows(rows, cfg=Config()):
    """Identical neutral loader for every term/label/lexicon family."""
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


def role(i, j, n):
    if i == 0 and j == n:
        return 'WHOLE'
    if i == 0:
        return 'INITIAL'
    if j == n:
        return 'FINAL'
    return 'MEDIAL'


def units(strings, side, cfg=Config(), budget=None):
    found = {}
    for s in strings:
        if budget:
            budget.check()
        spans = {(i, j) for i in range(len(s)) for j in range(i + 1, min(len(s), i + 4) + 1)}
        spans.add((0, len(s)))
        for i, j in sorted(spans):
            key = (role(i, j, len(s)), s[i:j])
            if key not in found:
                found[key] = {'id': digest([side, *key]), 'side': side, 'role': key[0],
                              'surface': key[1], 'length': j-i, 'provenance': []}
            found[key]['provenance'].append([digest(s), i, j])
        if len(found) > cfg.max_units:
            raise Limit('candidate_units')
    for u in found.values():
        u['occurrence_support'] = len(u['provenance'])
        u['independent_example_support'] = len({p[0] for p in u['provenance']})
    return sorted(found.values(), key=lambda u: (u['role'], u['surface']))


def rule_key(r):
    return tuple(r)


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
def positions(s, ro, unit):
    return [(i, i+len(unit)) for i in range(len(s)-len(unit)+1)
            if s.startswith(unit, i) and role(i, i+len(unit), len(s)) == ro]


def candidates(ts, ls, cfg, budget):
    su, tu = units(ts, 'source', cfg, budget), units(ls, 'target', cfg, budget)
    ranked = []
    for a in su:
        budget.check()
        na = a['independent_example_support']
        if na < 3:
            continue
        for b in tu:
            nb = b['independent_example_support']
            if nb < 3 or a['role'] != b['role']:
                continue
            bits = literal(a['surface']) + literal(b['surface'])
            potential = min(na, nb) * (bits-2) - (bits+10)
            ranked.append((-potential, abs(na-nb), (a['role'], a['surface'], b['surface'])))
    return [r[2] for r in sorted(ranked)[:cfg.pool]], {'source': su, 'target': tu}


def align(a, b, rules, k=2, budget=None):
    """Top-k DAG paths: same-role substitutions/compositions and explicit gaps."""
    if budget:
        budget.check()
    relevant = tuple(r for r in map(tuple, rules) if positions(a,r[0],r[1]) and positions(b,r[0],r[2]))
    return _align_cached(a,b,relevant,len(rules),k)


@lru_cache(maxsize=1024)
def _align_cached(a, b, rules, model_size, k):
    ref = 1 + math.ceil(math.log2(model_size+1))
    ops = defaultdict(list)
    for ro, x, y in rules:
        for i, ii in positions(a, ro, x):
            for j, jj in positions(b, ro, y):
                ops[i, j].append((ii, jj, ref, ('R', i, ii, j, jj, ro, x, y)))
    cells = {(0, 0): [(0, ())]}
    for i in range(len(a)+1):
        for j in range(len(b)+1):
            paths = cells.get((i, j), ())
            if not paths:
                continue
            edges = list(ops.get((i, j), ()))
            if i < len(a):
                edges.append((i+1, j, literal(a[i])+1, ('S', i, i+1, j, j)))
            if j < len(b):
                edges.append((i, j+1, literal(b[j])+1, ('T', i, i, j, j+1)))
            for ii, jj, cost, op in edges:
                dest = cells.setdefault((ii, jj), [])
                dest.extend((c+cost, p+(op,)) for c, p in paths)
                # Different interleavings of the same gaps are equivalent.
                unique = {}
                for c, p in dest:
                    signature = tuple(z for z in p if z[0] == 'R')
                    old = unique.get(signature)
                    if old is None or (c, p) < old:
                        unique[signature] = (c, p)
                cells[ii, jj] = sorted(unique.values())[:k]
    out = []
    for cost, path in cells[len(a), len(b)]:
        trace = [list(p) for p in path]
        covered = sum(p[4]-p[3] for p in path if p[0] == 'R')
        gaps = [p for p in path if p[0] != 'R']
        out.append({'cost': cost, 'path': trace, 'explained_target': covered,
                    'source_segmentation': [a[p[1]:p[2]] for p in path if p[2]>p[1]],
                    'target_segmentation': [b[p[3]:p[4]] for p in path if p[4]>p[3]],
                    'candidate_rules': [list(p[5:]) for p in path if p[0]=='R'],
                    'unexplained_units': [list(p) for p in gaps],
                    'local_complexity': len({tuple(p[5:]) for p in path if p[0]=='R'}),
                    'local_score': literal(a)+literal(b)-cost,
                    'coverage': covered/len(b), 'full': not gaps})
    return out


def transform(s, rules):
    validate_rules(rules)
    outputs = {0: {''}}
    for i in range(len(s)):
        for ro, a, b in rules:
            j = i+len(a)
            if s.startswith(a, i) and role(i, j, len(s)) == ro:
                outputs.setdefault(j, set()).update(x+b for x in outputs.get(i, ()))
    values = sorted(outputs.get(len(s), ()))
    return {'status': 'AMBIGUOUS' if len(values)>1 else 'UNIQUE' if values else 'UNEXPLAINED',
            'outputs': values}


def hungarian(weights, budget=None, forbidden=None):
    """Exact rectangular maximum-weight matching; dummy columns allow unmatched."""
    n = len(weights)
    if not n:
        return [], 0.0
    width = len(weights[0])
    m = width+n
    u, v, p, way = [0.]*(n+1), [0.]*(m+1), [0]*(m+1), [0]*(m+1)
    for i in range(1, n+1):
        if budget:
            budget.check()
        p[0], j0 = i, 0
        dist, used = [float('inf')]*(m+1), [False]*(m+1)
        while True:
            used[j0] = True
            i0, delta, j1 = p[j0], float('inf'), 0
            for j in range(1, m+1):
                if used[j]:
                    continue
                w = weights[i0-1][j-1] if j <= width else 0.
                if forbidden == (i0-1, j-1):
                    w = -1e12
                cur = -w-u[i0]-v[j]
                if cur < dist[j]:
                    dist[j], way[j] = cur, j0
                if dist[j] < delta:
                    delta, j1 = dist[j], j
            for j in range(m+1):
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
    pairs = sorted((p[j]-1, j-1) for j in range(1, width+1)
                   if p[j] and weights[p[j]-1][j-1] > 0 and forbidden != (p[j]-1, j-1))
    return pairs, sum(weights[i][j] for i, j in pairs)


def support(rules, assignments):
    counts = {tuple(r): {'occurrence_support': 0, 'pairs': set(), 'terms': set(), 'labels': set()}
              for r in rules}
    for record in assignments:
        a, b = record['term'], record['label']
        for op in record['alignment']['path']:
            if op[0] != 'R':
                continue
            c = counts[tuple(op[5:])]
            c['occurrence_support'] += 1
            c['pairs'].add((a, b))
            c['terms'].add(a)
            c['labels'].add(b)
    return [{'rule': list(r), 'occurrence_support': c['occurrence_support'],
             'pair_support': len(c['pairs']), 'distinct_term_support': len(c['terms']),
             'distinct_label_support': len(c['labels'])} for r, c in sorted(counts.items())]


def evaluate(ts, ls, rules, cfg=Config(), budget=None, training=True, alternatives=False):
    rules = sorted(map(tuple, rules))
    validate_rules(rules)
    baseline = sum(map(literal, ts))+sum(map(literal, ls))
    assignment_bits = math.log2(len(ts)+1)+math.log2(len(ls)+1)
    paths, weights = {}, []
    for i, a in enumerate(ts):
        if budget:
            budget.check()
        row = []
        for j, b in enumerate(ls):
            # Sparse matching graph: absent rule overlap cannot produce positive saving.
            relevant = [r for r in rules if positions(a, r[0], r[1]) and positions(b, r[0], r[2])]
            if not relevant:
                row.append(0.)
                continue
            pp = align(a, b, rules, cfg.top_paths, budget)
            paths[i, j] = pp
            row.append(max(0., literal(a)+literal(b)-pp[0]['cost']-assignment_bits))
        weights.append(row)
    pairs, optimum = hungarian(weights, budget)
    assigned = [{'term': ts[i], 'label': ls[j], 'alignment': paths[i, j][0]}
                for i, j in pairs]
    sup = support(rules, assigned)
    valid = all(min(x['pair_support'], x['distinct_term_support'], x['distinct_label_support'])>=3 for x in sup)
    components = {'RULE_CODE_LENGTH': sum(10+literal(r[1])+literal(r[2]) for r in rules),
                  'ASSIGNMENT_CODE_LENGTH': len(pairs)*assignment_bits,
                  'EXPLAINED_ALIGNMENT_COST': 0, 'UNEXPLAINED_SOURCE_COST': 0,
                  'UNEXPLAINED_TARGET_COST': 0, 'EXCEPTION_COST': 0}
    used_t, used_l = {i for i, j in pairs}, {j for i, j in pairs}
    components['UNEXPLAINED_SOURCE_COST'] = sum(literal(t) for i, t in enumerate(ts) if i not in used_t)
    components['UNEXPLAINED_TARGET_COST'] = sum(literal(t) for j, t in enumerate(ls) if j not in used_l)
    ref = 1+math.ceil(math.log2(len(rules)+1))
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
    gain = (baseline-total)/baseline if baseline else 0.
    margins, alts = [], []
    if alternatives:
        for i, j in pairs:
            alternative, score = hungarian(weights, budget, (i, j))
            margin = max(0., optimum-score)
            margins.append({'term': ts[i], 'label': ls[j], 'margin': margin})
            if margin < 1e-9:
                alt = [[ts[x], ls[y]] for x, y in alternative]
                if alt not in alts and len(alts) < cfg.top_models:
                    alts.append(alt)
    pairwise = []
    if alternatives:
        for i,a in enumerate(ts):
            for j,b in enumerate(ls):
                pp = paths.get((i,j)) or align(a,b,rules,cfg.top_paths,budget)
                pairwise.append({'term': a, 'label': b, 'weight': weights[i][j], 'explanations': pp})
    return {'pairwise': pairwise, 'rules': [list(r) for r in rules], 'support': sup, 'valid_support': valid,
            'assignments': assigned, 'components': components, 'baseline': baseline,
            'total': total, 'gain': gain, 'objective': gain if valid or not training else None,
            'coverage': sum(a['alignment']['full'] for a in assigned)/len(ls) if ls else 0.,
            'character_coverage': sum(a['alignment']['explained_target'] for a in assigned)/sum(map(len, ls)) if ls else 0.,
            'margins': margins, 'alternative_optima': alts,
            'unmatched_terms': [t for i, t in enumerate(ts) if i not in used_t],
            'unmatched_labels': [s for j, s in enumerate(ls) if j not in used_l]}


def atomic_json(path, value):
    path = Path(path)
    tmp = path.with_name(path.name+'.tmp')
    with tmp.open('w', encoding='utf-8') as f:
        f.write(canonical(value)+'\n')
        f.flush()
        os.fsync(f.fileno())
    os.replace(tmp, path)
    fd = os.open(path.parent, os.O_DIRECTORY)
    try:
        os.fsync(fd)
    finally:
        os.close(fd)


def checkpoint_write(path, payload):
    atomic_json(path, {'sha256': digest(payload), 'payload': payload})


def checkpoint_read(path):
    obj = json.loads(Path(path).read_text())
    if obj['sha256'] != digest(obj['payload']):
        raise ValueError('checkpoint checksum mismatch')
    return obj['payload']


def model_rank(m):
    return (-m['gain'], canonical(m['rules']))


def search(term_rows, label_rows, cfg=Config(), checkpoint=None, resume=False, pause_after=None):
    budget = Budget(cfg)
    try:
        ts, _ = load_rows(term_rows, cfg)
        ls, _ = load_rows(label_rows, cfg)
        fingerprint = digest([ts, ls, asdict(cfg), hashlib.sha256(Path(__file__).read_bytes()).hexdigest()])
        pool, inventory = candidates(ts, ls, cfg, budget)
        empty = evaluate(ts, ls, [], cfg, budget)
        beam, seen, count, iteration = [empty], {digest([])}, 1, 0
        finished = False
        if resume:
            state = checkpoint_read(checkpoint)
            if state['fingerprint'] != fingerprint:
                raise ValueError('checkpoint engine/config/input mismatch')
            budget.elapsed = state['elapsed']
            beam = [evaluate(ts, ls, r, cfg, budget) for r in state['beam']]
            seen, count, iteration = set(state['seen']), state['count'], state['iteration']
            finished = state['finished']
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
                m = evaluate(ts, ls, rr, cfg, budget)
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
            if checkpoint:
                checkpoint_write(checkpoint, {'fingerprint': fingerprint, 'beam': [m['rules'] for m in beam],
                    'seen': sorted(seen), 'count': count, 'iteration': iteration, 'finished': finished,
                    'elapsed': budget.elapsed+time.monotonic()-budget.start})
            if pause_after is not None and iteration >= pause_after:
                return {'SEARCH_STATUS': 'INCOMPLETE', 'reason': 'checkpoint_pause'}
            if not progressed:
                break
        finals = [evaluate(ts, ls, m['rules'], cfg, budget, alternatives=(i==0)) for i,m in enumerate(beam[:cfg.top_models])]
        best = finals[0]
        return {'SEARCH_STATUS': 'COMPLETE', 'data_status': 'EMPTY' if not ts or not ls else 'SMALL' if min(len(ts), len(ls))<3 else 'OK',
                'SEARCH_TYPE': 'BOUNDED_DETERMINISTIC_OPTIMIZATION', 'GLOBAL_OPTIMUM_CLAIMED': 'NO',
                'states': count, 'iterations': iteration, 'model': best, 'top_models': finals,
                'candidate_pool': [list(r) for r in pool], 'units': inventory,
                'input_hash': digest([ts, ls]), 'config': asdict(cfg)}
    except (Limit, MemoryError) as e:
        return {'SEARCH_STATUS': 'INCOMPLETE', 'reason': str(e) or 'memory'}


def heldout(train, train_terms, train_labels, terms, labels, cfg=Config()):
    """Apply trained rules, charging the same model complexity without refitting."""
    if train['SEARCH_STATUS'] != 'COMPLETE':
        return {'SEARCH_STATUS': 'INCOMPLETE', 'reason': 'train_incomplete'}
    budget = Budget(cfg)
    try:
        tt, _ = load_rows(train_terms, cfg)
        tl, _ = load_rows(train_labels, cfg)
        ht, _ = load_rows(terms, cfg)
        hl, _ = load_rows(labels, cfg)
        if set(tt)&set(ht) or set(tl)&set(hl):
            raise ValueError('train/heldout instance overlap')
        result = evaluate(ht, hl, train['model']['rules'], cfg, budget, training=False, alternatives=True)
        return {'SEARCH_STATUS': 'COMPLETE', 'model': result}
    except (Limit, MemoryError) as e:
        return {'SEARCH_STATUS': 'INCOMPLETE', 'reason': str(e) or 'memory'}


def accepted(train, test):
    if train['SEARCH_STATUS'] != 'COMPLETE' or test['SEARCH_STATUS'] != 'COMPLETE':
        return False
    a, b = train['model'], test['model']
    return bool(a['rules'] and a['valid_support'] and a['gain']>0 and b['gain']>0
                and min(a['coverage'], b['coverage'])>=.70
                and b['coverage']/max(a['coverage'], 1e-12)>=.70
                and not a['alternative_optima'] and not b['alternative_optima'])
