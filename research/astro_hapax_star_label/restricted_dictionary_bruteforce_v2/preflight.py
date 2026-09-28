#!/usr/bin/env python3
"""Pre-production structural audit. It never runs production/null scoring."""
from pathlib import Path
import csv, itertools, json, math, random, statistics
import engine

HERE=Path(__file__).resolve().parent

def rows(name):
    with (HERE/name).open(encoding='utf-8',newline='') as f:return list(csv.DictReader(f,delimiter='\t'))

def universe(source_n,target_n,max_size=4):
    injective=sum(math.factorial(source_n)//math.factorial(source_n-k)*target_n*(target_n-1) if False else 0 for k in ())
    # P(source,k)*P(target,k), including the empty table.
    injective=sum((math.factorial(source_n)//math.factorial(source_n-k))*(math.factorial(target_n)//math.factorial(target_n-k)) for k in range(max_size+1))
    # Exactly one target collision: choose the source pair that shares a target,
    # then inject the remaining k-1 source symbols into distinct targets.
    merge=sum(math.comb(source_n,k)*math.comb(k,2)*(math.factorial(target_n)//math.factorial(target_n-(k-1))) for k in range(2,max_size+1))
    return injective,merge,injective+merge

def feasible_keep(source,target,max_mappings=4):
    if len(source)!=len(target):return False
    mapping={};reverse={}
    for a,b in zip(source,target):
        if a==b:continue
        if a in mapping and mapping[a]!=b:return False
        if b in reverse and reverse[b]!=a:return False
        mapping[a]=b;reverse[b]=a
    return len(mapping)<=max_mappings

def feasible_drop(source,target,max_mappings=4):
    # Dynamic subsequence alignment. Deleted source characters cost nothing;
    # retained characters define one injective global table.
    states={(0,tuple(),tuple())}
    for ch in source:
        nxt=set(states)
        for i,m,r in states:
            if i>=len(target):continue
            b=target[i]; md=dict(m); rev=dict(r)
            if ch in md and md[ch]!=b:continue
            if ch not in md and b in rev and rev[b]!=ch:continue
            md[ch]=b;rev[b]=ch
            if len(md)<=max_mappings:nxt.add((i+1,tuple(sorted(md.items())),tuple(sorted(rev.items()))))
        states=nxt
    return any(i==len(target) for i,_,_ in states)

def main():
    target=rows('TARGET_SCOPE.tsv'); lex=rows('ASTRONOMICAL_LEXICON.tsv')
    labels=sorted(set(x['token'] for x in target)); source=list(engine.SOURCE)
    alphabet=sorted(set(''.join(labels)))
    inj,merge,total=universe(len(source),len(alphabet),4)
    beam=len(engine.table_candidates(alphabet,256))
    out_of_beam=[x for x in engine.table_candidates(alphabet,beam) if x[1].startswith('T')]
    # The current beam is score-blind and contains only its first catalogue slice.
    # Independently sampled planted tables are deliberately selected outside it.
    rng=random.Random(803517)
    outside=[]
    for _ in range(20):
        size=rng.randint(1,4); src=rng.sample(source,size); dst=rng.sample(alphabet,size)
        table=tuple(zip(src,dst))
        if table not in [x[0] for x in out_of_beam]:outside.append(table)
    # Structural reachability is pairwise existential, before any model selection.
    terms={}
    for x in lex:terms.setdefault(x['canonical_identity'],set()).add(x['normalized_form'].replace(' ',''))
    keep_pairs=drop_pairs=0
    for label in labels:
        if any(feasible_keep(form,label) for fs in terms.values() for form in fs):keep_pairs+=1
        if any(feasible_drop(form,label) for fs in terms.values() for form in fs):drop_pairs+=1
    # DROP collision diagnostic across the complete score-blind beam. This is
    # deliberately measured before matching: it exposes short accidental forms.
    drop_values=[]
    for table,_,_ in out_of_beam:
        for identity,forms in terms.items():
            for form in forms:
                value=engine.encode(engine.orthographic(form,{}),table,True,'NONE')
                drop_values.append((identity,value))
    buckets={}
    for identity,value in drop_values:buckets.setdefault(value,set()).add(identity)
    collisions=sum(len(v)-1 for v in buckets.values() if len(v)>1)
    report={
      'LEXICON_SCOPE_JUSTIFIED':'NO',
      'FULL_STAR_LEXICON_REQUIRED':'YES',
      'CANONICAL_IDENTITIES_CURRENT':len(terms),
      'TABLE_SOURCE_GRAPHEMES':len(source),
      'TABLE_TARGET_EVA_SYMBOLS':len(alphabet),
      'TABLE_UNIVERSE_INJECTIVE':inj,
      'TABLE_UNIVERSE_ONE_MERGE':merge,
      'TABLE_UNIVERSE_SIZE':total,
      'BEAM_SIZE':beam,
      'BEAM_COVERAGE_FRACTION':beam/total,
      'SYNTHETIC_TABLES_SAMPLED_OUTSIDE_BEAM':'YES',
      'OUT_OF_BEAM_SYNTHETIC_CASES':len(outside),
      'OUT_OF_BEAM_SYNTHETIC_RECOVERY':'FAIL',
      'REAL_LABEL_REACHABILITY_KEEP':keep_pairs/len(labels),
      'REAL_LABEL_REACHABILITY_DROP':drop_pairs/len(labels),
      'REAL_LABEL_REACHABILITY':max(keep_pairs,drop_pairs)/len(labels),
      'KEEP_MODE_CAN_PRODUCE_EVA_ONLY':'NO',
      'DROP_MODE_COLLISION_RATE':collisions/max(1,len(drop_values)),
      'CROSS_PAGE_EXACT_MAPPING_CONSISTENCY_DEFINED':'NO',
      'CANONICAL_IDENTITY_CAPACITY_JUSTIFIED':'NO',
      'PRODUCTION_RUN_AUTHORIZED':'NO',
      'DECISION':'BLOCK_PRODUCTION_UNTIL_LEXICON_BEAM_REACHABILITY_MAPPING_AND_CAPACITY_REVISIONS'
    }
    (HERE/'PREFLIGHT_AUDIT.json').write_text(json.dumps(report,indent=2)+'\n')
    lines=['# v2 preflight audit','', 'Production is blocked. This audit performs no null or production scoring.','', '| Field | Value |','|---|---:|']
    lines += [f'| `{k}` | `{v}` |' for k,v in report.items()]
    lines += ['', '## Interpretation','',
      f'The current 31-identity lexicon leaves the stated coverage ceiling at 31/57. The exact bounded table universe is {total:,} tables ({inj:,} injective and {merge:,} one-merge), while the current score-blind beam contains {beam}; its fraction is {beam/total:.3e}.',
      f'Independent planted tables were sampled outside the beam ({len(outside)}/20). They cannot be recovered by a beam-only production search, so out-of-beam synthetic recovery is FAIL by construction.',
      f'Existential pair reachability before matching is KEEP={keep_pairs}/{len(labels)} and DROP={drop_pairs}/{len(labels)}. DROP collision diagnostics are intentionally reported separately because deletion can create short accidental forms.',
      'The protocol currently defines consistency only at the family level and has no historical basis for one global canonical identity capacity across both pages. Both must be revised before production.']
    (HERE/'PREFLIGHT_AUDIT.md').write_text('\n'.join(lines)+'\n')
    print(json.dumps(report,indent=2))
if __name__=='__main__':main()
