# Validation report

Status: `PASS_WITH_PREREGISTERED_GATE_STOP`.

- PASS: upstream legacy, augmented-reference, preparation, plan, and occurrence checksums.
- PASS: legacy counts 7/7, 10/12, 53/67, 54, and 77 reproduce exactly.
- PASS: target LABEL counts are 37/33/22 and all 92 IDs are unique.
- PASS: every migrated legacy record and token occurrence exists in frozen inputs.
- PASS: no cross-panel match and no automatic many-to-many resolution.
- PASS: unmatched and sensitivity-only records are excluded from primary mapping.
- PASS: all human-added LABEL remain without invented legacy provenance.
- PASS: the eight-member group is one LABEL/group observation.
- PASS: original frozen plan SHA-256 remains `02e1f553dd77a60e40b8ae3cae3b3aa50abe0fef50997455ce463fc63d246e7a`.
- PASS: amendment SHA-256 is `9f413ff52b5ba065a2b896e2096c008688400072bfd1dc1a0aa8688ffa25763f` and predates the gate calculation.
- PASS: no corpus-local/page-local hapax definition was used; enrichment did not run.
- PASS: no dictionary/lexicon search ran.
- PASS: deterministic visual sheets include all ambiguous, all human-added, all f68r1 provenance rows, and a fixed primary sample.
- EXPECTED STOP: mapped ungrouped = 0; mapped panels = 1, so enrichment is not authorized.
