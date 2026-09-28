"""Bounded candidate-search contracts for the exploratory package.

The real-data entry point is intentionally absent. This module is exercised on
synthetic labels only and is suitable for future staged integration with the
frozen CP-SAT snapshot.
"""
from dataclasses import dataclass
from hashlib import sha256
from itertools import combinations

CLASSES = ('EXACT_GLOBAL', 'NEAR_LOCAL_1', 'WEAK_LOCAL_2', 'UNMATCHED')

@dataclass(frozen=True)
class Edge:
    label: str
    name: str
    match_class: str
    support: int

def compatibility(labels, names):
    """Create safe synthetic compatibility edges; singleton evidence is rejected."""
    out=[]
    for label in labels:
        for name in names:
            support=sum(a==b for a,b in zip(label, name))
            if support >= 2:
                cls='EXACT_GLOBAL' if label==name else ('NEAR_LOCAL_1' if support >= max(2,len(label)-1) else 'WEAK_LOCAL_2')
                out.append(Edge(label,name,cls,support))
    return out

def score(edges, profile='BALANCED'):
    weights={'EXACT_GLOBAL':1000,'NEAR_LOCAL_1':30,'WEAK_LOCAL_2':5,'UNMATCHED':-12}
    return sum(weights[e.match_class] for e in edges)

def top_k_systems(edges, k=50, cooptimal_limit=500):
    """Return diverse bounded systems under one-to-one name capacity."""
    labels=sorted({e.label for e in edges}); systems=[]
    for r in range(1, min(len(labels), 4)+1):
        for chosen in combinations(edges, r):
            if len({e.label for e in chosen}) != r or len({e.name for e in chosen}) != r: continue
            systems.append((score(chosen), tuple(sorted(chosen,key=lambda e:(e.label,e.name)))))
    systems.sort(key=lambda x:(-x[0],tuple((e.label,e.name,e.match_class) for e in x[1]))); return systems[:min(k,cooptimal_limit)]

def canonical_symmetry_key(edges):
    return sha256('|'.join(f'{e.label}>{e.name}' for e in sorted(edges,key=lambda x:(x.label,x.name))).encode()).hexdigest()
