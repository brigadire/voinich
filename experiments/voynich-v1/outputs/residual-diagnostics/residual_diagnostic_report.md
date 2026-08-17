# Residual diagnostic report

## Answer

held-out residual means drift across physical blocks; the apparent residual regime is block-specific and not cross-block recurrent; metadata survives leakage-safe whitening, so the current conditioning model remains insufficient.

The frozen diagnostic uses only the preselected `window=500`, `K=2`, k-medoids solution. No new scale, K, method, or token feature search was performed.

Training-fold residual mean is numerical zero (maximum L2 `8.029847394782325e-17`), while the largest held-out mean L2 is `0.09084345856745533`. This distinction is essential: each held-out physical block was transformed using estimates from training blocks only.

The headline Currier values compare the frozen global original baseline (`0.7133`) with the frozen residual winner (`0.7192`). NMI is not monotone under mean centering: centering changes the geometry and the selected residual partition isolates a metadata-pure physical block, so its label association can increase even though its silhouette does not beat the null.

The single-block cluster is exactly cluster `0`: `130` windows, Currier `2`, hand `2`, joint class `2/2`, physical block `2/2#8` (`[14800,21757)`), with window coverage `[14800,21750)`. It is one contiguous run.

## Representation comparison

| representation | silhouette | Currier NMI | hand NMI | joint NMI | block NMI |
|---|---:|---:|---:|---:|---:|
| original_features | 0.1969 | 0.6895 | 0.7818 | 0.6895 | 0.6017 |
| mean_residual | 0.3333 | 0.7192 | 0.5735 | 0.7192 | 0.6513 |
| whitened_residual | 0.4500 | 0.7192 | 0.5735 | 0.7192 | 0.6513 |

Norm-only K=2 reproduces the frozen labels only partially (ARI `0.4133`, NMI `0.4620`), while frozen cluster–physical-block NMI is `0.6513`. Thus magnitude contributes, but the decisive geometry is also block-specific.

## Interpretation guardrails

- Mean centering removes training class means, not held-out drift, covariance, dispersion, sparsity, block identity, or position.
- Whitening uses `0.9 Σ + 0.1 diag(Σ)` and a `1e-6 × largest-eigenvalue` floor, estimated within each training fold only.
- A cluster confined to a contiguous physical block is not called a reproducible regime. Recurrence requires both clusters to pass the training within-cluster reference threshold in held-out blocks.
- Block-aware permutations preserve physical grouping; no random window split is used.

## Decision

held-out residual means drift across physical blocks; the apparent residual regime is block-specific and not cross-block recurrent; metadata survives leakage-safe whitening, so the current conditioning model remains insufficient. Until the stated recurrence and metadata-removal criteria are met, residual discovery should not be expanded.
