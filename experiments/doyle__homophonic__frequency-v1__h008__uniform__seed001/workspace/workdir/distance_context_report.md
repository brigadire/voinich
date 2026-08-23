# Distance-specific context analysis

This report describes formal token-sequence distributions and does not assign meanings to tokens. The main analysis treats all non-empty lines as one continuous sequence; pages are not used. Physical lines are retained only for the line-bounded control.

## Parameters and corpus

Corpus tokens: 43713; target pairs: 0; maximum exact distance: 20; minimum observations: 30; requested primary mode: `continuous`. Similarity at every distance is computed separately, never from a pooled window. Frequency-matched baseline pairs have unordered counts within a factor of 2. Effective support is `exp(Shannon entropy)` using natural-log units.

## Right-context baseline

| Distance | Pairs | Median JS | P90 | P95 |
|---:|---:|---:|---:|---:|
| +1 | 13647 | 0.0291 | 0.1615 | 0.2132 |
| +2 | 13647 | 0.0455 | 0.1075 | 0.1314 |
| +3 | 13647 | 0.0616 | 0.1355 | 0.1618 |
| +4 | 13647 | 0.0625 | 0.1326 | 0.1542 |
| +5 | 13647 | 0.0635 | 0.1306 | 0.1517 |
| +6 | 13647 | 0.0639 | 0.1359 | 0.1582 |
| +7 | 13647 | 0.0616 | 0.1325 | 0.1533 |
| +8 | 13647 | 0.0644 | 0.1318 | 0.1525 |
| +9 | 13647 | 0.0614 | 0.1302 | 0.1517 |
| +10 | 13647 | 0.0616 | 0.1339 | 0.1582 |
| +11 | 13647 | 0.0625 | 0.1302 | 0.1512 |
| +12 | 13647 | 0.0652 | 0.1342 | 0.1557 |
| +13 | 13647 | 0.0609 | 0.1289 | 0.1525 |
| +14 | 13647 | 0.0598 | 0.1289 | 0.1500 |
| +15 | 13647 | 0.0635 | 0.1353 | 0.1577 |
| +16 | 13647 | 0.0620 | 0.1302 | 0.1515 |
| +17 | 13647 | 0.0598 | 0.1285 | 0.1504 |
| +18 | 13647 | 0.0620 | 0.1305 | 0.1527 |
| +19 | 13647 | 0.0632 | 0.1342 | 0.1551 |
| +20 | 13647 | 0.0619 | 0.1332 | 0.1568 |

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
