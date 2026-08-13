# Structural–graphemic analysis

This analysis compares two independent coordinates. The existing structural similarity was copied unchanged from `raw_similarity`; spelling was used only for the graphemic metrics.

## Scope

- Tokens: 537
- Pairs: 143916
- Pearson correlation (graphemic similarity vs structural similarity): 0.149328
- Spearman correlation: 0.121699

Pair observations share tokens and are not independent; the correlations are descriptive and ordinary pairwise p-values are intentionally not reported.

## Structural similarity by normalized grapheme-distance bin

| Distance | Pairs | Mean | Median | P90 | P95 |
|---|---:|---:|---:|---:|---:|
| 0.0–0.1 | 0 | 0.0000 | 0.0000 | 0.0000 | 0.0000 |
| 0.1–0.2 | 571 | 0.3396 | 0.3225 | 0.4764 | 0.5659 |
| 0.2–0.3 | 1554 | 0.3247 | 0.3063 | 0.4699 | 0.5299 |
| 0.3–0.4 | 2277 | 0.3066 | 0.2968 | 0.4433 | 0.4855 |
| 0.4–0.5 | 4003 | 0.2964 | 0.2887 | 0.4221 | 0.4714 |
| 0.5–0.6 | 11330 | 0.2807 | 0.2768 | 0.3944 | 0.4383 |
| 0.6–0.7 | 18492 | 0.2732 | 0.2706 | 0.3898 | 0.4302 |
| 0.7–0.8 | 13756 | 0.2645 | 0.2634 | 0.3676 | 0.4060 |
| 0.8–0.9 | 39424 | 0.2554 | 0.2554 | 0.3595 | 0.3944 |
| 0.9–1.0 | 52509 | 0.2540 | 0.2539 | 0.3685 | 0.4077 |

## Selection and percentile view

Configurable distant selection uses structural similarity ≥ 0.650, reliability ≥ 0.700, and normalized grapheme distance ≥ 0.600. It yields 27 pairs. Ranking is `structural_similarity × reliability × normalized_grapheme_distance`; this score is neither a probability nor statistical significance.

As a threshold-free companion view, corpus cutoffs are structural P95 = 0.4161, reliability P75 = 0.6795, and grapheme-distance P75 = 1.0000. Their intersection yields 2143 pairs, ranked by the same transparent score. These percentiles describe ranking coordinates and do not define the class.

## Structurally close / graphically distant

| A | B | Counts | Structural | Reliability | Norm. distance | Discovery score |
|---|---|---:|---:|---:|---:|---:|
| or | s | 388/350 | 0.7594 | 0.9336 | 1.0000 | 0.7090 |
| chol | daiin | 395/847 | 0.7206 | 0.9336 | 1.0000 | 0.6728 |
| r | s | 168/350 | 0.7010 | 0.9336 | 1.0000 | 0.6545 |
| ol | y | 557/304 | 0.6787 | 0.9336 | 1.0000 | 0.6336 |
| dar | ol | 323/557 | 0.6742 | 0.9336 | 1.0000 | 0.6295 |
| ar | ol | 402/557 | 0.6657 | 0.9336 | 1.0000 | 0.6215 |
| chor | daiin | 211/847 | 0.6630 | 0.9336 | 1.0000 | 0.6190 |
| chey | ol | 346/557 | 0.6528 | 0.9336 | 1.0000 | 0.6095 |
| lchedy | qokar | 116/156 | 0.6606 | 0.9053 | 1.0000 | 0.5980 |
| lchedy | qokain | 116/273 | 0.6529 | 0.9072 | 1.0000 | 0.5923 |
| lchedy | qol | 116/139 | 0.6523 | 0.8963 | 1.0000 | 0.5846 |
| okaiin | ol | 215/557 | 0.6837 | 0.9336 | 0.8333 | 0.5320 |
| okain | ol | 140/557 | 0.6815 | 0.9228 | 0.8000 | 0.5031 |
| qokaiin | qol | 264/139 | 0.7386 | 0.9222 | 0.7143 | 0.4865 |
| daiin | dol | 847/109 | 0.6674 | 0.9020 | 0.8000 | 0.4816 |
| qokain | qol | 273/139 | 0.7666 | 0.9222 | 0.6667 | 0.4713 |
| aiin | ar | 504/402 | 0.6710 | 0.9336 | 0.7500 | 0.4699 |
| chol | cthy | 395/103 | 0.6832 | 0.8972 | 0.7500 | 0.4598 |
| lchedy | qokeey | 116/306 | 0.7038 | 0.9072 | 0.6667 | 0.4257 |
| qokedy | qol | 276/139 | 0.6862 | 0.9222 | 0.6667 | 0.4219 |

## Percentile-ranked distant view

| A | B | Counts | Structural | Reliability | Norm. distance | Discovery score |
|---|---|---:|---:|---:|---:|---:|
| or | s | 388/350 | 0.7594 | 0.9336 | 1.0000 | 0.7090 |
| chol | daiin | 395/847 | 0.7206 | 0.9336 | 1.0000 | 0.6728 |
| r | s | 168/350 | 0.7010 | 0.9336 | 1.0000 | 0.6545 |
| ol | y | 557/304 | 0.6787 | 0.9336 | 1.0000 | 0.6336 |
| dar | ol | 323/557 | 0.6742 | 0.9336 | 1.0000 | 0.6295 |
| ar | ol | 402/557 | 0.6657 | 0.9336 | 1.0000 | 0.6215 |
| chor | daiin | 211/847 | 0.6630 | 0.9336 | 1.0000 | 0.6190 |
| chey | ol | 346/557 | 0.6528 | 0.9336 | 1.0000 | 0.6095 |
| dain | ol | 213/557 | 0.6481 | 0.9336 | 1.0000 | 0.6051 |
| chol | dy | 395/274 | 0.6460 | 0.9336 | 1.0000 | 0.6032 |
| ar | s | 402/350 | 0.6452 | 0.9336 | 1.0000 | 0.6023 |
| lchedy | qokar | 116/156 | 0.6606 | 0.9053 | 1.0000 | 0.5980 |
| lchedy | qokain | 116/273 | 0.6529 | 0.9072 | 1.0000 | 0.5923 |
| daiin | ol | 847/557 | 0.6333 | 0.9336 | 1.0000 | 0.5912 |
| chedy | ol | 504/557 | 0.6274 | 0.9336 | 1.0000 | 0.5857 |
| ol | shedy | 557/433 | 0.6273 | 0.9336 | 1.0000 | 0.5857 |
| lchedy | qol | 116/139 | 0.6523 | 0.8963 | 1.0000 | 0.5846 |
| dal | qoky | 242/145 | 0.6301 | 0.9257 | 1.0000 | 0.5833 |
| chey | okaiin | 346/215 | 0.6247 | 0.9336 | 1.0000 | 0.5832 |
| lchedy | qokal | 116/197 | 0.6377 | 0.9072 | 1.0000 | 0.5785 |

## Structurally close / graphically close (control)

| A | B | Counts | Structural | Reliability | Norm. distance | Discovery score |
|---|---|---:|---:|---:|---:|---:|
| qokedy | qokeedy | 276/306 | 0.8433 | 0.9336 | 0.1429 | 0.1125 |
| qokaiin | qokain | 264/273 | 0.7971 | 0.9336 | 0.1429 | 0.1063 |
| chedy | shedy | 504/433 | 0.8428 | 0.9336 | 0.2000 | 0.1574 |
| qokeedy | qokeey | 306/306 | 0.7797 | 0.9336 | 0.1429 | 0.1040 |
| qokedy | qokeey | 276/306 | 0.7647 | 0.9336 | 0.1667 | 0.1190 |
| qokeey | qokey | 306/105 | 0.7434 | 0.8989 | 0.1667 | 0.1114 |
| okeey | qokeey | 183/306 | 0.7143 | 0.9336 | 0.1667 | 0.1111 |
| shedy | sheedy | 433/83 | 0.7562 | 0.8786 | 0.1667 | 0.1107 |
| shedy | shey | 433/269 | 0.7248 | 0.9336 | 0.2000 | 0.1353 |
| daiin | dain | 847/213 | 0.7240 | 0.9336 | 0.2000 | 0.1352 |
| chedy | chey | 504/346 | 0.7235 | 0.9336 | 0.2000 | 0.1351 |
| qokal | qokar | 197/156 | 0.7092 | 0.9316 | 0.2000 | 0.1321 |
| okedy | qokedy | 120/276 | 0.6941 | 0.9101 | 0.1667 | 0.1053 |
| qokedy | qokey | 276/105 | 0.7018 | 0.8989 | 0.1667 | 0.1051 |
| cheey | chey | 182/346 | 0.7020 | 0.9336 | 0.2000 | 0.1311 |
| chol | chor | 395/211 | 0.7479 | 0.9336 | 0.2500 | 0.1746 |
| qokeedy | qoteedy | 306/77 | 0.6933 | 0.8718 | 0.1429 | 0.0864 |
| chey | shey | 346/269 | 0.7357 | 0.9336 | 0.2500 | 0.1717 |
| aiin | ain | 504/109 | 0.7509 | 0.9020 | 0.2500 | 0.1693 |
| shedy | sheey | 433/148 | 0.6819 | 0.9273 | 0.2000 | 0.1265 |

## Frequency control

| Minimum count for both tokens | Pairs | Pearson | Spearman | Distant candidates |
|---:|---:|---:|---:|---:|
| 10 | 143916 | 0.14933 | 0.12170 | 27 |
| 20 | 41905 | 0.18898 | 0.14484 | 27 |
| 50 | 6441 | 0.37376 | 0.31405 | 27 |
| 100 | 1830 | 0.50054 | 0.43332 | 25 |

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
