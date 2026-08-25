# G1 generation and stability contract

Generated TOKEN counts are `max(1, floor(scale*N+0.5))` for frozen scales 0.5,
1.0, and 2.0; `N` is the compared partition TOKEN count. Scale 1.0 is primary.
Replicate indices start at zero and checkpoints are 4, 8, 16, 32. At least 8
replicates run and no more than 32 run.

At each checkpoint from 8 onward, compare every applicable metric median with
the preceding checkpoint. A comparison passes when
`abs(current-previous) <= 1e-4 + .01*max(abs(current),abs(previous))` at every
scale. Stop after two consecutive passing comparisons, so the earliest stop is
16. Otherwise run 32 and mark `NON_CONVERGED`, hence `NUMERICALLY_UNSTABLE`.

At 32, coefficient of variation is `sample_sd/abs(mean)`; mean zero uses sample
sd directly. A value above 0.25 is excessive run variation. Any NaN, unexpected
Inf, probability-normalization error above `1e-12`, or undefined-score fraction
above zero is numerical instability. A zero-probability PM1/PM2 is separately
the frozen `HELDOUT_DEGENERATE` outcome.

Complexity growth uses nested DEVELOPMENT leaf prefixes after numeric leaf
sorting at fractions 0.25, 0.50, 0.75, 1.00 (ceil count, duplicates removed).
Fit the Theil-Sen median pairwise slope of `log2 Complexity` on `log2 scored
units`. A deterministic 1,000-resample percentile bootstrap uses the seed
contract and resamples the four scale points with replacement, discarding
samples with fewer than two distinct x values. Sort slopes; the lower endpoint
is nearest-rank index `ceil(.025*n)`. `COMPLEXITY_UNBOUNDED` requires
point slope >1.10 and lower 95% endpoint >1.00.
