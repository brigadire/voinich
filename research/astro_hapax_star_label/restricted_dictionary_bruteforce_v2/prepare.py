#!/usr/bin/env python3
"""Create v2 inputs from verified v1 artifacts; never alter v1."""
from pathlib import Path
import csv,hashlib,json,shutil

HERE=Path(__file__).resolve().parent; V1=HERE.parent/'restricted_dictionary_bruteforce_v1'; OUT=HERE.parent

def digest(p):return hashlib.sha256(p.read_bytes()).hexdigest()
def copy(name):shutil.copyfile(V1/name,HERE/name)
def main():
    if (HERE/'PREPARATION_MANIFEST.json').exists(): raise SystemExit('v2 already prepared; refusing overwrite')
    for name in ('TARGET_SCOPE.tsv','ASTRONOMICAL_LEXICON.tsv','LEXICON_PROVENANCE.md'):
        copy(name)
    # v2 additions are a review registry, initially empty: no post-hoc spellings.
    (HERE/'LEXICON_ADDITIONS.tsv').write_text('term_id\tcanonical_identity\tform\tlanguage\tdate\tsource\tprovenance\tstatus\n')
    (HERE/'SUBSTITUTION_ALPHABET.tsv').write_text('source_grapheme\tclass\tmax_support\n'+'\n'.join(f'{x}\tFIXED\t2' for x in ('a','b','c','d','e','f','g','h','i','k','l','m','n','o','p','q','r','t','u','v','w','x','y','z','kh','gh','sh','th','dh'))+'\n')
    (HERE/'NULL_CONTROL_REGISTRY.tsv').write_text('family\treplicates\tselection\tstatus\n'+'\n'.join(f'{x}\t10000\tFULL_GRID\tFROZEN' for x in ('SHUFFLED_LABEL','SHUFFLED_IDENTITIES','LENGTH_ENDPOINT','UNIGRAM','BIGRAM','HISTORICAL_NAMES','MEDIEVAL_LATIN','ARABO_LATIN','OTHER_SECTION_LABEL','CIRCULAR_TEXT','INTRO_PROSE'))+'\n')
    (HERE/'SEARCH_GRID_MANIFEST.json').write_text(json.dumps({'version':'v2','beam':256,'max_table_entries':4,'max_complexity':8,'max_support_single_mapping':2,'systems':'64 orthographic x bounded global substitution tables x global deletion/abbreviation','table_modes':['INJECTIVE','MERGE'],'abbreviations':['NONE','DROP_FINAL','PREFIX_4'],'selection':'100*matched-complexity','seeds':{'production':570000,'synthetic':800000,'resampling':900000}},indent=2)+'\n')
    inputs={name:digest(HERE/name) for name in ('TARGET_SCOPE.tsv','ASTRONOMICAL_LEXICON.tsv','LEXICON_PROVENANCE.md','LEXICON_ADDITIONS.tsv','SUBSTITUTION_ALPHABET.tsv','NULL_CONTROL_REGISTRY.tsv','SEARCH_GRID_MANIFEST.json','BRUTEFORCE_PROTOCOL.md','engine.py','prepare.py','run.py','test_engine.py') if (HERE/name).exists()}
    (HERE/'PREPARATION_MANIFEST.json').write_text(json.dumps({'package':'restricted_dictionary_bruteforce_v2','v1_source':str(V1.relative_to(OUT)),'target_labels':57,'f68r1':30,'f68r2':27,'target_hapax':26,'canonical_star_identities':31,'inputs':inputs,'status':'PREPARED_NOT_RUN'},indent=2)+'\n')
    print(json.dumps({'status':'PREPARED_NOT_RUN','target_labels':57,'canonical_star_identities':31},indent=2))
if __name__=='__main__':main()
