package main

import (
	"math"
	"os"
	"path/filepath"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestLoadDatasetAndBuildOutput(t *testing.T) {
	dictionary := []DictionaryToken{
		{
			Token: "A", Count: 3,
			PositionInString: []Position{{Position: 0, Count: 2}, {Position: 1, Count: 1}},
			WordBefore:       []Neighbor{{Token: "A", Count: 1}},
			WordAfter:        []Neighbor{{Token: "A", Count: 1}, {Token: "B", Count: 1}},
			LineStartCount:   2, LineEndCount: 1,
		},
		{
			Token: "B", Count: 1,
			PositionInString: []Position{{Position: 1, Count: 1}},
			WordBefore:       []Neighbor{{Token: "A", Count: 1}},
			LineEndCount:     1,
		},
	}
	analyses := []TokenAnalysisInput{
		{Token: "A", Count: 3, StartProbability: 2.0 / 3.0, EndProbability: 1.0 / 3.0, Left: EnvironmentAnalysis{Unique: 1}, Right: EnvironmentAnalysis{Unique: 2, Entropy: 1}},
		{Token: "B", Count: 1, EndProbability: 1, Left: EnvironmentAnalysis{Unique: 1}},
	}

	dataset := loadTestDataset(t, dictionary, analyses)
	if dataset.Meta.TokenOccurrences != 4 || dataset.Meta.Lines != 2 || dataset.Meta.Transitions != 2 {
		t.Fatalf("unexpected meta: %+v", dataset.Meta)
	}
	parameters := Parameters{
		MinTokenCountForRanking:  1,
		MinTransitionCount:       1,
		MinContextObservations:   1,
		MinSelfTransitionCount:   1,
		ReliabilityPriorCount:    1,
		MinEquivalenceSimilarity: 0,
		DominantContextLimit:     3,
	}
	output := buildOutput(dataset, parameters)
	if len(output.SignificantTransitions) != 2 {
		t.Fatalf("significant transitions = %d, want 2", len(output.SignificantTransitions))
	}
	transition := findTransition(t, output.SignificantTransitions, "A", "B")
	assertClose(t, "P(B|A)", transition.Probability, 0.5)
	assertClose(t, "expected A->B", transition.Expected, 1)
	assertClose(t, "PMI A->B", transition.PMI, 0)
	if len(output.SelfTransitions) != 1 || output.SelfTransitions[0].Token != "A" {
		t.Fatalf("unexpected self transitions: %+v", output.SelfTransitions)
	}
	assertClose(t, "self expected", output.SelfTransitions[0].Expected, 1)
	assertClose(t, "self enrichment", output.SelfTransitions[0].Enrichment, 1)
	assertClose(t, "self reliability", output.SelfTransitions[0].Reliability, 0.5)
	assertClose(t, "self ranking score", output.SelfTransitions[0].RankingScore, 0.5)
	if len(output.EquivalenceCandidates) != 1 {
		t.Fatalf("equivalence candidates = %d, want 1", len(output.EquivalenceCandidates))
	}
	if output.PositionalSpecialization[0].Reliability >= 1 {
		t.Fatalf("reliability must shrink ranking: %+v", output.PositionalSpecialization[0])
	}
}

func TestSelfTransitionsUseReliabilityForRanking(t *testing.T) {
	dataset := &Dataset{
		Tokens:   []DictionaryToken{{Token: "low", Count: 10}, {Token: "supported", Count: 20}},
		Right:    map[string]map[string]int{"low": {"low": 1}, "supported": {"supported": 4}},
		Outgoing: map[string]int{"low": 10, "supported": 20}, Incoming: map[string]int{"low": 10, "supported": 20},
		Meta: Meta{Transitions: 100},
	}
	result := selfTransitionRanking(dataset, Parameters{MinTokenCountForRanking: 1, MinSelfTransitionCount: 1, ReliabilityPriorCount: 10})
	if len(result) != 2 || result[0].Token != "supported" {
		t.Fatalf("self-transition reliability ranking = %+v", result)
	}
}

func TestEquivalenceCandidatesHaveIndependentLimit(t *testing.T) {
	dataset := &Dataset{
		Tokens:    []DictionaryToken{{Token: "A", Count: 10}, {Token: "B", Count: 10}, {Token: "C", Count: 10}, {Token: "D", Count: 10}},
		Positions: map[string]map[int]int{"A": {0: 10}, "B": {0: 10}, "C": {0: 10}, "D": {0: 10}},
		Left:      map[string]map[string]int{}, Right: map[string]map[string]int{},
	}
	parameters := Parameters{MinTokenCountForRanking: 10, MinEquivalenceSimilarity: 0, MaxItemsPerSection: 1}
	if got := len(equivalenceRanking(dataset, parameters)); got != 6 {
		t.Fatalf("unlimited equivalence candidates = %d, want 6", got)
	}
	parameters.MaxEquivalenceCandidates = 2
	if got := len(equivalenceRanking(dataset, parameters)); got != 2 {
		t.Fatalf("limited equivalence candidates = %d, want 2", got)
	}
}

func TestRankingObservationThresholds(t *testing.T) {
	dataset := &Dataset{
		Tokens: []DictionaryToken{
			{Token: "sparse", Count: 20},
			{Token: "supported", Count: 20},
		},
		Left: map[string]map[string]int{
			"sparse":    {"x": 1},
			"supported": {"x": 10},
		},
		Right: map[string]map[string]int{
			"sparse":    {"sparse": 1},
			"supported": {"supported": 3, "x": 7},
		},
		Outgoing: map[string]int{"sparse": 1, "supported": 10},
		Incoming: map[string]int{"sparse": 1, "supported": 10},
		Meta:     Meta{Transitions: 11},
	}
	parameters := Parameters{
		MinTokenCountForRanking: 10,
		MinContextObservations:  10,
		MinSelfTransitionCount:  3,
		ReliabilityPriorCount:   10,
	}

	successors := predictabilityRanking(dataset, parameters, true)
	if len(successors) != 1 || successors[0].Token != "supported" {
		t.Fatalf("successor predictability = %+v, want only supported", successors)
	}
	predecessors := predictabilityRanking(dataset, parameters, false)
	if len(predecessors) != 1 || predecessors[0].Token != "supported" {
		t.Fatalf("predecessor predictability = %+v, want only supported", predecessors)
	}
	self := selfTransitionRanking(dataset, parameters)
	if len(self) != 1 || self[0].Token != "supported" || self[0].Count != 3 {
		t.Fatalf("self transitions = %+v, want only supported", self)
	}
}

func TestLoadDatasetRejectsStaleEndProbability(t *testing.T) {
	dictionary := []DictionaryToken{{Token: "A", Count: 1, PositionInString: []Position{{Position: 0, Count: 1}}, LineStartCount: 1, LineEndCount: 1}}
	analyses := []TokenAnalysisInput{{Token: "A", Count: 1, StartProbability: 1, EndProbability: 0}}

	dictionaryPath := writeYAML(t, "dictionary.yaml", dictionary)
	analysisPath := writeYAML(t, "tokens_analysis.yaml", analyses)
	if _, err := loadDataset(dictionaryPath, analysisPath); err == nil {
		t.Fatal("loadDataset returned nil error for stale end_probability")
	}
}

func TestMetrics(t *testing.T) {
	assertClose(t, "identical position similarity", 1-positionJSD(map[int]int{0: 2, 1: 1}, map[int]int{0: 4, 1: 2}), 1)
	assertClose(t, "disjoint position divergence", positionJSD(map[int]int{0: 1}, map[int]int{1: 1}), 1)
	assertClose(t, "identical cosine", cosineSimilarity(map[string]int{"x": 2, "y": 1}, map[string]int{"x": 4, "y": 2}), 1)
	assertClose(t, "orthogonal cosine", cosineSimilarity(map[string]int{"x": 1}, map[string]int{"y": 1}), 0)
	if value := logLikelihood2x2(10, 20, 30, 100); value <= 0 {
		t.Fatalf("log likelihood = %v, want positive", value)
	}
}

func loadTestDataset(t *testing.T, dictionary []DictionaryToken, analyses []TokenAnalysisInput) *Dataset {
	t.Helper()
	dictionaryPath := writeYAML(t, "dictionary.yaml", dictionary)
	analysisPath := writeYAML(t, "tokens_analysis.yaml", analyses)
	dataset, err := loadDataset(dictionaryPath, analysisPath)
	if err != nil {
		t.Fatalf("loadDataset: %v", err)
	}
	return dataset
}

func writeYAML(t *testing.T, name string, value any) string {
	t.Helper()
	data, err := yaml.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(t.TempDir(), name)
	if err := os.WriteFile(path, data, 0o600); err != nil {
		t.Fatal(err)
	}
	return path
}

func findTransition(t *testing.T, transitions []SignificantTransition, from, to string) SignificantTransition {
	t.Helper()
	for _, transition := range transitions {
		if transition.From == from && transition.To == to {
			return transition
		}
	}
	t.Fatalf("transition %s -> %s not found", from, to)
	return SignificantTransition{}
}

func assertClose(t *testing.T, name string, got, want float64) {
	t.Helper()
	if math.Abs(got-want) > 1e-12 {
		t.Fatalf("%s = %.15g, want %.15g", name, got, want)
	}
}
