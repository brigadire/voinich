# Pipeline scalability refactor report (Task53)

## Implemented

Stage 13 now performs two semantic-preserving hoists:

1. log-count normalization for every control candidate is computed once per
   full pair file, instead of being sorted for every target;
2. pair decompositions are memoized by the canonical unordered pair key.

The memoized value is the same `decompose` result. It changes neither the
target order, control selection, thresholds, family edges, floating-point
operations inside a decomposition, nor deterministic serialization order.

Derived SVG rendering can be excluded from the critical path with
`structural-pair-decompose -no-svg`; the scientific artifacts are still
written. This is presentation-only and reversible by running without the
flag.

## Oracle and validation status

The existing `internal/pairdecomposition` tests remain the local control
selection oracle. The optimized control index uses the identical cost formula
and tie-breaker. A full before/after T8 output comparison is required before
claiming byte identity; this checkout contains the recorded pre-change run but
does not claim a post-change full-production measurement here.

The available pre-change T8 run measured:

| stage | wall | user CPU | sys CPU | peak RSS |
|---|---:|---:|---:|---:|
| 13 | 1702.24s | 658.95s | 257.20s | 31,257,956 KB |
| 14 | 1098.06s | 1420.29s | 5.58s | 6,104,512 KB |
| 17 | 8157.66s | 7542.02s | 19.64s | 2,536,916 KB |
| 23 | 6457.29s | 1504.18s | 19.76s | 2,324,556 KB |
| 27 | 328.14s | 322.55s | 20.95s | 663,816 KB |

No scientific output is deleted or sampled. The full T8 benchmark should be
run after targeted tests using `-no-svg` for the performance measurement and
the default mode for presentation identity.

## Non-goals

No family-edge cap, target cap, control reduction, threshold change, floating
point tolerance, or new aggregate comparison metric was introduced. Stage 14,
17, 23 and 27 need separate benchmark/profile runs before any further code
change; their scaling variables are demonstrably different.
