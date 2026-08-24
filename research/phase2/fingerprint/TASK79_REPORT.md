# Task79 report — page structure, hierarchy, stability and freeze readiness

## Result

Task79 is implemented and the canonical run completed.  The output is a
versioned **`FINGERPRINT_V2_CANDIDATE`**, not a frozen fingerprint.  The formal
decision is **`TASK79_B_REQUIRED`**: page/hierarchy structure is measurable and
mostly stable across folio halves, but cross-transcription stability,
historical notation/table controls and out-of-sample hierarchical-vs-flat
validation remain critical gaps.  No `FINGERPRINT_V2_FROZEN` marker was
created.

Canonical configuration:
`experiments/fingerprint-v2-task79-v1/canonical.yaml`; output:
`experiments/fingerprint-v2-task79-v1/canonical-out/`.  Corpus SHA-256 is
`f46f4190af65b85d145ec5bb957c1f56029b567e4bef12ac7baa1797f358d692`;
configuration SHA-256 is
`0c3c571d260fac8e4b4f337cf1ab42f1f21f923ed8480eaf4ea32f68b17042e7`.
The run used seed `20260824`, 100 task75/task77 repetitions, 1,000 Task79
permutations and 1,000 folio-bootstrap replicates.

## Input and metadata audit

LP1–LP4, EF1–EF5, C-GRAMMAR, edit-family validation and cross-scale code were
recomputed in the same invocation.  Their blocks are `REPRODUCED`; the sole
`NOT_TESTABLE` audit item is cross-transcription stability because only one
strictly aligned transcription is locally available.  Random-sampled task77
statistics use the declared Task79 seed; their qualitative conclusions are
unchanged and raw distributions are retained.

Strict IVTFF alignment yielded 39,380 tokens, 33,995 within-line transitions,
5,385 lines, 5,385 loci and 227 folios, with zero nesting violations and zero
missing folio/locus-type/section/hand values.  Currier is absent for 14,595
tokens and `$X` for 36,181, so those contrasts use complete cases.  The known
`f116r.37` case is present with 13 occurrences; the versioned correction layer
therefore records `VERIFIED_PRESENT_NO_PATCH` and never mutates the frozen
source.

The factual IVTFF model differs from the ideal hierarchy in one important
way: each parsed textual locus record is also one line, hence
`token -> line == locus -> folio -> $I/section`.  Separate LC3/LC4 within-locus
versus between-line inference is not identifiable from these records.  Locus
type, paragraph and `$I`/`$X` classes remain usable categorical metadata.  The
39,380-row `occurrence_metadata.jsonl` records every requested field and its
missing status.

## Canonical findings

All p-values below are empirical and all q-values use one BH family over the
new Task79 confirmatory tests.

| Metric | Observed | Effect vs null (SD) | q | Result |
|---|---:|---:|---:|---|
| LS2 positional lexical NMI | 0.11059 | 76.57 | 0.00142 | supported |
| LS3 boundary length asymmetry | 0.36065 | 16.56 | 0.00142 | supported |
| LS4 adjacent exact repetition | 0.00912 | 0.49 | 0.399 | not supported |
| BP1 boundary-token NMI | 0.08422 | 86.86 | 0.00142 | supported |
| LC1 locus-type NMI | 0.00759 | 23.40 | 0.00142 | supported |
| LC2 label/text NMI | 0.00485 | 25.56 | 0.00142 | supported |
| PF2 folio coherence | 0.17304 | 29.69 | 0.00142 | supported |
| PF3 adjacent-folio continuity | 0.49281 | -5.18 | 1.0 | not supported |
| PF4 recto/verso coherence | 0.51920 | -2.97 | 1.0 | not supported |
| PF5 within-folio progression | 0.36370 | 14.72 | 0.00131 | supported |
| HR1 folio variance share | 0.27327 | 52.50 | 0.00142 | supported |
| HR1 section variance share | 0.12526 | 181.54 | 0.00142 | supported |
| 2DL1 layout-position MI | 0.00471 | 17.32 | 0.00142 | supported |

LS1 line-length CV is `0.70654`, with a 95% hierarchical folio-bootstrap
interval `[0.61222, 0.81057]`.  Currier/section association is large
(`NMI=0.64862`), demonstrating the expected confounding rather than independent
regimes.  Eight exploratory profile-distance peaks passed the fixed
mean-plus-two-SD penalty; 4/8 coincide with section changes.  They are
`PARTIALLY_SUPPORTED` and cannot be interpreted as topics, languages or coding
stages.

Boundary effects establish only additional structure at line edges.  They do
not establish an acrostic, telestich, grille, plaintext, direction of reading
or cipher.  Likewise, `2D-LITE` uses IVTFF order and categorical markers, not
scan coordinates.

Negative results are conservatively registered as `ABSENCE_OF_EVIDENCE`, not
`EVIDENCE_OF_ABSENCE`, because no equivalence margin was preregistered.  This
applies in particular to LS4, PF3 and PF4.  The recto/verso verdict is therefore
`INCONCLUSIVE` despite a non-significant one-sided test.

## Stability, redundancy and controls

All retained new core candidates reproduce direction across deterministic
alternating-folio halves; LS1 additionally has the hierarchical bootstrap
above.  Transcription stability is `INSUFFICIENT_DATA`, so the aggregate
stability verdict is only `PARTIALLY_SUPPORTED`.  The line-profile redundancy
matrix shows the largest expected correlations between line length and token
entropy (`r=0.845`) / transition entropy (`r=0.872`); entropy summaries remain
supporting rather than additional core evidence.

Pseudo-blind comparison against the configured Doyle control was performed
only after metric definitions were fixed.  Twenty-one eligible contrasts are
serialized.  Metadata-dependent LC/PF/HR metrics are excluded for that control
because it lacks IVTFF metadata.  The result demonstrates a functioning
comparison interface, not a trained classifier or adequate discriminative
portfolio.

The candidate registry contains 33 summaries: the new Task79 metrics plus
unchanged LP/EF/cross-scale summaries.  Thirteen are candidate `CORE` metrics
and twenty are `SUPPORTING`; this classification is not a freeze because the
required stability/control gates have not passed.

## Artifact map

The canonical output contains the full fingerprint/candidate, copied config,
raw null distributions, occurrence metadata, line profiles, Task75/77 audit,
IVTFF audit, metric and null registries, stability/redundancy matrices,
coverage and negative-evidence registries, discrimination and segmentation
results, freeze manifest, verdicts, warnings/errors and compact report.
`TASK79_METADATA_CORRECTIONS_V1.yaml` is the separate patch layer;
`TASK79_B_SCOPE.md` is the bounded follow-up scope.

## Required verdicts

| Verdict | Value |
|---|---|
| `INPUT_RESULTS_REPRODUCED` | `PARTIALLY_SUPPORTED` |
| `METADATA_INTEGRITY_ACCEPTABLE` | `SUPPORTED` |
| `LINE_STRUCTURE_SUPPORTED` | `SUPPORTED` |
| `BOUNDARY_STRUCTURE_SUPPORTED` | `SUPPORTED` |
| `LOCUS_STRUCTURE_SUPPORTED` | `SUPPORTED` |
| `FOLIO_STRUCTURE_SUPPORTED` | `SUPPORTED` |
| `RECTO_VERSO_DEPENDENCE_SUPPORTED` | `INCONCLUSIVE` |
| `HIERARCHICAL_ORGANIZATION_SUPPORTED` | `SUPPORTED` |
| `HIERARCHICAL_MODEL_OUTPERFORMS_FLAT` | `INCONCLUSIVE` |
| `UNANNOTATED_CHANGE_POINTS_SUPPORTED` | `PARTIALLY_SUPPORTED` |
| `CORE_METRICS_STABLE` | `PARTIALLY_SUPPORTED` |
| `CORE_METRICS_NONREDUNDANT` | `PARTIALLY_SUPPORTED` |
| `NEGATIVE_EVIDENCE_REGISTERED` | `SUPPORTED` |
| `ALTERNATIVE_EXPLANATION_COVERAGE_ACCEPTABLE` | `NOT_SUPPORTED` |
| `MODEL_COMPARISON_INTERFACE_READY` | `PARTIALLY_SUPPORTED` |
| `FINGERPRINT_V2_FREEZE_STATUS` | **`TASK79_B_REQUIRED`** |

The decisive blockers are not the absence of significant page effects.  They
are the unavailable second transcription, insufficient historical
shorthand/table/procedural positive controls, non-identifiable line-versus-
locus level, and unvalidated out-of-sample hierarchical-vs-flat comparison.
The exact permitted follow-up is frozen in `TASK79_B_SCOPE.md`; it excludes
Fontana models and any post-hoc tuning against future model outputs.
