#!/usr/bin/env python3
import hashlib,json,random,time
from collections import defaultdict
from pathlib import Path
import sys
sys.path.insert(0,str(Path(__file__).resolve().parent/'snapshot'))
from scorer import encode_word
def assignment_hash(a): return hashlib.sha256(json.dumps(sorted(a),separators=(',',':')).encode()).hexdigest()
def matching(labels,lex,table):
 adj={}
 for l in labels: adj[l['label_id']]=sorted(i for i,fs in lex.items() if any(encode_word(f,table,'DROP_UNMAPPED','NONE')==l['token'] for f in fs))
 mi={}
 def dfs(lid,seen):
  for i in adj[lid]:
   if i in seen: continue
   seen.add(i)
   if i not in mi or dfs(mi[i],seen): mi[i]=lid; return True
  return False
 for lid in sorted(adj,key=lambda x:(len(adj[x]),x)): dfs(lid,set())
 return sorted((i,l) for i,l in mi.items())
def support_audit(table,assignment,labels,lex):
 by={x['label_id']:x for x in labels}; types={r:set() for r in table}; deps=[]
 for ident,lid in assignment:
  label=by[lid]; used=[]
  for form in lex[ident]:
   if encode_word(form,table,'DROP_UNMAPPED','NONE')!=label['token']: continue
   used=[s for s in table if s in form.replace(' ','')]
   break
  deps.append({'identity':ident,'label_id':lid,'token':label['token'],'rules':sorted(used)})
  for s in used: types[s].add(label['token'])
 audit=[]
 for s,t in sorted(table.items()):
  used=bool(types[s]); n=len(types[s]); audit.append({'source_unit':s,'eva_unit':t,'used_in_assignment':used,'distinct_eva_type_count':n,'support_types':','.join(sorted(types[s])),'status':'SUPPORTED' if not used or n>=2 else 'UNSUPPORTED_SINGLETON'})
 violations=sum(x['status']!='SUPPORTED' for x in audit)
 return audit,deps,violations
def evaluate(table,labels,lex):
 a=matching(labels,lex,table); audit,deps,v=support_audit(table,a,labels,lex); used=[x for x in audit if x['used_in_assignment']]
 valid=bool(a) and all(x['status']=='SUPPORTED' for x in used)
 toks={x['token'] for _,lid in a for x in labels if x['label_id']==lid}; ids={i for i,_ in a}
 return {'table':dict(sorted(table.items())),'assignment':a,'raw_coverage':len(a),'distinct_eva_coverage':len(toks),'distinct_identity_coverage':len(ids),'support_valid':valid,'support_violations':v,'support_audit':audit,'dependencies':deps,'assignment_hash':assignment_hash(a),'unmatched_labels':sorted({x['label_id'] for x in labels}-{lid for _,lid in a})}
def search(labels,lex,rule_pool,budget,seed):
 rng=random.Random(seed); start=time.monotonic(); best_raw=None; best_valid=None; inc=[]; seen=set()
 while time.monotonic()-start<budget:
  n=rng.randint(1,5); chosen=[]
  for _ in range(50):
   candidate=rng.choice(rule_pool)
   if candidate[0] in {x[0] for x in chosen} or candidate[1] in {x[1] for x in chosen}: continue
   chosen.append(candidate)
   if len(chosen)==n: break
  if len(chosen)!=n: continue
  table=dict(chosen); key=tuple(sorted(table.items()))
  if key in seen: continue
  seen.add(key); r=evaluate(table,labels,lex)
  improved=(best_raw is None or r['raw_coverage']>best_raw['raw_coverage'])
  valid_improved=r['support_valid'] and (best_valid is None or r['raw_coverage']>best_valid['raw_coverage'])
  if improved: best_raw=r; inc.append({'elapsed_sec':round(time.monotonic()-start,4),'criterion':'MAX_RAW','result':r})
  if valid_improved: best_valid=r; inc.append({'elapsed_sec':round(time.monotonic()-start,4),'criterion':'MAX_SUPPORT_VALID','result':r})
 return {'seed':seed,'budget_sec':budget,'runtime_sec':round(time.monotonic()-start,4),'tables_seen':len(seen),'best_raw':best_raw,'best_support_valid':best_valid,'incumbents':inc,'status':'TIME_BOUNDED'}
