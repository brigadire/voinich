# Distance-specific context analysis

This report describes formal token-sequence distributions and does not assign meanings to tokens. The main analysis treats all non-empty lines as one continuous sequence; pages are not used. Physical lines are retained only for the line-bounded control.

## Parameters and corpus

Corpus tokens: 43713; target pairs: 0; maximum exact distance: 20; minimum observations: 30; requested primary mode: `continuous`. Similarity at every distance is computed separately, never from a pooled window. Frequency-matched baseline pairs have unordered counts within a factor of 2. Effective support is `exp(Shannon entropy)` using natural-log units.

## Right-context baseline

| Distance | Pairs | Median JS | P90 | P95 |
|---:|---:|---:|---:|---:|
| +1 | 13231 | 0.0282 | 0.1579 | 0.2092 |
| +2 | 13231 | 0.0386 | 0.0951 | 0.1162 |
| +3 | 13231 | 0.0581 | 0.1221 | 0.1406 |
| +4 | 13231 | 0.0565 | 0.1197 | 0.1387 |
| +5 | 13231 | 0.0580 | 0.1166 | 0.1334 |
| +6 | 13231 | 0.0590 | 0.1203 | 0.1392 |
| +7 | 13231 | 0.0580 | 0.1183 | 0.1366 |
| +8 | 13231 | 0.0574 | 0.1178 | 0.1359 |
| +9 | 13231 | 0.0574 | 0.1165 | 0.1336 |
| +10 | 13231 | 0.0572 | 0.1197 | 0.1388 |
| +11 | 13231 | 0.0588 | 0.1186 | 0.1358 |
| +12 | 13231 | 0.0595 | 0.1210 | 0.1373 |
| +13 | 13231 | 0.0554 | 0.1166 | 0.1358 |
| +14 | 13231 | 0.0553 | 0.1151 | 0.1331 |
| +15 | 13231 | 0.0597 | 0.1210 | 0.1402 |
| +16 | 13231 | 0.0584 | 0.1184 | 0.1365 |
| +17 | 13231 | 0.0546 | 0.1140 | 0.1312 |
| +18 | 13231 | 0.0572 | 0.1156 | 0.1330 |
| +19 | 13231 | 0.0580 | 0.1192 | 0.1366 |
| +20 | 13231 | 0.0565 | 0.1193 | 0.1375 |

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
