package propertytrajectory

import (
	"math"
	"math/rand"
	"reflect"
	"testing"

	"gopkg.in/yaml.v3"
)

func fixtureCorpus() corpus {
	lines := [][]string{{"a", "x", "b", "y", "a", "x"}, {"b", "y", "a", "z", "b", "x"}}
	c := corpus{Lines: lines, Counts: map[string]int{}}
	for _, l := range lines {
		c.Tokens = append(c.Tokens, l...)
		for _, x := range l {
			c.Counts[x]++
		}
	}
	return c
}
func almost(a, b float64) bool { return math.Abs(a-b) < 1e-9 }

func TestPropertyVectorConstructionAndGlobalNormalization(t *testing.T) {
	c := fixtureCorpus()
	p, stats := buildProperties(c, []structuralEdge{{"a", "b", .8, 1}, {"a", "x", .5, 1}}, 1)
	if p["a"].Properties["grapheme_length"].Raw != 1 {
		t.Fatal("grapheme length")
	}
	if p["a"].Properties["global_count"].Raw != 3 {
		t.Fatal("count")
	}
	if !almost(p["a"].Properties["max_structural_similarity"].Raw, .8) {
		t.Fatal("structural maximum")
	}
	for _, n := range allPropertyNames() {
		sum := 0.
		for _, v := range p {
			sum += v.Properties[n].Normalized
		}
		if math.Abs(sum/float64(len(p))) > 1e-8 {
			t.Fatalf("%s is not globally centered", n)
		}
		if stats[n]["stddev"] <= 0 {
			t.Fatalf("%s invalid scale", n)
		}
	}
}

func TestExactDistanceAggregationAndDeltas(t *testing.T) {
	c := fixtureCorpus()
	p, _ := buildProperties(c, nil, 1)
	n, r, obs, excluded := aggregate(c.Tokens, "a", 1, 2, p, []string{"global_count"})
	if obs != 3 || excluded != 0 {
		t.Fatalf("observations=%d excluded=%d", obs, excluded)
	}
	if len(n["global_count"]) != 3 || len(r["global_count"]) != 3 {
		t.Fatal("aggregation size")
	}
	a := map[string][]float64{"q": {1, 3}}
	b := map[string][]float64{"q": {0, 2}}
	s := summarize(a, b, a, b, []string{"q"})["q"]
	if s.MeanA != 2 || s.MeanB != 1 || s.Delta != 1 {
		t.Fatalf("bad delta: %#v", s)
	}
}

func TestSimilarityAndTrajectoryCorrelation(t *testing.T) {
	cos, e, m, c := similarity([]float64{1, 2}, []float64{2, 4})
	if !almost(cos, 1) || !almost(c, 1) || e <= 0 || m <= 0 {
		t.Fatalf("unexpected metrics %v %v %v %v", cos, e, m, c)
	}
	if !almost(pearson([]float64{1, 2, 3}, []float64{3, 5, 7}), 1) {
		t.Fatal("trajectory correlation")
	}
}

func TestMatchedBaselineDeterministic(t *testing.T) {
	c := fixtureCorpus()
	eligible := []string{"a", "b", "x", "y", "z"}
	ws := prepareMatchWorkspace(c, eligible)
	a := fallbackMatched(pair{"a", "b"}, c, ws, 3, rand.New(rand.NewSource(9)))
	b := fallbackMatched(pair{"a", "b"}, c, ws, 3, rand.New(rand.NewSource(9)))
	if !reflect.DeepEqual(a, b) {
		t.Fatalf("matching is not deterministic: %v %v", a, b)
	}
	if percentileRank([]float64{.1, .2, .3}, .2) != 100*2.0/3.0 {
		t.Fatal("percentile")
	}
}

func TestShuffledControlsPreserveCountsAndDeterminism(t *testing.T) {
	c := fixtureCorpus()
	a := shuffleCorpus(c, "global", rand.New(rand.NewSource(3)))
	b := shuffleCorpus(c, "global", rand.New(rand.NewSource(3)))
	if !reflect.DeepEqual(a, b) {
		t.Fatal("global shuffle is not deterministic")
	}
	count := func(x []string) map[string]int {
		m := map[string]int{}
		for _, q := range x {
			m[q]++
		}
		return m
	}
	if !reflect.DeepEqual(count(a), c.Counts) {
		t.Fatal("shuffle changed counts")
	}
	line := shuffleCorpus(c, "line-preserving", rand.New(rand.NewSource(4)))
	if len(line) != len(c.Tokens) {
		t.Fatal("line shuffle size")
	}
}

func TestPropertyGroupAblation(t *testing.T) {
	all := modeNames("all-properties")
	without := modeNames("all-minus-frequency")
	freq := modeNames("frequency-only")
	if len(all)-len(freq) != len(without) {
		t.Fatalf("ablation sizes all=%d freq=%d without=%d", len(all), len(freq), len(without))
	}
	for _, n := range without {
		for _, f := range freq {
			if n == f {
				t.Fatalf("frequency property %s survived ablation", n)
			}
		}
	}
	if len(modeNames("graphemic-form-only")) != 2 {
		t.Fatal("graphemic-form alias")
	}
}

func TestDeterministicYAMLOutput(t *testing.T) {
	x := PairResult{TokenA: "a", TokenB: "b", DistanceProfiles: []DistanceProfile{{Distance: 1, Properties: map[string]PropertySummary{"x": {MeanA: 1, MeanB: 2, Delta: -1}}}}}
	a, e := yaml.Marshal(x)
	if e != nil {
		t.Fatal(e)
	}
	b, e := yaml.Marshal(x)
	if e != nil {
		t.Fatal(e)
	}
	if string(a) != string(b) {
		t.Fatal("YAML output differs")
	}
}

func TestSelectPairsSkipsInapplicableVoynichReferences(t *testing.T) {
	previous := []pair{{"alpha", "beta"}}
	got, err := selectPairs(previous, "", map[string]int{
		"alpha": 2, "beta": 2, "or": 1, "s": 1,
	})
	if err != nil {
		t.Fatal(err)
	}
	want := []pair{{"alpha", "beta"}, {"or", "s"}}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("pairs = %v, want %v", got, want)
	}

	// An explicit selection remains strict: analyze validates its presence
	// in the corpus instead of silently dropping it.
	got, err = selectPairs(nil, "chedy,qokeey", map[string]int{})
	if err != nil || !reflect.DeepEqual(got, []pair{{"chedy", "qokeey"}}) {
		t.Fatalf("explicit pair = %v, err = %v", got, err)
	}
}
