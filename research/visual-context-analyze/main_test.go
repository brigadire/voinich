package main

import (
	"encoding/json"
	"math"
	"math/rand"
	"reflect"
	"testing"
)

func TestLineProfileJSONTags(t *testing.T) {
	var p lineProfile
	if err := json.Unmarshal([]byte(`{"folio":"f1r","token_count":7,"exact_repetition_rate":0.25,"first_token":"a"}`), &p); err != nil {
		t.Fatal(err)
	}
	if p.Folio != "f1r" || p.TokenCount != 7 || p.ExactRepetitionRate != .25 || p.FirstToken != "a" {
		t.Fatalf("bad decode: %+v", p)
	}
}

func TestLeafID(t *testing.T) {
	for in, want := range map[string]string{"f67r1": "f67r", "f68v3": "f68v", "f1r": "f1r", "fRos": "fRos"} {
		if got := leafID(in); got != want {
			t.Fatalf("leafID(%q)=%q want %q", in, got, want)
		}
	}
}

func TestPermutationPreservesLabelsAndLeafShapes(t *testing.T) {
	labels := []string{"A", "B", "C", "D", "E", "F"}
	groups := [][]int{{0, 1}, {2, 3}, {4}, {5}}
	got := permuteByLeafShape(labels, groups, rand.New(rand.NewSource(1)))
	a, b := append([]string(nil), got[:4]...), append([]string(nil), got[4:]...)
	sortStrings(a)
	sortStrings(b)
	if !reflect.DeepEqual(a, []string{"A", "B", "C", "D"}) || !reflect.DeepEqual(b, []string{"E", "F"}) {
		t.Fatalf("shape buckets not preserved: %v", got)
	}
}

func TestWithinQuirePermutationPreservesQuireLabels(t *testing.T) {
	ps := []*page{{Quire: "A"}, {Quire: "A"}, {Quire: "B"}, {Quire: "B"}}
	labels := []string{"H", "T", "B", "S"}
	groups := [][]int{{0}, {1}, {2}, {3}}
	got := permuteLabelsWithinQuireByLeafShape(labels, ps, groups, rand.New(rand.NewSource(3)))
	a, b := append([]string(nil), got[:2]...), append([]string(nil), got[2:]...)
	sortStrings(a)
	sortStrings(b)
	if !reflect.DeepEqual(a, []string{"H", "T"}) || !reflect.DeepEqual(b, []string{"B", "S"}) {
		t.Fatalf("quire composition changed: %v", got)
	}
}

func sortStrings(x []string) {
	for i := range x {
		for j := i + 1; j < len(x); j++ {
			if x[j] < x[i] {
				x[i], x[j] = x[j], x[i]
			}
		}
	}
}

func TestProjectRankAndFit(t *testing.T) {
	x := [][]float64{{1, 0, 0}, {1, 1, 2}, {1, 2, 4}, {1, 3, 6}}
	y := []float64{1, 3, 5, 7}
	fit, rank := project(x, y)
	if rank != 2 {
		t.Fatalf("rank=%d", rank)
	}
	if sse(y, fit) > 1e-20 {
		t.Fatalf("SSE=%g", sse(y, fit))
	}
}

func TestBH(t *testing.T) {
	got := bh([]float64{.01, .04, .03})
	want := []float64{.03, .04, .04}
	for i := range got {
		if math.Abs(got[i]-want[i]) > 1e-12 {
			t.Fatalf("got %v want %v", got, want)
		}
	}
}
