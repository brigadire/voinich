# Blind metadata validation

Alignment result: **PASS**. Validation category: **Weak alignment**. This category describes association with reference metadata, not accuracy against ground truth.

The frozen IVTT `-x7 ASCII Full` sequence (39026 tokens) and all frozen distributional boundaries and cluster assignments were used unchanged. The IVTFF parser supplies metadata only. Corpus SHA256 recorded during this run: `360d99583145ec549b80edfafdc3f93534f3a11b85a0d52997ba8425e92b87c2`; it was not historically stored by discovery.

## Boundary association

| Metadata | support | tolerance | matched / blind | uniform percentile | circular percentile |
|---|---:|---:|---:|---:|---:|
| line | ≥4/5 | ±50 | 32/32 | 100.0 | 100.0 |
| paragraph | ≥4/5 | ±50 | 29/32 | 49.3 | 50.6 |
| folio | ≥4/5 | ±50 | 17/32 | 60.5 | 56.5 |
| currier | ≥4/5 | ±50 | 1/32 | 38.2 | 43.5 |
| hand | ≥4/5 | ±50 | 4/32 | 87.0 | 83.5 |
| quire | ≥4/5 | ±50 | 4/32 | 98.7 | 98.6 |

Line transitions are a dense sanity control and are not interpreted from raw overlap alone. Known→unknown and unknown→known changes are excluded from Currier, hand and quire transitions. Fixed tolerances were prespecified and are all retained.

## Frozen cluster association

`cluster_metadata_association.tsv` reports MI, NMI, ARI, homogeneity, completeness and conditional entropy for every frozen window scale, method and K, both for all known-majority windows and purity ≥0.8/≥0.9 subsets. Metadata never changes clusters. The max-over-K permutation control shuffles labels among contiguous blocks while retaining block lengths.

## Residual structure

`unexplained_distributional_structure.tsv` contains support ≥4/5 boundaries farther than 200 tokens from every available tested transition. These are neutral unexplained distributional structures, not proposed sections or languages.

<!-- BEGIN GLOBAL MULTIPLE-COMPARISON CORRECTION (cluster-metadata-global) -->
## Global multiple-comparison correction

This section corrects the observed association between frozen blind distributional regimes and Currier/hand metadata for every choice that was already available before metadata was consulted: window size in {50, 100, 200, 500, 1000}, method in {contiguous_segmentation, hierarchical, k_medoids} and K in 2..15. Discovery (windows, clustering, boundaries) is unchanged from `global_distributional_*`; only metadata labels are permuted, using 10000 block-aware permutations (seed 1) that preserve contiguous metadata block lengths and the unknown-token mask exactly, so the same set of valid windows is compared for every observed and null statistic. Empirical p-values use `(exceedances + 1) / (permutations + 1)` and are never reported as exactly zero.

### Currier

Per-method correction (max NMI over window size x K, frozen K=2..15):

| Method | Observed NMI | window | K | null mean | null P95 | null P99 | empirical p |
|---|---:|---:|---:|---:|---:|---:|---:|
| contiguous_segmentation | 0.713 | 50 | 11 | 0.485 | 0.618 | 0.658 | 0.0002 |
| hierarchical | 0.659 | 1000 | 15 | 0.446 | 0.581 | 0.625 | 0.0034 |
| k_medoids | 0.664 | 1000 | 8 | 0.500 | 0.621 | 0.659 | 0.0071 |

Global correction (max NMI over window size x method x K, the complete frozen search space): observed 0.713 at window=50, method=contiguous_segmentation, K=11; null mean 0.511, P95 0.633, P99 0.671, empirical p 0.0008.

Scale-persistence (mean and minimum of the five prespecified scale-specific max-over-K values, no maximum selection across scales):

| Method | mean-across-scales | empirical p | min-across-scales | empirical p |
|---|---:|---:|---:|---:|
| contiguous_segmentation | 0.659 | 0.0026 | 0.629 | 0.0046 |
| hierarchical | 0.323 | 0.0003 | 0.075 | 0.0207 |
| k_medoids | 0.591 | <0.0001 | 0.451 | 0.0002 |

k-medoids remains significant after correcting for window size and K within the frozen k-medoids search space (empirical p=0.0071): association is not explained by post-hoc selection of window size or K within that method. The global maximum remains significant after correcting across the complete frozen window size x method x K search space (empirical p=0.0008): the association survives correction across every clustering choice considered before metadata was consulted. The scale-persistence statistic is also significant (empirical p=<0.0001): the association is reproduced across multiple prespecified scales, not concentrated in a single window size.

### Hand

Per-method correction (max NMI over window size x K, frozen K=2..15):

| Method | Observed NMI | window | K | null mean | null P95 | null P99 | empirical p |
|---|---:|---:|---:|---:|---:|---:|---:|
| contiguous_segmentation | 0.543 | 500 | 12 | 0.456 | 0.562 | 0.598 | 0.0958 |
| hierarchical | 0.631 | 500 | 14 | 0.466 | 0.576 | 0.615 | 0.0054 |
| k_medoids | 0.598 | 200 | 5 | 0.487 | 0.586 | 0.618 | 0.0295 |

Global correction (max NMI over window size x method x K, the complete frozen search space): observed 0.631 at window=500, method=hierarchical, K=14; null mean 0.495, P95 0.596, P99 0.631, empirical p 0.0104.

Scale-persistence (mean and minimum of the five prespecified scale-specific max-over-K values, no maximum selection across scales):

| Method | mean-across-scales | empirical p | min-across-scales | empirical p |
|---|---:|---:|---:|---:|
| contiguous_segmentation | 0.530 | 0.0651 | 0.519 | 0.0498 |
| hierarchical | 0.345 | 0.0009 | 0.129 | 0.0228 |
| k_medoids | 0.540 | <0.0001 | 0.403 | <0.0001 |

k-medoids remains significant after correcting for window size and K within the frozen k-medoids search space (empirical p=0.0295): association is not explained by post-hoc selection of window size or K within that method. The global maximum remains significant after correcting across the complete frozen window size x method x K search space (empirical p=0.0104): the association survives correction across every clustering choice considered before metadata was consulted. The scale-persistence statistic is also significant (empirical p=<0.0001): the association is reproduced across multiple prespecified scales, not concentrated in a single window size.

### Purity sensitivity analysis

The same global correction, repeated over purity >= 0.8 and purity >= 0.9 mixed-window subsets. These thresholds were fixed in advance and are reported as **sensitivity analysis**, not primary evidence; the primary test above always uses all windows with a known majority metadata label.

| Metadata | Scope | Method | Observed max NMI | empirical p | Global max NMI | empirical p |
|---|---|---|---:|---:|---:|---:|
| currier | purity_0.8 | contiguous_segmentation | 0.822 | 0.0038 | 0.829 | 0.0268 |
| currier | purity_0.8 | hierarchical | 0.828 | 0.0194 | 0.829 | 0.0268 |
| currier | purity_0.8 | k_medoids | 0.829 | 0.0073 | 0.829 | 0.0268 |
| currier | purity_0.9 | contiguous_segmentation | 0.857 | 0.0083 | 0.857 | 0.0372 |
| currier | purity_0.9 | hierarchical | 0.823 | 0.0555 | 0.857 | 0.0372 |
| currier | purity_0.9 | k_medoids | 0.829 | 0.0218 | 0.857 | 0.0372 |
| hand | purity_0.8 | contiguous_segmentation | 0.659 | 0.1375 | 0.745 | 0.0862 |
| hand | purity_0.8 | hierarchical | 0.745 | 0.0431 | 0.745 | 0.0862 |
| hand | purity_0.8 | k_medoids | 0.738 | 0.0527 | 0.745 | 0.0862 |
| hand | purity_0.9 | contiguous_segmentation | 0.714 | 0.1532 | 0.773 | 0.1502 |
| hand | purity_0.9 | hierarchical | 0.773 | 0.0722 | 0.773 | 0.1502 |
| hand | purity_0.9 | k_medoids | 0.754 | 0.1061 | 0.773 | 0.1502 |

Secondary metric: the same three primary statistics (per-method max, global max, scale persistence) were also computed for Adjusted Rand Index; see `cluster_metadata_global_summary.tsv` and `cluster_metadata_scale_persistence.tsv` (metric=ARI) for full values, at every prespecified scope.

<!-- END GLOBAL MULTIPLE-COMPARISON CORRECTION (cluster-metadata-global) -->
