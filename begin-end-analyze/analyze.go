package main

import (
	"fmt"
	"math"
	"math/rand"
	"sort"
	"strings"
)

type scopeSequence struct {
	tokens    []string
	positions map[string][]int
}

type permutationMoments struct {
	sum        []float64
	sumSquares []float64
	exceed     []int
}

func runAnalysis(dictionaryPath, corpusPath string, parameters Parameters) (Output, error) {
	if parameters.MaxWindow < 1 {
		return Output{}, fmt.Errorf("max-window must be at least 1")
	}
	if parameters.Permutations < 0 {
		return Output{}, fmt.Errorf("permutations cannot be negative")
	}
	if parameters.MinTokenCount < 1 {
		return Output{}, fmt.Errorf("min-frequency must be at least 1")
	}
	if parameters.MaxCandidates < 0 {
		return Output{}, fmt.Errorf("max-candidates cannot be negative")
	}
	if parameters.PermutationMode != "page" && parameters.PermutationMode != "line" {
		return Output{}, fmt.Errorf("permutation-mode must be page or line")
	}
	parameters.Windows = configuredWindows(parameters.MaxWindow)
	dictionary, err := loadDictionary(dictionaryPath)
	if err != nil {
		return Output{}, fmt.Errorf("load dictionary: %w", err)
	}
	c, err := loadCorpus(corpusPath)
	if err != nil {
		return Output{}, fmt.Errorf("load corpus: %w", err)
	}
	if err := validateCorpusDictionary(c, dictionary); err != nil {
		return Output{}, err
	}

	byToken := make(map[string]DictionaryToken, len(dictionary))
	eligible := make([]string, 0)
	unclearExcluded := 0
	for _, token := range dictionary {
		byToken[token.Token] = token
		if token.Count < parameters.MinTokenCount {
			continue
		}
		if strings.Contains(token.Token, "?") && !parameters.IncludeUnclear {
			unclearExcluded++
			continue
		}
		eligible = append(eligible, token.Token)
	}
	sort.Strings(eligible)
	if len(eligible) < 2 {
		return Output{}, fmt.Errorf("fewer than two eligible tokens")
	}
	ids := make(map[string]int, len(eligible))
	for i, token := range eligible {
		ids[token] = i
	}
	lineSequences, pageSequences := buildScopeSequences(c)
	position := calculatePositions(c, dictionary)
	pageCounts := calculatePageCounts(c)
	pageBoundariesKnown := c.ExplicitBreaks > 0

	k := len(eligible)
	observedLine := enumerateWindowHits(lineSequences, ids, parameters.MaxWindow)
	observedPage := enumerateWindowHits(pageSequences, ids, parameters.MaxWindow)
	lineMoments, pageMoments := runPermutations(c, ids, parameters, observedLine, observedPage)

	candidates := make([]Candidate, 0, k*(k-1))
	for ai, a := range eligible {
		for bi, b := range eligible {
			if ai == bi {
				continue
			}
			line := directedDistance(a, b, lineSequences, parameters.Windows, parameters.MaxWindow, "line")
			page := directedDistance(a, b, pageSequences, parameters.Windows, parameters.MaxWindow, "page")
			index := ai*k + bi
			reverseIndex := bi*k + ai
			adjacent := neighborCount(byToken[a].WordAfter, b)
			localShare := ratio(adjacent, page.Observations)
			direction := directionality(page.Probability, ratio(observedPage[reverseIndex], byToken[b].Count), "page")
			if !pageBoundariesKnown {
				direction = directionality(line.Probability, ratio(observedLine[reverseIndex], byToken[b].Count), "line")
			}
			candidate := Candidate{
				BeginCandidate: a, EndCandidate: b, ContainsUnclear: strings.Contains(a, "?") || strings.Contains(b, "?"),
				Counts:     CountsResult{Begin: byToken[a].Count, End: byToken[b].Count},
				Position:   PositionResult{Begin: position[a], End: position[b]},
				WithinLine: line, WithinPage: page,
				Directionality:     direction,
				PageBalance:        pageBalance(a, b, pageCounts),
				SignificanceLine:   significance(observedLine[index], lineMoments, index, parameters.Permutations),
				SignificancePage:   significance(observedPage[index], pageMoments, index, parameters.Permutations),
				LocalCompatibility: LocalCompatibility{AdjacentCount: adjacent, AdjacentShare: localShare, LikelyLocal: page.Observations > 0 && localShare >= 0.8},
				Reliability:        pairReliability(byToken[a].Count, byToken[b].Count),
			}
			candidate.Position.Complementarity = positionalComplementarity(candidate.Position.Begin, candidate.Position.End)
			candidate.Score = candidateScore(candidate)
			candidates = append(candidates, candidate)
		}
	}
	calibratePageBalance(candidates, len(c.Pages) > 1)
	for i := range candidates {
		candidates[i].Score = candidateScore(candidates[i])
	}
	sortCandidates(candidates)
	// Nesting is more expensive and is evaluated for every record that can reach
	// an output report, followed by a final transparent re-ranking.
	nestingLimit := len(candidates)
	if parameters.MaxCandidates > 0 && nestingLimit > parameters.MaxCandidates*2 {
		nestingLimit = parameters.MaxCandidates * 2
	}
	for i := 0; i < nestingLimit; i++ {
		candidates[i].Nesting = nestingCounts(candidates[i].BeginCandidate, candidates[i].EndCandidate, pageSequences)
		candidates[i].Score = candidateScore(candidates[i])
	}
	sortCandidates(candidates)

	local := make([]Candidate, 0)
	main := make([]Candidate, 0)
	for _, candidate := range candidates {
		if candidate.LocalCompatibility.LikelyLocal {
			if len(local) < 100 {
				local = append(local, candidate)
			}
			continue
		}
		if parameters.MaxCandidates == 0 || len(main) < parameters.MaxCandidates {
			main = append(main, candidate)
		}
	}
	occurrences := 0
	for _, count := range c.Counts {
		occurrences += count
	}
	return Output{
		Meta:       Meta{TokenOccurrences: occurrences, UniqueTokens: len(dictionary), EligibleTokens: len(eligible), Lines: len(c.Lines), Pages: len(c.Pages), ExplicitPageBreaks: c.ExplicitBreaks, PageBoundariesKnown: pageBoundariesKnown, CandidatePairs: len(candidates), UnclearExcluded: unclearExcluded},
		Parameters: parameters,
		Methodology: Methodology{
			Tokenization:       "Go strings.Fields; @NNN; remains one token and no punctuation is removed",
			Pages:              "blank lines, form-feed, # page: markers, and === page ... === markers delimit pages; absent markers imply one page and page_boundaries_known=false",
			DirectedDependency: "for each opening-candidate occurrence, use the nearest later closing-candidate occurrence in the same scope; distance 1 means adjacency",
			Permutation:        "empirical one-sided p=(1+random counts >= observed)/(permutations+1); page mode shuffles within each page and restores original line lengths, line mode shuffles independently within each line",
			Nesting:            "counts sliding groups of four in each page after filtering to the two paired tokens; the four reported orders are descriptive",
			Score:              "reliability * (0.30 significance + 0.20 positive directionality + 0.15 positional complementarity + 0.10 page balance + 0.10 nesting contrast + 0.15 nonlocal directed coverage); reliability=min_pair_count/(min_pair_count+20); line scope replaces page scope when page boundaries are unknown",
			Interpretation:     "ranked pairs are neutral directed-dependency candidates, not semantic operators or proof of structure",
		},
		Candidates: main, LikelyLocalPairs: local,
	}, nil
}

func configuredWindows(max int) []int {
	base := []int{1, 2, 3, 5, 8, 13, 21, 34, 55}
	result := make([]int, 0, len(base)+1)
	for _, value := range base {
		if value <= max {
			result = append(result, value)
		}
	}
	if len(result) == 0 || result[len(result)-1] != max {
		result = append(result, max)
	}
	return result
}

func buildScopeSequences(c *corpus) ([]scopeSequence, []scopeSequence) {
	lines := make([]scopeSequence, 0, len(c.Lines))
	for _, line := range c.Lines {
		lines = append(lines, newScopeSequence(line.Tokens))
	}
	pages := make([]scopeSequence, 0, len(c.Pages))
	for _, lineIndexes := range c.Pages {
		var tokens []string
		for _, index := range lineIndexes {
			tokens = append(tokens, c.Lines[index].Tokens...)
		}
		pages = append(pages, newScopeSequence(tokens))
	}
	return lines, pages
}

func newScopeSequence(tokens []string) scopeSequence {
	s := scopeSequence{tokens: append([]string(nil), tokens...), positions: make(map[string][]int)}
	for i, token := range tokens {
		s.positions[token] = append(s.positions[token], i)
	}
	return s
}

func directedDistance(a, b string, sequences []scopeSequence, windows []int, maxWindow int, scope string) DistanceResult {
	result := DistanceResult{Scope: scope, Histogram: make(map[int]int)}
	distances := make([]int, 0)
	endWithout := 0
	allScopeHits := 0
	for _, sequence := range sequences {
		as, bs := sequence.positions[a], sequence.positions[b]
		result.BeginOccurrences += len(as)
		for _, position := range as {
			index := sort.SearchInts(bs, position+1)
			if index < len(bs) {
				allScopeHits++
				distance := bs[index] - position
				if distance <= maxWindow {
					distances = append(distances, distance)
					result.Histogram[distance]++
				}
			}
		}
		for _, position := range bs {
			index := sort.SearchInts(as, position)
			if index == 0 || position-as[index-1] > maxWindow {
				endWithout++
			}
		}
	}
	sort.Ints(distances)
	result.Observations = len(distances)
	result.Probability = ratio(result.Observations, result.BeginOccurrences)
	result.WithoutEnd = result.BeginOccurrences - result.Observations
	result.EndWithoutPriorBegin = endWithout
	if len(distances) > 0 {
		total := 0
		for _, distance := range distances {
			total += distance
		}
		result.Mean = float64(total) / float64(len(distances))
		middle := len(distances) / 2
		if len(distances)%2 == 1 {
			result.Median = float64(distances[middle])
		} else {
			result.Median = float64(distances[middle-1]+distances[middle]) / 2
		}
	}
	for _, window := range windows {
		count := sort.Search(len(distances), func(i int) bool { return distances[i] > window })
		result.Windows = append(result.Windows, WindowResult{Window: fmt.Sprintf("%d", window), Observations: count, Probability: ratio(count, result.BeginOccurrences)})
	}
	result.Windows = append(result.Windows, WindowResult{Window: "to_end_of_" + scope, Observations: allScopeHits, Probability: ratio(allScopeHits, result.BeginOccurrences)})
	return result
}

func enumerateWindowHits(sequences []scopeSequence, ids map[string]int, maxWindow int) []int {
	k := len(ids)
	result := make([]int, k*k)
	seen := make([]int, k)
	stamp := 0
	for _, sequence := range sequences {
		for i, a := range sequence.tokens {
			ai, ok := ids[a]
			if !ok {
				continue
			}
			stamp++
			limit := i + maxWindow + 1
			if limit > len(sequence.tokens) {
				limit = len(sequence.tokens)
			}
			for j := i + 1; j < limit; j++ {
				bi, ok := ids[sequence.tokens[j]]
				if !ok || bi == ai || seen[bi] == stamp {
					continue
				}
				seen[bi] = stamp
				result[ai*k+bi]++
			}
		}
	}
	return result
}

func runPermutations(c *corpus, ids map[string]int, parameters Parameters, observedLine, observedPage []int) (permutationMoments, permutationMoments) {
	size := len(observedLine)
	lineMoments := permutationMoments{sum: make([]float64, size), sumSquares: make([]float64, size), exceed: make([]int, size)}
	pageMoments := permutationMoments{sum: make([]float64, size), sumSquares: make([]float64, size), exceed: make([]int, size)}
	rng := rand.New(rand.NewSource(parameters.RandomSeed))
	for run := 0; run < parameters.Permutations; run++ {
		shuffled := shuffledCorpus(c, parameters.PermutationMode, rng)
		lines, pages := buildScopeSequences(shuffled)
		lineHits := enumerateWindowHits(lines, ids, parameters.MaxWindow)
		pageHits := enumerateWindowHits(pages, ids, parameters.MaxWindow)
		accumulateMoments(&lineMoments, lineHits, observedLine)
		accumulateMoments(&pageMoments, pageHits, observedPage)
	}
	return lineMoments, pageMoments
}

func shuffledCorpus(c *corpus, mode string, rng *rand.Rand) *corpus {
	result := &corpus{Pages: make([][]int, len(c.Pages)), Counts: c.Counts, ExplicitBreaks: c.ExplicitBreaks}
	for pageIndex, indexes := range c.Pages {
		if mode == "page" {
			var flat []string
			for _, index := range indexes {
				flat = append(flat, c.Lines[index].Tokens...)
			}
			rng.Shuffle(len(flat), func(i, j int) { flat[i], flat[j] = flat[j], flat[i] })
			offset := 0
			for _, index := range indexes {
				length := len(c.Lines[index].Tokens)
				lineIndex := len(result.Lines)
				result.Lines = append(result.Lines, corpusLine{Tokens: append([]string(nil), flat[offset:offset+length]...), Page: pageIndex})
				result.Pages[pageIndex] = append(result.Pages[pageIndex], lineIndex)
				offset += length
			}
		} else {
			for _, index := range indexes {
				tokens := append([]string(nil), c.Lines[index].Tokens...)
				rng.Shuffle(len(tokens), func(i, j int) { tokens[i], tokens[j] = tokens[j], tokens[i] })
				lineIndex := len(result.Lines)
				result.Lines = append(result.Lines, corpusLine{Tokens: tokens, Page: pageIndex})
				result.Pages[pageIndex] = append(result.Pages[pageIndex], lineIndex)
			}
		}
	}
	return result
}

func accumulateMoments(m *permutationMoments, values, observed []int) {
	for i, value := range values {
		v := float64(value)
		m.sum[i] += v
		m.sumSquares[i] += v * v
		if value >= observed[i] {
			m.exceed[i]++
		}
	}
}

func significance(observed int, moments permutationMoments, index, runs int) SignificanceResult {
	result := SignificanceResult{PermutationP: 1, Permutations: runs}
	if runs == 0 {
		return result
	}
	result.ExpectedMean = moments.sum[index] / float64(runs)
	variance := moments.sumSquares[index]/float64(runs) - result.ExpectedMean*result.ExpectedMean
	if variance < 0 && variance > -1e-12 {
		variance = 0
	}
	result.ExpectedStddev = math.Sqrt(math.Max(0, variance))
	result.PermutationP = float64(moments.exceed[index]+1) / float64(runs+1)
	if result.ExpectedStddev > 0 {
		result.ZScore = (float64(observed) - result.ExpectedMean) / result.ExpectedStddev
	}
	return result
}

func calculatePositions(c *corpus, dictionary []DictionaryToken) map[string]PositionSide {
	normalizedSum := make(map[string]float64)
	for _, line := range c.Lines {
		for position, token := range line.Tokens {
			value := 0.0
			if len(line.Tokens) > 1 {
				value = float64(position) / float64(len(line.Tokens)-1)
			}
			normalizedSum[token] += value
		}
	}
	result := make(map[string]PositionSide, len(dictionary))
	for _, token := range dictionary {
		absolute := 0
		for _, position := range token.PositionInString {
			absolute += position.Position * position.Count
		}
		result[token.Token] = PositionSide{StartProbability: ratio(token.LineStartCount, token.Count), EndProbability: ratio(token.LineEndCount, token.Count), MeanPosition: float64(absolute) / float64(token.Count), MeanNormalizedPosition: normalizedSum[token.Token] / float64(token.Count)}
	}
	return result
}

func calculatePageCounts(c *corpus) []map[string]int {
	result := make([]map[string]int, len(c.Pages))
	for page, indexes := range c.Pages {
		result[page] = make(map[string]int)
		for _, index := range indexes {
			for _, token := range c.Lines[index].Tokens {
				result[page][token]++
			}
		}
	}
	return result
}

func pageBalance(a, b string, pages []map[string]int) PageBalanceResult {
	result := PageBalanceResult{}
	if len(pages) == 0 {
		return result
	}
	diffs := make([]float64, len(pages))
	sum := 0.0
	for i, page := range pages {
		diffs[i] = float64(page[a] - page[b])
		sum += diffs[i]
		result.MeanAbsoluteDifference += math.Abs(diffs[i])
		if math.Abs(diffs[i]) <= 1 {
			result.NearZeroFraction++
		}
	}
	result.MeanDifference = sum / float64(len(pages))
	result.MeanAbsoluteDifference /= float64(len(pages))
	result.NearZeroFraction /= float64(len(pages))
	for _, d := range diffs {
		delta := d - result.MeanDifference
		result.StddevDifference += delta * delta
	}
	result.StddevDifference = math.Sqrt(result.StddevDifference / float64(len(pages)))
	return result
}

type balanceBin struct {
	low  int
	high int
}

type balanceAggregate struct {
	sum   float64
	count int
}

func calibratePageBalance(candidates []Candidate, meaningful bool) {
	aggregates := make(map[balanceBin]balanceAggregate)
	for _, candidate := range candidates {
		key := frequencyBin(candidate.Counts.Begin, candidate.Counts.End)
		item := aggregates[key]
		item.sum += candidate.PageBalance.StddevDifference
		item.count++
		aggregates[key] = item
	}
	for i := range candidates {
		aggregate := aggregates[frequencyBin(candidates[i].Counts.Begin, candidates[i].Counts.End)]
		if aggregate.count > 0 {
			candidates[i].PageBalance.ComparableStddev = aggregate.sum / float64(aggregate.count)
		}
		baseline := candidates[i].PageBalance.ComparableStddev
		if baseline > 0 {
			candidates[i].PageBalance.StddevRatio = candidates[i].PageBalance.StddevDifference / baseline
			if meaningful {
				candidates[i].PageBalance.RelativeScore = clamp01(1 - candidates[i].PageBalance.StddevRatio)
			}
		}
	}
}

func frequencyBin(a, b int) balanceBin {
	if a > b {
		a, b = b, a
	}
	return balanceBin{low: countBin(a), high: countBin(b)}
}

func countBin(count int) int {
	bin := 0
	for count > 1 {
		count >>= 1
		bin++
	}
	return bin
}

func nestingCounts(a, b string, pages []scopeSequence) NestingResult {
	result := NestingResult{}
	for _, page := range pages {
		filtered := make([]string, 0)
		for _, token := range page.tokens {
			if token == a || token == b {
				filtered = append(filtered, token)
			}
		}
		for i := 0; i+4 <= len(filtered); i++ {
			pattern := filtered[i : i+4]
			switch {
			case pattern[0] == a && pattern[1] == a && pattern[2] == b && pattern[3] == b:
				result.AABB++
			case pattern[0] == a && pattern[1] == b && pattern[2] == a && pattern[3] == b:
				result.ABAB++
			case pattern[0] == a && pattern[1] == b && pattern[2] == b && pattern[3] == a:
				result.ABBA++
			case pattern[0] == b && pattern[1] == a && pattern[2] == a && pattern[3] == b:
				result.BAAB++
			}
		}
	}
	return result
}

func neighborCount(neighbors []NeighborInput, token string) int {
	for _, n := range neighbors {
		if n.Token == token {
			return n.Count
		}
	}
	return 0
}

func directionality(forward, reverse float64, scope string) DirectionalityResult {
	const epsilon = 1e-9
	return DirectionalityResult{Scope: scope, AToB: forward, BToA: reverse, Score: forward - reverse, LogRatio: math.Log((forward + epsilon) / (reverse + epsilon))}
}

func positionalComplementarity(a, b PositionSide) float64 {
	order := clamp01((b.MeanNormalizedPosition - a.MeanNormalizedPosition + 1) / 2)
	boundary := clamp01(((a.StartProbability - a.EndProbability) + (b.EndProbability - b.StartProbability) + 2) / 4)
	return (order + boundary) / 2
}

func candidateScore(c Candidate) float64 {
	significanceResult := c.SignificancePage
	coverage := c.WithinPage.Probability
	if c.Directionality.Scope == "line" {
		significanceResult = c.SignificanceLine
		coverage = c.WithinLine.Probability
	}
	significance := clamp01(math.Max(0, significanceResult.ZScore) / 5)
	direction := clamp01(math.Max(0, c.Directionality.Score))
	balance := c.PageBalance.RelativeScore
	nestingTotal := c.Nesting.AABB + c.Nesting.ABAB + c.Nesting.ABBA + c.Nesting.BAAB
	nesting := 0.0
	if nestingTotal > 0 {
		nesting = clamp01(float64(c.Nesting.AABB-c.Nesting.ABAB) / float64(nestingTotal))
	}
	nonlocal := clamp01(coverage) * (1 - clamp01(c.LocalCompatibility.AdjacentShare))
	base := 0.30*significance + 0.20*direction + 0.15*c.Position.Complementarity + 0.10*balance + 0.10*nesting + 0.15*nonlocal
	return c.Reliability * base
}

func pairReliability(a, b int) float64 {
	if a > b {
		a = b
	}
	return float64(a) / float64(a+20)
}

func sortCandidates(items []Candidate) {
	sort.Slice(items, func(i, j int) bool {
		if items[i].Score != items[j].Score {
			return items[i].Score > items[j].Score
		}
		if items[i].BeginCandidate != items[j].BeginCandidate {
			return items[i].BeginCandidate < items[j].BeginCandidate
		}
		return items[i].EndCandidate < items[j].EndCandidate
	})
}
func ratio(numerator, denominator int) float64 {
	if denominator <= 0 {
		return 0
	}
	return float64(numerator) / float64(denominator)
}
func clamp01(value float64) float64 {
	if value < 0 {
		return 0
	}
	if value > 1 {
		return 1
	}
	return value
}
