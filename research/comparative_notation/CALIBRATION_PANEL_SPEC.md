# Calibration panel spec v1 (B01)

Executable implementation: `internal/notation/calibration.go`. CLI:
`notation-corpus calibrate`.

## 1. Purpose

The panel answers "what is the natural magnitude of variation of each
generic metric under controlled structural perturbation?" — never "what
looks like VM?". It is generated entirely from frozen, versioned parameters
and never reads VM or C01-C09 data (`RunCalibrationPanel` takes no corpus
argument at all; `TestC5NoCandidateOrVMDataInGenerators` asserts this
structurally).

## 2. Generators (all frozen, version `1.0`)

| generator_id | preserved | destroyed |
|---|---|---|
| `CAL-IID` | symbol frequency marginal | all sequence/token structure |
| `CAL-MARKOV1` | first-order symbol transition matrix | second-order+ structure |
| `CAL-MARKOV2` | second-order symbol transition structure | third-order+ structure |
| `CAL-CGRAMMAR` | a frozen three-slot (initial/medial/final symbol class) production grammar | everything not expressible by that grammar |
| `CAL-TOKEN-SHUFFLE` | token-frequency marginal | sequence order |
| `CAL-WITHIN-TOKEN-SHUFFLE` | token length distribution | within-token symbol order |
| `CAL-LINE-SHUFFLE` | within-line structure | document progression |
| `CAL-HIERARCHY-SHUFFLE` | lower (line-internal) grammar | hierarchy/page progression |

The four from-scratch generators (`CalibrationGenerators()`) use a shared
24-symbol alphabet and a geometric-like token-length draw
(`poissonish`, mean 5) so their token-formation statistics are on the same
order as VM's without being tuned to match it. `CAL-MARKOV2` caches a fresh
Dirichlet(0.3) transition row per second-order context so its transition
structure is non-uniform but frozen by seed. `CAL-CGRAMMAR` partitions the
alphabet into three fixed 8-symbol classes and always draws the first
symbol from the initial class, then randomly from medial/final.

The four shuffle-derived controls (`shuffleDerivedGeneratorIDs`,
`DeriveShuffleCorpus`) are each applied to their **own replicate's**
`CAL-MARKOV1` draw (never to VM or a candidate), reusing the already-frozen
`ShuffleTokenOrder`/`ShuffleWithinTokens`/`ShuffleLines`/`ShufflePages`
transforms from the metamorphic-test suite.

No generator parameter is tuned against a VM or candidate distance; none of
the eight generators ever receives VM or candidate data as input.

## 3. Sizes and replicate count

Corpora are generated directly at each frozen checkpoint size (5k/10k/20k/
39,380) rather than generated large and rarefied down, because for these
synthetic single-document generators direct generation preserves the
intended hierarchy exactly as well and is cheaper.

`CalibrationCorpora = 40` independent corpora per generator per checkpoint
(the task's stated minimum), frozen before any scale was computed. A
benchmark at the full 39,380-token checkpoint (2 replicates × 8 generators)
measured ≈57s (≈3.6s/corpus); at 40 replicates × 4 checkpoints this
extrapolates to roughly half an hour of one-time CPU work, which was
accepted (no reduction below the task's stated minimum was needed).

## 4. Scale estimator

`EstimateScale`: `s = 1.4826 * MAD(X)` over the pooled sample for a given
`(metric_id, family, support_regime, checkpoint)` stratum (pooled across all
eight generators, all replicates). If `MAD == 0`, falls back to
`IQR / 1.349`; if that is also `0`, the stratum is `DEGENERATE` and never
produces a normalized scalar distance. No epsilon is invented. Only metrics
whose frozen output type (`METRIC_OUTPUT_TYPES.tsv`) is `SCALAR` ever enter
this pool — `DESCRIPTIVE_ONLY`, `CATEGORICAL_DISTRIBUTION`,
`ORDERED_DISTRIBUTION`, and `CURVE` metrics are excluded by construction
(`AnalyzeCalibrationRuns`).

## 5. Stratification key

`metric_id, metric_version, support_regime, checkpoint` — a fresh scale per
combination (`CALIBRATION_SCALES.tsv`). No single scale is reused across
different vocabulary supports or checkpoints.

## 6. Validation (C1-C5)

All five pass in `internal/notation/calibration_test.go`:

- **C1** non-zero coverage: at least 30% of pooled scalar strata are `OK`
  (not `DEGENERATE`) on the test-size panel; production coverage is reported
  in `CALIBRATION_PANEL_REPORT.md`.
- **C2** generator dominance is a diagnostic
  (`generator_observation_counts` in the panel's `--output-diagnostics`
  JSON): every generator contributes the same number of replicates by
  construction, so dominance would only appear via differential
  `NOT_COMPARABLE` rates, which the diagnostic also exposes per generator.
- **C3** leave-one-generator-family-out (`LeaveOneGeneratorFamilyOut`) is
  computed from the *same* cached per-run metric values (no re-analysis),
  confirmed to use the identical frozen estimator, and is reported as a
  diagnostic only — it is never used to pick an estimator
  (`TestC3LeaveOneGeneratorFamilyOutIsDiagnosticOnly`).
- **C4** reproducibility: two independent `RunCalibrationPanel` calls with
  the same seed schedule produce a byte-identical `CALIBRATION_SCALES.tsv`
  (`TestC4CalibrationReproducibility`) — this required the same
  `ShuffleTokenOrder` determinism fix noted in `BOOTSTRAP_PROTOCOL.md`,
  since `CAL-TOKEN-SHUFFLE` calls it directly.
- **C5** no candidate access: `RunCalibrationPanel` has no corpus/candidate
  parameter, and manifests/logs never reference a `C01`-`C09` path.

## 7. B01 decision gate

Per section 33: calibration panel frozen (above); generator parameters
frozen (table in §2, immutable in code); estimator frozen (§4); scale table
produced (`CALIBRATION_SCALES.tsv`); degenerate metrics explicitly marked
(`status=DEGENERATE`); `Compare()` accepts scale only from an explicit
`[]Scale` argument sourced from `CALIBRATION_SCALES.tsv` and errors on a
non-positive spread — there is no pair-derived-scale code path anywhere in
`internal/notation`.

`B01 = CLOSED` once `CALIBRATION_SCALES.tsv` and
`CALIBRATION_PANEL_REPORT.md` are frozen for the production checkpoints
(see `CALIBRATION_PANEL_REPORT.md`).
