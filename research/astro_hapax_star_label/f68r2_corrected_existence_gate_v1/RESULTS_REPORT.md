# Corrected existence gate

The frozen scorer-consistent graph contains 25517 paths and 19848 rule signatures. No path has an empty rule signature. Therefore a single selected path cannot satisfy support >=2 distinct EVA token types.

For two selected paths, every active mapping must be used by both paths; consequently their complete rule signatures must be identical. Exhaustive enumeration of all signature groups found zero pairs with distinct LABEL occurrences, distinct canonical identities, distinct EVA tokens, injective mapping, and full scorer replay.

Thus coverage 1 and coverage 2 are infeasible under the corrected support semantics. Coverage >=3 is closed by the coverage-2 subset bound. The corrected maximum for this frozen graph is exactly 0. E3 and null are not authorized.
