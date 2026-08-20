package vocabularygrowth

import (
	"fmt"
	"math"
	"math/rand"
	"sort"
)

func Checkpoints(n int) []int {
	if n <= 0 {
		return nil
	}
	base := []int{100, 200, 500}
	for x := 1000; x <= n; {
		base = append(base, x)
		if x > n/2 {
			break
		}
		x *= 2
	}
	base = append(base, n)
	sort.Ints(base)
	out := base[:0]
	for _, x := range base {
		if x > 0 && x <= n && (len(out) == 0 || out[len(out)-1] != x) {
			out = append(out, x)
		}
	}
	return out
}

func Analyze(tokens []string, p Parameters) (Result, error) {
	if p.NullPermutations < 0 {
		return Result{}, fmt.Errorf("null permutations must be non-negative")
	}
	if len(p.WindowSizes) == 0 {
		p.WindowSizes = []int{500, 1000, 2000}
	}
	if len(p.SegmentCounts) == 0 {
		p.SegmentCounts = []int{4, 8}
	}
	if p.Seed == 0 {
		p.Seed = 1
	}
	if len(p.Checkpoints) == 0 {
		p.Checkpoints = Checkpoints(len(tokens))
	} else {
		p.Checkpoints = canonicalCheckpoints(p.Checkpoints, len(tokens))
	}
	if len(tokens) == 0 {
		return Result{Parameters: p}, nil
	}
	r := Result{TotalTokens: len(tokens), Checkpoints: p.Checkpoints, Parameters: p}
	r.Growth = trajectory(tokens, p.Checkpoints)
	r.Final = r.Growth[len(r.Growth)-1]
	r.Fit = fit(r.Growth, p.FitMinN, p.FitMaxN)
	r.Windows = windows(tokens, p.WindowSizes)
	r.Segments = segments(tokens, p.SegmentCounts)
	if p.NullPermutations > 0 {
		r.Null = nulls(tokens, r.Growth, p)
	}
	return r, nil
}

func canonicalCheckpoints(xs []int, n int) []int {
	ys := append([]int(nil), xs...)
	sort.Ints(ys)
	out := []int{}
	for _, x := range ys {
		if x > 0 && x <= n && (len(out) == 0 || out[len(out)-1] != x) {
			out = append(out, x)
		}
	}
	if n > 0 && (len(out) == 0 || out[len(out)-1] != n) {
		out = append(out, n)
	}
	return out
}
func trajectory(tokens []string, cps []int) []Point {
	seen := map[string]int{}
	freq := map[int]int{}
	out := make([]Point, 0, len(cps))
	ci := 0
	for i, t := range tokens {
		old := seen[t]
		if old > 0 {
			freq[old]--
		}
		seen[t]++
		freq[seen[t]]++
		if ci < len(cps) && i+1 == cps[ci] {
			pt := Point{N: i + 1, Vocabulary: len(seen), Hapax: freq[1], Dis: freq[2], Tri: freq[3], TTR: float64(len(seen)) / float64(i+1)}
			if len(out) > 0 {
				pt.BetaEffective = effective(out[len(out)-1], pt)
			}
			out = append(out, pt)
			ci++
		}
	}
	return out
}
func effective(a, b Point) float64 {
	if a.N <= 0 || b.N <= a.N || a.Vocabulary <= 0 || b.Vocabulary <= 0 {
		return 0
	}
	return (math.Log(float64(b.Vocabulary)) - math.Log(float64(a.Vocabulary))) / (math.Log(float64(b.N)) - math.Log(float64(a.N)))
}
func windows(tokens []string, sizes []int) []WindowPoint {
	sort.Ints(sizes)
	out := []WindowPoint{}
	for _, size := range sizes {
		if size <= 0 {
			continue
		}
		seenBefore := map[string]bool{}
		for start := 0; start < len(tokens); start += size {
			end := start + size
			if end > len(tokens) {
				end = len(tokens)
			}
			newTypes := 0
			windowSeen := map[string]bool{}
			for _, t := range tokens[start:end] {
				if !seenBefore[t] && !windowSeen[t] {
					newTypes++
					windowSeen[t] = true
				}
				seenBefore[t] = true
			}
			out = append(out, WindowPoint{Start: start, End: end, Tokens: end - start, NewTypes: newTypes, NewTypeRate: float64(newTypes) / float64(end-start)})
			if end == len(tokens) {
				break
			}
		}
	}
	return out
}
func frequencies(tokens []string) map[string]int {
	m := map[string]int{}
	for _, t := range tokens {
		m[t]++
	}
	return m
}
func fit(points []Point, minN, maxN int) Fit {
	x, y := []float64{}, []float64{}
	for _, p := range points {
		if p.Vocabulary <= 0 || (minN > 0 && p.N < minN) || (maxN > 0 && p.N > maxN) {
			continue
		}
		x = append(x, math.Log(float64(p.N)))
		y = append(y, math.Log(float64(p.Vocabulary)))
	}
	f := Fit{Points: len(x)}
	if len(x) < 2 {
		return f
	}
	mx, my := mean(x), mean(y)
	sxx, sxy := 0., 0.
	for i := range x {
		sxx += (x[i] - mx) * (x[i] - mx)
		sxy += (x[i] - mx) * (y[i] - my)
	}
	if sxx == 0 {
		return f
	}
	b := sxy / sxx
	a := my - b*mx
	f.Beta, f.K = b, math.Exp(a)
	f.NMin, f.NMax = int(math.Exp(x[0])), int(math.Exp(x[len(x)-1]))
	sst, sse, maxResidual := 0., 0., 0.
	for i := range x {
		residual := y[i] - (a + b*x[i])
		sse += residual * residual
		if math.Abs(residual) > maxResidual {
			maxResidual = math.Abs(residual)
		}
		sst += (y[i] - my) * (y[i] - my)
	}
	f.SSE, f.MaxAbsResidual = sse, maxResidual
	if sst > 0 {
		f.R2 = 1 - sse/sst
	}
	return f
}
func mean(x []float64) float64 {
	s := 0.
	for _, v := range x {
		s += v
	}
	return s / float64(len(x))
}
func nulls(tokens []string, observed []Point, p Parameters) []NullPoint {
	vals := make([][]float64, len(observed))
	for i := range vals {
		vals[i] = []float64{}
	}
	for rep := 0; rep < p.NullPermutations; rep++ {
		cp := append([]string(nil), tokens...)
		rng := rand.New(rand.NewSource(deriveSeed(p.Seed, rep)))
		rng.Shuffle(len(cp), func(i, j int) { cp[i], cp[j] = cp[j], cp[i] })
		pts := trajectory(cp, p.Checkpoints)
		for i, pt := range pts {
			vals[i] = append(vals[i], float64(pt.Vocabulary))
		}
	}
	out := make([]NullPoint, len(observed))
	for i, o := range observed {
		xs := vals[i]
		m := mean(xs)
		sd := 0.
		ge := 0
		for _, v := range xs {
			sd += (v - m) * (v - m)
			if v >= float64(o.Vocabulary) {
				ge++
			}
		}
		if len(xs) > 0 {
			sd = math.Sqrt(sd / float64(len(xs)))
		}
		effect := 0.
		if sd > 0 {
			effect = (float64(o.Vocabulary) - m) / sd
		}
		out[i] = NullPoint{N: o.N, Observed: float64(o.Vocabulary), NullMean: m, NullSD: sd, Effect: effect, EmpiricalP: float64(ge+1) / float64(len(xs)+1)}
	}
	return out
}
func deriveSeed(base int64, rep int) int64 {
	z := uint64(base) + 0x9e3779b97f4a7c15 + uint64(rep)*0xbf58476d1ce4e5b9
	z = (z ^ (z >> 30)) * 0xbf58476d1ce4e5b9
	z = (z ^ (z >> 27)) * 0x94d049bb133111eb
	return int64(z ^ (z >> 31))
}
func segments(tokens []string, counts []int) []SegmentPoint {
	out := []SegmentPoint{}
	for _, num := range counts {
		if num <= 0 || len(tokens) < num {
			continue
		}
		for s := 0; s < num; s++ {
			start := s * len(tokens) / num
			end := (s + 1) * len(tokens) / num
			if end <= start {
				continue
			}
			part := tokens[start:end]
			cps := Checkpoints(len(part))
			pts := trajectory(part, cps)
			f := fit(pts, 0, 0)
			rate := 0.
			if len(part) > 0 {
				seen := map[string]bool{}
				for _, t := range part {
					seen[t] = true
				}
				rate = float64(len(seen)) / float64(len(part))
			}
			for _, pt := range pts {
				out = append(out, SegmentPoint{Segments: num, Segment: s, CheckpointN: pt.N, Vocabulary: pt.Vocabulary, HeapsBeta: f.Beta, BetaEffective: pt.BetaEffective, NewTypeRate: rate})
			}
		}
	}
	return out
}

// FinalProfile returns the final frequency-of-frequencies profile without
// rereading the corpus. It is intentionally language-agnostic.
func FinalProfile(tokens []string) map[string]float64 {
	f := frequencies(tokens)
	v1, v2, v3, v5 := 0, 0, 0, 0
	for _, n := range f {
		switch n {
		case 1:
			v1++
		case 2:
			v2++
		case 3:
			v3++
		default:
			if n >= 5 {
				v5++
			}
		}
	}
	n := float64(len(tokens))
	u := float64(len(f))
	m := map[string]float64{"total_tokens": n, "unique_tokens": u, "type_token_ratio": u / n, "hapax": float64(v1), "dis_legomena": float64(v2), "tris_legomena": float64(v3), "v5_plus": float64(v5), "hapax_fraction_of_types": float64(v1) / u, "hapax_fraction_of_tokens": float64(v1) / n, "dis_fraction_of_types": float64(v2) / u, "dis_fraction_of_tokens": 2 * float64(v2) / n}
	if v2 > 0 {
		m["singleton_to_doubleton_ratio"] = float64(v1) / float64(v2)
	} else {
		m["singleton_to_doubleton_ratio"] = math.Inf(1)
	}
	return m
}
