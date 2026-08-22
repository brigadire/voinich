package tokenrepetition

import (
	"math/rand"
	"testing"
)

func newTestRand(seed int64) *rand.Rand { return rand.New(rand.NewSource(seed)) }

func glyphs(s string) []string {
	out := make([]string, len(s))
	for i, r := range s {
		out[i] = string(r)
	}
	return out
}

func TestLevenshteinKnownCases(t *testing.T) {
	cases := []struct{ a, b string; want int }{
		{"abc", "abd", 1},
		{"abc", "abcd", 1},
		{"abcd", "abc", 1},
		{"abc", "abc", 0},
		{"abc", "xyz", 3},
		{"", "abc", 3},
	}
	for _, c := range cases {
		if got := LevenshteinGlyphs(glyphs(c.a), glyphs(c.b)); got != c.want {
			t.Fatalf("Levenshtein(%q,%q)=%d, want %d", c.a, c.b, got, c.want)
		}
	}
}

func TestClassifyDistanceOneSubstitution(t *testing.T) {
	op, pos, src, tgt, ok := ClassifyDistanceOne(glyphs("abc"), glyphs("abd"))
	if !ok || op != "SUBSTITUTION" || pos != 2 || src != "c" || tgt != "d" {
		t.Fatalf("got op=%s pos=%d src=%s tgt=%s ok=%v", op, pos, src, tgt, ok)
	}
}
func TestClassifyDistanceOneInsertion(t *testing.T) {
	op, pos, _, tgt, ok := ClassifyDistanceOne(glyphs("abc"), glyphs("abcd"))
	if !ok || op != "INSERTION" || pos != 3 || tgt != "d" {
		t.Fatalf("got op=%s pos=%d tgt=%s ok=%v", op, pos, tgt, ok)
	}
	// insertion in the middle
	op, pos, _, tgt, ok = ClassifyDistanceOne(glyphs("ac"), glyphs("abc"))
	if !ok || op != "INSERTION" || pos != 1 || tgt != "b" {
		t.Fatalf("got op=%s pos=%d tgt=%s ok=%v", op, pos, tgt, ok)
	}
}
func TestClassifyDistanceOneDeletion(t *testing.T) {
	op, pos, src, _, ok := ClassifyDistanceOne(glyphs("abcd"), glyphs("abc"))
	if !ok || op != "DELETION" || pos != 3 || src != "d" {
		t.Fatalf("got op=%s pos=%d src=%s ok=%v", op, pos, src, ok)
	}
}
func TestClassifyDistanceOneRejectsNonDistanceOne(t *testing.T) {
	if _, _, _, _, ok := ClassifyDistanceOne(glyphs("abc"), glyphs("xyz")); ok {
		t.Fatal("expected ok=false for a distance-3 pair")
	}
}
func TestPositionClassThirds(t *testing.T) {
	if PositionClass(0, 6) != "BEGIN" {
		t.Fatal("expected BEGIN")
	}
	if PositionClass(3, 6) != "MIDDLE" {
		t.Fatal("expected MIDDLE")
	}
	if PositionClass(5, 6) != "END" {
		t.Fatal("expected END")
	}
}

func TestExactRunsNoDoubleCounting(t *testing.T) {
	runs := ExactRuns([]string{"a", "a", "a", "a", "a", "b", "c", "c"}, nil)
	if len(runs) != 2 {
		t.Fatalf("expected 2 runs, got %d: %+v", len(runs), runs)
	}
	if runs[0].RunLength != 5 || runs[0].Token != "a" {
		t.Fatalf("expected maximal run of 5 a's, got %+v", runs[0])
	}
	if runs[1].RunLength != 2 || runs[1].Token != "c" {
		t.Fatalf("expected run of 2 c's, got %+v", runs[1])
	}
}

func TestExactRunsSingleRepeatIsOneEvent(t *testing.T) {
	runs := ExactRuns([]string{"x", "a", "a", "y"}, nil)
	if len(runs) != 1 || runs[0].RunLength != 2 {
		t.Fatalf("expected exactly one run of length 2, got %+v", runs)
	}
}

func TestAdjacentRepetitionR2(t *testing.T) {
	st := AdjacentRepetition([]string{"a", "a", "b", "c", "c", "c"}, nil, 10)
	// pairs: (a,a)=repeat (b,c)= no (c,c)=repeat (c,c)=repeat -> valid=5, repeats=3
	if st.ValidPairs != 5 {
		t.Fatalf("expected 5 valid pairs, got %d", st.ValidPairs)
	}
	if st.RepeatPairs != 3 {
		t.Fatalf("expected 3 repeat pairs, got %d", st.RepeatPairs)
	}
	if st.Tokens["c"].MaximumRun != 3 {
		t.Fatalf("expected c's max run = 3, got %+v", st.Tokens["c"])
	}
}

func TestWithinLineShufflePreservesLineComposition(t *testing.T) {
	tokens := []string{"a", "b", "c", "d", "e"}
	lineOf := []int{0, 0, 0, 1, 1}
	out := WithinLineShuffle(tokens, lineOf, newTestRand(1))
	if len(out) != len(tokens) {
		t.Fatal("length changed")
	}
	line0 := map[string]bool{out[0]: true, out[1]: true, out[2]: true}
	for _, want := range []string{"a", "b", "c"} {
		if !line0[want] {
			t.Fatalf("line 0 composition changed: %v", out[:3])
		}
	}
}

func TestGlobalShufflePreservesFrequencies(t *testing.T) {
	tokens := []string{"a", "a", "b", "c", "c", "c"}
	out := GlobalShuffle(tokens, newTestRand(1))
	freqBefore, freqAfter := map[string]int{}, map[string]int{}
	for _, t := range tokens {
		freqBefore[t]++
	}
	for _, t := range out {
		freqAfter[t]++
	}
	for k, v := range freqBefore {
		if freqAfter[k] != v {
			t.Fatalf("frequency of %q changed: %d != %d", k, freqAfter[k], v)
		}
	}
}
