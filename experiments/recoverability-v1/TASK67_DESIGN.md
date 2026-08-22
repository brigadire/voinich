# Task67 design

Synthetic known-plaintext recoverability study. Corpora are split block-wise into TRAIN/VALIDATION/TEST; TEST is read only for final measurement. Task66 artifacts under experiments/mechanism-space-v1/ are authoritative. Voynich is never decoded, used for training, or used for candidate selection.

Corruption uses deterministic seeds and rates 0, 0.1%, 0.25%, 0.5%, 1%, 2%, 5%; stochastic jobs have 100 conceptual replicates in the report contract. Boundary, conflation, splitting, cascade, reset, oracle, and generator-control tables are explicit.
