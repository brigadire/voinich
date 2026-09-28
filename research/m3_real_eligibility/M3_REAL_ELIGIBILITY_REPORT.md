# M3 real-data eligibility

This is a structure-only audit of frozen D1 STAR source forms and confirmed STAR label types. No semantic assignment, M1/M2 mapping, spelling-driven split, or brute-force search was used.

The source side has 63 STAR rows and repeated lexical edge chunks, but the target side has 54 confirmed STAR token types with one confirmed occurrence each. Because the frozen label cross-section does not expose source-concept pairing or a repeated target component graph tied to concepts, a valid novel-combination split cannot be formed. A split with at least five held-out novel combinations would require inventing pairings or weakening M3P1_V constraints.

```text
M3_REAL_DATA_ELIGIBILITY=NOT_ELIGIBLE
ELIGIBLE_GENERATIVE_FAMILIES=NONE
STAR_LABEL_TYPES=54
SOURCE_CONCEPTS=63
SOURCE_RECURRENT_COMPONENTS_GE3=10
TARGET_RECURRENT_COMPONENTS_GE3=11
MAX_FEASIBLE_TRAIN=0
MAX_FEASIBLE_NOVEL_HELDOUT=0
ELIGIBLE_SPLITS=0
MATCHED_NULL_FEASIBLE=NO
REAL_M3_BRUTEFORCE_AUTHORIZED=NO
```

The result is `NOT_ELIGIBLE`, not a license to adapt segmentation or support thresholds. A future eligibility run would require a frozen source-label occurrence table with independently observed repeated component combinations.
