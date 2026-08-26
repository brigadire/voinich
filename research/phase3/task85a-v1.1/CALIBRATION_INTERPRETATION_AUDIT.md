# Calibration interpretation audit

The workload is unambiguous: `3 generators * 16 populations * 84 candidates =
4032` fit jobs, each population has 20000/5000/5000 tokens, and structural
distance compares a candidate generation to that population's HELDOUT analogue.
One generation at HELDOUT size per job follows the recorded workload.

The seed-variation quantity is different. The frozen calibration prose calls it
replicate-minus-replicate-median variation. Task86R performs one stochastic
generation for each independently generated and independently fitted
population, then treats the 16 resulting population-heldout metrics as the
replicate axis. This combines training-population, HELDOUT-population, and RNG
variation and is not a replicate experiment on a fixed fitted population.
No same-fit replicate outputs exist, so the correct counterfactual threshold
cannot be reconstructed without forbidden new generation.

Additionally, the contract states that nonfinite input invalidates calibration.
Task86R drops failed/nonfinite jobs and declares the remaining 948 thresholds
valid; there are no PM6 calibration threshold rows and no M3/M4 thresholds.
That is a separate undisclosed change, not merely an R6 interpretation.

Existing thresholds do implement nearest-rank q0.95 per generator and maximum
across MFC0/1/2 for every complete 16-value row. Alternative aggregation of the
same rows was therefore not a contract-consistent alternative; the sensitivity
lies in the missing true replicate axis and missing invalid rows.

`R6_CALIBRATION = SCIENTIFIC_CHANGE`.

