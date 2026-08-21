// Package corpustransform implements task46's deterministic historical
// cipher corpus transformer: token-level rectangular columnar transposition
// and token-level homophonic substitution over an existing plain-text
// corpus.
//
// The package treats the input corpus purely as a flat sequence of
// whitespace-delimited tokens (task46 section 2). It performs no
// linguistic tokenization, case folding, or punctuation handling beyond
// whitespace splitting, and it never reads or writes anything about the
// Voynich manuscript - it is a mechanistic experiment-input generator, not
// a stage of the scientific pipeline (see TRANSFORMATION_METHODS.md).
package corpustransform

import (
	"bufio"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"strings"
)

// SchemaVersion is the corpus-transform manifest schema version.
const SchemaVersion = 1

// Transformer is the fixed tool identity recorded in every manifest.
const Transformer = "corpus-transform"

const (
	MethodTransposition = "transposition"
	MethodHomophonic    = "homophonic"

	OrderNatural = "natural"
	OrderKeyed   = "keyed"

	SelectionUniform        = "uniform"
	SelectionWeighted       = "weighted"
	SelectionUniformVersion = "uniform-v1"

	HomophoneModelFixed     = "fixed"
	HomophoneModelFrequency = "frequency"

	LinePolicyPreserve = "preserve"
	LinePolicyReflow   = "reflow"
)

// ReflowTokensPerLine is the fixed, corpus-content-independent line width
// used by -line-policy reflow (see TRANSFORMATION_METHODS.md, "reflow-v1").
// It is a pure serialization constant: it is never derived from a corpus,
// a cipher parameter, or Voynich statistics.
const ReflowTokensPerLine = 10

// Corpus is a corpus read as a flat token stream plus the original
// per-line token counts, used only for -line-policy preserve output
// chunking (see WriteCorpus).
type Corpus struct {
	Tokens         []string
	LineLengths    []int
	InputSHA256Hex string
}

// ReadCorpus reads path as a whitespace-tokenized flat token stream,
// recording the natural per-line token counts alongside it. It performs no
// linguistic tokenization and does not require the input to already be a
// codex_prepare canonical corpus (task46 accepts any whitespace-tokenized
// text, e.g. data_test/pg2097-2.txt directly).
func ReadCorpus(path string) (Corpus, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Corpus{}, err
	}
	sum := sha256.Sum256(data)
	var tokens []string
	var lineLengths []int
	s := bufio.NewScanner(strings.NewReader(string(data)))
	s.Buffer(make([]byte, 64*1024), 64*1024*1024)
	for s.Scan() {
		fields := strings.Fields(s.Text())
		lineLengths = append(lineLengths, len(fields))
		tokens = append(tokens, fields...)
	}
	if err := s.Err(); err != nil {
		return Corpus{}, err
	}
	return Corpus{Tokens: tokens, LineLengths: lineLengths, InputSHA256Hex: hex.EncodeToString(sum[:])}, nil
}

// WriteCorpus serializes tokens (already in final, post-transform order) to
// a plain canonical text corpus, applying the given line policy. It never
// adds headers, comments, metadata, or markers (task46 section 9).
//
//   - LinePolicyPreserve chunks tokens using origLineLengths, the input
//     corpus's own natural per-line token counts, in order. For homophonic
//     substitution (which never reorders tokens) this reproduces the
//     original line boundaries exactly. For transposition (which
//     necessarily reorders tokens globally) this reproduces only the
//     original *line-length sequence*, not the original line contents: it
//     keeps the natural-line count and length distribution identical to
//     the untransformed corpus, so a downstream generic-pipeline line-based
//     block partition (internal/genericsegmentation) sees the same
//     partition granularity it would see on the plaintext control. It does
//     NOT mean "line k of the output corresponds to line k of the input".
//   - LinePolicyReflow ignores origLineLengths and wraps tokens at a fixed
//     ReflowTokensPerLine width. This changes the natural-line count
//     relative to the plaintext control and therefore the generic
//     pipeline's block-partition granularity; see TRANSFORMATION_METHODS.md
//     for why this is not the default.
func WriteCorpus(tokens []string, origLineLengths []int, linePolicy string) ([]byte, error) {
	switch linePolicy {
	case LinePolicyPreserve:
		return chunkAndSerialize(tokens, origLineLengths)
	case LinePolicyReflow:
		return chunkAndSerialize(tokens, reflowLengths(len(tokens), ReflowTokensPerLine))
	default:
		return nil, fmt.Errorf("unsupported line policy %q", linePolicy)
	}
}

func reflowLengths(n, width int) []int {
	if n == 0 {
		return nil
	}
	if width < 1 {
		width = 1
	}
	var lens []int
	for remaining := n; remaining > 0; {
		l := min(width, remaining)
		lens = append(lens, l)
		remaining -= l
	}
	return lens
}

func chunkAndSerialize(tokens []string, lineLengths []int) ([]byte, error) {
	total := 0
	for _, l := range lineLengths {
		total += l
	}
	if total != len(tokens) {
		return nil, fmt.Errorf("corpustransform: line length sequence sums to %d tokens, output has %d", total, len(tokens))
	}
	var b strings.Builder
	pos := 0
	for _, l := range lineLengths {
		b.WriteString(strings.Join(tokens[pos:pos+l], " "))
		b.WriteByte('\n')
		pos += l
	}
	return []byte(b.String()), nil
}

// ShaBytes returns the lowercase hex SHA256 digest of b.
func ShaBytes(b []byte) string {
	h := sha256.Sum256(b)
	return hex.EncodeToString(h[:])
}

// TokenCounts returns the occurrence count of every distinct token.
func TokenCounts(tokens []string) map[string]int {
	counts := make(map[string]int, len(tokens))
	for _, t := range tokens {
		counts[t]++
	}
	return counts
}

// MultisetEqual reports whether a and b contain exactly the same tokens
// with the same multiplicities, regardless of order.
func MultisetEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	ca, cb := TokenCounts(a), TokenCounts(b)
	if len(ca) != len(cb) {
		return false
	}
	for k, v := range ca {
		if cb[k] != v {
			return false
		}
	}
	return true
}
