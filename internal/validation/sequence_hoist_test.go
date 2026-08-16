package validation

import (
	"fmt"
	"math/rand"
	"reflect"
	"sort"
	"strings"
	"testing"
)

// referenceSequenceNGram and referenceAnalyzeSequences mirror sequenceNGram
// and AnalyzeSequences exactly as they stood before the dense-key
// rewrite: NGrams keyed by tokenKey's length-prefixed text concatenation,
// and per-n-gram line membership tracked by a map[int]bool rather than a
// last-line/distinct-count pair.
type referenceSequenceNGram struct {
	Tokens      []string
	Count       int
	Lines       map[int]bool
	Occurrences []Coordinate
}

type referenceSequenceMetrics struct {
	CrossLine map[int]int
	MaxLength int
	Contexts  map[int]ContextMetrics
	NGrams    map[int]map[string]*referenceSequenceNGram
}

func referenceAnalyzeSequences(corpus Corpus, minN, maxN, maxContext int) referenceSequenceMetrics {
	result := referenceSequenceMetrics{
		CrossLine: make(map[int]int), Contexts: make(map[int]ContextMetrics),
		NGrams: make(map[int]map[string]*referenceSequenceNGram),
	}
	contexts := make(map[int]map[string]*contextCount)
	for n := minN; n <= maxN; n++ {
		result.NGrams[n] = make(map[string]*referenceSequenceNGram)
	}
	for length := 1; length <= maxContext; length++ {
		contexts[length] = make(map[string]*contextCount)
	}
	for _, line := range corpus.Lines {
		for n := minN; n <= maxN && n <= len(line.Tokens); n++ {
			for offset := 0; offset+n <= len(line.Tokens); offset++ {
				tokens := line.Tokens[offset : offset+n]
				key := tokenKey(tokens)
				item := result.NGrams[n][key]
				if item == nil {
					item = &referenceSequenceNGram{Tokens: append([]string(nil), tokens...), Lines: make(map[int]bool)}
					result.NGrams[n][key] = item
				}
				item.Count++
				item.Lines[line.ID] = true
				item.Occurrences = append(item.Occurrences, Coordinate{LineID: line.ID, TokenOffset: offset})
			}
		}
		for length := 1; length <= maxContext && length < len(line.Tokens); length++ {
			for offset := 0; offset+length < len(line.Tokens); offset++ {
				key := tokenKey(line.Tokens[offset : offset+length])
				item := contexts[length][key]
				if item == nil {
					item = &contextCount{Next: make(map[string]int)}
					contexts[length][key] = item
				}
				item.Count++
				item.Next[line.Tokens[offset+length]]++
			}
		}
	}
	for n := minN; n <= maxN; n++ {
		for _, item := range result.NGrams[n] {
			if item.Count >= 2 && len(item.Lines) >= 2 {
				result.CrossLine[n]++
				if n > result.MaxLength {
					result.MaxLength = n
				}
			}
		}
	}
	for length := 1; length <= maxContext; length++ {
		metric := ContextMetrics{ContextLength: length}
		observations, repeatedObservations := 0, 0
		weightedEntropy, repeatedWeightedEntropy := 0.0, 0.0
		keys := make([]string, 0, len(contexts[length]))
		for key := range contexts[length] {
			keys = append(keys, key)
		}
		sort.Strings(keys)
		for _, key := range keys {
			item := contexts[length][key]
			e := entropy(item.Next)
			observations += item.Count
			weightedEntropy += float64(item.Count) * e
			if item.Count >= 2 {
				repeatedObservations += item.Count
				repeatedWeightedEntropy += float64(item.Count) * e
			}
		}
		if observations > 0 {
			metric.ConditionalEntropy = weightedEntropy / float64(observations)
			metric.RepeatedContextCoverage = float64(repeatedObservations) / float64(observations)
		}
		if repeatedObservations > 0 {
			metric.RepeatedContextConditionalEntropy = repeatedWeightedEntropy / float64(repeatedObservations)
		}
		result.Contexts[length] = metric
	}
	return result
}

// referenceNewCrossLineSequences mirrors NewCrossLineSequences against the
// reference (map[int]bool-based) metrics type.
func referenceNewCrossLineSequences(raw Corpus, rawMetrics, structuralMetrics referenceSequenceMetrics, minN, maxN int) []NewCrossLineSequence {
	rawLines := make(map[int][]string, len(raw.Lines))
	for _, line := range raw.Lines {
		rawLines[line.ID] = line.Tokens
	}
	var result []NewCrossLineSequence
	for n := minN; n <= maxN; n++ {
		for key, item := range structuralMetrics.NGrams[n] {
			if item.Count < 2 || len(item.Lines) < 2 {
				continue
			}
			rawItem := rawMetrics.NGrams[n][key]
			if rawItem != nil && rawItem.Count >= 2 && len(rawItem.Lines) >= 2 {
				continue
			}
			entry := NewCrossLineSequence{
				NormalizedTokens: append([]string(nil), item.Tokens...), N: n,
				Count: item.Count, LineCount: len(item.Lines),
			}
			for _, occurrence := range item.Occurrences {
				tokens := rawLines[occurrence.LineID]
				if occurrence.TokenOffset+n > len(tokens) {
					continue
				}
				entry.Occurrences = append(entry.Occurrences, SurfaceRealization{
					LineID: occurrence.LineID, TokenOffset: occurrence.TokenOffset,
					RawTokens: append([]string(nil), tokens[occurrence.TokenOffset:occurrence.TokenOffset+n]...),
				})
			}
			result = append(result, entry)
		}
	}
	sort.Slice(result, func(i, j int) bool {
		if result[i].N != result[j].N {
			return result[i].N < result[j].N
		}
		if result[i].LineCount != result[j].LineCount {
			return result[i].LineCount > result[j].LineCount
		}
		if result[i].Count != result[j].Count {
			return result[i].Count > result[j].Count
		}
		return strings.Join(result[i].NormalizedTokens, "\x00") < strings.Join(result[j].NormalizedTokens, "\x00")
	})
	return result
}

// metricsEqual compares only the fields AnalyzeSequences' callers ever
// observe (CrossLine, MaxLength, Contexts) - the internal NGrams map keys
// are not comparable across the reference (string) and rewritten (dense)
// implementations by construction, and were never meant to be.
func metricsEqual(t *testing.T, name string, got sequenceMetrics, want referenceSequenceMetrics) {
	t.Helper()
	if !reflect.DeepEqual(got.CrossLine, want.CrossLine) {
		t.Fatalf("%s: CrossLine diverged\ngot=%v\nwant=%v", name, got.CrossLine, want.CrossLine)
	}
	if got.MaxLength != want.MaxLength {
		t.Fatalf("%s: MaxLength diverged: got %d want %d", name, got.MaxLength, want.MaxLength)
	}
	if !reflect.DeepEqual(got.Contexts, want.Contexts) {
		t.Fatalf("%s: Contexts diverged\ngot=%v\nwant=%v", name, got.Contexts, want.Contexts)
	}
}

func fixtureSequenceCorpus(nLines, seed int) Corpus {
	r := rand.New(rand.NewSource(int64(seed)))
	vocab := []string{"aiin", "chey", "shey", "ol", "or", "dy", "qokeey", "s", "daiin"}
	corpus := Corpus{Counts: map[string]int{}}
	for li := 0; li < nLines; li++ {
		length := 2 + r.Intn(10)
		tokens := make([]string, length)
		for i := range tokens {
			tokens[i] = vocab[r.Intn(len(vocab))]
			corpus.Counts[tokens[i]]++
		}
		corpus.Lines = append(corpus.Lines, Line{ID: li + 1, Tokens: tokens})
		corpus.Occurrences += length
	}
	return corpus
}

func TestAnalyzeSequencesHoistMatchesReference(t *testing.T) {
	for _, nLines := range []int{0, 1, 5, 40} {
		corpus := fixtureSequenceCorpus(nLines, nLines*97+11)
		vocab := newVocabIndex(corpus)
		want := referenceAnalyzeSequences(corpus, 2, 5, 3)
		got := AnalyzeSequences(corpus, 2, 5, 3, vocab)
		metricsEqual(t, fmt.Sprintf("nLines=%d", nLines), got, want)
	}
}

// TestAnalyzeSequencesHoistMatchesReferenceWithSyntheticTokens exercises
// the exact scenario that made a statically-built vocabulary unsafe: a
// mapped corpus introducing a token string (a synthetic class ID) absent
// from the original corpus.
func TestAnalyzeSequencesHoistMatchesReferenceWithSyntheticTokens(t *testing.T) {
	raw := testCorpus([][]string{{"x", "A", "z", "w"}, {"x", "B", "z", "w"}, {"x", "A", "z", "w"}})
	normalized := applyMapping(raw, map[string]string{"A": "C0001", "B": "C0001"})
	vocab := newVocabIndex(raw)
	gotRaw := AnalyzeSequences(raw, 2, 4, 3, vocab)
	gotNormalized := AnalyzeSequences(normalized, 2, 4, 3, vocab)
	wantRaw := referenceAnalyzeSequences(raw, 2, 4, 3)
	wantNormalized := referenceAnalyzeSequences(normalized, 2, 4, 3)
	metricsEqual(t, "raw", gotRaw, wantRaw)
	metricsEqual(t, "normalized", gotNormalized, wantNormalized)

	gotItems := NewCrossLineSequences(raw, gotRaw, gotNormalized, 4, 4)
	wantItems := referenceNewCrossLineSequences(raw, wantRaw, wantNormalized, 4, 4)
	if !reflect.DeepEqual(gotItems, wantItems) {
		t.Fatalf("NewCrossLineSequences diverged\ngot=%+v\nwant=%+v", gotItems, wantItems)
	}
}

func benchSequenceCorpus() Corpus {
	return fixtureSequenceCorpus(500, 42)
}

func BenchmarkAnalyzeSequencesReference(b *testing.B) {
	corpus := benchSequenceCorpus()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		referenceAnalyzeSequences(corpus, 2, 8, 7)
	}
}

func BenchmarkAnalyzeSequencesHoisted(b *testing.B) {
	corpus := benchSequenceCorpus()
	vocab := newVocabIndex(corpus)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		AnalyzeSequences(corpus, 2, 8, 7, vocab)
	}
}
