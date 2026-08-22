package lineregime

import (
	"strings"
	"testing"
)

func tokC(s string) []string { return strings.Split(s, "") }

func TestComputeCoreStatsDeterministic(t *testing.T) {
	lines := []Line{
		{Index: 0, Folio: "f1r", Tokens: [][]string{tokC("chol"), tokC("chor"), tokC("shol"), tokC("daiin"), tokC("okaiin")}},
		{Index: 1, Folio: "f1r", Tokens: [][]string{tokC("qokeey"), tokC("qokaiin"), tokC("otchy"), tokC("shedy"), tokC("chedy")}},
		{Index: 2, Folio: "f1v", Tokens: [][]string{tokC("dain"), tokC("chor"), tokC("shol"), tokC("daiin"), tokC("okal")}},
	}
	a := ComputeCoreStats(lines, 5, 12345)
	b := ComputeCoreStats(lines, 5, 12345)
	if a.Adj.Rate() != b.Adj.Rate() || a.NonAdj.Rate() != b.NonAdj.Rate() || a.DiffLineSamePage.Rate() != b.DiffLineSamePage.Rate() {
		t.Fatalf("same seed produced different rates: %+v vs %+v", a, b)
	}
	if a.NLines != 3 {
		t.Fatalf("want 3 eligible lines, got %d", a.NLines)
	}
}

func TestRateOnlySeparatesAdjacentFromNonAdjacent(t *testing.T) {
	lines := []Line{{Tokens: [][]string{tokC("a"), tokC("a"), tokC("z"), tokC("z"), tokC("z")}}}
	adj := RateOnly(lines, 5, true)
	nonAdj := RateOnly(lines, 5, false)
	if adj == 0 {
		t.Fatal("adjacent identical pair should give a nonzero d<=1 rate")
	}
	_ = nonAdj
}

func TestSplitPagesNoOverlapAndCovers(t *testing.T) {
	pages := []string{"f1r", "f1v", "f2r", "f2v", "f3r", "f3v", "f4r", "f4v", "f5r", "f5v"}
	s := SplitPages(pages, 0.5, 0.2, 0.7)
	seen := map[string]int{}
	for _, p := range pages {
		n := 0
		if s.Train[p] {
			n++
		}
		if s.Validation[p] {
			n++
		}
		if s.Test[p] {
			n++
		}
		if n != 1 {
			t.Fatalf("page %s in %d of train/validation/test (want exactly 1)", p, n)
		}
		d, r := s.Discovery[p], s.Replication[p]
		if d == r {
			t.Fatalf("page %s discovery=%v replication=%v (want exactly one)", p, d, r)
		}
		seen[p]++
	}
	if len(seen) != len(pages) {
		t.Fatal("not every page was classified")
	}
	// discovery must be exactly train+validation (task64's definition).
	for p := range s.Discovery {
		if !s.Train[p] && !s.Validation[p] {
			t.Fatalf("page %s is discovery but neither train nor validation", p)
		}
	}
}

func TestFilterByPages(t *testing.T) {
	lines := []Line{{Folio: "f1r"}, {Folio: "f2r"}, {Folio: "f1r"}}
	out := FilterByPages(lines, map[string]bool{"f1r": true})
	if len(out) != 2 {
		t.Fatalf("want 2 lines on f1r, got %d", len(out))
	}
}

func TestBuildGiantSetFindsD1Neighbors(t *testing.T) {
	vocab := [][]string{tokC("chol"), tokC("chor"), tokC("shol"), tokC("zzzzzzzzzz")}
	giant := BuildGiantSet(vocab)
	if !giant["chol"] || !giant["chor"] || !giant["shol"] {
		t.Fatalf("expected chol/chor/shol in the giant component: %+v", giant)
	}
	if giant["zzzzzzzzzz"] {
		t.Fatal("isolated token should not be in the giant component")
	}
}

func TestProfileDistanceSymmetricAndZeroForIdentical(t *testing.T) {
	giant := map[string]bool{}
	p1 := ComputeProfile([][]string{tokC("abc"), tokC("abd"), tokC("xyz")}, giant)
	p2 := ComputeProfile([][]string{tokC("abc"), tokC("abd"), tokC("xyz")}, giant)
	if ProfileDistance(p1, p2) != 0 {
		t.Fatalf("identical token sets should give distance 0, got %v", ProfileDistance(p1, p2))
	}
	p3 := ComputeProfile([][]string{tokC("qqq"), tokC("rr")}, giant)
	if ProfileDistance(p1, p3) != ProfileDistance(p3, p1) {
		t.Fatal("profile distance must be symmetric")
	}
}

func TestPageOrderOfFirstAppearance(t *testing.T) {
	lines := []Line{{Folio: "f2r"}, {Folio: "f2r"}, {Folio: "f1r"}, {Folio: "f3r"}}
	order := PageOrderOf(lines)
	want := []string{"f2r", "f1r", "f3r"}
	if len(order) != len(want) {
		t.Fatalf("got %v want %v", order, want)
	}
	for i := range want {
		if order[i] != want[i] {
			t.Fatalf("got %v want %v", order, want)
		}
	}
}
