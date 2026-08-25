package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

func writeDevFits(fits []DevFitResult) error {
	w, err := NewTSVWriter(filepath.Join(outDir, "G1_MODEL_FITS.tsv"),
		[]string{"transcription", "model_class", "candidate_id", "failed", "failure_reason", "dev_pm2", "complexity_total", "structure_cost", "lexicon_cost", "exception_cost", "free_params", "states", "rules", "components"})
	if err != nil {
		return err
	}
	for _, f := range fits {
		if err := w.Row(f.Transcription, f.ModelClass, f.CandidateID, boolStr(f.Failed), f.FailureReason,
			f64(f.DevPM2), f64(f.Complexity.Total()), f64(f.Complexity.StructureCost), f64(f.Complexity.LexiconCost), f64(f.Complexity.ExceptionCost),
			i64(f.Complexity.FreeParams), i64(f.Complexity.States), i64(f.Complexity.Rules), i64(f.Complexity.Components)); err != nil {
			return err
		}
	}
	return w.Close()
}

func writeModelSelection(selByT map[string]StageCSelection, bitsRealByT map[string]float64, devByT, valByT map[string][]TokenOccurrence) error {
	w, err := NewTSVWriter(filepath.Join(outDir, "G1_MODEL_SELECTION.tsv"),
		[]string{"transcription", "model_class", "candidate_id", "parameters_json", "validation_pm1", "validation_pm2", "validation_pm4", "validation_pm5", "validation_pm6", "structure_cost", "lexicon_cost", "exception_cost", "complexity", "failure_status", "selected_for_confirmatory"})
	if err != nil {
		return err
	}
	classes := []string{"M0", "M1", "M2", "M3", "M4", "M5"}
	for _, tname := range []string{"ZL3b", "IT2a"} {
		sel := selByT[tname]
		val := valByT[tname]
		for _, class := range classes {
			sr := sel.ByClass[class]
			status := "OK"
			if sr.TrainingFailed {
				status = sr.FailureReason
				if err := w.Row(tname, class, sr.Candidate.CandidateID, paramsJSON(sr.Candidate), "NaN", "NaN", "NaN", "NaN", "NOT_APPLICABLE", "NaN", "NaN", "NaN", "NaN", status, "FALSE"); err != nil {
					return err
				}
				continue
			}
			pm := ComputePM1PM2PM3PM5(sr.Model, val)
			pm.PM4 = ComputePM4(sr.Model, val, devVocabulary(devByT[tname]))
			c := sr.Model.Complexity()
			if err := w.Row(tname, class, sr.Candidate.CandidateID, paramsJSON(sr.Candidate),
				f64(pm.PM1), f64(pm.PM2), f64(pm.PM4), f64(pm.PM5), "NOT_APPLICABLE",
				f64(c.StructureCost), f64(c.LexiconCost), f64(c.ExceptionCost), f64(c.Total()), status, "TRUE"); err != nil {
				return err
			}
		}
	}
	return w.Close()
}

func paramsJSON(c Candidate) string {
	b, _ := json.Marshal(c.Params)
	return string(b)
}

func writeSelectionFreeze(calibFreezeHash string, selByT map[string]StageCSelection) (string, error) {
	selHash, err := sha256Path(filepath.Join(outDir, "G1_MODEL_SELECTION.tsv"))
	if err != nil {
		return "", err
	}
	fitsHash, err := sha256Path(filepath.Join(outDir, "G1_MODEL_FITS.tsv"))
	if err != nil {
		return "", err
	}
	manifest := map[string]interface{}{
		"schema":                    "GRAMMAR_MODEL_SELECTION_FROZEN_MANIFEST_V1",
		"calibration_freeze_sha256": calibFreezeHash,
		"model_fits_sha256":         fitsHash,
		"model_selection_sha256":    selHash,
		"code_revision":             gitHeadShort(),
	}
	b, _ := json.MarshalIndent(manifest, "", "  ")
	if err := os.WriteFile(filepath.Join(outDir, "GRAMMAR_MODEL_SELECTION_FROZEN_MANIFEST.json"), append(b, '\n'), 0o644); err != nil {
		return "", err
	}
	content := fmt.Sprintf("GRAMMAR_MODEL_SELECTION_FROZEN\nversion=task86r-v1\ncalibration_freeze_sha256=%s\nmodel_fits_sha256=%s\nmodel_selection_sha256=%s\ncode_revision=%s\nheldout_opened_before_this_freeze=false\n",
		calibFreezeHash, fitsHash, selHash, gitHeadShort())
	if err := os.WriteFile(filepath.Join(outDir, "GRAMMAR_MODEL_SELECTION_FROZEN"), []byte(content), 0o644); err != nil {
		return "", err
	}
	return sha256Bytes([]byte(content)), nil
}
