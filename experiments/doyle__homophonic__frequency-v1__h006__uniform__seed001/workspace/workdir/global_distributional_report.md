# Global distributional regime discovery

The corpus is treated as one continuous token sequence. Discovery uses no folio, page, illustration, Currier, hand, or section metadata. Full window token distributions retain all probability mass; no RARE bucket is used.

## Continuous change profile

Adjacent-window Jensen–Shannon distance is the primary result. Weighted overlap and cosine similarity are retained as companion diagnostics. Local peaks, broad transitions, and low/high-variation intervals are descriptive labels rather than assumed true boundaries.

## Stable boundaries

| position | scale support | mean jump | uncertainty |
|---:|---:|---:|---:|
| 10523 | 4/5 | 0.08567 | 27.5 |
| 23808 | 4/5 | 0.08529 | 22.5 |
| 18003 | 4/5 | 0.08502 | 17.5 |
| 22898 | 4/5 | 0.08408 | 7.5 |
| 38615 | 4/5 | 0.08311 | 35.0 |
| 5733 | 4/5 | 0.08265 | 32.5 |
| 2638 | 4/5 | 0.08248 | 37.5 |
| 21623 | 4/5 | 0.08184 | 27.5 |
| 6800 | 4/5 | 0.08160 | 20.0 |
| 5883 | 4/5 | 0.08016 | 22.5 |
| 24175 | 4/5 | 0.08011 | 25.0 |
| 29670 | 4/5 | 0.07677 | 30.0 |
| 38842 | 3/5 | 0.09373 | 21.7 |
| 23300 | 3/5 | 0.09302 | 20.0 |
| 23713 | 3/5 | 0.09198 | 13.3 |
| 17663 | 3/5 | 0.09189 | 16.7 |
| 23203 | 3/5 | 0.09176 | 36.7 |
| 26063 | 3/5 | 0.09129 | 36.7 |
| 23393 | 3/5 | 0.09110 | 33.3 |
| 10240 | 3/5 | 0.09098 | 20.0 |
| 21043 | 3/5 | 0.09089 | 6.7 |
| 11597 | 3/5 | 0.09089 | 6.7 |
| 17103 | 3/5 | 0.09089 | 6.7 |
| 2853 | 3/5 | 0.09078 | 16.7 |
| 3903 | 3/5 | 0.09072 | 26.7 |
| 23047 | 3/5 | 0.09056 | 26.7 |
| 17267 | 3/5 | 0.08991 | 46.7 |
| 23910 | 3/5 | 0.08959 | 40.0 |
| 29343 | 3/5 | 0.08932 | 6.7 |
| 16390 | 3/5 | 0.08892 | 10.0 |

Candidates combine threshold peaks, PELT on the distributional JS-jump series, and binary segmentation. Cross-scale matches use ±0.5 × the smaller window size. Multiple detector hits at one scale count as one scale vote.

## Clustering diagnostics

Hierarchical single-link clustering and JS-distance k-medoids are unconstrained, so distant windows may share a regime. Contiguous binary segmentation is reported separately. K=2..15 is swept without selecting K by interpretability. To bound quadratic fitting, up to 200 sequence-wide uniformly spaced windows are used for fit and silhouette diagnostics; every window is then assigned to a fitted distribution centroid. Cluster sizes, transitions, fragmentation, and the assignment table cover all windows. These clusterings are diagnostics and do not replace the continuous profile.
