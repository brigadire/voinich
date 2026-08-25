package main

import (
	"math"
	"sort"
)

// leafOrderedPrefixes returns, for each frozen fraction 0.25/0.50/0.75/1.00,
// the leaf-aligned DEVELOPMENT prefix (numeric leaf order) whose cumulative
// TOKEN count first reaches ceil(fraction*N_dev); duplicate boundaries
// (distinct fractions landing on the same leaf count) collapse to one
// entry (G1_STABILITY_CONTRACT.md).
func leafOrderedPrefixes(dev []TokenOccurrence) []struct {
	Fraction float64
	Occ      []TokenOccurrence
} {
	leafOrder := map[string][]TokenOccurrence{}
	for _, o := range dev {
		leafOrder[o.Leaf] = append(leafOrder[o.Leaf], o)
	}
	leaves := make([]string, 0, len(leafOrder))
	for l := range leafOrder {
		leaves = append(leaves, l)
	}
	sort.Slice(leaves, func(i, j int) bool {
		ni, nj := leafNumber(leaves[i]), leafNumber(leaves[j])
		if ni != nj {
			return ni < nj
		}
		return leaves[i] < leaves[j]
	})
	n := len(dev)
	fractions := []float64{0.25, 0.50, 0.75, 1.00}
	var out []struct {
		Fraction float64
		Occ      []TokenOccurrence
	}
	seenCount := map[int]bool{}
	for _, f := range fractions {
		target := int(math.Ceil(f * float64(n)))
		cum := 0
		var prefix []TokenOccurrence
		for _, l := range leaves {
			prefix = append(prefix, leafOrder[l]...)
			cum += len(leafOrder[l])
			if cum >= target {
				break
			}
		}
		if seenCount[len(prefix)] {
			continue
		}
		seenCount[len(prefix)] = true
		out = append(out, struct {
			Fraction float64
			Occ      []TokenOccurrence
		}{Fraction: f, Occ: prefix})
	}
	return out
}

// ComplexityGrowthResult is the Theil-Sen point slope and deterministic
// bootstrap lower 95% CI endpoint of log2 Complexity vs log2 scored units
// across the nested leaf prefixes.
type ComplexityGrowthResult struct {
	PointSlope   float64
	LowerCI      float64
	Unbounded    bool
	Points       int
}

const complexityGrowthBootstrapResamples = 1000

func computeComplexityGrowth(namespace, transcription, candidateID, modelClass string, dev []TokenOccurrence, cand Candidate) ComplexityGrowthResult {
	prefixes := leafOrderedPrefixes(dev)
	var xs, ys []float64
	for _, p := range prefixes {
		bitsReal := bitsPerRealParameter(len(p.Occ))
		model := FitCandidate(p.Occ, cand, bitsReal)
		if failed, _ := model.TrainingFailed(); failed {
			continue
		}
		scoredUnits := 0
		for _, o := range p.Occ {
			scoredUnits += model.ScoredUnits(o.Glyphs)
		}
		complexity := model.Complexity().Total()
		if complexity <= 0 || scoredUnits <= 0 {
			continue
		}
		xs = append(xs, log2(float64(scoredUnits)))
		ys = append(ys, log2(complexity))
	}
	if len(xs) < 2 {
		return ComplexityGrowthResult{PointSlope: math.NaN(), LowerCI: math.NaN(), Points: len(xs)}
	}
	point := theilSenSlope(xs, ys)
	seed := SeedFields{Namespace: namespace, ModelClass: modelClass, CandidateID: candidateID, CorpusID: "COMPLEXITY_GROWTH", Transcription: transcription, Partition: "DEVELOPMENT", Scale: 1.0, Replicate: 0}
	prng := NewSeededPRNG(seed)
	var slopes []float64
	for r := 0; r < complexityGrowthBootstrapResamples; r++ {
		n := len(xs)
		rx := make([]float64, n)
		ry := make([]float64, n)
		distinct := map[float64]bool{}
		for i := 0; i < n; i++ {
			idx := int(prng.Float64() * float64(n))
			if idx >= n {
				idx = n - 1
			}
			rx[i], ry[i] = xs[idx], ys[idx]
			distinct[xs[idx]] = true
		}
		if len(distinct) < 2 {
			continue
		}
		slopes = append(slopes, theilSenSlope(rx, ry))
	}
	lowerCI := math.NaN()
	if len(slopes) > 0 {
		lowerCI = percentileNearestRank(slopes, 0.025)
	}
	unbounded := point > 1.10 && lowerCI > 1.00
	return ComplexityGrowthResult{PointSlope: point, LowerCI: lowerCI, Unbounded: unbounded, Points: len(xs)}
}
