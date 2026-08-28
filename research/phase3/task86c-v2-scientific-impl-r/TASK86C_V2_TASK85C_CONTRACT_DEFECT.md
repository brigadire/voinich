# Task86C-v2-scientific-impl-R contract defect

## Classification

`CONTRACT_DEFECT` at implementation point `A17 / evidence-schema compiler`, before production scientific handlers or inputs were constructed.

## Authoritative identity

The supplied Task85c identity matches: executable contract SHA-256 `275b29c592be6d3cb80c20df9b9348348009d758e2770ed39da0066004b11bca`; Task85c artifact root `273913473e3e37d6a776c79b0eb214753a90e9dbaf5d78e186dcb65a0c32c351`; evidence schema root `8462ceb7f34efce1674528af7e69bdbf6855cd4938494d6d5034247245235ed0`; golden root `b7443a962a82dd5c0cd67b71e24d8acea73fc9be4863fca4078bc53e468c7e51`. The defect is therefore in the frozen closure, not a hash mismatch.

## Exact contradiction

`G1V2_STATUS_REGISTRY.tsv` permits `PASS` and `FAIL` only from a gate/verifier and permits `NOT_REACHED` only from the DAG materializer. `G1V2_EXECUTABLE_CONTRACT.md` likewise says suppressed cells emit `NOT_REACHED`, and the task requires schemas to reject forbidden status combinations.

However, every frozen evidence schema uses the same status enum. In particular:

- `schemas/not_reached.schema.json` admits `status=PASS` and explicitly requires the ordinary NOT_REACHED payload for PASS/FAIL/NOT_ASSESSABLE.
- `schemas/fit.schema.json` admits `status=FAIL` and explicitly requires only `candidate_id` and `control_instance_id` for it.
- The prose restriction in `x-status-rules` is an annotation: JSON Schema does not enforce unknown `x-*` keywords.

Thus the machine-readable schemas, which have precedence over explanatory prose, accept evidence the status registry declares illegal.

## Smallest reproducing examples

`contract-defect/not_reached_with_pass.json` has a correct G1V2-CJ-1 content hash and satisfies the frozen `not_reached` schema while using the scientifically impossible status PASS. `contract-defect/fit_with_fail.json` likewise satisfies the frozen fit schema while representing a fit as scientific negative evidence. Run:

```text
python3 research/phase3/task86c-v2-scientific-impl-r/reproduce_contract_defect.py
```

## Scientifically distinct permitted implementations

1. A schema-first implementation accepts these records because machine-readable artifacts have frozen precedence. A fit FAIL can then participate as negative evidence, and a PASS not-reached record can participate in aggregation.
2. A registry/prose-first implementation rejects them and preserves failure/missingness as non-negative evidence.

These paths can produce different PredictiveAdequacy, NONE, NOT_IDENTIFIABLE, reachability, and final verdicts from the same evidence bytes. Selecting either is an unauthorized scientific decision.

## Required resolution

Task85c must publish a new immutable revision whose per-type schemas enforce legal producer/status combinations with standard JSON Schema constraints, define status-specific payloads, and update all dependent hashes and golden valid/invalid vectors. Task86C-v2-scientific-impl-R must then restart against that recorded identity. This task did not patch Task85c.
