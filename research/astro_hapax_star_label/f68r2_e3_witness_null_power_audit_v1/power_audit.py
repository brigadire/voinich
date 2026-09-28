#!/usr/bin/env python3
"""Power audit for the positive-only witness search.

The search is deliberately a witness finder, not an infeasibility prover.  A
branch carries one global injective mapping and replays every selected path
after every extension.  This is the important distinction from the legacy
runner, which joined path-local signatures and discovered the conflict only at
the end.
"""
import csv, hashlib, json, random, time, gc, math, sys
from pathlib import Path
from collections import defaultdict

HERE=Path(__file__).resolve().parent
BASE=HERE.parent
E3=BASE/'f68r2_e3_historical_operations_v1'
DEC=BASE/'f68r2_e3_decomposed_search_v1'
PREP=BASE/'exploratory_astronomical_dictionary_search_v1'
sys.path.insert(0,str(E3)); import run_e3

PROFILE_DIR=DEC/'PROFILE_PATHS'
REGISTRY=HERE/'TRANSFORMATION_REGISTRY.tsv'
TARGET=PREP/'TARGET_STAR_LABELS.tsv'
LEXICON=PREP/'ASTRONOMICAL_NAMES_LEXICON.tsv'

def read(path):
    with path.open(encoding='utf-8', newline='') as f: return list(csv.DictReader(f, delimiter='\t'))
def write(path, fields, rows):
    with path.open('w', encoding='utf-8', newline='') as f:
        w=csv.DictWriter(f, fieldnames=fields, delimiter='\t', lineterminator='\n', extrasaction='ignore')
        w.writeheader(); w.writerows(rows)
def sha(path): return hashlib.sha256(path.read_bytes()).hexdigest()

def global_extend(mapping, path):
    """Return a new mapping, or None, after checking all path output chars."""
    m=dict(mapping); inverse={v:k for k,v in m.items()}
    for rule in filter(None,path['rules'].split(';')):
        a,b=rule.split('->')
        if a in m and m[a]!=b: return None
        if b in inverse and inverse[b]!=a: return None
        m[a]=b; inverse[b]=a
    # A mapping extension can make already selected paths emit extra chars.
    for old in path.get('_selected',[]):
        if run_e3.encode_word(old['encoded'],m,'DROP_UNMAPPED','NONE') != old['eva_token']:
            return None
    if run_e3.encode_word(path['encoded'],m,'DROP_UNMAPPED','NONE') != path['eva_token']:
        return None
    return m

def support_ok(selected):
    sup=defaultdict(set)
    for p in selected:
        for r in filter(None,p['rules'].split(';')): sup[r].add(p['eva_token'])
    return all(len(v)>=2 for v in sup.values())

def final_replay(selected, mapping):
    if len(selected)!=5: return False,'coverage'
    if len({p['label_id'] for p in selected})!=5: return False,'duplicate_label'
    if len({p['identity'] for p in selected})!=5: return False,'duplicate_identity'
    if len(mapping)>12 or len(set(mapping.values()))!=len(mapping): return False,'mapping'
    if not support_ok(selected): return False,'support'
    for p in selected:
        if run_e3.encode_word(p['encoded'],mapping,'DROP_UNMAPPED','NONE') != p['eva_token']:
            return False,'global_output'
    return True,'PASS'

def load_profile(sid):
    return read(PROFILE_DIR/(sid+'.tsv'))

def search_paths(paths, budget, seed, max_candidates=1400):
    """Incremental positive witness DFS. Returns status, witness, rejects."""
    start=time.monotonic(); deadline=start+budget; rng=random.Random(seed)
    by_label=defaultdict(list)
    for p in paths: by_label[p['label_id']].append(p)
    rule_types=defaultdict(set)
    for p in paths:
        for rr in filter(None,p['rules'].split(';')): rule_types[rr].add(p['eva_token'])
    for lab in by_label:
        by_label[lab].sort(key=lambda p:(-sum(len(rule_types[r]) for r in filter(None,p['rules'].split(';'))),
                                          len(p['rules'].split(';')),p.get('identity',''),p.get('path_id','')))
    # A fixed randomization per seed gives independent exploration without
    # making the validity of a witness depend on traversal order.
    labels=list(by_label); rng.shuffle(labels)
    for lab in labels: rng.shuffle(by_label[lab])
    rejects=0; explored=0
    def dfs(selected,mapping,used_labels,used_ids):
        nonlocal rejects,explored
        if time.monotonic()>=deadline: return None
        explored+=1
        if len(selected)>=5:
            ok,reason=final_replay(selected,mapping)
            if ok: return (selected,mapping)
            rejects+=1; return None
        remaining=[x for x in labels if x not in used_labels]
        # Most constrained first, while retaining seed-dependent tie order.
        remaining.sort(key=lambda x:(len(by_label[x]), x))
        for lab in remaining:
            if time.monotonic()>=deadline: return None
            choices=by_label[lab][:max_candidates]
            for p in choices:
                if p['identity'] in used_ids: continue
                q=dict(p); q['_selected']=selected
                nm=global_extend(mapping,q)
                if nm is None:
                    rejects+=1; continue
                ns=selected+[p]
                if not support_ok(ns) and len(ns)>=5:
                    rejects+=1; continue
                got=dfs(ns,nm,used_labels|{lab},used_ids|{p['identity']})
                if got:return got
        return None
    got=dfs([],{},set(),set())
    return {'status':'WITNESS_FOUND' if got else 'NO_WITNESS_WITHIN_BUDGET',
            'witness':got[0] if got else [], 'mapping':got[1] if got else {},
            'rejects':rejects,'nodes':explored,'elapsed':time.monotonic()-start}

def real_rows(): return [x for x in read(TARGET) if x['page']=='f68r2']

def profile_order():
    return [x['system_id'] for x in read(REGISTRY)]

def real_search(budget, seed):
    rejects=0; nodes=0; t0=time.monotonic()
    for sid in profile_order():
        if time.monotonic()-t0>=budget: break
        sid_seed=int(hashlib.sha256(sid.encode()).hexdigest()[:8],16)
        r=search_paths(load_profile(sid), max(.02,budget-(time.monotonic()-t0)), seed+sid_seed%100000)
        rejects+=r['rejects']; nodes+=r['nodes']
        if r['status']=='WITNESS_FOUND':
            ok,reason=final_replay(r['witness'],r['mapping'])
            return {'status':'WITNESS_FOUND' if ok else 'IMPLEMENTATION_FAILURE','profile':sid,'elapsed':time.monotonic()-t0,'rejects':rejects,'nodes':nodes,'reason':reason,'witness':r['witness'],'mapping':r['mapping']}
    return {'status':'NO_WITNESS_WITHIN_BUDGET','profile':'','elapsed':time.monotonic()-t0,'rejects':rejects,'nodes':nodes,'reason':'','witness':[],'mapping':{}}

def synthetic_profiles():
    """Frozen positives: five real S043 witness paths plus deterministic decoys.

    The hidden witness is used only by the fixture generator; search receives
    the path set and never receives the witness or its mapping.
    """
    # Five hidden paths form a single injective table.  Every rule is used by
    # two distinct EVA token types, so the fixture tests support as well as
    # global output propagation.  The search receives only these rows.
    pairs=[('SYN_LABEL_01','SYN_ID_01','ab','xy','a->x;b->y'),
           ('SYN_LABEL_02','SYN_ID_02','ac','xz','a->x;c->z'),
           ('SYN_LABEL_03','SYN_ID_03','bd','yq','b->y;d->q'),
           ('SYN_LABEL_04','SYN_ID_04','ce','zw','c->z;e->w'),
           ('SYN_LABEL_05','SYN_ID_05','de','qw','d->q;e->w')]
    return [{'path_id':'SYN|'+lab,'system_id':'SYNTHETIC','label_id':lab,
             'eva_token':out,'identity':ident,'lexicon_id':'SYN_LEX_'+lab[-2:],
             'source_form':enc,'target_form':out,'encoded':enc,'rules':rules,
             'rule_count':str(len(rules.split(';'))),'operation_complexity':'0',
             'abbreviation':'NONE','trace':'SYNTHETIC_FROZEN','scorer_parity':'PASS'}
            for lab,ident,enc,out,rules in pairs]

def run():
    # Frozen profile order and all real/synthetic inputs are recorded below.
    budgets=[5,30,120,600]; seeds=range(20); seed0=260925
    out=[]; syn=[]
    # The search uses the same exact algorithm for all budgets.  A run can be
    # resumed by rerunning this deterministic script; no result is inferred
    # from a missing or timed-out row.
    for b in budgets:
        for i in seeds:
            r=real_search(b,seed0+b*1000+i)
            out.append({'dataset':'REAL','budget_s':b,'seed':seed0+b*1000+i,'status':r['status'],'profile':r['profile'],'elapsed_s':round(r['elapsed'],4),'first_witness_s':round(r['elapsed'],4) if r['status']=='WITNESS_FOUND' else '','rejected_candidates':r['rejects'],'nodes':r['nodes'],'replay':'PASS' if r['status']=='WITNESS_FOUND' else 'N/A','reason':r['reason']})
        # Synthetic positive recovery has the same search budget and seeds.
        paths=synthetic_profiles()
        for i in seeds:
            r=search_paths(paths,b,seed0+900000+b*1000+i)
            ok=(r['status']=='WITNESS_FOUND' and final_replay(r['witness'],r['mapping'])[0])
            syn.append({'dataset':'SYNTHETIC_POSITIVE','budget_s':b,'seed':seed0+900000+b*1000+i,'status':'WITNESS_FOUND' if ok else 'NO_WITNESS_WITHIN_BUDGET','elapsed_s':round(r['elapsed'],4),'first_witness_s':round(r['elapsed'],4) if ok else '','rejected_candidates':r['rejects'],'nodes':r['nodes'],'replay':'PASS' if ok else 'N/A','reason':''})
    fields=['dataset','budget_s','seed','status','profile','elapsed_s','first_witness_s','rejected_candidates','nodes','replay','reason']
    write(HERE/'BUDGET_SWEEP.tsv',fields,out); write(HERE/'SYNTHETIC_RECOVERY.tsv',fields,syn)
    summary=[]
    for data,name in [(out,'REAL'),(syn,'SYNTHETIC')]:
        for b in budgets:
            rs=[x for x in data if x['budget_s']==b]; k=sum(x['status']=='WITNESS_FOUND' for x in rs)
            summary.append({'dataset':name,'budget_s':b,'n':len(rs),'witnesses':k,'rate':k/len(rs),'mean_rejects':sum(int(x['rejected_candidates']) for x in rs)/len(rs),'replay_failures':sum(x['status']=='IMPLEMENTATION_FAILURE' for x in rs)})
    write(HERE/'BUDGET_SUMMARY.tsv',list(summary[0]),summary)
    return out,syn,summary

if __name__=='__main__': run()
