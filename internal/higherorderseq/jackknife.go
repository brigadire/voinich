package higherorderseq

// blocksExcluding returns every block except the one with the given ID.
func blocksExcluding(blocks []Block, excludeID string) []Block {
	out := make([]Block, 0, len(blocks))
	for _, b := range blocks {
		if b.ID != excludeID {
			out = append(out, b)
		}
	}
	return out
}

// jackknifeRow implements task22 Part J sections 51-54: leave-one-eligible-
// physical-block-out recomputation of pooled conditional enrichment, CMI
// around B, and the M1-vs-M2 predictive advantage, exactly one block
// removed at a time (never several at once, so each realization isolates
// the influence of a single physical block).
func jackknifeRow(cand Candidate, blocks []Block, eligibleRows []ConditionalRow) JackknifeRow {
	_, _, fullEnrichment := pooledEnrichment(eligibleRows)
	fullMeans, _ := loboBlockDeltas(cand, blocks)
	fullMeanDelta, _ := meanSD(fullMeans)

	row := JackknifeRow{Sequence: cand.Sequence, Realizations: len(eligibleRows)}
	var enrichments, cmis, deltas []float64
	signFlips := false
	for _, held := range eligibleRows {
		remaining := make([]ConditionalRow, 0, len(eligibleRows)-1)
		for _, r := range eligibleRows {
			if r.Block != held.Block {
				remaining = append(remaining, r)
			}
		}
		filtered := blocksExcluding(blocks, held.Block)

		_, _, enrichment := pooledEnrichment(remaining)
		enrichments = append(enrichments, enrichment)
		if (enrichment-1)*(fullEnrichment-1) < 0 {
			signFlips = true
		}

		cmi := cmiBits(collectBNeighbors(cand.B(), filtered))
		cmis = append(cmis, cmi)

		means, _ := loboBlockDeltas(cand, filtered)
		meanDelta, _ := meanSD(means)
		deltas = append(deltas, meanDelta)
		if meanDelta*fullMeanDelta < 0 {
			signFlips = true
		}
	}
	if len(enrichments) > 0 {
		row.EnrichmentMin, row.EnrichmentMax = minMax(enrichments)
		row.EnrichmentMedian = median(enrichments)
		_, row.EnrichmentSD = meanSD(enrichments)
		row.CMIMin, row.CMIMax = minMax(cmis)
		row.CMIMedian = median(cmis)
		_, row.CMISD = meanSD(cmis)
		row.DeltaLogLossMin, row.DeltaLogLossMax = minMax(deltas)
		row.DeltaLogLossMedian = median(deltas)
		_, row.DeltaLogLossSD = meanSD(deltas)
	}
	row.SingleBlockSensitive = signFlips
	return row
}
