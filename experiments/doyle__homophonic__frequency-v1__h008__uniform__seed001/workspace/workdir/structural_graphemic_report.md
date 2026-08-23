# Structural–graphemic analysis

This analysis compares two independent coordinates. The existing structural similarity was copied unchanged from `raw_similarity`; spelling was used only for the graphemic metrics.

## Scope

- Tokens: 667
- Pairs: 222111
- Pearson correlation (graphemic similarity vs structural similarity): 0.107017
- Spearman correlation: 0.080195

Pair observations share tokens and are not independent; the correlations are descriptive and ordinary pairwise p-values are intentionally not reported.

## Structural similarity by normalized grapheme-distance bin

| Distance | Pairs | Mean | Median | P90 | P95 |
|---|---:|---:|---:|---:|---:|
| 0.0–0.1 | 0 | 0.0000 | 0.0000 | 0.0000 | 0.0000 |
| 0.1–0.2 | 6967 | 0.2693 | 0.2606 | 0.3671 | 0.4322 |
| 0.2–0.3 | 60323 | 0.2505 | 0.2481 | 0.3294 | 0.3548 |
| 0.3–0.4 | 0 | 0.0000 | 0.0000 | 0.0000 | 0.0000 |
| 0.4–0.5 | 152162 | 0.2425 | 0.2429 | 0.3102 | 0.3308 |
| 0.5–0.6 | 2659 | 0.2014 | 0.2068 | 0.2603 | 0.2740 |
| 0.6–0.7 | 0 | 0.0000 | 0.0000 | 0.0000 | 0.0000 |
| 0.7–0.8 | 0 | 0.0000 | 0.0000 | 0.0000 | 0.0000 |
| 0.8–0.9 | 0 | 0.0000 | 0.0000 | 0.0000 | 0.0000 |
| 0.9–1.0 | 0 | 0.0000 | 0.0000 | 0.0000 | 0.0000 |

## Selection and percentile view

Configurable distant selection uses structural similarity ≥ 0.650, reliability ≥ 0.700, and normalized grapheme distance ≥ 0.600. It yields 0 pairs. Ranking is `structural_similarity × reliability × normalized_grapheme_distance`; this score is neither a probability nor statistical significance.

As a threshold-free companion view, corpus cutoffs are structural P95 = 0.3405, reliability P75 = 0.6160, and grapheme-distance P75 = 0.4286. Their intersection yields 4368 pairs, ranked by the same transparent score. These percentiles describe ranking coordinates and do not define the class.

## Structurally close / graphically distant

| A | B | Counts | Structural | Reliability | Norm. distance | Discovery score |
|---|---|---:|---:|---:|---:|---:|

## Percentile-ranked distant view

| A | B | Counts | Structural | Reliability | Norm. distance | Discovery score |
|---|---|---:|---:|---:|---:|---:|
| x000025 | x000100 | 159/59 | 0.5067 | 0.8299 | 0.4286 | 0.1802 |
| x000052 | x000100 | 107/59 | 0.5206 | 0.7912 | 0.4286 | 0.1765 |
| x000026 | x000100 | 177/59 | 0.4890 | 0.8304 | 0.4286 | 0.1740 |
| x000028 | x000100 | 158/59 | 0.4886 | 0.8293 | 0.4286 | 0.1737 |
| x000022 | x000100 | 167/59 | 0.4862 | 0.8304 | 0.4286 | 0.1730 |
| x000056 | x000100 | 110/59 | 0.5082 | 0.7939 | 0.4286 | 0.1729 |
| x000051 | x000100 | 99/59 | 0.5121 | 0.7832 | 0.4286 | 0.1719 |
| x000054 | x000100 | 98/59 | 0.5108 | 0.7822 | 0.4286 | 0.1712 |
| x000056 | x000124 | 110/58 | 0.5026 | 0.7920 | 0.4286 | 0.1706 |
| x000027 | x000100 | 159/59 | 0.4793 | 0.8299 | 0.4286 | 0.1705 |
| x000053 | x000122 | 81/55 | 0.5246 | 0.7546 | 0.4286 | 0.1696 |
| x000051 | x000124 | 99/58 | 0.5060 | 0.7813 | 0.4286 | 0.1694 |
| x000043 | x000182 | 106/34 | 0.5431 | 0.7269 | 0.4286 | 0.1692 |
| x000051 | x000122 | 99/55 | 0.5061 | 0.7754 | 0.4286 | 0.1682 |
| x000054 | x000122 | 98/55 | 0.5067 | 0.7743 | 0.4286 | 0.1681 |
| x000023 | x000100 | 149/59 | 0.4700 | 0.8237 | 0.4286 | 0.1659 |
| x000050 | x000122 | 90/55 | 0.5055 | 0.7656 | 0.4286 | 0.1658 |
| x000053 | x000121 | 81/54 | 0.5114 | 0.7526 | 0.4286 | 0.1649 |
| x000014 | x000130 | 166/61 | 0.4590 | 0.8345 | 0.4286 | 0.1642 |
| x000055 | x000124 | 94/58 | 0.4930 | 0.7760 | 0.4286 | 0.1639 |

## Structurally close / graphically close (control)

| A | B | Counts | Structural | Reliability | Norm. distance | Discovery score |
|---|---|---:|---:|---:|---:|---:|

## Frequency control

| Minimum count for both tokens | Pairs | Pearson | Spearman | Distant candidates |
|---:|---:|---:|---:|---:|
| 10 | 222111 | 0.10702 | 0.08019 | 0 |
| 20 | 55611 | 0.15228 | 0.11457 | 0 |
| 50 | 7503 | 0.24576 | 0.23318 | 0 |
| 100 | 1225 | 0.29767 | 0.22776 | 0 |

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
