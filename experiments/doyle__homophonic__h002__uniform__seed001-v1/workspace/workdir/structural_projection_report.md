# Soft structural projection analysis

This is a formal structural-neighbourhood experiment. It makes no semantic, grammatical, morphological, or syntactic claims. Token-level results remain the reference and are never replaced by projected metrics.

## Projection and anti-circularity

For an observed token `X`, `W(X,X)=1`. Each eligible neighbour receives `raw structural similarity × evidence reliability`; a row is normalized to sum to one. Thus `P_projected(Y|A,d) = Σ_X P(X|A,d) W_normalized(X,Y)`. The future-context ablation reconstructs weights from position and left-context components only; the past-context ablation analogously excludes left context.

Family projection is reported only as a coarse control and keeps all non-family tokens as singletons. No graphemic quantity is loaded or used.

## Main ranking

| Pair | token JS 1–5 | full projected | ablated | full gain | random percentile | random P95 | smoothing P95 |
|---|---:|---:|---:|---:|---:|---:|---:|
| `x005842` / `x010617` | 0.1945 | 0.3229 | 0.4429 | +0.1283 | P82.5 | +0.1333 | +0.6434 |
| `x005843` / `x010616` | 0.1852 | 0.3073 | 0.4465 | +0.1221 | P91.0 | +0.1239 | +0.6528 |
| `x005843` / `x010617` | 0.2032 | 0.3156 | 0.4366 | +0.1125 | P98.0 | +0.1095 | +0.6389 |
| `x008352` / `x010617` | 0.2161 | 0.3280 | 0.4503 | +0.1119 | P31.0 | +0.1254 | +0.6382 |
| `x006050` / `x010484` | 0.1875 | 0.2972 | 0.4413 | +0.1098 | P58.0 | +0.1181 | +0.6512 |
| `x009497` / `x010315` | 0.2832 | 0.3924 | 0.5341 | +0.1093 | P96.0 | +0.1089 | +0.5858 |
| `x009496` / `x010315` | 0.2947 | 0.3978 | 0.5372 | +0.1031 | P22.0 | +0.1125 | +0.5902 |
| `x008353` / `x010617` | 0.2370 | 0.3399 | 0.5158 | +0.1029 | P27.0 | +0.1163 | +0.6151 |
| `x005730` / `x010617` | 0.2589 | 0.3540 | 0.4915 | +0.0952 | P22.0 | +0.1073 | +0.6065 |
| `x004929` / `x010054` | 0.2407 | 0.3355 | 0.4885 | +0.0948 | P57.5 | +0.1013 | +0.6360 |
| `x005730` / `x010485` | 0.2002 | 0.2936 | 0.4620 | +0.0934 | P4.5 | +0.1209 | +0.6354 |
| `x005731` / `x010617` | 0.2301 | 0.3220 | 0.4736 | +0.0920 | P1.0 | +0.1222 | +0.6255 |
| `x008248` / `x010617` | 0.2386 | 0.3260 | 0.4804 | +0.0874 | P1.0 | +0.1110 | +0.5979 |
| `x003821` / `x010055` | 0.2712 | 0.3540 | 0.5145 | +0.0828 | P14.0 | +0.0953 | +0.6090 |
| `x006383` / `x010054` | 0.3387 | 0.4215 | 0.5381 | +0.0828 | P64.5 | +0.0881 | +0.5723 |
| `x006348` / `x010055` | 0.3343 | 0.4168 | 0.5279 | +0.0825 | P25.5 | +0.0883 | +0.5823 |
| `x004718` / `x010054` | 0.3346 | 0.4118 | 0.5293 | +0.0772 | P0.0 | +0.0914 | +0.5977 |
| `x006383` / `x010055` | 0.3310 | 0.4080 | 0.5332 | +0.0770 | P9.0 | +0.0895 | +0.5808 |
| `x004985` / `x010279` | 0.3682 | 0.4448 | 0.5418 | +0.0766 | P0.0 | +0.0907 | +0.5757 |
| `x004718` / `x010055` | 0.3483 | 0.4178 | 0.5323 | +0.0695 | P0.0 | +0.0843 | +0.5772 |
| `x004625` / `x010314` | 0.3819 | 0.4497 | 0.5898 | +0.0679 | P0.5 | +0.0802 | +0.5411 |
| `x004624` / `x010315` | 0.4217 | 0.4856 | 0.6110 | +0.0640 | P0.0 | +0.0746 | +0.5122 |
| `x004984` / `x010279` | 0.4014 | 0.4644 | 0.5609 | +0.0631 | P1.0 | +0.0733 | +0.5508 |
| `x004719` / `x010526` | 0.3809 | 0.4415 | 0.5420 | +0.0606 | P0.5 | +0.0706 | +0.5647 |
| `x004718` / `x010526` | 0.3583 | 0.4178 | 0.5248 | +0.0595 | P0.0 | +0.0717 | +0.5862 |
| `x004719` / `x010527` | 0.3733 | 0.4304 | 0.5272 | +0.0571 | P7.0 | +0.0653 | +0.5714 |

Ablated random-space and smoothing percentiles are stored alongside every pair and in the controls TSV.

## Prespecified sensitivity sweep

No parameter is selected from these results.

| Method | Parameter | Full mean gain | Ablated mean gain |
|---|---:|---:|---:|
| threshold | 0.50 | +0.2318 | +0.3880 |
| threshold | 0.60 | +0.1371 | +0.3008 |
| threshold | 0.65 | +0.0878 | +0.2177 |
| threshold | 0.70 | +0.0471 | +0.1671 |
| threshold | 0.75 | +0.0360 | +0.1193 |
| knn | 3.00 | +0.1936 | +0.2249 |
| knn | 5.00 | +0.2412 | +0.2704 |
| knn | 10.00 | +0.2931 | +0.3202 |
| knn | 20.00 | +0.3312 | +0.3566 |
| family_control | 0.00 | +0.0284 | +0.0284 |

## Shuffled-corpus controls

| Shuffle | token JS | projected JS | gain |
|---|---:|---:|---:|
| global | 0.2373 | 0.3524 | +0.1152 |
| line-preserving | 0.2075 | 0.3195 | +0.1120 |

## Suffix sequences

`projected_sequence_context.yaml` reports exact suffix JS, each position's projected JS, and their transparent arithmetic-mean sequence kernel for lengths 2 and 3. This tests structural resemblance without constructing a dense Cartesian product.

## Families and transitions

`structural_projection_families.yaml` reports cohesion, within-family dispersion, medoids, and matched percentiles for token/full/ablated spaces. `strongest_structural_transitions` in the pair YAML contains directed expected soft transitions ranked by lift over the product-of-marginals frequency baseline. Family 2 is plotted separately.

## Controls and limitations

Random spaces preserve each row's degree and weights while permuting destinations within log2-frequency bins. Generic smoothing uses the same degree but uniform random neighbours in the same bins. Global and line-preserving corpus shuffles retain token frequencies. Smoothing necessarily tends to increase distributional similarity; only gain beyond these nulls is structurally specific. Structural components were learned elsewhere on the full corpus, so the ablation removes the direct context component but is not a full train/test reconstruction. Cross-validation is intentionally not claimed by this run. Pair observations and distances are dependent; percentiles are deterministic empirical diagnostics, not classical independent-sample p-values.
