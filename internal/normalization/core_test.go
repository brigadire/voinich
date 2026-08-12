package normalization

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestCompleteLinkThresholdAndChaining(t *testing.T) {
	corpus := testCorpus(map[string]int{"A": 20, "B": 20, "C": 20})
	structural := structuralWith(
		candidate("A", "B", .90),
		candidate("B", "C", .85),
		candidate("A", "C", .70),
	)
	models, _, err := BuildModels(corpus, structural, Config{Thresholds: []float64{.8}, MinTokenCount: 10})
	if err != nil {
		t.Fatal(err)
	}
	multi := multiClasses(models[0])
	if len(multi) != 1 || !memberTokensEqual(multi[0], []string{"A", "B"}) {
		t.Fatalf("chaining produced wrong classes: %+v", multi)
	}
	if models[0].Stats.SingletonTokens != 1 {
		t.Fatalf("singleton count = %d, want 1", models[0].Stats.SingletonTokens)
	}
}

func TestBelowThresholdAndThreeCompatibleTokens(t *testing.T) {
	corpus := testCorpus(map[string]int{"A": 12, "B": 13, "C": 14})
	below, _, err := BuildModels(corpus, structuralWith(candidate("A", "B", .79)), Config{Thresholds: []float64{.8}, MinTokenCount: 10})
	if err != nil {
		t.Fatal(err)
	}
	if len(multiClasses(below[0])) != 0 {
		t.Fatal("below-threshold pair was merged")
	}
	all := structuralWith(candidate("A", "B", .9), candidate("A", "C", .85), candidate("B", "C", .88))
	models, _, err := BuildModels(corpus, all, Config{Thresholds: []float64{.8}, MinTokenCount: 10})
	if err != nil {
		t.Fatal(err)
	}
	multi := multiClasses(models[0])
	if len(multi) != 1 || !memberTokensEqual(multi[0], []string{"A", "B", "C"}) {
		t.Fatalf("fully compatible triple was not merged: %+v", multi)
	}
}

func TestMultipleThresholdsAndDeterministicIDs(t *testing.T) {
	corpus := testCorpus(map[string]int{"z": 20, "a": 20, "b": 20})
	structural := structuralWith(candidate("a", "b", .82), candidate("b", "z", .92), candidate("a", "z", .81))
	config := Config{Thresholds: []float64{.8, .9}, MinTokenCount: 10}
	first, _, err := BuildModels(corpus, structural, config)
	if err != nil {
		t.Fatal(err)
	}
	second, _, _ := BuildModels(corpus, structural, config)
	if !reflect.DeepEqual(first, second) {
		t.Fatal("class construction is not deterministic")
	}
	if len(first) != 2 || first[0].Label != "080" || first[1].Label != "090" {
		t.Fatalf("unexpected thresholds: %+v", first)
	}
	for _, model := range first {
		for index, class := range model.Classes {
			if want := fmt.Sprintf("C%04d", index+1); class.ID != want {
				t.Fatalf("class ID = %s, want %s", class.ID, want)
			}
		}
	}
}

func TestNormalizationPreservesCorpusAndSingletonModes(t *testing.T) {
	corpus := Corpus{
		Lines: [][]string{{"A", "B"}, {}, {"C"}}, Counts: map[string]int{"A": 1, "B": 1, "C": 1},
		Occurrences: 3, NonEmpty: 2, Transitions: 1,
	}
	model := Model{Classes: []Class{
		{ID: "C0001", Size: 2, Members: []Member{{Token: "A", Count: 1}, {Token: "B", Count: 1}}},
		{ID: "C0002", Size: 1, Members: []Member{{Token: "C", Count: 1}}},
	}}
	path := filepath.Join(t.TempDir(), "normalized.txt")
	if err := WriteNormalized(path, corpus, Mapping(model, "preserve")); err != nil {
		t.Fatal(err)
	}
	data, _ := os.ReadFile(path)
	if string(data) != "C0001 C0001\n\nC\n" {
		t.Fatalf("preserve output = %q", data)
	}
	loaded, err := LoadCorpus(path)
	if err != nil {
		t.Fatal(err)
	}
	if loaded.Occurrences != corpus.Occurrences || loaded.NonEmpty != corpus.NonEmpty || loaded.Transitions != corpus.Transitions || len(loaded.Lines) != len(corpus.Lines) {
		t.Fatalf("normalization changed corpus shape: %+v", loaded)
	}
	if err := WriteNormalized(path, corpus, Mapping(model, "class")); err != nil {
		t.Fatal(err)
	}
	data, _ = os.ReadFile(path)
	if string(data) != "C0001 C0001\n\nC0002\n" {
		t.Fatalf("class output = %q", data)
	}
}

func TestRandomModelSizeBinsAndSeed(t *testing.T) {
	counts := map[string]int{"a": 10, "b": 12, "c": 20, "d": 24, "e": 40, "f": 48}
	corpus := testCorpus(counts)
	structural := Model{Threshold: .8, Label: "080", Classes: []Class{
		{Size: 2, Members: []Member{{Token: "a", Count: 10}, {Token: "b", Count: 12}}},
		{Size: 3, Members: []Member{{Token: "c", Count: 20}, {Token: "d", Count: 24}, {Token: "e", Count: 40}}},
		{Size: 1, Members: []Member{{Token: "f", Count: 48}}},
	}}
	first := RandomModel(structural, corpus, 10, 1, 0)
	second := RandomModel(structural, corpus, 10, 1, 0)
	if !reflect.DeepEqual(first, second) {
		t.Fatal("random model is not deterministic for a fixed seed")
	}
	wantSizes := []int{2, 3}
	var gotSizes []int
	for _, class := range first.Classes {
		if class.Size > 1 {
			gotSizes = append(gotSizes, class.Size)
			for _, member := range class.Members {
				if frequencyBin(member.Count) < 3 || frequencyBin(member.Count) > 5 {
					t.Fatalf("unexpected frequency bin for %+v", member)
				}
			}
		}
	}
	sortInts(gotSizes)
	if !reflect.DeepEqual(gotSizes, wantSizes) {
		t.Fatalf("random class sizes = %v, want %v", gotSizes, wantSizes)
	}
}

func candidate(a, b string, similarity float64) Candidate {
	return Candidate{TokenA: a, TokenB: b, Similarity: similarity, PositionSimilarity: similarity, LeftContextSimilarity: similarity, RightContextSimilarity: similarity}
}

func structuralWith(candidates ...Candidate) StructuralInput {
	return StructuralInput{EquivalenceCandidates: candidates}
}

func testCorpus(counts map[string]int) Corpus {
	occurrences := 0
	for _, count := range counts {
		occurrences += count
	}
	return Corpus{Counts: counts, Occurrences: occurrences}
}

func multiClasses(model Model) []Class {
	var result []Class
	for _, class := range model.Classes {
		if class.Size > 1 {
			result = append(result, class)
		}
	}
	return result
}

func memberTokensEqual(class Class, want []string) bool {
	var got []string
	for _, member := range class.Members {
		got = append(got, member.Token)
	}
	return reflect.DeepEqual(got, want)
}

func sortInts(values []int) {
	for i := range values {
		for j := i + 1; j < len(values); j++ {
			if values[j] < values[i] {
				values[i], values[j] = values[j], values[i]
			}
		}
	}
}
