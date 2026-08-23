# Global distributional regime discovery

The corpus is treated as one continuous token sequence. Discovery uses no folio, page, illustration, Currier, hand, or section metadata. Full window token distributions retain all probability mass; no RARE bucket is used.

## Continuous change profile

Adjacent-window Jensen–Shannon distance is the primary result. Weighted overlap and cosine similarity are retained as companion diagnostics. Local peaks, broad transitions, and low/high-variation intervals are descriptive labels rather than assumed true boundaries.

## Stable boundaries

| position | scale support | mean jump | uncertainty |
|---:|---:|---:|---:|
| 16405 | 4/5 | 0.09225 | 15.0 |
| 23818 | 4/5 | 0.09002 | 37.5 |
| 10028 | 4/5 | 0.08747 | 27.5 |
| 11778 | 4/5 | 0.08555 | 27.5 |
| 35120 | 4/5 | 0.08459 | 30.0 |
| 3103 | 4/5 | 0.08322 | 27.5 |
| 22883 | 4/5 | 0.08251 | 42.5 |
| 3003 | 4/5 | 0.08248 | 7.5 |
| 35805 | 3/5 | 0.09758 | 25.0 |
| 23293 | 3/5 | 0.09518 | 33.3 |
| 30783 | 3/5 | 0.09430 | 16.7 |
| 16163 | 3/5 | 0.09378 | 23.3 |
| 2853 | 3/5 | 0.09365 | 16.7 |
| 23393 | 3/5 | 0.09339 | 33.3 |
| 23203 | 3/5 | 0.09328 | 36.7 |
| 3857 | 3/5 | 0.09328 | 6.7 |
| 10547 | 3/5 | 0.09316 | 26.7 |
| 17947 | 3/5 | 0.09213 | 26.7 |
| 17147 | 3/5 | 0.09183 | 6.7 |
| 17673 | 3/5 | 0.09176 | 36.7 |
| 23597 | 3/5 | 0.09171 | 26.7 |
| 25137 | 3/5 | 0.09170 | 36.7 |
| 30927 | 3/5 | 0.09148 | 36.7 |
| 29343 | 3/5 | 0.09127 | 6.7 |
| 20247 | 3/5 | 0.09113 | 26.7 |
| 18583 | 3/5 | 0.09112 | 23.3 |
| 33220 | 3/5 | 0.09098 | 20.0 |
| 20483 | 3/5 | 0.09096 | 33.3 |
| 21913 | 3/5 | 0.09088 | 36.7 |
| 24913 | 3/5 | 0.09088 | 26.7 |

Candidates combine threshold peaks, PELT on the distributional JS-jump series, and binary segmentation. Cross-scale matches use ±0.5 × the smaller window size. Multiple detector hits at one scale count as one scale vote.

## Clustering diagnostics

Hierarchical single-link clustering and JS-distance k-medoids are unconstrained, so distant windows may share a regime. Contiguous binary segmentation is reported separately. K=2..15 is swept without selecting K by interpretability. To bound quadratic fitting, up to 200 sequence-wide uniformly spaced windows are used for fit and silhouette diagnostics; every window is then assigned to a fitted distribution centroid. Cluster sizes, transitions, fragmentation, and the assignment table cover all windows. These clusterings are diagnostics and do not replace the continuous profile.
