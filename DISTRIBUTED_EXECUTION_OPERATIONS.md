# Remote Distributed Execution Operations

Task33 added a trusted-machine HTTP coordinator/worker mode to
`conditional-regime-analyze`. Task34 adds mutual-TLS authentication backed by
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
`-dns`/`-ip` values from Section 2), then presents its own certificate.
There is no insecure-skip-verify path in any build configuration. A worker
that cannot yet reach the coordinator (not started, restarting) backs off
and keeps retrying rather than exiting - a worker's lifecycle is controlled
by the operator (SIGINT), not by any one coordinator run. Start any number
of workers, in any order relative to the coordinator, each with its own
certificate.

SIGINT initiates graceful shutdown on both sides: the coordinator's HTTP
server drains in-flight requests for up to ten seconds, and a worker's lease
loops exit once their current job (if any) is posted back.

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

At coordinator startup both input files are SHA256-hashed by the existing
loader, exactly as in Task33. A worker fetches each input once, by hash,
from `GET /v1/input/<sha256>` on the coordinator (the direction is reversed
from Task33's coordinator-pushes-to-worker `PUT`, since the worker is now
the one dialing in) and caches it locally under `-remote-cache-dir`,
skipping any hash it already has on disk. Each lease's job is scoped to one
experiment fingerprint; a worker recomputes that fingerprint from its own
staged corpus/metadata bytes and parameters before ever computing a job, so
a stale cache object or mismatched configuration cannot execute as the
requested experiment.

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
  checkpoint intact) for one to reconnect; re-run a worker at the same
  `-coordinator` URL to resume progress.

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

- A running coordinator (Sections 1-4 above) reachable from every worker
  host over HTTPS.
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
ansible-playbook -i inventory.yml deploy-workers.yml                     # all workers
ansible-playbook -i inventory.yml deploy-workers.yml --limit worker1     # one host
ansible-playbook -i inventory.yml deploy-workers.yml -e serial_size=1    # rolling, one at a time
```

A successful run means: the binary and this host's unique credentials are
installed under `voynich_worker_install_dir` (default `/tmp/voynich-worker`),
the worker process is started (surviving the SSH session that launched it),
and - the deployment fails otherwise - a `GET /v1/handshake` probe from the
worker host, using the exact credentials just installed, returned HTTP 200
from the coordinator. That is the same mTLS-authenticated path the real
worker uses at its own startup handshake: a revoked, foreign, expired or
otherwise rejected certificate fails this check exactly as it would fail
the real worker, so "deployed" always means "authenticated," not just
"a process is running."

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
