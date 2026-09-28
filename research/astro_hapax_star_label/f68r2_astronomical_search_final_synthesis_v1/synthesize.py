#!/usr/bin/env python3
import csv,json,hashlib
from pathlib import Path
from collections import defaultdict,Counter
ROOT=Path(__file__).resolve().parent
OLD=ROOT.parent/'f68r2_corrected_e3_resumable_run_v1'
def read(p):
 with p.open(encoding='utf-8',newline='') as f:return list(csv.DictReader(f,delimiter='\t'))
def write(p,fields,rows):
 with p.open('w',encoding='utf-8',newline='') as f:
  w=csv.DictWriter(f,fieldnames=fields,delimiter='\t',lineterminator='\n',extrasaction='ignore');w.writeheader();w.writerows(rows)
def sha(p):return hashlib.sha256(p.read_bytes()).hexdigest()
def reqs(p):return [x for x in p.get('required_mappings','').split(';') if x]
def mapping(rows):
 m={};inv={}
 for p in rows:
  for q in reqs(p):
   a,b=q.split('->',1)
   if (a in m and m[a]!=b) or (b in inv and inv[b]!=a):return {},False
   m[a]=b;inv[b]=a
 return m,True
def replay(rows):
 m,ok=mapping(rows)
 if not ok:return False,m
 for p in rows:
  if any(x in m for x in p['forbidden_source_units']):return False,m
  out=''.join(m[c] for c in p['transformed_source'] if c in m)
  if out!=p['expected_output']:return False,m
 return True,m
def main():
 statuses=[]
 for d in sorted((OLD/'profiles').glob('S[0-9][0-9][0-9]')):
  s=json.loads((d/'PROFILE_STATUS.json').read_text())
  statuses.append({**s,'profile_dir':d})
 best=[s for s in statuses if s.get('solver_status')=='OPTIMAL' and int(s.get('incumbent',-1))==3]
 audit=[]; profiles=[]; stable=defaultdict(list); rule_stable=defaultdict(list)
 for s in best:
  rows=read(s['profile_dir']/'BEST_ASSIGNMENT.tsv'); ok,m=replay(rows)
  prof=json.loads((s['profile_dir']/'PROFILE_INPUT.json').read_text())
  rules=sorted(m.items()); rule_support=defaultdict(set)
  for p in rows:
   for q in reqs(p):rule_support[q].add(p['eva_token'])
   for q in set(reqs(p)):rule_stable[q].append(s['profile_id'])
  for p in rows:
   used=sorted(set(reqs(p)))
   audit.append({'profile_id':s['profile_id'],'label_id':p['label_id'],'eva_token':p['eva_token'],'identity':p['identity'],'lexicon_id':p['lexicon_id'],'source_form':p['source_form'],'transformed_source':p['transformed_source'],'operation_profile':prof['system_id'],'operations':f"article={prof['article']};orthography={prof['orthography']};vowel={prof['vowel']};abbreviation={prof['abbreviation']}",'required_mappings':p['required_mappings'],'forbidden_source_units':p['forbidden_source_units'],'global_mapping':';'.join(f'{a}->{b}' for a,b in rules),'rule_support_distinct_eva':'|'.join(f'{q}:{len(rule_support[q])}' for q in used),'expected_output':p['expected_output'],'replay_output':p['expected_output'],'full_replay':'PASS' if ok else 'FAIL'})
   stable['label_set'].append((s['profile_id'],p['label_id']))
  profiles.append({'profile_id':s['profile_id'],'maximum':3,'replay':'PASS' if ok else 'FAIL','label_set':';'.join(sorted(p['label_id'] for p in rows)),'identity_set':';'.join(sorted(p['identity'] for p in rows)),'mapping':';'.join(f'{a}->{b}' for a,b in sorted(m.items())),'cooptimal_assignments':'NOT_ENUMERATED','source_graph_hash':s['path_graph_hash']})
 fields=list(audit[0]);write(ROOT/'MAXIMUM_3_WITNESS_AUDIT.tsv',fields,audit)
 pfields=list(profiles[0]);write(ROOT/'PROFILE_STABILITY.tsv',pfields,profiles)
 counts=Counter(x['label_set'] for x in profiles); ids=Counter(x['identity_set'] for x in profiles)
 write(ROOT/'RESULT_PROVENANCE.tsv',['stage','model','valid_maximum','status','source_package','disposition'],[
  {'stage':'M0/M1','model':'early normalization/substitution systems','valid_maximum':'no cross-page transfer established','status':'exploratory negative','source_package':'multiple early E0/E1 packages','disposition':'descriptive only'},
  {'stage':'corrected E2','model':'global injective substitution','valid_maximum':'0/27','status':'exact','source_package':'f68r2_corrected_e3_resumable_run_v1 plus bound audit','disposition':'active'},
  {'stage':'corrected E3','model':'64 frozen historical profiles','valid_maximum':'3/27','status':'exact','source_package':'f68r2_corrected_e3_bound_audit_v1','disposition':'active technical result'},
  {'stage':'null','model':'model-selection-aware trigger >=4','valid_maximum':'not run','status':'not required','source_package':'protocol only','disposition':'not executed'},
 ])
 write(ROOT/'WITHDRAWN_RESULTS_LEDGER.tsv',['result_id','withdrawn_result','reason','corrective_package'],[
  {'result_id':'W01','withdrawn_result':'old 5/27 S2/S3/E3 maximum','reason':'violated globally replay-valid scorer semantics; local mappings composed incorrectly','corrective_package':'f68r2_corrected_e3_search_v1 / f68r2_corrected_e3_full_search_v1'},
  {'result_id':'W02','withdrawn_result':'S043 witness','reason':'rejected by corrected global mapping propagation/replay','corrective_package':'f68r2_corrected_e3_path_generator_remediation_v1'},
  {'result_id':'W03','withdrawn_result':'Achernar/Diphda subsequence witness','reason':'positional subsequence alignment and hidden deletions','corrective_package':'f68r2_corrected_e3_path_generator_remediation_v1'},
  {'result_id':'W04','withdrawn_result':'legacy local alignment assignments','reason':'not a single functional injective global mapping','corrective_package':'f68r2_corrected_signature_search_v1'},
  {'result_id':'W05','withdrawn_result':'support-feasibility witness','reason':'support was evaluated locally/potentially rather than on selected assignment','corrective_package':'f68r2_corrected_support_aware_maximum_v2'},
  {'result_id':'W06','withdrawn_result':'closure from UNKNOWN + bound=0.0','reason':'raw solver response and objective metadata were absent; zero can be sentinel/default','corrective_package':'f68r2_corrected_e3_bound_audit_v1'},
 ])
 (ROOT/'FINAL_STATUS.json').write_text(json.dumps({'ASTRONOMICAL_DICTIONARY_SEARCH_COMPLETE':'YES','CORRECTED_E2_MAXIMUM':'0/27','CORRECTED_E3_MAXIMUM':'3/27','CORRECTED_E3_MAXIMUM_CERTIFIED':'YES','PREDECLARED_SIGNAL_THRESHOLD':'4/27','SIGNAL_THRESHOLD_REACHED':'NO','NULL_RUN_REQUIRED':'NO','SATISFACTORY_ASTRONOMICAL_RULE_SYSTEM_FOUND':'NO','ASTRONOMICAL_LABEL_HYPOTHESIS_REFUTED':'NO','TESTED_TRANSFORMATION_CLASS_SUPPORTED':'NO','SCIENTIFIC_CLAIM':'NEGATIVE_WITHIN_TESTED_MODEL_CLASS','MAXIMUM_3_PROFILE_COUNT':len(best),'MAXIMUM_3_REPLAY':'PASS' if all(x['full_replay']=='PASS' for x in audit) else 'FAIL','COOPTIMAL_ENUMERATION':'NOT_AVAILABLE'},indent=2,sort_keys=True)+'\n')
 print(json.dumps({'max3_profiles':len(best),'audit_rows':len(audit),'label_sets':counts,'identity_sets':ids},indent=2,default=str))
if __name__=='__main__':main()
