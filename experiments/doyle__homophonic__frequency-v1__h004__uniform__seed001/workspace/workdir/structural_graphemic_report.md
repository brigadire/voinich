# Structural–graphemic analysis

This analysis compares two independent coordinates. The existing structural similarity was copied unchanged from `raw_similarity`; spelling was used only for the graphemic metrics.

## Scope

- Tokens: 616
- Pairs: 189420
- Pearson correlation (graphemic similarity vs structural similarity): 0.093836
- Spearman correlation: 0.068064

Pair observations share tokens and are not independent; the correlations are descriptive and ordinary pairwise p-values are intentionally not reported.

## Structural similarity by normalized grapheme-distance bin

| Distance | Pairs | Mean | Median | P90 | P95 |
|---|---:|---:|---:|---:|---:|
| 0.0–0.1 | 0 | 0.0000 | 0.0000 | 0.0000 | 0.0000 |
| 0.1–0.2 | 6435 | 0.2904 | 0.2763 | 0.4134 | 0.4936 |
| 0.2–0.3 | 53739 | 0.2702 | 0.2649 | 0.3633 | 0.4033 |
| 0.3–0.4 | 0 | 0.0000 | 0.0000 | 0.0000 | 0.0000 |
| 0.4–0.5 | 129246 | 0.2596 | 0.2567 | 0.3394 | 0.3706 |
| 0.5–0.6 | 0 | 0.0000 | 0.0000 | 0.0000 | 0.0000 |
| 0.6–0.7 | 0 | 0.0000 | 0.0000 | 0.0000 | 0.0000 |
| 0.7–0.8 | 0 | 0.0000 | 0.0000 | 0.0000 | 0.0000 |
| 0.8–0.9 | 0 | 0.0000 | 0.0000 | 0.0000 | 0.0000 |
| 0.9–1.0 | 0 | 0.0000 | 0.0000 | 0.0000 | 0.0000 |

## Selection and percentile view

Configurable distant selection uses structural similarity ≥ 0.650, reliability ≥ 0.700, and normalized grapheme distance ≥ 0.600. It yields 0 pairs. Ranking is `structural_similarity × reliability × normalized_grapheme_distance`; this score is neither a probability nor statistical significance.

As a threshold-free companion view, corpus cutoffs are structural P95 = 0.3841, reliability P75 = 0.6446, and grapheme-distance P75 = 0.4286. Their intersection yields 3079 pairs, ranked by the same transparent score. These percentiles describe ranking coordinates and do not define the class.

## Structurally close / graphically distant

| A | B | Counts | Structural | Reliability | Norm. distance | Discovery score |
|---|---|---:|---:|---:|---:|---:|

## Percentile-ranked distant view

| A | B | Counts | Structural | Reliability | Norm. distance | Discovery score |
|---|---|---:|---:|---:|---:|---:|
| x000023 | x000105 | 214/64 | 0.6682 | 0.8286 | 0.4286 | 0.2373 |
| x000022 | x000103 | 227/69 | 0.6399 | 0.8357 | 0.4286 | 0.2292 |
| x000022 | x000105 | 227/64 | 0.6290 | 0.8286 | 0.4286 | 0.2233 |
| x000012 | x000105 | 378/64 | 0.6256 | 0.8286 | 0.4286 | 0.2222 |
| x000012 | x000103 | 378/69 | 0.6177 | 0.8357 | 0.4286 | 0.2212 |
| x000024 | x000103 | 238/69 | 0.6177 | 0.8357 | 0.4286 | 0.2212 |
| x000011 | x000105 | 375/64 | 0.6224 | 0.8286 | 0.4286 | 0.2210 |
| x000052 | x000103 | 122/69 | 0.6178 | 0.8159 | 0.4286 | 0.2160 |
| x000011 | x000103 | 375/69 | 0.6022 | 0.8357 | 0.4286 | 0.2157 |
| x000024 | x000105 | 238/64 | 0.6039 | 0.8286 | 0.4286 | 0.2145 |
| x000023 | x000115 | 214/61 | 0.5881 | 0.8240 | 0.4286 | 0.2077 |
| x000047 | x000133 | 109/53 | 0.6140 | 0.7831 | 0.4286 | 0.2060 |
| x000022 | x000115 | 227/61 | 0.5789 | 0.8240 | 0.4286 | 0.2044 |
| x000029 | x000138 | 193/50 | 0.5900 | 0.8045 | 0.4286 | 0.2034 |
| x000014 | x000103 | 390/69 | 0.5663 | 0.8357 | 0.4286 | 0.2028 |
| x000022 | x000148 | 227/43 | 0.5994 | 0.7892 | 0.4286 | 0.2028 |
| x000022 | x000150 | 227/54 | 0.5817 | 0.8121 | 0.4286 | 0.2025 |
| x000048 | x000133 | 130/53 | 0.5936 | 0.7957 | 0.4286 | 0.2024 |
| x000046 | x000133 | 118/53 | 0.5967 | 0.7888 | 0.4286 | 0.2017 |
| x000043 | x000105 | 137/64 | 0.5750 | 0.8174 | 0.4286 | 0.2014 |

## Structurally close / graphically close (control)

| A | B | Counts | Structural | Reliability | Norm. distance | Discovery score |
|---|---|---:|---:|---:|---:|---:|
| x000005 | x000006 | 436/397 | 0.8429 | 0.9086 | 0.1429 | 0.1094 |
| x000004 | x000005 | 403/436 | 0.8037 | 0.9086 | 0.1429 | 0.1043 |
| x000004 | x000006 | 403/397 | 0.7986 | 0.9086 | 0.1429 | 0.1037 |
| x000031 | x000032 | 180/187 | 0.7877 | 0.9086 | 0.1429 | 0.1022 |
| x000040 | x000042 | 169/157 | 0.7839 | 0.9071 | 0.1429 | 0.1016 |
| x000031 | x000033 | 180/207 | 0.7799 | 0.9086 | 0.1429 | 0.1012 |
| x000025 | x000027 | 230/233 | 0.7671 | 0.9086 | 0.1429 | 0.0996 |
| x000041 | x000042 | 135/157 | 0.7777 | 0.8931 | 0.1429 | 0.0992 |
| x000025 | x000026 | 230/182 | 0.7628 | 0.9086 | 0.1429 | 0.0990 |
| x000040 | x000041 | 169/135 | 0.7731 | 0.8947 | 0.1429 | 0.0988 |
| x000001 | x000002 | 602/576 | 0.7496 | 0.9086 | 0.1429 | 0.0973 |
| x000020 | x000021 | 251/218 | 0.7481 | 0.9086 | 0.1429 | 0.0971 |
| x000000 | x000002 | 572/576 | 0.7431 | 0.9086 | 0.1429 | 0.0965 |
| x000001 | x000003 | 602/575 | 0.7401 | 0.9086 | 0.1429 | 0.0961 |
| x000010 | x000011 | 360/375 | 0.7390 | 0.9086 | 0.1429 | 0.0959 |
| x000002 | x000003 | 576/575 | 0.7325 | 0.9086 | 0.1429 | 0.0951 |
| x000026 | x000027 | 182/233 | 0.7277 | 0.9086 | 0.1429 | 0.0945 |
| x000058 | x000059 | 115/127 | 0.7646 | 0.8630 | 0.1429 | 0.0943 |
| x000032 | x000033 | 187/207 | 0.7253 | 0.9086 | 0.1429 | 0.0941 |
| x000032 | x000042 | 187/157 | 0.7234 | 0.9071 | 0.1429 | 0.0937 |

## Frequency control

| Minimum count for both tokens | Pairs | Pearson | Spearman | Distant candidates |
|---:|---:|---:|---:|---:|
| 10 | 189420 | 0.09384 | 0.06806 | 0 |
| 20 | 51681 | 0.11798 | 0.07471 | 0 |
| 50 | 8778 | 0.21228 | 0.20492 | 0 |
| 100 | 2145 | 0.15629 | 0.09671 | 0 |

## Families

The same explicit edge criteria form connected components. There are 19 graphemic-structural families and 0 structural-distant families. Components can contain token pairs that are connected only through intermediate tokens; inspect the saved edge list before interpreting a whole component. The neutral term “family” denotes a graph component only.

### Graphemic-structural families

- Family 1 (4 tokens): x000000, x000001, x000002, x000003
- Family 2 (6 tokens): x000004, x000005, x000006, x000025, x000026, x000027
- Family 3 (3 tokens): x000010, x000011, x000012
- Family 4 (3 tokens): x000013, x000014, x000015
- Family 5 (3 tokens): x000016, x000017, x000018
- Family 6 (2 tokens): x000020, x000021
- Family 7 (3 tokens): x000022, x000023, x000024
- Family 8 (2 tokens): x000028, x000029
- Family 9 (6 tokens): x000031, x000032, x000033, x000040, x000041, x000042
- Family 10 (3 tokens): x000034, x000035, x000036
- Family 11 (2 tokens): x000037, x000038
- Family 12 (3 tokens): x000046, x000047, x000048
- Family 13 (2 tokens): x000056, x000057
- Family 14 (2 tokens): x000058, x000059
- Family 15 (3 tokens): x000070, x000071, x000072
- Family 16 (3 tokens): x000076, x000077, x000078
- Family 17 (3 tokens): x000082, x000083, x000084
- Family 18 (2 tokens): x000133, x000135
- Family 19 (2 tokens): x000154, x000156

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
