package profilestability

import (
	"math"
	"reflect"
	"testing"

	"zcore.dev/voinich/internal/normalization"
	"zcore.dev/voinich/internal/validation"
)

func TestStableProfilesAndEligibility(t *testing.T) {
	corpus := corpusFromLines([][]string{{"L", "A", "R"}, {"L", "B", "R"}, {"L", "A", "R"}, {"L", "B", "R"}})
	profiles := BuildProfiles(corpus)
	components := Compare(profiles["A"], profiles["B"])
	assertClose(t, "stable combined", components.Similarity, 1)
	if got := Eligible(profiles, 2); !reflect.DeepEqual(got, []string{"A", "B", "L", "R"}) {
		t.Fatalf("eligible=%v", got)
	}
}

func TestNoCrossFoldContaminationAndDeterministicAssignment(t *testing.T) {
	corpus := corpusFromLines([][]string{{"A"}, {"A"}, {"TEST"}, {"B"}, {"B"}})
	first, err := validation.SplitFolds(corpus.Lines, 2, 7)
	if err != nil {
		t.Fatal(err)
	}
	second, _ := validation.SplitFolds(corpus.Lines, 2, 7)
	if !reflect.DeepEqual(first, second) {
		t.Fatal("fold assignment is not deterministic")
	}
	train, test, err := validation.Partition(corpus, []int{2})
	if err != nil {
		t.Fatal(err)
	}
	if _, exists := BuildProfiles(train)["TEST"]; exists {
		t.Fatal("TEST-only token contaminated TRAIN")
	}
	if BuildProfiles(test)["TEST"].Count != 1 {
		t.Fatal("TEST profile missing")
	}
}

func TestSelfProfileComparison(t *testing.T) {
	left := Profile{Positions: map[int]int{1: 10}, Left: map[string]int{"L": 10}, Right: map[string]int{"R": 10}}
	right := Profile{Positions: map[int]int{1: 5}, Left: map[string]int{"L": 5}, Right: map[string]int{"R": 5}}
	got := Compare(left, right)
	assertClose(t, "position", got.PositionSimilarity, 1)
	assertClose(t, "left", got.LeftSimilarity, 1)
	assertClose(t, "right", got.RightSimilarity, 1)
}

func TestThresholdNoiseAndUnstableGeometry(t *testing.T) {
	folds := []FoldSimilarity{{Components: Components{Similarity: .69}}, {Components: Components{Similarity: .72}}, {Components: Components{Similarity: .71}}, {Components: Components{Similarity: .68}}, {Components: Components{Similarity: .73}}}
	cross := thresholdCrossings(folds, []float64{.70})[0]
	if cross.FoldsAboveThreshold != 3 || cross.FoldsBelowThreshold != 2 || cross.ThresholdCrossingCount != 3 {
		t.Fatalf("crossing=%+v", cross)
	}
	left := []Neighbor{{Token: "A"}, {Token: "B"}, {Token: "C"}}
	right := []Neighbor{{Token: "X"}, {Token: "Y"}, {Token: "Z"}}
	if Jaccard(left, right) != 0 {
		t.Fatal("disjoint geometry has nonzero Jaccard")
	}
}

func TestNearestNeighborAndTop1Recovery(t *testing.T) {
	profiles := map[string]Profile{
		"A": {Positions: map[int]int{0: 10}, Left: map[string]int{"L": 10}, Right: map[string]int{"R": 10}},
		"B": {Positions: map[int]int{0: 8}, Left: map[string]int{"L": 8}, Right: map[string]int{"R": 8}},
		"C": {Positions: map[int]int{1: 8}, Left: map[string]int{"X": 8}, Right: map[string]int{"Y": 8}},
	}
	items := NearestNeighbors(profiles, "A", []string{"A", "B", "C"}, 2)
	if items[0].Token != "B" || items[0].Rank != 1 {
		t.Fatalf("neighbors=%+v", items)
	}
	folds := []foldData{{trainEligible: map[string]bool{"A": true, "B": true, "C": true}, neighbors: map[string][]Neighbor{"A": items}}, {trainEligible: map[string]bool{"A": true, "B": true, "C": true}, neighbors: map[string][]Neighbor{"A": items}}}
	stability := buildNeighborStability("A", items, folds, 2)
	if stability.Top1SameFraction != 1 || stability.Top1RecoveryFraction != 1 {
		t.Fatalf("stability=%+v", stability)
	}
}

func TestSpearmanAndBootstrapPercentiles(t *testing.T) {
	left := []Neighbor{{Token: "A"}, {Token: "B"}, {Token: "C"}}
	right := []Neighbor{{Token: "A"}, {Token: "B"}, {Token: "C"}}
	rho, n, ok := spearmanCommon(left, right)
	if !ok || n != 3 || rho != 1 {
		t.Fatalf("rho=%v n=%d ok=%v", rho, n, ok)
	}
	d := Summarize([]float64{1, 2, 3, 4, 5}, true)
	if d.Percentile50 != 3 || d.Percentile025 >= d.Percentile975 {
		t.Fatalf("distribution=%+v", d)
	}
}

func TestDeterministicBootstrapAndFrequencyUncertainty(t *testing.T) {
	var lines [][]string
	for i := 0; i < 10; i++ {
		left, right := "L1", "R1"
		if i%2 == 1 {
			left, right = "L2", "R2"
		}
		lines = append(lines, []string{left, "A", right}, []string{left, "B", right})
	}
	for i := 0; i < 100; i++ {
		left, right := "L1", "R1"
		if i%2 == 1 {
			left, right = "L2", "R2"
		}
		lines = append(lines, []string{left, "F", right}, []string{left, "G", right})
	}
	corpus := corpusFromLines(lines)
	pairs := map[pairKey]bool{makePair("A", "B"): true, makePair("F", "G"): true}
	config := Config{BootstrapRuns: 80, BootstrapSeed: 11, MinTokenCount: 1, Threshold: .7}
	first := runBootstrap(corpus, pairs, config)
	second := runBootstrap(corpus, pairs, config)
	if !reflect.DeepEqual(first, second) {
		t.Fatal("bootstrap is not deterministic")
	}
	byPair := map[pairKey]BootstrapPair{}
	for _, item := range first {
		byPair[makePair(item.TokenA, item.TokenB)] = item
	}
	if byPair[makePair("A", "B")].Similarity.Stddev <= byPair[makePair("F", "G")].Similarity.Stddev {
		t.Fatalf("rare CI not wider: rare=%+v frequent=%+v", byPair[makePair("A", "B")].Similarity, byPair[makePair("F", "G")].Similarity)
	}
	if byPair[makePair("A", "B")].ProbabilityAboveThreshold < 0 || byPair[makePair("A", "B")].ProbabilityAboveThreshold > 1 {
		t.Fatal("invalid probability")
	}
}

func TestFrequencyBinsAndComponentAblationReferenceReport(t *testing.T) {
	for count, want := range map[int]string{10: "10-19", 20: "20-39", 40: "40-79", 80: "80-159", 160: "160+"} {
		if got := frequencyBin(count); got != want {
			t.Fatalf("bin(%d)=%s", count, got)
		}
	}
	model := normalization.Model{Classes: []normalization.Class{{ID: "C1", Size: 2, Members: []normalization.Member{{Token: "A", Count: 10}, {Token: "B", Count: 10}}}}}
	profiles := map[string]Profile{"A": {Positions: map[int]int{0: 10}, Left: map[string]int{"L": 10}, Right: map[string]int{"R": 10}}, "B": {Positions: map[int]int{0: 10}, Left: map[string]int{"L": 10}, Right: map[string]int{"X": 10}}}
	component := Compare(profiles["A"], profiles["B"])
	pair := PairStability{TokenA: "A", TokenB: "B", Full: component, Summary: Distribution{Mean: component.Similarity, Min: component.Similarity, Max: component.Similarity}}
	boot := BootstrapPair{TokenA: "A", TokenB: "B", Similarity: Distribution{Mean: component.Similarity, Percentile025: .5, Percentile975: .8}}
	classes, diagnostics := buildReferenceReports(model, map[pairKey]PairStability{makePair("A", "B"): pair}, map[pairKey]BootstrapPair{makePair("A", "B"): boot}, map[string]NeighborStability{}, profiles)
	if len(classes) != 1 || len(classes[0].Pairwise) != 1 || classes[0].WeakestPair == nil {
		t.Fatalf("classes=%+v", classes)
	}
	if len(diagnostics) != 1 {
		t.Fatalf("diagnostics=%+v", diagnostics)
	}
	wantWithoutRight := (component.PositionSimilarity + component.LeftSimilarity) / 2
	assertClose(t, "without right", diagnostics[0].SimilarityWithoutRight, wantWithoutRight)
}

func corpusFromLines(lines [][]string) validation.Corpus {
	result := validation.Corpus{Counts: make(map[string]int)}
	for i, tokens := range lines {
		line := validation.Line{ID: i + 1, Tokens: tokens}
		result.Lines = append(result.Lines, line)
		if len(tokens) > 0 {
			result.Occurrences += len(tokens)
			result.Transitions += len(tokens) - 1
			for _, token := range tokens {
				result.Counts[token]++
			}
		}
	}
	return result
}

func assertClose(t *testing.T, name string, got, want float64) {
	t.Helper()
	if math.Abs(got-want) > 1e-12 {
		t.Fatalf("%s=%v want %v", name, got, want)
	}
}
