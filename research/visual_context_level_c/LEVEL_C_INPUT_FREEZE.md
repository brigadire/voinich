# Level-C input freeze

Inputs were frozen before interpreting Level-C output. The experiment uses the existing page-level fingerprint table and broad visual taxonomy, the frozen visual descriptor schema/protocol/manifest, and neutral metadata already present in those datasets. No schema or textual fingerprint was modified.

Design constants: seed `20260831`; deterministic ordering; 100 constrained within-section permutations with physical-leaf blocks preserved. Descriptor and textual-metric families were fixed before computation. Missing visual values are excluded per descriptor/test and never imputed.

The machine-readable hashes and statuses are recorded in `LEVEL_C_INPUT_MANIFEST.json` and `LEVEL_C_RESULTS_MANIFEST.json`.
