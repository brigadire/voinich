# Remote Worker Lifecycle (Task42)

## Problem

Task33-40 built a real mTLS distributed executor (coordinator + worker,
lease queue, deterministic reduction, PKI, Ansible deployment), but every
worker's process lifetime was implicitly scoped to **one coordinator
run**: `RunRemoteWorker` did exactly one handshake at startup, built its
computer state once from that handshake, and then leased jobs against that
one `ExperimentID` until the process exited. Nothing rebuilt that state or
re-handshook if the coordinator disappeared and came back as something
else.

In practice this meant:

- After an experiment finished (or the coordinator was stopped/restarted),
  already-deployed workers could not be reused for the next experiment
  without restarting the worker process.
- The Ansible role's readiness check (`GET /v1/handshake` must return 200
  within ~20s or deployment fails) assumed a coordinator was already
  running at deploy time - which directly contradicts "deploy once, run
  many experiments," since most of that lifetime has no coordinator
  running at all. This is the concrete hang referenced in Task42's own
  problem statement (`voynich_worker : readiness - probe the coordinator
  with this worker's own mTLS credentials`): deploying workers *before*
  starting the first experiment made the deployment itself fail/hang.

## Design

### The worker is a generation loop, not a single connection

`internal/conditionalregime/remote.go` now separates two concerns that
used to be one function:

- **`runWorkerGeneration`** (the old `RunRemoteWorker` body, still exported
  as `RunRemoteWorker` for tests and any caller that deliberately wants
  exactly one connection): one handshake, one computer-state build, one
  lease loop, until ctx is cancelled or the coordinator stops recognizing
  this worker's experiment identity.
- **`RunPersistentRemoteWorker`** (the new deployment entry point, used by
  `conditional-regime-analyze -coordinator ...`): calls
  `runWorkerGeneration` in a loop for as long as `ctx` is alive. Every
  generation starts from zero - a brand-new handshake and a freshly built
  computer state - so nothing from generation N can leak into generation
  N+1.

```
start
  |
  v
connect coordinator  -----------------------------+
  |                                                |
  +-- success -> register -> accept jobs           |
  |         |                                      |
  |         +-- coordinator rejects our             |
  |             experiment identity (409) ----------+  (re-handshake:
  |                                                     new experiment,
  +-- disconnect (transport failure) ------------------ new computer state)
              |
              v
          bounded backoff + jitter (1s .. 60s)
              |
              v
          connect again
```

### Detecting "the coordinator moved on" without polling forever

A worker's lease loop identifies itself by the `ExperimentID` it captured
at its last handshake. If the coordinator has since restarted for a
*different* experiment, `POST /v1/lease` now returns **HTTP 409** with a
distinguishable sentinel (`errStaleExperiment`) instead of a generic
error. Previously this fell into the same infinite-backoff-and-retry path
as a transient network failure, so a worker holding a stale experiment
identity would poll the wrong `ExperimentID` forever and never notice a
new experiment existed. Now the lease loop returns immediately on that
sentinel, and the outer generation loop re-handshakes right away.

If the coordinator instead restarted for the **same** experiment (same
corpus/config, same fingerprint), nothing about the identity check
changes: ordinary HTTP/TLS reconnection (already retried with its own
short backoff inside one generation) picks the connection back up without
ever needing a new handshake or rebuilt computer state. Restart-for-the-
same-experiment and restart-for-a-different-experiment are both covered,
by two different, independently minimal mechanisms.

### Reconnect policy: bounded exponential backoff with jitter

`RunPersistentRemoteWorker` backs off between generations starting at 1s,
doubling to a 60s ceiling, with full jitter (`sleepWithJitter` sleeps a
random duration in `[0, backoff)`) - never a tight loop
(`TestPersistentWorkerReconnectUsesBoundedBackoffNotTightLoop` bounds a
real black-box connection-attempt count over several seconds). A
generation that stayed connected long enough to actually do work resets
the backoff to its floor, so one bad start doesn't leave a healthy,
long-running worker sluggish to reconnect after a later, unrelated blip.

This is a *different, outer* backoff from the pre-existing
`remoteLeaseBackoff`/`remoteMaxBackoff` (200ms-5s) that already paced
"no work available yet" polling and transient lease-request retries
*within* one generation - that inner policy is unchanged.

### Lifecycle logging: once per transition, not once per second

`workerStateLogger` prints `worker: <state>[: <detail>]` only when the
state actually changes:

```
coordinator unavailable -> reconnecting -> connected -> authenticated -> registered -> disconnected -> reconnecting -> ...
```

A worker idling with no pending work, or backed off waiting for a
coordinator, never repeats the same line every second - satisfying Task42
section 10's "do not spam the same message" requirement while still
making the current state visible (`grep '^worker: '
.../worker.log | tail -n 1`, which is exactly what the new Ansible
`status` operation does - see below).

### Permanent vs. transient failures

`isPermanentWorkerError` classifies an mTLS/certificate error
(`x509.UnknownAuthorityError`, `x509.HostnameError`,
`x509.CertificateInvalidError`, `*tls.CertificateVerificationError`) as
**permanent**: `RunPersistentRemoteWorker` logs it clearly and returns
instead of retrying. Everything else (connection refused/reset, timeout,
coordinator not listening yet, EOF) is **transient** and retried with
backoff. An untrusted CA or a revoked/foreign certificate will never
resolve itself by waiting, unlike a coordinator that simply has not
started its next experiment yet; conflating the two would either mask a
real misconfiguration behind endless retries, or fail deployment for a
perfectly normal idle period. `TestPersistentWorkerPermanentAuthFailureStopsRetrying`
proves the failure surfaces in well under a second, not after exhausting a
multi-second timeout.

### mTLS identity is never weakened

Nothing about certificate verification changes across a reconnect: each
new TCP connection is a brand-new TLS handshake, so the CA chain,
`serverAuth`/SAN (worker verifying the coordinator) and `clientAuth`/URI
identity (coordinator verifying the worker) checks in
`internal/pki` run in full every single time, not just at the first
connection. There is no cached "already authenticated once" shortcut to
weaken.

### Experiment independence and cross-experiment isolation

The worker process takes no `-experiment-id` flag and never did; Job/
experiment identity has always come entirely from the coordinator's
handshake, per job type. What Task42 fixes is *runtime* behavior: because
every generation rebuilds its computer state from a fresh handshake
(never reusing the previous generation's `workerState`/`TrialWorker`/
`normalizationComputer` value), a worker that served experiment A cannot
answer an experiment B job using A's corpus, metadata, or classes.yaml -
proven by `TestPersistentWorkerHandlesSequentialExperimentsWithoutContamination`,
which runs two scientifically distinguishable experiments back-to-back
through the exact same long-lived worker process and worker cache
directory and checks experiment B's result against *both* its own oracle
(must match) and experiment A's oracle (must not match).

### Corpus/artifact cache: already safe by construction

Task42 section 13 asks for cache keys that include scientific identity, or
else clearing experiment-local state between experiments. The existing
`-remote-cache-dir` staging (Task34) already keys every cached file by its
own SHA256 content hash (`GET /v1/input/<hash>`, verified again on
receipt). Two different experiments' inputs cannot collide in that cache
unless they are byte-identical - in which case reusing the cached copy is
scientifically correct, not contamination. No new cache-invalidation code
was needed; this is a case where the existing design already satisfied
the requirement, not a gap.

Unbounded growth of that cache directory across many sequential
experiments over a long deployment's lifetime is a real but separate
operational concern (disk usage, not correctness); the Ansible `absent`
operation removes it entirely, and operators running very many
experiments against one long-lived deployment can prune it manually
between runs if needed.

## Ansible: deploy is no longer one experiment's lifecycle

`ansible/roles/voynich_worker` gains three operations alongside the
existing `present`/`absent` (see its README for full detail):

- **`started`**: ensure the already-deployed process is running - no
  redeploy, no coordinator-reachability requirement.
- **`stopped`**: gracefully stop the process, leave everything else
  deployed.
- **`status`**: report deployed/running/last-lifecycle-state per host;
  never fails the play.

And `present`'s readiness check no longer fails deployment just because no
coordinator is reachable yet - only for a real process-startup failure or
a genuine mTLS identity error (see `readiness.yml`, and
`DISTRIBUTED_EXECUTION_OPERATIONS.md`'s "Between experiments" section).
This is the direct fix for the hang in Task42's own problem statement.

## What was deliberately not built

- No change to which certificates a worker trusts, how it derives
  `WorkerID`, or any lease/retry/JobID semantics - Task42 is purely about
  *when* a worker (re)connects and rebuilds state, never *whether* it is
  authenticated correctly.
- No new coordinator-side "expected worker count" gate.
  `pipeline-orchestrate` and every coordinator binary already serve
  whichever authenticated workers happen to be connected whenever they
  connect - a persistent worker arriving before, during, or after the
  coordinator's own startup all reach the same steady state, so there was
  no "wait for N workers" barrier to add. The orchestrator's existing
  once-a-minute "stage still running, elapsed Ts" heartbeat
  (`pipeline-orchestrate/exec.go`) already gives elapsed-time visibility
  while a distributed stage runs; the coordinator now additionally logs
  each successful worker handshake, so `tail -f` on a stage log shows
  workers arriving in real time.
- No systemd unit / boot-time persistence: workers still live under
  `/tmp` and are expected not to survive a reboot (Task35's original
  design); Task42's persistence is about surviving coordinator lifecycle
  changes while the worker process itself keeps running, not about
  surviving the host rebooting.

## Validation

- `go build ./...`, `go vet ./...`, `go test ./...`, `go test -race ./...`
  all pass (see `NORMALIZATION_COMPARE_DISTRIBUTION_AUDIT.md`'s
  validation section for the exact commands and scope).
- New tests in `internal/conditionalregime/persistent_worker_test.go`:
  reconnect after coordinator restart (same experiment), sequential
  experiments without contamination (different experiment, same worker
  process and cache directory), coordinator starting late, bounded
  backoff (not a tight loop), and fail-fast on a permanent CA/certificate
  failure.
- Ansible role changes were syntax-checked
  (`ansible-playbook --syntax-check`) for every new state against both the
  role directly and the real `ansible/deploy-workers.yml` +
  `inventory.example.yml`. They were **not** exercised against real
  inventory hosts in this session: a real Doyle-corpus production pipeline
  run (`experiments/doyle-sign-of-four-v2`) was actively using the real
  worker fleet and coordinator port at the time (see the audit document's
  scaling-study note) - deliberately avoided to not disturb it. Validate
  the Ansible operations against a real inventory before relying on them
  in production, or once that run has completed.
