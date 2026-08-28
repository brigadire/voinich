# Frozen G1-v2 worker deployment contract

## Required closure

A production worker is a Linux/amd64 machine running the exact `g1v2-executor` artifact with SHA-256 `6b015b2e4078b9b5f109ebf3aa8d73918888e431bde267e0d10c3013b524f718`. The compatibility tuple is `g1v2-execution-v1`, `g1v2-evidence-v1`, and `canonical-json-nfc-v1`. Deployment may not build, download, patch, or silently replace this artifact during the frozen run.

The worker receives `ca.crt` and one unique worker certificate/private-key pair. It never receives `ca.key`. The certificate must contain a `voinich-worker:` URI SAN; authenticated identity is derived by the coordinator from that SAN. Inventory hostname and `--host` are telemetry only.

## Operational configuration

The HTTPS coordinator URL, host targeting, storage paths, slot count, service backend and resource limits are operational inputs. They are excluded from scientific manifest identity. Changing them must not change any seed, threshold, dependency, JobID, result schema, evidence rule, retry rule, or aggregation rule.

Each slot runs as the dedicated `voynich-worker` account under systemd or OpenRC, starts at boot, restarts after process failure, and continually reconnects to the coordinator. The worker opens outbound TCP/TLS to the configured coordinator only; no application ingress is required on workers.

The local cache and publication temp tree live under `/var/lib/voinich-g1v2`, not `/tmp`. Cache data survives service restart and decommission by default. Preseed objects are named by and verified against SHA-256. Only stale invalid `.publish-*` temporary files may be cleaned automatically; verified evidence is outside worker cleanup scope.

Coordinator-visible authoritative evidence storage is distinct from worker-local cache. When the coordinator evidence path is managed on a host, deployment must verify at least 120,000,000,000 free bytes. Worker cache must independently meet the coordinator-required operational free-space threshold before service start.

## Admission and failure behavior

Deployment fails before service start for an incompatible platform/service manager, non-HTTPS coordinator, wrong frozen tuple, CPU/RAM oversubscription, missing input, wrong source binary hash, missing worker URI SAN, shared private key within a fleet play, installed binary hash drift, preseed hash mismatch, or insufficient configured capacity. A coordinator separately performs the authoritative mTLS and compatibility handshake.

Infrastructure loss expires a lease and requeues the same immutable bundle. It does not create a scientific failure or modify the manifest. Equal verified content is counted once under the frozen executor rules; Ansible implements no scheduler, queue, retry policy, JobID scheme, or evidence store.

## Change control

For this frozen run, binary/config template or credential changes trigger a controlled worker service restart. A future executable requires a new explicit frozen hash and experiment change boundary; operators must never alter the hash merely to make an unreviewed binary deploy. `state=absent` stops and disables services and removes executable/credentials; cache removal requires the explicit `voynich_g1v2_remove_cache_on_absent=true` decision. Certificate revocation/deny-list update is an offline coordinator-PKI operation.
