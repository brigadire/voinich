#!/usr/bin/env python3
import csv, hashlib, json, re, time, sys
from collections import defaultdict
from pathlib import Path
from ortools.sat.python import cp_model

ROOT = Path(__file__).resolve().parent
BASE = ROOT.parent
AUDIT = BASE / 'exploratory_astronomical_dictionary_search_e2_structural_support_audit_v1'
sys.path.insert(0, str(AUDIT)); from audit import alignments as frozen_alignments
sys.path.insert(0, str(AUDIT / 'snapshot')); from scorer import encode_word
PREP = BASE / 'exploratory_astronomical_dictionary_search_v1'
LEX = PREP / 'ASTRONOMICAL_NAMES_LEXICON.tsv'
SCOPE = PREP / 'TARGET_STAR_LABELS.tsv'
REGISTRY = ROOT / 'TRANSFORMATION_REGISTRY.tsv'

ENDINGS = ('ibus','orum','arum','ium','ius','ae','is','us','um','ii','am','em','as','es','os','i','o','a','e')

def rows(path):
    with path.open(encoding='utf-8', newline='') as f:
        return list(csv.DictReader(f, delimiter='\t'))

def write(name, fields, data):
    with (ROOT / name).open('w', encoding='utf-8', newline='') as f:
        w = csv.DictWriter(f, fieldnames=fields, delimiter='\t', lineterminator='\n', extrasaction='ignore')
        w.writeheader(); w.writerows(data)

def sha(path):
    return hashlib.sha256(path.read_bytes()).hexdigest()

def normalize_words(form):
    return re.findall('[a-z]+', form.lower())

def transform(form, rule):
    words = normalize_words(form)
    steps = [('INPUT', ' '.join(words))]
    if rule['article'] == 'DROP_AL':
        kept = [w for w in words if w != 'al'] or words
        words = kept
        steps.append(('ARTICLE_DROP_AL', ''.join(words)))
    else:
        steps.append(('ARTICLE_KEEP', ''.join(words)))
    s = ''.join(words)
    if rule['orthography'] == 'IJ_UV':
        s = s.translate(str.maketrans({'j':'i', 'v':'u'})); steps.append(('IJ_UV', s))
    elif rule['orthography'] in ('ARABIC_LATIN', 'VELAR_COLLAPSE'):
        for a, b in (('kh','h'),('gh','g'),('sh','s'),('th','t'),('dh','d')):
            s = s.replace(a, b)
        s = s.translate(str.maketrans({'j':'i', 'w':'u'})); steps.append((rule['orthography'], s))
    else:
        steps.append(('IDENTITY', s))
    if rule['orthography'] == 'VELAR_COLLAPSE':
        s = s.translate(str.maketrans({'q':'k', 'c':'k'})); steps.append(('VELAR_COLLAPSE', s))
    if rule['vowel'] == 'CONTRACT_INTERNAL' and len(s) > 2:
        s = s[0] + re.sub('[aeiouy]', '', s[1:-1]) + s[-1]
        steps.append(('CONTRACT_INTERNAL', s))
    else:
        steps.append(('VOWELS_KEEP', s))
    abbr = rule['abbreviation']
    if abbr == 'SUSPEND_1' and len(s) > 3:
        s = s[:-1]; steps.append(('SUSPEND_1', s))
    elif abbr == 'PREFIX_4':
        s = s[:4]; steps.append(('PREFIX_4', s))
    elif abbr == 'STRIP_LATIN':
        for ending in ENDINGS:
            if s.endswith(ending) and len(s) - len(ending) >= 3:
                s = s[:-len(ending)]; steps.append(('STRIP_LATIN:' + ending, s)); break
        else:
            steps.append(('STRIP_LATIN:none', s))
    else:
        steps.append(('ABBREVIATION_NONE', s))
    return s, steps

def alignments(source, target):
    """Enumerate mappings under the frozen DROP_UNMAPPED scorer, capped at five."""
    out = set()
    def visit(i, j, forward, reverse):
        if i == len(source):
            if j == len(target): out.add(tuple(sorted(forward.items())))
            return
        ch = source[i]
        if ch in forward:
            if j < len(target) and forward[ch] == target[j]: visit(i+1, j+1, forward, reverse)
            return
        visit(i+1, j, forward, reverse)
        if j < len(target) and len(forward) < 5:
            dst = target[j]
            if dst not in reverse:
                nf, nr = dict(forward), dict(reverse); nf[ch] = dst; nr[dst] = ch
                visit(i+1, j+1, nf, nr)
    visit(0, 0, {}, {})
    return sorted(out)

def trace_alignment(source, target, rules):
    mapping = dict(x.split('->') for x in rules); ti = 0; kept = []; dropped = []
    for i, ch in enumerate(source):
        if ti < len(target) and ch in mapping and mapping[ch] == target[ti]:
            kept.append(i); ti += 1
        else:
            dropped.append(i)
    return ','.join(map(str, dropped)), ''.join(source[i] for i in dropped), kept

def op_cost(r):
    return int(r['article'] != 'KEEP') + int(r['orthography'] != 'IDENTITY') + int(r['vowel'] != 'KEEP') + {'NONE':0,'SUSPEND_1':1,'PREFIX_4':2,'STRIP_LATIN':1}[r['abbreviation']]

def build_paths(labels, lexicon, registry):
    out = {}
    alignment_cache = {}
    transform_cache = {}
    for rule in registry:
        transformed = {}
        for term in lexicon:
            key = (term['lexicon_id'], rule['system_id'])
            transformed[term['lexicon_id']] = transform(term['normalized_form'], rule)
        for label in labels:
            for term in lexicon:
                encoded, steps = transformed[term['lexicon_id']]
                if len(label['zl3b_token']) > len(encoded) or len(set(label['zl3b_token'])) > len(set(encoded)):
                    continue
                cache_key = (encoded, label['zl3b_token'])
                if cache_key not in alignment_cache:
                    alignment_cache[cache_key] = [tuple(f'{a}->{b}' for a,b in mp) for mp in alignments(encoded, label['zl3b_token'])]
                matches = alignment_cache[cache_key]
                if not matches: continue
                for path_no, rules in enumerate(matches):
                    mapping = dict(x.split('->') for x in rules)
                    if encode_word(encoded, mapping, 'DROP_UNMAPPED', 'NONE') != label['zl3b_token']:
                        continue
                    dropped_positions, dropped, kept = trace_alignment(encoded, label['zl3b_token'], rules)
                    key = (rule['system_id'], label['label_id'], term['canonical_identity_id'], ';'.join(rules), label['zl3b_token'])
                    candidate = {
                    'path_id': f"{rule['system_id']}|{label['label_id']}|{term['lexicon_id']}|{hashlib.sha1((';'.join(rules)).encode()).hexdigest()[:10]}",
                    'system_id': rule['system_id'], 'label_id': label['label_id'], 'eva_token': label['zl3b_token'],
                    'identity': term['canonical_identity_id'], 'lexicon_id': term['lexicon_id'],
                    'source_form': term['normalized_form'], 'target_form': label['zl3b_token'],
                    'encoded': encoded, 'rules': ';'.join(rules), 'rule_count': len(rules),
                    'operation_complexity': op_cost(rule), 'abbreviation': rule['abbreviation'],
                    'trace': ' > '.join(f'{name}:{value}' for name, value in steps) + f' > DROP_UNMAPPED:{dropped or "NONE"} > OUTPUT:{label["zl3b_token"]}',
                    'dropped_source_positions': dropped_positions,
                    'alternative_count': 1,
                    'scorer_parity': 'PASS'
                    }
                    if key in out:
                        out[key]['alternative_count'] += 1
                        out[key]['alternative_sources'] += ' || ' + term['lexicon_id'] + ':' + term['normalized_form']
                    else:
                        candidate['alternative_sources'] = term['lexicon_id'] + ':' + term['normalized_form']
                        out[key] = candidate
    return list(out.values())

def solve(paths, k, seconds=300):
    if not paths:
        return {'status':'INFEASIBLE_CERTIFIED','coverage':0,'objective':0,'best_bound':0,'gap':0,'elapsed_sec':0,'chosen':[],'rules':[],'tokens':[],'system_id':''}
    m = cp_model.CpModel(); x = [m.NewBoolVar(f'p{i}') for i in range(len(paths))]
    rules = sorted({r for p in paths for r in p['rules'].split(';')}); ri = {r:i for i,r in enumerate(rules)}
    toks = sorted({p['eva_token'] for p in paths}); ti = {t:i for i,t in enumerate(toks)}
    y = [m.NewBoolVar(f'r{i}') for i in range(len(rules))]; u = [m.NewBoolVar(f'u{i}') for i in range(len(toks))]
    z = {(r,t):m.NewBoolVar(f'z{r}_{t}') for r in range(len(rules)) for t in range(len(toks))}
    by_label = defaultdict(list); by_id = defaultdict(list); by_rule = defaultdict(list); by_rule_tok = defaultdict(list)
    for i, p in enumerate(paths):
        by_label[p['label_id']].append(x[i]); by_id[p['identity']].append(x[i]); m.AddImplication(x[i], u[ti[p['eva_token']]])
        for rr in p['rules'].split(';'):
            r = ri[rr]; by_rule[r].append(x[i]); by_rule_tok[r,ti[p['eva_token']]].append(x[i])
            m.AddImplication(x[i], y[r]); m.AddImplication(x[i], z[r,ti[p['eva_token']]])
    for vs in by_label.values(): m.Add(sum(vs) <= 1)
    for vs in by_id.values(): m.Add(sum(vs) <= 1)
    for r, vs in by_rule.items():
        m.Add(y[r] <= sum(vs)); m.Add(sum(z[r,t] for t in range(len(toks))) >= 2*y[r])
        for t in range(len(toks)): m.Add(z[r,t] <= sum(by_rule_tok.get((r,t), [])))
    for a in sorted({r.split('->')[0] for r in rules}): m.Add(sum(y[i] for i,r in enumerate(rules) if r.split('->')[0] == a) <= 1)
    for b in sorted({r.split('->')[1] for r in rules}): m.Add(sum(y[i] for i,r in enumerate(rules) if r.split('->')[1] == b) <= 1)
    m.Add(sum(y) <= k)
    m.Maximize(1000000*sum(x) + 1000*sum(u) - 10*sum(y))
    s = cp_model.CpSolver(); s.parameters.max_time_in_seconds = seconds; s.parameters.num_search_workers = 1; s.parameters.random_seed = 6803
    t0 = time.monotonic(); st = s.Solve(m)
    chosen = [i for i,v in enumerate(x) if s.Value(v)] if st in (cp_model.OPTIMAL, cp_model.FEASIBLE) else []
    return {'status':s.StatusName(st), 'coverage':len(chosen), 'objective':s.ObjectiveValue(), 'best_bound':s.BestObjectiveBound(), 'gap':0 if st == cp_model.OPTIMAL else 'UNKNOWN', 'elapsed_sec':round(time.monotonic()-t0,4), 'chosen':[paths[i] for i in chosen], 'rules':rules, 'tokens':toks, 'system_id':paths[0]['system_id']}

def main():
    labels = [r for r in rows(SCOPE) if r['page'] == 'f68r2']
    assert len(labels) == 27 and all(r['page'] == 'f68r2' for r in labels)
    lexicon, registry = rows(LEX), rows(REGISTRY)
    paths = build_paths(labels, lexicon, registry)
    write('E3_PATH_GRAPH.tsv', list(paths[0]), paths)
    results = []
    by_system = defaultdict(list)
    for p in paths: by_system[p['system_id']].append(p)
    for rule in registry:
        sys_paths = by_system[rule['system_id']]
        r12 = solve(sys_paths, 12); r12.update(system_id=rule['system_id'], operation_complexity=op_cost(rule), abbreviation_applied=int(rule['abbreviation'] != 'NONE'), mapping_count=len({z for p in r12['chosen'] for z in p['rules'].split(';')}), k=12)
        results.append(r12)
        if r12['coverage'] > 0:
            for k in range(1, 12):
                r = solve(sys_paths, k); r.update(system_id=rule['system_id'], operation_complexity=op_cost(rule), abbreviation_applied=int(rule['abbreviation'] != 'NONE'), mapping_count=len({z for p in r['chosen'] for z in p['rules'].split(';')}), k=k)
                results.append(r)
        else:
            for k in range(1, 12):
                results.append({'system_id':rule['system_id'],'k':k,'status':'INFEASIBLE_CERTIFIED_BY_K12_ZERO','coverage':0,'best_bound':0,'gap':0,'elapsed_sec':0,'mapping_count':0,'abbreviation_applied':int(rule['abbreviation'] != 'NONE'),'operation_complexity':op_cost(rule),'objective':0,'chosen':[],'rules':[],'tokens':[]})
    best_by_k = []
    for k in range(1,13):
        candidates = [r for r in results if r['k'] == k]
        best_by_k.append(max(candidates, key=lambda r:(r['coverage'], len({p['eva_token'] for p in r['chosen']}), len({p['identity'] for p in r['chosen']}), -r['mapping_count'], -r['abbreviation_applied'], -r['operation_complexity'], r['system_id'])))
    curve_fields = ['k','system_id','status','coverage','best_bound','gap','elapsed_sec','mapping_count','abbreviation_applied','operation_complexity','objective']
    write('E3_COVERAGE_BY_COMPLEXITY.tsv', curve_fields, best_by_k)
    write('E3_SYSTEM_RESULTS.tsv', ['system_id','k','status','coverage','best_bound','gap','elapsed_sec','mapping_count','abbreviation_applied','operation_complexity','objective'], results)
    final = best_by_k[-1]
    assignments = [dict(p, k=final['k'], system_id=final['system_id'], status=final['status']) for p in final['chosen']]
    write('E3_MATCH_ASSIGNMENTS.tsv', list(assignments[0]) if assignments else ['k','system_id','status'], assignments)
    trace = [dict(p, k=final['k'], selected='YES') for p in final['chosen']]
    write('E3_TRANSFORMATION_TRACE.tsv', list(trace[0]) if trace else ['k','selected'], trace)
    support = defaultdict(set)
    for p in final['chosen']:
        for rule in p['rules'].split(';'): support[rule].add(p['eva_token'])
    write('E3_RULE_SUPPORT.tsv', ['rule','distinct_eva_types','support_types','status'], [{'rule':r,'distinct_eva_types':len(v),'support_types':','.join(sorted(v)),'status':'PASS' if len(v)>=2 else 'FAIL'} for r,v in sorted(support.items())])
    write('E3_EXISTENCE_TESTS.tsv', ['threshold','status','coverage','best_bound','gap'], [{'threshold':t,'status':'FEASIBLE' if final['coverage']>=t else 'INFEASIBLE_CERTIFIED' if final['status']=='OPTIMAL' else 'UNKNOWN_TIMEOUT','coverage':final['coverage'],'best_bound':final['best_bound'],'gap':final['gap']} for t in (6,8,11)])
    status = {'E3_STATUS':'E3_MAXIMUM_CERTIFIED' if final['status']=='OPTIMAL' else 'E3_TIME_BOUNDED','F68R1_USED':'NO','GROUP_CROSSWALK_USED':'NO','PATH_COUNT':len(paths),'SYSTEM_COUNT':len(registry),'MAXIMUM_COVERAGE':final['coverage'],'THRESHOLDS_6_8_11':'FEASIBLE' if final['coverage']>=6 else 'INFEASIBLE_CERTIFIED','NULL_CONTROL':'REQUIRED' if final['coverage']>=6 else 'NOT_RUN_THRESHOLD_6_NOT_REACHED','SCIENTIFIC_CLAIM':'NONE','RESULTS_INTERPRETATION':'WITHHELD_UNTIL_NULL'}
    (ROOT/'RUN_STATUS.json').write_text(json.dumps(status, indent=2, sort_keys=True)+'\n')
    freeze = {'scope_sha256':sha(SCOPE),'scope_filter':'page == f68r2','scope_count':len(labels),'lexicon_sha256':sha(LEX),'registry_sha256':sha(REGISTRY),'scorer':'exact character-by-character injective alignment; DROP_UNMAPPED; one global operation system','operation_count':len(registry),'mapping_limits':'1..12','f68r1_used':'NO','group_crosswalk_used':'NO'}
    (ROOT/'INPUT_FREEZE.json').write_text(json.dumps(freeze, indent=2, sort_keys=True)+'\n')
    print(json.dumps(status, sort_keys=True))

if __name__ == '__main__': main()
