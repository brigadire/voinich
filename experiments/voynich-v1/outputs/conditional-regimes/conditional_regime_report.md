# Conditional residual distributional structure

Main question: after conditioning on Currier and Davis hand, does reproducible distributional structure remain in the corpus?

## Reproducibility

- corpus: `data_work/ZL3b-x7.txt`, SHA256 `360d99583145ec549b80edfafdc3f93534f3a11b85a0d52997ba8425e92b87c2`, 39026 tokens
- metadata map: `workdir/metadata-validation/token_metadata_map.tsv`, SHA256 `148745adbc889150ad1b59715bbfa75fa17e24b566694d94a0445d06393a7e68`
- excluded (unknown-metadata) tokens: 14453
- window sizes (within-class): [50 100 200 500]; residual window sizes: [50 100 200 500 1000]
- min class tokens: 1000; min block tokens: 500
- K: 2..10 (within-class), 2..15 (residual)
- permutations: 1000 primary, 10000 refinement (top 5 qualifying candidates, empirical_p<0.01 and effect_size>=2.0 at the primary pass)
- seed: 1

## Eligible joint Currier x hand classes

| Class | Total tokens | Blocks | Largest block | Eligible |
|---|---:|---:|---:|---|
| 1/1 | 7423 | 10 | 4145 | true |
| 2/2 | 9318 | 9 | 6957 | true |
| 2/5 | 467 | 3 | 217 | false |
| 3/2 | 1855 | 2 | 1352 | true |
| 3/4 | 540 | 1 | 540 | false |
| 4/1 | 870 | 3 | 489 | false |
| 5/3 | 535 | 1 | 535 | false |
| X/3 | 2787 | 2 | 1911 | true |
| Y/3 | 778 | 1 | 778 | false |

## Part A — within-class discovery

Per-class, per-window-size, per-method significance against Null A (within-block token shuffle), at each method's best-observed K (`within_class_permutations.yaml` has every candidate; `conditional_class_inventory.tsv`/`within_class_regimes.tsv`/`within_class_stability.tsv` have the full diagnostics):

| Class | Scheme | Window | Method | K | Silhouette | Effect size | Empirical p | Refined |
|---|---|---:|---|---:|---:|---:|---:|---|
| 3 | currier_only | 50 | k_medoids | 10 | 0.091 | 9.41 | <0.0001 | true |
| 4 | hand_only | 100 | hierarchical | 4 | 0.114 | 9.07 | <0.0001 | true |
| 4 | hand_only | 100 | k_medoids | 9 | 0.177 | 7.98 | <0.0001 | true |
| 2/2 | joint | 200 | hierarchical | 7 | 0.209 | 7.80 | 0.0006 | true |
| 3 | hand_only | 200 | k_medoids | 8 | 0.120 | 7.60 | <0.0001 | true |
| 3/2 | joint | 50 | k_medoids | 10 | 0.143 | 7.24 | 0.0010 | false |
| 3 | currier_only | 100 | k_medoids | 10 | 0.266 | 7.04 | 0.0010 | false |
| X/3 | joint | 100 | hierarchical | 8 | 0.192 | 6.38 | 0.0010 | false |
| X | currier_only | 100 | hierarchical | 8 | 0.192 | 6.38 | 0.0010 | false |
| X/3 | joint | 100 | k_medoids | 10 | 0.233 | 6.38 | 0.0010 | false |
| X | currier_only | 100 | k_medoids | 10 | 0.233 | 6.38 | 0.0010 | false |
| 4 | hand_only | 200 | k_medoids | 8 | 0.351 | 6.10 | 0.0010 | false |
| 2 | currier_only | 200 | hierarchical | 9 | 0.212 | 5.78 | 0.0040 | false |
| X/3 | joint | 200 | k_medoids | 6 | 0.373 | 5.74 | 0.0010 | false |
| X | currier_only | 200 | k_medoids | 6 | 0.373 | 5.74 | 0.0010 | false |
| 3 | hand_only | 500 | k_medoids | 10 | 0.306 | 5.58 | 0.0010 | false |
| 1/1 | joint | 200 | k_medoids | 10 | 0.205 | 5.58 | 0.0010 | false |
| 1 | currier_only | 200 | k_medoids | 10 | 0.205 | 5.58 | 0.0010 | false |
| 3 | currier_only | 100 | hierarchical | 10 | 0.244 | 4.88 | 0.0010 | false |
| X/3 | joint | 50 | k_medoids | 10 | 0.091 | 4.31 | 0.0010 | false |
| X | currier_only | 50 | k_medoids | 10 | 0.091 | 4.31 | 0.0010 | false |
| 3 | hand_only | 500 | hierarchical | 7 | 0.227 | 4.01 | 0.0010 | false |
| 1 | hand_only | 200 | k_medoids | 9 | 0.164 | 3.89 | 0.0010 | false |
| 3/2 | joint | 100 | k_medoids | 8 | 0.290 | 3.81 | 0.0010 | false |
| 3 | currier_only | 200 | k_medoids | 5 | 0.314 | 3.64 | 0.0010 | false |
| X/3 | joint | 200 | hierarchical | 6 | 0.272 | 3.05 | 0.0010 | false |
| X | currier_only | 200 | hierarchical | 6 | 0.272 | 3.05 | 0.0010 | false |
| 1 | hand_only | 200 | hierarchical | 10 | 0.125 | 2.89 | 0.0030 | false |
| 3 | currier_only | 200 | hierarchical | 4 | 0.270 | 2.37 | 0.0010 | false |
| 3 | currier_only | 50 | hierarchical | 10 | 0.096 | 2.14 | 0.0260 | false |
| 1/1 | joint | 500 | k_medoids | 4 | 0.285 | 2.14 | 0.0020 | false |
| 1 | currier_only | 500 | k_medoids | 4 | 0.285 | 2.14 | 0.0020 | false |
| 1/1 | joint | 100 | k_medoids | 2 | 0.086 | 1.52 | 0.1059 | false |
| 1 | currier_only | 100 | k_medoids | 2 | 0.086 | 1.52 | 0.1059 | false |
| 4 | hand_only | 50 | k_medoids | 9 | 0.055 | 1.37 | 0.1059 | false |
| 3/2 | joint | 50 | hierarchical | 10 | 0.099 | 1.15 | 0.0989 | false |
| 1 | hand_only | 500 | k_medoids | 6 | 0.322 | 1.04 | 0.0230 | false |
| 1 | hand_only | 100 | k_medoids | 4 | 0.079 | 1.01 | 0.1149 | false |
| 3/2 | joint | 200 | k_medoids | 3 | 0.291 | 0.74 | 0.2258 | false |
| 1 | hand_only | 50 | k_medoids | 3 | 0.032 | 0.61 | 0.2398 | false |
| 3 | hand_only | 100 | k_medoids | 2 | 0.070 | 0.55 | 0.4176 | false |
| 3/2 | joint | 200 | hierarchical | 3 | 0.291 | 0.49 | 0.4006 | false |
| 1 | hand_only | 50 | hierarchical | 2 | 0.096 | 0.44 | 0.3317 | false |
| 1 | hand_only | 500 | hierarchical | 4 | 0.268 | 0.42 | 0.3357 | false |
| 2/2 | joint | 50 | k_medoids | 3 | 0.067 | 0.29 | 0.4725 | false |
| 2 | currier_only | 50 | hierarchical | 2 | 0.124 | 0.28 | 0.3806 | false |
| 3 | hand_only | 50 | k_medoids | 2 | 0.019 | 0.25 | 0.1938 | false |
| 3 | hand_only | 50 | hierarchical | 2 | 0.073 | 0.23 | 0.4156 | false |
| 4 | hand_only | 200 | hierarchical | 6 | 0.194 | 0.17 | 0.4985 | false |
| 2 | hand_only | 50 | hierarchical | 2 | 0.105 | 0.12 | 0.4565 | false |
| 2 | currier_only | 50 | k_medoids | 4 | 0.049 | 0.07 | 0.3886 | false |
| 2 | hand_only | 50 | k_medoids | 3 | 0.044 | 0.06 | 0.4845 | false |
| X/3 | joint | 50 | hierarchical | 10 | 0.120 | 0.04 | 0.5015 | false |
| X | currier_only | 50 | hierarchical | 10 | 0.120 | 0.04 | 0.5015 | false |
| 2/2 | joint | 50 | hierarchical | 2 | 0.115 | -0.04 | 0.5065 | false |
| 2 | hand_only | 100 | k_medoids | 3 | 0.098 | -0.09 | 0.7892 | false |
| 3/2 | joint | 100 | hierarchical | 7 | 0.174 | -0.18 | 0.5764 | false |
| 1 | hand_only | 100 | hierarchical | 2 | 0.122 | -0.25 | 0.7013 | false |
| 2 | currier_only | 100 | k_medoids | 8 | 0.089 | -0.37 | 0.4336 | false |
| 1/1 | joint | 200 | hierarchical | 10 | 0.159 | -0.42 | 0.6673 | false |
| 1 | currier_only | 200 | hierarchical | 10 | 0.159 | -0.42 | 0.6673 | false |
| 2 | hand_only | 200 | k_medoids | 7 | 0.180 | -0.47 | 0.9171 | false |
| 2/2 | joint | 100 | k_medoids | 4 | 0.103 | -0.78 | 0.8901 | false |
| 3 | hand_only | 200 | hierarchical | 2 | 0.119 | -1.01 | 0.9471 | false |
| 2 | hand_only | 500 | hierarchical | 2 | 0.326 | -1.12 | 0.9750 | false |
| 3 | hand_only | 100 | hierarchical | 2 | 0.070 | -1.37 | 0.9111 | false |
| 1/1 | joint | 50 | k_medoids | 9 | 0.020 | -1.45 | 0.9301 | false |
| 1 | currier_only | 50 | k_medoids | 9 | 0.020 | -1.45 | 0.9301 | false |
| 1/1 | joint | 500 | hierarchical | 2 | 0.227 | -1.55 | 0.9431 | false |
| 1 | currier_only | 500 | hierarchical | 2 | 0.227 | -1.55 | 0.9431 | false |
| 2 | currier_only | 200 | k_medoids | 2 | 0.193 | -2.48 | 0.9930 | false |
| 2 | hand_only | 200 | hierarchical | 10 | 0.205 | -2.74 | 0.9810 | false |
| 2 | hand_only | 500 | k_medoids | 2 | 0.326 | -3.85 | 1.0000 | false |
| 1/1 | joint | 50 | hierarchical | 4 | 0.021 | -4.60 | 1.0000 | false |
| 1 | currier_only | 50 | hierarchical | 4 | 0.021 | -4.60 | 1.0000 | false |
| 4 | hand_only | 500 | hierarchical | 2 | 0.271 | -4.71 | 1.0000 | false |
| 4 | hand_only | 500 | k_medoids | 2 | 0.271 | -4.77 | 1.0000 | false |
| 2/2 | joint | 200 | k_medoids | 2 | 0.191 | -5.38 | 0.9980 | false |
| 1/1 | joint | 100 | hierarchical | 5 | 0.032 | -5.77 | 1.0000 | false |
| 1 | currier_only | 100 | hierarchical | 5 | 0.032 | -5.77 | 1.0000 | false |
| 2/2 | joint | 100 | hierarchical | 9 | 0.126 | -6.14 | 1.0000 | false |
| 4 | hand_only | 50 | hierarchical | 10 | 0.073 | -6.24 | 1.0000 | false |
| 2 | currier_only | 500 | k_medoids | 2 | 0.295 | -7.14 | 1.0000 | false |
| 2 | currier_only | 500 | hierarchical | 2 | 0.295 | -7.23 | 1.0000 | false |
| 2 | hand_only | 100 | hierarchical | 9 | 0.090 | -8.71 | 1.0000 | false |
| 2 | currier_only | 100 | hierarchical | 10 | 0.126 | -8.82 | 1.0000 | false |
| 2/2 | joint | 500 | hierarchical | 2 | 0.291 | -13.47 | 1.0000 | false |
| 2/2 | joint | 500 | k_medoids | 2 | 0.291 | -13.67 | 1.0000 | false |

## Part B — metadata-residualized feature space

Pooled k_medoids clustering over the raw residual R_w = X_w - mu_(C,H) (training-fold-only centering; no held-out leakage), across the frozen scale x K residual search space:

- hierarchical|raw: global max silhouette observed 0.333, null mean 0.408, P95 0.417, P99 0.422, effect size -12.16, empirical p 1.0000
- k_medoids|raw: global max silhouette observed 0.330, null mean 0.407, P95 0.417, P99 0.421, effect size -12.27, empirical p 1.0000

Winning combination (k_medoids, raw): window_size=500, K=2.

Metadata independence check (residual vs original global max NMI, task18's frozen `cluster_metadata_global_summary.tsv`):

| Metadata | Original NMI | Residual NMI | Residual ARI | Information reduction |
|---|---:|---:|---:|---:|
| currier | 0.713 | 0.719 | 0.690 | 0.00 |
| hand | 0.631 | 0.574 | 0.555 | 0.09 |

Cross-metadata/cross-block recurrence and the descriptive composite ranking are in `residual_regime_candidates.tsv`; the composite score is an unweighted sum of coverage fractions and is not an inferential statistic.

## Part C — conditional boundaries and residual transitions

404 change points found within controlled physical blocks (never crossing a metadata transition); 222 recurring boundary types (grouped by dominant-token signature rather than absolute position, since absolute positions of different blocks are not comparable). Full detail in `conditional_stable_boundaries.tsv`.

0 of 4 residual R_i -> R_j transition cells are enriched relative to the Null-B within-block window-order shuffle at p<0.05 (`residual_transition_matrix.tsv`).

## Interpretation

**Outcome B — weak/inconclusive residual structure.** Some within-class or residual clusters reached significance, but recurrence was limited to one block, cross-metadata recurrence was weak, or significance was marginal. Evidence for additional organization beyond Currier/hand is weak or inconclusive.

This outcome describes only whether additional reproducible distributional structure exists beyond tested Currier/hand metadata. It does not identify instructions, operators, operands, grammar, recipes, a cipher, or natural language, even under Outcome C.

**Main limitation.** Currier and Davis hand are metadata models, not the complete cause of distributional heterogeneity. "Residual after Currier x hand" means unexplained by these annotations, not independent of scribal or material effects in general.
