package structuralreliability

import (
	"math/rand"
	"sort"

	"zcore.dev/voinich/internal/normalization"
	"zcore.dev/voinich/internal/profilestability"
	"zcore.dev/voinich/internal/validation"
)

type pairKey struct{ a, b string }

func makePair(a, b string) pairKey {
	if a > b {
		a, b = b, a
	}
	return pairKey{a, b}
}

func sortedPairs(pairs map[pairKey]bool) []pairKey {
	result := make([]pairKey, 0, len(pairs))
	for pair := range pairs {
		result = append(result, pair)
	}
	sort.Slice(result, func(i, j int) bool {
		if result[i].a != result[j].a {
			return result[i].a < result[j].a
		}
		return result[i].b < result[j].b
	})
	return result
}

// specialReferencePairs are always included in the master candidate set,
// regardless of whether they also arise from neighbor or class membership,
// because task section 25 requires reliability diagnostics for them by name.
var specialReferencePairs = [][2]string{{"chedy", "shedy"}, {"qokedy", "qokeedy"}}

// buildMasterPairs mirrors structural-profile-stability's own candidate-pair
// construction (full-corpus similarity>=threshold, top-K neighbor pairs,
// reference-class member pairs) using only exported profilestability
// primitives - no similarity formula is duplicated, only the orchestration
// of which pairs to examine.
func buildMasterPairs(fullWs map[string]profilestability.SortedProfile, fullEligible []string, fullNeighbors map[string][]profilestability.Neighbor, referenceModel normalization.Model, threshold float64) map[pairKey]bool {
	candidates := make(map[pairKey]bool)
	for token, neighbors := range fullNeighbors {
		for _, neighbor := range neighbors {
			candidates[makePair(token, neighbor.Token)] = true
		}
	}
	for i, tokenA := range fullEligible {
		for _, tokenB := range fullEligible[i+1:] {
			if profilestability.CompareSorted(fullWs[tokenA], fullWs[tokenB]).Similarity >= threshold {
				candidates[makePair(tokenA, tokenB)] = true
			}
		}
	}
	for _, class := range referenceModel.Classes {
		if class.Size < 2 {
			continue
		}
		for i := 0; i < len(class.Members); i++ {
			for j := i + 1; j < len(class.Members); j++ {
				candidates[makePair(class.Members[i].Token, class.Members[j].Token)] = true
			}
		}
	}
	for _, pair := range specialReferencePairs {
		candidates[makePair(pair[0], pair[1])] = true
	}
	return candidates
}

// pairFoldSimilarities returns, for each pair, the TRAIN-fold similarities
// observed whenever both tokens are eligible (count>=minCount) in that fold -
// the same fold-level comparison structural-profile-stability performs, just
// scoped to the master candidate pairs.
func pairFoldSimilarities(pairs []pairKey, folds []foldProfiles, minCount int) map[pairKey][]float64 {
	result := make(map[pairKey][]float64, len(pairs))
	for _, pair := range pairs {
		var values []float64
		for _, fold := range folds {
			a, b := fold.trainProfiles[pair.a], fold.trainProfiles[pair.b]
			if a.Count < minCount || b.Count < minCount {
				continue
			}
			values = append(values, profilestability.CompareSorted(fold.trainWs[pair.a], fold.trainWs[pair.b]).Similarity)
		}
		result[pair] = values
	}
	return result
}

type bootstrapResult struct {
	Observations              int
	Mean                      float64
	Percentile025             float64
	Percentile975             float64
	CIWidth                   float64
	ProbabilityAboveThreshold float64
}

// runBootstrap resamples physical lines with replacement (preserving line
// count), exactly as structural-profile-stability's own bootstrap does, and
// records per-pair similarity distributions. It is computed once, at the
// base min-token-count, and reused (by filtering pairs) at every cumulative
// threshold - a pair's bootstrap availability does not change with the
// downstream frequency grouping.
func runBootstrap(corpus validation.Corpus, pairs []pairKey, runs int, seed int64, minCount int, threshold float64, progress func(string)) map[pairKey]bootstrapResult {
	values := make(map[pairKey][]float64, len(pairs))
	for _, pair := range pairs {
		values[pair] = nil
	}
	rng := rand.New(rand.NewSource(seed))
	for run := 0; run < runs; run++ {
		sample := validation.Corpus{Counts: make(map[string]int)}
		for index := 0; index < len(corpus.Lines); index++ {
			source := corpus.Lines[rng.Intn(len(corpus.Lines))]
			tokens := append([]string(nil), source.Tokens...)
			sample.Lines = append(sample.Lines, validation.Line{ID: index + 1, Tokens: tokens})
		}
		profiles := profilestability.BuildProfiles(sample)
		ws := profilestability.PrecomputeAll(profiles)
		for _, pair := range pairs {
			a, b := profiles[pair.a], profiles[pair.b]
			if a.Count < minCount || b.Count < minCount {
				continue
			}
			values[pair] = append(values[pair], profilestability.CompareSorted(ws[pair.a], ws[pair.b]).Similarity)
		}
		if progress != nil && ((run+1)%25 == 0 || run+1 == runs) {
			progress("structural-reliability bootstrap " + itoa(run+1) + "/" + itoa(runs))
		}
	}
	result := make(map[pairKey]bootstrapResult, len(pairs))
	for _, pair := range pairs {
		observed := values[pair]
		if len(observed) == 0 {
			result[pair] = bootstrapResult{}
			continue
		}
		distribution := profilestability.Summarize(observed, true)
		above := 0
		for _, value := range observed {
			if value >= threshold {
				above++
			}
		}
		result[pair] = bootstrapResult{
			Observations: len(observed), Mean: distribution.Mean,
			Percentile025: distribution.Percentile025, Percentile975: distribution.Percentile975,
			CIWidth:                   distribution.Percentile975 - distribution.Percentile025,
			ProbabilityAboveThreshold: float64(above) / float64(len(observed)),
		}
	}
	return result
}

func itoa(value int) string {
	if value == 0 {
		return "0"
	}
	digits := make([]byte, 0, 10)
	for value > 0 {
		digits = append(digits, byte('0'+value%10))
		value /= 10
	}
	for i, j := 0, len(digits)-1; i < j; i, j = i+1, j-1 {
		digits[i], digits[j] = digits[j], digits[i]
	}
	return string(digits)
}
