package beginendanalyze

import (
	"math"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

func testConfig(dictionary, corpusPath string, p Parameters) Config {
	return Config{
		DictionaryPath: dictionary, CorpusPath: corpusPath,
		MaxWindow: p.MaxWindow, Permutations: p.Permutations, MinTokenCount: p.MinTokenCount,
		RandomSeed: p.RandomSeed, PermutationMode: p.PermutationMode, IncludeUnclear: p.IncludeUnclear,
		MaxCandidates: p.MaxCandidates,
	}
}

func TestAnalysisFindsDirectedNonLocalPairDeterministically(t *testing.T) {
	corpusText := strings.Join([]string{
		"A x x B",
		"A y y B",
		"A x y B",
		"A y x B",
		"",
		"B x y A",
	}, "\n")
	dictionary, corpusPath := testInputs(t, corpusText)
	parameters := Parameters{MaxWindow: 5, Permutations: 20, MinTokenCount: 2, RandomSeed: 7, PermutationMode: "line", MaxCandidates: 20}
	first, err := RunAndWrite(testConfig(dictionary, corpusPath, parameters))
	if err != nil {
		t.Fatal(err)
	}
	second, err := RunAndWrite(testConfig(dictionary, corpusPath, parameters))
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(first, second) {
		t.Fatal("fixed seed did not produce deterministic output")
	}
	candidate := findCandidate(t, first.Candidates, "A", "B")
	if candidate.WithinLine.Observations != 4 || candidate.WithinLine.Histogram[3] != 4 {
		t.Fatalf("unexpected line distance: %+v", candidate.WithinLine)
	}
	if candidate.Directionality.Scope != "page" {
		t.Fatalf("known page boundaries should select page scope: %+v", candidate.Directionality)
	}
	if first.Meta.Pages != 2 || !first.Meta.PageBoundariesKnown {
		t.Fatalf("page metadata: %+v", first.Meta)
	}
	if len(candidate.WithinLine.Windows) != len(parametersForWindows(5))+1 {
		t.Fatalf("windows: %+v", candidate.WithinLine.Windows)
	}
}

func TestAnalysisMatchesAcrossExecutorWorkerCounts(t *testing.T) {
	corpusText := strings.Join([]string{
		"A x x B", "A y y B", "A x y B", "A y x B", "",
		"B x y A", "A x B y", "B y A x", "",
		"A x x x B", "A y y y B",
	}, "\n")
	dictionary, corpusPath := testInputs(t, corpusText)
	parameters := Parameters{MaxWindow: 5, Permutations: 20, MinTokenCount: 2, RandomSeed: 7, PermutationMode: "line", MaxCandidates: 50}
	base := testConfig(dictionary, corpusPath, parameters)

	sequential, err := RunAndWrite(base)
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		name      string
		workers   int
		batchSize int
	}{
		{"single-batch", 1, 10_000},
		{"tiny-batches-1-worker", 1, 1},
		{"tiny-batches-4-workers", 4, 1},
		{"medium-batches-3-workers", 3, 3},
	} {
		t.Run(tc.name, func(t *testing.T) {
			c := base
			c.Workers = tc.workers
			c.CandidateBatchSize = tc.batchSize
			got, err := RunAndWrite(c)
			if err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(got, sequential) {
				t.Fatalf("executor/batch-size variant %q differs from sequential oracle", tc.name)
			}
		})
	}
}

func TestDirectionalityComparesBothOrders(t *testing.T) {
	result := directionality(0.8, 0.2, "page")
	if math.Abs(result.Score-0.6) > 1e-12 || result.LogRatio <= 0 {
		t.Fatalf("unexpected directionality: %+v", result)
	}
}

func TestQuestionMarkExcludedAndAtTokenPreserved(t *testing.T) {
	dictionary, corpusPath := testInputs(t, "A @192; q?\nA @192; q?\n")
	output, err := RunAndWrite(testConfig(dictionary, corpusPath, Parameters{MaxWindow: 3, MinTokenCount: 2, PermutationMode: "page", MaxCandidates: 20}))
	if err != nil {
		t.Fatal(err)
	}
	if output.Meta.UnclearExcluded != 1 {
		t.Fatalf("unclear excluded=%d", output.Meta.UnclearExcluded)
	}
	found := false
	for _, c := range append(output.Candidates, output.LikelyLocalPairs...) {
		if c.BeginCandidate == "A" && c.EndCandidate == "@192;" {
			found = true
		}
	}
	if !found {
		t.Fatal("@192; pair was not preserved")
	}
}

func TestReportsContainRequiredSections(t *testing.T) {
	dictionary, corpusPath := testInputs(t, "A x B\nA y B\n")
	output, err := RunAndWrite(testConfig(dictionary, corpusPath, Parameters{MaxWindow: 3, MinTokenCount: 2, PermutationMode: "page", MaxCandidates: 20}))
	if err != nil {
		t.Fatal(err)
	}
	directory := t.TempDir()
	if err := WriteReports(directory, output); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"begin_end_candidates.yaml", "begin_end_top.tsv", "begin_end_report.md"} {
		if info, err := os.Stat(filepath.Join(directory, name)); err != nil || info.Size() == 0 {
			t.Fatalf("report %s missing or empty: %v", name, err)
		}
	}
	report, err := os.ReadFile(filepath.Join(directory, "begin_end_report.md"))
	if err != nil {
		t.Fatal(err)
	}
	for _, section := range []string{"Best pairs within a line", "Best pairs within a page", "Strongest directionality", "Most expressed page balance", "nesting-like"} {
		if !strings.Contains(string(report), section) {
			t.Fatalf("report lacks %q", section)
		}
	}
}

func testInputs(t *testing.T, text string) (string, string) {
	t.Helper()
	dir := t.TempDir()
	corpusPath := filepath.Join(dir, "corpus.txt")
	if err := os.WriteFile(corpusPath, []byte(text), 0o600); err != nil {
		t.Fatal(err)
	}
	c, err := loadCorpus(corpusPath)
	if err != nil {
		t.Fatal(err)
	}
	counts := make(map[string]*DictionaryToken)
	for _, line := range c.Lines {
		for position, token := range line.Tokens {
			item := counts[token]
			if item == nil {
				item = &DictionaryToken{Token: token}
				counts[token] = item
			}
			item.Count++
			found := false
			for i := range item.PositionInString {
				if item.PositionInString[i].Position == position {
					item.PositionInString[i].Count++
					found = true
				}
			}
			if !found {
				item.PositionInString = append(item.PositionInString, PositionInput{Position: position, Count: 1})
			}
			if position == 0 {
				item.LineStartCount++
			}
			if position == len(line.Tokens)-1 {
				item.LineEndCount++
			}
			if position > 0 {
				addNeighbor(&item.WordBefore, line.Tokens[position-1])
			}
			if position+1 < len(line.Tokens) {
				addNeighbor(&item.WordAfter, line.Tokens[position+1])
			}
		}
	}
	dictionary := make([]DictionaryToken, 0, len(counts))
	for _, item := range counts {
		dictionary = append(dictionary, *item)
	}
	data, err := yaml.Marshal(dictionary)
	if err != nil {
		t.Fatal(err)
	}
	dictionaryPath := filepath.Join(dir, "dictionary.yaml")
	if err := os.WriteFile(dictionaryPath, data, 0o600); err != nil {
		t.Fatal(err)
	}
	return dictionaryPath, corpusPath
}

func addNeighbor(items *[]NeighborInput, token string) {
	for i := range *items {
		if (*items)[i].Token == token {
			(*items)[i].Count++
			return
		}
	}
	*items = append(*items, NeighborInput{Token: token, Count: 1})
}
func findCandidate(t *testing.T, items []Candidate, a, b string) Candidate {
	t.Helper()
	for _, item := range items {
		if item.BeginCandidate == a && item.EndCandidate == b {
			return item
		}
	}
	t.Fatalf("candidate %s -> %s not found", a, b)
	return Candidate{}
}
func parametersForWindows(max int) []int { return configuredWindows(max) }
