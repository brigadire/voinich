# G1-v2 evidence preservation contract

Every scientific artifact is immutable, content-addressed SHA-256, canonical UTF-8, and connected to a producing `job_id`, sorted input hashes, executable/code hash, scientific-config hash, seed, schema version, and a plain-language scientific meaning. `G1V2_EVIDENCE_ARTIFACT_REGISTRY.tsv` is normative. The final manifest lists root verdict artifacts and all transitive nodes; an unreferenced intermediate cannot support a verdict.

The evidence graph is `input → fit/induction record → model artifact → PM records → per-PM gates → predictive verdict → generation batches → F2 metric records → family/scale/replicate gates → structural verdict → complexity comparisons → final verdict`. Fit and generation diagnostics remain results even on failure. Operational telemetry is stored beside, not inside, the scientific hash.

Atomic publication writes a temporary bundle, validates schema and hashes, then renames it under its hash. A result index maps `job_id` to one or more verified copies. Inputs, thresholds, registries, or manifests are read-only to workers. Duplicate identical copies provide reproducibility evidence; conflicting copies are all quarantined.

Retention is delete-free through freeze. For each scheduled downstream cell, the bundle contains either evidence or `NOT_REACHED` plus reason, upstream gate, and dependency hash. Missing is never an implicit status. The freeze verifier checks complete PM records, complete applicable F2 records, explicit reachability, exact code/config/seed closure, and all hashes before aggregation.

The Task86C-a CF1/CF2 tables remain engineering diagnostics only and are never imported as scientific classifications or threshold data.
