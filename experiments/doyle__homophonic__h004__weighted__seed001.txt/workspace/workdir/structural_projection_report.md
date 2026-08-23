# Soft structural projection analysis

This is a formal structural-neighbourhood experiment. It makes no semantic, grammatical, morphological, or syntactic claims. Token-level results remain the reference and are never replaced by projected metrics.

## Projection and anti-circularity

For an observed token `X`, `W(X,X)=1`. Each eligible neighbour receives `raw structural similarity × evidence reliability`; a row is normalized to sum to one. Thus `P_projected(Y|A,d) = Σ_X P(X|A,d) W_normalized(X,Y)`. The future-context ablation reconstructs weights from position and left-context components only; the past-context ablation analogously excludes left context.

Family projection is reported only as a coarse control and keeps all non-family tokens as singletons. No graphemic quantity is loaded or used.

## Main ranking

| Pair | token JS 1–5 | full projected | ablated | full gain | random percentile | random P95 | smoothing P95 |
|---|---:|---:|---:|---:|---:|---:|---:|
| `x009970` / `x020558` | 0.1842 | 0.2594 | 0.3576 | +0.0753 | P5.0 | +0.0884 | +0.7443 |
| `x009969` / `x020558` | 0.2239 | 0.2897 | 0.3792 | +0.0658 | P8.0 | +0.0767 | +0.7170 |
| `x009970` / `x020556` | 0.2162 | 0.2759 | 0.3552 | +0.0597 | P5.0 | +0.0708 | +0.7225 |
| `x009971` / `x020556` | 0.1708 | 0.2250 | 0.3298 | +0.0542 | P0.5 | +0.0689 | +0.7580 |
| `x009968` / `x020557` | 0.2645 | 0.3140 | 0.3937 | +0.0495 | P19.0 | +0.0558 | +0.6932 |
| `x009969` / `x020557` | 0.2518 | 0.3003 | 0.3924 | +0.0485 | P0.0 | +0.0679 | +0.6965 |
| `x009436` / `x012697` | 0.3391 | 0.3767 | 0.4462 | +0.0376 | P4.5 | +0.0438 | +0.6337 |

Ablated random-space and smoothing percentiles are stored alongside every pair and in the controls TSV.

## Prespecified sensitivity sweep

No parameter is selected from these results.

| Method | Parameter | Full mean gain | Ablated mean gain |
|---|---:|---:|---:|
| threshold | 0.50 | +0.1698 | +0.3466 |
| threshold | 0.60 | +0.0840 | +0.2332 |
| threshold | 0.65 | +0.0558 | +0.1434 |
| threshold | 0.70 | +0.0215 | +0.1012 |
| threshold | 0.75 | +0.0084 | +0.0668 |
| knn | 3.00 | +0.1750 | +0.2094 |
| knn | 5.00 | +0.2124 | +0.2502 |
| knn | 10.00 | +0.2505 | +0.2910 |
| knn | 20.00 | +0.2803 | +0.3221 |
| family_control | 0.00 | +0.0078 | +0.0078 |

## Shuffled-corpus controls

| Shuffle | token JS | projected JS | gain |
|---|---:|---:|---:|
| global | 0.2109 | 0.2750 | +0.0641 |
| line-preserving | 0.1804 | 0.2500 | +0.0696 |

## Suffix sequences

`projected_sequence_context.yaml` reports exact suffix JS, each position's projected JS, and their transparent arithmetic-mean sequence kernel for lengths 2 and 3. This tests structural resemblance without constructing a dense Cartesian product.

## Families and transitions

`structural_projection_families.yaml` reports cohesion, within-family dispersion, medoids, and matched percentiles for token/full/ablated spaces. `strongest_structural_transitions` in the pair YAML contains directed expected soft transitions ranked by lift over the product-of-marginals frequency baseline. Family 2 is plotted separately.

## Controls and limitations

Random spaces preserve each row's degree and weights while permuting destinations within log2-frequency bins. Generic smoothing uses the same degree but uniform random neighbours in the same bins. Global and line-preserving corpus shuffles retain token frequencies. Smoothing necessarily tends to increase distributional similarity; only gain beyond these nulls is structurally specific. Structural components were learned elsewhere on the full corpus, so the ablation removes the direct context component but is not a full train/test reconstruction. Cross-validation is intentionally not claimed by this run. Pair observations and distances are dependent; percentiles are deterministic empirical diagnostics, not classical independent-sample p-values.
