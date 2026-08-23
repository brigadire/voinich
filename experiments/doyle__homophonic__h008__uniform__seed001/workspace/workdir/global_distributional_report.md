# Global distributional regime discovery

The corpus is treated as one continuous token sequence. Discovery uses no folio, page, illustration, Currier, hand, or section metadata. Full window token distributions retain all probability mass; no RARE bucket is used.

## Continuous change profile

Adjacent-window Jensen–Shannon distance is the primary result. Weighted overlap and cosine similarity are retained as companion diagnostics. Local peaks, broad transitions, and low/high-variation intervals are descriptive labels rather than assumed true boundaries.

## Stable boundaries

| position | scale support | mean jump | uncertainty |
|---:|---:|---:|---:|
| 24170 | 4/5 | 0.08998 | 40.0 |
| 5733 | 4/5 | 0.08992 | 32.5 |
| 17998 | 4/5 | 0.08566 | 27.5 |
| 6318 | 4/5 | 0.08466 | 32.5 |
| 22878 | 4/5 | 0.08368 | 37.5 |
| 23293 | 3/5 | 0.09504 | 33.3 |
| 29380 | 3/5 | 0.09431 | 40.0 |
| 23817 | 3/5 | 0.09426 | 23.3 |
| 28267 | 3/5 | 0.09423 | 33.3 |
| 33720 | 3/5 | 0.09418 | 20.0 |
| 13402 | 3/5 | 0.09399 | 3.3 |
| 10513 | 3/5 | 0.09394 | 43.3 |
| 24823 | 3/5 | 0.09386 | 43.3 |
| 17333 | 3/5 | 0.09370 | 16.7 |
| 15303 | 3/5 | 0.09348 | 6.7 |
| 22733 | 3/5 | 0.09347 | 43.3 |
| 15913 | 3/5 | 0.09344 | 36.7 |
| 23080 | 3/5 | 0.09335 | 40.0 |
| 18910 | 3/5 | 0.09312 | 40.0 |
| 2827 | 3/5 | 0.09299 | 36.7 |
| 28173 | 3/5 | 0.09299 | 33.3 |
| 22067 | 3/5 | 0.09285 | 33.3 |
| 35270 | 3/5 | 0.09280 | 30.0 |
| 8977 | 3/5 | 0.09268 | 26.7 |
| 4533 | 3/5 | 0.09207 | 16.7 |
| 23440 | 3/5 | 0.09197 | 20.0 |
| 23173 | 3/5 | 0.09192 | 26.7 |
| 19813 | 3/5 | 0.09171 | 36.7 |
| 4797 | 3/5 | 0.09138 | 26.7 |
| 4270 | 3/5 | 0.09121 | 30.0 |

Candidates combine threshold peaks, PELT on the distributional JS-jump series, and binary segmentation. Cross-scale matches use ±0.5 × the smaller window size. Multiple detector hits at one scale count as one scale vote.

## Clustering diagnostics

Hierarchical single-link clustering and JS-distance k-medoids are unconstrained, so distant windows may share a regime. Contiguous binary segmentation is reported separately. K=2..15 is swept without selecting K by interpretability. To bound quadratic fitting, up to 200 sequence-wide uniformly spaced windows are used for fit and silhouette diagnostics; every window is then assigned to a fitted distribution centroid. Cluster sizes, transitions, fragmentation, and the assignment table cover all windows. These clusterings are diagnostics and do not replace the continuous profile.
