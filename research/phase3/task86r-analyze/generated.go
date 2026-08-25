package main

// glyphsForGenerated recovers the GLYPH sequence of a generated token for
// structural (F2) comparison purposes. M1-M5 already generate glyphs
// directly; M0 generates a raw TOKEN string (or the literal "<UNK>"
// reserved token) which must be decoded back to glyphs via rawToGlyphs --
// splitGlyphs for a corpus whose TOKEN.Raw is itself glyph-joined (MFC
// synthetic populations), or evaglyph.CollapseEVA for a real transcription
// whose TOKEN.Raw is the actual EVA field. The literal reserved token is
// always treated as its own atomic one-glyph placeholder.
func glyphsForGenerated(model FittedModel, g GeneratedToken, rawToGlyphs func(string) []string) []string {
	if model.Unit() != "TOKEN" {
		return g.Glyphs
	}
	if g.Raw == unkTokenLiteral {
		return []string{unkTokenLiteral}
	}
	return rawToGlyphs(g.Raw)
}
