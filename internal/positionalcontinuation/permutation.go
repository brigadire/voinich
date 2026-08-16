package positionalcontinuation

import (
	"math"
	"math/rand"
)

// positionalTestResult bundles Part E/F/G's output for one positional
// variable ("line_position" or "block_position_coarse").
type positionalTestResult struct {
	Dependence PositionDependenceRow
	Entropy    []PositionalEntropyRow
	CheyEffect []CheyEffectRow
}

// runPositionalTests implements task23 Parts E, F and G for one positional
// variable: the primary I(X;position) permutation test (section 24-29), the
// per-category entropy-narrowing effect (sections 30-35), and the
// chey-specific positional enrichment (sections 36-41) - all three sharing
// one within-block label-permutation null (section 26), since they are three
// descriptive facets of the same "does position change what follows s aiin"
// question.
func runPositionalTests(occs []SAiinOccurrence, variable string, categories []string, permutations int, seed int64) positionalTestResult {
	var xs, labels, blockIDs []string
	for _, o := range occs {
		if o.X == "" {
			continue
		}
		xs = append(xs, o.X)
		blockIDs = append(blockIDs, o.Block)
		if variable == "line_position" {
			labels = append(labels, o.LineCategory)
		} else {
			labels = append(labels, o.BlockBinCoarse)
		}
	}

	ws := newPositionalWorkspace(xs, blockIDs, labels)
	hGlobal, globalCheyP := ws.hGlobalAndCheyP()
	observed := ws.statsFor(ws.labelIdx, categories)

	r := rand.New(rand.NewSource(seed))
	miNull := make([]float64, 0, permutations)
	entropyDiffNull := map[string][]float64{}
	enrichmentNull := map[string][]float64{}
	for _, cat := range categories {
		entropyDiffNull[cat] = make([]float64, 0, permutations)
		enrichmentNull[cat] = make([]float64, 0, permutations)
	}
	// permutations<=0 (used by the jackknife, task23 Part L) skips the null
	// entirely and returns point estimates only - jackknife stability is
	// judged by how the point estimate moves across realizations, not by a
	// fresh 10000-permutation p-value on every one-block-removed subset.
	for p := 0; p < permutations; p++ {
		permLabelIdx := ws.permute(r)
		s := ws.statsFor(permLabelIdx, categories)
		miNull = append(miNull, s.mi)
		for ci, cat := range categories {
			c := s.cat[ci]
			entropyDiffNull[cat] = append(entropyDiffNull[cat], hGlobal-c.h)
			enrich := 0.0
			if globalCheyP > 0 {
				enrich = c.cheyP / globalCheyP
			}
			enrichmentNull[cat] = append(enrichmentNull[cat], enrich)
		}
	}

	mean, sd := meanSD(miNull)
	result := positionalTestResult{
		Dependence: PositionDependenceRow{
			PositionVariable: variable, ObservedMIBits: observed.mi,
			NullMeanMIBits: mean, NullSDMIBits: sd, Permutations: permutations,
			EmpiricalP: empiricalP(observed.mi, miNull),
		},
	}
	for ci, cat := range categories {
		o := observed.cat[ci]
		entropyDiff := hGlobal - o.h
		enrichment := 0.0
		if globalCheyP > 0 {
			enrichment = o.cheyP / globalCheyP
		}
		result.Entropy = append(result.Entropy, PositionalEntropyRow{
			PositionVariable: variable, Stratum: cat, OccurrenceCount: o.n,
			EntropyBits: o.h, EntropyGlobalBits: hGlobal, EntropyDifference: entropyDiff,
			EffectiveContinuationCount: pow2(o.h), UniqueContinuations: o.unique,
			EmpiricalP: empiricalP(entropyDiff, entropyDiffNull[cat]), Permutations: permutations,
		})
		result.CheyEffect = append(result.CheyEffect, CheyEffectRow{
			PositionVariable: variable, Stratum: cat, OccurrenceCount: o.n, CheyCount: o.cheyN,
			PCheyGivenPosition: o.cheyP, PCheyGlobal: globalCheyP, PositionalEnrichment: enrichment,
			EmpiricalP: empiricalP(enrichment, enrichmentNull[cat]), Permutations: permutations,
		})
	}
	return result
}

func pow2(h float64) float64 { return math.Pow(2, h) }
