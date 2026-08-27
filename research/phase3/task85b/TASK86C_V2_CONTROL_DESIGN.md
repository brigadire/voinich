# Task86C-v2 blind control design

Task86C-v2 validates G1-v2; it does not analyze Voynich. Stage A tests schemas, corruption detection, resume, duplicate conflict, representative cross-node M0–M5 equivalence, and 1/2/4-worker scaling. Stage B runs only open development controls, freezes derived thresholds, validates negative-space coverage and known-correct end-to-end survival, then freezes the final code/config/registry hash closure. Failure returns to a new design task; it cannot tune confirmation.

Stage C analyzes new opaque synthetic corpora from at least two independently authored generators per M0–M5 class at 2k/8k/32k tokens and four replicates. Generator definition, theoretical/minimal class, parameters, seed, and implementation hash are escrowed. Analysis sees only opaque corpus IDs. These are not the Task86C populations, parameters, or seeds. Stage D analyzes fixed English, Latin, and Sanskrit sources with provenance and licenses. Decode UTF-8, NFC normalize, lowercase Unicode letter runs, retain apostrophe only when internal, make no stemming/grammar-specific changes, take seeded contiguous occurrence windows, and split 60/20/20. Sampling matches synthetic scales.

The measurement contract, candidate grid, thresholds, code, seeds, recovery criteria, and analysis manifest freeze before Stage C/D inputs are opened. Analysts cannot read escrow mapping. Stage-E unblinding requires a frozen analysis-results hash and independent authorization; then mapping and generator hashes are disclosed. Stage F applies `G1V2_RECOVERY_CRITERIA.tsv` mechanically.

Pass requires all of:

- `MODEL_RECOVERY_VALIDATED`: every RC_EXACT through RC_STABILITY criterion passes;
- `NATURAL_LANGUAGE_APPLICABILITY_VALIDATED`: RC_NATURAL passes for all three languages; NONE is not a meaningful ordinary-language result here;
- `EVIDENCE_CHAIN_VALIDATED`: all paths regenerate byte-identically and all negative corruption tests fail closed;
- `DISTRIBUTED_EXECUTION_VALIDATED`: cross-node representatives pass, no unresolved conflict exists, manifest verifier passes, resume/failure suite passes, and scalability thresholds pass.

Synthetic pass alone is insufficient. Any failed condition yields `G1-v2 remains unvalidated`; no Voynich task may start.
