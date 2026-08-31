# Level-C v2 reproducibility report

Two executions used identical frozen inputs, seed `20260831`, protocol, registry, and 10,000 evaluated positions. Run IDs differed only by immutable bundle identifier. After normalizing that identifier, every TSV scientific output was byte-identical.

`LEVEL_C_V2_REPRODUCIBLE=true`

- `LEVEL_C_V2_CONFOUNDER_DIAGNOSTICS.tsv`: identical=True
- `LEVEL_C_V2_DESCRIPTOR_ASSOCIATIONS.tsv`: identical=True
- `LEVEL_C_V2_FAMILY_ASSOCIATIONS.tsv`: identical=True
- `LEVEL_C_V2_NEGATIVE_CONTROLS.tsv`: identical=True
- `LEVEL_C_V2_OVERLAP_DIAGNOSTICS.tsv`: identical=True
- `LEVEL_C_V2_PERMUTATION_AUDIT.tsv`: identical=True
- `LEVEL_C_V2_PRIMARY_TEST.tsv`: identical=True
- `LEVEL_C_V2_RAW_PERMUTATION_SUMMARY.tsv`: identical=True
- `LEVEL_C_V2_SECTION_CONSISTENCY.tsv`: identical=True
- `LEVEL_C_V2_SENSITIVITY.tsv`: identical=True
