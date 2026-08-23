package main

import (
	"os"
	"path/filepath"
	"testing"
)

// 10. Voynich execution must be impossible without a METHOD_FROZEN marker
// recording a passed validation gate.
func TestVoynichRefusesWithoutMethodFrozen(t *testing.T) {
	dir := t.TempDir()
	if _, err := checkMethodFrozen(dir); err == nil {
		t.Fatal("expected an error when METHOD_FROZEN is absent")
	}
}

func TestVoynichRefusesWhenGateDidNotPass(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "METHOD_FROZEN"), []byte("marker\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "manifest.json"), []byte(`{"gate_pass": false, "config": {"Threshold": 0.3}}`), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := checkMethodFrozen(dir); err == nil {
		t.Fatal("expected an error when manifest.json records gate_pass=false")
	}
}

func TestVoynichAcceptsPassedGate(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "METHOD_FROZEN"), []byte("marker\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "manifest.json"), []byte(`{"gate_pass": true, "config": {"Threshold": 0.27}}`), 0o644); err != nil {
		t.Fatal(err)
	}
	tau, err := checkMethodFrozen(dir)
	if err != nil {
		t.Fatalf("expected success, got %v", err)
	}
	if tau != 0.27 {
		t.Fatalf("expected tau=0.27, got %v", tau)
	}
}

// run() with the "voynich" subcommand must go through the same guard,
// even without any flags overriding the default corpus/out-dir.
func TestRunVoynichRefusesWithoutFrozenMethod(t *testing.T) {
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(cwd)

	code := run([]string{"voynich"})
	if code == 0 {
		t.Fatal("expected non-zero exit when no synthetic-validation directory exists at all")
	}
}
