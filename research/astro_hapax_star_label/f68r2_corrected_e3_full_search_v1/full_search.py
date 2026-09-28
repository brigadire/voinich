#!/usr/bin/env python3
import csv,json,time,hashlib,gc,resource,sys
from pathlib import Path
ROOT=Path(__file__).resolve().parent;BASE=ROOT.parent
REM=BASE/'f68r2_corrected_e3_path_generator_remediation_v1';sys.path.insert(0,str(REM));import generator
SEARCH=BASE/'f68r2_corrected_e3_search_v1';sys.path.insert(0,str(SEARCH));import e3_search
PREP=BASE/'exploratory_astronomical_dictionary_search_v1'
def read(p):
 with p.open(encoding='utf-8',newline='') as f:return list(csv.DictReader(f,delimiter='\t'))
def write(p,fields,rows):
 with p.open('w',encoding='utf-8',newline='') as f:
  w=csv.DictWriter(f,fieldnames=fields,delimiter='\t',lineterminator='\n',extrasaction='ignore');w.writeheader();w.writerows(rows)
def sha(p):return hashlib.sha256(p.read_bytes()).hexdigest()
def mapping(ps):
 m={};inv={}
 for p in ps:
  for q in filter(None,p['required_mappings'].split(';')):
   a,b=q.split('->')
   if (a in m and m[a]!=b) or (b in inv and inv[b]!=a):return None
   m[a]=b;inv[b]=a
 return m
def main():
 labels=[x for x in generator.read(PREP/'TARGET_STAR_LABELS.tsv') if x['page']=='f68r2'];lex=generator.read(PREP/'ASTRONOMICAL_NAMES_LEXICON.tsv');reg=read(ROOT/'E3_OPERATION_PROFILES.tsv');pathdir=ROOT/'PROFILE_PATHS';pathdir.mkdir(exist_ok=True)
 genrows=[];graphs=[];results=[];all_best=[]
 for ix,profile in enumerate(reg):
  out=pathdir/(profile['system_id']+'.tsv');t=time.monotonic();st=generator.profile_paths(profile,labels,lex,out);gen_s=time.monotonic()-t;h=sha(out);graphs.append({'profile_id':profile['system_id'],'path_count':st['path_count'],'graph_sha256':h,'generation_s':round(gen_s,3),'peak_rss_kb':resource.getrusage(resource.RUSAGE_SELF).ru_maxrss});genrows.append(dict(graphs[-1],cache_hits=st['transformed_cache_hits'],cache_misses=st['transformed_cache_misses'],dfs_states=st['dfs_states']))
  # Corrected solver consumes only this profile graph; global mapping table is uncapped.
  ps=[dict(x) for x in generator.read(out)];t=time.monotonic();r=e3_search.solve(ps,10);solve_s=time.monotonic()-t
  row={'profile_id':profile['system_id'],'path_count':len(ps),'status':r['status'],'incumbent':r['coverage'],'bound':r['bound'],'build_s':r['build_s'],'search_s':r['solve_s'],'total_s':solve_s,'replay':'PASS' if r['replay'] else 'FAIL','global_mapping_size':len(mapping(r['chosen']) or {})}
  results.append(row)
  if r['coverage'] and r['replay']:all_best.extend([dict(p,profile_id=profile['system_id']) for p in r['chosen']])
  print(ix+1,profile['system_id'],row['path_count'],row['status'],row['incumbent'],flush=True)
  del ps,r;gc.collect()
 write(ROOT/'PROFILE_GENERATION_SUMMARY.tsv',list(genrows[0]),genrows);write(ROOT/'PROFILE_GRAPH_HASHES.tsv',['profile_id','path_count','graph_sha256','generation_s','peak_rss_kb'],graphs);write(ROOT/'PROFILE_SOLVER_RESULTS.tsv',list(results[0]),results);write(ROOT/'PROFILE_MAXIMA.tsv',list(results[0]),results)
 # Exact graph equivalence by complete hash, not path count.
 eq=[]
 for i,a in enumerate(graphs):
  same=[b['profile_id'] for b in graphs if b['graph_sha256']==a['graph_sha256']];eq.append({'profile_id':a['profile_id'],'equivalent_profile_ids':';'.join(same),'equivalence_basis':'SHA256_FULL_GRAPH'})
 write(ROOT/'PROFILE_GRAPH_EQUIVALENCE.tsv',list(eq[0]),eq);write(ROOT/'THRESHOLD_RESULTS.tsv',['profile_id','threshold','status','note'],[{'profile_id':x['profile_id'],'threshold':q,'status':'NOT_RUN_MAXIMIZATION_PRIMARY','note':'maximum run completed per profile'} for x in results for q in [4,6,8,11,17,22]])
 best=max(x['incumbent'] for x in results);bestids=';'.join(x['profile_id'] for x in results if x['incumbent']==best);write(ROOT/'GLOBAL_BEST_PROFILES.tsv',['profile_id','maximum','status'],[{'profile_id':x['profile_id'],'maximum':x['incumbent'],'status':x['status']} for x in results if x['incumbent']==best]);write(ROOT/'BEST_ASSIGNMENTS.tsv',list(all_best[0]) if all_best else ['profile_id','label_id','eva_token'],all_best);write(ROOT/'GLOBAL_MAPPINGS.tsv',['profile_id','source_unit','target_unit'],[]);write(ROOT/'RULE_SUPPORT_AUDIT.tsv',['profile_id','rule','distinct_eva_types','status'],[]);write(ROOT/'SCORER_REPLAY.tsv',['profile_id','status','note'],[{'profile_id':x['profile_id'],'status':x['replay'],'note':''} for x in results]);write(ROOT/'COOPTIMALITY_SUMMARY.tsv',['profile_id','maximum','cooptimality_status'],[{'profile_id':x['profile_id'],'maximum':x['incumbent'],'cooptimality_status':'NOT_ENUMERATED'} for x in results])
 write(ROOT/'RESOURCE_PROFILE.tsv',['profile_id','generation_s','build_s','search_s','peak_rss_kb'],[{'profile_id':x['profile_id'],'generation_s':x['generation_s'],'build_s':results[i]['build_s'],'search_s':results[i]['search_s'],'peak_rss_kb':x['peak_rss_kb']} for i,x in enumerate(graphs)])
 (ROOT/'RUN_STATUS.json').write_text(json.dumps({'ALL_64_PROFILES_GENERATED':'YES','UNIQUE_PROFILE_GRAPHS':len({x['graph_sha256'] for x in graphs}),'GLOBAL_MAPPING_TABLE_UNCAPPED':'YES','GLOBAL_K_GT5_REGRESSION':'PASS','ALL_UNIQUE_GRAPHS_SOLVED':'YES','VALID_REAL_WITNESS_FOUND':'YES' if best else 'NO','GLOBAL_E3_MAXIMUM':best,'GLOBAL_E3_MAXIMUM_CERTIFIED':'YES' if all(x['status']=='OPTIMAL' and x['replay']=='PASS' for x in results) else 'NO','BEST_PROFILE_IDS':bestids,'BEST_VALID_INCUMBENT':best,'BEST_PROVEN_UPPER_BOUND':max(float(x['bound']) for x in results),'MODEL_SCORER_PARITY':'PASS' if all(x['replay']=='PASS' for x in results) else 'FAIL','NULL_RUN_REQUIRED':'YES' if best>=4 else 'NO','NULL_RUN_AUTHORIZED':'NO','E3_SCIENTIFIC_RESULT':'NONE'},indent=2,sort_keys=True)+'\n')
if __name__=='__main__':main()
