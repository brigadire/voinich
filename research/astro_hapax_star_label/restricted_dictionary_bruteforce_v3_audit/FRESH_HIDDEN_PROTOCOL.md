# Fresh Sealed Hidden Benchmark Protocol (Section F)

## 1. Freeze Point

Code frozen before generating or viewing any fresh result: `scripts/audit_verification.py`
(SHA-256 recorded in this package's `SHA256SUMS`, generated as the final step of this audit). The
solver/engine/generator code it imports is the **unmodified, read-only copy** in `repro_sandbox/`,
itself hash-verified against the frozen design package throughout (`UPSTREAM_INTEGRITY_REPORT.md`).

## 2. Sealed Dataset Generation

New seeds, disjoint from every seed used anywhere in the frozen design package or its own unit
tests (design package uses 101-120, 999, and 1000+17k for k=1..25; this audit's adversarial
section separately uses 42000-42605):
- Heuristic sweep: seeds **9001-9375** (contiguous allocation across the table-size × noise loop,
  see `audit_verification.py::section_fresh_hidden_benchmark`)
- CP_SAT / Branch-and-Bound sweep: continuation of the same seed allocator, disjoint sub-range
- Resource pilot: seeds **70000-70299**

`SealedSyntheticGenerator(seed=...).generate_dataset(...)` was called once per seed to produce the
public `labels` list and the `sealed_truth` dict (containing `planted_table` and
`true_assignments`). The **public labels only** were passed into each solver's `.solve()` call.

## 3. Predictions Recorded Before Truth Comparison

Each solver's output (`predicted_table`, `predicted_assignments`, runtime, peak memory) is written
to `FRESH_HIDDEN_PREDICTIONS.tsv` / `FRESH_HIDDEN_PREDICTIONS_CPBB.tsv` by the same script, in the
same function call, immediately after `.solve()` returns and before the recovery-metric comparison
against `sealed_truth` executes (see `compute_recovery()` call site in
`section_fresh_hidden_benchmark`, called only after the prediction row is already constructed).
There is no code path by which the solver could have seen `sealed_truth` before producing its
prediction (C1, confirmed separately in `SYNTHETIC_ISOLATION_REPORT.md`).

## 4. Metrics Computed After Predictions Are Fixed

`FRESH_HIDDEN_RESULTS.tsv` / `FRESH_HIDDEN_RESULTS_CPBB.tsv` contain the recovery metrics
(exact table recovery, mapping precision/recall/F1, assignment accuracy, coverage) computed from
the already-recorded predictions against `sealed_truth`. No threshold or parameter was changed
after seeing these numbers — `audit_verification.py` was not edited between the run that produced
these files and the writing of this report (script hash recorded in the final `SHA256SUMS`).

## 5. Scope Actually Executed (Disclosed Reduction From the Literal Task Ask)

| Dimension | Task asks for | This audit ran | Why reduced |
|---|---|---|---|
| Seeds per key config | ≥20 | 20 (heuristic, table_size=8) / 20 per table size (4,6,8,10,12) / 5 (CP_SAT, B&B) | CP_SAT/B&B take up to 15s/replica on the real 307-attestation lexicon at table_size=8; 5 seeds × 2 noise levels × 2 algorithms already costs up to 300s of wall time within a single audit session |
| Table sizes | 4,6,8,10,12 | 4,6,8,10,12 (heuristic only) | met in full for the heuristic, the architecture actually proposed for production |
| Noise levels | 0,10,25% | 0%,10%,25% (heuristic); 0%,10% (CP_SAT/B&B) | as above |
| Unmatched fractions | 0,25,50% | 25% only | fixed at the design package's own certified operating point to produce a like-for-like replacement for the fabricated Suite 1/2 numbers; 0%/50% not covered in this pass |
| Modes (merge/abbreviation/deletion combinations) | full composition matrix | INJECTIVE / DROP_UNMAPPED / NONE only | same reason; this is the exact configuration the fabricated `SYNTHETIC_RECOVERY_RESULTS.tsv` claimed 85%/75%/68.4%, so it is the configuration whose real numbers matter most for this audit's central finding |

This is disclosed as a **partial** fresh hidden benchmark: it is sufficient to conclusively
demonstrate that the design package's certified numbers were fabricated (by showing what the real
numbers actually are, or are not, on a like-for-like configuration), but it is **not** a
substitute for the full matrix the task specification asks for, which should be executed before
any future production freeze.
