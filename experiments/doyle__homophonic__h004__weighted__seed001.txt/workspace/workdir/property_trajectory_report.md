# Property-trajectory analysis

This analysis compares exact-distance trajectories of formal, intrinsic properties of subsequent tokens. It uses neither token classes, smoothing, nor structural projection. Rare subsequent tokens below the configured threshold are excluded and counted in every distance profile.

## Main results

| Pair | cosine 1–5 | 6–10 | 11–20 | matched percentile | random percentile |
|---|---:|---:|---:|---:|---:|
| `x009436` / `x012697` | 0.9966 | 0.9971 | 0.9972 | P100.0 | P99.8 |
| `x009968` / `x020557` | 0.9964 | 0.9926 | 0.9947 | P100.0 | P99.8 |
| `x009969` / `x020557` | 0.9916 | 0.9909 | 0.9932 | P100.0 | P97.5 |
| `x009970` / `x020556` | 0.9907 | 0.9917 | 0.9944 | P100.0 | P97.1 |
| `x009969` / `x020558` | 0.9920 | 0.9911 | 0.9924 | P100.0 | P98.9 |
| `x009970` / `x020558` | 0.9924 | 0.9887 | 0.9929 | P100.0 | P99.6 |
| `x009971` / `x020556` | 0.9771 | 0.9855 | 0.9904 | P100.0 | P86.9 |

## Critical null comparison

The target set does not clear the joint diagnostic criterion (P95 against both matched and random controls, plus a decrease under both shuffles). The property-trajectory hypothesis is therefore not supported as a general explanation by this run.
 Mean target cosine 1–5 is 0.9910; mean matched percentile is 100.0, mean random percentile is 97.1, global-shuffle cosine is 0.9938, and line-preserving-shuffle cosine is 0.9946. Pair-level values remain in the controls TSV and YAML.

## What drives the score

Each pair stores frequency-only, graphemic-form-only, position-only, context-complexity-only, structural-centrality-only, all-properties, and five leave-one-group-out scores. Per-property normalized deltas, trajectory correlations, and the strongest matching/differing rankings remain inspectable rather than being replaced by one score.

## Limits

Cosine in a globally z-scored property space can be negative. Empirical percentiles are deterministic diagnostics, not independent-sample p-values. Structural inputs contribute only each subsequent token's centrality statistics; no pair projection or family membership is used. No semantic labels or latent states are inferred.
