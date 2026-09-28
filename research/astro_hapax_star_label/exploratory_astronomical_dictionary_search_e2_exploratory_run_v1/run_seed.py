#!/usr/bin/env python3
import csv,json,sys,time
from collections import defaultdict
from pathlib import Path
from decomposed_search import master_search,support_audit
ROOT=Path(__file__).resolve().parent; PREP=ROOT.parent/'exploratory_astronomical_dictionary_search_v1'
def rows(p): return list(csv.DictReader(p.open(),delimiter='\t'))
def main():
 seed=int(sys.argv[1]); budget=int(sys.argv[2]); out=ROOT/f'SEED_{seed:02d}.json'; log=ROOT/f'SEED_{seed:02d}.heartbeat'; log.write_text('STARTED\n')
 scope=rows(PREP/'TARGET_STAR_LABELS.tsv'); raw=rows(PREP/'ASTRONOMICAL_NAMES_LEXICON.tsv'); lex=defaultdict(set)
 for x in raw: lex[x['canonical_identity_id']].add(x['normalized_form'])
 labels=[{'label_id':x['label_id'],'occurrence_id':x['label_id'],'token':x['zl3b_token'],'page_id':x['page'],'page':x['page']} for x in scope]
 source=tuple('abcdefghijklmnopqrstuvwxyz'); target=tuple(sorted(set(''.join(x['token'] for x in labels))))
 result=master_search(labels,dict(lex),source,target,5,'GLOBAL_CAPACITY_1',budget=budget,beam_width=32,checkpoint=ROOT/f'SEED_{seed:02d}_CHECKPOINT.json',seed=seed)
 for inc in result['incumbents']:
  inc['seed']=seed
  inc['support_audit']=support_audit(inc['table'],inc['assignment'],labels,lex); inc['unmatched_labels']=sorted(set(x['label_id'] for x in labels)-{l for _,l in inc['assignment']}); inc['distinct_eva_types']=len({next(x['token'] for x in labels if x['label_id']==lid) for _,lid in inc['assignment']}); inc['distinct_identities']=len({i for i,_ in inc['assignment']})
 log.write_text('COMPLETED\n')
 out.write_text(json.dumps({'seed':seed,'budget_sec':budget,'config':{'k':5,'capacity':'GLOBAL_CAPACITY_1','profile':'BALANCED'},'result':result},indent=2,sort_keys=True)+'\n')
 print(json.dumps({'seed':seed,'incumbents':len(result['incumbents']),'best':result['best']['objective'] if result['best'] else None,'runtime':result['runtime_sec']}),flush=True)
if __name__=='__main__':main()
