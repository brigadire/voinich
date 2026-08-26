# B2 baseline audit

The inherited `GRAMMAR_BASELINE_REGISTRY.tsv` defines B2 as a low-order M1/M2
instance (order 2 is an example) and says its order is selected on VALIDATION.
Task85a uses B2 in the gate but does not replace that construction with a
specific candidate or alpha objective.

Task86R instead restricts B2 to M1 `order == 2` and chooses alpha by VALIDATION
PM2. The exact B2 candidate ID and its validation values were not persisted.
Fixing order 2 differs from the inherited order-selection procedure; choosing
alpha by PM2 also fills an unspecified objective. A different B2 changes all
strict B2 comparisons and their calibrated nulls.

`R3_B2 = SCIENTIFIC_CHANGE`.

