# Structural–graphemic analysis

This analysis compares two independent coordinates. The existing structural similarity was copied unchanged from `raw_similarity`; spelling was used only for the graphemic metrics.

## Scope

- Tokens: 660
- Pairs: 217470
- Pearson correlation (graphemic similarity vs structural similarity): 0.040656
- Spearman correlation: 0.006138

Pair observations share tokens and are not independent; the correlations are descriptive and ordinary pairwise p-values are intentionally not reported.

## Structural similarity by normalized grapheme-distance bin

| Distance | Pairs | Mean | Median | P90 | P95 |
|---|---:|---:|---:|---:|---:|
| 0.0–0.1 | 0 | 0.0000 | 0.0000 | 0.0000 | 0.0000 |
| 0.1–0.2 | 1401 | 0.3271 | 0.3163 | 0.4755 | 0.5213 |
| 0.2–0.3 | 4630 | 0.2554 | 0.2503 | 0.3360 | 0.3742 |
| 0.3–0.4 | 0 | 0.0000 | 0.0000 | 0.0000 | 0.0000 |
| 0.4–0.5 | 26119 | 0.2447 | 0.2430 | 0.3187 | 0.3429 |
| 0.5–0.6 | 84540 | 0.2426 | 0.2421 | 0.3157 | 0.3388 |
| 0.6–0.7 | 0 | 0.0000 | 0.0000 | 0.0000 | 0.0000 |
| 0.7–0.8 | 100780 | 0.2448 | 0.2437 | 0.3173 | 0.3406 |
| 0.8–0.9 | 0 | 0.0000 | 0.0000 | 0.0000 | 0.0000 |
| 0.9–1.0 | 0 | 0.0000 | 0.0000 | 0.0000 | 0.0000 |

## Selection and percentile view

Configurable distant selection uses structural similarity ≥ 0.650, reliability ≥ 0.700, and normalized grapheme distance ≥ 0.600. It yields 0 pairs. Ranking is `structural_similarity × reliability × normalized_grapheme_distance`; this score is neither a probability nor statistical significance.

As a threshold-free companion view, corpus cutoffs are structural P95 = 0.3424, reliability P75 = 0.6013, and grapheme-distance P75 = 0.7143. Their intersection yields 3971 pairs, ranked by the same transparent score. These percentiles describe ranking coordinates and do not define the class.

## Structurally close / graphically distant

| A | B | Counts | Structural | Reliability | Norm. distance | Discovery score |
|---|---|---:|---:|---:|---:|---:|

## Percentile-ranked distant view

| A | B | Counts | Structural | Reliability | Norm. distance | Discovery score |
|---|---|---:|---:|---:|---:|---:|
| x019936 | x041112 | 122/119 | 0.6252 | 0.8618 | 0.7143 | 0.3849 |
| x018872 | x025395 | 151/169 | 0.5748 | 0.9159 | 0.7143 | 0.3761 |
| x018872 | x025396 | 151/133 | 0.5851 | 0.8962 | 0.7143 | 0.3745 |
| x018872 | x025393 | 151/194 | 0.5714 | 0.9159 | 0.7143 | 0.3738 |
| x018872 | x025394 | 151/189 | 0.5657 | 0.9159 | 0.7143 | 0.3701 |
| x018873 | x025394 | 138/189 | 0.5661 | 0.9062 | 0.7143 | 0.3664 |
| x018873 | x025392 | 138/250 | 0.5611 | 0.9062 | 0.7143 | 0.3632 |
| x018875 | x025392 | 97/250 | 0.5789 | 0.8665 | 0.7143 | 0.3583 |
| x018873 | x025395 | 138/169 | 0.5496 | 0.9062 | 0.7143 | 0.3558 |
| x018873 | x025396 | 138/133 | 0.5499 | 0.8867 | 0.7143 | 0.3483 |
| x018875 | x025393 | 97/194 | 0.5618 | 0.8665 | 0.7143 | 0.3477 |
| x018874 | x025395 | 94/169 | 0.5631 | 0.8629 | 0.7143 | 0.3471 |
| x018874 | x025392 | 94/250 | 0.5612 | 0.8629 | 0.7143 | 0.3459 |
| x018874 | x025393 | 94/194 | 0.5572 | 0.8629 | 0.7143 | 0.3434 |
| x002336 | x018872 | 74/151 | 0.5780 | 0.8298 | 0.7143 | 0.3426 |
| x000104 | x037858 | 238/367 | 0.5200 | 0.9221 | 0.7143 | 0.3425 |
| x000104 | x037859 | 238/339 | 0.5172 | 0.9221 | 0.7143 | 0.3407 |
| x019936 | x041115 | 122/79 | 0.5790 | 0.8165 | 0.7143 | 0.3377 |
| x018875 | x025396 | 97/133 | 0.5495 | 0.8483 | 0.7143 | 0.3330 |
| x000104 | x037857 | 238/483 | 0.5054 | 0.9221 | 0.7143 | 0.3329 |

## Structurally close / graphically close (control)

| A | B | Counts | Structural | Reliability | Norm. distance | Discovery score |
|---|---|---:|---:|---:|---:|---:|

## Frequency control

| Minimum count for both tokens | Pairs | Pearson | Spearman | Distant candidates |
|---:|---:|---:|---:|---:|
| 10 | 217470 | 0.04066 | 0.00614 | 0 |
| 20 | 52975 | 0.11300 | 0.03500 | 0 |
| 50 | 7750 | 0.25678 | 0.09127 | 0 |
| 100 | 1225 | 0.44267 | 0.20580 | 0 |

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
