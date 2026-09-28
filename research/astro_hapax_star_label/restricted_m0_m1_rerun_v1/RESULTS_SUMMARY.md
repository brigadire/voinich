# Restricted M0-R/M1-R results

The unchanged engines were applied through an external input adapter to the frozen page-defined scope. Original M0/M1 code and artifacts were not modified.

| Direction | Train | Held-out | M0-R | M1-R |
|---|---:|---:|---:|---:|
| f68r1 → f68r2 | 30 | 27 | 0/30; 0/27 | 9/30 (0.300); 0/27 (0.000) |
| f68r2 → f68r1 | 27 | 30 | 0/27; 0/30 | 8/27 (0.296); 0/30 (0.000) |

M0-R therefore finds no exact transformed-string compatibility in the fixed 640-system grid. M1-R finds partial TRAIN compatibility, but no cross-page held-out transfer in either direction. This pattern is consistent with a global substitution-table overfit, not with a CROSS_PAGE_SIGNAL.

The requested expanded null controls were not completed in this run. The unchanged M1 optimizer is computationally expensive at 10,000 full-search replicates per family and direction; all null outputs must remain `NOT_RUN` until that budget is completed. No p-value or null advantage is claimed here. Hapax/non-hapax breakdown is available from the frozen target registry but was not used to select a model.

```text
M0_R_STATUS=COMPLETE_OBSERVED_ONLY
M1_R_STATUS=COMPLETE_OBSERVED_ONLY
M0_R_F68R1_TO_F68R2=0/30 -> 0/27
M0_R_F68R2_TO_F68R1=0/27 -> 0/30
M1_R_F68R1_TO_F68R2=9/30 -> 0/27
M1_R_F68R2_TO_F68R1=8/27 -> 0/30
NULL_REPLICATES_COMPLETED=0
CROSS_PAGE_SIGNAL=NOT_SUPPORTED
M2R_REQUIRED=NOT_YET (await expanded null completion)
```
