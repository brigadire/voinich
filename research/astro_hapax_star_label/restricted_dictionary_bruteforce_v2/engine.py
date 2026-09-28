"""Bounded global substitution search for v2; no per-pair fitting."""
import itertools, re, unicodedata
from collections import defaultdict

SOURCE = tuple('abcdefghiklmnopqrtuvwxyz') + ('kh','gh','sh','th','dh')
TABLE_MODES=('INJECTIVE','MERGE')
ABBREVIATIONS=('NONE','DROP_FINAL','PREFIX_4')

def normalize(s):
    s=unicodedata.normalize('NFKD',s).encode('ascii','ignore').decode().lower()
    return re.findall('[a-z]+',s)

def orthographic(s, rule):
    words=normalize(s)
    if rule.get('drop_al'): words=[w for w in words if w!='al'] or words
    value=''.join(words)
    if rule.get('ij_uv'): value=value.translate(str.maketrans({'j':'i','v':'u'}))
    if rule.get('arabic_latin'):
        for a,b in (('kh','h'),('gh','g'),('sh','s'),('th','t'),('dh','d')): value=value.replace(a,b)
        value=value.translate(str.maketrans({'j':'i','w':'u'}))
    if rule.get('contract') and len(value)>2: value=value[0]+re.sub('[aeiouy]','',value[1:-1])+value[-1]
    return value

def table_candidates(target_alphabet, beam=256):
    """Fixed, score-blind table catalogue; complexity is bounded before matching."""
    targets=tuple(sorted(target_alphabet))
    rows=[((), 'T0000', 0)]
    n=1
    for size in range(1,5):
        for src in itertools.combinations(SOURCE,size):
            for dst in itertools.product(targets, repeat=size):
                if len(set(dst))<size: continue
                complexity=size
                rows.append((tuple(zip(src,dst)),f'T{n:04d}',complexity)); n+=1
                if len(rows)>=beam*8: break
            if len(rows)>=beam*8: break
        if len(rows)>=beam*8: break
    return sorted(rows,key=lambda x:(x[2],x[1]))[:beam]

def encode(value, table, delete=False, abbreviation='NONE'):
    # Longest source graphemes first; table is one global function.
    mapping=dict(table); out=[]; i=0
    keys=sorted(mapping,key=len,reverse=True)
    while i<len(value):
        key=next((k for k in keys if value.startswith(k,i)),None)
        if key: out.append(mapping[key]); i+=len(key)
        elif delete: i+=1
        else: out.append(value[i]); i+=1
    value=''.join(out)
    if abbreviation=='DROP_FINAL' and len(value)>3:value=value[:-1]
    elif abbreviation=='PREFIX_4':value=value[:4]
    return value

def index(terms, rule, table, delete=False, abbreviation='NONE'):
    out=defaultdict(set)
    for term in terms:
        for form in term['forms']:
            value=encode(orthographic(form,rule),table,delete,abbreviation)
            if value: out[value].add(term['identity'])
    return {k:tuple(sorted(v)) for k,v in out.items()}

def match(labels, idx, unavailable=()):
    blocked=set(unavailable); owners={}
    adjacency={l['occurrence_id']:[x for x in idx.get(l['token'],()) if x not in blocked] for l in labels}
    def visit(label,seen):
        for term in adjacency[label]:
            if term in seen:continue
            seen.add(term)
            if term not in owners or visit(owners[term],seen):owners[term]=label;return True
        return False
    for label in sorted(adjacency):visit(label,set())
    return {label:term for term,label in owners.items()}

def evaluate(labels, idx, system_id, complexity):
    pairs=match(labels,idx)
    return dict(system_id=system_id,complexity=complexity,matched=len(pairs),n=len(labels),
                coverage=len(pairs)/len(labels) if labels else 0,score=100*len(pairs)-complexity,pairs=pairs)

def global_systems(target_alphabet, beam=256):
    tables=table_candidates(target_alphabet,beam)
    orth=[dict(drop_al=False,ij_uv=False,arabic_latin=False,contract=False),
          dict(drop_al=True,ij_uv=True,arabic_latin=True,contract=False)]
    out=[]
    for oi,rule in enumerate(orth):
        for table,tid,tc in tables:
            for delete in (False,True):
                for abbr in ABBREVIATIONS:
                    complexity=tc+int(delete)+{'NONE':0,'DROP_FINAL':1,'PREFIX_4':2}[abbr]
                    if complexity>8:continue
                    out.append((f'V2_{oi}_{tid}_{int(delete)}_{abbr}',rule,table,delete,abbr,complexity))
    return out

def select(labels, systems):
    best=None
    for sid,rule,table,delete,abbr,cost in systems:
        idx=index(CURRENT_TERMS,rule,table,delete,abbr)
        value=evaluate(labels,idx,sid,cost)
        if best is None or (value['score'],-cost,sid)>(best['score'],-best['complexity'],best['system_id']):
            best=dict(value,system=(sid,rule,table,delete,abbr,cost),index=idx)
    return best

CURRENT_TERMS=[]
