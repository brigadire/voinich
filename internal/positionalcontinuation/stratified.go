package positionalcontinuation

import (
	"math/rand"
)

// stratifiedObs is one aiin occurrence with X present, reduced to what the
// Part I permutation test needs: its (block, position) stratum, whether its
// predecessor is "s", and whether its continuation is "chey".
type stratifiedObs struct {
	stratum string
	isS     bool
	isChey  bool
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
	ws := newStratifiedWorkspace(obs)
	observed := ws.observedStatistic()

	r := rand.New(rand.NewSource(seed))
	null := make([]float64, permutations)
	for p := 0; p < permutations; p++ {
		null[p] = ws.permuteAndStatistic(r)
	}
	mean, sd := meanSD(null)
	return StratifiedPredecessorRow{
		PositionVariable: variable, ObservedStatistic: observed,
		NullMeanStatistic: mean, NullSDStatistic: sd, Permutations: permutations,
		EmpiricalP: empiricalP(observed, null),
	}
}
