# Task85c-a expanded contract defect

## Classification

`CONTRACT_DEFECT_EXPANDED` at the mandatory evidence-type × status matrix construction step. The original A17 schema permissiveness is reproducible, but the frozen normative sources do not define one consistent status/reachability space that schemas can enforce without changing scientific semantics.

## Findings

### E01 — unregistered reachability status

`G1V2_REACHABILITY_CONTRACT.tsv` uses upstream status `SCIENTIFIC_FAILURE` for FIT, PREDICTIVE, GENERATION, and STRUCTURAL. `G1V2_STATUS_REGISTRY.tsv` does not register `SCIENTIFIC_FAILURE`; it registers five distinct failures: `FIT_FAILURE`, `NUMERICAL_FAILURE`, `INDUCTION_CAP`, `GENERATION_FAILURE`, and `PROTOCOL_VETO`.

Consequently a schema author must either invent `SCIENTIFIC_FAILURE` as a status/alias or ignore normative reachability rows. Both are forbidden scientific choices.

### E02 — registered failures have no reachability rows

The normative reachability table has no rows for any of the five registered failure statuses. In particular, it cannot determine downstream evidence for `FIT_FAILURE`, `NUMERICAL_FAILURE`, `INDUCTION_CAP`, `GENERATION_FAILURE`, or `PROTOCOL_VETO`, although Task85c requires total reachability and status-specific downstream behavior.

### E03 — generation failure action conflicts

The status registry freezes `GENERATION_FAILURE` downstream behavior as `structural NOT_REACHED`. The only apparent generic failure surrogate in the reachability table is `GENERATION / SCIENTIFIC_FAILURE`, which orders `STRUCTURAL / RUN`. Treating the unregistered generic value as an alias therefore produces the opposite action.

### E04 — FIT negative-state contradiction

The reachability table contains `FIT / FAIL` and `FIT / NOT_ASSESSABLE`. The status registry restricts `FAIL` to `gate/verifier` with a finite statistic and threshold, while the executable contract states that fit, induction, and numerical failures are never negative class evidence. The registered FIT-stage states are instead `FIT_FAILURE`, `NUMERICAL_FAILURE`, and `INDUCTION_CAP`, none of which has a reachability row.

## Smallest distinct interpretations

For a generation job that exhausts the frozen 64-glyph cap:

1. Status-registry implementation emits `GENERATION_FAILURE` and makes structural jobs `NOT_REACHED`.
2. Reachability-table implementation maps it to `SCIENTIFIC_FAILURE` and runs structural jobs.

These yield different evidence, dependency behavior, and potentially `NOT_IDENTIFIABLE` reconstruction. Standard JSON Schema cannot resolve the contradiction; it can only enforce whichever matrix Task85c-a selects.

## Stop-rule consequence

Task85c-a section 34 says to stop with `TASK85C_A_CONTRACT_DEFECT_EXPANDED` if reachability contradicts normative status semantics. Repairing E01–E04 would require a controlled status/reachability contract revision beyond A17 enforcement. Task85c-a therefore did not create V1_1 schemas or a refrozen V1_1 contract.

