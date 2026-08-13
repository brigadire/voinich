package structuralprojection

import (
	"bytes"
	"math"
	"reflect"
	"strings"
	"testing"
	"time"
)

func close(a, b float64) bool { return math.Abs(a-b) < 1e-12 }
func testEdges() []Edge {
	return []Edge{{A: "a", B: "b", Position: .9, Left: .8, Right: .1, Similarity: .8, PositionReliability: 1, LeftReliability: 1, RightReliability: 1, Reliability: 1}, {A: "a", B: "c", Position: .4, Left: .4, Right: .9, Similarity: .7, PositionReliability: 1, LeftReliability: 1, RightReliability: 1, Reliability: 1}}
}

func TestSoftProjectionNormalizationAndSelfWeight(t *testing.T) {
	p := BuildProjection([]string{"a", "b", "c"}, testEdges(), .5, .5, 0, "full")
	for token, row := range p {
		s := 0.0
		for _, v := range row {
			s += v
		}
		if !close(s, 1) {
			t.Fatalf("row %s sums to %g", token, s)
		}
		if row[token] <= 0 {
			t.Fatalf("missing self weight for %s", token)
		}
	}
	if !close(p["b"]["b"], 1/1.8) {
		t.Fatalf("self weight was not one before normalization: %#v", p["b"])
	}
}
func TestThresholdProjection(t *testing.T) {
	p := BuildProjection([]string{"a", "b", "c"}, testEdges(), .75, .5, 0, "full")
	if p["a"]["b"] == 0 || p["a"]["c"] != 0 {
		t.Fatalf("threshold not applied: %#v", p["a"])
	}
}
func TestKNNProjection(t *testing.T) {
	p := BuildProjection([]string{"a", "b", "c"}, testEdges(), 0, 0, 1, "full")
	if len(p["a"]) != 2 || p["a"]["b"] == 0 {
		t.Fatalf("expected self plus strongest neighbour: %#v", p["a"])
	}
}
func TestAblatedProjectionExcludesRight(t *testing.T) {
	p := BuildProjection([]string{"a", "b", "c"}, testEdges(), .6, .5, 0, "future-ablated")
	if p["a"]["b"] == 0 || p["a"]["c"] != 0 {
		t.Fatalf("future ablation used right component: %#v", p["a"])
	}
}
func TestProjectedExactDistanceDistribution(t *testing.T) {
	p := Projection{"x": {"x": .5, "z": .5}, "y": {"y": 1}}
	got := ProjectDistribution(map[string]int{"x": 1, "y": 1}, p)
	want := map[string]float64{"x": .25, "z": .25, "y": .5}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %#v want %#v", got, want)
	}
}
func TestProjectionGain(t *testing.T) {
	p := Projection{"x": {"z": 1}, "y": {"z": 1}}
	g := gain(map[string]int{"x": 2}, map[string]int{"y": 2}, p)
	if !close(g, 1) {
		t.Fatalf("gain=%g want 1", g)
	}
}
func TestRandomSpaceControlDeterministicAndPreservesRowMass(t *testing.T) {
	p := Projection{"a": {"a": .5, "b": .5}, "b": {"b": .5, "a": .5}, "c": {"c": 1}}
	counts := map[string]int{"a": 10, "b": 11, "c": 12}
	x := RandomizeProjection(p, counts, 7)
	y := RandomizeProjection(p, counts, 7)
	if !reflect.DeepEqual(x, y) {
		t.Fatal("same seed produced different random spaces")
	}
	for k, row := range x {
		if len(row) != len(p[k]) {
			t.Fatalf("row %s degree changed: %d != %d", k, len(row), len(p[k]))
		}
		s := 0.0
		for _, v := range row {
			s += v
		}
		if !close(s, 1) {
			t.Fatalf("row %s mass=%g", k, s)
		}
	}
}
func TestGenericSmoothingControlDeterministic(t *testing.T) {
	tokens := []string{"a", "b", "c"}
	counts := map[string]int{"a": 10, "b": 11, "c": 12}
	p := Projection{"a": {"a": .5, "b": .5}, "b": {"b": 1}, "c": {"c": 1}}
	x := GenericSmoothing(tokens, counts, p, 9)
	y := GenericSmoothing(tokens, counts, p, 9)
	if !reflect.DeepEqual(x, y) {
		t.Fatal("generic smoothing is not deterministic")
	}
	if len(x["a"]) != 2 {
		t.Fatalf("degree not preserved: %#v", x["a"])
	}
}
func TestProjectedSuffixSimilarity(t *testing.T) {
	p := profiles{"A": {Suffix: []map[string]int{{"x\x1fy": 2}, {"x\x1fy\x1fz": 2}}}, "B": {Suffix: []map[string]int{{"u\x1fv": 2}, {"u\x1fv\x1fw": 2}}}}
	proj := Projection{"x": {"q": 1}, "u": {"q": 1}, "y": {"r": 1}, "v": {"r": 1}, "z": {"s": 1}, "w": {"s": 1}}
	r := sequenceResults(pair{"A", "B"}, p, proj, proj, 1)
	if len(r) != 2 || !close(r[0].ProjectedSimilarityFull, 1) || !close(r[1].ProjectedSimilarityFull, 1) || r[0].ExactSimilarity != 0 {
		t.Fatalf("unexpected sequence result: %#v", r)
	}
}
func TestShuffledCorpusControlPreservesFrequenciesAndSeed(t *testing.T) {
	c := corpus{Lines: [][]string{{"a", "b"}, {"a", "c"}}, Tokens: []string{"a", "b", "a", "c"}, Counts: map[string]int{"a": 2, "b": 1, "c": 1}}
	x := shuffledCorpus(c, "global", 5)
	y := shuffledCorpus(c, "global", 5)
	if !reflect.DeepEqual(x.Tokens, y.Tokens) {
		t.Fatal("shuffle is not deterministic")
	}
	got := map[string]int{}
	for _, v := range x.Tokens {
		got[v]++
	}
	if !reflect.DeepEqual(got, c.Counts) {
		t.Fatalf("frequencies changed: %#v", got)
	}
}

func TestProgressReportsPercentElapsedAndETA(t *testing.T) {
	var out bytes.Buffer
	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	p := newProgress(&out)
	p.clock = func() time.Time { return now }
	p.begin(4, "Controls")
	now = now.Add(10 * time.Second)
	p.update(2, 10, "Controls")
	s := out.String()
	for _, want := range []string{"[4/7]", "2/10 (20%)", "elapsed 00:10", "ETA 00:40"} {
		if !strings.Contains(s, want) {
			t.Fatalf("progress %q lacks %q", s, want)
		}
	}
}

func TestProgressThrottlesIntermediateUpdatesButPrintsCompletion(t *testing.T) {
	var out bytes.Buffer
	now := time.Unix(0, 0)
	p := newProgress(&out)
	p.clock = func() time.Time { return now }
	p.begin(1, "Work")
	p.update(1, 3, "Work")
	before := out.Len()
	now = now.Add(100 * time.Millisecond)
	p.update(2, 3, "Work")
	if out.Len() != before {
		t.Fatal("intermediate update was not throttled")
	}
	p.update(3, 3, "Work")
	if !strings.Contains(out.String(), "3/3 (100%)") {
		t.Fatal("completion was not printed")
	}
}
