# Structural–graphemic analysis

This analysis compares two independent coordinates. The existing structural similarity was copied unchanged from `raw_similarity`; spelling was used only for the graphemic metrics.

## Scope

- Tokens: 667
- Pairs: 222111
- Pearson correlation (graphemic similarity vs structural similarity): 0.057067
- Spearman correlation: 0.026661

Pair observations share tokens and are not independent; the correlations are descriptive and ordinary pairwise p-values are intentionally not reported.

## Structural similarity by normalized grapheme-distance bin

| Distance | Pairs | Mean | Median | P90 | P95 |
|---|---:|---:|---:|---:|---:|
| 0.0–0.1 | 0 | 0.0000 | 0.0000 | 0.0000 | 0.0000 |
| 0.1–0.2 | 1437 | 0.3378 | 0.3175 | 0.5150 | 0.5648 |
| 0.2–0.3 | 5509 | 0.2559 | 0.2526 | 0.3302 | 0.3695 |
| 0.3–0.4 | 0 | 0.0000 | 0.0000 | 0.0000 | 0.0000 |
| 0.4–0.5 | 32621 | 0.2480 | 0.2466 | 0.3199 | 0.3443 |
| 0.5–0.6 | 93523 | 0.2463 | 0.2452 | 0.3197 | 0.3440 |
| 0.6–0.7 | 0 | 0.0000 | 0.0000 | 0.0000 | 0.0000 |
| 0.7–0.8 | 89021 | 0.2457 | 0.2449 | 0.3189 | 0.3453 |
| 0.8–0.9 | 0 | 0.0000 | 0.0000 | 0.0000 | 0.0000 |
| 0.9–1.0 | 0 | 0.0000 | 0.0000 | 0.0000 | 0.0000 |

## Selection and percentile view

Configurable distant selection uses structural similarity ≥ 0.650, reliability ≥ 0.700, and normalized grapheme distance ≥ 0.600. It yields 0 pairs. Ranking is `structural_similarity × reliability × normalized_grapheme_distance`; this score is neither a probability nor statistical significance.

As a threshold-free companion view, corpus cutoffs are structural P95 = 0.3468, reliability P75 = 0.6126, and grapheme-distance P75 = 0.7143. Their intersection yields 3376 pairs, ranked by the same transparent score. These percentiles describe ranking coordinates and do not define the class.

## Structurally close / graphically distant

| A | B | Counts | Structural | Reliability | Norm. distance | Discovery score |
|---|---|---:|---:|---:|---:|---:|

## Percentile-ranked distant view

| A | B | Counts | Structural | Reliability | Norm. distance | Discovery score |
|---|---|---:|---:|---:|---:|---:|
| x014952 | x030834 | 88/90 | 0.5888 | 0.8091 | 0.7143 | 0.3403 |
| x014956 | x030835 | 87/90 | 0.5863 | 0.8078 | 0.7143 | 0.3383 |
| x000078 | x028394 | 176/402 | 0.5053 | 0.9371 | 0.7143 | 0.3382 |
| x000078 | x028395 | 176/382 | 0.5047 | 0.9371 | 0.7143 | 0.3378 |
| x000078 | x028397 | 176/394 | 0.5042 | 0.9371 | 0.7143 | 0.3375 |
| x014956 | x030837 | 87/96 | 0.5769 | 0.8148 | 0.7143 | 0.3357 |
| x014157 | x028906 | 103/176 | 0.5295 | 0.8871 | 0.7143 | 0.3355 |
| x014956 | x030839 | 87/108 | 0.5675 | 0.8272 | 0.7143 | 0.3353 |
| x000078 | x028396 | 176/375 | 0.5007 | 0.9371 | 0.7143 | 0.3352 |
| x000081 | x028394 | 194/402 | 0.4991 | 0.9371 | 0.7143 | 0.3341 |
| x000081 | x028393 | 194/409 | 0.4960 | 0.9371 | 0.7143 | 0.3320 |
| x014952 | x030839 | 88/108 | 0.5605 | 0.8285 | 0.7143 | 0.3317 |
| x014952 | x030837 | 88/96 | 0.5679 | 0.8160 | 0.7143 | 0.3310 |
| x014154 | x028905 | 108/203 | 0.5176 | 0.8927 | 0.7143 | 0.3300 |
| x000081 | x028397 | 194/394 | 0.4930 | 0.9371 | 0.7143 | 0.3300 |
| x014957 | x030838 | 70/99 | 0.5802 | 0.7950 | 0.7143 | 0.3295 |
| x014952 | x030835 | 88/90 | 0.5693 | 0.8091 | 0.7143 | 0.3290 |
| x014154 | x028906 | 108/176 | 0.5152 | 0.8927 | 0.7143 | 0.3285 |
| x000078 | x028392 | 176/363 | 0.4885 | 0.9371 | 0.7143 | 0.3269 |
| x014157 | x028905 | 103/203 | 0.5150 | 0.8871 | 0.7143 | 0.3264 |

## Structurally close / graphically close (control)

| A | B | Counts | Structural | Reliability | Norm. distance | Discovery score |
|---|---|---:|---:|---:|---:|---:|

## Frequency control

| Minimum count for both tokens | Pairs | Pearson | Spearman | Distant candidates |
|---:|---:|---:|---:|---:|
| 10 | 222111 | 0.05707 | 0.02666 | 0 |
| 20 | 57970 | 0.11457 | 0.04531 | 0 |
| 50 | 8001 | 0.26862 | 0.05695 | 0 |
| 100 | 1711 | 0.40501 | 0.10783 | 0 |

## Families

The same explicit edge criteria form connected components. There are 0 graphemic-structural families and 0 structural-distant families. Components can contain token pairs that are connected only through intermediate tokens; inspect the saved edge list before interpreting a whole component. The neutral term “family” denotes a graph component only.

### Graphemic-structural families

No components met the configured criteria.

### Structural-distant families

No components met the configured criteria.

## Graphemic-distance distribution

The bin counts above are the full empirical distribution of normalized edit distance. Edit operations are performed on grapheme sequences: `@NNN;` is one grapheme, `?` is one unknown grapheme, and no signs are deleted or normalized.

## Limitations

- Pair rows are dependent because each token occurs in many pairs.
- Reliability and frequency reduce, but cannot eliminate, instability of sparse profiles.
- Levenshtein distance assigns equal cost to every insertion, deletion, and substitution and contains no palaeographic model.
- Connected components are threshold-sensitive descriptive groups, not linguistic categories.
- Correlation does not establish that one coordinate causes the other.
- The analysis makes no claim about language, morphology, commands, operators, or cipher mechanisms.
