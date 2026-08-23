# Distance-specific context analysis

This report describes formal token-sequence distributions and does not assign meanings to tokens. The main analysis treats all non-empty lines as one continuous sequence; pages are not used. Physical lines are retained only for the line-bounded control.

## Parameters and corpus

Corpus tokens: 43713; target pairs: 0; maximum exact distance: 20; minimum observations: 30; requested primary mode: `continuous`. Similarity at every distance is computed separately, never from a pooled window. Frequency-matched baseline pairs have unordered counts within a factor of 2. Effective support is `exp(Shannon entropy)` using natural-log units.

## Right-context baseline

| Distance | Pairs | Median JS | P90 | P95 |
|---:|---:|---:|---:|---:|
| +1 | 11708 | 0.0552 | 0.2245 | 0.3000 |
| +2 | 11708 | 0.1036 | 0.1936 | 0.2264 |
| +3 | 11708 | 0.1160 | 0.2230 | 0.2697 |
| +4 | 11708 | 0.1227 | 0.2261 | 0.2711 |
| +5 | 11708 | 0.1209 | 0.2310 | 0.2676 |
| +6 | 11708 | 0.1243 | 0.2298 | 0.2774 |
| +7 | 11708 | 0.1211 | 0.2240 | 0.2752 |
| +8 | 11708 | 0.1207 | 0.2300 | 0.2823 |
| +9 | 11708 | 0.1229 | 0.2214 | 0.2702 |
| +10 | 11708 | 0.1192 | 0.2252 | 0.2746 |
| +11 | 11708 | 0.1260 | 0.2263 | 0.2756 |
| +12 | 11708 | 0.1231 | 0.2304 | 0.2779 |
| +13 | 11708 | 0.1227 | 0.2248 | 0.2724 |
| +14 | 11708 | 0.1213 | 0.2238 | 0.2709 |
| +15 | 11708 | 0.1250 | 0.2312 | 0.2833 |
| +16 | 11708 | 0.1245 | 0.2280 | 0.2748 |
| +17 | 11708 | 0.1226 | 0.2287 | 0.2714 |
| +18 | 11708 | 0.1173 | 0.2261 | 0.2757 |
| +19 | 11708 | 0.1258 | 0.2350 | 0.2800 |
| +20 | 11708 | 0.1160 | 0.2216 | 0.2674 |

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
