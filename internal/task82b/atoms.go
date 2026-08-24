package task82b

import (
	"strings"
	"unicode"
)

// glyphsOf returns the lowercase letter/digit glyphs of a token, exactly
// matching internal/evaglyph.NaturalGlyphs's per-character rule (one
// glyph per Unicode letter/number rune, punctuation dropped), so an
// extraction operator's notion of "glyph" agrees with what
// fingerprintv2's GlyphMode=natural will later see. Reimplemented locally
// (rather than imported) because evaglyph.NaturalGlyphs returns
// single-rune strings already, which is exactly what is needed here too.
func glyphsOf(token string) []string {
	var out []string
	for _, r := range strings.ToLower(token) {
		if unicode.IsLetter(r) || unicode.IsNumber(r) {
			out = append(out, string(r))
		}
	}
	return out
}

// TokenAtom is one addressable token position in a carrier corpus.
type TokenAtom struct {
	Line      int
	IdxInLine int
	Text      string
}

// GlyphAtom is one addressable letter/digit glyph position within a
// carrier corpus token.
type GlyphAtom struct {
	Line            int
	TokenIdxInLine  int
	GlyphIdxInToken int
	Ch              string
}

// BuildAtoms flattens a Lines-style [][]string (tokens grouped by natural
// line) into the two addressable atom streams every extraction operator
// and every null model in this package selects from, in corpus order.
func BuildAtoms(groups [][]string) ([]TokenAtom, []GlyphAtom) {
	var tokens []TokenAtom
	var glyphs []GlyphAtom
	for li, line := range groups {
		for ti, tok := range line {
			tokens = append(tokens, TokenAtom{Line: li, IdxInLine: ti, Text: tok})
			for gi, g := range glyphsOf(tok) {
				glyphs = append(glyphs, GlyphAtom{Line: li, TokenIdxInLine: ti, GlyphIdxInToken: gi, Ch: g})
			}
		}
	}
	return tokens, glyphs
}
