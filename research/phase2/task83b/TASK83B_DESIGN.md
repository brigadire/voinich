# Task83b design: deterministic Fingerprint V2.1 reconstruction

## Scope and authority

Task83b reconstructs the Fingerprint V2 target from the Task83a-verified ZL3b
and IT2a raw inputs.  It does not use Task83 rankings, distances, endpoint or
trajectory results, and it performs no Fontana, shorthand, or extraction model
selection.  The old `3fb953...` IT2a value is retained only as
`HISTORICAL_UNRESOLVED_METADATA / NON_AUTHORITATIVE`.

The scientific definitions are the immutable rows of
`research/phase2/fingerprint/F2_METRIC_REGISTRY_FINAL.tsv`.  No metric,
classification, estimator, null, seed, repetition count, threshold,
normalization, missingness rule, or family weighting may be changed here.

## Reconstruction

Each clean run independently executes:

1. frozen raw ZL3b/IT2a -> `ivtff-x7-extract`;
2. extracted text -> `codex_prepare prepare` with UTF-8, preserved case, and
   preserved lines;
3. the canonical, IT, and held-out-control Fingerprint V2 configurations;
4. the corrected PF4 leaf-paired null and frozen HR3/HR5 protocol;
5. the frozen family-balanced distance and Pareto aggregation.

RUN_A uses `GOMAXPROCS=1`, RUN_B uses `GOMAXPROCS=2`, and RUN_C uses the
runtime default.  The production pipeline has no worker-count option and no
concurrent stochastic jobs.  Every normative artifact is compared by SHA-256.

## Determinism gate

The repository audit covers map traversal before random draws, map-dependent
floating-point reductions and ordering, concurrent result collection, sorting
ties, filesystem traversal, seeds, shared RNGs, serialization, and scheduling.
All unordered groups are converted to explicitly sorted key sequences before
PRNG consumption or numeric aggregation.  Regression tests perturb insertion
order and run the same payload in separate OS processes at multiple
`GOMAXPROCS` settings.

## Scientific and provenance gates

The old/new audit is computed for every F2 row and separately for Monte Carlo
outputs.  The 13 CORE metrics are reclassified with the frozen Task79c rules.
PF4, HR3/HR5, control ordering, and Pareto conclusions are regenerated.  The
Task83a verifier must accept the new transitive manifest in ordinary
authoritative mode, and its negative tests must continue to pass.

Task81, Task82, Task82a, Task82a.1, and Task82b are audited only through their
freeze contracts and manifests.  Target-blind portfolios are neither rerun nor
retuned.  Task83b succeeds only after complete byte identity, provenance, and
verifier gates have passed.
