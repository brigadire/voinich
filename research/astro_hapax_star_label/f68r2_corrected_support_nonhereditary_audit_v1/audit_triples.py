#!/usr/bin/env python3
import csv,itertools,json,hashlib
from pathlib import Path
from collections import defaultdict
import sys
BASE=Path(__file__).resolve().parent.parent
sys.path.insert(0,str(BASE/'f68r2_e3_historical_operations_v1'));import run_e3
SRC=BASE/'f68r2_occurrence_search_v1/F68R2_PATH_GRAPH.tsv';ROOT=Path(__file__).resolve().parent
def read(p):
 with p.open(encoding='utf-8',newline='') as f:return list(csv.DictReader(f,delimiter='\t'))
def write(p,fields,rs):
 with p.open('w',encoding='utf-8',newline='') as f:
  w=csv.DictWriter(f,fieldnames=fields,delimiter='\t',lineterminator='\n');w.writeheader();w.writerows(rs)
def mapping(paths):
 m={};inv={}
 for p in paths:
  for q in filter(None,p['rules'].split(';')):
   a,b=q.split('->')
   if (a in m and m[a]!=b) or (b in inv and inv[b]!=a):return None
   m[a]=b;inv[b]=a
 for p in paths:
  # Required mappings must generate the exact target; any additional global
  # mapping is caught here by the full executable scorer.
  if run_e3.encode_word(p['encoded'],m,'DROP_UNMAPPED','NONE')!=p['token']:return None
 return m
def valid3(ps):
 if len({p['label_id'] for p in ps})<3 or len({p['identity'] for p in ps})<3 or len({p['token'] for p in ps})<3:return None
 m=mapping(ps)
 if m is None:return None
 counts=defaultdict(set)
 for p in ps:
  for r in filter(None,p['rules'].split(';')):counts[r].add(p['token'])
 if any(len(v)<2 for v in counts.values()):return None
 # A local path forbids mapping any source unit not in its local lhs.
 for p in ps:
  lhs={r.split('->')[0] for r in filter(None,p['rules'].split(';'))}
  if any(a in m for a in set(p['encoded'])-lhs):return None
 return m
def main():
 rows=read(SRC); groups=defaultdict(list)
 for p in rows:groups[frozenset(filter(None,p['rules'].split(';')))].append(p)
 sigs=list(groups); exact={s:i for i,s in enumerate(sigs)}
 # Inverted subset index lets us enumerate only signatures that can be the
 # third member of a support-valid triple.
 subset=defaultdict(set)
 for i,s in enumerate(sigs):
  for n in range(len(s)+1):
   for q in itertools.combinations(sorted(s),n):subset[frozenset(q)].add(i)
 found=[]; tested=0; sigtriples=0
 for ia,A in enumerate(sigs):
  for ib in range(ia,len(sigs)):
   B=sigs[ib]; delta=A^B; union=A|B
   if len(delta)>5:continue
   # C must contain the symmetric difference and cannot contain a new rule.
   cand=[]
   for n in range(len(delta),min(5,len(union))+1):
    for C in (frozenset(x) for x in itertools.combinations(sorted(union),n)):
     if delta<=C and C in exact:cand.append(C)
   for C in set(cand):
    sigtriples+=1
    for a,b,c in itertools.product(groups[A],groups[B],groups[C]):
     tested+=1;m=valid3((a,b,c))
     if m is not None:
      found.append({'a':a,'b':b,'c':c,'mapping':m,'signature_a':';'.join(sorted(A)),'signature_b':';'.join(sorted(B)),'signature_c':';'.join(sorted(C))});return found,rows,len(groups),sigtriples,tested
 return found,rows,len(groups),sigtriples,tested
if __name__=='__main__':
 found,rows,ng,st,tested=main()
ROOT.joinpath('TRIPLE_SEARCH.tsv').write_text('status\tvalue\nEXACT_SIGNATURE_GROUPS\t'+str(ng)+'\nSIGNATURE_TRIPLES_TESTED\t'+str(st)+'\nPATH_TRIPLES_TESTED\t'+str(tested)+'\nVALID_TRIPLES_FOUND\t'+str(len(found))+'\n')
if found:
 r=[]
 for i,p in enumerate([found[0]['a'],found[0]['b'],found[0]['c']],1):r.append(dict(p,slot=i))
 write(ROOT/'VALID_TRIPLE.tsv',list(r[0]),r)
