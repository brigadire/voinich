# Synthetic recovery benchmark

The frozen M1 optimiser was evaluated on generated STAR-only labels with hidden mappings. Ground truth was not passed to search. S6_CONTEXTUAL is explicitly out of model space. Clean S0 at 20/6 recovered TRAIN 1.000 and HELD_OUT 1.000; detailed family, noise, dictionary and beam results are in TSV.

The critical negative control failed: random targets reached P95=0.950 and maximum=1.000. Thus high recovery is not discriminative at this sample size; positive and random controls overlap materially.

```text
ASTRO_SEARCH_RECOVERY_BENCHMARK=NON_DISCRIMINATIVE
CLEAN_RECOVERY_TRAIN=1.000000
CLEAN_RECOVERY_HELDOUT=1.000000
NOISE20_RECOVERY=0.656250
NOISE30_RECOVERY=0.714286
RANDOM_NULL_P95=0.950000
RANDOM_NULL_MAX=1.000000
GROUND_TRUTH_MAPPING_RECOVERY=1.000000
SAMPLE_20_6_RECOVERY=1.000000
SAMPLE_40_10_RECOVERY=0.775000
SAMPLE_80_20_RECOVERY=0.362500
REAL_DICTIONARY_SEARCH_AUTHORIZED=NO
```

Because the random-null separation criterion fails, no real astronomical dictionary search is authorized by this benchmark.
