# Local-regime / non-stationarity analysis

The corpus is treated as a continuous token sequence. Profiles exclude the central gap; no semantic labels, manuscript metadata, projection, or property trajectories are used.

## Pair decomposition

| pair | regime JS | distance JS 1–5 | expected | residual |
|---|---:|---:|---:|---:|
| `x009436` / `x012697` | 0.9225 | 0.3391 | 0.8762 | -0.5371 |
| `x009968` / `x020557` | 0.8433 | 0.2645 | 0.7326 | -0.4681 |
| `x009969` / `x020557` | 0.8248 | 0.2518 | 0.6861 | -0.4343 |
| `x009970` / `x020556` | 0.7992 | 0.2162 | 0.6088 | -0.3926 |
| `x009969` / `x020558` | 0.7903 | 0.2239 | 0.6580 | -0.4341 |
| `x009970` / `x020558` | 0.7762 | 0.1842 | 0.6009 | -0.4168 |
| `x009971` / `x020556` | 0.7453 | 0.1708 | 0.4570 | -0.2862 |

## Correlation diagnostic

regime_js_vs_distance_js_1_5: Pearson 0.9893, Spearman 0.9643 (N=7). These are descriptive diagnostics; ordinary independent-pair p-values are not used.

Detected 3119 neutral distributional change points by a mean-plus-one-standard-deviation JS-jump threshold. Parameter sweeps are retained in YAML and no radius, gap, block size, or number of regimes was selected post hoc.

## Interpretation

Compare original, regime-expected, and local-block-shuffle series. A small residual and high retained effect support shared local composition; a large positive residual and a decrease after block shuffle support additional sequential structure. Retained fractions are ratios, not probabilities.
