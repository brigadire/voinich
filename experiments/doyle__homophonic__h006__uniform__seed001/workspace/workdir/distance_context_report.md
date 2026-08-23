# Distance-specific context analysis

This report describes formal token-sequence distributions and does not assign meanings to tokens. The main analysis treats all non-empty lines as one continuous sequence; pages are not used. Physical lines are retained only for the line-bounded control.

## Parameters and corpus

Corpus tokens: 43713; target pairs: 0; maximum exact distance: 20; minimum observations: 30; requested primary mode: `continuous`. Similarity at every distance is computed separately, never from a pooled window. Frequency-matched baseline pairs have unordered counts within a factor of 2. Effective support is `exp(Shannon entropy)` using natural-log units.

## Right-context baseline

| Distance | Pairs | Median JS | P90 | P95 |
|---:|---:|---:|---:|---:|
| +1 | 12632 | 0.0378 | 0.1841 | 0.2473 |
| +2 | 12632 | 0.0534 | 0.1202 | 0.1496 |
| +3 | 12632 | 0.0736 | 0.1504 | 0.1762 |
| +4 | 12632 | 0.0739 | 0.1511 | 0.1770 |
| +5 | 12632 | 0.0753 | 0.1474 | 0.1722 |
| +6 | 12632 | 0.0746 | 0.1535 | 0.1795 |
| +7 | 12632 | 0.0721 | 0.1516 | 0.1767 |
| +8 | 12632 | 0.0752 | 0.1512 | 0.1764 |
| +9 | 12632 | 0.0714 | 0.1512 | 0.1747 |
| +10 | 12632 | 0.0748 | 0.1530 | 0.1814 |
| +11 | 12632 | 0.0738 | 0.1490 | 0.1712 |
| +12 | 12632 | 0.0769 | 0.1520 | 0.1753 |
| +13 | 12632 | 0.0706 | 0.1497 | 0.1759 |
| +14 | 12632 | 0.0710 | 0.1443 | 0.1706 |
| +15 | 12632 | 0.0735 | 0.1552 | 0.1827 |
| +16 | 12632 | 0.0755 | 0.1501 | 0.1731 |
| +17 | 12632 | 0.0718 | 0.1474 | 0.1706 |
| +18 | 12632 | 0.0727 | 0.1478 | 0.1724 |
| +19 | 12632 | 0.0749 | 0.1521 | 0.1776 |
| +20 | 12632 | 0.0728 | 0.1498 | 0.1760 |

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
