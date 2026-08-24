package task82a

import (
	"bufio"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"unicode"

	"zcore.dev/voinich/internal/mnemonicspace"
)

// Latin23 is the frozen alphabet already used by Task82's bounded adapter
// (internal/task82.loadCorpora) and by the F01/F08 registry parameters. It
// is reused verbatim so the literal mapping remains an unmodified extension
// of Task82's adapter, not a new convention.
const Latin23 = "ABCDEFGHIKLMNOPQRSTVXYZ"

// CorpusPaths are the same three natural-language controls Task82 used,
// with the same source/provenance rule: repository-committed text, never
// copied into results, identified only by SHA-256/size/letter-count.
var CorpusPaths = map[string]string{
	"Doyle":      "data_test/pg2097-2.txt",
	"Longfellow": "data_test/pg30795-mod.txt",
	"Astafiev":   "data_test/astafiev-1000-culinar-receipts-utf8.txt",
}

// SourceCorpus holds the extended (corpus-scale) adapter output for one
// natural-language control: a long literal Latin23 stream and a long
// cue-addressable item-ID stream, both mechanically derived and truncated
// to a matched length across all three corpora.
type SourceCorpus struct {
	ID             string   `json:"id"`
	Path           string   `json:"path"`
	SHA256         string   `json:"sha256"`
	Bytes          int      `json:"bytes"`
	UnicodeLetters int      `json:"unicode_letters_available"`
	WordTokens     int      `json:"word_tokens_available"`
	Preprocessing  string   `json:"preprocessing"`
	Letters        []string `json:"-"` // Latin23 single-character symbols, in order
	Items          []string `json:"-"` // "I"+16 hex chars, in order
}

// loadSourceCorpora reads the full committed corpus files and produces the
// mechanical literal/cue streams, capped at capLetters/capItems so that no
// more of the file is materialized than the largest frozen corpus_scale
// will ever need. The mapping rule is byte-for-byte identical to Task82's:
// upper(codepoint mod 23) for letters; SHA-256(lowercase token)[:16 hex]
// prefixed with "I" for word items.
func loadSourceCorpora(root string, capLetters, capItems int) (map[string]SourceCorpus, error) {
	out := map[string]SourceCorpus{}
	for id, rel := range CorpusPaths {
		b, err := os.ReadFile(filepath.Join(root, rel))
		if err != nil {
			return nil, err
		}
		c := SourceCorpus{
			ID:            id,
			Path:          rel,
			SHA256:        sum(b),
			Bytes:         len(b),
			Preprocessing: "Unicode letters -> upper(codepoint mod 23) over Latin23; whitespace-delimited word tokens -> \"I\"+SHA-256(lower(token))[:16 hex]; both streams truncated to a corpus-scale-matched length (see TASK82A_DESIGN.md); mechanically identical to Task82's bounded adapter, only length differs",
		}
		words := bufio.NewScanner(strings.NewReader(string(b)))
		words.Split(bufio.ScanWords)
		words.Buffer(make([]byte, 1024), 1024*1024)
		for words.Scan() {
			token := strings.TrimFunc(words.Text(), func(r rune) bool { return !unicode.IsLetter(r) })
			if token == "" {
				continue
			}
			c.WordTokens++
			if len(c.Items) < capItems {
				h := sha256.Sum256([]byte(strings.ToLower(token)))
				c.Items = append(c.Items, "I"+hex.EncodeToString(h[:8]))
			}
		}
		for _, r := range string(b) {
			if unicode.IsLetter(r) {
				c.UnicodeLetters++
				if len(c.Letters) < capLetters {
					c.Letters = append(c.Letters, string(Latin23[int(r)%len(Latin23)]))
				}
			}
		}
		if len(c.Letters) < capLetters || len(c.Items) < capItems {
			return nil, fmt.Errorf("corpus %s has only %d letters/%d items, need %d/%d", id, len(c.Letters), len(c.Items), capLetters, capItems)
		}
		out[id] = c
	}
	return out, nil
}

func sum(b []byte) string { h := sha256.Sum256(b); return hex.EncodeToString(h[:]) }

// itemsToCue derives the same opaque-cue convention Task82 used
// ("C"+index), applied here per chunk under whichever CueNamespace policy
// is frozen for the job.
func cueLabel(n int) mnemonicspace.Cue {
	return mnemonicspace.Cue(fmt.Sprintf("C%d", n))
}
