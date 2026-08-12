package validation

import (
	"math"
	"sort"
	"strconv"
	"strings"
)

type sequenceMetrics struct {
	CrossLine map[int]int
	MaxLength int
	Contexts  map[int]ContextMetrics
	NGrams    map[int]map[string]*sequenceNGram
}

type sequenceNGram struct {
	Tokens      []string
	Count       int
	Lines       map[int]bool
	Occurrences []Coordinate
}

type contextCount struct {
	Count int
	Next  map[string]int
}

func AnalyzeSequences(corpus Corpus, minN, maxN, maxContext int) sequenceMetrics {
	result := sequenceMetrics{
		CrossLine: make(map[int]int), Contexts: make(map[int]ContextMetrics),
		NGrams: make(map[int]map[string]*sequenceNGram),
	}
	contexts := make(map[int]map[string]*contextCount)
	for n := minN; n <= maxN; n++ {
		result.NGrams[n] = make(map[string]*sequenceNGram)
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
					item = &sequenceNGram{Tokens: append([]string(nil), tokens...), Lines: make(map[int]bool)}
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
			entropy := entropy(item.Next)
			observations += item.Count
			weightedEntropy += float64(item.Count) * entropy
			if item.Count >= 2 {
				repeatedObservations += item.Count
				repeatedWeightedEntropy += float64(item.Count) * entropy
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

func NewCrossLineSequences(raw Corpus, rawMetrics, structuralMetrics sequenceMetrics, minN, maxN int) []NewCrossLineSequence {
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

func tokenKey(tokens []string) string {
	var builder strings.Builder
	for _, token := range tokens {
		builder.WriteString(strconv.Itoa(len(token)))
		builder.WriteByte(':')
		builder.WriteString(token)
	}
	return builder.String()
}

func entropy(counts map[string]int) float64 {
	total := sumIntCounts(counts)
	if total == 0 {
		return 0
	}
	value := 0.0
	keys := make([]string, 0, len(counts))
	for key := range counts {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		count := counts[key]
		if count == 0 {
			continue
		}
		probability := float64(count) / float64(total)
		value -= probability * math.Log2(probability)
	}
	return value
}
