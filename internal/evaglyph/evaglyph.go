// Package evaglyph is the single EVA composite-collapsing glyph parser
// shared by the independent Task58 (Rozanova-Temerev replication), Task59
// (glyph positional specialization), and Task60 (token repetition) analyzers,
// so all three interpret Voynich glyphs identically instead of each keeping
// its own copy (task58 section 7, task59 section 3, task60 section 2).
package evaglyph

import (
	"math"
	"sort"
	"strings"
	"unicode"
)

// composites is the longest-first EVA multi-character-composite -> atomic
// symbol mapping already used by the Task58 replication before this
// package existed. Order matters: three-character composites are listed
// before the two-character ones they could otherwise partially match.
var composites = [][2]string{
	{"cth", "C"}, {"ckh", "K"}, {"cph", "P"}, {"cfh", "F"},
	{"iin", "N"}, {"ain", "A"},
	{"ch", "H"}, {"sh", "S"}, {"ee", "E"}, {"in", "I"},
}

// CollapseEVA lowercases s, greedily collapses the known EVA composites
// (longest first) into single atomic uppercase symbols, and returns the
// remaining runes as one glyph each. This is the Voynich glyph
// tokenization used throughout Task58/59/60.
func CollapseEVA(s string) []string {
	s = strings.ToLower(s)
	for _, p := range composites {
		s = strings.ReplaceAll(s, p[0], p[1])
	}
	out := make([]string, 0, len(s))
	for _, r := range s {
		out = append(out, string(r))
	}
	return out
}

// NaturalGlyphs returns the lowercase Unicode letters/digits of s, one
// glyph per character, for natural-language controls (Doyle, Longfellow,
// Astafiev): punctuation and whitespace are removed, case is folded.
func NaturalGlyphs(s string) []string {
	out := make([]string, 0, len(s))
	for _, r := range strings.ToLower(s) {
		if unicode.IsLetter(r) || unicode.IsNumber(r) {
			out = append(out, string(r))
		}
	}
	return out
}

// Classify returns the within-token position class of the i-th glyph of
// an n-glyph token: SINGLETON for a one-glyph token, otherwise INITIAL /
// MEDIAL / FINAL (task59 section 4).
func Classify(n, i int) string {
	if n == 1 {
		return "SINGLETON"
	}
	if i == 0 {
		return "INITIAL"
	}
	if i == n-1 {
		return "FINAL"
	}
	return "MEDIAL"
}

// entropy returns the Shannon entropy (bits) of the empirical distribution
// given by counts. Keys are visited in sorted order so the float64
// accumulation order - and therefore the result's low-order bits - does
// not depend on Go's randomized map iteration order (see project memory:
// sort map keys before float64 accumulation for run-to-run determinism).
func entropy(counts map[string]int) float64 {
	n := 0
	for _, v := range counts {
		n += v
	}
	if n == 0 {
		return 0
	}
	keys := make([]string, 0, len(counts))
	for k := range counts {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	h := 0.0
	for _, k := range keys {
		v := counts[k]
		if v == 0 {
			continue
		}
		p := float64(v) / float64(n)
		h -= p * math.Log2(p)
	}
	return h
}

// MI returns the discrete mutual information (bits) between two paired
// sequences a and b of equal length, estimated from their empirical joint
// and marginal distributions (no bias correction) - the shared estimator
// behind Task58's token-order/glyph-edge metrics and Task59's glyph-edge
// comparison.
func MI(a, b []string) float64 {
	if len(a) == 0 {
		return 0
	}
	x, y, j := map[string]int{}, map[string]int{}, map[string]int{}
	for i := range a {
		x[a[i]]++
		y[b[i]]++
		j[a[i]+"\x00"+b[i]]++
	}
	return entropy(x) + entropy(y) - entropy(j)
}
