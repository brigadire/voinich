package transitionnetwork

import (
	"fmt"
	"math"
	"math/rand"
	"testing"
)

// randomCorpus builds a synthetic multi-block corpus with a modest
// vocabulary and skewed token frequencies (so eligibility and opportunity
// thresholds produce a non-trivial mix of qualifying/non-qualifying
// tokens), used to exercise the permutation workspace at a scale closer to
// real data than the single-digit synthetic() fixture.
func randomCorpus(seed int64, nBlocks, blockLen, vocabSize int) ([]Token, []Block) {
	r := rand.New(rand.NewSource(seed))
	alphabet := make([]string, vocabSize)
	for i := range alphabet {
		alphabet[i] = fmt.Sprintf("t%02d", i)
	}
	var toks []Token
	var blocks []Block
	pos := 0
	for b := 0; b < nBlocks; b++ {
		currier := []string{"A", "B", "C"}[b%3]
		hand := []string{"H1", "H2"}[b%2]
		joint := currier + "/" + hand
		var bt []Token
		for i := 0; i < blockLen; i++ {
			// Zipf-ish skew: low indices drawn far more often.
			idx := int(math.Abs(r.NormFloat64()) * float64(vocabSize) / 3)
			if idx >= vocabSize {
				idx = idx % vocabSize
			}
			tok := Token{pos, alphabet[idx], currier, hand, joint}
			bt = append(bt, tok)
			toks = append(toks, tok)
			pos++
		}
		blocks = append(blocks, Block{fmt.Sprintf("%s#%d", joint, b), currier, hand, joint, bt})
	}
	return toks, blocks
}

func buildTestAnalysis(seed int64, nBlocks, blockLen, vocabSize, minCount int) *analysis {
	ts, bs := randomCorpus(seed, nBlocks, blockLen, vocabSize)
	counts, vocab, edges, data := buildData(ts, bs, minCount)
	return &analysis{Tokens: ts, Blocks: bs, Counts: counts, Vocab: vocab, Edges: edges, Data: data}
}

func almostEqual(a, b, tol float64) bool {
	if math.IsNaN(a) && math.IsNaN(b) {
		return true
	}
	return math.Abs(a-b) <= tol
}

// TestPermWorkspaceMatchesReferenceEdgeStats verifies that the indexed,
// buffer-reusing hot path (PermWorkspace.run) computes byte-identical
// per-edge median permuted log2-enrichment to the original map-based
// permutedStatistics, replicate for replicate, across several corpus
// shapes and seeds. Edge statistics only depend on invariant integer
// arithmetic (opportunity counts) plus a formula evaluated in the same
// operation order in both implementations, so no floating-point tolerance
// is needed here.
func TestPermWorkspaceMatchesReferenceEdgeStats(t *testing.T) {
	cases := []struct{ nBlocks, blockLen, vocabSize, minCount, minBlock int }{
		{4, 60, 10, 3, 2},
		{6, 120, 18, 5, 3},
		{3, 30, 6, 2, 1},
	}
	for ci, c := range cases {
		a := buildTestAnalysis(int64(1000+ci), c.nBlocks, c.blockLen, c.vocabSize, c.minCount)
		if len(a.Vocab) == 0 || len(a.Edges) == 0 {
			t.Fatalf("case %d: degenerate corpus (vocab=%d edges=%d)", ci, len(a.Vocab), len(a.Edges))
		}
		ws := newPermWorkspace(a, c.minBlock)
		for rep := 0; rep < 5; rep++ {
			wantEs, _, _ := permutedStatistics(a, rep, 42, c.minBlock)
			gotEs, _, _ := ws.run(42, rep, true)
			if len(wantEs) != len(gotEs) {
				t.Fatalf("case %d rep %d: edge count mismatch want=%d got=%d", ci, rep, len(wantEs), len(gotEs))
			}
			for e, want := range wantEs {
				got, ok := gotEs[e]
				if !ok {
					t.Fatalf("case %d rep %d: missing edge %v", ci, rep, e)
				}
				if want != got {
					t.Fatalf("case %d rep %d edge %v: want %.17g got %.17g", ci, rep, e, want, got)
				}
			}
		}
	}
}

// TestPermWorkspaceMatchesReferenceProfileStats verifies that the LOBO
// profile-null statistics (outgoing/incoming correlation, sign agreement,
// entropy effect) computed by the dense workspace match the original
// map-based implementation within the documented floating-point tolerance.
// The workspace derives each leave-one-out mean from a cached total
// instead of a fresh per-exclusion sum, so results are not required to be
// bit-identical, only equal within roundoff.
func TestPermWorkspaceMatchesReferenceProfileStats(t *testing.T) {
	a := buildTestAnalysis(2024, 6, 150, 16, 4)
	minBlock := 3
	ws := newPermWorkspace(a, minBlock)
	const tol = 1e-9
	for rep := 0; rep < 8; rep++ {
		_, wantOut, wantIn := permutedStatistics(a, rep, 7, minBlock)
		_, gotOut, gotIn := ws.run(7, rep, true)
		check := func(dir string, want, got map[string]profileNullStat) {
			if len(want) != len(got) {
				t.Fatalf("rep %d %s: token count mismatch want=%d got=%d", rep, dir, len(want), len(got))
			}
			for tok, w := range want {
				g, ok := got[tok]
				if !ok {
					t.Fatalf("rep %d %s: missing token %q", rep, dir, tok)
				}
				if !almostEqual(w.Correlation, g.Correlation, tol) {
					t.Errorf("rep %d %s %q: correlation want %.17g got %.17g", rep, dir, tok, w.Correlation, g.Correlation)
				}
				if !almostEqual(w.SignAgreement, g.SignAgreement, tol) {
					t.Errorf("rep %d %s %q: sign agreement want %.17g got %.17g", rep, dir, tok, w.SignAgreement, g.SignAgreement)
				}
				if !almostEqual(w.EntropyEffect, g.EntropyEffect, tol) {
					t.Errorf("rep %d %s %q: entropy effect want %.17g got %.17g", rep, dir, tok, w.EntropyEffect, g.EntropyEffect)
				}
			}
		}
		check("outgoing", wantOut, gotOut)
		check("incoming", wantIn, gotIn)
	}
}

// TestPermWorkspaceComputeProfilesFlagDoesNotAffectEdgeStats verifies that
// skipping the profile-vector/entropy work (as done during the
// pre-specified refinement pass, where those results are discarded) never
// changes the edge exceedance statistics that ARE consulted.
func TestPermWorkspaceComputeProfilesFlagDoesNotAffectEdgeStats(t *testing.T) {
	a := buildTestAnalysis(99, 5, 80, 12, 3)
	minBlock := 2
	ws := newPermWorkspace(a, minBlock)
	for rep := 0; rep < 6; rep++ {
		withProfiles, _, _ := ws.run(11, rep, true)
		withoutProfiles, nilOut, nilIn := ws.run(11, rep, false)
		if nilOut != nil || nilIn != nil {
			t.Fatalf("rep %d: expected nil profile maps when computeProfiles=false", rep)
		}
		if len(withProfiles) != len(withoutProfiles) {
			t.Fatalf("rep %d: edge count mismatch %d vs %d", rep, len(withProfiles), len(withoutProfiles))
		}
		for e, v := range withProfiles {
			if withoutProfiles[e] != v {
				t.Fatalf("rep %d edge %v: computeProfiles changed edge stat %.17g vs %.17g", rep, e, v, withoutProfiles[e])
			}
		}
	}
}

// TestPermWorkspaceDeterministicAcrossCalls verifies that re-running the
// same replicate on a fresh workspace reproduces identical results,
// guarding against any accidental cross-replicate state leaking through a
// reused scratch buffer.
func TestPermWorkspaceDeterministicAcrossCalls(t *testing.T) {
	a := buildTestAnalysis(5, 5, 90, 14, 3)
	minBlock := 2
	ws1 := newPermWorkspace(a, minBlock)
	ws2 := newPermWorkspace(a, minBlock)
	for rep := 0; rep < 4; rep++ {
		es1, out1, in1 := ws1.run(3, rep, true)
		es2, out2, in2 := ws2.run(3, rep, true)
		for e, v := range es1 {
			if es2[e] != v {
				t.Fatalf("rep %d: non-deterministic edge stat for %v", rep, e)
			}
		}
		for tok, v := range out1 {
			if out2[tok] != v {
				t.Fatalf("rep %d: non-deterministic outgoing stat for %q", rep, tok)
			}
		}
		for tok, v := range in1 {
			if in2[tok] != v {
				t.Fatalf("rep %d: non-deterministic incoming stat for %q", rep, tok)
			}
		}
	}
	// Re-running rep 0 on ws1 after later replicates must reproduce the
	// same result as the very first call: buffers are reused, not leaked.
	es0, _, _ := ws1.run(3, 0, true)
	es0Again, _, _ := ws1.run(3, 0, true)
	for e, v := range es0 {
		if es0Again[e] != v {
			t.Fatalf("re-running replicate 0 changed edge stat for %v: %.17g vs %.17g", e, v, es0Again[e])
		}
	}
}
