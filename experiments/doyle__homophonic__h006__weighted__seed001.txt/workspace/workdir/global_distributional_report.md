# Global distributional regime discovery

The corpus is treated as one continuous token sequence. Discovery uses no folio, page, illustration, Currier, hand, or section metadata. Full window token distributions retain all probability mass; no RARE bucket is used.

## Continuous change profile

Adjacent-window Jensen–Shannon distance is the primary result. Weighted overlap and cosine similarity are retained as companion diagnostics. Local peaks, broad transitions, and low/high-variation intervals are descriptive labels rather than assumed true boundaries.

## Stable boundaries

| position | scale support | mean jump | uncertainty |
|---:|---:|---:|---:|
| 18915 | 4/5 | 0.09274 | 25.0 |
| 20258 | 4/5 | 0.09002 | 12.5 |
| 7015 | 4/5 | 0.08598 | 35.0 |
| 16910 | 4/5 | 0.08398 | 30.0 |
| 23723 | 4/5 | 0.08307 | 27.5 |
| 6288 | 4/5 | 0.08260 | 37.5 |
| 11798 | 4/5 | 0.08246 | 7.5 |
| 6810 | 4/5 | 0.08166 | 10.0 |
| 20100 | 4/5 | 0.08117 | 20.0 |
| 18428 | 4/5 | 0.08035 | 27.5 |
| 24483 | 4/5 | 0.07450 | 42.5 |
| 17427 | 3/5 | 0.09512 | 33.3 |
| 10227 | 3/5 | 0.09240 | 23.3 |
| 23170 | 3/5 | 0.09165 | 30.0 |
| 4277 | 3/5 | 0.09127 | 26.7 |
| 33713 | 3/5 | 0.09125 | 13.3 |
| 23653 | 3/5 | 0.09098 | 26.7 |
| 29350 | 3/5 | 0.09090 | 10.0 |
| 17327 | 3/5 | 0.09072 | 33.3 |
| 18583 | 3/5 | 0.09059 | 23.3 |
| 23443 | 3/5 | 0.09058 | 23.3 |
| 17148 | 3/5 | 0.09049 | 8.3 |
| 2710 | 3/5 | 0.09038 | 10.0 |
| 10547 | 3/5 | 0.09035 | 26.7 |
| 20840 | 3/5 | 0.09014 | 30.0 |
| 6153 | 3/5 | 0.09002 | 16.7 |
| 3940 | 3/5 | 0.08977 | 10.0 |
| 38620 | 3/5 | 0.08957 | 30.0 |
| 20557 | 3/5 | 0.08925 | 43.3 |
| 18723 | 3/5 | 0.08910 | 36.7 |

Candidates combine threshold peaks, PELT on the distributional JS-jump series, and binary segmentation. Cross-scale matches use ±0.5 × the smaller window size. Multiple detector hits at one scale count as one scale vote.

## Clustering diagnostics

Hierarchical single-link clustering and JS-distance k-medoids are unconstrained, so distant windows may share a regime. Contiguous binary segmentation is reported separately. K=2..15 is swept without selecting K by interpretability. To bound quadratic fitting, up to 200 sequence-wide uniformly spaced windows are used for fit and silhouette diagnostics; every window is then assigned to a fitted distribution centroid. Cluster sizes, transitions, fragmentation, and the assignment table cover all windows. These clusterings are diagnostics and do not replace the continuous profile.
