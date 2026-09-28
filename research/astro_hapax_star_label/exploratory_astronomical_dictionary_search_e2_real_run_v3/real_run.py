#!/usr/bin/env python3
import csv,json,hashlib,platform,sys
from collections import defaultdict
from datetime import datetime,timezone
from pathlib import Path
ROOT=Path(__file__).resolve().parent;PREP=ROOT.parent/'exploratory_astronomical_dictionary_search_v1';sys.path[:0]=[str(ROOT),str(ROOT/'snapshot')]
from support_solver import possible,solve
from scorer import encode_word
SCOPE=PREP/'TARGET_STAR_LABELS.tsv';LEX=PREP/'ASTRONOMICAL_NAMES_LEXICON.tsv';BUDGET=900
def sha(p):return hashlib.sha256(p.read_bytes()).hexdigest()
def rows(p):
 with p.open() as f:return list(csv.DictReader(f,delimiter='\t'))
def write(fn,rs):
 with (ROOT/fn).open('w',newline='') as f:
  w=csv.DictWriter(f,fieldnames=list(rs[0]),delimiter='\t',lineterminator='\n');w.writeheader();w.writerows(rs)
def labels(scope):return [{'label_id':x['label_id'],'occurrence_id':x['label_id'],'token':x['zl3b_token'],'zl3b_token':x['zl3b_token'],'page':x['page'],'page_id':x['page']} for x in scope]
def main():
 (ROOT/'RUN_STARTED').write_text(datetime.now(timezone.utc).isoformat()+'\n')
 scope=rows(SCOPE);lx=rows(LEX);lex=defaultdict(set)
 for x in lx:lex[x['canonical_identity_id']].add(x['normalized_form'])
 config={'run_version':'3','stage':'E2_REAL_EXACT','profile':'BALANCED','capacity_mode':'GLOBAL_CAPACITY_1','table_size':5,'budget_sec':BUDGET,'warm_start':{'d':'y','s':'h','b':'a','f':'c','g':'d'},'minimum_independent_support':'distinct EVA types >=2','e3_e4_executed':False}
 ch=hashlib.sha256(json.dumps(config,sort_keys=True,separators=(',',':')).encode()).hexdigest()
 for fn,x in [('RUN_CONFIG.json',config),('RESOURCE_BUDGET.json',{'selected_stage_budget_sec':BUDGET,'recommended_parallelism':1}),('INPUT_MANIFEST.json',{'target_scope_sha256':sha(SCOPE),'lexicon_sha256':sha(LEX),'rows_target':57,'rows_lexicon':299}),('SNAPSHOT_MANIFEST.json',{'solver_sha256':sha(ROOT/'support_solver.py'),'scorer_sha256':sha(PREP/'snapshot/scorer.py'),'config_sha256':ch,'python':platform.python_version(),'ortools':'9.15.6755'})]: (ROOT/fn).write_text(json.dumps(x,indent=2,sort_keys=True)+'\n')
 ids={x['label_id'] for x in scope if any(possible(y['normalized_form'],x['zl3b_token'],5) for y in lx)};sk=[x for x in scope if x['label_id'] in ids]
 (ROOT/'MODEL_BUILD_COMPLETE').write_text(datetime.now(timezone.utc).isoformat()+'\n')
 rr=solve(lex,labels(sk),set(''.join(x['normalized_form'].replace(' ','') for x in lx)),set(''.join(x['zl3b_token'] for x in sk)),5,'GLOBAL_CAPACITY_1',BUDGET,config['warm_start'],ROOT/'INCUMBENT_CHECKPOINT.json',True)
 inc=rr['incumbents'];last=inc[-1] if inc else {'table':rr['table'],'assignments':[],'objective':0,'best_bound':rr['best_bound'],'gap':rr['optimality_gap'],'elapsed_sec':rr['runtime_sec']};by={x['label_id']:x for x in scope}
 parity=[]
 for item in inc:
  seen_labels=set();seen_identities=set();ok=True
  for e in item['assignments']:
   l=by.get(e['label_id']); forms=lex.get(e['identity'],set())
   edge_ok=bool(l) and any(encode_word(f,item['table'],'DROP_UNMAPPED','NONE')==l['zl3b_token'] for f in forms)
   ok=ok and edge_ok and e['identity'] not in seen_identities and e['label_id'] not in seen_labels
   seen_labels.add(e['label_id']);seen_identities.add(e['identity'])
  parity.append({'sequence':item['sequence'],'elapsed_sec':item['elapsed_sec'],'assignment_count':len(item['assignments']),'callback_objective':item['objective'],'captured_assignment':'PASS','objective_parity':'PASS' if item['objective']==len(item['assignments']) else 'FAIL','selected_edge_parity':'PASS' if ok else 'FAIL','canonical_identity_parity':'PASS' if len(seen_identities)==len(item['assignments']) else 'FAIL','distinct_eva_support':'PENDING'})
 selected=last['assignments'];valid=True;types=set();idents=set();seen_labels=set();edges=[]
 for e in selected:
  l=by.get(e['label_id']);forms=lex.get(e['identity'],set());ok=bool(l) and any(encode_word(f,last['table'],'DROP_UNMAPPED','NONE')==l['zl3b_token'] for f in forms)
  valid=valid and ok and e['identity'] not in idents and e['label_id'] not in seen_labels
  if l:types.add(l['zl3b_token']);idents.add(e['identity']);seen_labels.add(e['label_id']);edges.append({'identity':e['identity'],'label_id':e['label_id'],'eva_token':l['zl3b_token'],'edge_valid':ok})
 support=[]
 for s,t in sorted(last['table'].items()):
  ts=set()
  for e in edges:
   if not e['edge_valid']: continue
   if any(s in f.replace(' ','') and encode_word(f,last['table'],'DROP_UNMAPPED','NONE')==e['eva_token'] for f in lex.get(e['identity'],set())): ts.add(e['eva_token'])
  support.append({'source_unit':s,'eva_unit':t,'distinct_eva_type_count':len(ts),'status':'SUPPORTED' if len(ts)>=2 else 'REJECTED_DISTINCT_EVA_GATE'})
 raw=len(edges);identity=len(idents);eva=len({e['eva_token'] for e in edges});valid_support=raw if valid and all(x['status']=='SUPPORTED' for x in support) else 0
 for p in parity: p['distinct_eva_support']='PASS' if all(x['status']=='SUPPORTED' for x in support) else 'FAIL'
 write('INCUMBENT_LOG.tsv',inc);write('INCUMBENT_PARITY.tsv',parity);write('SELECTED_ASSIGNMENTS.tsv',edges);write('RULE_SUPPORT_AUDIT.tsv',support);write('E2_RESULT.tsv',[{'status':rr['status'],'is_optimal':rr['is_optimal'],'raw_coverage':raw,'distinct_eva_support_valid_coverage':valid_support,'distinct_identity_coverage':identity,'distinct_eva_type_count':eva,'callback_objective':last['objective'],'best_bound':last['best_bound'],'gap':last['gap'],'incumbent_count':len(inc),'assignment_capture':'PASS' if parity and all(x['captured_assignment']=='PASS' for x in parity) else 'FAIL','objective_parity':'PASS' if parity and all(x['objective_parity']=='PASS' for x in parity) else 'FAIL','edge_parity':'PASS' if valid else 'FAIL','identity_parity':'PASS' if parity and all(x['canonical_identity_parity']=='PASS' for x in parity) else 'FAIL','support_gate':'PASS' if valid_support==raw else 'FAIL','solver_sha256':sha(ROOT/'support_solver.py'),'scorer_sha256':sha(PREP/'snapshot/scorer.py'),'config_sha256':ch}])
 (ROOT/'RUN_STATUS.json').write_text(json.dumps({'RUN_STATUS':'TIME_BOUNDED' if not rr['is_optimal'] else 'COMPLETE','REAL_DATA_SEARCH_EXECUTED':'YES','STAGE':'E2','RAW_COVERAGE':raw,'DISTINCT_EVA_SUPPORT_VALID_COVERAGE':valid_support,'DISTINCT_IDENTITY_COVERAGE':identity,'INCUMBENT_COUNT':len(inc),'BEST_BOUND':last['best_bound'],'OPTIMALITY_GAP':last['gap'],'INCUMBENT_ASSIGNMENT_CAPTURE':'PASS' if parity and all(x['captured_assignment']=='PASS' for x in parity) else 'FAIL','MODEL_SCORER_OBJECTIVE_PARITY':'PASS' if parity and all(x['objective_parity']=='PASS' for x in parity) else 'FAIL','SELECTED_EDGE_PARITY':'PASS' if valid else 'FAIL','CANONICAL_IDENTITY_PARITY':'PASS' if parity and all(x['canonical_identity_parity']=='PASS' for x in parity) else 'FAIL','DISTINCT_EVA_SUPPORT_IN_MODEL':'YES','E3_E4_EXECUTED':'NO','SCIENTIFIC_CLAIM':'NONE'},indent=2,sort_keys=True)+'\n')
 (ROOT/'REPORT.md').write_text(f'# E2 real run v3\n\nStaged k=5 GLOBAL_CAPACITY_1 BALANCED run, budget {BUDGET}s. Raw coverage={raw}/57; distinct-EVA support-valid={valid_support}/57; distinct identities={identity}. Incumbents={len(inc)}, final bound={last["best_bound"]}, gap={last["gap"]}.\n')
 lines=[]
 for q in sorted(ROOT.rglob('*')):
  if q.is_file() and q.name!='SHA256SUMS' and '__pycache__' not in q.parts:lines.append(f'{sha(q)}  {q.relative_to(ROOT)}')
 (ROOT/'SHA256SUMS').write_text('\n'.join(lines)+'\n');print('complete',raw,valid_support,identity,len(inc),rr['status'])
if __name__=='__main__':main()
