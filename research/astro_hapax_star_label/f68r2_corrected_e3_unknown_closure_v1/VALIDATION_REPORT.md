# Validation report

- Imported checkpoint verification: PASS (64/64).
- Frozen graph hashes and checkpoint manifests: PASS.
- Decision model on committed exact profiles: PASS; four `coverage >=4` queries were INFEASIBLE.
- Synthetic non-hereditary quadruple: PASS; coverage 4 was FEASIBLE with full replay.
- Required/forbidden, functional/injective mapping, LABEL/EVA/identity capacities, and distinct-EVA support are encoded in the decision model.
- Independent full replay of every accepted validation witness: PASS.
- Real E3 profiles rerun: 0; all 24 UNKNOWN profiles were closed by their committed upper bound ≤3.
- Null controls: not run.
