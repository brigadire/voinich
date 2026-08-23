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

## Task53-1 execution record (2026-08-21)

The frozen pre-change input was reused verbatim:

* corpus: `doyle__transposition__w008__natural__seed001.txt`, SHA-256
  `2020b5359bf2118df395d2765daa52b4fc7fe003c21b84fee7501081a191f782`;
* Stage 13 parameters: defaults (`top=50`, `controls=3`, `context-limit=12`);
* pre-run cardinality: 3,170 target pairs, one family, 3,170 family edges.

The pre-run output accounting is 1,139,358,148 scientific bytes in five
YAML/TSV/Markdown files, plus 6,349 SVG files (18,942,402 bytes). The
recorded pre-run resource line remains:

| mode | wall | user CPU | sys CPU | peak RSS | output bytes |
|---|---:|---:|---:|---:|---:|
| pre, default | 1702.24 s | 658.95 s | 257.20 s | 31,257,956 KB | 1,158,300,550 |
| post, default | **not completed** | **not completed** | **not completed** | **not completed** | 0 |
| post, `-no-svg` | **not completed** | **not completed** | **not completed** | **not completed** | 0 |

Both post runs used the current optimized binary and the frozen input
artifacts. They emitted the expected cardinality diagnostics but the process
group was killed before `WriteAll` completed; no post scientific artifact was
created. Repeating with `GOMEMLIMIT=20GiB GOGC=20` also ended before output.
This is an observed memory-pressure failure, not a scientific result, so no
post wall/RSS value or speedup is claimed. Consequently byte identity,
identical target pairs, controls, decompositions, and YAML/TSV/Markdown
identity are still **unverified**. The `-no-svg` flag itself is verified by
code inspection to affect only derived SVG writes.

This run exposed the remaining Stage 13 bottleneck: `Output` retains full
context distributions for every target/control decomposition and the YAML
marshaller creates another large in-memory representation. The next
semantic-preserving optimization should stream scientific serialization (or
write decompositions incrementally) while retaining deterministic ordering;
SVG generation can remain optional. No Stage 14, 17, 23, or 27 code was
changed before profiling.

## Stage 14/17/23/27 profile status and hotspot hypotheses

The existing T8 run-state provides measured wall/CPU/RSS, while the available
Stage 17 one-trial pprof audit provides CPU and allocation evidence. Fresh
full-production profiling was not completed after the Stage 13 OOM, and the
following recommendations are therefore profiling targets, not claimed
optimizations:

| stage | measured T8 wall / CPU / peak RSS | observed or likely hotspot | semantic-preserving next step |
|---|---:|---|---|
| 14 distance-context-analyze | 1098.06 s / 1420.29 s / 6,104,512 KB | family-edge × exact-distance/context matrix and retained pair rows | index reusable corpus distance/context counts and stream rows; preserve all distances and ordering |
| 17 structural-projection-analyze | 8157.66 s / 7542.02 s / 2,536,916 KB | one-trial pprof: `metricsFloat`, `normalize`, `GenericSmoothing`, `ProjectDistribution`; allocation leaders `countsFloat`, `normalize`, `ProjectDistribution` | reuse immutable normalized vectors and scratch buffers per isolated trial; keep trial seeds and ordered reduction unchanged |
| 23 token-relation-validate | 6457.29 s / 1504.18 s / 2,324,556 KB | six permutation batteries over frozen candidates/families; executor/serialization overhead is secondary to replicate work | profile one replicate and reuse immutable candidate/profile preparation; retain per-replicate seeds and canonical reduction |
| 27 transition-network-validate | 328.14 s / 322.55 s / 663,816 KB | directed-edge permutation/null loop and profile aggregation | pre-index eligible directed edges and reuse immutable adjacency data; retain primary/refinement seed continuation and counters |

The Stage 17 pprof evidence is recorded in
`STRUCTURAL_PROJECTION_DISTRIBUTED_AUDIT.md`; the run-state evidence is in
`experiments/doyle__transposition__w008__natural__seed001-v1/run-state.json`.
These measurements confirm that the long intervals are CPU-backed and are
not the shutdown gap in the original filesystem timestamps.

## Non-goals

No family-edge cap, target cap, control reduction, threshold change, floating
point tolerance, or new aggregate comparison metric was introduced. Stage 14,
17, 23 and 27 need separate benchmark/profile runs before any further code
change; their scaling variables are demonstrably different.
