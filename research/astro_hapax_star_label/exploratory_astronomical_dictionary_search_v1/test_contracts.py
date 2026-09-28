#!/usr/bin/env python3
from candidate_search import compatibility, top_k_systems, canonical_symmetry_key

def main():
    labels=['abca','defg','hijk','lmno']; names=['abca','defg','hijk','lmno','xxxx']
    edges=compatibility(labels,names)
    assert all(e.support>=2 for e in edges)
    systems=top_k_systems(edges,50,500)
    assert systems and systems[0][0] >= systems[-1][0]
    assert len({canonical_symmetry_key(s[1]) for s in systems}) == len(systems)
    # A local perturbation does not create a global exact edge.
    assert not any(e.match_class=='EXACT_GLOBAL' and e.label=='abca' and e.name=='abda' for e in compatibility(['abca'],['abda']))
    print('synthetic contract tests: PASS')

if __name__=='__main__': main()
