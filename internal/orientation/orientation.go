// Package orientation transforms canonical corpora without changing their
// logical line order or tokenization.
package orientation

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"

	"zcore.dev/voinich/internal/corpusprep"
)

const (
	SchemaVersion         = 1
	Transformation        = "orientation"
	TransformationVersion = 1

	TokenReverse = "TOKEN_REVERSE"
	GlyphReverse = "GLYPH_REVERSE"
	FullReverse  = "FULL_REVERSE"
)

type Stats struct {
	Tokens       int
	UniqueTokens int
	Lines        int
	Counts       map[string]int
	PerLine      []int
	Lengths      []int
}

type Manifest struct {
	SchemaVersion         int    `json:"schema_version"`
	Transformation        string `json:"transformation"`
	TransformationVersion int    `json:"transformation_version"`
	Mode                  string `json:"mode"`
	InputPath             string `json:"input_path"`
	InputSHA256           string `json:"input_sha256"`
	OutputPath            string `json:"output_path"`
	OutputSHA256          string `json:"output_sha256"`
	InputTokens           int    `json:"input_tokens"`
	OutputTokens          int    `json:"output_tokens"`
	InputUniqueTokens     int    `json:"input_unique_tokens"`
	OutputUniqueTokens    int    `json:"output_unique_tokens"`
	InputLines            int    `json:"input_lines"`
	OutputLines           int    `json:"output_lines"`
	TokenOrderReversed    bool   `json:"token_order_reversed"`
	GlyphOrderReversed    bool   `json:"glyph_order_reversed"`
	LineOrderReversed     bool   `json:"line_order_reversed"`
	Deterministic         bool   `json:"deterministic"`
}

func ValidMode(mode string) bool {
	switch mode {
	case TokenReverse, GlyphReverse, FullReverse:
		return true
	default:
		return false
	}
}

func Transform(input []byte, mode string) ([]byte, Stats, Stats, error) {
	if !ValidMode(mode) {
		return nil, Stats{}, Stats{}, fmt.Errorf("unsupported orientation mode %q", mode)
	}
	check, err := corpusprep.Check(input)
	if err != nil {
		return nil, Stats{}, Stats{}, fmt.Errorf("validate canonical corpus: %w", err)
	}
	if !check.Valid {
		return nil, Stats{}, Stats{}, fmt.Errorf("canonical corpus invalid: %s", check.Reason)
	}

	lines, trailingNewline := splitLines(string(input))
	before := stats(lines)
	for i, line := range lines {
		if line == "" {
			continue
		}
		tokens := strings.Split(line, " ")
		if mode == TokenReverse || mode == FullReverse {
			reverseStrings(tokens)
		}
		if mode == GlyphReverse || mode == FullReverse {
			for j := range tokens {
				tokens[j] = reverseRunes(tokens[j])
			}
		}
		lines[i] = strings.Join(tokens, " ")
	}
	output := []byte(strings.Join(lines, "\n"))
	if trailingNewline {
		output = append(output, '\n')
	}
	after := stats(lines)
	if err := verifyInvariants(mode, before, after); err != nil {
		return nil, Stats{}, Stats{}, err
	}
	return output, before, after, nil
}

func NewManifest(mode, inputPath, outputPath string, input, output []byte, before, after Stats) Manifest {
	inputHash := sha256.Sum256(input)
	outputHash := sha256.Sum256(output)
	tokenOrder, glyphOrder := mode == TokenReverse || mode == FullReverse, mode == GlyphReverse || mode == FullReverse
	return Manifest{
		SchemaVersion: SchemaVersion, Transformation: Transformation, TransformationVersion: TransformationVersion,
		Mode: mode, InputPath: filepath.Clean(inputPath), InputSHA256: hex.EncodeToString(inputHash[:]),
		OutputPath: filepath.Clean(outputPath), OutputSHA256: hex.EncodeToString(outputHash[:]),
		InputTokens: before.Tokens, OutputTokens: after.Tokens,
		InputUniqueTokens: before.UniqueTokens, OutputUniqueTokens: after.UniqueTokens,
		InputLines: before.Lines, OutputLines: after.Lines,
		TokenOrderReversed: tokenOrder, GlyphOrderReversed: glyphOrder, LineOrderReversed: false, Deterministic: true,
	}
}

func MarshalManifest(manifest Manifest) ([]byte, error) {
	data, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		return nil, err
	}
	return append(data, '\n'), nil
}

func splitLines(text string) ([]string, bool) {
	trailingNewline := strings.HasSuffix(text, "\n")
	lines := strings.Split(text, "\n")
	if trailingNewline {
		lines = lines[:len(lines)-1]
	}
	return lines, trailingNewline
}

func stats(lines []string) Stats {
	result := Stats{Lines: len(lines), Counts: make(map[string]int), PerLine: make([]int, len(lines))}
	for i, line := range lines {
		if line == "" {
			continue
		}
		tokens := strings.Split(line, " ")
		result.PerLine[i] = len(tokens)
		result.Tokens += len(tokens)
		for _, token := range tokens {
			result.Counts[token]++
			result.Lengths = append(result.Lengths, len([]rune(token)))
		}
	}
	result.UniqueTokens = len(result.Counts)
	return result
}

func verifyInvariants(mode string, before, after Stats) error {
	if before.Tokens != after.Tokens || before.Lines != after.Lines || !equalInts(before.PerLine, after.PerLine) {
		return fmt.Errorf("%s invariant failed: token or line structure changed", mode)
	}
	if mode == TokenReverse && !equalCounts(before.Counts, after.Counts) {
		return fmt.Errorf("%s invariant failed: token multiset changed", mode)
	}
	if mode == GlyphReverse || mode == FullReverse {
		if !equalLengthMultiset(before.Lengths, after.Lengths) {
			return fmt.Errorf("%s invariant failed: token-length multiset changed", mode)
		}
	}
	return nil
}

func reverseRunes(token string) string {
	runes := []rune(token)
	reverseStrings(runes)
	return string(runes)
}

func reverseStrings[T ~[]E, E any](items T) {
	for left, right := 0, len(items)-1; left < right; left, right = left+1, right-1 {
		items[left], items[right] = items[right], items[left]
	}
}

func equalCounts(left, right map[string]int) bool {
	if len(left) != len(right) {
		return false
	}
	for token, count := range left {
		if right[token] != count {
			return false
		}
	}
	return true
}

func equalInts(left, right []int) bool {
	if len(left) != len(right) {
		return false
	}
	for i := range left {
		if left[i] != right[i] {
			return false
		}
	}
	return true
}

func equalLengthMultiset(left, right []int) bool {
	if len(left) != len(right) {
		return false
	}
	counts := make(map[int]int, len(left))
	for _, n := range left {
		counts[n]++
	}
	for _, n := range right {
		counts[n]--
		if counts[n] < 0 {
			return false
		}
	}
	return true
}
