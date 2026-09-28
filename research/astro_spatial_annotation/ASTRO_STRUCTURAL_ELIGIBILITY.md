# Astronomical structural eligibility

This audit uses spatial structure only; no lexical matching or real M3 brute force was run.

| Audit | Result | Reason |
|---|---|---|
| Cyclic-order eligibility | `NOT_ELIGIBLE` | One panel has a provisional order of 29 stars, but labels are anonymous/low-confidence and no independent order agreement exists. |
| Sequence eligibility | `NOT_ELIGIBLE` | No independently observed and adjudicated label sequence exists. |
| Group eligibility | `EXPLORATORY_ONLY` | Frozen radius components exist only as `DERIVED_CLUSTER`; stability cannot be assessed without the missing panels/pass B. |
| M3 eligibility | `NOT_ELIGIBLE` | No independently observed repeated label↔object pairing/combinatorial structure was established. |

```text
REAL_M3_BRUTEFORCE_RUN=NO
M3_SPATIAL_RECHECK=NOT_ELIGIBLE
PRIMARY_BLOCKERS=INCOMPLETE_OBJECT_COVERAGE,LOW_CONFIDENCE_LABEL_BOXES,NO_DOUBLE_ANNOTATION,NO_SPATIAL_STOLFI_CROSSWALK
```
