package positionalcontinuation

import (
	"math"
	"math/rand"
	"sort"
)

// permuteLabelsWithinBlocks shuffles the position-label values among
// occurrences independently within each physical block (task23 sections
// 26-27): it preserves continuation-token identity per occurrence, the
// per-block label composition, and block membership, while destroying the
// label<->continuation pairing. Nothing is ever permuted across blocks.
func permuteLabelsWithinBlocks(blockIDs, labels []string, r *rand.Rand) []string {
	out := make([]string, len(labels))
	copy(out, labels)
	byBlock := map[string][]int{}
	for i, b := range blockIDs {
		byBlock[b] = append(byBlock[b], i)
	}
	keys := make([]string, 0, len(byBlock))
	for b := range byBlock {
		keys = append(keys, b)
	}
	sort.Strings(keys)
	for _, b := range keys {
		idxs := byBlock[b]
		vals := make([]string, len(idxs))
		for j, idx := range idxs {
			vals[j] = out[idx]
		}
		r.Shuffle(len(vals), func(a, c int) { vals[a], vals[c] = vals[c], vals[a] })
		for j, idx := range idxs {
			out[idx] = vals[j]
		}
	}
	return out
}

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

	hGlobal := countEntropyBits(countMap(xs))
	globalCheyP := 0.0
	if len(xs) > 0 {
		globalCheyP = float64(countMap(xs)[FrozenChey]) / float64(len(xs))
	}
	observedMI := mutualInformationBits(xs, labels)

	type catObs struct {
		n, cheyN     int
		h            float64
		entropyDiff  float64
		effectiveCnt float64
		unique       int
		cheyP        float64
		enrichment   float64
	}
	obsByCat := map[string]catObs{}
	for _, cat := range categories {
		var catXs []string
		for i, l := range labels {
			if l == cat {
				catXs = append(catXs, xs[i])
			}
		}
		counts := countMap(catXs)
		h := countEntropyBits(counts)
		o := catObs{n: len(catXs), h: h, entropyDiff: hGlobal - h, unique: len(counts)}
		o.effectiveCnt = pow2(h)
		o.cheyN = counts[FrozenChey]
		if o.n > 0 {
			o.cheyP = float64(o.cheyN) / float64(o.n)
		}
		if globalCheyP > 0 {
			o.enrichment = o.cheyP / globalCheyP
		}
		obsByCat[cat] = o
	}

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
		permLabels := permuteLabelsWithinBlocks(blockIDs, labels, r)
		miNull = append(miNull, mutualInformationBits(xs, permLabels))
		for _, cat := range categories {
			var catXs []string
			for i, l := range permLabels {
				if l == cat {
					catXs = append(catXs, xs[i])
				}
			}
			counts := countMap(catXs)
			h := countEntropyBits(counts)
			entropyDiffNull[cat] = append(entropyDiffNull[cat], hGlobal-h)
			cheyP := 0.0
			if len(catXs) > 0 {
				cheyP = float64(counts[FrozenChey]) / float64(len(catXs))
			}
			enrich := 0.0
			if globalCheyP > 0 {
				enrich = cheyP / globalCheyP
			}
			enrichmentNull[cat] = append(enrichmentNull[cat], enrich)
		}
	}

	mean, sd := meanSD(miNull)
	result := positionalTestResult{
		Dependence: PositionDependenceRow{
			PositionVariable: variable, ObservedMIBits: observedMI,
			NullMeanMIBits: mean, NullSDMIBits: sd, Permutations: permutations,
			EmpiricalP: empiricalP(observedMI, miNull),
		},
	}
	for _, cat := range categories {
		o := obsByCat[cat]
		result.Entropy = append(result.Entropy, PositionalEntropyRow{
			PositionVariable: variable, Stratum: cat, OccurrenceCount: o.n,
			EntropyBits: o.h, EntropyGlobalBits: hGlobal, EntropyDifference: o.entropyDiff,
			EffectiveContinuationCount: o.effectiveCnt, UniqueContinuations: o.unique,
			EmpiricalP: empiricalP(o.entropyDiff, entropyDiffNull[cat]), Permutations: permutations,
		})
		result.CheyEffect = append(result.CheyEffect, CheyEffectRow{
			PositionVariable: variable, Stratum: cat, OccurrenceCount: o.n, CheyCount: o.cheyN,
			PCheyGivenPosition: o.cheyP, PCheyGlobal: globalCheyP, PositionalEnrichment: o.enrichment,
			EmpiricalP: empiricalP(o.enrichment, enrichmentNull[cat]), Permutations: permutations,
		})
	}
	return result
}

func pow2(h float64) float64 { return math.Pow(2, h) }
