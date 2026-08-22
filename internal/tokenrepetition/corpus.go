package tokenrepetition

import (
	"regexp"

	"zcore.dev/voinich/internal/genericsegmentation"
)

var opaqueTokenRE = regexp.MustCompile(`^x[0-9]{6}$`)

// LoadCorpus reads a corpus with internal/genericsegmentation.ReadCorpus
// (the same reader every generic pipeline stage uses) and detects whether
// it is a Task46/55 opaque-token ciphertext corpus (task60 section 27):
// opaque if a supermajority of its distinct token types match the
// x%06d convention.
func LoadCorpus(path, name string) (Corpus, error) {
	tokens, lineOfToken, sha, err := genericsegmentation.ReadCorpus(path)
	if err != nil {
		return Corpus{}, err
	}
	types := map[string]bool{}
	for _, t := range tokens {
		types[t] = true
	}
	opaqueCount := 0
	for t := range types {
		if opaqueTokenRE.MatchString(t) {
			opaqueCount++
		}
	}
	opaque := len(types) > 0 && float64(opaqueCount)/float64(len(types)) > 0.5
	return Corpus{Name: name, Path: path, SHA256: sha, Tokens: tokens, LineOfToken: lineOfToken, Opaque: opaque}, nil
}

// GlyphMode selects how a corpus's tokens are split into glyph sequences.
type GlyphMode int

const (
	GlyphVoynich GlyphMode = iota
	GlyphNatural
)
