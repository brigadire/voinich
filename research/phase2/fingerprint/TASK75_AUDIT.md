# Task75 audit (task77 stage 1)

Status: **audit complete**. This is the mandatory pre-computation audit
required by `tasks_ph2/task77.txt` §Этап 1: it checks reproducibility,
infrastructure adequacy and pipeline correctness of Task75's LP1-LP4/EF1-EF4
block *before* any new cross-scale metric is trusted. Verdicts use exactly
the four values task77 requires: `REPRODUCED`, `REPRODUCED_WITH_DEVIATION`,
`NOT_REPRODUCED`, `NOT_TESTABLE`.

Method: re-read every line of `internal/fingerprintv2`, re-ran its full unit
and integration test suite, and — critically — ran the pipeline for the
first time against the real frozen corpus (`data/ZL3b-n.txt` +
`data_work/ZL3b-x7.canonical.txt`, both present locally under this
repository's external-data discipline; see `DATA.md`). Task75's own report
explicitly states it never had corpus bytes available and never exercised
this path. Several defects below were found only because this audit is the
first real-corpus run this pipeline has ever had.

## 1. LP1-LP4 reproducibility

| Item | Verdict | Notes |
|---|---|---|
| LP1 census determinism | `REPRODUCED` | `TestSeededPipelineIsDeterministic` reruns the whole pipeline twice from the same seed/config and byte-compares the JSON fingerprint; still passes. |
| LP1 rule taxonomy correctness | `REPRODUCED` | Directed-pair classification (`ruleFor`), zone/position-class labeling reviewed against `tokenrepetition.ClassifyEditDistanceOne`/`PositionClass`; consistent with Task60's declared convention. |
| LP2 C-GRAMMAR/C-LEN/C-FREQ nulls and BH FDR | `REPRODUCED_WITH_DEVIATION` | The BH implementation (`fdr`) is correct (verified by hand-tracing the cummin-from-the-top procedure). The deviation is infrastructural, not statistical: the frequency-aware C-GRAMMAR generator's per-length uniqueness search (`uniqueForms`) had an attempt budget (`want*200`, min 1000) tuned only against small fixtures; on the real corpus's skewed short-length glyph distributions it failed outright (see §4). Fixed by raising the budget (`want*5000`, min 20000); LP2 itself was never wrong, it just could not run to completion. |
| LP3 family/branching/depth/overlap/locality | `REPRODUCED_WITH_DEVIATION` | Logic is correct, but `lp3()` rebuilt `glyphByToken(c)` (an O(corpus size) map construction) inside a per-edge, per-component loop — O(edges × corpus size) — which is silently fine on tiny fixtures and catastrophic on the real ~39,000-token/~8,200-type corpus. Fixed by hoisting the map build once per call. No output changed; only wall-clock did. |
| LP4 core/affix MI and permutation null | `REPRODUCED_WITH_DEVIATION` | Same class of defect: `attachmentValues`, `eligibleAttachmentCount`, `attachmentPermutation` and `attachmentTriples` each called `glyphByToken(c)` *inside* a loop over `vocabulary(c)`, rebuilding the whole map on every vocabulary token. This alone made a single LP4 call take **over 70 seconds** on the real corpus (measured directly, see §4). Fixed by hoisting the map build once per function call; LP4's formulas and null construction are otherwise unchanged and verified correct. |

## 2. EF1-EF4 reproducibility

| Item | Verdict | Notes |
|---|---|---|
| EF1 degree/component distributions | `REPRODUCED` | Straightforward BFS-based components; determinism verified. |
| EF2 clustering/triangle/4-cycle census | `REPRODUCED` | Algorithm re-derived by hand on a small graph and matches; on the real corpus's actual (non-pathological) degree distribution (max degree 75-83, `sum(deg^2)` ~1.5-2.7M) it runs in well under a second, so no change was needed here — the earlier suspicion that this exact metric was the bottleneck (a plausible hypothesis given a known "giant-component perf gotcha" from Task66's history) was checked directly with stage-level timing and **ruled out**. |
| EF3 Spearman degree/log-frequency + C-FREQ control | `REPRODUCED` | No issues found. |
| EF4 consolidated grammar-boundedness verdict | `REPRODUCED_WITH_DEVIATION` | Correct given valid EF1-EF3/LP2 inputs; inherits the LP2 C-GRAMMAR deviation above. |

## 3. C-GRAMMAR adequacy

`structure-preserving` mode validated cleanly on the real corpus (exact
token-count/length/alphabet marginals held; positional/endpoint/bigram
total-variation diagnostics were inspected and are within the configured
tolerance in every run observed during this audit).

`frequency-aware` mode, once the attempt-budget fix (§4) was applied, also
completed; its diagnostics are reported in `raw_results.json` for the
canonical run (see `TASK77_REPORT.md`) rather than duplicated here.

Verdict: `REPRODUCED_WITH_DEVIATION` (the generator-adequacy mechanism is
sound; its *tuning* was fixture-only and needed a real-corpus correction).

## 4. Corpus partitions, identifiers and infrastructure defects found and fixed

This is the audit item task77 explicitly calls out as in-scope
("Исправление локальных дефектов pipeline входит в задачу"). All four are
now fixed and covered by tests/regression; none required redesigning
Task75's approach.

1. **IVTFF alignment normalization was wrong on real data.**
   `internal/metadatavalidation.NormalizeForAlignment` treated apostrophe
   (`'`), question mark (`?`), `@` and `;` as ordinary content. Against the
   real `data/ZL3b-n.txt` + `data_work/ZL3b-x7.canonical.txt` pair, strict
   alignment failed at token 47 of the very first folio. Byte-level
   comparison against the canonical corpus showed all four characters are
   always word breaks in the historical `-x7` conversion (the canonical
   file contains **zero** apostrophes, question marks, `@`s or `;`s; e.g.
   `{c'y}` → `c y`, `d?n` → `d n`, bare `@192;chy` → `192 chy`). Fixed by
   adding all four to the existing word-break replacer, with the exact
   evidence recorded in code comments and in
   `internal/metadatavalidation/metadatavalidation_test.go`. This was a
   **correctness** bug, not a performance one: no corpus using
   `ivtff_path` had ever been checked against real IVTFF text before.
   `Section` was also added to `TokenMetadata` (from `$I`), which task75
   never populated because nothing consumed it yet.
2. **`k`-core decomposition had a peeling-order bug.** The Batagelj-
   Zaversnik degeneracy algorithm requires the *running max* of removal-
   time degree to be threaded through the peel order live; the first
   implementation instead re-sorted by raw removal-degree afterward, which
   is not equivalent when ties in the peel order are broken arbitrarily. A
   synthetic triangle-plus-pendant unit test (`TestKCoreOnTriangleWithPendant`)
   caught this immediately (`b` and `c` reported coreness 1/0 instead of
   the correct 2/2). This is new task77 code, not a task75 regression, but
   is recorded here because it directly affects `CORE`/`PERIPHERY` family
   roles used throughout the cross-scale block.
3. **O(vocabulary × corpus) map rebuilds in LP3/LP4.** Detailed in §1
   above; measured directly with stage-level timing
   (`analyzeBare` on the real corpus dropped from **72.7s to 0.67s** after
   the fix; a full `lp3()` call similarly dropped from an unmeasured
   multi-minute stall to sub-second). Without this fix, the canonical run
   required by task77 was not computationally feasible at all — a single
   corpus analysis (never mind the grammar replicates, stability battery,
   or cross-scale block, several of which call `analyzeBare`/`lp3`
   repeatedly) would not have completed in bounded time.
4. **Frequency-aware C-GRAMMAR generator attempt budget.** Detailed in §1;
   fixed by raising `uniqueForms`'s attempt cap from `max(1000, want*200)`
   to `max(20000, want*5000)`. This does not change what the generator
   produces when it succeeds, only how hard it tries before declaring
   failure, and is safe for existing fixtures (which need far fewer
   attempts already).

## 5. Leakage between train/held-out partitions

Task75 itself defines no held-out split, so there is nothing to audit for
leakage in Task75's own code. Verdict: `NOT_TESTABLE` (feature did not
exist). Task77's own grouped-folio held-out validation (§8 of the task) is
new and is audited separately in `TASK77_REPORT.md`.

## 6. Multiple-testing correction

`REPRODUCED`. Traced `fdr()` by hand against the standard Benjamini-Hochberg
definition (`q_(i) = min_{j>=i} p_(j)*n/j`, cumulative minimum taken from
the largest p-value down): the implementation's "sort descending by p,
carry the running minimum down" loop computes exactly this. No defect
found.

## 7. Dependence on rare/high-frequency tokens

`NOT_TESTABLE` under Task75 alone (no such sensitivity sweep existed).
Task77 adds this directly as part of its own stability battery (rare-token
cutoff is one of the required perturbations); see `TASK77_REPORT.md` §9.

## 8. Robustness to preprocessing profiles

`NOT_TESTABLE` under Task75 alone. Task77's stability battery reports a
`preprocessing_profile` row as `NOT_TESTABLE` for the *same-vocabulary*
comparison (EVA and natural glyph modes tokenize into disjoint alphabets,
so ARI/NMI over a shared node set is undefined by construction, not because
the check was skipped) and instead documents a structural comparison; see
`TASK77_REPORT.md` §9.

## 9. Machine-readable output vs. declared schema

`REPRODUCED`. `fingerprint.json`/`raw_results.json` fields match
`FINGERPRINT_V2_SCHEMA.md` for every field task75 declared. Task77 adds new
fields (`metrics.ef5`, `edit_graph_validation`, `cross_scale`) which are
additive and documented in the schema update accompanying this audit.

## Overall

`TASK75_RESULTS_REPRODUCED`: **PARTIALLY_SUPPORTED**. Every formula and
null-model construction in LP1-LP4/EF1-EF4 was confirmed correct by direct
re-derivation and by running it (for the first time) against real data. The
"deviation" in four of the above rows is entirely about infrastructure
readiness for a real-scale corpus (alignment normalization, two
performance defects, one generator-tuning defect), not about any reported
statistic being wrong. All four are fixed in this branch and covered by
regression tests or the existing determinism test. No defect required
redesigning Task75's statistical approach, consistent with task77's
instruction that local-defect fixes are in scope but a Task75 redesign is
not.
