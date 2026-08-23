# Corpus size audit

Task 37 audited the repository for assumptions that the corpus contains
39,026 tokens (or the earlier 38,887-token inventory). The audit covered Go
sources and tests, shell/orchestration files, Ansible templates and manifests,
documentation, task specifications, generated work directories, and the
frozen Voynich Baseline v1 snapshot.

## Result

No algorithmic corpus-size hardcode was found in production code. Corpus
lengths, window bounds, allocations, validation counts, and denominators are
already derived from the loaded token slice or an authoritative metadata
structure. No scientific parameter was changed.

An executable repository guard now rejects `39026`, `39_026`, `38887`, and
`38_887` in production Go/C, shell, JSON/YAML, Ansible, and orchestration
sources. Historical artifacts, documentation, task specifications, and test
fixtures are excluded intentionally. A separate regression exercises fixed
1,000-token windows with actual corpus sizes 1,000, 8,000, 39,026, and 60,017.

## Occurrence classification

### A. Legitimate historical/reference metadata (left unchanged)

- `experiments/voynich-v1/outputs/**` and `experiments/voynich-v1/logs/**`:
  token counts in the frozen output fields, reports, and logs describe the
  concrete Baseline v1 corpus. The exact-size fields occur in
  `begin_end_{candidates,report}`, `sequence_analysis*`,
  `distance_context_{pairs,report}`, `global_distributional_regimes`,
  `metadata_validation*`, `alignment_report`, `structural_{analysis,
  validation,reliability,profile_stability}`, `conditional-regime*`,
  `residual-diagnostic*`, `token-relation-validation*`,
  `higher-order-sequences*`, and `transition-network*`, plus the corresponding
  completion logs. They are results, not inputs to bounds or normalization.
- `experiments/voynich-v1/{FROZEN,REPORT.md,manifest.json,checksums.sha256}`:
  the immutable baseline identity and checksum oracle. The corpus SHA, rather
  than a size constant, identifies the canonical input.
- `tasks/task02.txt`, `task03-01.txt`, `task03.txt`, `task04.txt`, `task05.txt`,
  `task07.txt`, `task17.txt`, `task19.txt` through `task24.txt`, and
  `task37.txt`: historical requirements and recorded results mentioning 38,887
  or 39,026. Task text is not executable.
- `internal/tokenrelationvalidation/write.go`: the phrase “38,887-token
  corpus” documents the provenance of an old frozen candidate inventory. The
  same report explicitly states that validation statistics are recomputed on
  the current corpus.
- `internal/tokenrelationvalidation/tokenrelationvalidation_test.go`: the
  38,887 fixture verifies that candidate identity/settings from that old
  inventory remain readable while statistics are recomputed. The number does
  not size an algorithmic buffer or loop.
- Local untracked/generated `workdir/**` contains current 39,026-token results;
  local `workdir.v0/**` contains old 38,887-token results. Neither directory is
  production source or versioned baseline input.

### B. Algorithmic hardcode (fixed)

None found. In particular, searches found no production equality check,
allocation, loop bound, normalization denominator, window count, shell
condition, manifest field, or orchestration argument driven by either
historical corpus size.

The following existing authoritative patterns were verified:

- corpus loaders construct token slices from input and reject empty input;
- analyses report and normalize with `len(tokens)` or counts accumulated from
  the loaded corpus;
- metadata-consuming stages compare metadata length with the loaded corpus
  length before indexed access;
- sliding-window stages derive their count and end bounds from `len(tokens)`;
- frozen discovery counts are cross-input consistency checks, not constants.

### C. Suspicious but harmless (left unchanged)

- `README.md` contains a 38,887 `position_observations` YAML example. It is a
  displayed historical output value and is never parsed. Changing it would not
  improve runtime behavior and would rewrite reference documentation.
- `PERFORMANCE_AUDIT.md`, `PERFORMANCE_REFACTOR_REPORT.md`,
  and `DISTRIBUTED_EXECUTION_AUDIT.md` use approximately 39,026 as the measured
  workload scale. These values explain benchmark results only.
- `internal/metadatavalidation/hoist_test.go` uses 39,026 in equivalence tests
  and benchmarks representing the real workload. It does not affect shipped
  code. The new multi-size regression prevents that representative benchmark
  fixture from being mistaken for a supported-size constraint.
- `internal/globalregime/globalregime_test.go` intentionally contains 39,026
  as one of four required regression sizes. `corpus_size_audit_test.go` and
  this report name both historical sizes in order to reject/document them;
  none of these files is production code.

Numeric substrings inside floating-point results (for example
`0.39026...` or `...38887...`) were rejected as search false positives and are
not corpus-size occurrences.

## Small-corpus behavior

Fixed scientific parameters are retained. A corpus shorter than a configured
window is not silently rescaled: `global-regime-analyze` returns a diagnostic
of the form `window size 1000 exceeds corpus length 999`. The regression test
confirms this path returns an error without panic or out-of-bounds access.
Internal window construction also returns no windows when the fixed size is
larger than the available token slice. Existing empty-corpus and
metadata/corpus mismatch checks remain unchanged.

## Tests and validation

- `TestSlidingWindowsUsesActualCorpusSize` checks window counts, every bound,
  and per-window normalization at N=1,000, 8,000, 39,026, and 60,017.
- `TestShortCorpusReportsFixedWindowAsNotApplicable` checks the clear failure
  diagnostic for an inapplicable fixed 1,000-token window.
- `TestNoVoynichCorpusSizeInAlgorithmicSources` prevents either known
  historical size from entering production or orchestration source.
- `go build ./...`: PASS.
- `go vet ./...`: PASS.
- `go test ./...`: PASS.
- `git diff --check`: PASS.

## Voynich byte-equivalence oracle

`global-regime-analyze` was rerun on `data_work/ZL3b-x7.txt` with the frozen
default parameters and a separate output directory. All nine generated files
(YAML, four TSV files, Markdown report, stable-boundary TSV, and two SVG plots)
were byte-for-byte identical to `experiments/voynich-v1/outputs/**`.

In addition, all 340 files in the frozen snapshot passed
`experiments/voynich-v1/checksums.sha256`. Thus the audited stage and complete
stored Voynich Baseline v1 oracle remain byte-identical.
