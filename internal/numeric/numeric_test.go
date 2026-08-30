package numeric

import (
	"math"
	"reflect"
	"testing"
)

func tiny() Corpus {
	return Corpus{Name: "x", Alphabet: []byte{'a', 'b', 'c'}, UniqueTokenCount: 4, Tokens: []Token{{Text: "ab", Glyphs: []byte("ab"), Line: 0, IndexInLine: 0, Folio: "f1"}, {Text: "ac", Glyphs: []byte("ac"), Line: 0, IndexInLine: 1, Folio: "f1"}, {Text: "bc", Glyphs: []byte("bc"), Line: 0, IndexInLine: 2, Folio: "f1"}, {Text: "a", Glyphs: []byte("a"), Line: 1, IndexInLine: 0, Folio: "f2"}}}
}
func TestValues(t *testing.T) {
	v, _ := Values(tiny(), BaselineMapping(3))
	want := []float64{1, 2, 5, 0}
	if !reflect.DeepEqual(v, want) {
		t.Fatalf("got %v want %v", v, want)
	}
}
func TestSubstitutionIdentity(t *testing.T) {
	if got := Compute(tiny(), []int{2, 0, 1}).EditSubstitutionConsistency; got != 1 {
		t.Fatalf("got %v", got)
	}
}
func TestControlsDeterministic(t *testing.T) {
	a := Control(tiny(), "C1_WITHIN_TOKEN_GLYPH_SHUFFLE", 7)
	b := Control(tiny(), "C1_WITHIN_TOKEN_GLYPH_SHUFFLE", 7)
	if !reflect.DeepEqual(a.Tokens, b.Tokens) {
		t.Fatal("control is not deterministic")
	}
}
func TestLeadingZeroCollision(t *testing.T) {
	c := tiny()
	m := []int{0, 1, 2}
	x := Compute(c, m)
	if x.LeadingZeroFraction <= 0 || math.IsNaN(x.Score) {
		t.Fatalf("bad metric: %+v", x)
	}
}
