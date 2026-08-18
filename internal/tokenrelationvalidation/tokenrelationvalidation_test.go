package tokenrelationvalidation

import (
	"bytes"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func toks(words []string, line, c, h string) []Token {
	x := make([]Token, len(words))
	for n, w := range words {
		x[n] = Token{Position: n, Text: w, Line: line, LineIndex: n, Currier: c, Hand: h, Joint: c + "/" + h}
	}
	return x
}

func TestPhysicalBlockExtraction(t *testing.T) {
	x := append(toks([]string{"a", "b"}, "l1", "1", "1"), toks([]string{"c"}, "l2", "2", "1")...)
	x = append(x, toks([]string{"z"}, "l3", "?", "1")...)
	b, u := ExtractBlocks(x)
	if len(b) != 2 || b[0].Start != 0 || b[0].End != 2 || b[1].Joint != "2/1" || u != 1 {
		t.Fatalf("blocks=%+v unknown=%d", b, u)
	}
}

func TestNoRelationCrossesBlockOrLineBoundary(t *testing.T) {
	b1 := Block{ID: "1/1#1", Currier: "1", Hand: "1", Joint: "1/1", Tokens: toks([]string{"x", "a"}, "l1", "1", "1")}
	b2 := Block{ID: "2/1#1", Currier: "2", Hand: "1", Joint: "2/1", Tokens: toks([]string{"b", "x"}, "l2", "2", "1")}
	if x := directionForBlock(Candidate{ID: "d", A: "a", B: "b"}, b1, 5); x.ABeforeB != 0 {
		t.Fatal("cross-block observation")
	}
	one := Block{Tokens: append(toks([]string{"a"}, "l1", "1", "1"), toks([]string{"b"}, "l2", "1", "1")...)}
	if x := directionForBlock(Candidate{ID: "d", A: "a", B: "b"}, one, 5); x.ABeforeB != 0 {
		t.Fatal("cross-line observation")
	}
	_ = b2
}

func TestBlockEligibilityDirectionAndEnrichment(t *testing.T) {
	words := strings.Fields("a b x a b x a b x a b x a b x b b")
	x := directionForBlock(Candidate{ID: "d", A: "a", B: "b"}, Block{ID: "b", Tokens: toks(words, "l", "1", "1")}, 1)
	if !x.Eligible || x.Score <= 0 || x.EnrichmentAB <= 1 {
		t.Fatalf("unexpected %+v", x)
	}
}

func TestExactDistanceBlockProfiles(t *testing.T) {
	b := Block{Tokens: toks(strings.Fields("a x q b x q a x q b x q a x q b x q a x q b x q a x q b x q a x q b x q a x q b x q a x q b x q a x q b x q a x q b x q"), "l", "1", "1")}
	p := buildLocalProfiles(b, 3)
	if p.D["a"][1][1]["q"] != 10 {
		t.Fatalf("d2=%v", p.D["a"][1][1])
	}
	x := profileForBlock(Candidate{ID: "p", Family: "distance-profile", A: "a", B: "b"}, b, p, 3)
	if !x.EligiblePrimary || x.Distance <= 0 {
		t.Fatalf("profile=%+v", x)
	}
}

func TestLeaveOneOutReferenceNoLeakage(t *testing.T) {
	x := []DirectionBlock{{Score: 1}, {Score: .5}, {Score: -1}}
	train := append([]DirectionBlock(nil), x[:1]...)
	train = append(train, x[2:]...)
	if got := meanDirection(train); got != 0 {
		t.Fatalf("held-out leaked: %v", got)
	}
}

func TestMetadataTransfersCrossCurrierAndHand(t *testing.T) {
	s := []RelationSummary{{CandidateID: "d", Family: "directional"}}
	d := []DirectionBlock{{CandidateID: "d", BlockID: "b1", Currier: "1", Hand: "1", Score: 1, Eligible: true}, {CandidateID: "d", BlockID: "b2", Currier: "2", Hand: "2", Score: 1, Eligible: true}}
	m := buildMetadataTransfers(s, d, nil, false)
	seenC, seenH := false, false
	for _, x := range m {
		if x.Training != x.Heldout && x.Fraction == 1 {
			seenC = seenC || x.Dimension == "Currier"
			seenH = seenH || x.Dimension == "hand"
		}
	}
	if !seenC || !seenH {
		t.Fatalf("transfers=%+v", m)
	}
}

func TestFrequencyMatching(t *testing.T) {
	if !matched(10, 20) || matched(10, 21) {
		t.Fatal("factor-two matching")
	}
}

func TestWithinBlockPermutationDeterministicAndPreserving(t *testing.T) {
	b := []Block{{ID: "x", Tokens: toks(strings.Fields("a b c d e"), "l", "1", "1")}, {ID: "y", Tokens: toks(strings.Fields("q r s"), "l", "2", "1")}}
	x := PermuteWithinBlocks(b, 7)
	y := PermuteWithinBlocks(b, 7)
	if !reflect.DeepEqual(x, y) {
		t.Fatal("non-deterministic")
	}
	for n := range b {
		before, after := map[string]int{}, map[string]int{}
		for _, z := range b[n].Tokens {
			before[z.Text]++
		}
		for _, z := range x[n].Tokens {
			after[z.Text]++
		}
		if !reflect.DeepEqual(before, after) {
			t.Fatal("block vocabulary changed")
		}
	}
}

func TestBatchedDirectionalPermutationScore(t *testing.T) {
	b := Block{ID: "b", Tokens: toks(strings.Fields("a b x a b x a b x a b x a b x b b"), "l", "1", "1")}
	c := Candidate{ID: "d", A: "a", B: "b"}
	direct := directionForBlock(c, b, 1)
	want := 0.
	if direct.Eligible {
		p, n := 0, 0
		if direct.Score > 0 {
			p++
		} else if direct.Score < 0 {
			n++
		}
		want = float64(max(p, n))
	}
	candidates := map[string]Candidate{"d": c}
	got := directionScoresAll([]Block{b}, candidates, buildDirectionEdges(candidates, 1))["d"]
	if got != want {
		t.Fatalf("batched=%v direct=%v", got, want)
	}
}

func TestBHFDR(t *testing.T) {
	q := BH([]float64{.01, .04, .03})
	want := []float64{.03, .04, .04}
	for i := range q {
		if q[i] != want[i] {
			t.Fatalf("q=%v", q)
		}
	}
}

func TestClassificationAndRuleLike(t *testing.T) {
	s := RelationSummary{Family: "directional", EligibleBlocks: 3, PhysicalBlocks: 3, JointClasses: 2, CurrierClasses: 2, Hands: 2, SignConsistency: .75, MedianEnrichment: 1.2, TransferSuccess: .67, FDRQ: .05, ProfileMedian: .7}
	if got := Classify(s, true, true); got != "UNIVERSAL" {
		t.Fatal(got)
	}
	if !RuleLike(s) {
		t.Fatal("expected rule-like")
	}
	s.FDRQ = .051
	if RuleLike(s) {
		t.Fatal("q threshold ignored")
	}
	if got := Classify(RelationSummary{EligibleBlocks: 1, PhysicalBlocks: 1, TransferSuccess: .2}, false, false); got != "BLOCK_SPECIFIC" {
		t.Fatal(got)
	}
}

func TestCandidateSetOnlyFrozenAndOldInventoryAccepted(t *testing.T) {
	d := t.TempDir()
	files := map[string]string{
		"begin_end_candidates.yaml":   "meta: {token_occurrences: 38887}\nparameters: {max_window: 3}\ncandidates:\n  - {begin_candidate: olda, end_candidate: oldb}\n",
		"distance_context_pairs.yaml": "token_count: 38887\nparameters: {max_distance: 20}\npairs: []\n",
		"sequence_analysis.yaml":      "meta: {token_occurrences: 38887}\nrepeated_ngrams: {}\n",
		"structural_reliability.yaml": "meta: {token_occurrences: 38887}\nparameters: {threshold: 0.7}\nreference_pairs: []\n",
		"structural_classes.yaml":     "models: []\n",
		"soft_structural_space.yaml":  "parameters: {graph_min_similarity: 0.6}\n",
		"soft_structural_pairs.tsv":   "token_a\ttoken_b\traw_similarity\nnewtail\tx\t0.2\n",
	}
	for n, v := range files {
		if e := os.WriteFile(filepath.Join(d, n), []byte(v), 0644); e != nil {
			t.Fatal(e)
		}
	}
	c, _, _, e := loadCandidates(d)
	if e != nil {
		t.Fatal(e)
	}
	if len(c) != 1 || c[0].A != "olda" || c[0].StoredTokenCount != 38887 {
		t.Fatalf("candidates=%+v", c)
	}
	for _, x := range c {
		if x.A == "newtail" {
			t.Fatal("candidate discovered from canonical tail")
		}
	}
}

func TestCanonicalStatisticsRecomputed(t *testing.T) {
	b := Block{Tokens: toks(strings.Fields("olda oldb olda oldb olda oldb olda oldb olda oldb"), "l", "1", "1")}
	x := directionForBlock(Candidate{A: "olda", B: "oldb", StoredTokenCount: 38887}, b, 1)
	if x.CountA != 5 || x.CountB != 5 {
		t.Fatalf("not recomputed: %+v", x)
	}
}

func TestStatusBar(t *testing.T) {
	var w bytes.Buffer
	p := newProgress(&w)
	p.begin(3, "Testing")
	p.update(1, 2, "Testing")
	p.update(2, 2, "Testing")
	s := w.String()
	for _, want := range []string{"[3/8]", "[==========..........]", "elapsed"} {
		if !strings.Contains(s, want) {
			t.Fatalf("status %q lacks %q", s, want)
		}
	}
}
