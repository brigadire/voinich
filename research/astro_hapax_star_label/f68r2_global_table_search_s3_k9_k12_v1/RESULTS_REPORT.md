# f68r2 S3: global table limits k=9..12

This package evaluates the final controlled expansion of the frozen f68r2 occurrence-level model.

## Frozen semantics

- 27 f68r2 occurrences only; f68r1 and group crosswalk are excluded.
- 25,517 scorer-consistent paths from the frozen graph.
- `DROP_UNMAPPED`, `NONE` abbreviation, exact full-token scorer parity.
- One path per LABEL occurrence and one use per canonical identity.
- Injective global character mapping, `GLOBAL_CAPACITY_1`.
- At least two distinct EVA token types for every active rule.
- Individual paths require at most five mappings. The S3 change is only the global union limit, 9 through 12; the graph was not rebuilt.

## Exact curve

| Global mapping limit | Status | Maximum coverage | Best bound | Time (s) |
|---:|---|---:|---:|---:|
| 1 | OPTIMAL | 0/27 | 0 | 2.0285 |
| 2 | OPTIMAL | 0/27 | 0 | 16.0214 |
| 3 | OPTIMAL | 0/27 | 0 | 16.2533 |
| 4 | OPTIMAL | 3/27 | 3 | 16.4383 |
| 5 | OPTIMAL | 3/27 | 3 | 16.3352 |
| 6 | OPTIMAL | 5/27 | 5 | 16.3689 |
| 7 | OPTIMAL | 5/27 | 5 | 16.3729 |
| 8 | OPTIMAL | 5/27 | 5 | 16.8279 |
| 9 | OPTIMAL | 5/27 | 5 | 16.4491 |
| 10 | OPTIMAL | 5/27 | 5 | 16.3333 |
| 11 | OPTIMAL | 5/27 | 5 | 16.2933 |
| 12 | OPTIMAL | 5/27 | 5 | 16.5792 |

The k=9..12 runs all have zero optimality gap. Thresholds 6, 8, and 11 are therefore `INFEASIBLE_CERTIFIED` for the tested frozen path graph and semantics.

The best k=12 assignment has five occurrences and six active mappings. Every active rule passes the distinct-EVA support gate. The same five-coverage plateau already appears at k=6 and does not improve through k=12.

## Interpretation

This is an exact result for the frozen scorer-consistent path graph and the global mapping limits tested. It is not a correlation claim about the astronomical dictionary. The preregistered threshold for launching a model-selection-aware null is 6/27; that threshold was not reached, so no null pilot was run. The 5/27 result remains exploratory and uncalibrated against the dictionary-selection null.

The controlled expansion is closed at k=12. Any further work should change the transformation class under a separately specified E3 protocol rather than increase the global table size within this model.
