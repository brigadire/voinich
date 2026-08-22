// Package localregimetopology contains Task65's window/coordinate,
// change-point and clustering primitives. Token/glyph distance and the
// primary structural profile are never redefined here: both come from
// the authoritative internal/lineregime (task64 sections 4, 22-23; task65
// sections 4, 11).
package localregimetopology

import (
	"math"
	"sort"

	"zcore.dev/voinich/internal/lineregime"
)

// TokenRecord places one corpus token on the unified manuscript
// coordinate x (task65 section 5).
type TokenRecord struct {
	GlobalIndex                   int
	Line, PageIndex               int
	Folio, Currier, Hand, Section string
	Glyphs                        []string
}

// Window is one fixed-size span of tokens in GlobalIndex order (task65
// sections 6-7). Windows may cross line/page boundaries; CrossesLine and
// CrossesPage record that instead of forbidding it, so the boundary
// effect can be measured directly (sections 15-17).
type Window struct {
	Index                    int
	StartGlobal, EndGlobal   int
	Folio                    string
	Line                     int
	CrossesLine, CrossesPage bool
	Tokens                   [][]string
}

// BuildWindows builds overlapping fixed-size windows of w tokens stepping
// by step tokens over records in GlobalIndex order (task65 sections 6-7).
// step<1 is treated as step=1 rather than looping forever or crashing.
func BuildWindows(records []TokenRecord, w, step int) []Window {
	if step < 1 {
		step = 1
	}
	var out []Window
	idx := 0
	for start := 0; start+w <= len(records); start += step {
		end := start + w
		win := Window{Index: idx, StartGlobal: records[start].GlobalIndex, EndGlobal: records[end-1].GlobalIndex,
			Folio: records[start].Folio, Line: records[start].Line}
		toks := make([][]string, w)
		for i := 0; i < w; i++ {
			toks[i] = records[start+i].Glyphs
			if records[start+i].Folio != win.Folio {
				win.CrossesPage = true
			}
			if records[start+i].Line != win.Line {
				win.CrossesLine = true
			}
		}
		win.Tokens = toks
		out = append(out, win)
		idx++
	}
	return out
}

// ChangePoint is one candidate boundary between two adjacent
// non-overlapping windows (task65 section 18).
type ChangePoint struct {
	Position int
	Score    float64
}

// ScanChangePoints scores every boundary between adjacent, non-overlapping
// windows of size w using the authoritative Task64 profile distance
// (task65 section 18's simple deterministic method). Non-overlapping scan
// windows are used deliberately, so scores are not pseudo-replicated the
// way an every-token overlapping scan would be (section 7).
func ScanChangePoints(records []TokenRecord, w int, giant map[string]bool) []ChangePoint {
	var out []ChangePoint
	for start := w; start+w <= len(records); start += w {
		left := make([][]string, w)
		right := make([][]string, w)
		for i := 0; i < w; i++ {
			left[i] = records[start-w+i].Glyphs
			right[i] = records[start+i].Glyphs
		}
		pl := lineregime.ComputeProfile(left, giant)
		pr := lineregime.ComputeProfile(right, giant)
		out = append(out, ChangePoint{Position: records[start].GlobalIndex, Score: lineregime.ProfileDistance(pl, pr)})
	}
	return out
}

// CUSUMMax is task65 section 18's second, "formal" change-point method: it
// applies the classical single-change-point CUSUM statistic to a scalar
// score series (e.g. ScanChangePoints' scores), returning the position of
// the maximum absolute cumulative deviation from the series mean and that
// maximum itself. Significance is established by the caller via a
// permutation null over the same series (task65 section 19).
func CUSUMMax(scores []float64) (peakIndex int, peakValue float64) {
	if len(scores) == 0 {
		return -1, 0
	}
	mean := 0.0
	for _, s := range scores {
		mean += s
	}
	mean /= float64(len(scores))
	cum := 0.0
	for i, s := range scores {
		cum += s - mean
		if math.Abs(cum) > peakValue {
			peakValue = math.Abs(cum)
			peakIndex = i
		}
	}
	return
}

// KMedoids is a small deterministic PAM-style k-medoids clustering over
// arbitrary vectors using dist (task65 section 36). Medoid initialization
// is deterministic (evenly spaced indices over a stably sorted-by-first-
// coordinate order), so the same input always yields the same clustering
// (task65 section 73/76 test 18).
func KMedoids(vectors [][]float64, k, iters int) (assign []int, medoids []int) {
	n := len(vectors)
	if n == 0 || k <= 0 {
		return nil, nil
	}
	if k > n {
		k = n
	}
	order := make([]int, n)
	for i := range order {
		order[i] = i
	}
	sort.SliceStable(order, func(a, b int) bool {
		if len(vectors[order[a]]) == 0 || len(vectors[order[b]]) == 0 {
			return order[a] < order[b]
		}
		return vectors[order[a]][0] < vectors[order[b]][0]
	})
	medoids = make([]int, k)
	for i := 0; i < k; i++ {
		medoids[i] = order[i*n/k]
	}
	assign = make([]int, n)
	dist := func(a, b []float64) float64 {
		s := 0.0
		for i := range a {
			d := a[i] - b[i]
			s += d * d
		}
		return math.Sqrt(s)
	}
	for iter := 0; iter < iters; iter++ {
		for i, v := range vectors {
			best, bestD := 0, math.Inf(1)
			for mi, m := range medoids {
				d := dist(v, vectors[m])
				if d < bestD {
					bestD, best = d, mi
				}
			}
			assign[i] = best
		}
		changed := false
		for mi := range medoids {
			bestMedoid, bestCost := medoids[mi], math.Inf(1)
			for i, v := range vectors {
				if assign[i] != mi {
					continue
				}
				cost := 0.0
				for j, w := range vectors {
					if assign[j] == mi {
						cost += dist(v, w)
					}
				}
				if cost < bestCost {
					bestCost, bestMedoid = cost, i
				}
			}
			if bestMedoid != medoids[mi] {
				medoids[mi] = bestMedoid
				changed = true
			}
		}
		if !changed {
			break
		}
	}
	return assign, medoids
}

// StandardizeColumns z-scores every column of vectors using the mean/sd
// estimated from trainRows only (task65 section 10: normalization
// parameters come from TRAIN/DISCOVERY, never per-window). A zero-sd
// column is left at zero for every row (not divided) rather than
// producing NaN.
func StandardizeColumns(vectors [][]float64, trainRows []int) [][]float64 {
	if len(vectors) == 0 {
		return vectors
	}
	dims := len(vectors[0])
	mean := make([]float64, dims)
	for _, i := range trainRows {
		for d := 0; d < dims; d++ {
			mean[d] += vectors[i][d]
		}
	}
	n := float64(max(1, len(trainRows)))
	for d := range mean {
		mean[d] /= n
	}
	sd := make([]float64, dims)
	for _, i := range trainRows {
		for d := 0; d < dims; d++ {
			x := vectors[i][d] - mean[d]
			sd[d] += x * x
		}
	}
	for d := range sd {
		sd[d] = math.Sqrt(sd[d] / n)
	}
	out := make([][]float64, len(vectors))
	for i, v := range vectors {
		row := make([]float64, dims)
		for d := 0; d < dims; d++ {
			if sd[d] > 0 {
				row[d] = (v[d] - mean[d]) / sd[d]
			}
		}
		out[i] = row
	}
	return out
}
