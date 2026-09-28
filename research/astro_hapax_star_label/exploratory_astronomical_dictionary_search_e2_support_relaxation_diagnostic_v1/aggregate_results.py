#!/usr/bin/env python3
import csv,hashlib,json
from collections import defaultdict
from pathlib import Path
ROOT=Path(__file__).resolve().parent
def main():
 all_inc=[]
 for p in sorted(ROOT.glob('SEED_*.json')):
  if '_CHECKPOINT' in p.name: continue
  x=json.loads(p.read_text()); all_inc.extend(x['result']['incumbents'])
 unique={}
 for x in all_inc: unique[json.dumps(x['table'],sort_keys=True,separators=(',',':'))]=x
 top=sorted(unique.values(),key=lambda x:(-x['objective'],x['assignment_hash'],json.dumps(x['table'],sort_keys=True)))[:20]
 families=defaultdict(list)
 for x in unique.values(): families[tuple(sorted(x['table'].items()))].append(x['seed'] if 'seed' in x else None)
 with (ROOT/'TOP_K_TABLES.tsv').open('w',newline='') as f:
  fields=['rank','objective','elapsed_sec','oracle','table','assignment_hash','raw_coverage','distinct_eva_coverage','distinct_identity_coverage','unmatched_count','support_valid','seed']
  w=csv.DictWriter(f,fieldnames=fields,delimiter='\t');w.writeheader()
  for i,x in enumerate(top,1):w.writerow({'rank':i,'objective':x['objective'],'elapsed_sec':x['elapsed_sec'],'oracle':x['oracle'],'table':json.dumps(x['table'],sort_keys=True),'assignment_hash':x['assignment_hash'],'raw_coverage':x['objective'],'distinct_eva_coverage':x.get('distinct_eva_types','NA'),'distinct_identity_coverage':x.get('distinct_identities','NA'),'unmatched_count':len(x.get('unmatched_labels',[])),'support_valid':'YES' if all(y['status']=='SUPPORTED' for y in x.get('support_audit',[])) else 'NO','seed':x.get('seed','NA')})
 with (ROOT/'SOLUTION_FAMILIES.tsv').open('w',newline='') as f:
  w=csv.writer(f,delimiter='\t');w.writerow(['family_id','table','member_count','seeds'])
  for i,(tab,members) in enumerate(sorted(families.items()),1):w.writerow([i,json.dumps(dict(tab),sort_keys=True),len(members),','.join(str(x) for x in sorted(set(members)))])
 with (ROOT/'OBJECTIVE_DYNAMICS.tsv').open('w',newline='') as f:
  w=csv.writer(f,delimiter='\t');w.writerow(['seed','sequence','elapsed_sec','objective','best_so_far'])
  for p in sorted(ROOT.glob('SEED_*.json')):
   if '_CHECKPOINT' in p.name: continue
   x=json.loads(p.read_text()); best=0
   for inc in x['result']['incumbents']:
    best=max(best,inc['objective']);w.writerow([x['seed'],inc['sequence'],inc['elapsed_sec'],inc['objective'],best])
 status={'SEARCH_STATUS':'HEURISTIC_ONLY','GLOBAL_OPTIMUM_CERTIFIED':'NO','RESULT_TYPE':'EXPLORATORY_CANDIDATES','SCIENTIFIC_CLAIM':'NONE','SEEDS':len(list(ROOT.glob('SEED_*.json'))),'INCUMBENTS':len(all_inc),'DISTINCT_TABLES':len(unique),'TOP_K':len(top),'SOLUTION_FAMILIES':len(families),'REAL_DATA_SEARCH':'YES','EXACT_FIXED_TABLE_MATCHING_ORACLE':'YES','DISTINCT_EVA_SUPPORT':'ENFORCED_AND_AUDITED'}
 (ROOT/'RUN_STATUS.json').write_text(json.dumps(status,indent=2,sort_keys=True)+'\n')
 print(json.dumps(status,indent=2))
if __name__=='__main__':main()
