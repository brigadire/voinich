package notation

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// realComparativeNotationDir is the actual frozen protocol directory this
// package ships. Regression tests here exercise the shipped
// GLOBAL_FREEZE_MANIFEST.json directly, so a change that silently breaks
// the real freeze binding fails the test suite, not just a synthetic
// fixture.
const realComparativeNotationDir = "../../research/comparative_notation"

func withSyntheticRequiredArtifacts(t *testing.T, specs []FreezeArtifactSpec) {
	t.Helper()
	orig := RequiredGlobalFreezeArtifacts
	RequiredGlobalFreezeArtifacts = specs
	t.Cleanup(func() { RequiredGlobalFreezeArtifacts = orig })
}

func writeSyntheticFreezeFixture(t *testing.T, dir string) (specs []FreezeArtifactSpec, hashes map[string]string) {
	t.Helper()
	files := map[string]string{
		"FIXTURE_PROTOCOL.md":   "# fixture protocol\n\nfrozen text A\n",
		"FIXTURE_SCALES.tsv":    "metric_id\tvalue\nG01\t1.0\n",
	}
	specs = []FreezeArtifactSpec{
		{"FIXTURE_PROTOCOL.md", RoleProtocol, ""},
		{"FIXTURE_SCALES.tsv", RoleCalibration, ""},
	}
	hashes = map[string]string{}
	for name, content := range files {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0644); err != nil {
			t.Fatal(err)
		}
		h, err := FileSHA256(filepath.Join(dir, name))
		if err != nil {
			t.Fatal(err)
		}
		hashes[name] = h
	}
	return specs, hashes
}

func writeSyntheticManifest(t *testing.T, dir string, entries []FreezeArtifactEntry) {
	t.Helper()
	m := GlobalFreezeManifest{
		SchemaVersion: GlobalFreezeManifestSchemaVersion,
		Artifacts:     entries,
		ProtocolStatus: GlobalFreezeProtocolStatus{
			GlobalComparisonProtocolFrozen: true, GlobalFreezeCryptographicallyBound: true, ProductionComparativeRunAuthorized: false,
		},
	}
	b, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "GLOBAL_FREEZE_MANIFEST.json"), b, 0644); err != nil {
		t.Fatal(err)
	}
}

// TestGlobalFreezeValidCompleteManifest is the positive control: a
// synthetic manifest that binds every required artifact correctly must
// verify with zero errors.
func TestGlobalFreezeValidCompleteManifest(t *testing.T) {
	dir := t.TempDir()
	specs, hashes := writeSyntheticFreezeFixture(t, dir)
	withSyntheticRequiredArtifacts(t, specs)
	writeSyntheticManifest(t, dir, []FreezeArtifactEntry{
		{Path: "FIXTURE_PROTOCOL.md", SHA256: hashes["FIXTURE_PROTOCOL.md"], Role: RoleProtocol},
		{Path: "FIXTURE_SCALES.tsv", SHA256: hashes["FIXTURE_SCALES.tsv"], Role: RoleCalibration},
	})
	errs, err := VerifyGlobalFreezeManifest(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(errs) != 0 {
		t.Fatalf("expected a fully valid manifest, got errors: %v", errs)
	}
}

// TestGlobalFreezeMissingMandatoryBinding: a manifest that never mentions a
// required artifact at all must fail closed.
func TestGlobalFreezeMissingMandatoryBinding(t *testing.T) {
	dir := t.TempDir()
	specs, hashes := writeSyntheticFreezeFixture(t, dir)
	withSyntheticRequiredArtifacts(t, specs)
	writeSyntheticManifest(t, dir, []FreezeArtifactEntry{
		{Path: "FIXTURE_PROTOCOL.md", SHA256: hashes["FIXTURE_PROTOCOL.md"], Role: RoleProtocol},
		// FIXTURE_SCALES.tsv binding omitted entirely.
	})
	errs, err := VerifyGlobalFreezeManifest(dir)
	if err != nil {
		t.Fatal(err)
	}
	if !containsSubstring(errs, "missing mandatory binding: FIXTURE_SCALES.tsv") {
		t.Fatalf("expected a missing-mandatory-binding error, got: %v", errs)
	}
}

// TestGlobalFreezeMissingArtifactOnDisk: the manifest binds a path that no
// longer exists on disk.
func TestGlobalFreezeMissingArtifactOnDisk(t *testing.T) {
	dir := t.TempDir()
	specs, hashes := writeSyntheticFreezeFixture(t, dir)
	withSyntheticRequiredArtifacts(t, specs)
	writeSyntheticManifest(t, dir, []FreezeArtifactEntry{
		{Path: "FIXTURE_PROTOCOL.md", SHA256: hashes["FIXTURE_PROTOCOL.md"], Role: RoleProtocol},
		{Path: "FIXTURE_SCALES.tsv", SHA256: hashes["FIXTURE_SCALES.tsv"], Role: RoleCalibration},
	})
	if err := os.Remove(filepath.Join(dir, "FIXTURE_SCALES.tsv")); err != nil {
		t.Fatal(err)
	}
	errs, err := VerifyGlobalFreezeManifest(dir)
	if err != nil {
		t.Fatal(err)
	}
	if !containsSubstring(errs, "FIXTURE_SCALES.tsv: cannot read artifact") {
		t.Fatalf("expected a cannot-read-artifact error, got: %v", errs)
	}
}

// TestGlobalFreezeModifiedArtifactAfterFreeze: the artifact's bytes were
// edited after the manifest recorded its hash — this must be detected as a
// hash mismatch (content drift), never silently accepted.
func TestGlobalFreezeModifiedArtifactAfterFreeze(t *testing.T) {
	dir := t.TempDir()
	specs, hashes := writeSyntheticFreezeFixture(t, dir)
	withSyntheticRequiredArtifacts(t, specs)
	writeSyntheticManifest(t, dir, []FreezeArtifactEntry{
		{Path: "FIXTURE_PROTOCOL.md", SHA256: hashes["FIXTURE_PROTOCOL.md"], Role: RoleProtocol},
		{Path: "FIXTURE_SCALES.tsv", SHA256: hashes["FIXTURE_SCALES.tsv"], Role: RoleCalibration},
	})
	if err := os.WriteFile(filepath.Join(dir, "FIXTURE_PROTOCOL.md"), []byte("# fixture protocol\n\ntampered text\n"), 0644); err != nil {
		t.Fatal(err)
	}
	errs, err := VerifyGlobalFreezeManifest(dir)
	if err != nil {
		t.Fatal(err)
	}
	if !containsSubstring(errs, "FIXTURE_PROTOCOL.md hash mismatch") {
		t.Fatalf("expected a hash-mismatch error for the modified artifact, got: %v", errs)
	}
}

// TestGlobalFreezeWrongChecksum: the manifest's recorded hash was simply
// wrong from the start (fat-fingered or fabricated), independent of any
// later drift.
func TestGlobalFreezeWrongChecksum(t *testing.T) {
	dir := t.TempDir()
	specs, hashes := writeSyntheticFreezeFixture(t, dir)
	withSyntheticRequiredArtifacts(t, specs)
	writeSyntheticManifest(t, dir, []FreezeArtifactEntry{
		{Path: "FIXTURE_PROTOCOL.md", SHA256: strings.Repeat("0", 64), Role: RoleProtocol},
		{Path: "FIXTURE_SCALES.tsv", SHA256: hashes["FIXTURE_SCALES.tsv"], Role: RoleCalibration},
	})
	errs, err := VerifyGlobalFreezeManifest(dir)
	if err != nil {
		t.Fatal(err)
	}
	if !containsSubstring(errs, "FIXTURE_PROTOCOL.md hash mismatch") {
		t.Fatalf("expected a hash-mismatch error for the wrong checksum, got: %v", errs)
	}
}

// TestGlobalFreezeDuplicateEntry: the same path is bound twice in the
// artifacts array.
func TestGlobalFreezeDuplicateEntry(t *testing.T) {
	dir := t.TempDir()
	specs, hashes := writeSyntheticFreezeFixture(t, dir)
	withSyntheticRequiredArtifacts(t, specs)
	writeSyntheticManifest(t, dir, []FreezeArtifactEntry{
		{Path: "FIXTURE_PROTOCOL.md", SHA256: hashes["FIXTURE_PROTOCOL.md"], Role: RoleProtocol},
		{Path: "FIXTURE_PROTOCOL.md", SHA256: hashes["FIXTURE_PROTOCOL.md"], Role: RoleProtocol},
		{Path: "FIXTURE_SCALES.tsv", SHA256: hashes["FIXTURE_SCALES.tsv"], Role: RoleCalibration},
	})
	errs, err := VerifyGlobalFreezeManifest(dir)
	if err != nil {
		t.Fatal(err)
	}
	if !containsSubstring(errs, "duplicate manifest entry for FIXTURE_PROTOCOL.md") {
		t.Fatalf("expected a duplicate-entry error, got: %v", errs)
	}
}

// TestGlobalFreezeRealManifestValid pins the actual shipped
// GLOBAL_FREEZE_MANIFEST.json under research/comparative_notation: every
// mandatory binding and every cross-reference check must pass with zero
// errors. A future edit that silently breaks the real freeze binding (a
// modified frozen artifact, a removed mandatory file, a drifted
// cross-reference) fails this test, not just a synthetic fixture.
func TestGlobalFreezeRealManifestValid(t *testing.T) {
	if _, err := os.Stat(filepath.Join(realComparativeNotationDir, "GLOBAL_FREEZE_MANIFEST.json")); err != nil {
		t.Skipf("research/comparative_notation not present in this checkout: %v", err)
	}
	bindingErrs, err := VerifyGlobalFreezeManifest(realComparativeNotationDir)
	if err != nil {
		t.Fatal(err)
	}
	if len(bindingErrs) != 0 {
		t.Fatalf("real GLOBAL_FREEZE_MANIFEST.json has binding errors: %v", bindingErrs)
	}
	crossErrs, err := GlobalFreezeCrossReferenceChecks(realComparativeNotationDir)
	if err != nil {
		t.Fatal(err)
	}
	if len(crossErrs) != 0 {
		t.Fatalf("real frozen artifacts have cross-reference inconsistencies: %v", crossErrs)
	}
}

func containsSubstring(errs []string, substr string) bool {
	for _, e := range errs {
		if strings.Contains(e, substr) {
			return true
		}
	}
	return false
}
