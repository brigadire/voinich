# Global distributional regime discovery

The corpus is treated as one continuous token sequence. Discovery uses no folio, page, illustration, Currier, hand, or section metadata. Full window token distributions retain all probability mass; no RARE bucket is used.

## Continuous change profile

Adjacent-window Jensen–Shannon distance is the primary result. Weighted overlap and cosine similarity are retained as companion diagnostics. Local peaks, broad transitions, and low/high-variation intervals are descriptive labels rather than assumed true boundaries.

## Stable boundaries

| position | scale support | mean jump | uncertainty |
|---:|---:|---:|---:|
| 17673 | 4/5 | 0.08856 | 27.5 |
| 17288 | 4/5 | 0.08821 | 17.5 |
| 4260 | 4/5 | 0.08798 | 40.0 |
| 16075 | 4/5 | 0.08617 | 25.0 |
| 24175 | 4/5 | 0.08592 | 25.0 |
| 17110 | 4/5 | 0.08567 | 30.0 |
| 19990 | 4/5 | 0.08557 | 30.0 |
| 23010 | 4/5 | 0.08533 | 40.0 |
| 16878 | 4/5 | 0.08271 | 27.5 |
| 10980 | 4/5 | 0.08118 | 30.0 |
| 23300 | 3/5 | 0.09479 | 0.0 |
| 17463 | 3/5 | 0.09355 | 36.7 |
| 32427 | 3/5 | 0.09300 | 26.7 |
| 15303 | 3/5 | 0.09238 | 6.7 |
| 2830 | 3/5 | 0.09234 | 30.0 |
| 10530 | 3/5 | 0.09223 | 20.0 |
| 20247 | 3/5 | 0.09211 | 26.7 |
| 28173 | 3/5 | 0.09194 | 33.3 |
| 15427 | 3/5 | 0.09157 | 36.7 |
| 29407 | 3/5 | 0.09146 | 13.3 |
| 5727 | 3/5 | 0.09131 | 33.3 |
| 41273 | 3/5 | 0.09127 | 36.7 |
| 22890 | 3/5 | 0.09095 | 20.0 |
| 23647 | 3/5 | 0.09091 | 46.7 |
| 24853 | 3/5 | 0.09086 | 16.7 |
| 21773 | 3/5 | 0.09086 | 23.3 |
| 14340 | 3/5 | 0.09058 | 20.0 |
| 22733 | 3/5 | 0.09057 | 43.3 |
| 25023 | 3/5 | 0.09056 | 26.7 |
| 24687 | 3/5 | 0.09047 | 13.3 |

Candidates combine threshold peaks, PELT on the distributional JS-jump series, and binary segmentation. Cross-scale matches use ±0.5 × the smaller window size. Multiple detector hits at one scale count as one scale vote.

## Clustering diagnostics

Hierarchical single-link clustering and JS-distance k-medoids are unconstrained, so distant windows may share a regime. Contiguous binary segmentation is reported separately. K=2..15 is swept without selecting K by interpretability. To bound quadratic fitting, up to 200 sequence-wide uniformly spaced windows are used for fit and silhouette diagnostics; every window is then assigned to a fitted distribution centroid. Cluster sizes, transitions, fragmentation, and the assignment table cover all windows. These clusterings are diagnostics and do not replace the continuous profile.
