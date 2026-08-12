package validation

import (
	"math"
	"sort"

	"zcore.dev/voinich/internal/normalization"
)

type trainingStats struct {
	Counts    map[string]int
	Positions map[string]map[int]int
	Left      map[string]map[string]int
	Right     map[string]map[string]int
}

func BuildTrainModel(train Corpus, config Config) (normalization.Model, map[string]bool, error) {
	stats := collectTrainingStats(train)
	eligible := make(map[string]bool)
	for token, count := range stats.Counts {
		if count >= config.MinTokenCount {
			eligible[token] = true
		}
	}
	candidates := equivalenceCandidates(stats, config.MinTokenCount, math.Min(config.Threshold, 0.70), 100)
	input := normalization.StructuralInput{EquivalenceCandidates: candidates}
	input.Parameters.MinTokenCountForRanking = config.MinTokenCount
	models, _, err := normalization.BuildModels(normalizationCorpus(train), input, normalization.Config{
		Thresholds: []float64{config.Threshold}, MinTokenCount: config.MinTokenCount, SingletonMode: "preserve",
	})
	if err != nil {
		return normalization.Model{}, nil, err
	}
	return models[0], eligible, nil
}

func collectTrainingStats(corpus Corpus) trainingStats {
	stats := trainingStats{
		Counts: make(map[string]int), Positions: make(map[string]map[int]int),
		Left: make(map[string]map[string]int), Right: make(map[string]map[string]int),
	}
	allPositions := make(map[string]map[int]int)
	for _, line := range corpus.Lines {
		for position, token := range line.Tokens {
			stats.Counts[token]++
			if allPositions[token] == nil {
				allPositions[token] = make(map[int]int)
			}
			allPositions[token][position]++
			if position > 0 {
				incrementNeighbor(stats.Left, token, line.Tokens[position-1])
			}
			if position+1 < len(line.Tokens) {
				incrementNeighbor(stats.Right, token, line.Tokens[position+1])
			}
		}
	}
	for token, positions := range allPositions {
		stats.Positions[token] = topPositionCounts(positions, 3)
	}
	return stats
}

func incrementNeighbor(contexts map[string]map[string]int, token, neighbor string) {
	if contexts[token] == nil {
		contexts[token] = make(map[string]int)
	}
	contexts[token][neighbor]++
}

func topPositionCounts(counts map[int]int, limit int) map[int]int {
	type item struct{ position, count int }
	items := make([]item, 0, len(counts))
	for position, count := range counts {
		items = append(items, item{position, count})
	}
	sort.Slice(items, func(i, j int) bool {
		if items[i].count != items[j].count {
			return items[i].count > items[j].count
		}
		return items[i].position < items[j].position
	})
	if len(items) > limit {
		items = items[:limit]
	}
	result := make(map[int]int, len(items))
	for _, item := range items {
		result[item.position] = item.count
	}
	return result
}

func equivalenceCandidates(stats trainingStats, minCount int, minSimilarity float64, limit int) []normalization.Candidate {
	var tokens []string
	for token, count := range stats.Counts {
		if count >= minCount {
			tokens = append(tokens, token)
		}
	}
	sort.Slice(tokens, func(i, j int) bool {
		if stats.Counts[tokens[i]] != stats.Counts[tokens[j]] {
			return stats.Counts[tokens[i]] > stats.Counts[tokens[j]]
		}
		return tokens[i] < tokens[j]
	})
	type ranked struct {
		candidate normalization.Candidate
		score     float64
	}
	var candidates []ranked
	for i, left := range tokens {
		for _, right := range tokens[i+1:] {
			position := 1 - positionJSD(stats.Positions[left], stats.Positions[right])
			leftContext := cosineSimilarity(stats.Left[left], stats.Left[right])
			rightContext := cosineSimilarity(stats.Right[left], stats.Right[right])
			similarity := (position + leftContext + rightContext) / 3
			if similarity < minSimilarity {
				continue
			}
			minimumCount := min(stats.Counts[left], stats.Counts[right])
			leftCoverage := ratioInt(sumIntCounts(stats.Positions[left]), stats.Counts[left])
			rightCoverage := ratioInt(sumIntCounts(stats.Positions[right]), stats.Counts[right])
			reliability := float64(minimumCount) / float64(minimumCount+10)
			reliability *= math.Sqrt(leftCoverage * rightCoverage)
			candidates = append(candidates, ranked{candidate: normalization.Candidate{
				TokenA: left, TokenB: right, CountA: stats.Counts[left], CountB: stats.Counts[right],
				Similarity: similarity, PositionSimilarity: position,
				LeftContextSimilarity: leftContext, RightContextSimilarity: rightContext,
			}, score: similarity * reliability})
		}
	}
	sort.Slice(candidates, func(i, j int) bool {
		if candidates[i].score != candidates[j].score {
			return candidates[i].score > candidates[j].score
		}
		if candidates[i].candidate.TokenA != candidates[j].candidate.TokenA {
			return candidates[i].candidate.TokenA < candidates[j].candidate.TokenA
		}
		return candidates[i].candidate.TokenB < candidates[j].candidate.TokenB
	})
	if limit > 0 && len(candidates) > limit {
		candidates = candidates[:limit]
	}
	result := make([]normalization.Candidate, len(candidates))
	for i := range candidates {
		result[i] = candidates[i].candidate
	}
	return result
}

func positionJSD(left, right map[int]int) float64 {
	leftTotal, rightTotal := sumIntCounts(left), sumIntCounts(right)
	if leftTotal == 0 || rightTotal == 0 {
		return 1
	}
	positions := make(map[int]bool)
	for position := range left {
		positions[position] = true
	}
	for position := range right {
		positions[position] = true
	}
	ordered := make([]int, 0, len(positions))
	for position := range positions {
		ordered = append(ordered, position)
	}
	sort.Ints(ordered)
	value := 0.0
	for _, position := range ordered {
		p := ratioInt(left[position], leftTotal)
		q := ratioInt(right[position], rightTotal)
		middle := (p + q) / 2
		if p > 0 {
			value += .5 * p * math.Log2(p/middle)
		}
		if q > 0 {
			value += .5 * q * math.Log2(q/middle)
		}
	}
	return clamp(value, 0, 1)
}

func cosineSimilarity(left, right map[string]int) float64 {
	if len(left) == 0 || len(right) == 0 {
		return 0
	}
	dot, leftNorm, rightNorm := 0.0, 0.0, 0.0
	for token, count := range left {
		dot += float64(count * right[token])
		leftNorm += float64(count * count)
	}
	for _, count := range right {
		rightNorm += float64(count * count)
	}
	if leftNorm == 0 || rightNorm == 0 {
		return 0
	}
	return clamp(dot/(math.Sqrt(leftNorm)*math.Sqrt(rightNorm)), 0, 1)
}

func sumIntCounts[K comparable](counts map[K]int) int {
	total := 0
	for _, count := range counts {
		total += count
	}
	return total
}

func ratioInt(numerator, denominator int) float64 {
	if numerator <= 0 || denominator <= 0 {
		return 0
	}
	return float64(numerator) / float64(denominator)
}

func clamp(value, minimum, maximum float64) float64 {
	if value < minimum {
		return minimum
	}
	if value > maximum {
		return maximum
	}
	return value
}
