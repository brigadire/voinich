package tokenrepetition

import (
	"os"
	"path/filepath"
	"testing"
)

const fixtureIVTFF = `<f1r>
<f1r.1,@P0>      qokeedy.qokedy
<f1r.2,@P0>      chol.shol
<f2r>
<f2r.1,@L0>      ytoail
<f2r.2,@Lf>      keer.dal
<f2r.3,@Lf>      keer.dal
`

func writeFixtureIVTFF(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "fixture.ivtff")
	if err := os.WriteFile(path, []byte(fixtureIVTFF), 0o644); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestExtractLabelsOnlyReadsLabelLoci(t *testing.T) {
	path := writeFixtureIVTFF(t)
	labels, err := ExtractLabels(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(labels) != 3 {
		t.Fatalf("expected 3 label loci (paragraph loci excluded), got %d: %+v", len(labels), labels)
	}
	if labels[0].Tokens[0] != "ytoail" {
		t.Fatalf("unexpected first label tokens: %+v", labels[0])
	}
}

func TestExtractLabelsReproducible(t *testing.T) {
	path := writeFixtureIVTFF(t)
	a, err := ExtractLabels(path)
	if err != nil {
		t.Fatal(err)
	}
	b, err := ExtractLabels(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(a) != len(b) {
		t.Fatalf("non-reproducible extraction: %d != %d", len(a), len(b))
	}
	for i := range a {
		if a[i].RawLabel != b[i].RawLabel || len(a[i].Tokens) != len(b[i].Tokens) {
			t.Fatalf("non-reproducible extraction at %d: %+v != %+v", i, a[i], b[i])
		}
	}
}

func TestSummarizeLabelRepetitionCountsRepeatedCompleteLabels(t *testing.T) {
	path := writeFixtureIVTFF(t)
	labels, err := ExtractLabels(path)
	if err != nil {
		t.Fatal(err)
	}
	stats := SummarizeLabelRepetition(labels)
	if stats.Instances != 3 {
		t.Fatalf("expected 3 instances, got %d", stats.Instances)
	}
	// "keer dal" appears twice (f2r.2 and f2r.3): one repeated complete label.
	if stats.RepeatedCompleteLabels != 1 {
		t.Fatalf("expected 1 repeated complete label, got %d", stats.RepeatedCompleteLabels)
	}
}
