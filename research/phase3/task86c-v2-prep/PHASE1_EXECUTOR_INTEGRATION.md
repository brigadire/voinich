# Phase-I executor integration

The implementation reuses `internal/pki` unchanged and follows `internal/conditionalregime/remote.go`: one mTLS coordinator listens, persistent workers dial out, pull ready work, receive a lease ID distinct from JobID, and submit a result bound to the authenticated certificate identity. Lease expiry requeues the same immutable bundle. Workers reconnect/poll; verified CAS indexes reconstruct completion after coordinator restart. The Phase-I content-hash/cache and Ansible deployment convention remain the deployment path.

`internal/g1v2/remote.go` is the G1-v2 protocol adapter required because the old unexported `remotePool` messages are tied to individual Phase-I numeric/blob workloads and cannot carry the frozen multi-stage bundle. This is not a second scheduling policy: readiness, claiming, expiry and identity semantics are the same; the manifest supplies the DAG.

The final cross-node run used coordinator `adelie` (10.10.24.107) and authenticated worker `cognition-v2` on `cognition` (10.10.24.105). The worker dynamically leased and published 193/193 jobs. Its JobID→result-hash graph was `051301c45b12207581d20f42d7349f78b6e5adf8aade1d0bfabfbec6c24a9206`, identical to the local graph. The CA private key was never copied to the worker.
