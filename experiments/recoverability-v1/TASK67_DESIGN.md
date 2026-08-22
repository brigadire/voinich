# Task67 design

Synthetic known-plaintext recoverability study. Corpora are split block-wise into TRAIN/VALIDATION/TEST. Decoder tables are fitted on TRAIN+VALIDATION; final measurements use the first content-blind 128-word TEST block, with explicit 8/16/32/64/128-unit clean sub-blocks. Task66 artifacts under experiments/mechanism-space-v1/ are authoritative. Voynich is never decoded, used for training, or used for candidate selection.

Corruption uses deterministic seeds and rates 0, 0.1%, 0.25%, 0.5%, 1%, 2%, 5%. Every candidate/corpus/channel/rate cell has 100 executed raw replicates in ERROR_RECOVERABILITY.tsv. Conflation and splitting have 30 executed replicates per fraction. Single-error, boundary, cascade, reset, oracle, wrong-language, and generator controls all run encode -> transform ciphertext -> decode.
