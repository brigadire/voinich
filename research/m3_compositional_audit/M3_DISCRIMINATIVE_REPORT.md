# M3 compositional-generalization audit

M3 freezes source/target segmentation, recurrent component rules and role-preserving composition. TRAIN and HELDOUT share components but have disjoint combinations; held-out rematching, exceptions and new rules are forbidden.

Formal metrics: `NOVEL_COMBINATION_ACCURACY = correctly predicted unseen combinations / unseen combinations`; `EXACT_TOKEN_PREDICTION` requires the full generated token without additional search; `COMPONENT_RULE_PRECISION/RECALL` compare inferred reusable rules with hidden rules; `COMPOSITIONAL_GENERALIZATION_GAP = positive novel-combination median - matched-null P95`.

Independent validation passes the preferred thresholds for the basic role-composition profile: positive novel-combination median 0.86, P05 0.70, held-out exact median 0.82, null P95 0.43, gap 0.43. However this is limited to the exact compositional family; G2–G6 are not uniformly supported.

```text
M3_COMPOSITIONAL_MODEL=FAMILY_LIMITED
BEST_M3_PROFILE=M3P1_V
POSITIVE_NOVEL_COMBINATION_MEDIAN=0.860000
POSITIVE_NOVEL_COMBINATION_P05=0.700000
HELDOUT_EXACT_TOKEN_MEDIAN=0.820000
NULL_NOVEL_COMBINATION_P95=0.430000
NULL_NOVEL_COMBINATION_MAX=0.600000
COMPOSITIONAL_GENERALIZATION_GAP=0.430000
COMPONENT_RULE_PRECISION=0.900000
COMPONENT_RULE_RECALL=0.880000
SUPPORTED_GENERATIVE_FAMILIES=G0,G1
REAL_DICTIONARY_SEARCH_AUTHORIZED=NO
```

M3 is promising for strict compositional systems but is not broad enough to authorize real astronomical brute force. A separate real-data eligibility check and broader family validation are required.
