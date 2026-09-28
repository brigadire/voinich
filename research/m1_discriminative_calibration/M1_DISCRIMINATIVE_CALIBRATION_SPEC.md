# M1 discriminative calibration report

## Formal metric definitions

`TRAIN_RECOVERY = matched_train_labels / train_labels`. `HELDOUT_RECOVERY = matched_heldout_labels / heldout_labels`, with held-out rows excluded from search. `EXACT_PARAMETER_RECOVERY` is the fraction of source mappings identical to the hidden table. `FUNCTIONAL_MAPPING_RECOVERY` counts mappings equivalent on all generated outputs. `OUTPUT_RECOVERY` is the fraction of generated target tokens reproduced; `ASSIGNMENT_RECOVERY` is the fraction of source-label assignments recovered. `RANDOM_NULL_P95` is the 95th percentile of `MAX_SCORE_OVER_SEARCH` across independent matched null replicates. `DISCRIMINATION_GAP = positive_recovery - matched_random_null`. `SINGLETON_FRACTION = unsupported_or_singleton_mappings / all_mappings`.

## Mandatory calibration result

The frozen grid does not produce an independently validated preferred profile. Restricting mapping size/support/two-unit mappings suppresses nulls only by also degrading S1–S5 recovery; less restrictive profiles retain high null scores. The apparent S0 gap is not stable enough under the frozen acceptance rules.

```text
M1_DISCRIMINATIVE_CALIBRATION=MODEL_FAMILY_NON_DISCRIMINATIVE
BEST_PROFILE_ID=NONE
POSITIVE_VALIDATION_MEDIAN=0.850000
POSITIVE_VALIDATION_P05=0.600000
HELDOUT_VALIDATION_MEDIAN=0.800000
RANDOM_NULL_P95=0.750000
RANDOM_NULL_MAX=0.900000
DISCRIMINATION_GAP=0.100000
SUPPORTED_SYNTHETIC_FAMILIES=S0_ONLY_WITHOUT_ACCEPTANCE
REAL_DICTIONARY_SEARCH_AUTHORIZED=NO
```

Calibration and validation seeds are disjoint (1000-series vs 2000-series). No profile meets the preferred or weaker acceptance rule. Real-dictionary brute force remains prohibited.
