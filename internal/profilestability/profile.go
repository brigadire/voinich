package profilestability

import (
	"math"
	"sort"

	"zcore.dev/voinich/internal/validation"
)

func BuildProfiles(corpus validation.Corpus) map[string]Profile {
	profiles := make(map[string]Profile)
	for _, line := range corpus.Lines {
		for position, token := range line.Tokens {
			profile := profiles[token]
			if profile.Positions == nil {
				profile.Positions = make(map[int]int)
				profile.Left = make(map[string]int)
				profile.Right = make(map[string]int)
			}
			profile.Count++
			profile.Positions[position]++
			if position > 0 {
				profile.Left[line.Tokens[position-1]]++
			}
			if position+1 < len(line.Tokens) {
				profile.Right[line.Tokens[position+1]]++
			}
			profiles[token] = profile
		}
	}
	return profiles
}

func Compare(left, right Profile) Components {
	result := Components{
		PositionSimilarity: 1 - positionJSD(left.Positions, right.Positions),
		LeftSimilarity:     cosine(left.Left, right.Left),
		RightSimilarity:    cosine(left.Right, right.Right),
	}
	result.Similarity = (result.PositionSimilarity + result.LeftSimilarity + result.RightSimilarity) / 3
	return result
}

func Eligible(profiles map[string]Profile, minCount int) []string {
	result := make([]string, 0)
	for token, profile := range profiles {
		if profile.Count >= minCount {
			result = append(result, token)
		}
	}
	sort.Strings(result)
	return result
}

func NearestNeighbors(profiles map[string]Profile, token string, eligible []string, limit int) []Neighbor {
	neighbors := make([]Neighbor, 0, len(eligible)-1)
	for _, other := range eligible {
		if other == token {
			continue
		}
		neighbors = append(neighbors, Neighbor{Token: other, Components: Compare(profiles[token], profiles[other])})
	}
	sort.Slice(neighbors, func(i, j int) bool {
		if neighbors[i].Similarity != neighbors[j].Similarity {
			return neighbors[i].Similarity > neighbors[j].Similarity
		}
		return neighbors[i].Token < neighbors[j].Token
	})
	if limit > 0 && len(neighbors) > limit {
		neighbors = neighbors[:limit]
	}
	for i := range neighbors {
		neighbors[i].Rank = i + 1
	}
	return neighbors
}

func positionJSD(left, right map[int]int) float64 {
	leftTotal, rightTotal := sumCounts(left), sumCounts(right)
	if leftTotal == 0 || rightTotal == 0 {
		return 1
	}
	positions := make(map[int]bool, len(left)+len(right))
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
		p := float64(left[position]) / float64(leftTotal)
		q := float64(right[position]) / float64(rightTotal)
		middle := (p + q) / 2
		if p > 0 {
			value += .5 * p * math.Log2(p/middle)
		}
		if q > 0 {
			value += .5 * q * math.Log2(q/middle)
		}
	}
	return clamp(value)
}

func cosine(left, right map[string]int) float64 {
	if len(left) == 0 || len(right) == 0 {
		return 0
	}
	keys := make([]string, 0, len(left))
	for key := range left {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	dot, leftNorm := 0.0, 0.0
	for _, key := range keys {
		count := left[key]
		dot += float64(count * right[key])
		leftNorm += float64(count * count)
	}
	keys = keys[:0]
	for key := range right {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	rightNorm := 0.0
	for _, key := range keys {
		rightNorm += float64(right[key] * right[key])
	}
	if leftNorm == 0 || rightNorm == 0 {
		return 0
	}
	return clamp(dot / (math.Sqrt(leftNorm) * math.Sqrt(rightNorm)))
}

func sumCounts[K comparable](counts map[K]int) int {
	total := 0
	for _, count := range counts {
		total += count
	}
	return total
}

func clamp(value float64) float64 {
	if value < 0 {
		return 0
	}
	if value > 1 {
		return 1
	}
	return value
}
