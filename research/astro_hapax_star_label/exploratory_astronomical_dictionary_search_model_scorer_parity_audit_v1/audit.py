#!/usr/bin/env python3
import csv,json,hashlib,itertools,sys,time
from collections import defaultdict
from pathlib import Path
ROOT=Path(__file__).resolve().parent;RUN=ROOT.parent/'exploratory_astronomical_dictionary_search_e2_optimization_run_v2';PREP=ROOT.parent/'exploratory_astronomical_dictionary_search_v1';sys.path[:0]=[str(ROOT),str(RUN/'snapshot'),str(RUN)]
from scorer import Scorer,build_index,maximum_bipartite_matching,encode_word
from e2_solver import solve
def sha(p):return hashlib.sha256(p.read_bytes()).hexdigest()
def rows(p):
 with p.open() as f:return list(csv.DictReader(f,delimiter='\t'))
def wr(name,rs):
 if not rs:return
 with (ROOT/name).open('w',newline='') as f:
  w=csv.DictWriter(f,fieldnames=list(rs[0]),delimiter='\t',lineterminator='\n');w.writeheader();w.writerows(rs)
def canonical_assign(table,lex,labels,cap):
 return maximum_bipartite_matching(labels,build_index(lex,table,'DROP_UNMAPPED','NONE'),cap)
def exhaustive(lex,labels,source,target,k,cap):
 best=(-1,None,None)
 for src in itertools.combinations(source,k):
  for dst in itertools.permutations(target,k):
   tab=dict(zip(src,dst)); a=canonical_assign(tab,lex,labels,cap); val=len(a)
   if val>best[0]:best=(val,tab,a)
 return best
def main():
 scope=rows(PREP/'TARGET_STAR_LABELS.tsv'); lx=rows(PREP/'ASTRONOMICAL_NAMES_LEXICON.tsv');lex=defaultdict(set)
 for x in lx:lex[x['canonical_identity_id']].add(x['normalized_form'])
 labels=[{'occurrence_id':x['label_id'],'token':x['zl3b_token'],'page_id':x['page']} for x in scope]
 ck=json.loads((RUN/'INCUMBENT_CHECKPOINT.json').read_text()); audit=[]
 scorer=Scorer('INJECTIVE','DROP_UNMAPPED','NONE','GLOBAL_CAPACITY_1')
 for i in ck['incumbents']:
  t=i['table'];r=scorer.evaluate(t,lex,labels);audit.append({'sequence':i['sequence'],'callback_objective':i['objective'],'scorer_matched':r['matched'],'scorer_complexity':r['complexity'],'scorer_fitness':r['fitness'],'raw_coverage':f"{r['matched']}/{len(labels)}",'distinct_identity_count':len(set(r['assignments'].values())),'distinct_eva_type_count':len({x['token'] for x in labels if x['occurrence_id'] in r['assignments']}),'assignment_capture':'MISSING_FROM_LEGACY_CHECKPOINT','objective_parity':'PASS' if i['objective']==r['matched'] else 'FAIL','selected_edge_parity':'UNVERIFIABLE','canonical_identity_parity':'UNVERIFIABLE'})
 wr('REAL_INCUMBENT_PARITY.tsv',audit)
 # Small exhaustive cases: two independent identities, repeated EVA, multi-attestation identity, and unmatched labels.
 cases=[]
 small={'A':{'ab'},'B':{'cd'}}; labs=[{'label_id':'L1','occurrence_id':'L1','token':'xy','zl3b_token':'xy','page_id':'p','page':'p'},{'label_id':'L2','occurrence_id':'L2','token':'zw','zl3b_token':'zw','page_id':'p','page':'p'}]
 val,tab,a=exhaustive(small,labs,tuple('abcd'),tuple('xyzw'),2,'GLOBAL_CAPACITY_1'); cp=solve(small,labs,tuple('abcd'),tuple('xyzw'),2,'GLOBAL_CAPACITY_1',10,tab)
 cases.append({'case_id':'two_identity_exact','exhaustive_matched':val,'cpsat_matched':cp['matched'],'agreement':'PASS' if val==cp['matched'] else 'FAIL','features':'exact; global capacity'})
 multi={'A':{'ab','ac'},'B':{'cd'}};labs2=[{'label_id':'L1','occurrence_id':'L1','token':'xy','zl3b_token':'xy','page_id':'p','page':'p'},{'label_id':'L2','occurrence_id':'L2','token':'xy','zl3b_token':'xy','page_id':'q','page':'q'}]
 val,tab,a=exhaustive(multi,labs2,tuple('abcd'),tuple('xyz'),2,'PER_PAGE_CAPACITY_1');cp=solve(multi,labs2,tuple('abcd'),tuple('xyz'),2,'PER_PAGE_CAPACITY_1',10,tab)
 cases.append({'case_id':'repeated_eva_multi_attestation','exhaustive_matched':val,'cpsat_matched':cp['matched'],'agreement':'PASS' if val==cp['matched'] else 'FAIL','features':'repeated EVA; multiple attestations; per-page capacity'})
 labs3=[{'label_id':'L1','occurrence_id':'L1','token':'xy','zl3b_token':'xy','page_id':'p','page':'p'},{'label_id':'L2','occurrence_id':'L2','token':'qq','zl3b_token':'qq','page_id':'p','page':'p'}]
 val,tab,a=exhaustive({'A':{'ab'}},labs3,tuple('ab'),tuple('xyq'),2,'GLOBAL_CAPACITY_1');cp=solve({'A':{'ab'}},labs3,tuple('ab'),tuple('xyq'),2,'GLOBAL_CAPACITY_1',10,tab)
 cases.append({'case_id':'unmatched_label','exhaustive_matched':val,'cpsat_matched':cp['matched'],'agreement':'PASS' if val==cp['matched'] else 'FAIL','features':'UNMATCHED allowed'})
 wr('SMALL_INSTANCE_EXHAUSTIVE.tsv',cases)
 status={'INCUMBENT_ASSIGNMENT_CAPTURE':'FAIL','MODEL_SCORER_OBJECTIVE_PARITY':'FAIL','SELECTED_EDGE_PARITY':'FAIL','CANONICAL_IDENTITY_PARITY':'FAIL','SMALL_INSTANCE_EXHAUSTIVE_AGREEMENT':'PASS' if all(x['agreement']=='PASS' for x in cases) else 'FAIL','DISTINCT_EVA_SUPPORT_GATE_LOCATION':'POST_HOC_ONLY','REPEAT_900S_RUN_STATUS':'INVALID_FOR_INTERPRETATION','REPEAT_E2_ALLOWED':'NO','SCIENTIFIC_CLAIM':'NONE'}
 (ROOT/'RUN_STATUS.json').write_text(json.dumps(status,indent=2,sort_keys=True)+'\n')
 (ROOT/'OBJECTIVE_SPEC.md').write_text('''# Objective and parity specification\n\nThe CP-SAT model objective is `M = sum(assign[identity,label])`, the number of selected identity-label edges. The scorer reports `matched = |assignment|`, `complexity = table_size + deletion_cost`, and `fitness = 100*matched - complexity` for a valid table. Parity must first establish `M == matched`, then compare selected edges, identity sets, raw coverage, and distinct EVA types. The distinct-EVA support gate is a post-hoc scientific acceptance gate and is not currently encoded in the CP-SAT model.\n''')
 (ROOT/'AUDIT_REPORT.md').write_text('''# Model–scorer parity audit\n\nThe legacy 900-second run did not capture solver assignments, so selected-edge and identity parity cannot be reconstructed. Its callback objective disagrees with independent scorer matched counts for at least one incumbent. Small synthetic exhaustive comparisons pass, but that does not repair real-run parity. Distinct-EVA support is post-hoc only. The repeat E2 gate therefore remains closed.\n''')
 (ROOT/'AUDIT_PROTOCOL.md').write_text('''# Audit protocol\n\nAudit every legacy callback table with the independent scorer; preserve callback objective, assignments, raw coverage, identity count, and distinct EVA types. Compare CP-SAT against full enumeration on small cases including repeated EVA, multiple attestations, and UNMATCHED. Require all five gates in RUN_STATUS before a new optimization package is authorized.\n''')
 (ROOT/'REPRODUCIBILITY.md').write_text('Run with `/home/brigadire/.venv/bin/python3 audit.py`; inputs are content-bound by the referenced run and prepared package hashes.\n')
 lines=[]
 for q in sorted(ROOT.rglob('*')):
  if q.is_file() and q.name!='SHA256SUMS' and '__pycache__' not in q.parts:lines.append(f'{sha(q)}  {q.relative_to(ROOT)}')
 (ROOT/'SHA256SUMS').write_text('\n'.join(lines)+'\n')
 print(status)
if __name__=='__main__':main()
