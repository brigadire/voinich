package replicatedlocalaudit

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func writeGenericCorpus(t *testing.T, lines, tokensPerLine int) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "corpus.txt")
	var b strings.Builder
	for l := range lines {
		for i := range tokensPerLine {
			fmt.Fprintf(&b, "w%d_%d ", l, i)
		}
		b.WriteByte('\n')
	}
	if err := os.WriteFile(path, []byte(b.String()), 0644); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestLoadGenericTokensDeterministicAndKnown(t *testing.T) {
	path := writeGenericCorpus(t, 300, 4)
	corpus := readGenericCorpusTokens(t, path)

	a, err := loadGenericTokens(path, corpus)
	if err != nil {
		t.Fatal(err)
	}
	if len(a) != len(corpus) {
		t.Fatalf("token count %d != corpus token count %d", len(a), len(corpus))
	}
	for _, tok := range a {
		if tok.Hand != "generic" {
			t.Fatalf("Hand must always be the generic sentinel, got %q", tok.Hand)
		}
		// Mirrors loadInputs's local known() gate: a generic token's
		// Currier/Hand must never be empty/"?"/"null", or every physical
		// block would be silently discarded as "unknown metadata".
		if tok.Currier == "" || tok.Hand == "" {
			t.Fatalf("generic tokens must never have an empty Currier/Hand: %+v", tok)
		}
	}

	b, err := loadGenericTokens(path, corpus)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(a, b) {
		t.Fatal("loadGenericTokens is not deterministic across identical calls")
	}
}

func readGenericCorpusTokens(t *testing.T, path string) []string {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return strings.Fields(string(b))
}
