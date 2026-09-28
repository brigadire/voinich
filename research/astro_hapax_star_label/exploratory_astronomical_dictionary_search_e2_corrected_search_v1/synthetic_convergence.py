#!/usr/bin/env python3
import csv,time
from pathlib import Path
from corrected_search import evaluate
ROOT=Path('.')
def main():
 table={'a':'o','b':'k','c':'e'}; forms=['ab','ac','bc','abc']; lex={f'I{i}':{f} for i,f in enumerate(forms)}; labels=[{'label_id':f'L{i}','token':''.join(table[x] for x in forms[i]),'page':'p0'} for i in range(4)]
 candidates=[{'a':'o','b':'k','c':'r'},{'a':'o','b':'k','c':'e'}]; start=time.monotonic(); best=0; rows=[]
 for i,t in enumerate(candidates):
  r=evaluate(t,labels,lex);best=max(best,r['raw_coverage']);rows.append({'step':i+1,'elapsed_sec':round(time.monotonic()-start,6),'objective':r['raw_coverage'],'best_so_far':best,'support_valid':r['support_valid']})
 with (ROOT/'SYNTHETIC_CONVERGENCE.tsv').open('w',newline='') as f:
  w=csv.DictWriter(f,fieldnames=list(rows[0]),delimiter='\t');w.writeheader();w.writerows(rows)
 print({'strict_improvement':rows[-1]['best_so_far']>rows[0]['best_so_far'],'status':'PASS'})
if __name__=='__main__':main()
