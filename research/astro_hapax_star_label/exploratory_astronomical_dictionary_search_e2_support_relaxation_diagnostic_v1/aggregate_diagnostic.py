#!/usr/bin/env python3
import csv,json,hashlib
from pathlib import Path
ROOT=Path('.')
def main():
 records=[]
 for p in sorted(ROOT.glob('SEED_*_DIAGNOSTIC.json')): records += json.loads(p.read_text())['records']
 fields=['seed','support_threshold','sequence','elapsed_sec','raw_coverage','support_valid_coverage','support_violations','coverage_after_removing_unsupported_rules','support_0_rules','support_1_rules','support_2_plus_rules','assignment_hash','table','dependencies']
 with (ROOT/'DIAGNOSTIC_TABLES.tsv').open('w',newline='') as f:
  w=csv.DictWriter(f,fieldnames=fields,delimiter='\t');w.writeheader()
  for r in records:
   m=r['metrics']; w.writerow({'seed':r['seed'],'support_threshold':r['support_threshold'],'sequence':r['sequence'],'elapsed_sec':r['elapsed_sec'],'raw_coverage':m['raw_coverage'],'support_valid_coverage':m['support_valid_coverage'],'support_violations':m['support_violations'],'coverage_after_removing_unsupported_rules':m['coverage_after_removing_unsupported_rules'],'support_0_rules':m['support_counts']['0'],'support_1_rules':m['support_counts']['1'],'support_2_plus_rules':m['support_counts']['2_plus'],'assignment_hash':m['assignment_hash'],'table':json.dumps(m['table'],sort_keys=True),'dependencies':json.dumps(m['dependencies'],sort_keys=True)})
 criteria=[]
 if records:
  criteria=[('MAX_RAW_COVERAGE',max(records,key=lambda r:r['metrics']['raw_coverage'])),('MAX_SUPPORT_VALID_COVERAGE',max(records,key=lambda r:r['metrics']['support_valid_coverage'])),('MIN_VIOLATIONS_AT_MAX_RAW',min((r for r in records if r['metrics']['raw_coverage']==max(x['metrics']['raw_coverage'] for x in records)),key=lambda r:r['metrics']['support_violations']))]
 with (ROOT/'BEST_BY_CRITERION.tsv').open('w',newline='') as f:
  w=csv.writer(f,delimiter='\t');w.writerow(['criterion','seed','threshold','raw_coverage','support_valid_coverage','violations','table','assignment_hash'])
  for name,r in criteria:w.writerow([name,r['seed'],r['support_threshold'],r['metrics']['raw_coverage'],r['metrics']['support_valid_coverage'],r['metrics']['support_violations'],json.dumps(r['metrics']['table'],sort_keys=True),r['metrics']['assignment_hash']])
 with (ROOT/'SUPPORT_THRESHOLD_SUMMARY.tsv').open('w',newline='') as f:
  w=csv.writer(f,delimiter='\t');w.writerow(['threshold','tables_recorded','max_raw_coverage','max_support_valid_coverage','min_violations'])
  for t in (0,1,2):
   rr=[r for r in records if r['support_threshold']==t]; w.writerow([t,len(rr),max((r['metrics']['raw_coverage'] for r in rr),default=0),max((r['metrics']['support_valid_coverage'] for r in rr),default=0),min((r['metrics']['support_violations'] for r in rr),default='NA')])
 status={'SEARCH_STATUS':'HEURISTIC_ONLY','GLOBAL_OPTIMUM_CERTIFIED':'NO','RESULT_TYPE':'SUPPORT_RELAXATION_DIAGNOSTIC','SCIENTIFIC_CLAIM':'NONE','SEEDS_COMPLETED':3,'BUDGET_SECONDS_PER_THRESHOLD':30,'THRESHOLDS':[0,1,2],'TABLES_RECORDED':len(records),'MAX_RAW_COVERAGE':max((r['metrics']['raw_coverage'] for r in records),default=0),'MAX_SUPPORT_VALID_COVERAGE':max((r['metrics']['support_valid_coverage'] for r in records),default=0),'REAL_DATA_SEARCH':'YES','INTERPRETATION':'DIAGNOSTIC_ONLY'}
 (ROOT/'RUN_STATUS.json').write_text(json.dumps(status,indent=2,sort_keys=True)+'\n');print(json.dumps(status,indent=2))
if __name__=='__main__':main()
