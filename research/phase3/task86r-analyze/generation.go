package main

import "math"

// GenerationResult is one (candidate, transcription, scale) generative
// validation run (G1_STABILITY_CONTRACT.md).
type GenerationResult struct {
	Scale        float64
	Replicates   int
	Converged    bool
	MedianAtStop map[string]float64
	CV           map[string]float64
	ExcessiveCV  bool
	MetricPass   map[string]bool
	EditPass     bool
	EditCount    int
	LexicalPass  bool
	LexicalCount int
	StructuralOverallPass bool
}

func tokenCountFormula(scale float64, n int) int {
	c := int(math.Floor(scale*float64(n) + 0.5))
	if c < 1 {
		c = 1
	}
	return c
}

func convergencePass(current, previous map[string]float64) bool {
	for metric, cur := range current {
		prev, ok := previous[metric]
		if !ok || !isFinite(cur) || !isFinite(prev) {
			return false
		}
		diff := math.Abs(cur - prev)
		tol := 1e-4 + 0.01*math.Max(math.Abs(cur), math.Abs(prev))
		if diff > tol {
			return false
		}
	}
	return true
}

func medianPerMetric(vals []map[string]float64) map[string]float64 {
	out := map[string]float64{}
	for _, metric := range StructuralMetricIDs {
		var xs []float64
		for _, v := range vals {
			if x, ok := v[metric]; ok && isFinite(x) {
				xs = append(xs, x)
			}
		}
		if len(xs) > 0 {
			out[metric] = median(xs)
		}
	}
	return out
}

// runGeneration generates up to 32 replicate populations at one scale,
// evaluating the frozen convergence/stability rules and F2 metric medians.
func runGeneration(namespace, transcription, class, candID string, model FittedModel, scale float64, partitionN int, rawToGlyphs func(string) []string, alias *GlyphAlias, workDir string) GenerationResult {
	n := tokenCountFormula(scale, partitionN)
	var perReplicate []map[string]float64
	generateUpTo := func(target int) {
		for r := len(perReplicate); r < target; r++ {
			seed := SeedFields{Namespace: namespace, ModelClass: class, CandidateID: candID, CorpusID: transcription, Transcription: transcription, Partition: "HELDOUT", Scale: scale, Replicate: r}
			prng := NewSeededPRNG(seed)
			pops := make([][]string, 0, n)
			for i := 0; i < n; i++ {
				g := model.Generate(prng)
				if g.NonGenerative {
					continue
				}
				pops = append(pops, glyphsForGenerated(model, g, rawToGlyphs))
			}
			f2, ok, _ := StructuralMetrics(alias, pops, int64(seed.Seed()), workDir)
			if !ok {
				f2 = map[string]float64{}
			}
			perReplicate = append(perReplicate, f2)
		}
	}

	generateUpTo(4)
	m4 := medianPerMetric(perReplicate)
	generateUpTo(8)
	m8 := medianPerMetric(perReplicate)
	c8 := convergencePass(m8, m4)

	res := GenerationResult{Scale: scale}
	if c8 {
		generateUpTo(16)
		m16 := medianPerMetric(perReplicate)
		c16 := convergencePass(m16, m8)
		if c16 {
			res.Replicates = 16
			res.Converged = true
			res.MedianAtStop = m16
		} else {
			generateUpTo(32)
			m32 := medianPerMetric(perReplicate)
			c32 := convergencePass(m32, m16)
			res.Replicates = 32
			res.Converged = c16 && c32
			res.MedianAtStop = m32
		}
	} else {
		generateUpTo(16)
		m16 := medianPerMetric(perReplicate)
		c16 := convergencePass(m16, m8)
		generateUpTo(32)
		m32 := medianPerMetric(perReplicate)
		c32 := convergencePass(m32, m16)
		res.Replicates = 32
		res.Converged = c16 && c32
		res.MedianAtStop = m32
	}

	res.CV = map[string]float64{}
	res.ExcessiveCV = false
	last32 := perReplicate
	if len(last32) > 32 {
		last32 = last32[:32]
	}
	for _, metric := range StructuralMetricIDs {
		var xs []float64
		for _, v := range last32 {
			if x, ok := v[metric]; ok && isFinite(x) {
				xs = append(xs, x)
			}
		}
		if len(xs) >= 2 {
			cv := coefficientOfVariation(xs)
			res.CV[metric] = cv
			if cv > 0.25 {
				res.ExcessiveCV = true
			}
		}
	}
	return res
}
