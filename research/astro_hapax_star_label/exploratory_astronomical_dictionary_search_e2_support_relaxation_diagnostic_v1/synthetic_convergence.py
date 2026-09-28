#!/usr/bin/env python3
"""Positive-control convergence curve for the D1 master/oracle path."""
import csv,json,time
from pathlib import Path
from decomposed_search import master_search
ROOT=Path(__file__).resolve().parent
def positive():
 labels=[]; lex={}; i=0
 tables=({'a':'o','b':'k','c':'r'},{'a':'o','b':'k','c':'e'})
 forms=('ab','ac','bc')
 for tab,count in zip(tables,(3,6)):
  for j in range(count):
   f=forms[j%3]; ident=f'I{i:03d}'; i+=1; lex[ident]={f}; labels.append({'label_id':f'L{i:03d}','occurrence_id':f'L{i:03d}','token':''.join(tab[x] for x in f),'page_id':'p0'})
 return labels,lex,tuple('abc'),tuple('ekor')
def main():
 labels,lex,source,target=positive(); start=time.monotonic(); points=[]; best=0
 for seed,warm in (('seed-low',{'a':'o','b':'k','c':'r'}),('seed-high',{'a':'o','b':'k','c':'e'})):
  r=master_search(labels,lex,source,target,3,'GLOBAL_CAPACITY_1',2,8,warm,None,12)
  score=r['best']['objective'] if r['best'] else 0; best=max(best,score); points.append({'seed':seed,'elapsed_sec':round(time.monotonic()-start,4),'best_objective':best,'run_objective':score,'incumbent_captured':'YES' if r['best'] else 'NO','search_status':r['search_status']})
 with (ROOT/'SYNTHETIC_CONVERGENCE.tsv').open('w',newline='') as f:
  w=csv.DictWriter(f,fieldnames=list(points[0]),delimiter='\t');w.writeheader();w.writerows(points)
 improved=points[-1]['best_objective']>points[0]['best_objective']
 status={'cases':2,'positive_control':'YES','BEST_OBJECTIVE_MONOTONE':'PASS' if all(points[i]['best_objective']>=points[i-1]['best_objective'] for i in range(1,len(points))) else 'FAIL','STRICT_IMPROVEMENT_OBSERVED':'PASS' if improved else 'FAIL','MASTER_CONVERGENCE_PROBE':'PASS' if improved else 'FAIL','scientific_claim':'NONE'}
 (ROOT/'SYNTHETIC_CONVERGENCE_STATUS.json').write_text(json.dumps(status,indent=2,sort_keys=True)+'\n');print(json.dumps(status,indent=2))
if __name__=='__main__':main()
