package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNormalizeCorpusNameStripsUniformSuffix(t *testing.T) {
	if got := normalizeCorpusName("Doyle-H4-uniform"); got != "Doyle-H4" {
		t.Fatalf("got %q", got)
	}
	if got := normalizeCorpusName("Voynich"); got != "Voynich" {
		t.Fatalf("got %q", got)
	}
}

func TestJoinTask58ReproducesStoredMetric(t *testing.T) {
	fixtureDir := t.TempDir()
	fixture := filepath.Join(fixtureDir, "comparison.tsv")
	content := "corpus\tpath\ttokens\tlines\tpairs\ttypes\ttoken_observed_bits\ttoken_shuffle_mean_bits\ttoken_shuffle_sd_bits\ttoken_corrected_bits\ttoken_share\tglyph_status\tedge_observed_bits\tedge_shuffle_mean_bits\tedge_shuffle_sd_bits\tedge_corrected_bits\tedge_share\n" +
		"Voynich\tp\t1\t1\t1\t1\t3.05\t2.96\t0.01\t0.09\t0.011\tAPPLICABLE\t0.25\t0.03\t0.001\t0.21\t0.06\n"
	if err := os.WriteFile(fixture, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	oldWD, _ := os.Getwd()
	chdirTemp := t.TempDir()
	if err := os.MkdirAll(filepath.Join(chdirTemp, "experiments/rozanova-temerev-v1"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Rename(fixture, filepath.Join(chdirTemp, "experiments/rozanova-temerev-v1/comparison.tsv")); err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(chdirTemp); err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(oldWD)

	rep := newReport(chdirTemp)
	rep.summary("Voynich")
	if err := joinTask58(rep, chdirTemp); err != nil {
		t.Fatal(err)
	}
	s := rep.summaries["Voynich"]
	if !s.HasTask58 {
		t.Fatal("expected HasTask58=true")
	}
	if s.TokenOrderMI != 3.05 || s.TokenShare != 0.011 || s.GlyphEdgeMI != 0.25 {
		t.Fatalf("stored metrics not reproduced exactly: %+v", s)
	}
}

func TestJoinTask59ReproducesStoredMetric(t *testing.T) {
	chdirTemp := t.TempDir()
	if err := os.MkdirAll(filepath.Join(chdirTemp, "experiments/glyph-position-v1"), 0o755); err != nil {
		t.Fatal(err)
	}
	content := "Corpus\tTokens\tGlyphVocab\tWeightedEntropy\tStrictSpecialists\tP95Specialists\tHighFreqSpecialists\tExclusions\n" +
		"Voynich\t39380\t45\t0.5964\t1\t7\t6\t4\n"
	if err := os.WriteFile(filepath.Join(chdirTemp, "experiments/glyph-position-v1/POSITIONAL_SPECIALIZATION_COMPARISON.tsv"), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	oldWD, _ := os.Getwd()
	if err := os.Chdir(chdirTemp); err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(oldWD)

	rep := newReport(chdirTemp)
	rep.summary("Voynich")
	if err := joinTask59(rep, chdirTemp); err != nil {
		t.Fatal(err)
	}
	s := rep.summaries["Voynich"]
	if !s.HasTask59 || s.HighFreqSpecialists != 6 || s.WeightedEntropy != 0.5964 {
		t.Fatalf("stored metrics not reproduced exactly: %+v", s)
	}
}
