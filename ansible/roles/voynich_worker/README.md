# voynich_worker

Task35 Ansible role: builds, installs, configures, starts, verifies and
removes Task33/34 `conditional-regime-analyze` remote mTLS workers. Workers
are deliberately ephemeral - they live entirely under `/tmp` on the target
host - and the role never deploys a coordinator or a CA private key.

See `DISTRIBUTED_EXECUTION_OPERATIONS.md` at the repo root for the
end-to-end operational story (CA generation, certificate issuance,
starting the coordinator). This README covers the role itself.

## What this role manages, and what it never touches

- Manages: one worker process per target host, its binary, its own
  certificate/key, the copied project CA certificate, its cache/log
  directories, and a PID file - all under one directory
  (`voynich_worker_install_dir`, default `/tmp/voynich-worker`).
- Never manages: the coordinator process, `ca.key` (the project CA's
  private key - only `ca.crt` is ever copied to a worker), or any
  scientific/RNG/JobID/checkpoint state (those live entirely on the
  coordinator and are untouched by this role).

## Requirements

- Target hosts: Linux with `/proc` (used for safe PID validation), `curl`
  (used for the post-start mTLS readiness probe), and either a Go
  toolchain (`build_on_target`/default controller build path) or none at
  all (`prebuilt` mode).
- Controller: Go toolchain (default `copy_from_controller_build` mode
  builds there), `git` (best-effort, for deployed-binary diagnostics only).
- Already-issued PKI material from `conditional-regime-pki` (Task34):
  `ca.crt` and one `worker-<id>.crt`/`worker-<id>.key` pair per host. This
  role never generates or renews certificates itself - see
  `DISTRIBUTED_EXECUTION_OPERATIONS.md`.

## Core state model

```
voynich_worker_state: present|started|stopped|status|absent
voynich_worker_install_dir: /tmp/voynich-worker   # default
```

The worker binary is a Task42 **persistent** worker: it reconnects to its
coordinator across restarts, absences, and different experiments entirely
on its own (see `REMOTE_WORKER_LIFECYCLE.md`). That decouples *deploy*
from *experiment lifecycle* - the normal workflow is deploy once, then run
as many experiments as you like:

```
present (once)  ->  started  ->  run experiment A  ->  run experiment B  ->  ...  ->  absent (once)
                       ^________________________________________________________|
                       (started is also how you resume after a deliberate `stopped`)
```

`present`: build/obtain the worker binary, create the managed directory,
install the binary/CA/this host's certificate+key, start the worker, and
verify readiness. Readiness still fails the play on a real identity
problem (rejected certificate, untrusted CA) or if the process itself
never came up - but **not** merely because no coordinator is reachable
yet, which is expected and normal before the first experiment exists.

`started`: ensure the already-deployed worker process is running - no
redeploy, no certificate/binary copy, and (like `present`) does not
require a coordinator to be reachable right now. This is what a "run the
next experiment" playbook uses; it fails clearly if the host was never
`present`-deployed at all.

`stopped`: gracefully stop the managed process (SIGINT, bounded wait,
SIGKILL only after that timeout) but leave the binary, certificates,
cache, and logs in place. Use this to pause a fleet, or before rotating a
certificate/binary out of band. `started` resumes it.

`status`: read-only report of what this host's worker is actually doing
right now (not deployed / deployed-but-stopped / running plus its most
recent lifecycle log line - `coordinator unavailable`, `connected`,
`authenticated`, `registered`, or `disconnected`). Never fails the play,
so it is safe to run across a whole inventory as a health check.

`absent`: stop the exact managed process (graceful SIGINT, bounded wait,
SIGKILL only after that timeout), then remove every managed artifact
(credentials, caches, binary, the complete managed directory) and verify
no managed process remains. Safe to run repeatedly. This is the only state
that needs redeploying afterward (`present` again) before workers can run
another job.

## Variables

| Variable | Default | Meaning |
|---|---|---|
| `voynich_worker_state` | `present` | `present`, `started`, `stopped`, `status`, or `absent`. |
| `voynich_worker_install_dir` | `/tmp/voynich-worker` | Managed directory root. Must be under `/tmp`. |
| `voynich_worker_coordinator_url` | `""` (required) | `-coordinator` value; must be `https://`. |
| `voynich_worker_concurrency` | `1` | `-remote-concurrency` value. |
| `voynich_worker_ca_src` | `""` (required) | Controller path to `ca.crt`. Shared across hosts - it is only a certificate. |
| `voynich_worker_cert_src` | `""` (required) | Controller path to **this host's own** `worker-<id>.crt`. |
| `voynich_worker_key_src` | `""` (required) | Controller path to **this host's own** `worker-<id>.key`. Never share across hosts (see phase 7 guard below). |
| `voynich_worker_id` | `{{ inventory_hostname }}` | Cosmetic label only (log/metadata file naming). The real identity always comes from the certificate's `voynich-worker://` URI SAN - this variable is never treated as authoritative and is never sent to the coordinator. |
| `voynich_worker_allow_shared_key` | `false` | Set `true` to permit two or more hosts in one play to resolve to an identical private key (scratch/test environments only). |
| `voynich_worker_build_mode` | `copy_from_controller_build` | `copy_from_controller_build`, `build_on_target`, or `prebuilt`. |
| `voynich_worker_repo_path` | `{{ playbook_dir }}/..` | Repo root (contains `go.mod`); used by the two build-from-source modes. |
| `voynich_worker_go_bin` | `go` | Go toolchain to invoke. |
| `voynich_worker_controller_build_dir` | `/tmp/voynich-worker-controller-build` | Controller-local scratch build output; never part of the git tree. |
| `voynich_worker_binary_src` | `""` | Prebuilt binary path; required when `build_mode: prebuilt`. |
| `voynich_worker_goos` / `voynich_worker_goarch` | `""` (auto-detected) | Override when a target's architecture can't be inferred from `ansible_facts`. |
| `voynich_worker_process_manager` | `pidfile` | Only `pidfile` is implemented (see below); reserved for a future `systemd-run` mode. |
| `voynich_worker_log_dir` | `{{ install_dir }}/log` | Worker stdout/stderr. |
| `voynich_worker_cache_dir` | `{{ install_dir }}/cache` | `-remote-cache-dir`. |
| `voynich_worker_stop_timeout_seconds` | `15` | Bounded graceful-shutdown wait before SIGKILL. |
| `voynich_worker_readiness_timeout_seconds` | `20` | Bounded wait for the post-start mTLS probe. |
| `voynich_worker_readiness_poll_interval_seconds` | `1` | Probe poll interval. |

## Build strategy (phase 4)

`copy_from_controller_build` (default) computes the set of **unique**
GOOS/GOARCH pairs actually present among the play's target hosts and runs
`go build` on the controller exactly once per pair - never once per host,
even across a large fleet of identical machines - then copies the matching
binary to each host. `build_on_target` runs `go build` directly on each
host (needs a Go toolchain and this repo's source there).  `prebuilt`
copies an already-built binary from `voynich_worker_binary_src`, for
offline/reproducible-artifact workflows.

Every mode writes `{{ install_dir }}/bin/VERSION.json` after installing the
binary, recording the deploy timestamp, build mode, GOOS/GOARCH, the
binary's own SHA256, and (best-effort) the controller's git commit - so a
deployed worker's exact provenance is always inspectable on the host
without running it. `deployed_at` refreshes on every `present` run by
design (it is a "last verified" timestamp); this does not restart the
worker and is not treated as a meaningful change.

## Process lifecycle (phase 6)

A rendered `run/start.sh` launches the worker with
`setsid nohup ... &`, redirected stdio, and `disown` - it survives the SSH
session that launched it. Its PID is written to `worker.pid`. Stopping
never trusts a bare PID number: every stop/restart/removal path first
checks that `/proc/<pid>/cmdline` still references this exact
`{{ install_dir }}/bin/conditional-regime-analyze` path (see
`tasks/is_running.yml`) before signaling anything - never a broad
process-name match, and never an action on a PID that has since been
recycled by an unrelated process. Graceful shutdown sends **SIGINT**
(matching the worker's own `signal.NotifyContext(context.Background(),
os.Interrupt)` shutdown path in Go), waits up to
`voynich_worker_stop_timeout_seconds`, and only then sends SIGKILL -
`tasks/stop.yml` fails loudly if the process somehow survives both.

Re-running `present` when nothing changed is a no-op (no restart). A
changed binary, certificate, key, or start script (any of which can arise
from a normal redeploy) notifies a handler that stops-then-starts exactly
once, applied via `meta: flush_handlers` before the idempotent
"ensure running" check - so a real change never produces a half-restarted
or double-started worker.

## Per-host mTLS mapping (phase 7)

`voynich_worker_cert_src`/`voynich_worker_key_src` are meant to be set in
`host_vars` (see `inventory.example.yml`), one unique pair per host. Before
installing anything, the role checksums (on the controller) every target
host's resolved private-key path and refuses to proceed if two or more
hosts in the same play resolve to an identical file, unless
`voynich_worker_allow_shared_key: true` is set explicitly - intended only
for a scratch/test environment where distinguishing workers doesn't
matter. WorkerID is always the identity the coordinator derives from the
certificate's `voynich-worker://` URI SAN (Task34); `voynich_worker_id` is
a display label only.

## Readiness (phase 8)

After starting, the role verifies more than "the process exists": it
confirms the PID is alive (safe cmdline check, above), then - from the
worker host itself, using the exact certificate/key/CA files just
installed - calls `GET /v1/handshake` on the coordinator with `curl`. That
endpoint is the same mTLS-authenticated path the real worker process uses
at its own startup handshake (`internal/conditionalregime/remote.go`), so
a `200` is proof the coordinator's CA trusts this certificate, the EKU/SAN
checks passed, and the certificate is not revoked - not merely that some
TCP port answered. A revoked, foreign, expired, or otherwise rejected
certificate fails this probe the same way it would fail the real worker,
and the whole deployment fails (not just warns) if this does not succeed
within `voynich_worker_readiness_timeout_seconds`.

## Removal (phase 9)

`state: absent` stops the exact managed process (graceful, bounded, force
only after timeout, with an internal verification that the *specific* PID
we signaled is actually gone - recorded before any file is deleted, since
by the time removal finishes `worker.pid` no longer exists to re-check),
then removes: the worker key and certificate and the copied CA, the cache
directory, the binary, and finally the complete managed directory. It only
ever touches `voynich_worker_install_dir` - never a broader `/tmp` sweep -
and is safe to run again with nothing to do.

## `/tmp`/reboot semantics (phase 10)

Workers are deliberately not made to survive a reboot: there is no
systemd/init boot-persistence unit installed by default, and `/tmp` being
cleared (by the OS or by hand) is an expected, tolerated event - the
coordinator already tolerates a worker vanishing mid-lease and reassigns
its work (Task34). Re-running `present` after `/tmp` loss recreates the
worker from scratch with no special handling required.

## Fleet targeting (phase 11)

Ordinary Ansible primitives, no custom scheduler:

```bash
ansible-playbook -i inventory.yml deploy-workers.yml                                   # all workers
ansible-playbook -i inventory.yml deploy-workers.yml --limit worker1.example.internal  # one host
ansible-playbook -i inventory.yml deploy-workers.yml --limit a_group                    # a subset/group
ansible-playbook -i inventory.yml deploy-workers.yml -e serial_size=1                   # rolling, one at a time
ansible-playbook -i inventory.yml deploy-workers.yml -e voynich_worker_state=absent     # remove everywhere targeted
```

## Security (phase 12)

- `ca.key` is never a valid value for `voynich_worker_ca_src`/`_cert_src`/`_key_src`
  (the role asserts none of them resolve to a file literally named `ca.key`,
  and that key/cert/ca sources are three distinct files).
- The private-key install task uses `no_log: true` and `diff: false`, so
  neither its content nor a diff of it is ever printed.
- `voynich_worker_coordinator_url` must be `https://` - asserted before
  anything else runs.
- The worker binary never has an insecure-TLS mode to accidentally enable;
  this role has no "skip verification" variable because the underlying
  binary has no such flag (Task34).
- **Protecting worker private keys in Ansible material**: this role never
  requires a new secret-management service. If `voynich_worker_key_src`
  points at files checked into an Ansible project (rather than issued
  fresh per deploy and kept elsewhere), encrypt them with
  [Ansible Vault](https://docs.ansible.com/ansible/latest/vault_guide/index.html)
  (`ansible-vault encrypt files/certs/worker-1/worker.key`) or an
  operator-controlled equivalent (e.g. a password manager's file share, a
  restricted-permission path outside the repo). The role only reads
  whatever path `voynich_worker_key_src` names at run time - it does not
  care where or how that file is protected at rest.

## Idempotency and `--check` (phase 13)

Every state transition below was exercised end-to-end against a real
Task34 coordinator+CA (see "End-to-end acceptance" below): present on a
clean host; present again (idempotent - only a diagnostic timestamp
changes, the worker process itself is not restarted); confirmed
running/authenticated; absent; absent again; present after absent; present
after simulated `/tmp` loss (process killed, directory deleted, `present`
recreates cleanly); a renewed certificate triggering exactly one
controlled restart; and a changed binary triggering exactly one controlled
restart.

`--check` mode works for everything that does not require a real running
process: variable/HTTPS/credential-source assertions, the shared-key
guard, and what `copy`/`template` would change. It cannot start a real
process or perform a real mTLS handshake (there is nothing real to
start/probe), so the idempotent-start and readiness steps are skipped
under `--check`, and `stop.yml`'s wait/force-terminate/confirm sequence is
skipped there too (the SIGINT/SIGKILL themselves are only simulated by
Ansible under `--check`, so there is nothing to wait for or confirm).

## Example inventory/playbook (phase 14)

See `../inventory.example.yml` and `../deploy-workers.yml` at the `ansible/`
root - two placeholder hosts, each with its own certificate/key path, no
real hosts/keys/passwords/production certificates.

## End-to-end acceptance (phase 15)

Validated locally (`ansible_connection: local`, a real Task34 coordinator
and CA built from this repository, two distinct worker identities): deploy
succeeds; the mTLS readiness probe passes; the coordinator's own log and
`/v1/metrics` counters show the deployed worker(s) leasing and completing
jobs; a small deterministic distributed run through the deployed worker(s)
produced output byte-for-byte identical (`sha256sum`/`diff -rq`) to the
existing sequential-goroutine oracle, both with one worker and with two
workers using distinct certificates; `state: absent` stopped the process
and removed `/tmp/voynich-worker-<host>` completely; and a subsequent
`present` redeployed successfully from nothing.
