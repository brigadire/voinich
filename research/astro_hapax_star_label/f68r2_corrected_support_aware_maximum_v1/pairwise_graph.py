#!/usr/bin/env python3
"""Executable pairwise compatibility predicate for corrected paths.

Pairwise compatibility is sufficient for mapping/replay constraints because
all global conflicts are binary: a source unit can have one target, a target
can have one source, and a path can forbid a source unit.  Support is still a
hyperedge property and is intentionally not encoded by this graph.
"""
from collections import defaultdict
def rules(path): return tuple(filter(None,path['rules'].split(';')))
def forbidden(path):
    lhs={x.split('->')[0] for x in rules(path)}
    return set(path['encoded'])-lhs
def mapping(path): return {x.split('->')[0]:x.split('->')[1] for x in rules(path)}
def pair_compatible(a,b,scorer):
    if len({a['label_id'],b['label_id']})<2:return False
    if len({a['identity'],b['identity']})<2:return False
    if len({a['token'],b['token']})<2:return False
    ma,mb=mapping(a),mapping(b)
    for x,y in ma.items():
        if x in mb and mb[x]!=y:return False
        if y in mb.values() and mb.get(x)!=y:return False
    if forbidden(a)&set(mb):return False
    if forbidden(b)&set(ma):return False
    union=dict(ma);union.update(mb)
    return (scorer(a['encoded'],union)==a['token'] and
            scorer(b['encoded'],union)==b['token'])
def support_valid(paths):
    s=defaultdict(set)
    for p in paths:
        for r in rules(p):s[r].add(p['token'])
    return all(len(v)>=2 for v in s.values())
