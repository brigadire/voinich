package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"zcore.dev/voinich/internal/fingerprintv2"
)

// StructuralMetricIDs are the seven frozen G1-applicable Fingerprint V2
// metrics (G1_EXECUTABLE_CONTRACT.json metrics.structural).
var StructuralMetricIDs = []string{
	"EF1_GIANT_COMPONENT_SHARE", "EF1_ISOLATE_SHARE", "EF2_GLOBAL_CLUSTERING",
	"EF3_DEGREE_FREQUENCY_SPEARMAN", "LP1_RULE_SUPPORT_GINI",
	"LP4_PREFIX_ATTACHMENT_NMI", "LP4_SUFFIX_ATTACHMENT_NMI",
}

// aliasPool is a deterministic pool of single-rune, letter-classified,
// case-fold-stable code points used to encode an arbitrary (possibly
// multi-character, EVA-composite) glyph alphabet into internal/
// fingerprintv2's "natural" glyph mode (one rune = one glyph, no EVA
// re-collapsing of composite symbols like "CTH"). This avoids the
// alternative -- writing composite glyphs back out as lowercase EVA
// substrings and letting fingerprintv2 re-run CollapseEVA -- which risks
// spurious re-collapsing at glyph-boundary junctions.
func aliasPool() []rune {
	var pool []rune
	for r := rune(0x03B1); r <= 0x03C9; r++ { // Greek lowercase
		pool = append(pool, r)
	}
	for r := rune(0x0430); r <= 0x044F; r++ { // Cyrillic lowercase
		pool = append(pool, r)
	}
	for r := 'a'; r <= 'z'; r++ {
		pool = append(pool, r)
	}
	return pool
}

// GlyphAlias is a bijective glyph<->rune encoding fixed for one alphabet.
type GlyphAlias struct {
	toRune map[string]rune
}

func NewGlyphAlias(alphabet []string) *GlyphAlias {
	sorted := append([]string(nil), alphabet...)
	sorted = append(sorted, unkTokenLiteral, unkGlyphSymbol, eosSymbol)
	sort.Strings(sorted)
	sorted = dedupSorted(sorted)
	pool := aliasPool()
	if len(sorted) > len(pool) {
		panic("alias pool too small for alphabet")
	}
	m := map[string]rune{}
	for i, g := range sorted {
		m[g] = pool[i]
	}
	return &GlyphAlias{toRune: m}
}

func dedupSorted(s []string) []string {
	out := s[:0]
	var prev string
	first := true
	for _, v := range s {
		if first || v != prev {
			out = append(out, v)
			prev = v
			first = false
		}
	}
	return out
}

func (a *GlyphAlias) Encode(glyphs []string) string {
	var b strings.Builder
	for _, g := range glyphs {
		r, ok := a.toRune[g]
		if !ok {
			panic("glyph outside alias alphabet: " + g)
		}
		b.WriteRune(r)
	}
	return b.String()
}

// StructuralMetrics computes the seven frozen G1 F2 metrics over a
// population of glyph-sequence tokens, via a temporary plain-text corpus
// and the frozen internal/fingerprintv2 extractor (natural glyph mode).
func StructuralMetrics(alias *GlyphAlias, populations [][]string, seed int64, workDir string) (map[string]float64, bool, error) {
	dir, err := os.MkdirTemp(workDir, "f2pop-")
	if err != nil {
		return nil, false, err
	}
	defer os.RemoveAll(dir)
	path := filepath.Join(dir, "corpus.txt")
	f, err := os.Create(path)
	if err != nil {
		return nil, false, err
	}
	for _, glyphs := range populations {
		if _, err := fmt.Fprintln(f, alias.Encode(glyphs)); err != nil {
			f.Close()
			return nil, false, err
		}
	}
	if err := f.Close(); err != nil {
		return nil, false, err
	}
	cfg := fingerprintv2.Config{
		Version:             fingerprintv2.Version,
		OutputDir:           dir,
		Primary:             fingerprintv2.CorpusConfig{ID: "pop", Path: path, GlyphMode: "natural"},
		Seed:                seed,
		Repetitions:         1,
		MinRuleSupport:      3,
		Alpha:               0.05,
		GraphSwaps:          10,
		DiagnosticTolerance: 0.20,
		Grammar:             fingerprintv2.GrammarConfig{Modes: []string{"structure-preserving", "frequency-aware"}},
	}
	fp, err := fingerprintv2.Run(cfg)
	if err != nil {
		return nil, false, err
	}
	if fp.Primary.Metrics.LP2.ProductivityState == "INSUFFICIENT_SUPPORT" {
		return nil, false, nil
	}
	m := fp.Primary.Metrics
	return map[string]float64{
		"EF1_GIANT_COMPONENT_SHARE":     m.EF1.GiantComponentShare,
		"EF1_ISOLATE_SHARE":             m.EF1.IsolateShare,
		"EF2_GLOBAL_CLUSTERING":         m.EF2.GlobalClustering,
		"EF3_DEGREE_FREQUENCY_SPEARMAN": m.EF3.SpearmanDegreeLogFrequency,
		"LP1_RULE_SUPPORT_GINI":         m.LP1.SupportGini,
		"LP4_PREFIX_ATTACHMENT_NMI":     m.LP4.Prefix.NormalizedMI,
		"LP4_SUFFIX_ATTACHMENT_NMI":     m.LP4.Suffix.NormalizedMI,
	}, true, nil
}
