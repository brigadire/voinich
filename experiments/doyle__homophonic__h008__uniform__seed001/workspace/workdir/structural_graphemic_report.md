# Structural–graphemic analysis

This analysis compares two independent coordinates. The existing structural similarity was copied unchanged from `raw_similarity`; spelling was used only for the graphemic metrics.

## Scope

- Tokens: 664
- Pairs: 220116
- Pearson correlation (graphemic similarity vs structural similarity): 0.045556
- Spearman correlation: 0.013701

Pair observations share tokens and are not independent; the correlations are descriptive and ordinary pairwise p-values are intentionally not reported.

## Structural similarity by normalized grapheme-distance bin

| Distance | Pairs | Mean | Median | P90 | P95 |
|---|---:|---:|---:|---:|---:|
| 0.0–0.1 | 0 | 0.0000 | 0.0000 | 0.0000 | 0.0000 |
| 0.1–0.2 | 1631 | 0.3095 | 0.2953 | 0.4473 | 0.4919 |
| 0.2–0.3 | 4846 | 0.2529 | 0.2469 | 0.3340 | 0.3780 |
| 0.3–0.4 | 0 | 0.0000 | 0.0000 | 0.0000 | 0.0000 |
| 0.4–0.5 | 25962 | 0.2417 | 0.2417 | 0.3122 | 0.3341 |
| 0.5–0.6 | 84529 | 0.2403 | 0.2403 | 0.3115 | 0.3325 |
| 0.6–0.7 | 0 | 0.0000 | 0.0000 | 0.0000 | 0.0000 |
| 0.7–0.8 | 103148 | 0.2414 | 0.2411 | 0.3119 | 0.3337 |
| 0.8–0.9 | 0 | 0.0000 | 0.0000 | 0.0000 | 0.0000 |
| 0.9–1.0 | 0 | 0.0000 | 0.0000 | 0.0000 | 0.0000 |

## Selection and percentile view

Configurable distant selection uses structural similarity ≥ 0.650, reliability ≥ 0.700, and normalized grapheme distance ≥ 0.600. It yields 0 pairs. Ranking is `structural_similarity × reliability × normalized_grapheme_distance`; this score is neither a probability nor statistical significance.

As a threshold-free companion view, corpus cutoffs are structural P95 = 0.3355, reliability P75 = 0.5911, and grapheme-distance P75 = 0.7143. Their intersection yields 4197 pairs, ranked by the same transparent score. These percentiles describe ranking coordinates and do not define the class.

## Structurally close / graphically distant

| A | B | Counts | Structural | Reliability | Norm. distance | Discovery score |
|---|---|---:|---:|---:|---:|---:|

## Percentile-ranked distant view

| A | B | Counts | Structural | Reliability | Norm. distance | Discovery score |
|---|---|---:|---:|---:|---:|---:|
| x018874 | x025395 | 94/144 | 0.5376 | 0.8410 | 0.7143 | 0.3229 |
| x018874 | x025392 | 94/139 | 0.5337 | 0.8374 | 0.7143 | 0.3192 |
| x018874 | x025398 | 94/140 | 0.5280 | 0.8381 | 0.7143 | 0.3161 |
| x018872 | x025395 | 77/144 | 0.5327 | 0.8171 | 0.7143 | 0.3109 |
| x018873 | x025397 | 89/143 | 0.5215 | 0.8337 | 0.7143 | 0.3105 |
| x018873 | x025398 | 89/140 | 0.5221 | 0.8315 | 0.7143 | 0.3101 |
| x018874 | x025396 | 94/146 | 0.5152 | 0.8424 | 0.7143 | 0.3100 |
| x018876 | x025393 | 75/130 | 0.5372 | 0.8043 | 0.7143 | 0.3086 |
| x018876 | x025398 | 75/140 | 0.5301 | 0.8115 | 0.7143 | 0.3073 |
| x018878 | x025397 | 91/143 | 0.5136 | 0.8364 | 0.7143 | 0.3068 |
| x018873 | x025394 | 89/136 | 0.5161 | 0.8286 | 0.7143 | 0.3055 |
| x018874 | x025393 | 94/130 | 0.5145 | 0.8305 | 0.7143 | 0.3052 |
| x018878 | x025393 | 91/130 | 0.5169 | 0.8267 | 0.7143 | 0.3052 |
| x018879 | x025393 | 91/130 | 0.5160 | 0.8267 | 0.7143 | 0.3047 |
| x000108 | x037863 | 152/281 | 0.4698 | 0.9079 | 0.7143 | 0.3047 |
| x018873 | x025395 | 89/144 | 0.5097 | 0.8344 | 0.7143 | 0.3038 |
| x018872 | x025398 | 77/140 | 0.5222 | 0.8144 | 0.7143 | 0.3038 |
| x018875 | x025397 | 78/143 | 0.5180 | 0.8178 | 0.7143 | 0.3026 |
| x018875 | x025393 | 78/130 | 0.5219 | 0.8084 | 0.7143 | 0.3014 |
| x018876 | x025397 | 75/143 | 0.5176 | 0.8136 | 0.7143 | 0.3008 |

## Structurally close / graphically close (control)

| A | B | Counts | Structural | Reliability | Norm. distance | Discovery score |
|---|---|---:|---:|---:|---:|---:|

## Frequency control

| Minimum count for both tokens | Pairs | Pearson | Spearman | Distant candidates |
|---:|---:|---:|---:|---:|
| 10 | 220116 | 0.04556 | 0.01370 | 0 |
| 20 | 54615 | 0.11008 | 0.03541 | 0 |
| 50 | 7381 | 0.28458 | 0.08060 | 0 |
| 100 | 1128 | 0.57509 | 0.33223 | 0 |

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
