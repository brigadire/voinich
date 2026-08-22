// Package lineregime contains Task64's line-extraction, null-generation and
// minimal-regime-model primitives. Token/glyph distance is never redefined
// here: pairwise distance delegates to the authoritative Task60/63
// implementation in internal/tokentransition (task64 section 4).
package lineregime

import (
	"math/rand"
	"sort"

	"zcore.dev/voinich/internal/tokentransition"
)

// Line is one physical transcription line (task64 section 5). Folio/Currier/
// Hand are empty when authoritative IVTFF metadata could not be aligned;
// callers must treat that as NOT_APPLICABLE rather than guessing (section 32).
type Line struct {
	Index   int
	Folio   string
	Currier string
	Hand    string
	Tokens  [][]string
}

func (l Line) N() int { return len(l.Tokens) }

// BuildLines pairs each pre-split source line with its authoritative
// per-line folio/Currier/Hand metadata. When metaOK is false, or a source
// line has no corresponding metadata row, Folio/Currier/Hand stay empty.
func BuildLines(tokensByLine [][][]string, folio, currier, hand []string, metaOK bool) []Line {
	out := make([]Line, len(tokensByLine))
	for i, toks := range tokensByLine {
		l := Line{Index: i, Tokens: toks}
		if metaOK && i < len(folio) {
			l.Folio, l.Currier, l.Hand = folio[i], currier[i], hand[i]
		}
		out[i] = l
	}
	return out
}

// Eligible returns the lines with at least minN tokens (task64 section 6).
func Eligible(lines []Line, minN int) []Line {
	out := make([]Line, 0, len(lines))
	for _, l := range lines {
		if l.N() >= minN {
			out = append(out, l)
		}
	}
	return out
}

// Pair is one within-line token-pair observation (task64 section 7).
type Pair struct {
	I, J, Separation, Distance int
}

// WithinLinePairs returns every i<j token pair inside l. Because pairs are
// built only from l.Tokens, they can never reference a token from another
// line (task64 section 7, test 2).
func WithinLinePairs(l Line) []Pair {
	out := make([]Pair, 0, l.N()*(l.N()-1)/2)
	for i := 0; i < l.N(); i++ {
		for j := i + 1; j < l.N(); j++ {
			out = append(out, Pair{I: i, J: j, Separation: j - i,
				Distance: tokentransition.EditDistance(l.Tokens[i], l.Tokens[j])})
		}
	}
	return out
}

// SeparationBucket implements the task64 section 10 adjacency stratification.
func SeparationBucket(sep int) string {
	switch sep {
	case 1:
		return "SEP1"
	case 2:
		return "SEP2"
	default:
		return "SEP3+"
	}
}

// ShuffleWithinLine permutes token order inside l, preserving line
// membership and the line's token multiset while destroying order
// (task64 section 16).
func ShuffleWithinLine(l Line, r *rand.Rand) Line {
	out := Line{Index: l.Index, Folio: l.Folio, Currier: l.Currier, Hand: l.Hand,
		Tokens: append([][]string(nil), l.Tokens...)}
	r.Shuffle(len(out.Tokens), func(a, b int) { out.Tokens[a], out.Tokens[b] = out.Tokens[b], out.Tokens[a] })
	return out
}

// ShuffleLineMembership redistributes every token uniformly at random across
// all lines, preserving the line-length sequence and the global token
// multiset while destroying actual line membership (task64 section 17).
func ShuffleLineMembership(lines []Line, r *rand.Rand) []Line {
	flat := make([][]string, 0)
	for _, l := range lines {
		flat = append(flat, l.Tokens...)
	}
	r.Shuffle(len(flat), func(a, b int) { flat[a], flat[b] = flat[b], flat[a] })
	out := make([]Line, len(lines))
	q := 0
	for i, l := range lines {
		out[i] = Line{Index: l.Index, Folio: l.Folio, Currier: l.Currier, Hand: l.Hand,
			Tokens: append([][]string(nil), flat[q:q+l.N()]...)}
		q += l.N()
	}
	return out
}

// ShuffleLineMembershipWithinPage is the page-conditional variant of
// ShuffleLineMembership: token identities are redistributed only among
// lines sharing the same Folio, preserving each page's token multiset and
// every line's length (task64 section 18).
func ShuffleLineMembershipWithinPage(lines []Line, r *rand.Rand) []Line {
	groups := map[string][]int{}
	for i, l := range lines {
		groups[l.Folio] = append(groups[l.Folio], i)
	}
	pages := make([]string, 0, len(groups))
	for k := range groups {
		pages = append(pages, k)
	}
	sort.Strings(pages)
	out := make([]Line, len(lines))
	copy(out, lines)
	for _, p := range pages {
		idx := groups[p]
		flat := make([][]string, 0)
		for _, i := range idx {
			flat = append(flat, lines[i].Tokens...)
		}
		r.Shuffle(len(flat), func(a, b int) { flat[a], flat[b] = flat[b], flat[a] })
		q := 0
		for _, i := range idx {
			n := lines[i].N()
			l := lines[i]
			l.Tokens = append([][]string(nil), flat[q:q+n]...)
			out[i] = l
			q += n
		}
	}
	return out
}

// PseudoLineGlobal builds synthetic lines the same sizes as lines, sampling
// tokens with replacement from pool (task64 section 15 variant A).
func PseudoLineGlobal(lines []Line, pool [][]string, r *rand.Rand) []Line {
	out := make([]Line, len(lines))
	for i, l := range lines {
		toks := make([][]string, l.N())
		for j := range toks {
			toks[j] = pool[r.Intn(len(pool))]
		}
		out[i] = Line{Index: l.Index, Folio: l.Folio, Currier: l.Currier, Hand: l.Hand, Tokens: toks}
	}
	return out
}

// PseudoLineSamePage is PseudoLineGlobal restricted to each line's own page
// pool (task64 section 15 variant B). A page with an empty pool falls back
// to recycling the line's own tokens rather than crashing.
func PseudoLineSamePage(lines []Line, pagePool map[string][][]string, r *rand.Rand) []Line {
	out := make([]Line, len(lines))
	for i, l := range lines {
		pool := pagePool[l.Folio]
		toks := make([][]string, l.N())
		for j := range toks {
			if len(pool) == 0 {
				toks[j] = l.Tokens[j]
				continue
			}
			toks[j] = pool[r.Intn(len(pool))]
		}
		out[i] = Line{Index: l.Index, Folio: l.Folio, Currier: l.Currier, Hand: l.Hand, Tokens: toks}
	}
	return out
}

// PseudoLineLengthPreserving keeps each line's empirical token-length
// sequence but resamples, at each position, a random token of matching
// length from lengthPool (task64 section 15 variant C).
func PseudoLineLengthPreserving(lines []Line, lengthPool map[int][][]string, r *rand.Rand) []Line {
	out := make([]Line, len(lines))
	for i, l := range lines {
		toks := make([][]string, l.N())
		for j, t := range l.Tokens {
			cand := lengthPool[len(t)]
			if len(cand) == 0 {
				toks[j] = t
				continue
			}
			toks[j] = cand[r.Intn(len(cand))]
		}
		out[i] = Line{Index: l.Index, Folio: l.Folio, Currier: l.Currier, Hand: l.Hand, Tokens: toks}
	}
	return out
}

// ShiftedBlocks re-chunks flat into contiguous blocks with the same sizes as
// real lines (sizes), but with every boundary displaced by offset tokens
// relative to the real line boundaries; the leading offset tokens are
// dropped as a partial block (task64 section 26). Calling it twice with the
// same arguments always returns the same blocks (test 11).
func ShiftedBlocks(flat [][]string, sizes []int, offset int) [][][]string {
	if offset < 0 {
		offset = 0
	}
	if offset > len(flat) {
		offset = len(flat)
	}
	out := make([][][]string, 0, len(sizes))
	pos := offset
	for _, n := range sizes {
		if pos+n > len(flat) {
			break
		}
		out = append(out, append([][]string(nil), flat[pos:pos+n]...))
		pos += n
	}
	return out
}

// FixedWindows re-chunks flat into non-overlapping contiguous windows of
// exactly w tokens (task64 section 27); a final shorter remainder is
// dropped rather than padded.
func FixedWindows(flat [][]string, w int) [][][]string {
	out := make([][][]string, 0)
	for pos := 0; pos+w <= len(flat); pos += w {
		out = append(out, append([][]string(nil), flat[pos:pos+w]...))
	}
	return out
}

// Categorical is a normalized-on-demand discrete distribution over string
// labels, used by the minimal frozen regime model (task64 section 44).
type Categorical struct {
	Keys    []string
	Weights []float64
}

// NewCategorical builds a deterministic Categorical from counts: keys are
// sorted so float64 accumulation order never depends on map iteration order
// (project convention).
func NewCategorical(counts map[string]int) Categorical {
	keys := make([]string, 0, len(counts))
	for k := range counts {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	w := make([]float64, len(keys))
	for i, k := range keys {
		w[i] = float64(counts[k])
	}
	return Categorical{Keys: keys, Weights: w}
}

// Sample draws one key proportional to its weight. It returns "" for an
// empty distribution rather than panicking.
func (c Categorical) Sample(r *rand.Rand) string {
	total := 0.0
	for _, w := range c.Weights {
		total += w
	}
	if total <= 0 || len(c.Keys) == 0 {
		return ""
	}
	x := r.Float64() * total
	for i, w := range c.Weights {
		x -= w
		if x <= 0 {
			return c.Keys[i]
		}
	}
	return c.Keys[len(c.Keys)-1]
}
