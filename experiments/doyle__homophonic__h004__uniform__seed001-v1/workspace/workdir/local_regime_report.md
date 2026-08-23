# Local-regime / non-stationarity analysis

The corpus is treated as a continuous token sequence. Profiles exclude the central gap; no semantic labels, manuscript metadata, projection, or property trajectories are used.

## Pair decomposition

| pair | regime JS | distance JS 1–5 | expected | residual |
|---|---:|---:|---:|---:|
| `x009968` / `x020559` | 0.8005 | 0.2253 | 0.6881 | -0.4628 |
| `x009969` / `x020558` | 0.7917 | 0.2244 | 0.6509 | -0.4264 |
| `x009970` / `x020558` | 0.7987 | 0.2148 | 0.6732 | -0.4584 |
| `x009971` / `x020558` | 0.7911 | 0.2095 | 0.6317 | -0.4222 |
| `x009970` / `x020556` | 0.7986 | 0.2046 | 0.6801 | -0.4755 |
| `x009970` / `x020559` | 0.7958 | 0.2015 | 0.6774 | -0.4759 |
| `x009971` / `x020556` | 0.7905 | 0.1809 | 0.6527 | -0.4718 |
| `x009971` / `x020559` | 0.7803 | 0.1798 | 0.6200 | -0.4401 |

## Correlation diagnostic

regime_js_vs_distance_js_1_5: Pearson 0.6777, Spearman 0.7381 (N=8). These are descriptive diagnostics; ordinary independent-pair p-values are not used.

Detected 2659 neutral distributional change points by a mean-plus-one-standard-deviation JS-jump threshold. Parameter sweeps are retained in YAML and no radius, gap, block size, or number of regimes was selected post hoc.

## Interpretation

Compare original, regime-expected, and local-block-shuffle series. A small residual and high retained effect support shared local composition; a large positive residual and a decrease after block shuffle support additional sequential structure. Retained fractions are ratios, not probabilities.
