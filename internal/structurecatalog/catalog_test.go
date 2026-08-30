package structurecatalog

import (
	"compress/gzip"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunPreservesZeroRules(t *testing.T) {
	d := t.TempDir()
	corpus := filepath.Join(d, "c.txt")
	if err := os.WriteFile(corpus, []byte("ab ac\nab b\n"), 0644); err != nil {
		t.Fatal(err)
	}
	out := filepath.Join(d, "out")
	cat, err := Run(Config{CorpusPath: corpus, OutputDir: out, MinFrequency: 1})
	if err != nil {
		t.Fatal(err)
	}
	if len(cat.Rules) == 0 {
		t.Fatal("no rules")
	}
	b, err := os.ReadFile(filepath.Join(out, "GLYPH_BIGRAM_RULES.tsv"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(b), "UNOBSERVED\tNEVER_OBSERVED") {
		t.Fatal("zero bigrams were not retained")
	}
	b, err = os.ReadFile(filepath.Join(out, "TOKEN_TRANSITIONS_UNOBSERVED.tsv"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(b), "\t0\t") {
		t.Fatal("zero token transitions were not retained")
	}
	f, err := os.Open(filepath.Join(out, "TOKEN_TRANSITION_COMPLEMENT.json.gz"))
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	gz, err := gzip.NewReader(f)
	if err != nil {
		t.Fatal(err)
	}
	var x any
	if err = json.NewDecoder(gz).Decode(&x); err != nil {
		t.Fatal(err)
	}
}
