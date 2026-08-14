package positionalcontinuation

import "math"

var lineCategories = []string{"LINE_START", "LINE_EARLY", "LINE_MIDDLE", "LINE_LATE", "LINE_END"}
var blockCoarseCategories = []string{"BLOCK_START", "BLOCK_MIDDLE", "BLOCK_END"}
var blockFixedCategories = []string{"B0", "B1", "B2", "B3", "B4", "B5", "B6", "B7", "B8", "B9"}

// presentX returns the X tokens (skipping boundary occurrences with no
// continuation) for a set of s-aiin occurrences.
func presentX(occs []SAiinOccurrence) []string {
	var xs []string
	for _, o := range occs {
		if o.X != "" {
			xs = append(xs, o.X)
		}
	}
	return xs
}

func filterByLineCategory(occs []SAiinOccurrence, cat string) []SAiinOccurrence {
	var out []SAiinOccurrence
	for _, o := range occs {
		if o.LineCategory == cat {
			out = append(out, o)
		}
	}
	return out
}

func filterByBlockCoarse(occs []SAiinOccurrence, cat string) []SAiinOccurrence {
	var out []SAiinOccurrence
	for _, o := range occs {
		if o.BlockBinCoarse == cat {
			out = append(out, o)
		}
	}
	return out
}

func filterByBlockFixed(occs []SAiinOccurrence, cat string) []SAiinOccurrence {
	var out []SAiinOccurrence
	for _, o := range occs {
		if o.BlockBinFixed == cat {
			out = append(out, o)
		}
	}
	return out
}

func countMap(xs []string) map[string]int {
	m := map[string]int{}
	for _, x := range xs {
		m[x]++
	}
	return m
}

// distributionSummary implements task23 Part D section 20 for one stratum's
// continuation sample.
func distributionSummary(context, stratum, stratumType string, xs []string) DistributionSummaryRow {
	counts := countMap(xs)
	probs := toProbabilities(counts)
	h := entropyBits(probs)
	row := DistributionSummaryRow{
		Context: context, Stratum: stratum, StratumType: stratumType,
		OccurrenceCount: len(xs), UniqueContinuations: len(counts), EntropyBits: h,
	}
	if len(counts) > 1 {
		row.NormalizedEntropy = h / math.Log2(float64(len(counts)))
	}
	top, topN := "", -1
	for _, k := range stringKeysInt(counts) {
		if counts[k] > topN {
			top, topN = k, counts[k]
		}
	}
	row.TopContinuation = top
	if len(xs) > 0 {
		row.TopContinuationProb = float64(topN) / float64(len(xs))
		row.CheyProbability = float64(counts[FrozenChey]) / float64(len(xs))
	}
	return row
}

func continuationRowsForStratum(context, stratum, stratumType string, xs []string) []ContinuationRow {
	counts := countMap(xs)
	var rows []ContinuationRow
	for _, k := range stringKeysInt(counts) {
		rows = append(rows, ContinuationRow{
			Context: context, Stratum: stratum, StratumType: stratumType, Token: k,
			Count: counts[k], Probability: float64(counts[k]) / float64(len(xs)),
		})
	}
	return rows
}

// buildSAiinContinuationDistributions implements task23 Part D in full
// (sections 16-20) for the "s aiin" context.
func buildSAiinContinuationDistributions(occs []SAiinOccurrence) ([]ContinuationRow, []DistributionSummaryRow) {
	var rows []ContinuationRow
	var summaries []DistributionSummaryRow

	global := presentX(occs)
	rows = append(rows, continuationRowsForStratum("s_aiin", "global", "", global)...)
	summaries = append(summaries, distributionSummary("s_aiin", "global", "", global))

	for _, cat := range lineCategories {
		xs := presentX(filterByLineCategory(occs, cat))
		rows = append(rows, continuationRowsForStratum("s_aiin", cat, "line_position", xs)...)
		summaries = append(summaries, distributionSummary("s_aiin", cat, "line_position", xs))
	}
	for _, cat := range blockCoarseCategories {
		xs := presentX(filterByBlockCoarse(occs, cat))
		rows = append(rows, continuationRowsForStratum("s_aiin", cat, "block_position_coarse", xs)...)
		summaries = append(summaries, distributionSummary("s_aiin", cat, "block_position_coarse", xs))
	}
	for _, cat := range blockFixedCategories {
		xs := presentX(filterByBlockFixed(occs, cat))
		rows = append(rows, continuationRowsForStratum("s_aiin", cat, "block_position_fixed", xs)...)
		summaries = append(summaries, distributionSummary("s_aiin", cat, "block_position_fixed", xs))
	}
	return rows, summaries
}
