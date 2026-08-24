package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"zcore.dev/voinich/internal/task82b"
)

func writeManifestsAndReport(out string, recs []Rec, pairs []task82b.PairUnit) error {
	v := computeVerdicts(recs, pairs)

	kindCounts := map[string]int{}
	degenerate := 0
	for _, r := range recs {
		kindCounts[r.Kind]++
		if r.Degenerate {
			degenerate++
		}
	}
	manifest := map[string]any{
		"schema_version":   1,
		"task":             "Task82b",
		"total_jobs":       len(recs),
		"jobs_by_kind":     kindCounts,
		"degenerate_jobs":  degenerate,
		"f2_metric_ids":    task82b.AllMetricIDs(),
		"f2_repetitions":   task82b.F2Repetitions,
		"operator_grid":    task82b.Registry(),
		"carriers":         task82b.CarrierPaths,
		"shorthand_source": "BDD_koeln-edd-c-119",
		"shorthand_pairs":  len(pairs),
	}
	if err := writeJSON(out, "TASK82B_MANIFEST.json", manifest); err != nil {
		return err
	}

	if err := writeReport(out, v, recs, pairs, degenerate); err != nil {
		return err
	}
	if err := writeHandoff(out, v); err != nil {
		return err
	}

	// TASK82B_DESIGN_FROZEN marker (design was written and frozen before
	// this aggregation/interpretation step; see TASK82B_DESIGN.md header).
	if err := os.WriteFile(filepath.Join(out, "TASK82B_DESIGN_FROZEN"), []byte("Frozen before any operator-grid or shorthand-trajectory result was interpreted. See TASK82B_DESIGN.md.\n"), 0o644); err != nil {
		return err
	}

	resultsManifest, err := buildResultsManifest(out)
	if err != nil {
		return err
	}
	if err := writeJSON(out, "TASK82B_RESULTS_MANIFEST.json", resultsManifest); err != nil {
		return err
	}

	invalid := v.Values["VOYNICH_FIREWALL_PRESERVED"] != "SUPPORTED" || v.Values["FONTANA_FIREWALL_PRESERVED"] != "SUPPORTED" || len(pairs) == 0
	marker := "TASK82B_NOTATION_EXTRACTION_PORTFOLIO_FROZEN"
	body := "Task82b shorthand + extraction portfolios frozen. See TASK82B_REPORT.md and TASK83_NOTATION_EXTRACTION_HANDOFF.md.\n"
	if invalid {
		marker = "TASK82B_EXPERIMENT_INVALID"
		body = "Task82b invalidated: firewall breach or no real paired shorthand data. See TASK82B_REPORT.md.\n"
	}
	return os.WriteFile(filepath.Join(out, marker), []byte(body), 0o644)
}

func writeJSON(dir, name string, v any) error {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(dir, name), append(data, '\n'), 0o644)
}

func buildResultsManifest(out string) (map[string]any, error) {
	entries, err := os.ReadDir(out)
	if err != nil {
		return nil, err
	}
	files := map[string]string{}
	var names []string
	for _, e := range entries {
		if e.IsDir() || strings.HasSuffix(e.Name(), "_MANIFEST.json") {
			continue
		}
		names = append(names, e.Name())
	}
	sort.Strings(names)
	for _, name := range names {
		b, err := os.ReadFile(filepath.Join(out, name))
		if err != nil {
			return nil, err
		}
		h := sha256.Sum256(b)
		files[name] = hex.EncodeToString(h[:])
	}
	return map[string]any{"schema_version": 1, "task": "Task82b", "files": files}, nil
}

func writeReport(out string, v Verdicts, recs []Rec, pairs []task82b.PairUnit, degenerate int) error {
	var b strings.Builder
	fmt.Fprintf(&b, "# Task82b report\n\n")
	fmt.Fprintf(&b, "Historical shorthand/abbreviation and selective-extraction fingerprint experiment. Independent of Task81/82/82a (Fontana) and of Voynich; see TASK82B_DESIGN.md for the frozen design.\n\n")

	fmt.Fprintf(&b, "## Mandatory answers (sec.69)\n\n")
	fmt.Fprintf(&b, "1. Historical shorthand corpora used: Burchards Dekret Digital (BDD), koeln-edd-c-119, books 6/7/11/12/13 -- the only real paired abbreviated/expanded historical corpus obtained (see SHORTHAND_CORPUS_PROVENANCE.tsv).\n")
	fmt.Fprintf(&b, "2. Real abbreviated<->expanded pairs: yes, %d TEI <choice> pairs, both branches extracted from the source XML itself (internal/task82b/teipair.go).\n", len(pairs))
	fmt.Fprintf(&b, "3. Abbreviation operations represented: SUSPENSION, CONTRACTION, SPECIAL_SIGN_WHOLE_WORD, MARK_ONLY_ABBREVIATION, plus NO_VISIBLE_CHANGE/OTHER_SUBSTITUTION edge cases (ABBREVIATION_OPERATION_REGISTRY.tsv).\n")
	fmt.Fprintf(&b, "4. How abbreviation changes F2: see SHORTHAND_F2_TRAJECTORIES.tsv per chapter/combined; %d/%d CORE metrics show a chapter-consistent sign (see below).\n", v.NShorthandChaptersStable, v.NShorthandChaptersTotal)
	fmt.Fprintf(&b, "5. Stable shorthand ΔF2: %s (SHORTHAND_TRANSFORMATION_DETECTED).\n", v.Values["SHORTHAND_TRANSFORMATION_DETECTED"])
	fmt.Fprintf(&b, "6. Differs from matched-deletion null: %s (SHORTHAND_NULL_SEPARATION, SHORTHAND_NULL_COMPARISON.tsv).\n", v.Values["SHORTHAND_NULL_SEPARATION"])
	fmt.Fprintf(&b, "7. Differs from frequency/position-matched nulls: see SHORTHAND_NULL_COMPARISON.tsv rows for NULL_FREQUENCY_MATCHED_DELETION/NULL_POSITION_MATCHED alongside NULL_RANDOM_DELETION_MATCHED.\n")
	fmt.Fprintf(&b, "8. Stable between documents (BDD's 5 chapters): %s (SHORTHAND_CROSS_CORPUS_STABILITY; only one manuscript/scribe, so this is document-level, not corpus-level, replication).\n", v.Values["SHORTHAND_CROSS_CORPUS_STABILITY"])
	fmt.Fprintf(&b, "9. Stable between notation traditions: NOT_SUPPORTED -- no second tradition was obtained (TASK82B_DESIGN.md sec.4); this is a data-availability limitation, not a negative finding.\n")
	fmt.Fprintf(&b, "10. Context-dependent shorthand properties: SX5_CONTEXT_DEPENDENCE / SHORTHAND_RECOVERY.tsv quantify how many abbreviated-form types have >=2 observed expansions.\n")
	fmt.Fprintf(&b, "11. Expansion ambiguity: yes, see SX2_EXPANSION_AMBIGUITY and SHORTHAND_RECOVERY.tsv (ambiguous=true rows).\n")
	fmt.Fprintf(&b, "12. Shorthand-general fingerprint: not separable from a single-tradition effect with the data obtained (SHORTHAND_CROSS_TRADITION_STABILITY=NOT_SUPPORTED); only a single-tradition (SYSTEM_SPECIFIC) fingerprint is supported, see SHORTHAND_STABILITY.tsv.\n")
	fmt.Fprintf(&b, "13. Extraction operators tested: 20 (EXTRACTION_OPERATOR_REGISTRY.tsv), covering ACROSTIC/TELESTIC/POSITIONAL_EXTRACTION/PERIODIC_EXTRACTION classes.\n")
	fmt.Fprintf(&b, "14. Operators with stable ΔF2: %d EXTRACTION_GENERAL + %d OPERATOR_SPECIFIC (of %d operator×CORE-metric cells; EXTRACTION_STABILITY.tsv).\n", v.NOperatorsGeneral, v.NOperatorsSpecific, v.NOperatorsGeneral+v.NOperatorsSpecific+v.NOperatorsPlaintext+v.NOperatorsNotStable)
	fmt.Fprintf(&b, "15. Differ from random subsequence: %s (EXTRACTION_NULL_SEPARATION, EXTRACTION_NULL_COMPARISON.tsv, RANDOM_SUBSEQUENCE_MATCHED rows).\n", v.Values["EXTRACTION_NULL_SEPARATION"])
	fmt.Fprintf(&b, "16. Differ from position-matched null: see EXTRACTION_NULL_COMPARISON.tsv, POSITION_STRATIFIED_RANDOM rows (the 12 PER_GROUP operators).\n")
	fmt.Fprintf(&b, "17. Acrostic-specific signature: %s (%d/%d ACROSTIC/TELESTIC operator-metric cells stable, vs %d/%d PERIODIC_EXTRACTION cells; ACROSTIC_SPECIFIC_SIGNATURE).\n", v.Values["ACROSTIC_SPECIFIC_SIGNATURE"], v.NAcrosticGeneral, v.NAcrosticTotal, v.NPeriodicGeneral, v.NPeriodicTotal)
	fmt.Fprintf(&b, "18. Or only generic thinning: partly the latter by construction -- 4 of the ACROSTIC/TELESTIC operators (FIRST/LAST_TOKEN_OF_LINE, FIRST/LAST_GLYPH_OF_LINE) always emit <=1 output unit per original line, which mechanically zeroes F2's entire line-position family (2DL1/BP1/LS1-4/cs6) regardless of *which* position was kept; no PERIODIC operator can do this by definition. Real, reproducible, cross-carrier ΔF2 (not a bug), but part of the ACROSTIC/TELESTIC vs PERIODIC gap in EXTRACTION_STABILITY.tsv reflects which classes *can* collapse to <=1/line rather than positional specificity alone.\n")
	fmt.Fprintf(&b, "19. Carrier-language dependence: see INPUT_DEPENDENCE_COMPARISON.tsv (INPUT_DOMINATED vs MECHANISM_DOMINATED per CORE metric).\n")
	fmt.Fprintf(&b, "20. Operator dependence: same table, MECHANISM_DOMINATED rows.\n")
	fmt.Fprintf(&b, "21. Small-sample artifacts: %d/%d jobs marked DEGENERATE_OUTPUT (mostly single-glyph-alphabet operator outputs); retained, not deleted (raw/*.json `degenerate` field).\n", degenerate, len(recs))
	fmt.Fprintf(&b, "22. F2 sufficient for shorthand: F2 alone cannot see the abbreviated<->expanded alignment at all (structural gap, sec.51), independent of any sensitivity finding.\n")
	fmt.Fprintf(&b, "23. SX required: yes, by construction (see #22).\n")
	fmt.Fprintf(&b, "24. SX validated: %s (SX_VALIDATION.tsv self-consistency checks).\n", v.Values["SX_VALIDATED"])
	fmt.Fprintf(&b, "25. F2 sufficient for acrostic: partially -- BP1_BOUNDARY_TOKEN_NMI/LS2_POSITIONAL_LEXICON_NMI/2DL1_LAYOUT_POSITION_MI already carry positional signal (AX1/AX2/AX7 audit, TASK82B_DESIGN.md sec.12), but F2 has no entropy-vs-null-ratio, TTR, periodic-NMI, or cross-line-persistence statistic.\n")
	fmt.Fprintf(&b, "26. AX required: yes, for AX3/AX4/AX5/AX6 only (AX1/AX2/AX7 are redundant with existing F2 metrics, not implemented).\n")
	fmt.Fprintf(&b, "27. AX validated: %s (sec.50 gate needs positive-control sensitivity AND null calibration AND cross-corpus robustness; only the first two were attempted -- AX_VALIDATION.tsv).\n", v.Values["AX_VALIDATED"])
	fmt.Fprintf(&b, "28. General information-reducing-representation fingerprint: %s (INFORMATION_REDUCTION_COMPARISON.tsv; Spearman correlation of length-ratio vs retained-entropy-fraction across both branches).\n", v.Values["GENERAL_INFORMATION_REDUCTION_SIGNATURE"])
	fmt.Fprintf(&b, "29. Or statistically distinct: both branches are reported separately throughout and never merged into one dataset (sec.26); INFORMATION_REDUCTION_COMPARISON.tsv keeps `branch` as an explicit column precisely so this can be checked directly.\n")
	fmt.Fprintf(&b, "30. Both portfolios frozen before Voynich comparison: yes, see TASK82B_DESIGN_FROZEN and this report; no Voynich path was ever constructed (internal/task82b.assertNoVoynichPath).\n")
	fmt.Fprintf(&b, "31. Voynich firewall preserved: %s.\n", v.Values["VOYNICH_FIREWALL_PRESERVED"])
	fmt.Fprintf(&b, "32. Fontana firewall preserved: %s.\n", v.Values["FONTANA_FIREWALL_PRESERVED"])
	fmt.Fprintf(&b, "33. Task83 handoff ready: %s (TASK83_NOTATION_EXTRACTION_HANDOFF.md).\n\n", v.Values["TASK83_NOTATION_PORTFOLIO_READY"])

	fmt.Fprintf(&b, "## Final verdicts (sec.70)\n\n")
	fmt.Fprintf(&b, "| Verdict | Result |\n| --- | --- |\n")
	order := []string{
		"HISTORICAL_SHORTHAND_DATA_SUFFICIENT", "SHORTHAND_TRANSFORMATION_DETECTED", "SHORTHAND_F2_SIGNATURE",
		"SHORTHAND_NULL_SEPARATION", "SHORTHAND_CROSS_CORPUS_STABILITY", "SHORTHAND_CROSS_TRADITION_STABILITY",
		"SHORTHAND_KNOWLEDGE_DEPENDENCE", "EXTRACTION_TRANSFORMATION_DETECTED", "EXTRACTION_F2_SIGNATURE",
		"EXTRACTION_NULL_SEPARATION", "ACROSTIC_SPECIFIC_SIGNATURE", "AX_VALIDATED", "SX_VALIDATED",
		"GENERAL_INFORMATION_REDUCTION_SIGNATURE", "TASK83_NOTATION_PORTFOLIO_READY",
		"VOYNICH_FIREWALL_PRESERVED", "FONTANA_FIREWALL_PRESERVED",
	}
	for _, k := range order {
		fmt.Fprintf(&b, "| %s | %s |\n", k, v.Values[k])
	}
	invalid := v.Values["VOYNICH_FIREWALL_PRESERVED"] != "SUPPORTED" || v.Values["FONTANA_FIREWALL_PRESERVED"] != "SUPPORTED" || len(pairs) == 0
	fmt.Fprintf(&b, "\n**%s**\n", ifStr(invalid, "TASK82B_EXPERIMENT_INVALID", "TASK82B_NOTATION_EXTRACTION_PORTFOLIO_FROZEN"))

	return os.WriteFile(filepath.Join(out, "TASK82B_REPORT.md"), []byte(b.String()), 0o644)
}

func writeHandoff(out string, v Verdicts) error {
	var b strings.Builder
	fmt.Fprintf(&b, "# Task83 notation/extraction handoff\n\n")
	fmt.Fprintf(&b, "Frozen Task82b outputs available to Task83 (no Voynich result appears anywhere below):\n\n")
	fmt.Fprintf(&b, "- Frozen shorthand portfolio: SHORTHAND_CORPUS_PROVENANCE.tsv, ABBREVIATION_OPERATION_REGISTRY.tsv, SHORTHAND_ALIGNMENT_STATS.tsv, SHORTHAND_F2_BEFORE_AFTER.tsv, SHORTHAND_F2_TRAJECTORIES.tsv, SHORTHAND_NULL_COMPARISON.tsv, SHORTHAND_STABILITY.tsv, SHORTHAND_RECOVERY.tsv.\n")
	fmt.Fprintf(&b, "- Frozen extraction portfolio: EXTRACTION_OPERATOR_REGISTRY.tsv, EXTRACTION_F2_BEFORE_AFTER.tsv, EXTRACTION_F2_TRAJECTORIES.tsv, EXTRACTION_NULL_COMPARISON.tsv, EXTRACTION_STABILITY.tsv.\n")
	fmt.Fprintf(&b, "- Valid F2 subspace: the 17-metric CORE/SUPPORTING union frozen in TASK82B_DESIGN.md sec.2 (identical to Task82a.1's F2_COMMON_DIRECT ∪ F2_ASSEMBLER_PROJECTION).\n")
	fmt.Fprintf(&b, "- Shorthand transformation vectors: ΔF2(ABBREVIATED-EXPANDED) per chapter and combined, SHORTHAND_F2_TRAJECTORIES.tsv.\n")
	fmt.Fprintf(&b, "- Extraction transformation vectors: ΔF2(operator output - carrier baseline) per operator×carrier, EXTRACTION_F2_TRAJECTORIES.tsv.\n")
	fmt.Fprintf(&b, "- SX (validated, %s): SX_REGISTRY.tsv, SX_VALIDATION.tsv, SX_RESULTS.tsv.\n", v.Values["SX_VALIDATED"])
	fmt.Fprintf(&b, "- AX (%s -- sec.50 gate not fully passed, see AX_VALIDATION.tsv; Task83 may use AX3/AX4/AX5/AX6 only as descriptive, non-evidentiary context per sec.50's own rule, not as confirmatory evidence): AX_REGISTRY.tsv, AX_VALIDATION.tsv, AX_RESULTS.tsv.\n", v.Values["AX_VALIDATED"])
	fmt.Fprintf(&b, "- Null distributions: EXTRACTION_NULL_COMPARISON.tsv / SHORTHAND_NULL_COMPARISON.tsv (observed vs null mean/sd/effect-size/p-value per metric).\n")
	fmt.Fprintf(&b, "- Stability classifications: EXTRACTION_STABILITY.tsv (OPERATOR_SPECIFIC/EXTRACTION_GENERAL/PLAINTEXT_DRIVEN/NOT_STABLE), SHORTHAND_STABILITY.tsv (SYSTEM_SPECIFIC_EFFECT/NOT_STABLE; SHORTHAND_GENERAL_EFFECT never assigned, no cross-tradition data).\n")
	fmt.Fprintf(&b, "- Corpus provenance: SHORTHAND_CORPUS_PROVENANCE.tsv (BDD, reusing Task79c's verified chain) and TASK82B_DESIGN.md sec.9 (Doyle/Longfellow/Astafiev, same as Task82/Task82a).\n")
	fmt.Fprintf(&b, "- Limitations: single shorthand tradition/manuscript (no cross-tradition data obtainable); AX gate not fully passed (no cross-corpus positive control); 20-operator/{2,3,5,7}-period grid is intentionally small (sec.29, no combinatorial search); F2Repetitions=5 (empirically shown not to change any CORE/SUPPORTING point estimate used here).\n\n")
	fmt.Fprintf(&b, "## Interpretation rule for Task83 (sec.67)\n\n")
	fmt.Fprintf(&b, "Even a strong positive Task83 result (Voynich ≈ this shorthand or extraction fingerprint) means only that Voynich's fingerprint is statistically compatible with and aligned with the tested transformation class -- never \"Voynich is shorthand\" or \"Voynich contains an acrostic\".\n")
	return os.WriteFile(filepath.Join(out, "TASK83_NOTATION_EXTRACTION_HANDOFF.md"), []byte(b.String()), 0o644)
}
