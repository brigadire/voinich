# Remote Distributed Execution Operations

Task33 adds a deliberately small trusted-machine HTTP coordinator/worker
mode to `conditional-regime-analyze`. It is not a cluster manager: operators
start workers, protect the network, and give the coordinator a fixed endpoint
list. The existing local `goroutine` and `process` modes remain defaults and
need no remote service.

## Trust and compatibility requirements

The protocol is intended only for loopback, a private/VPN network, or an
SSH tunnel. It does not provide TLS, authorization policy, tenant isolation,
or sandboxing. Always set the same strong bearer token on coordinator and
workers, restrict the listening socket with a firewall, and never expose a
worker directly to an untrusted network. A worker executes CPU- and
memory-intensive scientific jobs selected by the caller and stores uploaded
input, so treating an unauthenticated public listener as safe is incorrect.

For initial deployment the worker enforces the same protocol version,
scientific compatibility ID, GOOS, GOARCH, and exact `runtime.Version()` as
the coordinator. This conservative restriction bounds the validated
compatibility envelope; do not bypass it. Task33 has passed the
frozen SHA256 oracle on Intel i7-8850H and AMD Ryzen 7 5700X workers. Any new
OS, architecture, Go runtime, or CPU family should be checked the same way
before production use.

## Build and start two workers

Build the same source revision with the same Go toolchain on every machine:

```bash
go build -buildvcs=false -o conditional-regime-analyze ./conditional-regime-analyze
export CONDITIONAL_REGIME_REMOTE_TOKEN='replace-with-a-long-random-secret'
```

On worker 1:

```bash
./conditional-regime-analyze \
  -remote-worker-listen 10.20.0.11:8091 \
  -remote-cache-dir /var/lib/conditional-regime-worker/cache \
  -remote-concurrency 4
```

On worker 2:

```bash
./conditional-regime-analyze \
  -remote-worker-listen 10.20.0.12:8091 \
  -remote-cache-dir /var/lib/conditional-regime-worker/cache \
  -remote-concurrency 4
```

SIGINT initiates graceful HTTP shutdown and allows in-flight requests up to
ten seconds to finish. A worker keeps only immutable SHA256-named input files
and in-memory caches of pure derived state; restarting it cannot change a job.

## Start or resume the coordinator

The normal scientific flags are unchanged. `-workers` is the coordinator's
global in-flight-job bound, not a per-host value:

```bash
export CONDITIONAL_REGIME_REMOTE_TOKEN='replace-with-a-long-random-secret'
./conditional-regime-analyze \
  -executor remote \
  -remote-workers http://10.20.0.11:8091,http://10.20.0.12:8091 \
  -workers 8 -remote-timeout 20m -remote-retries 3 \
  -corpus data_work/ZL3b-x7.txt \
  -token-metadata-map workdir/metadata-validation/token_metadata_map.tsv \
  -output-dir workdir/conditional-regimes \
  -checkpoint-path workdir/conditional-regimes/checkpoint.json \
  -permutations 1000 -seed 1
```

Run the identical command after interruption to resume. The coordinator
loads only a checkpoint with the exact experiment fingerprint, skips every
completed `JobID`, retries missing jobs, and deletes the checkpoint only
after final outputs are written. Worker order/count may change between
runs. A late response from another experiment is rejected, and duplicate
delivery cannot contribute twice.

Progress uses the existing stderr status display. There is no web UI. The
worker's `GET /v1/info` endpoint is a small authenticated operational probe;
for example:

```bash
curl -H "Authorization: Bearer $CONDITIONAL_REGIME_REMOTE_TOKEN" \
  http://10.20.0.11:8091/v1/info
```

Machine-readable counters and resource readings are available without a UI:

```bash
curl -H "Authorization: Bearer $CONDITIONAL_REGIME_REMOTE_TOKEN" \
  http://10.20.0.11:8091/v1/metrics
```

The counters include cold input and staging time, cache hits/misses, job and
result payload bytes, failures, process CPU ticks, current/peak RSS, and Go
heap. On Linux divide CPU ticks by `getconf CLK_TCK` (100 on both measured
Task33 hosts). Counter deltas before/after a run isolate that run.

## Input staging and cache verification

At coordinator startup, both input files are SHA256-hashed by the existing
loader. For each worker the coordinator performs `HEAD /v1/input/<sha256>`
and uploads a missing object once with a bounded `PUT`. The worker hashes the
bytes before atomically installing the file as `<cache>/<sha256>`. Each job
names both hashes and the experiment fingerprint; the worker reloads and
recomputes the fingerprint before creating cached experiment state. Thus a
stale path or mutated cache object cannot execute as the requested
experiment. Local filesystem path text never enters scientific computation.

Cold staging traffic for the frozen corpus is measured as 2,440,409 bytes
per worker (234,466-byte corpus + 2,205,943-byte metadata map), plus small
HTTP headers. A warm run uploads zero input bytes. Each job request
and result is bounded to 1 MiB; input objects are currently bounded to 64 MiB.
Use `/v1/metrics` deltas for exact application-payload accounting; HTTP/TCP
header bytes remain transport overhead and require interface-level tooling.

## Failure behavior

- HTTP 5xx, 429, disconnects, and timeouts are transport failures and are
  retried against the configured endpoints. Diagnostics include endpoint,
  worker hostname when available, and exact `JobID`.
- Protocol, runtime, input, experiment, malformed-job, and scientific errors
  are explicit non-retryable failures.
- Removing a worker merely causes retry on another endpoint. Re-add it at the
  same endpoint; its immutable disk cache survives restart.
- If all attempts fail, the coordinator exits with its atomic checkpoint
  intact. Re-run the command after workers recover.

## Future multi-corpus scheduling

The present fingerprint is the `ExperimentID`; input hashes provide the
future `CorpusID`; and `JobID` remains `(stage, combination, replicate)`.
A future scheduler can queue `(ExperimentID, CorpusID, JobID)`, use
round-robin/deficit scheduling across experiments so small corpora are not
starved, and retain a separate canonical replicate-index reducer and
checkpoint per experiment. No network arrival order should ever cross an
experiment's reduction boundary.
