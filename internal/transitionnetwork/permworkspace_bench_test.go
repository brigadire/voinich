package transitionnetwork

import "testing"

// benchCorpus is a fixed, representative synthetic corpus used for both the
// old (reference) and new (indexed workspace) benchmarks, sized closer to
// the real -min-token-count=10 run (539 eligible tokens, 32 blocks) than
// the tiny fixtures in the correctness tests, so the reported ns/op, B/op,
// and allocs/op are informative for the actual hot path shape.
func benchCorpus() (*analysis, int) {
	minBlock := 5
	a := buildTestAnalysis(7, 32, 1200, 550, 10)
	return a, minBlock
}

// BenchmarkPermutedStatisticsReference measures the original map-based
// implementation: fresh map[string]... and map[EdgeKey]... allocations for
// every block, every replicate.
func BenchmarkPermutedStatisticsReference(b *testing.B) {
	a, minBlock := benchCorpus()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		permutedStatistics(a, i, 1, minBlock)
	}
}

// BenchmarkPermWorkspaceRun measures the indexed workspace with profile
// vectors computed (the primary-permutations path).
func BenchmarkPermWorkspaceRun(b *testing.B) {
	a, minBlock := benchCorpus()
	ws := newPermWorkspace(a, minBlock)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ws.run(1, i, true)
	}
}

// BenchmarkPermWorkspaceRunEdgesOnly measures the indexed workspace with
// profile-vector computation skipped, matching the pre-specified
// refinement path where only edge exceedance is consulted.
func BenchmarkPermWorkspaceRunEdgesOnly(b *testing.B) {
	a, minBlock := benchCorpus()
	ws := newPermWorkspace(a, minBlock)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ws.run(1, i, false)
	}
}

// BenchmarkCollectEffectVectorsReference measures the single biggest
// allocator identified by the task25 diagnostic profile in isolation.
func BenchmarkCollectEffectVectorsReference(b *testing.B) {
	a, minBlock := benchCorpus()
	d := a.Data[0]
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		out, in := map[string][]map[string]float64{}, map[string][]map[string]float64{}
		collectEffectVectors(d, a.Vocab, minBlock, &out, &in)
	}
}
