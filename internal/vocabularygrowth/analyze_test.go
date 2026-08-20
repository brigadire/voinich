package vocabularygrowth

import (
	"reflect"
	"testing"
)

func TestAdaptiveCheckpoints(t *testing.T) {
	for _, n := range []int{1, 100, 1000, 8000, 39026, 85000, 120000} {
		cps := Checkpoints(n)
		if len(cps) == 0 || cps[len(cps)-1] != n {
			t.Fatalf("n=%d checkpoints=%v", n, cps)
		}
		for _, c := range cps {
			if c > n {
				t.Fatalf("checkpoint %d exceeds n=%d", c, n)
			}
		}
	}
}
func TestAllIdentical(t *testing.T) {
	toks := make([]string, 100)
	for i := range toks {
		toks[i] = "a"
	}
	r, err := Analyze(toks, Parameters{Checkpoints: []int{1, 50, 100}})
	if err != nil {
		t.Fatal(err)
	}
	for _, p := range r.Growth {
		if p.Vocabulary != 1 {
			t.Fatalf("point=%+v", p)
		}
	}
}
func TestAllUnique(t *testing.T) {
	toks := make([]string, 100)
	for i := range toks {
		toks[i] = string(rune(i + 1))
	}
	r, err := Analyze(toks, Parameters{Checkpoints: []int{1, 50, 100}})
	if err != nil {
		t.Fatal(err)
	}
	for _, p := range r.Growth {
		if p.Vocabulary != p.N || p.Hapax != p.N {
			t.Fatalf("point=%+v", p)
		}
	}
}
func TestOrderChangesTrajectoryButNotFinal(t *testing.T) {
	a := []string{"a", "a", "a", "b", "b", "b"}
	b := []string{"a", "b", "a", "b", "a", "b"}
	ra, _ := Analyze(a, Parameters{Checkpoints: []int{2, 4, 6}, NullPermutations: 0})
	rb, _ := Analyze(b, Parameters{Checkpoints: []int{2, 4, 6}, NullPermutations: 0})
	if ra.Final.Vocabulary != rb.Final.Vocabulary || reflect.DeepEqual(ra.Growth, rb.Growth) {
		t.Fatalf("order should alter trajectory: %+v %+v", ra.Growth, rb.Growth)
	}
}
func TestNullDeterminism(t *testing.T) {
	toks := []string{"a", "a", "b", "c", "d", "e", "f", "g"}
	p := Parameters{Checkpoints: []int{2, 4, 8}, NullPermutations: 7, Seed: 19}
	a, _ := Analyze(toks, p)
	b, _ := Analyze(toks, p)
	if !reflect.DeepEqual(a.Null, b.Null) {
		t.Fatalf("null output not deterministic\n%+v\n%+v", a.Null, b.Null)
	}
}
func TestTinySegments(t *testing.T) {
	r, err := Analyze([]string{"a", "b", "a"}, Parameters{SegmentCounts: []int{4, 8}, NullPermutations: 0})
	if err != nil {
		t.Fatal(err)
	}
	if len(r.Segments) != 0 {
		t.Fatalf("segments too short should be unavailable: %d", len(r.Segments))
	}
}
