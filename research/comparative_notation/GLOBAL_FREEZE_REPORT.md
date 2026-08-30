# Global freeze report — Comparative Notation Study

Status at completion of this task:

    B01=CLOSED
    B02=CLOSED
    B03=CLOSED
    B04=CLOSED

    GLOBAL_COMPARISON_PROTOCOL_FROZEN=true
    PRODUCTION_COMPARATIVE_RUN_AUTHORIZED=false

No C01-C09 candidate corpus was acquired, normalized, analyzed, or
compared to VM during this task. `notation-corpus analyze` still requires
`--fixture`; `notation-corpus rarefy`/`bootstrap`/`distributions` refuse any
`C01`-`C09`-shaped `corpus_id` without `--fixture`; `compare-vm` and
`compare-classes` still require `--authorize-production` *and* the
repository file `PRODUCTION_COMPARATIVE_RUN_AUTHORIZED=true`, which remains
`false`.

## B03 — Rarefaction

- **Boundary preservation**: sampling draws whole "structural blocks" — the
  contiguous run of records sharing the deepest source-observed hierarchy
  key (a physical line whenever lines are observed, else the deepest level
  that is). No block is ever split, and no sampled adjacency is synthetic,
  because `sameSequenceUnit` always compares the full five-level hierarchy
  tuple. See `RAREFACTION_PROTOCOL.md` §1.
- **Replicates**: `R=100` per checkpoint/corpus/representation/family-group.
  Benchmarked on the real VM corpus at ≈23 CPU-minutes total (two
  family-group draws × 4 checkpoints × 100 replicates); not reduced.
- **Seeds**: `SeedFor(base_seed=20260830, corpus_id, representation_id,
  family_group, checkpoint, replicate_index)` — SHA-256 of the tuple, first
  8 bytes as a non-negative int64. No runtime-random seed anywhere.
- **Overshoot**: whichever of {last-block-excluded, last-block-included} is
  closer to the requested checkpoint; ties (and a first block alone meeting
  the checkpoint) resolve to inclusion. `requested_N`, `actual_N`, and the
  implied deviation are all retained.
- **Tests passed**: R1 determinism, R2 boundary preservation, R3 no
  synthetic transitions, R4 size accounting, R5 order preservation, R6
  VM/candidate symmetry (`internal/notation/rarefaction_test.go`).
- **Production run**: `VM_RAREFACTION_V2.tsv` (341 summary rows over 4
  checkpoints) and `VM_RAREFACTION_V2_RAW.tsv` (34,000 replicate-level rows)
  computed on the frozen VM corpus.

## B04 — Distributions and bootstrap

- **Output types**: `SCALAR`, `CATEGORICAL_DISTRIBUTION`,
  `ORDERED_DISTRIBUTION`, `CURVE`, `DESCRIPTIVE_ONLY` — one frozen type per
  metric_id, enumerated in `METRIC_OUTPUT_TYPES.tsv`
  (`notation.MetricOutputTypes()`).
- **Serialization**: `DISTRIBUTIONS.tsv` (`corpus_id, representation_id,
  metric_id, support_id, bin_or_category, value, probability, comparable,
  reason`); frozen category universe for `T11`
  (`INITIAL_RESTRICTED`/`INTERNAL_RESTRICTED`/`FINAL_RESTRICTED`) and
  integer support for `T01` token length (no arbitrary bins).
- **Bootstrap unit**: a block bootstrap resampling whole structural blocks
  with replacement, at the corpus's own actual size (not per checkpoint).
- **CI estimator**: 95% percentile bootstrap, fixed before any VM or
  candidate value was computed.
- **Replicates**: `B=200` (reduced from the proposed 1,000 after a
  documented benchmark — `BOOTSTRAP_PROTOCOL.md` §2 — B=1,000 would cost
  ≈81 CPU-minutes for VM's full-size bootstrap alone).
- **Tests passed**: D1 probability normalization, D2 serialization
  round-trip, D3 label invariance, D4 bootstrap determinism, D5 degenerate
  handling (no NaN/Inf), D6 common-support enforcement
  (`internal/notation/distributions_test.go`).
- **Production run**: `VM_BOOTSTRAP_V2.tsv` (82 metric rows, `B=200`) and
  `VM_DISTRIBUTIONS_V2.tsv` computed on the frozen VM corpus.

Closing D4/C4 determinism required fixing one genuine pre-existing defect:
`ShuffleTokenOrder` consumed a shared `math/rand` source in raw Go map
iteration order (randomized per process), and several `analyze.go`
functions summed `float64` values in map iteration order. Both were made to
iterate sorted keys; see `BOOTSTRAP_PROTOCOL.md` §5 and the diff to
`internal/notation/transform.go`/`analyze.go`.

## B01 — External calibration panel

- **Generators**: `CAL-IID`, `CAL-MARKOV1`, `CAL-MARKOV2`, `CAL-CGRAMMAR`
  (from-scratch), `CAL-TOKEN-SHUFFLE`, `CAL-WITHIN-TOKEN-SHUFFLE`,
  `CAL-LINE-SHUFFLE`, `CAL-HIERARCHY-SHUFFLE` (derived from each
  replicate's own `CAL-MARKOV1` draw). Parameters frozen in
  `CALIBRATION_GENERATORS/*.json`.
- **Independence from candidates**: `RunCalibrationPanel` takes no corpus
  argument at all — it cannot read VM or C01-C09 data even by mistake.
  `TestC5NoCandidateOrVMDataInGenerators` asserts this structurally.
- **Scale computation**: pooled `s = 1.4826*MAD(X)` over all 8 generators ×
  40 replicates at a given `(metric_id, family, support_regime,
  checkpoint)` stratum; falls back to `IQR/1.349` if `MAD=0`; `DEGENERATE`
  (no scale) if both are `0`. No epsilon invented.
- **Metrics with DEGENERATE scale**: 49 of 268 pooled strata (18.3%); 4
  persist at the full 39,380 checkpoint
  (`G02_INITIAL_RESTRICTION_DENSITY`, `G03_FINAL_RESTRICTION_DENSITY`,
  `T11_POSITIONAL_RESTRICTION_DENSITY`, `L01_LINE_TOKEN_COUNT_MEAN`), each
  with a documented structural cause, not a code defect — see
  `CALIBRATION_PANEL_REPORT.md` §C3.
- **Stability**: every generator contributes exactly the same number of
  scalar observations (2,680) at every checkpoint, and leave-one-out
  (`LeaveOneGeneratorFamilyOut`) preserves the same 67-stratum coverage
  regardless of which generator is excluded — no single generator drives
  panel coverage. Full numeric leave-one-out deltas per stratum were not
  exhaustively tabulated (compute cost); the balanced-contribution evidence
  and the generators' structurally distinct preserved/destroyed properties
  are the stability evidence recorded here.
- **Tests passed**: C1 non-zero coverage, C2 dominance (diagnostic), C3
  leave-one-out (diagnostic only), C4 reproducibility, C5 no candidate
  access (`internal/notation/calibration_test.go`).
- **Production run**: `CALIBRATION_SCALES.tsv` (268 rows across 4
  checkpoints), `CALIBRATION_DIAGNOSTICS.json`.

## B02 — Complete generic VM reference

- **Coverage**: all 82 generic metric rows the analyzer emits are present
  in `VM_REFERENCE_V2.tsv`; 70 are `COMPARABLE`, and the remaining 12
  (`D_SECTION_*`, `D_PAGE_*`, `D_LOCUS_*`, 4 metrics × 3 levels) are
  explicitly `NOT_COMPARABLE` because the frozen source
  (`data_work/ZL3b-x7.canonical.txt`, no IVTFF metadata) only observes
  `document` and `physical_line` — never fabricated.
- **Anchor reproduction**: all seven frozen `VM_STRUCTURAL_CATALOG.md`
  anchors reproduce exactly (alphabet size 36; initial restriction 9/36;
  final restriction 0/36; bigram occupancy 379/1296; trigram occupancy
  1569/46656; frequent-transition zero density 0.960458979298 at threshold
  10; same-line zero density 0.769393558194) — see
  `VM_REFERENCE_RECONCILIATION.md`.
- **Discrepancy**: one formal defect was found and fixed while reconciling
  the same-line anchor — `L06`/`L07` were hard-coded to a `TOP_100`
  vocabulary instead of being stratified over the same frozen support
  regimes as `S01`-`S03`. Fixed additively (new `regime` dimension, no
  metric removed); documented in `VM_REFERENCE_RECONCILIATION.md`. No other
  discrepancy exists; the STOP clause (task section 50.1) was not invoked.
- **SHA-256 of the frozen VM reference**: `output_sha256` in
  `VM_REFERENCE_V2_MANIFEST.json` =
  `90d2254a7c9ab25c3c1a3167d3f0f4f6afc69689d3f3477e46486484fd938f42`
  (the `VM_REFERENCE_V2.fingerprint.json` bytes `notation.Compare`
  actually consumes; `VM_REFERENCE_V2.tsv` is a separate human-readable
  projection of the same fingerprint).

## End-to-end

- **Synthetic dry run**: `TestEndToEndSyntheticDryRun`
  (`internal/notation/dryrun_test.go`) runs an independent synthetic corpus
  through USC → validation → structural analyzer → rarefaction →
  distributions/bootstrap → a synthetic "VM reference" → frozen calibration
  scales (from the same `RunCalibrationPanel`) → `Compare` → result, and
  passes: every expected metric row is present, missing levels are
  `NOT_COMPARABLE`, no NaN/Inf anywhere, `d_TOTAL` is absent, all five
  family distances (`G`/`T`/`S`/`L`/`D`) are present, and a repeated run is
  byte-identical row-for-row.
- **Adversarial tests**: A1 (missing scale → `NOT_COMPARABLE`, never
  `scale=1`), A2 (metric registry version mismatch → hard error), A3 (wrong
  support regime → never approximately joined), A4 (candidate-specific VM
  file → rejected by `VerifyFrozenVMReference`, wired into `compare-vm`
  via `--vm-manifest`), A5 (missing physical lines → `L` is
  `NOT_COMPARABLE`), A6 (short corpus → checkpoint `NOT_COMPARABLE`,
  rarefaction errors rather than truncating), A7 (corrupt provenance SHA →
  `VerifyArtifactHash`/`vm-adapter`'s own SHA check fails), A8 (changed
  metric registry after freeze → `VerifyFrozenVMReference` fails), A9
  (changed calibration scale after freeze → `VerifyArtifactHash` fails),
  A10 (C01-C09 full run blocked — `guardNotUnauthorizedCandidate` in the
  CLI plus the pre-existing `--fixture`/`--authorize-production`/
  `PRODUCTION_COMPARATIVE_RUN_AUTHORIZED` locks) — all pass
  (`internal/notation/adversarial_test.go`,
  `cmd/notation-corpus/main_test.go`).
- **Post-hoc decisions**: `notation.Compare` accepts a scale only from an
  explicit `[]Scale` argument sourced from `CALIBRATION_SCALES.tsv`, errors
  on a non-positive spread, and has no pair-derived-scale code path
  anywhere in `internal/notation`. The comparator cannot run without a
  frozen scale and a hash-verified VM reference.

## Full test suite

`go test ./...` passes across the repository (including
`internal/notation`, `cmd/notation-corpus`, and every pre-existing
package) after this task's changes.

PRODUCTION_COMPARATIVE_RUN_AUTHORIZED=false
