# EI01 — blind-ID algorithm ambiguity

## Classification

`EXECUTION_IDENTITY_DEFECT`. Blind IDs affect run-instance and JobID identity, but not scientific model mathematics or outcomes.

## Conflicting/under-specified machine sources

`G1V2_RNG_DOMAIN_REGISTRY.tsv` defines domain `BLIND_ID`, namespace `g1v2/blind/id`, counters `generator_index,scale_index,replicate`, output `digest`, consumer `escrow builder`.

`G1V2_BLIND_ESCROW_SCHEMA.json` instead states: random 32-byte escrow key; HMAC-SHA256 over a canonical truth record; blind ID is the first 20 hex characters. It does not define:

- the exact truth-record object and required/forbidden fields;
- whether `blind_id`, which appears in `secret_fields`, is excluded as an output or included recursively;
- whether token count, scale, replicate, generator index, or RNG digest enters the HMAC message;
- any HMAC domain separator or exact message framing;
- whether the RNG `BLIND_ID` digest is the ID, HMAC input, key material, or unused.

## Non-equivalent implementations

At least these readings fit parts of the frozen machine closure:

1. use the first 20 hex characters of the G1V2-RNG-1 `BLIND_ID` digest;
2. HMAC the canonical truth fields excluding the output `blind_id`;
3. HMAC a truth record that additionally binds the RNG `BLIND_ID` digest/counters.

They yield different opaque `control_instance_id` values for identical generated content. Since `control_instance_id` enters every JobID payload, the concrete JobID set and DAG roots also differ.

## Required repair

Freeze one machine-readable escrow schema with an exact HMAC input schema, required/forbidden fields, domain framing, output encoding, and explicit relationship to G1V2-RNG-1 `BLIND_ID`. Preserve blind secrecy and scientific control generation. Then restart R2R before generating a production escrow key.
