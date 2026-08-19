# Remote Distributed Execution Operations

Task33 added a trusted-machine HTTP coordinator/worker mode to
`conditional-regime-analyze`. Task40 extends the same protocol, coordinator,
worker binary, PKI, leases and deployment to `structural-projection-analyze`;
no second distributed service is involved. Task34 adds mutual-TLS authentication backed by
a small project-controlled CA and, to make that authentication meaningful,
inverts which side of the connection listens: **the coordinator is now the
mTLS/HTTPS server** (a fixed, addressable identity with DNS/IP SANs), and
**every worker is a TLS client** with its own individual certificate that
dials in, leases a job, computes it, and posts the result back. This is a
transport change only: JobID, RNG, scheduling, checkpoints and every
scientific output are exactly as Task33 left them (see
`DISTRIBUTED_EXECUTION_IMPLEMENTATION.md`'s Task34 section for why the
connection direction had to invert and why nothing scientific did).

There is no bearer token anymore and no unauthenticated loopback exception:
every connection is mutually authenticated by certificate, always.

## 1. Generate the project CA (once, offline)

```bash
go build -buildvcs=false -o conditional-regime-pki ./conditional-regime-pki
./conditional-regime-pki ca -out-dir pki
```

This writes `pki/ca.crt` and `pki/ca.key`. **`ca.key` must never be copied to
a coordinator or worker machine.** Move it to offline/cold storage (an
encrypted volume, a hardware token, or an access-controlled secrets vault)
immediately after issuing the credentials below, and keep at least one
offline backup of it - losing it means you cannot issue or renew any
credential without rotating the whole CA (Section 6). `ca.crt` is not
sensitive: it is the CA's public certificate and is copied to every
coordinator and worker.

## 2. Issue the coordinator's certificate

The coordinator certificate needs every DNS name and/or IP address workers
will actually dial:

```bash
./conditional-regime-pki issue-coordinator \
  -ca-cert pki/ca.crt -ca-key pki/ca.key \
  -dns coordinator.internal -ip 10.20.0.10 \
  -out-dir pki
```

This writes `pki/coordinator.crt` and `pki/coordinator.key`. Identity is
carried only in the SANs, never in Common Name - point every worker's
`-coordinator` URL at a name or address that appears in `-dns`/`-ip` above.

## 3. Issue one certificate per worker

Every worker gets its own certificate and key - never share one across
machines:

```bash
./conditional-regime-pki issue-worker -ca-cert pki/ca.crt -ca-key pki/ca.key -worker-id worker-1 -out-dir pki
./conditional-regime-pki issue-worker -ca-cert pki/ca.crt -ca-key pki/ca.key -worker-id worker-2 -out-dir pki
```

This writes `pki/worker-worker-1.crt`/`.key` and `pki/worker-worker-2.crt`/`.key`.
The worker's identity (`worker-1`, `worker-2`, ...) is carried in a
`voinich-worker://` URI SAN and is what the coordinator will report as
`WorkerID` - it is never taken from any request field. Worker IDs must match
`^[a-z0-9]([a-z0-9-]{0,61}[a-z0-9])?$`.

Copy each worker's own `.crt`/`.key` (never another worker's) and `ca.crt`
(never `ca.key`) to that worker's machine.

## 4. Start the coordinator

The normal scientific flags are unchanged. `-workers` is still the bound on
concurrently in-flight jobs, not a per-host value:

```bash
./conditional-regime-analyze \
  -executor remote \
  -remote-listen 0.0.0.0:8443 \
  -tls-cert pki/coordinator.crt -tls-key pki/coordinator.key -client-ca pki/ca.crt \
  -workers 8 -remote-timeout 20m -remote-retries 3 \
  -corpus data_work/ZL3b-x7.txt \
  -token-metadata-map workdir/metadata-validation/token_metadata_map.tsv \
  -output-dir workdir/conditional-regimes \
  -checkpoint-path workdir/conditional-regimes/checkpoint.json \
  -permutations 1000 -seed 1
```

`-remote-listen` is the coordinator's own mTLS bind address. It never dials
out to a worker; workers dial in. `ca.key` is never given to the coordinator
- only `ca.crt` (as `-client-ca`), to verify connecting workers.

Run the identical command after interruption to resume: the coordinator
loads a checkpoint only with the exact experiment fingerprint, skips every
completed `JobID`, retries missing jobs, and deletes the checkpoint only
after final outputs are written - unaffected by which workers were
connected, or under which certificates, in the run that was interrupted.

For stage 17 the coordinator command is:

```bash
./structural-projection-analyze \
  -executor remote -workers 10 \
  -remote-listen 0.0.0.0:8443 \
  -tls-cert pki/coordinator.crt -tls-key pki/coordinator.key -client-ca pki/ca.crt \
  -remote-timeout 20m -remote-retries 3 \
  -checkpoint workdir/structural-projection-checkpoint.json
```

The production scientific defaults, including `-random-projections 200`, are
unchanged. Keep the operational checkpoint outside frozen outputs.

Task42 adds a third distributable job type, `normalization_compare_baseline`
(one random-baseline trial per threshold), for `normalization-compare`:

```bash
./normalization-compare \
  -executor remote -workers 10 \
  -remote-listen 0.0.0.0:8443 \
  -tls-cert pki/coordinator.crt -tls-key pki/coordinator.key -client-ca pki/ca.crt \
  -remote-timeout 15m -remote-retries 3 \
  -input data_test/pg2097-2.txt \
  -classes workdir/structural_classes.yaml -raw-analysis workdir/sequence_analysis.yaml \
  -normalized-pattern workdir/normalized_%s.txt -analysis-pattern workdir/sequence_analysis_%s.yaml \
  -output workdir/normalization_comparison.yaml
```

See `NORMALIZATION_COMPARE_DISTRIBUTION_AUDIT.md` for the profiling/audit
that justified distributing this stage and the resulting scaling study. It
has no checkpoint flag: unlike conditional-regime-analyze/
structural-projection-analyze, normalization-compare had no pre-existing
resume mechanism, and Task42 did not add one (see that document's stop
condition/scope notes).

Task44 adds five more distributable job types, one per generic-eligible
stage 23-27, each following the exact same `-executor remote -workers N
-remote-listen ... -tls-cert ... -tls-key ... -client-ca ...` shape as
above - only the stage binary and its own scientific/input flags change:

```bash
./token-relation-validate \
  -executor remote -workers 10 \
  -remote-listen 0.0.0.0:8443 \
  -tls-cert pki/coordinator.crt -tls-key pki/coordinator.key -client-ca pki/ca.crt \
  -remote-timeout 15m -remote-retries 3 \
  -corpus data_work/ZL3b-x7.txt -discovery-dir workdir \
  -output-dir workdir/token-relation-validation

./replicated-local-structure-audit \
  -executor remote -workers 10 \
  -remote-listen 0.0.0.0:8443 \
  -tls-cert pki/coordinator.crt -tls-key pki/coordinator.key -client-ca pki/ca.crt \
  -remote-timeout 15m -remote-retries 3 \
  -corpus data_work/ZL3b-x7.txt \
  -relation-dir workdir/token-relation-validation -discovery-dir workdir \
  -output-dir workdir/replicated-local-structure

./higher-order-sequence-validate \
  -executor remote -workers 10 \
  -remote-listen 0.0.0.0:8443 \
  -tls-cert pki/coordinator.crt -tls-key pki/coordinator.key -client-ca pki/ca.crt \
  -remote-timeout 15m -remote-retries 3 \
  -corpus data_work/ZL3b-x7.txt \
  -audit-dir workdir/replicated-local-structure -discovery-dir workdir \
  -output-dir workdir/higher-order-sequences

./positional-continuation-validate \
  -executor remote -workers 10 \
  -remote-listen 0.0.0.0:8443 \
  -tls-cert pki/coordinator.crt -tls-key pki/coordinator.key -client-ca pki/ca.crt \
  -remote-timeout 15m -remote-retries 3 \
  -corpus data_work/ZL3b-x7.txt \
  -higher-order-dir workdir/higher-order-sequences \
  -output-dir workdir/positional-continuation

./transition-network-validate \
  -executor remote -workers 10 \
  -remote-listen 0.0.0.0:8443 \
  -tls-cert pki/coordinator.crt -tls-key pki/coordinator.key -client-ca pki/ca.crt \
  -remote-timeout 15m -remote-retries 3 \
  -corpus data_work/ZL3b-x7.txt \
  -token-metadata-map workdir/metadata-validation/token_metadata_map.tsv \
  -output-dir workdir/transition-network
```

Every one of these five accepts `-generic-corpus` in place of
`-token-metadata-map`, exactly like every other generic-eligible stage (see
`GENERIC_STAGE_APPLICABILITY_AUDIT.md`). `token-relation-validate` and
`replicated-local-structure-audit` still need `-checkpoint-path`
overridden explicitly if the default (`<output-dir>/checkpoint.json`)
should live elsewhere; the other three follow the same convention. See
`DISTRIBUTED_GENERIC_STAGES_AUDIT.md` for the dependency/parallelism audit
that justified distributing each of these five stages (three at the
permutation-replicate level, one - higher-order-sequence-validate - at the
whole-candidate level, and one - positional-continuation-validate - at the
whole-battery level) and the resulting scaling study.

Task47 adds a sixth distributable job type, `begin_end_candidate_batch`
(a batch of candidate-pair indexes, not a permutation replicate - see
`BEGIN_END_ANALYZE_DISTRIBUTED_AUDIT.md` section 2), for
`begin-end-analyze`:

```bash
./begin-end-analyze \
  -executor remote -workers 5 \
  -remote-listen 0.0.0.0:8443 \
  -tls-cert pki/coordinator.crt -tls-key pki/coordinator.key -client-ca pki/ca.crt \
  -remote-timeout 15m -remote-retries 3 \
  -corpus data_work/ZL3b-x7.txt -dictionary workdir/dataset/dictionary.yaml \
  -output-dir workdir
```

Like `normalization-compare`, it has no checkpoint flag (no pre-existing
resume mechanism, none added). `-candidate-batch-size` (default 2048)
controls the work-unit granularity; raise it if a production run has many
more eligible tokens than the Astafiev corpus this default was measured
against, but stay mindful of `maxRemoteMessageBytes` (32 MiB as of Task47 -
see `BEGIN_END_ANALYZE_DISTRIBUTED_AUDIT.md` section 4b for why this cap
exists and what happens silently if a batch's marshaled result exceeds
it: not a loud error, but a hang until `-remote-timeout` elapses, repeated
once per retry). See `BEGIN_END_ANALYZE_DISTRIBUTED_AUDIT.md` for the
profiling audit, the granularity study that chose the default batch size,
and the resulting worker-count scaling study (measured on the real
Astafiev corpus; recommend 5 workers - going higher did not help on that
single-machine measurement).

## 5. Start workers

```bash
./conditional-regime-analyze \
  -coordinator https://coordinator.internal:8443 \
  -ca pki/ca.crt -tls-cert pki/worker-worker-1.crt -tls-key pki/worker-worker-1.key \
  -remote-cache-dir /var/lib/conditional-regime-worker/cache \
  -remote-concurrency 4
```

A worker verifies the coordinator's certificate chain and that its SAN
matches the dialed name (`coordinator.internal` above must be one of the
`-dns`/`-ip` values from Section 2), then presents its own certificate -
on every connection, including every reconnect, never only at first
startup. There is no insecure-skip-verify path in any build configuration.
Start any number of workers, in any order relative to the coordinator,
each with its own certificate.

The same worker command serves every supported stage/job type
(conditional-regime part A/B, `structural_projection_trial`,
`normalization_compare_baseline`, and Task44's five new types -
`token_relation_permutation`, `replicated_local_null`,
`higher_order_candidate`, `positional_continuation_battery`,
`transition_network_permutation`): the coordinator's handshake declares
which workload it is running, and the worker builds the matching
computation from that, not from any command-line flag of its own. This is
why none of the five new stage binaries needed a `-coordinator`/persistent-
worker mode of their own - `conditional-regime-analyze -coordinator ...`
already serves every job type any coordinator declares.
Structural projection intentionally executes one trial at a time in each
worker process even when `-remote-concurrency` is larger, because its
scientific core reuses package-level scratch buffers. Scale it with worker
processes/hosts rather than concurrent trials inside one address space.
`higher_order_candidate` and `positional_continuation_battery` jobs are
naturally few in number (one per frozen candidate / one per named battery
- see `DISTRIBUTED_GENERIC_STAGES_AUDIT.md`), so beyond a handful of
workers there is simply no more work to hand out; scale those two stages'
worker count to their job count, not higher.

**Task42: workers are persistent by default.** A worker started with
`-coordinator` (`conditionalregime.RunPersistentRemoteWorker`) is not tied
to one coordinator process or one experiment: it reconnects with bounded
exponential backoff and jitter (1s up to 60s, never a tight loop) whenever
the coordinator is not yet running, has stopped, has restarted, or has
moved on to a different experiment, logging each lifecycle transition
(`coordinator unavailable` / `reconnecting` / `connected` / `authenticated`
/ `registered` / `disconnected`) once per change rather than once per
attempt. It never rebuilds its computer state (corpus, classes/metadata,
scientific fingerprint) mid-connection - only across a reconnect, so one
experiment's data can never leak into another's job. This is what makes
"deploy once, run many experiments" possible; see
`REMOTE_WORKER_LIFECYCLE.md` for the full design and
`ansible/roles/voynich_worker/README.md` for the `present`/`started`/
`stopped`/`status`/`absent` operations built on top of it.

An mTLS identity failure (untrusted CA, rejected/expired certificate) is
never retried forever: the worker classifies it as permanent, logs a clear
diagnostic, and exits non-zero instead of looping - only transport-level
failures (connection refused/reset, timeout, coordinator not listening
yet) are treated as transient and retried.

SIGINT and SIGTERM both initiate graceful shutdown on the worker side (a
persistent worker deployed under Ansible or a process manager is commonly
stopped with SIGTERM, an interactive one with SIGINT - both take the same
path). The coordinator's HTTP server drains in-flight requests for up to
ten seconds on SIGINT, and a worker's lease loops exit once their current
job (if any) is posted back; the outer reconnect loop then exits instead
of scheduling another attempt.

## Compatibility envelope

Every worker still checks the coordinator's declared protocol version,
scientific compatibility ID, GOOS, GOARCH and exact `runtime.Version()` at
handshake, and the coordinator checks the same fields (declared by the
worker) before issuing a lease - both directions of the check Task33 had,
just carried over the new handshake/lease messages instead of one `info`
probe. Build the identical source revision with the identical Go toolchain
on every machine. Any new OS, architecture, Go runtime or CPU family should
be verified against the frozen SHA256 oracle before production use.

## Input staging

At coordinator startup all stage inputs are SHA256-hashed by the existing
loader, exactly as in Task33. A worker fetches each input once, by hash,
from `GET /v1/input/<sha256>` on the coordinator (the direction is reversed
from Task33's coordinator-pushes-to-worker `PUT`, since the worker is now
the one dialing in) and caches it locally under `-remote-cache-dir`,
skipping any hash it already has on disk. Each lease's job is scoped to one
experiment fingerprint; a worker recomputes that fingerprint from its own
staged corpus/metadata bytes and parameters before ever computing a job, so
a stale cache object or mismatched configuration cannot execute as the
requested experiment.

Conditional regime stages the corpus and `token_metadata_map.tsv`;
structural projection stages the corpus, structural-pair table, distance-pair
YAML and family YAML. Invariant inputs are fetched once, never per trial.
Task44's five stages stage the corpus, the metadata map (skipped in
generic mode), and every file in the upstream frozen directory each one
reads (`token-relation-validate`'s and `transition-network-validate`'s own
`-discovery-dir`; `replicated-local-structure-audit`'s `-relation-dir` +
`-discovery-dir`; `higher-order-sequence-validate`'s `-audit-dir` +
`-discovery-dir`; `positional-continuation-validate`'s
`-higher-order-dir`) - staged under content-hash keys prefixed by which
directory they came from (`"discovery:"`, `"relation:"`, `"audit:"`,
`"higherorder:"`) so a worker can reconstruct each directory under its
original filenames (some of these stages' loaders hardcode specific
filenames within one directory) via `reconstructNamedDir`.

## Lease/retry model

The coordinator holds the same "one goroutine per in-flight job" bound
Task33 had (`-workers`), realized now as a lease queue: each in-flight job
is queued once, handed out as a lease (with a unique `LeaseID` - phase 8's
"execution attempt identity", distinct from both `JobID` and `WorkerID`) to
whichever authenticated worker asks next, and reclaimed for another worker
if `-remote-timeout` passes without a result. A job fails outright only
after `-remote-retries` reassignments are all unanswered. Duplicate or late
result delivery for a job already resolved is accepted and ignored, never
double-applied - the same idempotency guarantee Task33 had.

## Revocation

The coordinator supports an explicit deny-list keyed by certificate serial
and/or authenticated `WorkerID`:

```bash
./conditional-regime-pki revoke -deny-list pki/deny.json -worker-id worker-3
./conditional-regime-analyze -executor remote ... -remote-deny-list pki/deny.json
```

The list is read once at coordinator startup; restart the coordinator to
pick up a change. This is the whole revocation mechanism - a full CRL/OCSP
stack is disproportionate for a PKI this small.

## Lifecycle procedures

- **Add a worker**: issue a new `worker-<id>` certificate (Section 3), copy
  it and `ca.crt` to the new machine, start it with `-coordinator` pointed
  at the running coordinator. Nothing else changes.
- **Renew a worker**: re-run `issue-worker` for the same `-worker-id` with
  `-force` (a fresh key and serial, same identity); replace the two files on
  that worker and restart it. In-flight/queued jobs are unaffected - a
  worker's certificate never participates in scientific computation
  (phase 8), only in authenticating the connection.
- **Renew the coordinator**: re-run `issue-coordinator` with `-force`
  before the current certificate expires; replace `coordinator.crt`/`.key`
  and restart the coordinator. Workers reconnect and re-verify normally.
- **Replace a compromised worker**: revoke its identity and/or certificate
  serial (above), issue it (or a new identity) a fresh certificate, and
  decommission the old key. The deny-list keeps the compromised credential
  rejected even though it was never expired.
- **CA rotation**: generate a new CA, issue new coordinator/worker
  certificates from it, and during the transition set `-client-ca`/`-ca` to
  a bundle containing *both* the old and new CA certificates
  (`cat old-ca.crt new-ca.crt > bundle.crt`) so nodes holding either
  certificate remain trusted until every node has rotated. Then drop the old
  CA from the bundle.

## Failure behavior

- A lease that expires unanswered (worker crash, network partition) is
  reassigned to another connected worker, up to `-remote-retries` times,
  exactly like Task33's per-endpoint retry.
- Protocol, runtime, experiment-identity, malformed-request and scientific
  errors are explicit, non-retryable failures surfaced with the JobID.
  Certificate/authentication failures (missing cert, foreign CA, wrong EKU,
  expired, revoked, wrong coordinator SAN) fail the connection at the TLS
  layer before any job-level logic runs.
- If every worker disconnects, the coordinator simply waits (with its
  checkpoint intact) for one to reconnect; a Task42 persistent worker does
  this on its own with bounded backoff, so nothing needs to be re-run
  manually.
- When a coordinator restarts for the *same* experiment (same
  corpus/config, same fingerprint), an already-running persistent worker
  reconnects and resumes leasing without rebuilding anything. When a
  coordinator starts a *different* experiment on the same address, the
  worker's next lease request is rejected (409) instead of silently
  retried forever; it re-handshakes, rebuilds its computer state from the
  new experiment's inputs, and resumes - the old experiment's state never
  contaminates the new one's results.

## Never log secrets

No code path in the coordinator or worker ever logs a private key, or any
field of `ca.key`. The coordinator's default TLS error log records rejected
connection attempts (missing/foreign/expired/revoked certificates) for
audit purposes - this is a security signal, not a secret.

## Fleet deployment with the voynich_worker Ansible role (Task35)

Operators do not need to SSH to every worker host by hand. `ansible/`
(repo root) provides a role that builds, installs, configures, starts,
verifies and removes ephemeral workers under `/tmp` - it manages workers
only, never the coordinator and never `ca.key`. Full variable reference,
lifecycle mechanics, and idempotency evidence are in
`ansible/roles/voynich_worker/README.md`; this section is the quick
end-to-end path.

### Prerequisites

- The coordinator's address and PKI trust anchor (`ca.crt`) - not
  necessarily a coordinator that is running yet. Task42's persistent worker
  connects whenever one first appears, so deploying workers before
  starting the first experiment (Sections 1-4 above) is normal, not an
  error; see `REMOTE_WORKER_LIFECYCLE.md`.
- `ca.crt` and one `worker-<id>.crt`/`worker-<id>.key` pair per worker host,
  already issued via `conditional-regime-pki issue-worker` (Section 3) -
  the role never generates or renews certificates itself.
- Go toolchain on the Ansible controller (default build mode compiles
  there) or a prebuilt binary; `curl` on every worker host (used for the
  post-start mTLS readiness check).

### Variables and per-host certificate mapping

Every worker host gets its own `voynich_worker_cert_src`/`_key_src` in
`host_vars` (or inline per-host, as in `ansible/inventory.example.yml`);
`voynich_worker_ca_src` and `voynich_worker_coordinator_url` are typically
group-level since only `ca.crt` (never `ca.key`) and the coordinator's
address are safe to share. The role refuses to proceed if two hosts in the
same run resolve to an identical private key, unless
`voynich_worker_allow_shared_key: true` is set for a deliberate scratch
environment. See the README's variable table for the complete list
(`voynich_worker_state`, `_install_dir`, `_coordinator_url`,
`_concurrency`, `_ca_src`, `_cert_src`, `_key_src`, `_build_mode`,
`_log_dir`, and more).

### Build strategy

Default `voynich_worker_build_mode: copy_from_controller_build` builds
`go build ./conditional-regime-analyze` once per unique GOOS/GOARCH pair
actually present among the target hosts (never once per identical host),
then copies the matching binary to each; `build_on_target` builds directly
on each host; `prebuilt` copies an already-built binary. Every mode writes
`bin/VERSION.json` on the target recording build mode, GOOS/GOARCH, the
binary's SHA256, and the controller's git commit, so a deployed worker's
provenance is always inspectable without running it.

### Deploy, verify, roll out

```bash
cd ansible
cp inventory.example.yml inventory.yml   # then edit: real hosts, real cert/key paths
ansible-playbook -i inventory.yml deploy-workers.yml                     # all workers, once
ansible-playbook -i inventory.yml deploy-workers.yml --limit worker1     # one host
ansible-playbook -i inventory.yml deploy-workers.yml -e serial_size=1    # rolling, one at a time
```

A successful run means: the binary and this host's unique credentials are
installed under `voynich_worker_install_dir` (default `/tmp/voynich-worker`)
and the worker process is started (surviving the SSH session that launched
it). Readiness still verifies more than "a process exists": it fails the
deployment if the process never came up, or if a `GET /v1/handshake` probe
from the worker host, using the exact credentials just installed, fails
with a certificate/identity error (revoked, foreign CA, wrong EKU,
expired) - the same mTLS-authenticated path the real worker uses at its
own startup handshake. It does **not** fail deployment just because that
probe timed out with no coordinator to answer it: a Task42 persistent
worker deployed before the first experiment exists is the normal case, not
a broken one (see `REMOTE_WORKER_LIFECYCLE.md`). That run instead prints a
warning and keeps going; confirm connectivity later with `status` below.

### Between experiments

Once deployed, a fleet does not need `present` again for a new experiment
- only when the binary, config, or a certificate actually changes:

```bash
ansible-playbook -i inventory.yml deploy-workers.yml -e voynich_worker_state=started  # ensure running, no redeploy
ansible-playbook -i inventory.yml deploy-workers.yml -e voynich_worker_state=status   # health check, never fails
ansible-playbook -i inventory.yml deploy-workers.yml -e voynich_worker_state=stopped  # pause, keep deployed
```

`status` reports each host as not-deployed, deployed-but-stopped, or
running with its most recent lifecycle log line (`coordinator unavailable`
/ `connected` / `authenticated` / `registered` / `disconnected`) - useful
both before starting a coordinator (confirm workers are up and waiting)
and after (confirm they actually connected).

### Removal

```bash
ansible-playbook -i inventory.yml deploy-workers.yml -e voynich_worker_state=absent                  # everywhere
ansible-playbook -i inventory.yml deploy-workers.yml -e voynich_worker_state=absent --limit worker1  # one host
```

Stops the exact managed process (graceful SIGINT, bounded wait, SIGKILL
only after that timeout, with an internal check that the *specific* PID
signaled is actually gone), then removes the worker's key/certificate/copied
CA, cache, binary, and the complete managed directory. Only
`voynich_worker_install_dir` is ever touched; repeated `absent` runs are
safe no-ops.

### `/tmp` and reboot behavior

Workers are deliberately not persisted across reboots (no boot-time service
is installed by default): `/tmp` being cleared is expected and tolerated,
matching the coordinator's own tolerance for a worker vanishing mid-lease
(above). Re-running the role after `/tmp` loss recreates the worker from
scratch.

### Security

Never deploy `ca.key` to a worker (the role asserts none of
`_ca_src`/`_cert_src`/`_key_src` resolve to a file literally named `ca.key`);
private key files are installed `0600` with `no_log`/`diff: false`; the
coordinator URL must be `https://`; and there is no insecure-TLS variable
to accidentally enable, because the underlying binary has no such flag.
To protect worker private keys committed alongside Ansible material, use
[Ansible Vault](https://docs.ansible.com/ansible/latest/vault_guide/index.html)
(`ansible-vault encrypt files/certs/worker-1/worker.key`) or an
operator-controlled equivalent - this role requires no new
secret-management service.

## Running the full pipeline across multiple nodes (Task36)

This is the end-to-end runbook that ties the pieces above together:
`pipeline-orchestrate` (`pipeline-orchestrate/README.md`) runs all 27
current pipeline stages in order. `structural-projection-analyze` (stage 17)
and `conditional-regime-analyze` (stage 21) spread work across the nodes
deployed via the `voynich_worker` Ansible role. Every other stage runs locally on
the machine that runs `pipeline-orchestrate`, regardless of how many nodes
are deployed here.

### 0. Prerequisites

- A Go toolchain and this repository checked out on the machine that will
  run `pipeline-orchestrate` (the "controller/coordinator machine" below -
  it plays both roles: it is the Ansible controller for the worker fleet
  *and* the Task34 mTLS coordinator).
- An Ansible inventory naming the node(s) you intend to use as workers,
  reachable over SSH, with `ansible_host`/`ansible_user`/credentials
  already resolvable the normal Ansible way (`ansible -i inventory.yml all
  -m ping` should succeed before proceeding). **Read every host's
  `host_vars`/`group_vars` before treating an inventory as disposable test
  capacity** - a name like "dev" or "test" does not guarantee the hosts
  aren't running real production services; deploying a CPU-heavy research
  workload onto a shared production host can degrade it.
- One address the worker nodes can actually reach for the coordinator:
  either this machine's LAN IP (workers on the same network) or a
  public/NAT-forwarded address (workers elsewhere on the internet) -
  confirmed reachable *before* relying on it (see step 3).

### 1. Generate the project CA and one certificate per node

```bash
go build -o bin/conditional-regime-pki ./conditional-regime-pki
go build -o bin/pipeline-orchestrate ./pipeline-orchestrate

PKI=/path/to/pki   # outside the repo; never commit ca.key
bin/conditional-regime-pki ca -out-dir "$PKI"
bin/conditional-regime-pki issue-coordinator -ca-cert "$PKI/ca.crt" -ca-key "$PKI/ca.key" \
  -out-dir "$PKI" -dns coordinator.internal -ip 10.0.0.10 -ip 203.0.113.10
# one -worker-id per node, unique, matching the inventory hostnames used below:
bin/conditional-regime-pki issue-worker -ca-cert "$PKI/ca.crt" -ca-key "$PKI/ca.key" -out-dir "$PKI" -worker-id worker1
bin/conditional-regime-pki issue-worker -ca-cert "$PKI/ca.crt" -ca-key "$PKI/ca.key" -out-dir "$PKI" -worker-id worker2
```

Pass every address the coordinator will actually be dialed as to
`issue-coordinator` as `-dns`/`-ip` (repeatable): a node on the same LAN
and a node reached over the public internet may need to dial in on two
different addresses, and both must be in the certificate's SAN list (see
`ansible/inventory.example.yml`'s per-host `voynich_worker_coordinator_url`
override for exactly this case - one worker used the coordinator's LAN
address while the rest used its public one, because the LAN host's own
router could not route its own public IP back to itself).

### 2. Prepare the Ansible inventory

Combine your real connection inventory with a small vars file naming the
per-node certificate/key and the coordinator URL - `ansible/inventory.example.yml`
is a template for this. Keep it outside the repo (or gitignored) if it
names real infrastructure:

```yaml
all:
  children:
    voynich_workers:
      hosts:
        worker1: { voynich_worker_cert_src: "{{ pki }}/worker-worker1.crt", voynich_worker_key_src: "{{ pki }}/worker-worker1.key" }
        worker2: { voynich_worker_cert_src: "{{ pki }}/worker-worker2.crt", voynich_worker_key_src: "{{ pki }}/worker-worker2.key" }
      vars:
        voynich_worker_ca_src: "{{ pki }}/ca.crt"
        voynich_worker_coordinator_url: "https://203.0.113.10:8443"
        voynich_worker_concurrency: 3     # keep modest on shared/production hosts
        voynich_worker_repo_path: /path/to/this/repo/checkout
```

If your real hosts already live in a separate ops inventory with its own
`ansible.cfg` (vault password, SSH key, `remote_user`), pass both:

```bash
ANSIBLE_CONFIG=/path/to/ops/ansible.cfg \
ANSIBLE_ROLES_PATH=/path/to/this/repo/ansible/roles \
ansible-playbook -i /path/to/ops/inventory/hosts -i node-vars.yml deploy-workers.yml
```

### 3. Deploy the workers (before starting the coordinator)

```bash
ansible-playbook -i inventory.yml deploy-workers.yml
```

**This is expected to report `readiness` as `failed` right now** - the
coordinator (step 4) is not running yet, so there is nothing to
authenticate against. The worker daemon itself starts successfully
regardless and will keep retrying with backoff until a coordinator
appears (see `ansible/roles/voynich_worker/tasks/readiness.yml`); nothing
further to do here. Confirm the daemons are actually up:

```bash
ansible -i inventory.yml all -m shell -a "pgrep -af conditional-regime-analyze"
```

To sanity-check the network path before relying on it for the real run,
probe the coordinator's future address from one worker host once you know
it (harmless - either times out cleanly or gets a TLS-level rejection,
never actually authenticates without a real coordinator running):

```bash
ansible -i inventory.yml worker1 -m shell -a \
  "curl -sS -k -o /dev/null -w '%{http_code}\n' --connect-timeout 5 https://203.0.113.10:8443/v1/handshake"
```

### 4. Write the immutable manifest

Do this once, before the real run starts - `manifest` freezes every
scientific parameter (already each stage's own default, see
`pipeline-orchestrate/README.md`) and the exact worker list:

```bash
bin/pipeline-orchestrate manifest -experiment-dir experiments/my-run \
  -executor remote -workers 64 \
  -remote-listen 0.0.0.0:8443 \
  -tls-cert "$PKI/coordinator.crt" -tls-key "$PKI/coordinator.key" -client-ca "$PKI/ca.crt" \
  -remote-timeout 15m -remote-retries 5 \
  -remote-worker worker1 -remote-worker worker2
```

`-workers` bounds the coordinator's total in-flight job slots (independent
of how many nodes connect); size it to comfortably exceed
`node_count × voynich_worker_concurrency`. Inspect both distributed stages in
`experiments/my-run/manifest.json` before proceeding - specifically that
their recorded `args` actually contain your
`-remote-listen`/`-tls-*`/`-client-ca` values, not empty strings.

### 5. Run

```bash
bin/pipeline-orchestrate run -experiment-dir experiments/my-run
```

Stages run one at a time, in order; stages 17 and 21 each start the shared
mTLS coordinator in turn and wait on the deployed nodes. Other stages run
locally, and the nodes remain idle during them. If the
orchestrator process dies, re-running the identical command resumes at
the first incomplete stage (`run-state.json`); a node that drops mid-lease
is simply reassigned, per Task34's normal retry behavior.

To watch progress without disturbing anything, poll the coordinator's own
metrics from any machine holding a valid worker certificate (any node's
own cert works, or the coordinator's own copy in `$PKI`):

```bash
curl -sS --cacert "$PKI/ca.crt" --cert "$PKI/worker-worker1.crt" --key "$PKI/worker-worker1.key" \
  https://203.0.113.10:8443/v1/metrics
# {"handshakes":2,"leases_issued":...,"pending_jobs":...,"outstanding_leases":...}
```

`outstanding_leases` reaching `node_count Ă— voynich_worker_concurrency`
confirms every node is actually contributing, not merely connected.

### 6. Remove the workers, then freeze

Once stage 21 has finished (check `pipeline-orchestrate status`; no need
to wait for the whole pipeline), the nodes are no longer needed:

```bash
ansible-playbook -i inventory.yml deploy-workers.yml -e voynich_worker_state=absent
```

After every stage completes:

```bash
bin/pipeline-orchestrate freeze -experiment-dir experiments/my-run
bin/pipeline-orchestrate verify -experiment-dir experiments/my-run   # optional, re-checks checksums.sha256
```

`freeze` refuses to run against an incomplete run, snapshots only the
files this run actually produced (by modification time - `workdir/` is a
shared, long-lived scratch area, not exclusive to one experiment),
checksums them, writes `REPORT.md`, and writes a read-only `FROZEN`
marker: every later `run`/`freeze`/`manifest` against that directory
refuses without an explicit `-force`, so no later pipeline change can
silently overwrite a frozen baseline.

### Scaling note

Only `conditional-regime-analyze` benefits from more nodes at all - and
only unevenly: its Part B global permutation correction parallelizes well
(measured ~6x faster going from 1 to 10 real nodes), while Part A
significance/refinement showed no measurable benefit from extra nodes in
the same test (too few independent jobs at typical scale to spread
further). Every other stage, and especially `structural-projection-analyze`
(no executor at all, several hours fixed at production scale), sets a wall-
time floor that no number of worker nodes reduces. Budget node count
against Part B's job count, not against the pipeline's total duration.
