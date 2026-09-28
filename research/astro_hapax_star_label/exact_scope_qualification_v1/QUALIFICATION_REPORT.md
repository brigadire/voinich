# Qualification report

The qualification stopped at the pre-sealed verifier gate. The declared scope is explicit and neighborhood-only modes are excluded, but this checkout has no CP-SAT dependency and no frozen exact verifier. The available M1 beam search cannot provide a global certificate, so running it would violate the protocol.

The synthetic manifests and ground truth were created without accessing real STAR LABEL data. The 100 oracle cases are recorded, while CP-SAT agreement, solver invariance, qualification metrics, and null feasibility remain unrun.

```text
EXACT_SCOPE_QUALIFICATION=INVALID
EXACT_SCOPE_SCIENTIFIC_RESULT=NOT_EVALUATED
QUALIFICATION_INVALID_REASON=MISSING_FROZEN_EXACT_VERIFIER
PRIOR_REMEDIATION_RESULTS_TRUST_STATUS=UNVERIFIED
NEW_PRODUCTION_PROTOCOL_AUTHORIZED=NO
REAL_DATA_SEARCH_AUTHORIZED=NO
FULL_TRANSFORMATION_HYPOTHESIS_QUALIFIED=NO
EXACT_SCOPE_DEFINED=YES
ALL_INCLUDED_MODES_HAVE_EXACT_VERIFIER=NO
NEIGHBORHOOD_ONLY_MODES_EXCLUDED=YES
SMALL_INSTANCE_OPTIMUM_AGREEMENT=NA
```

Reason: the requested fixed exact verifier is absent; substituting bounded beam search would invalidate sealed qualification.
