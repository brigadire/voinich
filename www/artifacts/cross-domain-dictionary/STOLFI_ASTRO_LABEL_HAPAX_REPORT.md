# Positive-label hapax enrichment test

## Result

`POSITIVE_LABEL_HAPAX_ENRICHMENT=DETECTED`.

Of 112 independently matched Stolfi `LABEL` token occurrences, 80 are hapax relative to the complete 901-token Astronomical section. The observed fraction is 0.714285714; the panel-conditioned null mean is 0.611207143. The enrichment ratio is 1.168647524, the absolute difference is 0.103078571, and the one-sided permutation p-value is 0.008799120. The nearest-rank 95% null interval is [0.526785714, 0.687500000].

The result satisfies the frozen decision rule: the effect is positive, `p < 0.05`, and remains positive after every leave-one-panel-out exclusion (8/8_POSITIVE_DIRECTION;8/8_P_LT_0.05). This supports only the tested statement: **independently identified astronomical labels are enriched for section-local hapax**. It does not imply that all hapax are labels.

## Design

- Panels: `f67r1, f67r2, f67v1, f68r1, f68r2, f68r3, f68v2, f68v1` only.
- Confirmed labels: the 112 distinct frozen absolute token positions from rows marked `MATCHED` in `STOLFI_ASTRO_LABEL_MATCHES.tsv`; repeated Stolfi transcriber records and multi-token labels are deduplicated by absolute token position.
- Hapax: token frequency exactly one over all 901 frozen occurrences whose metadata section is `A`; there are 518 such occurrences.
- Null: independently within each panel, sample without replacement exactly the observed number of confirmed label occurrences, then pool the eight samples. The sampling frame contains every panel token, including confirmed labels. No complement is called `NON_LABEL`.
- Permutations: 10,000; base seed `20260901`. Named streams are derived by SHA-256 and sampled with the repository-local SplitMix64 implementation.
- P-value: `(1 + number(null >= observed)) / (B + 1)`, upper-tail. The 95% interval uses empirical nearest-rank 2.5% and 97.5% quantiles.
- No images, spatial interpretation, unmatched-record compensation, or post-hoc inventory completion is used.

## By panel

| panel | labels | hapax | label hapax fraction | panel background fraction |
|---|---:|---:|---:|---:|
| f67r1 | 14 | 5 | 0.357142857 | 0.437125749 |
| f67r2 | 16 | 12 | 0.750000000 | 0.554973822 |
| f67v1 | 6 | 3 | 0.500000000 | 0.671232877 |
| f68r1 | 24 | 20 | 0.833333333 | 0.695652174 |
| f68r2 | 21 | 19 | 0.904761905 | 0.719101124 |
| f68r3 | 8 | 6 | 0.750000000 | 0.609090909 |
| f68v2 | 13 | 8 | 0.615384615 | 0.576923077 |
| f68v1 | 10 | 7 | 0.700000000 | 0.520408163 |

## Secondary families

| family | labels | hapax | observed | null mean | ratio | p (upper) | scope |
|---|---:|---:|---:|---:|---:|---:|---|
| STAR | 54 | 46 | 0.851851852 | 0.690868519 | 1.233015876 | 0.002099790 | ADEQUATELY_REPRESENTED |
| PLANET_MOON | 9 | 6 | 0.666666667 | 0.555100000 | 1.200984808 | 0.371262874 | SMALL_FAMILY_INTERPRET_CAUTIOUSLY |
| CIRCLE_SECTOR | 14 | 5 | 0.357142857 | 0.436200000 | 0.818759416 | 0.810818918 | MODEST_FAMILY_INTERPRET_CAUTIOUSLY |


The secondary results are descriptive robustness checks, not separate decision gates. Star labels are the only well-sized family. Planet/moon and circle-sector results are retained but explicitly treated as small or modest samples. The known unmapped `f68v2.X/Y` series remains outside the confirmed label set and is not compensated post hoc.

## Leave one panel out

All 8/8 exclusions retain a positive observed-minus-null difference; 8/8 remain individually below `p=0.05`. Loss of significance in a smaller LOPO sample is not treated as reversal. Exact values and null intervals are in `STOLFI_ASTRO_LABEL_HAPAX_LOPO.tsv`.

## Provenance and limitations

This test inherits the partial-inventory limitation documented in `STOLFI_ASTRO_LABEL_AUDIT.md`: 112 mapped positive token occurrences are usable, but the complement is not an independently validated negative class. In particular, the unmapped f68v2 sector labels can affect representativeness. The analysis therefore tests enrichment among confirmed positives against panel-matched token draws; it does not estimate sensitivity for all astronomical labels.

No claim is made that all hapax are labels, and no spatial or semantic conclusion is drawn.

## Final status

```text
POSITIVE_LABEL_HAPAX_ENRICHMENT=DETECTED
CONFIRMED_LABEL_OCCURRENCES=112
LABEL_HAPAX_FRACTION=0.714285714
BACKGROUND_EXPECTED_HAPAX_FRACTION=0.611207143
ENRICHMENT_RATIO=1.168647524
PERMUTATION_P=0.008799120
LOPO_STABILITY=8/8_POSITIVE_DIRECTION;8/8_P_LT_0.05
```
