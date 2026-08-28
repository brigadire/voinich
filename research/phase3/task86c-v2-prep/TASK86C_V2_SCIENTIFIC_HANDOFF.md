# Task86C-v2 scientific handoff

Task86C-v2 receives the unchanged Task85b contract plus:

- executable SHA-256 `6b015b2e4078b9b5f109ebf3aa8d73918888e431bde267e0d10c3013b524f718`;
- protocol `g1v2-execution-v1`, evidence schema `g1v2-evidence-v1`, canonicalization `canonical-json-nfc-v1`;
- source/component and authoritative registry hashes in `EXECUTION_COMPONENT_MANIFEST.json`;
- manifest/DAG adapter, Phase-I mTLS transport, CAS publisher and evidence-only verifier in `internal/g1v2`;
- executor configuration: pull-by-lease, compatible workers only, immutable bundle, exact-code/config closure, atomic verified index, conflict quarantine;
- the deployment and worker-expansion contracts in this directory.

Before the blind start, place the frozen scientific manifest and all immutable inputs in the coordinator-visible/preseeded CAS, provision at least 120 GB evidence capacity, issue unique worker credentials offline, verify every worker handshake and archive the exact executable. Additional compatible workers require only inventory and credential changes.

No scheduler, protocol, JobID, retry, evidence, duplicate, reachability or aggregation redesign is delegated to Task86C-v2. No scientific verdict is produced by this preparation.
