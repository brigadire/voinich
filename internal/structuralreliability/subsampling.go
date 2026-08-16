package structuralreliability

import (
	"math/rand"
	"sort"

	"zcore.dev/voinich/internal/profilestability"
)

// baseSubsampleSizes are the sample sizes task section 12 asks for by name.
// Config.SubsampleMinFullCount filters this list down when it is smaller
// than 160, and a token additionally skips any size larger than its own
// occurrence count (section 13: "если исходных occurrences достаточно").
var baseSubsampleSizes = []int{10, 20, 40, 80, 160}

func subsampleSizes(minFullCount int) []int {
	var sizes []int
	for _, size := range baseSubsampleSizes {
		if size <= minFullCount {
			sizes = append(sizes, size)
		}
	}
	return sizes
}

type subsampleObservation struct {
	position, left, right float64
}

// runSubsampling is the controlled, same-token experiment of sections 12-17:
// for every sufficiently frequent token it repeatedly rebuilds a profile from
// a random subset of that token's real occurrences (never a synthetic text)
// and compares it, via the unmodified profilestability.Compare, against the
// token's full-corpus reference profile.
func runSubsampling(occurrences map[string][]Occurrence, fullProfiles map[string]profilestability.Profile, fullWs map[string]profilestability.SortedProfile, config Config) Subsampling {
	var tokens []string
	for token, profile := range fullProfiles {
		if profile.Count >= config.SubsampleMinFullCount {
			tokens = append(tokens, token)
		}
	}
	sort.Strings(tokens)
	sizes := subsampleSizes(config.SubsampleMinFullCount)

	rng := rand.New(rand.NewSource(config.SubsampleSeed))
	perSize := make(map[int][]subsampleObservation, len(sizes))
	perTokenSize := make(map[string]map[int][]subsampleObservation, len(tokens))
	for _, token := range tokens {
		referenceSorted := fullWs[token]
		occurrencesForToken := occurrences[token]
		perTokenSize[token] = make(map[int][]subsampleObservation, len(sizes))
		for _, size := range sizes {
			if len(occurrencesForToken) < size {
				continue
			}
			observations := make([]subsampleObservation, 0, config.SubsampleRuns)
			for run := 0; run < config.SubsampleRuns; run++ {
				sample := SampleOccurrences(occurrencesForToken, size, rng)
				profile := ProfileFromOccurrences(sample)
				components := profilestability.CompareSorted(profilestability.Precompute(profile), referenceSorted)
				observations = append(observations, subsampleObservation{position: components.PositionSimilarity, left: components.LeftSimilarity, right: components.RightSimilarity})
			}
			perTokenSize[token][size] = observations
			perSize[size] = append(perSize[size], observations...)
		}
	}

	tokensForSize := make(map[int]int, len(sizes))
	for _, size := range sizes {
		for _, token := range tokens {
			if _, ok := perTokenSize[token][size]; ok {
				tokensForSize[size]++
			}
		}
	}

	result := Subsampling{MinFullCount: config.SubsampleMinFullCount, SampleSizes: sizes, Runs: config.SubsampleRuns, Tokens: len(tokens)}
	for _, size := range sizes {
		observations := perSize[size]
		result.Results = append(result.Results, SubsamplingResult{
			SampleSize:   size,
			Position:     componentSizeStat(size, tokensForSize[size], observations, func(o subsampleObservation) float64 { return o.position }),
			LeftContext:  componentSizeStat(size, tokensForSize[size], observations, func(o subsampleObservation) float64 { return o.left }),
			RightContext: componentSizeStat(size, tokensForSize[size], observations, func(o subsampleObservation) float64 { return o.right }),
		})
	}
	for _, token := range tokens {
		entry := PerTokenSubsampling{Token: token, FullCount: fullProfiles[token].Count}
		for _, size := range sizes {
			observations, ok := perTokenSize[token][size]
			if !ok {
				continue
			}
			entry.SampleSizes = append(entry.SampleSizes, PerTokenSampleSize{
				N: size, Runs: len(observations),
				PositionMean: meanOf(observations, func(o subsampleObservation) float64 { return o.position }),
				LeftMean:     meanOf(observations, func(o subsampleObservation) float64 { return o.left }),
				RightMean:    meanOf(observations, func(o subsampleObservation) float64 { return o.right }),
			})
		}
		result.PerToken = append(result.PerToken, entry)
	}
	for _, size := range sizes {
		var positionMeans, leftMeans, rightMeans []float64
		for _, token := range tokens {
			observations, ok := perTokenSize[token][size]
			if !ok {
				continue
			}
			positionMeans = append(positionMeans, meanOf(observations, func(o subsampleObservation) float64 { return o.position }))
			leftMeans = append(leftMeans, meanOf(observations, func(o subsampleObservation) float64 { return o.left }))
			rightMeans = append(rightMeans, meanOf(observations, func(o subsampleObservation) float64 { return o.right }))
		}
		result.Heterogeneity = append(result.Heterogeneity, Heterogeneity{
			SampleSize:   size,
			Position:     heterogeneityComponent(positionMeans),
			LeftContext:  heterogeneityComponent(leftMeans),
			RightContext: heterogeneityComponent(rightMeans),
		})
	}
	return result
}

func componentSizeStat(size, tokenCount int, observations []subsampleObservation, field func(subsampleObservation) float64) ComponentSizeStat {
	values := make([]float64, len(observations))
	for i, observation := range observations {
		values[i] = field(observation)
	}
	stat := SummarizeStat(values)
	return ComponentSizeStat{
		Size: size, Tokens: tokenCount, Runs: len(values), MeanSimilarity: stat.Mean, MedianSimilarity: stat.Median,
		Percentile05: PercentileOf(values, .05), Percentile95: PercentileOf(values, .95),
	}
}

func heterogeneityComponent(means []float64) HeterogeneityComponent {
	return HeterogeneityComponent{
		Percentile10: PercentileOf(means, .10), Percentile25: PercentileOf(means, .25),
		Median: PercentileOf(means, .50), Percentile75: PercentileOf(means, .75), Percentile90: PercentileOf(means, .90),
	}
}

func meanOf(observations []subsampleObservation, field func(subsampleObservation) float64) float64 {
	if len(observations) == 0 {
		return 0
	}
	total := 0.0
	for _, observation := range observations {
		total += field(observation)
	}
	return total / float64(len(observations))
}
