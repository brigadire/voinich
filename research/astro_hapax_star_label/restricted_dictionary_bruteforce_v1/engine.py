"""Finite, anonymous exact-matching engine. No trained parameters or EVA recoding.

Normalization conventions originate in research/astro_token_formation/main.py
(words/apply_rule); this independent implementation changes the grid and removes
its import-time/output dependencies. Matching uses augmenting paths.
"""
import collections
import itertools
import re
import unicodedata

ENDINGS = ('ibus','orum','arum','ium','ius','ae','is','us','um','ii','am','em','as','es','os','i','o','a','e')

def words(form):
    return re.findall('[a-z]+', unicodedata.normalize('NFKD', form).encode('ascii','ignore').decode().lower())

def grid():
    result = []
    for i, (article, orth, vowel, abbr) in enumerate(itertools.product(
            ('KEEP','DROP_AL'), ('IDENTITY','IJ_UV','ARABIC_LATIN','VELAR_COLLAPSE'),
            ('KEEP','CONTRACT_INTERNAL'), ('NONE','SUSPEND_1','PREFIX_4','STRIP_LATIN'))):
        complexity = (article != 'KEEP') + (orth != 'IDENTITY') + (vowel != 'KEEP') + {'NONE':0,'SUSPEND_1':1,'PREFIX_4':2,'STRIP_LATIN':1}[abbr]
        result.append(dict(system_id=f'S{i:03}', article=article, orthography=orth,
                           vowel=vowel, abbreviation=abbr, complexity=complexity,
                           level='A' if article=='KEEP' and vowel=='KEEP' and abbr=='NONE' else 'B'))
    return result

GRID = grid()

def transform(form, rule):
    ws = words(form)
    if rule['article']=='DROP_AL':
        ws = [w for w in ws if w != 'al'] or ws
    return transform_joined(''.join(ws), rule)

def transform_joined(s, rule):
    orth=rule['orthography']
    if orth=='IJ_UV': s=s.translate(str.maketrans({'j':'i','v':'u'}))
    if orth in ('ARABIC_LATIN','VELAR_COLLAPSE'):
        for a,b in (('kh','h'),('gh','g'),('sh','s'),('th','t'),('dh','d')): s=s.replace(a,b)
        s=s.translate(str.maketrans({'j':'i','w':'u'}))
    if orth=='VELAR_COLLAPSE': s=s.translate(str.maketrans({'q':'k','c':'k'}))
    if rule['vowel']=='CONTRACT_INTERNAL' and len(s)>2: s=s[0]+re.sub('[aeiouy]','',s[1:-1])+s[-1]
    a=rule['abbreviation']
    if a=='SUSPEND_1' and len(s)>3: s=s[:-1]
    if a=='PREFIX_4': s=s[:4]
    if a=='STRIP_LATIN':
        for ending in ENDINGS:
            if s.endswith(ending) and len(s)-len(ending)>=3:
                s=s[:-len(ending)]; break
    return s

def form_outputs(form):
    ws=words(form)
    joined={'KEEP':''.join(ws), 'DROP_AL':''.join([w for w in ws if w!='al'] or ws)}
    return [transform_joined(joined[r['article']],r) for r in GRID]

def indexes(terms, cache=None):
    result=[collections.defaultdict(set) for _ in GRID]
    for term in terms:
        for form in term['forms']:
            if cache is not None:
                if form not in cache: cache[form]=form_outputs(form)
                values=cache[form]
            else: values=form_outputs(form)
            for index,value in zip(result,values):
                if value: index[value].add(term['identity'])
    return [{s:tuple(sorted(ids)) for s,ids in x.items()} for x in result]

def matching(labels, index, unavailable=(), forbidden=None):
    banned=set(unavailable)
    adjacency={l['occurrence_id']:tuple(t for t in index.get(l['token'],())
                  if t not in banned and (l['occurrence_id'],t)!=forbidden) for l in labels}
    owner={}
    def visit(label,seen):
        for term in adjacency[label]:
            if term in seen: continue
            seen.add(term)
            if term not in owner or visit(owner[term],seen):
                owner[term]=label; return True
        return False
    for label in sorted(adjacency): visit(label,set())
    return {label:term for term,label in owner.items()}

def evaluate(labels, index, rule, unavailable=()):
    pairs=matching(labels,index,unavailable)
    # Integer selection statistic avoids floating point tie ambiguities.
    return dict(system_id=rule['system_id'], complexity=rule['complexity'],
                score=100*len(pairs)-rule['complexity'], matched=len(pairs),
                coverage=len(pairs)/len(labels) if labels else 0,
                n=len(labels), pairs=pairs)

def select(labels, indices):
    # Train only. In particular neither held-out tokens nor hapax enter tie-breaking.
    best=None
    for i,(idx,rule) in enumerate(zip(indices,GRID)):
        result=evaluate(labels,idx,rule)
        if best is None or result['score']>best['score']:
            best=dict(result,index=i)
    return best

def search(labels, indices):
    joint=select(labels,indices)
    out={'JOINT':joint}
    for train_page,held_page in [('f68r1','f68r2'),('f68r2','f68r1')]:
        train=[l for l in labels if l['page_id']==train_page]
        held=[l for l in labels if l['page_id']==held_page]
        selected=select(train,indices)
        result=evaluate(held,indices[selected['index']],GRID[selected['index']],selected['pairs'].values())
        out[train_page+'_TO_'+held_page]=dict(result,index=selected['index'],
            train_pairs=selected['pairs'],train_score=selected['score'],train_matched=selected['matched'])
    return out

def synthetic_dictionary(terms,rng,family):
    alphabet=''.join(''.join(words(f)) for t in terms for f in t['forms'])
    def alter(w):
        if len(w)<3:return w
        if family=='LENGTH_ENDPOINT': return w[0]+''.join(rng.choice(alphabet) for _ in w[1:-1])+w[-1]
        if family=='UNIGRAM':
            inner=list(w[1:-1]);rng.shuffle(inner);return w[0]+''.join(inner)+w[-1]
        # A randomized Euler traversal preserves directed bigram multiplicities,
        # length, unigram counts and endpoints EXACTLY; unchanged trails retained.
        edges=collections.defaultdict(list)
        for a,b in zip(w,w[1:]):edges[a].append(b)
        for e in edges.values():rng.shuffle(e)
        stack=[w[0]];trail=[]
        while stack:
            if edges[stack[-1]]:stack.append(edges[stack[-1]].pop())
            else:trail.append(stack.pop())
        return ''.join(reversed(trail))
    return [dict(identity=t['identity'],forms=[' '.join(alter(w) for w in words(f)) for f in t['forms']]) for t in terms]

def sampled_dictionary(terms,pool,rng):
    """Sample real attested forms into anonymous capacity blocks matching astronomy.

    These blocks are NOT asserted to be historical name identities. They hold
    variant opportunities/capacity fixed. No input is invented or truncated.
    """
    available=list(pool);out=[];distances=[]
    for t in terms:
        forms=[]
        for f in t['forms']:
            n=len(''.join(words(f)))
            delta=min(abs(len(s)-n) for s in available)
            candidates=[s for s in available if abs(len(s)-n)==delta]
            chosen=rng.choice(candidates);available.remove(chosen)
            forms.append(chosen);distances.append(delta)
        out.append(dict(identity=t['identity'],forms=forms))
    return out,distances
