package main

import (
	"fmt"
	"math/rand"

	"zcore.dev/voinich/internal/evaglyph"
	"zcore.dev/voinich/internal/tokenrepetition"
)

type homophonicSpec struct {
	Series, Model, PlaintextPath, CipherPath, AllocationPath string
	H                                                        int
}

func homophonicSpecs() []homophonicSpec {
	tr := "data_test/transformed/"
	doyle := "data_test/pg2097-2.txt"
	longfellow := "data_test/pg30795-mod.txt"
	astafiev := "data_test/astafiev-1000-culinar-receipts-prepared.txt"
	var specs []homophonicSpec
	for _, h := range []int{2, 4, 6, 8} {
		specs = append(specs, homophonicSpec{"Doyle", "global-uniform", doyle,
			fmt.Sprintf("%sdoyle__homophonic__h%03d__uniform__seed001.txt", tr, h), "", h})
	}
	for _, h := range []int{4, 6, 8} {
		specs = append(specs, homophonicSpec{"Doyle", "global-weighted", doyle,
			fmt.Sprintf("%sdoyle__homophonic__h%03d__weighted__seed001.txt", tr, h), "", h})
	}
	for _, h := range []int{4, 6, 8} {
		cipher := fmt.Sprintf("%sdoyle__homophonic-frequency-v1__hmax%03d__uniform__seed001.txt", tr, h)
		specs = append(specs, homophonicSpec{"Doyle", "frequency-v1-uniform", doyle, cipher, cipher + ".homophone_allocation.tsv", h})
	}
	for _, h := range []int{2, 4, 8} {
		specs = append(specs, homophonicSpec{"Longfellow", "global-uniform", longfellow,
			fmt.Sprintf("%slongfellow__homophonic__h%03d__uniform__seed001.txt", tr, h), "", h})
	}
	specs = append(specs, homophonicSpec{"Astafiev", "global-uniform", astafiev,
		fmt.Sprintf("%sastafiev__homophonic__h%03d__uniform__seed001.txt", tr, 4), "", 4})
	return specs
}

// runHomophonyDoseResponse implements task60 sections 11-14: for every
// already-prepared Task46/55 homophonic corpus, compare the empirical
// exact-run dose-response (R2, runs>=k, max run) against Doyle/
// Longfellow/Astafiev's own unperturbed values, and check each
// individual plaintext run's survival against the theoretical
// prediction (section 11's (1/H)^(k-1) / weighted sum_j p_j^k, or, for
// frequency-v1, the per-token allocated H from its own
// homophone_allocation.tsv sidecar - never a substituted Hmax, section 14).
func runHomophonyDoseResponse(w *writers) error {
	plaintextCache := map[string]tokenrepetition.Corpus{}
	loadPlain := func(path string) (tokenrepetition.Corpus, error) {
		if c, ok := plaintextCache[path]; ok {
			return c, nil
		}
		c, err := tokenrepetition.LoadCorpus(path, path)
		if err == nil {
			plaintextCache[path] = c
		}
		return c, err
	}
	// Baseline (H=0, no transformation) rows for each plaintext series.
	baselineDone := map[string]bool{}
	for _, spec := range homophonicSpecs() {
		if !baselineDone[spec.Series] {
			baselineDone[spec.Series] = true
			plain, err := loadPlain(spec.PlaintextPath)
			if err != nil {
				return err
			}
			runs := tokenrepetition.ExactRuns(plain.Tokens, plain.LineOfToken)
			adj := tokenrepetition.AdjacentRepetition(plain.Tokens, plain.LineOfToken, 0)
			w.homDoseResponse.row(spec.Series, "0", "plaintext", f8(adj.R2), i(countGE(runs, 3)), i(countGE(runs, 4)), i(countGE(runs, 5)), i(tokenrepetition.MaxObservedRun(runs)))
		}
	}

	for _, spec := range homophonicSpecs() {
		plain, err := loadPlain(spec.PlaintextPath)
		if err != nil {
			return err
		}
		cipher, err := tokenrepetition.LoadCorpus(spec.CipherPath, spec.CipherPath)
		if err != nil {
			return err
		}
		runs := tokenrepetition.ExactRuns(cipher.Tokens, cipher.LineOfToken)
		adj := tokenrepetition.AdjacentRepetition(cipher.Tokens, cipher.LineOfToken, 0)
		w.homDoseResponse.row(spec.Series, i(spec.H), spec.Model, f8(adj.R2), i(countGE(runs, 3)), i(countGE(runs, 4)), i(countGE(runs, 5)), i(tokenrepetition.MaxObservedRun(runs)))

		var hOf func(string) int
		var weightsOf func(int) []float64
		switch spec.Model {
		case "global-uniform":
			hOf = func(string) int { return spec.H }
			weightsOf = tokenrepetition.UniformWeights
		case "global-weighted":
			hOf = func(string) int { return spec.H }
			weightsOf = tokenrepetition.TriangularWeights
		case "frequency-v1-uniform":
			alloc, err := tokenrepetition.LoadAllocation(spec.AllocationPath)
			if err != nil {
				return err
			}
			hOf = func(t string) int {
				if e, ok := alloc[t]; ok && e.AllocatedH > 0 {
					return e.AllocatedH
				}
				return 1
			}
			weightsOf = tokenrepetition.UniformWeights
		}
		rows := tokenrepetition.RunSurvivalDoseResponse(plain.Tokens, cipher.Tokens, plain.LineOfToken, hOf, weightsOf)
		for _, agg := range tokenrepetition.AggregateSurvivalByLength(rows) {
			w.homTheoretical.row(spec.Series, i(spec.H), spec.Model, i(agg.RunLength), i(agg.Count), f8(agg.MeanPredicted), f8(agg.ObservedFraction))
		}
	}
	return nil
}

// runGlyphHomophonyNearRepetition implements task60 section 28: since
// Task46/55 ciphertext tokens are opaque (section 27), glyph-level near-
// repetition controls instead apply Task59's shared, fixed,
// position-independent homophone generator (evaglyph.RandomHomophony)
// directly to a natural corpus's own glyph representation, producing
// real synthetic glyph strings.
func runGlyphHomophonyNearRepetition(rep *report, rng *rand.Rand) error {
	doyle, err := tokenrepetition.LoadCorpus("data_test/pg2097-2.txt", "Doyle")
	if err != nil {
		return err
	}
	glyphTokens := make([][]string, len(doyle.Tokens))
	for i, t := range doyle.Tokens {
		glyphTokens[i] = evaglyph.NaturalGlyphs(t)
	}
	plainGlyphSeqs := map[string][]string{}
	for i, t := range doyle.Tokens {
		plainGlyphSeqs[t] = glyphTokens[i]
	}
	plainDists := tokenrepetition.AdjacentEditDistances(doyle.Tokens, doyle.LineOfToken, plainGlyphSeqs)
	plainRate := tokenrepetition.SummarizeDistances(plainDists, plainGlyphSeqs).PLe1
	rep.glyphHomophonyPlainRate = plainRate
	rep.note("Glyph-level homophony control (Doyle, natural characters as glyphs): plaintext P(d<=1)=%.6f.", plainRate)

	for _, h := range []int{2, 4, 8} {
		homGlyphTokens := evaglyph.RandomHomophony(glyphTokens, h, rng)
		homTokenStrings := make([]string, len(homGlyphTokens))
		homGlyphSeqs := map[string][]string{}
		for i, g := range homGlyphTokens {
			key := joinGlyphsCLI(g)
			homTokenStrings[i] = key
			homGlyphSeqs[key] = g
		}
		dists := tokenrepetition.AdjacentEditDistances(homTokenStrings, doyle.LineOfToken, homGlyphSeqs)
		rate := tokenrepetition.SummarizeDistances(dists, homGlyphSeqs).PLe1
		rep.glyphHomophonyRates = append(rep.glyphHomophonyRates, homophonyGlyphRate{H: h, Rate: rate})
		rep.note("Glyph-level homophony control (Doyle, H=%d, position-independent, shared with Task59): P(d<=1)=%.6f (plaintext %.6f).", h, rate, plainRate)
	}
	return nil
}

func joinGlyphsCLI(g []string) string {
	out := ""
	for _, x := range g {
		out += x
	}
	return out
}
