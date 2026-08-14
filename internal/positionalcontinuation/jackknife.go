package positionalcontinuation

import "sort"

func eligibleBlockIDs(sAiinOccs []SAiinOccurrence) []string {
	seen := map[string]bool{}
	for _, o := range sAiinOccs {
		seen[o.Block] = true
	}
	ids := make([]string, 0, len(seen))
	for b := range seen {
		ids = append(ids, b)
	}
	sort.Strings(ids)
	return ids
}

func maxOf(xs []float64) float64 {
	m := 0.0
	for i, v := range xs {
		if i == 0 || v > m {
			m = v
		}
	}
	return m
}

func excludeBlock(sAiinOccs []SAiinOccurrence, id string) []SAiinOccurrence {
	var out []SAiinOccurrence
	for _, o := range sAiinOccs {
		if o.Block != id {
			out = append(out, o)
		}
	}
	return out
}

func excludeBlockAiin(aiinOccs []AiinOccurrence, id string) []AiinOccurrence {
	var out []AiinOccurrence
	for _, o := range aiinOccs {
		if o.Block != id {
			out = append(out, o)
		}
	}
	return out
}

// jackknifeSensitivityThreshold: a realization is SINGLE_BLOCK_SENSITIVE if
// removing it flips the sign of the point estimate relative to the full
// estimate (task23 section 71).
func signFlips(full, sub float64) bool { return full*sub < 0 }

// runJackknife implements task23 Part L sections 69-72: leave-one-eligible-
// physical-block-out recomputation of the point estimates from Parts E-I
// (never a fresh 10000-permutation p-value per realization - see
// permutation.go's permutations<=0 fast path), for both the line-position
// and coarse block-position variables.
func runJackknife(sAiinOccs []SAiinOccurrence, aiinOccs []AiinOccurrence, seed int64) []PositionalJackknifeRow {
	ids := eligibleBlockIDs(sAiinOccs)
	variables := []struct {
		name       string
		categories []string
	}{
		{"line_position", lineCategories},
		{"block_position_coarse", blockCoarseCategories},
	}

	fullByVar := map[string]positionalTestResult{}
	fullStratByVar := map[string]StratifiedPredecessorRow{}
	for _, v := range variables {
		fullByVar[v.name] = runPositionalTests(sAiinOccs, v.name, v.categories, 0, seed)
		fullStratByVar[v.name] = runStratifiedPredecessorTest(aiinOccs, v.name, 0, seed)
	}

	var rows []PositionalJackknifeRow
	for _, v := range variables {
		full := fullByVar[v.name]
		fullStrat := fullStratByVar[v.name]
		fullEntropyEffect := maxEntropyDiff(full.Entropy)
		fullCheyEnrichment := maxEnrichment(full.CheyEffect)

		var mis, entropyEffects, cheyEnrichments, stratS []float64
		sensitive := false
		for _, id := range ids {
			subSAiin := excludeBlock(sAiinOccs, id)
			subAiin := excludeBlockAiin(aiinOccs, id)
			sub := runPositionalTests(subSAiin, v.name, v.categories, 0, seed)
			subStrat := runStratifiedPredecessorTest(subAiin, v.name, 0, seed)

			mis = append(mis, sub.Dependence.ObservedMIBits)
			ee := maxEntropyDiff(sub.Entropy)
			entropyEffects = append(entropyEffects, ee)
			ce := maxEnrichment(sub.CheyEffect)
			cheyEnrichments = append(cheyEnrichments, ce)
			stratS = append(stratS, subStrat.ObservedStatistic)

			if signFlips(fullEntropyEffect, ee) || signFlips(fullCheyEnrichment-1, ce-1) || signFlips(fullStrat.ObservedStatistic, subStrat.ObservedStatistic) {
				sensitive = true
			}
		}

		row := PositionalJackknifeRow{PositionVariable: v.name, Realizations: len(ids), SingleBlockSensitive: sensitive}
		if len(ids) > 0 {
			row.MIMin, row.MIMax = minMax(mis)
			row.MIMedian = median(mis)
			_, row.MISD = meanSD(mis)
			row.EntropyEffectMin, row.EntropyEffectMax = minMax(entropyEffects)
			row.EntropyEffectMedian = median(entropyEffects)
			_, row.EntropyEffectSD = meanSD(entropyEffects)
			row.CheyEnrichmentMin, row.CheyEnrichmentMax = minMax(cheyEnrichments)
			row.CheyEnrichmentMedian = median(cheyEnrichments)
			_, row.CheyEnrichmentSD = meanSD(cheyEnrichments)
			row.StratifiedSMin, row.StratifiedSMax = minMax(stratS)
			row.StratifiedSMedian = median(stratS)
			_, row.StratifiedSSD = meanSD(stratS)
		}
		rows = append(rows, row)
	}
	return rows
}

func maxEntropyDiff(rows []PositionalEntropyRow) float64 {
	xs := make([]float64, len(rows))
	for i, r := range rows {
		xs[i] = r.EntropyDifference
	}
	return maxOf(xs)
}

func maxEnrichment(rows []CheyEffectRow) float64 {
	best := 0.0
	for i, r := range rows {
		if i == 0 || r.PositionalEnrichment > best {
			best = r.PositionalEnrichment
		}
	}
	return best
}
