package positionalcontinuation

import "math"

// pearsonR is the standard Pearson correlation coefficient.
func pearsonR(xs, ys []float64) float64 {
	n := len(xs)
	if n < 2 {
		return 0
	}
	mx, my := mean(xs), mean(ys)
	var sxy, sxx, syy float64
	for i := range xs {
		dx, dy := xs[i]-mx, ys[i]-my
		sxy += dx * dy
		sxx += dx * dx
		syy += dy * dy
	}
	if sxx == 0 || syy == 0 {
		return 0
	}
	return sxy / math.Sqrt(sxx*syy)
}

// buildLineVsBlockRows implements task23 Part M sections 73-77: the
// line-position <-> block-position association, then chey probability
// stratified each way controlling for the other.
func buildLineVsBlockRows(occs []SAiinOccurrence) ([]LineVsBlockRow, float64, string) {
	var normLine, normBlock []float64
	for _, o := range occs {
		normLine = append(normLine, o.NormalizedLinePosition)
		normBlock = append(normBlock, o.NormalizedBlockPosition)
	}
	r := pearsonR(normLine, normBlock)

	var rows []LineVsBlockRow
	// Association cross-tab.
	for _, lc := range lineCategories {
		for _, bc := range blockCoarseCategories {
			var xs []string
			n := 0
			for _, o := range occs {
				if o.LineCategory == lc && o.BlockBinCoarse == bc {
					n++
					if o.X != "" {
						xs = append(xs, o.X)
					}
				}
			}
			if n == 0 {
				continue
			}
			rows = append(rows, LineVsBlockRow{
				Analysis: "association", LineCategory: lc, BlockCoarseBucket: bc,
				OccurrenceCount: n, CheyProbability: cheyProbOf(xs),
			})
		}
	}

	// A: line-position effect controlling coarse block position.
	lineRangesPerBlock := map[string]float64{}
	for _, bc := range blockCoarseCategories {
		var probs []float64
		for _, lc := range lineCategories {
			var xs []string
			n := 0
			for _, o := range occs {
				if o.BlockBinCoarse == bc && o.LineCategory == lc {
					n++
					if o.X != "" {
						xs = append(xs, o.X)
					}
				}
			}
			if n == 0 {
				continue
			}
			p := cheyProbOf(xs)
			probs = append(probs, p)
			rows = append(rows, LineVsBlockRow{
				Analysis: "line_controlling_block", LineCategory: lc, BlockCoarseBucket: bc,
				OccurrenceCount: n, CheyProbability: p,
			})
		}
		if len(probs) >= 2 {
			lo, hi := minMax(probs)
			lineRangesPerBlock[bc] = hi - lo
		}
	}

	// B: block-position effect controlling line position.
	blockRangesPerLine := map[string]float64{}
	for _, lc := range lineCategories {
		var probs []float64
		for _, bc := range blockCoarseCategories {
			var xs []string
			n := 0
			for _, o := range occs {
				if o.LineCategory == lc && o.BlockBinCoarse == bc {
					n++
					if o.X != "" {
						xs = append(xs, o.X)
					}
				}
			}
			if n == 0 {
				continue
			}
			p := cheyProbOf(xs)
			probs = append(probs, p)
			rows = append(rows, LineVsBlockRow{
				Analysis: "block_controlling_line", LineCategory: lc, BlockCoarseBucket: bc,
				OccurrenceCount: n, CheyProbability: p,
			})
		}
		if len(probs) >= 2 {
			lo, hi := minMax(probs)
			blockRangesPerLine[lc] = hi - lo
		}
	}

	lineEffectPersists := len(lineRangesPerBlock) > 0 && maxMapValue(lineRangesPerBlock) > 0
	blockEffectPersists := len(blockRangesPerLine) > 0 && maxMapValue(blockRangesPerLine) > 0

	source := "UNRESOLVED"
	switch {
	case len(lineRangesPerBlock) == 0 && len(blockRangesPerLine) == 0:
		source = "UNRESOLVED"
	case lineEffectPersists && blockEffectPersists:
		source = "BOTH"
	case lineEffectPersists:
		source = "LINE"
	case blockEffectPersists:
		source = "BLOCK"
	}
	return rows, r, source
}

func cheyProbOf(xs []string) float64 {
	if len(xs) == 0 {
		return 0
	}
	n := 0
	for _, x := range xs {
		if x == FrozenChey {
			n++
		}
	}
	return float64(n) / float64(len(xs))
}

func maxMapValue(m map[string]float64) float64 {
	best := 0.0
	first := true
	for _, k := range stringKeysFloat(m) {
		if first || m[k] > best {
			best = m[k]
			first = false
		}
	}
	return best
}
