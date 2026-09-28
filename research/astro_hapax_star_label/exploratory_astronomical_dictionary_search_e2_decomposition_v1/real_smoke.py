#!/usr/bin/env python3
import csv,json,time
from collections import defaultdict
from pathlib import Path
from decomposed_search import master_search
ROOT=Path(__file__).resolve().parent; PREP=ROOT.parent/'exploratory_astronomical_dictionary_search_v1'
def rows(p): return list(csv.DictReader(p.open(),delimiter='\t'))
def main():
 scope=rows(PREP/'TARGET_STAR_LABELS.tsv'); raw=rows(PREP/'ASTRONOMICAL_NAMES_LEXICON.tsv'); lex=defaultdict(set)
 for x in raw: lex[x['canonical_identity_id']].add(x['normalized_form'])
 labels=[{'label_id':x['label_id'],'occurrence_id':x['label_id'],'token':x['zl3b_token'],'page_id':x['page'],'page':x['page']} for x in scope]
 source=tuple('abcdefghijklmnopqrstuvwxyz'); target=tuple(sorted(set(''.join(x['token'] for x in labels)))); start=time.monotonic()
 r=master_search(labels,dict(lex),source,target,5,'GLOBAL_CAPACITY_1',budget=5,beam_width=16,checkpoint=ROOT/'REAL_SMOKE_CHECKPOINT.json')
 best=r['best']; row={'executed':'YES','status':'COMPLETE','build_completed':'YES','search_started':'YES','wall_seconds':round(time.monotonic()-start,4),'first_feasible_incumbent':'YES' if best else 'NO','objective':best['objective'] if best else 'NA','parity':'PASS' if best and best['objective']==len(best['assignment']) else ('NOT_EVALUATED' if not best else 'FAIL'),'scientific_interpretation':'NOT_ALLOWED'}
 with (ROOT/'REAL_SMOKE_RESULTS.tsv').open('w',newline='') as f:
  w=csv.DictWriter(f,fieldnames=list(row),delimiter='\t');w.writeheader();w.writerow(row)
 (ROOT/'REAL_SMOKE_STATUS.json').write_text(json.dumps(row,indent=2,sort_keys=True)+'\n');print(json.dumps(row,indent=2))
if __name__=='__main__':main()
