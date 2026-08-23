# Global distributional regime discovery

The corpus is treated as one continuous token sequence. Discovery uses no folio, page, illustration, Currier, hand, or section metadata. Full window token distributions retain all probability mass; no RARE bucket is used.

## Continuous change profile

Adjacent-window Jensen–Shannon distance is the primary result. Weighted overlap and cosine similarity are retained as companion diagnostics. Local peaks, broad transitions, and low/high-variation intervals are descriptive labels rather than assumed true boundaries.

## Stable boundaries

| position | scale support | mean jump | uncertainty |
|---:|---:|---:|---:|
| 17283 | 5/5 | 0.08985 | 18.0 |
| 22894 | 5/5 | 0.08755 | 14.0 |
| 3096 | 5/5 | 0.08721 | 24.0 |
| 23839 | 4/5 | 0.09323 | 13.8 |
| 23179 | 4/5 | 0.09304 | 21.2 |
| 7181 | 4/5 | 0.09214 | 18.8 |
| 7004 | 4/5 | 0.09187 | 6.2 |
| 15305 | 4/5 | 0.09175 | 5.0 |
| 28188 | 4/5 | 0.09125 | 12.5 |
| 20091 | 4/5 | 0.09120 | 18.8 |
| 22416 | 4/5 | 0.09118 | 23.8 |
| 18915 | 4/5 | 0.09115 | 25.0 |
| 29601 | 4/5 | 0.09085 | 3.8 |
| 2848 | 4/5 | 0.09069 | 12.5 |
| 29361 | 4/5 | 0.09064 | 23.8 |
| 29998 | 4/5 | 0.09063 | 7.5 |
| 18203 | 4/5 | 0.09055 | 7.5 |
| 10561 | 4/5 | 0.09037 | 18.8 |
| 9649 | 4/5 | 0.09035 | 21.2 |
| 10210 | 4/5 | 0.09022 | 10.0 |
| 23505 | 4/5 | 0.09014 | 15.0 |
| 11799 | 4/5 | 0.08991 | 21.2 |
| 5384 | 4/5 | 0.08958 | 18.8 |
| 20610 | 4/5 | 0.08924 | 20.0 |
| 18000 | 4/5 | 0.08846 | 20.0 |
| 28999 | 4/5 | 0.08838 | 13.8 |
| 31814 | 4/5 | 0.08715 | 16.2 |
| 28313 | 4/5 | 0.08682 | 27.5 |
| 24206 | 4/5 | 0.08667 | 8.8 |
| 37366 | 4/5 | 0.08664 | 23.8 |

Candidates combine threshold peaks, PELT on the distributional JS-jump series, and binary segmentation. Cross-scale matches use ±0.5 × the smaller window size. Multiple detector hits at one scale count as one scale vote.

## Clustering diagnostics

Hierarchical single-link clustering and JS-distance k-medoids are unconstrained, so distant windows may share a regime. Contiguous binary segmentation is reported separately. K=2..15 is swept without selecting K by interpretability. To bound quadratic fitting, up to 200 sequence-wide uniformly spaced windows are used for fit and silhouette diagnostics; every window is then assigned to a fitted distribution centroid. Cluster sizes, transitions, fragmentation, and the assignment table cover all windows. These clusterings are diagnostics and do not replace the continuous profile.
