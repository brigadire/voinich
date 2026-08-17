# Global distributional regime discovery

The corpus is treated as one continuous token sequence. Discovery uses no folio, page, illustration, Currier, hand, or section metadata. Full window token distributions retain all probability mass; no RARE bucket is used.

## Continuous change profile

Adjacent-window Jensen–Shannon distance is the primary result. Weighted overlap and cosine similarity are retained as companion diagnostics. Local peaks, broad transitions, and low/high-variation intervals are descriptive labels rather than assumed true boundaries.

## Stable boundaries

| position | scale support | mean jump | uncertainty |
|---:|---:|---:|---:|
| 12504 | 5/5 | 0.08264 | 26.0 |
| 11486 | 5/5 | 0.07959 | 16.0 |
| 12230 | 4/5 | 0.09548 | 20.0 |
| 12051 | 4/5 | 0.09293 | 16.2 |
| 14499 | 4/5 | 0.09269 | 3.8 |
| 12686 | 4/5 | 0.09022 | 16.2 |
| 23960 | 4/5 | 0.08987 | 20.0 |
| 38003 | 4/5 | 0.08959 | 17.5 |
| 26898 | 4/5 | 0.08959 | 7.5 |
| 30894 | 4/5 | 0.08953 | 13.8 |
| 28159 | 4/5 | 0.08913 | 21.2 |
| 24409 | 4/5 | 0.08888 | 11.2 |
| 38094 | 4/5 | 0.08805 | 18.8 |
| 22298 | 4/5 | 0.08794 | 22.5 |
| 25358 | 4/5 | 0.08763 | 22.5 |
| 12178 | 4/5 | 0.08745 | 27.5 |
| 38186 | 4/5 | 0.08742 | 13.8 |
| 14205 | 4/5 | 0.08740 | 5.0 |
| 28114 | 4/5 | 0.08701 | 26.2 |
| 36386 | 4/5 | 0.08670 | 13.8 |
| 14284 | 4/5 | 0.08665 | 16.2 |
| 26506 | 4/5 | 0.08638 | 13.8 |
| 22185 | 4/5 | 0.08569 | 15.0 |
| 10703 | 4/5 | 0.08474 | 7.5 |
| 26593 | 4/5 | 0.08467 | 12.5 |
| 9091 | 4/5 | 0.08378 | 11.2 |
| 36886 | 4/5 | 0.08284 | 13.8 |
| 24694 | 4/5 | 0.08216 | 18.8 |
| 11803 | 4/5 | 0.08123 | 7.5 |
| 31591 | 4/5 | 0.07669 | 16.2 |

Candidates combine threshold peaks, PELT on the distributional JS-jump series, and binary segmentation. Cross-scale matches use ±0.5 × the smaller window size. Multiple detector hits at one scale count as one scale vote.

## Clustering diagnostics

Hierarchical single-link clustering and JS-distance k-medoids are unconstrained, so distant windows may share a regime. Contiguous binary segmentation is reported separately. K=2..15 is swept without selecting K by interpretability. To bound quadratic fitting, up to 200 sequence-wide uniformly spaced windows are used for fit and silhouette diagnostics; every window is then assigned to a fitted distribution centroid. Cluster sizes, transitions, fragmentation, and the assignment table cover all windows. These clusterings are diagnostics and do not replace the continuous profile.
