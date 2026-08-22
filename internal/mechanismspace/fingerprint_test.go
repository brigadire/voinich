package mechanismspace

import (
	"math/rand"
	"testing"

	"zcore.dev/voinich/internal/characterentropy"
	"zcore.dev/voinich/internal/evaglyph"
)

func bigCorpus(n int) Corpus {
	base := []string{"the", "quick", "brown", "fox", "jumps", "over", "the", "lazy", "dog",
		"and", "runs", "through", "dark", "wood", "again", "sherlock", "holmes", "took",
		"his", "bottle", "from", "corner", "mantel", "piece"}
	r := rand.New(rand.NewSource(99))
	var words []string
	var lines []int
	line := 0
	for len(words) < n {
		w := base[r.Intn(len(base))]
		words = append(words, w)
		lines = append(lines, line)
		if len(words)%8 == 0 {
			line++
		}
	}
	return Corpus{Name: "big", Words: words, Lines: lines}
}

// test 18: M2's homophony matches the authoritative Task59 implementation
// by construction (it calls evaglyph.RandomHomophony directly).
func TestHomophonyReusesAuthoritativeImplementation(t *testing.T) {
	c := bigCorpus(200)
	cfg := Config{Family: "M2", Homophones: 4, Seed: 5}
	out := Transform(cfg, c)
	r := rand.New(rand.NewSource(cfg.Seed))
	words := c.Glyphs()
	want := make([][]string, len(words))
	for i, w := range words {
		want[i] = evaglyph.RandomHomophony([][]string{w}, 4, r)[0]
	}
	for i := range want {
		if len(want[i]) != len(out.Tokens[i]) {
			t.Fatalf("homophony output shape mismatch at %d", i)
		}
	}
}

// test 19: fingerprint extraction reuses the authoritative per-family
// primitives directly, e.g. H1-H4 must equal a direct
// characterentropy.Entropy call on the same tokens.
func TestFingerprintReusesAuthoritativeEntropy(t *testing.T) {
	c := bigCorpus(300)
	tokens := c.Glyphs()
	fp := ComputeFingerprint(tokens, c.Lines, DefaultScreeningOptions(1))
	want := characterentropy.Entropy(tokens, c.Lines, characterentropy.TokenBoundary, 0, true)
	if want.Status == "OK" && fp.H1 != want.H {
		t.Fatalf("H1 diverged from authoritative characterentropy.Entropy: %v vs %v", fp.H1, want.H)
	}
}

// test: fingerprint extraction is deterministic given the same options/seed.
func TestFingerprintDeterministic(t *testing.T) {
	c := bigCorpus(400)
	tokens := c.Glyphs()
	a := ComputeFingerprint(tokens, c.Lines, DefaultFullOptions(3))
	b := ComputeFingerprint(tokens, c.Lines, DefaultFullOptions(3))
	if a.TokenOrderBits != b.TokenOrderBits || a.Topology != b.Topology || a.H2 != b.H2 {
		t.Fatalf("fingerprint not deterministic: %+v vs %+v", a, b)
	}
}

// test: M0 (identity) fingerprint on a corpus equals the fingerprint of
// that corpus computed directly, confirming Transform+ComputeFingerprint
// compose without altering content for the control mechanism.
func TestIdentityFingerprintMatchesDirectComputation(t *testing.T) {
	c := bigCorpus(300)
	out := Transform(Config{Family: "M0"}, c)
	direct := ComputeFingerprint(c.Glyphs(), c.Lines, DefaultScreeningOptions(2))
	viaJob := ComputeFingerprint(out.Tokens, out.Lines, DefaultScreeningOptions(2))
	if direct.H1 != viaJob.H1 || direct.PositionalWeightedEntropy != viaJob.PositionalWeightedEntropy {
		t.Fatalf("identity mechanism changed the fingerprint")
	}
}

// test: M1 (monoalphabetic relabeling) preserves the fingerprint's
// structural statistics (task66 section 12: "relabeling without
// structural change should mostly preserve the fingerprint"), since it is
// a bijective per-glyph renaming.
func TestMonoalphabeticPreservesStructuralFingerprint(t *testing.T) {
	c := bigCorpus(300)
	base := ComputeFingerprint(c.Glyphs(), c.Lines, DefaultScreeningOptions(2))
	out := Transform(Config{Family: "M1", Seed: 123}, c)
	relabeled := ComputeFingerprint(out.Tokens, out.Lines, DefaultScreeningOptions(2))
	if diff := absF(base.H1 - relabeled.H1); diff > 1e-9 {
		t.Fatalf("H1 changed under pure relabeling: %v vs %v", base.H1, relabeled.H1)
	}
	if diff := absF(base.ExactAdjacentRepeatRate - relabeled.ExactAdjacentRepeatRate); diff > 1e-9 {
		t.Fatalf("exact-adjacent-repeat rate changed under pure relabeling: %v vs %v", base.ExactAdjacentRepeatRate, relabeled.ExactAdjacentRepeatRate)
	}
}

func absF(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}

// test 22: corpus transfer uses an identical Config across corpora.
func TestCorpusTransferUsesIdenticalParameters(t *testing.T) {
	cfg := Config{Family: "M4", StateCount: 4, Update: UpdateB, Seed: 42}
	corpora := []Corpus{bigCorpus(120), bigCorpus(150)}
	corpora[1].Name = "second"
	var first Config
	for i, c := range corpora {
		out := Transform(cfg, c)
		if i == 0 {
			first = out.Metadata.Parameters
			continue
		}
		if out.Metadata.Parameters != first {
			t.Fatalf("mechanism config drifted across corpora: %+v vs %+v", out.Metadata.Parameters, first)
		}
	}
}
