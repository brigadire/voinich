#!/usr/bin/env python3
import csv,json,hashlib,random
import sys
from collections import defaultdict
from itertools import combinations
from pathlib import Path
ROOT=Path(__file__).resolve().parent; PREP=ROOT.parent/'exploratory_astronomical_dictionary_search_v1'; DIAG=ROOT.parent/'exploratory_astronomical_dictionary_search_e2_support_relaxation_diagnostic_v1'
sys.path.insert(0,str(ROOT/'snapshot'))
from scorer import encode_word
def rows(p): return list(csv.DictReader(p.open(),delimiter='\t'))
def alignments(form,token,k):
 form=form.replace(' ',''); out=[]
 if len(token)>len(form): return out
 for pos in combinations(range(len(form)),len(token)):
  mapping={}; used={}; ok=True
  for i,j in zip(pos,range(len(token))):
   s=form[i]; t=token[j]
   if (s in mapping and mapping[s]!=t) or (t in used and used[t]!=s): ok=False; break
   mapping[s]=t; used[t]=s
  if ok and len(mapping)<=k: out.append(tuple(sorted(mapping.items())))
 return sorted(set(out))
def matching(labels,lex,table):
 adj={}
 for l in labels:
  adj[l['label_id']]=sorted(i for i,fs in lex.items() if any(encode_word(f,table,'DROP_UNMAPPED','NONE')==l['token'] for f in fs))
 mi={}
 def dfs(lid,seen):
  for ident in adj[lid]:
   if ident in seen: continue
   seen.add(ident)
   if ident not in mi or dfs(mi[ident],seen): mi[ident]=lid; return True
  return False
 for lid in sorted(adj,key=lambda x:(len(adj[x]),x)): dfs(lid,set())
 return sorted((i,l) for i,l in mi.items())
def max_matching(edges):
 adj=defaultdict(list)
 for i,l in edges: adj[l].append(i)
 mi={}
 def dfs(lid,seen):
  for ident in sorted(adj[lid]):
   if ident in seen: continue
   seen.add(ident)
   if ident not in mi or dfs(mi[ident],seen): mi[ident]=lid; return True
  return False
 for lid in sorted(adj,key=lambda x:(len(adj[x]),x)): dfs(lid,set())
 return len(mi)
def main():
 labels=rows(PREP/'TARGET_STAR_LABELS.tsv'); labels=[dict(x,token=x['zl3b_token'],page_id=x['page']) for x in labels]; raw=rows(PREP/'ASTRONOMICAL_NAMES_LEXICON.tsv'); lex=defaultdict(set)
 for x in raw: lex[x['canonical_identity_id']].add(x['normalized_form'])
 tokens=sorted({x['zl3b_token'] for x in labels}); target=sorted(set(''.join(tokens))); source=tuple('abcdefghijklmnopqrstuvwxyz')
 edge_maps={}; rule_edges=defaultdict(set); rule_types=defaultdict(set); rule_ids=defaultdict(set); rule_pages=defaultdict(set); co=defaultdict(set)
 for ident,forms in lex.items():
  for form in forms:
   for l in labels:
    maps=alignments(form,l['zl3b_token'],5)
    if maps: edge_maps[ident,l['label_id']] = (form,maps)
    for mp in maps:
     for rule in mp:
      rule_edges[rule].add((ident,l['label_id'],form)); rule_types[rule].add(l['zl3b_token']); rule_ids[rule].add(ident); rule_pages[rule].add(l['page'])
     for a,b in combinations(mp,2): co[tuple(sorted((a,b)))].add((ident,l['label_id']))
 support_rules={k for k,v in rule_types.items() if len(v)>=2}
 with (ROOT/'RULE_SUPPORT_INDEX.tsv').open('w',newline='') as f:
  fields=['source_unit','eva_unit','potential_support_labels','potential_support_eva_types','potential_support_identities','potential_support_pages','compatibility_edge_count','compatibility_edges'];w=csv.DictWriter(f,fieldnames=fields,delimiter='\t');w.writeheader()
  for s,t in sorted(rule_edges):
   es=sorted(rule_edges[s,t]);w.writerow({'source_unit':s,'eva_unit':t,'potential_support_labels':len({e[1] for e in es}),'potential_support_eva_types':len(rule_types[s,t]),'potential_support_identities':len(rule_ids[s,t]),'potential_support_pages':len(rule_pages[s,t]),'compatibility_edge_count':len(es),'compatibility_edges':json.dumps(es,sort_keys=True)})
 with (ROOT/'RULE_CO_OCCURRENCE.tsv').open('w',newline='') as f:
  w=csv.writer(f,delimiter='\t');w.writerow(['source_1','eva_1','source_2','eva_2','cooccurrence_edge_count','edges'])
  for (a,b),es in sorted(co.items()):w.writerow([a[0],a[1],b[0],b[1],len(es),json.dumps(sorted(es))])
 frontier=[]; random.seed(20260924)
 for k in range(1,6):
  raw_edges=[]; potential_edges=[]
  for (ident,lid),(form,maps) in edge_maps.items():
   if any(len(mp)<=k for mp in maps):
    raw_edges.append((ident,lid))
    if any(len(mp)<=k and all(r in support_rules for r in mp) for mp in maps): potential_edges.append((ident,lid))
  supportive=[r for r in support_rules if r[0] in source and r[1] in target]
  exists=bool(supportive)
  # A potential-support table of size <=k exists if any supported rule exists;
  # the table may be padded only when additional supported rules are available.
  best_real=0
  for _ in range(100):
   if not supportive: break
   n=random.randint(1,min(k,len(supportive))); chosen=random.sample(supportive,n); table=dict(chosen)
   if len(table)<n: continue
   ass=matching(labels,lex,table); audit=defaultdict(set)
   by={x['label_id']:x for x in labels}
   for ident,lid in ass:
    label=by[lid]
    for s,t in table.items():
     if any(s in form.replace(' ','') and encode_word(form,table,'DROP_UNMAPPED','NONE')==label['token'] for form in lex[ident]): audit[s].add(label['token'])
   if all(len(audit[s])>=2 for s in table): best_real=max(best_real,len(ass))
  frontier.append({'k':k,'potential_rules':len(supportive),'exists_table_leq_k':'YES' if exists else 'NO','max_raw_coverage_upper_bound':max_matching(raw_edges),'max_potential_support_coverage_upper_bound':max_matching(potential_edges),'max_realized_support_valid_found':best_real,'raw_edge_count':len(raw_edges),'potential_edge_count':len(potential_edges)})
 with (ROOT/'SUPPORT_FEASIBILITY_BY_K.tsv').open('w',newline='') as f:
  w=csv.DictWriter(f,fieldnames=list(frontier[0]),delimiter='\t');w.writeheader();w.writerows(frontier)
 with (ROOT/'COVERAGE_SUPPORT_FRONTIER.tsv').open('w',newline='') as f:
  w=csv.writer(f,delimiter='\t');w.writerow(['k','raw_upper_bound','potential_support_upper_bound','realized_support_valid_found']);[w.writerow([x['k'],x['max_raw_coverage_upper_bound'],x['max_potential_support_coverage_upper_bound'],x['max_realized_support_valid_found']]) for x in frontier]
 audit_three(ROOT,labels,lex,edge_maps,rule_types)
 (ROOT/'AUDIT_DATA.json').write_text(json.dumps({'labels':len(labels),'lexicon_rows':len(raw),'identities':len(lex),'rules':len(rule_edges),'supportive_rules':len(support_rules)},indent=2,sort_keys=True)+'\n')
def audit_three(root,labels,lex,edge_maps,rule_types):
 out=[]
 for p in sorted(DIAG.glob('SEED_*_DIAGNOSTIC.json')):
  x=json.loads(p.read_text())
  for r in x['records']:
   m=r['metrics']
   if m['raw_coverage']!=3: continue
   by={z['label_id']:z for z in labels}; table=m['table']; ass=m['assignment']; audit={q['source_unit']:q for q in m['support_audit']}; table_hash=hashlib.sha256(json.dumps(table,sort_keys=True,separators=(',',':')).encode()).hexdigest()
   for ident,lid in ass:
    l=by[lid]; form=next((f for f in lex[ident] if encode_word(f,table,'DROP_UNMAPPED','NONE')==l['token']), '')
    chain=[]
    for c in form.replace(' ',''): chain.append(f'{c}->{table[c]}' if c in table else f'{c}->DROP')
    used=sorted({c for c in table if c in form.replace(' ','')}); out.append({'seed':r['seed'],'table_hash':table_hash,'label_id':lid,'page':l['page'],'token':l['token'],'identity':ident,'dictionary_form':form,'transformation_chain':' '.join(chain),'used_rules':json.dumps(used),'rule_support':json.dumps({c:audit[c]['distinct_eva_type_count'] for c in used}),'repeated_hy_dependency':'YES' if l['token']=='hy' else 'NO','capacity':'GLOBAL_CAPACITY_1'})
 with (root/'THREE_MATCH_AUDIT.tsv').open('w',newline='') as f:
  fields=['seed','table_hash','label_id','page','token','identity','dictionary_form','transformation_chain','used_rules','rule_support','repeated_hy_dependency','capacity'];w=csv.DictWriter(f,fieldnames=fields,delimiter='\t');w.writeheader();w.writerows(out)
if __name__=='__main__':main()
