package graphemic

import (
	"reflect"
	"testing"
)

func TestTokenizeGraphemes(t *testing.T) {
	want := []string{"@135;", "o", "d", "a", "i", "i", "n", "?"}
	got := TokenizeGraphemes("@135;odaiin?")
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %#v, want %#v", got, want)
	}
	got = TokenizeGraphemes("@x;")
	want = []string{"@", "x", ";"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("malformed code changed: %#v", got)
	}
}

func TestQuestionMarkIsOneUnknownGrapheme(t *testing.T) {
	if got := TokenizeGraphemes("?"); !reflect.DeepEqual(got, []string{"?"}) {
		t.Fatalf("got %#v", got)
	}
}
func TestLevenshteinAndMetrics(t *testing.T) {
	if d := Levenshtein(TokenizeGraphemes("kitten"), TokenizeGraphemes("sitting")); d != 3 {
		t.Fatalf("distance=%d", d)
	}
	d, n, s, p, suf, ld := GraphemeMetrics("@135;odaiin", "@148;odain")
	if d != 2 || n != 2.0/7 || s != 5.0/7 || p != 0 || suf != 2 || ld != 1 {
		t.Fatalf("metrics: %d %v %v %d %d %d", d, n, s, p, suf, ld)
	}
}
func TestPrefixSuffix(t *testing.T) {
	_, _, _, p, s, l := GraphemeMetrics("qokeedy", "qokeey")
	if p != 5 || s != 1 || l != 1 {
		t.Fatalf("prefix=%d suffix=%d length=%d", p, s, l)
	}
}
func TestDeterministicRanking(t *testing.T) {
	a := []Pair{{TokenA: "b", TokenB: "c", DiscoveryScore: .5}, {TokenA: "a", TokenB: "z", DiscoveryScore: .5}, {TokenA: "a", TokenB: "b", DiscoveryScore: .7}}
	sortPairs(a, func(p Pair) float64 { return p.DiscoveryScore })
	got := []string{a[0].TokenA + a[0].TokenB, a[1].TokenA + a[1].TokenB, a[2].TokenA + a[2].TokenB}
	want := []string{"ab", "az", "bc"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v", got)
	}
}
