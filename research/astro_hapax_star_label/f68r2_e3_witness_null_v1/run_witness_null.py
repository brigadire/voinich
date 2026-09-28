#!/usr/bin/env python3
import csv,json,random,time,hashlib,gc,math,sys
from pathlib import Path
from collections import defaultdict
ROOT=Path(__file__).resolve().parent;BASE=ROOT.parent;E3=BASE/'f68r2_e3_historical_operations_v1';DEC=BASE/'f68r2_e3_decomposed_search_v1';PREP=BASE/'exploratory_astronomical_dictionary_search_v1'
sys.path.insert(0,str(E3));import run_e3
sys.path.insert(0,str(DEC));import run_decomposed
N=100;SEED_BASE=90250925;BUDGET=5;THRESHOLD=5
def rows(p):
 with p.open(encoding='utf-8',newline='') as f:return list(csv.DictReader(f,delimiter='\t'))
def write(name,fields,data):
 with (ROOT/name).open('w',encoding='utf-8',newline='') as f:
  w=csv.DictWriter(f,fieldnames=fields,delimiter='\t',lineterminator='\n',extrasaction='ignore');w.writeheader();w.writerows(data)
def labels_for(base,rng):
 chars=list(''.join(x['zl3b_token'] for x in base));rng.shuffle(chars);out=[];i=0
 for x in base:
  y=dict(x);n=len(x['zl3b_token']);y['zl3b_token']=''.join(chars[i:i+n]);i+=n;out.append(y)
 return out
def paths_for(labels,lex,profile,deadline):
 out={};cache={}
 for lab in labels:
  for term in lex:
   if time.monotonic()>=deadline:return list(out.values()),'BUDGET'
   enc,_=run_e3.transform(term['normalized_form'],profile);key=(enc,lab['zl3b_token'])
   if len(key[1])>len(key[0]) or len(set(key[1]))>len(set(key[0])):continue
   if key not in cache:cache[key]=run_e3.alignments(*key)
   for sig in cache[key]:
    mp=dict(sig)
    if run_e3.encode_word(enc,mp,'DROP_UNMAPPED','NONE')!=lab['zl3b_token']:continue
    rules=';'.join(f'{a}->{b}' for a,b in sig);k=(lab['label_id'],term['canonical_identity_id'],rules,lab['zl3b_token'])
    out[k]={'system_id':profile['system_id'],'label_id':lab['label_id'],'eva_token':lab['zl3b_token'],'identity':term['canonical_identity_id'],'lexicon_id':term['lexicon_id'],'source_form':term['normalized_form'],'encoded':enc,'rules':rules,'trace':'NULL','scorer_parity':'PASS'}
 return list(out.values()),'COMPLETE'
def replay(w):
 if len(w)!=5 or len({x['label_id'] for x in w})!=5 or len({x['identity'] for x in w})!=5:return False,'cardinality'
 rules={};support=defaultdict(set)
 for x in w:
  for q in x['rules'].split(';'):
   a,b=q.split('->');
   if a in rules and rules[a]!=b:return False,'mapping_conflict'
   if b in rules.values() and a not in rules:return False,'noninjective'
   rules[a]=b;support[q].add(x['eva_token'])
  if ''.join(rules.get(c,'') for c in x['encoded'])!='': pass
 if any(len(v)<2 for v in support.values()):return False,'support'
 for x in w:
  if run_e3.encode_word(x['encoded'],rules,'DROP_UNMAPPED','NONE')!=x['eva_token']:return False,'global_scorer_output'
 return True,'PASS'
def ci_lower(k,n):
 try:
  from scipy.stats import beta
  return 0.0 if k==0 else beta.ppf(.05,k,n-k+1)
 except Exception:
  return 0.0 if k==0 else max(0,(k/n)-1.645*math.sqrt((k/n)*(1-k/n)/n))
def calibration():
 paths=rows(DEC/'PROFILE_PATHS/S043.tsv');r=run_decomposed.solve(paths,12,30,decision=5);return r['status'] in ('FEASIBLE','OPTIMAL') and r['coverage']>=5
def main():
 base=[x for x in rows(PREP/'TARGET_STAR_LABELS.tsv') if x['page']=='f68r2'];lex=rows(PREP/'ASTRONOMICAL_NAMES_LEXICON.tsv');reg=rows(ROOT/'TRANSFORMATION_REGISTRY.tsv');sc=rows(DEC/'E3_PROFILE_SCALING.tsv');upper={x['system_id']:int(x['relaxed_bipartite_upper_bound']) for x in sc};first=['S000','S003','S008','S011','S016','S019','S024','S027','S043'];order=first+[x['system_id'] for x in sorted(reg,key=lambda x:(-upper[x['system_id']],x['system_id'])) if x['system_id'] not in first]
 cal=calibration();reps=[];witnesses=[];started=time.monotonic()
 for rep in range(N):
  rng=random.Random(SEED_BASE+rep);ls=labels_for(base,rng);deadline=time.monotonic()+BUDGET;status='NO_WITNESS_WITHIN_BUDGET';found=None;processed=0
  try:
   for sid in order:
    if time.monotonic()>=deadline:break
    profile=next(x for x in reg if x['system_id']==sid);ps,st=paths_for(ls,lex,profile,deadline);processed+=1
    if ps:
     rr=run_decomposed.solve(ps,12,max(.05,deadline-time.monotonic()),decision=5)
     if rr['status'] in ('FEASIBLE','OPTIMAL') and rr['coverage']>=5:
      ok,reason=replay(rr['chosen'])
      if not ok: status='IMPLEMENTATION_FAILURE';break
      status='WITNESS_FOUND';found={'replicate':rep,'seed':SEED_BASE+rep,'profile':sid,'coverage':rr['coverage'],'assignments':rr['chosen']};break
    del ps;gc.collect()
  except Exception as e: status='IMPLEMENTATION_FAILURE';reps.append({'replicate':rep,'seed':SEED_BASE+rep,'status':status,'profiles_processed':processed,'error':repr(e)});continue
  reps.append({'replicate':rep,'seed':SEED_BASE+rep,'status':status,'profiles_processed':processed,'error':''})
  if found:witnesses.append(found)
  if rep%10==9:print(json.dumps({'replicates':rep+1,'witnesses':len(witnesses),'elapsed_sec':round(time.monotonic()-started,1)}),flush=True)
 write('NULL_REPLICATES.tsv',['replicate','seed','status','profiles_processed','error'],reps);write('NULL_WITNESSES.tsv',['replicate','seed','profile','coverage'],[{k:w[k] for k in ['replicate','seed','profile','coverage']} for w in witnesses]);assign=[]
 for w in witnesses:
  for x in w['assignments']:assign.append(dict(x,replicate=w['replicate'],seed=w['seed'],profile=w['profile']))
 write('WITNESS_ASSIGNMENTS.tsv',list(assign[0]) if assign else ['replicate','seed','profile'],assign)
 write('SCORER_REPLAY.tsv',['replicate','status','reason'],[{'replicate':x['replicate'],'status':'PASS' if x['status']=='WITNESS_FOUND' else 'NOT_APPLICABLE','reason':''} for x in reps])
 k=len(witnesses);p=k/N;status={'REAL_COVERAGE':'5/27','NULL_REPLICATES_RUN':N,'VALID_WITNESSES_FOUND':k,'OBSERVED_WITNESS_RATE':p,'CONSERVATIVE_LOWER_BOUND':ci_lower(k,N),'RESULT_COMPATIBLE_WITH_RANDOM':'YES' if ci_lower(k,N)>.1 else 'NOT_ESTABLISHED','NULL_RARITY_ESTABLISHED':'NO','SCIENTIFIC_CLAIM':'NONE','CALIBRATION_REAL_WITNESS':'PASS' if cal else 'FAIL','WITNESS_SEARCH_MODE':'LAZY_POSITIVE_ONLY','BUDGET_SECONDS_PER_REPLICATE':BUDGET}
 (ROOT/'RUN_STATUS.json').write_text(json.dumps(status,indent=2,sort_keys=True)+'\n');(ROOT/'CALIBRATION_REPORT.md').write_text(f'# Calibration\n\nKnown real S043 witness search: {"PASS" if cal else "FAIL"}. Null runs count only independently replayed witnesses.\n')
 (ROOT/'INPUT_FREEZE.json').write_text(json.dumps({'scope':'27 f68r2 occurrences','registry_sha256':hashlib.sha256((ROOT/'TRANSFORMATION_REGISTRY.tsv').read_bytes()).hexdigest(),'null':'global character shuffle preserving lengths','profiles':64,'seed_base':SEED_BASE,'replicates':N,'budget_seconds':BUDGET},indent=2,sort_keys=True)+'\n')
if __name__=='__main__':main()
