package corpustransform

import (
	"reflect"
	"strings"
	"testing"
)

func TestHomophonicPreservesTokenCount(t *testing.T) {
	in := words("the quick brown fox the lazy dog the fox")
	mapping, err := BuildMapping(in, HomophonicParams{Model: HomophoneModelFixed, Homophones: 4, Selection: SelectionUniform, Seed: 1})
	if err != nil {
		t.Fatal(err)
	}
	out := Encode(in, mapping, 1)
	if len(out) != len(in) {
		t.Fatalf("got %d tokens, want %d", len(out), len(in))
	}
}

func TestHomophonicMappingDomainsDoNotOverlap(t *testing.T) {
	in := words("alpha beta gamma delta alpha beta")
	mapping, err := BuildMapping(in, HomophonicParams{Model: HomophoneModelFixed, Homophones: 4, Selection: SelectionUniform, Seed: 1})
	if err != nil {
		t.Fatal(err)
	}
	if got := MappingCollisions(mapping); got != 0 {
		t.Fatalf("mapping collisions = %d, want 0", got)
	}
}

func TestHomophonicOutputMapsToExactlyOnePlaintextToken(t *testing.T) {
	in := words("alpha beta gamma delta alpha beta alpha gamma")
	mapping, err := BuildMapping(in, HomophonicParams{Model: HomophoneModelFixed, Homophones: 3, Selection: SelectionUniform, Seed: 2})
	if err != nil {
		t.Fatal(err)
	}
	out := Encode(in, mapping, 2)
	reverse := make(map[string]string)
	for _, t := range mapping.Vocabulary {
		for _, e := range mapping.Entries[t] {
			if prev, ok := reverse[e.CipherToken]; ok && prev != t {
				panic("cipher token owned by two plaintext tokens")
			}
			reverse[e.CipherToken] = t
		}
	}
	for _, c := range out {
		if _, ok := reverse[c]; !ok {
			t.Fatalf("output token %q has no plaintext owner", c)
		}
	}
}

func TestHomophonicH1IsMonoalphabetic(t *testing.T) {
	in := words("alpha beta alpha gamma alpha beta beta")
	mapping, err := BuildMapping(in, HomophonicParams{Model: HomophoneModelFixed, Homophones: 1, Selection: SelectionUniform, Seed: 1})
	if err != nil {
		t.Fatal(err)
	}
	out := Encode(in, mapping, 1)
	seen := make(map[string]string)
	for i, plain := range in {
		if prev, ok := seen[plain]; ok && prev != out[i] {
			t.Fatalf("H=1 substitution not monoalphabetic: %q mapped to both %q and %q", plain, prev, out[i])
		}
		seen[plain] = out[i]
	}
}

func TestHomophonicVocabularyBounds(t *testing.T) {
	in := words("a b c d e f g h a b c d")
	for _, h := range []int{2, 4, 8} {
		mapping, err := BuildMapping(in, HomophonicParams{Model: HomophoneModelFixed, Homophones: h, Selection: SelectionUniform, Seed: 1})
		if err != nil {
			t.Fatal(err)
		}
		wantCipherVocab := len(mapping.Vocabulary) * h
		gotCipherVocab := 0
		for _, tok := range mapping.Vocabulary {
			gotCipherVocab += len(mapping.Entries[tok])
		}
		if gotCipherVocab != wantCipherVocab {
			t.Fatalf("H=%d: cipher vocab = %d, want %d", h, gotCipherVocab, wantCipherVocab)
		}
	}
}

func TestHomophonicSameSeedByteIdentical(t *testing.T) {
	in := words(strings.Repeat("alpha beta gamma delta epsilon ", 20))
	p := HomophonicParams{Model: HomophoneModelFixed, Homophones: 4, Selection: SelectionWeighted, Seed: 123}
	m1, err := BuildMapping(in, p)
	if err != nil {
		t.Fatal(err)
	}
	out1 := Encode(in, m1, p.Seed)
	m2, err := BuildMapping(in, p)
	if err != nil {
		t.Fatal(err)
	}
	out2 := Encode(in, m2, p.Seed)
	if !reflect.DeepEqual(out1, out2) {
		t.Fatal("same seed produced different output")
	}
}

func TestHomophonicDifferentSeedsDifferAssignment(t *testing.T) {
	in := words(strings.Repeat("alpha beta gamma delta epsilon ", 20))
	base := HomophonicParams{Model: HomophoneModelFixed, Homophones: 4, Selection: SelectionUniform}
	m, err := BuildMapping(in, base)
	if err != nil {
		t.Fatal(err)
	}
	out1 := Encode(in, m, 1)
	out2 := Encode(in, m, 2)
	if reflect.DeepEqual(out1, out2) {
		t.Fatal("different seeds produced identical occurrence assignment for a non-trivial corpus")
	}
}

func TestHomophonicUniformDistributionSanity(t *testing.T) {
	n := 20000
	tokens := make([]string, n)
	for i := range tokens {
		tokens[i] = "w"
	}
	p := HomophonicParams{Model: HomophoneModelFixed, Homophones: 4, Selection: SelectionUniform, Seed: 7}
	mapping, err := BuildMapping(tokens, p)
	if err != nil {
		t.Fatal(err)
	}
	out := Encode(tokens, mapping, p.Seed)
	counts := make(map[string]int)
	for _, c := range out {
		counts[c]++
	}
	if len(counts) != 4 {
		t.Fatalf("got %d distinct homophones, want 4", len(counts))
	}
	for c, n := range counts {
		frac := float64(n) / float64(len(out))
		if frac < 0.20 || frac > 0.30 {
			t.Fatalf("homophone %q occurrence fraction %.4f outside uniform sanity band [0.20,0.30]", c, frac)
		}
	}
}

func TestHomophonicMappingSerializationDeterministic(t *testing.T) {
	in := words("delta charlie alpha bravo delta alpha")
	p := HomophonicParams{Model: HomophoneModelFixed, Homophones: 2, Selection: SelectionWeighted, Seed: 1}
	m1, err := BuildMapping(in, p)
	if err != nil {
		t.Fatal(err)
	}
	m2, err := BuildMapping(in, p)
	if err != nil {
		t.Fatal(err)
	}
	tsv1 := MarshalMappingTSV(m1)
	tsv2 := MarshalMappingTSV(m2)
	if string(tsv1) != string(tsv2) {
		t.Fatal("mapping TSV serialization is not deterministic")
	}
	lines := strings.Split(strings.TrimRight(string(tsv1), "\n"), "\n")
	if lines[0] != "plaintext_token\tcipher_token\tprobability" {
		t.Fatalf("unexpected header: %q", lines[0])
	}
	for i := 2; i < len(lines); i++ {
		prevTok := strings.Split(lines[i-1], "\t")[0]
		curTok := strings.Split(lines[i], "\t")[0]
		if curTok < prevTok {
			t.Fatalf("mapping TSV not sorted by plaintext token: %q before %q", prevTok, curTok)
		}
	}
}

func TestHomophonicNoPlaintextLeakageInCipherNames(t *testing.T) {
	in := words("secret hidden plaintext token corpus")
	mapping, err := BuildMapping(in, HomophonicParams{Model: HomophoneModelFixed, Homophones: 3, Selection: SelectionUniform, Seed: 1})
	if err != nil {
		t.Fatal(err)
	}
	for _, plain := range mapping.Vocabulary {
		for _, e := range mapping.Entries[plain] {
			if strings.Contains(e.CipherToken, plain) {
				t.Fatalf("cipher token %q leaks plaintext token %q", e.CipherToken, plain)
			}
			if e.CipherToken == plain {
				t.Fatalf("cipher token equals plaintext token %q", plain)
			}
		}
	}
}

func TestHomophonicRoundTrip(t *testing.T) {
	in := words("the quick brown fox jumps over the lazy dog the fox runs")
	for _, sel := range []string{SelectionUniform, SelectionWeighted} {
		for _, h := range []int{1, 2, 4} {
			p := HomophonicParams{Model: HomophoneModelFixed, Homophones: h, Selection: sel, Seed: 11}
			mapping, err := BuildMapping(in, p)
			if err != nil {
				t.Fatal(err)
			}
			out := Encode(in, mapping, p.Seed)
			back, err := Decode(out, mapping)
			if err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(back, in) {
				t.Fatalf("selection=%s H=%d: decode(transform(T)) != T", sel, h)
			}
		}
	}
}

func TestHomophonicFrequencyModelIsBacklog(t *testing.T) {
	_, err := BuildMapping(words("a b c"), HomophonicParams{Model: HomophoneModelFrequency, Homophones: 4, Selection: SelectionUniform, Seed: 1})
	if err == nil {
		t.Fatal("expected frequency model to be rejected as not implemented")
	}
}
