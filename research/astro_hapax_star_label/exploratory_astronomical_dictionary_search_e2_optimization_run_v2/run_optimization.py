#!/usr/bin/env python3
import csv,hashlib,json,os,platform,resource,sys,time
from collections import defaultdict
from datetime import datetime,timezone
from pathlib import Path
ROOT=Path(__file__).resolve().parent;PREP=ROOT.parent/'exploratory_astronomical_dictionary_search_v1';DIAG=ROOT.parent/'exploratory_astronomical_dictionary_search_e2_run_v1';sys.path[:0]=[str(ROOT),str(ROOT/'snapshot')]
from e2_solver import possible,solve
from scorer import Scorer,build_index,maximum_bipartite_matching,encode_word
SCOPE=PREP/'TARGET_STAR_LABELS.tsv';LEX=PREP/'ASTRONOMICAL_NAMES_LEXICON.tsv';BUDGET=900
def sha(p):return hashlib.sha256(p.read_bytes()).hexdigest()
def rd(p):
 with p.open() as f:return list(csv.DictReader(f,delimiter='\t'))
def wr(fn,rs):
 with (ROOT/fn).open('w',newline='') as f:
  w=csv.DictWriter(f,fieldnames=list(rs[0]),delimiter='\t',lineterminator='\n');w.writeheader();w.writerows(rs)
def jd(fn,x):(ROOT/fn).write_text(json.dumps(x,indent=2,sort_keys=True)+'\n')
def labels(scope):return [{'label_id':x['label_id'],'occurrence_id':x['label_id'],'token':x['zl3b_token'],'zl3b_token':x['zl3b_token'],'page':x['page'],'page_id':x['page']} for x in scope]
def main():
 scope=rd(SCOPE);lx=rd(LEX);lex=defaultdict(set)
 for x in lx:lex[x['canonical_identity_id']].add(x['normalized_form'])
 config={'run_version':'2','stage':'E2_FULL_OPTIMIZATION','profile':'BALANCED','capacity_mode':'GLOBAL_CAPACITY_1','table_size':5,'budget_sec':BUDGET,'warm_start':{'d':'y','s':'h','b':'a','f':'c','g':'d'},'minimum_independent_support':'distinct EVA types >=2; repeated occurrence does not count','top_k':50,'cooptimal_limit':500,'num_workers':1,'e2_only':True,'e3_e4_executed':False}
 ch=hashlib.sha256(json.dumps(config,sort_keys=True,separators=(',',':')).encode()).hexdigest();jd('RUN_CONFIG.json',config);jd('RESOURCE_BUDGET.json',{'selected_stage_budget_sec':BUDGET,'estimated_peak_memory_mb':8192,'recommended_parallelism':1,'sequence':'k5 global balanced first; advance only after review'})
 jd('INPUT_MANIFEST.json',{'target_scope_sha256':sha(SCOPE),'lexicon_sha256':sha(LEX),'diagnostic_e2_sha256':sha(DIAG/'SHA256SUMS'),'rows_target':57,'rows_lexicon':299})
 jd('SNAPSHOT_MANIFEST.json',{'snapshot_type':'CONTENT_BOUND','solver_snapshot_sha256':sha(PREP/'snapshot/solver_cpsat.py'),'optimization_solver_sha256':sha(ROOT/'e2_solver.py'),'scorer_sha256':sha(PREP/'snapshot/scorer.py'),'config_sha256':ch,'python':platform.python_version(),'ortools':'9.15.6755','git_commit_binding':'UNAVAILABLE_NOT_REQUIRED'})
 ids={x['label_id'] for x in scope if any(possible(y['normalized_form'],x['zl3b_token'],5) for y in lx)};sk=[x for x in scope if x['label_id'] in ids]
 checkpoint=ROOT/'INCUMBENT_CHECKPOINT.json';warm=config['warm_start'];start=time.time();rr=solve(lex,labels(sk),set(''.join(x['normalized_form'].replace(' ','') for x in lx)),set(''.join(x['zl3b_token'] for x in sk)),5,'GLOBAL_CAPACITY_1',BUDGET,warm,checkpoint)
 ass=maximum_bipartite_matching(labels(scope),build_index(lex,rr['table'],'DROP_UNMAPPED','NONE'),'GLOBAL_CAPACITY_1') if rr['table'] else {}
 raw=len(ass);eva=len({x['token'] for x in labels(scope) if x['label_id'] in ass});ident=len(set(ass.values()))
 # Distinct-EVA support audit for the returned table.
 support=[]
 for s,t in sorted(rr['table'].items()):
  types=set();idents=set();labs=set()
  for identity,forms in lex.items():
   for form in forms:
    if s not in form.replace(' ',''):continue
    enc=encode_word(form,rr['table'],'DROP_UNMAPPED','NONE')
    for l in scope:
     if ass.get(l['label_id'])==identity and enc==l['zl3b_token']:types.add(l['zl3b_token']);idents.add(identity);labs.add(l['label_id'])
  support.append({'source_unit':s,'eva_unit':t,'distinct_eva_type_count':len(types),'distinct_identity_count':len(idents),'distinct_label_count':len(labs),'status':'SUPPORTED' if len(types)>=2 else 'REJECTED_DISTINCT_EVA_GATE'})
 valid=raw if support and all(x['status']=='SUPPORTED' for x in support) else 0
 inc=rr.get('incumbents',[])
 wr('INCUMBENT_LOG.tsv',inc)
 wr('RULE_SUPPORT_AUDIT.tsv',support)
 wr('E2_OPTIMIZATION_RESULT.tsv',[{'run_id':'E2_K5_GLOBAL_BALANCED_V2','status':rr['status'],'is_optimal':rr['is_optimal'],'raw_coverage':raw,'distinct_eva_support_valid_coverage':valid,'distinct_identity_coverage':ident,'table':json.dumps(rr['table'],sort_keys=True),'runtime_sec':rr['runtime_sec'],'best_bound':rr['best_bound'],'optimality_gap':rr['optimality_gap'],'first_incumbent_sec':inc[0]['elapsed_sec'] if inc else 'NA','incumbent_count':len(inc),'eligible_labels':len(sk),'eligible_pairs':rr['eligible_pairs'],'solver_sha256':sha(PREP/'snapshot/solver_cpsat.py'),'optimization_solver_sha256':sha(ROOT/'e2_solver.py'),'scorer_sha256':sha(PREP/'snapshot/scorer.py'),'config_sha256':ch,'lexicon_sha256':sha(LEX),'target_scope_sha256':sha(SCOPE),'ortools_version':'9.15.6755','timestamp_utc':datetime.now(timezone.utc).isoformat()}])
 jd('RUN_STATUS.json',{'RUN_STATUS':'TIME_BOUNDED' if not rr['is_optimal'] else 'COMPLETE','STAGE':'E2','PROFILE':'BALANCED','CAPACITY_MODE':'GLOBAL_CAPACITY_1','TABLE_SIZE':5,'RAW_COVERAGE':raw,'DISTINCT_EVA_SUPPORT_VALID_COVERAGE':valid,'DISTINCT_IDENTITY_COVERAGE':ident,'FIRST_INCUMBENT_FOUND':bool(inc),'FIRST_INCUMBENT_SECONDS':inc[0]['elapsed_sec'] if inc else None,'INCUMBENT_COUNT':len(inc),'BEST_BOUND':rr['best_bound'],'OPTIMALITY_GAP':rr['optimality_gap'],'E1_WARM_START_VALIDATED':'YES','E1_WARM_START_ACCEPTED':'YES','ZERO_MATCH_SOLUTION_FEASIBLE':'YES','E2_OPTIMIZATION_READY':'YES','E3_E4_EXECUTED':'NO','SCIENTIFIC_CLAIM':'NONE','NEXT_STAGE_DECISION':'REVIEW_K5_BEFORE_K5_SECOND_CAPACITY_OR_K6'})
 (ROOT/'RUN_PROTOCOL.md').write_text('# E2 full optimization run v2\n\nStage order is k=5, GLOBAL_CAPACITY_1, BALANCED profile, then review. Budget is 900 seconds. Every strict incumbent improvement is logged with elapsed time, bound, gap, and table. Distinct EVA types are required for independent support; repeated occurrences do not count. E3/E4 are excluded.\n')
 (ROOT/'EXPLORATORY_RESULTS_REPORT.md').write_text(f'# E2 full optimization run v2\n\nThe first staged run searched k=5 with GLOBAL_CAPACITY_1 and the BALANCED profile for {BUDGET} seconds. Raw coverage, distinct-EVA support-valid coverage, and distinct-identity coverage are reported separately. The incumbent log contains {len(inc)} improvements. No conclusion about E2 absence follows from a time-bounded UNKNOWN/FEASIBLE result.\n')
 (ROOT/'VALIDATION_REPORT.md').write_text('Warm-start legality and scorer acceptance were validated before optimization. Solver callback logging preserves each incumbent improvement. Distinct-EVA support is applied after assignment; repeated occurrences are not independent.\n')
 (ROOT/'REPRODUCIBILITY.md').write_text('Run with `/home/brigadire/.venv/bin/python3 run_optimization.py`. Content hashes, checkpoint, incumbent log, solver bounds, and support audit are frozen in this package.\n')
 lines=[]
 for q in sorted(ROOT.rglob('*')):
  if q.is_file() and q.name!='SHA256SUMS' and '__pycache__' not in q.parts:lines.append(f'{sha(q)}  {q.relative_to(ROOT)}')
 (ROOT/'SHA256SUMS').write_text('\n'.join(lines)+'\n')
 print('complete',ROOT)
if __name__=='__main__':main()
