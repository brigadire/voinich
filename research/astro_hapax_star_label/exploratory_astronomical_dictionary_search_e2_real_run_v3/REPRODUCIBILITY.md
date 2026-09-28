# Reproducibility record

The package records hashes for the solver, scorer, configuration, frozen inputs, and all generated outputs in `SHA256SUMS`. `SNAPSHOT_MANIFEST.json` records the Python and OR-Tools versions. The solver canonicalizes source and target alphabets before assigning indices, and checkpoint payloads contain full incumbent tables and assignments.

To reproduce the run, use the locked environment and execute `real_run.py` from this directory. The run is single-worker and uses the fixed 900-second budget. Do not reuse the legacy v2 outputs or alter the sealed inputs after execution.
