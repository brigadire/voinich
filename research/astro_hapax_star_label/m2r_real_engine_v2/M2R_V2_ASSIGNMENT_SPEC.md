# Assignment subproblem

For fixed rules, align every possible source-target pair. Weight is literal baseline
saving minus path cost and log2(n_terms+1)+log2(n_labels+1) assignment bits. The exact
rectangular Hungarian solver augments each source row with dummy columns of weight
zero. Positive real edges produce assignments; other rows/labels remain unmatched.
This is maximum-weight one-to-one partial matching, not matching of parallel arrays.

Canonical surfaces order rows and columns, and all ties use ascending column order.
Opaque input IDs do not enter costs or tie-breaking. Excluding each selected edge
and resolving exactly gives its global assignment margin. Zero margins produce
alternative optimal solutions (up to three distinct representatives); this is not
exhaustive enumeration of an exponentially large equivalence class. Predictive
acceptance requires no alternative optimum, avoiding arbitrary symmetric claims.

The test oracle exhaustively enumerates rectangular matchings for independently
sampled integer matrices with negative and positive weights. Additional tests cover
unmatched choices, permutations, neutral ID changes and explicit ambiguity.
