# Validation report

Validated by `main.py`: all included synthetic ground-truth rows are capacity-feasible, contain no composition or abbreviation rule, and use disjoint declared seeds for DEV, qualification, and hard-negative sets. The qualification set is isolated from real data.

The implementation gate fails because no exact CP-SAT verifier is available. Therefore no solver result is interpreted as a qualification result. Prior remediation outputs remain `UNVERIFIED` because their executable linkage has not been established.

```text
EXACT_SCOPE_SCIENTIFIC_RESULT=NOT_EVALUATED
QUALIFICATION_INVALID_REASON=MISSING_FROZEN_EXACT_VERIFIER
PRIOR_REMEDIATION_RESULTS_TRUST_STATUS=UNVERIFIED
REAL_DATA_SEARCH_AUTHORIZED=NO
```

`QUALIFICATION_INVALID` is the only permitted status for this incomplete run.
