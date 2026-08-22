package main

import (
	"fmt"
	"strings"

	"zcore.dev/voinich/internal/mechanismspace"
)

// ErrorRates is task66 section 67's frozen level set.
var ErrorRates = []float64{0.001, 0.005, 0.01, 0.02, 0.05}

// RunErrorRobustness is task66 section 67: for the best DEVELOPMENT
// candidates (the Pareto frontier), inject scribal-like errors at each
// frozen rate and re-measure a few key fingerprint fields, writing
// ERROR_ROBUSTNESS.tsv. This is a secondary descriptive probe only - it
// is never used for model selection (section 67's own prohibition), so it
// runs after the frontier is already fixed.
func RunErrorRobustness(candidates []string, grid []GridEntry, corpora map[string]mechanismspace.Corpus, opt mechanismspace.FingerprintOptions, path string) {
	byName := map[string]GridEntry{}
	for _, e := range grid {
		byName[e.Name] = e
	}
	var b strings.Builder
	b.WriteString("mechanism\tcorpus\terror_rate\ttoken_order_bits\tweighted_entropy\th2\tgiant_component_fraction\n")
	for _, name := range candidates {
		e, ok := byName[name]
		if !ok {
			continue
		}
		for cname, c := range corpora {
			out := mechanismspace.Transform(e.Config, c)
			for _, rate := range append([]float64{0}, ErrorRates...) {
				tokens := mechanismspace.InjectErrors(out.Tokens, rate, 999)
				fp := mechanismspace.ComputeFingerprint(tokens, nil, opt)
				b.WriteString(fmt.Sprintf("%s\t%s\t%.4g\t%.9g\t%.9g\t%.9g\t%.9g\n", name, cname, rate, fp.TokenOrderBits, fp.PositionalWeightedEntropy, fp.H2, fp.GiantComponentFraction))
			}
		}
	}
	writeFile(path, b.String())
}
