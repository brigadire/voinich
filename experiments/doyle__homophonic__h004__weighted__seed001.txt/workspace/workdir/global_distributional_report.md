# Global distributional regime discovery

The corpus is treated as one continuous token sequence. Discovery uses no folio, page, illustration, Currier, hand, or section metadata. Full window token distributions retain all probability mass; no RARE bucket is used.

## Continuous change profile

Adjacent-window Jensen–Shannon distance is the primary result. Weighted overlap and cosine similarity are retained as companion diagnostics. Local peaks, broad transitions, and low/high-variation intervals are descriptive labels rather than assumed true boundaries.

## Stable boundaries

| position | scale support | mean jump | uncertainty |
|---:|---:|---:|---:|
| 6300 | 5/5 | 0.08857 | 20.0 |
| 19003 | 5/5 | 0.07679 | 12.0 |
| 23239 | 4/5 | 0.09302 | 11.2 |
| 17255 | 4/5 | 0.09299 | 25.0 |
| 22905 | 4/5 | 0.09230 | 15.0 |
| 18591 | 4/5 | 0.09188 | 16.2 |
| 2849 | 4/5 | 0.09160 | 11.2 |
| 18485 | 4/5 | 0.09135 | 15.0 |
| 23106 | 4/5 | 0.09132 | 13.8 |
| 10264 | 4/5 | 0.09025 | 13.8 |
| 20844 | 4/5 | 0.09012 | 16.2 |
| 29496 | 4/5 | 0.08995 | 6.2 |
| 18906 | 4/5 | 0.08991 | 13.8 |
| 23498 | 4/5 | 0.08945 | 7.5 |
| 11991 | 4/5 | 0.08853 | 11.2 |
| 12100 | 4/5 | 0.08849 | 20.0 |
| 23793 | 4/5 | 0.08843 | 17.5 |
| 10508 | 4/5 | 0.08815 | 17.5 |
| 3509 | 4/5 | 0.08723 | 21.2 |
| 4191 | 4/5 | 0.08643 | 11.2 |
| 6803 | 4/5 | 0.08611 | 17.5 |
| 18005 | 4/5 | 0.08607 | 15.0 |
| 24101 | 4/5 | 0.08529 | 8.8 |
| 2600 | 4/5 | 0.08432 | 0.0 |
| 14701 | 4/5 | 0.08374 | 8.8 |
| 7386 | 4/5 | 0.08317 | 13.8 |
| 17315 | 3/5 | 0.09670 | 20.0 |
| 27175 | 3/5 | 0.09661 | 15.0 |
| 11817 | 3/5 | 0.09657 | 6.7 |
| 18533 | 3/5 | 0.09645 | 26.7 |

Candidates combine threshold peaks, PELT on the distributional JS-jump series, and binary segmentation. Cross-scale matches use ±0.5 × the smaller window size. Multiple detector hits at one scale count as one scale vote.

## Clustering diagnostics

Hierarchical single-link clustering and JS-distance k-medoids are unconstrained, so distant windows may share a regime. Contiguous binary segmentation is reported separately. K=2..15 is swept without selecting K by interpretability. To bound quadratic fitting, up to 200 sequence-wide uniformly spaced windows are used for fit and silhouette diagnostics; every window is then assigned to a fitted distribution centroid. Cluster sizes, transitions, fragmentation, and the assignment table cover all windows. These clusterings are diagnostics and do not replace the continuous profile.
