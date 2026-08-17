package normalizationcompare

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"zcore.dev/voinich/internal/sequenceanalyze"
)

const runTestClassesYAML = `
meta:
  input_corpus: corpus.txt
  singleton_mode: token
  min_token_count: 1
  random_matching: "matched by frequency bin"
models:
  - threshold: 0.8
    label: "080"
    stats:
      multi_member_classes: 1
      classes: 2
    classes:
      - id: C0001
        size: 2
        members:
          - token: a
            count: 60
          - token: b
            count: 60
`

// runFixture writes a self-consistent corpus/classes/normalized/raw-analysis
// set: the "structural" pass reads normalized_080.txt back unchanged, so its
// SequenceMeta trivially matches the raw analysis's, exactly like a real
// structural-normalize output that made no interesting merges for this
// threshold would.
func runFixture(t *testing.T) Config {
	t.Helper()
	d := t.TempDir()
	write := func(name, body string) string {
		p := filepath.Join(d, name)
		if err := os.WriteFile(p, []byte(body), 0600); err != nil {
			t.Fatal(err)
		}
		return p
	}
	corpus := strings.Repeat("a b c a c b\n", 30)
	corpusPath := write("corpus.txt", corpus)
	classesPath := write("structural_classes.yaml", runTestClassesYAML)
	write("normalized_080.txt", corpus)

	params := sequenceanalyze.DefaultParameters()
	rawOutput, err := sequenceanalyze.AnalyzeFile(corpusPath, params)
	if err != nil {
		t.Fatal(err)
	}
	rawAnalysisPath := filepath.Join(d, "sequence_analysis.yaml")
	if err := WriteAnalysisYAML(rawAnalysisPath, rawOutput); err != nil {
		t.Fatal(err)
	}

	return Config{
		ClassesPath: classesPath, InputPath: corpusPath, RawAnalysisPath: rawAnalysisPath,
		NormalizedPattern: filepath.Join(d, "normalized_%s.txt"), AnalysisPattern: filepath.Join(d, "sequence_analysis_%s.yaml"),
		SequenceAnalyzerPath: "bin/sequence-analyze", OutputPath: filepath.Join(d, "out.yaml"),
		RandomRuns: 5, RandomSeed: 1, Workers: 1, Context: context.Background(),
	}
}

func TestRunAndWriteLocalConcurrentWorkersMatchSequential(t *testing.T) {
	sequential := runFixture(t)
	sequential.OutputPath = filepath.Join(t.TempDir(), "seq.yaml")
	sequential.Workers = 1
	if err := RunAndWrite(sequential); err != nil {
		t.Fatal(err)
	}
	sequentialBytes, err := os.ReadFile(sequential.OutputPath)
	if err != nil {
		t.Fatal(err)
	}

	concurrent := runFixture(t)
	concurrent.OutputPath = filepath.Join(t.TempDir(), "conc.yaml")
	concurrent.Workers = 4
	if err := RunAndWrite(concurrent); err != nil {
		t.Fatal(err)
	}
	concurrentBytes, err := os.ReadFile(concurrent.OutputPath)
	if err != nil {
		t.Fatal(err)
	}

	if string(sequentialBytes) != string(concurrentBytes) {
		t.Fatalf("local goroutine executor is not order-independent:\nsequential=%s\nconcurrent=%s", sequentialBytes, concurrentBytes)
	}
}

func TestRunAndWriteRejectsCorpusInvariantMismatch(t *testing.T) {
	c := runFixture(t)
	// Overwrite the raw analysis with a corpus that has different invariants
	// than the actual corpus/normalized files this run reads: RunAndWrite
	// must hard-fail rather than silently comparing mismatched corpora.
	other := "x y\nx y z\n"
	otherPath := filepath.Join(t.TempDir(), "other.txt")
	if err := os.WriteFile(otherPath, []byte(other), 0600); err != nil {
		t.Fatal(err)
	}
	otherOutput, err := sequenceanalyze.AnalyzeFile(otherPath, sequenceanalyze.DefaultParameters())
	if err != nil {
		t.Fatal(err)
	}
	if err := WriteAnalysisYAML(c.RawAnalysisPath, otherOutput); err != nil {
		t.Fatal(err)
	}
	if err := RunAndWrite(c); err == nil {
		t.Fatal("expected corpus invariant mismatch to be a hard error")
	}
}
