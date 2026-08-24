# Task83 notation/extraction handoff

Frozen Task82b outputs available to Task83 (no Voynich result appears anywhere below):

- Frozen shorthand portfolio: SHORTHAND_CORPUS_PROVENANCE.tsv, ABBREVIATION_OPERATION_REGISTRY.tsv, SHORTHAND_ALIGNMENT_STATS.tsv, SHORTHAND_F2_BEFORE_AFTER.tsv, SHORTHAND_F2_TRAJECTORIES.tsv, SHORTHAND_NULL_COMPARISON.tsv, SHORTHAND_STABILITY.tsv, SHORTHAND_RECOVERY.tsv.
- Frozen extraction portfolio: EXTRACTION_OPERATOR_REGISTRY.tsv, EXTRACTION_F2_BEFORE_AFTER.tsv, EXTRACTION_F2_TRAJECTORIES.tsv, EXTRACTION_NULL_COMPARISON.tsv, EXTRACTION_STABILITY.tsv.
- Valid F2 subspace: the 17-metric CORE/SUPPORTING union frozen in TASK82B_DESIGN.md sec.2 (identical to Task82a.1's F2_COMMON_DIRECT ∪ F2_ASSEMBLER_PROJECTION).
- Shorthand transformation vectors: ΔF2(ABBREVIATED-EXPANDED) per chapter and combined, SHORTHAND_F2_TRAJECTORIES.tsv.
- Extraction transformation vectors: ΔF2(operator output - carrier baseline) per operator×carrier, EXTRACTION_F2_TRAJECTORIES.tsv.
- SX (validated, SUPPORTED): SX_REGISTRY.tsv, SX_VALIDATION.tsv, SX_RESULTS.tsv.
- AX (NOT_SUPPORTED -- sec.50 gate not fully passed, see AX_VALIDATION.tsv; Task83 may use AX3/AX4/AX5/AX6 only as descriptive, non-evidentiary context per sec.50's own rule, not as confirmatory evidence): AX_REGISTRY.tsv, AX_VALIDATION.tsv, AX_RESULTS.tsv.
- Null distributions: EXTRACTION_NULL_COMPARISON.tsv / SHORTHAND_NULL_COMPARISON.tsv (observed vs null mean/sd/effect-size/p-value per metric).
- Stability classifications: EXTRACTION_STABILITY.tsv (OPERATOR_SPECIFIC/EXTRACTION_GENERAL/PLAINTEXT_DRIVEN/NOT_STABLE), SHORTHAND_STABILITY.tsv (SYSTEM_SPECIFIC_EFFECT/NOT_STABLE; SHORTHAND_GENERAL_EFFECT never assigned, no cross-tradition data).
- Corpus provenance: SHORTHAND_CORPUS_PROVENANCE.tsv (BDD, reusing Task79c's verified chain) and TASK82B_DESIGN.md sec.9 (Doyle/Longfellow/Astafiev, same as Task82/Task82a).
- Limitations: single shorthand tradition/manuscript (no cross-tradition data obtainable); AX gate not fully passed (no cross-corpus positive control); 20-operator/{2,3,5,7}-period grid is intentionally small (sec.29, no combinatorial search); F2Repetitions=5 (empirically shown not to change any CORE/SUPPORTING point estimate used here).

## Interpretation rule for Task83 (sec.67)

Even a strong positive Task83 result (Voynich ≈ this shorthand or extraction fingerprint) means only that Voynich's fingerprint is statistically compatible with and aligned with the tested transformation class -- never "Voynich is shorthand" or "Voynich contains an acrostic".
