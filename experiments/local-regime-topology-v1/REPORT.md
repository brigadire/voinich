# Task65 report: local regime topology, drift and change-point analysis

LOCAL_STRUCTURE: **CONFIRMED**
TOPOLOGY: **MIXED_DRIFT_AND_STATES**
METADATA: **PARTIALLY_METADATA_EXPLAINED**
TASK64_SPLIT: **TRUE_REGIME_HETEROGENEITY**

R is called a *regime* or *local distributional state* throughout, never a
topic, language state or cipher key (task65's own safeguard, echoing task64
section 65). No decipherment, language or semantic claim is made. A discrete
or hierarchical finding does not rule out the scribal/layout alternative
(line filling, glyph-choice drift, copying batches, page planning) - it is
discussed explicitly below rather than assumed away (section 51).

## 1-2. Does local similarity depend on distance, and does it decay smoothly?

LOCAL_REGIME_DECAY.tsv reports mean/median/CI/null-mean profile distance for every token/line/page lag, for the primary W=20 window and the W=10/40/80 sensitivity windows, against a shuffle-order null - LOCAL_STRUCTURE=CONFIRMED answers question 1 directly (whether the lag=1 curve sits below its null). Section 14's instruction not to assume exponential decay is honored: no functional form is fit here, only the empirical curve.

## 3. Are there statistically supported abrupt change points?

CHANGE_POINTS.tsv reports every non-overlapping-window boundary's score, significance against the STATIONARY synthetic null's P95/P99 (CHANGE_POINT_NULLS.tsv), discovery/replication fold and nearest metadata boundary. CHANGE_POINT_METADATA_OVERLAP.tsv gives the enrichment of significant change points near real Currier/Hand/Section/page boundaries over a position-permutation null.

## 4. Do similar regimes recur in distant parts of the manuscript?

DISTANT_REGIME_RECURRENCE.tsv compares, for a sample of windows, the nearest profile distance among windows >=5 pages away against the same nearest-distance search repeated after globally shuffling which profile sits at which position (both are a minimum over the same-size candidate pool, so the comparison isn't biased by the fact that a minimum over many candidates is always smaller than a single random draw) - this is independent of the discrete-clustering result and answers the recurrence question even if clustering itself is unstable (section 46). Observed mean nearest-distant distance is 0.046108 against a properly-calibrated null of 0.002320: the real manuscript does not find closer distant matches than the shuffled-position null once the minimum-of-many-candidates bias is corrected for, so this analysis does NOT support recurring discrete regimes beyond chance on its own.

## 5-6. How much does Currier/Hand/Section/Page explain, and does a residual local regime remain after conditioning?

METADATA_EFFECTS.tsv (between vs within-group page-profile distance) and HIERARCHICAL_VARIANCE.tsv (variance share per factor on MeanTokenLength/GiantFraction) give METADATA=PARTIALLY_METADATA_EXPLAINED. METADATA_CONDITIONED_DECAY.tsv recomputes the token-lag decay restricted to windows sharing the same Currier/Hand/Section; the SAME_PAGE(LINE) rows in that same file are Task64's own within-page line-lag decay, i.e. the strict same-page conditioning of section 31. Its lag=1 excess similarity is 0.060906200: decay survives strictly within one page, which cannot be explained by Currier/Hand/Section/page-level composition since page is held fixed (task65 section 31, critical result B).

## 7. What explains Task64's discovery (~0.030) vs replication (~0.003) gap?

TASK64_SPLIT_DIAGNOSIS.tsv first reproduces Task64's own numbers independently: discovery=0.030069436, replication=0.003132198 (compare to Task64's published 0.030069436/0.003132198). It then reports each fold's Currier/Hand/Section/page composition and a standardized per-stratum comparison; TASK64_SPLIT=TRUE_REGIME_HETEROGENEITY. LOCAL_EFFECT_BY_REGION.tsv gives the region-wise map (Currier/Hand/Section, N/effect/CI/reliability) requested by section 70, so if the effect concentrates in one region that is visible directly rather than averaged away.

## Discrete clustering (H2) and mixed drift+states (H4)

A stable K=5 clustering was selected on held-out (VALIDATION-fold) distance-to-medoid among bootstrap-stable K values (REGIME_CLUSTER_SELECTION.tsv/REGIME_CLUSTER_STABILITY.tsv), fit on DISCOVERY only and only compared to metadata afterward (no label leakage). REGIME_TRANSITIONS.tsv/REGIME_DWELL.tsv compare transition/dwell structure to a label-shuffled null on non-overlapping windows (avoiding the overlapping-window pseudoreplication warned about in section 7); WITHIN_REGIME_DRIFT.tsv tests whether decay still exists inside each cluster (section 43/H4). K summary: K=2 stability=1.000 validationWithin=2.541
K=3 stability=0.627 validationWithin=2.449
K=4 stability=0.550 validationWithin=2.242
K=5 stability=0.506 validationWithin=1.979
K=6 stability=0.433 validationWithin=2.015
K=7 stability=0.451 validationWithin=1.885
K=8 stability=0.435 validationWithin=1.820


## Controls

SYNTHETIC_CONTROL_TOPOLOGY.tsv validates the pipeline itself: STATIONARY should show near-zero correlation length and a change-point count near the calibration control's own count (it IS that control); SMOOTH_DRIFT should show a real correlation length without an inflated change-point count (section 21); DISCRETE and MIXED report boundary-recovery rate against their known, injected boundaries (sections 20/22). NATURAL_CONTROL_TOPOLOGY.tsv applies the identical pipeline to Doyle/Longfellow/Astafiev (section 48, not to classify genre but to see whether drift/regime topology is a generic property of any finite natural-language corpus). TASK62_STATIONARY_CONTROL.tsv applies it to Task62's frozen G-only generation, which should read as close to the STATIONARY control (section 49).

## Scope

This is a diagnostic study only: no G+R or other generative model was built (section 2 - that is explicitly left to a future task once this topology is settled). Stages 1-28 were not touched and no Stage29 was added. No claim is made here about language, semantics, a specific cipher, or decipherment.
