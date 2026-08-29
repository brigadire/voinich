# Task86C-v2 scientific implementation R2 — frozen golden contradiction

## Classification

`GOLDEN_CONTRACT_CONTRADICTION` at the mandatory pre-implementation contract/golden closure gate. Production scientific implementation and freeze are prohibited by sections 6 and 136 of the R2 task.

## Identity closure

The supplied files have the required identities:

- executable contract: `5c3cd272c1dbae9bfe1d7a100155faf102e86d34660da239e1cb31704ad470b0`;
- Task85c-c artifact root: `dc27660d3e03a06f91aaa1c3aa6c48226328457746c9dc819e7c0a355ce8f09c`;
- evidence schema root: `4744ca82532cd47a0d02bb680796b26a11ceca57d6229f0b312df69a103f784b`;
- V1.1 golden root: `967e63f9c3ba66d99cbff4da83c819d705f1f14c4ace923dbcd1cb8e7dfd1ff8`;
- V2 status/reachability SHA-256: `fc1ca07d8123ed5d44bc24ecba98fca54d5b05781ecbaba820d44079319038b9`;
- V2 status/reachability root: `51b3b517f50a050f93524c1dbe74701efd244821b48e6d23607b63ddf39c1f0f`.

The Task85c-c results-manifest validator and cross-artifact validator pass. The defect is therefore inside the frozen transitive closure, not a supplied-file identity mismatch.

## R2-G01: contract-version contradiction

The V1.1 machine contract fixes `contract_version` to `G1_V2_EXECUTABLE_CONTRACT_V1_1`. R2 also requires the production manifest and evidence to bind V1.1. The inherited frozen `JOBID` golden vector instead places `G1_V2_EXECUTABLE_CONTRACT_V1` in the scientific identity payload and expects:

`j-d85279815a36c30515b0be66387c99c3303fa09e`.

Changing only that payload value to the authoritative V1.1 identity, while applying the same frozen G1V2-CJ-1 and JobID formula, produces:

`j-f7c26e7460fa192e3186873428d5e2a37caa6285`.

Both cannot be the JobID for the required V1.1 production cell. Retaining V1 satisfies the golden but fails to bind the authoritative contract; using V1.1 satisfies the normative production identity but fails the unmodifiable golden.

## R2-G02: dependency-field contradiction

The V1.1 `G1V2_DAG_CONTRACT.json` gives the exact JobID payload field `dependency_job_ids`. The `JOBID` golden vector uses the distinct field `dependencies`. With V1.1 and the registry field name, the same FIT example produces:

`j-186f1406add6d4d4d7f788907efb76500468a5f7`.

Because canonical object member names are hashed, these representations are not aliases. The declared precedence puts the machine registry before golden vectors, while R2 forbids changing the expected golden output.

## Minimal reproducer

Run:

```text
python3 research/phase3/task86c-v2-scientific-impl-r2/reproduce_golden_contract_contradiction.py
```

The reproducer reads only frozen artifacts, recalculates all three identifiers, verifies the golden value, and exits only after proving they are distinct.

## Required resolution

Publish a new immutable contract revision that fixes one exact JobID payload schema and a matching golden vector, then updates all affected roots and downstream identities. Historical artifacts must remain unchanged. R2 must restart against that new authority; this implementation task cannot choose which identity to preserve.
