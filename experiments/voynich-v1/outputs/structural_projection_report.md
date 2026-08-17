# Soft structural projection analysis

This is a formal structural-neighbourhood experiment. It makes no semantic, grammatical, morphological, or syntactic claims. Token-level results remain the reference and are never replaced by projected metrics.

## Projection and anti-circularity

For an observed token `X`, `W(X,X)=1`. Each eligible neighbour receives `raw structural similarity × evidence reliability`; a row is normalized to sum to one. Thus `P_projected(Y|A,d) = Σ_X P(X|A,d) W_normalized(X,Y)`. The future-context ablation reconstructs weights from position and left-context components only; the past-context ablation analogously excludes left context.

Family projection is reported only as a coarse control and keeps all non-family tokens as singletons. No graphemic quantity is loaded or used.

## Main ranking

| Pair | token JS 1–5 | full projected | ablated | full gain | random percentile | random P95 | smoothing P95 |
|---|---:|---:|---:|---:|---:|---:|---:|
| `qol` / `qotain` | 0.2654 | 0.4051 | 0.5902 | +0.1397 | P47.0 | +0.1474 | +0.6206 |
| `lchedy` / `qol` | 0.3328 | 0.4670 | 0.5998 | +0.1343 | P31.5 | +0.1419 | +0.6007 |
| `qokedy` / `qol` | 0.3631 | 0.4883 | 0.6155 | +0.1252 | P0.5 | +0.1371 | +0.5834 |
| `qokar` / `qol` | 0.3597 | 0.4793 | 0.6238 | +0.1196 | P7.0 | +0.1314 | +0.5876 |
| `lchedy` / `qokar` | 0.3086 | 0.4178 | 0.5605 | +0.1092 | P0.0 | +0.1273 | +0.6357 |
| `qokaiin` / `qol` | 0.3689 | 0.4779 | 0.6243 | +0.1089 | P16.0 | +0.1170 | +0.5716 |
| `okar` / `otain` | 0.2372 | 0.3447 | 0.5088 | +0.1076 | P0.0 | +0.1287 | +0.6961 |
| `qokain` / `qol` | 0.4036 | 0.5095 | 0.6504 | +0.1059 | P2.0 | +0.1167 | +0.5441 |
| `lchedy` / `qokeey` | 0.3624 | 0.4553 | 0.5866 | +0.0929 | P0.0 | +0.1049 | +0.5952 |
| `ain` / `al` | 0.2539 | 0.3445 | 0.4750 | +0.0905 | P0.0 | +0.1024 | +0.6992 |
| `lchedy` / `qokain` | 0.3702 | 0.4592 | 0.5854 | +0.0890 | P0.0 | +0.1075 | +0.5852 |
| `ain` / `ar` | 0.2952 | 0.3803 | 0.5062 | +0.0851 | P0.5 | +0.0947 | +0.6622 |
| `okain` / `ol` | 0.3445 | 0.4266 | 0.5630 | +0.0821 | P0.0 | +0.0994 | +0.6182 |
| `daiin` / `dol` | 0.3045 | 0.3839 | 0.5027 | +0.0795 | P0.0 | +0.0902 | +0.6570 |
| `or` / `r` | 0.3472 | 0.4168 | 0.5285 | +0.0696 | P1.0 | +0.0779 | +0.6210 |
| `okaiin` / `ol` | 0.3910 | 0.4557 | 0.5765 | +0.0646 | P0.0 | +0.0762 | +0.5711 |
| `r` / `s` | 0.3672 | 0.4300 | 0.5371 | +0.0628 | P3.5 | +0.0700 | +0.5978 |
| `chol` / `cthy` | 0.3033 | 0.3604 | 0.4933 | +0.0571 | P1.0 | +0.0677 | +0.6586 |
| `dal` / `ol` | 0.4093 | 0.4656 | 0.5655 | +0.0562 | P0.0 | +0.0645 | +0.5629 |
| `or` / `s` | 0.4041 | 0.4580 | 0.5485 | +0.0539 | P1.0 | +0.0609 | +0.5688 |
| `dar` / `ol` | 0.4148 | 0.4664 | 0.5655 | +0.0516 | P0.0 | +0.0609 | +0.5559 |
| `chey` / `ol` | 0.4410 | 0.4885 | 0.5842 | +0.0475 | P0.0 | +0.0609 | +0.5342 |
| `chedy` / `qokeey` | 0.4718 | 0.5168 | 0.6102 | +0.0449 | P0.0 | +0.0578 | +0.5079 |
| `ar` / `ol` | 0.4049 | 0.4497 | 0.5520 | +0.0448 | P0.0 | +0.0579 | +0.5612 |
| `ol` / `y` | 0.4329 | 0.4754 | 0.5746 | +0.0425 | P0.0 | +0.0527 | +0.5420 |
| `aiin` / `ar` | 0.4016 | 0.4434 | 0.5364 | +0.0418 | P0.0 | +0.0509 | +0.5694 |
| `chor` / `daiin` | 0.3799 | 0.4213 | 0.5265 | +0.0414 | P0.0 | +0.0550 | +0.5930 |
| `chol` / `daiin` | 0.4432 | 0.4759 | 0.5618 | +0.0327 | P0.0 | +0.0428 | +0.5359 |

Ablated random-space and smoothing percentiles are stored alongside every pair and in the controls TSV.

## Prespecified sensitivity sweep

No parameter is selected from these results.

| Method | Parameter | Full mean gain | Ablated mean gain |
|---|---:|---:|---:|
| threshold | 0.50 | +0.1984 | +0.3323 |
| threshold | 0.60 | +0.1110 | +0.2642 |
| threshold | 0.65 | +0.0779 | +0.1989 |
| threshold | 0.70 | +0.0385 | +0.1320 |
| threshold | 0.75 | +0.0172 | +0.0839 |
| knn | 3.00 | +0.1858 | +0.2157 |
| knn | 5.00 | +0.2199 | +0.2529 |
| knn | 10.00 | +0.2573 | +0.2917 |
| knn | 20.00 | +0.2831 | +0.3182 |
| family_control | 0.00 | +0.0527 | +0.0527 |

## Shuffled-corpus controls

| Shuffle | token JS | projected JS | gain |
|---|---:|---:|---:|
| global | 0.3193 | 0.3898 | +0.0705 |
| line-preserving | 0.3192 | 0.4156 | +0.0964 |

## Suffix sequences

`projected_sequence_context.yaml` reports exact suffix JS, each position's projected JS, and their transparent arithmetic-mean sequence kernel for lengths 2 and 3. This tests structural resemblance without constructing a dense Cartesian product.

## Families and transitions

`structural_projection_families.yaml` reports cohesion, within-family dispersion, medoids, and matched percentiles for token/full/ablated spaces. `strongest_structural_transitions` in the pair YAML contains directed expected soft transitions ranked by lift over the product-of-marginals frequency baseline. Family 2 is plotted separately.

## Controls and limitations

Random spaces preserve each row's degree and weights while permuting destinations within log2-frequency bins. Generic smoothing uses the same degree but uniform random neighbours in the same bins. Global and line-preserving corpus shuffles retain token frequencies. Smoothing necessarily tends to increase distributional similarity; only gain beyond these nulls is structurally specific. Structural components were learned elsewhere on the full corpus, so the ablation removes the direct context component but is not a full train/test reconstruction. Cross-validation is intentionally not claimed by this run. Pair observations and distances are dependent; percentiles are deterministic empirical diagnostics, not classical independent-sample p-values.
