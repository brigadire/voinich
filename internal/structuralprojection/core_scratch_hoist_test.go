package structuralprojection

import (
	"fmt"
	"math"
	"math/rand"
	"reflect"
	"sort"
	"testing"
)

// referenceNormalizeAllocating and referenceMetricsFloatAllocating are
// normalize/metricsFloat exactly as they stood before task28 Phase 2's
// scratch-buffer-reuse optimization: each allocated a fresh []string (and,
// for metricsFloat, a fresh map[string]bool) on every call instead of
// reusing a package-level scratch buffer. They are the correctness oracles
// for that specific change.

func referenceNormalizeAllocating(m map[string]float64) map[string]float64 {
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

func referenceMetricsFloatAllocating(a, b map[string]float64) (js, overlap, jaccard float64) {
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

func fixtureFloatMap(rng *rand.Rand, maxKeys int) map[string]float64 {
	m := map[string]float64{}
	n := rng.Intn(maxKeys + 1)
	for i := 0; i < n; i++ {
		m[fmt.Sprintf("tok%03d", rng.Intn(maxKeys*3))] = rng.Float64() * 10
	}
	return m
}

func TestNormalizeScratchReuseMatchesReference(t *testing.T) {
	rng := rand.New(rand.NewSource(11))
	sizes := []int{0, 1, 2, 5, 20, 60}
	for _, size := range sizes {
		for trial := 0; trial < 8; trial++ {
			m := fixtureFloatMap(rng, size)
			want := referenceNormalizeAllocating(m)
			got := normalize(m)
			if !reflect.DeepEqual(want, got) {
				t.Fatalf("size=%d trial=%d: want %v got %v", size, trial, want, got)
			}
		}
	}
}

// TestNormalizeScratchReuseInterleavedCalls proves the shared
// normalizeKeysScratch backing array carries no state between calls that
// could affect a later call's result - alternating differently-shaped
// fixtures (which forces the scratch slice to grow, shrink via [:0], and
// grow again) and checking each call against its own independent reference.
func TestNormalizeScratchReuseInterleavedCalls(t *testing.T) {
	rng := rand.New(rand.NewSource(12))
	sizes := []int{60, 1, 40, 0, 20, 5, 60, 2}
	for i, size := range sizes {
		m := fixtureFloatMap(rng, size)
		want := referenceNormalizeAllocating(m)
		got := normalize(m)
		if !reflect.DeepEqual(want, got) {
			t.Fatalf("call %d (size=%d): want %v got %v", i, size, want, got)
		}
	}
}

func TestMetricsFloatScratchReuseMatchesReference(t *testing.T) {
	rng := rand.New(rand.NewSource(21))
	sizes := []int{0, 1, 2, 5, 20, 60}
	for _, sizeA := range sizes {
		for _, sizeB := range sizes {
			for trial := 0; trial < 3; trial++ {
				a := fixtureFloatMap(rng, sizeA)
				b := fixtureFloatMap(rng, sizeB)
				wantJS, wantOverlap, wantJaccard := referenceMetricsFloatAllocating(a, b)
				gotJS, gotOverlap, gotJaccard := metricsFloat(a, b)
				if wantJS != gotJS || wantOverlap != gotOverlap || wantJaccard != gotJaccard {
					t.Fatalf("sizeA=%d sizeB=%d trial=%d: want (%v,%v,%v) got (%v,%v,%v)",
						sizeA, sizeB, trial, wantJS, wantOverlap, wantJaccard, gotJS, gotOverlap, gotJaccard)
				}
			}
		}
	}
}

func TestMetricsFloatScratchReuseInterleavedCalls(t *testing.T) {
	rng := rand.New(rand.NewSource(22))
	sizes := []int{60, 1, 40, 0, 20, 5, 60, 2}
	for i, size := range sizes {
		a := fixtureFloatMap(rng, size)
		b := fixtureFloatMap(rng, size+3)
		wantJS, wantOverlap, wantJaccard := referenceMetricsFloatAllocating(a, b)
		gotJS, gotOverlap, gotJaccard := metricsFloat(a, b)
		if wantJS != gotJS || wantOverlap != gotOverlap || wantJaccard != gotJaccard {
			t.Fatalf("call %d (size=%d): want (%v,%v,%v) got (%v,%v,%v)",
				i, size, wantJS, wantOverlap, wantJaccard, gotJS, gotOverlap, gotJaccard)
		}
	}
}

// Determinism-across-repeated-calls for normalize/metricsFloat on identical
// input is already covered by TestNormalizeDeterministicAcrossCalls and
// TestMetricsFloatDeterministicAcrossCalls in determinism_test.go (task27
// item 3) - those tests exercise normalize/metricsFloat directly, so they
// equally cover the scratch-reuse change here without duplication.

// benchFloatMap builds a deterministic (non-rand-package-dependent, so
// benchmarks are reproducible run to run) map of the given size for
// scaling benchmarks below.
func benchFloatMap(size int, offset int) map[string]float64 {
	m := make(map[string]float64, size)
	for i := 0; i < size; i++ {
		m[fmt.Sprintf("tok%06d", (i*7+offset)%(size*3+1))] = float64(i%11) + 0.5
	}
	return m
}

// benchSizes spans small (typical per-token row/degree-bounded) maps up to
// real-vocabulary scale (~8363 tokens), since metricsFloat's arguments are
// projected distributions whose size is not degree-bounded the way
// normalize's input is - see the dependency analysis in
// PERFORMANCE_REFACTOR_REPORT.md's task28 Phase 2 section.
var benchSizes = []int{10, 100, 500, 1000, 8363}

func BenchmarkNormalizeAllocatingScaling(b *testing.B) {
	for _, size := range benchSizes {
		b.Run(fmt.Sprintf("V=%d", size), func(b *testing.B) {
			m := benchFloatMap(size, 0)
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_ = referenceNormalizeAllocating(m)
			}
		})
	}
}

func BenchmarkNormalizeScratchReuseScaling(b *testing.B) {
	for _, size := range benchSizes {
		b.Run(fmt.Sprintf("V=%d", size), func(b *testing.B) {
			m := benchFloatMap(size, 0)
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_ = normalize(m)
			}
		})
	}
}

func BenchmarkMetricsFloatAllocatingScaling(b *testing.B) {
	for _, size := range benchSizes {
		b.Run(fmt.Sprintf("V=%d", size), func(b *testing.B) {
			a, bb := benchFloatMap(size, 0), benchFloatMap(size, 1)
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_, _, _ = referenceMetricsFloatAllocating(a, bb)
			}
		})
	}
}

func BenchmarkMetricsFloatScratchReuseScaling(b *testing.B) {
	for _, size := range benchSizes {
		b.Run(fmt.Sprintf("V=%d", size), func(b *testing.B) {
			a, bb := benchFloatMap(size, 0), benchFloatMap(size, 1)
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_, _, _ = metricsFloat(a, bb)
			}
		})
	}
}
