# Global distributional regime discovery

The corpus is treated as one continuous token sequence. Discovery uses no folio, page, illustration, Currier, hand, or section metadata. Full window token distributions retain all probability mass; no RARE bucket is used.

## Continuous change profile

Adjacent-window Jensen–Shannon distance is the primary result. Weighted overlap and cosine similarity are retained as companion diagnostics. Local peaks, broad transitions, and low/high-variation intervals are descriptive labels rather than assumed true boundaries.

## Stable boundaries

| position | scale support | mean jump | uncertainty |
|---:|---:|---:|---:|
| 2968 | 4/5 | 0.08980 | 47.5 |
| 21018 | 4/5 | 0.08807 | 32.5 |
| 4273 | 4/5 | 0.08750 | 27.5 |
| 17088 | 4/5 | 0.08522 | 37.5 |
| 17443 | 3/5 | 0.09686 | 33.3 |
| 27203 | 3/5 | 0.09655 | 26.7 |
| 28120 | 3/5 | 0.09636 | 0.0 |
| 23393 | 3/5 | 0.09477 | 33.3 |
| 20575 | 3/5 | 0.09396 | 35.0 |
| 23833 | 3/5 | 0.09365 | 23.3 |
| 10210 | 3/5 | 0.09357 | 10.0 |
| 17147 | 3/5 | 0.09352 | 6.7 |
| 14990 | 3/5 | 0.09326 | 40.0 |
| 17313 | 3/5 | 0.09318 | 13.3 |
| 29343 | 3/5 | 0.09306 | 6.7 |
| 25017 | 3/5 | 0.09288 | 36.7 |
| 5240 | 3/5 | 0.09266 | 10.0 |
| 35277 | 3/5 | 0.09223 | 26.7 |
| 14377 | 3/5 | 0.09208 | 36.7 |
| 2273 | 3/5 | 0.09200 | 46.7 |
| 20067 | 3/5 | 0.09182 | 33.3 |
| 23257 | 3/5 | 0.09171 | 23.3 |
| 26080 | 3/5 | 0.09160 | 40.0 |
| 12967 | 3/5 | 0.09147 | 16.7 |
| 31070 | 3/5 | 0.09144 | 30.0 |
| 11993 | 3/5 | 0.09132 | 13.3 |
| 20260 | 3/5 | 0.09122 | 10.0 |
| 17670 | 3/5 | 0.09114 | 20.0 |
| 29927 | 3/5 | 0.09102 | 33.3 |
| 25917 | 3/5 | 0.09096 | 36.7 |

Candidates combine threshold peaks, PELT on the distributional JS-jump series, and binary segmentation. Cross-scale matches use ±0.5 × the smaller window size. Multiple detector hits at one scale count as one scale vote.

## Clustering diagnostics

Hierarchical single-link clustering and JS-distance k-medoids are unconstrained, so distant windows may share a regime. Contiguous binary segmentation is reported separately. K=2..15 is swept without selecting K by interpretability. To bound quadratic fitting, up to 200 sequence-wide uniformly spaced windows are used for fit and silhouette diagnostics; every window is then assigned to a fitted distribution centroid. Cluster sizes, transitions, fragmentation, and the assignment table cover all windows. These clusterings are diagnostics and do not replace the continuous profile.
