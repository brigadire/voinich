# Validation report

For each of the 30 previously observed rule tables, every graph path whose active rules are a subset of that table was collected. All unordered triples were enumerated and retained only when:

- labels are distinct;
- canonical identities are distinct;
- EVA token types are distinct;
- every active rule has support from at least two distinct EVA token types;
- `GLOBAL_CAPACITY_1` is satisfied.

The enumeration completed without timeout or implementation failure. It found 2,488 valid triples and zero cross-page triples. The result is therefore `COVERAGE_GE3_FEASIBLE=YES` and `CROSS_PAGE_INFEASIBLE_CERTIFIED=YES` for this tested table family.
