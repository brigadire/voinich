package positionalcontinuation

import (
	"math/rand"
	"sort"
)

// stratifiedObs is one aiin occurrence with X present, reduced to what the
// Part I permutation test needs: its (block, position) stratum, whether its
// predecessor is "s", and whether its continuation is "chey".
type stratifiedObs struct {
	stratum    string
	isS        bool
	isChey     bool
}

// permuteIsSWithinStrata shuffles the predecessor-is-s label among
// occurrences sharing the same (physical block, position category) stratum
// (task23 sections 48-53): it preserves aiin, position, block membership and
// continuation-token identity per occurrence, while destroying the
// predecessor<->continuation association. Nothing is ever permuted across
// strata.
func permuteIsSWithinStrata(obs []stratifiedObs, r *rand.Rand) []bool {
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

func statisticCheyAmongS(obs []stratifiedObs, isS []bool) float64 {
	n := 0.0
	for i, o := range obs {
		if isS[i] && o.isChey {
			n++
		}
	}
	return n
}

// runStratifiedPredecessorTest implements task23 Part I: does the
// predecessor "s" still predict "chey" once aiin, position and physical
// block are all held fixed? The test statistic is the pooled count of chey
// continuations among predecessor==s occurrences, stratified by (block,
// position) so the permutation null never mixes evidence across blocks or
// positional categories (section 53).
func runStratifiedPredecessorTest(aiinOccs []AiinOccurrence, variable string, permutations int, seed int64) StratifiedPredecessorRow {
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
	observed := statisticCheyAmongS(obs, baseline)

	r := rand.New(rand.NewSource(seed))
	null := make([]float64, permutations)
	for p := 0; p < permutations; p++ {
		perm := permuteIsSWithinStrata(obs, r)
		null[p] = statisticCheyAmongS(obs, perm)
	}
	mean, sd := meanSD(null)
	return StratifiedPredecessorRow{
		PositionVariable: variable, ObservedStatistic: observed,
		NullMeanStatistic: mean, NullSDStatistic: sd, Permutations: permutations,
		EmpiricalP: empiricalP(observed, null),
	}
}
