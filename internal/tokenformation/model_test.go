package tokenformation

import (
	"math/rand"
	"testing"
)

func TestModelsAndDeterminism(t *testing.T) {
	c := Corpus{Tokens: [][]string{{"a", "b"}, {"a", "c", "b"}}, Lines: []int{0, 1}}
	for _, k := range []Kind{IID, PosIID, Markov1, Markov2, PosMarkov1} {
		m := Fit(c, k, .1)
		a := m.Generate(10, rand.New(rand.NewSource(3)))
		b := m.Generate(10, rand.New(rand.NewSource(3)))
		if len(a) != len(b) || a[0][0] != b[0][0] {
			t.Fatal(k, "not deterministic")
		}
		if m.CrossEntropy(c.Tokens) <= 0 {
			t.Fatal(k, "bad CE")
		}
	}
}
