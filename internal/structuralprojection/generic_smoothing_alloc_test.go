package structuralprojection

import (
	"fmt"
	"math"
	"math/rand"
	"reflect"
	"testing"
)

// referenceGenericSmoothingAllocating is GenericSmoothing exactly as it
// stood after the frequencyBins hoist (task27 backlog item 3) and before
// the buffer-reuse optimization below: it still allocates a fresh pool
// slice and a fresh per-bin group copy on every one of the V token
// iterations. It is the correctness oracle for that specific change — the
// separately-preserved referenceGenericSmoothing in hoist_bench_test.go is
// the oracle for the *earlier* (bins-recomputed-per-call) stage and is not
// reused here to keep each optimization's proof isolated to what it
// actually changed.
func referenceGenericSmoothingAllocating(fb frequencyBins, p Projection, seed int64) Projection {
	r := rand.New(rand.NewSource(seed))
	out := Projection{}
	for _, src := range fb.sortedTokens {
		degree := len(p[src]) - 1
		if degree < 0 {
			degree = 0
		}
		b := fb.tokenBin[src]
		m := map[string]float64{src: 1}
		pool := make([]string, 0, len(fb.sortedTokens))
		appendBin := func(bin int) {
			group := append([]string(nil), fb.bins[bin]...)
			r.Shuffle(len(group), func(i, j int) { group[i], group[j] = group[j], group[i] })
			pool = append(pool, group...)
		}
		for delta := 0; delta <= fb.maxBin+1; delta++ {
			if b-delta >= 0 {
				appendBin(b - delta)
			}
			if delta > 0 && b+delta <= fb.maxBin {
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

// gsFixture is a named synthetic (Projection, counts, tokens) triple used
// to exercise GenericSmoothing's dependency surface: token identity
// (implicit in iteration), frequency (via counts -> bins), and projection
// degree (via len(p[src])). None of these fixtures depend on distance or
// metadata, matching GenericSmoothing's actual dependency surface.
type gsFixture struct {
	name   string
	p      Projection
	counts map[string]int
	tokens []string
}

func gsFixtures() []gsFixture {
	var out []gsFixture

	// Small vocabulary, mixed degree.
	{
		tokens := []string{"a", "b", "c", "d", "e"}
		counts := map[string]int{"a": 5, "b": 5, "c": 20, "d": 20, "e": 100}
		p := Projection{
			"a": {"a": 1, "b": 0.5},
			"b": {"b": 1},
			"c": {"c": 1, "d": 0.3, "e": 0.2},
			"d": {"d": 1},
			"e": {"e": 1, "a": 0.1, "b": 0.1, "c": 0.1, "d": 0.1},
		}
		out = append(out, gsFixture{"small_mixed_degree", p, counts, tokens})
	}

	// Equal frequencies: every token lands in the same log2 bin.
	{
		n := 40
		tokens := make([]string, n)
		counts := map[string]int{}
		p := Projection{}
		for i := 0; i < n; i++ {
			tokens[i] = fmt.Sprintf("eq%02d", i)
			counts[tokens[i]] = 10 // identical count -> identical bin for all
		}
		for i, t := range tokens {
			row := map[string]float64{t: 1}
			for k := 0; k < 3; k++ {
				row[tokens[(i+k+1)%n]] = 0.2
			}
			p[t] = row
		}
		out = append(out, gsFixture{"equal_frequencies_one_bin", p, counts, tokens})
	}

	// Highly skewed frequencies: exponentially growing counts, spanning
	// many distinct bins, several of them singletons.
	{
		n := 30
		tokens := make([]string, n)
		counts := map[string]int{}
		p := Projection{}
		for i := 0; i < n; i++ {
			tokens[i] = fmt.Sprintf("sk%02d", i)
			counts[tokens[i]] = 1 << i // each token its own bin (singleton bins)
		}
		for i, t := range tokens {
			row := map[string]float64{t: 1}
			if i+1 < n {
				row[tokens[i+1]] = 0.5
			}
			p[t] = row
		}
		out = append(out, gsFixture{"skewed_singleton_bins", p, counts, tokens})
	}

	// Zero eligible neighbours (degree 0: p[src] has only the self entry,
	// or is entirely absent) mixed with tokens needing many neighbours
	// (degree close to len(tokens)-1, exhausting most of the pool).
	{
		n := 25
		tokens := make([]string, n)
		counts := map[string]int{}
		p := Projection{}
		for i := 0; i < n; i++ {
			tokens[i] = fmt.Sprintf("de%02d", i)
			counts[tokens[i]] = 1 + i*3
		}
		for i, t := range tokens {
			switch {
			case i%3 == 0:
				// no eligible match: only the self weight, degree 0
				p[t] = map[string]float64{t: 1}
			case i%3 == 1:
				// all eligible matches: as many rows as the vocabulary allows
				row := map[string]float64{t: 1}
				for j := 0; j < n; j++ {
					if tokens[j] != t {
						row[tokens[j]] = 1
					}
				}
				p[t] = row
			default:
				row := map[string]float64{t: 1}
				row[tokens[(i+1)%n]] = 0.4
				p[t] = row
			}
		}
		out = append(out, gsFixture{"zero_and_max_degree", p, counts, tokens})
	}

	// Rare tokens (count == 1, bin 0) at both the start and end of the
	// sorted token order, alongside common tokens in high bins.
	{
		tokens := []string{"rare1", "rare2", "common1", "common2", "zzrare3"}
		counts := map[string]int{"rare1": 1, "rare2": 1, "common1": 500, "common2": 500, "zzrare3": 1}
		p := Projection{
			"rare1":   {"rare1": 1, "rare2": 0.5},
			"rare2":   {"rare2": 1},
			"common1": {"common1": 1, "common2": 0.5, "rare1": 0.2},
			"common2": {"common2": 1},
			"zzrare3": {"zzrare3": 1},
		}
		out = append(out, gsFixture{"rare_and_common_tokens", p, counts, tokens})
	}

	// Boundary bins: a token that IS the lowest bin (b=0, so the b-delta
	// branch is immediately exhausted) and a token that IS the highest bin
	// (b=maxBin, so the b+delta branch is immediately exhausted).
	{
		tokens := []string{"lo", "mid1", "mid2", "hi"}
		counts := map[string]int{"lo": 1, "mid1": 8, "mid2": 32, "hi": 4096}
		p := Projection{
			"lo":   {"lo": 1, "mid1": 0.5},
			"mid1": {"mid1": 1, "mid2": 0.3},
			"mid2": {"mid2": 1},
			"hi":   {"hi": 1, "lo": 0.1, "mid1": 0.1, "mid2": 0.1},
		}
		out = append(out, gsFixture{"boundary_lowest_and_highest_bin", p, counts, tokens})
	}

	// Large vocabulary (for a stronger multi-seed sweep, not just the
	// scaling benchmarks below).
	{
		n := 600
		r := rand.New(rand.NewSource(4242))
		tokens := make([]string, n)
		counts := map[string]int{}
		p := Projection{}
		for i := 0; i < n; i++ {
			tokens[i] = fmt.Sprintf("lg%04d", i)
		}
		for i, t := range tokens {
			counts[t] = 1 + r.Intn(3000)
			degree := r.Intn(15)
			row := map[string]float64{t: 1}
			for k := 0; k < degree; k++ {
				row[tokens[(i+k+1)%n]] = 0.1
			}
			p[t] = row
		}
		out = append(out, gsFixture{"large_vocabulary", p, counts, tokens})
	}

	return out
}

// TestGenericSmoothingBufferReuseMatchesReference is the reference-vs-
// optimized oracle for the pool/group buffer-reuse optimization: for every
// fixture above (small/large vocabulary, equal/skewed frequencies,
// singleton bins, zero/maximal degree, rare/common tokens, boundary bins)
// and 25 seeds each, the buffer-reusing GenericSmoothing must produce a
// byte-identical Projection to the allocating reference.
func TestGenericSmoothingBufferReuseMatchesReference(t *testing.T) {
	for _, fx := range gsFixtures() {
		fb := buildFrequencyBins(fx.tokens, fx.counts)
		for seed := int64(0); seed < 25; seed++ {
			want := referenceGenericSmoothingAllocating(fb, fx.p, seed)
			got := GenericSmoothing(fb, fx.p, seed)
			if !reflect.DeepEqual(got, want) {
				t.Fatalf("fixture=%s seed=%d: GenericSmoothing with reused buffers diverged from allocating reference\ngot=%#v\nwant=%#v", fx.name, seed, got, want)
			}
		}
	}
}

// TestGenericSmoothingBufferReuseDeterministicAcrossCalls is a determinism
// regression check specific to this change: confirms the buffer-reuse
// rewrite introduced no new map-iteration-order dependency (map iteration
// affecting RNG candidate order, float accumulation, or slice construction
// subsequently sampled by RNG) by calling it many times on identical input
// and requiring bit-identical results.
func TestGenericSmoothingBufferReuseDeterministicAcrossCalls(t *testing.T) {
	for _, fx := range gsFixtures() {
		fb := buildFrequencyBins(fx.tokens, fx.counts)
		seen := map[string]bool{}
		for i := 0; i < 30; i++ {
			out := GenericSmoothing(fb, fx.p, 777)
			bits := make([]byte, 0, 256)
			for _, t := range fx.tokens {
				row := out[t]
				keys := make([]string, 0, len(row))
				for k := range row {
					keys = append(keys, k)
				}
				sortStringsForTest(keys)
				for _, k := range keys {
					bits = append(bits, []byte(fmt.Sprintf("%s=%s:%d;", t, k, math.Float64bits(row[k])))...)
				}
			}
			seen[string(bits)] = true
		}
		if len(seen) != 1 {
			t.Fatalf("fixture=%s: GenericSmoothing produced %d distinct results across 30 calls with the same seed", fx.name, len(seen))
		}
	}
}

func sortStringsForTest(x []string) {
	for i := 1; i < len(x); i++ {
		for j := i; j > 0 && x[j] < x[j-1]; j-- {
			x[j], x[j-1] = x[j-1], x[j]
		}
	}
}
