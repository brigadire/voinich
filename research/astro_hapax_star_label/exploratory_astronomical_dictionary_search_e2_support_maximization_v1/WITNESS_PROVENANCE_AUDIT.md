# Witness provenance audit

The previously reported Achernar/Diphda witness was rechecked against the executable E2 scorer before maximization. It fails exact scorer parity.

For `acarnar` with `a->l`, `c->o`, `n->h`, `r->c`, the scorer maps every occurrence of a mapped source character and returns `lolchlc`, not `olchc`. For `rana secunda`, the same table returns `clhlohl`, not `chol`. The earlier path was therefore a subsequence alignment artifact: it dropped repeated mapped characters without representing those drops in the scorer semantics.

Consequences:

- the previous witness is rejected as a warm start;
- its provenance cannot support a real-data claim;
- exact scorer parity is now required for every compatibility path;
- the corrected graph contains 52,067 parity-valid paths and no retained Achernar/Diphda witness.

The source rows themselves remain historically documented in the frozen lexicon: Achernar `LEX_0002` is dated 1252–1270, and Diphda `LEX_0137` is dated 1175. This provenance fact does not rescue the invalid transformation.
