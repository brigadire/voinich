# Global distributional regime discovery

The corpus is treated as one continuous token sequence. Discovery uses no folio, page, illustration, Currier, hand, or section metadata. Full window token distributions retain all probability mass; no RARE bucket is used.

## Continuous change profile

Adjacent-window Jensen–Shannon distance is the primary result. Weighted overlap and cosine similarity are retained as companion diagnostics. Local peaks, broad transitions, and low/high-variation intervals are descriptive labels rather than assumed true boundaries.

## Stable boundaries

| position | scale support | mean jump | uncertainty |
|---:|---:|---:|---:|
| 22904 | 5/5 | 0.07627 | 6.0 |
| 23253 | 4/5 | 0.08604 | 12.5 |
| 21501 | 4/5 | 0.08526 | 3.8 |
| 29984 | 4/5 | 0.08425 | 23.8 |
| 9886 | 4/5 | 0.08411 | 16.2 |
| 23500 | 4/5 | 0.08334 | 0.0 |
| 21003 | 4/5 | 0.08319 | 7.5 |
| 15306 | 4/5 | 0.08312 | 8.8 |
| 18490 | 4/5 | 0.08307 | 20.0 |
| 10501 | 4/5 | 0.08300 | 18.8 |
| 11384 | 4/5 | 0.08272 | 18.8 |
| 5768 | 4/5 | 0.08232 | 22.5 |
| 29738 | 4/5 | 0.08143 | 17.5 |
| 22399 | 4/5 | 0.08080 | 8.8 |
| 18005 | 4/5 | 0.08071 | 15.0 |
| 6300 | 4/5 | 0.08063 | 0.0 |
| 589 | 4/5 | 0.08039 | 18.8 |
| 23690 | 4/5 | 0.07821 | 10.0 |
| 7009 | 4/5 | 0.07674 | 11.2 |
| 24203 | 4/5 | 0.07597 | 7.5 |
| 913 | 4/5 | 0.07523 | 17.5 |
| 3090 | 4/5 | 0.07314 | 10.0 |
| 2600 | 4/5 | 0.07199 | 0.0 |
| 14906 | 4/5 | 0.07125 | 8.8 |
| 20005 | 4/5 | 0.07111 | 15.0 |
| 2788 | 3/5 | 0.09594 | 23.3 |
| 4048 | 3/5 | 0.09441 | 11.7 |
| 18247 | 3/5 | 0.09441 | 13.3 |
| 21042 | 3/5 | 0.09413 | 3.3 |
| 3977 | 3/5 | 0.09388 | 6.7 |

Candidates combine threshold peaks, PELT on the distributional JS-jump series, and binary segmentation. Cross-scale matches use ±0.5 × the smaller window size. Multiple detector hits at one scale count as one scale vote.

## Clustering diagnostics

Hierarchical single-link clustering and JS-distance k-medoids are unconstrained, so distant windows may share a regime. Contiguous binary segmentation is reported separately. K=2..15 is swept without selecting K by interpretability. To bound quadratic fitting, up to 200 sequence-wide uniformly spaced windows are used for fit and silhouette diagnostics; every window is then assigned to a fitted distribution centroid. Cluster sizes, transitions, fragmentation, and the assignment table cover all windows. These clusterings are diagnostics and do not replace the continuous profile.
