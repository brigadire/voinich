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

// sortedProfileKeys returns p's keys in sorted order, so callers that
// accumulate a single running sum fed by every key of a profile do so
// deterministically (Go's map iteration order is randomized per range
// execution - see determinism_test.go).
func sortedProfileKeys(p profile) []string {
	keys := make([]string, 0, len(p))
	for k := range p {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

// sortedUnionKeys returns the sorted union of a's and b's keys.
func sortedUnionKeys(a, b profile) []string {
	keys := make([]string, 0, len(a)+len(b))
	seen := make(map[string]bool, len(a)+len(b))
	for k := range a {
		seen[k] = true
		keys = append(keys, k)
	}
	for k := range b {
		if !seen[k] {
			keys = append(keys, k)
		}
	}
	sort.Strings(keys)
	return keys
}

func jsSimilarity(a, b profile) float64 {
	d := 0.
	for _, k := range sortedUnionKeys(a, b) {
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
	for _, k := range sortedProfileKeys(a) {
		x := a[k]
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

// dotProduct is sum_k a[k]*b[k] over a's own keys (matching cosine's
// original loop exactly: b's keys not present in a never contribute, since
// the numerator only ever ranges over a).
func dotProduct(a, b profile) float64 {
	n := 0.
	for _, k := range sortedProfileKeys(a) {
		n += a[k] * b[k]
	}
	return n
}
func cosine(a, b profile) float64 {
	aa, bb := 0., 0.
	for _, k := range sortedProfileKeys(a) {
		x := a[k]
		aa += x * x
	}
	for _, k := range sortedProfileKeys(b) {
		x := b[k]
		bb += x * x
	}
	if aa == 0 || bb == 0 {
		return 0
	}
	return dotProduct(a, b) / math.Sqrt(aa*bb)
}

// sortedProfile pairs a profile with its own keys sorted once, so
// jsSimilaritySorted can merge-walk two already-sorted key lists instead of
// re-sorting their union on every pairwise call. dispersion/
// pairwiseDispersion each compare the same profiles against many others
// (O(n) and O(n^2) respectively), so sorting each profile's keys exactly
// once here - instead of on every one of those repeated comparisons -
// matters: profiling conditionalregime's equivalent hot path (task27)
// showed the sort-per-call approach alone was over 70% of that CLI's total
// CPU time.
type sortedProfile struct {
	p    profile
	keys []string
}

func sortProfile(p profile) sortedProfile {
	return sortedProfile{p: p, keys: sortedProfileKeys(p)}
}

// jsSimilaritySorted is jsSimilarity, but merge-walking a's and b's
// already-sorted key lists instead of re-sorting their union - the merge
// visits the sorted union in the same order sorting it directly would, so
// this is bit-identical to jsSimilarity(a.p, b.p).
func jsSimilaritySorted(a, b sortedProfile) float64 {
	ak, bk := a.keys, b.keys
	i, j := 0, 0
	d := 0.0
	for i < len(ak) && j < len(bk) {
		switch {
		case ak[i] < bk[j]:
			x := a.p[ak[i]]
			if x > 0 {
				d += .5 * x * math.Log2(2)
			}
			i++
		case ak[i] > bk[j]:
			y := b.p[bk[j]]
			if y > 0 {
				d += .5 * y * math.Log2(2)
			}
			j++
		default:
			x, y := a.p[ak[i]], b.p[bk[j]]
			m := (x + y) / 2
			if x > 0 {
				d += .5 * x * math.Log2(x/m)
			}
			if y > 0 {
				d += .5 * y * math.Log2(y/m)
			}
			i++
			j++
		}
	}
	for ; i < len(ak); i++ {
		x := a.p[ak[i]]
		if x > 0 {
			d += .5 * x * math.Log2(2)
		}
	}
	for ; j < len(bk); j++ {
		y := b.p[bk[j]]
		if y > 0 {
			d += .5 * y * math.Log2(2)
		}
	}
	v := 1 - d
	if v < 0 {
		return 0
	}
	return v
}

// weightedOverlapSorted, dotProductSorted and cosineSorted mirror
// weightedOverlap/dotProduct/cosine, but read a's/b's keys off an
// already-sorted sortedProfile instead of sorting them fresh - useful when
// the same (a,b) pair already had its keys sorted once for a
// jsSimilaritySorted call and would otherwise pay for sorting the same
// keys again for these related metrics.
func weightedOverlapSorted(a, b sortedProfile) float64 {
	s := 0.
	for _, k := range a.keys {
		x := a.p[k]
		if b.p[k] < x {
			x = b.p[k]
		}
		s += x
	}
	return s
}
func dotProductSorted(a, b sortedProfile) float64 {
	n := 0.
	for _, k := range a.keys {
		n += a.p[k] * b.p[k]
	}
	return n
}
func cosineSorted(a, b sortedProfile) float64 {
	aa, bb := 0., 0.
	for _, k := range a.keys {
		x := a.p[k]
		aa += x * x
	}
	for _, k := range b.keys {
		x := b.p[k]
		bb += x * x
	}
	if aa == 0 || bb == 0 {
		return 0
	}
	return dotProductSorted(a, b) / math.Sqrt(aa*bb)
}

func dispersion(ps []profile, c profile) float64 {
	if len(ps) == 0 {
		return 0
	}
	sc := sortProfile(c)
	s := 0.
	for _, p := range ps {
		s += 1 - jsSimilaritySorted(sortProfile(p), sc)
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
	var touched []int
	for i := 0; i < len(ps); i += step {
		touched = append(touched, i)
	}
	sorted := make(map[int]sortedProfile, len(touched))
	for _, i := range touched {
		sorted[i] = sortProfile(ps[i])
	}
	s, n := 0., 0
	for _, i := range touched {
		for _, j := range touched {
			if j <= i {
				continue
			}
			s += 1 - jsSimilaritySorted(sorted[i], sorted[j])
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
	return distanceDistributionAt(c, positions(c, t), d, respect)
}

// distanceDistributionAt is distanceDistributionMode with pos precomputed:
// positions(c,t) doesn't depend on d, so a caller sweeping d=1..MaxDistance
// for a fixed (corpus variant, token) should compute it once and pass it in
// here, instead of distanceDistributionMode recomputing it on every d.
func distanceDistributionAt(c corpus, pos []int, d int, respect bool) profile {
	m := map[string]int{}
	for _, i := range pos {
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
	for _, k := range sortedProfileKeys(p) {
		v := p[k]
		s += v * v
	}
	return s
}
func concentrationSorted(sp sortedProfile) float64 {
	s := 0.
	for _, k := range sp.keys {
		v := sp.p[k]
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
