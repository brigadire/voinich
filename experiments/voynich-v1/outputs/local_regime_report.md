# Local-regime / non-stationarity analysis

The corpus is treated as a continuous token sequence. Profiles exclude the central gap; no semantic labels, manuscript metadata, projection, or property trajectories are used.

## Pair decomposition

| pair | regime JS | distance JS 1–5 | expected | residual |
|---|---:|---:|---:|---:|
| `chedy` / `qokeey` | 0.9195 | 0.4718 | 0.8328 | -0.3609 |
| `chol` / `daiin` | 0.9318 | 0.4432 | 0.8588 | -0.4156 |
| `ol` / `y` | 0.8889 | 0.4329 | 0.8083 | -0.3754 |
| `chey` / `ol` | 0.9204 | 0.4410 | 0.8454 | -0.4044 |
| `dar` / `ol` | 0.9096 | 0.4148 | 0.8330 | -0.4182 |
| `ar` / `ol` | 0.8793 | 0.4049 | 0.7633 | -0.3583 |
| `qokain` / `qol` | 0.9088 | 0.4036 | 0.8243 | -0.4207 |
| `dal` / `ol` | 0.8920 | 0.4093 | 0.8042 | -0.3948 |
| `aiin` / `ar` | 0.9380 | 0.4016 | 0.8788 | -0.4772 |
| `okaiin` / `ol` | 0.8932 | 0.3910 | 0.7843 | -0.3933 |
| `chor` / `daiin` | 0.9016 | 0.3799 | 0.7750 | -0.3951 |
| `or` / `s` | 0.8771 | 0.4041 | 0.7936 | -0.3895 |
| `qokaiin` / `qol` | 0.8402 | 0.3689 | 0.6925 | -0.3235 |
| `lchedy` / `qokeey` | 0.8808 | 0.3624 | 0.7615 | -0.3990 |
| `qokedy` / `qol` | 0.8830 | 0.3631 | 0.7919 | -0.4289 |
| `lchedy` / `qokain` | 0.8824 | 0.3702 | 0.7419 | -0.3717 |
| `qokar` / `qol` | 0.8205 | 0.3597 | 0.6788 | -0.3191 |
| `r` / `s` | 0.8669 | 0.3672 | 0.7951 | -0.4279 |
| `okain` / `ol` | 0.8773 | 0.3445 | 0.7526 | -0.4082 |
| `or` / `r` | 0.8605 | 0.3472 | 0.7378 | -0.3906 |
| `lchedy` / `qol` | 0.8621 | 0.3328 | 0.7431 | -0.4103 |
| `lchedy` / `qokar` | 0.8195 | 0.3086 | 0.6666 | -0.3579 |
| `daiin` / `dol` | 0.8761 | 0.3045 | 0.7310 | -0.4265 |
| `chol` / `cthy` | 0.8309 | 0.3033 | 0.7136 | -0.4103 |
| `ain` / `ar` | 0.8542 | 0.2952 | 0.7107 | -0.4154 |
| `qol` / `qotain` | 0.8248 | 0.2654 | 0.6397 | -0.3743 |
| `ain` / `al` | 0.8472 | 0.2539 | 0.7195 | -0.4655 |
| `okar` / `otain` | 0.8225 | 0.2372 | 0.6821 | -0.4449 |

## Correlation diagnostic

regime_js_vs_distance_js_1_5: Pearson 0.7977, Spearman 0.8216 (N=28). These are descriptive diagnostics; ordinary independent-pair p-values are not used.

Detected 2481 neutral distributional change points by a mean-plus-one-standard-deviation JS-jump threshold. Parameter sweeps are retained in YAML and no radius, gap, block size, or number of regimes was selected post hoc.

## Interpretation

Compare original, regime-expected, and local-block-shuffle series. A small residual and high retained effect support shared local composition; a large positive residual and a decrease after block shuffle support additional sequential structure. Retained fractions are ratios, not probabilities.
