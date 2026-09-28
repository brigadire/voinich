# M2 morphology/syllable-constrained audit

## Formal metrics

`TOKEN_COVERAGE = matched labels / labels`; `HELDOUT_COVERAGE` uses rows excluded from search. `RULE_PRECISION = recovered true rules / inferred rules`; `RULE_RECALL = recovered true rules / hidden rules`; `FUNCTIONAL_RULE_RECOVERY` counts rules equivalent on all generated outputs; `ASSIGNMENT_RECOVERY` is recovered concept-label assignments. `MEAN_RULE_SUPPORT` is the mean number of independent concepts supporting an inferred rule. `SINGLETON_FRACTION` is singleton/total rules. `COMPRESSION_RATIO = explained description length / total description length`. `DISCRIMINATION_GAP = positive median - matched-null P95`.

## Mandatory synthetic gate

M2 uses recurrent units, role-preserving composition, minimum independent support, singleton prohibition and compression scoring. Positive and negative corpora preserve surface statistics and use disjoint 1000-series calibration and 2000-series validation seeds. M2 lowers the null relative to M1, but no profile meets the frozen acceptance thresholds: validation null P95 remains 0.68 and the gap is 0.12. Recovery outside P0 is also insufficient.

```text
M2_TOKEN_FORMATION_MODEL=MODEL_FAMILY_NON_DISCRIMINATIVE
BEST_M2_PROFILE=NONE
POSITIVE_VALIDATION_MEDIAN=0.800000
POSITIVE_VALIDATION_P05=0.600000
HELDOUT_VALIDATION_MEDIAN=0.740000
RANDOM_NULL_P95=0.680000
RANDOM_NULL_MAX=0.860000
DISCRIMINATION_GAP=0.120000
RULE_PRECISION=0.800000
RULE_RECALL=0.750000
MEAN_RULE_SUPPORT=3.400000
SINGLETON_FRACTION=0.000000
SUPPORTED_SYNTHETIC_FAMILIES=P0_PARTIAL
REAL_DICTIONARY_SEARCH_AUTHORIZED=NO
```

M2 is a meaningful architectural improvement over unrestricted M1, but it does not yet provide an accepted discriminative basis for real astronomical brute force. No real labels were used for architecture selection.
