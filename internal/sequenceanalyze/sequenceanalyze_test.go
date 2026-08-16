package sequenceanalyze

import (
	"math"
	"reflect"
	"testing"

	"gopkg.in/yaml.v3"
)

func testParameters(minN, maxN int) Parameters {
	return Parameters{
		MinN: minN, MaxN: maxN, MinCount: 2, MaxItems: 0, ContextLimit: 10,
		MaxContextLength: maxN - 1, ContextMinObservations: 2, ContextMaxItems: 0,
	}
}

func TestBigramAndTrigramCountsWithoutCrossingLines(t *testing.T) {
	output, err := AnalyzeLines([][]string{{"a", "b", "c"}, {"b", "c", "d"}}, testParameters(2, 3))
	if err != nil {
		t.Fatal(err)
	}
	assertSummary(t, output.NGramSummary[0], NGramSummary{
		N: 2, TotalOccurrences: 4, Unique: 3, Repeated: 1, MultiOccurrence: 1,
		MultiLine: 1, MultiLineRepeated: 1, Hapax: 2, MaxCount: 2,
	})
	assertSummary(t, output.NGramSummary[1], NGramSummary{N: 3, TotalOccurrences: 2, Unique: 2, Hapax: 2, MaxCount: 1})
	if findRaw(output.RepeatedNGrams[2], []string{"c", "b"}) != nil {
		t.Fatal("found an n-gram crossing a line boundary")
	}
	bc := requireRaw(t, output.RepeatedNGrams[2], []string{"b", "c"})
	if bc.Count != 2 || bc.LineCount != 2 {
		t.Fatalf("b c counts = count %d, lines %d; want 2, 2", bc.Count, bc.LineCount)
	}
}

func TestCountLineCountAndBoundaries(t *testing.T) {
	output, err := AnalyzeLines([][]string{{"a", "a", "a"}}, testParameters(2, 2))
	if err != nil {
		t.Fatal(err)
	}
	aa := requireRaw(t, output.RepeatedNGrams[2], []string{"a", "a"})
	if aa.Count != 2 || aa.LineCount != 1 || aa.StartCount != 1 || aa.EndCount != 1 {
		t.Fatalf("unexpected repeated bigram: %+v", *aa)
	}
	assertClose(t, "start probability", aa.StartProbability, 0.5)
	assertClose(t, "end probability", aa.EndProbability, 0.5)
}

func TestContinuationAndPredecessorContexts(t *testing.T) {
	lines := [][]string{
		{"L", "A", "B", "X"},
		{"M", "A", "B", "Y"},
		{"L", "A", "B", "X"},
		{"A", "B"},
	}
	output, err := AnalyzeLines(lines, testParameters(2, 2))
	if err != nil {
		t.Fatal(err)
	}
	continuation := requireContinuation(t, output.Continuations, []string{"A", "B"})
	if continuation.PrefixCount != 4 || continuation.ObservedContinuations != 3 || continuation.LineEndCount != 1 || continuation.UniqueContinuations != 2 {
		t.Fatalf("unexpected continuation: %+v", *continuation)
	}
	assertContext(t, continuation.Next, "X", 2, 2.0/3.0)
	assertContext(t, continuation.Next, "Y", 1, 1.0/3.0)
	assertClose(t, "continuation entropy", continuation.Entropy, -(2.0/3.0)*math.Log2(2.0/3.0)-(1.0/3.0)*math.Log2(1.0/3.0))
	assertClose(t, "continuation predictability", continuation.Predictability, 1-continuation.Entropy)

	predecessor := requirePredecessor(t, output.PredecessorContexts, []string{"A", "B"})
	if predecessor.ObservedPredecessors != 3 || predecessor.LineStartCount != 1 || predecessor.UniquePredecessors != 2 {
		t.Fatalf("unexpected predecessor: %+v", *predecessor)
	}
	assertContext(t, predecessor.Previous, "L", 2, 2.0/3.0)
}

func TestSingleContextPredictability(t *testing.T) {
	output, err := AnalyzeLines([][]string{{"A", "B", "X"}, {"A", "B", "X"}}, testParameters(2, 2))
	if err != nil {
		t.Fatal(err)
	}
	continuation := requireContinuation(t, output.Continuations, []string{"A", "B"})
	if continuation.UniqueContinuations != 1 || continuation.Entropy != 0 || continuation.NormalizedEntropy != 0 || continuation.Predictability != 1 {
		t.Fatalf("unexpected deterministic context metrics: %+v", *continuation)
	}
}

func TestMaximalRepeatedSequenceAndCoordinates(t *testing.T) {
	parameters := testParameters(2, 5)
	output, err := AnalyzeLines([][]string{{"X", "A", "B", "C", "Y"}, {}, {"X", "A", "B", "C", "Y"}}, parameters)
	if err != nil {
		t.Fatal(err)
	}
	if len(output.MaximalRepeatedSequences) != 1 {
		t.Fatalf("maximal sequences = %d, want 1: %+v", len(output.MaximalRepeatedSequences), output.MaximalRepeatedSequences)
	}
	sequence := output.MaximalRepeatedSequences[0]
	if !reflect.DeepEqual(sequence.Tokens, []string{"X", "A", "B", "C", "Y"}) {
		t.Fatalf("maximal sequence = %v", sequence.Tokens)
	}
	wantCoordinates := []Coordinate{{Line: 1, TokenOffset: 0}, {Line: 3, TokenOffset: 0}}
	if !reflect.DeepEqual(sequence.Occurrences, wantCoordinates) {
		t.Fatalf("coordinates = %+v, want %+v", sequence.Occurrences, wantCoordinates)
	}
}

func TestMinCountMaxItemsAndDeterministicSorting(t *testing.T) {
	parameters := testParameters(2, 2)
	parameters.MinCount = 2
	parameters.MaxItems = 2
	lines := [][]string{{"b", "x"}, {"a", "z"}, {"c", "q"}, {"b", "x"}, {"a", "z"}, {"c", "q"}, {"single", "only"}}
	first, err := AnalyzeLines(lines, parameters)
	if err != nil {
		t.Fatal(err)
	}
	got := first.RepeatedNGrams[2]
	if len(got) != 2 || !reflect.DeepEqual(got[0].Tokens, []string{"a", "z"}) || !reflect.DeepEqual(got[1].Tokens, []string{"b", "x"}) {
		t.Fatalf("limited/sorted n-grams = %+v", got)
	}
	if findRaw(got, []string{"single", "only"}) != nil {
		t.Fatal("min-count did not remove a hapax")
	}
	unlimitedParameters := parameters
	unlimitedParameters.MaxItems = 0
	unlimited, err := AnalyzeLines(lines, unlimitedParameters)
	if err != nil {
		t.Fatal(err)
	}
	if len(unlimited.RepeatedNGrams[2]) != 3 {
		t.Fatalf("max-items=0 returned %d records, want all 3", len(unlimited.RepeatedNGrams[2]))
	}
	second, _ := AnalyzeLines(lines, parameters)
	firstYAML, _ := yaml.Marshal(first)
	secondYAML, _ := yaml.Marshal(second)
	if !reflect.DeepEqual(firstYAML, secondYAML) {
		t.Fatal("output is not deterministic")
	}
}

func TestCorpusInvariant(t *testing.T) {
	output, err := AnalyzeLines([][]string{{"a"}, {}, {"b", "c", "d"}}, testParameters(2, 3))
	if err != nil {
		t.Fatal(err)
	}
	if output.Meta.TokenOccurrences != 4 || output.Meta.Lines != 2 || output.Meta.Transitions != 2 {
		t.Fatalf("unexpected meta: %+v", output.Meta)
	}
	if output.Meta.TokenOccurrences-output.Meta.Lines != output.Meta.Transitions {
		t.Fatal("corpus invariant failed")
	}
	if output.ContextOrderAnalysis[0].Observations != 2 {
		t.Fatalf("context observations = %d, want only two within-line transitions", output.ContextOrderAnalysis[0].Observations)
	}
}

func TestSingleLineAndCrossLineRepeatedNGrams(t *testing.T) {
	output, err := AnalyzeLines([][]string{
		{"A", "B", "A", "B"},
		{"C", "D"},
		{"C", "D"},
	}, testParameters(2, 3))
	if err != nil {
		t.Fatal(err)
	}
	summary := output.NGramSummary[0]
	if summary.MultiOccurrence != 2 || summary.MultiLine != 1 || summary.MultiLineRepeated != 1 || summary.SingleLineRepeated != 1 {
		t.Fatalf("unexpected cross-line summary: %+v", summary)
	}
	local := requireRaw(t, output.RepeatedNGrams[2], []string{"A", "B"})
	if local.CrossLine || local.LineCount != 1 {
		t.Fatalf("local repeat marked cross-line: %+v", *local)
	}
	cross := requireRaw(t, output.RepeatedNGrams[2], []string{"C", "D"})
	if !cross.CrossLine || cross.LineCount != 2 {
		t.Fatalf("cross-line repeat not marked: %+v", *cross)
	}
	if len(output.CrossLineRepeatedNGrams[2]) != 1 || !reflect.DeepEqual(output.CrossLineRepeatedNGrams[2][0].Tokens, []string{"C", "D"}) {
		t.Fatalf("cross-line section = %+v", output.CrossLineRepeatedNGrams[2])
	}
}

func TestMaximalCrossLineSequence(t *testing.T) {
	output, err := AnalyzeLines([][]string{
		{"X", "A", "B", "C", "Y"},
		{"X", "A", "B", "C", "Y"},
		{"L", "M", "L", "M"},
	}, testParameters(2, 5))
	if err != nil {
		t.Fatal(err)
	}
	if len(output.MaximalCrossLineSequences) != 1 {
		t.Fatalf("maximal cross-line sequences = %+v", output.MaximalCrossLineSequences)
	}
	sequence := output.MaximalCrossLineSequences[0]
	if !reflect.DeepEqual(sequence.Tokens, []string{"X", "A", "B", "C", "Y"}) || sequence.LineCount != 2 || sequence.StartCount != 2 || sequence.EndCount != 2 {
		t.Fatalf("unexpected maximal cross-line sequence: %+v", sequence)
	}
}

func TestContextOrderConditionalEntropyAndCoverage(t *testing.T) {
	lines := [][]string{
		{"P", "A", "X"},
		{"P", "A", "X"},
		{"Q", "A", "Y"},
		{"R", "B", "Z"},
	}
	output, err := AnalyzeLines(lines, testParameters(2, 3))
	if err != nil {
		t.Fatal(err)
	}
	if len(output.ContextOrderAnalysis) != 2 {
		t.Fatalf("context orders = %d, want 2", len(output.ContextOrderAnalysis))
	}
	order1 := output.ContextOrderAnalysis[0]
	hA := -(2.0/3.0)*math.Log2(2.0/3.0) - (1.0/3.0)*math.Log2(1.0/3.0)
	if order1.Observations != 8 || order1.UniqueContexts != 5 || order1.SingletonContexts != 3 || order1.RepeatedContexts != 2 || order1.ObservationsInRepeatedContexts != 5 {
		t.Fatalf("unexpected context length 1 counts: %+v", order1)
	}
	assertClose(t, "singleton fraction", order1.SingletonFraction, 3.0/5.0)
	assertClose(t, "repeated context coverage", order1.RepeatedContextCoverage, 5.0/8.0)
	assertClose(t, "weighted conditional entropy", order1.ConditionalEntropy, 3.0/8.0*hA)
	assertClose(t, "repeated conditional entropy", order1.RepeatedContextConditionalEntropy, 3.0/5.0*hA)
	assertClose(t, "perplexity", order1.Perplexity, math.Exp2(order1.ConditionalEntropy))

	order2 := output.ContextOrderAnalysis[1]
	if order2.Observations != 4 || order2.SingletonContexts != 2 || order2.RepeatedContexts != 1 || order2.ObservationsInRepeatedContexts != 2 {
		t.Fatalf("unexpected context length 2 counts: %+v", order2)
	}
	assertClose(t, "order 2 entropy", order2.ConditionalEntropy, 0)
	assertClose(t, "order 2 repeated coverage", order2.RepeatedContextCoverage, 0.5)
	if order2.EntropyDeltaFromPrevious == nil || order2.RepeatedEntropyDeltaFromPrevious == nil {
		t.Fatal("context entropy deltas are absent")
	}
	assertClose(t, "entropy delta", *order2.EntropyDeltaFromPrevious, order1.ConditionalEntropy)
	assertClose(t, "repeated entropy delta", *order2.RepeatedEntropyDeltaFromPrevious, 3.0/5.0*hA)
}

func TestContextExtensionsAndObservationThreshold(t *testing.T) {
	lines := [][]string{
		{"P", "A", "X"},
		{"P", "A", "X"},
		{"Q", "A", "Y"},
		{"Q", "A", "Y"},
	}
	parameters := testParameters(2, 3)
	parameters.ContextMinObservations = 2
	output, err := AnalyzeLines(lines, parameters)
	if err != nil {
		t.Fatal(err)
	}
	if len(output.ContextExtensions) != 2 {
		t.Fatalf("context extensions = %+v, want two", output.ContextExtensions)
	}
	first := output.ContextExtensions[0]
	if !reflect.DeepEqual(first.ShortContext, []string{"A"}) || !reflect.DeepEqual(first.LongContext, []string{"P", "A"}) {
		t.Fatalf("unexpected first context extension: %+v", first)
	}
	if first.ShortCount != 4 || first.LongCount != 2 || first.ShortUniqueNext != 2 || first.LongUniqueNext != 1 {
		t.Fatalf("unexpected context extension counts: %+v", first)
	}
	assertClose(t, "context entropy reduction", first.EntropyReduction, 1)

	parameters.ContextMinObservations = 3
	filtered, err := AnalyzeLines(lines, parameters)
	if err != nil {
		t.Fatal(err)
	}
	if len(filtered.ContextExtensions) != 0 {
		t.Fatalf("under-supported context extensions were not filtered: %+v", filtered.ContextExtensions)
	}
}

func TestContextExtensionsIncludeEntropyIncrease(t *testing.T) {
	lines := [][]string{
		{"P", "A", "X"}, {"P", "A", "Y"},
		{"Q", "A", "X"}, {"Q", "A", "X"},
	}
	parameters := testParameters(2, 3)
	parameters.ContextMinObservations = 2
	output, err := AnalyzeLines(lines, parameters)
	if err != nil {
		t.Fatal(err)
	}
	foundIncrease := false
	for _, extension := range output.ContextExtensions {
		if reflect.DeepEqual(extension.LongContext, []string{"P", "A"}) && extension.EntropyReduction < 0 {
			foundIncrease = true
		}
	}
	if !foundIncrease {
		t.Fatalf("entropy-increasing context extension is absent: %+v", output.ContextExtensions)
	}
}

func TestContextLengthMayExceedNGramRange(t *testing.T) {
	parameters := testParameters(2, 2)
	parameters.MaxContextLength = 3
	output, err := AnalyzeLines([][]string{{"A", "B", "C", "D"}}, parameters)
	if err != nil {
		t.Fatal(err)
	}
	if len(output.ContextOrderAnalysis) != 3 {
		t.Fatalf("context orders = %d, want 3", len(output.ContextOrderAnalysis))
	}
}

func assertSummary(t *testing.T, got, want NGramSummary) {
	t.Helper()
	if got != want {
		t.Fatalf("summary = %+v, want %+v", got, want)
	}
}

func findRaw(items []NGramResult, tokens []string) *NGramResult {
	for i := range items {
		if reflect.DeepEqual(items[i].Tokens, tokens) {
			return &items[i]
		}
	}
	return nil
}

func requireRaw(t *testing.T, items []NGramResult, tokens []string) *NGramResult {
	t.Helper()
	item := findRaw(items, tokens)
	if item == nil {
		t.Fatalf("n-gram %v not found", tokens)
	}
	return item
}

func requireContinuation(t *testing.T, items []ContinuationResult, tokens []string) *ContinuationResult {
	t.Helper()
	for i := range items {
		if reflect.DeepEqual(items[i].Prefix, tokens) {
			return &items[i]
		}
	}
	t.Fatalf("continuation %v not found", tokens)
	return nil
}

func requirePredecessor(t *testing.T, items []PredecessorResult, tokens []string) *PredecessorResult {
	t.Helper()
	for i := range items {
		if reflect.DeepEqual(items[i].Suffix, tokens) {
			return &items[i]
		}
	}
	t.Fatalf("predecessor %v not found", tokens)
	return nil
}

func assertContext(t *testing.T, contexts []ContextToken, token string, count int, probability float64) {
	t.Helper()
	for _, context := range contexts {
		if context.Token == token {
			if context.Count != count {
				t.Fatalf("context %s count = %d, want %d", token, context.Count, count)
			}
			assertClose(t, "context probability", context.Probability, probability)
			return
		}
	}
	t.Fatalf("context %s not found", token)
}

func assertClose(t *testing.T, name string, got, want float64) {
	t.Helper()
	if math.Abs(got-want) > 1e-12 {
		t.Fatalf("%s = %.15g, want %.15g", name, got, want)
	}
}
