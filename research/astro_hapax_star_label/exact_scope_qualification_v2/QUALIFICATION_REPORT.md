# Exact-scope qualification v2 result

The content-addressed solver snapshot and preregistered gates were frozen before the replacement sealed dataset. The earlier attempts are excluded under `invalid_attempt_1/`, `invalid_attempt_2/`, and `invalid_attempt_3/`.

The final sealed run contains 30 zero-noise, 30 noisy, and 30 hard-negative cases. Predictions were generated from public data only; truth was opened only by this scoring step.

```text
EXACT_SCOPE_QUALIFICATION=EXACT_SCOPE_NOT_QUALIFIED
EXACT_SOLVER_SNAPSHOT=CONTENT_BOUND
GIT_COMMIT_BINDING=UNAVAILABLE_NOT_REQUIRED
CORRECTNESS_PRECHECK=PASS
SEALED_QUALIFICATION_RUN=COMPLETE
PREREGISTRATION_REQUIRED=SATISFIED
REAL_DATA_SEARCH_AUTHORIZED=NO
```

Metrics: zero-noise recovery 0.233; noisy recovery 0.200; precision 0.525; recall 0.525; held-out coverage 0.667; false certificates 0; hard-negative false accepts 0.

The exact verifier passes the independent correctness precheck, but the synthetic recovery gates fail. The result does not authorize real-data search and does not qualify the full transformation hypothesis.
