# Structural projection distributed execution audit (Task40)

Date: 2026-08-17. The scientific oracle is the implementation at the start
of Task40. Production parameters remain unchanged; profiling alone used one
random projection instead of 200.

## Before profile

Fresh runs were made against `data_work/ZL3b-x7.txt` and
`data_test/pg2097-2.txt`, with the production inputs and parameters and
`-random-projections 1`. The profiles are external diagnostic artifacts, not
experiment outputs.

| corpus | wall | CPU samples | trial loop CPU | non-trial/final controls CPU | allocations | live heap at exit |
|---|---:|---:|---:|---:|---:|---:|
| Voynich | 244.107 s | 264.39 s | about 72.5 s | family controls 164.58 s; preparation and output below 2 s | 27.856 GiB | 4.52 MiB |
| Doyle | 147.461 s | 157.87 s | about 35 s | family controls 109.57 s; preparation and output below 2 s | 13.426 GiB | 3.02 MiB |

The heap profile records live heap at exit, not maximum resident set size;
peak RSS was not captured by the profiling invocation and is therefore not
misreported as peak memory. The main CPU costs were `metricsFloat`, string
sorting/map operations, `GenericSmoothing`, `normalize`, and
`ProjectDistribution`. Doyle allocation-space leaders were `countsFloat`
(6.69 GiB), `normalize` (4.18 GiB), and `ProjectDistribution` (1.22 GiB).

The one-trial runs contain a large serial family-analysis floor, but at the
production value of 200 the repeated trials are projected to exceed 98% of
stage CPU. Distribution therefore passes the stop condition. This is an
extrapolation from the fresh profile, not a substitute for the scaling runs
reported below.

## Parallelism and RNG

A trial consumes only the four immutable input files, the staged corpus and
profiles, scientific parameters, base seed, and trial index. Existing seed
semantics are retained exactly:

```text
full random       seed + trial*7919
full smoothing    seed + trial*104729
future random     seed + trial*15485863
future smoothing  seed + trial*32452843
```

There is no global RNG and no dependence on worker identity, the previous
trial, completion order, retry count, or worker count. `TrialWorker` stages
the corpus, profiles, projections, frequency bins, and selected-pair order
once per process.

The package does reuse package-level normalization and metric scratch
buffers. They are not safe for concurrent trial goroutines. Consequently a
`TrialWorker` serializes `Run`, and a remote worker deliberately reduces its
structural concurrency to one. Parallelism is across isolated worker
processes. No scientific-core rewrite or algorithm change was made.

## Architecture and deterministic reduction

Task40 extends the Task31--35 process and mTLS coordinator/worker protocol
with workload `structural_projection` and job stage
`structural_projection_trial`; it does not create another service or PKI.
The coordinator publishes four SHA256-addressed invariant objects (corpus,
structural pairs, distance pairs, families), which each worker verifies and
caches once.

The JobID is `(stage, experiment fingerprint, trial index)`. The fingerprint
hashes the contents of all four inputs and every relevant scientific
parameter. Hostname and IP are excluded. Remote JSON float encoding uses
Go's round-trip representation. Results are stored by trial index, validated,
then reduced strictly in order `0..N-1`; network arrival order is never a
float-reduction order.

The existing lease expiry, reassignment, retry, late-result rejection, mTLS
identity, revocation, and JobID deduplication paths are shared by both job
types. A crash-safe JSON checkpoint is atomically written after every newly
accepted trial. Its fingerprint must match, completed trials are skipped on
restart, and the checkpoint is removed only after all scientific outputs are
written successfully.

## Progress and pipeline integration

Non-quiet runs report completed/total trials, active workers, outstanding
work, retries, elapsed time, and ETA. `-quiet` suppresses this operational
output. `pipeline-orchestrate` emits a one-minute stage heartbeat even when
the child log is otherwise quiet.

Stage 17 and stage 21 are both marked executor-capable. A new manifest records
the executor arguments for each; old frozen manifests continue to use their
recorded command lines. Structural projection receives an operational
checkpoint path which is deleted on success and therefore is not frozen.

## Correctness evidence

- `go test -race ./internal/structuralprojection ./internal/conditionalregime ./pipeline-orchestrate` passes when loopback sockets are permitted.
- `go test ./...` passes.
- Automated tests cover out-of-order canonical collection, checkpoint skip
  and resume, missing trials, process protocol identity checks, mTLS local vs
  remote trial equality, two-worker remote execution, lease expiry and
  reassignment, late/duplicate results, stale experiments, revocation, and
  remote checkpoint resume. These transport fault paths are job-type agnostic.
- A real Voynich one-trial old-local oracle and the new process executor were
  compared recursively: every persisted byte was identical. The old-local
  wall time was 244.107 s; the one-process run was 211 s.
- `pipeline-orchestrate verify -experiment-dir experiments/voynich-v1`
  verifies all 340 frozen files. The frozen baseline was not modified.

## Doyle scaling study

The required real 1/2/5/10-host measurement could not be completed from this
coordinator. Ten existing Task35 workers were deployed with the extended
binary, but every node failed to reach `https://109.87.30.62:8443`; external
TCP/8443 is closed. The attempted coordinator was stopped, and Ansible
cleanup was verified on all ten nodes: no process, binary, cache, input,
certificate, private key, or PID file remains. No new CA or certificate was
created.

Accordingly, no invented wall/RSS/retry values are presented as measurements:

| remote workers | projection time | total time | speedup | efficiency |
|---:|---:|---:|---:|---:|
| 1 | blocked: coordinator ingress closed | -- | -- | -- |
| 2 | blocked: coordinator ingress closed | -- | -- | -- |
| 5 | blocked: coordinator ingress closed | -- | -- | -- |
| 10 | blocked: coordinator ingress closed | -- | -- | -- |

For capacity planning only, the fresh Doyle R=1 profile implies roughly 35 s
per trial and about a 112 s serial floor. Ignoring network and contention,
the production estimate is approximately 119, 60, 25, and 14 minutes total
at 1, 2, 5, and 10 workers respectively. These are estimates, not Definition
of Done scaling measurements. Real runs require routable coordinator ingress
on TCP/8443 (and a certificate SAN matching the address workers use), after
which the exact commands in `DISTRIBUTED_EXECUTION_OPERATIONS.md` can be used.

## Conclusion

The implementation and byte-identity evidence satisfy the code, protocol,
resume, determinism, pipeline, and Voynich-regression portions of Task40.
The only incomplete Definition of Done item is the real Doyle 1/2/5/10
scaling table, blocked by external network reachability rather than code.
