# Upstream Integrity Report (A1)

## Checksums

Baseline SHA-256 checksums of `restricted_dictionary_bruteforce_v1`,
`restricted_dictionary_bruteforce_v2`, and `restricted_dictionary_bruteforce_v3_design` were
captured **before** any audit activity into
`checkpoints/BASELINE_restricted_dictionary_bruteforce_{v1,v2,v3_design}.sha256` (951 + 20 + 34
files respectively — v1/v2 include cached benchmark artifacts from their own prior work; v3_design
is the full 34-file package under audit).

These baselines were re-verified via `sha256sum -c` at multiple points during the audit (after
setting up the reproduction sandbox, after running the reachability/small-instance re-execution,
and again at the end of the audit before writing this report): **all three packages verify OK, no
mismatches, at every checkpoint.**

The design package's own internal `SHA256SUMS` (covering its 20 top-level report/data files) was
also independently re-verified: **all 20 entries OK** — those files have not been edited since the
design package's own manifest was written. See `DATA_ACCESS_GRAPH.json` entry for `SHA256SUMS`
(finding F008 — MINOR): this internal manifest does not cover `scripts/`, `tests/`, or
`checkpoints/`, so it would not by itself have caught any post-hoc edit to the algorithm code.

## Target Scope / Seeds / Configs

- `restricted_dictionary_bruteforce_v2/TARGET_SCOPE.tsv` (the real EVA source file) is covered by
  the v2 baseline hash set and was not modified.
- No seed file exists as a separate artifact in the design package; seeds are inline integer
  literals in the benchmark-driver scripts, which are themselves covered by the v3-design baseline
  hash.
- No config files outside the package's own `.md`/`.tsv` documents were found to require separate
  hashing.

## Disposition

```text
UPSTREAM_INTEGRITY = PASS
```

No unauthorized modification of v1, v2, v3-design, the target scope, the historical lexicon, the
synthetic datasets, the source code, configurations, or seeds was detected at any point in this
audit. This finding is orthogonal to, and does not excuse, the CRITICAL content-level findings
elsewhere in this audit (F001-F003, F009) — the package's files are exactly what its own authors
last wrote; the problem is what several of those files claim, not whether they were tampered with
afterward.
