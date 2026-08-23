# Structural–graphemic analysis

This analysis compares two independent coordinates. The existing structural similarity was copied unchanged from `raw_similarity`; spelling was used only for the graphemic metrics.

## Scope

- Tokens: 560
- Pairs: 156520
- Pearson correlation (graphemic similarity vs structural similarity): 0.018396
- Spearman correlation: 0.000273

Pair observations share tokens and are not independent; the correlations are descriptive and ordinary pairwise p-values are intentionally not reported.

## Structural similarity by normalized grapheme-distance bin

| Distance | Pairs | Mean | Median | P90 | P95 |
|---|---:|---:|---:|---:|---:|
| 0.0–0.1 | 0 | 0.0000 | 0.0000 | 0.0000 | 0.0000 |
| 0.1–0.2 | 839 | 0.3626 | 0.3179 | 0.6023 | 0.6843 |
| 0.2–0.3 | 7792 | 0.2864 | 0.2775 | 0.3914 | 0.4363 |
| 0.3–0.4 | 0 | 0.0000 | 0.0000 | 0.0000 | 0.0000 |
| 0.4–0.5 | 44716 | 0.2793 | 0.2718 | 0.3803 | 0.4229 |
| 0.5–0.6 | 87958 | 0.2792 | 0.2718 | 0.3801 | 0.4233 |
| 0.6–0.7 | 0 | 0.0000 | 0.0000 | 0.0000 | 0.0000 |
| 0.7–0.8 | 15215 | 0.2849 | 0.2782 | 0.3889 | 0.4332 |
| 0.8–0.9 | 0 | 0.0000 | 0.0000 | 0.0000 | 0.0000 |
| 0.9–1.0 | 0 | 0.0000 | 0.0000 | 0.0000 | 0.0000 |

## Selection and percentile view

Configurable distant selection uses structural similarity ≥ 0.650, reliability ≥ 0.700, and normalized grapheme distance ≥ 0.600. It yields 26 pairs. Ranking is `structural_similarity × reliability × normalized_grapheme_distance`; this score is neither a probability nor statistical significance.

As a threshold-free companion view, corpus cutoffs are structural P95 = 0.4262, reliability P75 = 0.6973, and grapheme-distance P75 = 0.5714. Their intersection yields 2901 pairs, ranked by the same transparent score. These percentiles describe ranking coordinates and do not define the class.

## Structurally close / graphically distant

| A | B | Counts | Structural | Reliability | Norm. distance | Discovery score |
|---|---|---:|---:|---:|---:|---:|
| x004984 | x010279 | 232/303 | 0.8339 | 0.9243 | 0.7143 | 0.5505 |
| x004985 | x010279 | 229/303 | 0.8249 | 0.9243 | 0.7143 | 0.5446 |
| x004625 | x010314 | 623/167 | 0.6964 | 0.9243 | 0.7143 | 0.4597 |
| x004624 | x010315 | 613/179 | 0.6834 | 0.9243 | 0.7143 | 0.4512 |
| x004718 | x010054 | 338/102 | 0.6961 | 0.8933 | 0.7143 | 0.4442 |
| x004719 | x010526 | 341/203 | 0.6578 | 0.9243 | 0.7143 | 0.4343 |
| x004718 | x010526 | 338/203 | 0.6569 | 0.9243 | 0.7143 | 0.4337 |
| x004719 | x010527 | 341/182 | 0.6538 | 0.9243 | 0.7143 | 0.4316 |
| x006348 | x010055 | 549/91 | 0.6745 | 0.8852 | 0.7143 | 0.4265 |
| x004718 | x010055 | 338/91 | 0.6603 | 0.8852 | 0.7143 | 0.4175 |
| x009496 | x010315 | 71/179 | 0.6707 | 0.8659 | 0.7143 | 0.4148 |
| x006050 | x010484 | 40/50 | 0.7812 | 0.7403 | 0.7143 | 0.4130 |
| x006383 | x010055 | 84/91 | 0.6787 | 0.8429 | 0.7143 | 0.4086 |
| x005730 | x010617 | 32/67 | 0.7691 | 0.7411 | 0.7143 | 0.4071 |
| x006383 | x010054 | 84/102 | 0.6591 | 0.8505 | 0.7143 | 0.4004 |
| x009497 | x010315 | 53/179 | 0.6567 | 0.8406 | 0.7143 | 0.3943 |
| x008352 | x010617 | 32/67 | 0.7397 | 0.7411 | 0.7143 | 0.3916 |
| x003821 | x010055 | 77/91 | 0.6518 | 0.8366 | 0.7143 | 0.3895 |
| x005731 | x010617 | 38/67 | 0.6948 | 0.7572 | 0.7143 | 0.3758 |
| x005843 | x010617 | 24/67 | 0.7366 | 0.7134 | 0.7143 | 0.3754 |

## Percentile-ranked distant view

| A | B | Counts | Structural | Reliability | Norm. distance | Discovery score |
|---|---|---:|---:|---:|---:|---:|
| x004984 | x010279 | 232/303 | 0.8339 | 0.9243 | 0.7143 | 0.5505 |
| x004985 | x010279 | 229/303 | 0.8249 | 0.9243 | 0.7143 | 0.5446 |
| x004625 | x010314 | 623/167 | 0.6964 | 0.9243 | 0.7143 | 0.4597 |
| x004624 | x010315 | 613/179 | 0.6834 | 0.9243 | 0.7143 | 0.4512 |
| x004718 | x010054 | 338/102 | 0.6961 | 0.8933 | 0.7143 | 0.4442 |
| x004719 | x010526 | 341/203 | 0.6578 | 0.9243 | 0.7143 | 0.4343 |
| x004718 | x010526 | 338/203 | 0.6569 | 0.9243 | 0.7143 | 0.4337 |
| x004984 | x010278 | 232/271 | 0.8177 | 0.9243 | 0.5714 | 0.4319 |
| x004719 | x010527 | 341/182 | 0.6538 | 0.9243 | 0.7143 | 0.4316 |
| x004624 | x010701 | 613/278 | 0.6494 | 0.9243 | 0.7143 | 0.4287 |
| x006348 | x010055 | 549/91 | 0.6745 | 0.8852 | 0.7143 | 0.4265 |
| x004625 | x010701 | 623/278 | 0.6436 | 0.9243 | 0.7143 | 0.4249 |
| x004985 | x010278 | 229/271 | 0.7950 | 0.9243 | 0.5714 | 0.4199 |
| x004718 | x010055 | 338/91 | 0.6603 | 0.8852 | 0.7143 | 0.4175 |
| x009496 | x010315 | 71/179 | 0.6707 | 0.8659 | 0.7143 | 0.4148 |
| x006050 | x010484 | 40/50 | 0.7812 | 0.7403 | 0.7143 | 0.4130 |
| x006349 | x010055 | 564/91 | 0.6494 | 0.8852 | 0.7143 | 0.4106 |
| x004719 | x010054 | 341/102 | 0.6419 | 0.8933 | 0.7143 | 0.4096 |
| x006383 | x010055 | 84/91 | 0.6787 | 0.8429 | 0.7143 | 0.4086 |
| x006349 | x010526 | 564/203 | 0.6183 | 0.9243 | 0.7143 | 0.4082 |

## Structurally close / graphically close (control)

| A | B | Counts | Structural | Reliability | Norm. distance | Discovery score |
|---|---|---:|---:|---:|---:|---:|
| x004624 | x004625 | 613/623 | 0.8897 | 0.9243 | 0.1429 | 0.1175 |
| x004984 | x004985 | 232/229 | 0.8824 | 0.9243 | 0.1429 | 0.1165 |
| x004344 | x004345 | 175/168 | 0.8741 | 0.9243 | 0.1429 | 0.1154 |
| x010278 | x010279 | 271/303 | 0.8698 | 0.9243 | 0.1429 | 0.1148 |
| x004352 | x004353 | 323/322 | 0.8560 | 0.9243 | 0.1429 | 0.1130 |
| x005004 | x005005 | 349/348 | 0.8532 | 0.9243 | 0.1429 | 0.1127 |
| x009464 | x009465 | 1174/1151 | 0.8520 | 0.9243 | 0.1429 | 0.1125 |
| x004224 | x004225 | 172/185 | 0.8428 | 0.9243 | 0.1429 | 0.1113 |
| x010700 | x010701 | 265/278 | 0.8262 | 0.9243 | 0.1429 | 0.1091 |
| x009634 | x009635 | 546/547 | 0.8069 | 0.9243 | 0.1429 | 0.1065 |
| x000026 | x000027 | 520/571 | 0.8062 | 0.9243 | 0.1429 | 0.1064 |
| x000744 | x000745 | 140/121 | 0.8251 | 0.8964 | 0.1429 | 0.1057 |
| x006348 | x006349 | 549/564 | 0.7983 | 0.9243 | 0.1429 | 0.1054 |
| x009462 | x009463 | 317/314 | 0.7821 | 0.9243 | 0.1429 | 0.1033 |
| x007926 | x007927 | 116/108 | 0.8134 | 0.8762 | 0.1429 | 0.1018 |
| x010314 | x010315 | 167/179 | 0.7683 | 0.9243 | 0.1429 | 0.1014 |
| x002078 | x002079 | 73/71 | 0.8541 | 0.8146 | 0.1429 | 0.0994 |
| x004718 | x004719 | 338/341 | 0.7481 | 0.9243 | 0.1429 | 0.0988 |
| x005738 | x005739 | 112/114 | 0.7856 | 0.8775 | 0.1429 | 0.0985 |
| x004472 | x004473 | 264/216 | 0.7395 | 0.9243 | 0.1429 | 0.0976 |

## Frequency control

| Minimum count for both tokens | Pairs | Pearson | Spearman | Distant candidates |
|---:|---:|---:|---:|---:|
| 10 | 156520 | 0.01840 | 0.00027 | 26 |
| 20 | 47278 | 0.03144 | -0.00296 | 26 |
| 50 | 8256 | 0.08968 | 0.03984 | 15 |
| 100 | 1953 | 0.14861 | 0.04943 | 8 |

## Families

The same explicit edge criteria form connected components. There are 47 graphemic-structural families and 6 structural-distant families. Components can contain token pairs that are connected only through intermediate tokens; inspect the saved edge list before interpreting a whole component. The neutral term “family” denotes a graph component only.

### Graphemic-structural families

- Family 1 (2 tokens): x000026, x000027
- Family 2 (2 tokens): x000356, x000357
- Family 3 (2 tokens): x000472, x000473
- Family 4 (2 tokens): x000526, x000527
- Family 5 (2 tokens): x000584, x000585
- Family 6 (2 tokens): x000744, x000745
- Family 7 (2 tokens): x000798, x000799
- Family 8 (2 tokens): x001312, x001313
- Family 9 (2 tokens): x002078, x002079
- Family 10 (2 tokens): x002734, x002735
- Family 11 (2 tokens): x004224, x004225
- Family 12 (2 tokens): x004344, x004345
- Family 13 (2 tokens): x004352, x004353
- Family 14 (2 tokens): x004458, x004459
- Family 15 (2 tokens): x004472, x004473
- Family 16 (2 tokens): x004498, x004499
- Family 17 (2 tokens): x004624, x004625
- Family 18 (2 tokens): x004640, x004641
- Family 19 (2 tokens): x004718, x004719
- Family 20 (2 tokens): x004928, x004929

### Structural-distant families

- Family 1 (10 tokens): x003821, x004718, x004719, x004929, x006348, x006383, x010054, x010055, x010526, x010527
- Family 2 (4 tokens): x004624, x009496, x009497, x010315
- Family 3 (2 tokens): x004625, x010314
- Family 4 (3 tokens): x004984, x004985, x010279
- Family 5 (10 tokens): x005730, x005731, x005842, x005843, x008248, x008352, x008353, x010485, x010616, x010617
- Family 6 (2 tokens): x006050, x010484

## Graphemic-distance distribution

The bin counts above are the full empirical distribution of normalized edit distance. Edit operations are performed on grapheme sequences: `@NNN;` is one grapheme, `?` is one unknown grapheme, and no signs are deleted or normalized.

## Limitations

- Pair rows are dependent because each token occurs in many pairs.
- Reliability and frequency reduce, but cannot eliminate, instability of sparse profiles.
- Levenshtein distance assigns equal cost to every insertion, deletion, and substitution and contains no palaeographic model.
- Connected components are threshold-sensitive descriptive groups, not linguistic categories.
- Correlation does not establish that one coordinate causes the other.
- The analysis makes no claim about language, morphology, commands, operators, or cipher mechanisms.
