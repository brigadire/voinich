# Local-regime / non-stationarity analysis

The corpus is treated as a continuous token sequence. Profiles exclude the central gap; no semantic labels, manuscript metadata, projection, or property trajectories are used.

## Pair decomposition

| pair | regime JS | distance JS 1–5 | expected | residual |
|---|---:|---:|---:|---:|
| `x004624` / `x010315` | 0.8858 | 0.4217 | 0.6811 | -0.2594 |
| `x004984` / `x010279` | 0.8835 | 0.4014 | 0.7895 | -0.3882 |
| `x004625` / `x010314` | 0.8876 | 0.3819 | 0.6854 | -0.3036 |
| `x004985` / `x010279` | 0.8759 | 0.3682 | 0.7663 | -0.3980 |
| `x004719` / `x010526` | 0.9255 | 0.3809 | 0.8486 | -0.4677 |
| `x004718` / `x010526` | 0.9235 | 0.3583 | 0.8286 | -0.4703 |
| `x004719` / `x010527` | 0.9183 | 0.3733 | 0.8376 | -0.4643 |
| `x009496` / `x010315` | 0.8237 | 0.2947 | 0.5843 | -0.2897 |
| `x006383` / `x010054` | 0.8308 | 0.3387 | 0.7232 | -0.3845 |
| `x004718` / `x010054` | 0.8847 | 0.3346 | 0.7057 | -0.3711 |
| `x009497` / `x010315` | 0.8163 | 0.2832 | 0.5334 | -0.2503 |
| `x006348` / `x010055` | 0.8843 | 0.3343 | 0.6812 | -0.3469 |
| `x004718` / `x010055` | 0.8791 | 0.3483 | 0.6973 | -0.3490 |
| `x003821` / `x010055` | 0.8322 | 0.2712 | 0.7126 | -0.4414 |
| `x006383` / `x010055` | 0.8118 | 0.3310 | 0.6709 | -0.3399 |
| `x005730` / `x010617` | 0.7666 | 0.2589 | 0.4988 | -0.2399 |
| `x008248` / `x010617` | 0.7939 | 0.2386 | 0.6139 | -0.3753 |
| `x008353` / `x010617` | 0.7945 | 0.2370 | 0.5745 | -0.3375 |
| `x005731` / `x010617` | 0.7774 | 0.2301 | 0.6055 | -0.3754 |
| `x004929` / `x010054` | 0.8020 | 0.2407 | 0.5919 | -0.3513 |
| `x005842` / `x010617` | 0.7817 | 0.1945 | 0.5373 | -0.3428 |
| `x008352` / `x010617` | 0.7924 | 0.2161 | 0.5753 | -0.3592 |
| `x005843` / `x010617` | 0.7599 | 0.2032 | 0.4715 | -0.2683 |
| `x006050` / `x010484` | 0.7813 | 0.1875 | 0.5480 | -0.3605 |
| `x005730` / `x010485` | 0.7706 | 0.2002 | 0.5570 | -0.3568 |
| `x005843` / `x010616` | 0.7560 | 0.1852 | 0.4562 | -0.2710 |

## Correlation diagnostic

regime_js_vs_distance_js_1_5: Pearson 0.9045, Spearman 0.9002 (N=26). These are descriptive diagnostics; ordinary independent-pair p-values are not used.

Detected 2910 neutral distributional change points by a mean-plus-one-standard-deviation JS-jump threshold. Parameter sweeps are retained in YAML and no radius, gap, block size, or number of regimes was selected post hoc.

## Interpretation

Compare original, regime-expected, and local-block-shuffle series. A small residual and high retained effect support shared local composition; a large positive residual and a decrease after block shuffle support additional sequential structure. Retained fractions are ratios, not probabilities.
