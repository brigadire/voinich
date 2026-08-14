package positionalcontinuation

import "sort"

// trainAiinModels builds M1 (P(X|aiin)), M2 (P(X|aiin,line_position)) and M3
// (P(X|predecessor in {s,other},aiin,line_position)) training statistics
// from a set of training aiin occurrences, over a shared vocabulary (every
// token ever observed to follow aiin in training) - task23 Part J sections
// 54-55, 59.
func trainAiinModels(train []AiinOccurrence) (
	countM1 map[string]int, totalM1 int,
	countM2 map[string]map[string]int, totalM2 map[string]int,
	countM3 map[string]map[string]int, totalM3 map[string]int,
	vocab map[string]bool,
) {
	countM1 = map[string]int{}
	countM2 = map[string]map[string]int{}
	totalM2 = map[string]int{}
	countM3 = map[string]map[string]int{}
	totalM3 = map[string]int{}
	vocab = map[string]bool{}
	for _, o := range train {
		if o.X == "" {
			continue
		}
		countM1[o.X]++
		totalM1++
		vocab[o.X] = true

		if countM2[o.LineCategory] == nil {
			countM2[o.LineCategory] = map[string]int{}
		}
		countM2[o.LineCategory][o.X]++
		totalM2[o.LineCategory]++

		predKey := "other"
		if o.PredecessorIsS {
			predKey = "s"
		}
		key := predKey + "|" + o.LineCategory
		if countM3[key] == nil {
			countM3[key] = map[string]int{}
		}
		countM3[key][o.X]++
		totalM3[key]++
	}
	return
}

// runModelLOBO implements task23 Part J in full: M1/M2/M3 leave-one-
// physical-block-out comparison (never a random split - section 56), fixed
// alpha=0.5 additive smoothing.
func runModelLOBO(aiinOccs []AiinOccurrence) ModelLOBORow {
	byBlock := map[string][]AiinOccurrence{}
	for _, o := range aiinOccs {
		byBlock[o.Block] = append(byBlock[o.Block], o)
	}
	blockIDs := make([]string, 0, len(byBlock))
	for b := range byBlock {
		blockIDs = append(blockIDs, b)
	}
	sort.Strings(blockIDs)

	row := ModelLOBORow{}
	var deltas21, deltas32 []float64
	for _, held := range blockIDs {
		test := byBlock[held]
		var testWithX []AiinOccurrence
		for _, o := range test {
			if o.X != "" {
				testWithX = append(testWithX, o)
			}
		}
		if len(testWithX) == 0 {
			continue
		}
		var train []AiinOccurrence
		for _, b := range blockIDs {
			if b != held {
				train = append(train, byBlock[b]...)
			}
		}
		countM1, totalM1, countM2, totalM2, countM3, totalM3, vocab := trainAiinModels(train)

		var d21s, d32s []float64
		for _, o := range testWithX {
			p1 := smoothedProb(countM1, vocab, o.X, totalM1, smoothingAlpha)
			p2 := smoothedProb(countM2[o.LineCategory], vocab, o.X, totalM2[o.LineCategory], smoothingAlpha)
			predKey := "other"
			if o.PredecessorIsS {
				predKey = "s"
			}
			key := predKey + "|" + o.LineCategory
			p3 := smoothedProb(countM3[key], vocab, o.X, totalM3[key], smoothingAlpha)

			l1, l2, l3 := log2Loss(p1), log2Loss(p2), log2Loss(p3)
			d21s = append(d21s, l1-l2)
			d32s = append(d32s, l2-l3)
		}
		m21, _ := meanSD(d21s)
		m32, _ := meanSD(d32s)
		deltas21 = append(deltas21, m21)
		deltas32 = append(deltas32, m32)
		row.TestedBlocks++
		switch {
		case m21 > 1e-12:
			row.BlocksM2BetterM1++
		case m21 < -1e-12:
			row.BlocksM1BetterM2++
		}
		switch {
		case m32 > 1e-12:
			row.BlocksM3BetterM2++
		case m32 < -1e-12:
			row.BlocksM2BetterM3++
		}
	}
	row.MeanDelta21, _ = meanSD(deltas21)
	row.MedianDelta21 = median(deltas21)
	row.MeanDelta32, _ = meanSD(deltas32)
	row.MedianDelta32 = median(deltas32)
	return row
}
