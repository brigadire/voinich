package localregime

import (
	"math"
	"math/rand"
	"sort"
)

func normalizeCounts(m map[string]int) profile {
	p := profile{}
	n := 0
	for _, v := range m {
		n += v
	}
	if n == 0 {
		return p
	}
	for k, v := range m {
		p[k] = float64(v) / float64(n)
	}
	return p
}
func localProfile(c corpus, pos, radius, gap int, side string, respect bool) profile {
	m := map[string]int{}
	for d := gap + 1; d <= radius; d++ {
		if side != "right" {
			j := pos - d
			if j >= 0 && (!respect || c.LineAt[j] == c.LineAt[pos]) {
				m[c.Tokens[j]]++
			}
		}
		if side != "left" {
			j := pos + d
			if j < len(c.Tokens) && (!respect || c.LineAt[j] == c.LineAt[pos]) {
				m[c.Tokens[j]]++
			}
		}
	}
	return normalizeCounts(m)
}

type offsetCounts map[int]map[string]int

func buildOffsetCounts(c corpus, pos []int, radius int, respect bool) offsetCounts {
	out := offsetCounts{}
	for d := -radius; d <= radius; d++ {
		if d != 0 {
			out[d] = map[string]int{}
		}
	}
	for _, center := range pos {
		for d := -radius; d <= radius; d++ {
			if d == 0 {
				continue
			}
			j := center + d
			if j >= 0 && j < len(c.Tokens) && (!respect || c.LineAt[j] == c.LineAt[center]) {
				out[d][c.Tokens[j]]++
			}
		}
	}
	return out
}

func offsetProfile(x offsetCounts, radius, gap int, side string) profile {
	m := map[string]int{}
	for d := gap + 1; d <= radius; d++ {
		if side != "right" {
			for k, n := range x[-d] {
				m[k] += n
			}
		}
		if side != "left" {
			for k, n := range x[d] {
				m[k] += n
			}
		}
	}
	return normalizeCounts(m)
}
func aggregate(ps []profile) profile {
	m := profile{}
	for _, p := range ps {
		for k, v := range p {
			m[k] += v
		}
	}
	if len(ps) > 0 {
		for k := range m {
			m[k] /= float64(len(ps))
		}
	}
	return m
}
func jsSimilarity(a, b profile) float64 {
	keys := map[string]bool{}
	for k := range a {
		keys[k] = true
	}
	for k := range b {
		keys[k] = true
	}
	d := 0.
	for k := range keys {
		x, y := a[k], b[k]
		m := (x + y) / 2
		if x > 0 {
			d += .5 * x * math.Log2(x/m)
		}
		if y > 0 {
			d += .5 * y * math.Log2(y/m)
		}
	}
	v := 1 - d
	if v < 0 {
		return 0
	}
	return v
}
func weightedOverlap(a, b profile) float64 {
	s := 0.
	for k, x := range a {
		if b[k] < x {
			x = b[k]
		}
		s += x
	}
	return s
}
func jaccard(a, b profile) float64 {
	u, i := 0, 0
	seen := map[string]bool{}
	for k := range a {
		seen[k] = true
	}
	for k := range b {
		seen[k] = true
	}
	for k := range seen {
		u++
		if a[k] > 0 && b[k] > 0 {
			i++
		}
	}
	if u == 0 {
		return 0
	}
	return float64(i) / float64(u)
}
func cosine(a, b profile) float64 {
	n, aa, bb := 0., 0., 0.
	for k, x := range a {
		n += x * b[k]
		aa += x * x
	}
	for _, x := range b {
		bb += x * x
	}
	if aa == 0 || bb == 0 {
		return 0
	}
	return n / math.Sqrt(aa*bb)
}
func dispersion(ps []profile, c profile) float64 {
	if len(ps) == 0 {
		return 0
	}
	s := 0.
	for _, p := range ps {
		s += 1 - jsSimilarity(p, c)
	}
	return s / float64(len(ps))
}
func pairwiseDispersion(ps []profile) float64 {
	if len(ps) < 2 {
		return 0
	}
	// Deterministic subsampling keeps the diagnostic bounded for very frequent
	// tokens while retaining profiles from across the full occurrence series.
	step := 1
	if len(ps) > 200 {
		step = (len(ps) + 199) / 200
	}
	s, n := 0., 0
	for i := 0; i < len(ps); i += step {
		for j := i + step; j < len(ps); j += step {
			s += 1 - jsSimilarity(ps[i], ps[j])
			n++
		}
	}
	if n == 0 {
		return 0
	}
	return s / float64(n)
}
func positions(c corpus, t string) []int {
	var x []int
	for i, v := range c.Tokens {
		if v == t {
			x = append(x, i)
		}
	}
	return x
}
func distanceDistribution(c corpus, t string, d int) profile {
	return distanceDistributionMode(c, t, d, false)
}
func distanceDistributionMode(c corpus, t string, d int, respect bool) profile {
	m := map[string]int{}
	for _, i := range positions(c, t) {
		j := i + d
		if j < len(c.Tokens) && (!respect || c.LineAt[j] == c.LineAt[i]) {
			m[c.Tokens[j]]++
		}
	}
	return normalizeCounts(m)
}
func residualDependency(observed, expected float64) float64 { return observed - expected }
func shuffleCorpus(c corpus, mode string, block int, seed int64) corpus {
	x := c
	x.Tokens = append([]string(nil), c.Tokens...)
	r := rand.New(rand.NewSource(seed))
	switch mode {
	case "global":
		r.Shuffle(len(x.Tokens), func(i, j int) { x.Tokens[i], x.Tokens[j] = x.Tokens[j], x.Tokens[i] })
	case "line":
		for _, line := range c.Lines {
			_ = line
		}
		start := 0
		for _, line := range c.Lines {
			r.Shuffle(len(line), func(i, j int) { x.Tokens[start+i], x.Tokens[start+j] = x.Tokens[start+j], x.Tokens[start+i] })
			start += len(line)
		}
	case "block":
		for start := 0; start < len(x.Tokens); start += block {
			end := start + block
			if end > len(x.Tokens) {
				end = len(x.Tokens)
			}
			r.Shuffle(end-start, func(i, j int) { x.Tokens[start+i], x.Tokens[start+j] = x.Tokens[start+j], x.Tokens[start+i] })
		}
	}
	return x
}
func slidingProfiles(c corpus, size, step int) ([]profile, []WindowRow) {
	var ps []profile
	var rows []WindowRow
	for start := 0; start+size <= len(c.Tokens); start += step {
		m := map[string]int{}
		for _, t := range c.Tokens[start : start+size] {
			m[t]++
		}
		p := normalizeCounts(m)
		row := WindowRow{Size: size, Index: len(ps), Start: start, End: start + size, Concentration: concentration(p)}
		if len(ps) > 0 {
			row.AdjacentJSDistance = 1 - jsSimilarity(ps[len(ps)-1], p)
		}
		ps = append(ps, p)
		rows = append(rows, row)
	}
	return ps, rows
}
func concentration(p profile) float64 {
	s := 0.
	for _, v := range p {
		s += v * v
	}
	return s
}
func blockShuffle(tokens []string, block int, seed int64) []string {
	x := append([]string(nil), tokens...)
	r := rand.New(rand.NewSource(seed))
	for i := 0; i < len(x); i += block {
		e := i + block
		if e > len(x) {
			e = len(x)
		}
		r.Shuffle(e-i, func(a, b int) { x[i+a], x[i+b] = x[i+b], x[i+a] })
	}
	return x
}
func changePoints(rows []WindowRow) []ChangePoint {
	if len(rows) < 3 {
		return nil
	}
	v := make([]float64, 0, len(rows)-1)
	for _, r := range rows[1:] {
		v = append(v, r.AdjacentJSDistance)
	}
	// This is a diagnostic threshold, not a fitted number of regimes.
	threshold := mean(v) + stddev(v)
	var out []ChangePoint
	for _, r := range rows[1:] {
		if r.AdjacentJSDistance > threshold {
			out = append(out, ChangePoint{r.Size, r.Start, r.AdjacentJSDistance, threshold})
		}
	}
	return out
}
func mean(x []float64) float64 {
	if len(x) == 0 {
		return 0
	}
	s := 0.
	for _, v := range x {
		s += v
	}
	return s / float64(len(x))
}
func stddev(x []float64) float64 {
	m := mean(x)
	s := 0.
	for _, v := range x {
		s += (v - m) * (v - m)
	}
	if len(x) == 0 {
		return 0
	}
	return math.Sqrt(s / float64(len(x)))
}
func pearson(a, b []float64) float64 {
	if len(a) != len(b) || len(a) < 2 {
		return 0
	}
	ma, mb := mean(a), mean(b)
	n, da, db := 0., 0., 0.
	for i := range a {
		x, y := a[i]-ma, b[i]-mb
		n += x * y
		da += x * x
		db += y * y
	}
	if da == 0 || db == 0 {
		return 0
	}
	return n / math.Sqrt(da*db)
}
func ranks(x []float64) []float64 {
	idx := make([]int, len(x))
	for i := range idx {
		idx[i] = i
	}
	sort.SliceStable(idx, func(i, j int) bool { return x[idx[i]] < x[idx[j]] })
	r := make([]float64, len(x))
	for i := 0; i < len(idx); {
		j := i + 1
		for j < len(idx) && x[idx[j]] == x[idx[i]] {
			j++
		}
		q := float64(i+j-1)/2 + 1
		for k := i; k < j; k++ {
			r[idx[k]] = q
		}
		i = j
	}
	return r
}
func spearman(a, b []float64) float64 { return pearson(ranks(a), ranks(b)) }
func retainedEffect(original, shuffled, baseline float64) float64 {
	den := original - baseline
	if math.Abs(den) < 1e-12 {
		return 0
	}
	return (shuffled - baseline) / den
}
