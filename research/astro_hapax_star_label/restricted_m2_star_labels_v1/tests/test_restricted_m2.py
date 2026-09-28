from pathlib import Path
import csv,json,hashlib
P=Path(__file__).resolve().parents[1]
def rows(n): return list(csv.DictReader((P/n).open(encoding='utf-8'),delimiter='\t'))
scope=rows('ROLE_SCOPE_REGISTRY.tsv'); target=[x for x in scope if x['category']=='STAR_LABEL']
assert target and all(x['panel'] in ('f68r1','f68r2') for x in target)
assert not any(x['category']=='AMBIGUOUS_SCOPE' for x in scope)
assert not any(x['panel']=='f68r3' for x in scope)
assert all(x['category'] not in ('INTRO_PROSE','CIRCULAR_TEXT') for x in target)
for x in target: assert int(x['normalized_frequency'])>=1
sets=rows('TARGET_SETS.tsv'); keys={(x['set_id'],x['occurrence_id']) for x in sets}; assert len(keys)==len(sets)
m=json.loads((P/'MANIFEST.json').read_text()); assert m['restricted_m2_status']=='BLOCKED' and m['restricted_real_run_authorized']=='NO'
assert m['f68r3_included']=='NO' and m['singleton_rules_accepted']==0 and m['token_null_replicates']==0 and m['lexicon_null_replicates']==0
assert m['frozen_inputs_unchanged']=='YES'
print('restricted M2 integrity tests: PASS')
