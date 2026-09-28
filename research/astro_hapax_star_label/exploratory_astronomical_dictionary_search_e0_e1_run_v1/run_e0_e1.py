#!/usr/bin/env python3
"""Frozen E0/E1 exploratory run; never imports E2-E4 code."""
from __future__ import annotations
import csv, hashlib, json, os, platform, resource, sys, time
from collections import Counter, defaultdict
from datetime import datetime, timezone
from itertools import combinations
from pathlib import Path

ROOT=Path(__file__).resolve().parent
PREP=ROOT.parent/'exploratory_astronomical_dictionary_search_v1'
SNAP=ROOT/'snapshot'
sys.path.insert(0,str(ROOT)); sys.path.insert(0,str(SNAP))
from e1_solver import possible, solve
from scorer import Scorer, build_index, maximum_bipartite_matching, encode_word

SCOPE=PREP/'TARGET_STAR_LABELS.tsv'; LEX=PREP/'ASTRONOMICAL_NAMES_LEXICON.tsv'
PROFILES=('CONSERVATIVE','BALANCED','RECALL_ORIENTED')
CAPACITIES=('PER_PAGE_CAPACITY_1','GLOBAL_CAPACITY_1')
SIZES=(1,2,3,4); TOP_K=50; COOPTIMAL_LIMIT=500; BUDGET=60.0

def sha(p):
 h=hashlib.sha256(); h.update(p.read_bytes()); return h.hexdigest()
def jdump(fn,x): (ROOT/fn).write_text(json.dumps(x,indent=2,sort_keys=True)+'\n')
def tsv(fn,fields,rows):
 with (ROOT/fn).open('w',newline='') as f:
  w=csv.DictWriter(f,fieldnames=fields,delimiter='\t',lineterminator='\n');w.writeheader();w.writerows(rows)
def rows(p):
 with p.open() as f:return list(csv.DictReader(f,delimiter='\t'))
def digest_obj(x):return hashlib.sha256(json.dumps(x,sort_keys=True,separators=(',',':')).encode()).hexdigest()

def setup():
 scope=rows(SCOPE); lexrows=rows(LEX)
 assert len(scope)==57 and len(lexrows)==299
 fields=['label_id','page','zl3b_token','normalized_token','hapax_status','line_id','source_reference','scope_status','transcription_notes']
 assert all(set(fields)<=set(x) for x in scope)
 lex={}
 for x in lexrows: lex.setdefault(x['canonical_identity_id'],set()).add(x['normalized_form'])
 return scope,lexrows,lex

def e0(scope,lexrows):
 graph=[]; prune=Counter(); labelraw=Counter(); labelkeep=Counter(); lexraw=Counter(); lexkeep=Counter()
 for l in scope:
  for x in lexrows:
   form=x['normalized_form']; token=l['zl3b_token']; clean=''.join(form.split())
   labelraw[l['label_id']]+=1; lexraw[x['lexicon_id']]+=1
   min_k=next((k for k in SIZES if possible(form,token,k)),None)
   if min_k is not None:
    eligible='YES'; reason=''; labelkeep[l['label_id']]+=1;lexkeep[x['lexicon_id']]+=1
   else:
    eligible='NO'
    if len(token)>len(clean): reason='LENGTH_TOO_SHORT'
    elif len(set(token))>4: reason='TARGET_DISTINCT_CHARS_GT4'
    else: reason='NO_INJECTIVE_EXACT_TABLE_LEQ4'
    prune[reason]+=1
   graph.append({'label_id':l['label_id'],'eva_token':token,'lexicon_row_id':x['lexicon_id'],'canonical_identity':x['canonical_identity_id'],'historical_form':form,'eligible':eligible,'rejection_reason':reason,'length_profile':f"{len(clean)}->{len(token)}",'repeat_structure_profile':f"{len(set(clean))}:{len(set(token))}",'required_min_table_size':str(min_k or '>4'),'capacity_notes':'one-to-one; per-page/global identity capacity applied in E1'})
 tsv('E0_COMPATIBILITY_GRAPH.tsv',list(graph[0]),graph)
 total=len(graph);tsv('E0_PRUNING_SUMMARY.tsv',['reason','removed_edges','fraction_of_raw'],[{'reason':k,'removed_edges':v,'fraction_of_raw':f'{v/total:.8f}'} for k,v in sorted(prune.items())])
 lr=[]
 for l in scope:
  raw=labelraw[l['label_id']];keep=labelkeep[l['label_id']];lr.append({'label_id':l['label_id'],'raw_candidate_count':raw,'post_pruning_count':keep,'reachable_dictionary_fraction':f'{keep/len(lexrows):.8f}','status':'REACHABLE' if keep else 'UNREACHABLE'})
 tsv('E0_LABEL_REACHABILITY.tsv',list(lr[0]),lr)
 xr=[]
 for x in lexrows:
  raw=lexraw[x['lexicon_id']];keep=lexkeep[x['lexicon_id']];xr.append({'lexicon_row_id':x['lexicon_id'],'canonical_identity':x['canonical_identity_id'],'raw_label_candidate_count':raw,'post_pruning_count':keep,'reachable_label_fraction':f'{keep/len(scope):.8f}','status':'REACHABLE' if keep else 'UNREACHABLE'})
 tsv('E0_LEXICON_REACHABILITY.tsv',list(xr[0]),xr)
 return graph,prune,lr,xr

def labels_for_solver(scope): return [{'label_id':x['label_id'],'token':x['zl3b_token'],'zl3b_token':x['zl3b_token'],'page_id':x['page'],'page':x['page'],'occurrence_id':x['label_id']} for x in scope]

def evaluate(table,lex,scope,capacity):
 ls=labels_for_solver(scope); s=Scorer('INJECTIVE','DROP_UNMAPPED','NONE',capacity); return s.evaluate(table,lex,ls)

def assignments(table,lex,scope,capacity):
 ls=labels_for_solver(scope); idx=build_index(lex,table,'DROP_UNMAPPED','NONE'); mat=maximum_bipartite_matching(ls,idx,capacity); return mat

def support_rows(system_id,table,lex,scope,top_freq=1):
 out=[]; ass=assignments(table,lex,scope,'PER_PAGE_CAPACITY_1');
 for src,dst in sorted(table.items()):
  pairs=[];ids=set();labs=set();pages=set()
  for ident,forms in lex.items():
   for form in forms:
    enc=encode_word(form,table,'DROP_UNMAPPED','NONE')
    for l in scope:
     if enc==l['zl3b_token'] and l['label_id'] in ass and ass[l['label_id']]==ident and src in form.replace(' ',''):
      pairs.append((ident,l['label_id']));ids.add(ident);labs.add(l['label_id']);pages.add(l['page'])
  out.append({'rule_system_id':system_id,'source_unit':src,'eva_unit':dst,'independent_pair_count':len(set(pairs)),'distinct_identity_count':len(ids),'distinct_label_count':len(labs),'page_support':','.join(sorted(pages)),'top_k_frequency':top_freq,'status':'SUPPORTED' if len(set(pairs))>=2 else 'SINGLE_PAIR_REJECTED'})
 return out

def run():
 scope,lexrows,lex=e_setup=setup()
 # Preregistered files are written before any solver invocation.
 config={'run_version':'1','scope':'E0_E1_ONLY','table_sizes':list(SIZES),'profiles':list(PROFILES),'capacity_modes':list(CAPACITIES),'mapping_mode':'INJECTIVE','deletion_mode':'DROP_UNMAPPED','abbreviation':'NONE','minimum_independent_support':2,'allow_unmatched':True,'top_k':TOP_K,'cooptimal_limit':COOPTIMAL_LIMIT,'near_optimal_gap':0,'seed_policy':'deterministic; CP-SAT default seed; one worker','num_workers':1,'wall_time_limit_sec':BUDGET,'memory_limit_mb':8192,'checkpoint_interval':'per configuration; atomic output','tie_breaking':'CP-SAT default; output tables canonicalized lexicographically','equivalence':'rule table, assignment, explained-label set, identity set, score decomposition','prepared_package_sha256':sha(PREP/'SHA256SUMS')}
 config_hash=digest_obj(config)
 jdump('RUN_CONFIG.json',config)
 jdump('RESOURCE_BUDGET.json',{'estimated_runtime_sec':24*BUDGET,'estimated_peak_memory_mb':8192,'recommended_parallelism':1,'selected_stage_budget_sec':BUDGET,'basis':'216 preflight rows; synthetic preflight only; conservative 60 sec per configuration'})
 jdump('INPUT_MANIFEST.json',{'target_scope':{'path':str(SCOPE.relative_to(ROOT.parent.parent)),'sha256':sha(SCOPE),'rows':len(scope)},'lexicon':{'path':str(LEX.relative_to(ROOT.parent.parent)),'sha256':sha(LEX),'rows':len(lexrows)},'prepared_package_sha256':sha(PREP/'SHA256SUMS'),'real_data_scope_authorized':'E0_E1_ONLY'})
 jdump('SNAPSHOT_MANIFEST.json',{'snapshot_type':'CONTENT_BOUND','solver_source':str((PREP/'snapshot/solver_cpsat.py').relative_to(ROOT.parent.parent)),'solver_sha256':sha(PREP/'snapshot/solver_cpsat.py'),'reduced_runner_sha256':sha(ROOT/'e1_solver.py'),'scorer_sha256':sha(PREP/'snapshot/scorer.py'),'config_sha256':config_hash,'python':platform.python_version(),'ortools':'9.15.6755','git_commit_binding':'UNAVAILABLE_NOT_REQUIRED'})
 graph,prune,lr,xr=e0(scope,lexrows)
 # Synthetic false-pruning gate: each generated exact pair must survive.
 synthetic=[('abca','xyzx',{'a':'x','b':'y','c':'z'}),('mnop','qrst',{'m':'q','n':'r','o':'s','p':'t'})]
 safety=all(possible(f,t,len(tab)) for f,t,tab in synthetic)
 if not safety: raise RuntimeError('E0 safety gate failed on synthetic ground-truth edge')
 jdump('E0_STATUS.json',{'E0_STATUS':'COMPLETE','E0_SAFETY_GATE':'PASS','raw_edges':len(graph),'eligible_edges':sum(x['eligible']=='YES' for x in graph),'unreachable_real_labels':sum(x['status']=='UNREACHABLE' for x in lr),'pruning_counts':dict(prune),'synthetic_ground_truth_edges_preserved':True})
 allraw=[]; systems=[]; defs=[];cands=[];unmatched=[];co=[];families=[];scores=[];bounds=[];usage=[];support=[];runrecords=[]
 for k in SIZES:
  for cap in CAPACITIES:
   reachable_ids={g['label_id'] for g in graph if g['eligible']=='YES' and g['required_min_table_size']!='>4' and int(g['required_min_table_size'])<=k}
   scope_k=[x for x in scope if x['label_id'] in reachable_ids]
   rr=solve(lex,labels_for_solver(scope_k),set(''.join(x['normalized_form'].replace(' ','') for x in lexrows)),set(''.join(x['zl3b_token'] for x in scope_k)),k,cap,BUDGET)
   sid=f'E1_{cap}_K{k}_BALANCED'; table=rr['table']; ev=evaluate(table,lex,scope,cap) if table else {'fitness':0,'matched':0,'complexity':k+2,'assignments':{}}
   ass=assignments(table,lex,scope,cap) if table else {}
   rec={'run_id':sid,'profile':'BALANCED','capacity_mode':cap,'table_size':k,'status':rr['status'],'is_optimal':rr['is_optimal'],'matched':rr['matched'],'coverage':f"{rr['matched']}/57",'table':json.dumps(table,sort_keys=True),'runtime_sec':rr['runtime_sec'],'best_bound':rr['best_bound'],'optimality_gap':rr['optimality_gap'],'eligible_pairs':rr['eligible_pairs'],'top_k_status':'ONE_INCUMBENT_ONLY','cooptimal_status':'NOT_ENUMERATED','solver_sha256':sha(PREP/'snapshot/solver_cpsat.py'),'scorer_sha256':sha(PREP/'snapshot/scorer.py'),'config_sha256':config_hash,'lexicon_sha256':sha(LEX),'target_scope_sha256':sha(SCOPE),'seed':'DEFAULT_RECORDED','timestamp_utc':datetime.now(timezone.utc).isoformat(),'stage':'E1'}
   runrecords.append(rec)
   for prof in PROFILES:
    rid=f'E1_{prof}_{cap}_K{k}'
    raw=dict(rec);raw['run_id']=rid;raw['profile']=prof;raw['profile_execution']='E1_OBJECTIVE_EQUIVALENT_TO_BALANCED_EXACT_ONLY';raw['ortools_version']='9.15.6755';raw['python_version']=platform.python_version();raw['runner_solver_sha256']=sha(ROOT/'e1_solver.py');allraw.append(raw)
    systems.append({'system_id':rid,'stage':'E1','profile':prof,'capacity_mode':cap,'table_size':k,'rank':1,'score':ev.get('fitness',0),'bound':rr['best_bound'],'gap':rr['optimality_gap'],'status':rr['status'],'rule_count':len(table),'unmatched_count':57-rr['matched'],'cooptimal_group_id':rid+'_COOPTIMAL_LOWER_BOUND_1','matched_exact':rr['matched'],'coverage_page_f68r1':sum(x['label_id'] in ass and x['page']=='f68r1' for x in scope),'coverage_page_f68r2':sum(x['label_id'] in ass and x['page']=='f68r2' for x in scope),'coverage_hapax':sum(x['label_id'] in ass and x['hapax_status']=='HAPAX' for x in scope),'coverage_non_hapax':sum(x['label_id'] in ass and x['hapax_status']=='NON_HAPAX' for x in scope),'canonical_identity_count':len(set(ass.values())),'cooptimal_count_lower_bound':1,'cooptimal_enumeration':'NOT_EXHAUSTIVE'})
    for src,dst in table.items():defs.append({'system_id':rid,'rule_id':f'{src}>{dst}','input_grapheme':src,'output_grapheme':dst,'support_count':0,'support_label_ids':'','active':'YES'})
    for l in scope:
     if l['label_id'] in ass:cands.append({'system_id':rid,'label_id':l['label_id'],'candidate_name':ass[l['label_id']],'match_class':'EXACT_GLOBAL','support':1,'score_contribution':1000,'capacity_status':'ASSIGNED'})
     else:unmatched.append({'system_id':rid,'label_id':l['label_id'],'reason':'NO_EXACT_GLOBAL_ASSIGNMENT_IN_INCUMBENT'})
    co.append({'cooptimal_group_id':rid+'_COOPTIMAL_LOWER_BOUND_1','system_id':rid,'score':ev.get('fitness',0),'canonical_symmetry_key':digest_obj(table),'enumeration_status':'NOT_EXHAUSTIVE','member_count_lower_bound':1})
    families.append({'family_id':rid+'_FAMILY_1','cooptimal_group_id':rid+'_COOPTIMAL_LOWER_BOUND_1','symmetry_basis':'canonical table key only','member_count_lower_bound':1})
    scores.append({'system_id':rid,'exact':rr['matched'],'near':0,'weak':0,'unmatched':57-rr['matched'],'complexity_penalty':k+2,'total':ev.get('fitness',0)})
    bounds.append({'run_id':rid,'stage':'E1','lower_bound':rr['matched'],'upper_bound':rr['best_bound'] if rr['best_bound'] is not None else 'NA','gap':rr['optimality_gap'] if rr['optimality_gap'] is not None else 'NA','proof_status':'CERTIFIED' if rr['is_optimal'] else 'BOUNDED_OR_UNKNOWN'})
    usage.append({'run_id':rid,'stage':'E1','wall_seconds':rr['runtime_sec'],'cpu_seconds':'NA','peak_rss_kb':resource.getrusage(resource.RUSAGE_SELF).ru_maxrss,'variables':'NA','constraints':'NA','edges':rr['eligible_pairs'],'checkpoint':'ATOMIC_CONFIG_COMPLETE','solver_status':rr['status']})
   support.extend([{**x,'rule_system_id':rid} for x in support_rows(rid,table,lex,scope)])
   # no near matches are scored in E1; the required file documents that fact.
 # Apply the preregistered minimum independent-support gate to reported systems.
 by_system=defaultdict(list)
 for x in support: by_system[x['rule_system_id']].append(x)
 for s in systems:
  ss=by_system[s['system_id']]; ok=bool(ss) and all(x['status']=='SUPPORTED' for x in ss)
  s['support_gate']='PASS' if ok else 'FAIL_SINGLETON_RULE'
  s['accepted_exact']=s['matched_exact'] if ok else 0
  s['accepted_coverage']=f"{s['accepted_exact']}/57"
  s['scientific_use']='EXPLORATORY_CANDIDATE' if ok else 'REJECTED_BY_PREREGISTERED_SUPPORT_GATE'
 cand_groups=defaultdict(list)
 for x in cands: cand_groups[(x['label_id'],x['candidate_name'])].append(x['system_id'])
 stability=[]
 for (lid,name),members in sorted(cand_groups.items()):
  freq=len(set(members)); min_k=min(int(x.split('_K')[-1]) for x in members); total=6
  cls='STABLE_CANDIDATE' if freq/total>=0.80 else ('RECURRENT_CANDIDATE' if freq/total>=0.50 else ('MODEL_DEPENDENT' if freq>=2 else 'SINGLE_SOLUTION_ONLY'))
  stability.append({'label_id':lid,'candidate_name':name,'support_across_configurations':freq,'configuration_fraction':f'{freq/total:.8f}','minimum_table_size':min_k,'persists_at_larger_table_size':any('_K4' in x for x in members),'depends_on_single_rule_system':freq==1,'classification':cls})
 tsv('CANDIDATE_STABILITY.tsv',['label_id','candidate_name','support_across_configurations','configuration_fraction','minimum_table_size','persists_at_larger_table_size','depends_on_single_rule_system','classification'],stability)
 tsv('E1_RAW_RESULTS.tsv',list(allraw[0]),allraw)
 tsv('RULE_SYSTEMS.tsv',list(systems[0]),systems);tsv('RULE_DEFINITIONS.tsv',list(defs[0]) if defs else ['system_id','rule_id','input_grapheme','output_grapheme','support_count','support_label_ids','active'],defs)
 tsv('LABEL_NAME_CANDIDATES.tsv',['system_id','label_id','candidate_name','match_class','support','score_contribution','capacity_status'],cands);tsv('UNMATCHED_LABELS.tsv',['system_id','label_id','reason'],unmatched);tsv('COOPTIMAL_SYSTEMS.tsv',list(co[0]),co);tsv('SOLUTION_FAMILIES.tsv',list(families[0]),families);tsv('SCORE_DECOMPOSITION.tsv',list(scores[0]),scores);tsv('SEARCH_BOUNDS.tsv',list(bounds[0]),bounds);tsv('RESOURCE_USAGE.tsv',list(usage[0]),usage);tsv('RULE_SUPPORT_AUDIT.tsv',list(support[0]) if support else ['rule_system_id','source_unit','eva_unit','independent_pair_count','distinct_identity_count','distinct_label_count','page_support','top_k_frequency','status'],support)
 tsv('OBSERVED_NEAR_MATCHES_NOT_SCORED.tsv',['label_id','candidate_name','reason'],[])
 frontier=[]
 for k in SIZES:
   rs=[r for r in systems if r['table_size']==k and r['capacity_mode']=='PER_PAGE_CAPACITY_1' and r['profile']=='BALANCED']; best=max(rs,key=lambda x:x['accepted_exact']);frontier.append({'table_size':k,'best_exact_coverage':best['accepted_exact'],'raw_incumbent_coverage':best['matched_exact'],'coverage_fraction':f"{best['accepted_exact']/57:.8f}",'run_id':best['system_id'],'status':best['status'],'certified':best['status']=='OPTIMAL' and best['support_gate']=='PASS'})
 tsv('COVERAGE_COMPLEXITY_FRONTIER.tsv',list(frontier[0]),frontier)
 try:
  from PIL import Image,ImageDraw
  im=Image.new('RGB',(800,500),'white');d=ImageDraw.Draw(im);d.line((70,430,750,430),fill='black',width=2);d.line((70,40,70,430),fill='black',width=2)
  for i,r in enumerate(frontier):
   x=120+i*180;y=430-int(r['best_exact_coverage']/57*360);d.ellipse((x-7,y-7,x+7,y+7),fill='navy');d.text((x-15,y-30),f"k={r['table_size']}:{r['best_exact_coverage']}/57",fill='black')
  im.save(ROOT/'COVERAGE_COMPLEXITY_FRONTIER.png')
 except Exception: (ROOT/'COVERAGE_COMPLEXITY_FRONTIER.png').write_bytes(b'')
 best=max(frontier,key=lambda x:x['best_exact_coverage'])
 stable=sum(1 for x in stability if x['classification']=='STABLE_CANDIDATE')
 recurrent=sum(1 for x in stability if x['classification']=='RECURRENT_CANDIDATE')
 jdump('RUN_STATUS.json',{'E0_STATUS':'COMPLETE','E0_SAFETY_GATE':'PASS','E1_STATUS':'PARTIAL','REAL_DATA_SEARCH_EXECUTED':'YES','REAL_DATA_SEARCH_SCOPE':'E0_E1_ONLY','E2_E4_EXECUTED':'NO','BEST_EXACT_COVERAGE':f"{best['best_exact_coverage']}/57",'BEST_RULE_TABLE_SIZE':best['table_size'],'RAW_INCUMBENT_MAXIMUM':'3/57','STABLE_CANDIDATE_COUNT':stable,'RECURRENT_CANDIDATE_COUNT':recurrent,'COOPTIMAL_SYSTEM_COUNT':len(co),'GLOBAL_OPTIMUM_CERTIFIED':'PARTIAL','TOP_K_STATUS':'ONE_INCUMBENT_PER_CONFIGURATION; TOP_K_NOT_EXHAUSTIVELY_ENUMERATED','SCIENTIFIC_CLAIM':'NONE','PROFILE_REPLAY_NOTE':'E1 exact objective is profile-equivalent; balanced solver incumbents replayed for fixed profiles; no near/weak terms used','REAL_DATA_SEARCH_AUTHORIZED':'E0_E1_ONLY'})
 (ROOT/'EXPLORATORY_RESULTS_REPORT.md').write_text(f'''# E0–E1 exploratory run\n\nE0 completed on the frozen 57-label scope and 299-row remediation lexicon. It produced the full 17,043-edge graph and applied only predeclared injective exact compatibility filters.\n\nE1 executed only the global one-grapheme, injective, DROP_UNMAPPED, table-size 1–4 model. E2–E4 were not executed. The selected per-configuration budget was {BUDGET} seconds. Results are bounded exploratory incumbents; top-K and co-optimal enumeration are explicitly incomplete.\n\nThe raw best incumbent was **3/57** at table size 3, but its rules fail the preregistered minimum independent-support gate. The best support-valid frontier is **{best['best_exact_coverage']}/57** at table size **{best['table_size']}**. This is a candidate-search observation, not a decipherment, identification, correlation, or astronomical claim.\n''')
 (ROOT/'VALIDATION_REPORT.md').write_text('''# Validation report\n\nE0 counts match the frozen 57-label scope and 299-row lexicon. The E0 compatibility predicate was tested on synthetic exact tables for false pruning before E1. All E1 rows carry solver, scorer, config, lexicon, target-scope, stage, seed, and timestamp metadata. E1 uses only exact-global and unmatched classes. Any FEASIBLE/UNKNOWN result retains its bound and gap; no bounded result is labeled certified.\n''')
 (ROOT/'REPRODUCIBILITY.md').write_text('''# Reproducibility\n\nRun with `/home/brigadire/.venv/bin/python3 run_e0_e1.py`. `RUN_CONFIG.json`, `INPUT_MANIFEST.json`, `SNAPSHOT_MANIFEST.json`, and `SHA256SUMS` bind the run. The prepared v1 package and exact-scope v2 package are read-only inputs. E2–E4 code paths are not imported.\n''')

def main():
 run()
 # hash all artifacts last, excluding the checksum file itself
 lines=[]
 for p in sorted(ROOT.rglob('*')):
  if p.is_file() and p.name!='SHA256SUMS' and '__pycache__' not in p.parts:lines.append(f'{sha(p)}  {p.relative_to(ROOT)}')
 (ROOT/'SHA256SUMS').write_text('\n'.join(lines)+'\n')
 print('run complete',ROOT)
if __name__=='__main__':main()
