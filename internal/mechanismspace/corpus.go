// Package mechanismspace implements Task66's target-blind transformation
// families and the generic fingerprint extraction used to compare a
// transformed plaintext corpus against the authoritative Task58-65 Voynich
// fingerprint. It deliberately reuses the shared, already-generic Task58-65
// packages (evaglyph, tokenrepetition, characterentropy, tokenformation,
// tokentransition, lineregime, localregimetopology) rather than
// reinterpreting their metrics a second time (task66 sections 8, 44-51).
package mechanismspace

import (
	"bufio"
	"math/rand"
	"os"
	"strings"

	"zcore.dev/voinich/internal/evaglyph"
)

// Corpus is a plaintext corpus at the WORD granularity: one natural-language
// word per element of Words, grouped into physical Lines by index (Lines[i]
// is the line number of Words[i]). It carries no Voynich-specific content;
// Glyphs() applies the same NaturalGlyphs tokenisation used by every other
// independent Task58-65 analyzer for natural-language controls.
type Corpus struct {
	Name  string
	Words []string
	Lines []int
}

// Glyphs returns the glyph-token representation of the corpus: one
// []string per word, using evaglyph.NaturalGlyphs (task58/59/60's shared
// natural-language convention).
func (c Corpus) Glyphs() [][]string {
	out := make([][]string, len(c.Words))
	for i, w := range c.Words {
		out[i] = evaglyph.NaturalGlyphs(w)
	}
	return out
}

// LoadNatural reads a plaintext corpus file (one physical line of the
// source per row, already-cleaned lowercase text as used throughout this
// repository for Doyle/Longfellow/Astafiev, e.g. data_test/pg2097-2.txt).
func LoadNatural(name, path string) (Corpus, error) {
	f, err := os.Open(path)
	if err != nil {
		return Corpus{}, err
	}
	defer f.Close()
	c := Corpus{Name: name}
	sc := bufio.NewScanner(f)
	sc.Buffer(make([]byte, 4096), 16<<20)
	line := 0
	for sc.Scan() {
		fields := strings.Fields(sc.Text())
		for _, w := range fields {
			g := evaglyph.NaturalGlyphs(w)
			if len(g) == 0 {
				continue
			}
			c.Words = append(c.Words, strings.Join(g, ""))
			c.Lines = append(c.Lines, line)
		}
		if len(fields) > 0 {
			line++
		}
	}
	return c, sc.Err()
}

// MatchedSample deterministically extracts a contiguous block of n words
// starting at a seed-derived (but fixed, not content-optimized) offset, so
// comparisons across corpora of different native sizes use a Voynich-
// comparable sample size rather than letting corpus length itself drive
// entropy/vocabulary-graph/local-regime metrics (task66 section 7). Offset
// selection depends only on corpus length and seed, never on corpus
// content, so it cannot be tuned toward a good result.
func (c Corpus) MatchedSample(n int, seed int64) Corpus {
	if n <= 0 || n >= len(c.Words) {
		return c
	}
	r := rand.New(rand.NewSource(seed))
	start := r.Intn(len(c.Words) - n + 1)
	out := Corpus{Name: c.Name, Words: append([]string(nil), c.Words[start:start+n]...), Lines: append([]int(nil), c.Lines[start:start+n]...)}
	return out
}

// ShufflePlaintextWords is the plaintext ablation control (task66 section
// 65): reorders word occurrences with a seeded shuffle, preserving each
// word type's marginal frequency exactly while destroying sequential
// structure.
func (c Corpus) ShufflePlaintextWords(seed int64) Corpus {
	r := rand.New(rand.NewSource(seed))
	words := append([]string(nil), c.Words...)
	r.Shuffle(len(words), func(i, j int) { words[i], words[j] = words[j], words[i] })
	return Corpus{Name: c.Name + "-shuffled", Words: words, Lines: append([]int(nil), c.Lines...)}
}

// ShufflePlaintextGlyphs is the STREAM-mode plaintext ablation: it
// preserves the marginal glyph frequency of the corpus but destroys
// sequential order at the character level (task66 section 65, applied to
// the continuous-stream input mode rather than the word stream).
func (c Corpus) ShufflePlaintextGlyphs(seed int64) []string {
	r := rand.New(rand.NewSource(seed))
	var all []string
	for _, w := range c.Words {
		all = append(all, evaglyph.NaturalGlyphs(w)...)
	}
	r.Shuffle(len(all), func(i, j int) { all[i], all[j] = all[j], all[i] })
	return all
}

// LinesOf groups glyph tokens by physical line, for the native-layout
// mode of task66 section 53A.
func LinesOf(tokens [][]string, lines []int) [][][]string {
	if len(lines) == 0 {
		return [][][]string{tokens}
	}
	maxLine := 0
	for _, l := range lines {
		if l > maxLine {
			maxLine = l
		}
	}
	out := make([][][]string, maxLine+1)
	for i, t := range tokens {
		out[lines[i]] = append(out[lines[i]], t)
	}
	return out
}

// LayoutCounts is the coarse (token-identity-free) layout shape used for
// the layout-matched secondary mode (task66 section 53B): only counts, so
// no token identity crosses from Voynich into a natural-language corpus.
type LayoutCounts struct {
	TokensPerLine []int
	LinesPerPage  []int
}

// ApplyLayout re-chops a flat output-token stream into synthetic lines and
// pages using only coarse counts (never token identities), cycling through
// the observed counts if the stream is longer than one full pass (task66
// section 53B).
func ApplyLayout(tokens [][]string, layout LayoutCounts) (lines [][][]string, pages [][][][]string) {
	if len(layout.TokensPerLine) == 0 {
		return [][][]string{tokens}, nil
	}
	li := 0
	for pos := 0; pos < len(tokens); {
		n := layout.TokensPerLine[li%len(layout.TokensPerLine)]
		if n <= 0 {
			n = 1
		}
		end := pos + n
		if end > len(tokens) {
			end = len(tokens)
		}
		lines = append(lines, tokens[pos:end])
		pos = end
		li++
	}
	if len(layout.LinesPerPage) == 0 {
		return lines, nil
	}
	pi, idx := 0, 0
	for idx < len(lines) {
		n := layout.LinesPerPage[pi%len(layout.LinesPerPage)]
		if n <= 0 {
			n = 1
		}
		end := idx + n
		if end > len(lines) {
			end = len(lines)
		}
		pages = append(pages, lines[idx:end])
		idx = end
		pi++
	}
	return lines, pages
}
