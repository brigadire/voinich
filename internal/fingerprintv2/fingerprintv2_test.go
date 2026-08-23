package fingerprintv2

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"testing"
)

func fixtureDir(t *testing.T) string {
	t.Helper()
	dir, err := os.MkdirTemp(".", "fingerprintv2-test-")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.RemoveAll(dir) })
	return dir
}

func writeFixture(t *testing.T, dir, name, contents string) string {
	t.Helper()
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte(contents), 0644); err != nil {
		t.Fatal(err)
	}
	return path
}

func syntheticCorpus(t *testing.T) corpus {
	t.Helper()
	dir := fixtureDir(t)
	path := writeFixture(t, dir, "synthetic.txt", "caa cab\ndaa dab\nfaa fab\n")
	c, err := loadCorpus(CorpusConfig{ID: "synthetic", Path: path, GlyphMode: "natural"})
	if err != nil {
		t.Fatal(err)
	}
	return c
}

func testConfig(path string) Config {
	return Config{
		OutputDir: "unused", Primary: CorpusConfig{ID: "synthetic", Path: path, GlyphMode: "natural"},
		Seed: 17, Repetitions: 4, MinRuleSupport: 3, Alpha: 0.05, GraphSwaps: 2,
		Grammar: GrammarConfig{Modes: []string{"structure-preserving", "frequency-aware"}},
	}
}

func TestSyntheticProductiveRuleAndNegativeControl(t *testing.T) {
	c := syntheticCorpus(t)
	cfg, err := testConfig("unused").normalized()
	if err != nil {
		t.Fatal(err)
	}
	positive := analyzeBare(c, cfg)
	const suffixAB = "SUBSTITUTION|SUFFIX|END|a→b"
	if !positive.candidates[suffixAB] {
		t.Fatalf("productive suffix rule %q absent from %+v", suffixAB, positive.lp1.Rules)
	}
	families, _ := lp3(c, positive.candidates, positive.graph, 3, rand.New(rand.NewSource(3)))
	if families.ProductiveRuleCount == 0 || families.SmallFamilyCount == 0 {
		t.Fatalf("expected productive-rule graph components, got %+v", families)
	}

	dir := fixtureDir(t)
	path := writeFixture(t, dir, "negative.txt", "abc\ndef\nghi\njkl\n")
	negative, err := loadCorpus(CorpusConfig{ID: "negative", Path: path, GlyphMode: "natural"})
	if err != nil {
		t.Fatal(err)
	}
	negativeMetrics := analyzeBare(negative, cfg)
	if negativeMetrics.lp1.DirectedPairCount != 0 || len(negativeMetrics.candidates) != 0 {
		t.Fatalf("negative corpus unexpectedly has productive edit rules: %+v", negativeMetrics.lp1)
	}
}

func TestCGrammarPreservationAndDeterminism(t *testing.T) {
	c := syntheticCorpus(t)
	model := newGrammarModel(c)
	for _, mode := range []string{"structure-preserving", "frequency-aware"} {
		first, err := model.generate(c, mode, rand.New(rand.NewSource(44)))
		if err != nil {
			t.Fatalf("%s first generation: %v", mode, err)
		}
		second, err := model.generate(c, mode, rand.New(rand.NewSource(44)))
		if err != nil {
			t.Fatalf("%s second generation: %v", mode, err)
		}
		firstTokens, secondTokens := make([]string, len(first.records)), make([]string, len(second.records))
		for i := range first.records {
			firstTokens[i], secondTokens[i] = first.records[i].Token, second.records[i].Token
		}
		if !reflect.DeepEqual(firstTokens, secondTokens) {
			t.Fatalf("%s generation is not deterministic: %v != %v", mode, firstTokens, secondTokens)
		}
		d := grammarDiagnostic(c, first)
		if !d.TokenCountExact || !d.LengthDistributionExact || !d.AlphabetExact {
			t.Fatalf("%s failed exact C-GRAMMAR invariant: %+v", mode, d)
		}
		if mode == "frequency-aware" && d.TokenFrequencyTV != 0 {
			t.Fatalf("frequency-aware generator did not preserve frequency ranks: %+v", d)
		}
	}
}

func TestSeededPipelineIsDeterministic(t *testing.T) {
	dir := fixtureDir(t)
	path := writeFixture(t, dir, "deterministic.txt", "caa cab\ndaa dab\nfaa fab\n")
	cfg := testConfig(path)
	first, err := Run(cfg)
	if err != nil {
		t.Fatal(err)
	}
	second, err := Run(cfg)
	if err != nil {
		t.Fatal(err)
	}
	a, err := json.Marshal(first)
	if err != nil {
		t.Fatal(err)
	}
	b, err := json.Marshal(second)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(a, b) {
		t.Fatalf("same seed/config produced different fingerprints\n%s\n%s", a, b)
	}
}

func TestPipelineCommandWithStrictIVTFFFixture(t *testing.T) {
	dir := fixtureDir(t)
	corpusPath := writeFixture(t, dir, "primary.txt", "caa cab\ndaa dab\nfaa fab\n")
	controlPath := writeFixture(t, dir, "control.txt", "abc abd\nijk ijl\nmno mnp\n")
	ivtffPath := writeFixture(t, dir, "primary.ivtff", "<f1> <! $C=A >\n<f1.1,@P0> caa cab\n<f1.2,@P0> daa dab\n<f1.3,@P0> faa fab\n")
	outputDir := filepath.Join(dir, "out")
	absolute := func(path string) string {
		v, err := filepath.Abs(path)
		if err != nil {
			t.Fatal(err)
		}
		return v
	}
	configPath := writeFixture(t, dir, "analysis.yaml", fmt.Sprintf(`version: fingerprint-v2-lexical-paradigms-v1
output_dir: %q
seed: 19
repetitions: 3
min_rule_support: 2
graph_swaps: 2
grammar:
  modes: [structure-preserving, frequency-aware]
primary:
  id: fixture-primary
  path: %q
  glyph_mode: natural
  ivtff_path: %q
controls:
  - name: fixture-control
    corpus:
      path: %q
      glyph_mode: natural
`, absolute(outputDir), absolute(corpusPath), absolute(ivtffPath), absolute(controlPath)))
	repo, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		t.Fatal(err)
	}
	command := exec.Command("go", "run", "./cmd/fingerprint-v2-analyze", "-config", absolute(configPath))
	command.Dir = repo
	if output, err := command.CombinedOutput(); err != nil {
		t.Fatalf("command failed: %v\n%s", err, output)
	}
	for _, name := range []string{"fingerprint.json", "raw_results.json", "config.yaml", "warnings.json", "errors.json", "report.md"} {
		if _, err := os.Stat(filepath.Join(outputDir, name)); err != nil {
			t.Fatalf("missing %s: %v", name, err)
		}
	}
	fingerprint, err := os.ReadFile(filepath.Join(outputDir, "fingerprint.json"))
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Contains(fingerprint, []byte(`"metadata_alignment": "strict IVTFF aligned"`)) {
		t.Fatalf("strict alignment was not recorded:\n%s", fingerprint)
	}
	for _, key := range []string{`"ef5"`, `"edit_graph_validation"`, `"cross_scale"`, `"graph_representations"`, `"null_registry"`, `"EDIT_CROSS_SCALE_BLOCK_READY"`} {
		if !bytes.Contains(fingerprint, []byte(key)) {
			t.Fatalf("task77 block missing key %s:\n%s", key, fingerprint)
		}
	}
}

func TestRunFileRejectsDeclaredInvalidNumericParameters(t *testing.T) {
	dir := fixtureDir(t)
	path := writeFixture(t, dir, "invalid.yaml", "output_dir: output\nprimary:\n  path: corpus.txt\nrepetitions: -1\n")
	if _, err := RunFile(path); err == nil {
		t.Fatal("RunFile accepted a declared negative repetitions value")
	}
}

func TestSingleTypeCorpusIsInsufficientSupport(t *testing.T) {
	dir := fixtureDir(t)
	path := writeFixture(t, dir, "single.txt", "abc abc\n")
	cfg := testConfig(path)
	result, err := Run(cfg)
	if err != nil {
		t.Fatal(err)
	}
	if result.Primary.Metrics.EF4.Verdict != "INSUFFICIENT_SUPPORT" {
		t.Fatalf("single-type corpus should not produce a grammar verdict: %+v", result.Primary.Metrics.EF4)
	}
	for _, verdict := range verdicts(result) {
		if verdict.ID == "EDIT_NEIGHBORHOODS_EXCEED_GRAMMAR_NULL" && verdict.Value != "INCONCLUSIVE" {
			t.Fatalf("single-type corpus reported a substantive edit-neighborhood verdict: %+v", verdict)
		}
	}
}
