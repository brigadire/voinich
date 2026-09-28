#!/usr/bin/env python3
import csv,json,sys
csv.field_size_limit(10**8)
from collections import defaultdict
from pathlib import Path
from corrected_search import search
ROOT=Path(__file__).resolve().parent; PREP=ROOT.parent/'exploratory_astronomical_dictionary_search_v1'; STRUCT=ROOT.parent/'exploratory_astronomical_dictionary_search_e2_structural_support_audit_v1'
def rows(p):
 return list(csv.DictReader(p.open(),delimiter='\t'))
def main():
 seed=int(sys.argv[1]); budget=int(sys.argv[2]); scope=rows(PREP/'TARGET_STAR_LABELS.tsv'); raw=rows(PREP/'ASTRONOMICAL_NAMES_LEXICON.tsv'); lex=defaultdict(set)
 for x in raw: lex[x['canonical_identity_id']].add(x['normalized_form'])
 labels=[{'label_id':x['label_id'],'occurrence_id':x['label_id'],'token':x['zl3b_token'],'page':x['page']} for x in scope]
 pool=[(x['source_unit'],x['eva_unit']) for x in rows(STRUCT/'RULE_SUPPORT_INDEX.tsv') if int(x['potential_support_eva_types'])>=2]
 r=search(labels,dict(lex),pool,budget,seed); (ROOT/f'SEED_{seed:02d}.json').write_text(json.dumps(r,indent=2,sort_keys=True)+'\n'); print(json.dumps({'seed':seed,'seen':r['tables_seen'],'raw':r['best_raw']['raw_coverage'] if r['best_raw'] else 0,'support_valid':r['best_support_valid']['raw_coverage'] if r['best_support_valid'] else 0}),flush=True)
if __name__=='__main__':main()
