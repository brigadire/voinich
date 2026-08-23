# Soft structural projection analysis

This is a formal structural-neighbourhood experiment. It makes no semantic, grammatical, morphological, or syntactic claims. Token-level results remain the reference and are never replaced by projected metrics.

## Projection and anti-circularity

For an observed token `X`, `W(X,X)=1`. Each eligible neighbour receives `raw structural similarity × evidence reliability`; a row is normalized to sum to one. Thus `P_projected(Y|A,d) = Σ_X P(X|A,d) W_normalized(X,Y)`. The future-context ablation reconstructs weights from position and left-context components only; the past-context ablation analogously excludes left context.

Family projection is reported only as a coarse control and keeps all non-family tokens as singletons. No graphemic quantity is loaded or used.

## Main ranking

| Pair | token JS 1–5 | full projected | ablated | full gain | random percentile | random P95 | smoothing P95 |
|---|---:|---:|---:|---:|---:|---:|---:|
| `x009971` / `x020558` | 0.2095 | 0.2601 | 0.3404 | +0.0507 | P2.0 | +0.0618 | +0.7532 |
| `x009970` / `x020558` | 0.2148 | 0.2630 | 0.3540 | +0.0482 | P0.0 | +0.0611 | +0.7470 |
| `x009969` / `x020558` | 0.2244 | 0.2641 | 0.3511 | +0.0397 | P2.5 | +0.0518 | +0.7353 |
| `x009971` / `x020559` | 0.1798 | 0.2160 | 0.3136 | +0.0361 | P0.0 | +0.0557 | +0.7829 |
| `x009970` / `x020559` | 0.2015 | 0.2369 | 0.3273 | +0.0353 | P0.0 | +0.0538 | +0.7494 |
| `x009971` / `x020556` | 0.1809 | 0.2099 | 0.3067 | +0.0290 | P0.0 | +0.0475 | +0.7836 |
| `x009970` / `x020556` | 0.2046 | 0.2332 | 0.3330 | +0.0286 | P0.0 | +0.0468 | +0.7533 |
| `x009968` / `x020559` | 0.2253 | 0.2528 | 0.3388 | +0.0275 | P0.0 | +0.0401 | +0.7455 |

Ablated random-space and smoothing percentiles are stored alongside every pair and in the controls TSV.

## Prespecified sensitivity sweep

No parameter is selected from these results.

| Method | Parameter | Full mean gain | Ablated mean gain |
|---|---:|---:|---:|
| threshold | 0.50 | +0.1644 | +0.3536 |
| threshold | 0.60 | +0.0851 | +0.2419 |
| threshold | 0.65 | +0.0369 | +0.1280 |
| threshold | 0.70 | +0.0169 | +0.0855 |
| threshold | 0.75 | +0.0000 | +0.0583 |
| knn | 3.00 | +0.1807 | +0.2139 |
| knn | 5.00 | +0.2200 | +0.2590 |
| knn | 10.00 | +0.2621 | +0.3021 |
| knn | 20.00 | +0.2903 | +0.3315 |
| family_control | 0.00 | +0.0082 | +0.0082 |

## Shuffled-corpus controls

| Shuffle | token JS | projected JS | gain |
|---|---:|---:|---:|
| global | 0.1751 | 0.2202 | +0.0451 |
| line-preserving | 0.1607 | 0.2160 | +0.0553 |

## Suffix sequences

`projected_sequence_context.yaml` reports exact suffix JS, each position's projected JS, and their transparent arithmetic-mean sequence kernel for lengths 2 and 3. This tests structural resemblance without constructing a dense Cartesian product.

## Families and transitions

`structural_projection_families.yaml` reports cohesion, within-family dispersion, medoids, and matched percentiles for token/full/ablated spaces. `strongest_structural_transitions` in the pair YAML contains directed expected soft transitions ranked by lift over the product-of-marginals frequency baseline. Family 2 is plotted separately.

## Controls and limitations

Random spaces preserve each row's degree and weights while permuting destinations within log2-frequency bins. Generic smoothing uses the same degree but uniform random neighbours in the same bins. Global and line-preserving corpus shuffles retain token frequencies. Smoothing necessarily tends to increase distributional similarity; only gain beyond these nulls is structurally specific. Structural components were learned elsewhere on the full corpus, so the ablation removes the direct context component but is not a full train/test reconstruction. Cross-validation is intentionally not claimed by this run. Pair observations and distances are dependent; percentiles are deterministic empirical diagnostics, not classical independent-sample p-values.
