package sequenceanalyze

import (
	"bufio"
	"fmt"
	"math"
	"os"
	"sort"
	"strconv"
	"strings"
)

type ngramStats struct {
	Tokens      []string
	Count       int
	LineCount   int
	StartCount  int
	EndCount    int
	Left        map[string]int
	Right       map[string]int
	Occurrences []Coordinate
}

type corpusStats struct {
	Meta     Meta
	ByN      map[int]map[string]*ngramStats
	Contexts map[int]map[string]*contextStats
	Params   Parameters
}

type contextStats struct {
	Tokens []string
	Count  int
	Next   map[string]int
}

func AnalyzeFile(path string, parameters Parameters) (Output, error) {
	file, err := os.Open(path)
	if err != nil {
		return Output{}, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	scanner.Buffer(make([]byte, 64*1024), 16*1024*1024)
	lines := make([][]string, 0)
	for scanner.Scan() {
		tokens := strings.Fields(scanner.Text())
		// Keep empty physical lines in the slice so coordinates retain the
		// original 1-based file line number. analyzeLines ignores their content.
		lines = append(lines, tokens)
	}
	if err := scanner.Err(); err != nil {
		return Output{}, err
	}
	return AnalyzeLines(lines, parameters)
}

func AnalyzeLines(lines [][]string, parameters Parameters) (Output, error) {
	if err := validateParameters(parameters); err != nil {
		return Output{}, err
	}
	stats := &corpusStats{
		ByN:      make(map[int]map[string]*ngramStats),
		Contexts: make(map[int]map[string]*contextStats),
		Params:   parameters,
	}
	for n := parameters.MinN; n <= parameters.MaxN; n++ {
		stats.ByN[n] = make(map[string]*ngramStats)
	}
	for length := 1; length <= parameters.MaxContextLength; length++ {
		stats.Contexts[length] = make(map[string]*contextStats)
	}

	for lineIndex, tokens := range lines {
		if len(tokens) == 0 {
			continue
		}
		stats.Meta.Lines++
		stats.Meta.TokenOccurrences += len(tokens)
		stats.Meta.Transitions += len(tokens) - 1
		seenOnLine := make(map[string]struct{})
		for n := parameters.MinN; n <= parameters.MaxN && n <= len(tokens); n++ {
			for offset := 0; offset+n <= len(tokens); offset++ {
				sequence := tokens[offset : offset+n]
				key := ngramKey(sequence)
				item := stats.ByN[n][key]
				if item == nil {
					item = &ngramStats{Tokens: append([]string(nil), sequence...), Left: make(map[string]int), Right: make(map[string]int)}
					stats.ByN[n][key] = item
				}
				item.Count++
				lineKey := strconv.Itoa(n) + ":" + key
				if _, exists := seenOnLine[lineKey]; !exists {
					item.LineCount++
					seenOnLine[lineKey] = struct{}{}
				}
				if offset == 0 {
					item.StartCount++
				} else {
					item.Left[tokens[offset-1]]++
				}
				if offset+n == len(tokens) {
					item.EndCount++
				} else {
					item.Right[tokens[offset+n]]++
				}
				item.Occurrences = append(item.Occurrences, Coordinate{Line: lineIndex + 1, TokenOffset: offset})
			}
		}
		for length := 1; length <= parameters.MaxContextLength && length < len(tokens); length++ {
			for offset := 0; offset+length < len(tokens); offset++ {
				context := tokens[offset : offset+length]
				key := ngramKey(context)
				item := stats.Contexts[length][key]
				if item == nil {
					item = &contextStats{Tokens: append([]string(nil), context...), Next: make(map[string]int)}
					stats.Contexts[length][key] = item
				}
				item.Count++
				item.Next[tokens[offset+length]]++
			}
		}
	}
	if stats.Meta.TokenOccurrences-stats.Meta.Lines != stats.Meta.Transitions {
		return Output{}, fmt.Errorf("corpus invariant failed: occurrences-lines=%d, transitions=%d", stats.Meta.TokenOccurrences-stats.Meta.Lines, stats.Meta.Transitions)
	}
	return buildOutput(stats), nil
}

func validateParameters(parameters Parameters) error {
	if parameters.MinN < 1 {
		return fmt.Errorf("min-n must be at least 1")
	}
	if parameters.MaxN < parameters.MinN {
		return fmt.Errorf("max-n must be greater than or equal to min-n")
	}
	if parameters.MinCount < 1 {
		return fmt.Errorf("min-count must be at least 1")
	}
	if parameters.MaxItems < 0 {
		return fmt.Errorf("max-items cannot be negative")
	}
	if parameters.ContextLimit < 0 {
		return fmt.Errorf("context-limit cannot be negative")
	}
	if parameters.MaxContextLength < 1 {
		return fmt.Errorf("max-context-length must be at least 1")
	}
	if parameters.ContextMinObservations < 2 {
		return fmt.Errorf("context-min-observations must be at least 2")
	}
	if parameters.ContextMaxItems < 0 {
		return fmt.Errorf("context-max-items cannot be negative")
	}
	return nil
}

func ngramKey(tokens []string) string {
	var builder strings.Builder
	for _, token := range tokens {
		builder.WriteString(strconv.Itoa(len(token)))
		builder.WriteByte(':')
		builder.WriteString(token)
	}
	return builder.String()
}

func buildOutput(stats *corpusStats) Output {
	output := Output{
		Meta:       stats.Meta,
		Parameters: stats.Params,
		Methodology: Methodology{
			SequenceBoundary:   "each non-empty input line is an independent sequence; n-grams never cross line boundaries",
			Tokenization:       "Go strings.Fields; token case, punctuation, and content are preserved",
			Count:              "total number of overlapping occurrences",
			LineCount:          "number of distinct non-empty input lines containing the sequence",
			Entropy:            "Shannon entropy in bits: -sum(p * log2(p)), calculated from all observed contexts",
			NormalizedEntropy:  "entropy/log2(unique contexts) when unique contexts > 1; otherwise 0",
			Predictability:     "1-normalized_entropy; equals 1 when exactly one context is observed",
			MaximalRepeated:    "a sequence with count >= min_count for which no identical left or right extension of length <= max_n preserves its full count",
			Coordinates:        "line is the 1-based physical input-file line; token_offset is 0-based within strings.Fields(line)",
			OutputLimits:       "min_count and max_items filter output records per n; ngram_summary always uses all observed n-grams",
			CrossLine:          "repeated means count >= 2; cross-line repeated additionally requires occurrences in at least two distinct physical input lines",
			ConditionalEntropy: "H(next|context length k) is the context-count-weighted mean of each observed next-token distribution entropy",
			ContextCoverage:    "a singleton context has one observed continuation; repeated_context_coverage is observations belonging to contexts with count >= 2 divided by all observations",
			EntropyDelta:       "entropy at context length k-1 minus entropy at k; a positive value is a diagnostic reduction in uncertainty",
			Interpretation:     "entropy can fall because contexts become sparse; all metrics are formal sequence statistics and have no semantic interpretation",
		},
		RepeatedNGrams:          make(map[int][]NGramResult),
		CrossLineRepeatedNGrams: make(map[int][]NGramResult),
	}

	for n := stats.Params.MinN; n <= stats.Params.MaxN; n++ {
		all := allSorted(stats.ByN[n])
		summary := NGramSummary{N: n, Unique: len(all)}
		for _, item := range all {
			summary.TotalOccurrences += item.Count
			if item.Count == 1 {
				summary.Hapax++
			} else {
				summary.Repeated++
				summary.MultiOccurrence++
				if item.LineCount >= 2 {
					summary.MultiLine++
					summary.MultiLineRepeated++
				} else {
					summary.SingleLineRepeated++
				}
			}
			if item.Count > summary.MaxCount {
				summary.MaxCount = item.Count
			}
		}
		output.NGramSummary = append(output.NGramSummary, summary)

		selected := repeatedSorted(stats.ByN[n], stats.Params.MinCount, stats.Params.MaxItems)
		output.RepeatedNGrams[n] = make([]NGramResult, 0, len(selected))
		for _, item := range selected {
			output.RepeatedNGrams[n] = append(output.RepeatedNGrams[n], ngramResult(item))
			if sumCounts(item.Right) > 0 {
				output.Continuations = append(output.Continuations, continuationResult(item, stats.Params.ContextLimit))
			}
			if sumCounts(item.Left) > 0 {
				output.PredecessorContexts = append(output.PredecessorContexts, predecessorResult(item, stats.Params.ContextLimit))
			}
			output.Extensions = append(output.Extensions, extensionResult(item))
		}
		crossLine := crossLineSorted(stats.ByN[n], stats.Params.MinCount, stats.Params.MaxItems)
		output.CrossLineRepeatedNGrams[n] = make([]NGramResult, 0, len(crossLine))
		for _, item := range crossLine {
			output.CrossLineRepeatedNGrams[n] = append(output.CrossLineRepeatedNGrams[n], ngramResult(item))
		}

		maximal := maximalForN(stats, n)
		if stats.Params.MaxItems > 0 && len(maximal) > stats.Params.MaxItems {
			maximal = maximal[:stats.Params.MaxItems]
		}
		for _, item := range maximal {
			output.MaximalRepeatedSequences = append(output.MaximalRepeatedSequences, MaximalRepeatedSequence{
				Tokens: append([]string(nil), item.Tokens...), N: len(item.Tokens), Count: item.Count, LineCount: item.LineCount,
				StartCount: item.StartCount, EndCount: item.EndCount,
				Occurrences: append([]Coordinate(nil), item.Occurrences...),
			})
		}
		maximalCrossLine := maximalCrossLineForN(stats, n)
		if stats.Params.MaxItems > 0 && len(maximalCrossLine) > stats.Params.MaxItems {
			maximalCrossLine = maximalCrossLine[:stats.Params.MaxItems]
		}
		for _, item := range maximalCrossLine {
			output.MaximalCrossLineSequences = append(output.MaximalCrossLineSequences, MaximalRepeatedSequence{
				Tokens: append([]string(nil), item.Tokens...), N: len(item.Tokens), Count: item.Count, LineCount: item.LineCount,
				StartCount: item.StartCount, EndCount: item.EndCount,
				Occurrences: append([]Coordinate(nil), item.Occurrences...),
			})
		}
	}
	sort.Slice(output.MaximalCrossLineSequences, func(i, j int) bool {
		left := output.MaximalCrossLineSequences[i]
		right := output.MaximalCrossLineSequences[j]
		if left.N != right.N {
			return left.N > right.N
		}
		if left.LineCount != right.LineCount {
			return left.LineCount > right.LineCount
		}
		if left.Count != right.Count {
			return left.Count > right.Count
		}
		return compareTokens(left.Tokens, right.Tokens) < 0
	})
	sort.Slice(output.MaximalRepeatedSequences, func(i, j int) bool {
		left := output.MaximalRepeatedSequences[i]
		right := output.MaximalRepeatedSequences[j]
		if left.N != right.N {
			return left.N > right.N
		}
		if left.LineCount != right.LineCount {
			return left.LineCount > right.LineCount
		}
		if left.Count != right.Count {
			return left.Count > right.Count
		}
		return compareTokens(left.Tokens, right.Tokens) < 0
	})
	output.ContextOrderAnalysis = contextOrderAnalysis(stats)
	output.ContextExtensions = contextExtensions(stats)
	return output
}

func allSorted(items map[string]*ngramStats) []*ngramStats {
	result := make([]*ngramStats, 0, len(items))
	for _, item := range items {
		result = append(result, item)
	}
	sortStats(result)
	return result
}

func repeatedSorted(items map[string]*ngramStats, minCount, maxItems int) []*ngramStats {
	result := make([]*ngramStats, 0)
	for _, item := range items {
		if item.Count >= minCount {
			result = append(result, item)
		}
	}
	sortStats(result)
	if maxItems > 0 && len(result) > maxItems {
		result = result[:maxItems]
	}
	return result
}

func crossLineSorted(items map[string]*ngramStats, minCount, maxItems int) []*ngramStats {
	result := make([]*ngramStats, 0)
	for _, item := range items {
		if item.Count >= minCount && item.LineCount >= 2 {
			result = append(result, item)
		}
	}
	sort.Slice(result, func(i, j int) bool {
		if result[i].LineCount != result[j].LineCount {
			return result[i].LineCount > result[j].LineCount
		}
		if result[i].Count != result[j].Count {
			return result[i].Count > result[j].Count
		}
		return compareTokens(result[i].Tokens, result[j].Tokens) < 0
	})
	if maxItems > 0 && len(result) > maxItems {
		result = result[:maxItems]
	}
	return result
}

func sortStats(items []*ngramStats) {
	sort.Slice(items, func(i, j int) bool {
		if items[i].Count != items[j].Count {
			return items[i].Count > items[j].Count
		}
		return compareTokens(items[i].Tokens, items[j].Tokens) < 0
	})
}

func compareTokens(left, right []string) int {
	limit := len(left)
	if len(right) < limit {
		limit = len(right)
	}
	for i := 0; i < limit; i++ {
		if left[i] < right[i] {
			return -1
		}
		if left[i] > right[i] {
			return 1
		}
	}
	if len(left) < len(right) {
		return -1
	}
	if len(left) > len(right) {
		return 1
	}
	return 0
}

func ngramResult(item *ngramStats) NGramResult {
	return NGramResult{
		Tokens: append([]string(nil), item.Tokens...), N: len(item.Tokens), Count: item.Count, LineCount: item.LineCount,
		CrossLine:  item.LineCount >= 2,
		StartCount: item.StartCount, EndCount: item.EndCount,
		StartProbability: ratio(item.StartCount, item.Count), EndProbability: ratio(item.EndCount, item.Count),
	}
}

func continuationResult(item *ngramStats, limit int) ContinuationResult {
	observed := sumCounts(item.Right)
	unique, entropyValue, normalized, predictability := contextMetrics(item.Right)
	return ContinuationResult{
		Prefix: append([]string(nil), item.Tokens...), N: len(item.Tokens), PrefixCount: item.Count,
		ObservedContinuations: observed, LineEndCount: item.EndCount, UniqueContinuations: unique,
		Entropy: entropyValue, NormalizedEntropy: normalized, Predictability: predictability,
		Next: sortedContexts(item.Right, limit),
	}
}

func predecessorResult(item *ngramStats, limit int) PredecessorResult {
	observed := sumCounts(item.Left)
	unique, entropyValue, normalized, predictability := contextMetrics(item.Left)
	return PredecessorResult{
		Suffix: append([]string(nil), item.Tokens...), N: len(item.Tokens), SuffixCount: item.Count,
		ObservedPredecessors: observed, LineStartCount: item.StartCount, UniquePredecessors: unique,
		Entropy: entropyValue, NormalizedEntropy: normalized, Predictability: predictability,
		Previous: sortedContexts(item.Left, limit),
	}
}

func contextMetrics(counts map[string]int) (int, float64, float64, float64) {
	total := sumCounts(counts)
	if total == 0 {
		return 0, 0, 0, 0
	}
	unique := len(counts)
	entropyValue := 0.0
	keys := make([]string, 0, len(counts))
	for token := range counts {
		keys = append(keys, token)
	}
	sort.Strings(keys)
	for _, token := range keys {
		count := counts[token]
		probability := float64(count) / float64(total)
		entropyValue -= probability * math.Log2(probability)
	}
	if unique == 1 {
		return 1, 0, 0, 1
	}
	normalized := entropyValue / math.Log2(float64(unique))
	return unique, entropyValue, normalized, 1 - normalized
}

func sortedContexts(counts map[string]int, limit int) []ContextToken {
	total := sumCounts(counts)
	result := make([]ContextToken, 0, len(counts))
	for token, count := range counts {
		result = append(result, ContextToken{Token: token, Count: count, Probability: ratio(count, total)})
	}
	sort.Slice(result, func(i, j int) bool {
		if result[i].Count != result[j].Count {
			return result[i].Count > result[j].Count
		}
		return result[i].Token < result[j].Token
	})
	if limit > 0 && len(result) > limit {
		result = result[:limit]
	}
	return result
}

func extensionResult(item *ngramStats) ExtensionResult {
	return ExtensionResult{
		Sequence: append([]string(nil), item.Tokens...), N: len(item.Tokens), Count: item.Count,
		Left: extensionSide(item.Left), Right: extensionSide(item.Right),
		LineStartCount: item.StartCount, LineEndCount: item.EndCount,
	}
}

func extensionSide(counts map[string]int) ExtensionSide {
	contexts := sortedContexts(counts, 1)
	result := ExtensionSide{Observed: sumCounts(counts), Unique: len(counts)}
	if len(contexts) > 0 {
		result.Dominant = &DominantExtension{Token: contexts[0].Token, Count: contexts[0].Count, Probability: contexts[0].Probability}
	}
	return result
}

func maximalForN(stats *corpusStats, n int) []*ngramStats {
	items := repeatedSorted(stats.ByN[n], stats.Params.MinCount, 0)
	result := make([]*ngramStats, 0, len(items))
	for _, item := range items {
		if n < stats.Params.MaxN && (hasCountPreservingExtension(item.Left, item.Count) || hasCountPreservingExtension(item.Right, item.Count)) {
			continue
		}
		result = append(result, item)
	}
	return result
}

func maximalCrossLineForN(stats *corpusStats, n int) []*ngramStats {
	items := crossLineSorted(stats.ByN[n], stats.Params.MinCount, 0)
	result := make([]*ngramStats, 0, len(items))
	for _, item := range items {
		if n < stats.Params.MaxN && (hasCountPreservingExtension(item.Left, item.Count) || hasCountPreservingExtension(item.Right, item.Count)) {
			continue
		}
		result = append(result, item)
	}
	return result
}

func contextOrderAnalysis(stats *corpusStats) []ContextOrderResult {
	results := make([]ContextOrderResult, 0, stats.Params.MaxContextLength)
	for length := 1; length <= stats.Params.MaxContextLength; length++ {
		result := ContextOrderResult{ContextLength: length, UniqueContexts: len(stats.Contexts[length])}
		weightedEntropy := 0.0
		repeatedWeightedEntropy := 0.0
		contextKeys := make([]string, 0, len(stats.Contexts[length]))
		for key := range stats.Contexts[length] {
			contextKeys = append(contextKeys, key)
		}
		sort.Strings(contextKeys)
		for _, key := range contextKeys {
			context := stats.Contexts[length][key]
			_, entropyValue, _, _ := contextMetrics(context.Next)
			result.Observations += context.Count
			weightedEntropy += float64(context.Count) * entropyValue
			if context.Count == 1 {
				result.SingletonContexts++
			} else {
				result.RepeatedContexts++
				result.ObservationsInRepeatedContexts += context.Count
				repeatedWeightedEntropy += float64(context.Count) * entropyValue
			}
		}
		result.SingletonFraction = ratio(result.SingletonContexts, result.UniqueContexts)
		result.RepeatedContextCoverage = ratio(result.ObservationsInRepeatedContexts, result.Observations)
		if result.Observations > 0 {
			result.ConditionalEntropy = weightedEntropy / float64(result.Observations)
			result.Perplexity = math.Exp2(result.ConditionalEntropy)
		}
		if result.ObservationsInRepeatedContexts > 0 {
			result.RepeatedContextConditionalEntropy = repeatedWeightedEntropy / float64(result.ObservationsInRepeatedContexts)
			result.RepeatedContextPerplexity = math.Exp2(result.RepeatedContextConditionalEntropy)
		}
		if len(results) > 0 {
			previous := results[len(results)-1]
			delta := previous.ConditionalEntropy - result.ConditionalEntropy
			repeatedDelta := previous.RepeatedContextConditionalEntropy - result.RepeatedContextConditionalEntropy
			result.EntropyDeltaFromPrevious = &delta
			result.RepeatedEntropyDeltaFromPrevious = &repeatedDelta
		}
		results = append(results, result)
	}
	return results
}

func contextExtensions(stats *corpusStats) []ContextExtensionResult {
	results := make([]ContextExtensionResult, 0)
	for length := 2; length <= stats.Params.MaxContextLength; length++ {
		for _, long := range stats.Contexts[length] {
			if long.Count < stats.Params.ContextMinObservations {
				continue
			}
			shortTokens := long.Tokens[1:]
			short := stats.Contexts[length-1][ngramKey(shortTokens)]
			if short == nil {
				continue
			}
			shortUnique, shortEntropy, _, _ := contextMetrics(short.Next)
			longUnique, longEntropy, _, _ := contextMetrics(long.Next)
			reduction := shortEntropy - longEntropy
			if reduction == 0 {
				continue
			}
			results = append(results, ContextExtensionResult{
				ShortContext:      append([]string(nil), short.Tokens...),
				LongContext:       append([]string(nil), long.Tokens...),
				ShortCount:        short.Count,
				LongCount:         long.Count,
				ShortEntropy:      shortEntropy,
				LongEntropy:       longEntropy,
				EntropyReduction:  reduction,
				ShortUniqueNext:   shortUnique,
				LongUniqueNext:    longUnique,
				ShortDominantNext: dominantNext(short.Next),
				LongDominantNext:  dominantNext(long.Next),
			})
		}
	}
	sort.Slice(results, func(i, j int) bool {
		leftChange := math.Abs(results[i].EntropyReduction)
		rightChange := math.Abs(results[j].EntropyReduction)
		if leftChange != rightChange {
			return leftChange > rightChange
		}
		if results[i].LongCount != results[j].LongCount {
			return results[i].LongCount > results[j].LongCount
		}
		return compareTokens(results[i].LongContext, results[j].LongContext) < 0
	})
	if stats.Params.ContextMaxItems > 0 && len(results) > stats.Params.ContextMaxItems {
		results = results[:stats.Params.ContextMaxItems]
	}
	return results
}

func dominantNext(counts map[string]int) DominantNext {
	contexts := sortedContexts(counts, 1)
	if len(contexts) == 0 {
		return DominantNext{}
	}
	return DominantNext{Token: contexts[0].Token, Count: contexts[0].Count, Probability: contexts[0].Probability}
}

func hasCountPreservingExtension(contexts map[string]int, count int) bool {
	for _, contextCount := range contexts {
		if contextCount == count {
			return true
		}
	}
	return false
}

func sumCounts(counts map[string]int) int {
	total := 0
	for _, count := range counts {
		total += count
	}
	return total
}

func ratio(numerator, denominator int) float64 {
	if numerator == 0 || denominator == 0 {
		return 0
	}
	return float64(numerator) / float64(denominator)
}
