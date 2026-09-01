# Matching-selection bias audit

## Decision

`MATCHING_SELECTION_BIAS=COULD_EXPLAIN_EFFECT`.

The earlier enrichment is robust within the mapped data, but mapping is not composition-neutral. At the physical-coordinate level there are 89 matched and 54 unmatched labels. Panel and family composition, coordinate-bridge availability, and exact lexical anchors differ between the two groups; the entire `f68v2.X/Y` sector series is unmatched. This violates the conditions required for `UNLIKELY_TO_EXPLAIN_EFFECT`.

The audit does **not** show that selection bias caused the enrichment. The high-coverage, no-f68v2, and star-only confirmed subsets all retain positive enrichment, and panel/family-background imputation retains the direction. However, the status of the unmapped labels is unobserved, and the formal pessimistic bound reverses the effect. The defensible conclusion is therefore that matching incompleteness *could* explain some or all of the observed enrichment.

## Units and pre-hapax features

- Primary selection unit: 143 physical `panel.group.number` coordinates (89 matched, 54 unmatched).
- Secondary descriptive unit: 191 Stolfi source records (130 matched, 61 unmatched). Record variants at one coordinate are not treated as independent.
- Family definitions are frozen from Stolfi fields: `STAR`, `PLANET_MOON`, `CIRCLE_SECTOR`, `RADIAL_TEXT`, `SECTOR_LABEL`, `OUTER_TITLE`, and `OTHER_DIAGRAM_LABEL`.
- Glyph and token lengths come only from the Stolfi EVA strings. `*` is recorded as a wildcard and excluded from known-glyph length.
- Exact-anchor availability is `raw_lexical_candidate_count > 0`; bridge/type fields are taken from the completed matching audit before any hapax status is joined.
- Actual mapped locus type is reported descriptively only: `NONE` for an unmatched coordinate is a consequence of matching and cannot serve as an independent predictor.

Association strength uses Cramer's V for categorical features and pooled standardized mean difference for numeric features. Coordinate-level p-values are two-sided, fixed-matched-count permutation tests with 10,000 draws. These diagnose selection structure; they are not causal tests.

## Composition findings

| feature | effect measure | effect | permutation p | band |
|---|---|---:|---:|---|
| PANEL | CRAMERS_V | 0.497179667 | 0.000099990 | STRONG |
| LABEL_FAMILY | CRAMERS_V | 0.564105221 | 0.000099990 | STRONG |
| STOLFI_GROUP | CRAMERS_V | 0.535268642 | 0.000099990 | STRONG |
| PANEL_SERIES | CRAMERS_V | 0.575397092 | 0.000099990 | STRONG |
| WILDCARD_PRESENT | CRAMERS_V | 0.087313359 | 0.366663334 | WEAK |
| EXACT_LEXICAL_ANCHOR | CRAMERS_V | 0.985178223 | 0.000099990 | STRONG |
| COORDINATE_BRIDGE_AVAILABLE | CRAMERS_V | 0.455676104 | 0.000099990 | STRONG |
| VALIDATED_IVTFF_LOCUS_TYPE | CRAMERS_V | 0.463628844 | 0.000099990 | STRONG |
| LABEL_LENGTH_GLYPHS | STANDARDIZED_MEAN_DIFFERENCE | -0.509262397 | 0.004199580 | STRONG |
| LABEL_LENGTH_TOKENS | STANDARDIZED_MEAN_DIFFERENCE | -0.295732450 | 0.108689131 | WEAK |
| TRANSCRIBER_VARIANTS | STANDARDIZED_MEAN_DIFFERENCE | 0.740173080 | 0.000099990 | STRONG |


The detailed panel/family table shows the substantive imbalance. `f68v2` has 6/28 mapped coordinates; `SECTOR_LABEL` is 0/15 and `OUTER_TITLE` is 0/1. In contrast, planet/moon is 7/7, circle-sector is 10/12, and star labels are 53/67 overall. Thus unmatched coordinates are concentrated in specific panels and families, not missing uniformly.

Matched coordinates also have shorter Stolfi readings by known-glyph length and more transcriber variants on average. Wildcard presence is weakly associated with matching, while exact lexical anchoring is nearly deterministic because it is a core matching requirement. These are properties of selection into the confirmed set, not evidence about hapax status by themselves.

## Sensitivity bounds

Unmatched coordinates are not inserted into the corpus. For bounds only, each of the 54 coordinates contributes the maximum Stolfi token count seen among its transcriber variants, for 77 potential token occurrences. This is an accounting device, not a reconstructed label corpus.

| scenario | added potential occurrences | resulting/expected hapax fraction | conditioned background | difference | direction |
|---|---:|---:|---:|---:|---|
| pessimistic: all non-hapax | 77 | 0.423280423 | 0.603673298 | -0.180392875 | NON_POSITIVE |
| optimistic: all hapax | 77 | 0.830687831 | 0.603673298 | 0.227014533 | POSITIVE |
| panel/family background imputation | 77 | 0.665047248 | 0.603673298 | 0.061373949 | POSITIVE |

The conditioned imputation uses only each panel's observed 901-token background hapax rate, applied within the actual unmatched panel/family cells. It does not use the desired label result. It preserves the positive direction but necessarily shrinks it. The pessimistic bound is not a realistic estimate; it establishes that the missing data are numerous enough, in principle, to reverse the result.

## Confirmed-label subset checks

| subset | labels | observed | null mean | ratio | difference | p (upper) |
|---|---:|---:|---:|---:|---:|---:|
| high-coverage series | 68 | 0.735294118 | 0.631241176 | 1.164838647 | 0.104052941 | 0.026297370 |
| without f68v2 | 99 | 0.727272727 | 0.614942424 | 1.182668000 | 0.112330303 | 0.006699330 |
| star family | 54 | 0.851851852 | 0.690924074 | 1.232916733 | 0.160927778 | 0.001899810 |

The high-coverage definition is frozen mechanically at `panel.group >= 80%` physical-coordinate coverage; qualifying series are `f67r1.S, f67r2.L, f68r1.S, f68r2.S`. These checks show that f68v2 alone is not responsible for the positive result, but they cannot recover outcomes for excluded or unmapped labels.

## Scope

This audit asks only whether mapping incompleteness can explain the previously detected enrichment. It neither invalidates the positive-label test nor upgrades unmatched records to labels or non-labels. No image interpretation, semantic matching, or desired-result imputation is used.

## Final status

```text
MATCHING_SELECTION_BIAS=COULD_EXPLAIN_EFFECT
MATCHED_RECORDS=130
UNMATCHED_RECORDS=61
HIGH_COVERAGE_SUBSET_EFFECT=ratio=1.164838647;difference=0.104052941;p=0.026297370
WITHOUT_F68V2_EFFECT=ratio=1.182668000;difference=0.112330303;p=0.006699330
STAR_SUBSET_EFFECT=ratio=1.232916733;difference=0.160927778;p=0.001899810
WORST_CASE_EFFECT=ratio=0.701174666;difference=-0.180392875;direction=NON_POSITIVE
```
