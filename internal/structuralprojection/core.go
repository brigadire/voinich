package structuralprojection

import (
	"math"
	"math/rand"
	"sort"
)

func normalize(m map[string]float64) map[string]float64 {
	// Sum in sorted key order: map iteration order is randomized
	// independently per range statement execution, so summing in `range m`
	// order made this float64 accumulation nondeterministic across
	// otherwise byte-identical calls (see determinism_test.go).
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	s := 0.0
	for _, k := range keys {
		if v := m[k]; v > 0 {
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
	// Visit the observed tokens (the outer accumulation source) in sorted
	// order: several distinct x's can route weight to the same destination
	// y, accumulating into the shared out[y], so the order distinct x's are
	// visited in must not depend on Go's randomized map iteration (see
	// determinism_test.go). The inner row loop needs no such fix: within one
	// x's row, every destination y is visited exactly once, so its order
	// cannot change which value lands in any single out[y].
	keys := make([]string, 0, len(counts))
	for x := range counts {
		keys = append(keys, x)
	}
	sort.Strings(keys)
	for _, x := range keys {
		n := counts[x]
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
	keySet := map[string]bool{}
	for k := range a {
		keySet[k] = true
	}
	for k := range b {
		keySet[k] = true
	}
	// Accumulate in sorted key order: overlap and div are single running
	// sums fed by every key in one call, so summing in `range keySet` order
	// made them nondeterministic across otherwise byte-identical calls (see
	// determinism_test.go). inter is an exact integer count and needs no
	// such fix.
	keys := make([]string, 0, len(keySet))
	for k := range keySet {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	inter, div := 0, 0.0
	for _, k := range keys {
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

// frequencyBins groups a fixed token universe into log2-frequency bins, in
// the same sorted-token order RandomizeProjection and GenericSmoothing each
// used to build internally. Both derive this grouping from the identical
// (sorted corpus vocabulary, corpus counts) pair, and the grouping does not
// depend on which trial or which Projection (full vs future/past-ablated)
// is being processed — the corpus vocabulary and its counts are fixed for
// an entire analyze() call — so it only needs to be built once and reused
// across every one of the 200-trial random/smoothing control loop's calls,
// instead of rebuilt on all ~800 of them (200 trials x 2 functions x 2
// projections). See PERFORMANCE_REFACTOR_REPORT.md for the profiler
// evidence and the equivalence argument.
type frequencyBins struct {
	sortedTokens []string
	binKeys      []int
	bins         map[int][]string
	tokenBin     map[string]int
	maxBin       int
}

func buildFrequencyBins(tokens []string, counts map[string]int) frequencyBins {
	sorted := make([]string, len(tokens))
	copy(sorted, tokens)
	sort.Strings(sorted)
	bins := map[int][]string{}
	tokenBin := map[string]int{}
	maxBin := 0
	for _, t := range sorted {
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
	binKeys := make([]int, 0, len(bins))
	for b := range bins {
		binKeys = append(binKeys, b)
	}
	sort.Ints(binKeys)
	return frequencyBins{sortedTokens: sorted, binKeys: binKeys, bins: bins, tokenBin: tokenBin, maxBin: maxBin}
}

func RandomizeProjection(p Projection, fb frequencyBins, seed int64) Projection {
	r := rand.New(rand.NewSource(seed))
	// Permute destinations within log2-frequency bins, preserving every row's
	// degree and weights while approximately preserving neighbour frequency.
	perm := map[string]string{}
	// Visit bins in ascending index order, not map iteration order: a single
	// shared r is consumed across every bin's shuffle, so which bin consumes
	// which slice of the random stream must not depend on Go's randomized
	// map iteration (see determinism_test.go).
	for _, b := range fb.binKeys {
		xs := fb.bins[b]
		ys := append([]string(nil), xs...)
		r.Shuffle(len(ys), func(i, j int) { ys[i], ys[j] = ys[j], ys[i] })
		for i, x := range xs {
			perm[x] = ys[i]
		}
	}
	out := Projection{}
	for _, src := range fb.sortedTokens {
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
					b := fb.tokenBin[dst]
					for _, candidate := range fb.bins[b] {
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

// GenericSmoothing's outward bin sweep (see the delta loop below) always
// visits every one of the fb.maxBin+1 populated bins exactly once,
// regardless of the starting bin b — so pool always ends up holding a
// full-vocabulary-sized sequence (len(fb.sortedTokens) entries) for every
// one of the V tokens processed: O(V) work x V tokens = O(V²) total. This
// is required, not accidental: r is a single stream shared across every
// bin of every token, consumed strictly in the sweep's visitation order,
// so skipping or truncating any bin's shuffle (even one whose candidates
// are never read by the current token, because degree was already
// satisfied by an earlier bin) would consume fewer random draws and shift
// which numbers every subsequent token in fb.sortedTokens receives,
// changing the output for a given seed. What is NOT required is
// reallocating pool's backing array (and a separate per-bin copy) from
// scratch on every one of the V iterations — the buffers below are
// preallocated once and reused, with pool itself doubling as the
// per-bin shuffle scratch (appending a bin's raw content and shuffling the
// newly-appended segment in place is equivalent to the former
// copy-then-shuffle-then-append: Fisher-Yates swaps are positional and
// don't care whether the slice being shuffled is a fresh copy or a live
// segment of a larger backing array), removing the O(V²) allocation this
// produced without changing the shuffle work or RNG consumption at all.
func GenericSmoothing(fb frequencyBins, p Projection, seed int64) Projection {
	r := rand.New(rand.NewSource(seed))
	out := make(Projection, len(fb.sortedTokens))
	pool := make([]string, 0, len(fb.sortedTokens))
	m := map[string]float64{}
	for _, src := range fb.sortedTokens {
		degree := len(p[src]) - 1
		if degree < 0 {
			degree = 0
		}
		b := fb.tokenBin[src]
		pool = pool[:0]
		appendBin := func(bin int) {
			start := len(pool)
			pool = append(pool, fb.bins[bin]...)
			segment := pool[start:]
			r.Shuffle(len(segment), func(i, j int) { segment[i], segment[j] = segment[j], segment[i] })
		}
		for delta := 0; delta <= fb.maxBin+1; delta++ {
			if b-delta >= 0 {
				appendBin(b - delta)
			}
			if delta > 0 && b+delta <= fb.maxBin {
				appendBin(b + delta)
			}
		}
		clear(m)
		m[src] = 1
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
