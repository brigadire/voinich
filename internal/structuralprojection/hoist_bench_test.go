package structuralprojection

import (
	"fmt"
	"math"
	"math/rand"
	"reflect"
	"sort"
	"testing"
)

// referenceRandomizeProjection is the pre-hoist implementation: it builds
// the log2-frequency-bin grouping from (p's keys, counts) fresh on every
// call, instead of accepting a precomputed frequencyBins. It already
// visits bins in sorted-key order (the separate determinism fix — see
// determinism_test.go), so this isolates the hoist itself from that fix.
func referenceRandomizeProjection(p Projection, counts map[string]int, seed int64) Projection {
	r := rand.New(rand.NewSource(seed))
	tokens := make([]string, 0, len(p))
	for t := range p {
		tokens = append(tokens, t)
	}
	sort.Strings(tokens)
	bins := map[int][]string{}
	for _, t := range tokens {
		b := 0
		if counts[t] > 0 {
			b = int(math.Log2(float64(counts[t])))
		}
		bins[b] = append(bins[b], t)
	}
	binKeys := make([]int, 0, len(bins))
	for b := range bins {
		binKeys = append(binKeys, b)
	}
	sort.Ints(binKeys)
	perm := map[string]string{}
	for _, b := range binKeys {
		xs := bins[b]
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
				if target == src {
					target = dst
				}
				m[target] += w
			}
		}
		out[src] = normalize(m)
	}
	return out
}

// referenceGenericSmoothing is the pre-hoist implementation: same
// treatment as referenceRandomizeProjection above.
func referenceGenericSmoothing(tokens []string, counts map[string]int, p Projection, seed int64) Projection {
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

func randomProjectionFixture(seed int64, n int) (Projection, map[string]int, []string) {
	r := rand.New(rand.NewSource(seed))
	tokens := make([]string, n)
	counts := map[string]int{}
	p := Projection{}
	for i := 0; i < n; i++ {
		tokens[i] = fmt.Sprintf("tok%04d", i)
	}
	for i, t := range tokens {
		counts[t] = 1 + r.Intn(2000)
		row := map[string]float64{t: 1}
		degree := 1 + r.Intn(8)
		for k := 0; k < degree; k++ {
			row[tokens[(i+k+1)%n]] = 0.1 + 0.05*float64(k)
		}
		p[t] = row
	}
	return p, counts, tokens
}

// TestRandomizeProjectionHoistedBinsMatchesReference proves the
// frequencyBins hoist (analyze.go computes it once per analyze() call
// instead of RandomizeProjection rebuilding it from scratch on every one of
// the 200-trial loop's ~400 calls) produces byte-identical output to the
// pre-hoist implementation, across several corpus shapes and seeds.
func TestRandomizeProjectionHoistedBinsMatchesReference(t *testing.T) {
	shapes := []int{5, 40, 137}
	for _, n := range shapes {
		p, counts, tokens := randomProjectionFixture(int64(n*97+3), n)
		fb := buildFrequencyBins(tokens, counts)
		for seed := int64(0); seed < 15; seed++ {
			want := referenceRandomizeProjection(p, counts, seed)
			got := RandomizeProjection(p, fb, seed)
			if !reflect.DeepEqual(got, want) {
				t.Fatalf("n=%d seed=%d: RandomizeProjection with hoisted bins diverged from reference", n, seed)
			}
		}
	}
}

// TestGenericSmoothingHoistedBinsMatchesReference is the same equivalence
// proof for GenericSmoothing.
func TestGenericSmoothingHoistedBinsMatchesReference(t *testing.T) {
	shapes := []int{5, 40, 137}
	for _, n := range shapes {
		p, counts, tokens := randomProjectionFixture(int64(n*197+11), n)
		fb := buildFrequencyBins(tokens, counts)
		for seed := int64(0); seed < 15; seed++ {
			want := referenceGenericSmoothing(tokens, counts, p, seed)
			got := GenericSmoothing(fb, p, seed)
			if !reflect.DeepEqual(got, want) {
				t.Fatalf("n=%d seed=%d: GenericSmoothing with hoisted bins diverged from reference", n, seed)
			}
		}
	}
}

// referenceMatchedGroupFamilyAnalysis is the pre-hoist implementation: it
// recomputes matchedGroup(f.Tokens, candidates, counts, trial) inside the
// per-distance loop, for every one of the c.MaxDistance x 200 (distance,
// trial) combinations, instead of hoisting the 200 trial-indexed groups out
// to compute once per family.
func referenceMatchedGroupFamilyAnalysis(f family, p profiles, full, abl Projection, counts map[string]int, c Config) FamilyResult {
	r := FamilyResult{ID: f.ID, Tokens: f.Tokens}
	candidates := make([]string, 0, len(p))
	for t := range p {
		if counts[t] >= c.MinObservations {
			candidates = append(candidates, t)
		}
	}
	sort.Strings(candidates)
	for d := 0; d < c.MaxDistance; d++ {
		fullCache, ablCache := map[string]map[string]float64{}, map[string]map[string]float64{}
		get := func(t string, proj Projection, cache map[string]map[string]float64) map[string]float64 {
			if x := cache[t]; x != nil {
				return x
			}
			x := ProjectDistribution(p[t].Right[d], proj)
			cache[t] = x
			return x
		}
		coh := func(group []string, mode string) (meanV, disp float64, medoid string) {
			if len(group) == 0 {
				return
			}
			sums := make([]float64, len(group))
			vals := []float64{}
			for i := range group {
				for j := 0; j < i; j++ {
					var a, b map[string]float64
					switch mode {
					case "token":
						a, b = countsFloat(p[group[i]].Right[d]), countsFloat(p[group[j]].Right[d])
					case "full":
						a, b = get(group[i], full, fullCache), get(group[j], full, fullCache)
					default:
						a, b = get(group[i], abl, ablCache), get(group[j], abl, ablCache)
					}
					v, _, _ := metricsFloat(a, b)
					vals = append(vals, v)
					sums[i] += v
					sums[j] += v
				}
			}
			meanV = mean(vals)
			for _, v := range vals {
				disp += (v - meanV) * (v - meanV)
			}
			if len(vals) > 0 {
				disp = math.Sqrt(disp / float64(len(vals)))
			}
			best := 0
			for i := 1; i < len(sums); i++ {
				if sums[i] > sums[best] {
					best = i
				}
			}
			medoid = group[best]
			return
		}
		tc, td, tm := coh(f.Tokens, "token")
		fc, fd, fm := coh(f.Tokens, "full")
		ac, ad, am := coh(f.Tokens, "ablated")
		tv, fv, av := []float64{}, []float64{}, []float64{}
		for trial := 0; trial < 200; trial++ {
			g := matchedGroup(f.Tokens, candidates, counts, trial)
			if len(g) != len(f.Tokens) {
				continue
			}
			x, _, _ := coh(g, "token")
			y, _, _ := coh(g, "full")
			z, _, _ := coh(g, "ablated")
			tv = append(tv, x)
			fv = append(fv, y)
			av = append(av, z)
		}
		sort.Float64s(tv)
		sort.Float64s(fv)
		sort.Float64s(av)
		r.Distances = append(r.Distances, FamilyDistance{Distance: d + 1, TokenCohesion: tc, ProjectedCohesionFull: fc, ProjectedCohesionAblated: ac, TokenDispersion: td, ProjectedDispersionFull: fd, ProjectedDispersionAblated: ad, TokenMedoid: tm, FullMedoid: fm, AblatedMedoid: am, MatchedPercentileToken: percentile(tv, tc), MatchedPercentileFull: percentile(fv, fc), MatchedPercentileAblated: percentile(av, ac)})
	}
	return r
}

func familyAnalysisFixture(seed int64) (family, profiles, Projection, Projection, map[string]int, Config) {
	p, counts, tokens := randomProjectionFixture(seed, 50)
	prof := profiles{}
	r := rand.New(rand.NewSource(seed + 1))
	for _, t := range tokens {
		right := make([]map[string]int, 6)
		for d := range right {
			m := map[string]int{}
			for k := 0; k < 4; k++ {
				m[tokens[r.Intn(len(tokens))]] += 1 + r.Intn(5)
			}
			right[d] = m
		}
		prof[t] = &profile{Right: right}
	}
	f := family{ID: 1, Tokens: []string{tokens[0], tokens[1], tokens[2]}}
	c := Config{MaxDistance: 6, MinObservations: 1}
	return f, prof, p, p, counts, c
}

// TestFamilyAnalysisHoistedMatchedGroupMatchesReference proves hoisting
// matchedGroup's 200 trial-indexed groups out of the c.MaxDistance-iteration
// distance loop (computing them once per family instead of once per
// (distance, trial) pair) produces a byte-identical FamilyResult to the
// pre-hoist implementation.
func TestFamilyAnalysisHoistedMatchedGroupMatchesReference(t *testing.T) {
	for seed := int64(0); seed < 8; seed++ {
		f, prof, full, abl, counts, c := familyAnalysisFixture(seed)
		want := referenceMatchedGroupFamilyAnalysis(f, prof, full, abl, counts, c)
		got := familyAnalysis(f, prof, full, abl, counts, c)
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("seed=%d: familyAnalysis with hoisted matchedGroup diverged from reference\ngot=%+v\nwant=%+v", seed, got, want)
		}
	}
}

// benchProjectionFixture is sized closer to the real vocabulary
// (~8,300 distinct tokens) than the small correctness fixtures above, so
// the reported ns/op, B/op, and allocs/op reflect the actual hot-path shape.
func benchProjectionFixture() (Projection, map[string]int, []string) {
	return randomProjectionFixture(99, 4000)
}

func BenchmarkRandomizeProjectionReference(b *testing.B) {
	p, counts, _ := benchProjectionFixture()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		referenceRandomizeProjection(p, counts, int64(i))
	}
}

func BenchmarkRandomizeProjectionHoistedBins(b *testing.B) {
	p, counts, tokens := benchProjectionFixture()
	fb := buildFrequencyBins(tokens, counts)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		RandomizeProjection(p, fb, int64(i))
	}
}

func BenchmarkGenericSmoothingReference(b *testing.B) {
	p, counts, tokens := benchProjectionFixture()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		referenceGenericSmoothing(tokens, counts, p, int64(i))
	}
}

func BenchmarkGenericSmoothingHoistedBins(b *testing.B) {
	p, counts, tokens := benchProjectionFixture()
	fb := buildFrequencyBins(tokens, counts)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		GenericSmoothing(fb, p, int64(i))
	}
}

func BenchmarkFamilyAnalysisReferenceMatchedGroup(b *testing.B) {
	f, prof, full, abl, counts, c := familyAnalysisFixture(7)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		referenceMatchedGroupFamilyAnalysis(f, prof, full, abl, counts, c)
	}
}

func BenchmarkFamilyAnalysisHoistedMatchedGroup(b *testing.B) {
	f, prof, full, abl, counts, c := familyAnalysisFixture(7)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		familyAnalysis(f, prof, full, abl, counts, c)
	}
}
