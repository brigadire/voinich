package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

func writeExecutionLedger(stageA StageAResult, stageD map[string]map[string]*StageDResult) error {
	w, err := NewTSVWriter(filepath.Join(outDir, "TASK86R_EXECUTION_LEDGER.tsv"),
		[]string{"job_id", "stage", "model_class", "candidate_id", "corpus", "transcription", "partition", "status", "failure_class"})
	if err != nil {
		return err
	}
	id := 0
	next := func() string { id++; return fmt.Sprintf("job-%06d", id) }
	for _, j := range stageA.Jobs {
		status := "OK"
		fc := ""
		if j.Failed {
			status, fc = "FAILED", j.FailureWhy
		}
		if err := w.Row(next(), "CALIBRATION", j.ModelClass, j.CandidateID, j.Generator, fmt.Sprintf("P%02d", j.Population), "CALIBRATION", status, fc); err != nil {
			return err
		}
	}
	for _, tname := range []string{"ZL3b", "IT2a"} {
		for _, class := range []string{"M0", "M1", "M2", "M3", "M4", "M5"} {
			d := stageD[tname][class]
			status := "OK"
			fc := ""
			if d.Model == nil {
				status, fc = "FAILED", "TRAINING_FAILED"
			} else if len(d.FailureClasses) > 0 {
				status, fc = "DEGRADED", d.FailureClasses[0]
			}
			if err := w.Row(next(), "HELDOUT_CONFIRMATORY", class, d.CandidateID, "REAL", tname, "HELDOUT", status, fc); err != nil {
				return err
			}
		}
	}
	return w.Close()
}

func writeResultsManifest(calibFreezeHash, selectionFreezeHash string, synth SynthResult) error {
	ledgerHash, _ := sha256Path(filepath.Join(outDir, "TASK86R_EXECUTION_LEDGER.tsv"))
	manifest := map[string]interface{}{
		"schema":                    "TASK86R_RESULTS_MANIFEST_V1",
		"code_revision":             gitHeadShort(),
		"calibration_freeze_sha256": calibFreezeHash,
		"selection_freeze_sha256":   selectionFreezeHash,
		"execution_ledger_sha256":   ledgerHash,
		"g1_minimal_class":          synth.G1MinimalClass,
		"token_formation_depth":     synth.TokenFormationDepth,
		"explicit_rule_grammar_required": synth.ExplicitRuleGrammarRequired,
		"grammar_sufficient":        synth.GrammarSufficient,
	}
	b, _ := json.MarshalIndent(manifest, "", "  ")
	return os.WriteFile(filepath.Join(outDir, "TASK86R_RESULTS_MANIFEST.json"), append(b, '\n'), 0o644)
}

func writeFinalFreeze(synth SynthResult) (string, error) {
	marker := "TOKEN_GRAMMAR_FROZEN"
	content := fmt.Sprintf("%s\nversion=task86r-v1\ncode_revision=%s\ng1_minimal_class=%s\ntoken_formation_depth=%s\nexplicit_rule_grammar_required=%s\ngrammar_sufficient=%s\n",
		marker, gitHeadShort(), synth.G1MinimalClass, synth.TokenFormationDepth, synth.ExplicitRuleGrammarRequired, synth.GrammarSufficient)
	if err := os.WriteFile(filepath.Join(outDir, marker), []byte(content), 0o644); err != nil {
		return "", err
	}
	return marker, nil
}
