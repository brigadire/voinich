package structuralreliability

import (
	"math"
	"math/rand"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"gopkg.in/yaml.v3"
	"zcore.dev/voinich/internal/normalization"
	"zcore.dev/voinich/internal/profilestability"
	"zcore.dev/voinich/internal/validation"
)

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

func assertClose(t *testing.T, name string, got, want, tolerance float64) {
	t.Helper()
	if math.Abs(got-want) > tolerance {
		t.Fatalf("%s=%v want %v (tolerance %v)", name, got, want, tolerance)
	}
}

// buildRepeatedCorpus repeats a token in a fixed structural context n times,
// interleaved so every fold sees a proportional share of every token.
func buildRepeatedCorpus(spec map[string][3]string, counts map[string]int) validation.Corpus {
	max := 0
	for _, count := range counts {
		if count > max {
			max = count
		}
	}
	var lines [][]string
	for i := 0; i < max; i++ {
		for token, count := range counts {
			if i >= count {
				continue
			}
			context := spec[token]
			lines = append(lines, []string{context[0], token, context[1]})
			_ = context[2]
		}
	}
	return corpusFromLines(lines)
}

// --- eligibility filtering + neighbors recalculated inside threshold ---

func TestNeighborsRecomputedWithinThreshold(t *testing.T) {
	spec := map[string][3]string{
		"A": {"L", "R", ""},
		"B": {"L", "R", ""},
		"C": {"L", "R", ""},
		"D": {"L", "R", ""},
	}
	corpus := buildRepeatedCorpus(spec, map[string]int{"A": 30, "B": 28, "C": 25, "D": 8})
	full := profilestability.BuildProfiles(corpus)
	folds, err := BuildFolds(corpus, 2, 1)
	if err != nil {
		t.Fatal(err)
	}

	fullWs := profilestability.PrecomputeAll(full)
	items, fullEligible, fullNeighbors := buildTokenMetrics(full, fullWs, folds, 10, 3)
	for _, token := range fullEligible {
		if token == "D" {
			t.Fatalf("D (count=8) must not be eligible at min-count=10")
		}
	}
	for token, neighbors := range fullNeighbors {
		for _, neighbor := range neighbors {
			if full[neighbor.Token].Count < 10 {
				t.Fatalf("neighbor %s of %s has count %d, below the threshold used to build this neighbor set", neighbor.Token, token, full[neighbor.Token].Count)
			}
			if neighbor.Token == "D" {
				t.Fatalf("ineligible token D leaked into neighbor geometry of %s", token)
			}
		}
	}

	lowItems, lowEligible, lowNeighbors := buildTokenMetrics(full, fullWs, folds, 5, 3)
	foundD := false
	for _, token := range lowEligible {
		if token == "D" {
			foundD = true
		}
	}
	if !foundD {
		t.Fatal("D (count=8) must be eligible at min-count=5")
	}
	_ = items
	_ = lowItems
	_ = lowNeighbors
}

// --- cumulative thresholds / independent frequency bins ---

func TestCumulativeThresholdsShrinkAsMinCountGrows(t *testing.T) {
	spec := map[string][3]string{"A": {"L", "R", ""}, "B": {"L", "R", ""}, "C": {"L", "R", ""}}
	corpus := buildRepeatedCorpus(spec, map[string]int{"A": 100, "B": 50, "C": 15})
	full := profilestability.BuildProfiles(corpus)
	folds, err := BuildFolds(corpus, 2, 1)
	if err != nil {
		t.Fatal(err)
	}
	fullWs := profilestability.PrecomputeAll(full)
	_, eligible10, _ := buildTokenMetrics(full, fullWs, folds, 10, 3)
	_, eligible60, _ := buildTokenMetrics(full, fullWs, folds, 60, 3)
	if len(eligible10) <= len(eligible60) {
		t.Fatalf("higher cumulative threshold must not increase eligible tokens: %d vs %d", len(eligible10), len(eligible60))
	}
	want := map[string]bool{"A": true, "L": true, "R": true}
	for _, token := range eligible60 {
		if !want[token] {
			t.Fatalf("only A, L and R have count>=60 (A=100, B=50, C=15, L=R=165), got eligible=%v", eligible60)
		}
	}
	if len(eligible60) != len(want) {
		t.Fatalf("expected exactly %v, got %v", want, eligible60)
	}
}

func TestBinLabelAndFilterPairs(t *testing.T) {
	upper20 := 20
	if got := binLabel(10, &upper20); got != "10-19" {
		t.Fatalf("binLabel(10,20)=%s", got)
	}
	if got := binLabel(320, nil); got != "320+" {
		t.Fatalf("binLabel(320,nil)=%s", got)
	}
	full := map[string]profilestability.Profile{"A": {Count: 15}, "B": {Count: 25}, "C": {Count: 5}}
	pairs := []pairKey{makePair("A", "B"), makePair("A", "C"), makePair("B", "C")}
	got := filterPairs(pairs, full, func(count int) bool { return count >= 10 })
	if len(got) != 1 || got[0] != makePair("A", "B") {
		t.Fatalf("filterPairs=%v", got)
	}
}

func TestAggregateTokenMetricsMeanOfMeans(t *testing.T) {
	items := []tokenMetric{
		{ContinuousTokenMetric: ContinuousTokenMetric{Token: "A", PositionTrainTrainStability: 0.8, TrainTrainObservations: 2}},
		{ContinuousTokenMetric: ContinuousTokenMetric{Token: "B", PositionTrainTrainStability: 0.6, TrainTrainObservations: 2}},
		{ContinuousTokenMetric: ContinuousTokenMetric{Token: "C", PositionTrainTrainStability: 0.9, TrainTrainObservations: 0}},
	}
	self, _, _ := aggregateTokenMetrics(items)
	if self.Position.Observations != 2 {
		t.Fatalf("expected zero-observation token C excluded, got observations=%d", self.Position.Observations)
	}
	assertClose(t, "mean of means", self.Position.Mean, 0.7, 1e-12)
}

// --- Spearman correlation ---

func TestSpearmanCorrelation(t *testing.T) {
	x := []float64{1, 2, 3, 4, 5}
	y := []float64{10, 20, 30, 40, 50}
	rho, n := Spearman(x, y)
	if n != 5 {
		t.Fatalf("n=%d", n)
	}
	assertClose(t, "perfect positive rho", rho, 1, 1e-12)

	yInverse := []float64{50, 40, 30, 20, 10}
	rho, _ = Spearman(x, yInverse)
	assertClose(t, "perfect negative rho", rho, -1, 1e-12)

	yTied := []float64{1, 1, 2, 2, 3}
	rho, n = Spearman(x, yTied)
	if n != 5 || rho < 0.9 {
		t.Fatalf("tied rho=%v n=%d, expected strong positive correlation", rho, n)
	}

	if _, n := Spearman([]float64{1, 2}, []float64{1, 2}); n != 2 {
		t.Fatalf("below minimum length should still report observation count, got n=%d", n)
	}
}

// --- pair min/geometric frequency ---

func TestPairMinAndGeometricFrequency(t *testing.T) {
	assertClose(t, "geometric mean", GeometricMean(9, 16), 12, 1e-12)
	if GeometricMean(0, 5) != 0 {
		t.Fatal("geometric mean with a zero count must be zero, not NaN")
	}

	folds := []foldProfiles{
		{trainProfiles: map[string]profilestability.Profile{"A": {Count: 12, Positions: map[int]int{0: 12}, Left: map[string]int{"L": 12}, Right: map[string]int{"R": 12}}, "B": {Count: 12, Positions: map[int]int{0: 12}, Left: map[string]int{"L": 12}, Right: map[string]int{"R": 12}}}},
		{trainProfiles: map[string]profilestability.Profile{"A": {Count: 3, Positions: map[int]int{0: 3}, Left: map[string]int{"L": 3}, Right: map[string]int{"R": 3}}, "B": {Count: 12, Positions: map[int]int{0: 12}, Left: map[string]int{"L": 12}, Right: map[string]int{"R": 12}}}},
	}
	pair := makePair("A", "B")
	values := pairFoldSimilarities([]pairKey{pair}, folds, 10)
	if len(values[pair]) != 1 {
		t.Fatalf("only the fold where both members meet min-count should contribute an observation, got %v", values[pair])
	}
}

// --- deterministic subsampling + occurrence-level subsampling ---

func TestOccurrenceLevelSamplingIsARealSubset(t *testing.T) {
	var occurrences []Occurrence
	for i := 0; i < 20; i++ {
		occurrences = append(occurrences, Occurrence{Position: i, HasLeft: true, Left: "L", HasRight: true, Right: "R"})
	}
	rng := rand.New(rand.NewSource(1))
	sample := SampleOccurrences(occurrences, 7, rng)
	if len(sample) != 7 {
		t.Fatalf("len(sample)=%d", len(sample))
	}
	seen := map[int]bool{}
	for _, occurrence := range sample {
		if seen[occurrence.Position] {
			t.Fatalf("occurrence at position %d sampled twice", occurrence.Position)
		}
		seen[occurrence.Position] = true
	}
	profile := ProfileFromOccurrences(sample)
	if profile.Count != 7 {
		t.Fatalf("ProfileFromOccurrences count=%d", profile.Count)
	}
}

func TestDeterministicSubsampling(t *testing.T) {
	occurrences := map[string][]Occurrence{}
	for i := 0; i < 200; i++ {
		occurrences["A"] = append(occurrences["A"], Occurrence{Position: i % 3, HasLeft: true, Left: []string{"P0", "P1", "P2"}[i%3], HasRight: true, Right: []string{"S0", "S1", "S2"}[(i+1)%3]})
	}
	full := map[string]profilestability.Profile{"A": ProfileFromOccurrences(occurrences["A"])}
	fullWs := profilestability.PrecomputeAll(full)
	config := Config{SubsampleMinFullCount: 160, SubsampleRuns: 20, SubsampleSeed: 7}
	first := runSubsampling(occurrences, full, fullWs, config)
	second := runSubsampling(occurrences, full, fullWs, config)
	if !reflect.DeepEqual(first, second) {
		t.Fatal("subsampling is not deterministic for a fixed seed")
	}
}

// --- reference profile comparison ---

func TestSubsampleAgainstReferenceProfile(t *testing.T) {
	var occurrences []Occurrence
	for i := 0; i < 50; i++ {
		occurrences = append(occurrences, Occurrence{Position: i % 4, HasLeft: true, Left: "L", HasRight: true, Right: "R"})
	}
	reference := ProfileFromOccurrences(occurrences)
	full := profilestability.Compare(ProfileFromOccurrences(occurrences), reference)
	assertClose(t, "full-set self comparison", full.Similarity, 1, 1e-12)

	rng := rand.New(rand.NewSource(3))
	sample := ProfileFromOccurrences(SampleOccurrences(occurrences, 10, rng))
	components := profilestability.Compare(sample, reference)
	if components.Similarity < 0 || components.Similarity > 1 {
		t.Fatalf("similarity out of range: %v", components.Similarity)
	}
}

// --- reliability lookup, log2 interpolation, lower/upper bounds ---

func TestReliabilityTableInterpolationAndBounds(t *testing.T) {
	table := NewReliabilityTable(map[int]float64{10: 0.5, 20: 0.7, 40: 0.9})
	assertClose(t, "exact lookup at 10", table.Reliability(10), 0.5, 1e-12)
	assertClose(t, "exact lookup at 20", table.Reliability(20), 0.7, 1e-12)
	assertClose(t, "exact lookup at 40", table.Reliability(40), 0.9, 1e-12)
	assertClose(t, "below minimum clamps to minimum value", table.Reliability(1), 0.5, 1e-12)
	assertClose(t, "above maximum clamps to maximum value (never extrapolated)", table.Reliability(10000), 0.9, 1e-12)

	midpoint := int(math.Round(math.Sqrt(10 * 20)))
	assertClose(t, "log2-midpoint interpolation between 10 and 20", table.Reliability(midpoint), 0.6, 0.02)

	empty := NewReliabilityTable(map[int]float64{})
	if empty.Reliability(50) != 0 {
		t.Fatal("empty table must not panic and must return zero")
	}
}

func TestReliabilityThresholdsFor(t *testing.T) {
	sizes := []int{10, 20, 40, 80, 160}
	curve := map[int]float64{10: 0.5, 20: 0.82, 40: 0.91, 80: 0.94, 160: 0.94}
	thresholds := ReliabilityThresholdsFor(sizes, curve)
	if thresholds.R80 == nil || *thresholds.R80 != 20 {
		t.Fatalf("r80=%v", thresholds.R80)
	}
	if thresholds.R90 == nil || *thresholds.R90 != 40 {
		t.Fatalf("r90=%v", thresholds.R90)
	}
	if thresholds.R95 != nil {
		t.Fatalf("r95 must stay null when the curve never reaches 0.95, got %v", *thresholds.R95)
	}
}

// --- pair reliability ---

func TestPairReliability(t *testing.T) {
	table := NewReliabilityTable(map[int]float64{10: 0.5, 20: 0.7})
	got := PairReliability(table, 10, 20)
	assertClose(t, "pair reliability is the geometric mean of member reliabilities", got, GeometricMean(0.5, 0.7), 1e-12)
	assertClose(t, "component support", ComponentSupport(0.8, 0.5), 0.4, 1e-12)
}

// --- context diversity + effective observations ---

func TestContextDiversityAndEffectiveObservations(t *testing.T) {
	full := map[string]profilestability.Profile{
		"A": {Count: 100, Left: map[string]int{"X": 50, "Y": 50}, Right: map[string]int{"Z": 100}},
	}
	diversity := buildContextDiversity(full, []string{"A"})
	if len(diversity) != 1 {
		t.Fatalf("len=%d", len(diversity))
	}
	entry := diversity[0]
	if entry.UniquePredecessors != 2 || entry.UniqueSuccessors != 1 {
		t.Fatalf("unique predecessors/successors=%d/%d", entry.UniquePredecessors, entry.UniqueSuccessors)
	}
	assertClose(t, "left entropy of a 50/50 split is 1 bit", entry.LeftEntropy, 1, 1e-12)
	assertClose(t, "right entropy of a single successor is 0 bits", entry.RightEntropy, 0, 1e-12)
	assertClose(t, "effective left observations = count/unique predecessors", entry.EffectiveLeftObservations, 50, 1e-12)
	assertClose(t, "effective right observations = count/unique successors", entry.EffectiveRightObservations, 100, 1e-12)

	zero := buildContextDiversity(map[string]profilestability.Profile{"B": {Count: 5}}, []string{"B"})[0]
	assertClose(t, "no predecessors falls back to dividing by 1, not 0", zero.EffectiveLeftObservations, 5, 1e-12)
}

// --- percentile calculations ---

func TestPercentileAndStatHelpers(t *testing.T) {
	values := []float64{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	assertClose(t, "median", PercentileOf(values, .5), 5.5, 1e-12)
	stat := SummarizeStat(values)
	if stat.Observations != 10 {
		t.Fatalf("observations=%d", stat.Observations)
	}
	assertClose(t, "mean", stat.Mean, 5.5, 1e-12)
	percentileStat := SummarizePercentileStat(values)
	if percentileStat.Percentile90 <= percentileStat.Median || percentileStat.Percentile95 <= percentileStat.Percentile90 {
		t.Fatalf("percentiles must be increasing: median=%v p90=%v p95=%v", percentileStat.Median, percentileStat.Percentile90, percentileStat.Percentile95)
	}
	if empty := SummarizeStat(nil); empty.Observations != 0 {
		t.Fatalf("empty input must report zero observations, got %d", empty.Observations)
	}
}

// --- deterministic YAML output (full Run, end to end) ---

func TestRunIsDeterministicAndStructurallyConsistent(t *testing.T) {
	dir := t.TempDir()
	corpus := "L A R\nL B R\nL A R\nL B R\nX C Y\nL A R\nL B R\nL A R\nL B R\nX C Y\n"
	inputPath := filepath.Join(dir, "corpus.txt")
	if err := os.WriteFile(inputPath, []byte(corpus), 0o644); err != nil {
		t.Fatal(err)
	}
	classes := normalization.ClassesOutput{Models: []normalization.Model{{
		Threshold: .70, Stats: normalization.ModelStats{TokenOccurrenceCoverage: .5},
		Classes: []normalization.Class{{ID: "C0001", Size: 2, Members: []normalization.Member{{Token: "A", Count: 8}, {Token: "B", Count: 8}}}},
	}}}
	classesData, err := yaml.Marshal(classes)
	if err != nil {
		t.Fatal(err)
	}
	classesPath := filepath.Join(dir, "classes.yaml")
	if err := os.WriteFile(classesPath, classesData, 0o644); err != nil {
		t.Fatal(err)
	}

	config := Config{
		InputPath: inputPath, ClassesPath: classesPath, Folds: 2, FoldSeed: 1,
		MinTokenCount: 2, Neighbors: 3, BootstrapRuns: 5, BootstrapSeed: 1,
		Threshold: .70, ThresholdMargin: .05, CountThresholds: []int{2, 4},
		SubsampleMinFullCount: 4, SubsampleRuns: 5, SubsampleSeed: 1,
	}
	first, err := Run(config)
	if err != nil {
		t.Fatal(err)
	}
	second, err := Run(config)
	if err != nil {
		t.Fatal(err)
	}
	firstYAML, err := yaml.Marshal(first)
	if err != nil {
		t.Fatal(err)
	}
	secondYAML, err := yaml.Marshal(second)
	if err != nil {
		t.Fatal(err)
	}
	if string(firstYAML) != string(secondYAML) {
		t.Fatal("two runs with identical seeds must produce byte-identical YAML")
	}

	if len(first.CumulativeThresholds) != 2 {
		t.Fatalf("expected one cumulative_thresholds entry per configured threshold, got %d", len(first.CumulativeThresholds))
	}
	if first.CumulativeThresholds[0].MinCount != 2 || first.CumulativeThresholds[1].MinCount != 4 {
		t.Fatalf("cumulative_thresholds must be sorted ascending: %+v", first.CumulativeThresholds)
	}
	if first.CumulativeThresholds[0].EligibleTokens < first.CumulativeThresholds[1].EligibleTokens {
		t.Fatalf("a smaller min_count must never have fewer eligible tokens than a larger one")
	}
	if len(first.FrequencyBins) != 2 {
		t.Fatalf("expected one frequency bin per configured threshold, got %d", len(first.FrequencyBins))
	}
	if first.FrequencyBins[0].Bin != "2-3" || first.FrequencyBins[1].Bin != "4+" {
		t.Fatalf("unexpected bin labels: %s %s", first.FrequencyBins[0].Bin, first.FrequencyBins[1].Bin)
	}
}

// --- synthetic convergence test ---

func TestSubsamplingConvergesWithSampleSize(t *testing.T) {
	var occurrences []Occurrence
	predecessors := []string{"P0", "P1", "P2", "P3", "P4"}
	successors := []string{"S0", "S1", "S2", "S3", "S4"}
	for i := 0; i < 320; i++ {
		occurrences = append(occurrences, Occurrence{
			Position: i % 5, HasLeft: true, Left: predecessors[i%len(predecessors)],
			HasRight: true, Right: successors[(i+1)%len(successors)],
		})
	}
	occurrenceMap := map[string][]Occurrence{"token": occurrences}
	full := map[string]profilestability.Profile{"token": ProfileFromOccurrences(occurrences)}
	fullWs := profilestability.PrecomputeAll(full)
	config := Config{SubsampleMinFullCount: 160, SubsampleRuns: 200, SubsampleSeed: 1}
	result := runSubsampling(occurrenceMap, full, fullWs, config)
	if len(result.Results) == 0 {
		t.Fatal("no subsampling results produced")
	}
	smallest, largest := result.Results[0], result.Results[len(result.Results)-1]
	if smallest.SampleSize >= largest.SampleSize {
		t.Fatalf("results must be ordered by ascending sample size, got %d then %d", smallest.SampleSize, largest.SampleSize)
	}
	if largest.Position.MeanSimilarity <= smallest.Position.MeanSimilarity {
		t.Fatalf("position reliability should trend upward from n=%d (%v) to n=%d (%v)", smallest.SampleSize, smallest.Position.MeanSimilarity, largest.SampleSize, largest.Position.MeanSimilarity)
	}
	if largest.LeftContext.MeanSimilarity <= smallest.LeftContext.MeanSimilarity {
		t.Fatalf("left-context reliability should trend upward from n=%d (%v) to n=%d (%v)", smallest.SampleSize, smallest.LeftContext.MeanSimilarity, largest.SampleSize, largest.LeftContext.MeanSimilarity)
	}
	if largest.RightContext.MeanSimilarity <= smallest.RightContext.MeanSimilarity {
		t.Fatalf("right-context reliability should trend upward from n=%d (%v) to n=%d (%v)", smallest.SampleSize, smallest.RightContext.MeanSimilarity, largest.SampleSize, largest.RightContext.MeanSimilarity)
	}
}
