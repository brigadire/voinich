# Level-C experiment protocol

Primary null: permute frozen visual descriptor vectors within broad visual section while preserving physical-leaf blocks. The primary multivariate diagnostic is the negative mean L2 distance between paired frozen visual vectors; descriptor-wise tests use binary mean differences, ordinal rank association, or categorical eta-squared according to descriptor type.

The ten pre-existing textual metric families are fixed: token length, type/token ratio, token entropy, exact and near-edit adjacency repetition, line-transition entropy, mean line tokens, line-length CV, boundary asymmetry, and line-token entropy. Benjamini–Hochberg correction is applied within the descriptor-wise family; the multivariate test is reported separately.

Partially annotated pages contribute only where a descriptor is observed. No imputation is performed. The run is diagnostic and uses a fixed 100-permutation budget; this is below the recommended 10,000 and therefore cannot support a positive Level-C claim. Confounder-controlled incremental models were not identifiable from the available frozen join and are explicitly marked incomplete.
