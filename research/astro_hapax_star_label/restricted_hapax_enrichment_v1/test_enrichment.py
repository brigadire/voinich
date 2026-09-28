from pathlib import Path
import csv,json
P=Path(__file__).parent
r=list(csv.DictReader((P/'ENRICHMENT_RESULTS.tsv').open(),delimiter='\t'))
m=json.loads((P/'MANIFEST.json').read_text())
assert m['target_n']==57 and m['target_hapax']==26
assert len(r)==4 and all(int(x['replicates'])==100000 for x in r)
assert {x['control'] for x in r}=={'INTRO_PROSE','CIRCULAR_TEXT','OTHER_ASTRONOMICAL','LENGTH_MATCHED_OTHER_SECTION'}
print('hapax enrichment integrity: PASS')
