# G1-v2 worker operations

## Inventory inputs

For each Linux/amd64 worker, issue a unique certificate/key whose URI SAN is `voinich-worker:<identity>`. Keep the CA key offline. Put only paths to `ca.crt`, that host's certificate/key, the HTTPS coordinator URL, and a conservative slot count in inventory or command-line variables. Start at 2–4 GB RAM per heavy slot and never exceed either discovered vCPU count or the configured RAM floor.

The supplied inventory does not include `cognition` in `voynich_workers`, so the audited invocation uses the documented target selector instead of editing the shared external inventory:

```bash
ANSIBLE_CONFIG=/home/brigadire/devops/workdir/ansible/ansible.cfg \
ansible-playbook \
  -i /home/brigadire/devops/workdir/inventory_dev/voinich/hosts \
  ansible/deploy-workers.yml \
  -e voynich_worker_target_group=cognition \
  -e voynich_worker_coordinator_url=https://10.10.24.107:38490 \
  -e voynich_worker_concurrency=4 \
  -e voynich_g1v2_binary_src=/path/to/frozen/g1v2-executor \
  -e voynich_worker_ca_src=/secure/path/ca.crt \
  -e voynich_worker_cert_src=/secure/path/worker-cognition.crt \
  -e voynich_worker_key_src=/secure/path/worker-cognition.key \
  -e voynich_g1v2_evidence_dir=/usr/local/data/voinich-evidence
```

The controller-side executable must hash to `6b015b2e4078b9b5f109ebf3aa8d73918888e431bde267e0d10c3013b524f718`. For the full fleet, add each host to `voynich_workers`, place unique paths in `host_vars`, and omit the target override. Use `serial_size=1` for rolling replacement.

## Preflight and deployment

1. Confirm the coordinator certificate covers its DNS name or IP and that workers can make outbound TCP/TLS connections to its port. No worker ingress rule is needed.
2. Archive and independently hash the frozen executable. Never build it on target.
3. Run `ansible-playbook --syntax-check`, then `--check` with the production variables.
4. Run `state=present`. The role checks platform, resources, source hash, PKI inputs, persistent capacity, installed hash and service activity.
5. Inspect the service and executable hash. Then run only the approved engineering compatibility fixture through the actual coordinator. A process merely being `started` is not a substitute for successful mTLS/compatibility admission.
6. Run the identical deployment again and require `changed=0` before freezing the fleet.

## Cache preseed

Supply immutable inputs as a list:

```yaml
voynich_g1v2_preseed:
  - src: /secure/frozen-inputs/object.bin
    sha256: 0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef
```

The role copies each object to `{{ voynich_g1v2_cache_dir }}/<sha256>` as `0440` and verifies its content afterward. Run preseed before the blind start and repeat the role to verify reuse. Do not put authoritative evidence in the worker cache.

## Routine lifecycle

Use `-e voynich_worker_state=status` for a read-only slot report, `started` to enable/start an already hash-verified deployment, and `stopped` for a clean pause. These states refuse a missing or drifted binary and do not redeploy it. Systemd/OpenRC automatically restarts a crashed process and the frozen worker reconnects after coordinator/network interruption.

When changing credentials, service configuration, or a reviewed frozen payload, use `present` with the complete inputs and preferably `serial_size=1`. The role restarts configured slots only when managed content changes. Do not change executable hash during a frozen run.

## Capacity and evidence

The worker cache and temp paths default to `/var/lib/voinich-g1v2/cache` and `/var/lib/voinich-g1v2/tmp`. The authoritative evidence volume is coordinator-visible and separate. For the audited coordinator host it is `/usr/local/data/voinich-evidence`; the role observed 178,870,026,240 bytes free and enforced a minimum of 120,000,000,000. Capacity failure is a deployment failure, never a reason to delete verified evidence or change scientific batch definitions.

## Addition and removal

To add a worker: prepare Linux/amd64 and systemd/OpenRC; issue its unique certificate; add inventory/host vars; choose slots from CPU/RAM; set persistent cache and network endpoint; run `present`; verify hash/service/mTLS handshake; and require one accepted engineering job. No scientific manifest edit is allowed.

To remove one: run `stopped` (or `absent` for decommission); allow outstanding leases to expire and verify coordinator completion/requeue; optionally add the certificate identity to the existing coordinator deny list or revoke it through the established PKI procedure; then remove the host from inventory. `absent` removes credentials and executable but preserves cache unless explicitly told otherwise. No result, JobID, seed, or manifest changes.

## Incident checks

On failure, check platform facts, installed SHA-256, certificate validity/URI SAN, service logs under journald or `/var/lib/voinich-g1v2/log`, outbound TLS reachability, cache/evidence free bytes, and slot resource sizing. Never disable TLS verification, share a worker key, copy `ca.key`, loosen executable ownership, retry a scientific failure as infrastructure, or tune scientific definitions to fit a host.
