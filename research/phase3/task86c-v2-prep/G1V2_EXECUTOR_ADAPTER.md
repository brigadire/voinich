# G1-v2 executor adapter

`internal/g1v2` validates the frozen fields, canonicalizes sorted hash/dependency lists, computes JobID from scientific fields only, validates acyclicity, and exposes ready jobs to the Phase-I pull/lease transport. The supported stages are exactly FIT, PREDICTIVE, GENERATION, STRUCTURAL and AGGREGATION. The scheduler does not inspect verdicts and cannot invent reachability.

The immutable bundle includes experiment/protocol, stage, corpus, model/candidate, scale/replicate, seed, input/dependency hashes, code/config hashes, output schema, declared DAG parents and the frozen execution descriptor. Worker, host, lease, attempts and timestamps exist only in telemetry.

The engineering `sha256-chain-v1` handler proves execution semantics for fully known M0–M5 labels. It is explicitly not a scientific M0–M5 implementation and its outputs cannot support recovery. Future frozen scientific handlers bind to the same bundle interface without changing scheduler, JobID, evidence, retry or aggregation semantics.
