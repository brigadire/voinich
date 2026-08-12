package validation

import (
	"reflect"
	"sort"
	"testing"

	"zcore.dev/voinich/internal/normalization"
)

func TestDeterministicFoldsAndPartitionInvariants(t *testing.T) {
	corpus := testCorpus([][]string{{"a", "b"}, {"c"}, {}, {"d", "e", "f"}, {"g"}, {"h"}, {"i"}})
	first, err := SplitFolds(corpus.Lines, 3, 17)
	if err != nil {
		t.Fatal(err)
	}
	second, err := SplitFolds(corpus.Lines, 3, 17)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(first, second) {
		t.Fatal("same seed produced different folds")
	}
	seen := make(map[int]int)
	for _, indexes := range first {
		train, test, err := Partition(corpus, indexes)
		if err != nil {
			t.Fatal(err)
		}
		trainIDs, testIDs := lineIDSet(train), lineIDSet(test)
		for id := range testIDs {
			seen[id]++
			if trainIDs[id] {
				t.Fatalf("line %d occurs in TRAIN and TEST", id)
			}
		}
		if len(train.Lines)+len(test.Lines) != len(corpus.Lines) {
			t.Fatal("TRAIN + TEST does not equal corpus")
		}
		if train.Occurrences+test.Occurrences != corpus.Occurrences {
			t.Fatal("token occurrence partition mismatch")
		}
	}
	for _, line := range corpus.Lines {
		if seen[line.ID] != 1 {
			t.Fatalf("line %d is TEST %d times", line.ID, seen[line.ID])
		}
	}
}

func TestTestOnlySimilarityCannotCreateClass(t *testing.T) {
	train := testCorpus([][]string{
		{"left-x", "X", "right-x"}, {"left-x", "X", "right-x"}, {"left-x", "X", "right-x"},
		{"left-y", "Y", "right-y"}, {"left-y", "Y", "right-y"}, {"left-y", "Y", "right-y"},
	})
	model, _, err := BuildTrainModel(train, Config{Threshold: .70, MinTokenCount: 3})
	if err != nil {
		t.Fatal(err)
	}
	if sameClass(model, "X", "Y") {
		t.Fatal("TRAIN-incompatible X/Y were merged")
	}
	test := testCorpus([][]string{{"same", "X", "tail"}, {"same", "Y", "tail"}, {"same", "X", "tail"}, {"same", "Y", "tail"}})
	normalized := applyMapping(test, normalization.Mapping(model, "preserve"))
	if normalized.Lines[0].Tokens[1] == normalized.Lines[1].Tokens[1] {
		t.Fatal("TEST-only similarity changed TRAIN classes")
	}
}

func TestTrainClassRemainsFixedUnderDifferentTestContexts(t *testing.T) {
	train := testCorpus([][]string{
		{"L", "A", "R"}, {"L", "A", "R"}, {"L", "A", "R"},
		{"L", "B", "R"}, {"L", "B", "R"}, {"L", "B", "R"},
	})
	model, _, err := BuildTrainModel(train, Config{Threshold: .70, MinTokenCount: 3})
	if err != nil {
		t.Fatal(err)
	}
	if !sameClass(model, "A", "B") {
		t.Fatal("TRAIN-equivalent A/B were not merged")
	}
	test := testCorpus([][]string{{"A", "x", "x"}, {"q", "q", "B"}})
	normalized := applyMapping(test, normalization.Mapping(model, "preserve"))
	if normalized.Lines[0].Tokens[0] != normalized.Lines[1].Tokens[2] {
		t.Fatal("fixed TRAIN class was changed by TEST contexts")
	}
}

func TestTrainingPreservesAllPositionObservations(t *testing.T) {
	train := testCorpus([][]string{{"X"}, {"a", "X"}, {"a", "b", "X"}, {"a", "b", "c", "X"}})
	stats := collectTrainingStats(train)
	if len(stats.Positions["X"]) != 4 || sumIntCounts(stats.Positions["X"]) != 4 {
		t.Fatalf("TRAIN positions were truncated: %+v", stats.Positions["X"])
	}
}

func TestTestOnlyTokenIsPreservedAndRandomUsesTrainCounts(t *testing.T) {
	train := testCorpus([][]string{
		{"L", "A", "R"}, {"L", "A", "R"}, {"L", "A", "R"},
		{"L", "B", "R"}, {"L", "B", "R"}, {"L", "B", "R"},
	})
	model, eligible, err := BuildTrainModel(train, Config{Threshold: .70, MinTokenCount: 3})
	if err != nil {
		t.Fatal(err)
	}
	if eligible["TEST"] {
		t.Fatal("TEST-only token became TRAIN eligible")
	}
	random := normalization.RandomModel(model, normalizationCorpus(train), 3, 9, 0)
	for _, class := range multiMemberClasses(random) {
		for _, member := range class.Members {
			if train.Counts[member.Token] != member.Count {
				t.Fatalf("random member %q does not use TRAIN count", member.Token)
			}
			if member.Token == "TEST" {
				t.Fatal("TEST token entered random class")
			}
		}
	}
	test := testCorpus([][]string{{"TEST", "TEST", "TEST", "TEST"}})
	normalized := applyMapping(test, normalization.Mapping(random, "preserve"))
	if normalized.Lines[0].Tokens[0] != "TEST" {
		t.Fatal("TEST-only token was not preserved")
	}
}

func TestClassStability(t *testing.T) {
	models := []normalization.Model{
		modelWithClasses([][]string{{"A", "B"}}),
		modelWithClasses([][]string{{"A", "B"}}),
		modelWithClasses(nil),
	}
	eligible := []map[string]bool{{"A": true, "B": true}, {"A": true, "B": true}, {"A": true, "B": true}}
	result := BuildClassStability(models, eligible)
	if len(result.Pairs) != 1 || result.Pairs[0].FoldsBothEligible != 3 || result.Pairs[0].FoldsSameClass != 2 {
		t.Fatalf("unexpected stability: %+v", result)
	}
	if result.Pairs[0].Stability != 2.0/3.0 || result.UnstablePairs != 1 {
		t.Fatalf("unexpected ratio: %+v", result)
	}
}

func TestNewSequenceSurfaceReconstruction(t *testing.T) {
	raw := testCorpus([][]string{{"x", "A", "z", "w"}, {"x", "B", "z", "w"}})
	normalized := applyMapping(raw, map[string]string{"A": "C0001", "B": "C0001"})
	rawMetrics := AnalyzeSequences(raw, 2, 4, 3)
	normalizedMetrics := AnalyzeSequences(normalized, 2, 4, 3)
	items := NewCrossLineSequences(raw, rawMetrics, normalizedMetrics, 4, 4)
	if len(items) != 1 || items[0].Count != 2 || len(items[0].Occurrences) != 2 {
		t.Fatalf("unexpected reconstructed sequences: %+v", items)
	}
	if !reflect.DeepEqual(items[0].Occurrences[0].RawTokens, []string{"x", "A", "z", "w"}) {
		t.Fatalf("wrong surface realization: %+v", items[0])
	}
}

func TestAggregateAndEmpiricalP(t *testing.T) {
	folds := []FoldResult{
		{SequenceComparison: SequenceComparison{NGrams: []NGramComparison{{N: 2, RawCrossLineRepeated: 2, StructuralCrossLineRepeated: 4, AbsoluteDelta: 2}}}},
		{SequenceComparison: SequenceComparison{NGrams: []NGramComparison{{N: 2, RawCrossLineRepeated: 3, StructuralCrossLineRepeated: 3, AbsoluteDelta: 0}}}},
		{SequenceComparison: SequenceComparison{NGrams: []NGramComparison{{N: 2, RawCrossLineRepeated: 5, StructuralCrossLineRepeated: 4, AbsoluteDelta: -1}}}},
	}
	result := AggregateFolds(folds, 2, 2)
	item := result.CrossLineNGrams[0]
	if item.FoldsPositive != 1 || item.FoldsZero != 1 || item.FoldsNegative != 1 || item.MedianDelta != 0 {
		t.Fatalf("unexpected aggregate: %+v", item)
	}
	if result.PooledTest[0].RawCrossLineRepeated != 10 || result.PooledTest[0].StructuralCrossLineRepeated != 11 {
		t.Fatalf("unexpected pooled result: %+v", result.PooledTest[0])
	}
	if got := EmpiricalP([]float64{1, 2, 3}, 3, true); got != .5 {
		t.Fatalf("empirical p=%v, want .5", got)
	}
}

func TestLeaveOneClassOutAndMemberAblation(t *testing.T) {
	corpus := testCorpus([][]string{
		{"x", "A", "z", "w"}, {"x", "B", "z", "w"}, {"x", "E", "z", "w"},
		{"p", "C", "q", "r"}, {"p", "D", "q", "r"},
	})
	model := modelWithClasses([][]string{{"A", "B", "E"}, {"C", "D"}})
	result, members := runAblations(corpus, model, Config{MinN: 2, MaxN: 4, MaxContext: 3})
	if len(result.Variants) != 2 {
		t.Fatalf("got %d leave-one-class variants", len(result.Variants))
	}
	if len(members) != 3 {
		t.Fatalf("got %d member ablations, want 3", len(members))
	}
	var large LeaveOneOutVariant
	for _, variant := range result.Variants {
		if len(variant.ClassMembers) == 3 {
			large = variant
		}
	}
	if large.ContributionN4 != 1 {
		t.Fatalf("large-class n4 contribution=%d, want 1", large.ContributionN4)
	}
	for _, item := range members {
		if len(item.RemainingNormalizedMembers) != 2 {
			t.Fatalf("wrong member ablation: %+v", item)
		}
	}
}

func testCorpus(lines [][]string) Corpus {
	result := Corpus{Counts: make(map[string]int)}
	for i, tokens := range lines {
		addLine(&result, Line{ID: i + 1, Tokens: append([]string(nil), tokens...)})
	}
	return result
}

func lineIDSet(corpus Corpus) map[int]bool {
	result := make(map[int]bool)
	for _, line := range corpus.Lines {
		result[line.ID] = true
	}
	return result
}

func modelWithClasses(groups [][]string) normalization.Model {
	model := normalization.Model{Threshold: .7, Label: "070"}
	for i, group := range groups {
		sorted := append([]string(nil), group...)
		sort.Strings(sorted)
		class := normalization.Class{ID: "C000" + string(rune('1'+i)), Size: len(sorted)}
		for _, token := range sorted {
			class.Members = append(class.Members, normalization.Member{Token: token, Count: 1})
		}
		model.Classes = append(model.Classes, class)
	}
	return model
}
