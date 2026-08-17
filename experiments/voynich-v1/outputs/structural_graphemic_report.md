# Structural–graphemic analysis

This analysis compares two independent coordinates. The existing structural similarity was copied unchanged from `raw_similarity`; spelling was used only for the graphemic metrics.

## Scope

- Tokens: 539
- Pairs: 144991
- Pearson correlation (graphemic similarity vs structural similarity): 0.149532
- Spearman correlation: 0.122062

Pair observations share tokens and are not independent; the correlations are descriptive and ordinary pairwise p-values are intentionally not reported.

## Structural similarity by normalized grapheme-distance bin

| Distance | Pairs | Mean | Median | P90 | P95 |
|---|---:|---:|---:|---:|---:|
| 0.0–0.1 | 0 | 0.0000 | 0.0000 | 0.0000 | 0.0000 |
| 0.1–0.2 | 574 | 0.3395 | 0.3228 | 0.4801 | 0.5654 |
| 0.2–0.3 | 1564 | 0.3247 | 0.3060 | 0.4713 | 0.5316 |
| 0.3–0.4 | 2288 | 0.3065 | 0.2970 | 0.4435 | 0.4858 |
| 0.4–0.5 | 4029 | 0.2963 | 0.2883 | 0.4219 | 0.4723 |
| 0.5–0.6 | 11400 | 0.2806 | 0.2767 | 0.3943 | 0.4385 |
| 0.6–0.7 | 18660 | 0.2729 | 0.2702 | 0.3895 | 0.4292 |
| 0.7–0.8 | 13813 | 0.2644 | 0.2633 | 0.3675 | 0.4058 |
| 0.8–0.9 | 39746 | 0.2552 | 0.2553 | 0.3592 | 0.3939 |
| 0.9–1.0 | 52917 | 0.2538 | 0.2536 | 0.3682 | 0.4076 |

## Selection and percentile view

Configurable distant selection uses structural similarity ≥ 0.650, reliability ≥ 0.700, and normalized grapheme distance ≥ 0.600. It yields 26 pairs. Ranking is `structural_similarity × reliability × normalized_grapheme_distance`; this score is neither a probability nor statistical significance.

As a threshold-free companion view, corpus cutoffs are structural P95 = 0.4159, reliability P75 = 0.6794, and grapheme-distance P75 = 1.0000. Their intersection yields 2157 pairs, ranked by the same transparent score. These percentiles describe ranking coordinates and do not define the class.

## Structurally close / graphically distant

| A | B | Counts | Structural | Reliability | Norm. distance | Discovery score |
|---|---|---:|---:|---:|---:|---:|
| or | s | 390/350 | 0.7588 | 0.9332 | 1.0000 | 0.7081 |
| chol | daiin | 395/848 | 0.7202 | 0.9332 | 1.0000 | 0.6721 |
| r | s | 169/350 | 0.6989 | 0.9332 | 1.0000 | 0.6522 |
| ol | y | 560/305 | 0.6788 | 0.9332 | 1.0000 | 0.6334 |
| dar | ol | 323/560 | 0.6746 | 0.9332 | 1.0000 | 0.6295 |
| ar | ol | 403/560 | 0.6639 | 0.9332 | 1.0000 | 0.6195 |
| chor | daiin | 211/848 | 0.6627 | 0.9332 | 1.0000 | 0.6184 |
| chey | ol | 351/560 | 0.6529 | 0.9332 | 1.0000 | 0.6093 |
| lchedy | qokar | 116/159 | 0.6585 | 0.9065 | 1.0000 | 0.5969 |
| lchedy | qol | 116/142 | 0.6556 | 0.8977 | 1.0000 | 0.5886 |
| okaiin | ol | 216/560 | 0.6885 | 0.9332 | 0.8333 | 0.5354 |
| okain | ol | 141/560 | 0.6828 | 0.9230 | 0.8000 | 0.5042 |
| qokaiin | qol | 265/142 | 0.7398 | 0.9236 | 0.7143 | 0.4881 |
| daiin | dol | 848/109 | 0.6672 | 0.9018 | 0.8000 | 0.4813 |
| aiin | ar | 504/403 | 0.6715 | 0.9332 | 0.7500 | 0.4700 |
| qokain | qol | 279/142 | 0.7614 | 0.9236 | 0.6667 | 0.4688 |
| chol | cthy | 395/103 | 0.6832 | 0.8970 | 0.7500 | 0.4596 |
| lchedy | qokeey | 116/307 | 0.7030 | 0.9069 | 0.6667 | 0.4251 |
| chedy | qokeey | 506/307 | 0.6743 | 0.9332 | 0.6667 | 0.4195 |
| qokedy | qol | 276/142 | 0.6804 | 0.9236 | 0.6667 | 0.4189 |

## Percentile-ranked distant view

| A | B | Counts | Structural | Reliability | Norm. distance | Discovery score |
|---|---|---:|---:|---:|---:|---:|
| or | s | 390/350 | 0.7588 | 0.9332 | 1.0000 | 0.7081 |
| chol | daiin | 395/848 | 0.7202 | 0.9332 | 1.0000 | 0.6721 |
| r | s | 169/350 | 0.6989 | 0.9332 | 1.0000 | 0.6522 |
| ol | y | 560/305 | 0.6788 | 0.9332 | 1.0000 | 0.6334 |
| dar | ol | 323/560 | 0.6746 | 0.9332 | 1.0000 | 0.6295 |
| ar | ol | 403/560 | 0.6639 | 0.9332 | 1.0000 | 0.6195 |
| chor | daiin | 211/848 | 0.6627 | 0.9332 | 1.0000 | 0.6184 |
| chey | ol | 351/560 | 0.6529 | 0.9332 | 1.0000 | 0.6093 |
| dain | ol | 214/560 | 0.6494 | 0.9332 | 1.0000 | 0.6060 |
| chol | dy | 395/274 | 0.6460 | 0.9332 | 1.0000 | 0.6029 |
| ar | s | 403/350 | 0.6434 | 0.9332 | 1.0000 | 0.6005 |
| lchedy | qokar | 116/159 | 0.6585 | 0.9065 | 1.0000 | 0.5969 |
| daiin | ol | 848/560 | 0.6327 | 0.9332 | 1.0000 | 0.5904 |
| lchedy | qol | 116/142 | 0.6556 | 0.8977 | 1.0000 | 0.5886 |
| lchedy | qokain | 116/279 | 0.6475 | 0.9069 | 1.0000 | 0.5873 |
| ol | shedy | 560/434 | 0.6282 | 0.9332 | 1.0000 | 0.5862 |
| chedy | ol | 506/560 | 0.6275 | 0.9332 | 1.0000 | 0.5856 |
| dal | qoky | 242/145 | 0.6301 | 0.9253 | 1.0000 | 0.5830 |
| chey | okaiin | 351/216 | 0.6223 | 0.9332 | 1.0000 | 0.5807 |
| lchedy | qokal | 116/197 | 0.6377 | 0.9069 | 1.0000 | 0.5783 |

## Structurally close / graphically close (control)

| A | B | Counts | Structural | Reliability | Norm. distance | Discovery score |
|---|---|---:|---:|---:|---:|---:|
| qokedy | qokeedy | 276/307 | 0.8446 | 0.9332 | 0.1429 | 0.1126 |
| qokaiin | qokain | 265/279 | 0.7983 | 0.9332 | 0.1429 | 0.1064 |
| chedy | shedy | 506/434 | 0.8446 | 0.9332 | 0.2000 | 0.1576 |
| qokeedy | qokeey | 307/307 | 0.7825 | 0.9332 | 0.1429 | 0.1043 |
| qokedy | qokeey | 276/307 | 0.7642 | 0.9332 | 0.1667 | 0.1189 |
| okeey | qokeey | 184/307 | 0.7171 | 0.9332 | 0.1667 | 0.1115 |
| shedy | sheedy | 434/83 | 0.7556 | 0.8785 | 0.1667 | 0.1106 |
| qokeey | qokey | 307/108 | 0.7340 | 0.9010 | 0.1667 | 0.1102 |
| chedy | chey | 506/351 | 0.7310 | 0.9332 | 0.2000 | 0.1364 |
| daiin | dain | 848/214 | 0.7266 | 0.9332 | 0.2000 | 0.1356 |
| shedy | shey | 434/278 | 0.7203 | 0.9332 | 0.2000 | 0.1344 |
| qokal | qokar | 197/159 | 0.7080 | 0.9327 | 0.2000 | 0.1321 |
| okedy | qokedy | 120/276 | 0.6941 | 0.9098 | 0.1667 | 0.1052 |
| cheey | chey | 183/351 | 0.7013 | 0.9332 | 0.2000 | 0.1309 |
| chol | chor | 395/211 | 0.7479 | 0.9332 | 0.2500 | 0.1745 |
| qokeedy | qoteedy | 307/77 | 0.6960 | 0.8717 | 0.1429 | 0.0867 |
| qokedy | qokey | 276/108 | 0.6925 | 0.9010 | 0.1667 | 0.1040 |
| chey | shey | 351/278 | 0.7369 | 0.9332 | 0.2500 | 0.1719 |
| aiin | ain | 504/113 | 0.7462 | 0.9048 | 0.2500 | 0.1688 |
| shedy | sheey | 434/149 | 0.6816 | 0.9275 | 0.2000 | 0.1264 |

## Frequency control

| Minimum count for both tokens | Pairs | Pearson | Spearman | Distant candidates |
|---:|---:|---:|---:|---:|
| 10 | 144991 | 0.14953 | 0.12206 | 26 |
| 20 | 41905 | 0.18888 | 0.14457 | 26 |
| 50 | 6441 | 0.37323 | 0.31310 | 26 |
| 100 | 1830 | 0.49979 | 0.43231 | 24 |

## Families

The same explicit edge criteria form connected components. There are 9 graphemic-structural families and 5 structural-distant families. Components can contain token pairs that are connected only through intermediate tokens; inspect the saved edge list before interpreting a whole component. The neutral term “family” denotes a graph component only.

### Graphemic-structural families

- Family 1 (3 tokens): aiiin, aiin, ain
- Family 2 (9 tokens): chdy, chedy, cheedy, cheey, chey, shedy, sheedy, sheey, shey
- Family 3 (5 tokens): chol, chor, cthol, cthor, shol
- Family 4 (2 tokens): chy, cthy
- Family 5 (2 tokens): daiin, dain
- Family 6 (2 tokens): okar, otar
- Family 7 (9 tokens): okedy, okeedy, okeey, qokedy, qokeedy, qokeey, qokey, qotedy, qoteedy
- Family 8 (3 tokens): qokaiin, qokain, qotain
- Family 9 (3 tokens): qokal, qokar, qotal

### Structural-distant families

- Family 1 (11 tokens): aiin, ain, al, ar, chey, dal, dar, okaiin, okain, ol, y
- Family 2 (9 tokens): chedy, lchedy, qokaiin, qokain, qokar, qokedy, qokeey, qol, qotain
- Family 3 (5 tokens): chol, chor, cthy, daiin, dol
- Family 4 (2 tokens): okar, otain
- Family 5 (3 tokens): or, r, s

## Graphemic-distance distribution

The bin counts above are the full empirical distribution of normalized edit distance. Edit operations are performed on grapheme sequences: `@NNN;` is one grapheme, `?` is one unknown grapheme, and no signs are deleted or normalized.

## Limitations

- Pair rows are dependent because each token occurs in many pairs.
- Reliability and frequency reduce, but cannot eliminate, instability of sparse profiles.
- Levenshtein distance assigns equal cost to every insertion, deletion, and substitution and contains no palaeographic model.
- Connected components are threshold-sensitive descriptive groups, not linguistic categories.
- Correlation does not establish that one coordinate causes the other.
- The analysis makes no claim about language, morphology, commands, operators, or cipher mechanisms.
