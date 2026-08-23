# Task54-1 pre-Voynich audit

Objective frozen as **structural-v2** before any Voynich search. Corpus: Doyle (`data_test/pg2097-2.txt`), 43713 tokens.

## Raw metric calibration

| control | transition | relation | sequence-2 | sequence-3 | raw mean |
|---|---:|---:|---:|---:|---:|
| Doyle | 0.700458 | 0.013249 | 0.520635 | 0.139530 | 0.343468 |
| T2 | 0.679999 | 0.005604 | 0.419770 | 0.048844 | 0.288554 |
| T4 | 0.673464 | 0.002783 | 0.369578 | 0.026698 | 0.268131 |
| T8 | 0.672405 | 0.002440 | 0.357911 | 0.023976 | 0.264183 |

The relation family is on a materially smaller raw scale than the other three families; raw means therefore do not provide approximately equal contributions. `structural-v2` uses candidate-set min-max balancing, fixed before Voynich inspection.

| metric | empirical range (Doyle/T2/T4/T8) | variance |
|---|---:|---:|
| transition | 0.672405..0.700458 | 0.00012723 |
| relation | 0.002440..0.013249 | 0.00001893 |
| sequence-2 | 0.357911..0.520635 | 0.00412206 |
| sequence-3 | 0.023976..0.139530 | 0.00221401 |

## Correlation matrix (synthetic transposition controls)

Pearson correlations over Doyle plus widths 2..16, natural and keyed order, seed 1.

| | transition | relation | sequence-2 | sequence-3 |
|---|---:|---:|---:|---:|
| transition | 1.0000 | 0.9985 | 0.9789 | 0.9903 |
| relation | 0.9985 | 1.0000 | 0.9845 | 0.9912 |
| sequence-2 | 0.9789 | 0.9845 | 1.0000 | 0.9752 |
| sequence-3 | 0.9903 | 0.9912 | 0.9752 | 1.0000 |

The controls are strongly redundant (all off-diagonal correlations exceed 0.97). The four families are retained as pre-registered diagnostics, but each receives one quarter of the balanced objective; no additional data-dependent decorrelation is introduced.

## Blind recovery

Search candidate set: widths 2..16, natural/keyed, one round; the oracle is used only for validation below.

| control | true rank | top-1 | top-3 | exact inverse | structural recovery | random top-1 baseline |
|---|---:|---:|---:|---:|---:|---:|
| T2 | 2 | false | true | true (rank 2) | true | 1/30 = 0.0333 |
| T4 | 1 | true | true | true (rank 1) | true | 1/30 = 0.0333 |
| T8 | 1 | true | true | true (rank 1) | true | 1/30 = 0.0333 |

Synthetic recovery is substantially above the random-search top-1 baseline (2/3 true-parameter top-1; 3/3 top-3 and exact inverse recovery). T2 has an order-equivalent candidate at rank 1; this is a structural equivalence, not a failure of inverse recovery. No Voynich discovery run was performed.
