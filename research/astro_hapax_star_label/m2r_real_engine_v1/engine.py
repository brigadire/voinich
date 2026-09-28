#!/usr/bin/env python3
"""Small deterministic M2R morphology engine; development is synthetic-only."""
from collections import Counter,defaultdict
def units(s): return tuple(s)
def role(i,n): return 'initial' if i==0 else 'final' if i==n-1 else 'medial'
def induce(terms,labels,min_support=3):
 pairs=[]
 for t,l in zip(terms,labels):
  for i,(a,b) in enumerate(zip(t,l)): pairs.append((role(i,len(t)),a,role(i,len(l)),b))
 c=Counter(pairs); rules=[{'source':a,'source_role':ra,'target':b,'target_role':rb,'support':v} for (ra,a,rb,b),v in sorted(c.items()) if v>=min_support and ra==rb]
 return rules
def transform(s,rules):
 d={(r['source_role'],r['source']):r['target'] for r in rules}; out=''; unexpl=0
 for i,a in enumerate(s):
  z=d.get((role(i,len(s)),a)); out+=z if z else '?'; unexpl+=z is None
 return out,unexpl
def score(terms,labels,rules):
 fit=0;un=0
 for t,l in zip(terms,labels):
  z,u=transform(t,rules);fit+=sum(a==b for a,b in zip(z,l));un+=u+abs(len(z)-len(l))
 complexity=len(rules)+sum(max(0,3-r['support']) for r in rules)
 return {'data_fit':fit,'unexplained':un,'complexity':complexity,'score':fit-2*un-complexity,'coverage':fit/sum(map(len,labels))}
def search(terms,labels):
 best=None
 for k in (2,3,4):
  r=induce(terms,labels,k);s=score(terms,labels,r); cand=(s,r)
  if best is None or s['score']>best[0]['score']:best=cand
 return {'rules':best[1],'metrics':best[0],'search_type':'BOUNDED_OPTIMIZATION','global_optimum_claimed':'NO'}
