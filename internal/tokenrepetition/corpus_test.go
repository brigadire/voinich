package tokenrepetition

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadCorpusDetectsOpaqueTokens(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "opaque.txt")
	if err := os.WriteFile(path, []byte("x000001 x000002 x000001 x000003\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	c, err := LoadCorpus(path, "opaque")
	if err != nil {
		t.Fatal(err)
	}
	if !c.Opaque {
		t.Fatal("expected opaque corpus to be detected as such")
	}
}

func TestLoadCorpusDetectsNonOpaqueTokens(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "natural.txt")
	if err := os.WriteFile(path, []byte("the quick brown fox\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	c, err := LoadCorpus(path, "natural")
	if err != nil {
		t.Fatal(err)
	}
	if c.Opaque {
		t.Fatal("expected natural-language corpus to NOT be detected as opaque")
	}
}
