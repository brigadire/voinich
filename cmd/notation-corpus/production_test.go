package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestBlockedProductionArtifactsStayUnauthorizedAndContainNoResults(t *testing.T) {
	repo := t.TempDir()
	s := preflightSnapshot{
		RunID:      productionRunID,
		Valid:      false,
		Gates:      []preflightGate{{ID: "production_candidates", Passed: false, Detail: "missing inputs"}},
		Candidates: []candidateSnapshot{{ClassID: "C01", ProductionReady: false, InputFiles: []string{}, MissingFields: []string{"USC_input"}}},
	}
	if err := writeBlockedProductionArtifacts(repo, s); err != nil {
		t.Fatal(err)
	}
	base := filepath.Join(repo, "research", "comparative_notation")
	for _, rel := range []string{
		"PRODUCTION_CANDIDATE_MANIFEST.json",
		"PRODUCTION_RUN_MANIFEST.json",
		"PRODUCTION_PREFLIGHT_REPORT.md",
		"PRODUCTION_RUN_AUTHORIZATION.json",
		"PRODUCTION_COMPARATIVE_RUN_REPORT.md",
		"PRODUCTION_COMPARATIVE_RUN_MANIFEST.json",
		"production_run/SHA256SUMS",
	} {
		if _, err := os.Stat(filepath.Join(base, filepath.FromSlash(rel))); err != nil {
			t.Fatalf("missing %s: %v", rel, err)
		}
	}
	b, err := os.ReadFile(filepath.Join(base, "PRODUCTION_RUN_AUTHORIZATION.json"))
	if err != nil {
		t.Fatal(err)
	}
	var auth struct {
		Authorized bool `json:"PRODUCTION_COMPARATIVE_RUN_AUTHORIZED"`
	}
	if err := json.Unmarshal(b, &auth); err != nil {
		t.Fatal(err)
	}
	if auth.Authorized {
		t.Fatal("a failed preflight must never emit positive authorization")
	}
	var manifest struct {
		Status      string   `json:"status"`
		ResultFiles []string `json:"result_files"`
	}
	b, err = os.ReadFile(filepath.Join(base, "PRODUCTION_COMPARATIVE_RUN_MANIFEST.json"))
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(b, &manifest); err != nil {
		t.Fatal(err)
	}
	if manifest.Status != "NOT_STARTED" || len(manifest.ResultFiles) != 0 {
		t.Fatalf("blocked run was partially materialized: %+v", manifest)
	}
}

func TestMandatoryFreezeBindingsIncludeProductionContract(t *testing.T) {
	want := map[string]bool{
		"GLOBAL_FREEZE_REPORT.md":          false,
		"CALIBRATION_PANEL_SPEC.md":        false,
		"CALIBRATION_PANEL_REPORT.md":      false,
		"CALIBRATION_SCALES.tsv":           false,
		"RAREFACTION_PROTOCOL.md":          false,
		"RAREFACTION_SCHEMA.md":            false,
		"DISTRIBUTION_OUTPUT_CONTRACT.md":  false,
		"BOOTSTRAP_PROTOCOL.md":            false,
		"VM_REFERENCE_V2.tsv":              false,
		"VM_REFERENCE_V2.fingerprint.json": false,
		"VM_REFERENCE_V2_MANIFEST.json":    false,
		"VM_REFERENCE_RECONCILIATION.md":   false,
	}
	for _, name := range requiredFrozenArtifacts {
		if _, ok := want[name]; ok {
			want[name] = true
		}
	}
	for name, present := range want {
		if !present {
			t.Errorf("mandatory frozen artifact %s is not checked", name)
		}
	}
}
