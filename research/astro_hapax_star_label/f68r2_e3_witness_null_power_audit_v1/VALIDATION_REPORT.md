# Validation report

- `GLOBAL_MAPPING_PROPAGATION=PASS` on the synthetic positive and small rejection fixtures.
- Full scorer replay is performed after every incremental extension and again at acceptance.
- Synthetic recovery is 20/20 at every listed budget.
- Historical S043 calibration assignment: `FAIL` under corrected global replay.
- Real 5-second rediscovery sweep: 0/20.
- Longer real sweeps: not run after the mandatory calibration gate failure.
- Old witness-null package was not modified.

The appropriate status is `CALIBRATION_GATE_FAILED`; the result is a power audit, not a null probability estimate.
