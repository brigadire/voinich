package metadatavalidation

import (
	"bufio"
	"fmt"
	"math"
	"math/rand"
	"os"
	"sort"
	"strconv"
	"strings"
)

var metadataKinds = []string{"line", "paragraph", "folio", "currier", "hand", "quire"}

func ExtractBoundaries(records []TokenMetadata) map[string][]MetadataBoundary {
	out := map[string][]MetadataBoundary{}
	if len(records) < 2 {
		return out
	}
	add := func(kind string, p int, a, b string) { out[kind] = append(out[kind], MetadataBoundary{p, kind, a, b}) }
	for i := 1; i < len(records); i++ {
		a, b := records[i-1], records[i]
		p := b.Position
		if a.LineID != b.LineID {
			add("line", p, a.LineID, b.LineID)
		}
		if a.Folio != b.Folio {
			add("folio", p, a.Folio, b.Folio)
		}
		if a.Folio != b.Folio || a.ParagraphID != b.ParagraphID {
			add("paragraph", p, fmt.Sprint(a.ParagraphID), fmt.Sprint(b.ParagraphID))
		}
		knownChange := func(kind, x, y string) {
			if x != "" && y != "" && x != "?" && y != "?" && x != y {
				add(kind, p, x, y)
			}
		}
		knownChange("currier", a.Currier, b.Currier)
		knownChange("hand", a.Hand, b.Hand)
		knownChange("quire", a.Quire, b.Quire)
	}
	return out
}

func positions(x []MetadataBoundary) []int {
	r := make([]int, len(x))
	for i, v := range x {
		r[i] = v.Position
	}
	sort.Ints(r)
	return r
}
func NearestBoundary(position int, boundaries []int) (int, int) {
	if len(boundaries) == 0 {
		return -1, -1
	}
	i := sort.SearchInts(boundaries, position)
	if i == 0 {
		return boundaries[0], abs(boundaries[0] - position)
	}
	if i == len(boundaries) {
		return boundaries[i-1], abs(boundaries[i-1] - position)
	}
	a, b := boundaries[i-1], boundaries[i]
	if position-a <= b-position {
		return a, position - a
	}
	return b, b - position
}
func MatchWithin(blind, bounds []int, tolerance int) int {
	n := 0
	for _, p := range blind {
		_, d := NearestBoundary(p, bounds)
		if d >= 0 && d <= tolerance {
			n++
		}
	}
	return n
}
func distances(blind, bounds []int) []int {
	r := make([]int, 0, len(blind))
	for _, p := range blind {
		_, d := NearestBoundary(p, bounds)
		if d >= 0 {
			r = append(r, d)
		}
	}
	return r
}

// UniformBoundaries draws `count` uniformly-random distinct positions in
// [1, n-1] by computing the same Fisher-Yates permutation
// *rand.Rand.Perm(n-1) does internally (same algorithm, same sequence of
// rng.Intn calls, same resulting values) and taking its first `count`
// entries, but into scratch instead of a freshly allocated slice.
// scratch's contents before the call are never read: rand.Perm's loop
// writes to position i (via `m[i] = m[j]`) before any later step can read
// it back (j is always <= i), so every element scratch[:n-1] touches gets
// overwritten before it is used, regardless of what it held on entry — see
// determinism_test.go for the proof this is called out and verified by.
// Callers own scratch and must reuse the same slice (with capacity >=
// n-1) across every call in a replicate loop instead of allocating fresh
// each time; UniformBoundaries never retains it.
func UniformBoundaries(n, count int, rng *rand.Rand, scratch []int) []int {
	if n <= 1 || count <= 0 {
		return nil
	}
	if count > n-1 {
		count = n - 1
	}
	m := scratch[:n-1]
	for i := 0; i < len(m); i++ {
		j := rng.Intn(i + 1)
		m[i] = m[j]
		m[j] = i
	}
	p := make([]int, count)
	copy(p, m[:count])
	for i := range p {
		p[i]++
	}
	sort.Ints(p)
	return p
}
func CircularShiftBoundaries(x []int, n, shift int) []int {
	r := make([]int, len(x))
	if n <= 1 {
		return r
	}
	m := n - 1
	for i, p := range x {
		r[i] = 1 + mod(p-1+shift, m)
	}
	sort.Ints(r)
	return r
}

func ValidateBoundaries(stable []StableBoundary, refs map[string][]MetadataBoundary, tolerances []int, permutations, n int, seed int64) ([]BoundaryValidation, map[string][][]float64) {
	rng := rand.New(rand.NewSource(seed))
	rows := []BoundaryValidation{}
	rawNull := map[string][][]float64{}
	// Reused across every UniformBoundaries call in the loops below instead
	// of a fresh n-1-length slice per call (900,000+ calls at default
	// settings) — see UniformBoundaries' doc comment for why this is safe.
	permScratch := make([]int, max(0, n-1))
	for _, support := range []int{3, 4, 5} {
		blind := []int{}
		for _, b := range stable {
			if b.Support >= support {
				blind = append(blind, b.Position)
			}
		}
		for _, kind := range metadataKinds {
			ref := positions(refs[kind])
			if len(ref) == 0 {
				continue
			}
			ds := distances(blind, ref)
			meanD, medD := meanInts(ds), medianInts(ds)
			for _, tol := range tolerances {
				obs := MatchWithin(blind, ref, tol)
				un := make([]float64, permutations)
				cir := make([]float64, permutations)
				for p := 0; p < permutations; p++ {
					un[p] = float64(MatchWithin(UniformBoundaries(n, len(blind), rng, permScratch), ref, tol))
					shift := 1 + rng.Intn(max(1, n-1))
					cir[p] = float64(MatchWithin(CircularShiftBoundaries(blind, n, shift), ref, tol))
				}
				key := fmt.Sprintf("%s/support_%d/tolerance_%d", kind, support, tol)
				rawNull[key] = [][]float64{un, cir}
				rows = append(rows, BoundaryValidation{kind, support, tol, len(blind), obs, ratio(obs, len(blind)), meanD, medD, meanFloat(un), percentileFloat(un, float64(obs)), meanFloat(cir), percentileFloat(cir, float64(obs))})
			}
		}
	}
	return rows, rawNull
}

type Assignment struct {
	WindowSize                    int
	Method                        string
	K, Index, Start, End, Cluster int
}

func readTSV(path string) ([]map[string]string, error) {
	f, e := os.Open(path)
	if e != nil {
		return nil, e
	}
	defer f.Close()
	s := bufio.NewScanner(f)
	s.Buffer(make([]byte, 64*1024), 16*1024*1024)
	if !s.Scan() {
		if e := s.Err(); e != nil {
			return nil, e
		}
		return nil, nil
	}
	header := strings.Split(strings.TrimSuffix(s.Text(), "\r"), "\t")
	out := []map[string]string{}
	for s.Scan() {
		row := strings.Split(strings.TrimSuffix(s.Text(), "\r"), "\t")
		m := map[string]string{}
		for i, k := range header {
			if i < len(row) {
				m[k] = row[i]
			}
		}
		out = append(out, m)
	}
	return out, s.Err()
}
func atoi(s string) int     { v, _ := strconv.Atoi(s); return v }
func atof(s string) float64 { v, _ := strconv.ParseFloat(s, 64); return v }
func LoadStable(path string) ([]StableBoundary, error) {
	x, e := readTSV(path)
	if e != nil {
		return nil, e
	}
	r := make([]StableBoundary, 0, len(x))
	for _, m := range x {
		r = append(r, StableBoundary{atoi(m["position"]), atoi(m["support_count"]), atof(m["mean_jump_strength"]), atof(m["position_uncertainty"])})
	}
	return r, nil
}
func LoadAssignments(path string) ([]Assignment, error) {
	x, e := readTSV(path)
	if e != nil {
		return nil, e
	}
	r := make([]Assignment, 0, len(x))
	for _, m := range x {
		r = append(r, Assignment{atoi(m["window_size"]), m["method"], atoi(m["k"]), atoi(m["window_index"]), atoi(m["start"]), atoi(m["end"]), atoi(m["cluster"])})
	}
	return r, nil
}

type WindowMetadata struct {
	Label       string
	Purity      float64
	Composition map[string]float64
}

func MetadataComposition(records []TokenMetadata, start, end int, kind string) WindowMetadata {
	if start < 0 {
		start = 0
	}
	if end > len(records) {
		end = len(records)
	}
	counts := map[string]int{}
	total := 0
	for _, r := range records[start:end] {
		v := ""
		if kind == "currier" {
			v = r.Currier
		} else {
			v = r.Hand
		}
		if v == "" || v == "?" {
			continue
		}
		counts[v]++
		total++
	}
	w := WindowMetadata{Composition: map[string]float64{}}
	best := 0
	for k, v := range counts {
		w.Composition[k] = ratio(v, total)
		if v > best || (v == best && k < w.Label) {
			best = v
			w.Label = k
		}
	}
	w.Purity = ratio(best, total)
	return w
}

type Metrics struct {
	MI, NMI, ARI, Homogeneity, Completeness float64
	Contingency                             string
}

func AssociationMetrics(a, b []string) Metrics {
	n := min(len(a), len(b))
	tab := map[string]map[string]int{}
	ra, cb := map[string]int{}, map[string]int{}
	valid := 0
	for i := 0; i < n; i++ {
		if a[i] == "" || b[i] == "" {
			continue
		}
		if tab[a[i]] == nil {
			tab[a[i]] = map[string]int{}
		}
		tab[a[i]][b[i]]++
		ra[a[i]]++
		cb[b[i]]++
		valid++
	}
	if valid == 0 {
		return Metrics{}
	}
	comb := func(x int) float64 { return float64(x*(x-1)) / 2 }
	// mi and sumCell are single running sums fed by every cell of tab, so
	// summing in map iteration order made them nondeterministic across
	// otherwise byte-identical calls (see determinism_test.go). Both loops
	// visit the same cells, so they're merged into one pass over the rows
	// (sorted) x columns (sorted per row) instead of two separate passes.
	rows := make([]string, 0, len(tab))
	for x := range tab {
		rows = append(rows, x)
	}
	sort.Strings(rows)
	mi, sumCell := 0., 0.
	for _, x := range rows {
		row := tab[x]
		cols := make([]string, 0, len(row))
		for y := range row {
			cols = append(cols, y)
		}
		sort.Strings(cols)
		for _, y := range cols {
			v := row[y]
			p := float64(v) / float64(valid)
			mi += p * math.Log(p/(float64(ra[x]*cb[y])/float64(valid*valid)))
			sumCell += comb(v)
		}
	}
	ha, hb := entropyCounts(ra, valid), entropyCounts(cb, valid)
	nmi := 0.
	if ha+hb > 0 {
		nmi = 2 * mi / (ha + hb)
	}
	hom, comp := 1., 1.
	if ha > 0 {
		hom = mi / ha
	}
	if hb > 0 {
		comp = mi / hb
	}
	// sumA/sumB have the same single-running-sum-over-a-map hazard as
	// mi/sumCell above.
	sumA, sumB := sortedIntMapSum(ra, comb), sortedIntMapSum(cb, comb)
	den := comb(valid)
	ari := 0.
	if den > 0 {
		expected := sumA * sumB / den
		d := .5*(sumA+sumB) - expected
		if d != 0 {
			ari = (sumCell - expected) / d
		}
	}
	return Metrics{mi, nmi, ari, hom, comp, formatContingency(tab)}
}
func formatContingency(t map[string]map[string]int) string {
	xs := make([]string, 0, len(t))
	for x := range t {
		xs = append(xs, x)
	}
	sort.Strings(xs)
	parts := []string{}
	for _, x := range xs {
		ys := make([]string, 0, len(t[x]))
		for y := range t[x] {
			ys = append(ys, y)
		}
		sort.Strings(ys)
		for _, y := range ys {
			parts = append(parts, fmt.Sprintf("%s:%s=%d", x, y, t[x][y]))
		}
	}
	return strings.Join(parts, ",")
}

// sortedIntMapSum sums f(v) over m's values in sorted-key order, so the
// result does not depend on Go's randomized map iteration order — needed
// anywhere a single running sum is fed by every entry of a map (see
// determinism_test.go).
func sortedIntMapSum(m map[string]int, f func(int) float64) float64 {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	s := 0.
	for _, k := range keys {
		s += f(m[k])
	}
	return s
}
func entropyCounts(c map[string]int, n int) float64 {
	return sortedIntMapSum(c, func(v int) float64 {
		p := float64(v) / float64(n)
		return -p * math.Log(p)
	})
}

func AnalyzeAssignments(x []Assignment, records []TokenMetadata) []Association {
	out := []Association{}
	compositionCache := map[string]WindowMetadata{}
	for _, kind := range []string{"currier", "hand"} {
		type key struct {
			w int
			m string
			k int
		}
		groups := map[key][]Assignment{}
		for _, a := range x {
			q := key{a.WindowSize, a.Method, a.K}
			groups[q] = append(groups[q], a)
		}
		keys := make([]key, 0, len(groups))
		for k := range groups {
			keys = append(keys, k)
		}
		sort.Slice(keys, func(i, j int) bool {
			if keys[i].w != keys[j].w {
				return keys[i].w < keys[j].w
			}
			if keys[i].m != keys[j].m {
				return keys[i].m < keys[j].m
			}
			return keys[i].k < keys[j].k
		})
		for _, q := range keys {
			g := groups[q]
			for _, threshold := range []float64{0, .8, .9} {
				labels, clusters := []string{}, []string{}
				for _, a := range g {
					cacheKey := fmt.Sprintf("%s/%d/%d", kind, a.Start, a.End)
					wm, ok := compositionCache[cacheKey]
					if !ok {
						wm = MetadataComposition(records, a.Start, a.End, kind)
						compositionCache[cacheKey] = wm
					}
					if wm.Label == "" || wm.Purity < threshold {
						continue
					}
					labels = append(labels, wm.Label)
					clusters = append(clusters, strconv.Itoa(a.Cluster))
				}
				m := AssociationMetrics(labels, clusters)
				subset := "all"
				if threshold > 0 {
					subset = fmt.Sprintf("purity>=%.1f", threshold)
				}
				he := metadataEntropy(labels)
				conditional := conditionalEntropy(labels, clusters)
				reduction := 0.
				if he > 0 {
					reduction = 1 - conditional/he
				}
				out = append(out, Association{q.w, q.m, q.k, kind, subset, len(labels), m.MI, m.NMI, m.ARI, m.Homogeneity, m.Completeness, conditional, reduction, m.Contingency})
			}
		}
	}
	return out
}
func metadataEntropy(x []string) float64 {
	c := map[string]int{}
	for _, v := range x {
		c[v]++
	}
	return entropyCounts(c, len(x))
}
func conditionalEntropy(labels, clusters []string) float64 {
	n := min(len(labels), len(clusters))
	by := map[string]map[string]int{}
	sizes := map[string]int{}
	for i := 0; i < n; i++ {
		if by[clusters[i]] == nil {
			by[clusters[i]] = map[string]int{}
		}
		by[clusters[i]][labels[i]]++
		sizes[clusters[i]]++
	}
	// h is a single running sum fed by every cluster key in by, so summing
	// in map iteration order made it nondeterministic across otherwise
	// byte-identical calls (see determinism_test.go).
	keys := make([]string, 0, len(by))
	for k := range by {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	h := 0.
	for _, k := range keys {
		h += float64(sizes[k]) / float64(n) * entropyCounts(by[k], sizes[k])
	}
	return h
}

// PermuteBlockLabels shuffles labels among contiguous blocks without changing
// any block length. It is the block-aware null used for metadata associations.
func PermuteBlockLabels(labels []string, rng *rand.Rand) []string {
	if len(labels) == 0 {
		return nil
	}
	type block struct {
		v string
		n int
	}
	bs := []block{}
	for _, v := range labels {
		if len(bs) == 0 || bs[len(bs)-1].v != v {
			bs = append(bs, block{v, 1})
		} else {
			bs[len(bs)-1].n++
		}
	}
	vals := make([]string, len(bs))
	for i, b := range bs {
		vals[i] = b.v
	}
	rng.Shuffle(len(vals), func(i, j int) { vals[i], vals[j] = vals[j], vals[i] })
	out := []string{}
	for i, b := range bs {
		for j := 0; j < b.n; j++ {
			out = append(out, vals[i])
		}
	}
	return out
}

func meanInts(x []int) float64 {
	if len(x) == 0 {
		return 0
	}
	s := 0
	for _, v := range x {
		s += v
	}
	return float64(s) / float64(len(x))
}
func medianInts(x []int) float64 {
	if len(x) == 0 {
		return 0
	}
	y := append([]int(nil), x...)
	sort.Ints(y)
	if len(y)%2 == 1 {
		return float64(y[len(y)/2])
	}
	return float64(y[len(y)/2-1]+y[len(y)/2]) / 2
}
func meanFloat(x []float64) float64 {
	if len(x) == 0 {
		return 0
	}
	s := 0.
	for _, v := range x {
		s += v
	}
	return s / float64(len(x))
}
func percentileFloat(x []float64, v float64) float64 {
	if len(x) == 0 {
		return 0
	}
	n := 0
	for _, q := range x {
		if q <= v {
			n++
		}
	}
	return 100 * float64(n) / float64(len(x))
}
func ratio(a, b int) float64 {
	if b == 0 {
		return 0
	}
	return float64(a) / float64(b)
}
func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}
func mod(a, b int) int {
	r := a % b
	if r < 0 {
		r += b
	}
	return r
}
