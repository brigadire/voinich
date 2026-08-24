package task82b

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"zcore.dev/voinich/internal/genericsegmentation"
)

// assertNoVoynichPath is the Voynich firewall (task82b.txt sec.4/71):
// this package must never build an F2 config over a Voynich transcription
// or IVTFF-derived artifact before Task83. Verbatim policy from
// internal/task82a/f2.go, reapplied independently so task82b never
// imports task82a.
func assertNoVoynichPath(path string) error {
	lower := strings.ToLower(path)
	for _, bad := range []string{"voynich", "zl3b", "it2a", "cd2a", "fg2a", "vt0e", "rf1b", "eva.txt", "data/ivtff", "data_work/"} {
		if strings.Contains(lower, bad) {
			return fmt.Errorf("VOYNICH_FIREWALL: refusing to build an F2 config over %q", path)
		}
	}
	return nil
}

// CarrierPaths are the natural-language extraction carriers (task82b.txt
// sec.31), the same three controls and paths Task82/Task82a already use.
var CarrierPaths = map[string]string{
	"Doyle":      "data_test/pg2097-2.txt",
	"Longfellow": "data_test/pg30795-mod.txt",
	"Astafiev":   "data_test/astafiev-1000-culinar-receipts-prepared.txt",
}

// CarrierOrder is CarrierPaths in a fixed, frozen iteration order.
var CarrierOrder = []string{"Doyle", "Longfellow", "Astafiev"}

// Lines is one loaded corpus: whitespace tokens grouped by natural line,
// plus provenance.
type Lines struct {
	ID      string
	Path    string
	SHA256  string
	Tokens  [][]string // per natural line
	NTokens int
}

// LoadLines reads a plain-text corpus into per-natural-line token groups,
// using the same tokenization genericsegmentation.ReadCorpus already uses
// (scan lines, split on whitespace) so every downstream operator/F2 call
// agrees on what a "line" and a "token" are.
func LoadLines(root, id, relPath string) (Lines, error) {
	full := filepath.Join(root, relPath)
	if err := assertNoVoynichPath(full); err != nil {
		return Lines{}, err
	}
	tokens, lineOf, sum, err := genericsegmentation.ReadCorpus(full)
	if err != nil {
		return Lines{}, err
	}
	var groups [][]string
	for i, tok := range tokens {
		li := lineOf[i]
		for len(groups) <= li {
			groups = append(groups, nil)
		}
		groups[li] = append(groups[li], tok)
	}
	return Lines{ID: id, Path: relPath, SHA256: sum, Tokens: groups, NTokens: len(tokens)}, nil
}

// WriteCorpusFile writes one physical text line per token group (the same
// ASSEMBLER_DEFINED convention internal/task82a/f2.go uses), so
// fingerprintv2 and genericsegmentation agree with this package's own
// notion of "line".
func WriteCorpusFile(path string, groups [][]string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	var b strings.Builder
	for _, g := range groups {
		b.WriteString(strings.Join(g, " "))
		b.WriteByte('\n')
	}
	return os.WriteFile(path, []byte(b.String()), 0o644)
}

// FlattenTokens returns the token stream and each token's line index, in
// corpus order, from a Lines value (or any [][]string groups).
func FlattenTokens(groups [][]string) (tokens []string, lineOf []int) {
	for li, g := range groups {
		for _, tok := range g {
			tokens = append(tokens, tok)
			lineOf = append(lineOf, li)
		}
	}
	return tokens, lineOf
}
