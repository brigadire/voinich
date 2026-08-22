package tokenrepetition

import (
	"math/rand"
	"sort"

	"zcore.dev/voinich/internal/evaglyph"
)

// GlyphSequences returns the parsed glyph sequence for every distinct
// token in tokens, using the given mode (task60 section 15: EVA glyph
// units for Voynich, natural characters for natural-language controls -
// never bytes/runes directly when they would differ).
func GlyphSequences(tokens []string, mode GlyphMode) map[string][]string {
	out := map[string][]string{}
	for _, t := range tokens {
		if _, ok := out[t]; ok {
			continue
		}
		if mode == GlyphVoynich {
			out[t] = evaglyph.CollapseEVA(t)
		} else {
			out[t] = evaglyph.NaturalGlyphs(t)
		}
	}
	return out
}

// AdjacentDistance is one adjacent pair's glyph-level edit distance
// (task60 section 15/18).
type AdjacentDistance struct {
	Position int
	A, B     string
	Distance int
	Normalized float64
}

// AdjacentEditDistances scans every adjacent token pair once (task60
// section 48: O(number of transitions), never all-pairs) and returns its
// glyph-level Levenshtein distance. A pair crossing a natural line
// boundary is skipped (see sameLine in exact.go); pass lineOfToken=nil to
// disable that check.
func AdjacentEditDistances(tokens []string, lineOfToken []int, glyphs map[string][]string) []AdjacentDistance {
	out := make([]AdjacentDistance, 0, len(tokens))
	for i := 0; i+1 < len(tokens); i++ {
		if !sameLine(lineOfToken, i) {
			continue
		}
		a, b := tokens[i], tokens[i+1]
		d := LevenshteinGlyphs(glyphs[a], glyphs[b])
		out = append(out, AdjacentDistance{Position: i, A: a, B: b, Distance: d, Normalized: NormalizedDistance(glyphs[a], glyphs[b], d)})
	}
	return out
}

// DistanceRateSummary is P(d=0),P(d=1),P(d<=1),P(d<=2) plus the
// same-length-conditional and substitution-only rates (task60 section 18).
type DistanceRateSummary struct {
	Total                    int
	PEq0, PEq1, PLe1, PLe2   float64
	SameLengthPairs          int
	PEq1GivenSameLength      float64
	SubstitutionOnlyRate     float64
}

func SummarizeDistances(dists []AdjacentDistance, glyphs map[string][]string) DistanceRateSummary {
	var s DistanceRateSummary
	s.Total = len(dists)
	if s.Total == 0 {
		return s
	}
	var eq0, eq1, le1, le2, sameLen, sameLenEq1, subOnly int
	for _, d := range dists {
		switch {
		case d.Distance == 0:
			eq0++
		case d.Distance == 1:
			eq1++
		}
		if d.Distance <= 1 {
			le1++
		}
		if d.Distance <= 2 {
			le2++
		}
		if len(glyphs[d.A]) == len(glyphs[d.B]) {
			sameLen++
			if d.Distance == 1 {
				sameLenEq1++
				subOnly++ // equal-length distance-1 pairs are exactly the substitutions
			}
		}
	}
	n := float64(s.Total)
	s.PEq0, s.PEq1, s.PLe1, s.PLe2 = float64(eq0)/n, float64(eq1)/n, float64(le1)/n, float64(le2)/n
	s.SameLengthPairs = sameLen
	if sameLen > 0 {
		s.PEq1GivenSameLength = float64(sameLenEq1) / float64(sameLen)
	}
	s.SubstitutionOnlyRate = float64(subOnly) / n
	return s
}

// vocabIndex buckets distinct tokens by (length, frequency rank) for the
// frequency/length-matched null (task60 section 20).
type vocabIndex struct {
	byLength map[int][]string // tokens of that glyph-length, sorted by descending frequency (rank order)
	rankOf   map[string]int
}

func buildVocabIndex(tokens []string, glyphs map[string][]string) vocabIndex {
	freq := map[string]int{}
	for _, t := range tokens {
		freq[t]++
	}
	keys := sortedTokenKeys(freq)
	sort.SliceStable(keys, func(i, j int) bool { return freq[keys[i]] > freq[keys[j]] })
	vi := vocabIndex{byLength: map[int][]string{}, rankOf: map[string]int{}}
	for rank, t := range keys {
		vi.rankOf[t] = rank
		l := len(glyphs[t])
		vi.byLength[l] = append(vi.byLength[l], t)
	}
	return vi
}

// matchedDraw picks a random token of the same glyph-length as t, whose
// frequency rank is within rankTolerance of t's own rank when enough
// candidates exist at that length, else the closest available rank.
func (vi vocabIndex) matchedDraw(t string, length, rankTolerance int, r *rand.Rand) string {
	pool := vi.byLength[length]
	if len(pool) == 0 {
		return t
	}
	rank := vi.rankOf[t]
	var candidates []string
	for _, c := range pool {
		if abs(vi.rankOf[c]-rank) <= rankTolerance {
			candidates = append(candidates, c)
		}
	}
	if len(candidates) == 0 {
		candidates = pool
	}
	return candidates[r.Intn(len(candidates))]
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

// MatchedNullRate draws matchedNullDraws frequency/length-matched random
// pairs per observed adjacent pair and returns the fraction of matched
// draws that are themselves edit-distance-1 (task60 section 20): the
// expected d=1 rate purely from vocabulary composition (frequency and
// length), against which the true adjacent-pair rate is compared.
func MatchedNullRate(dists []AdjacentDistance, glyphs map[string][]string, tokens []string, matchedNullDraws, rankTolerance int, r *rand.Rand) (rate float64, draws int) {
	vi := buildVocabIndex(tokens, glyphs)
	hits := 0
	for _, d := range dists {
		la, lb := len(glyphs[d.A]), len(glyphs[d.B])
		for k := 0; k < matchedNullDraws; k++ {
			a2 := vi.matchedDraw(d.A, la, rankTolerance, r)
			b2 := vi.matchedDraw(d.B, lb, rankTolerance, r)
			draws++
			if LevenshteinGlyphs(glyphs[a2], glyphs[b2]) == 1 {
				hits++
			}
		}
	}
	if draws == 0 {
		return 0, 0
	}
	return float64(hits) / float64(draws), draws
}
