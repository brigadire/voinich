package tokenrepetition

import "testing"

func TestBuildEditGraphAgreesWithBruteForce(t *testing.T) {
	vocab := []string{"qokeedy", "qokedy", "qokeey", "qokaedy", "daiin", "chol", "shol", "xyzxyz"}
	glyphSeqs := map[string][]string{}
	for _, v := range vocab {
		glyphSeqs[v] = glyphs(v)
	}
	got := BuildEditGraph(vocab, glyphSeqs)

	brute := map[string]map[string]bool{}
	for i := 0; i < len(vocab); i++ {
		for j := i + 1; j < len(vocab); j++ {
			if LevenshteinGlyphs(glyphSeqs[vocab[i]], glyphSeqs[vocab[j]]) == 1 {
				if brute[vocab[i]] == nil {
					brute[vocab[i]] = map[string]bool{}
				}
				if brute[vocab[j]] == nil {
					brute[vocab[j]] = map[string]bool{}
				}
				brute[vocab[i]][vocab[j]] = true
				brute[vocab[j]][vocab[i]] = true
			}
		}
	}
	for a, neigh := range brute {
		for b := range neigh {
			found := false
			for _, n := range got.Adjacency[a] {
				if n == b {
					found = true
				}
			}
			if !found {
				t.Fatalf("indexed graph missed brute-force edge %s-%s", a, b)
			}
		}
	}
	for a, neigh := range got.Adjacency {
		for _, b := range neigh {
			if !brute[a][b] {
				t.Fatalf("indexed graph has spurious edge %s-%s", a, b)
			}
		}
	}
}

func TestConnectedComponentsAndChains(t *testing.T) {
	// a-b-c-d is a path (chain), e is isolated (no edges, not in graph).
	vocab := []string{"a", "ab", "abc", "abcd"}
	seqs := map[string][]string{}
	for _, v := range vocab {
		seqs[v] = glyphs(v)
	}
	g := BuildEditGraph(vocab, seqs)
	comps := g.ConnectedComponents()
	if len(comps) != 1 || len(comps[0]) != 4 {
		t.Fatalf("expected one connected component of 4, got %+v", comps)
	}
	chains := g.GreedyChains(3)
	if len(chains) != 1 || len(chains[0].Tokens) != 4 {
		t.Fatalf("expected one chain of length 4, got %+v", chains)
	}
}

func TestIndependenceExpectedAdjacency(t *testing.T) {
	vocab := []string{"ab", "ac"}
	seqs := map[string][]string{"ab": glyphs("ab"), "ac": glyphs("ac")}
	g := BuildEditGraph(vocab, seqs)
	freq := map[string]int{"ab": 10, "ac": 5}
	expected, edges := IndependenceExpectedAdjacency(g, freq, 100)
	if edges != 1 {
		t.Fatalf("expected 1 edge, got %d", edges)
	}
	want := 10.0 * 5.0 / 99.0
	if expected < want-1e-9 || expected > want+1e-9 {
		t.Fatalf("expected %.6f, got %.6f", want, expected)
	}
}
