# Property-trajectory analysis

This analysis compares exact-distance trajectories of formal, intrinsic properties of subsequent tokens. It uses neither token classes, smoothing, nor structural projection. Rare subsequent tokens below the configured threshold are excluded and counted in every distance profile.

## Main results

| Pair | cosine 1–5 | 6–10 | 11–20 | matched percentile | random percentile |
|---|---:|---:|---:|---:|---:|
| `x009968` / `x020559` | 0.9921 | 0.9936 | 0.9923 | P100.0 | P98.2 |
| `x009969` / `x020558` | 0.9876 | 0.9865 | 0.9864 | P100.0 | P95.8 |
| `x009970` / `x020558` | 0.9903 | 0.9891 | 0.9923 | P100.0 | P98.3 |
| `x009971` / `x020558` | 0.9897 | 0.9842 | 0.9902 | P100.0 | P97.6 |
| `x009970` / `x020556` | 0.9834 | 0.9903 | 0.9929 | P100.0 | P88.5 |
| `x009970` / `x020559` | 0.9845 | 0.9920 | 0.9919 | P100.0 | P92.2 |
| `x009971` / `x020556` | 0.9862 | 0.9909 | 0.9918 | P100.0 | P93.2 |
| `x009971` / `x020559` | 0.9863 | 0.9922 | 0.9934 | P100.0 | P94.1 |

## Critical null comparison

The target set does not clear the joint diagnostic criterion (P95 against both matched and random controls, plus a decrease under both shuffles). The property-trajectory hypothesis is therefore not supported as a general explanation by this run.
 Mean target cosine 1–5 is 0.9875; mean matched percentile is 100.0, mean random percentile is 94.7, global-shuffle cosine is 0.9931, and line-preserving-shuffle cosine is 0.9930. Pair-level values remain in the controls TSV and YAML.

## What drives the score

Each pair stores frequency-only, graphemic-form-only, position-only, context-complexity-only, structural-centrality-only, all-properties, and five leave-one-group-out scores. Per-property normalized deltas, trajectory correlations, and the strongest matching/differing rankings remain inspectable rather than being replaced by one score.

## Limits

Cosine in a globally z-scored property space can be negative. Empirical percentiles are deterministic diagnostics, not independent-sample p-values. Structural inputs contribute only each subsequent token's centrality statistics; no pair projection or family membership is used. No semantic labels or latent states are inferred.
