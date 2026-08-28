# Worker compatibility contract

A worker is admitted only after project-CA mTLS authentication and an exact handshake match on protocol version, executable/code SHA-256, GOOS, GOARCH and the runtime embedded in the executable. It must also report usable content-addressed cache access and at least the coordinator-required free storage. Certificate identity is accepted only from the verified URI SAN; request JSON cannot choose it.

The frozen executable is Linux/amd64 SHA-256 `6b015b2e4078b9b5f109ebf3aa8d73918888e431bde267e0d10c3013b524f718`. Workers may have different hostnames, installed Go toolchains, CPU counts and storage paths. Those are operational provenance and never enter JobID or scientific hashes.

To add a worker: provision the unchanged binary; provision `ca.crt` and a unique worker certificate/key (never `ca.key`); configure the coordinator URL and a writable SHA-256 cache; verify free evidence/cache capacity; start `g1v2-executor worker`. No scientific manifest, seed, threshold, dependency or JobID changes are permitted.

Tests accepted `cognition-v2` from its certificate, rejected wrong code compatibility, and the inherited `internal/pki` suite rejects unknown-CA, wrong-EKU, expired, hostname-invalid and deny-listed identities.
