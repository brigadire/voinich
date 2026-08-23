# Structural–graphemic analysis

This analysis compares two independent coordinates. The existing structural similarity was copied unchanged from `raw_similarity`; spelling was used only for the graphemic metrics.

## Scope

- Tokens: 643
- Pairs: 206403
- Pearson correlation (graphemic similarity vs structural similarity): 0.107384
- Spearman correlation: 0.079847

Pair observations share tokens and are not independent; the correlations are descriptive and ordinary pairwise p-values are intentionally not reported.

## Structural similarity by normalized grapheme-distance bin

| Distance | Pairs | Mean | Median | P90 | P95 |
|---|---:|---:|---:|---:|---:|
| 0.0–0.1 | 0 | 0.0000 | 0.0000 | 0.0000 | 0.0000 |
| 0.1–0.2 | 6743 | 0.2781 | 0.2661 | 0.3847 | 0.4441 |
| 0.2–0.3 | 57155 | 0.2585 | 0.2543 | 0.3420 | 0.3740 |
| 0.3–0.4 | 0 | 0.0000 | 0.0000 | 0.0000 | 0.0000 |
| 0.4–0.5 | 141289 | 0.2487 | 0.2477 | 0.3193 | 0.3444 |
| 0.5–0.6 | 1216 | 0.2040 | 0.2046 | 0.2678 | 0.2860 |
| 0.6–0.7 | 0 | 0.0000 | 0.0000 | 0.0000 | 0.0000 |
| 0.7–0.8 | 0 | 0.0000 | 0.0000 | 0.0000 | 0.0000 |
| 0.8–0.9 | 0 | 0.0000 | 0.0000 | 0.0000 | 0.0000 |
| 0.9–1.0 | 0 | 0.0000 | 0.0000 | 0.0000 | 0.0000 |

## Selection and percentile view

Configurable distant selection uses structural similarity ≥ 0.650, reliability ≥ 0.700, and normalized grapheme distance ≥ 0.600. It yields 0 pairs. Ranking is `structural_similarity × reliability × normalized_grapheme_distance`; this score is neither a probability nor statistical significance.

As a threshold-free companion view, corpus cutoffs are structural P95 = 0.3564, reliability P75 = 0.6054, and grapheme-distance P75 = 0.4286. Their intersection yields 3907 pairs, ranked by the same transparent score. These percentiles describe ranking coordinates and do not define the class.

## Structurally close / graphically distant

| A | B | Counts | Structural | Reliability | Norm. distance | Discovery score |
|---|---|---:|---:|---:|---:|---:|

## Percentile-ranked distant view

| A | B | Counts | Structural | Reliability | Norm. distance | Discovery score |
|---|---|---:|---:|---:|---:|---:|
| x000098 | x000100 | 76/66 | 0.6536 | 0.7437 | 0.4286 | 0.2083 |
| x000097 | x000100 | 70/66 | 0.6135 | 0.7363 | 0.4286 | 0.1936 |
| x000099 | x000100 | 62/66 | 0.6158 | 0.7252 | 0.4286 | 0.1914 |
| x000016 | x000158 | 224/41 | 0.5871 | 0.7583 | 0.4286 | 0.1908 |
| x000096 | x000100 | 69/66 | 0.6049 | 0.7350 | 0.4286 | 0.1905 |
| x000078 | x000100 | 67/66 | 0.5965 | 0.7323 | 0.4286 | 0.1872 |
| x000016 | x000175 | 224/37 | 0.5817 | 0.7469 | 0.4286 | 0.1862 |
| x000031 | x000129 | 143/51 | 0.5583 | 0.7740 | 0.4286 | 0.1852 |
| x000034 | x000128 | 147/50 | 0.5573 | 0.7741 | 0.4286 | 0.1849 |
| x000019 | x000175 | 237/37 | 0.5758 | 0.7469 | 0.4286 | 0.1843 |
| x000036 | x000172 | 133/39 | 0.5816 | 0.7383 | 0.4286 | 0.1840 |
| x000034 | x000126 | 147/50 | 0.5533 | 0.7741 | 0.4286 | 0.1836 |
| x000017 | x000158 | 202/41 | 0.5613 | 0.7583 | 0.4286 | 0.1824 |
| x000036 | x000158 | 133/41 | 0.5691 | 0.7438 | 0.4286 | 0.1814 |
| x000019 | x000158 | 237/41 | 0.5559 | 0.7583 | 0.4286 | 0.1807 |
| x000037 | x000175 | 146/37 | 0.5684 | 0.7399 | 0.4286 | 0.1802 |
| x000033 | x000126 | 150/50 | 0.5421 | 0.7758 | 0.4286 | 0.1802 |
| x000039 | x000171 | 137/39 | 0.5677 | 0.7406 | 0.4286 | 0.1802 |
| x000018 | x000175 | 230/37 | 0.5607 | 0.7469 | 0.4286 | 0.1795 |
| x000036 | x000175 | 133/37 | 0.5697 | 0.7326 | 0.4286 | 0.1789 |

## Structurally close / graphically close (control)

| A | B | Counts | Structural | Reliability | Norm. distance | Discovery score |
|---|---|---:|---:|---:|---:|---:|
| x000008 | x000009 | 252/256 | 0.6836 | 0.9057 | 0.1429 | 0.0884 |
| x000006 | x000009 | 241/256 | 0.6687 | 0.9057 | 0.1429 | 0.0865 |
| x000066 | x000069 | 109/98 | 0.6578 | 0.8176 | 0.1429 | 0.0768 |

## Frequency control

| Minimum count for both tokens | Pairs | Pearson | Spearman | Distant candidates |
|---:|---:|---:|---:|---:|
| 10 | 206403 | 0.10738 | 0.07985 | 0 |
| 20 | 53956 | 0.12452 | 0.08983 | 0 |
| 50 | 7875 | 0.18191 | 0.13222 | 0 |
| 100 | 2016 | 0.18001 | 0.10426 | 0 |

## Families

The same explicit edge criteria form connected components. There are 2 graphemic-structural families and 0 structural-distant families. Components can contain token pairs that are connected only through intermediate tokens; inspect the saved edge list before interpreting a whole component. The neutral term “family” denotes a graph component only.

### Graphemic-structural families

- Family 1 (3 tokens): x000006, x000008, x000009
- Family 2 (2 tokens): x000066, x000069

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
