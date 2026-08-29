# Task85c-d report

## Outcome

Parent V1.1 identity and the R2 failure closure are supported. R2-G01 and R2-G02 are reproduced and semantically resolved by precedence:

- the JobID scientific identity binds the stable scientific baseline `G1_V2_EXECUTABLE_CONTRACT_V1_1`, not the patch artifact label;
- the sole dependency field is `dependency_job_ids`, containing direct parent JobIDs in deterministic UTF-8 byte order and present as `[]` for FIT.

No patch was issued because the mandatory complete-DAG proof needs 144 exact blind-safe IDs which the frozen closure does not contain. Parent rules derive them from HMAC-SHA256 over secret truth with a random escrow key; Task85c-d forbids constructing or inspecting those controls and forbids generating escrow. Placeholders would prove a different JobID set, while descriptive generator/scale IDs would disclose blind truth.

## Verdicts

`PARENT_V1_1_IDENTITY = SUPPORTED`  
`R2_FAILURE_REPRODUCED = SUPPORTED`  
`R2_G01 = CLOSED`  
`R2_G02 = CLOSED`  
`JOBID_NORMATIVE_SOURCE_CLOSED = SUPPORTED`  
`JOBID_PAYLOAD_UNIQUE = SUPPORTED` at the payload-schema level  
`COMPLETE_DAG_JOBID_BIJECTION = NOT_SUPPORTED`  
`COMPLETE_DAG_EDGE_CLOSURE = NOT_SUPPORTED`  
`EVIDENCE_SCHEMA_ROOT_CHANGED = NO`  
`EVIDENCE_SCHEMA_ROOT_SHA256 = 4744ca82532cd47a0d02bb680796b26a11ceca57d6229f0b312df69a103f784b`  
`CONTRACT_PATCH_REFREEZE = NOT_SUPPORTED`  
`TASK86C_V2_SCIENTIFIC_IMPL_R3_READY = NOT_SUPPORTED`.

`FAILURE_CLASS = DEFECT_SCOPE_EXPANDED`  
`BLOCKING_FINDINGS = D01-BLIND-ID-CLOSURE`  
`PRODUCTION_RUN_AUTHORIZED = NO`  
`TERMINAL_MARKER = TASK85C_D_CONTRACT_DEFECT_EXPANDED`.
