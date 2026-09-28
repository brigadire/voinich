# Post-qualification identifiability/objective audit

This audit pilot reads the immutable v2 public/truth record and independently enumerates every exact-scope table of size 4 for 12 cases selected by metadata only: two cases per `group × mapping_mode`. It does not modify v2, rerun CP-SAT, or access real STAR LABEL data.

Classification counts: `{'MULTIPLE_OPTIMA': 5, 'GROUND_TRUTH_NOT_OPTIMAL': 2, 'UNIQUE_CORRECT_OPTIMUM': 1, 'HARD_NEGATIVE_CONTROL': 4}`. Hard negatives are reported as `HARD_NEGATIVE_CONTROL` and are excluded from the positive identifiability/objective categories. For positive cases, `GROUND_TRUTH_NOT_OPTIMAL` means the ground-truth table scores below the global optimum. `MULTIPLE_OPTIMA` means the ground truth is optimal but more than one table attains the same optimum. `UNIQUE_CORRECT_OPTIMUM` means the ground truth is the sole optimum under the documented table representation.

The v2 dataset has table_size=4 throughout and does not vary numeric unmatched fraction or a separate collision parameter; those requested strata are therefore unavailable rather than inferred. Noise, hard-negative group, and mapping mode are reported.

```text
AUDIT_MODE=PILOT_METADATA_STRATIFIED
V2_IMMUTABLE=YES
GROUND_TRUTH_NOT_OPTIMAL=2
MULTIPLE_OPTIMA=5
UNIQUE_WRONG_OPTIMUM=0
REAL_DATA_SEARCH_AUTHORIZED=NO
```

The audit distinguishes identifiability failure from objective misalignment but does not redesign the objective or qualify a new solver.
