# Validation report

The enumeration is complete for coverage 2 under the frozen graph. Paths were grouped by their complete active rule signature. For two selected paths, distinct-EVA support requires every active rule to occur in both paths; therefore both paths must share the same signature. Each unordered pair in every signature group was then checked for distinct label and canonical identity, and identical EVA token types were excluded.

Observed result:

- 52,067 graph paths examined;
- 30 support-valid signatures;
- 34 valid unordered pairs;
- all pairs satisfy the support threshold and `GLOBAL_CAPACITY_1`;
- all pairs are on `f68r2`;
- no timeout and no implementation-gate failure.

Coverage 3 was not part of this package and remains untested here.
