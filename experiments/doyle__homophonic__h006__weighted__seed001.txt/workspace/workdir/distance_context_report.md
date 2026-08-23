# Distance-specific context analysis

This report describes formal token-sequence distributions and does not assign meanings to tokens. The main analysis treats all non-empty lines as one continuous sequence; pages are not used. Physical lines are retained only for the line-bounded control.

## Parameters and corpus

Corpus tokens: 43713; target pairs: 1; maximum exact distance: 20; minimum observations: 30; requested primary mode: `continuous`. Similarity at every distance is computed separately, never from a pooled window. Frequency-matched baseline pairs have unordered counts within a factor of 2. Effective support is `exp(Shannon entropy)` using natural-log units.

## Right-context baseline

| Distance | Pairs | Median JS | P90 | P95 |
|---:|---:|---:|---:|---:|
| +1 | 13312 | 0.0350 | 0.1780 | 0.2388 |
| +2 | 13312 | 0.0616 | 0.1348 | 0.1590 |
| +3 | 13312 | 0.0811 | 0.1608 | 0.1882 |
| +4 | 13312 | 0.0832 | 0.1618 | 0.1904 |
| +5 | 13312 | 0.0826 | 0.1611 | 0.1882 |
| +6 | 13312 | 0.0862 | 0.1631 | 0.1921 |
| +7 | 13312 | 0.0802 | 0.1574 | 0.1869 |
| +8 | 13312 | 0.0814 | 0.1604 | 0.1884 |
| +9 | 13312 | 0.0850 | 0.1592 | 0.1875 |
| +10 | 13312 | 0.0812 | 0.1581 | 0.1870 |
| +11 | 13312 | 0.0851 | 0.1610 | 0.1884 |
| +12 | 13312 | 0.0846 | 0.1617 | 0.1919 |
| +13 | 13312 | 0.0814 | 0.1589 | 0.1890 |
| +14 | 13312 | 0.0842 | 0.1566 | 0.1833 |
| +15 | 13312 | 0.0834 | 0.1603 | 0.1915 |
| +16 | 13312 | 0.0812 | 0.1583 | 0.1885 |
| +17 | 13312 | 0.0800 | 0.1561 | 0.1851 |
| +18 | 13312 | 0.0805 | 0.1599 | 0.1865 |
| +19 | 13312 | 0.0846 | 0.1598 | 0.1888 |
| +20 | 13312 | 0.0827 | 0.1576 | 0.1861 |

## Long-range structural persistence ranking

`persistence_1_5` is the transparent mean of baseline percentile ranks at exact distances +1 through +5. It is a ranking aid, not a replacement for the profiles.

| Pair | Persistence 1–5 | Mean JS 1–5 | 6–10 | 11–20 |
|---|---:|---:|---:|---:|
| `x014952` / `x030834` | 93.14 | 0.1911 | 0.1699 | 0.1721 |

## Target profiles, directionality, and boundary sensitivity

### `x014952` / `x030834`

Counts: 149/155. Right mean JS 1–5: 0.1911; left mean JS 1–5: 0.2453. Sequence suffix similarities n=2/n=3: 0.0000/0.0000.

| d | Right JS | Percentile | Left JS | Continuous | Line-bounded | Difference | Continuous obs A/B | Line obs A/B | Reliable C/L |
|---:|---:|---:|---:|---:|---:|---:|---:|---:|:---:|
| 1 | 0.2860 | P96.9 | 0.4689 | 0.2860 | 0.2994 | -0.0135 | 149/155 | 140/145 | true/true |
| 2 | 0.1638 | P95.5 | 0.2368 | 0.1638 | 0.1388 | +0.0250 | 149/155 | 130/133 | true/true |
| 3 | 0.1702 | P91.9 | 0.1563 | 0.1702 | 0.1301 | +0.0401 | 149/155 | 120/123 | true/true |
| 4 | 0.1506 | P86.8 | 0.2303 | 0.1506 | 0.1239 | +0.0266 | 149/155 | 108/110 | true/true |
| 5 | 0.1848 | P94.6 | 0.1340 | 0.1848 | 0.1424 | +0.0424 | 149/155 | 98/97 | true/true |
| 6 | 0.1726 | P92.0 | 0.2070 | 0.1726 | 0.0902 | +0.0824 | 149/155 | 87/86 | true/true |
| 7 | 0.1605 | P90.6 | 0.1852 | 0.1605 | 0.0332 | +0.1272 | 149/155 | 72/71 | true/true |
| 8 | 0.1315 | P80.8 | 0.1840 | 0.1315 | 0.0413 | +0.0902 | 149/155 | 58/57 | true/true |
| 9 | 0.2138 | P97.4 | 0.1735 | 0.2138 | 0.0690 | +0.1448 | 149/155 | 44/43 | true/true |
| 10 | 0.1710 | P92.7 | 0.1986 | 0.1710 | 0.0834 | +0.0876 | 149/155 | 38/34 | true/true |
| 11 | 0.2007 | P96.4 | 0.2084 | 0.2007 | 0.0000 | +0.2007 | 149/155 | 21/25 | true/false |
| 12 | 0.1955 | P95.5 | 0.1718 | 0.1955 | 0.0000 | +0.1955 | 149/155 | 14/11 | true/false |
| 13 | 0.1544 | P88.9 | 0.1721 | 0.1544 | 0.1680 | -0.0135 | 149/155 | 7/5 | true/false |
| 14 | 0.1807 | P94.7 | 0.1230 | 0.1807 | 0.0000 | +0.1807 | 149/155 | 2/2 | true/false |
| 15 | 0.1897 | P94.8 | 0.1504 | 0.1897 | 0.0000 | +0.1897 | 149/155 | 1/1 | true/false |
| 16 | 0.1551 | P89.3 | 0.2161 | 0.1551 | 0.0000 | +0.1551 | 149/155 | 1/0 | true/false |
| 17 | 0.1913 | P95.7 | 0.1712 | 0.1913 | 0.0000 | +0.1913 | 149/155 | 0/0 | true/false |
| 18 | 0.1970 | P96.2 | 0.1730 | 0.1970 | 0.0000 | +0.1970 | 149/155 | 0/0 | true/false |
| 19 | 0.1448 | P85.6 | 0.1938 | 0.1448 | 0.0000 | +0.1448 | 149/155 | 0/0 | true/false |
| 20 | 0.1116 | P70.4 | 0.2437 | 0.1116 | 0.0000 | +0.1116 | 149/155 | 0/0 | true/false |

## Matched negative controls

| Target | Controls | Target mean JS 1–5 | Control median mean JS 1–5 |
|---|---:|---:|---:|
| `x014952` / `x030834` | 3 | 0.1911 | 0.0950 |

## Family cohesion

Matrices and distance-specific medoids are in `family_distance_profiles.yaml`. Random controls are deterministic groups of equal size whose member frequencies match the corresponding family member within ×2.

| Family | d | Cohesion | Matched percentile | Medoid |
|---:|---:|---:|---:|---|
| 1 | 1 | 0.2860 | P95.5 | `x014952` |
| 1 | 2 | 0.1638 | P83.0 | `x014952` |
| 1 | 3 | 0.1702 | P59.0 | `x014952` |
| 1 | 4 | 0.1506 | P39.0 | `x014952` |
| 1 | 5 | 0.1848 | P72.5 | `x014952` |
| 1 | 6 | 0.1726 | P66.0 | `x014952` |
| 1 | 7 | 0.1605 | P53.0 | `x014952` |
| 1 | 8 | 0.1315 | P22.5 | `x014952` |
| 1 | 9 | 0.2138 | P90.0 | `x014952` |
| 1 | 10 | 0.1710 | P65.0 | `x014952` |
| 1 | 11 | 0.2007 | P82.0 | `x014952` |
| 1 | 12 | 0.1955 | P75.0 | `x014952` |
| 1 | 13 | 0.1544 | P39.5 | `x014952` |
| 1 | 14 | 0.1807 | P74.5 | `x014952` |
| 1 | 15 | 0.1897 | P73.0 | `x014952` |
| 1 | 16 | 0.1551 | P50.0 | `x014952` |
| 1 | 17 | 0.1913 | P81.0 | `x014952` |
| 1 | 18 | 0.1970 | P82.0 | `x014952` |
| 1 | 19 | 0.1448 | P39.0 | `x014952` |
| 1 | 20 | 0.1116 | P15.5 | `x014952` |

## Sequence-context signatures

The 2- and 3-token suffix distributions are compared separately with JS similarity, weighted overlap, support Jaccard, observation counts, effective support, and reliability in `sequence_context_pairs.yaml`.

## Limitations

Similarity at greater distance is descriptive context-similarity persistence, not physical decay. Corpus positions and pair comparisons are dependent. Frequency matching controls only a major sampling confound and is not a causal null model. Support overlap is sensitive to rare observations. A low-reliability value is retained but must not be treated as equivalent to a frequent comparison. Line-bounded results remove cross-line observations and therefore can change both distributions and sample sizes. Sequence suffixes become sparse quickly; only lengths 2 and 3 are used. No semantic interpretation is made.
