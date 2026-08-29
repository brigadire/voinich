# Task85c-f design

This audit is blind-safe and non-confirmatory. It pins the three authoritative
parents by SHA-256, reproduces PF-SC01 before proposing any repair, compares
rejection with conditional sampling, audits every fitted and synthetic token
generation path, and tests the preferred conditional rule with two independent
implementations.

The success gate is fail-closed. A V1.2 artifact may be issued only if frozen
model state plus generation state plus G1V2-RNG-1 uniquely determines corpus
bytes. If the complete audit exposes independent scientific choices outside
the nonempty-token boundary, the required result is
`TASK85C_F_SCIENTIFIC_CONTRACT_DEFECT_EXPANDED`; no V1.2 files are emitted.

No escrow key, blind controls, production DAG, natural confirmatory data,
Voynich data, thresholds, or downstream performance are accessed.
