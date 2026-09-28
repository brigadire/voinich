#!/usr/bin/env python3
"""D1 decomposition: deterministic table master plus an independent matcher.

The master is deliberately time bounded and reports HEURISTIC_ONLY.  The
fixed-table matcher is exact on small instances by exhaustive enumeration and
uses deterministic Hopcroft-style augmenting paths for the realistic beam;
every realistic result is support-audited before it can be an incumbent.
"""
from __future__ import annotations
import hashlib, json, time
import random
from collections import defaultdict
from itertools import combinations
from pathlib import Path
import sys
sys.path.insert(0, str(Path(__file__).resolve().parent / 'snapshot'))
from scorer import encode_word

def canonical_table(table):
    return tuple(sorted(table.items()))

def assignment_hash(assignment):
    return hashlib.sha256(json.dumps(sorted(assignment), sort_keys=True, separators=(',', ':')).encode()).hexdigest()

def support_audit(table, assignment, labels, lexicon, minimum=2):
    by={x['label_id']:x for x in labels}; support={s:set() for s in table}; rows=[]
    for ident,lid in assignment:
        label=by[lid]
        for s in table:
            if any(s in form.replace(' ','') and encode_word(form, table, 'DROP_UNMAPPED', 'NONE')==label['token'] for form in lexicon.get(ident, ())):
                support[s].add(label['token'])
    for s,t in sorted(table.items()):
        types=sorted(support[s]); rows.append({'source_unit':s,'eva_unit':t,'distinct_eva_type_count':len(types),'support_types':','.join(types),'status':'SUPPORTED' if len(types)>=minimum else 'REJECTED_SUPPORT_THRESHOLD'})
    return rows

def _edges(table, labels, lexicon):
    out=defaultdict(list)
    for label in sorted(labels, key=lambda x:x['label_id']):
        for ident in sorted(lexicon):
            if any(encode_word(form, table, 'DROP_UNMAPPED', 'NONE')==label['token'] for form in lexicon[ident]):
                out[label['label_id']].append(ident)
    return out

def _capacity_ok(assignment, labels, policy):
    if len({x[0] for x in assignment}) != len(assignment) or len({x[1] for x in assignment}) != len(assignment): return False
    if policy == 'GLOBAL_CAPACITY_1': return True
    counts=defaultdict(int); by={x['label_id']:x for x in labels}
    for ident,lid in assignment: counts[(ident, by[lid].get('page_id', by[lid].get('page','default')))] += 1
    return max(counts.values(), default=0) <= 1

def _valid_assignment(assignment, table, labels, lexicon, policy, minimum=2):
    if not _capacity_ok(assignment, labels, policy): return False
    by={x['label_id']:x for x in labels}
    for ident,lid in assignment:
        if not any(encode_word(f,table,'DROP_UNMAPPED','NONE')==by[lid]['token'] for f in lexicon.get(ident, ())): return False
    return all(x['status']=='SUPPORTED' for x in support_audit(table, assignment, labels, lexicon, minimum))

def exhaustive_matching(table, labels, lexicon, capacity='GLOBAL_CAPACITY_1', minimum=2):
    """Exact fixed-table oracle for small instances, including support."""
    adj=_edges(table,labels,lexicon); lids=sorted(adj, key=lambda x:(len(adj[x]),x)); best=[]
    def visit(i, used_i, used_l, cur):
        nonlocal best
        if len(cur)+(len(lids)-i)<=len(best): return
        if i==len(lids):
            if len(cur)>len(best) and _valid_assignment(cur,table,labels,lexicon,capacity,minimum): best=list(cur)
            return
        lid=lids[i]
        visit(i+1,used_i,used_l,cur)
        for ident in adj[lid]:
            if ident in used_i or lid in used_l: continue
            cur.append((ident,lid)); visit(i+1,used_i|{ident},used_l|{lid},cur); cur.pop()
    visit(0,set(),set(),[])
    return result(table,best,labels,lexicon,'EXHAUSTIVE_EXACT',capacity,minimum)

def deterministic_matching(table, labels, lexicon, capacity='GLOBAL_CAPACITY_1', minimum=2):
    """Deterministic maximum-cardinality matching, followed by support audit."""
    adj=_edges(table,labels,lexicon); match_i={}; match_l={}
    def dfs(lid, seen):
        for ident in adj[lid]:
            if ident in seen: continue
            seen.add(ident)
            if ident not in match_i or dfs(match_i[ident],seen):
                match_i[ident]=lid; match_l[lid]=ident; return True
        return False
    for lid in sorted(adj, key=lambda x:(len(adj[x]),x)): dfs(lid,set())
    pairs=sorted((ident,lid) for ident,lid in match_i.items())
    if not _valid_assignment(pairs,table,labels,lexicon,capacity,minimum): return result(table,[],labels,lexicon,'DETERMINISTIC_REJECTED_SUPPORT',capacity,minimum)
    return result(table,pairs,labels,lexicon,'DETERMINISTIC_MAX_MATCHING',capacity,minimum)

def result(table, assignment, labels, lexicon, mode, capacity, minimum=2):
    by={x['label_id']:x for x in labels}; types=sorted({by[l]['token'] for _,l in assignment}); identities=sorted({i for i,_ in assignment})
    unmatched=sorted(set(by)-{l for _,l in assignment})
    audit=support_audit(table,assignment,labels,lexicon,minimum)
    return {'table':dict(table),'assignment':sorted(assignment),'matched':len(assignment),'raw_coverage':len(assignment),'distinct_eva_coverage':len(types),'distinct_identity_coverage':len(identities),'unmatched_labels':unmatched,'support_audit':audit,'support_violations':sum(x['status']!='SUPPORTED' for x in audit),'assignment_hash':assignment_hash(assignment),'mode':mode,'capacity':capacity,'support_minimum':minimum,'status':'SUPPORT_VALID' if assignment and all(x['status']=='SUPPORTED' for x in audit) else ('EMPTY' if not assignment else 'SUPPORT_INVALID')}

def _optimistic(table, labels, lexicon, k):
    """Admissible coverage proxy for beam ordering; never an optimum claim."""
    by={x['label_id']:x for x in labels}; count=0
    for label in labels:
        if any(any(encode_word(form,table,'DROP_UNMAPPED','NONE')==label['token'] for form in fs) for fs in lexicon.values()): count+=1
    return count

def master_search(labels, lexicon, source, target, k=5, capacity='GLOBAL_CAPACITY_1', budget=30, beam_width=48, warm_start=None, checkpoint=None, exhaustive_threshold=12, seed=0, support_minimum=2):
    start=time.monotonic(); source=tuple(sorted(source)); target=tuple(sorted(target)); labels=sorted(labels,key=lambda x:x['label_id']); incumbents=[]; seen=set(); rng=random.Random(seed); search_source=list(source); rng.shuffle(search_source)
    if checkpoint and Path(checkpoint).exists():
        old=json.loads(Path(checkpoint).read_text()); incumbents=old.get('incumbents',[]); seen=set(old.get('seen',[]))
    beam=[{}]
    if warm_start: beam=[dict(warm_start)]
    def evaluate(table):
        nonlocal incumbents
        if len(table)!=k: return
        oracle=exhaustive_matching(table,labels,lexicon,capacity,support_minimum) if len(labels)<=exhaustive_threshold else deterministic_matching(table,labels,lexicon,capacity,support_minimum)
        if oracle['status']!='SUPPORT_VALID': return
        incumbent={'sequence':len(incumbents)+1,'elapsed_sec':round(time.monotonic()-start,4),'table':table,'assignment':oracle['assignment'],'objective':oracle['matched'],'best_bound':None,'gap':None,'oracle':oracle['mode'],'status':'HEURISTIC_ONLY'}
        if not incumbents or incumbent['objective']>incumbents[-1]['objective']:
            incumbents.append(incumbent)
            if checkpoint:
                tmp=str(checkpoint)+'.tmp'; Path(tmp).write_text(json.dumps({'status':'RUNNING','incumbents':incumbents,'seen':sorted(seen)},sort_keys=True)); Path(tmp).replace(checkpoint)
    if len(beam[0])==k: evaluate(beam[0])
    for depth in range(len(warm_start or {}), k):
        candidates=[]
        for partial in beam:
            last=max((search_source.index(x) for x in partial), default=-1)
            for si in range(last+1,len(search_source)):
                s=search_source[si]
                for t in target:
                    if t in partial.values(): continue
                    nxt=dict(partial); nxt[s]=t; key=canonical_table(nxt)
                    if key in seen: continue
                    seen.add(key); candidates.append((_optimistic(nxt,labels,lexicon,k),key,nxt))
                    if time.monotonic()-start>=budget: break
                if time.monotonic()-start>=budget: break
            if time.monotonic()-start>=budget: break
        tie={x[1]:rng.random() for x in candidates}; candidates.sort(key=lambda x:(-x[0],tie[x[1]],x[1])); beam=[x[2] for x in candidates[:beam_width]]
        for table in beam:
            if time.monotonic()-start>=budget: break
            evaluate(table)
        if time.monotonic()-start>=budget: break
    return {'status':'TIME_BOUNDED','search_status':'HEURISTIC_ONLY','seed':seed,'runtime_sec':round(time.monotonic()-start,4),'incumbents':incumbents,'best':incumbents[-1] if incumbents else None,'seen_tables':len(seen)}
