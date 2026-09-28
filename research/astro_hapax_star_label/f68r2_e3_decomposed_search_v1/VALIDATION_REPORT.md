# Validation report

- Frozen source graph SHA-256 is recorded in `INPUT_FREEZE.json`.
- Source graph rows are split by profile without changing path contents.
- Exact duplicate pruning passed; no dominated-path pruning was applied.
- Profile globality gate: PASS.
- All 64 profile maxima: OPTIMAL.
- All 64 coverage-6 decision tests: infeasible certified.
- Best assignment scorer parity: PASS on every selected row.
- Canonical identities and LABEL occurrences are unique in the best assignment.
- Every active rule in the best assignment has distinct-EVA support at least two.
- Operation traces include normalization, profile operations, `DROP_UNMAPPED`, and final output.
- Checkpoint JSON exists for every profile.
- Null control was correctly skipped because no profile reached 6/27.

The result is exact only for the frozen 27-occurrence f68r2 graph and the 64 frozen operation profiles.
