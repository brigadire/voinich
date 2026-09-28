#!/usr/bin/env python3
import csv, json, random, hashlib, time
from collections import defaultdict
from pathlib import Path
import sys
ROOT=Path(__file__).resolve().parent; BASE=ROOT.parent
E3=BASE/'f68r2_e3_historical_operations_v1'; DEC=BASE/'f68r2_e3_decomposed_search_v1'; PREP=BASE/'exploratory_astronomical_dictionary_search_v1'
sys.path.insert(0,str(E3)); import run_e3
sys.path.insert(0,str(DEC)); import run_decomposed
N_REPS=20; SEED_BASE=90250925
def rows(p):
 with p.open(encoding='utf-8',newline='') as f:return list(csv.DictReader(f,delimiter='\t'))
def write(name,fields,data):
 with (ROOT/name).open('w',encoding='utf-8',newline='') as f:
  w=csv.DictWriter(f,fieldnames=fields,delimiter='\t',lineterminator='\n',extrasaction='ignore');w.writeheader();w.writerows(data)
def make_labels(labels,rng):
 chars=list(''.join(x['zl3b_token'] for x in labels)); rng.shuffle(chars); pos=0; out=[]
 for x in labels:
  n=len(x['zl3b_token']); y=dict(x); y['zl3b_token']=''.join(chars[pos:pos+n]); pos+=n; out.append(y)
 return out
def graph(labels,lex,registry):
 out=[]; cache={}
 for rule in registry:
  for lab in labels:
   for term in lex:
    enc,_=run_e3.transform(term['normalized_form'],rule); key=(enc,lab['zl3b_token'])
    if key not in cache: cache[key]=run_e3.alignments(*key)
    for sig in cache[key]:
     mapping=dict(sig)
     if run_e3.encode_word(enc,mapping,'DROP_UNMAPPED','NONE')!=lab['zl3b_token']:continue
     out.append({'system_id':rule['system_id'],'label_id':lab['label_id'],'eva_token':lab['zl3b_token'],'identity':term['canonical_identity_id'],'lexicon_id':term['lexicon_id'],'source_form':term['normalized_form'],'target_form':lab['zl3b_token'],'encoded':enc,'rules':';'.join(f'{a}->{b}' for a,b in sig),'trace':'NULL','alternative_count':1,'scorer_parity':'PASS'})
 return out
def main():
 labels=[x for x in rows(PREP/'TARGET_STAR_LABELS.tsv') if x['page']=='f68r2']; lex=rows(PREP/'ASTRONOMICAL_NAMES_LEXICON.tsv'); reg=rows(ROOT/'TRANSFORMATION_REGISTRY.tsv')
 all_rows=[]; best_assign=[]; started=time.monotonic()
 for rep in range(N_REPS):
  rng=random.Random(SEED_BASE+rep); ls=make_labels(labels,rng); paths=graph(ls,lex,reg); by=defaultdict(list)
  for p in paths:by[p['system_id']].append(p)
  best=None
  for r in reg:
   ps=by[r['system_id']]; upper=run_decomposed.matching_bound(ps)
   if upper<5: result={'status':'INFEASIBLE_CERTIFIED_RELAXED_BOUND','coverage':0,'best_bound':upper,'chosen':[],'elapsed_sec':0,'mapping_count':0}
   else:
    result=run_decomposed.solve(ps,12,30); result['mapping_count']=len({z for p in result['chosen'] for z in p['rules'].split(';')})
   row={'replicate':rep,'seed':SEED_BASE+rep,'profile':r['system_id'],'operation_complexity':r['complexity'],'abbreviation':r['abbreviation'],'status':result['status'],'coverage':result['coverage'],'best_bound':result['best_bound'],'gap':result.get('gap',0),'mapping_count':result.get('mapping_count',0),'path_count':len(ps)};all_rows.append(row)
   if best is None or (row['coverage'],-row['mapping_count'],row['profile'])>(best['coverage'],-best['mapping_count'],best['profile']):best=row
  best_assign.append(best)
  if rep%10==9: print(json.dumps({'replicate':rep+1,'best_so_far':max(x['coverage'] for x in best_assign),'elapsed_sec':round(time.monotonic()-started,1)}),flush=True)
 write('NULL_PROFILE_RESULTS.tsv',['replicate','seed','profile','operation_complexity','abbreviation','status','coverage','best_bound','gap','mapping_count','path_count'],all_rows)
 write('NULL_BEST_BY_REPLICATE.tsv',list(best_assign[0]),best_assign)
 dist=defaultdict(int)
 for x in best_assign:dist[x['coverage']]+=1
 write('NULL_MAXIMUM_DISTRIBUTION.tsv',['coverage','replicates','fraction'],[{'coverage':k,'replicates':v,'fraction':v/N_REPS} for k,v in sorted(dist.items())])
 status={'NULL_TYPE':'POST_HOC_EXPLORATORY_MODEL_SELECTION_NULL','NULL_FAMILY':'GLOBAL_CHARACTER_SHUFFLE_PRESERVE_LENGTHS','REPLICATES':N_REPS,'SEED_BASE':SEED_BASE,'PROFILE_SELECTION':'ALL_64_PROFILES_PER_REPLICATE','OBSERVED_COVERAGE':5,'NULL_GE_5':sum(x['coverage']>=5 for x in best_assign),'NULL_GE_6':sum(x['coverage']>=6 for x in best_assign),'SCIENTIFIC_INFERENCE':'NONE','COMPLETION':'COMPLETE'}
 (ROOT/'RUN_STATUS.json').write_text(json.dumps(status,indent=2,sort_keys=True)+'\n');(ROOT/'INPUT_FREEZE.json').write_text(json.dumps({'scope_sha256':hashlib.sha256((PREP/'TARGET_STAR_LABELS.tsv').read_bytes()).hexdigest(),'lexicon_sha256':hashlib.sha256((PREP/'ASTRONOMICAL_NAMES_LEXICON.tsv').read_bytes()).hexdigest(),'registry_sha256':hashlib.sha256((ROOT/'TRANSFORMATION_REGISTRY.tsv').read_bytes()).hexdigest(),'null':'global character shuffle preserving 27 token lengths','replicates':N_REPS,'seed_base':SEED_BASE,'selection':'all 64 profiles; same support/scorer contract'},indent=2,sort_keys=True)+'\n')
if __name__=='__main__':main()
