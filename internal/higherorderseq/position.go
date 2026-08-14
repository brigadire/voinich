package higherorderseq

// abOccurrencePositions collects the normalized block position and, when the
// bigram sits entirely within one line, the line-position bucket for every
// in-block occurrence of the candidate's A immediately followed by B. It is
// the baseline task22 section 57 compares ABC's own position distribution
// against: if AB occurs everywhere in a block but ABC only appears near a
// boundary, that is a positional-rule candidate rather than a general
// second-order transition.
func abOccurrencePositions(cand Candidate, blocks []Block, lineLength map[string]int) (normPos []float64, linePos []string) {
	a, b := cand.A(), cand.B()
	for _, blk := range blocks {
		n := len(blk.Tokens)
		for k := 0; k+1 < n; k++ {
			if blk.Tokens[k].Text != a || blk.Tokens[k+1].Text != b {
				continue
			}
			normPos = append(normPos, normalizedPosition(k, n))
			if blk.Tokens[k].Line == blk.Tokens[k+1].Line {
				linePos = append(linePos, linePosition(blk.Tokens[k].TokenIndexLine, lineLength[blk.Tokens[k].Line]))
			}
		}
	}
	return
}

// positionRows implements task22 Part K sections 55-58: the normalized
// block-position bucket distribution and, where line metadata is available,
// the line-start/middle/end distribution, for ABC occurrences against the
// AB baseline.
func positionRows(cand Candidate, occs []Occurrence, blocks []Block, lineLength map[string]int) []PositionRow {
	abcBuckets := map[string]int{}
	abcLine := map[string]int{}
	for _, o := range occs {
		abcBuckets[positionBucket(o.NormalizedBlockPos)]++
		if o.WithinSameLine {
			abcLine[o.LinePosition]++
		}
	}
	abNormPos, abLinePos := abOccurrencePositions(cand, blocks, lineLength)
	abBuckets := map[string]int{}
	for _, p := range abNormPos {
		abBuckets[positionBucket(p)]++
	}
	abLine := map[string]int{}
	for _, lp := range abLinePos {
		abLine[lp]++
	}

	totalABC, totalAB := len(occs), len(abNormPos)
	var rows []PositionRow
	for _, bucket := range []string{"[0,0.1)", "[0.1,0.2)", "[0.2,0.3)", "[0.3,0.4)", "[0.4,0.5)", "[0.5,0.6)", "[0.6,0.7)", "[0.7,0.8)", "[0.8,0.9)", "[0.9,1.0]"} {
		row := PositionRow{Sequence: cand.Sequence, Metric: "block_position_bin", Bucket: bucket, ABCCount: abcBuckets[bucket], ABCount: abBuckets[bucket]}
		if totalABC > 0 {
			row.ABCFraction = float64(row.ABCCount) / float64(totalABC)
		}
		if totalAB > 0 {
			row.ABFraction = float64(row.ABCount) / float64(totalAB)
		}
		rows = append(rows, row)
	}
	totalABCLine, totalABLine := len(occs)-countCross(occs), len(abLinePos)
	for _, bucket := range []string{"start", "middle", "end"} {
		row := PositionRow{Sequence: cand.Sequence, Metric: "line_position", Bucket: bucket, ABCCount: abcLine[bucket], ABCount: abLine[bucket]}
		if totalABCLine > 0 {
			row.ABCFraction = float64(row.ABCCount) / float64(totalABCLine)
		}
		if totalABLine > 0 {
			row.ABFraction = float64(row.ABCount) / float64(totalABLine)
		}
		rows = append(rows, row)
	}
	return rows
}

func countCross(occs []Occurrence) int {
	n := 0
	for _, o := range occs {
		if o.CrossesLineBoundary {
			n++
		}
	}
	return n
}

// positionTVD computes total variation distance between the ABC and AB
// fractions for one Metric group of positionRows - used descriptively by
// the POSITION_DEPENDENT diagnostic, not as a new significance test.
func positionTVD(rows []PositionRow, metric string) float64 {
	sum := 0.0
	for _, r := range rows {
		if r.Metric != metric {
			continue
		}
		d := r.ABCFraction - r.ABFraction
		if d < 0 {
			d = -d
		}
		sum += d
	}
	return sum / 2
}
