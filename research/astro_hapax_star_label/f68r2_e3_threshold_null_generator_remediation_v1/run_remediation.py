#!/usr/bin/env python3
import csv, gc, hashlib, json, random, resource, sys, time
from collections import defaultdict
from pathlib import Path
ROOT=Path(__file__).resolve().parent; BASE=ROOT.parent; PREP=BASE/'exploratory_astronomical_dictionary_search_v1'; E3=BASE/'f68r2_e3_historical_operations_v1'; DEC=BASE/'f68r2_e3_decomposed_search_v1'
sys.path.insert(0,str(E3)); import run_e3
sys.path.insert(0,str(DEC)); import run_decomposed
REGISTRY=ROOT/'TRANSFORMATION_REGISTRY.tsv'; N_STAGE_A=20; SEED_BASE=90250925; THRESHOLD=5
def rows(p):
 with p.open(encoding='utf-8',newline='') as f:return list(csv.DictReader(f,delimiter='\t'))
def write(name,fields,data):
 with (ROOT/name).open('w',encoding='utf-8',newline='') as f:
  w=csv.DictWriter(f,fieldnames=fields,delimiter='\t',lineterminator='\n',extrasaction='ignore');w.writeheader();w.writerows(data)
def rss(): return resource.getrusage(resource.RUSAGE_SELF).ru_maxrss
def canonical_hash(ps):
 key=lambda x:tuple(x.get(k,'') for k in ('system_id','label_id','identity','rules','encoded'))
 return hashlib.sha256('\n'.join('\t'.join(key(x)) for x in sorted(ps,key=key)).encode()).hexdigest()
def make_labels(labels,rng):
 chars=list(''.join(x['zl3b_token'] for x in labels));rng.shuffle(chars);i=0;out=[]
 for x in labels:
  y=dict(x);n=len(x['zl3b_token']);y['zl3b_token']=''.join(chars[i:i+n]);i+=n;out.append(y)
 return out
def profile_paths(labels,lex,profile,transforms):
 out={}; t0=time.monotonic(); cache={}
 for lab in labels:
  for term in lex:
   enc,trace=transforms[(profile['system_id'],term['lexicon_id'])]; key=(enc,lab['zl3b_token'])
   if len(lab['zl3b_token'])>len(enc) or len(set(lab['zl3b_token']))>len(set(enc)): continue
   if key not in cache: cache[key]=run_e3.alignments(enc,lab['zl3b_token'])
   for sig in cache[key]:
    mapping=dict(sig)
    if run_e3.encode_word(enc,mapping,'DROP_UNMAPPED','NONE')!=lab['zl3b_token']:continue
    rules=';'.join(f'{a}->{b}' for a,b in sig); k=(lab['label_id'],term['canonical_identity_id'],rules,lab['zl3b_token'])
    out[k]={'system_id':profile['system_id'],'label_id':lab['label_id'],'eva_token':lab['zl3b_token'],'identity':term['canonical_identity_id'],'lexicon_id':term['lexicon_id'],'source_form':term['normalized_form'],'target_form':lab['zl3b_token'],'encoded':enc,'rules':rules,'trace':'NULL','scorer_parity':'PASS'}
 return list(out.values()),round(time.monotonic()-t0,4)
def main():
 labels=[x for x in rows(PREP/'TARGET_STAR_LABELS.tsv') if x['page']=='f68r2'];lex=rows(PREP/'ASTRONOMICAL_NAMES_LEXICON.tsv');reg=rows(REGISTRY)
 realmax=rows(DEC/'E3_PROFILE_MAXIMA.tsv'); first=['S000','S003','S008','S011','S016','S019','S024','S027','S043']; order=first+[x['system_id'] for x in sorted(reg,key=lambda x:(-int(next(y['relaxed_bipartite_upper_bound'] for y in rows(DEC/'E3_PROFILE_SCALING.tsv') if y['system_id']==x['system_id'])),x['system_id'])) if x['system_id'] not in first]
 transforms={(r['system_id'],t['lexicon_id']):run_e3.transform(t['normalized_form'],r) for r in reg for t in lex}
 rng=random.Random(SEED_BASE); null_labels=make_labels(labels,rng); profile_rows=[]; started=time.monotonic(); paths_count=0
 for profile_id in order:
  profile=next(r for r in reg if r['system_id']==profile_id); t0=time.monotonic(); ps,gen_sec=profile_paths(null_labels,lex,profile,transforms);paths_count+=len(ps);upper=run_decomposed.matching_bound(ps); build_t=time.monotonic();
  if upper<THRESHOLD: verdict='INFEASIBLE_CERTIFIED_RELAXED_UPPER_BOUND'; decision='INFEASIBLE_CERTIFIED'; coverage=0; bound=upper; solve_sec=0
  else:
   result=run_decomposed.solve(ps,12,60,decision=THRESHOLD); decision=result['status']; verdict='FEASIBLE' if decision in ('FEASIBLE','OPTIMAL') else 'INFEASIBLE_CERTIFIED' if decision=='INFEASIBLE' else 'INCOMPLETE';coverage=result['coverage'];bound=result['best_bound'];solve_sec=result['elapsed_sec']
  profile_rows.append({'profile':profile_id,'path_count':len(ps),'reachable_labels':len({x['label_id'] for x in ps}),'relaxed_upper':upper,'path_generation_sec':gen_sec,'model_build_plus_solve_sec':round(time.monotonic()-build_t+solve_sec,4),'peak_rss_kb':rss(),'decision_status':decision,'verdict':verdict,'replicate':0})
  if verdict=='FEASIBLE':
   null_verdict='NULL_GE_5_YES';break
  if verdict=='INCOMPLETE': null_verdict='INCOMPLETE';break
  del ps;gc.collect()
 else:null_verdict='NULL_GE_5_NO'
 write('PROFILE_STREAM_TRACE.tsv',list(profile_rows[0]),profile_rows);write('ONE_NULL_PROFILE_RESULTS.tsv',list(profile_rows[0]),profile_rows)
 status={'ONE_NULL_REPLICATE_COMPLETE':'YES' if null_verdict!='INCOMPLETE' else 'NO','NULL_GE_5':null_verdict,'PROFILES_PROCESSED':len(profile_rows),'PROFILES_TOTAL':64,'PROFILE_STREAMING_IMPLEMENTED':'YES','NULL_GENERATION_BOTTLENECK':'PROFILE_PATH_GENERATION','PEAK_MEMORY_KB':max(x['peak_rss_kb'] for x in profile_rows),'PATHS_PROCESSED':paths_count,'ELAPSED_SEC':round(time.monotonic()-started,3)}
 (ROOT/'ONE_NULL_STATUS.json').write_text(json.dumps(status,indent=2,sort_keys=True)+'\n')
 (ROOT/'INPUT_FREEZE.json').write_text(json.dumps({'scope_sha256':hashlib.sha256((PREP/'TARGET_STAR_LABELS.tsv').read_bytes()).hexdigest(),'lexicon_sha256':hashlib.sha256((PREP/'ASTRONOMICAL_NAMES_LEXICON.tsv').read_bytes()).hexdigest(),'registry_sha256':hashlib.sha256(REGISTRY.read_bytes()).hexdigest(),'threshold':THRESHOLD,'seed':SEED_BASE,'profile_order':order,'null':'global character shuffle preserving 27 lengths'},indent=2,sort_keys=True)+'\n')
if __name__=='__main__':main()
