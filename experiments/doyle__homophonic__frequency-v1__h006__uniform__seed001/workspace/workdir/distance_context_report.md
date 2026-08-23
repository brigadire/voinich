# Distance-specific context analysis

This report describes formal token-sequence distributions and does not assign meanings to tokens. The main analysis treats all non-empty lines as one continuous sequence; pages are not used. Physical lines are retained only for the line-bounded control.

## Parameters and corpus

Corpus tokens: 43713; target pairs: 0; maximum exact distance: 20; minimum observations: 30; requested primary mode: `continuous`. Similarity at every distance is computed separately, never from a pooled window. Frequency-matched baseline pairs have unordered counts within a factor of 2. Effective support is `exp(Shannon entropy)` using natural-log units.

## Right-context baseline

| Distance | Pairs | Median JS | P90 | P95 |
|---:|---:|---:|---:|---:|
| +1 | 12754 | 0.0425 | 0.1910 | 0.2567 |
| +2 | 12754 | 0.0618 | 0.1384 | 0.1720 |
| +3 | 12754 | 0.0807 | 0.1695 | 0.2048 |
| +4 | 12754 | 0.0814 | 0.1701 | 0.2051 |
| +5 | 12754 | 0.0845 | 0.1674 | 0.2012 |
| +6 | 12754 | 0.0845 | 0.1746 | 0.2084 |
| +7 | 12754 | 0.0793 | 0.1710 | 0.2051 |
| +8 | 12754 | 0.0837 | 0.1728 | 0.2061 |
| +9 | 12754 | 0.0820 | 0.1687 | 0.2028 |
| +10 | 12754 | 0.0817 | 0.1706 | 0.2072 |
| +11 | 12754 | 0.0856 | 0.1700 | 0.2016 |
| +12 | 12754 | 0.0838 | 0.1723 | 0.2047 |
| +13 | 12754 | 0.0791 | 0.1672 | 0.2058 |
| +14 | 12754 | 0.0813 | 0.1670 | 0.1999 |
| +15 | 12754 | 0.0817 | 0.1764 | 0.2090 |
| +16 | 12754 | 0.0818 | 0.1725 | 0.2048 |
| +17 | 12754 | 0.0841 | 0.1676 | 0.2030 |
| +18 | 12754 | 0.0800 | 0.1685 | 0.2022 |
| +19 | 12754 | 0.0850 | 0.1734 | 0.2081 |
| +20 | 12754 | 0.0802 | 0.1709 | 0.2075 |

## Long-range structural persistence ranking

`persistence_1_5` is the transparent mean of baseline percentile ranks at exact distances +1 through +5. It is a ranking aid, not a replacement for the profiles.

| Pair | Persistence 1–5 | Mean JS 1–5 | 6–10 | 11–20 |
|---|---:|---:|---:|---:|

## Target profiles, directionality, and boundary sensitivity

## Matched negative controls

| Target | Controls | Target mean JS 1–5 | Control median mean JS 1–5 |
|---|---:|---:|---:|

## Family cohesion

Matrices and distance-specific medoids are in `family_distance_profiles.yaml`. Random controls are deterministic groups of equal size whose member frequencies match the corresponding family member within ×2.

| Family | d | Cohesion | Matched percentile | Medoid |
|---:|---:|---:|---:|---|

## Sequence-context signatures

The 2- and 3-token suffix distributions are compared separately with JS similarity, weighted overlap, support Jaccard, observation counts, effective support, and reliability in `sequence_context_pairs.yaml`.

## Limitations

Similarity at greater distance is descriptive context-similarity persistence, not physical decay. Corpus positions and pair comparisons are dependent. Frequency matching controls only a major sampling confound and is not a causal null model. Support overlap is sensitive to rare observations. A low-reliability value is retained but must not be treated as equivalent to a frequent comparison. Line-bounded results remove cross-line observations and therefore can change both distributions and sample sizes. Sequence suffixes become sparse quickly; only lengths 2 and 3 are used. No semantic interpretation is made.
