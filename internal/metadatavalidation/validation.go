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

func UniformBoundaries(n, count int, rng *rand.Rand) []int {
	if n <= 1 || count <= 0 {
		return nil
	}
	if count > n-1 {
		count = n - 1
	}
	p := rng.Perm(n - 1)[:count]
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
					un[p] = float64(MatchWithin(UniformBoundaries(n, len(blind), rng), ref, tol))
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
	mi := 0.
	for x, row := range tab {
		for y, v := range row {
			p := float64(v) / float64(valid)
			mi += p * math.Log(p/(float64(ra[x]*cb[y])/float64(valid*valid)))
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
	comb := func(x int) float64 { return float64(x*(x-1)) / 2 }
	sumCell, sumA, sumB := 0., 0., 0.
	for _, row := range tab {
		for _, v := range row {
			sumCell += comb(v)
		}
	}
	for _, v := range ra {
		sumA += comb(v)
	}
	for _, v := range cb {
		sumB += comb(v)
	}
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
func entropyCounts(c map[string]int, n int) float64 {
	h := 0.
	for _, v := range c {
		p := float64(v) / float64(n)
		h -= p * math.Log(p)
	}
	return h
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
	h := 0.
	for k, c := range by {
		h += float64(sizes[k]) / float64(n) * entropyCounts(c, sizes[k])
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
