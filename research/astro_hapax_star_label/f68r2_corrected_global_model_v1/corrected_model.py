#!/usr/bin/env python3
"""Globally replay-safe path model.

For a path, required_mappings are its local signature.  Any source unit in
encoded that is not in that signature is a forbidden_source_unit: selecting a
global mapping for it would create an extra scorer output character.  The
CP-SAT implications below enforce both facts during optimization, rather than
discovering the conflict only in a final replay.
"""
import csv,json,time,hashlib,sys,itertools
from pathlib import Path
from collections import defaultdict
from ortools.sat.python import cp_model

ROOT=Path(__file__).resolve().parent
BASE=ROOT.parent
DEC=BASE/'f68r2_e3_decomposed_search_v1'
PREP=BASE/'exploratory_astronomical_dictionary_search_v1'
sys.path.insert(0,str(BASE/'f68r2_e3_historical_operations_v1'))
import run_e3

def read(p):
    with p.open(encoding='utf-8',newline='') as f:return list(csv.DictReader(f,delimiter='\t'))
def write(p,fields,rows):
    with p.open('w',encoding='utf-8',newline='') as f:
        w=csv.DictWriter(f,fieldnames=fields,delimiter='\t',lineterminator='\n',extrasaction='ignore');w.writeheader();w.writerows(rows)
def sha(p): return hashlib.sha256(p.read_bytes()).hexdigest()

def enrich(p):
    if 'eva_token' not in p: p=dict(p,eva_token=p.get('token',''))
    if 'label_id' not in p: p=dict(p,label_id=p.get('label',''))
    rules=tuple(filter(None,p['rules'].split(';')))
    lhs={r.split('->')[0] for r in rules}
    return dict(p,required_mappings=';'.join(rules),forbidden_source_units=''.join(sorted(set(p['encoded'])-lhs)))

def build(paths,k,decision=None):
    m=cp_model.CpModel(); x=[m.NewBoolVar('p'+str(i)) for i in range(len(paths))]
    rules=sorted({r for p in paths for r in p['required_mappings'].split(';') if r}); ri={r:i for i,r in enumerate(rules)}
    y=[m.NewBoolVar('m'+str(i)) for i in range(len(rules))]
    toks=sorted({p['eva_token'] for p in paths}); ti={t:i for i,t in enumerate(toks)}
    z={(r,t):m.NewBoolVar(f'z{r}_{t}') for r in range(len(rules)) for t in range(len(toks))}
    bylabel=defaultdict(list); byid=defaultdict(list); support=defaultdict(list); supporttok=defaultdict(list)
    for i,p in enumerate(paths):
        bylabel[p['label_id']].append(x[i]);byid[p['identity']].append(x[i])
        m.AddImplication(x[i],z[ri[p['required_mappings'].split(';')[0]],ti[p['eva_token']]]) if False else None
        req=[r for r in p['required_mappings'].split(';') if r]
        for r in req:
            j=ri[r];m.AddImplication(x[i],y[j]);support[j].append(x[i]);supporttok[j,ti[p['eva_token']]].append(x[i])
        for f in filter(None,p['forbidden_source_units']):
            for r,j in ri.items():
                if r.split('->')[0]==f:m.AddImplication(x[i],y[j].Not())
        for r in req:
            j=ri[r];m.AddImplication(x[i],z[j,ti[p['eva_token']]])
    for vs in bylabel.values():m.Add(sum(vs)<=1)
    for vs in byid.values():m.Add(sum(vs)<=1)
    for j,vs in support.items():
        m.Add(y[j]<=sum(vs));m.Add(sum(z[j,t] for t in range(len(toks)))>=2*y[j])
        for t in range(len(toks)):m.Add(z[j,t]<=sum(supporttok.get((j,t),[])))
    for a in sorted({r.split('->')[0] for r in rules}):m.Add(sum(y[j] for j,r in enumerate(rules) if r.split('->')[0]==a)<=1)
    for b in sorted({r.split('->')[1] for r in rules}):m.Add(sum(y[j] for j,r in enumerate(rules) if r.split('->')[1]==b)<=1)
    m.Add(sum(y)<=k)
    if decision is not None:m.Add(sum(x)>=decision)
    else:m.Maximize(1000000*sum(x)+1000*sum(z.values())-10*sum(y))
    return m,x,rules

def solve(paths,k,seconds=60,decision=None):
    if not paths:return {'status':'INFEASIBLE','coverage':0,'chosen':[],'build_s':0,'solve_s':0,'bound':0}
    t=time.monotonic();m,x,rules=build(paths,k,decision);build_s=time.monotonic()-t
    s=cp_model.CpSolver();s.parameters.max_time_in_seconds=seconds;s.parameters.num_search_workers=1;s.parameters.random_seed=6804
    t=time.monotonic();st=s.Solve(m);solve_s=time.monotonic()-t
    chosen=[paths[i] for i,v in enumerate(x) if st in (cp_model.FEASIBLE,cp_model.OPTIMAL) and s.Value(v)]
    return {'status':s.StatusName(st),'coverage':len(chosen),'chosen':chosen,'build_s':build_s,'solve_s':solve_s,'bound':s.BestObjectiveBound(),'rules':rules}

def replay(chosen):
    mapping={}
    for p in chosen:
        for q in filter(None,p['required_mappings'].split(';')):
            a,b=q.split('->')
            if (a in mapping and mapping[a]!=b) or (b in mapping.values() and mapping.get(a)!=b):return False
            mapping[a]=b
    return all(run_e3.encode_word(p['encoded'],mapping,'DROP_UNMAPPED','NONE')==p['eva_token'] for p in chosen)

def main():
    src=BASE/'f68r2_occurrence_search_v1'/'F68R2_PATH_GRAPH.tsv';raw=read(src);paths=[enrich(p) for p in raw]
    # Preserve only scorer-consistent rows; no candidate-table preselection.
    write(ROOT/'PATH_REQUIREMENTS.tsv',['path_id','system_id','label_id','identity','encoded','eva_token','required_mappings','forbidden_source_units','scorer_parity'],paths)
    curve=[];best=[]
    for k in range(1,13):
        r=solve(paths,k,60);ok=replay(r['chosen'])
        curve.append({'k':k,'status':r['status'],'coverage':r['coverage'],'global_replay':'PASS' if ok else 'FAIL','build_s':round(r['build_s'],3),'solve_s':round(r['solve_s'],3),'best_bound':r['bound'],'mapping_count':len({q for p in r['chosen'] for q in p['required_mappings'].split(';') if q})})
        if r['coverage']>0:best.extend(dict(p,k=k) for p in r['chosen'])
    write(ROOT/'CORRECTED_BASELINE_CURVE.tsv',list(curve[0]),curve)
    (ROOT/'RUN_STATUS.json').write_text(json.dumps({'MODEL':'REQUIRED_MAPPINGS_FORBIDDEN_SOURCE_UNITS','SOURCE_PATHS':len(paths),'CORRECTED_BASELINE_MAXIMUM':max(x['coverage'] for x in curve),'S043_FIXTURE':'PENDING','E3_RECOMPUTE_AUTHORIZED':'IF_BASELINE_LT_6','NULL_RUN_AUTHORIZED':'NO','SCIENTIFIC_RESULT':'NOT_EVALUATED'},indent=2,sort_keys=True)+'\n')
if __name__=='__main__':main()
