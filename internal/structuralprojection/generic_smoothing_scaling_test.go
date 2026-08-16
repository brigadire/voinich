package structuralprojection

import (
	"fmt"
	"math/rand"
	"testing"
)

// genericSmoothingScalingFixture builds a synthetic (Projection, counts,
// tokens) triple of a given vocabulary size n, with realistic frequency
// spread (so tokens land across many log2 bins, not just one) and modest
// per-token degree, for benchmarking GenericSmoothing's scaling behavior.
func genericSmoothingScalingFixture(n int) (Projection, map[string]int, []string) {
	r := rand.New(rand.NewSource(int64(n)*31 + 1))
	tokens := make([]string, n)
	counts := map[string]int{}
	p := Projection{}
	for i := 0; i < n; i++ {
		tokens[i] = fmt.Sprintf("tok%06d", i)
	}
	for i, t := range tokens {
		counts[t] = 1 + r.Intn(8000)
		degree := 1 + r.Intn(10)
		row := map[string]float64{t: 1}
		for k := 0; k < degree; k++ {
			row[tokens[(i+k+1)%n]] = 0.1
		}
		p[t] = row
	}
	return p, counts, tokens
}

// benchmarkGenericSmoothingAtSize runs both the allocating reference and
// the buffer-reusing implementation at a fixed vocabulary size n, reporting
// ns/op, B/op, and allocs/op for each — used at several V to show the
// empirical scaling (not inferred from source alone), per task27's ask.
func benchmarkGenericSmoothingAtSize(b *testing.B, n int) {
	p, counts, tokens := genericSmoothingScalingFixture(n)
	fb := buildFrequencyBins(tokens, counts)
	b.Run(fmt.Sprintf("Allocating/V=%d", n), func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			referenceGenericSmoothingAllocating(fb, p, int64(i))
		}
	})
	b.Run(fmt.Sprintf("BufferReused/V=%d", n), func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			GenericSmoothing(fb, p, int64(i))
		}
	})
}

func BenchmarkGenericSmoothingScalingV100(b *testing.B)  { benchmarkGenericSmoothingAtSize(b, 100) }
func BenchmarkGenericSmoothingScalingV500(b *testing.B)  { benchmarkGenericSmoothingAtSize(b, 500) }
func BenchmarkGenericSmoothingScalingV1000(b *testing.B) { benchmarkGenericSmoothingAtSize(b, 1000) }
func BenchmarkGenericSmoothingScalingV4000(b *testing.B) { benchmarkGenericSmoothingAtSize(b, 4000) }
func BenchmarkGenericSmoothingScalingV8363(b *testing.B) { benchmarkGenericSmoothingAtSize(b, 8363) }
