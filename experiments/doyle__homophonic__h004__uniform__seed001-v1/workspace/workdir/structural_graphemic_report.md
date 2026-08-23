# Structural–graphemic analysis

This analysis compares two independent coordinates. The existing structural similarity was copied unchanged from `raw_similarity`; spelling was used only for the graphemic metrics.

## Scope

- Tokens: 630
- Pairs: 198135
- Pearson correlation (graphemic similarity vs structural similarity): 0.017232
- Spearman correlation: -0.015438

Pair observations share tokens and are not independent; the correlations are descriptive and ordinary pairwise p-values are intentionally not reported.

## Structural similarity by normalized grapheme-distance bin

| Distance | Pairs | Mean | Median | P90 | P95 |
|---|---:|---:|---:|---:|---:|
| 0.0–0.1 | 0 | 0.0000 | 0.0000 | 0.0000 | 0.0000 |
| 0.1–0.2 | 1122 | 0.3518 | 0.3282 | 0.5358 | 0.6019 |
| 0.2–0.3 | 6279 | 0.2638 | 0.2589 | 0.3453 | 0.3816 |
| 0.3–0.4 | 0 | 0.0000 | 0.0000 | 0.0000 | 0.0000 |
| 0.4–0.5 | 35758 | 0.2565 | 0.2530 | 0.3366 | 0.3705 |
| 0.5–0.6 | 90700 | 0.2548 | 0.2514 | 0.3333 | 0.3646 |
| 0.6–0.7 | 0 | 0.0000 | 0.0000 | 0.0000 | 0.0000 |
| 0.7–0.8 | 64276 | 0.2602 | 0.2573 | 0.3392 | 0.3706 |
| 0.8–0.9 | 0 | 0.0000 | 0.0000 | 0.0000 | 0.0000 |
| 0.9–1.0 | 0 | 0.0000 | 0.0000 | 0.0000 | 0.0000 |

## Selection and percentile view

Configurable distant selection uses structural similarity ≥ 0.650, reliability ≥ 0.700, and normalized grapheme distance ≥ 0.600. It yields 8 pairs. Ranking is `structural_similarity × reliability × normalized_grapheme_distance`; this score is neither a probability nor statistical significance.

As a threshold-free companion view, corpus cutoffs are structural P95 = 0.3701, reliability P75 = 0.6310, and grapheme-distance P75 = 0.7143. Their intersection yields 2509 pairs, ranked by the same transparent score. These percentiles describe ranking coordinates and do not define the class.

## Structurally close / graphically distant

| A | B | Counts | Structural | Reliability | Norm. distance | Discovery score |
|---|---|---:|---:|---:|---:|---:|
| x009968 | x020559 | 133/154 | 0.6854 | 0.8927 | 0.7143 | 0.4370 |
| x009971 | x020558 | 110/149 | 0.6997 | 0.8719 | 0.7143 | 0.4358 |
| x009970 | x020559 | 119/154 | 0.6709 | 0.8822 | 0.7143 | 0.4228 |
| x009971 | x020556 | 110/137 | 0.6846 | 0.8646 | 0.7143 | 0.4227 |
| x009969 | x020558 | 99/149 | 0.6810 | 0.8618 | 0.7143 | 0.4192 |
| x009970 | x020556 | 119/137 | 0.6654 | 0.8719 | 0.7143 | 0.4144 |
| x009971 | x020559 | 110/154 | 0.6555 | 0.8747 | 0.7143 | 0.4096 |
| x009970 | x020558 | 119/149 | 0.6512 | 0.8793 | 0.7143 | 0.4090 |

## Percentile-ranked distant view

| A | B | Counts | Structural | Reliability | Norm. distance | Discovery score |
|---|---|---:|---:|---:|---:|---:|
| x009968 | x020559 | 133/154 | 0.6854 | 0.8927 | 0.7143 | 0.4370 |
| x009971 | x020558 | 110/149 | 0.6997 | 0.8719 | 0.7143 | 0.4358 |
| x009970 | x020559 | 119/154 | 0.6709 | 0.8822 | 0.7143 | 0.4228 |
| x009971 | x020556 | 110/137 | 0.6846 | 0.8646 | 0.7143 | 0.4227 |
| x009969 | x020558 | 99/149 | 0.6810 | 0.8618 | 0.7143 | 0.4192 |
| x009970 | x020556 | 119/137 | 0.6654 | 0.8719 | 0.7143 | 0.4144 |
| x009436 | x012697 | 166/280 | 0.6328 | 0.9131 | 0.7143 | 0.4127 |
| x009971 | x020559 | 110/154 | 0.6555 | 0.8747 | 0.7143 | 0.4096 |
| x009970 | x020558 | 119/149 | 0.6512 | 0.8793 | 0.7143 | 0.4090 |
| x009437 | x012698 | 172/289 | 0.6265 | 0.9131 | 0.7143 | 0.4086 |
| x009437 | x012696 | 172/269 | 0.6256 | 0.9131 | 0.7143 | 0.4080 |
| x009438 | x012696 | 159/269 | 0.6217 | 0.9125 | 0.7143 | 0.4052 |
| x009436 | x012698 | 166/289 | 0.6203 | 0.9131 | 0.7143 | 0.4046 |
| x009437 | x012699 | 172/275 | 0.6177 | 0.9131 | 0.7143 | 0.4029 |
| x009439 | x012698 | 182/289 | 0.6156 | 0.9131 | 0.7143 | 0.4015 |
| x009436 | x012699 | 166/275 | 0.6154 | 0.9131 | 0.7143 | 0.4014 |
| x009438 | x012697 | 159/280 | 0.6138 | 0.9125 | 0.7143 | 0.4001 |
| x009438 | x012699 | 159/275 | 0.6137 | 0.9125 | 0.7143 | 0.4000 |
| x009439 | x012696 | 182/269 | 0.6123 | 0.9131 | 0.7143 | 0.3993 |
| x009968 | x020557 | 133/134 | 0.6225 | 0.8802 | 0.7143 | 0.3913 |

## Structurally close / graphically close (control)

| A | B | Counts | Structural | Reliability | Norm. distance | Discovery score |
|---|---|---:|---:|---:|---:|---:|
| x009250 | x009251 | 314/309 | 0.7496 | 0.9131 | 0.1429 | 0.0978 |
| x020556 | x020558 | 137/149 | 0.7336 | 0.8925 | 0.1429 | 0.0935 |
| x009248 | x009249 | 297/316 | 0.7139 | 0.9131 | 0.1429 | 0.0931 |
| x018930 | x018931 | 576/575 | 0.7038 | 0.9131 | 0.1429 | 0.0918 |
| x020558 | x020559 | 149/154 | 0.6966 | 0.9031 | 0.1429 | 0.0899 |
| x020556 | x020559 | 137/154 | 0.6961 | 0.8954 | 0.1429 | 0.0890 |
| x008704 | x008706 | 163/148 | 0.6877 | 0.9060 | 0.1429 | 0.0890 |
| x012696 | x012697 | 269/280 | 0.6787 | 0.9131 | 0.1429 | 0.0885 |
| x018928 | x018929 | 572/602 | 0.6772 | 0.9131 | 0.1429 | 0.0883 |
| x012696 | x012698 | 269/289 | 0.6732 | 0.9131 | 0.1429 | 0.0878 |
| x012696 | x012699 | 269/275 | 0.6661 | 0.9131 | 0.1429 | 0.0869 |
| x008704 | x008705 | 163/160 | 0.6627 | 0.9131 | 0.1429 | 0.0864 |
| x020557 | x020558 | 134/149 | 0.6764 | 0.8904 | 0.1429 | 0.0860 |
| x009970 | x009971 | 119/110 | 0.7057 | 0.8520 | 0.1429 | 0.0859 |
| x010008 | x010009 | 180/169 | 0.6574 | 0.9131 | 0.1429 | 0.0858 |
| x012697 | x012698 | 280/289 | 0.6546 | 0.9131 | 0.1429 | 0.0854 |
| x012697 | x012699 | 280/275 | 0.6513 | 0.9131 | 0.1429 | 0.0850 |
| x020556 | x020557 | 137/134 | 0.6688 | 0.8828 | 0.1429 | 0.0843 |
| x009968 | x009969 | 133/99 | 0.6832 | 0.8520 | 0.1429 | 0.0831 |
| x008688 | x008689 | 94/81 | 0.7000 | 0.8026 | 0.1429 | 0.0803 |

## Frequency control

| Minimum count for both tokens | Pairs | Pearson | Spearman | Distant candidates |
|---:|---:|---:|---:|---:|
| 10 | 198135 | 0.01723 | -0.01544 | 8 |
| 20 | 55278 | 0.07590 | 0.02469 | 8 |
| 50 | 7875 | 0.18422 | 0.06314 | 8 |
| 100 | 1711 | 0.28404 | 0.09801 | 7 |

## Families

The same explicit edge criteria form connected components. There are 13 graphemic-structural families and 1 structural-distant families. Components can contain token pairs that are connected only through intermediate tokens; inspect the saved edge list before interpreting a whole component. The neutral term “family” denotes a graph component only.

### Graphemic-structural families

- Family 1 (2 tokens): x008688, x008689
- Family 2 (2 tokens): x008690, x008691
- Family 3 (3 tokens): x008704, x008705, x008706
- Family 4 (2 tokens): x009248, x009249
- Family 5 (2 tokens): x009250, x009251
- Family 6 (2 tokens): x009968, x009969
- Family 7 (2 tokens): x009970, x009971
- Family 8 (2 tokens): x010008, x010009
- Family 9 (4 tokens): x012696, x012697, x012698, x012699
- Family 10 (3 tokens): x015852, x015853, x015854
- Family 11 (2 tokens): x018928, x018929
- Family 12 (2 tokens): x018930, x018931
- Family 13 (4 tokens): x020556, x020557, x020558, x020559

### Structural-distant families

- Family 1 (7 tokens): x009968, x009969, x009970, x009971, x020556, x020558, x020559

## Graphemic-distance distribution

The bin counts above are the full empirical distribution of normalized edit distance. Edit operations are performed on grapheme sequences: `@NNN;` is one grapheme, `?` is one unknown grapheme, and no signs are deleted or normalized.

## Limitations

- Pair rows are dependent because each token occurs in many pairs.
- Reliability and frequency reduce, but cannot eliminate, instability of sparse profiles.
- Levenshtein distance assigns equal cost to every insertion, deletion, and substitution and contains no palaeographic model.
- Connected components are threshold-sensitive descriptive groups, not linguistic categories.
- Correlation does not establish that one coordinate causes the other.
- The analysis makes no claim about language, morphology, commands, operators, or cipher mechanisms.
