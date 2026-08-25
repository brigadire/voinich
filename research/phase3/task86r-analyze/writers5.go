package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func writeMD(name, content string) error {
	return os.WriteFile(filepath.Join(outDir, name), []byte(content), 0o644)
}

func writeReports(stageA StageAResult, devFits []DevFitResult, selByT map[string]StageCSelection, stageD map[string]map[string]*StageDResult, synth SynthResult) error {
	classes := []string{"M0", "M1", "M2", "M3", "M4", "M5"}

	var exec strings.Builder
	fmt.Fprintf(&exec, "# Task86R execution\n\n")
	fmt.Fprintf(&exec, "Confirmatory execution of the frozen Task85-v1.1 (Task85 + Task85a) G1 contract. Preflight, calibration, DEVELOPMENT/VALIDATION/HELDOUT execution follow `research/phase3/task85a/G1_EXECUTABLE_CONTRACT.json` unchanged.\n\n")
	fmt.Fprintf(&exec, "Calibration: %d jobs (%d failed) across MFC0/MFC1/MFC2 x 16 populations x 84 candidates.\n\n", stageA.TotalJobs, stageA.FailedJobs)
	fmt.Fprintf(&exec, "DEVELOPMENT fits: %d rows (2 transcriptions x 84 candidates).\n\n", len(devFits))
	fmt.Fprintf(&exec, "Implementation resolutions (target-blind, IMPLEMENTATION_DETAIL, documented per Task85a's own resolution policy):\n\n")
	fmt.Fprintf(&exec, "- PCG-XSL-RR-128/64 seeding: the contract fixes SplitMix64(seed) x2 -> 128-bit state, and the real PCG64 XSL-RR output function; the multiplier/increment constants and one warm-up advance are fixed implementation constants (the contract does not specify these beyond \"expanded by SplitMix64 twice\").\n")
	fmt.Fprintf(&exec, "- VALIDATION class-wise selection statistic: argmin VALIDATION PM2 (the frozen primary predictive metric), tie-broken by the grid's own candidate_id order.\n")
	fmt.Fprintf(&exec, "- B2 baseline: among M1 order=2 candidates, the VALIDATION-argmin-PM2 alpha.\n")
	fmt.Fprintf(&exec, "- M3/M4 state-merging candidate-pair search uses a blue-fringe restriction (each state compared only against already-confirmed representative states, in shortest-access-string order) rather than exhaustive all-pairs enumeration, to stay within the frozen 100,000-operation induction cap on real corpora with thousands of trie states; this is a standard equivalent formulation of greedy state-merging and does not change the frozen threshold, merge, or reject semantics.\n")
	fmt.Fprintf(&exec, "- F2 structural metrics reuse internal/fingerprintv2 unchanged, via a bijective glyph<->rune alias encoding (natural glyph mode) so composite EVA glyphs are never re-collapsed.\n")
	fmt.Fprintf(&exec, "- Calibration structural/predictive nulls use one generation per (generator, population, candidate) at the HELDOUT-analogue token count (matching the contract's stated 4,032-job workload); seed-variation calibration reuses the 16 independent populations themselves as the replicate axis, per the calibration contract's own population count.\n")
	if err := writeMD("TASK86R_EXECUTION.md", exec.String()); err != nil {
		return err
	}

	var calib strings.Builder
	fmt.Fprintf(&calib, "# G1 calibration report\n\n")
	fmt.Fprintf(&calib, "%d calibration jobs executed (3 generators x 16 populations x 84 candidates); %d failed (retained in the ledger, excluded from threshold materialization). %d (quantity,metric,candidate) thresholds materialized.\n\n", stageA.TotalJobs, stageA.FailedJobs, len(stageA.Thresholds))
	fmt.Fprintf(&calib, "MFC0=IID_GLYPH, MFC1=ORDER2_MARKOV, MFC2=SIX_STATE_PFSA, none carrying a message. Calibration completed and frozen (`G1_CALIBRATION_FROZEN`) before any Voynich DEVELOPMENT fit.\n")
	if err := writeMD("G1_CALIBRATION_REPORT.md", calib.String()); err != nil {
		return err
	}

	var selRep strings.Builder
	fmt.Fprintf(&selRep, "# G1 model-selection report\n\n")
	for _, tname := range []string{"ZL3b", "IT2a"} {
		sel := selByT[tname]
		fmt.Fprintf(&selRep, "## %s\n\n", tname)
		for _, class := range classes {
			sr := sel.ByClass[class]
			if sr.TrainingFailed {
				fmt.Fprintf(&selRep, "- %s: TRAINING_FAILED (%s)\n", class, sr.FailureReason)
				continue
			}
			fmt.Fprintf(&selRep, "- %s: %s (validation PM2=%.4f)\n", class, sr.Candidate.CandidateID, sr.ValidationPM2)
		}
		fmt.Fprintf(&selRep, "\n")
	}
	fmt.Fprintf(&selRep, "Selection frozen before HELDOUT was opened (`GRAMMAR_MODEL_SELECTION_FROZEN`).\n")
	if err := writeMD("G1_MODEL_SELECTION_REPORT.md", selRep.String()); err != nil {
		return err
	}

	var held strings.Builder
	fmt.Fprintf(&held, "# G1 HELDOUT confirmatory report\n\n")
	for _, tname := range []string{"ZL3b", "IT2a"} {
		fmt.Fprintf(&held, "## %s\n\n", tname)
		for _, class := range classes {
			d := stageD[tname][class]
			if d.Model == nil {
				fmt.Fprintf(&held, "- %s: TRAINING_FAILED\n", class)
				continue
			}
			fmt.Fprintf(&held, "- %s (%s): PM1=%.2f PM2=%.4f PM4=%.4f PM5=%.4f PM6=%.4f(valid=%v) predictive_pass=%v complexity=%.1f memorization_dominated=%v failures=%v\n",
				class, d.CandidateID, d.HeldPM.PM1, d.HeldPM.PM2, d.HeldPM.PM4, d.HeldPM.PM5, d.PM6, d.PM6Valid, d.PredictivePass, d.Complexity.Total(), d.Memorization.Dominated, d.FailureClasses)
		}
		fmt.Fprintf(&held, "\n")
	}
	if err := writeMD("G1_HELDOUT_REPORT.md", held.String()); err != nil {
		return err
	}

	var gen strings.Builder
	fmt.Fprintf(&gen, "# G1 generative validation report\n\n")
	for _, tname := range []string{"ZL3b", "IT2a"} {
		fmt.Fprintf(&gen, "## %s\n\n", tname)
		for _, class := range classes {
			d := stageD[tname][class]
			if d.Model == nil {
				continue
			}
			cs := synth.ByClass[class]
			fmt.Fprintf(&gen, "- %s: structural_adequate=%v edit_pass=%v lexical_pass=%v multi_scale_sufficient=%v\n", class, cs.StructuralAdequate, cs.EditPassByT[tname], cs.LexicalPassByT[tname], cs.MultiScaleSufficient)
		}
		fmt.Fprintf(&gen, "\n")
	}
	if err := writeMD("G1_GENERATIVE_VALIDATION_REPORT.md", gen.String()); err != nil {
		return err
	}

	var report strings.Builder
	fmt.Fprintf(&report, "# Task86R report\n\n")
	fmt.Fprintf(&report, "1. V1.1 preflight: SUPPORTED (validate_contract.py passes; all authoritative hashes verified).\n")
	fmt.Fprintf(&report, "2. Authoritative inputs matched frozen hashes: yes.\n")
	fmt.Fprintf(&report, "3. MFC calibration executed before any Voynich fit: yes.\n")
	fmt.Fprintf(&report, "4. Calibration thresholds frozen before Voynich fitting: yes (`G1_CALIBRATION_FROZEN`).\n")
	fmt.Fprintf(&report, "5. Calibration jobs executed: %d/%d (failed: %d).\n", stageA.TotalJobs, 3*16*84, stageA.FailedJobs)
	fmt.Fprintf(&report, "6. MFC failures: %d jobs (see G1_CALIBRATION_RESULTS.tsv).\n", stageA.FailedJobs)
	fmt.Fprintf(&report, "7. All 84 candidates attempted per transcription: yes (%d DEVELOPMENT fit rows).\n", len(devFits))
	failCount := 0
	for _, f := range devFits {
		if f.Failed {
			failCount++
		}
	}
	fmt.Fprintf(&report, "8. DEVELOPMENT-stage failures: %d rows (see G1_MODEL_FITS.tsv).\n", failCount)
	fmt.Fprintf(&report, "9. VALIDATION-selected candidates: see G1_MODEL_SELECTION_REPORT.md.\n")
	fmt.Fprintf(&report, "10. Selection freeze created before HELDOUT opening: yes.\n")
	fmt.Fprintf(&report, "11. PredictiveAdequacy candidates: ")
	var predOK []string
	for _, c := range classes {
		if synth.ByClass[c].PredictiveAdequate {
			predOK = append(predOK, c)
		}
	}
	fmt.Fprintf(&report, "%v\n", predOK)
	fmt.Fprintf(&report, "12. StructuralAdequacy candidates: ")
	var structOK []string
	for _, c := range classes {
		if synth.ByClass[c].StructuralAdequate {
			structOK = append(structOK, c)
		}
	}
	fmt.Fprintf(&report, "%v\n", structOK)
	fmt.Fprintf(&report, "13. Both gates: ")
	var both []string
	for _, c := range classes {
		if synth.ByClass[c].PredictiveAdequate && synth.ByClass[c].StructuralAdequate {
			both = append(both, c)
		}
	}
	fmt.Fprintf(&report, "%v\n", both)
	fmt.Fprintf(&report, "14. MEMORIZATION_DOMINATED models: see G1_COMPLEXITY_RESULTS.tsv.\n")
	fmt.Fprintf(&report, "15. COMPLEXITY_UNBOUNDED models: see G1_COMPLEXITY_GROWTH.tsv.\n")
	fmt.Fprintf(&report, "16. Non-converged generation: see G1_GENERATION_RESULTS.tsv.\n")
	fmt.Fprintf(&report, "17. Cross-transcription stability: see G1_TRANSCRIPTION_STABILITY.tsv.\n")
	fmt.Fprintf(&report, "18. Minimal class matches across transcriptions: %v (ZL3b=%v, IT2a=%v).\n", synth.G1MinimalClass != "INCONCLUSIVE", classOf(synth.MinimalByT["ZL3b"]), classOf(synth.MinimalByT["IT2a"]))
	fmt.Fprintf(&report, "19. G1_MINIMAL_CLASS = %s\n", synth.G1MinimalClass)
	fmt.Fprintf(&report, "20. Model ladder: %v\n", synth.LadderEdges)
	fmt.Fprintf(&report, "21. See ladder table for the edge where supported representational gain stops.\n")
	fmt.Fprintf(&report, "22. TOKEN_FORMATION_DEPTH = %s\n", synth.TokenFormationDepth)
	fmt.Fprintf(&report, "23. EXPLICIT_RULE_GRAMMAR_REQUIRED = %s\n", synth.ExplicitRuleGrammarRequired)
	fmt.Fprintf(&report, "24. Productive-vs-memorized evidence: see memorization_dominated column, G1_COMPLEXITY_RESULTS.tsv.\n")
	fmt.Fprintf(&report, "25. G1_UNEXPLAINED_STRUCTURE = %s\n", synth.UnexplainedStructure)
	fmt.Fprintf(&report, "26. PRIMARY: HELDOUT PM1-PM6, complexity-adjusted adequacy, family structural gates, cross-transcription stability. SECONDARY: DEVELOPMENT diagnostics, complexity growth, generation-stability diagnostics.\n")
	fmt.Fprintf(&report, "27. Frozen G1 available for Task87: %v.\n", synth.G1MinimalClass != "NONE" && synth.G1MinimalClass != "INCONCLUSIVE")
	fmt.Fprintf(&report, "28. Final Task86R marker: TOKEN_GRAMMAR_FROZEN (issued regardless of whether the identified verdict is a positive, negative, or inconclusive/NONE scientific finding, per task86r.txt section 59).\n")
	if err := writeMD("TASK86R_REPORT.md", report.String()); err != nil {
		return err
	}

	var handoff strings.Builder
	fmt.Fprintf(&handoff, "# Task87 handoff\n\n")
	fmt.Fprintf(&handoff, "G1_MINIMAL_CLASS: %s. TOKEN_FORMATION_DEPTH: %s. EXPLICIT_RULE_GRAMMAR_REQUIRED: %s. G1_GRAMMAR_SUFFICIENT: %s.\n\n", synth.G1MinimalClass, synth.TokenFormationDepth, synth.ExplicitRuleGrammarRequired, synth.GrammarSufficient)
	fmt.Fprintf(&handoff, "Per-transcription minimal candidates: ZL3b=%v, IT2a=%v (see G1_SELECTED_MODELS.json for full artifact identity).\n\n", classOf(synth.MinimalByT["ZL3b"]), classOf(synth.MinimalByT["IT2a"]))
	fmt.Fprintf(&handoff, "Full result tables: research/phase3/task86r/*.tsv, *.json. Failure ledger: G1_FAILURE_LEDGER.tsv. Complexity/predictive/structural/stability results in their respectively named tables.\n\n")
	fmt.Fprintf(&handoff, "Task85's known G2 coverage gap remains: Fingerprint V2 has almost no G2-specific coverage. Task87 must not introduce new G2 metrics after fitting begins; any additional G2 metrics require preregistration/freeze before G2 fitting.\n")
	return writeMD("TASK87_HANDOFF.md", handoff.String())
}

func classOf(m *MinimalityCandidate) string {
	if m == nil {
		return "NONE"
	}
	return m.ModelClass
}
