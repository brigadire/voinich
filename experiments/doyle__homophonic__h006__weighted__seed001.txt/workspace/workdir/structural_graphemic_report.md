# Structural–graphemic analysis

This analysis compares two independent coordinates. The existing structural similarity was copied unchanged from `raw_similarity`; spelling was used only for the graphemic metrics.

## Scope

- Tokens: 648
- Pairs: 209628
- Pearson correlation (graphemic similarity vs structural similarity): 0.053606
- Spearman correlation: 0.024016

Pair observations share tokens and are not independent; the correlations are descriptive and ordinary pairwise p-values are intentionally not reported.

## Structural similarity by normalized grapheme-distance bin

| Distance | Pairs | Mean | Median | P90 | P95 |
|---|---:|---:|---:|---:|---:|
| 0.0–0.1 | 0 | 0.0000 | 0.0000 | 0.0000 | 0.0000 |
| 0.1–0.2 | 1178 | 0.3536 | 0.3275 | 0.5341 | 0.5841 |
| 0.2–0.3 | 5007 | 0.2576 | 0.2521 | 0.3394 | 0.3795 |
| 0.3–0.4 | 0 | 0.0000 | 0.0000 | 0.0000 | 0.0000 |
| 0.4–0.5 | 31044 | 0.2499 | 0.2472 | 0.3265 | 0.3537 |
| 0.5–0.6 | 88834 | 0.2480 | 0.2460 | 0.3252 | 0.3525 |
| 0.6–0.7 | 0 | 0.0000 | 0.0000 | 0.0000 | 0.0000 |
| 0.7–0.8 | 83565 | 0.2477 | 0.2455 | 0.3254 | 0.3546 |
| 0.8–0.9 | 0 | 0.0000 | 0.0000 | 0.0000 | 0.0000 |
| 0.9–1.0 | 0 | 0.0000 | 0.0000 | 0.0000 | 0.0000 |

## Selection and percentile view

Configurable distant selection uses structural similarity ≥ 0.650, reliability ≥ 0.700, and normalized grapheme distance ≥ 0.600. It yields 1 pairs. Ranking is `structural_similarity × reliability × normalized_grapheme_distance`; this score is neither a probability nor statistical significance.

As a threshold-free companion view, corpus cutoffs are structural P95 = 0.3557, reliability P75 = 0.6122, and grapheme-distance P75 = 0.7143. Their intersection yields 3144 pairs, ranked by the same transparent score. These percentiles describe ranking coordinates and do not define the class.

## Structurally close / graphically distant

| A | B | Counts | Structural | Reliability | Norm. distance | Discovery score |
|---|---|---:|---:|---:|---:|---:|
| x014952 | x030834 | 149/155 | 0.6859 | 0.9008 | 0.7143 | 0.4413 |

## Percentile-ranked distant view

| A | B | Counts | Structural | Reliability | Norm. distance | Discovery score |
|---|---|---:|---:|---:|---:|---:|
| x014952 | x030834 | 149/155 | 0.6859 | 0.9008 | 0.7143 | 0.4413 |
| x014952 | x030836 | 149/115 | 0.6411 | 0.8715 | 0.7143 | 0.3991 |
| x014952 | x030837 | 149/92 | 0.6338 | 0.8486 | 0.7143 | 0.3841 |
| x014953 | x030834 | 94/155 | 0.6238 | 0.8543 | 0.7143 | 0.3807 |
| x014953 | x030836 | 94/115 | 0.6298 | 0.8272 | 0.7143 | 0.3721 |
| x001752 | x019044 | 103/316 | 0.5980 | 0.8667 | 0.7143 | 0.3702 |
| x000078 | x028394 | 297/446 | 0.5560 | 0.9108 | 0.7143 | 0.3617 |
| x014952 | x030835 | 149/124 | 0.5750 | 0.8790 | 0.7143 | 0.3611 |
| x013416 | x028394 | 134/446 | 0.5620 | 0.8934 | 0.7143 | 0.3586 |
| x000078 | x028393 | 297/563 | 0.5497 | 0.9108 | 0.7143 | 0.3576 |
| x014953 | x030835 | 94/124 | 0.5945 | 0.8342 | 0.7143 | 0.3542 |
| x013417 | x028393 | 137/563 | 0.5537 | 0.8956 | 0.7143 | 0.3542 |
| x014154 | x028902 | 189/311 | 0.5438 | 0.9108 | 0.7143 | 0.3538 |
| x013417 | x028392 | 137/661 | 0.5520 | 0.8956 | 0.7143 | 0.3532 |
| x001752 | x019045 | 103/257 | 0.5693 | 0.8667 | 0.7143 | 0.3524 |
| x014954 | x030836 | 86/115 | 0.6002 | 0.8182 | 0.7143 | 0.3508 |
| x013416 | x028392 | 134/661 | 0.5476 | 0.8934 | 0.7143 | 0.3494 |
| x000078 | x028392 | 297/661 | 0.5358 | 0.9108 | 0.7143 | 0.3486 |
| x000078 | x028395 | 297/324 | 0.5336 | 0.9108 | 0.7143 | 0.3471 |
| x014154 | x028905 | 189/145 | 0.5392 | 0.9012 | 0.7143 | 0.3471 |

## Structurally close / graphically close (control)

| A | B | Counts | Structural | Reliability | Norm. distance | Discovery score |
|---|---|---:|---:|---:|---:|---:|
| x013872 | x013873 | 343/310 | 0.7056 | 0.9108 | 0.1429 | 0.0918 |
| x013872 | x013874 | 343/235 | 0.6851 | 0.9108 | 0.1429 | 0.0891 |
| x028392 | x028394 | 661/446 | 0.6818 | 0.9108 | 0.1429 | 0.0887 |
| x028393 | x028394 | 563/446 | 0.6716 | 0.9108 | 0.1429 | 0.0874 |
| x019044 | x019045 | 316/257 | 0.6621 | 0.9108 | 0.1429 | 0.0862 |
| x013873 | x013874 | 310/235 | 0.6616 | 0.9108 | 0.1429 | 0.0861 |
| x028392 | x028393 | 661/563 | 0.6591 | 0.9108 | 0.1429 | 0.0858 |
| x030834 | x030836 | 155/115 | 0.6815 | 0.8751 | 0.1429 | 0.0852 |
| x014952 | x014954 | 149/86 | 0.6946 | 0.8415 | 0.1429 | 0.0835 |
| x030834 | x030835 | 155/124 | 0.6519 | 0.8827 | 0.1429 | 0.0822 |
| x013032 | x013033 | 103/79 | 0.6830 | 0.7995 | 0.1429 | 0.0780 |
| x013033 | x013034 | 79/73 | 0.6687 | 0.7675 | 0.1429 | 0.0733 |

## Frequency control

| Minimum count for both tokens | Pairs | Pearson | Spearman | Distant candidates |
|---:|---:|---:|---:|---:|
| 10 | 209628 | 0.05361 | 0.02402 | 1 |
| 20 | 52975 | 0.10317 | 0.03419 | 1 |
| 50 | 8128 | 0.20633 | 0.02965 | 1 |
| 100 | 1485 | 0.31929 | 0.03768 | 1 |

## Families

The same explicit edge criteria form connected components. There are 6 graphemic-structural families and 1 structural-distant families. Components can contain token pairs that are connected only through intermediate tokens; inspect the saved edge list before interpreting a whole component. The neutral term “family” denotes a graph component only.

### Graphemic-structural families

- Family 1 (3 tokens): x013032, x013033, x013034
- Family 2 (3 tokens): x013872, x013873, x013874
- Family 3 (2 tokens): x014952, x014954
- Family 4 (2 tokens): x019044, x019045
- Family 5 (3 tokens): x028392, x028393, x028394
- Family 6 (3 tokens): x030834, x030835, x030836

### Structural-distant families

- Family 1 (2 tokens): x014952, x030834

## Graphemic-distance distribution

The bin counts above are the full empirical distribution of normalized edit distance. Edit operations are performed on grapheme sequences: `@NNN;` is one grapheme, `?` is one unknown grapheme, and no signs are deleted or normalized.

## Limitations

- Pair rows are dependent because each token occurs in many pairs.
- Reliability and frequency reduce, but cannot eliminate, instability of sparse profiles.
- Levenshtein distance assigns equal cost to every insertion, deletion, and substitution and contains no palaeographic model.
- Connected components are threshold-sensitive descriptive groups, not linguistic categories.
- Correlation does not establish that one coordinate causes the other.
- The analysis makes no claim about language, morphology, commands, operators, or cipher mechanisms.
