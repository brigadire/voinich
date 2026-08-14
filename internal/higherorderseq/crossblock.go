package higherorderseq

// crossBlockRow implements task22 Part H sections 42-45: cross-block
// replication counted at the level of physical blocks (never individual
// occurrences within one block), plus how many distinct Currier, hand and
// joint metadata classes the eligible evidence spans.
func crossBlockRow(rows []ConditionalRow) CrossBlockRow {
	eligible := primaryEligible(rows)
	seq := ""
	if len(rows) > 0 {
		seq = rows[0].Sequence
	}
	row := CrossBlockRow{Sequence: seq, EligibleBlocks: len(eligible)}
	currier, hand, joint := map[string]bool{}, map[string]bool{}, map[string]bool{}
	for _, r := range eligible {
		currier[r.Currier] = true
		hand[r.Hand] = true
		joint[r.Joint] = true
		switch {
		case r.Enrichment > 1:
			row.PositiveEnrichmentBlocks++
		case r.Enrichment < 1:
			row.NegativeEnrichmentBlocks++
		}
	}
	if row.EligibleBlocks > 0 {
		row.SignConsistency = float64(row.PositiveEnrichmentBlocks) / float64(row.EligibleBlocks)
	}
	row.DistinctCurrier, row.DistinctHand, row.DistinctJoint = len(currier), len(hand), len(joint)
	row.CrossCurrier = row.DistinctCurrier >= 2
	row.CrossHand = row.DistinctHand >= 2
	row.CrossJoint = row.DistinctJoint >= 2
	return row
}
