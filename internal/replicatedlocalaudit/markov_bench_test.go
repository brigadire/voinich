package replicatedlocalaudit

import (
	"fmt"
	"math/rand"
	"reflect"
	"testing"
)

// referenceMarkovBlocks is the pre-optimization implementation, preserved
// verbatim as a correctness oracle: it rebuilds the leave-one-block-out
// training model for every held-out block on every call, instead of reusing
// a precomputed markovHeldOut slice. Any refactor of markovBlocks/
// buildMarkovTraining must keep producing byte-identical output to this.
func referenceMarkovBlocks(blocks []block, seed int64) ([]block, int) {
	r := rand.New(rand.NewSource(seed))
	out := make([]block, 0, len(blocks))
	available := 0
	for _, held := range blocks {
		var train []block
		for _, b := range blocks {
			if b.ID != held.ID && b.Joint == held.Joint {
				train = append(train, b)
			}
		}
		m := buildMarkov(train)
		if m == nil {
			continue
		}
		available++
		z := held
		z.Tokens = append([]token(nil), held.Tokens...)
		for i := 0; i < len(z.Tokens); {
			j := i + 1
			for j < len(z.Tokens) && z.Tokens[j].Line == z.Tokens[i].Line {
				j++
			}
			prev := weightedChoice(r, m.vocab, m.unigram)
			for k := i; k < j; k++ {
				if k > i {
					n := weightedChoice(r, m.next[prev], m.weights[prev])
					if n == "" {
						n = weightedChoice(r, m.vocab, m.unigram)
					}
					prev = n
				}
				z.Tokens[k].Text = prev
			}
			i = j
		}
		out = append(out, z)
	}
	return out, available
}

// randomMarkovCorpus builds a synthetic multi-block corpus with several
// Joint groups (so some blocks have leave-one-out training data and others,
// singleton Joint groups, do not), skewed token frequencies, and multi-line
// blocks, to exercise both the "no training data" skip path and the
// multi-line weightedChoice restart path.
func randomMarkovCorpus(seed int64, nBlocks, blockLen, vocabSize int) []block {
	r := rand.New(rand.NewSource(seed))
	alphabet := make([]string, vocabSize)
	for i := range alphabet {
		alphabet[i] = fmt.Sprintf("t%02d", i)
	}
	blocks := make([]block, 0, nBlocks)
	for b := 0; b < nBlocks; b++ {
		joint := fmt.Sprintf("J%d", b%4) // one Joint group (b%4==3) ends up a singleton for nBlocks<=7-ish sweeps
		var toks []token
		line := 0
		for i := 0; i < blockLen; i++ {
			if i > 0 && r.Intn(5) == 0 {
				line++
			}
			idx := r.Intn(vocabSize)
			toks = append(toks, token{Text: alphabet[idx], Line: fmt.Sprintf("%d", line), Joint: joint})
		}
		blocks = append(blocks, block{ID: fmt.Sprintf("%s#%d", joint, b), Joint: joint, Tokens: toks})
	}
	return blocks
}

// TestMarkovBlocksMatchesReference verifies that precomputing the
// leave-one-block-out training models once (buildMarkovTraining) and
// drawing replicates from them (markovBlocks) produces byte-identical
// output and availability counts to the reference implementation that
// rebuilds every model inline, across many synthetic corpora and seeds.
func TestMarkovBlocksMatchesReference(t *testing.T) {
	shapes := []struct{ nBlocks, blockLen, vocab int }{
		{2, 6, 3},
		{5, 20, 6},
		{9, 40, 12},
		{16, 80, 25},
	}
	for _, shape := range shapes {
		blocks := randomMarkovCorpus(int64(shape.nBlocks*1000+shape.blockLen), shape.nBlocks, shape.blockLen, shape.vocab)
		training := buildMarkovTraining(blocks)
		for seed := int64(0); seed < 20; seed++ {
			wantBlocks, wantAvailable := referenceMarkovBlocks(blocks, seed)
			gotBlocks, gotAvailable := markovBlocks(training, seed)
			if gotAvailable != wantAvailable {
				t.Fatalf("shape=%+v seed=%d: available = %d, want %d", shape, seed, gotAvailable, wantAvailable)
			}
			if !reflect.DeepEqual(gotBlocks, wantBlocks) {
				t.Fatalf("shape=%+v seed=%d: markovBlocks output diverged from reference", shape, seed)
			}
		}
	}
}

// TestMarkovBlocksEmptyAndSingletonJoint covers the edge cases called out
// in the task27 correctness-oracle requirements: a block whose Joint group
// has no other members (no training data, so it must be dropped) alongside
// ordinary blocks, and a wholly empty block set.
func TestMarkovBlocksEmptyAndSingletonJoint(t *testing.T) {
	if got, avail := markovBlocks(buildMarkovTraining(nil), 1); len(got) != 0 || avail != 0 {
		t.Fatalf("empty block set: got %d blocks, available=%d", len(got), avail)
	}
	blocks := []block{
		{ID: "solo#0", Joint: "solo", Tokens: []token{{Text: "x", Line: "1"}}},
		{ID: "pair#0", Joint: "pair", Tokens: []token{{Text: "a", Line: "1"}, {Text: "b", Line: "1"}}},
		{ID: "pair#1", Joint: "pair", Tokens: []token{{Text: "b", Line: "1"}, {Text: "a", Line: "1"}}},
	}
	training := buildMarkovTraining(blocks)
	want, wantAvailable := referenceMarkovBlocks(blocks, 7)
	got, gotAvailable := markovBlocks(training, 7)
	if gotAvailable != wantAvailable || wantAvailable != 2 {
		t.Fatalf("singleton Joint should be dropped: available=%d want=%d", gotAvailable, wantAvailable)
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("singleton-Joint case diverged from reference")
	}
	for _, z := range got {
		if z.ID == "solo#0" {
			t.Fatalf("singleton-Joint block should have been dropped, not resampled")
		}
	}
}

// benchMarkovCorpus is a fixed corpus sized closer to a real
// replicated-local-structure-audit run (dozens of blocks, hundreds of
// tokens each) than the small correctness fixtures above.
func benchMarkovCorpus() []block {
	return randomMarkovCorpus(99, 40, 300, 60)
}

// BenchmarkMarkovBlocksReference measures the original implementation,
// which rebuilds every held-out block's training model from scratch inside
// the replicate call.
func BenchmarkMarkovBlocksReference(b *testing.B) {
	blocks := benchMarkovCorpus()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		referenceMarkovBlocks(blocks, int64(i))
	}
}

// BenchmarkMarkovBlocksPrecomputed measures the optimized path: training
// models are built once outside the loop (as RunAndWrite now does) and
// reused across every replicate call.
func BenchmarkMarkovBlocksPrecomputed(b *testing.B) {
	blocks := benchMarkovCorpus()
	training := buildMarkovTraining(blocks)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		markovBlocks(training, int64(i))
	}
}
