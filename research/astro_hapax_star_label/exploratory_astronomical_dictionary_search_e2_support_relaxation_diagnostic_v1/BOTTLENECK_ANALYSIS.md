# Bottleneck analysis

The failed real E2 v3 run did not obtain a callback or checkpoint. The decomposition profile reproduces the failure on the monolithic builder before useful search output. The dominant structure is support-constraint construction: the legacy formulation creates support literals for combinations of source rule, target, EVA token type, identity, form, and occurrence label.

Its effective upper-order count is `O(|S| |T| |V| |I| |F| |L|)` in the literal construction path. The recorded realistic profile has 26 source units, 16 target units, 57 labels, 94 identities, and support-link estimates exceeding 1.5 million for the larger label subsets. This explains why the 900-second v3 result characterizes model construction/presolve rather than CP-SAT search.

D1 removes this cross-product from the master. A table is represented by at most five mappings, and matching is built only for the current table. The realistic synthetic benchmark builds and begins search immediately, captures a support-valid incumbent, and resumes deterministically. Its search status remains heuristic.
