package higherorderseq

import (
	"fmt"
	"math/rand"
	"reflect"
	"testing"
)

// referenceRunCMI is runCMI exactly as it stood before the cmiWorkspace
// rewrite.
func referenceRunCMI(cand Candidate, blocks []Block, permutations int, seed int64) CMIResult {
	obs := collectBNeighbors(cand.B(), blocks)
	observed := cmiBits(obs)
	r := rand.New(rand.NewSource(seed))
	null := make([]float64, permutations)
	for i := 0; i < permutations; i++ {
		null[i] = cmiBits(permuteWithinBlocks(obs, r))
	}
	mean, sd := meanSD(null)
	joint, left, right := jointTable(obs)
	n := float64(len(obs))
	contribution := 0.0
	if n > 0 {
		jn := float64(joint[[2]string{cand.A(), cand.C()}])
		pxy := jn / n
		px := float64(left[cand.A()]) / n
		py := float64(right[cand.C()]) / n
		if pxy > 0 && px > 0 && py > 0 {
			contribution = pxy * log2Ratio(pxy, px*py)
		}
	}
	return CMIResult{
		Sequence: cand.Sequence, CenterToken: cand.B(), Occurrences: len(obs),
		ObservedCMIBits: observed, NullMeanCMIBits: mean, NullSDCMIBits: sd,
		Permutations: permutations, EmpiricalP: empiricalP(observed, null),
		ContributionBits: contribution,
	}
}

// fixtureBlocksForCandidate builds n synthetic physical blocks whose tokens
// are drawn from a small shared vocabulary, so cand's central token B
// occurs with varied, overlapping-but-not-identical left/right neighbors
// across blocks of different lengths (including some too short to
// contribute any B-neighbor observation at all).
func fixtureBlocksForCandidate(nBlocks int, seed int64) (Candidate, []Block) {
	r := rand.New(rand.NewSource(seed))
	vocab := []string{"aiin", "chey", "shey", "ol", "or", "dy", "qokeey"}
	cand := Candidate{Sequence: "or aiin chey", Tokens: []string{"or", "aiin", "chey"}} // A=or, B=aiin, C=chey
	blocks := make([]Block, nBlocks)
	for bi := range blocks {
		length := 2 + r.Intn(10)
		toks := make([]Token, length)
		for i := range toks {
			toks[i] = Token{Text: vocab[r.Intn(len(vocab))]}
		}
		blocks[bi] = Block{ID: fmt.Sprintf("b%d", bi), Tokens: toks}
	}
	return cand, blocks
}

// TestRunCMIHoistMatchesReference proves the cmiWorkspace rewrite produces
// byte-identical CMIResult values to the pre-rewrite reference, across
// several block-set sizes (including ones with zero B-occurrences),
// permutation counts (including the jackknife's implicit non-loop shape via
// 0), and seeds.
func TestRunCMIHoistMatchesReference(t *testing.T) {
	sizes := []int{0, 1, 5, 40, 200}
	for _, n := range sizes {
		cand, blocks := fixtureBlocksForCandidate(n, int64(n)*97+11)
		for _, perms := range []int{0, 1, 25} {
			for seed := int64(0); seed < 4; seed++ {
				want := referenceRunCMI(cand, blocks, perms, seed)
				got := runCMI(cand, blocks, perms, seed)
				if !reflect.DeepEqual(got, want) {
					t.Fatalf("n=%d perms=%d seed=%d: diverged\ngot=%+v\nwant=%+v", n, perms, seed, got, want)
				}
			}
		}
	}
}

func benchCMIFixture() (Candidate, []Block) {
	return fixtureBlocksForCandidate(300, 42)
}

func BenchmarkRunCMIReference(b *testing.B) {
	cand, blocks := benchCMIFixture()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		referenceRunCMI(cand, blocks, 2000, int64(i))
	}
}

func BenchmarkRunCMIHoisted(b *testing.B) {
	cand, blocks := benchCMIFixture()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		runCMI(cand, blocks, 2000, int64(i))
	}
}
