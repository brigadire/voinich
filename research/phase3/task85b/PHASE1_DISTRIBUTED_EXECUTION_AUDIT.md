# Phase-I distributed execution audit

## Located mechanism

The reusable mechanism is the Task31–42 executor centered on `internal/conditionalregime/remote.go`, `internal/pki`, the executor adapters in `internal/conditionalregime/*_executor.go`, `pipeline-orchestrate`, and `ansible/roles/voynich_worker`. Its design and operations are documented in `docs/methods/DISTRIBUTED_EXECUTION_IMPLEMENTATION.md`, `DISTRIBUTED_EXECUTION_OPERATIONS.md`, and `REMOTE_WORKER_LIFECYCLE.md`.

It evolved from deterministic goroutine jobs to persistent subprocesses and then an mTLS coordinator with pull-by-lease workers. Coordinator and workers authenticate with the project CA; worker identity comes from the certificate. Workers handshake for an experiment, fetch inputs through a SHA-256 content-addressed cache, dynamically lease ready jobs, and submit results bound to worker and lease. Expired leases are reassigned. Persistent workers reconnect across coordinator restarts and rebuild state for a changed experiment. Checkpoints are keyed by deterministic scientific JobID; worker count/backend are operational and excluded from the scientific fingerprint. Reduction is in deterministic ID order.

## Reuse decision

| Capability | Existing | G1-v2 action |
|---|---|---|
| 1 orchestrator + N workers | yes | reuse |
| dynamic claiming and lease expiry | yes | reuse |
| mTLS identity and revocation | yes | reuse |
| content-addressed input cache | yes | reuse |
| backend/worker-count-independent resume | yes | reuse |
| duplicate suppression | yes | extend to retain identical provenance copies |
| deterministic reduction | yes | reuse |
| arbitrary scientific multi-stage DAG | partial; pipeline stages/adapters exist | add G1-v2 manifest adapter |
| immutable multi-artifact result graph | partial | add evidence bundle/index verifier |
| conflicting duplicate quarantine | test patterns exist, not G1 evidence semantics | add G1-v2 quarantine gate |
| scientific-vs-infrastructure taxonomy | partial | bind frozen registries |

No parallel scheduler is justified. The lease queue, PKI, cache, checkpoint, lifecycle, and deployment are reused. The required work is an adapter that maps the G1-v2 DAG/job schema to the executor and a strict result/evidence verifier. Therefore `PHASE1_DISTRIBUTED_REUSE=PARTIAL`.

Cross-node validation is mandatory on representatives M0/M1, M2, M3, M4, and M5. The repository demonstrates prior byte-identical remote work, but Task85b does not claim G1-v2 cross-node identity before the adapter exists. That validation belongs to Task86C-v2 Stage A.
