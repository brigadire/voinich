#!/usr/bin/env python3
import json,sys,time,random
from collections import defaultdict
from pathlib import Path
from decomposed_search import deterministic_matching,support_audit
from scorer import encode_word
ROOT=Path(__file__).resolve().parent; PREP=ROOT.parent/'exploratory_astronomical_dictionary_search_v1'
def rows(p):
 import csv
 return list(csv.DictReader(p.open(),delimiter='\t'))
def edge_dependencies(table,assignment,labels,lex):
 by={x['label_id']:x for x in labels}; out=[]
 for ident,lid in assignment:
  label=by[lid]; deps=[]
  for s in table:
   if any(s in f.replace(' ','') and encode_word(f,table,'DROP_UNMAPPED','NONE')==label['token'] for f in lex.get(ident,())): deps.append(s)
  out.append({'identity':ident,'label_id':lid,'eva_token':label['token'],'depends_on_rules':deps})
 return out
def metrics(table,labels,lex):
 raw=deterministic_matching(table,labels,lex,'GLOBAL_CAPACITY_1',0); audit=support_audit(table,raw['assignment'],labels,lex,2); counts={str(i):sum(x['distinct_eva_type_count']==i for x in audit) for i in (0,1)}; counts['2_plus']=sum(x['distinct_eva_type_count']>=2 for x in audit); valid=raw['matched'] if all(x['status']=='SUPPORTED' for x in audit) else 0; reduced={s:t for s,t in table.items() if next(x for x in audit if x['source_unit']==s)['status']=='SUPPORTED'}; after=deterministic_matching(reduced,labels,lex,'GLOBAL_CAPACITY_1',0)['matched'] if reduced else 0
 return {'table':table,'raw_coverage':raw['matched'],'support_valid_coverage':valid,'support_counts':counts,'support_violations':sum(x['status']!='SUPPORTED' for x in audit),'coverage_after_removing_unsupported_rules':after,'dependencies':edge_dependencies(table,raw['assignment'],labels,lex),'assignment':raw['assignment'],'support_audit':audit,'assignment_hash':raw['assignment_hash']}
def main():
 seed=int(sys.argv[1]); budget=int(sys.argv[2]); scope=rows(PREP/'TARGET_STAR_LABELS.tsv'); raw=rows(PREP/'ASTRONOMICAL_NAMES_LEXICON.tsv'); lex=defaultdict(set); rng=random.Random(seed)
 for x in raw: lex[x['canonical_identity_id']].add(x['normalized_form'])
 labels=[{'label_id':x['label_id'],'occurrence_id':x['label_id'],'token':x['zl3b_token'],'page_id':x['page'],'page':x['page']} for x in scope]; source=tuple('abcdefghijklmnopqrstuvwxyz'); target=tuple(sorted(set(''.join(x['token'] for x in labels)))); records=[]
 for minimum in (0,1,2):
  start=time.monotonic(); best=-1; sequence=0
  while time.monotonic()-start < budget:
   chosen=rng.sample(list(source),5); chosen_targets=rng.sample(list(target),5); table=dict(sorted(zip(chosen,chosen_targets)))
   oracle=deterministic_matching(table,labels,dict(lex),'GLOBAL_CAPACITY_1',0); m=metrics(table,labels,dict(lex)); supported=all(x['distinct_eva_type_count']>=minimum for x in m['support_audit'])
   if supported and oracle['matched']>best:
    best=oracle['matched']; sequence+=1; records.append({'seed':seed,'support_threshold':minimum,'sequence':sequence,'elapsed_sec':round(time.monotonic()-start,4),'search_status':'HEURISTIC_RANDOM_TABLE_PROBE','oracle':oracle['mode'],'metrics':m})
 (ROOT/f'SEED_{seed:02d}_DIAGNOSTIC.json').write_text(json.dumps({'seed':seed,'budget_sec_per_threshold':budget,'config':{'k':5,'capacity':'GLOBAL_CAPACITY_1','profile':'BALANCED'},'records':records},indent=2,sort_keys=True)+'\n'); print(json.dumps({'seed':seed,'records':len(records)}),flush=True)
if __name__=='__main__':main()
