package main

import (
	"fmt"
	"strings"

	"zcore.dev/voinich/internal/mechanismspace"
)

// representativePerFamily is one grid entry per M1-M11 family, used by
// the plaintext-sensitivity/information-retention/error-robustness passes
// so every family gets a section-64/65/67 result without rerunning the
// entire grid a second time under every control.
func representativePerFamily(grid []GridEntry) []GridEntry {
	seen := map[string]bool{}
	var out []GridEntry
	for _, e := range grid {
		fam := e.Config.Family
		if seen[fam] {
			continue
		}
		seen[fam] = true
		out = append(out, e)
	}
	return out
}

// RunPlaintextSensitivity is task66 sections 65-66: for one representative
// mechanism per family, compare the real-plaintext fingerprint against the
// same mechanism run on shuffled plaintext (marginal frequencies
// preserved), classify the sensitivity class, and write
// PLAINTEXT_SENSITIVITY.tsv.
func RunPlaintextSensitivity(grid []GridEntry, corpora map[string]mechanismspace.Corpus, opt mechanismspace.FingerprintOptions, path string) {
	var b strings.Builder
	b.WriteString("mechanism\tcorpus\tsensitivity_class\n")
	for _, e := range representativePerFamily(grid) {
		for name, c := range corpora {
			realOut := mechanismspace.Transform(e.Config, c)
			realFP := mechanismspace.ComputeFingerprint(realOut.Tokens, realOut.Lines, opt)
			shuffledCfg := e.Config
			shuffledCfg.ShufflePlaintext = true
			shufOut := mechanismspace.Transform(shuffledCfg, c)
			shufFP := mechanismspace.ComputeFingerprint(shufOut.Tokens, shufOut.Lines, opt)
			class := mechanismspace.ClassifySensitivity(realFP, shufFP)
			b.WriteString(fmt.Sprintf("%s\t%s\t%s\n", e.Name, name, class))
		}
	}
	writeFile(path, b.String())
}

// RunInformationRetention is task66 section 64: coarse input/output
// mutual information for one representative mechanism per family.
func RunInformationRetention(grid []GridEntry, corpora map[string]mechanismspace.Corpus, path string) {
	var b strings.Builder
	b.WriteString("mechanism\tcorpus\tmutual_information_bits\n")
	for _, e := range representativePerFamily(grid) {
		for name, c := range corpora {
			out := mechanismspace.Transform(e.Config, c)
			mi := mechanismspace.InformationRetention(c.Glyphs(), out.Tokens)
			b.WriteString(fmt.Sprintf("%s\t%s\t%.9g\n", e.Name, name, mi))
		}
	}
	writeFile(path, b.String())
}
