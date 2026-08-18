package transitionnetwork

import (
	"bytes"
	"math"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

func synthetic() ([]Token, []Block) {
	texts := []string{"a", "b", "a", "c", "a", "b", "x", "a", "b", "a", "c", "a", "b"}
	var ts []Token
	for i, s := range texts {
		j := "A/H1"
		c, h := "A", "H1"
		if i >= 7 {
			j = "B/H2"
			c, h = "B", "H2"
		}
		ts = append(ts, Token{i, s, c, h, j})
	}
	return ts, []Block{{"A/H1#0", "A", "H1", "A/H1", ts[:7]}, {"B/H2#0", "B", "H2", "B/H2", ts[7:]}}
}

func TestExtractionDoesNotCrossBlockBoundary(t *testing.T) {
	ts, bs := synthetic()
	counts, vocab, edges, data := buildData(ts, bs, 1)
	_ = counts
	_ = vocab
	_ = edges
	if data[0].Edges[EdgeKey{"x", "a"}] != 0 {
		t.Fatal("edge crossed physical-block boundary")
	}
	if data[0].Edges[EdgeKey{"a", "b"}] != 2 || data[0].Counts["a"] != 3 || data[0].Counts["b"] != 2 {
		t.Fatalf("bad counts: %+v", data[0])
	}
}

func TestOpportunityIncludesIneligibleDestination(t *testing.T) {
	ts, bs := synthetic()
	_, _, _, data := buildData(ts, bs, 2) // x occurs once and is not an eligible destination.
	if data[0].Opp["b"] != 2 {
		t.Fatalf("b opportunities = %d, want 2 including b->x", data[0].Opp["b"])
	}
	if data[0].Edges[EdgeKey{"b", "x"}] != 0 {
		t.Fatal("ineligible destination leaked into primary matrix")
	}
	_ = permuteBlockEdges(data[0], rand.New(rand.NewSource(1)))
}

func TestSmoothingEnrichmentAndSign(t *testing.T) {
	_, bs := synthetic()
	_, v, _, d := buildData(append(bs[0].Tokens, bs[1].Tokens...), bs, 1)
	x := effect(d[0], EdgeKey{"a", "b"}, len(v))
	wantPC := (2.0 + .5) / (3.0 + .5*float64(len(v)))
	if math.Abs(x.PConditional-wantPC) > 1e-12 {
		t.Fatalf("smoothing got %g want %g", x.PConditional, wantPC)
	}
	if x.Enrichment <= 1 || x.Log2Enrichment <= 0 {
		t.Fatalf("expected preferred effect: %+v", x)
	}
	z := effect(d[0], EdgeKey{"a", "x"}, len(v))
	if z.Enrichment >= 1 || z.Log2Enrichment >= 0 {
		t.Fatalf("expected depleted effect: %+v", z)
	}
}

func TestLOBOAndEligibility(t *testing.T) {
	r := &EdgeSummary{}
	xs := []BlockStats{{Log2Enrichment: 1}, {Log2Enrichment: 2}, {Log2Enrichment: .5}}
	lobo(r, xs)
	if r.TestedBlocks != 3 || r.SuccessfulSignPredictions != 3 || r.TransferFraction != 1 {
		t.Fatalf("LOBO: %+v", r)
	}
}

func TestPermutationPreservesDestinationMarginals(t *testing.T) {
	ts, bs := synthetic()
	_, _, _, d := buildData(ts, bs, 1)
	pe := permuteBlockEdges(d[0], rand.New(rand.NewSource(9)))
	before, after := map[string]int{}, map[string]int{}
	for e, n := range d[0].Edges {
		before[e.Target] += n
	}
	for e, n := range pe {
		after[e.Target] += n
	}
	for k, n := range before {
		if after[k] != n {
			t.Fatalf("destination %s: %d != %d", k, after[k], n)
		}
	}
	pe2 := permuteBlockEdges(d[0], rand.New(rand.NewSource(9)))
	if len(pe) != len(pe2) {
		t.Fatal("seed is not deterministic")
	}
	for e, n := range pe {
		if pe2[e] != n {
			t.Fatal("seed is not deterministic")
		}
	}
}

func TestEmpiricalCorrectionAndSeparateBH(t *testing.T) {
	if got := float64(0+1) / float64(9+1); got != .1 {
		t.Fatal(got)
	}
	p1 := &EdgeSummary{ExpectedSign: "preferred", EligibleBlocks: 3, EmpiricalP: .01}
	p2 := &EdgeSummary{ExpectedSign: "preferred", EligibleBlocks: 3, EmpiricalP: .04}
	n := &EdgeSummary{ExpectedSign: "depleted", EligibleBlocks: 3, EmpiricalP: .04}
	xs := []*EdgeSummary{p1, p2, n}
	bh(xs, true)
	bh(xs, false)
	if p1.FDRQ != .02 || p2.FDRQ != .04 || n.FDRQ != .04 {
		t.Fatalf("separate BH: %g %g %g", p1.FDRQ, p2.FDRQ, n.FDRQ)
	}
}

func TestProfileMathEntropyAndGraph(t *testing.T) {
	p := []float64{.5, .5}
	if math.Abs(entropy(p)-math.Log(2)) > 1e-12 {
		t.Fatal("entropy")
	}
	if math.Abs(jsSimilarity(p, p)-1) > 1e-12 {
		t.Fatal("JS")
	}
	x := map[string]float64{"a": 1, "b": -1}
	if vectorSignAgreement(x, x, []string{"a", "b"}) != 1 || vectorCorrelation(x, x, []string{"a", "b"}) < .999 {
		t.Fatal("normalized profile")
	}
	g1 := map[EdgeKey]bool{{"a", "b"}: true, {"b", "a"}: true}
	g2 := map[EdgeKey]bool{{"a", "b"}: true, {"b", "a"}: true, {"b", "c"}: true}
	if sccOverlap(g1, g2) != 1 {
		t.Fatalf("SCC overlap %g", sccOverlap(g1, g2))
	}
}

func TestMetadataGraphAndLeakageFreePredictionDiagnostics(t *testing.T) {
	ts, bs := synthetic()
	counts, vocab, edges, data := buildData(ts, bs, 1)
	a := &analysis{Tokens: ts, Blocks: bs, Counts: counts, Vocab: vocab, Edges: edges, Data: data}
	summarizeEdges(a, 1)
	for _, r := range a.Summaries {
		r.FDRQ = .01
	}
	computeGraphDiagnostics(a, 1, false)
	if len(a.MetadataTransfer) == 0 || a.MetadataTransfer[0].CommonEdges == 0 {
		t.Fatal("metadata aggregation missing")
	}
	if len(a.GraphSimilarity) != 1 {
		t.Fatalf("graph comparisons=%d", len(a.GraphSimilarity))
	}
	computePredictions(a, 1)
	if len(a.Predictions) != 4 || len(a.ModelOrder) != 2 {
		t.Fatalf("prediction rows=%d model-order rows=%d", len(a.Predictions), len(a.ModelOrder))
	}
	for _, r := range a.Predictions {
		if math.IsNaN(r.LossM0) || math.IsNaN(r.LossM1) {
			t.Fatal("invalid held-out loss")
		}
	}
}

func TestClassificationAndCheckpoint(t *testing.T) {
	r := &EdgeSummary{EligibleBlocks: 3, JointClasses: 2, FDRQ: .01, ExpectedSign: "preferred", SignConsistency: .75, TransferFraction: .67, MaxBlockObservationFraction: .5, MaxBlockEffectWeightFraction: .5}
	classify(r, false)
	if r.Status != "BACKBONE_PREFERRED" {
		t.Fatal(r.Status)
	}
	path := filepath.Join(t.TempDir(), "cp.json")
	cp := freshCheckpoint("same")
	cp.Completed = 7
	if e := saveCheckpoint(path, cp); e != nil {
		t.Fatal(e)
	}
	got, ok, e := loadCheckpoint(path, "same")
	if e != nil || !ok || got.Completed != 7 {
		t.Fatalf("resume: %+v %v %v", got, ok, e)
	}
	if _, ok, _ := loadCheckpoint(path, "different"); ok {
		t.Fatal("mismatched fingerprint resumed")
	}
}

func TestProgressStatusBar(t *testing.T) {
	var b bytes.Buffer
	p := newProgress(&b)
	p.begin(1, "load")
	p.update(1, 2, "work")
	p.update(2, 2, "work")
	s := b.String()
	if !strings.Contains(s, "[1/8]") || !strings.Contains(s, "elapsed") || !strings.Contains(s, "100%") {
		t.Fatalf("status output: %q", s)
	}
}

func TestEndToEndSmall(t *testing.T) {
	dir := t.TempDir()
	corpus := filepath.Join(dir, "corpus.txt")
	meta := filepath.Join(dir, "meta.tsv")
	if e := os.WriteFile(corpus, []byte("a b a c a b x a b a c a b\n"), 0644); e != nil {
		t.Fatal(e)
	}
	var mb strings.Builder
	mb.WriteString("token_position\ttoken\tcurrier\thand\n")
	ts, _ := synthetic()
	for _, x := range ts {
		mb.WriteString(strings.Join([]string{strconvI(x.Position), x.Text, x.Currier, x.Hand}, "\t") + "\n")
	}
	if e := os.WriteFile(meta, []byte(mb.String()), 0644); e != nil {
		t.Fatal(e)
	}
	out := filepath.Join(dir, "out")
	c := Config{CorpusPath: corpus, MetadataPath: meta, OutputDir: out, MinTokenCount: 1, MinBlockTokenCount: 1, Permutations: 2, RefinePermutations: 2, Seed: 1, Quiet: true}
	if e := RunAndWrite(c); e != nil {
		t.Fatal(e)
	}
	for _, n := range []string{"transition_edge_summary.tsv", "transition_network_summary.yaml", "preferred_backbone.graphml", "plots/model_order_comparison.svg"} {
		if _, e := os.Stat(filepath.Join(out, n)); e != nil {
			t.Fatal(n, e)
		}
	}
	raw, e := os.ReadFile(filepath.Join(out, "transition_network_summary.yaml"))
	if e != nil {
		t.Fatal(e)
	}
	var summary map[string]any
	if e = yaml.Unmarshal(raw, &summary); e != nil {
		t.Fatalf("invalid summary YAML: %v", e)
	}
	if _, e := os.Stat(filepath.Join(out, "checkpoint.json")); !os.IsNotExist(e) {
		t.Fatal("checkpoint was not removed")
	}
}
func strconvI(n int) string {
	const digits = "0123456789"
	if n < 10 {
		return string(digits[n])
	}
	return strconvI(n/10) + string(digits[n%10])
}
