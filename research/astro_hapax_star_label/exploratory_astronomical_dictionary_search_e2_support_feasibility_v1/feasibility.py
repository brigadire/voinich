#!/usr/bin/env python3
import csv,json,hashlib,time,sys
from collections import defaultdict
from itertools import combinations
from pathlib import Path
sys.path.insert(0,str(Path(__file__).resolve().parent.parent/'exploratory_astronomical_dictionary_search_e2_structural_support_audit_v1'))
from audit import alignments
from ortools.sat.python import cp_model
ROOT=Path(__file__).resolve().parent; PREP=ROOT.parent/'exploratory_astronomical_dictionary_search_v1'
def rows(p): return list(csv.DictReader(p.open(),delimiter='\t'))
def main():
 labels=rows(PREP/'TARGET_STAR_LABELS.tsv'); labels=[dict(x,token=x['zl3b_token'],page_id=x['page']) for x in labels]; raw=rows(PREP/'ASTRONOMICAL_NAMES_LEXICON.tsv'); lex=defaultdict(set)
 for x in raw: lex[x['canonical_identity_id']].add(x['normalized_form'])
 variants={}; started=time.time()
 for ident,forms in lex.items():
  for form in forms:
   for label in labels:
    for mapping in alignments(form,label['token'],5):
     key=(ident,label['label_id'],mapping)
     variants.setdefault(key,{'identity':ident,'label_id':label['label_id'],'token':label['token'],'form':form,'rules':mapping})
 (ROOT/'BUILD_HEARTBEAT').write_text(json.dumps({'phase':'VARIANT_GENERATION','variants':len(variants),'elapsed_sec':time.time()-started})+'\n')
 model=cp_model.CpModel(); vars_y=[]; by_label=defaultdict(list); by_ident=defaultdict(list); by_rule=defaultdict(list); by_rule_token=defaultdict(list); rules=set()
 for n,v in enumerate(variants.values()):
  y=model.NewBoolVar(f'y_{n}'); vars_y.append(y); by_label[v['label_id']].append(y); by_ident[v['identity']].append(y)
  for r in v['rules']: rules.add(r); by_rule[r].append(y); by_rule_token[r,v['token']].append(y)
 for vs in by_label.values(): model.Add(sum(vs)<=1)
 for vs in by_ident.values(): model.Add(sum(vs)<=1)
 x={r:model.NewBoolVar('x_'+str(i)) for i,r in enumerate(sorted(rules))}
 for r,v in x.items():
  model.AddMaxEquality(v,by_rule[r]);
  for y in by_rule[r]: model.AddImplication(y,v)
 model.Add(sum(x.values())<=5); 
 for s in sorted({r[0] for r in rules}): model.Add(sum(v for r,v in x.items() if r[0]==s)<=1)
 for t in sorted({r[1] for r in rules}): model.Add(sum(v for r,v in x.items() if r[1]==t)<=1)
 z={}
 for r in sorted(rules):
  for token in sorted({v['token'] for v in variants.values()}):
   z[r,token]=model.NewBoolVar(f'z_{r[0]}_{r[1]}_{token}')
   terms=by_rule_token.get((r,token),[])
   if terms:model.AddMaxEquality(z[r,token],terms)
   else:model.Add(z[r,token]==0)
  model.Add(sum(z[r,t] for t in sorted({v['token'] for v in variants.values()}))>=2*x[r])
 model.Add(sum(vars_y)>=2); model.Maximize(sum(vars_y))
 (ROOT/'BUILD_HEARTBEAT').write_text(json.dumps({'phase':'MODEL_BUILT','variants':len(variants),'rules':len(rules),'elapsed_sec':time.time()-started})+'\n')
 solver=cp_model.CpSolver(); solver.parameters.max_time_in_seconds=300; solver.parameters.num_workers=1; status=solver.Solve(model); feasible=status in (cp_model.OPTIMAL,cp_model.FEASIBLE)
 selected=[]; table={}
 if feasible:
  selected=[v for y,v in zip(vars_y,variants.values()) if solver.Value(y)]
  table={r[0]:r[1] for r in rules if solver.Value(x[r])}
 support=[]
 for r in sorted(table):
  types=sorted({v['token'] for v in selected if r in v['rules']}); support.append({'source_unit':r,'eva_unit':table[r],'distinct_eva_type_count':len(types),'support_types':','.join(types),'status':'SUPPORTED' if len(types)>=2 else 'FAIL'})
 with (ROOT/'ASSIGNMENT.tsv').open('w',newline='') as f:
  w=csv.writer(f,delimiter='\t');w.writerow(['identity','label_id','token','dictionary_form','rules']);[w.writerow([v['identity'],v['label_id'],v['token'],v['form'],json.dumps(v['rules'])]) for v in selected]
 with (ROOT/'RULE_SUPPORT.tsv').open('w',newline='') as f:
  w=csv.DictWriter(f,fieldnames=['source_unit','eva_unit','distinct_eva_type_count','support_types','status'],delimiter='\t');w.writeheader();w.writerows(support)
 result={'solver_status':solver.StatusName(status),'feasible':feasible,'objective':len(selected),'rule_count':len(table),'variant_count':len(variants),'rule_support_pass':'PASS' if feasible and all(x['status']=='SUPPORTED' for x in support) else 'FAIL','table':table,'elapsed_sec':round(time.time()-started,4)}
 (ROOT/'FEASIBILITY_RESULT.json').write_text(json.dumps(result,indent=2,sort_keys=True)+'\n'); print(json.dumps(result,indent=2))
if __name__=='__main__':main()
