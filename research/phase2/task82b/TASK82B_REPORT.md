# Task82b report

Historical shorthand/abbreviation and selective-extraction fingerprint experiment. Independent of Task81/82/82a (Fontana) and of Voynich; see TASK82B_DESIGN.md for the frozen design.

## Mandatory answers (sec.69)

1. Historical shorthand corpora used: Burchards Dekret Digital (BDD), koeln-edd-c-119, books 6/7/11/12/13 -- the only real paired abbreviated/expanded historical corpus obtained (see SHORTHAND_CORPUS_PROVENANCE.tsv).
2. Real abbreviated<->expanded pairs: yes, 7150 TEI <choice> pairs, both branches extracted from the source XML itself (internal/task82b/teipair.go).
3. Abbreviation operations represented: SUSPENSION, CONTRACTION, SPECIAL_SIGN_WHOLE_WORD, MARK_ONLY_ABBREVIATION, plus NO_VISIBLE_CHANGE/OTHER_SUBSTITUTION edge cases (ABBREVIATION_OPERATION_REGISTRY.tsv).
4. How abbreviation changes F2: see SHORTHAND_F2_TRAJECTORIES.tsv per chapter/combined; 1/7 CORE metrics show a chapter-consistent sign (see below).
5. Stable shorthand ΔF2: PARTIAL (SHORTHAND_TRANSFORMATION_DETECTED).
6. Differs from matched-deletion null: SUPPORTED (SHORTHAND_NULL_SEPARATION, SHORTHAND_NULL_COMPARISON.tsv).
7. Differs from frequency/position-matched nulls: see SHORTHAND_NULL_COMPARISON.tsv rows for NULL_FREQUENCY_MATCHED_DELETION/NULL_POSITION_MATCHED alongside NULL_RANDOM_DELETION_MATCHED.
8. Stable between documents (BDD's 5 chapters): PARTIAL (SHORTHAND_CROSS_CORPUS_STABILITY; only one manuscript/scribe, so this is document-level, not corpus-level, replication).
9. Stable between notation traditions: NOT_SUPPORTED -- no second tradition was obtained (TASK82B_DESIGN.md sec.4); this is a data-availability limitation, not a negative finding.
10. Context-dependent shorthand properties: SX5_CONTEXT_DEPENDENCE / SHORTHAND_RECOVERY.tsv quantify how many abbreviated-form types have >=2 observed expansions.
11. Expansion ambiguity: yes, see SX2_EXPANSION_AMBIGUITY and SHORTHAND_RECOVERY.tsv (ambiguous=true rows).
12. Shorthand-general fingerprint: not separable from a single-tradition effect with the data obtained (SHORTHAND_CROSS_TRADITION_STABILITY=NOT_SUPPORTED); only a single-tradition (SYSTEM_SPECIFIC) fingerprint is supported, see SHORTHAND_STABILITY.tsv.
13. Extraction operators tested: 20 (EXTRACTION_OPERATOR_REGISTRY.tsv), covering ACROSTIC/TELESTIC/POSITIONAL_EXTRACTION/PERIODIC_EXTRACTION classes.
14. Operators with stable ΔF2: 21 EXTRACTION_GENERAL + 20 OPERATOR_SPECIFIC (of 140 operator×CORE-metric cells; EXTRACTION_STABILITY.tsv).
15. Differ from random subsequence: SUPPORTED (EXTRACTION_NULL_SEPARATION, EXTRACTION_NULL_COMPARISON.tsv, RANDOM_SUBSEQUENCE_MATCHED rows).
16. Differ from position-matched null: see EXTRACTION_NULL_COMPARISON.tsv, POSITION_STRATIFIED_RANDOM rows (the 12 PER_GROUP operators).
17. Acrostic-specific signature: SUPPORTED (15/42 ACROSTIC/TELESTIC operator-metric cells stable, vs 9/56 PERIODIC_EXTRACTION cells; ACROSTIC_SPECIFIC_SIGNATURE).
18. Or only generic thinning: partly the latter by construction -- 4 of the ACROSTIC/TELESTIC operators (FIRST/LAST_TOKEN_OF_LINE, FIRST/LAST_GLYPH_OF_LINE) always emit <=1 output unit per original line, which mechanically zeroes F2's entire line-position family (2DL1/BP1/LS1-4/cs6) regardless of *which* position was kept; no PERIODIC operator can do this by definition. Real, reproducible, cross-carrier ΔF2 (not a bug), but part of the ACROSTIC/TELESTIC vs PERIODIC gap in EXTRACTION_STABILITY.tsv reflects which classes *can* collapse to <=1/line rather than positional specificity alone.
19. Carrier-language dependence: see INPUT_DEPENDENCE_COMPARISON.tsv (INPUT_DOMINATED vs MECHANISM_DOMINATED per CORE metric).
20. Operator dependence: same table, MECHANISM_DOMINATED rows.
21. Small-sample artifacts: 35/495 jobs marked DEGENERATE_OUTPUT (mostly single-glyph-alphabet operator outputs); retained, not deleted (raw/*.json `degenerate` field).
22. F2 sufficient for shorthand: F2 alone cannot see the abbreviated<->expanded alignment at all (structural gap, sec.51), independent of any sensitivity finding.
23. SX required: yes, by construction (see #22).
24. SX validated: SUPPORTED (SX_VALIDATION.tsv self-consistency checks).
25. F2 sufficient for acrostic: partially -- BP1_BOUNDARY_TOKEN_NMI/LS2_POSITIONAL_LEXICON_NMI/2DL1_LAYOUT_POSITION_MI already carry positional signal (AX1/AX2/AX7 audit, TASK82B_DESIGN.md sec.12), but F2 has no entropy-vs-null-ratio, TTR, periodic-NMI, or cross-line-persistence statistic.
26. AX required: yes, for AX3/AX4/AX5/AX6 only (AX1/AX2/AX7 are redundant with existing F2 metrics, not implemented).
27. AX validated: NOT_SUPPORTED (sec.50 gate needs positive-control sensitivity AND null calibration AND cross-corpus robustness; only the first two were attempted -- AX_VALIDATION.tsv).
28. General information-reducing-representation fingerprint: NOT_SUPPORTED (INFORMATION_REDUCTION_COMPARISON.tsv; Spearman correlation of length-ratio vs retained-entropy-fraction across both branches).
29. Or statistically distinct: both branches are reported separately throughout and never merged into one dataset (sec.26); INFORMATION_REDUCTION_COMPARISON.tsv keeps `branch` as an explicit column precisely so this can be checked directly.
30. Both portfolios frozen before Voynich comparison: yes, see TASK82B_DESIGN_FROZEN and this report; no Voynich path was ever constructed (internal/task82b.assertNoVoynichPath).
31. Voynich firewall preserved: SUPPORTED.
32. Fontana firewall preserved: SUPPORTED.
33. Task83 handoff ready: SUPPORTED (TASK83_NOTATION_EXTRACTION_HANDOFF.md).

## Final verdicts (sec.70)

| Verdict | Result |
| --- | --- |
| HISTORICAL_SHORTHAND_DATA_SUFFICIENT | SUPPORTED |
| SHORTHAND_TRANSFORMATION_DETECTED | PARTIAL |
| SHORTHAND_F2_SIGNATURE | PARTIAL |
| SHORTHAND_NULL_SEPARATION | SUPPORTED |
| SHORTHAND_CROSS_CORPUS_STABILITY | PARTIAL |
| SHORTHAND_CROSS_TRADITION_STABILITY | NOT_SUPPORTED |
| SHORTHAND_KNOWLEDGE_DEPENDENCE | SUPPORTED |
| EXTRACTION_TRANSFORMATION_DETECTED | PARTIAL |
| EXTRACTION_F2_SIGNATURE | PARTIAL |
| EXTRACTION_NULL_SEPARATION | SUPPORTED |
| ACROSTIC_SPECIFIC_SIGNATURE | SUPPORTED |
| AX_VALIDATED | NOT_SUPPORTED |
| SX_VALIDATED | SUPPORTED |
| GENERAL_INFORMATION_REDUCTION_SIGNATURE | NOT_SUPPORTED |
| TASK83_NOTATION_PORTFOLIO_READY | SUPPORTED |
| VOYNICH_FIREWALL_PRESERVED | SUPPORTED |
| FONTANA_FIREWALL_PRESERVED | SUPPORTED |

**TASK82B_NOTATION_EXTRACTION_PORTFOLIO_FROZEN**
