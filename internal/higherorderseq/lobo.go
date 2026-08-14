package higherorderseq

const smoothingAlpha = 0.5 // task22 section 26: fixed, never optimized

// abContinuations returns, for one block, the actual continuation token
// following every in-block occurrence of the candidate's A immediately
// followed by B (i.e. every AB occurrence that itself has a following
// token). This is the held-out test set for one LOBO fold.
func abContinuations(cand Candidate, blk Block) []string {
	a, b := cand.A(), cand.B()
	var out []string
	n := len(blk.Tokens)
	for k := 0; k+2 < n; k++ {
		if blk.Tokens[k].Text == a && blk.Tokens[k+1].Text == b {
			out = append(out, blk.Tokens[k+2].Text)
		}
	}
	return out
}

// trainModels builds M1 (P(x|B)) and M2 (P(x|A,B)) training statistics from
// a set of training blocks: right-neighbor counts of B for M1, and
// right-neighbor counts of the specific AB bigram for M2, over a shared
// vocabulary (the set of tokens ever observed to follow B in training).
func trainModels(cand Candidate, trainBlocks []Block) (countB map[string]int, totalB int, countAB map[string]int, totalAB int, vocab map[string]bool) {
	a, b := cand.A(), cand.B()
	countB = map[string]int{}
	countAB = map[string]int{}
	for _, blk := range trainBlocks {
		n := len(blk.Tokens)
		for k := 0; k+1 < n; k++ {
			if blk.Tokens[k].Text == b {
				x := blk.Tokens[k+1].Text
				countB[x]++
				totalB++
			}
		}
		for k := 0; k+2 < n; k++ {
			if blk.Tokens[k].Text == a && blk.Tokens[k+1].Text == b {
				x := blk.Tokens[k+2].Text
				countAB[x]++
				totalAB++
			}
		}
	}
	vocab = map[string]bool{}
	for x := range countB {
		vocab[x] = true
	}
	return
}

// loboBlockDeltas runs leave-one-physical-block-out evaluation (task22
// sections 23-27) over exactly the given blocks: for each held-out block
// with at least one AB occurrence, train M1/M2 on every other block in the
// slice only (zero leakage), then score the held-out block's actual AB
// continuations under both models with fixed alpha=0.5 additive smoothing.
// It returns one mean delta_log_loss per tested block plus the grand sum
// (the held-out log-likelihood ratio in bits) across every individual
// observation.
func loboBlockDeltas(cand Candidate, blocks []Block) (blockMeans []float64, sumAll float64) {
	for i, held := range blocks {
		test := abContinuations(cand, held)
		if len(test) == 0 {
			continue
		}
		var train []Block
		for j, b := range blocks {
			if j != i {
				train = append(train, b)
			}
		}
		countB, totalB, countAB, totalAB, vocab := trainModels(cand, train)
		var deltas []float64
		for _, x := range test {
			p1 := smoothedProb(countB, vocab, x, totalB, smoothingAlpha)
			p2 := smoothedProb(countAB, vocab, x, totalAB, smoothingAlpha)
			delta := log2Loss(p1) - log2Loss(p2)
			deltas = append(deltas, delta)
			sumAll += delta
		}
		m, _ := meanSD(deltas)
		blockMeans = append(blockMeans, m)
	}
	return
}

// runLOBO implements task22 Part E in full for one candidate.
func runLOBO(cand Candidate, blocks []Block) LOBORow {
	blockMeans, sumAll := loboBlockDeltas(cand, blocks)
	row := LOBORow{Sequence: cand.Sequence, TestedBlocks: len(blockMeans), HeldoutLogLikelihoodRatio: sumAll}
	for _, m := range blockMeans {
		switch {
		case m > 1e-12:
			row.M2BetterBlocks++
		case m < -1e-12:
			row.M1BetterBlocks++
		default:
			row.Ties++
		}
	}
	row.MeanDeltaLogLoss, _ = meanSD(blockMeans)
	row.MedianDeltaLogLoss = median(blockMeans)
	return row
}
