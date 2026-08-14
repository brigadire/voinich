package higherorderseq

// Primary/descriptive block eligibility thresholds, task22 section 12.
const (
	primaryMinCountAB     = 3
	primaryMinCountB      = 10
	descriptiveMinCountAB = 2
)

// countsInBlock implements task22 Part B sections 9 and Part C section 14:
// the four raw counts a candidate ABC needs within one physical block,
// computed by scanning simple token adjacency (the same continuous-corpus
// convention the frozen sequence analysis used - section 7), never crossing
// the block boundary because it only ever looks within one Block's Tokens.
func countsInBlock(cand Candidate, blk Block) BlockCounts {
	a, b, c := cand.A(), cand.B(), cand.C()
	bc := BlockCounts{Sequence: cand.Sequence, Block: blk.ID, Currier: blk.Currier, Hand: blk.Hand, Joint: blk.Joint}
	n := len(blk.Tokens)
	for k := 0; k < n; k++ {
		if blk.Tokens[k].Text == b {
			bc.CountB++
		}
		if k+1 < n && blk.Tokens[k].Text == a && blk.Tokens[k+1].Text == b {
			bc.CountAB++
		}
		if k+1 < n && blk.Tokens[k].Text == b && blk.Tokens[k+1].Text == c {
			bc.CountBC++
		}
		if k+2 < n && blk.Tokens[k].Text == a && blk.Tokens[k+1].Text == b && blk.Tokens[k+2].Text == c {
			bc.CountABC++
		}
	}
	return bc
}

// conditionalRow turns raw counts into task22 Part B's primary
// conditional-enrichment statistics and Part C's reverse-conditional
// statistics (sections 10, 11, 14, 15). P(A|B) is the fraction of B's
// immediately preceded by A, i.e. count(AB)/count(B); P(A|B,C) is the
// fraction of B,C pairs immediately preceded by A, i.e. count(ABC)/count(BC).
// Both directions reuse the same four counts - task22 section 16 treats them
// as two characteristics of one frozen relation, not separate discoveries.
func conditionalRow(bc BlockCounts) ConditionalRow {
	row := ConditionalRow{BlockCounts: bc}
	if bc.CountB > 0 {
		row.PCGivenB = float64(bc.CountBC) / float64(bc.CountB)
		row.PAGivenB = float64(bc.CountAB) / float64(bc.CountB)
	}
	if bc.CountAB > 0 {
		row.PCGivenAB = float64(bc.CountABC) / float64(bc.CountAB)
	}
	if bc.CountBC > 0 {
		row.PAGivenBC = float64(bc.CountABC) / float64(bc.CountBC)
	}
	if row.PCGivenB > 0 {
		row.Enrichment = row.PCGivenAB / row.PCGivenB
	}
	row.DeltaProbability = row.PCGivenAB - row.PCGivenB
	if row.PAGivenB > 0 {
		row.ReverseEnrichment = row.PAGivenBC / row.PAGivenB
	}
	row.ReverseDeltaProbability = row.PAGivenBC - row.PAGivenB
	row.EligiblePrimary = bc.CountAB >= primaryMinCountAB && bc.CountB >= primaryMinCountB
	row.EligibleDescriptive = bc.CountAB >= descriptiveMinCountAB
	return row
}

// conditionalRowsForCandidate computes one ConditionalRow per physical
// block for a candidate.
func conditionalRowsForCandidate(cand Candidate, blocks []Block) []ConditionalRow {
	rows := make([]ConditionalRow, 0, len(blocks))
	for _, blk := range blocks {
		rows = append(rows, conditionalRow(countsInBlock(cand, blk)))
	}
	return rows
}

func primaryEligible(rows []ConditionalRow) []ConditionalRow {
	var out []ConditionalRow
	for _, r := range rows {
		if r.EligiblePrimary {
			out = append(out, r)
		}
	}
	return out
}

// pooledEnrichment pools counts across a set of blocks before computing
// P(C|B), P(C|A,B) and their ratio - used by cross-block meta-analysis and
// jackknife, which must never treat several occurrences within one block as
// independent replications (task22 section 45).
func pooledEnrichment(rows []ConditionalRow) (pCGivenB, pCGivenAB, enrichment float64) {
	var sumB, sumAB, sumBC, sumABC int
	for _, r := range rows {
		sumB += r.CountB
		sumAB += r.CountAB
		sumBC += r.CountBC
		sumABC += r.CountABC
	}
	if sumB > 0 {
		pCGivenB = float64(sumBC) / float64(sumB)
	}
	if sumAB > 0 {
		pCGivenAB = float64(sumABC) / float64(sumAB)
	}
	if pCGivenB > 0 {
		enrichment = pCGivenAB / pCGivenB
	}
	return
}
