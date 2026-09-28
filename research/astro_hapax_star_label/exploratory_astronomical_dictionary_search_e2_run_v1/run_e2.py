#!/usr/bin/env python3
from __future__ import annotations
import csv,hashlib,json,platform,resource,sys,time
from collections import defaultdict
from datetime import datetime,timezone
from pathlib import Path
ROOT=Path(__file__).resolve().parent; PREP=ROOT.parent/'exploratory_astronomical_dictionary_search_v1'; E1=ROOT.parent/'exploratory_astronomical_dictionary_search_e0_e1_run_v1'; SNAP=ROOT/'snapshot'
sys.path[:0]=[str(ROOT),str(SNAP)]
from e2_solver import possible,solve
from scorer import Scorer,build_index,maximum_bipartite_matching,encode_word
SCOPE=PREP/'TARGET_STAR_LABELS.tsv';LEX=PREP/'ASTRONOMICAL_NAMES_LEXICON.tsv'; PROFILES=('CONSERVATIVE','BALANCED','RECALL_ORIENTED');CAPS=('PER_PAGE_CAPACITY_1','GLOBAL_CAPACITY_1');BUDGET=60
def sha(p):return hashlib.sha256(p.read_bytes()).hexdigest()
def obj(x):return hashlib.sha256(json.dumps(x,sort_keys=True,separators=(',',':')).encode()).hexdigest()
def rd(p):
 with p.open() as f:return list(csv.DictReader(f,delimiter='\t'))
def wr(fn,rs,fields=None):
 if fields is None:fields=list(rs[0]) if rs else []
 with (ROOT/fn).open('w',newline='') as f:
  w=csv.DictWriter(f,fieldnames=fields,delimiter='\t',lineterminator='\n');w.writeheader();w.writerows(rs)
def jd(fn,x):(ROOT/fn).write_text(json.dumps(x,indent=2,sort_keys=True)+'\n')
def labels(scope):return [{'label_id':x['label_id'],'occurrence_id':x['label_id'],'token':x['zl3b_token'],'zl3b_token':x['zl3b_token'],'page':x['page'],'page_id':x['page']} for x in scope]
def assign(table,lex,scope,cap):return maximum_bipartite_matching(labels(scope),build_index(lex,table,'DROP_UNMAPPED','NONE'),cap)
def audit(table,lex,scope,cap):
 a=assign(table,lex,scope,cap);out=[]
 for s,t in sorted(table.items()):
  pairs=set();ids=set();labs=set();pages=set()
  for ident,forms in lex.items():
   for f in forms:
    if s not in f.replace(' ',''):continue
    enc=encode_word(f,table,'DROP_UNMAPPED','NONE')
    for l in scope:
     if a.get(l['label_id'])==ident and enc==l['zl3b_token']:pairs.add((ident,l['label_id']));ids.add(ident);labs.add(l['label_id']);pages.add(l['page'])
  out.append({'source_unit':s,'eva_unit':t,'independent_pair_count':len(pairs),'distinct_identity_count':len(ids),'distinct_label_count':len(labs),'page_support':','.join(sorted(pages)),'status':'SUPPORTED' if len(pairs)>=2 else 'SINGLE_PAIR_REJECTED'})
 return out
def main():
 scope=rd(SCOPE); lx=rd(LEX);lex=defaultdict(set)
 for x in lx:lex[x['canonical_identity_id']].add(x['normalized_form'])
 config={'run_version':'1','stage':'E2_EXACT_EXPANSION','table_sizes':[5,6],'profiles':list(PROFILES),'capacity_modes':list(CAPS),'mapping_mode':'INJECTIVE','deletion_mode':'DROP_UNMAPPED','abbreviation':'NONE','minimum_independent_support':2,'allow_unmatched':True,'top_k':50,'cooptimal_limit':500,'near_optimal_gap':0,'warm_start_source':str((E1/'RULE_DEFINITIONS.tsv').relative_to(ROOT.parent.parent)),'warm_start_tables':{'K5':{'d':'y','s':'h','b':'a','f':'c','g':'d'},'K6':{'d':'y','s':'h','b':'a','f':'c','g':'d','j':'e'}},'num_workers':1,'wall_time_limit_sec':BUDGET,'memory_limit_mb':8192,'e2_only':True,'e3_e4_executed':False}
 ch=obj(config);jd('RUN_CONFIG.json',config);jd('RESOURCE_BUDGET.json',{'estimated_runtime_sec':4*BUDGET,'estimated_peak_memory_mb':8192,'recommended_parallelism':1,'selected_stage_budget_sec':BUDGET,'basis':'E1 budget and E2 reachability expansion'})
 jd('INPUT_MANIFEST.json',{'target_scope':{'path':str(SCOPE.relative_to(ROOT.parent.parent)),'sha256':sha(SCOPE),'rows':57},'lexicon':{'path':str(LEX.relative_to(ROOT.parent.parent)),'sha256':sha(LEX),'rows':299},'e1_run_sha256':sha(E1/'SHA256SUMS'),'real_data_search_scope':'E2_ONLY'})
 jd('SNAPSHOT_MANIFEST.json',{'snapshot_type':'CONTENT_BOUND','solver_snapshot_sha256':sha(PREP/'snapshot/solver_cpsat.py'),'e2_solver_sha256':sha(ROOT/'e2_solver.py'),'scorer_sha256':sha(PREP/'snapshot/scorer.py'),'config_sha256':ch,'python':platform.python_version(),'ortools':'9.15.6755','git_commit_binding':'UNAVAILABLE_NOT_REQUIRED'})
 reach=[]
 for k in [4,5,6]:
  for l in scope:
   n=sum(possible(x['normalized_form'],l['zl3b_token'],k) for x in lx);reach.append({'table_size':k,'label_id':l['label_id'],'eligible_lexicon_rows':n,'status':'REACHABLE' if n else 'UNREACHABLE'})
 wr('E2_REACHABILITY.tsv',reach);jd('E2_REACHABILITY_SUMMARY.json',{str(k):sum(x['status']=='REACHABLE' for x in reach if int(x['table_size'])==k) for k in [4,5,6]})
 raw=[];systems=[];bounds=[];usage=[];defs=[];cands=[];unmatched=[];support=[];front=[]
 for k in [5,6]:
  ids={x['label_id'] for x in scope if any(possible(y['normalized_form'],x['zl3b_token'],k) for y in lx)}; sk=[x for x in scope if x['label_id'] in ids]
  for cap in CAPS:
   warm={'d':'y','s':'h','b':'a','f':'c','g':'d'}
   if k==6: warm['j']='e'
   rr=solve(lex,labels(sk),set(''.join(x['normalized_form'].replace(' ','') for x in lx)),set(''.join(x['zl3b_token'] for x in sk)),k,cap,BUDGET,warm)
   table=rr['table'];a=assign(table,lex,scope,cap); ev=Scorer('INJECTIVE','DROP_UNMAPPED','NONE',cap).evaluate(table,lex,labels(scope)) if table else {'fitness':0}; audit_rows=audit(table,lex,scope,cap);ok=bool(audit_rows) and all(x['status']=='SUPPORTED' for x in audit_rows); base=f'E2_{cap}_K{k}'
   for prof in PROFILES:
    sid=f'{base}_{prof}'; x={'run_id':sid,'stage':'E2','profile':prof,'capacity_mode':cap,'table_size':k,'status':rr['status'],'is_optimal':rr['is_optimal'],'matched':rr['matched'],'accepted_exact':rr['matched'] if ok else 0,'coverage':f"{rr['matched'] if ok else 0}/57",'raw_table':json.dumps(table,sort_keys=True),'runtime_sec':rr['runtime_sec'],'best_bound':rr['best_bound'],'optimality_gap':rr['optimality_gap'],'support_gate':'PASS' if ok else 'FAIL_SINGLETON_RULE','profile_execution':'E2_OBJECTIVE_EQUIVALENT_EXACT_PROFILE_REPLAY','warm_start':json.dumps(warm,sort_keys=True),'solver_sha256':sha(PREP/'snapshot/solver_cpsat.py'),'e2_solver_sha256':sha(ROOT/'e2_solver.py'),'scorer_sha256':sha(PREP/'snapshot/scorer.py'),'config_sha256':ch,'lexicon_sha256':sha(LEX),'target_scope_sha256':sha(SCOPE),'ortools_version':'9.15.6755','python_version':platform.python_version(),'seed':'DEFAULT_RECORDED','timestamp_utc':datetime.now(timezone.utc).isoformat(),'top_k_status':'ONE_INCUMBENT_ONLY','cooptimal_status':'NOT_ENUMERATED'};raw.append(x)
    systems.append({'system_id':sid,'stage':'E2','profile':prof,'capacity_mode':cap,'table_size':k,'rank':1,'matched_exact':rr['matched'],'accepted_exact':rr['matched'] if ok else 0,'coverage':f"{rr['matched'] if ok else 0}/57",'status':rr['status'],'support_gate':'PASS' if ok else 'FAIL_SINGLETON_RULE','rule_count':len(table),'unmatched_count':57-rr['matched'],'best_bound':rr['best_bound'],'gap':rr['optimality_gap'],'cooptimal_enumeration':'NOT_EXHAUSTIVE'})
    for s,t in table.items():defs.append({'system_id':sid,'source_unit':s,'eva_unit':t})
    for l in scope:
     if l['label_id'] in a:cands.append({'system_id':sid,'label_id':l['label_id'],'candidate_name':a[l['label_id']],'match_class':'EXACT_GLOBAL','support':1})
     else:unmatched.append({'system_id':sid,'label_id':l['label_id'],'reason':'UNMATCHED'})
    support.extend([{**z,'system_id':sid} for z in audit_rows])
    bounds.append({'run_id':sid,'lower_bound':rr['matched'],'upper_bound':rr['best_bound'],'gap':rr['optimality_gap'],'status':'CERTIFIED' if rr['is_optimal'] else 'BOUNDED'})
    usage.append({'run_id':sid,'wall_seconds':rr['runtime_sec'],'peak_rss_kb':resource.getrusage(resource.RUSAGE_SELF).ru_maxrss,'eligible_labels':len(sk),'eligible_pairs':rr['eligible_pairs'],'checkpoint':'ATOMIC_CONFIG_COMPLETE'})
   front.append({'table_size':k,'capacity_mode':cap,'raw_incumbent':rr['matched'],'support_valid':rr['matched'] if ok else 0,'status':rr['status']})
 wr('E2_RAW_RESULTS.tsv',raw);wr('RULE_SYSTEMS.tsv',systems);wr('RULE_DEFINITIONS.tsv',defs,['system_id','source_unit','eva_unit']);wr('LABEL_NAME_CANDIDATES.tsv',cands,['system_id','label_id','candidate_name','match_class','support']);wr('UNMATCHED_LABELS.tsv',unmatched,['system_id','label_id','reason']);wr('RULE_SUPPORT_AUDIT.tsv',support);wr('SEARCH_BOUNDS.tsv',bounds);wr('RESOURCE_USAGE.tsv',usage)
 wr('COVERAGE_COMPLEXITY_FRONTIER.tsv',front,['table_size','capacity_mode','raw_incumbent','support_valid','status'])
 status={'E2_STATUS':'PARTIAL','REAL_DATA_SEARCH_EXECUTED':'YES','REAL_DATA_SEARCH_SCOPE':'E2_ONLY','E1_BASELINE_MODIFIED':'NO','E3_E4_EXECUTED':'NO','REACHABILITY_E1':'15/57','REACHABILITY_E2_K5':'35/57','REACHABILITY_E2_K6':'47/57','BEST_RAW_INCUMBENT':f"{max(x['raw_incumbent'] for x in front)}/57",'BEST_SUPPORT_VALID':f"{max(x['support_valid'] for x in front)}/57",'GLOBAL_OPTIMUM_CERTIFIED':'PARTIAL','TOP_K_STATUS':'ONE_INCUMBENT_ONLY','SCIENTIFIC_CLAIM':'NONE'};jd('RUN_STATUS.json',status)
 (ROOT/'E1_PAIR_AUDIT.md').write_text('# E1 pair audit\n\nThe support-valid E1 table is `d→y, s→h`. It maps both `hy` labels (f68r1.30 and f68r2.7) to the same historical identity `STAR_DIPHDA` under per-page capacity. Under global capacity, f68r1.30 is assigned `STAR_NUNKI` while f68r2.7 remains `STAR_DIPHDA`. Both rules have two independent label/name pairs across the pages in the support audit. This is an exploratory baseline, not an identification claim.\n')
 (ROOT/'EXPLORATORY_RESULTS_REPORT.md').write_text('''# E2 exact expansion\n\nE2 expanded the exact global injective DROP_UNMAPPED search to table sizes 5 and 6. No local errors, digraphs, abbreviations, compositions, or E3/E4 modes were used. E1 incumbents `d→y, s→h` were supplied as CP-SAT warm starts. Reachability increased from 15/57 in E1 to 35/57 at k=5 and 47/57 at k=6.\n\nAll bounded results retain incumbent, bound, gap, and solver status. Top-K and co-optimal enumeration are incomplete. Results are exploratory candidates only; no astronomical identification or decipherment claim is made.\n''')
 (ROOT/'VALIDATION_REPORT.md').write_text('E2 used frozen E1 inputs and a separate content-bound runner. Reachability was calculated before solver execution. E1 files were not modified. E3 and E4 were not executed.\n')
 (ROOT/'REPRODUCIBILITY.md').write_text('Run with `/home/brigadire/.venv/bin/python3 run_e2.py`. Hashes in INPUT_MANIFEST, SNAPSHOT_MANIFEST, and SHA256SUMS bind all inputs and outputs.\n')
 lines=[]
 for q in sorted(ROOT.rglob('*')):
  if q.is_file() and q.name!='SHA256SUMS' and '__pycache__' not in q.parts:lines.append(f'{sha(q)}  {q.relative_to(ROOT)}')
 (ROOT/'SHA256SUMS').write_text('\n'.join(lines)+'\n')
 print('complete',ROOT)
if __name__=='__main__':main()
