# Fixed-table maxima

All 30 scorer-consistent coverage-2-derived rule tables were optimized exactly. In every table:

- `UNCONSTRAINED_PAGE` maximum coverage is 3;
- `F68R2_ONLY` maximum coverage is 3;
- `CROSS_PAGE_REQUIRED` is infeasible with maximum coverage 0.

The 30 unrestricted maxima have 2,488 optimal assignments in total, matching the 2,488 previously enumerated support-valid triples. Each prior triple is contained in an optimal assignment for its own table. The maximum remains three labels, and every optimal assignment is same-page `f68r2`.

Co-optimality is substantial: each table has one optimal LABEL set but many identity assignments. Across tables there are 2,488 path assignments and many interchangeable canonical identities. This is consistent with local combinatorial ambiguity rather than a uniquely identified decoding.

The result is exact only for the family of 30 tables derived from coverage-2 signatures. It does not exclude tables with distributed support that have no support-valid coverage-2 subset sharing a complete signature.
