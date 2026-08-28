# Task86C-v2 preparation design

This task implements the frozen Task85b execution contract without executing the blind experiment. It accessed only a generated hash-chain fixture whose M0–M5 names are routing labels. No IVTFF/Voynich input, confirmatory mapping, escrow or future truth was opened.

The scientific manifest is canonical. The adapter validates its schema and DAG and computes each JobID from canonical UTF-8 NFC JSON scientific fields. The coordinator may choose only which compatible worker receives a ready node and when. Workers validate bundles, execute a named frozen handler and return facts. Decision logic is confined to the evidence-only verifier.

Publication is temp file → scientific schema/hash/closure validation → fsync → content-addressed rename → atomic JobID index. Equal duplicate hashes retain all telemetry copies. Different hashes set `CONFLICT`, quarantine the index and prevent descendants from becoming ready. Telemetry is a separate object and cannot change the scientific hash.

Restart constructs a new coordinator around the same CAS/index. Verified nodes are absent from the ready set; unfinished/expired leases return with unchanged JobID. Scientific outcomes are valid evidence and are not infrastructure retries. `NOT_REACHED` requires reason, upstream job and dependency hash; absence never substitutes for it.

Acceptance combines unit mutation/failure tests, loopback mTLS, a 193-job 1/2/4/8 worker benchmark, an independent run on `cognition`, and a real cross-node mTLS lease run. The blind Task86C-v2 control was not run.
