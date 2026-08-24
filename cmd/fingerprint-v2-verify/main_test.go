package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func sum(b []byte) string { h := sha256.Sum256(b); return hex.EncodeToString(h[:]) }

func fixture(t *testing.T) (string, string, string) {
	t.Helper()
	d := t.TempDir()
	raw := []byte("raw\n")
	prepared := []byte("prepared\n")
	if err := os.WriteFile(filepath.Join(d, "raw.txt"), raw, 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(d, "prepared.txt"), prepared, 0644); err != nil {
		t.Fatal(err)
	}
	top, _ := json.Marshal(map[string]any{"corpus": map[string]any{"sha256": sum(prepared)}})
	if err := os.WriteFile(filepath.Join(d, "top.json"), top, 0644); err != nil {
		t.Fatal(err)
	}
	m := Manifest{SchemaVersion: 1, Version: "test", Status: "AUTHORITATIVE", Artifacts: []Artifact{
		{ID: "raw", Path: "raw.txt", SHA256: sum(raw), Kind: "raw"},
		{ID: "prepared", Path: "prepared.txt", SHA256: sum(prepared), Kind: "generated", Parents: []Parent{{ID: "raw", SHA256: sum(raw)}}},
		{ID: "top", Path: "top.json", SHA256: sum(top), Kind: "artifact", Parents: []Parent{{ID: "prepared", SHA256: sum(prepared)}}, Bindings: []Binding{{JSONPath: "corpus.sha256", ParentID: "prepared"}}},
	}}
	b, _ := json.Marshal(m)
	mp := filepath.Join(d, "manifest.json")
	if err := os.WriteFile(mp, b, 0644); err != nil {
		t.Fatal(err)
	}
	return d, mp, sum(prepared)
}

func TestRejectsNonAuthoritativeManifest(t *testing.T) {
	d, m, _ := fixture(t)
	b, err := os.ReadFile(m)
	if err != nil {
		t.Fatal(err)
	}
	b = []byte(strings.Replace(string(b), "AUTHORITATIVE", "PROVENANCE_UNRESOLVED", 1))
	if err = os.WriteFile(m, b, 0644); err != nil {
		t.Fatal(err)
	}
	if err = Verify(m, d); err == nil {
		t.Fatal("expected non-authoritative manifest to be rejected")
	}
	if err = verify(m, d, true); err != nil {
		t.Fatalf("audit mode should still verify graph integrity: %v", err)
	}
}

func TestVerifyValidTransitiveManifest(t *testing.T) {
	d, m, _ := fixture(t)
	if err := Verify(m, d); err != nil {
		t.Fatal(err)
	}
}
func TestRejectsChangedTransitiveChecksumInManifest(t *testing.T) {
	d, m, old := fixture(t)
	b, err := os.ReadFile(m)
	if err != nil {
		t.Fatal(err)
	}
	b = []byte(strings.Replace(string(b), old, strings.Repeat("0", 64), 1))
	if err = os.WriteFile(m, b, 0644); err != nil {
		t.Fatal(err)
	}
	if err = Verify(m, d); err == nil {
		t.Fatal("expected transitive checksum mismatch")
	}
}
func TestRejectsPreparedSubstitutionWithUnchangedTopArtifact(t *testing.T) {
	d, m, _ := fixture(t)
	if err := os.WriteFile(filepath.Join(d, "prepared.txt"), []byte("substitute\n"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := Verify(m, d); err == nil {
		t.Fatal("expected substituted prepared corpus to be rejected")
	}
}

func TestJSONStringAtArrayPath(t *testing.T) {
	p := filepath.Join(t.TempDir(), "array.json")
	if err := os.WriteFile(p, []byte(`{"items":[{"sha256":"abc"}]}`), 0644); err != nil {
		t.Fatal(err)
	}
	got, err := jsonStringAt(p, "items.0.sha256")
	if err != nil || got != "abc" {
		t.Fatalf("jsonStringAt array path = %q, %v", got, err)
	}
}
