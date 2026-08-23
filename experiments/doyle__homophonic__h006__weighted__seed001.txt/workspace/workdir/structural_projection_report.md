# Soft structural projection analysis

This is a formal structural-neighbourhood experiment. It makes no semantic, grammatical, morphological, or syntactic claims. Token-level results remain the reference and are never replaced by projected metrics.

## Projection and anti-circularity

For an observed token `X`, `W(X,X)=1`. Each eligible neighbour receives `raw structural similarity × evidence reliability`; a row is normalized to sum to one. Thus `P_projected(Y|A,d) = Σ_X P(X|A,d) W_normalized(X,Y)`. The future-context ablation reconstructs weights from position and left-context components only; the past-context ablation analogously excludes left context.

Family projection is reported only as a coarse control and keeps all non-family tokens as singletons. No graphemic quantity is loaded or used.

## Main ranking

| Pair | token JS 1–5 | full projected | ablated | full gain | random percentile | random P95 | smoothing P95 |
|---|---:|---:|---:|---:|---:|---:|---:|
| `x014952` / `x030834` | 0.1911 | 0.1978 | 0.2921 | +0.0067 | P0.0 | +0.0160 | +0.7957 |

Ablated random-space and smoothing percentiles are stored alongside every pair and in the controls TSV.

## Prespecified sensitivity sweep

No parameter is selected from these results.

| Method | Parameter | Full mean gain | Ablated mean gain |
|---|---:|---:|---:|
| threshold | 0.50 | +0.1268 | +0.3102 |
| threshold | 0.60 | +0.0338 | +0.1947 |
| threshold | 0.65 | +0.0067 | +0.1011 |
| threshold | 0.70 | +0.0008 | +0.0610 |
| threshold | 0.75 | +0.0000 | +0.0300 |
| knn | 3.00 | +0.1491 | +0.1890 |
| knn | 5.00 | +0.1816 | +0.2294 |
| knn | 10.00 | +0.2174 | +0.2707 |
| knn | 20.00 | +0.2404 | +0.2957 |
| family_control | 0.00 | +0.0013 | +0.0013 |

## Shuffled-corpus controls

| Shuffle | token JS | projected JS | gain |
|---|---:|---:|---:|
| global | 0.1976 | 0.2172 | +0.0197 |
| line-preserving | 0.1456 | 0.1580 | +0.0124 |

## Suffix sequences

`projected_sequence_context.yaml` reports exact suffix JS, each position's projected JS, and their transparent arithmetic-mean sequence kernel for lengths 2 and 3. This tests structural resemblance without constructing a dense Cartesian product.

## Families and transitions

`structural_projection_families.yaml` reports cohesion, within-family dispersion, medoids, and matched percentiles for token/full/ablated spaces. `strongest_structural_transitions` in the pair YAML contains directed expected soft transitions ranked by lift over the product-of-marginals frequency baseline. Family 2 is plotted separately.

## Controls and limitations

Random spaces preserve each row's degree and weights while permuting destinations within log2-frequency bins. Generic smoothing uses the same degree but uniform random neighbours in the same bins. Global and line-preserving corpus shuffles retain token frequencies. Smoothing necessarily tends to increase distributional similarity; only gain beyond these nulls is structurally specific. Structural components were learned elsewhere on the full corpus, so the ablation removes the direct context component but is not a full train/test reconstruction. Cross-validation is intentionally not claimed by this run. Pair observations and distances are dependent; percentiles are deterministic empirical diagnostics, not classical independent-sample p-values.
