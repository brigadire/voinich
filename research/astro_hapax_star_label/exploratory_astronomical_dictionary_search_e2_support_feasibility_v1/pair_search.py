#!/usr/bin/env python3
import csv,json,time
from collections import defaultdict
from pathlib import Path
import sys
sys.path.insert(0,str(Path(__file__).resolve().parent.parent/'exploratory_astronomical_dictionary_search_e2_structural_support_audit_v1'))
from audit import alignments
ROOT=Path(__file__).resolve().parent; PREP=ROOT.parent/'exploratory_astronomical_dictionary_search_v1'
def rows(p):
 return list(csv.DictReader(p.open(),delimiter='\t'))
def main():
 labels=rows(PREP/'TARGET_STAR_LABELS.tsv'); labels=[dict(x,token=x['zl3b_token']) for x in labels]; raw=rows(PREP/'ASTRONOMICAL_NAMES_LEXICON.tsv'); lex=defaultdict(set)
 for x in raw:lex[x['canonical_identity_id']].add(x['normalized_form'])
 cache={}; groups=defaultdict(list); start=time.time()
 for ident,forms in lex.items():
  for form in forms:
   for l in labels:
    key=(form,l['token']); maps=cache.get(key)
    if maps is None: maps=cache[key]=alignments(form,l['token'],5)
    for mp in maps:
     if len(mp)<=5: groups[mp].append({'identity':ident,'label_id':l['label_id'],'token':l['token'],'form':form,'rules':mp})
 (ROOT/'PAIR_BUILD_HEARTBEAT').write_text(json.dumps({'variant_count':sum(map(len,groups.values())),'groups':len(groups),'elapsed_sec':time.time()-start})+'\n')
 total=sum(map(len,groups.values()))
 for rules,items in groups.items():
  if len(items)<2: continue
  bytoken=defaultdict(list)
  for x in items:bytoken[x['token']].append(x)
  toks=list(bytoken)
  for a in range(len(toks)):
   for b in range(a+1,len(toks)):
    for x in bytoken[toks[a]]:
     for y in bytoken[toks[b]]:
      if x['identity']!=y['identity'] and x['label_id']!=y['label_id']:
       result={'status':'SUPPORT_FEASIBLE','edge_count':2,'rule_count':len(rules),'table':dict(rules),'assignment':[x,y],'elapsed_sec':round(time.time()-start,4),'variant_count':total}
       (ROOT/'PAIR_FEASIBILITY.json').write_text(json.dumps(result,indent=2,sort_keys=True)+'\n');print(json.dumps(result,indent=2));return
 result={'status':'NO_PAIR_FOUND','edge_count':0,'rule_count':0,'variant_count':total,'elapsed_sec':round(time.time()-start,4)}; (ROOT/'PAIR_FEASIBILITY.json').write_text(json.dumps(result,indent=2,sort_keys=True)+'\n');print(json.dumps(result,indent=2))
if __name__=='__main__':main()
