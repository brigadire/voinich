# Exact-scope qualification v2

This is a new synthetic sealed run. It uses only the copied solver snapshot in `snapshot/`, fresh seeds derived from `20261001`, and per-row solver/config/dependency hashes. The legacy remediation outputs are not inputs.

The precheck ran 100 instances: CP-SAT/oracle objective agreement is `1.000` and CP-SAT optimal status is `1.000`. These are solver-correctness and provenance checks, not evidence about real STAR LABEL data.

The snapshot is content-bound and hash-bound, but not yet commit-bound because this checkout's `.git` is read-only and refused creation of `index.lock`. The full qualification remains pending commit binding and preregistered qualification gates.

```text
CP_SAT_IMPLEMENTATION_EXISTENCE=VERIFIED
CP_SAT_CORRECTNESS_PRECHECK=PASS
HISTORICAL_RAW_RESULT_ATTRIBUTION=UNVERIFIABLE
PRIOR_REMEDIATION_RESULTS_USE=DIAGNOSTIC_ONLY
NEW_EXACT_SCOPE_QUALIFICATION_REQUIRED=YES
EXACT_SCOPE_SCIENTIFIC_RESULT=NOT_EVALUATED
REAL_DATA_SEARCH_AUTHORIZED=NO
```

No real data was read.
