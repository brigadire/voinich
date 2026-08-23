package main

import (
	"math"
	"os"
	"path/filepath"
	"testing"
)

func TestAnalyzeTokens(t *testing.T) {
	tokens := []Tokens{
		{
			Token: "A", Count: 4,
			PositionInString: []Position{{Position: 0, Count: 2}, {Position: 1, Count: 2}},
			WordBefore:       []Token{{Token: "A", Count: 1}, {Token: "B", Count: 1}},
			WordAfter:        []Token{{Token: "B", Count: 2}, {Token: "A", Count: 1}},
			LineStartCount:   2, LineEndCount: 1,
		},
		{
			Token: "B", Count: 3,
			PositionInString: []Position{{Position: 0, Count: 1}, {Position: 1, Count: 2}},
			WordBefore:       []Token{{Token: "A", Count: 2}},
			WordAfter:        []Token{{Token: "A", Count: 1}},
			LineStartCount:   1, LineEndCount: 2,
		},
	}
	before := map[string]map[string]int{
		"A": {"A": 1, "B": 1},
		"B": {"A": 2},
	}
	after := map[string]map[string]int{
		"A": {"A": 1, "B": 2},
		"B": {"A": 1},
	}

	analyses := analyzeTokens(tokens, before, after)
	a := analyses[0]
	assertClose(t, "start probability", a.StartProbability, 0.5)
	assertClose(t, "end probability", a.EndProbability, 0.25)
	if a.Left.Unique != 2 || a.Right.Unique != 2 {
		t.Fatalf("unique neighbors = left %d, right %d; want 2, 2", a.Left.Unique, a.Right.Unique)
	}
	if a.SelfTransition.Count != 1 {
		t.Fatalf("self-transition count = %d, want 1", a.SelfTransition.Count)
	}
	assertClose(t, "self-transition probability", a.SelfTransition.Probability, 1.0/3.0)

	toB := findTransition(t, a.Transitions, "B")
	assertClose(t, "P(B|A)", toB.Probability, 2.0/3.0)
	assertClose(t, "P(A|B)", toB.ReverseProbability, 1)
	assertClose(t, "A/B asymmetry", toB.Asymmetry, -0.2)
	if a.StructuralScores.PositionalSpecialization <= 0 || a.StructuralScores.PositionalSpecialization > 1 {
		t.Fatalf("positional specialization = %v, want value in (0, 1]", a.StructuralScores.PositionalSpecialization)
	}
}

func TestCountUniqueAndEntropy(t *testing.T) {
	unique, entropy := countUniqueAndEntropy(map[string]int{"x": 2, "y": 2, "zero": 0})
	if unique != 2 {
		t.Fatalf("unique = %d, want 2", unique)
	}
	assertClose(t, "entropy", entropy, 1)
}

func TestRestrictionUsesUniqueOutcomesNotObservationCount(t *testing.T) {
	if got := calculateRestriction(1, 2); got != 0 {
		t.Fatalf("restriction for uniform binary distribution = %v, want 0", got)
	}
	if got := calculateRestriction(0, 1); got != 1 {
		t.Fatalf("restriction for deterministic distribution = %v, want 1", got)
	}
	counts := map[string]int{"x": 500, "y": 500}
	unique, entropy := countUniqueAndEntropy(counts)
	if got := calculateRestriction(entropy, unique); got != 0 {
		t.Fatalf("restriction changed with sample size: %v", got)
	}
}

func TestPositionalSpecializationKnownValues(t *testing.T) {
	assertClose(t, "identical positions", positionalSpecialization([]Position{{Position: 0, Count: 2}, {Position: 1, Count: 2}}, map[int]int{0: 10, 1: 10}), 0)
	assertClose(t, "disjoint positions", positionalSpecialization([]Position{{Position: 0, Count: 2}}, map[int]int{1: 10}), 1)
}

func TestValidateTokenRejectsInvalidCountsAndNeighbors(t *testing.T) {
	tests := []Tokens{
		{Token: "x", Count: -1},
		{Token: "x", Count: 1, LineStartCount: 2},
		{Token: "x", Count: 1, LineEndCount: 2},
		{Token: "x", Count: 1, WordAfter: []Token{{Token: "", Count: 1}}},
		{Token: "x", Count: 1, WordBefore: []Token{{Token: "y", Count: -1}}},
	}
	for _, item := range tests {
		if err := validateToken(item); err == nil {
			t.Fatalf("validateToken(%+v) returned nil error", item)
		}
	}
}

func TestReadFileTokenRejectsDuplicateToken(t *testing.T) {
	fileName := filepath.Join(t.TempDir(), "dictionary.yaml")
	data := []byte("- token: a\n  count: 1\n- token: a\n  count: 1\n")
	if err := os.WriteFile(fileName, data, 0o600); err != nil {
		t.Fatal(err)
	}
	if _, _, _, err := readFileToken(fileName); err == nil {
		t.Fatal("readFileToken returned nil error for duplicate token")
	}
}

func findTransition(t *testing.T, transitions []TransitionAnalysis, token string) TransitionAnalysis {
	t.Helper()
	for _, transition := range transitions {
		if transition.Token == token {
			return transition
		}
	}
	t.Fatalf("transition to %q not found", token)
	return TransitionAnalysis{}
}

func assertClose(t *testing.T, name string, got, want float64) {
	t.Helper()
	if math.Abs(got-want) > 1e-12 {
		t.Fatalf("%s = %.15g, want %.15g", name, got, want)
	}
}
