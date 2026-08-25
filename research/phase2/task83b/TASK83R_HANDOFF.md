# Task83r handoff

- Authoritative target: `FINGERPRINT_V2.1_DETERMINISTIC_SCIENTIFIC_REFREEZE`.
- ZL3b target: `research/phase2/task83b/artifacts/zl/fingerprint.json`, SHA-256 `1661c21b080734f5313c3152814eee5b374b6565c5b3bad8c3c45aa3a9bfd847`.
- IT2a cross-transcription target: `research/phase2/task83b/artifacts/it/fingerprint.json`, SHA-256 `ab0de85fb01f4f7aef8acfdf5a8489facd4cb3ce73f9c07321812c7252df8bf4`.
- Metric registry: `research/phase2/task83b/F2_METRIC_REGISTRY_REFROZEN.tsv`, SHA-256 `7803afb24d5fd525eeeaae4aa6fec91263969aa49c85c92128adee9453389035`.
- Transitive manifest: `research/phase2/task83b/FINGERPRINT_V2_DETERMINISTIC_MANIFEST.json`, SHA-256 `a8c58ab9b42ab7cc62bcfb33a4dcabb1d86d4be9f495aea8251428f55bdfc512`.
- Verifier: `go run ./cmd/fingerprint-v2-verify -manifest research/phase2/task83b/FINGERPRINT_V2_DETERMINISTIC_MANIFEST.json -root .`.
- Expected verifier result: exit 0 and `Fingerprint V2 authoritative provenance verified`.
- Multi-run evidence: `MULTIRUN_REPRODUCIBILITY.tsv`; RUN_A/B/C are byte-identical under `GOMAXPROCS=1`, `2`, and default/NCPU.
- Downstream integrity: Task81, Task82, Task82a, Task82a.1, and Task82b are target-blind, valid, unchanged, and must not be retuned.
- Comparison contract: `research/phase2/task82a1/TASK83_COMPARISON_CONTRACT.md` remains valid; its scales and rules are pre-target.
- Ready status: `TASK83R_READY = SUPPORTED`.

Task83r must not read the old Task79/Task79c fingerprint as authoritative, the Task83 endpoint/trajectory/class rankings or quarantined distances, or any Task83 external-memory/shorthand/extraction evidence. The historical `3fb953…` checksum is metadata only and is explicitly non-authoritative. Task83 remains invalid; Task83r is a new confirmatory comparison against this refrozen target and the existing blind-frozen mechanism portfolios.
