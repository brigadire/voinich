# Property-trajectory analysis

This analysis compares exact-distance trajectories of formal, intrinsic properties of subsequent tokens. It uses neither token classes, smoothing, nor structural projection. Rare subsequent tokens below the configured threshold are excluded and counted in every distance profile.

## Main results

| Pair | cosine 1–5 | 6–10 | 11–20 | matched percentile | random percentile |
|---|---:|---:|---:|---:|---:|
| `x004624` / `x010315` | 0.9827 | 0.9977 | 0.9969 | P100.0 | P77.0 |
| `x004984` / `x010279` | 0.9968 | 0.9955 | 0.9972 | P100.0 | P99.6 |
| `x004625` / `x010314` | 0.9846 | 0.9963 | 0.9980 | P100.0 | P80.7 |
| `x004985` / `x010279` | 0.9963 | 0.9949 | 0.9973 | P100.0 | P99.8 |
| `x004719` / `x010526` | 0.9942 | 0.9984 | 0.9977 | P100.0 | P98.5 |
| `x004718` / `x010526` | 0.9932 | 0.9984 | 0.9981 | P100.0 | P97.8 |
| `x004719` / `x010527` | 0.9925 | 0.9981 | 0.9978 | P100.0 | P96.9 |
| `x009496` / `x010315` | 0.9888 | 0.9948 | 0.9937 | P100.0 | P92.1 |
| `x006383` / `x010054` | 0.9741 | 0.9950 | 0.9933 | P100.0 | P53.5 |
| `x004718` / `x010054` | 0.9929 | 0.9966 | 0.9958 | P100.0 | P98.2 |
| `x009497` / `x010315` | 0.9820 | 0.9917 | 0.9909 | P100.0 | P78.2 |
| `x006348` / `x010055` | 0.9915 | 0.9962 | 0.9965 | P100.0 | P96.8 |
| `x004718` / `x010055` | 0.9913 | 0.9950 | 0.9958 | P100.0 | P96.1 |
| `x003821` / `x010055` | 0.9845 | 0.9946 | 0.9940 | P100.0 | P84.0 |
| `x006383` / `x010055` | 0.9719 | 0.9903 | 0.9915 | P33.3 | P48.7 |
| `x005730` / `x010617` | 0.9828 | 0.9875 | 0.9863 | P100.0 | P88.4 |
| `x008248` / `x010617` | 0.9843 | 0.9810 | 0.9844 | P100.0 | P92.0 |
| `x008353` / `x010617` | 0.9846 | 0.9893 | 0.9876 | P100.0 | P91.5 |
| `x005731` / `x010617` | 0.9835 | 0.9831 | 0.9873 | P100.0 | P88.6 |
| `x004929` / `x010054` | 0.8673 | 0.9934 | 0.9894 | P0.0 | P14.0 |
| `x005842` / `x010617` | 0.9808 | 0.9877 | 0.9843 | P100.0 | P85.4 |
| `x008352` / `x010617` | 0.9863 | 0.9883 | 0.9811 | P100.0 | P94.9 |
| `x005843` / `x010617` | 0.9824 | 0.9877 | 0.9725 | P100.0 | P93.0 |
| `x006050` / `x010484` | 0.9808 | 0.9808 | 0.9841 | P100.0 | P83.6 |
| `x005730` / `x010485` | 0.9824 | 0.9874 | 0.9807 | P100.0 | P91.6 |
| `x005843` / `x010616` | 0.9848 | 0.9803 | 0.9691 | P100.0 | P95.8 |

## Critical null comparison

The target set does not clear the joint diagnostic criterion (P95 against both matched and random controls, plus a decrease under both shuffles). The property-trajectory hypothesis is therefore not supported as a general explanation by this run.
 Mean target cosine 1–5 is 0.9814; mean matched percentile is 93.6, mean random percentile is 85.3, global-shuffle cosine is 0.9902, and line-preserving-shuffle cosine is 0.9895. Pair-level values remain in the controls TSV and YAML.

## What drives the score

Each pair stores frequency-only, graphemic-form-only, position-only, context-complexity-only, structural-centrality-only, all-properties, and five leave-one-group-out scores. Per-property normalized deltas, trajectory correlations, and the strongest matching/differing rankings remain inspectable rather than being replaced by one score.

## Limits

Cosine in a globally z-scored property space can be negative. Empirical percentiles are deterministic diagnostics, not independent-sample p-values. Structural inputs contribute only each subsequent token's centrality statistics; no pair projection or family membership is used. No semantic labels or latent states are inferred.
