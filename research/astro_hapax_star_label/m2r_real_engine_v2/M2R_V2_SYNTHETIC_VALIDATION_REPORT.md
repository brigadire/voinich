# M2R-v2 synthetic validation

BLOCKED. The synthetic gates do not justify an engine freeze or independent pre-production readiness.

Failed gates: precision, recall, assignment, heldout_coverage, null_separation, order_invariance, checkpoint, complete_searches.

Metrics use full latent rule tuples and exact surface assignments, macro-averaged over the six hidden positive datasets. Negative and out-of-family datasets are reported separately. Every dataset uses the same isolated worker and fixed configuration. Only one replicate per listed condition was run; tail quantiles from eight nulls are descriptive and do not establish a population false-positive guarantee.

The shuffled-assignment null preserves the bags of surfaces and cannot be distinguished by an assignment-search engine. Bigram-preserving nulls may also preserve some entire tokens. No truth-aware score adjustment is used. The bounded candidate pool and rule search are heuristic; full global optimality is not claimed. All-pairs top-k derivations are retained for the selected model; search pool pruning can exclude true morphology.

A zero hard-negative acceptance count cannot compensate for poor positive recovery. The original v1 is checksum-verified unchanged. Real inputs were neither loaded nor searched; only authorized manifests and SHA-256 metadata were inspected.

Hidden validation has been disclosed. Do not repair or tune this v2 implementation. Further development requires M2R-v3 and fresh hidden seeds.

See [POST_VALIDATION_FINDINGS.md](POST_VALIDATION_FINDINGS.md) for the structural recall ceiling, inherited peak-memory accounting failure, and the correction that null false-positive rate is **NA**, not the raw incomplete-run eligibility fraction. These are reporting findings only; sealed code and configuration were not changed.

```text
M2R_V2_STATUS=BLOCKED
M2R_V1_MODIFIED=NO
REAL_DATA_CONTENTS_ACCESSED=NO
ASSIGNMENT_SEARCH=IMPLEMENTED
MULTISYMBOL_MORPHOLOGY=IMPLEMENTED
MIN_INDEPENDENT_SUPPORT=3
DUPLICATE_SUPPORT_GUARD=PASS
RULE_CONFLICT_HANDLING=PASS
SIZE_NORMALIZED_SCORING=PASS
SEARCH_TYPE=BOUNDED_DETERMINISTIC_OPTIMIZATION
ORDER_INVARIANCE=FAIL
CHECKPOINT_RESUME_IDENTICAL=NO
LATENT_RULE_PRECISION=0.2972222222222222
LATENT_RULE_RECALL=0.1
ASSIGNMENT_ACCURACY=0.022058823529411766
HELDOUT_COVERAGE=0.0
HELDOUT_COMPRESSION_GAIN=0.11531287217634459
SYNTHETIC_NULL_P95=NA
HARD_NEGATIVES_ACCEPTED=0/10
SINGLETON_FRACTION=0.0
CR01_CR09_REGRESSION=PASS
M2R_V2_SYNTHETIC_VALIDATION=FAIL
M2R_V2_ENGINE_FROZEN=NO
READY_FOR_INDEPENDENT_PREPRODUCTION_AUDIT=NO
REAL_DATA_SEARCH_AUTHORIZED=NO
RESULTS_REPRODUCIBLE=NO
```
