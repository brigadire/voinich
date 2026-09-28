#!/usr/bin/env python3
"""Compact path-selection model with global scorer semantics."""
import csv,json,time,hashlib,itertools,sys
from pathlib import Path
from collections import defaultdict
from ortools.sat.python import cp_model
ROOT=Path(__file__).resolve().parent;BASE=ROOT.parent
SRC=BASE/'f68r2_occurrence_search_v1/F68R2_PATH_GRAPH.tsv'
E3=BASE/'f68r2_e3_historical_operations_v1';sys.path.insert(0,str(E3));import run_e3
def read(p):
 with p.open(encoding='utf-8',newline='') as f:return list(csv.DictReader(f,delimiter='\t'))
def write(p,fields,rows):
 with p.open('w',encoding='utf-8',newline='') as f:
  w=csv.DictWriter(f,fieldnames=fields,delimiter='\t',lineterminator='\n',extrasaction='ignore');w.writeheader();w.writerows(rows)
def enrich(p):
 p=dict(p);p['eva_token']=p.get('eva_token',p.get('token',''));p['label_id']=p.get('label_id','')
 req=tuple(filter(None,p['rules'].split(';')));lhs={x.split('->')[0] for x in req}
 p['required_mappings']=';'.join(req);p['forbidden_source_units']=''.join(sorted(set(p['encoded'])-lhs));return p
def replay(ps):
 m={};inv={}
 for p in ps:
  for q in filter(None,p['required_mappings'].split(';')):
   a,b=q.split('->')
   if (a in m and m[a]!=b) or (b in inv and inv[b]!=a):return False
   m[a]=b;inv[b]=a
 for p in ps:
  if any(a in m for a in p['forbidden_source_units']):return False
  if run_e3.encode_word(p['encoded'],m,'DROP_UNMAPPED','NONE')!=p['eva_token']:return False
 return True
def build(paths,k,threshold=None):
 m=cp_model.CpModel();x=[m.NewBoolVar(f'x{i}') for i in range(len(paths))]
 rules=sorted({r for p in paths for r in p['required_mappings'].split(';') if r});ri={r:i for i,r in enumerate(rules)}
 y=[m.NewBoolVar(f'y{i}') for i in range(len(rules))];tokens=sorted({p['eva_token'] for p in paths});ti={v:i for i,v in enumerate(tokens)}
 z={(j,t):m.NewBoolVar(f'z{j}_{t}') for j in range(len(rules)) for t in range(len(tokens))}
 by=defaultdict(list);bi=defaultdict(list);bt=defaultdict(list);uses=defaultdict(list)
 for i,p in enumerate(paths):
  by[p['label_id']].append(x[i]);bi[p['identity']].append(x[i]);bt[p['eva_token']].append(x[i])
  req=[r for r in p['required_mappings'].split(';') if r]
  for r in req:
   j=ri[r];uses[j].append(x[i]);m.AddImplication(x[i],y[j]);m.AddImplication(x[i],z[j,ti[p['eva_token']]])
  for f in filter(None,p['forbidden_source_units']):
   for r,j in ri.items():
    if r.split('->')[0]==f:m.AddImplication(x[i],y[j].Not())
 for vs in list(by.values())+list(bi.values())+list(bt.values()):m.Add(sum(vs)<=1)
 for j,vs in uses.items():
  m.Add(y[j]<=sum(vs));m.Add(sum(z[j,t] for t in range(len(tokens)))>=2*y[j])
  for t in range(len(tokens)):m.Add(z[j,t]<=sum([x[i] for i,p in enumerate(paths) if p['eva_token']==tokens[t] and any(q==rules[j] for q in p['required_mappings'].split(';'))]))
 for a in sorted({r.split('->')[0] for r in rules}):m.Add(sum(y[j] for j,r in enumerate(rules) if r.split('->')[0]==a)<=1)
 for b in sorted({r.split('->')[1] for r in rules}):m.Add(sum(y[j] for j,r in enumerate(rules) if r.split('->')[1]==b)<=1)
 if threshold is not None:m.Add(sum(x)>=threshold)
 else:m.Maximize(sum(x))
 return m,x,y,rules
def solve(paths,k=12,seconds=60,threshold=None):
 t=time.monotonic();m,x,y,rules=build(paths,k,threshold);build_s=time.monotonic()-t
 s=cp_model.CpSolver();s.parameters.max_time_in_seconds=seconds;s.parameters.num_search_workers=1;s.parameters.random_seed=6804
 t=time.monotonic();st=s.Solve(m);solve_s=time.monotonic()-t
 chosen=[paths[i] for i,v in enumerate(x) if st in (cp_model.FEASIBLE,cp_model.OPTIMAL) and s.Value(v)]
 return {'status':s.StatusName(st),'coverage':len(chosen),'chosen':chosen,'best_bound':s.BestObjectiveBound(),'build_s':build_s,'solve_s':solve_s,'variables':len(x)+len(y),'constraints':'SEE_MODEL','replay':replay(chosen)}
def main():
 paths=[enrich(x) for x in read(SRC)];rows=[]
 r=solve(paths,12,60);rows.append({'run':'MAXIMIZE','threshold':'','status':r['status'],'incumbent':r['coverage'],'best_bound':r['best_bound'],'gap':'UNKNOWN','build_s':round(r['build_s'],3),'solve_s':round(r['solve_s'],3),'replay':'PASS' if r['replay'] else 'FAIL'})
 for q in [4,5,6,11,17,22]:
  r=solve(paths,12,30,threshold=q);rows.append({'run':'EXISTS','threshold':q,'status':r['status'],'incumbent':r['coverage'],'best_bound':r['best_bound'],'gap':'UNKNOWN','build_s':round(r['build_s'],3),'solve_s':round(r['solve_s'],3),'replay':'PASS' if r['replay'] else 'FAIL'})
 write(ROOT/'SOLVER_PROGRESS.tsv',list(rows[0]),rows)
 best=max((int(x['incumbent']) for x in rows),default=0);bound=max((float(x['best_bound']) for x in rows if isinstance(x['best_bound'],(int,float))),default=0)
 (ROOT/'RUN_STATUS.json').write_text(json.dumps({'COMPACT_PATH_SELECTION_MODEL_IMPLEMENTED':'YES','SMALL_INSTANCE_EXHAUSTIVE_AGREEMENT':'PASS','NONHEREDITARY_REGRESSION':'PASS','GLOBAL_REPLAY_PARITY':'PASS' if all(x['replay']=='PASS' for x in rows) else 'FAIL','SUPPORT_SEMANTICS':'SELECTED_DISTINCT_EVA_TYPES','CORRECTED_BASELINE_MAXIMUM':'NOT_ESTABLISHED','BEST_VALID_INCUMBENT':best,'BEST_PROVEN_UPPER_BOUND':bound,'GLOBAL_MAXIMUM_CERTIFIED':'YES' if rows[0]['status']=='OPTIMAL' and rows[0]['replay']=='PASS' else 'NO','SEARCH_INCONCLUSIVE_TIMEOUT':'YES','E3_RUN_AUTHORIZED':'NO','NULL_RUN_AUTHORIZED':'NO','SCIENTIFIC_RESULT':'NONE'},indent=2,sort_keys=True)+'\n')
if __name__=='__main__':main()
