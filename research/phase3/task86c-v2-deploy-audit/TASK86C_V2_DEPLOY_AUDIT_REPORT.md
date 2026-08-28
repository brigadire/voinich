# Task86C-v2 deployment audit report

## Outcome

The inherited role was not sufficient: it installed the legacy Phase-I executable, allowed builds, used volatile `/tmp`, and had no boot-persistent G1-v2 service, frozen hash gate, storage checks, preseed mechanism, or G1-v2 compatibility closure. All identified gaps were deployment bugs, missing infrastructure features, or documentation gaps. No scientific-contract conflict was found.

The corrected default G1-v2 profile now deploys the exact frozen executable, validates Linux/amd64 and resource sizing, provisions protected per-worker mTLS material without `ca.key`, creates persistent cache/temp paths, supports verified CAS preseed, validates cache and optional coordinator evidence capacity, and manages unprivileged systemd or OpenRC service slots. The old implementation remains isolated behind `voynich_worker_profile: phase1`.

## Actual deployment evidence

The supplied inventory's `cognition` host was used as the clean-equivalent target. The first run exposed and led to correction of two real portability/idempotency bugs: cognition uses OpenRC rather than systemd, and embedding changing free-space telemetry in service files caused unnecessary restarts. After correction, a repeat production run reported `ok=33 changed=0 failed=0`.

The remote executable hash was exactly `6b015b2e4078b9b5f109ebf3aa8d73918888e431bde267e0d10c3013b524f718`. Four OpenRC services were enabled in the default runlevel and active. The binary was root:root 0755; CA, worker certificate and worker key were root:voynich-worker 0640; cache/temp/evidence paths were voynich-worker:voynich-worker 0750; no `ca.key` existed below the install root. cognition reported Linux/x86_64. Available capacity was 12,645,380,096 bytes for the worker cache filesystem and 178,870,026,240 bytes for the evidence filesystem.

A newly started coordinator accepted the already-running workers without a worker restart and completed the 193-job open engineering fixture over real cross-node mTLS. A synchronized engineering fault test stopped all four slots for three seconds with two-second leases, restarted them, and completed 25/25 successfully with `retry_sum=4`, `max_retry=1`, and `lease_history_max=2`. This proves actual lease expiry/requeue, not merely reconnect.

Negative tests rejected `/bin/true` as the payload before copy, rejected a coordinator certificate lacking the worker URI SAN, and rejected an impossible cache threshold before service changes; each recap had `changed=0`. A preseeded non-confirmatory audit object was installed by content hash as 0440, post-copy verified, and remained byte-identical after a service restart.

Final validation passed Ansible syntax-check, a correct full check-mode run (`ok=24 changed=0 failed=0`), read-only lifecycle status (`ok=9 changed=0 failed=0`), `git diff --check`, both authoritative prep validators, and focused `internal/g1v2`/`cmd/g1v2-executor` Go tests. `ansible-lint` was not installed and is recorded as `NOT_TESTED`, not PASS. The inventory's configured callback emitted a non-fatal Ansible 2.19 compatibility warning; task execution and recaps were unaffected.

## Infrastructure changes

- `defaults/main.yml`: frozen values, persistent paths, capacity/resource/preseed settings, G1-v2 default profile.
- `tasks/main.yml`, `g1v2.yml`, `g1v2_present.yml`, `g1v2_lifecycle.yml`, `g1v2_absent.yml`: profile dispatch, fail-closed checks, deployment and lifecycle.
- `templates/g1v2-worker.service.j2`, `templates/g1v2-worker.openrc.j2`, `handlers/main.yml`: persistent systemd/OpenRC service management and controlled restarts.
- `ansible/deploy-workers.yml`: facts/become and ordinary selectable inventory target.
- role README and this audit directory: production contract, procedures, matrices and evidence.

These are `INFRASTRUCTURE_CHANGE`. No Go or scientific implementation was changed by this audit.

## Required verdicts

ROLE_BINARY_DEPLOYMENT = SUPPORTED

ROLE_BINARY_HASH_ENFORCEMENT = SUPPORTED

ROLE_PLATFORM_COMPATIBILITY = SUPPORTED

ROLE_MTLS_PROVISIONING = SUPPORTED

ROLE_CA_KEY_ISOLATION = SUPPORTED

ROLE_COORDINATOR_CONFIGURATION = SUPPORTED

ROLE_SERVICE_MANAGEMENT = SUPPORTED

ROLE_CACHE_CONFIGURATION = SUPPORTED

ROLE_STORAGE_VALIDATION = SUPPORTED

ROLE_CONCURRENCY_CONFIGURATION = SUPPORTED

ROLE_IDEMPOTENCY = SUPPORTED

ROLE_FRESH_HOST_DEPLOYMENT = SUPPORTED

ROLE_WORKER_EXPANSION = SUPPORTED

ROLE_SCIENTIFIC_CONTRACT_PRESERVED = SUPPORTED

VOYNICH_FIREWALL_PRESERVED = SUPPORTED

CONFIRMATORY_FIREWALL_PRESERVED = SUPPORTED

TASK86C_V2_DEPLOYMENT_READY = SUPPORTED

All execution in this audit used known engineering-only fixtures. No blind/confirmatory control, natural-language control, Voynich target, escrow, or scientific verdict was accessed.

TASK86C_V2_DEPLOYMENT_READY_FROZEN
