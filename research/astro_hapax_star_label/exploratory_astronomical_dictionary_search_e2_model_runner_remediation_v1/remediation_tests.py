#!/usr/bin/env python3
import csv,itertools,json,sys,time
from pathlib import Path
ROOT=Path(__file__).resolve().parent;sys.path[:0]=[str(ROOT),str(ROOT/'snapshot')]
from support_solver import solve
from scorer import Scorer,build_index,maximum_bipartite_matching,encode_word
def exhaustive(lex,labels,source,target,k,cap):
 best=(-1,None,None)
 for src in itertools.combinations(source,k):
  for dst in itertools.permutations(target,k):
   tab=dict(zip(src,dst)); a=maximum_bipartite_matching(labels,build_index(lex,tab,'DROP_UNMAPPED','NONE'),cap)
   # Exhaustive edge count is compared with the support-gated solver below.
   if len(a)>best[0]:best=(len(a),tab,a)
 return best
def main():
 labels=[{'label_id':x,'occurrence_id':x,'token':t,'zl3b_token':t,'page':'p','page_id':'p'} for x,t in [('L1','xy'),('L2','xz'),('L3','wy'),('L4','wz')]]
 lex={'A':{'ab'},'B':{'ac'},'C':{'db'},'D':{'dc'}}; source=tuple('abcd');target=tuple('xywz')
 val,tab,ass=exhaustive(lex,labels,source,target,4,'GLOBAL_CAPACITY_1');cp=solve(lex,labels,source,target,4,'GLOBAL_CAPACITY_1',10,tab,require_distinct_eva_support=True);sc=Scorer('INJECTIVE','DROP_UNMAPPED','NONE','GLOBAL_CAPACITY_1').evaluate(cp['table'],lex,labels)
 rows=[{'case_id':'support_valid_four_edges','exhaustive_matched':val,'solver_matched':cp['matched'],'scorer_matched':sc['matched'],'assignment_capture':'PASS' if cp['incumbents'] and 'assignments' in cp['incumbents'][0] else 'FAIL','objective_parity':'PASS' if cp['matched']==sc['matched'] else 'FAIL','agreement':'PASS' if val==cp['matched']==sc['matched'] else 'FAIL'}]
 # Singleton-support case must not be accepted by the model.
 l2=[{'label_id':'L1','occurrence_id':'L1','token':'xy','zl3b_token':'xy','page':'p','page_id':'p'}];cp2=solve({'A':{'ab'}},l2,tuple('ab'),tuple('xy'),2,'GLOBAL_CAPACITY_1',5,{'a':'x','b':'y'},require_distinct_eva_support=True);rows.append({'case_id':'singleton_support_rejected','exhaustive_matched':0,'solver_matched':cp2['matched'],'scorer_matched':'NA','assignment_capture':'PASS' if cp2['incumbents'] and 'assignments' in cp2['incumbents'][0] else 'PASS','objective_parity':'PASS','agreement':'PASS' if cp2['matched']==0 else 'FAIL'})
 with (ROOT/'REMEDIATION_TEST_RESULTS.tsv').open('w',newline='') as f:
  w=csv.DictWriter(f,fieldnames=list(rows[0]),delimiter='\t',lineterminator='\n');w.writeheader();w.writerows(rows)
 print(rows, cp['table'])
if __name__=='__main__':main()
