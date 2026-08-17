package normalizationcompare

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"zcore.dev/voinich/internal/sequenceanalyze"
)

// referenceLoadViaSubprocessPath reproduces exactly what this tool used to
// do when it shelled out to a sequence-analyze subprocess: marshal an
// analyzer Output to YAML, write it to disk, then unmarshal it back into the
// local SequenceAnalysis type via LoadSequence. It is the correctness oracle
// for FromAnalyzerOutput, which now takes an in-process sequenceanalyze.Output
// directly instead of round-tripping it through a file.
func referenceLoadViaSubprocessPath(t *testing.T, o sequenceanalyze.Output) SequenceAnalysis {
	t.Helper()
	path := filepath.Join(t.TempDir(), "analysis.yaml")
	if err := WriteAnalysisYAML(path, o); err != nil {
		t.Fatalf("WriteAnalysisYAML: %v", err)
	}
	got, err := LoadSequence(path)
	if err != nil {
		t.Fatalf("LoadSequence: %v", err)
	}
	return got
}

// TestFromAnalyzerOutputMatchesSubprocessRoundTrip is the reference-vs-
// optimized oracle for the normalization-compare subprocess-elimination
// change: it proves that reading sequenceanalyze.Output fields directly
// in-process (FromAnalyzerOutput) produces the exact same SequenceAnalysis
// value that the old code got by writing the analyzer's output to a YAML
// file and reading it back (the byte-identical marshal/unmarshal round trip
// a sequence-analyze subprocess used to perform).
func TestFromAnalyzerOutputMatchesSubprocessRoundTrip(t *testing.T) {
	corpusPath := filepath.Join(t.TempDir(), "corpus.txt")
	corpus := "the quick brown fox\nthe quick brown fox\nthe lazy dog runs\nthe quick cat sleeps\n"
	if err := os.WriteFile(corpusPath, []byte(corpus), 0o644); err != nil {
		t.Fatal(err)
	}
	output, err := sequenceanalyze.AnalyzeFile(corpusPath, sequenceanalyze.DefaultParameters())
	if err != nil {
		t.Fatalf("AnalyzeFile: %v", err)
	}

	got := FromAnalyzerOutput(output)
	want := referenceLoadViaSubprocessPath(t, output)

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("FromAnalyzerOutput diverged from subprocess round trip:\n got=%+v\nwant=%+v", got, want)
	}
	if len(got.NGramSummary) == 0 || len(got.ContextOrderAnalysis) == 0 {
		t.Fatal("fixture produced no n-grams/context orders; test would pass vacuously")
	}
}
