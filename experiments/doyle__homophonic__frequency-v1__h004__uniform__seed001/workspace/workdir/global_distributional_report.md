# Global distributional regime discovery

The corpus is treated as one continuous token sequence. Discovery uses no folio, page, illustration, Currier, hand, or section metadata. Full window token distributions retain all probability mass; no RARE bucket is used.

## Continuous change profile

Adjacent-window Jensen–Shannon distance is the primary result. Weighted overlap and cosine similarity are retained as companion diagnostics. Local peaks, broad transitions, and low/high-variation intervals are descriptive labels rather than assumed true boundaries.

## Stable boundaries

| position | scale support | mean jump | uncertainty |
|---:|---:|---:|---:|
| 6304 | 5/5 | 0.08335 | 16.0 |
| 2600 | 5/5 | 0.07906 | 0.0 |
| 17260 | 4/5 | 0.09216 | 20.0 |
| 10261 | 4/5 | 0.09210 | 11.2 |
| 20844 | 4/5 | 0.09125 | 16.2 |
| 17659 | 4/5 | 0.09015 | 21.2 |
| 23566 | 4/5 | 0.09000 | 16.2 |
| 17151 | 4/5 | 0.08994 | 11.2 |
| 913 | 4/5 | 0.08989 | 17.5 |
| 18901 | 4/5 | 0.08967 | 21.2 |
| 23293 | 4/5 | 0.08962 | 12.5 |
| 10694 | 4/5 | 0.08956 | 18.8 |
| 28186 | 4/5 | 0.08937 | 13.8 |
| 18510 | 4/5 | 0.08858 | 20.0 |
| 15306 | 4/5 | 0.08851 | 8.8 |
| 18000 | 4/5 | 0.08844 | 20.0 |
| 23010 | 4/5 | 0.08804 | 10.0 |
| 6536 | 4/5 | 0.08739 | 16.2 |
| 2711 | 4/5 | 0.08721 | 11.2 |
| 29601 | 4/5 | 0.08707 | 3.8 |
| 2843 | 4/5 | 0.08633 | 7.5 |
| 29338 | 4/5 | 0.08623 | 12.5 |
| 23505 | 4/5 | 0.08609 | 15.0 |
| 25839 | 4/5 | 0.08457 | 11.2 |
| 22748 | 4/5 | 0.08354 | 17.5 |
| 20611 | 4/5 | 0.08267 | 18.8 |
| 23796 | 4/5 | 0.08243 | 6.2 |
| 27994 | 4/5 | 0.07865 | 13.8 |
| 24198 | 4/5 | 0.07601 | 7.5 |
| 23720 | 3/5 | 0.09751 | 0.0 |

Candidates combine threshold peaks, PELT on the distributional JS-jump series, and binary segmentation. Cross-scale matches use ±0.5 × the smaller window size. Multiple detector hits at one scale count as one scale vote.

## Clustering diagnostics

Hierarchical single-link clustering and JS-distance k-medoids are unconstrained, so distant windows may share a regime. Contiguous binary segmentation is reported separately. K=2..15 is swept without selecting K by interpretability. To bound quadratic fitting, up to 200 sequence-wide uniformly spaced windows are used for fit and silhouette diagnostics; every window is then assigned to a fitted distribution centroid. Cluster sizes, transitions, fragmentation, and the assignment table cover all windows. These clusterings are diagnostics and do not replace the continuous profile.
