package positionalcontinuation

// buildReversePositionRows implements task23 Part P sections 88-91: is
// "s aiin" itself positionally specialized relative to "aiin" in general?
// Compares P(position|s,aiin) against P(position|aiin) for both the line and
// coarse block positional variables.
func buildReversePositionRows(sAiinOccs []SAiinOccurrence, aiinOccs []AiinOccurrence) []ReversePositionRow {
	var rows []ReversePositionRow
	sLine := map[string]int{}
	for _, o := range sAiinOccs {
		sLine[o.LineCategory]++
	}
	aLine := map[string]int{}
	for _, o := range aiinOccs {
		aLine[o.LineCategory]++
	}
	rows = append(rows, reverseRowsFor("line_position", lineCategories, sLine, len(sAiinOccs), aLine, len(aiinOccs))...)

	sBlock := map[string]int{}
	for _, o := range sAiinOccs {
		sBlock[o.BlockBinCoarse]++
	}
	aBlock := map[string]int{}
	for _, o := range aiinOccs {
		aBlock[o.BlockBinCoarse]++
	}
	rows = append(rows, reverseRowsFor("block_position_coarse", blockCoarseCategories, sBlock, len(sAiinOccs), aBlock, len(aiinOccs))...)
	return rows
}

func reverseRowsFor(variable string, categories []string, sCounts map[string]int, sTotal int, aCounts map[string]int, aTotal int) []ReversePositionRow {
	var rows []ReversePositionRow
	tvd := 0.0
	for _, cat := range categories {
		ps, pa := 0.0, 0.0
		if sTotal > 0 {
			ps = float64(sCounts[cat]) / float64(sTotal)
		}
		if aTotal > 0 {
			pa = float64(aCounts[cat]) / float64(aTotal)
		}
		d := ps - pa
		if d < 0 {
			d = -d
		}
		tvd += d
		rows = append(rows, ReversePositionRow{PositionVariable: variable, Stratum: cat, PGivenSAiin: ps, PGivenAiin: pa, TotalVariation: tvd / 2})
	}
	// Back-fill the final (whole-distribution) TVD onto every row for this
	// variable, so each row also carries the overall summary statistic.
	final := tvd / 2
	for i := range rows {
		if rows[i].PositionVariable == variable {
			rows[i].TotalVariation = final
		}
	}
	return rows
}
