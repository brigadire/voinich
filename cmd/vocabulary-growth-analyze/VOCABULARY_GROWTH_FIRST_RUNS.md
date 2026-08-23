# Task49 first scientific runs

These are the first fixed-parameter runs (`seed=1`, `null-permutations=10`)
on the existing corpus inputs. The table is descriptive and is not a class
diagnosis. `NA` means that the checkpoint exceeds that corpus length.

| corpus | N | V(N) | V(1000) | V(2000) | V(4000) | V(8000) | V(16000) | V(32000) | Heaps beta | Heaps R² | final hapax | hapax/type |
|---|---:|---:|---:|---:|---:|---:|---:|---:|---:|---:|---:|---:|
| Voynich | 39026 | 8363 | 550 | 903 | 1472 | 2504 | 4605 | 7359 | 0.761401 | 0.999530 | 5863 | 0.7013 |
| Doyle | 43713 | 5360 | 475 | 779 | 1290 | 2082 | 3093 | 4604 | 0.700603 | 0.995798 | 2852 | 0.5321 |
| Longfellow | 33077 | 4975 | 438 | 742 | 1255 | 2050 | 3182 | 4803 | 0.774558 | 0.994051 | 2304 | 0.4631 |
| MS-DOS 2.0 | 112162 | 7169 | 192 | 343 | 563 | 917 | 1627 | 2868 | 0.706636 | 0.992899 | 873 | 0.1217 |
| Astafiev recipes | 85280 | 6576 | 557 | 1012 | 1514 | 1904 | 3003 | 4008 | 0.624694 | 0.982709 | 3014 | 0.4583 |

The values were read from the machine-readable stage outputs. They are
included as a reproducibility snapshot; changing any fixed parameter requires
a new experiment generation/version.
