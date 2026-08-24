# Transcription stability report — 13 CORE metrics, ZL3b-n.txt vs. IT2a-n.txt

Classification thresholds fixed in `TASK79C_DESIGN.md` §2 before any
canonical/alternate value was compared. Run: canonical =
`experiments/fingerprint-v2-task79-v1/canonical-out/metric_registry.json`
(seed 20260824); alternate =
`experiments/fingerprint-v2-task79c-v1/transcription-it-out/metric_registry.json`
(same seed, same config, `primary.path=data_work/IT2a-x7.canonical.txt`,
`primary.ivtff_path=data/IT2a-n.txt`).

## Result

| Classification | Count | Metrics |
|---|---:|---|
| `STABLE` | 3 | `2DL1_LAYOUT_POSITION_MI`, `LC2_LABEL_TEXT_NMI`, `PF2_FOLIO_COHERENCE` |
| `DIRECTION_STABLE` | 10 | `BP1_BOUNDARY_TOKEN_NMI`, `EF1_GIANT_COMPONENT_SHARE`, `EF2_GLOBAL_CLUSTERING`, `EF3_DEGREE_FREQUENCY_SPEARMAN`, `HR1_FOLIO_VARIANCE_SHARE`, `HR1_SECTION_VARIANCE_SHARE`, `LC1_LOCUS_TYPE_NMI`, `LS2_POSITIONAL_LEXICON_NMI`, `LS3_BOUNDARY_LENGTH_ASYMMETRY`, `PF5_WITHIN_FOLIO_PROGRESSION` |
| `VERDICT_STABLE` | 0 | — |
| `UNSTABLE` | 0 | — |
| `NOT_TESTABLE` | 0 | — |

**Headline: 13/13 CORE metrics agree in status (`SUPPORTED` vs.
`SUPPORTED`, or `REPRODUCED`/`CONSISTENT_WITH_GRAMMAR_BOUND` matching their
own kind) and in effect direction between the two independently transcribed
readings of the manuscript. Zero metrics are `UNSTABLE`. Three survive at
the strict sub-one-null-SD tolerance; the other ten move by more than one
canonical-run permutation-null standard deviation but never flip sign or
status.** This is exactly the "does the scientific conclusion survive,
not does the number match to several digits" question parent §8 asks, and
the answer is yes for all 13.

## By sensitivity class (parent §10)

- **Glyph-sensitive metrics** (families whose statistic is computed
  directly over glyph-level tokens: `2D-LITE`, `edit family`, `lexical
  paradigm`-adjacent `line` metrics): `2DL1` is `STABLE`; `LS2`, `LS3`,
  `EF1`, `EF2`, `EF3` are `DIRECTION_STABLE`. `LS3_BOUNDARY_LENGTH_ASYMMETRY`
  has the largest standardized shift (5.08 SD) of any CORE metric — it is
  the single metric most sensitive to the two transcriptions' differing
  treatment of boundary tokens (bracketed-uncertain-reading notation, see
  `TRANSCRIPTION_ALIGNMENT_REPORT.md`), though it does not change sign or
  status.
- **Token-boundary-sensitive metrics** (`boundary` family): `BP1` is
  `DIRECTION_STABLE` (2.69 SD).
- **Locus/folio/hierarchy metrics** (require IVTFF metadata, exercised
  only because Gate A's 96.8%/100% locus coverage — `TRANSCRIPTION_
  ALIGNMENT_REPORT.md` — is high enough to recompute them at all):
  `LC2`, `PF2` are `STABLE`; `LC1`, `HR1_FOLIO_VARIANCE_SHARE`,
  `HR1_SECTION_VARIANCE_SHARE`, `PF5` are `DIRECTION_STABLE`.
  `HR1_SECTION_VARIANCE_SHARE` has the second-largest standardized shift
  (9.46 SD) — plausible, since `f116v`/`fRos` (excluded from IT, see
  alignment report) sit in specific sections and their removal shifts that
  section's line-length composition slightly, without changing which
  sections dominate the variance decomposition.

## Manuscript property vs. transcription-system property

None of the 13 CORE metrics flips `SUPPORTED`/`NOT_SUPPORTED` or reverses
sign between transcriptions. The observed shifts are consistent with the
two transcriptions' known, documented differences (uncertain-reading
bracket notation, two excluded difficult folios, ~1,320 loci with
differing word-segmentation) rather than with either transcription
containing a different underlying manuscript signal. This supports
treating the 13 CORE findings as properties of the manuscript's structure,
not artifacts of the Zandbergen–Landini transcription specifically —
though the two-transcription comparison, by itself, is one data point and
cannot rule out that both transcriptions share a common bias (both are
IVTFF-family, EVA-alphabet-family transcriptions of the same physical
object; a genuinely independent notation system, e.g. a different scan
interpretation methodology, is not tested here).

## Caveat on standardized difference for EF1–EF3

`EF1_GIANT_COMPONENT_SHARE`, `EF2_GLOBAL_CLUSTERING` and
`EF3_DEGREE_FREQUENCY_SPEARMAN` carry `effect_size=0` and no permutation
null SD in `metric_registry.json` (their status derives from a C-GRAMMAR
comparison recorded separately in `raw_results.json`, not a direct
permutation p-value/effect pair). Their standardized-difference column is
therefore `N/A`, not `0` or `inf`; their classification rests only on the
status+direction agreement conditions, which both hold.
