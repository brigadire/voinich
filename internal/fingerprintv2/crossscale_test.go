package fingerprintv2

import (
	"math/rand"
	"testing"
)

func lineRecords(lineIndex int, currier string, tokens []string) []tokenRecord {
	out := make([]tokenRecord, len(tokens))
	lineID := "L" + string(rune('A'+lineIndex%26)) + string(rune('0'+(lineIndex/26)%10))
	for i, tok := range tokens {
		out[i] = tokenRecord{
			Token: tok, Glyph: []string{tok}, Line: lineIndex, LineID: lineID,
			IndexInLine: i, LineLength: len(tokens), Currier: currier, LocusType: "P",
		}
	}
	return out
}

// TestCS1PositiveControl verifies that a corpus with a deterministic,
// maximal family/line-position dependence is detected: observed NMI should
// be large (here, exactly 1, a perfect bijection) and the permutation
// p-value should be small.
func TestCS1PositiveControl(t *testing.T) {
	c := corpus{}
	for i := 0; i < 30; i++ {
		c.records = append(c.records, lineRecords(i, "A", []string{"aaa", "bbb"})...)
	}
	familyOf := map[string]int{"aaa": 0, "bbb": 1}
	famTest, _, n, _ := cs1Test(c, familyOf, nil, 200, rand.New(rand.NewSource(1)))
	if n != 60 {
		t.Fatalf("expected 60 family-bearing occurrences, got %d", n)
	}
	if famTest.Observed < 0.9 {
		t.Fatalf("expected near-maximal NMI for a deterministic family/position bijection, got %v", famTest.Observed)
	}
	if famTest.PValue > 0.05 {
		t.Fatalf("expected a significant permutation p-value for a maximal true effect, got %v", famTest.PValue)
	}
}

// TestCS1NegativeControl verifies that a corpus where family membership is
// exactly balanced across line positions produces zero observed NMI and a
// non-significant permutation result.
func TestCS1NegativeControl(t *testing.T) {
	c := corpus{}
	for i := 0; i < 40; i++ {
		if i%2 == 0 {
			c.records = append(c.records, lineRecords(i, "A", []string{"aaa", "bbb"})...)
		} else {
			c.records = append(c.records, lineRecords(i, "A", []string{"bbb", "aaa"})...)
		}
	}
	familyOf := map[string]int{"aaa": 0, "bbb": 1}
	famTest, _, n, _ := cs1Test(c, familyOf, nil, 200, rand.New(rand.NewSource(2)))
	if n != 80 {
		t.Fatalf("expected 80 family-bearing occurrences, got %d", n)
	}
	if famTest.Observed > 1e-9 {
		t.Fatalf("expected exactly zero NMI under perfect position/family balance, got %v", famTest.Observed)
	}
	if famTest.PValue < 0.2 {
		t.Fatalf("expected a non-significant permutation result under true independence, got p=%v", famTest.PValue)
	}
}

// TestCS1ConfoundedByRegime is the required synthetic confounded case: each
// Currier regime uses a single, constant family throughout (so within
// either regime alone, family carries zero entropy and therefore cannot be
// associated with anything), but regime A's lines are longer than regime
// B's, so the pooled corpus shows a spurious family/position association
// that is entirely an artifact of regime composition and must disappear
// once the analysis conditions on regime (task77's central methodological
// requirement: two marginal effects do not establish X _||_ Y).
func TestCS1ConfoundedByRegime(t *testing.T) {
	c := corpus{}
	for i := 0; i < 15; i++ {
		c.records = append(c.records, lineRecords(i, "A", []string{"aaa", "aaa", "aaa"})...)
	}
	for i := 15; i < 30; i++ {
		c.records = append(c.records, lineRecords(i, "B", []string{"bbb", "bbb"})...)
	}
	familyOf := map[string]int{"aaa": 0, "bbb": 1}

	pooled, _, _, _ := cs1Test(c, familyOf, nil, 50, rand.New(rand.NewSource(3)))
	if pooled.Observed < 0.05 {
		t.Fatalf("expected a nonzero pooled NMI from the regime confound, got %v", pooled.Observed)
	}

	var strA, strB corpus
	for _, r := range c.records {
		if r.Currier == "A" {
			strA.records = append(strA.records, r)
		} else {
			strB.records = append(strB.records, r)
		}
	}
	condA, _, _, _ := cs1Test(strA, familyOf, nil, 50, rand.New(rand.NewSource(4)))
	condB, _, _, _ := cs1Test(strB, familyOf, nil, 50, rand.New(rand.NewSource(5)))
	if condA.Observed != 0 {
		t.Fatalf("expected exactly zero within-regime-A NMI once conditioned (single constant family), got %v", condA.Observed)
	}
	if condB.Observed != 0 {
		t.Fatalf("expected exactly zero within-regime-B NMI once conditioned (single constant family), got %v", condB.Observed)
	}
}

func TestArticulationPointsOnPathGraph(t *testing.T) {
	// a-b-c-d: b and c are cut vertices, a and d are not.
	g := editGraph{nodes: []string{"a", "b", "c", "d"}, adj: map[string]map[string]bool{
		"a": {"b": true}, "b": {"a": true, "c": true}, "c": {"b": true, "d": true}, "d": {"c": true},
	}}
	art, bridges := articulationPointsAndBridges(g, g.nodes)
	if art != 2 {
		t.Fatalf("expected 2 articulation points on a 4-node path, got %d", art)
	}
	if bridges != 3 {
		t.Fatalf("expected all 3 edges of a path to be bridges, got %d", bridges)
	}
}

func TestKCoreOnTriangleWithPendant(t *testing.T) {
	// triangle a-b-c, plus pendant d attached to a.
	g := editGraph{nodes: []string{"a", "b", "c", "d"}, adj: map[string]map[string]bool{
		"a": {"b": true, "c": true, "d": true}, "b": {"a": true, "c": true}, "c": {"a": true, "b": true}, "d": {"a": true},
	}}
	core := kCoreDecomposition(g, g.nodes)
	if core["a"] != 2 || core["b"] != 2 || core["c"] != 2 {
		t.Fatalf("expected the triangle members to have coreness 2, got %+v", core)
	}
	if core["d"] != 1 {
		t.Fatalf("expected the pendant to have coreness 1, got %+v", core)
	}
}

func TestClusterAgreementIdenticalPartitions(t *testing.T) {
	a := []string{"x", "x", "y", "y", "z"}
	b := []string{"1", "1", "2", "2", "3"}
	ari, nmi, vi := clusterAgreement(a, b)
	if ari < 0.99 {
		t.Fatalf("expected ARI ~1 for a relabeled-identical partition, got %v", ari)
	}
	if nmi < 0.99 {
		t.Fatalf("expected NMI ~1 for a relabeled-identical partition, got %v", nmi)
	}
	if vi > 1e-9 {
		t.Fatalf("expected VI ~0 for a relabeled-identical partition, got %v", vi)
	}
}

func TestHubRemovalReducesGiantComponentForStarGraph(t *testing.T) {
	nodes := []string{"hub", "a", "b", "c", "d", "e", "f", "g", "h", "i", "j"}
	g := editGraph{nodes: nodes, adj: map[string]map[string]bool{}}
	g.adj["hub"] = map[string]bool{}
	for _, n := range nodes[1:] {
		g.adj["hub"][n] = true
		g.adj[n] = map[string]bool{"hub": true}
	}
	dep := hubRemovalGiantShare(g, 0.1) // removes the single highest-degree node: the hub
	if dep.GiantShareBefore != 1.0 {
		t.Fatalf("expected the star graph to start fully connected, got %v", dep.GiantShareBefore)
	}
	if dep.GiantShareAfter >= 0.5 {
		t.Fatalf("expected removing the hub of a star graph to shatter it, got giant share %v after", dep.GiantShareAfter)
	}
}
