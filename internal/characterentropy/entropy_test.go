package characterentropy

import (
	"math/rand"
	"testing"
)

func TestEstimatorControls(t *testing.T) {
	if got := Entropy([][]string{{"a", "a", "a", "a"}}, nil, Continuous, 1, false).H; got != 0 {
		t.Fatalf("constant: %v", got)
	}
	if got := Entropy([][]string{{"a", "b", "a", "b", "a", "b"}}, nil, Continuous, 1, false).H; got > 1e-12 {
		t.Fatalf("alternating: %v", got)
	}
	// Four symbols in a deterministic cycle have h1=2 and h2=0.
	if got := Entropy([][]string{{"a", "b", "c", "d", "a", "b", "c", "d"}}, nil, Continuous, 0, false).H; got < 1.9 {
		t.Fatalf("h1: %v", got)
	}
	if got := Entropy([][]string{{"a", "b", "c", "d", "a", "b", "c", "d"}}, nil, Continuous, 1, false).H; got > 1e-12 {
		t.Fatalf("cycle conditional: %v", got)
	}
	a := [][]string{{"a", "b"}, {"c"}}
	b := TokenShuffle(a, rand.New(rand.NewSource(1)))
	if len(b) != 2 || len(b[0]) != 2 {
		t.Fatal("shuffle changed forms")
	}
}
