# Property-trajectory analysis

This analysis compares exact-distance trajectories of formal, intrinsic properties of subsequent tokens. It uses neither token classes, smoothing, nor structural projection. Rare subsequent tokens below the configured threshold are excluded and counted in every distance profile.

## Main results

| Pair | cosine 1–5 | 6–10 | 11–20 | matched percentile | random percentile |
|---|---:|---:|---:|---:|---:|
| `x014952` / `x030834` | 0.9873 | 0.9917 | 0.9918 | P100.0 | P98.6 |

## Critical null comparison

The target set does not clear the joint diagnostic criterion (P95 against both matched and random controls, plus a decrease under both shuffles). The property-trajectory hypothesis is therefore not supported as a general explanation by this run.
 Mean target cosine 1–5 is 0.9873; mean matched percentile is 100.0, mean random percentile is 98.6, global-shuffle cosine is 0.9930, and line-preserving-shuffle cosine is 0.9956. Pair-level values remain in the controls TSV and YAML.

## What drives the score

Each pair stores frequency-only, graphemic-form-only, position-only, context-complexity-only, structural-centrality-only, all-properties, and five leave-one-group-out scores. Per-property normalized deltas, trajectory correlations, and the strongest matching/differing rankings remain inspectable rather than being replaced by one score.

## Limits

Cosine in a globally z-scored property space can be negative. Empirical percentiles are deterministic diagnostics, not independent-sample p-values. Structural inputs contribute only each subsequent token's centrality statistics; no pair projection or family membership is used. No semantic labels or latent states are inferred.
