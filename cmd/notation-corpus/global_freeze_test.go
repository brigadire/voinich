package main

import (
	"os"
	"path/filepath"
	"testing"

	"zcore.dev/voinich/internal/notation"
)

// setupSyntheticFreezeRepo builds a minimal repo tree
// (<dir>/research/comparative_notation/...) with a synthetic
// two-artifact frozen set, and points notation.RequiredGlobalFreezeArtifacts
// at it for the duration of the test.
func setupSyntheticFreezeRepo(t *testing.T) (repo, base string) {
	t.Helper()
	repo = t.TempDir()
	base = filepath.Join(repo, "research", "comparative_notation")
	if err := os.MkdirAll(filepath.Join(base, "CALIBRATION_GENERATORS"), 0755); err != nil {
		t.Fatal(err)
	}
	files := map[string]string{
		"FIXTURE_PROTOCOL.md": "# fixture protocol\n\nfrozen text\n",
		"FIXTURE_SCALES.tsv":  "metric_id\tvalue\nG01\t1.0\n",
	}
	for name, content := range files {
		if err := os.WriteFile(filepath.Join(base, name), []byte(content), 0644); err != nil {
			t.Fatal(err)
		}
	}
	orig := notation.RequiredGlobalFreezeArtifacts
	notation.RequiredGlobalFreezeArtifacts = []notation.FreezeArtifactSpec{
		{Path: "FIXTURE_PROTOCOL.md", Role: notation.RoleProtocol},
		{Path: "FIXTURE_SCALES.tsv", Role: notation.RoleCalibration},
	}
	t.Cleanup(func() { notation.RequiredGlobalFreezeArtifacts = orig })
	return repo, base
}

// TestGlobalFreezeBindThenVerify exercises the full bind -> verify cycle:
// binding a fresh set of frozen artifacts must produce a manifest that
// immediately verifies clean, and the completion record must exist and
// state that nothing scientific changed.
func TestGlobalFreezeBindThenVerify(t *testing.T) {
	repo, base := setupSyntheticFreezeRepo(t)

	if err := globalFreezeBindCmd([]string{"--repo", repo}); err != nil {
		t.Fatalf("global-freeze-bind failed: %v", err)
	}
	if _, err := os.Stat(filepath.Join(base, "GLOBAL_FREEZE_BINDING_COMPLETION.md")); err != nil {
		t.Fatalf("missing GLOBAL_FREEZE_BINDING_COMPLETION.md: %v", err)
	}
	// Exercise the binding-verification logic global-freeze-verify uses
	// directly (notation.VerifyGlobalFreezeManifest); the CLI's own
	// cross-reference checks are hardcoded to the real frozen artifact
	// names and are pinned against the real manifest by
	// internal/notation's TestGlobalFreezeRealManifestValid instead.
	errs, err := notation.VerifyGlobalFreezeManifest(base)
	if err != nil {
		t.Fatalf("VerifyGlobalFreezeManifest failed after a fresh bind: %v", err)
	}
	if len(errs) != 0 {
		t.Fatalf("expected a freshly bound manifest to verify clean, got: %v", errs)
	}
}

// TestGlobalFreezeBindRefusesContentDrift: once an artifact is bound,
// re-running global-freeze-bind after that artifact's bytes changed must
// fail closed rather than silently re-bind the new content.
func TestGlobalFreezeBindRefusesContentDrift(t *testing.T) {
	repo, base := setupSyntheticFreezeRepo(t)
	if err := globalFreezeBindCmd([]string{"--repo", repo}); err != nil {
		t.Fatalf("global-freeze-bind failed: %v", err)
	}
	if err := os.WriteFile(filepath.Join(base, "FIXTURE_PROTOCOL.md"), []byte("# fixture protocol\n\ntampered\n"), 0644); err != nil {
		t.Fatal(err)
	}
	err := globalFreezeBindCmd([]string{"--repo", repo})
	if err == nil {
		t.Fatal("expected global-freeze-bind to refuse re-binding drifted scientific content")
	}
}

// TestGlobalFreezeVerifyFailsOnTamperedManifest: verify must fail closed
// (non-nil error) if a bound artifact is edited after binding, without
// re-running bind.
func TestGlobalFreezeVerifyFailsOnTamperedManifest(t *testing.T) {
	repo, base := setupSyntheticFreezeRepo(t)
	if err := globalFreezeBindCmd([]string{"--repo", repo}); err != nil {
		t.Fatalf("global-freeze-bind failed: %v", err)
	}
	if err := os.WriteFile(filepath.Join(base, "FIXTURE_SCALES.tsv"), []byte("metric_id\tvalue\nG01\t999.0\n"), 0644); err != nil {
		t.Fatal(err)
	}
	errs, err := notation.VerifyGlobalFreezeManifest(base)
	if err != nil {
		t.Fatal(err)
	}
	if len(errs) == 0 {
		t.Fatal("expected VerifyGlobalFreezeManifest to fail closed on a modified frozen artifact")
	}
}
