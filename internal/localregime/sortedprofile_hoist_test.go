package localregime

import (
	"fmt"
	"math"
	"math/rand"
	"reflect"
	"sort"
	"testing"
)

// referenceDispersion and referenceJSSimilarity mirror dispersion and
// jsSimilarity exactly as they stood before the sortedProfile caching
// optimization (still using the already-fixed, deterministic sorted-keys
// jsSimilarity - this test proves only the caching hoist, not the earlier
// determinism fix).
func referenceDispersion(ps []profile, c profile) float64 {
	if len(ps) == 0 {
		return 0
	}
	s := 0.
	for _, p := range ps {
		s += 1 - jsSimilarity(p, c)
	}
	return s / float64(len(ps))
}

func referencePairwiseDispersion(ps []profile) float64 {
	if len(ps) < 2 {
		return 0
	}
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

func fixtureProfileSlice(n int, seed int64) []profile {
	r := rand.New(rand.NewSource(seed))
	out := make([]profile, n)
	for i := range out {
		p := profile{}
		for t := 0; t < 6; t++ {
			tok := fmt.Sprintf("tok%d", (i+t)%15)
			p[tok] += r.Float64()
		}
		out[i] = p
	}
	return out
}

func TestJSSimilaritySortedMatchesReference(t *testing.T) {
	for _, n := range []int{1, 5, 50, 300} {
		for _, keep := range []float64{0.1, 0.5, 1.0} {
			r := rand.New(rand.NewSource(int64(n)*1000 + int64(keep*10)))
			a, b := profile{}, profile{}
			for i := 0; i < n; i++ {
				if r.Float64() < keep {
					a[fmt.Sprintf("t%03d", i)] = r.Float64()
				}
				if r.Float64() < keep {
					b[fmt.Sprintf("t%03d", i)] = r.Float64()
				}
			}
			want := jsSimilarity(a, b)
			got := jsSimilaritySorted(sortProfile(a), sortProfile(b))
			if math.Float64bits(got) != math.Float64bits(want) {
				t.Fatalf("n=%d keep=%v: got %v want %v", n, keep, got, want)
			}
		}
	}
}

func TestWeightedOverlapCosineConcentrationSortedMatchReference(t *testing.T) {
	for _, n := range []int{1, 5, 50, 300} {
		for _, keep := range []float64{0.1, 0.5, 1.0} {
			r := rand.New(rand.NewSource(int64(n)*2000 + int64(keep*10)))
			a, b := profile{}, profile{}
			for i := 0; i < n; i++ {
				if r.Float64() < keep {
					a[fmt.Sprintf("t%03d", i)] = r.Float64()
				}
				if r.Float64() < keep {
					b[fmt.Sprintf("t%03d", i)] = r.Float64()
				}
			}
			sa, sb := sortProfile(a), sortProfile(b)
			if got, want := weightedOverlapSorted(sa, sb), weightedOverlap(a, b); math.Float64bits(got) != math.Float64bits(want) {
				t.Fatalf("weightedOverlapSorted n=%d keep=%v: got %v want %v", n, keep, got, want)
			}
			if got, want := cosineSorted(sa, sb), cosine(a, b); math.Float64bits(got) != math.Float64bits(want) {
				t.Fatalf("cosineSorted n=%d keep=%v: got %v want %v", n, keep, got, want)
			}
			if got, want := concentrationSorted(sa), concentration(a); math.Float64bits(got) != math.Float64bits(want) {
				t.Fatalf("concentrationSorted(a) n=%d keep=%v: got %v want %v", n, keep, got, want)
			}
		}
	}
}

func TestDispersionAndPairwiseDispersionHoistMatchReference(t *testing.T) {
	for _, n := range []int{0, 1, 2, 40, 250} {
		ps := fixtureProfileSlice(n, int64(n)*97+11)
		c := aggregate(ps)
		wantD := referenceDispersion(ps, c)
		gotD := dispersion(ps, c)
		if math.Float64bits(gotD) != math.Float64bits(wantD) {
			t.Fatalf("n=%d: dispersion diverged: got %v want %v", n, gotD, wantD)
		}
		wantP := referencePairwiseDispersion(ps)
		gotP := pairwiseDispersion(ps)
		if math.Float64bits(gotP) != math.Float64bits(wantP) {
			t.Fatalf("n=%d: pairwiseDispersion diverged: got %v want %v", n, gotP, wantP)
		}
	}
}

// referenceMatchCandidate/referenceBuildControlPool/referenceMatchedExpected
// mirror matchCandidate/buildControlPool/matchedExpected exactly as they
// stood before caching each occurrence's and each pool candidate's sorted
// profile: dotProduct(p, x.profile) re-sorted p's keys on every one of the
// up to 512 pool iterations, and re-sorted x.profile's keys (implicitly,
// inside dotProduct's own sortedProfileKeys(a) call - here a is p, but
// dotProduct's b=x.profile is looked up unsorted either way) every time
// too.
type referenceMatchCandidate struct {
	pos     int
	token   string
	count   int
	profile profile
	norm    float64
}

func referenceBuildControlPool(c corpus, cfg Config) []referenceMatchCandidate {
	stride := 1
	if len(c.Tokens) > 512 {
		stride = (len(c.Tokens) + 511) / 512
	}
	out := make([]referenceMatchCandidate, 0)
	for i := 0; i < len(c.Tokens); i += stride {
		p := localProfile(c, i, cfg.RegimeRadius, cfg.RegimeGap, "symmetric", cfg.RespectLineBoundaries)
		out = append(out, referenceMatchCandidate{i, c.Tokens[i], c.Counts[c.Tokens[i]], p, math.Sqrt(concentration(p))})
	}
	return out
}

func referenceMatchedExpected(c corpus, token string, ps []profile, pool []referenceMatchCandidate, cfg Config) matchedFuture {
	counts := map[int]map[string]int{}
	for d := 1; d <= cfg.MaxDistance; d++ {
		counts[d] = map[string]int{}
	}
	for _, p := range ps {
		type scored struct {
			i int
			s float64
		}
		var best []scored
		targetCount := float64(max(1, c.Counts[token]))
		pn := math.Sqrt(concentration(p))
		for i, x := range pool {
			if x.token == token {
				continue
			}
			ratio := math.Abs(math.Log(float64(max(1, x.count)) / targetCount))
			sim := 0.0
			if pn > 0 && x.norm > 0 {
				sim = dotProduct(p, x.profile) / (pn * x.norm)
			}
			score := sim - .05*ratio
			if len(best) < cfg.RegimeControlsK {
				best = append(best, scored{i, score})
				sort.Slice(best, func(i, j int) bool { return best[i].s > best[j].s })
			} else if score > best[len(best)-1].s {
				best[len(best)-1] = scored{i, score}
				sort.Slice(best, func(i, j int) bool { return best[i].s > best[j].s })
			}
		}
		for _, b := range best {
			base := pool[b.i].pos
			for d := 1; d <= cfg.MaxDistance; d++ {
				if base+d < len(c.Tokens) && (!cfg.RespectLineBoundaries || c.LineAt[base+d] == c.LineAt[base]) {
					counts[d][c.Tokens[base+d]]++
				}
			}
		}
	}
	out := matchedFuture{}
	for d, m := range counts {
		out[d] = normalizeCounts(m)
	}
	return out
}

func TestMatchedExpectedCachingMatchesReference(t *testing.T) {
	c := fixtureCorpus(30, 55)
	cfg := defaults(Config{RegimeRadius: 20, RegimeGap: 2, RegimeControlsK: 3, MaxDistance: 5})
	pool := buildControlPool(c, cfg)
	refPool := referenceBuildControlPool(c, cfg)
	for _, tok := range []string{"aiin", "chey", "ol"} {
		pos := positions(c, tok)
		var ps []profile
		for _, i := range pos {
			ps = append(ps, localProfile(c, i, cfg.RegimeRadius, cfg.RegimeGap, "symmetric", cfg.RespectLineBoundaries))
		}
		want := referenceMatchedExpected(c, tok, ps, refPool, cfg)
		got := matchedExpected(c, tok, pos, ps, pool, cfg)
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("token=%s: matchedExpected diverged\ngot=%v\nwant=%v", tok, got, want)
		}
	}
}

func benchProfileSlice() []profile {
	return fixtureProfileSlice(500, 42)
}

func BenchmarkPairwiseDispersionReference(b *testing.B) {
	ps := benchProfileSlice()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		referencePairwiseDispersion(ps)
	}
}

func BenchmarkPairwiseDispersionHoisted(b *testing.B) {
	ps := benchProfileSlice()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		pairwiseDispersion(ps)
	}
}
