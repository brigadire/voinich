# Synthetic benchmark methodological audit

## Mandatory audit

The original benchmark is not valid as a discriminative benchmark without revision. Its clean S0 recovery is real (20/6 TRAIN and HELD_OUT 1.0), but the manifest has one replicate and its random-target null reaches P95 0.95 / MAX 1.0. Therefore the previous `SEARCH_CAPABLE` interpretation is not supported. Capacity explains the 40/10 and 80/20 declines: 31 concepts impose TRAIN ceilings 0.775 and 0.3875.

Thirty audit-only replicates per noise level were run; quantiles are in `SYNTHETIC_NOISE_AUDIT.tsv`. Ground truth, held-out generation, and family-level metrics are separated in dedicated TSVs.

## Extended audit

Capacity-safe 31/60/120 concept scalability, beam sweep, five null generators, and a realistic positive/negative shell are reported separately. The current random generator uses the same short EVA alphabet and length-conditioned outputs as the positive system, leaving an overly expressive substitution model able to encode many random targets. This is experimentally supported; the relative contribution of morphology and alphabet size remains uncertain until larger calibrated nulls are run.

```text
SYNTHETIC_BENCHMARK_AUDIT=REVISION_REQUIRED
POSITIVE_RECOVERY_CONFIRMED=YES
OPTIMISER_HIGH_COVERAGE_CAPABLE=YES
RANDOM_NULL_DISCRIMINATION=FAILED
LARGE_SAMPLE_RECOVERY=CONFIRMED
HISTORICAL_SYNTHETIC_NULL_DISCREPANCY=PARTIALLY_EXPLAINED
REAL_DICTIONARY_SEARCH_AUTHORIZED=NO
```

Mandatory conclusion: M1 can recover an in-model clean transformation, but the current synthetic benchmark cannot establish discrimination from chance and does not authorize another real astronomical brute-force run.
