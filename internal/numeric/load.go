package numeric

import (
	"crypto/sha256"
	"fmt"
	"os"
	"sort"
	"strings"

	"zcore.dev/voinich/internal/genericsegmentation"
	"zcore.dev/voinich/internal/metadatavalidation"
)

// LoadVoynich uses literal transcription characters. This deliberately does
// not equate EVA characters with physical glyphs. The frozen admission rule is
// lowercase ASCII a-z; a token containing any other character is excluded.
func LoadVoynich(name, corpusPath, ivtffPath string) (Corpus, error) {
	raw, lines, hash, err := genericsegmentation.ReadCorpus(corpusPath)
	if err != nil {
		return Corpus{}, err
	}
	doc, err := metadatavalidation.ParseIVTFF(ivtffPath)
	if err != nil {
		return Corpus{}, err
	}
	aligned, err := metadatavalidation.Align(doc, raw, hash)
	if err != nil {
		return Corpus{}, fmt.Errorf("strict IVTFF alignment: %w", err)
	}
	ivraw, err := os.ReadFile(ivtffPath)
	if err != nil {
		return Corpus{}, err
	}
	c := Corpus{Name: name, Path: corpusPath, SHA256: hash, IVTFFPath: ivtffPath,
		IVTFFSHA256: fmt.Sprintf("%x", sha256.Sum256(ivraw)), RawTokenCount: len(raw)}
	uniq, alphabet := map[string]bool{}, map[byte]bool{}
	for _, s := range raw {
		uniq[s] = true
	}
	c.UniqueTokenCount = len(uniq)
	lineMap, nextLine := map[string]int{}, 0
	linePos := map[int]int{}
	for i, s := range raw {
		m := aligned.Records[i]
		key := m.Folio + "\x00" + m.LineID
		ln, ok := lineMap[key]
		if !ok {
			ln = nextLine
			lineMap[key] = ln
			nextLine++
		}
		physicalPos := linePos[ln]
		linePos[ln]++
		good := s != ""
		for j := 0; j < len(s); j++ {
			if s[j] < 'a' || s[j] > 'z' {
				good = false
				break
			}
		}
		if !good {
			c.ExcludedTokenCount++
			continue
		}
		gs := append([]byte(nil), []byte(s)...)
		for _, g := range gs {
			alphabet[g] = true
		}
		c.Tokens = append(c.Tokens, Token{Text: s, Glyphs: gs, Line: ln,
			IndexInLine: physicalPos, Folio: m.Folio, Section: m.Section, LocusType: m.LocusType})
	}
	_ = lines // canonical line indices are not assumed equivalent to IVTFF lines.
	c.LineCount = nextLine
	for g := range alphabet {
		c.Alphabet = append(c.Alphabet, g)
	}
	sort.Slice(c.Alphabet, func(i, j int) bool { return c.Alphabet[i] < c.Alphabet[j] })
	return c, nil
}

// LoadNatural is a secondary negative control: Unicode-free ASCII words are
// case-folded and punctuation is stripped, with original physical lines kept.
func LoadNatural(path string, maxTokens int) (Corpus, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return Corpus{}, err
	}
	c := Corpus{Name: "LATIN_NATURAL", Path: path, SHA256: fmt.Sprintf("%x", sha256.Sum256(b))}
	alphabet := map[byte]bool{}
	uniq := map[string]bool{}
	for li, line := range strings.Split(string(b), "\n") {
		pos := 0
		for _, raw := range strings.Fields(line) {
			var x []byte
			for _, g := range []byte(strings.ToLower(raw)) {
				if g >= 'a' && g <= 'z' {
					x = append(x, g)
				}
			}
			if len(x) == 0 {
				continue
			}
			s := string(x)
			uniq[s] = true
			for _, g := range x {
				alphabet[g] = true
			}
			c.Tokens = append(c.Tokens, Token{Text: s, Glyphs: x, Line: li, IndexInLine: pos})
			pos++
			if maxTokens > 0 && len(c.Tokens) >= maxTokens {
				break
			}
		}
		if maxTokens > 0 && len(c.Tokens) >= maxTokens {
			break
		}
		c.LineCount = li + 1
	}
	c.RawTokenCount = len(c.Tokens)
	c.UniqueTokenCount = len(uniq)
	for g := range alphabet {
		c.Alphabet = append(c.Alphabet, g)
	}
	sort.Slice(c.Alphabet, func(i, j int) bool { return c.Alphabet[i] < c.Alphabet[j] })
	return c, nil
}
