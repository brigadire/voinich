# Validation report

- Scope filter: `page == f68r2`; 27 occurrences.
- f68r1 used: no.
- Group crosswalk used: no.
- Operation profiles frozen before result inspection: 64.
- Individual operation semantics: frozen in `CORRECTED_MODEL_SEMANTICS.md`.
- Scorer: exact character-level injective mapping with explicit `DROP_UNMAPPED`.
- Path graph: generated and written; every retained path passed scorer replay.
- Exact optimization: incomplete; no scientific result.
- Null control: not run.

The package deliberately reports timeout rather than treating partial solver output as a result.
