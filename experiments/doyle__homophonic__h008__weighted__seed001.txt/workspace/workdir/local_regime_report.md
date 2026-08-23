# Local-regime / non-stationarity analysis

The corpus is treated as a continuous token sequence. Profiles exclude the central gap; no semantic labels, manuscript metadata, projection, or property trajectories are used.

## Pair decomposition

| pair | regime JS | distance JS 1–5 | expected | residual |
|---|---:|---:|---:|---:|

## Correlation diagnostic

regime_js_vs_distance_js_1_5: Pearson 0.0000, Spearman 0.0000 (N=0). These are descriptive diagnostics; ordinary independent-pair p-values are not used.

Detected 3237 neutral distributional change points by a mean-plus-one-standard-deviation JS-jump threshold. Parameter sweeps are retained in YAML and no radius, gap, block size, or number of regimes was selected post hoc.

## Interpretation

Compare original, regime-expected, and local-block-shuffle series. A small residual and high retained effect support shared local composition; a large positive residual and a decrease after block shuffle support additional sequential structure. Retained fractions are ratios, not probabilities.
