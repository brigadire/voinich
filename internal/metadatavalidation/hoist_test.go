package metadatavalidation

import (
	"math/rand"
	"reflect"
	"sort"
	"strconv"
	"testing"
)

// referenceUniformBoundaries is *rand.Rand.Perm(n-1)[:count] exactly as
// UniformBoundaries called it before the scratch-buffer optimization: a
// fresh n-1-length slice allocated on every call.
func referenceUniformBoundaries(n, count int, rng *rand.Rand) []int {
	if n <= 1 || count <= 0 {
		return nil
	}
	if count > n-1 {
		count = n - 1
	}
	p := rng.Perm(n - 1)[:count]
	for i := range p {
		p[i]++
	}
	sort.Ints(p)
	return p
}

// TestUniformBoundariesScratchBufferMatchesReference proves the
// scratch-buffer optimization produces byte-identical output to the
// allocating reference, across several (n, count) shapes and seeds, using
// a FRESH scratch buffer per call (the simplest case).
func TestUniformBoundariesScratchBufferMatchesReference(t *testing.T) {
	shapes := []struct{ n, count int }{
		{2, 1}, {10, 1}, {10, 9}, {10, 100}, {39026, 5}, {39026, 500}, {39026, 5000},
	}
	for _, shape := range shapes {
		for seed := int64(0); seed < 10; seed++ {
			want := referenceUniformBoundaries(shape.n, shape.count, rand.New(rand.NewSource(seed)))
			scratch := make([]int, max(0, shape.n-1))
			got := UniformBoundaries(shape.n, shape.count, rand.New(rand.NewSource(seed)), scratch)
			if !reflect.DeepEqual(got, want) {
				t.Fatalf("shape=%+v seed=%d: got %v want %v", shape, seed, got, want)
			}
		}
	}
}

// TestUniformBoundariesScratchBufferReuseAcrossCallsMatchesReference is the
// critical proof for this optimization: a single scratch buffer, NEVER
// reset, reused across many consecutive calls sharing one *rand.Rand (the
// actual usage pattern in ValidateBoundaries's replicate loop), must
// produce the identical sequence of results as the reference calling
// rand.Perm(n-1) fresh every time on an equivalent *rand.Rand stream. This
// directly tests the doc-comment claim that scratch's leftover contents
// from a prior call are never read before being overwritten.
func TestUniformBoundariesScratchBufferReuseAcrossCallsMatchesReference(t *testing.T) {
	n, count := 500, 17
	refRng := rand.New(rand.NewSource(99))
	optRng := rand.New(rand.NewSource(99))
	scratch := make([]int, n-1)
	// Prime scratch with non-zero "garbage" so a bug that relies on a
	// zero-valued or freshly-allocated buffer would be caught.
	for i := range scratch {
		scratch[i] = 999999 - i
	}
	for call := 0; call < 50; call++ {
		want := referenceUniformBoundaries(n, count, refRng)
		got := UniformBoundaries(n, count, optRng, scratch)
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("call %d: reused-scratch result diverged from reference: got %v want %v", call, got, want)
		}
	}
}

// referenceClusterPermutationSummary is clusterPermutationSummary exactly
// as it stood before the clustersByK hoist: it rebuilds the Cluster-id
// string conversion for byK[k] fresh inside the per-replicate loop, once
// per (kind, method, replicate, k) combination instead of once per
// (kind, method, k).
func referenceClusterPermutationSummary(a []Assignment, records []TokenMetadata, n int, seed int64, p *progressReporter) map[string][]float64 {
	out := map[string][]float64{}
	rng := rand.New(rand.NewSource(seed))
	for _, kind := range []string{"currier", "hand"} {
		for _, method := range []string{"hierarchical", "k_medoids", "contiguous_segmentation"} {
			byK := map[int][]Assignment{}
			for _, x := range a {
				if x.WindowSize == 200 && x.Method == method {
					byK[x.K] = append(byK[x.K], x)
				}
			}
			if len(byK) == 0 {
				continue
			}
			ks := make([]int, 0, len(byK))
			for k := range byK {
				ks = append(ks, k)
			}
			sort.Ints(ks)
			base := byK[ks[0]]
			labels := make([]string, len(base))
			for i, x := range base {
				labels[i] = MetadataComposition(records, x.Start, x.End, kind).Label
			}
			vals := make([]float64, n)
			for z := 0; z < n; z++ {
				permuted := PermuteBlockLabels(labels, rng)
				best := 0.
				for _, k := range ks {
					clusters := make([]string, len(byK[k]))
					for i, x := range byK[k] {
						clusters[i] = strconv.Itoa(x.Cluster)
					}
					v := AssociationMetrics(permuted, clusters).NMI
					if v > best {
						best = v
					}
				}
				vals[z] = best
				if p != nil {
					p.update(z+1, n*6, "Block-aware metadata permutation controls")
				}
			}
			out[kind+"/"+method+"/window_200/max_nmi_over_k"] = vals
		}
	}
	return out
}

// clusterSummaryFixture builds a synthetic Assignment/TokenMetadata pair
// spanning multiple window sizes (only 200 is used), methods, and several
// K values per method, with varied Currier/Hand composition per window so
// MetadataComposition's Label/Purity differ across windows.
func clusterSummaryFixture() ([]Assignment, []TokenMetadata) {
	var records []TokenMetadata
	for i := 0; i < 400; i++ {
		currier := "A"
		if (i/50)%2 == 1 {
			currier = "B"
		}
		hand := "H1"
		if (i/33)%3 == 0 {
			hand = "H2"
		}
		records = append(records, TokenMetadata{Position: i, Currier: currier, Hand: hand})
	}
	var assignments []Assignment
	windowSize := 200
	numWindows := len(records) / windowSize
	for _, method := range []string{"hierarchical", "k_medoids", "contiguous_segmentation"} {
		for _, k := range []int{2, 3, 5} {
			for w := 0; w < numWindows; w++ {
				start, end := w*windowSize, (w+1)*windowSize
				assignments = append(assignments, Assignment{
					WindowSize: windowSize, Method: method, K: k, Index: w,
					Start: start, End: end, Cluster: (w + k) % k,
				})
			}
		}
	}
	// Also add a non-200 window size to confirm it's correctly excluded.
	assignments = append(assignments, Assignment{WindowSize: 100, Method: "hierarchical", K: 2, Index: 0, Start: 0, End: 100, Cluster: 0})
	return assignments, records
}

// TestClusterPermutationSummaryHoistMatchesReference proves the
// clustersByK precompute hoist produces byte-identical output to the
// pre-hoist reference across several permutation counts and seeds.
func TestClusterPermutationSummaryHoistMatchesReference(t *testing.T) {
	assignments, records := clusterSummaryFixture()
	for _, n := range []int{1, 5, 37} {
		for seed := int64(0); seed < 6; seed++ {
			want := referenceClusterPermutationSummary(assignments, records, n, seed, nil)
			got := clusterPermutationSummary(assignments, records, n, seed, nil)
			if !reflect.DeepEqual(got, want) {
				t.Fatalf("n=%d seed=%d: clusterPermutationSummary with hoisted clustersByK diverged from reference\ngot=%#v\nwant=%#v", n, seed, got, want)
			}
		}
	}
}

// TestClusterPermutationSummaryEmptyAndSingleWindowGroups covers the edge
// cases of no matching (window_size=200, method) group at all, and a group
// with only a single K value (ks has length 1, base==the only group).
func TestClusterPermutationSummaryEmptyAndSingleWindowGroups(t *testing.T) {
	records := []TokenMetadata{{Currier: "A", Hand: "H1"}, {Currier: "A", Hand: "H1"}}
	// No window_size==200 assignments at all.
	noneAssignments := []Assignment{{WindowSize: 50, Method: "hierarchical", K: 2, Start: 0, End: 2, Cluster: 0}}
	got := clusterPermutationSummary(noneAssignments, records, 3, 1, nil)
	if len(got) != 0 {
		t.Fatalf("expected no summaries with no window_size=200 assignments, got %v", got)
	}
	// Single K value for one method.
	singleK := []Assignment{
		{WindowSize: 200, Method: "hierarchical", K: 2, Start: 0, End: 2, Cluster: 0},
	}
	want := referenceClusterPermutationSummary(singleK, records, 3, 1, nil)
	got = clusterPermutationSummary(singleK, records, 3, 1, nil)
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("single-K fixture diverged: got %#v want %#v", got, want)
	}
}

// BenchmarkUniformBoundariesReference measures the pre-optimization
// implementation (fresh n-1-length allocation on every call).
func BenchmarkUniformBoundariesReference(b *testing.B) {
	rng := rand.New(rand.NewSource(1))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		referenceUniformBoundaries(39026, 200, rng)
	}
}

// BenchmarkUniformBoundariesScratchBuffer measures the reused-scratch-buffer
// implementation at the same (n, count) as the real default corpus size.
func BenchmarkUniformBoundariesScratchBuffer(b *testing.B) {
	rng := rand.New(rand.NewSource(1))
	scratch := make([]int, 39025)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		UniformBoundaries(39026, 200, rng, scratch)
	}
}

// BenchmarkClusterPermutationSummaryReference/HoistedClustersByK measure
// the clustersByK hoist in isolation, at a representative permutation
// count.
func BenchmarkClusterPermutationSummaryReference(b *testing.B) {
	assignments, records := clusterSummaryFixture()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		referenceClusterPermutationSummary(assignments, records, 100, int64(i), nil)
	}
}
func BenchmarkClusterPermutationSummaryHoistedClustersByK(b *testing.B) {
	assignments, records := clusterSummaryFixture()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		clusterPermutationSummary(assignments, records, 100, int64(i), nil)
	}
}

// clusterSummaryFixtureRealisticScale mirrors the real corpus's scale for
// window_size=200 over ~39,026 tokens (~195 windows) and the K-sweep range
// noted in the audit (up to ~14 distinct K values per method), so the
// clustersByK hoist's benefit is measurable rather than lost in noise from
// an unrealistically small synthetic fixture.
func clusterSummaryFixtureRealisticScale() ([]Assignment, []TokenMetadata) {
	const tokens = 39026
	const windowSize = 200
	var records []TokenMetadata
	for i := 0; i < tokens; i++ {
		currier := "A"
		if (i/500)%2 == 1 {
			currier = "B"
		}
		hand := "H1"
		if (i/333)%3 == 0 {
			hand = "H2"
		}
		records = append(records, TokenMetadata{Position: i, Currier: currier, Hand: hand})
	}
	numWindows := tokens / windowSize
	var assignments []Assignment
	for _, method := range []string{"hierarchical", "k_medoids", "contiguous_segmentation"} {
		for k := 2; k <= 15; k++ {
			for w := 0; w < numWindows; w++ {
				start, end := w*windowSize, (w+1)*windowSize
				assignments = append(assignments, Assignment{
					WindowSize: windowSize, Method: method, K: k, Index: w,
					Start: start, End: end, Cluster: (w + k) % k,
				})
			}
		}
	}
	return assignments, records
}

func BenchmarkClusterPermutationSummaryReferenceRealisticScale(b *testing.B) {
	assignments, records := clusterSummaryFixtureRealisticScale()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		referenceClusterPermutationSummary(assignments, records, 20, int64(i), nil)
	}
}
func BenchmarkClusterPermutationSummaryHoistedRealisticScale(b *testing.B) {
	assignments, records := clusterSummaryFixtureRealisticScale()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		clusterPermutationSummary(assignments, records, 20, int64(i), nil)
	}
}

// TestClusterPermutationSummaryHoistMatchesReferenceAtRealisticScale is the
// realistic-scale companion to TestClusterPermutationSummaryHoistMatchesReference,
// confirming exact equivalence at the ~195-window, 14-K-value scale the
// scaling benchmarks above use (a single seed/permutation-count pair, since
// this fixture is expensive to construct and the smaller fixture above
// already covers multiple seeds/counts).
func TestClusterPermutationSummaryHoistMatchesReferenceAtRealisticScale(t *testing.T) {
	assignments, records := clusterSummaryFixtureRealisticScale()
	want := referenceClusterPermutationSummary(assignments, records, 5, 42, nil)
	got := clusterPermutationSummary(assignments, records, 5, 42, nil)
	if !reflect.DeepEqual(got, want) {
		t.Fatal("realistic-scale fixture: clusterPermutationSummary with hoisted clustersByK diverged from reference")
	}
}
