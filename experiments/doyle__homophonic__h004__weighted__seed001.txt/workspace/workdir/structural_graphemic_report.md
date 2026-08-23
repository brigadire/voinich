# Structural–graphemic analysis

This analysis compares two independent coordinates. The existing structural similarity was copied unchanged from `raw_similarity`; spelling was used only for the graphemic metrics.

## Scope

- Tokens: 622
- Pairs: 193131
- Pearson correlation (graphemic similarity vs structural similarity): 0.035450
- Spearman correlation: 0.008452

Pair observations share tokens and are not independent; the correlations are descriptive and ordinary pairwise p-values are intentionally not reported.

## Structural similarity by normalized grapheme-distance bin

| Distance | Pairs | Mean | Median | P90 | P95 |
|---|---:|---:|---:|---:|---:|
| 0.0–0.1 | 0 | 0.0000 | 0.0000 | 0.0000 | 0.0000 |
| 0.1–0.2 | 989 | 0.3606 | 0.3351 | 0.5650 | 0.6258 |
| 0.2–0.3 | 5971 | 0.2687 | 0.2615 | 0.3553 | 0.3923 |
| 0.3–0.4 | 0 | 0.0000 | 0.0000 | 0.0000 | 0.0000 |
| 0.4–0.5 | 34503 | 0.2612 | 0.2566 | 0.3443 | 0.3794 |
| 0.5–0.6 | 87539 | 0.2592 | 0.2551 | 0.3415 | 0.3752 |
| 0.6–0.7 | 0 | 0.0000 | 0.0000 | 0.0000 | 0.0000 |
| 0.7–0.8 | 64129 | 0.2610 | 0.2568 | 0.3457 | 0.3791 |
| 0.8–0.9 | 0 | 0.0000 | 0.0000 | 0.0000 | 0.0000 |
| 0.9–1.0 | 0 | 0.0000 | 0.0000 | 0.0000 | 0.0000 |

## Selection and percentile view

Configurable distant selection uses structural similarity ≥ 0.650, reliability ≥ 0.700, and normalized grapheme distance ≥ 0.600. It yields 7 pairs. Ranking is `structural_similarity × reliability × normalized_grapheme_distance`; this score is neither a probability nor statistical significance.

As a threshold-free companion view, corpus cutoffs are structural P95 = 0.3798, reliability P75 = 0.6441, and grapheme-distance P75 = 0.7143. Their intersection yields 2288 pairs, ranked by the same transparent score. These percentiles describe ranking coordinates and do not define the class.

## Structurally close / graphically distant

| A | B | Counts | Structural | Reliability | Norm. distance | Discovery score |
|---|---|---:|---:|---:|---:|---:|
| x009968 | x020557 | 188/174 | 0.7192 | 0.9136 | 0.7143 | 0.4694 |
| x009969 | x020557 | 135/174 | 0.7146 | 0.8992 | 0.7143 | 0.4590 |
| x009969 | x020558 | 135/122 | 0.6983 | 0.8765 | 0.7143 | 0.4372 |
| x009436 | x012697 | 279/351 | 0.6602 | 0.9136 | 0.7143 | 0.4308 |
| x009970 | x020556 | 95/213 | 0.6880 | 0.8682 | 0.7143 | 0.4266 |
| x009970 | x020558 | 95/122 | 0.6553 | 0.8466 | 0.7143 | 0.3963 |
| x009971 | x020556 | 43/213 | 0.6619 | 0.7914 | 0.7143 | 0.3741 |

## Percentile-ranked distant view

| A | B | Counts | Structural | Reliability | Norm. distance | Discovery score |
|---|---|---:|---:|---:|---:|---:|
| x009968 | x020557 | 188/174 | 0.7192 | 0.9136 | 0.7143 | 0.4694 |
| x009969 | x020557 | 135/174 | 0.7146 | 0.8992 | 0.7143 | 0.4590 |
| x009969 | x020558 | 135/122 | 0.6983 | 0.8765 | 0.7143 | 0.4372 |
| x009436 | x012697 | 279/351 | 0.6602 | 0.9136 | 0.7143 | 0.4308 |
| x009970 | x020556 | 95/213 | 0.6880 | 0.8682 | 0.7143 | 0.4266 |
| x009437 | x012696 | 186/426 | 0.6431 | 0.9136 | 0.7143 | 0.4197 |
| x009436 | x012698 | 279/224 | 0.6397 | 0.9136 | 0.7143 | 0.4175 |
| x009438 | x012697 | 139/351 | 0.6233 | 0.9017 | 0.7143 | 0.4014 |
| x009436 | x021052 | 279/165 | 0.6079 | 0.9136 | 0.7143 | 0.3967 |
| x009970 | x020558 | 95/122 | 0.6553 | 0.8466 | 0.7143 | 0.3963 |
| x009438 | x012696 | 139/426 | 0.6133 | 0.9017 | 0.7143 | 0.3950 |
| x009436 | x012764 | 279/78 | 0.6498 | 0.8500 | 0.7143 | 0.3945 |
| x009436 | x020108 | 279/77 | 0.6493 | 0.8488 | 0.7143 | 0.3937 |
| x009437 | x012698 | 186/224 | 0.6000 | 0.9136 | 0.7143 | 0.3915 |
| x000052 | x018930 | 405/456 | 0.5978 | 0.9136 | 0.7143 | 0.3901 |
| x000052 | x018929 | 405/697 | 0.5963 | 0.9136 | 0.7143 | 0.3891 |
| x009437 | x021052 | 186/165 | 0.5938 | 0.9136 | 0.7143 | 0.3875 |
| x009438 | x021052 | 139/165 | 0.5992 | 0.9017 | 0.7143 | 0.3859 |
| x000052 | x018928 | 405/943 | 0.5902 | 0.9136 | 0.7143 | 0.3852 |
| x012696 | x020108 | 426/77 | 0.6327 | 0.8488 | 0.7143 | 0.3836 |

## Structurally close / graphically close (control)

| A | B | Counts | Structural | Reliability | Norm. distance | Discovery score |
|---|---|---:|---:|---:|---:|---:|
| x009248 | x009249 | 495/383 | 0.7953 | 0.9136 | 0.1429 | 0.1038 |
| x009968 | x009969 | 188/135 | 0.7891 | 0.8992 | 0.1429 | 0.1014 |
| x020556 | x020557 | 213/174 | 0.7743 | 0.9136 | 0.1429 | 0.1011 |
| x008704 | x008705 | 265/175 | 0.7681 | 0.9136 | 0.1429 | 0.1003 |
| x018928 | x018929 | 943/697 | 0.7638 | 0.9136 | 0.1429 | 0.0997 |
| x012696 | x012697 | 426/351 | 0.7271 | 0.9136 | 0.1429 | 0.0949 |
| x010008 | x010009 | 275/220 | 0.7216 | 0.9136 | 0.1429 | 0.0942 |
| x008688 | x008689 | 139/111 | 0.7351 | 0.8707 | 0.1429 | 0.0914 |
| x018924 | x018925 | 261/169 | 0.6987 | 0.9136 | 0.1429 | 0.0912 |
| x019268 | x019269 | 425/361 | 0.6964 | 0.9136 | 0.1429 | 0.0909 |
| x009436 | x009437 | 279/186 | 0.6892 | 0.9136 | 0.1429 | 0.0900 |
| x020556 | x020558 | 213/122 | 0.7004 | 0.8904 | 0.1429 | 0.0891 |
| x000052 | x000053 | 405/342 | 0.6822 | 0.9136 | 0.1429 | 0.0890 |
| x021400 | x021401 | 213/163 | 0.6802 | 0.9136 | 0.1429 | 0.0888 |
| x008448 | x008449 | 139/106 | 0.7150 | 0.8667 | 0.1429 | 0.0885 |
| x012696 | x012698 | 426/224 | 0.6688 | 0.9136 | 0.1429 | 0.0873 |
| x012697 | x012698 | 351/224 | 0.6552 | 0.9136 | 0.1429 | 0.0855 |
| x009436 | x009438 | 279/139 | 0.6636 | 0.9017 | 0.1429 | 0.0855 |
| x008704 | x008706 | 265/116 | 0.6647 | 0.8860 | 0.1429 | 0.0841 |
| x018924 | x018926 | 261/136 | 0.6522 | 0.8998 | 0.1429 | 0.0838 |

## Frequency control

| Minimum count for both tokens | Pairs | Pearson | Spearman | Distant candidates |
|---:|---:|---:|---:|---:|
| 10 | 193131 | 0.03545 | 0.00845 | 7 |
| 20 | 49455 | 0.06440 | 0.01406 | 7 |
| 50 | 8385 | 0.15102 | 0.05313 | 6 |
| 100 | 1830 | 0.23425 | 0.07032 | 4 |

## Families

The same explicit edge criteria form connected components. There are 20 graphemic-structural families and 2 structural-distant families. Components can contain token pairs that are connected only through intermediate tokens; inspect the saved edge list before interpreting a whole component. The neutral term “family” denotes a graph component only.

### Graphemic-structural families

- Family 1 (2 tokens): x000052, x000053
- Family 2 (2 tokens): x001488, x001489
- Family 3 (2 tokens): x004156, x004158
- Family 4 (2 tokens): x008448, x008449
- Family 5 (2 tokens): x008688, x008689
- Family 6 (4 tokens): x008704, x008705, x008706, x008707
- Family 7 (2 tokens): x009248, x009249
- Family 8 (2 tokens): x009250, x009251
- Family 9 (3 tokens): x009436, x009437, x009438
- Family 10 (2 tokens): x009968, x009969
- Family 11 (2 tokens): x009970, x009971
- Family 12 (2 tokens): x010008, x010009
- Family 13 (3 tokens): x012696, x012697, x012698
- Family 14 (3 tokens): x015852, x015853, x015854
- Family 15 (3 tokens): x018924, x018925, x018926
- Family 16 (2 tokens): x018928, x018929
- Family 17 (2 tokens): x018976, x018977
- Family 18 (2 tokens): x019268, x019269
- Family 19 (4 tokens): x020556, x020557, x020558, x020559
- Family 20 (2 tokens): x021400, x021401

### Structural-distant families

- Family 1 (2 tokens): x009436, x012697
- Family 2 (7 tokens): x009968, x009969, x009970, x009971, x020556, x020557, x020558

## Graphemic-distance distribution

The bin counts above are the full empirical distribution of normalized edit distance. Edit operations are performed on grapheme sequences: `@NNN;` is one grapheme, `?` is one unknown grapheme, and no signs are deleted or normalized.

## Limitations

- Pair rows are dependent because each token occurs in many pairs.
- Reliability and frequency reduce, but cannot eliminate, instability of sparse profiles.
- Levenshtein distance assigns equal cost to every insertion, deletion, and substitution and contains no palaeographic model.
- Connected components are threshold-sensitive descriptive groups, not linguistic categories.
- Correlation does not establish that one coordinate causes the other.
- The analysis makes no claim about language, morphology, commands, operators, or cipher mechanisms.
