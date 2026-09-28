# Graph build report

The corrected compatibility graph contains 52,067 scorer-consistent paths over the frozen 57 labels and 299 lexicon rows. The graph is suitable as input to a compact path-only search.

The earlier Achernar/Diphda witness is excluded. Under the claimed table `a->l, c->o, n->h, r->c`, the scorer returns `lolchlc` for `acarnar` and `clhlohl` for `rana secunda`, so the claimed `olchc` and `chol` paths were invalid alignment artifacts.

This package records graph construction, not a solution. No support-valid correspondence is claimed, no solution is certified absent, and the subsequent search remains timeout-inconclusive.
