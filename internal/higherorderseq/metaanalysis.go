package higherorderseq

import "math"

// blockLogEnrichment computes a Haldane-Anscombe smoothed log-enrichment
// for one block (task22 section 46: "log(enrichment) with appropriate
// smoothing", needed because raw enrichment is undefined whenever count_BC
// or count_ABC is zero) together with a sample-size weight for the
// meta-analysis.
func blockLogEnrichment(r ConditionalRow) (logE, weight float64) {
	p1 := (float64(r.CountABC) + 0.5) / (float64(r.CountAB) + 1)
	p0 := (float64(r.CountBC) + 0.5) / (float64(r.CountB) + 1)
	return math.Log(p1 / p0), float64(r.CountAB)
}

// metaAnalysisRow implements task22 Part I sections 46-50: unweighted and
// sample-size-weighted mean log-enrichment across eligible blocks, between-
// block variance, Cochran's Q and I^2 as descriptive heterogeneity
// statistics (section 49 explicitly asks for these as descriptive, not as a
// new significance test), and how much of the total evidence one block
// contributes (section 50).
func metaAnalysisRow(rows []ConditionalRow) MetaAnalysisRow {
	eligible := primaryEligible(rows)
	seq := ""
	if len(rows) > 0 {
		seq = rows[0].Sequence
	}
	row := MetaAnalysisRow{Sequence: seq, Blocks: len(eligible)}
	if len(eligible) == 0 {
		return row
	}
	logEs := make([]float64, len(eligible))
	weights := make([]float64, len(eligible))
	var sumW, sumWLogE float64
	for i, r := range eligible {
		logEs[i], weights[i] = blockLogEnrichment(r)
		sumW += weights[i]
		sumWLogE += weights[i] * logEs[i]
	}
	row.UnweightedMeanLogEnrichment, _ = meanSD(logEs)
	row.MedianLogEnrichment = median(logEs)
	if sumW > 0 {
		row.WeightedMeanLogEnrichment = sumWLogE / sumW
		maxW := weights[0]
		for _, w := range weights {
			if w > maxW {
				maxW = w
			}
		}
		row.MaxBlockWeightFraction = maxW / sumW
	}
	_, sd := meanSD(logEs)
	row.BetweenBlockVariance = sd * sd
	q := 0.0
	for i := range eligible {
		d := logEs[i] - row.WeightedMeanLogEnrichment
		q += weights[i] * d * d
	}
	row.CochranQ = q
	df := float64(len(eligible) - 1)
	if q > 0 && df > 0 {
		row.I2 = math.Max(0, (q-df)/q) * 100
	}
	return row
}
