package globalregime

import (
	"math"
	"sort"
)

type seriesCost struct{ sum, squares []float64 }

func newSeriesCost(x []float64) seriesCost {
	c := seriesCost{make([]float64, len(x)+1), make([]float64, len(x)+1)}
	for i, v := range x {
		c.sum[i+1] = c.sum[i] + v
		c.squares[i+1] = c.squares[i] + v*v
	}
	return c
}
func (c seriesCost) cost(a, b int) float64 {
	if b <= a {
		return 0
	}
	n := float64(b - a)
	s := c.sum[b] - c.sum[a]
	q := c.squares[b] - c.squares[a]
	v := q - s*s/n
	if v < 0 {
		return 0
	}
	return v
}

func jumpSeries(w []Window) []float64 {
	x := make([]float64, len(w))
	for i := 1; i < len(w); i++ {
		x[i] = w[i].JSDistance
	}
	return x
}

func thresholdPeaks(w []Window) []ChangePoint {
	if len(w) < 3 {
		return nil
	}
	x := jumpSeries(w)[1:]
	threshold := mean(x) + stddev(x)
	var out []ChangePoint
	for i := 1; i < len(w)-1; i++ {
		if w[i].LocalPeak && w[i].JSDistance >= threshold {
			out = append(out, ChangePoint{w[i].WindowSize, w[i].Center, "threshold", w[i].JSDistance, threshold})
		}
	}
	return out
}

// pelt applies penalised optimal partitioning to the distributional JS-jump
// series. The observations are distances between complete token distributions,
// never frequencies of a selected token.
func pelt(w []Window) []ChangePoint {
	if len(w) < 8 {
		return nil
	}
	x := jumpSeries(w)
	c := newSeriesCost(x)
	n := len(x)
	minSeg := max(3, w[0].WindowSize/max(1, w[0].Step)/2)
	variance := c.cost(1, n) / float64(max(1, n-1))
	penalty := maxFloat(1e-12, 2*variance*math.Log(float64(n)))
	dp := make([]float64, n+1)
	prev := make([]int, n+1)
	for i := range dp {
		dp[i] = math.Inf(1)
		prev[i] = -1
	}
	dp[0] = -penalty
	// PELT candidate pruning. Recent candidates are retained until they can
	// form a minimum-length segment; older candidates survive only while the
	// standard additive-cost pruning inequality remains possible.
	candidates := []int{0}
	for end := 1; end <= n; end++ {
		for _, start := range candidates {
			if end-start < minSeg || math.IsInf(dp[start], 1) {
				continue
			}
			v := dp[start] + c.cost(start, end) + penalty
			if v < dp[end] {
				dp[end], prev[end] = v, start
			}
		}
		next := candidates[:0]
		for _, start := range candidates {
			if end-start < minSeg || math.IsInf(dp[end], 1) || dp[start]+c.cost(start, end) <= dp[end] {
				next = append(next, start)
			}
		}
		candidates = next
		if !math.IsInf(dp[end], 1) {
			candidates = append(candidates, end)
		}
	}
	if prev[n] < 0 {
		return nil
	}
	var cuts []int
	for end := n; prev[end] > 0; end = prev[end] {
		cuts = append(cuts, prev[end])
	}
	sort.Ints(cuts)
	out := make([]ChangePoint, 0, len(cuts))
	for _, i := range cuts {
		if i > 0 && i < len(w) {
			out = append(out, ChangePoint{w[i].WindowSize, w[i].Center, "pelt", w[i].JSDistance, penalty})
		}
	}
	return out
}

type segment struct{ a, b int }

func binarySegments(w []Window, k int) []int {
	n := len(w)
	labels := make([]int, n)
	if n == 0 || k <= 1 {
		return labels
	}
	x := jumpSeries(w)
	c := newSeriesCost(x)
	minSeg := max(2, w[0].WindowSize/max(1, w[0].Step)/2)
	segs := []segment{{0, n}}
	for len(segs) < k {
		bestGain := math.Inf(-1)
		bestSI, bestCut := -1, -1
		for si, s := range segs {
			if s.b-s.a < 2*minSeg {
				continue
			}
			base := c.cost(s.a, s.b)
			for cut := s.a + minSeg; cut <= s.b-minSeg; cut++ {
				gain := base - c.cost(s.a, cut) - c.cost(cut, s.b)
				if gain > bestGain {
					bestGain, bestSI, bestCut = gain, si, cut
				}
			}
		}
		if bestSI < 0 {
			break
		}
		s := segs[bestSI]
		segs = append(segs, segment{})
		copy(segs[bestSI+2:], segs[bestSI+1:])
		segs[bestSI] = segment{s.a, bestCut}
		segs[bestSI+1] = segment{bestCut, s.b}
	}
	for label, s := range segs {
		for i := s.a; i < s.b; i++ {
			labels[i] = label
		}
	}
	return labels
}
func binaryChangePoints(w []Window) []ChangePoint {
	target := max(2, min(15, int(math.Sqrt(float64(len(w))/2))))
	labels := binarySegments(w, target)
	var out []ChangePoint
	for i := 1; i < len(labels); i++ {
		if labels[i] != labels[i-1] {
			out = append(out, ChangePoint{w[i].WindowSize, w[i].Center, "binary_segmentation", w[i].JSDistance, 0})
		}
	}
	return out
}

func stableBoundaries(changes []ChangePoint, sizes []int) []StableBoundary {
	x := append([]ChangePoint(nil), changes...)
	sort.SliceStable(x, func(i, j int) bool {
		if x[i].Strength == x[j].Strength {
			return x[i].Position < x[j].Position
		}
		return x[i].Strength > x[j].Strength
	})
	type group struct{ points []ChangePoint }
	var groups []group
	for _, p := range x {
		best := -1
		bestD := math.MaxInt
		for i, g := range groups {
			m := 0
			for _, q := range g.points {
				m += q.Position
			}
			m /= len(g.points)
			d := abs(p.Position - m)
			tol := min(p.WindowSize, g.points[0].WindowSize) / 2
			if d <= tol && d < bestD {
				best, bestD = i, d
			}
		}
		if best < 0 {
			groups = append(groups, group{[]ChangePoint{p}})
		} else {
			groups[best].points = append(groups[best].points, p)
		}
	}
	var out []StableBoundary
	for _, g := range groups {
		support := map[int]bool{}
		positions := []float64{}
		strengths := []float64{}
		seenPosition := map[int]bool{}
		for _, p := range g.points {
			support[p.WindowSize] = true
			strengths = append(strengths, p.Strength)
			if !seenPosition[p.WindowSize] {
				positions = append(positions, float64(p.Position))
				seenPosition[p.WindowSize] = true
			}
		}
		mp := mean(positions)
		spread := 0.
		for _, p := range positions {
			d := math.Abs(p - mp)
			if d > spread {
				spread = d
			}
		}
		mx := 0.
		for _, s := range strengths {
			if s > mx {
				mx = s
			}
		}
		out = append(out, StableBoundary{int(math.Round(mp)), support, len(support), float64(len(support)) / float64(len(sizes)), mp, mean(strengths), mx, spread})
	}
	sort.SliceStable(out, func(i, j int) bool {
		if out[i].SupportCount == out[j].SupportCount {
			return out[i].MeanJumpStrength > out[j].MeanJumpStrength
		}
		return out[i].SupportCount > out[j].SupportCount
	})
	return out
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}
func maxFloat(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}
