package notation

import (
	"math"
	"strings"
	"testing"
)

func TestComparatorIsVectorOnlyAndFailClosed(t *testing.T) {
	c := Fingerprint{Metadata: map[string]string{"metric_registry_version": MetricRegistryVersion}, Metrics: []Metric{{MetricID: "G01", Family: "G", Value: 2, Status: Comparable}, {MetricID: "L01", Family: "L", Status: NotComparable}}}
	r := Fingerprint{Metadata: map[string]string{"metric_registry_version": MetricRegistryVersion}, Metrics: []Metric{{MetricID: "G01", Family: "G", Value: 1, Status: Comparable}, {MetricID: "L01", Family: "L", Value: 1, Status: Comparable}}}
	rows, fam, err := Compare(c, r, []Scale{{MetricID: "G01", Center: 0, Spread: 2}})
	if err != nil {
		t.Fatal(err)
	}
	if rows[0].Distance != .5 || rows[0].Status != Comparable {
		t.Fatalf("row=%+v", rows[0])
	}
	if rows[1].Status != NotComparable {
		t.Fatal("missing candidate level was compared")
	}
	if len(fam) != 5 {
		t.Fatalf("families=%d", len(fam))
	}
	for _, f := range fam {
		if f.Family == "TOTAL" {
			t.Fatal("aggregate score emitted")
		}
	}
	var b strings.Builder
	if err := WriteComparisonTSV(&b, rows, fam); err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(b.String(), "metric_id\tfamily\tvm_value") || strings.Contains(b.String(), "d_TOTAL") {
		t.Fatal("invalid comparison schema")
	}
}

func TestFrozenDistanceFunctions(t *testing.T) {
	js, err := JensenShannon([]float64{1, 0}, []float64{0, 1})
	if err != nil || math.Abs(js-1) > 1e-12 {
		t.Fatalf("JS=%v %v", js, err)
	}
	w, err := Wasserstein1([]float64{0}, []float64{1}, []float64{2}, []float64{1})
	if err != nil || math.Abs(w-2) > 1e-12 {
		t.Fatalf("W=%v %v", w, err)
	}
	a, err := NormalizedCurveArea([]float64{0, 1}, []float64{0, 1}, []float64{0, 2})
	if err != nil || a <= 0 {
		t.Fatalf("area=%v %v", a, err)
	}
}

func TestNotationDeltaRespectsMissing(t *testing.T) {
	a := Fingerprint{Metrics: []Metric{{MetricID: "G", Family: "G", Value: 1, Status: Comparable}, {MetricID: "L", Family: "L", Status: NotComparable}}}
	b := Fingerprint{Metrics: []Metric{{MetricID: "G", Family: "G", Value: 3, Status: Comparable}, {MetricID: "L", Family: "L", Value: 4, Status: Comparable}}}
	d := NotationDelta(a, b)
	if d[0].Delta != 2 || d[0].Status != Comparable {
		t.Fatalf("delta=%+v", d[0])
	}
	if d[1].Status != NotComparable {
		t.Fatal("missing delta imputed")
	}
}

func TestIndexedEditGraphMatchesBruteForce(t *testing.T) {
	seqs := map[string][]string{"a": {"a"}, "ab": {"a", "b"}, "ac": {"a", "c"}, "bc": {"b", "c"}, "abc": {"a", "b", "c"}}
	vocab := []string{"a", "ab", "ac", "bc", "abc"}
	_, got := buildEditAdjacency(vocab, seqs)
	want := 0
	for i := range vocab {
		for j := i + 1; j < len(vocab); j++ {
			if sequenceEditOne(seqs[vocab[i]], seqs[vocab[j]]) {
				want++
			}
		}
	}
	if got != want {
		t.Fatalf("indexed edges=%d brute=%d", got, want)
	}
}

func TestRegistryCoversEveryEmittedMetric(t *testing.T) {
	defs := map[string]bool{}
	for _, d := range MetricRegistry() {
		defs[d.ID] = true
	}
	fp := mustAnalyze(t, baseFixture(t))
	for _, m := range fp.Metrics {
		if !defs[m.MetricID] {
			t.Errorf("unregistered metric %s", m.MetricID)
		}
	}
}
