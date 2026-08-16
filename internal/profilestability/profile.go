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
	return CompareSorted(Precompute(left), Precompute(right))
}

// SortedProfile is a Profile plus each context map's keys pre-sorted once.
// Compare/positionJSD/cosine all accumulate over keys in sorted order (for
// deterministic map-independent output), so a profile compared repeatedly -
// an O(E^2) nearest-neighbor sweep, a bootstrap replicate's candidate pairs,
// or an O(V^2) all-pairs ranking, all of which reuse the same profile across
// many Compare calls - re-sorted the same keys on every single call. Caching
// the sort once per profile and reusing it via CompareSorted/PrecomputeAll
// is the shared fix referenced by PERFORMANCE_AUDIT.md's cross-cutting
// `profilestability.Compare` finding (task27 item 10); it changes nothing
// about Compare's arithmetic or accumulation order (see
// sorted_hoist_test.go's reference-equivalence tests).
type SortedProfile struct {
	profile   Profile
	positions []int
	leftKeys  []string
	rightKeys []string
}

// Precompute builds a SortedProfile for a single profile.
func Precompute(p Profile) SortedProfile {
	return SortedProfile{
		profile:   p,
		positions: sortedIntKeys(p.Positions),
		leftKeys:  sortedStringKeys(p.Left),
		rightKeys: sortedStringKeys(p.Right),
	}
}

// PrecomputeAll builds a SortedProfile for every entry of profiles, so every
// downstream Compare against any of these profiles reuses the same sorted
// key slices instead of re-sorting.
func PrecomputeAll(profiles map[string]Profile) map[string]SortedProfile {
	result := make(map[string]SortedProfile, len(profiles))
	for token, p := range profiles {
		result[token] = Precompute(p)
	}
	return result
}

// CompareSorted is Compare's exact algorithm, driven by pre-sorted key
// slices instead of sorting left/right's context maps on every call.
func CompareSorted(left, right SortedProfile) Components {
	result := Components{
		PositionSimilarity: 1 - positionJSDSorted(left, right),
		LeftSimilarity:     cosineSorted(left.profile.Left, left.leftKeys, right.profile.Left, right.leftKeys),
		RightSimilarity:    cosineSorted(left.profile.Right, left.rightKeys, right.profile.Right, right.rightKeys),
	}
	result.Similarity = (result.PositionSimilarity + result.LeftSimilarity + result.RightSimilarity) / 3
	return result
}

func sortedIntKeys(m map[int]int) []int {
	if len(m) == 0 {
		return nil
	}
	keys := make([]int, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Ints(keys)
	return keys
}

func sortedStringKeys(m map[string]int) []string {
	if len(m) == 0 {
		return nil
	}
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
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

// NearestNeighborsIn is NearestNeighbors driven by a precomputed
// map[string]SortedProfile (see PrecomputeAll) instead of raw Profiles, so
// neither the fixed token's profile nor eligible's other profiles are
// re-sorted across the O(len(eligible)) Compare calls this makes - or across
// the repeated calls a caller like buildTokenMetrics/buildAllNeighbors makes
// for every token in eligible.
func NearestNeighborsIn(ws map[string]SortedProfile, token string, eligible []string, limit int) []Neighbor {
	left := ws[token]
	neighbors := make([]Neighbor, 0, len(eligible)-1)
	for _, other := range eligible {
		if other == token {
			continue
		}
		neighbors = append(neighbors, Neighbor{Token: other, Components: CompareSorted(left, ws[other])})
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

// NearestNeighbors is NearestNeighborsIn for callers that only need one
// token's neighbors and have no workspace to share across calls.
func NearestNeighbors(profiles map[string]Profile, token string, eligible []string, limit int) []Neighbor {
	return NearestNeighborsIn(PrecomputeAll(profiles), token, eligible, limit)
}

func positionJSDSorted(left, right SortedProfile) float64 {
	leftPositions, rightPositions := left.profile.Positions, right.profile.Positions
	leftTotal, rightTotal := sumCounts(leftPositions), sumCounts(rightPositions)
	if leftTotal == 0 || rightTotal == 0 {
		return 1
	}
	value := 0.0
	lp, rp := left.positions, right.positions
	i, j := 0, 0
	for i < len(lp) || j < len(rp) {
		var position int
		switch {
		case i >= len(lp):
			position = rp[j]
		case j >= len(rp):
			position = lp[i]
		case lp[i] <= rp[j]:
			position = lp[i]
		default:
			position = rp[j]
		}
		p := float64(leftPositions[position]) / float64(leftTotal)
		q := float64(rightPositions[position]) / float64(rightTotal)
		middle := (p + q) / 2
		if p > 0 {
			value += .5 * p * math.Log2(p/middle)
		}
		if q > 0 {
			value += .5 * q * math.Log2(q/middle)
		}
		if i < len(lp) && lp[i] == position {
			i++
		}
		if j < len(rp) && rp[j] == position {
			j++
		}
	}
	return clamp(value)
}

func cosineSorted(left map[string]int, leftKeys []string, right map[string]int, rightKeys []string) float64 {
	if len(left) == 0 || len(right) == 0 {
		return 0
	}
	dot, leftNorm := 0.0, 0.0
	for _, key := range leftKeys {
		count := left[key]
		dot += float64(count * right[key])
		leftNorm += float64(count * count)
	}
	rightNorm := 0.0
	for _, key := range rightKeys {
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
