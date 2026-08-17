# Property-trajectory analysis

This analysis compares exact-distance trajectories of formal, intrinsic properties of subsequent tokens. It uses neither token classes, smoothing, nor structural projection. Rare subsequent tokens below the configured threshold are excluded and counted in every distance profile.

## Main results

| Pair | cosine 1–5 | 6–10 | 11–20 | matched percentile | random percentile |
|---|---:|---:|---:|---:|---:|
| `chedy` / `qokeey` | 0.9983 | 0.9989 | 0.9986 | P100.0 | P100.0 |
| `chol` / `daiin` | 0.9960 | 0.9987 | 0.9981 | P100.0 | P99.2 |
| `ol` / `y` | 0.9962 | 0.9964 | 0.9960 | P100.0 | P99.4 |
| `chey` / `ol` | 0.9978 | 0.9978 | 0.9980 | P100.0 | P100.0 |
| `dar` / `ol` | 0.9970 | 0.9975 | 0.9964 | P100.0 | P100.0 |
| `ar` / `ol` | 0.9899 | 0.9962 | 0.9938 | P66.7 | P81.2 |
| `qokain` / `qol` | 0.9943 | 0.9978 | 0.9972 | P100.0 | P96.0 |
| `dal` / `ol` | 0.9982 | 0.9977 | 0.9972 | P100.0 | P100.0 |
| `aiin` / `ar` | 0.9954 | 0.9979 | 0.9975 | P100.0 | P98.1 |
| `okaiin` / `ol` | 0.9955 | 0.9952 | 0.9967 | P100.0 | P98.4 |
| `chor` / `daiin` | 0.9944 | 0.9959 | 0.9952 | P100.0 | P96.9 |
| `or` / `s` | 0.9971 | 0.9949 | 0.9959 | P100.0 | P99.9 |
| `qokaiin` / `qol` | 0.9934 | 0.9974 | 0.9954 | P100.0 | P92.9 |
| `lchedy` / `qokeey` | 0.9970 | 0.9960 | 0.9963 | P100.0 | P100.0 |
| `qokedy` / `qol` | 0.9969 | 0.9971 | 0.9971 | P100.0 | P99.9 |
| `lchedy` / `qokain` | 0.9919 | 0.9964 | 0.9967 | P75.0 | P91.0 |
| `qokar` / `qol` | 0.9925 | 0.9949 | 0.9919 | P100.0 | P89.6 |
| `r` / `s` | 0.9936 | 0.9933 | 0.9940 | P100.0 | P93.7 |
| `okain` / `ol` | 0.9949 | 0.9959 | 0.9962 | P100.0 | P98.3 |
| `or` / `r` | 0.9940 | 0.9967 | 0.9927 | P70.0 | P95.1 |
| `lchedy` / `qol` | 0.9956 | 0.9956 | 0.9952 | P100.0 | P99.5 |
| `lchedy` / `qokar` | 0.9909 | 0.9950 | 0.9925 | P100.0 | P87.0 |
| `daiin` / `dol` | 0.9971 | 0.9970 | 0.9954 | P100.0 | P100.0 |
| `chol` / `cthy` | 0.9945 | 0.9943 | 0.9912 | P100.0 | P96.8 |
| `ain` / `ar` | 0.9868 | 0.9950 | 0.9915 | P100.0 | P69.5 |
| `qol` / `qotain` | 0.9912 | 0.9938 | 0.9917 | P100.0 | P93.3 |
| `ain` / `al` | 0.9918 | 0.9926 | 0.9933 | P100.0 | P90.4 |
| `okar` / `otain` | 0.9927 | 0.9922 | 0.9918 | P100.0 | P92.8 |

## Critical null comparison

The target set does not clear the joint diagnostic criterion (P95 against both matched and random controls, plus a decrease under both shuffles). The property-trajectory hypothesis is therefore not supported as a general explanation by this run.
 Mean target cosine 1–5 is 0.9945; mean matched percentile is 96.8, mean random percentile is 95.0, global-shuffle cosine is 0.9961, and line-preserving-shuffle cosine is 0.9957. Pair-level values remain in the controls TSV and YAML.

## What drives the score

Each pair stores frequency-only, graphemic-form-only, position-only, context-complexity-only, structural-centrality-only, all-properties, and five leave-one-group-out scores. Per-property normalized deltas, trajectory correlations, and the strongest matching/differing rankings remain inspectable rather than being replaced by one score.

## Limits

Cosine in a globally z-scored property space can be negative. Empirical percentiles are deterministic diagnostics, not independent-sample p-values. Structural inputs contribute only each subsequent token's centrality statistics; no pair projection or family membership is used. No semantic labels or latent states are inferred.
