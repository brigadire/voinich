package globalregime

import (
	"bufio"
	"fmt"
	"math"
	"os"
	"sort"
	"strings"
)

func readCorpus(path string) ([]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	var tokens []string
	s := bufio.NewScanner(f)
	s.Buffer(make([]byte, 64<<10), 4<<20)
	for s.Scan() {
		tokens = append(tokens, strings.Fields(s.Text())...)
	}
	if err := s.Err(); err != nil {
		return nil, err
	}
	if len(tokens) == 0 {
		return nil, fmt.Errorf("corpus is empty")
	}
	return tokens, nil
}

func normalize(counts map[string]int, total int) profile {
	p := profile{}
	if total == 0 {
		return p
	}
	for token, n := range counts {
		p[token] = float64(n) / float64(total)
	}
	return p
}

// jsDistance's d is a single running sum fed by every key of the union of
// a and b, so it is accumulated in sorted key order: summing in map
// iteration order made it nondeterministic across otherwise
// byte-identical calls (see determinism_test.go).
func jsDistance(a, b profile) float64 {
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
	d := 0.0
	for _, k := range keys {
		x, y := a[k], b[k]
		m := (x + y) / 2
		if x > 0 {
			d += .5 * x * math.Log2(x/m)
		}
		if y > 0 {
			d += .5 * y * math.Log2(y/m)
		}
	}
	if d < 0 {
		return 0
	}
	if d > 1 {
		return 1
	}
	return d
}

// sortedProfile pairs a profile with its own keys sorted once, so
// jsDistanceSorted can merge-walk two already-sorted key lists instead of
// re-sorting their union on every pairwise call - each profile's own keys
// are sorted exactly once however many times it's compared against
// others. This matters for distanceMatrix/expandLabels, which compare the
// same window/centroid profiles against many others (O(n^2) and O(n*k)
// respectively): profiling conditionalregime's own equivalent hot path
// (task27) showed the sort-per-call approach was over 70% of that CLI's
// total CPU time.
type sortedProfile struct {
	p    profile
	keys []string
}

func sortProfile(p profile) sortedProfile {
	keys := make([]string, 0, len(p))
	for k := range p {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return sortedProfile{p: p, keys: keys}
}

// jsDistanceSorted is jsDistance, but merge-walking a's and b's
// already-sorted key lists instead of re-sorting their union - the merge
// visits the sorted union in the same order sorting it directly would, so
// this is bit-identical to jsDistance(a.p, b.p).
func jsDistanceSorted(a, b sortedProfile) float64 {
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
	if d < 0 {
		return 0
	}
	if d > 1 {
		return 1
	}
	return d
}

// sortedProfileKeys returns p's keys in sorted order, for accumulating a
// single running sum fed by every key of a map deterministically (map
// iteration order is randomized per range execution — see
// determinism_test.go).
func sortedProfileKeys(p profile) []string {
	keys := make([]string, 0, len(p))
	for k := range p {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
func overlap(a, b profile) float64 {
	s := 0.0
	for _, k := range sortedProfileKeys(a) {
		x := a[k]
		if b[k] < x {
			x = b[k]
		}
		s += x
	}
	return s
}
func cosine(a, b profile) float64 {
	dot, aa, bb := 0., 0., 0.
	for _, k := range sortedProfileKeys(a) {
		x := a[k]
		dot += x * b[k]
		aa += x * x
	}
	for _, k := range sortedProfileKeys(b) {
		y := b[k]
		bb += y * y
	}
	if aa == 0 || bb == 0 {
		return 0
	}
	return dot / math.Sqrt(aa*bb)
}

func slidingWindows(tokens []string, size, step int) []Window {
	if step <= 0 {
		step = max(1, size/10)
	}
	if size > len(tokens) {
		return nil
	}
	counts := map[string]int{}
	for _, t := range tokens[:size] {
		counts[t]++
	}
	var out []Window
	var prev profile
	for start := 0; start+size <= len(tokens); start += step {
		if start > 0 {
			if step <= size {
				oldStart := start - step
				oldEnd := start + size - step
				for _, t := range tokens[oldStart:start] {
					counts[t]--
					if counts[t] == 0 {
						delete(counts, t)
					}
				}
				for _, t := range tokens[oldEnd : start+size] {
					counts[t]++
				}
			} else {
				counts = map[string]int{}
				for _, t := range tokens[start : start+size] {
					counts[t]++
				}
			}
		}
		p := normalize(counts, size)
		w := Window{WindowSize: size, Index: len(out), Start: start, End: start + size, Center: start + size/2, Step: step, distribution: p}
		if prev != nil {
			w.JSDistance = jsDistance(prev, p)
			w.WeightedOverlap = overlap(prev, p)
			w.Cosine = cosine(prev, p)
		} else {
			w.WeightedOverlap = 1
			w.Cosine = 1
		}
		out = append(out, w)
		prev = p
	}
	classifyVariation(out)
	return out
}

func quantile(x []float64, q float64) float64 {
	if len(x) == 0 {
		return 0
	}
	y := append([]float64(nil), x...)
	sort.Float64s(y)
	pos := q * float64(len(y)-1)
	lo := int(pos)
	hi := min(lo+1, len(y)-1)
	f := pos - float64(lo)
	return y[lo]*(1-f) + y[hi]*f
}

func classifyVariation(w []Window) {
	if len(w) < 2 {
		return
	}
	v := make([]float64, 0, len(w)-1)
	for i := 1; i < len(w); i++ {
		v = append(v, w[i].JSDistance)
	}
	lo, hi := quantile(v, .25), quantile(v, .75)
	for i := 1; i < len(w); i++ {
		switch {
		case w[i].JSDistance <= lo:
			w[i].Variation = "low"
		case w[i].JSDistance >= hi:
			w[i].Variation = "high"
		default:
			w[i].Variation = "moderate"
		}
		if i < len(w)-1 && w[i].JSDistance > w[i-1].JSDistance && w[i].JSDistance >= w[i+1].JSDistance {
			w[i].LocalPeak = true
		}
	}
	// A broad transition needs elevated variation in a neighbourhood, not one spike.
	for i := 2; i < len(w)-2; i++ {
		n := 0
		for j := i - 2; j <= i+2; j++ {
			if w[j].JSDistance >= hi {
				n++
			}
		}
		w[i].BroadTransition = n >= 3
	}
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
	if len(x) == 0 {
		return 0
	}
	m := mean(x)
	s := 0.
	for _, v := range x {
		d := v - m
		s += d * d
	}
	return math.Sqrt(s / float64(len(x)))
}
