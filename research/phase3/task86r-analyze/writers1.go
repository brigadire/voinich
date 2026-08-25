package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

func writeCalibrationTables(res StageAResult) error {
	w, err := NewTSVWriter(filepath.Join(outDir, "G1_CALIBRATION_RESULTS.tsv"),
		[]string{"generator", "population", "model_class", "candidate_id", "failed", "failure_why", "dev_pm2", "val_pm1", "val_pm2", "val_pm4", "val_pm5", "held_pm1", "held_pm2", "held_pm4", "held_pm5", "pm6", "pm6_valid", "f2_valid"})
	if err != nil {
		return err
	}
	for _, j := range res.Jobs {
		if err := w.Row(j.Generator, i64(j.Population), j.ModelClass, j.CandidateID, boolStr(j.Failed), j.FailureWhy,
			f64(j.DevPM2), f64(j.ValPM.PM1), f64(j.ValPM.PM2), f64(j.ValPM.PM4), f64(j.ValPM.PM5),
			f64(j.HeldPM.PM1), f64(j.HeldPM.PM2), f64(j.HeldPM.PM4), f64(j.HeldPM.PM5),
			f64(j.PM6), boolStr(j.PM6Valid), boolStr(j.F2Valid)); err != nil {
			return err
		}
	}
	if err := w.Close(); err != nil {
		return err
	}

	w2, err := NewTSVWriter(filepath.Join(outDir, "G1_CALIBRATION_THRESHOLDS.tsv"),
		[]string{"quantity", "metric", "model_class", "candidate_id", "threshold", "mfc0", "mfc1", "mfc2"})
	if err != nil {
		return err
	}
	for _, t := range res.Thresholds {
		if err := w2.Row(t.Quantity, t.Metric, t.ModelClass, t.CandidateID, f64(t.Threshold),
			f64(t.PerGenerator["MFC0"]), f64(t.PerGenerator["MFC1"]), f64(t.PerGenerator["MFC2"])); err != nil {
			return err
		}
	}
	return w2.Close()
}

func sha256Path(path string) (string, error) { return sha256File(path) }

func sha256Bytes(b []byte) string {
	h := sha256.Sum256(b)
	return hex.EncodeToString(h[:])
}

func writeCalibrationFreeze(res StageAResult) (string, error) {
	resultsHash, err := sha256Path(filepath.Join(outDir, "G1_CALIBRATION_RESULTS.tsv"))
	if err != nil {
		return "", err
	}
	thresholdsHash, err := sha256Path(filepath.Join(outDir, "G1_CALIBRATION_THRESHOLDS.tsv"))
	if err != nil {
		return "", err
	}
	gitCommit := gitHeadShort()
	manifest := map[string]interface{}{
		"schema":                     "G1_CALIBRATION_MANIFEST_V1",
		"total_jobs":                 res.TotalJobs,
		"failed_jobs":                res.FailedJobs,
		"threshold_rows":             len(res.Thresholds),
		"calibration_results_sha256": resultsHash,
		"calibration_thresholds_sha256": thresholdsHash,
		"code_revision":              gitCommit,
	}
	b, _ := json.MarshalIndent(manifest, "", "  ")
	if err := os.WriteFile(filepath.Join(outDir, "G1_CALIBRATION_MANIFEST.json"), append(b, '\n'), 0o644); err != nil {
		return "", err
	}
	manifestHash := sha256Bytes(append(b, '\n'))

	freezeContent := fmt.Sprintf("G1_CALIBRATION_FROZEN\nversion=task86r-v1\ncode_revision=%s\ncalibration_results_sha256=%s\ncalibration_thresholds_sha256=%s\ncalibration_manifest_sha256=%s\ntotal_jobs=%d\nfailed_jobs=%d\n",
		gitCommit, resultsHash, thresholdsHash, manifestHash, res.TotalJobs, res.FailedJobs)
	if err := os.WriteFile(filepath.Join(outDir, "G1_CALIBRATION_FROZEN"), []byte(freezeContent), 0o644); err != nil {
		return "", err
	}
	return sha256Bytes([]byte(freezeContent)), nil
}

func gitHeadShort() string {
	b, err := os.ReadFile(".git/HEAD")
	if err != nil {
		return "unknown"
	}
	s := string(b)
	if len(s) > 5 && s[:5] == "ref: " {
		ref := s[5 : len(s)-1]
		b2, err := os.ReadFile(filepath.Join(".git", ref))
		if err == nil {
			return trimNL(string(b2))
		}
	}
	return trimNL(s)
}

func trimNL(s string) string {
	for len(s) > 0 && (s[len(s)-1] == '\n' || s[len(s)-1] == '\r') {
		s = s[:len(s)-1]
	}
	return s
}
