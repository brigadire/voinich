package propertytrajectory

import (
	"math"
	"math/rand"
	"sort"
	"strconv"
	"strings"

	"zcore.dev/voinich/internal/graphemic"
)

var propertyGroups = map[string][]string{
	"frequency":  {"global_count", "log_global_count", "global_frequency_probability", "frequency_rank", "normalized_frequency_rank"},
	"graphemic":  {"grapheme_length", "unique_grapheme_count"},
	"position":   {"line_start_probability", "line_end_probability", "mean_position", "median_position", "normalized_mean_position", "positional_entropy", "positional_specialization"},
	"context":    {"predecessor_entropy", "successor_entropy", "effective_predecessor_count", "effective_successor_count", "unique_predecessor_count", "unique_successor_count", "max_predecessor_probability", "max_successor_probability", "top3_predecessor_mass", "top3_successor_mass", "top5_predecessor_mass", "top5_successor_mass"},
	"structural": {"mean_structural_similarity_to_all", "max_structural_similarity", "structural_degree_above_0_5", "structural_degree_above_0_6", "structural_degree_above_0_7"},
}

func allPropertyNames() []string {
	var x []string
	for _, g := range []string{"frequency", "graphemic", "position", "context", "structural"} {
		x = append(x, propertyGroups[g]...)
	}
	return x
}

// entropy's h is a single running sum fed by every key of m, so it is
// accumulated in sorted key order: summing in map iteration order made it
// nondeterministic across otherwise byte-identical calls (see
// determinism_test.go). n (an integer sum) is unaffected by order.
func entropy(m map[string]int) float64 {
	n := 0
	for _, v := range m {
		n += v
	}
	if n == 0 {
		return 0
	}
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	h := 0.
	for _, k := range keys {
		p := float64(m[k]) / float64(n)
		h -= p * math.Log2(p)
	}
	return h
}
func masses(m map[string]int) (maxp, t3, t5 float64) {
	n := 0
	var v []int
	for _, q := range m {
		n += q
		v = append(v, q)
	}
	if n == 0 {
		return
	}
	sort.Sort(sort.Reverse(sort.IntSlice(v)))
	for i, q := range v {
		p := float64(q) / float64(n)
		if i == 0 {
			maxp = p
		}
		if i < 3 {
			t3 += p
		}
		if i < 5 {
			t5 += p
		}
	}
	return
}
func quantile(v []float64, p float64) float64 {
	if len(v) == 0 {
		return 0
	}
	x := append([]float64(nil), v...)
	sort.Float64s(x)
	q := p * float64(len(x)-1)
	lo, hi := int(math.Floor(q)), int(math.Ceil(q))
	if lo == hi {
		return x[lo]
	}
	return x[lo] + (x[hi]-x[lo])*(q-float64(lo))
}
func pearson(a, b []float64) float64 {
	if len(a) != len(b) || len(a) < 2 {
		return 0
	}
	ma, mb := mean(a), mean(b)
	var n, da, db float64
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
func percentileRank(x []float64, v float64) float64 {
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

func buildProperties(c corpus, edges []structuralEdge, minFreq int) (map[string]TokenProperties, map[string]map[string]float64) {
	pos := map[string][]float64{}
	npos := map[string][]float64{}
	starts, ends := map[string]int{}, map[string]int{}
	left, right := map[string]map[string]int{}, map[string]map[string]int{}
	for _, line := range c.Lines {
		for i, t := range line {
			pos[t] = append(pos[t], float64(i))
			den := max(1, len(line)-1)
			npos[t] = append(npos[t], float64(i)/float64(den))
			if i == 0 {
				starts[t]++
			} else {
				if left[t] == nil {
					left[t] = map[string]int{}
				}
				left[t][line[i-1]]++
			}
			if i == len(line)-1 {
				ends[t]++
			} else {
				if right[t] == nil {
					right[t] = map[string]int{}
				}
				right[t][line[i+1]]++
			}
		}
	}
	tokens := make([]string, 0)
	for t, n := range c.Counts {
		if n >= minFreq {
			tokens = append(tokens, t)
		}
	}
	sort.Slice(tokens, func(i, j int) bool {
		if c.Counts[tokens[i]] == c.Counts[tokens[j]] {
			return tokens[i] < tokens[j]
		}
		return c.Counts[tokens[i]] > c.Counts[tokens[j]]
	})
	rank := map[string]int{}
	for i, t := range tokens {
		rank[t] = i + 1
	}
	ssum, smax := map[string]float64{}, map[string]float64{}
	sn := map[string]int{}
	deg5, deg6, deg7 := map[string]int{}, map[string]int{}, map[string]int{}
	for _, e := range edges {
		if _, ok := rank[e.A]; !ok {
			continue
		}
		if _, ok := rank[e.B]; !ok {
			continue
		}
		q := e.Similarity
		ssum[e.A] += q
		ssum[e.B] += q
		sn[e.A]++
		sn[e.B]++
		smax[e.A] = math.Max(smax[e.A], q)
		smax[e.B] = math.Max(smax[e.B], q)
		for _, t := range []string{e.A, e.B} {
			if q >= .5 {
				deg5[t]++
			}
			if q >= .6 {
				deg6[t]++
			}
			if q >= .7 {
				deg7[t]++
			}
		}
	}
	raw := map[string]map[string]float64{}
	for _, t := range tokens {
		n := c.Counts[t]
		g := graphemic.TokenizeGraphemes(t)
		u := map[string]bool{}
		for _, x := range g {
			u[x] = true
		}
		ph := map[string]int{}
		for _, x := range pos[t] {
			ph[strings.TrimRight(strings.TrimRight(fmtFloat(x), "0"), ".")]++
		}
		lp, rp := left[t], right[t]
		le, re := entropy(lp), entropy(rp)
		lmax, l3, l5 := masses(lp)
		rmax, r3, r5 := masses(rp)
		sp := 0.
		if len(ph) > 1 {
			sp = 1 - entropy(ph)/math.Log2(float64(len(ph)))
		}
		sm := 0.
		if sn[t] > 0 {
			sm = ssum[t] / float64(sn[t])
		}
		raw[t] = map[string]float64{
			"global_count": float64(n), "log_global_count": math.Log1p(float64(n)), "global_frequency_probability": float64(n) / float64(len(c.Tokens)), "frequency_rank": float64(rank[t]), "normalized_frequency_rank": float64(rank[t]) / float64(len(tokens)),
			"grapheme_length": float64(len(g)), "unique_grapheme_count": float64(len(u)), "line_start_probability": float64(starts[t]) / float64(n), "line_end_probability": float64(ends[t]) / float64(n), "mean_position": mean(pos[t]), "median_position": quantile(pos[t], .5), "normalized_mean_position": mean(npos[t]), "positional_entropy": entropy(ph), "positional_specialization": sp,
			"predecessor_entropy": le, "successor_entropy": re, "effective_predecessor_count": math.Pow(2, le), "effective_successor_count": math.Pow(2, re), "unique_predecessor_count": float64(len(lp)), "unique_successor_count": float64(len(rp)), "max_predecessor_probability": lmax, "max_successor_probability": rmax, "top3_predecessor_mass": l3, "top3_successor_mass": r3, "top5_predecessor_mass": l5, "top5_successor_mass": r5,
			"mean_structural_similarity_to_all": sm, "max_structural_similarity": smax[t], "structural_degree_above_0_5": float64(deg5[t]), "structural_degree_above_0_6": float64(deg6[t]), "structural_degree_above_0_7": float64(deg7[t])}
	}
	stats := map[string]map[string]float64{}
	names := allPropertyNames()
	for _, name := range names {
		v := make([]float64, 0, len(tokens))
		for _, t := range tokens {
			x := raw[t][name]
			if name == "global_count" || strings.HasPrefix(name, "structural_degree") || strings.HasPrefix(name, "unique_") || strings.HasPrefix(name, "effective_") {
				x = math.Log1p(x)
			}
			v = append(v, x)
		}
		m := mean(v)
		sd := 0.
		for _, x := range v {
			sd += (x - m) * (x - m)
		}
		sd = math.Sqrt(sd / float64(len(v)))
		if sd == 0 {
			sd = 1
		}
		stats[name] = map[string]float64{"mean": m, "stddev": sd}
	}
	out := map[string]TokenProperties{}
	for _, t := range tokens {
		p := TokenProperties{Token: t, Count: c.Counts[t], Properties: map[string]PropertyValue{}}
		for _, name := range names {
			x := raw[t][name]
			zbase := x
			if name == "global_count" || strings.HasPrefix(name, "structural_degree") || strings.HasPrefix(name, "unique_") || strings.HasPrefix(name, "effective_") {
				zbase = math.Log1p(x)
			}
			p.Properties[name] = PropertyValue{Raw: x, Normalized: (zbase - stats[name]["mean"]) / stats[name]["stddev"]}
		}
		out[t] = p
	}
	return out, stats
}
func fmtFloat(v float64) string { return strconv.FormatFloat(v, 'f', 0, 64) }

func aggregate(tokens []string, origin string, d, maxD int, props map[string]TokenProperties, names []string) (map[string][]float64, map[string][]float64, int, int) {
	norm, raw := map[string][]float64{}, map[string][]float64{}
	obs, excluded := 0, 0
	for i, t := range tokens {
		if t != origin || i+d >= len(tokens) {
			continue
		}
		x := tokens[i+d]
		p, ok := props[x]
		if !ok {
			excluded++
			continue
		}
		obs++
		for _, n := range names {
			norm[n] = append(norm[n], p.Properties[n].Normalized)
			raw[n] = append(raw[n], p.Properties[n].Raw)
		}
	}
	return norm, raw, obs, excluded
}

type aggregateData struct {
	norm, raw     map[string][]float64
	obs, excluded int
}
type trajectoryCache map[string][]aggregateData

func buildTrajectoryCache(tokens []string, maxD int, props map[string]TokenProperties) trajectoryCache {
	names := allPropertyNames()
	out := trajectoryCache{}
	for origin := range props {
		x := make([]aggregateData, maxD)
		for i := range x {
			x[i].norm = map[string][]float64{}
			x[i].raw = map[string][]float64{}
		}
		out[origin] = x
	}
	for i, origin := range tokens {
		x, ok := out[origin]
		if !ok {
			continue
		}
		for d := 1; d <= maxD && i+d < len(tokens); d++ {
			p, eligible := props[tokens[i+d]]
			if !eligible {
				x[d-1].excluded++
				continue
			}
			x[d-1].obs++
			for _, n := range names {
				x[d-1].norm[n] = append(x[d-1].norm[n], p.Properties[n].Normalized)
				x[d-1].raw[n] = append(x[d-1].raw[n], p.Properties[n].Raw)
			}
		}
	}
	return out
}
func vector(m map[string][]float64, names []string) []float64 {
	x := make([]float64, len(names))
	for i, n := range names {
		x[i] = mean(m[n])
	}
	return x
}
func similarity(a, b []float64) (cos, euclid, manhattan, corr float64) {
	var dot, aa, bb float64
	for i := range a {
		dot += a[i] * b[i]
		aa += a[i] * a[i]
		bb += b[i] * b[i]
		d := a[i] - b[i]
		euclid += d * d
		manhattan += math.Abs(d)
	}
	if aa > 0 && bb > 0 {
		cos = dot / math.Sqrt(aa*bb)
	}
	euclid = math.Sqrt(euclid)
	corr = pearson(a, b)
	return
}
func summarize(a, b, ra, rb map[string][]float64, names []string) map[string]PropertySummary {
	out := map[string]PropertySummary{}
	for _, n := range names {
		ma, mb := mean(a[n]), mean(b[n])
		sd := func(x []float64) float64 {
			m := mean(x)
			s := 0.
			for _, v := range x {
				s += (v - m) * (v - m)
			}
			if len(x) > 0 {
				return math.Sqrt(s / float64(len(x)))
			}
			return 0
		}
		out[n] = PropertySummary{MeanA: ma, MeanB: mb, Delta: ma - mb, RawMeanA: mean(ra[n]), RawMeanB: mean(rb[n]), MedianA: quantile(a[n], .5), MedianB: quantile(b[n], .5), StddevA: sd(a[n]), StddevB: sd(b[n]), P25A: quantile(a[n], .25), P25B: quantile(b[n], .25), P75A: quantile(a[n], .75), P75B: quantile(b[n], .75)}
	}
	return out
}
func profile(tokens []string, p pair, maxD int, props map[string]TokenProperties, names []string) []DistanceProfile {
	out := make([]DistanceProfile, 0, maxD)
	for d := 1; d <= maxD; d++ {
		a, ra, oa, xa := aggregate(tokens, p.A, d, maxD, props, names)
		b, rb, ob, xb := aggregate(tokens, p.B, d, maxD, props, names)
		cos, e, m, c := similarity(vector(a, names), vector(b, names))
		out = append(out, DistanceProfile{Distance: d, ObservationsA: oa, ObservationsB: ob, ExcludedLowFrequencyA: xa, ExcludedLowFrequencyB: xb, CosineSimilarity: cos, EuclideanDistance: e, ManhattanDistance: m, Correlation: c, Properties: summarize(a, b, ra, rb, names)})
	}
	return out
}
func profileCached(cache trajectoryCache, p pair, maxD int, names []string) []DistanceProfile {
	out := make([]DistanceProfile, 0, maxD)
	aa, bb := cache[p.A], cache[p.B]
	for d := 1; d <= maxD; d++ {
		var a, b aggregateData
		if d <= len(aa) {
			a = aa[d-1]
		} else {
			a = aggregateData{norm: map[string][]float64{}, raw: map[string][]float64{}}
		}
		if d <= len(bb) {
			b = bb[d-1]
		} else {
			b = aggregateData{norm: map[string][]float64{}, raw: map[string][]float64{}}
		}
		cos, e, m, c := similarity(vector(a.norm, names), vector(b.norm, names))
		out = append(out, DistanceProfile{Distance: d, ObservationsA: a.obs, ObservationsB: b.obs, ExcludedLowFrequencyA: a.excluded, ExcludedLowFrequencyB: b.excluded, CosineSimilarity: cos, EuclideanDistance: e, ManhattanDistance: m, Correlation: c, Properties: summarize(a.norm, b.norm, a.raw, b.raw, names)})
	}
	return out
}
func cosinesCached(cache trajectoryCache, p pair, maxD int, names []string) []float64 {
	aa, bb := cache[p.A], cache[p.B]
	out := make([]float64, 0, maxD)
	for d := 0; d < maxD; d++ {
		if d >= len(aa) || d >= len(bb) || aa[d].obs == 0 || bb[d].obs == 0 {
			continue
		}
		cos, _, _, _ := similarity(vector(aa[d].norm, names), vector(bb[d].norm, names))
		out = append(out, cos)
	}
	return out
}
func scoreCached(cache trajectoryCache, p pair, maxD int, names []string) float64 {
	return mean(cosinesCached(cache, p, maxD, names))
}
func rangeCos(x []DistanceProfile, lo, hi int) float64 {
	var v []float64
	for _, d := range x {
		if d.Distance >= lo && d.Distance <= hi && d.ObservationsA > 0 && d.ObservationsB > 0 {
			v = append(v, d.CosineSimilarity)
		}
	}
	return mean(v)
}
func modeNames(mode string) []string {
	if mode == "all-properties" {
		return allPropertyNames()
	}
	if strings.HasPrefix(mode, "all-minus-") {
		drop := strings.TrimPrefix(mode, "all-minus-")
		var x []string
		for g, n := range propertyGroups {
			if g != drop {
				x = append(x, n...)
			}
		}
		sort.Strings(x)
		return x
	}
	key := strings.TrimSuffix(mode, "-only")
	key = strings.TrimSuffix(key, "-form")
	key = strings.TrimSuffix(key, "-complexity")
	key = strings.TrimSuffix(key, "-centrality")
	return propertyGroups[key]
}
func shuffleCorpus(c corpus, mode string, r *rand.Rand) []string {
	if mode == "global" {
		x := append([]string(nil), c.Tokens...)
		r.Shuffle(len(x), func(i, j int) { x[i], x[j] = x[j], x[i] })
		return x
	}
	var out []string
	for _, line := range c.Lines {
		x := append([]string(nil), line...)
		r.Shuffle(len(x), func(i, j int) { x[i], x[j] = x[j], x[i] })
		out = append(out, x...)
	}
	return out
}
