package lineregime

import (
	"math/rand"
	"sort"
	"strings"
	"testing"
)

func tok(s string) []string { return strings.Split(s, "") }

func multiset(toks [][]string) []string {
	out := make([]string, len(toks))
	for i, t := range toks {
		out[i] = strings.Join(t, "")
	}
	sort.Strings(out)
	return out
}

func TestBuildLinesAndEligible(t *testing.T) {
	tb := [][][]string{{tok("ab"), tok("cd")}, {tok("x")}, {tok("a"), tok("b"), tok("c")}}
	lines := BuildLines(tb, []string{"f1r", "f1r", "f1v"}, []string{"A", "A", "B"}, []string{"1", "1", "1"}, true)
	if len(lines) != 3 || lines[2].Folio != "f1v" || lines[0].N() != 2 {
		t.Fatalf("unexpected lines: %+v", lines)
	}
	elig := Eligible(lines, 2)
	if len(elig) != 2 || elig[0].N() < 2 || elig[1].N() < 2 {
		t.Fatalf("eligible filter wrong: %+v", elig)
	}
}

// test 2: same-line pairs never cross a line boundary.
func TestWithinLinePairsNeverCrossBoundary(t *testing.T) {
	l := Line{Tokens: [][]string{tok("aa"), tok("ab"), tok("ac")}}
	pairs := WithinLinePairs(l)
	if len(pairs) != 3 {
		t.Fatalf("want 3 pairs for 3 tokens, got %d", len(pairs))
	}
	for _, p := range pairs {
		if p.I < 0 || p.J >= l.N() || p.I >= p.J {
			t.Fatalf("pair indexes escape the line: %+v", p)
		}
	}
}

// test 3: non-adjacent stratification excludes separation 1.
func TestSeparationBucketExcludesAdjacent(t *testing.T) {
	l := Line{Tokens: [][]string{tok("a"), tok("b"), tok("c"), tok("d")}}
	nonAdjacent := 0
	for _, p := range WithinLinePairs(l) {
		b := SeparationBucket(p.Separation)
		if b == "SEP1" && p.Separation != 1 {
			t.Fatalf("SEP1 bucket leaked separation %d", p.Separation)
		}
		if p.Separation > 1 {
			nonAdjacent++
			if b != "SEP2" && b != "SEP3+" {
				t.Fatalf("non-adjacent pair mis-bucketed: %+v -> %s", p, b)
			}
		}
	}
	if nonAdjacent == 0 {
		t.Fatal("expected at least one non-adjacent pair")
	}
}

func TestShuffleWithinLinePreservesMultiset(t *testing.T) {
	l := Line{Index: 5, Folio: "f1r", Tokens: [][]string{tok("ab"), tok("cd"), tok("ef"), tok("gh")}}
	r := rand.New(rand.NewSource(1))
	out := ShuffleWithinLine(l, r)
	if out.Index != l.Index || out.Folio != l.Folio {
		t.Fatal("shuffle must not change line identity/metadata")
	}
	before, after := multiset(l.Tokens), multiset(out.Tokens)
	for i := range before {
		if before[i] != after[i] {
			t.Fatalf("multiset changed: %v -> %v", before, after)
		}
	}
}

// test 8: line-membership shuffle preserves the global multiset.
func TestShuffleLineMembershipPreservesGlobalMultiset(t *testing.T) {
	lines := []Line{
		{Folio: "f1r", Tokens: [][]string{tok("aa"), tok("bb")}},
		{Folio: "f1r", Tokens: [][]string{tok("cc"), tok("dd"), tok("ee")}},
		{Folio: "f2r", Tokens: [][]string{tok("ff")}},
	}
	r := rand.New(rand.NewSource(2))
	out := ShuffleLineMembership(lines, r)
	if len(out) != len(lines) {
		t.Fatal("line count changed")
	}
	for i := range lines {
		if out[i].N() != lines[i].N() {
			t.Fatalf("line %d length changed: %d -> %d", i, lines[i].N(), out[i].N())
		}
	}
	var before, after [][]string
	for _, l := range lines {
		before = append(before, l.Tokens...)
	}
	for _, l := range out {
		after = append(after, l.Tokens...)
	}
	b, a := multiset(before), multiset(after)
	for i := range b {
		if b[i] != a[i] {
			t.Fatalf("global multiset changed: %v -> %v", b, a)
		}
	}
}

// test 9: within-page line-membership shuffle preserves each page's
// multiset and never leaks tokens across pages.
func TestShuffleLineMembershipWithinPagePreservesPageMultiset(t *testing.T) {
	lines := []Line{
		{Folio: "f1r", Tokens: [][]string{tok("p1a"), tok("p1b")}},
		{Folio: "f1r", Tokens: [][]string{tok("p1c")}},
		{Folio: "f2r", Tokens: [][]string{tok("p2a"), tok("p2b"), tok("p2c")}},
	}
	r := rand.New(rand.NewSource(3))
	out := ShuffleLineMembershipWithinPage(lines, r)
	pageMultiset := func(ls []Line, folio string) []string {
		var flat [][]string
		for _, l := range ls {
			if l.Folio == folio {
				flat = append(flat, l.Tokens...)
			}
		}
		return multiset(flat)
	}
	for _, folio := range []string{"f1r", "f2r"} {
		b, a := pageMultiset(lines, folio), pageMultiset(out, folio)
		if len(b) != len(a) {
			t.Fatalf("page %s size changed", folio)
		}
		for i := range b {
			if b[i] != a[i] {
				t.Fatalf("page %s multiset changed: %v -> %v", folio, b, a)
			}
		}
	}
	for i := range lines {
		if out[i].N() != lines[i].N() || out[i].Folio != lines[i].Folio {
			t.Fatalf("line %d identity/length changed", i)
		}
	}
}

// test 6: global pseudo-lines preserve requested line sizes.
func TestPseudoLineGlobalPreservesSizes(t *testing.T) {
	lines := []Line{{Tokens: [][]string{tok("a"), tok("b"), tok("c")}}, {Tokens: [][]string{tok("d")}}}
	pool := [][]string{tok("x"), tok("y"), tok("z")}
	r := rand.New(rand.NewSource(4))
	out := PseudoLineGlobal(lines, pool, r)
	for i := range lines {
		if out[i].N() != lines[i].N() {
			t.Fatalf("size not preserved at line %d: %d vs %d", i, lines[i].N(), out[i].N())
		}
	}
}

// test 7: same-page pseudo-lines only draw from their own page's pool.
func TestPseudoLineSamePagePreservesPageMembership(t *testing.T) {
	lines := []Line{
		{Folio: "f1r", Tokens: [][]string{tok("a"), tok("b")}},
		{Folio: "f2r", Tokens: [][]string{tok("c"), tok("d"), tok("e")}},
	}
	pagePool := map[string][][]string{
		"f1r": {tok("p1")},
		"f2r": {tok("p2")},
	}
	r := rand.New(rand.NewSource(5))
	out := PseudoLineSamePage(lines, pagePool, r)
	for i, l := range out {
		if l.Folio != lines[i].Folio {
			t.Fatalf("page membership changed at line %d", i)
		}
		want := "p1"
		if l.Folio == "f2r" {
			want = "p2"
		}
		for _, tk := range l.Tokens {
			if strings.Join(tk, "") != want {
				t.Fatalf("line %d drew a token outside its page pool: %v", i, tk)
			}
		}
	}
}

func TestPseudoLineLengthPreserving(t *testing.T) {
	lines := []Line{{Tokens: [][]string{tok("ab"), tok("c")}}}
	pool := map[int][][]string{1: {tok("x")}, 2: {tok("yz")}}
	r := rand.New(rand.NewSource(6))
	out := PseudoLineLengthPreserving(lines, pool, r)
	if len(out[0].Tokens[0]) != 2 || len(out[0].Tokens[1]) != 1 {
		t.Fatalf("length sequence not preserved: %+v", out[0].Tokens)
	}
}

// test 11: shifted-block boundaries are deterministic.
func TestShiftedBlocksDeterministic(t *testing.T) {
	flat := [][]string{tok("a"), tok("b"), tok("c"), tok("d"), tok("e"), tok("f")}
	sizes := []int{2, 2, 2}
	b1 := ShiftedBlocks(flat, sizes, 1)
	b2 := ShiftedBlocks(flat, sizes, 1)
	if len(b1) != len(b2) {
		t.Fatalf("non-deterministic block count: %d vs %d", len(b1), len(b2))
	}
	for i := range b1 {
		if multisetJoin(b1[i]) != multisetJoin(b2[i]) {
			t.Fatalf("non-deterministic block content at %d", i)
		}
	}
	// offset 1 over sizes {2,2,2} on 6 tokens fits two full blocks
	// (positions 1..3 and 3..5); the third would need positions 5..7.
	if len(b1) != 2 || multisetJoin(b1[0]) != "b,c" || multisetJoin(b1[1]) != "d,e" {
		t.Fatalf("unexpected shifted blocks: %+v", b1)
	}
}

func multisetJoin(toks [][]string) string {
	parts := make([]string, len(toks))
	for i, t := range toks {
		parts[i] = strings.Join(t, "")
	}
	return strings.Join(parts, ",")
}

func TestFixedWindows(t *testing.T) {
	flat := make([][]string, 7)
	for i := range flat {
		flat[i] = tok("a")
	}
	out := FixedWindows(flat, 3)
	if len(out) != 2 {
		t.Fatalf("want 2 windows of 3 from 7 tokens, got %d", len(out))
	}
}

func TestCategoricalSampleDegenerate(t *testing.T) {
	c := NewCategorical(map[string]int{"only": 5})
	r := rand.New(rand.NewSource(7))
	for range 20 {
		if got := c.Sample(r); got != "only" {
			t.Fatalf("degenerate distribution returned %q", got)
		}
	}
	empty := NewCategorical(map[string]int{})
	if got := empty.Sample(r); got != "" {
		t.Fatalf("empty distribution should return \"\", got %q", got)
	}
}
