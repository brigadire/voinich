# Distance-specific context analysis

This report describes formal token-sequence distributions and does not assign meanings to tokens. The main analysis treats all non-empty lines as one continuous sequence; pages are not used. Physical lines are retained only for the line-bounded control.

## Parameters and corpus

Corpus tokens: 43713; target pairs: 0; maximum exact distance: 20; minimum observations: 30; requested primary mode: `continuous`. Similarity at every distance is computed separately, never from a pooled window. Frequency-matched baseline pairs have unordered counts within a factor of 2. Effective support is `exp(Shannon entropy)` using natural-log units.

## Right-context baseline

| Distance | Pairs | Median JS | P90 | P95 |
|---:|---:|---:|---:|---:|
| +1 | 13502 | 0.0258 | 0.1541 | 0.2080 |
| +2 | 13502 | 0.0469 | 0.1082 | 0.1299 |
| +3 | 13502 | 0.0646 | 0.1341 | 0.1588 |
| +4 | 13502 | 0.0656 | 0.1351 | 0.1587 |
| +5 | 13502 | 0.0669 | 0.1341 | 0.1558 |
| +6 | 13502 | 0.0662 | 0.1332 | 0.1567 |
| +7 | 13502 | 0.0655 | 0.1305 | 0.1544 |
| +8 | 13502 | 0.0651 | 0.1318 | 0.1533 |
| +9 | 13502 | 0.0651 | 0.1299 | 0.1529 |
| +10 | 13502 | 0.0658 | 0.1313 | 0.1533 |
| +11 | 13502 | 0.0671 | 0.1348 | 0.1572 |
| +12 | 13502 | 0.0666 | 0.1376 | 0.1601 |
| +13 | 13502 | 0.0647 | 0.1338 | 0.1582 |
| +14 | 13502 | 0.0649 | 0.1277 | 0.1512 |
| +15 | 13502 | 0.0648 | 0.1334 | 0.1590 |
| +16 | 13502 | 0.0656 | 0.1340 | 0.1571 |
| +17 | 13502 | 0.0626 | 0.1275 | 0.1507 |
| +18 | 13502 | 0.0663 | 0.1322 | 0.1555 |
| +19 | 13502 | 0.0658 | 0.1319 | 0.1553 |
| +20 | 13502 | 0.0643 | 0.1323 | 0.1562 |

## Long-range structural persistence ranking

`persistence_1_5` is the transparent mean of baseline percentile ranks at exact distances +1 through +5. It is a ranking aid, not a replacement for the profiles.

| Pair | Persistence 1–5 | Mean JS 1–5 | 6–10 | 11–20 |
|---|---:|---:|---:|---:|

## Target profiles, directionality, and boundary sensitivity

## Matched negative controls

| Target | Controls | Target mean JS 1–5 | Control median mean JS 1–5 |
|---|---:|---:|---:|

## Family cohesion

Matrices and distance-specific medoids are in `family_distance_profiles.yaml`. Random controls are deterministic groups of equal size whose member frequencies match the corresponding family member within ×2.

| Family | d | Cohesion | Matched percentile | Medoid |
|---:|---:|---:|---:|---|

## Sequence-context signatures

The 2- and 3-token suffix distributions are compared separately with JS similarity, weighted overlap, support Jaccard, observation counts, effective support, and reliability in `sequence_context_pairs.yaml`.

## Limitations

Similarity at greater distance is descriptive context-similarity persistence, not physical decay. Corpus positions and pair comparisons are dependent. Frequency matching controls only a major sampling confound and is not a causal null model. Support overlap is sensitive to rare observations. A low-reliability value is retained but must not be treated as equivalent to a frequent comparison. Line-bounded results remove cross-line observations and therefore can change both distributions and sample sizes. Sequence suffixes become sparse quickly; only lengths 2 and 3 are used. No semantic interpretation is made.
