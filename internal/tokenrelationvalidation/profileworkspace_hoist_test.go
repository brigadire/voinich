package tokenrelationvalidation

import (
	"fmt"
	"math/rand"
	"reflect"
	"testing"
)

// referenceProfilePermutationScores is profilePermutationScores exactly as
// it stood before the profileWorkspace rewrite: it calls the stateless,
// always-allocate buildLocalProfiles for every block on every call, instead
// of reusing a per-block-ID cached skeleton.
func referenceProfilePermutationScores(blocks []Block, candidates map[string]Candidate, maxD int) map[string]float64 {
	profiles := map[string]localProfiles{}
	for _, b := range blocks {
		profiles[b.ID] = buildLocalProfiles(b, maxD)
	}
	values := map[string][]float64{}
	for id, c := range candidates {
		if c.Family == "structural" {
			for _, b := range blocks {
				x := profileForBlock(c, b, profiles[b.ID], maxD)
				if x.EligiblePrimary {
					values[id] = append(values[id], x.Similarity)
				}
			}
		} else {
			for i := 0; i < len(blocks); i++ {
				pi := profileForBlock(c, blocks[i], profiles[blocks[i].ID], maxD)
				if !pi.EligiblePrimary {
					continue
				}
				for j := i + 1; j < len(blocks); j++ {
					pj := profileForBlock(c, blocks[j], profiles[blocks[j].ID], maxD)
					if pj.EligiblePrimary {
						v, _ := compareDistanceProfiles(profiles[blocks[i].ID], profiles[blocks[j].ID], c.A, c.B, maxD)
						values[id] = append(values[id], v)
					}
				}
			}
		}
	}
	out := map[string]float64{}
	for id, v := range values {
		_, out[id], _, _ = distribution(v)
	}
	return out
}

// fixtureProfileBlocksAndCandidates builds nBlocks synthetic blocks (each
// with several lines) drawn from a shared vocabulary, and a mix of
// "structural" and "distance-profile" candidates referencing pairs from
// that vocabulary.
func fixtureProfileBlocksAndCandidates(nBlocks, nCandidates int, seed int64) ([]Block, map[string]Candidate) {
	r := rand.New(rand.NewSource(seed))
	vocab := []string{"aiin", "chey", "shey", "ol", "or", "dy", "qokeey", "s", "daiin", "chol"}
	blocks := make([]Block, nBlocks)
	for bi := range blocks {
		var toks []Token
		nLines := 2 + r.Intn(4)
		for li := 0; li < nLines; li++ {
			lineLen := 3 + r.Intn(8)
			for k := 0; k < lineLen; k++ {
				toks = append(toks, Token{Text: vocab[r.Intn(len(vocab))], Line: fmt.Sprintf("l%d", li), LineIndex: k})
			}
		}
		blocks[bi] = Block{ID: fmt.Sprintf("b%d", bi), Tokens: toks}
	}
	candidates := map[string]Candidate{}
	for i := 0; i < nCandidates; i++ {
		id := fmt.Sprintf("c%d", i)
		a, b := vocab[r.Intn(len(vocab))], vocab[r.Intn(len(vocab))]
		family := "distance-profile"
		if i%2 == 0 {
			family = "structural"
		}
		candidates[id] = Candidate{ID: id, A: a, B: b, Family: family}
	}
	return blocks, candidates
}

// TestProfilePermutationScoresHoistMatchesReference proves the
// profileWorkspace rewrite produces byte-identical scores to the
// pre-rewrite reference across several permutation-like replicates of the
// SAME block IDs (mirroring how profilePermutations reuses one workspace
// across many PermuteWithinBlocks-shuffled block sets), several
// block/candidate set sizes, and several maxD values.
func TestProfilePermutationScoresHoistMatchesReference(t *testing.T) {
	sizes := []struct{ nBlocks, nCand int }{{0, 3}, {3, 0}, {5, 12}, {9, 30}}
	for _, sz := range sizes {
		blocks, candidates := fixtureProfileBlocksAndCandidates(sz.nBlocks, sz.nCand, int64(sz.nBlocks)*97+int64(sz.nCand)*13+11)
		for _, maxD := range []int{1, 3} {
			ws := newProfileWorkspace()
			r := rand.New(rand.NewSource(999))
			for rep := 0; rep < 6; rep++ {
				permuted := PermuteWithinBlocks(blocks, r.Int63())
				want := referenceProfilePermutationScores(permuted, candidates, maxD)
				got := profilePermutationScores(permuted, candidates, maxD, ws)
				if !reflect.DeepEqual(got, want) {
					t.Fatalf("nBlocks=%d nCand=%d maxD=%d rep=%d: diverged\ngot=%v\nwant=%v", sz.nBlocks, sz.nCand, maxD, rep, got, want)
				}
			}
		}
	}
}

// benchProfileFixture mirrors realistic production scale: ~32 physical
// blocks averaging ~1200 tokens each (the actual corpus's block count and
// size), ~300 distance-profile/structural candidates.
func benchProfileFixture() ([]Block, map[string]Candidate) {
	r := rand.New(rand.NewSource(1))
	vocab := make([]string, 400)
	for i := range vocab {
		vocab[i] = fmt.Sprintf("tok%03d", i)
	}
	blocks := make([]Block, 32)
	for bi := range blocks {
		var toks []Token
		nLines := 30 + r.Intn(20)
		for li := 0; li < nLines; li++ {
			lineLen := 6 + r.Intn(6)
			for k := 0; k < lineLen; k++ {
				toks = append(toks, Token{Text: vocab[r.Intn(len(vocab))], Line: fmt.Sprintf("l%d", li), LineIndex: k})
			}
		}
		blocks[bi] = Block{ID: fmt.Sprintf("b%d", bi), Tokens: toks}
	}
	candidates := map[string]Candidate{}
	for i := 0; i < 306; i++ {
		id := fmt.Sprintf("c%d", i)
		a, b := vocab[r.Intn(len(vocab))], vocab[r.Intn(len(vocab))]
		family := "distance-profile"
		if i%2 == 0 {
			family = "structural"
		}
		candidates[id] = Candidate{ID: id, A: a, B: b, Family: family}
	}
	return blocks, candidates
}

func BenchmarkProfilePermutationScoresReferenceSweep(b *testing.B) {
	blocks, candidates := benchProfileFixture()
	r := rand.New(rand.NewSource(7))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		permuted := PermuteWithinBlocks(blocks, r.Int63())
		referenceProfilePermutationScores(permuted, candidates, 3)
	}
}

func BenchmarkProfilePermutationScoresHoistedSweep(b *testing.B) {
	blocks, candidates := benchProfileFixture()
	ws := newProfileWorkspace()
	r := rand.New(rand.NewSource(7))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		permuted := PermuteWithinBlocks(blocks, r.Int63())
		profilePermutationScores(permuted, candidates, 3, ws)
	}
}
