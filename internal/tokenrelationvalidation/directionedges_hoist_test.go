package tokenrelationvalidation

import (
	"fmt"
	"math/rand"
	"reflect"
	"strings"
	"testing"
)

// referenceDirectionScoresAll is directionScoresAll exactly as it stood
// before the directionEdges hoist: it rebuilds the (from,to)-pair edge
// index from candidates/defaultMax on every call, even though those never
// change across a permutation replicate loop.
func referenceDirectionScoresAll(blocks []Block, candidates map[string]Candidate, defaultMax int) map[string]float64 {
	edges := map[Pair][]directedRef{}
	globalMax := defaultMax
	for id, c := range candidates {
		d := defaultMax
		if strings.Contains(c.Sources, "begin_end") && int(c.FrozenThreshold) > d {
			d = int(c.FrozenThreshold)
		}
		if d > globalMax {
			globalMax = d
		}
		edges[Pair{c.A, c.B}] = append(edges[Pair{c.A, c.B}], directedRef{id, true, d})
		edges[Pair{c.B, c.A}] = append(edges[Pair{c.B, c.A}], directedRef{id, false, d})
	}
	pos, neg, eligible := map[string]int{}, map[string]int{}, map[string]int{}
	for _, block := range blocks {
		freq := map[string]int{}
		for _, t := range block.Tokens {
			freq[t.Text]++
		}
		ab, ba := map[string]int{}, map[string]int{}
		for _, line := range splitLines(block) {
			for i, t := range line {
				for d := 1; d <= globalMax && i+d < len(line); d++ {
					refs := edges[Pair{t.Text, line[i+d].Text}]
					for _, ref := range refs {
						if d > ref.maxD {
							continue
						}
						if ref.forward {
							ab[ref.id]++
						} else {
							ba[ref.id]++
						}
					}
				}
			}
		}
		for id, c := range candidates {
			n := ab[id] + ba[id]
			if freq[c.A] < 5 || freq[c.B] < 5 || n < 5 {
				continue
			}
			eligible[id]++
			score := float64(ab[id]-ba[id]) / float64(n)
			if score > 0 {
				pos[id]++
			} else if score < 0 {
				neg[id]++
			}
		}
	}
	out := map[string]float64{}
	for id, n := range eligible {
		out[id] = float64(max(pos[id], neg[id])) / float64(n)
	}
	return out
}

// fixtureDirectionCandidatesAndBlocks builds a synthetic candidate set (some
// with a begin_end frozen threshold exceeding defaultMax, exercising the
// globalMax-widening branch) and a block set with enough repeated tokens
// for several candidates to clear the eligibility thresholds.
func fixtureDirectionCandidatesAndBlocks(nCandidates, nBlocks int, seed int64) (map[string]Candidate, []Block) {
	r := rand.New(rand.NewSource(seed))
	vocab := []string{"aiin", "chey", "shey", "ol", "or", "dy", "qokeey", "s"}
	candidates := map[string]Candidate{}
	for i := 0; i < nCandidates; i++ {
		id := fmt.Sprintf("c%d", i)
		a, b := vocab[r.Intn(len(vocab))], vocab[r.Intn(len(vocab))]
		cand := Candidate{ID: id, A: a, B: b, Family: "directional"}
		if i%3 == 0 {
			cand.Sources = "begin_end"
			cand.FrozenThreshold = float64(3 + r.Intn(3))
		}
		candidates[id] = cand
	}
	blocks := make([]Block, nBlocks)
	for bi := range blocks {
		length := 10 + r.Intn(40)
		toks := make([]Token, length)
		for i := range toks {
			toks[i] = Token{Text: vocab[r.Intn(len(vocab))], Line: fmt.Sprintf("l%d", i/8)}
		}
		blocks[bi] = Block{ID: fmt.Sprintf("b%d", bi), Tokens: toks}
	}
	return candidates, blocks
}

// TestDirectionScoresAllHoistMatchesReference proves the directionEdges
// hoist produces byte-identical scores to the pre-hoist reference, across
// several candidate/block set sizes and defaultMax values.
func TestDirectionScoresAllHoistMatchesReference(t *testing.T) {
	sizes := []struct{ nCand, nBlocks int }{{0, 5}, {5, 0}, {20, 15}, {80, 40}}
	for _, sz := range sizes {
		candidates, blocks := fixtureDirectionCandidatesAndBlocks(sz.nCand, sz.nBlocks, int64(sz.nCand)*97+int64(sz.nBlocks)*13+11)
		for _, defaultMax := range []int{1, 2, 5} {
			want := referenceDirectionScoresAll(blocks, candidates, defaultMax)
			de := buildDirectionEdges(candidates, defaultMax)
			got := directionScoresAll(blocks, candidates, de)
			if !reflect.DeepEqual(got, want) {
				t.Fatalf("nCand=%d nBlocks=%d defaultMax=%d: diverged\ngot=%v\nwant=%v", sz.nCand, sz.nBlocks, defaultMax, got, want)
			}
		}
	}
}

func benchDirectionFixture() (map[string]Candidate, []Block) {
	return fixtureDirectionCandidatesAndBlocks(1000, 200, 42)
}

func BenchmarkDirectionScoresAllReferenceSweep(b *testing.B) {
	candidates, blocks := benchDirectionFixture()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for rep := 0; rep < 5; rep++ {
			referenceDirectionScoresAll(blocks, candidates, 2)
		}
	}
}

func BenchmarkDirectionScoresAllHoistedSweep(b *testing.B) {
	candidates, blocks := benchDirectionFixture()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		de := buildDirectionEdges(candidates, 2)
		for rep := 0; rep < 5; rep++ {
			directionScoresAll(blocks, candidates, de)
		}
	}
}
