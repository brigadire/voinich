package positionalcontinuation

import (
	"math/rand"
	"reflect"
	"sort"
	"testing"
)

// referencePermuteLabelsWithinBlocks is permuteLabelsWithinBlocks exactly as
// it stood before the positionalWorkspace rewrite: it rebuilds byBlock via a
// map, sorted every call, even though block membership never changes across
// replicates.
func referencePermuteLabelsWithinBlocks(blockIDs, labels []string, r *rand.Rand) []string {
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

// referenceRunPositionalTests is runPositionalTests exactly as it stood
// before the positionalWorkspace rewrite.
func referenceRunPositionalTests(occs []SAiinOccurrence, variable string, categories []string, permutations int, seed int64) positionalTestResult {
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
	for p := 0; p < permutations; p++ {
		permLabels := referencePermuteLabelsWithinBlocks(blockIDs, labels, r)
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

// referencePermuteIsSWithinStrata is permuteIsSWithinStrata exactly as it
// stood before the stratifiedWorkspace rewrite.
func referencePermuteIsSWithinStrata(obs []stratifiedObs, r *rand.Rand) []bool {
	out := make([]bool, len(obs))
	for i, o := range obs {
		out[i] = o.isS
	}
	byStratum := map[string][]int{}
	for i, o := range obs {
		byStratum[o.stratum] = append(byStratum[o.stratum], i)
	}
	keys := make([]string, 0, len(byStratum))
	for k := range byStratum {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		idxs := byStratum[k]
		vals := make([]bool, len(idxs))
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

func referenceStatisticCheyAmongS(obs []stratifiedObs, isS []bool) float64 {
	n := 0.0
	for i, o := range obs {
		if isS[i] && o.isChey {
			n++
		}
	}
	return n
}

// referenceRunStratifiedPredecessorTest is runStratifiedPredecessorTest
// exactly as it stood before the stratifiedWorkspace rewrite.
func referenceRunStratifiedPredecessorTest(aiinOccs []AiinOccurrence, variable string, permutations int, seed int64) StratifiedPredecessorRow {
	var obs []stratifiedObs
	for _, o := range aiinOccs {
		if o.X == "" || !o.HasPredecessor {
			continue
		}
		cat := o.LineCategory
		if variable == "block_position_coarse" {
			cat = o.BlockBinCoarse
		}
		obs = append(obs, stratifiedObs{
			stratum: o.Block + "|" + cat, isS: o.PredecessorIsS, isChey: o.X == FrozenChey,
		})
	}
	baseline := make([]bool, len(obs))
	for i, o := range obs {
		baseline[i] = o.isS
	}
	observed := referenceStatisticCheyAmongS(obs, baseline)

	r := rand.New(rand.NewSource(seed))
	null := make([]float64, permutations)
	for p := 0; p < permutations; p++ {
		perm := referencePermuteIsSWithinStrata(obs, r)
		null[p] = referenceStatisticCheyAmongS(obs, perm)
	}
	mean, sd := meanSD(null)
	return StratifiedPredecessorRow{
		PositionVariable: variable, ObservedStatistic: observed,
		NullMeanStatistic: mean, NullSDStatistic: sd, Permutations: permutations,
		EmpiricalP: empiricalP(observed, null),
	}
}

// fixtureSAiinOccurrences builds a synthetic, non-degenerate SAiinOccurrence
// set: varied continuation tokens, blocks, and line/block position
// categories, so both the permutation null and the per-category statistics
// exercise realistic ties/overlaps.
func fixtureSAiinOccurrences(n int, seed int64) []SAiinOccurrence {
	r := rand.New(rand.NewSource(seed))
	xVocab := []string{"chey", "shey", "ol", "or", "dy", "aiin"}
	blocks := []string{"blk0", "blk1", "blk2", "blk3"}
	out := make([]SAiinOccurrence, n)
	for i := range out {
		out[i] = SAiinOccurrence{
			X:              xVocab[r.Intn(len(xVocab))],
			Block:          blocks[r.Intn(len(blocks))],
			LineCategory:   lineCategories[r.Intn(len(lineCategories))],
			BlockBinCoarse: blockCoarseCategories[r.Intn(len(blockCoarseCategories))],
		}
	}
	return out
}

func fixtureAiinOccurrences(n int, seed int64) []AiinOccurrence {
	r := rand.New(rand.NewSource(seed))
	xVocab := []string{"chey", "shey", "ol", "or", "dy", "aiin"}
	blocks := []string{"blk0", "blk1", "blk2", "blk3"}
	out := make([]AiinOccurrence, n)
	for i := range out {
		out[i] = AiinOccurrence{
			X:              xVocab[r.Intn(len(xVocab))],
			Block:          blocks[r.Intn(len(blocks))],
			LineCategory:   lineCategories[r.Intn(len(lineCategories))],
			BlockBinCoarse: blockCoarseCategories[r.Intn(len(blockCoarseCategories))],
			HasPredecessor: r.Intn(4) != 0,
			PredecessorIsS: r.Intn(2) == 0,
		}
	}
	return out
}

// TestRunPositionalTestsHoistMatchesReference proves the positionalWorkspace
// rewrite produces byte-identical results to the pre-rewrite reference
// across both position variables, several occurrence-set sizes (including
// the permutations<=0 jackknife path), and several seeds.
func TestRunPositionalTestsHoistMatchesReference(t *testing.T) {
	sizes := []int{0, 5, 60, 300}
	for _, n := range sizes {
		occs := fixtureSAiinOccurrences(n, int64(n)*97+11)
		for _, variable := range []string{"line_position", "block_position_coarse"} {
			cats := lineCategories
			if variable == "block_position_coarse" {
				cats = blockCoarseCategories
			}
			for _, perms := range []int{0, 1, 25} {
				for seed := int64(0); seed < 3; seed++ {
					want := referenceRunPositionalTests(occs, variable, cats, perms, seed)
					got := runPositionalTests(occs, variable, cats, perms, seed)
					if !reflect.DeepEqual(got, want) {
						t.Fatalf("n=%d variable=%s perms=%d seed=%d: diverged\ngot=%+v\nwant=%+v", n, variable, perms, seed, got, want)
					}
				}
			}
		}
	}
}

// TestRunStratifiedPredecessorTestHoistMatchesReference proves the
// stratifiedWorkspace rewrite produces byte-identical results to the
// pre-rewrite reference.
func TestRunStratifiedPredecessorTestHoistMatchesReference(t *testing.T) {
	sizes := []int{0, 5, 60, 300}
	for _, n := range sizes {
		occs := fixtureAiinOccurrences(n, int64(n)*131+17)
		for _, variable := range []string{"line_position", "block_position_coarse"} {
			for _, perms := range []int{0, 1, 25} {
				for seed := int64(0); seed < 3; seed++ {
					want := referenceRunStratifiedPredecessorTest(occs, variable, perms, seed)
					got := runStratifiedPredecessorTest(occs, variable, perms, seed)
					if !reflect.DeepEqual(got, want) {
						t.Fatalf("n=%d variable=%s perms=%d seed=%d: diverged\ngot=%+v\nwant=%+v", n, variable, perms, seed, got, want)
					}
				}
			}
		}
	}
}

func benchPositionalFixture() []SAiinOccurrence {
	return fixtureSAiinOccurrences(2000, 42)
}

func BenchmarkRunPositionalTestsReference(b *testing.B) {
	occs := benchPositionalFixture()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		referenceRunPositionalTests(occs, "line_position", lineCategories, 500, int64(i))
	}
}

func BenchmarkRunPositionalTestsHoisted(b *testing.B) {
	occs := benchPositionalFixture()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		runPositionalTests(occs, "line_position", lineCategories, 500, int64(i))
	}
}

func benchAiinFixture() []AiinOccurrence {
	return fixtureAiinOccurrences(2000, 43)
}

func BenchmarkRunStratifiedPredecessorTestReference(b *testing.B) {
	occs := benchAiinFixture()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		referenceRunStratifiedPredecessorTest(occs, "line_position", 500, int64(i))
	}
}

func BenchmarkRunStratifiedPredecessorTestHoisted(b *testing.B) {
	occs := benchAiinFixture()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		runStratifiedPredecessorTest(occs, "line_position", 500, int64(i))
	}
}
