package orientation

import (
	"bytes"
	"math/rand"
	"testing"
)

func TestTransformModesAndInvolution(t *testing.T) {
	input := []byte("a b c\nde fg h\n\nодин два\n")
	wants := map[string]string{
		TokenReverse: "c b a\nh fg de\n\nдва один\n",
		GlyphReverse: "a b c\ned gf h\n\nнидо авд\n",
		FullReverse:  "c b a\nh gf ed\n\nавд нидо\n",
	}
	for mode, want := range wants {
		t.Run(mode, func(t *testing.T) {
			got, before, after, err := Transform(input, mode)
			if err != nil {
				t.Fatal(err)
			}
			if string(got) != want {
				t.Fatalf("output = %q, want %q", got, want)
			}
			if before.Tokens != after.Tokens || before.Lines != after.Lines {
				t.Fatalf("structural invariants changed: before=%+v after=%+v", before, after)
			}
			roundTrip, _, _, err := Transform(got, mode)
			if err != nil {
				t.Fatal(err)
			}
			if !bytes.Equal(roundTrip, input) {
				t.Fatalf("double transform = %q, want %q", roundTrip, input)
			}
		})
	}
}

func TestTokenReversePreservesTokenMultiset(t *testing.T) {
	input := []byte("long x long\nаб вг\n")
	_, before, after, err := Transform(input, TokenReverse)
	if err != nil {
		t.Fatal(err)
	}
	if !equalCounts(before.Counts, after.Counts) || before.UniqueTokens != after.UniqueTokens {
		t.Fatalf("token multiset changed: before=%v after=%v", before.Counts, after.Counts)
	}
}

func TestTransformRejectsNonCanonicalInput(t *testing.T) {
	if _, _, _, err := Transform([]byte("a, b\n"), TokenReverse); err == nil {
		t.Fatal("non-canonical corpus accepted")
	}
}

func TestTransformRandomizedInvolutionAndTokenInvariants(t *testing.T) {
	rng := rand.New(rand.NewSource(51))
	vocabulary := []string{"a", "bb", "ccc", "ёж", "мир", "漢字"}
	for iteration := 0; iteration < 100; iteration++ {
		var corpus []byte
		for line := 0; line < 1+rng.Intn(10); line++ {
			if line > 0 && rng.Intn(5) == 0 {
				corpus = append(corpus, '\n')
				continue
			}
			for token := 0; token < 1+rng.Intn(8); token++ {
				if token > 0 {
					corpus = append(corpus, ' ')
				}
				corpus = append(corpus, vocabulary[rng.Intn(len(vocabulary))]...)
			}
			corpus = append(corpus, '\n')
		}
		for _, mode := range []string{TokenReverse, GlyphReverse, FullReverse} {
			transformed, before, after, err := Transform(corpus, mode)
			if err != nil {
				t.Fatalf("iteration %d mode %s: %v", iteration, mode, err)
			}
			roundTrip, _, _, err := Transform(transformed, mode)
			if err != nil {
				t.Fatal(err)
			}
			if !bytes.Equal(roundTrip, corpus) {
				t.Fatalf("iteration %d mode %s was not involutory", iteration, mode)
			}
			if mode == TokenReverse && !equalCounts(before.Counts, after.Counts) {
				t.Fatalf("iteration %d token multiset changed", iteration)
			}
		}
	}
}
