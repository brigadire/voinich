"""Prepare inputs only; never imports or runs model selection."""
from pathlib import Path
import collections,csv,hashlib,json,re,sys
from html.parser import HTMLParser
from engine import GRID,words
O=Path(__file__).resolve().parent
R=O.parents[2]
B=R/'research/astro_hapax_star_label'

def sha(p):return hashlib.sha256(p.read_bytes()).hexdigest()
def rd(p):
    with p.open(encoding='utf-8',newline='') as f:return list(csv.DictReader(f,delimiter='\t'))
def wr(name,rows,fields=None):
    with (O/name).open('w',encoding='utf-8',newline='') as f:
        w=csv.DictWriter(f,fieldnames=fields or list(rows[0]),delimiter='\t',lineterminator='\n');w.writeheader();w.writerows(rows)

def validate_upstream():
    checks=[]
    for directory, filenames in [(B/'restricted_m2_star_labels_v1',['TARGET_SETS.tsv','MANIFEST.json']),
            (R/'research/astro_dictionary_expansion_m1',['ASTRO_TERM_CORPUS_EXPANDED.tsv','ASTRO_TERM_CORPUS_EXPANSION_SOURCES.md'])]:
        registered={line.split(None,1)[1].strip():line.split()[0] for line in (directory/'SHA256SUMS').read_text().splitlines()}
        for name in filenames:
            p=directory/name
            assert sha(p)==registered[name],f'Upstream checksum mismatch: {p}'
            checks.append(dict(path=str(p.relative_to(R)),sha256=sha(p)))
    for row in rd(B/'INPUT_MANIFEST.tsv'):
        if row['logical_role']=='ZL3B_OCCURRENCE_METADATA':
            p=R/row['path'];assert sha(p)==row['sha256'];checks.append(dict(path=row['path'],sha256=sha(p)))
    checks.append(dict(path=str((B/'TRANSCRIPTION_TOKEN_CANDIDATES.tsv').relative_to(R)),sha256=sha(B/'TRANSCRIPTION_TOKEN_CANDIDATES.tsv')))
    return checks

class Text(HTMLParser):
    def __init__(self):super().__init__();self.parts=[]
    def handle_data(self,s):self.parts.append(s)
def htmltext(name):
    p=Text();p.feed((O/'sources'/name).read_text(encoding='cp1252'));return ' '.join(p.parts)

NAMES='''Aaron Abdenago Abdias Abdo Abel Abgarum Abia Abimelech Abrahae Abraham Abram Absalon Achab Achaz Achazias Adam Agar Aggaeus Ahia Alphaei Alphaeus Amasias Amathi Ammi Ammon Amon Amos Ananias Andreas Anna Aoth Aquila Arfaxat Asa Asaph Aser Athalia Azarias Bala Balaam Balac Balthasar Barabba Barach Barachia Bariona Barione Barnaban Barnabas Bartholomeus Beniamin Benoni Booz Cain Cainan Caiphas Caleph Canaan Cephan Cephas Cethura Chain Cham Chus Cleophae Coheleth Dalila Dan Daniel Danielum Darii David Debbora Didymus Dina Eleazar Eliachim Elisabeth Eman Emmanuel Enoch Enos Ephraim Esaia Esaiam Esaias Esau Esdras Ethan Eva Ezechias Ezechiel Gabriel Gad Gedeon Habacuc Heber Heli Heliam Helias Helisaeus Herode Herodes Hester Hieronymus Iacob Iacobi Iacobus Iahel Iair Iannes Iapheth Ididia Idithun Ieconias Iehu Iephte Ieremias Ieroboam Iessai Iesu Iesum Iesus Iezabel Ioachaz Ioachim Ioas Ioatha Ioatham Iob Iohanna Iohanne Iohannem Iohannes Iohannis Iohel Iona Ionas Ionathan Ioram Iosaphat Ioseph Iosias Iosue Isaac Iscariotes Ismahel Israel Issachar Iuda Iudae Iudas Iudith Laban Lamech Lazarus Lebbaeus Levi Lia Loth Lucas Magdalena Malachias Mambres Manahem Manasse Manasses Marcus Maria Mariam Martha Matthaeus Matthias Matusalam Melchisedech Michael Micheas Misac Misahel Moab Moysen Moyses Moysi Nabuchodonosor Nahum Nathan Nathanael Nathanei Nehemias Nembroth Nephtalim Noe Noemi Ochozias Olda Omri Ophni Othoniel Ozias Paulus Paulum Petri Petro Petrum Petrus Phaleg Pharao Pharaonis Phares Philippus Phinees Pilatus Pontius Raab Rachel Racheli Raphael Rebecca Roboam Ruben Ruth Salomon Salomonem Samson Samuel Samuele Sara Sarai Saram Sarra Saul Saule Saulem Saulo Saulum Saulus Sedechias Sella Sem Semeia Seth Sidrac Simeon Simon Simonis Sophonias Stephanus Susannam Thaddaeus Thamar Thara Thomas Tobiam Tola Vriel Zabulon Zacchaeus Zacharia Zacharias Zambri Zambria Zara Zebedaei Zebedaeus Zelpha Zorobabel'''.split()
# Manually selected ordinary words/technical terms, all independently checked
# against the Latin edition excerpt. Do not admit unreviewed OCR output.
ALCHEMY='''argentum aqua sale corpora salis ignem igne aurum argenti vivum sal alkali solutum album arsenicum argento olla furno ferri plumbum sicca auro natura panis aceto sulphur corporibus auri folia vase corpus plumbi arsenico noctem sapientiae aquam sulphuris armoniaci oleo condimentum luto calida soluto drachmam albot vitreata salem opus opera arsenici rubeum pondus limatura plumbo corporum vivo opere operationis unciam dies color spiritus fermentum urina sulphure atramenti armoniaco vivi ferrum calx vitro vitrum salibus humida operibus acetum quantitate aeris partes coagulatio diem condimenti limaturam cupro aluminis ferro rebus cuprum rubeo ignis lapis molitione mortario vitreato follibus operis colore albedinem rubedinem atramento fundo tinctura siccitas medicina alba massa partibus cochleari ferreo atramentum pallidi vas horis sepum vitreatum minium alumine mundo die unguentum olfactu stanni humiditas mane alumen cerussa aquae testa interiora frigida aere pallidum mineris exaltatio exaltatum albedine calidum aes frigidum colorem lapidum rubedinis horas cannam liquefactio minerae vitri clara albo lapide superficies albi pulvis stagnum nitro tincturam lapides usurub condimento spiritibus solutionis radix calcis frater modis folium foliato massae pura assatum animalium tempore ollam modum panno spisso coagulatum croco lutum foveam naturae terrae forti dissolutio sapientum aurifabri tritum optimum liquefactum humido pondere stannum armoniacus prunis nigra foliis siccum stagni faeces pauco unciis stanno aceti terra soluti grani sales'''.split()

def main():
    if (O/'FREEZE.json').exists():raise SystemExit('Already frozen. Refusing to overwrite v1 inputs.')
    checks=validate_upstream()
    metadata=[json.loads(s) for s in (R/'experiments/fingerprint-v2-task79-v1/canonical-out/occurrence_metadata.jsonl').open()]
    freq=collections.Counter(x['token'].replace('\x1f','/') for x in metadata)
    by_id={str(x['absolute_token_position']):x for x in metadata}
    target=[]
    source=rd(B/'restricted_m2_star_labels_v1/TARGET_SETS.tsv')
    for x in source:
        if x['subset']!='ALL_STAR_LABELS' or x['set_id'] not in ('F68R1_ALL_STAR_LABELS','F68R2_ALL_STAR_LABELS'):continue
        m=by_id[x['occurrence_id']]
        assert m['token'].replace('\x1f','/')==x['canonical_token_key']
        assert m['folio']==x['panel'] and m['locus_type']=='L'
        assert freq[x['canonical_token_key']]==int(x['normalized_frequency'])
        target.append(dict(occurrence_id=x['occurrence_id'],page_id=x['panel'],line_ref=x['line_ref'],
            block_id=x['functional_unit_id'],index_in_line=m['index_in_line'],token=x['readable_eva'],
            canonical_token_key=x['canonical_token_key'],normalized_frequency=freq[x['canonical_token_key']],
            hapax=int(freq[x['canonical_token_key']]==1)))
    assert len(target)==len({x['occurrence_id'] for x in target})==57
    assert collections.Counter(x['page_id'] for x in target)=={'f68r1':30,'f68r2':27}
    assert sum(x['hapax'] for x in target)==26
    wr('TARGET_SCOPE.tsv',target)
    lex=[]
    for i,x in enumerate(rd(R/'research/astro_dictionary_expansion_m1/ASTRO_TERM_CORPUS_EXPANDED.tsv')):
        if x['object_class']!='STAR':continue
        lex.append(dict(term_id=f'D1_ATTESTATION_{i:03}',canonical_identity=x['term_id'],
            canonical_star_identity=x['concept'],original_form=x['attested_form'],
            normalized_form=' '.join(words(x['normalized_form'])),transliterated_form=x['normalized_form'],
            language=x['language_layer'],tradition='MEDIEVAL_ASTROLABE_OR_CATALOGUE',source=x['source'],
            date=x['date'],variant=x['source_location'],historically_attested='YES',
            confidence=x['confidence'],provenance='research/astro_dictionary_expansion_m1/ASTRO_TERM_CORPUS_EXPANDED.tsv'))
    wr('ASTRONOMICAL_LEXICON.tsv',lex)
    pools=[]
    def addpool(family,forms,source,date,language,text):
        available=set(re.findall(r'\b[a-z]+\b',text.lower()))
        for s in sorted(set(w.lower() for w in forms)):
            assert s in available,(family,s)
            pools.append(dict(dictionary_id=family,term_id=family+'_'+s,original_form=s,normalized_form=s,
                language=language,source=source,date=date,historically_attested='YES',
                source_offset=text.lower().find(s)))
    addpool('HISTORICAL_NAMES',NAMES,'sources/isidore_7.html','early 7th century','LATIN',htmltext('isidore_7.html'))
    text17=htmltext('isidore_17.html')
    # Body only; all fully lowercase alphabetic tokens, unique types; no modern navigation.
    text17=text17[text17.index('I.'):text17.rfind('Isidore of Seville')]
    latin=sorted(set(re.findall(r'\b[a-z]{2,}\b',text17)))
    addpool('MEDIEVAL_LATIN',latin,'sources/isidore_17.html','early 7th century','LATIN',text17)
    ruska='\n'.join((O/'sources/ruska_1935.txt').read_text().splitlines()[2400:3948])
    addpool('ARABO_LATIN',ALCHEMY,'sources/ruska_1935.txt lines 2401-3948; Ruska 1935 pp.54-83',
            'medieval Arabic-Latin tradition; early print witness edited 1935','ARABIC_LATIN_TRANSLATION',ruska)
    wr('CONTROL_LEXICON_POOL.tsv',pools)
    expansions=dict(C='cth',K='ckh',P='cph',F='cfh',N='iin',A='ain',H='ch',S='sh',E='ee',I='in')
    labelpools=[]
    for x in rd(B/'TRANSCRIPTION_TOKEN_CANDIDATES.tsv'):
        family={'C':'CIRCULAR_TEXT','P':'INTRO_PROSE','Pb':'INTRO_PROSE'}.get(x['locus_type'])
        if x['panel'] in ('f68r1','f68r2') and family:
            labelpools.append(dict(family=family,occurrence_id=x['occurrence_id'],page_id=x['panel'],line_ref=x['line_ref'],token=x['readable_eva']))
    for m in metadata:
        if m['section']=='A' or m['locus_type']!='L':continue
        token=''.join(expansions.get(s,s) for s in m['token'].split('\x1f'))
        if not re.fullmatch('[a-z]+',token):continue
        labelpools.append(dict(family='OTHER_SECTION_LABEL',occurrence_id=str(m['absolute_token_position']),page_id=m['folio'],line_ref=m['line_identifier'].split('\x00')[-1],token=token))
    wr('LABEL_CONTROL_POOL.tsv',labelpools)
    registry=[]
    for family in ['LENGTH_ENDPOINT','UNIGRAM','BIGRAM','HISTORICAL_NAMES','MEDIEVAL_LATIN','ARABO_LATIN','OTHER_SECTION_LABEL','CIRCULAR_TEXT','INTRO_PROSE','SHUFFLED_LABEL','SHUFFLED_IDENTITIES']:
        registry.append(dict(dictionary_id=family,null_family=family,replicates=1 if family.startswith('SHUFFLED') else 10000,
            test='EXACT_INVARIANCE' if family.startswith('SHUFFLED') else 'MODEL_SELECTION_MONTE_CARLO',
            status='FROZEN',source='CONTROL_LEXICON_POOL.tsv' if family in ('HISTORICAL_NAMES','MEDIEVAL_LATIN','ARABO_LATIN') else 'LABEL_CONTROL_POOL.tsv' if family in ('OTHER_SECTION_LABEL','CIRCULAR_TEXT','INTRO_PROSE') else 'engine.py',
            identity_policy='same astronomy capacity blocks; controls are not claimed star identities'))
    wr('CONTROL_LEXICON_REGISTRY.tsv',registry)
    wr('TRANSFORMATION_REGISTRY.tsv',GRID)
    config=dict(version='restricted_dictionary_bruteforce_v1',systems=GRID,replicates=10000,
                seed=570000,synthetic_seeds=list(range(20)),synthetic_noise=[0,0.1,0.25],
                synthetic_system='S008',resampling_replicates=200,families=[x['dictionary_id'] for x in registry],
                wall_budget_seconds=14400,stage_C='EXCLUDED',score='100 * matched - complexity',
                matching='maximum cardinality exact equality, canonical capacity=1, unmatched cost=1, match cost=0')
    (O/'SEARCH_GRID_MANIFEST.json').write_text(json.dumps(config,indent=2)+'\n')
    wr('UPSTREAM_INPUTS.tsv',checks)
    # Hash all files outside this package in the relevant frozen upstream tree.
    upstream={str(p.relative_to(R)):sha(p) for directory in [B,R/'research/astro_dictionary_expansion_m1',R/'research/astro_token_formation',R/'research/astro_token_formation_m1']
              for p in directory.rglob('*') if p.is_file() and O not in p.parents and '__pycache__' not in p.parts}
    (O/'UPSTREAM_SNAPSHOT.json').write_text(json.dumps(upstream,sort_keys=True,indent=2)+'\n')
    (O/'PREPARATION_VALIDATION.json').write_text(json.dumps(dict(target_valid=True,target_count=57,pages={'f68r1':30,'f68r2':27},hapax=26,
        star_attestations=len(lex),canonical_identities=len({x['canonical_identity'] for x in lex}),
        control_pool_sizes=dict(collections.Counter(x['dictionary_id'] for x in pools)),
        label_pool_sizes=dict(collections.Counter(x['family'] for x in labelpools)),upstream_checksums='PASS'),indent=2)+'\n')
    print((O/'PREPARATION_VALIDATION.json').read_text())

if __name__=='__main__':main()
