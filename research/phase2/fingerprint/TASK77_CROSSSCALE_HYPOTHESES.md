# Task77 preregistered cross-scale hypothesis matrix

Status: **preregistered before the canonical run; implemented**. task77's
own methodological ban is the reason this document exists: "нельзя выводить
cross-scale dependency из двух отдельно значимых результатов" — two
separately-significant marginal results never establish `X ⊥̸ Y`. Every row
below is measured directly (joint or conditional statistic against a
matched null), not inferred from two marginals.

This CS1-CS8 numbering is **task77's own**, distinct from task73's
CS1-CS10 table in `FINGERPRINT_V2_SPEC.md` (which is glyph/repetition-
centric). Where the two overlap in substance, this is noted explicitly.
`metrics.ef5` in the JSON output is task73's EF5/CS4 (family × folio/
regime concentration, the same computation as `LP3.locality`, extended
with a C-REGIME comparison); it is referenced, not duplicated, by CS4
below.

| ID | Pair | Confirmatory statistic | Null | Confounders controlled | Status field |
|---|---|---|---|---|---|
| CS1 | edit family (and family role) × line position | `normalizedMI(position_class, family_label)` and `normalizedMI(position_class, role_label)`, restricted to family-bearing occurrences in lines of length ≥2 | N2 (within-line shuffle) | line length, token frequency (both held exactly fixed by the within-line, family-population-restricted shuffle) | `cross_scale.metrics[metric_id="cs1/family-line-position"]` |
| CS2 | transformation (family member / zone) × local context | `normalizedMI(prev_token_family, current_token_family)` and `normalizedMI(prev_token_family, current_token_zone_profile)` over adjacent-occurrence pairs | N2 (within-line sequence shuffle, i.e. reorder each line's own tokens) | which lines exist and their vocabulary (both fixed exactly; only intra-line order is randomized) | `cs2/prev-family-current-family`, `cs2/prev-family-current-zone` |
| CS3 | edit family × locus type | `normalizedMI(family_label, locus_class)`, locus_class ∈ {TEXT, LABEL, SPECIAL} | N4 (within-folio shuffle) | per-folio locus-type composition, per-folio family composition | `cs3/family-locus-type` |
| CS4 | edit family × folio/section/Currier regime | (a) EF5 same-folio/same-regime concentration (shared computation, `metrics.ef5`); (b) `normalizedMI(family_label, currier)`; (c) `normalizedMI(family_label, section)`, both restricted to family-bearing occurrences | (a) C-GLOBAL/C-REGIME; (b)/(c) folio-level metadata-label permutation | folio count and size per regime/section (permutation is at the folio level, matching the actual unit at which Currier/section are assigned) | `cs4/family-currier`, `cs4/family-section` (+ `metrics.ef5`) |
| CS5 | local context × larger regime (interaction) | range (max−min) of the CS2 same-family-adjacency rate across Currier×section strata | folio-level regime-label permutation (same construction as CS4b/c) | per-stratum sample size (strata with <5 pairs excluded) | `cs5/local-adjacency-x-regime` |
| CS6 | family composition × line/folio structure | \|Spearman(line length, per-line family-composition entropy)\| | N1 realized as a global token shuffle that leaves line-length boundaries untouched (positions, not identities, define a line) | corpus-wide vocabulary/frequency distribution (exactly preserved: only token *identities* are permuted) | `cs6/family-diversity-x-line-length` |
| CS7 | edit distance × structural distance | frequency-bin-weighted \|Spearman(raw glyph edit distance, minimum line distance between occurrences)\| over sampled vocabulary pairs | N6 (structural-distance permutation within the same combined-frequency bin) | combined endpoint frequency (via log2-frequency-bin stratification, matching the same convention as LP2's C-FREQ) | `cs7/edit-distance-x-structural-distance` |
| CS8 | conditional persistence of CS1 | CS1's family test recomputed separately within Currier A, Currier B, and TEXT-locus strata, each against its **own** within-line null | N2, applied independently per stratum | folio/regime composition that could confound a pooled estimate (this is exactly the `TestCS1ConfoundedByRegime` synthetic case) | `cs1/family-line-position.partition_stability` (stratum rows tagged `cs1_conditioning`) |

## Preregistered thresholds (fixed before the canonical run)

- Significance: Benjamini-Hochberg FDR at `alpha` (config default 0.05)
  over the declared cross-scale test family (all confirmatory CS1-CS7
  tests reported in one `fdr()` call — see `internal/fingerprintv2/
  crossscale_pipeline.go`).
- Minimum population floor: 20 qualifying occurrences/pairs per test,
  below which the metric is reported `INCONCLUSIVE`, not silently omitted
  or force-labeled `NOT_SUPPORTED`.
- Stability classification thresholds (§9 of the task, used for both the
  edit-family stability battery and CS8): `ARI >= 0.5` → `GLOBAL`,
  `0.2 <= ARI < 0.5` → `PARTITION_SPECIFIC`, `ARI < 0.2` → `UNSTABLE`,
  `<10` comparable units → `INSUFFICIENT_DATA`. These bands were chosen
  before the canonical run and are not tuned to its results.

## What is confirmatory vs. exploratory

Confirmatory (this table, fixed in advance): CS1-CS8 as specified above,
EF5, the edit-family stability battery (§2.3 of the task), and the
held-out validation attached to CS1 (family → line position) via grouped-
folio log loss.

Exploratory (reported separately in `cross_scale.exploratory_findings`,
never folded into a `SUPPORTED` verdict): connected-components-vs-label-
propagation disagreement, hub-removal sensitivity read as a *specific-
hub-identity* claim (rather than the preregistered aggregate giant-share
statistic), and any locus-type-code-level pattern inside the pooled
SPECIAL bucket.

## Deliberately out of scope for CS1-CS8 (documented, not silently dropped)

- Occurrence-level "direction of transformation" (insertion vs. deletion
  specifically, rather than which zone/family) is not tested per
  occurrence; CS2's zone-profile test is the closest confirmatory proxy.
  Flagged as a named limitation on every CS2 metric record.
- CS7 samples vocabulary pairs (`cross_scale.folds`-independent;
  `structural_sample` in config, default 2000) rather than enumerating
  all ~34M pairs; the sample size and seed are recorded in provenance so
  the sample is reproducible, and the cap is disclosed on the metric
  record, not hidden.
