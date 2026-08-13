package structuralprojection

import (
	"math"
	"math/rand"
	"sort"
)

func normalize(m map[string]float64) map[string]float64 {
	s := 0.0
	for _, v := range m {
		if v > 0 {
			s += v
		}
	}
	out := map[string]float64{}
	if s == 0 {
		return out
	}
	for k, v := range m {
		if v > 0 {
			out[k] = v / s
		}
	}
	return out
}

// BuildProjection constructs row-stochastic soft neighbourhoods. Self-weight is
// exactly one and is never subject to thresholds. mode is full, future-ablated
// (position+left), or past-ablated (position+right).
func BuildProjection(tokens []string, edges []Edge, minSim, minRel float64, k int, mode string) Projection {
	type candidate struct {
		token            string
		weight, sim, rel float64
	}
	rows := map[string][]candidate{}
	for _, e := range edges {
		sim, rel := e.Similarity, e.Reliability
		switch mode {
		case "future-ablated":
			den := e.PositionReliability + e.LeftReliability
			if den > 0 {
				sim = (e.Position*e.PositionReliability + e.Left*e.LeftReliability) / den
				rel = den / 2
			} else {
				sim, rel = 0, 0
			}
		case "past-ablated":
			den := e.PositionReliability + e.RightReliability
			if den > 0 {
				sim = (e.Position*e.PositionReliability + e.Right*e.RightReliability) / den
				rel = den / 2
			} else {
				sim, rel = 0, 0
			}
		}
		if sim >= minSim && rel >= minRel {
			w := sim * rel
			rows[e.A] = append(rows[e.A], candidate{e.B, w, sim, rel})
			rows[e.B] = append(rows[e.B], candidate{e.A, w, sim, rel})
		}
	}
	out := Projection{}
	for _, t := range tokens {
		x := rows[t]
		sort.Slice(x, func(i, j int) bool {
			if x[i].weight == x[j].weight {
				return x[i].token < x[j].token
			}
			return x[i].weight > x[j].weight
		})
		if k > 0 && len(x) > k {
			x = x[:k]
		}
		m := map[string]float64{t: 1}
		for _, c := range x {
			m[c.token] = c.weight
		}
		out[t] = normalize(m)
	}
	return out
}

func ProjectDistribution(counts map[string]int, p Projection) map[string]float64 {
	total := 0
	for _, n := range counts {
		total += n
	}
	out := map[string]float64{}
	if total == 0 {
		return out
	}
	for x, n := range counts {
		row := p[x]
		if len(row) == 0 {
			row = map[string]float64{x: 1}
		}
		mass := float64(n) / float64(total)
		for y, w := range row {
			out[y] += mass * w
		}
	}
	return out
}

func metricsFloat(a, b map[string]float64) (js, overlap, jaccard float64) {
	if len(a) == 0 || len(b) == 0 {
		return
	}
	keys := map[string]bool{}
	for k := range a {
		keys[k] = true
	}
	for k := range b {
		keys[k] = true
	}
	inter, div := 0, 0.0
	for k := range keys {
		pa, pb := a[k], b[k]
		if pa > 0 && pb > 0 {
			inter++
		}
		overlap += math.Min(pa, pb)
		m := (pa + pb) / 2
		if pa > 0 {
			div += .5 * pa * math.Log(pa/m)
		}
		if pb > 0 {
			div += .5 * pb * math.Log(pb/m)
		}
	}
	js = 1 - div/math.Ln2
	if js < 0 {
		js = 0
	}
	if js > 1 {
		js = 1
	}
	jaccard = float64(inter) / float64(len(keys))
	return
}
func countsFloat(a map[string]int) map[string]float64 {
	s := 0
	for _, n := range a {
		s += n
	}
	o := map[string]float64{}
	if s > 0 {
		for k, n := range a {
			o[k] = float64(n) / float64(s)
		}
	}
	return o
}
func total(a map[string]int) int {
	n := 0
	for _, v := range a {
		n += v
	}
	return n
}

func RandomizeProjection(p Projection, counts map[string]int, seed int64) Projection {
	r := rand.New(rand.NewSource(seed))
	tokens := make([]string, 0, len(p))
	for t := range p {
		tokens = append(tokens, t)
	}
	sort.Strings(tokens)
	// Permute destinations within log2-frequency bins, preserving every row's
	// degree and weights while approximately preserving neighbour frequency.
	bins := map[int][]string{}
	for _, t := range tokens {
		b := 0
		if counts[t] > 0 {
			b = int(math.Log2(float64(counts[t])))
		}
		bins[b] = append(bins[b], t)
	}
	perm := map[string]string{}
	for _, xs := range bins {
		ys := append([]string(nil), xs...)
		r.Shuffle(len(ys), func(i, j int) { ys[i], ys[j] = ys[j], ys[i] })
		for i, x := range xs {
			perm[x] = ys[i]
		}
	}
	out := Projection{}
	for _, src := range tokens {
		m := map[string]float64{}
		for dst, w := range p[src] {
			if dst == src {
				m[src] += w
			} else {
				target := perm[dst]
				if target == src {
					// perm[src] cannot be the image of another destination in
					// this row because the self edge is handled separately.
					target = perm[src]
				}
				if target == src || m[target] > 0 {
					b := 0
					if counts[dst] > 0 {
						b = int(math.Log2(float64(counts[dst])))
					}
					for _, candidate := range bins[b] {
						if candidate != src && m[candidate] == 0 {
							target = candidate
							break
						}
					}
				}
				if target == src { // only possible for a one-token frequency bin
					target = dst
				}
				m[target] += w
			}
		}
		out[src] = normalize(m)
	}
	return out
}

func GenericSmoothing(tokens []string, counts map[string]int, p Projection, seed int64) Projection {
	r := rand.New(rand.NewSource(seed))
	bins, tokenBin, maxBin := map[int][]string{}, map[string]int{}, 0
	for _, t := range tokens {
		b := 0
		if counts[t] > 0 {
			b = int(math.Log2(float64(counts[t])))
		}
		bins[b] = append(bins[b], t)
		tokenBin[t] = b
		if b > maxBin {
			maxBin = b
		}
	}
	out := Projection{}
	for _, src := range tokens {
		degree := len(p[src]) - 1
		if degree < 0 {
			degree = 0
		}
		b := tokenBin[src]
		m := map[string]float64{src: 1}
		pool := make([]string, 0, len(tokens))
		appendBin := func(bin int) {
			group := append([]string(nil), bins[bin]...)
			r.Shuffle(len(group), func(i, j int) { group[i], group[j] = group[j], group[i] })
			pool = append(pool, group...)
		}
		for delta := 0; delta <= maxBin+1; delta++ {
			if b-delta >= 0 {
				appendBin(b - delta)
			}
			if delta > 0 && b+delta <= maxBin {
				appendBin(b + delta)
			}
		}
		added := 0
		for _, candidate := range pool {
			if candidate != src {
				m[candidate] = 1
				added++
				if added == degree {
					break
				}
			}
		}
		out[src] = normalize(m)
	}
	return out
}

func gain(a, b map[string]int, p Projection) float64 {
	tj, _, _ := metricsFloat(countsFloat(a), countsFloat(b))
	pj, _, _ := metricsFloat(ProjectDistribution(a, p), ProjectDistribution(b, p))
	return pj - tj
}
