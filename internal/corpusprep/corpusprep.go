package corpusprep

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"
	"unicode"
	"unicode/utf8"

	"golang.org/x/text/encoding/charmap"
	"golang.org/x/text/transform"
)

const CanonicalCorpusVersion = 1

const (
	EncodingUTF8   = "utf-8"
	EncodingAuto   = "auto"
	EncodingCP1251 = "windows-1251"
	EncodingKOI8R  = "koi8-r"
	EncodingCP866  = "cp866"

	CasePreserve = "preserve"
	CaseLower    = "lower"

	LinePreserve = "preserve"
	LineReflow   = "reflow"
)

type Options struct {
	Encoding           string
	CasePolicy         string
	LinePolicy         string
	DropEmptyLines     bool
	StripLeadingLines  int
	StripTrailingLines int
}

type Corpus struct {
	Lines       [][]string
	Counts      map[string]int
	Occurrences int
	NonEmpty    int
	Transitions int
}

type Stats struct {
	CanonicalCorpusVersion  int            `json:"canonical_corpus_version"`
	InputEncoding           string         `json:"input_encoding"`
	OutputEncoding          string         `json:"output_encoding"`
	CasePolicy              string         `json:"case_policy"`
	PunctuationPolicy       string         `json:"punctuation_policy"`
	WhitespacePolicy        string         `json:"whitespace_policy"`
	LinePolicy              string         `json:"line_policy"`
	InputTokenCount         int            `json:"input_token_count"`
	OutputTokenCount        int            `json:"output_token_count"`
	OutputUniqueTokenCount  int            `json:"output_unique_token_count"`
	LineCount               int            `json:"line_count"`
	ReplacementCharCount    int            `json:"replacement_char_count"`
	InvalidUTF8Count        int            `json:"invalid_utf8_count"`
	ForbiddenControlCount   int            `json:"forbidden_control_count"`
	PunctuationCount        int            `json:"punctuation_count"`
	HighBitByteDistribution map[string]int `json:"high_bit_byte_distribution,omitempty"`
	SuspiciousBytes         map[string]int `json:"suspicious_bytes,omitempty"`
	UTF8Valid               bool           `json:"utf8_valid"`
}

type PrepareManifest struct {
	SchemaVersion          int    `json:"schema_version"`
	CanonicalCorpusVersion int    `json:"canonical_corpus_version"`
	ToolGitCommit          string `json:"tool_git_commit"`
	PreparedBy             string `json:"prepared_by"`
	InputPath              string `json:"input_path"`
	InputSHA256            string `json:"input_sha256"`
	InputSizeBytes         int64  `json:"input_size_bytes"`
	InputEncoding          string `json:"input_encoding"`
	OutputPath             string `json:"output_path"`
	OutputSHA256           string `json:"output_sha256"`
	OutputEncoding         string `json:"output_encoding"`
	CasePolicy             string `json:"case_policy"`
	PunctuationPolicy      string `json:"punctuation_policy"`
	WhitespacePolicy       string `json:"whitespace_policy"`
	LinePolicy             string `json:"line_policy"`
	InputTokenCount        int    `json:"input_token_count"`
	OutputTokenCount       int    `json:"output_token_count"`
	OutputUniqueTokenCount int    `json:"output_unique_token_count"`
	LineCount              int    `json:"line_count"`
	ReplacementCharCount   int    `json:"replacement_char_count"`
	InvalidUTF8Count       int    `json:"invalid_utf8_count"`
	ForbiddenControlCount  int    `json:"forbidden_control_count"`
	Deterministic          bool   `json:"deterministic"`
}

type PrepareResult struct {
	Corpus  Corpus
	Text    []byte
	Stats   Stats
	Decoded string
}

type InspectReport struct {
	Stats            Stats            `json:"stats"`
	Notes            []string         `json:"notes,omitempty"`
	ProposedOutput   *PrepareManifest `json:"proposed_output,omitempty"`
	ProposedChecksum string           `json:"proposed_checksum,omitempty"`
}

type CheckReport struct {
	Valid  bool   `json:"valid"`
	Reason string `json:"reason,omitempty"`
	Stats  Stats  `json:"stats"`
}

func Prepare(input []byte, opts Options, toolGitCommit, inputPath, outputPath string) (PrepareResult, PrepareManifest, error) {
	decoded, stats, err := decodeInput(input, opts)
	if err != nil {
		return PrepareResult{}, PrepareManifest{}, err
	}
	lines, corpus, lineStats, err := normalizeDecoded(decoded, opts)
	if err != nil {
		return PrepareResult{}, PrepareManifest{}, err
	}
	text := serialize(lines, opts.LinePolicy)
	stats.LineCount = lineStats.LineCount
	stats.InputTokenCount = lineStats.InputTokenCount
	stats.OutputTokenCount = corpus.Occurrences
	stats.OutputUniqueTokenCount = len(corpus.Counts)
	stats.PunctuationCount = lineStats.PunctuationCount
	stats.ReplacementCharCount = lineStats.ReplacementCharCount
	stats.InvalidUTF8Count = lineStats.InvalidUTF8Count
	stats.ForbiddenControlCount = lineStats.ForbiddenControlCount
	stats.UTF8Valid = true

	inputSum := sha256.Sum256(input)
	outputSum := sha256.Sum256(text)
	manifest := PrepareManifest{
		SchemaVersion:          1,
		CanonicalCorpusVersion: CanonicalCorpusVersion,
		ToolGitCommit:          toolGitCommit,
		PreparedBy:             preparedBy(toolGitCommit),
		InputPath:              inputPath,
		InputSHA256:            hex.EncodeToString(inputSum[:]),
		InputSizeBytes:         int64(len(input)),
		InputEncoding:          stats.InputEncoding,
		OutputPath:             outputPath,
		OutputSHA256:           hex.EncodeToString(outputSum[:]),
		OutputEncoding:         EncodingUTF8,
		CasePolicy:             stats.CasePolicy,
		PunctuationPolicy:      stats.PunctuationPolicy,
		WhitespacePolicy:       stats.WhitespacePolicy,
		LinePolicy:             stats.LinePolicy,
		InputTokenCount:        stats.InputTokenCount,
		OutputTokenCount:       stats.OutputTokenCount,
		OutputUniqueTokenCount: stats.OutputUniqueTokenCount,
		LineCount:              stats.LineCount,
		ReplacementCharCount:   stats.ReplacementCharCount,
		InvalidUTF8Count:       stats.InvalidUTF8Count,
		ForbiddenControlCount:  stats.ForbiddenControlCount,
		Deterministic:          true,
	}

	return PrepareResult{Corpus: corpus, Text: text, Stats: stats, Decoded: decoded}, manifest, nil
}

func Inspect(input []byte, opts Options, toolGitCommit, inputPath string) (InspectReport, error) {
	decoded, stats, err := decodeInput(input, opts)
	if err != nil {
		return InspectReport{}, err
	}
	lines, corpus, lineStats, err := normalizeDecoded(decoded, opts)
	if err != nil {
		return InspectReport{}, err
	}
	text := serialize(lines, opts.LinePolicy)
	stats.LineCount = lineStats.LineCount
	stats.InputTokenCount = lineStats.InputTokenCount
	stats.OutputTokenCount = corpus.Occurrences
	stats.OutputUniqueTokenCount = len(corpus.Counts)
	stats.PunctuationCount = lineStats.PunctuationCount
	stats.ReplacementCharCount = lineStats.ReplacementCharCount
	stats.InvalidUTF8Count = lineStats.InvalidUTF8Count
	stats.ForbiddenControlCount = lineStats.ForbiddenControlCount
	stats.UTF8Valid = true

	inputSum := sha256.Sum256(input)
	outputSum := sha256.Sum256(text)
	manifest := PrepareManifest{
		SchemaVersion:          1,
		CanonicalCorpusVersion: CanonicalCorpusVersion,
		ToolGitCommit:          toolGitCommit,
		PreparedBy:             preparedBy(toolGitCommit),
		InputPath:              inputPath,
		InputSHA256:            hex.EncodeToString(inputSum[:]),
		InputSizeBytes:         int64(len(input)),
		InputEncoding:          stats.InputEncoding,
		OutputEncoding:         EncodingUTF8,
		CasePolicy:             stats.CasePolicy,
		PunctuationPolicy:      stats.PunctuationPolicy,
		WhitespacePolicy:       stats.WhitespacePolicy,
		LinePolicy:             stats.LinePolicy,
		InputTokenCount:        stats.InputTokenCount,
		OutputTokenCount:       stats.OutputTokenCount,
		OutputUniqueTokenCount: stats.OutputUniqueTokenCount,
		LineCount:              stats.LineCount,
		ReplacementCharCount:   stats.ReplacementCharCount,
		InvalidUTF8Count:       stats.InvalidUTF8Count,
		ForbiddenControlCount:  stats.ForbiddenControlCount,
		Deterministic:          true,
	}

	return InspectReport{
		Stats:            stats,
		Notes:            inspectNotes(input, stats),
		ProposedOutput:   &manifest,
		ProposedChecksum: hex.EncodeToString(outputSum[:]),
	}, nil
}

func Check(input []byte) (CheckReport, error) {
	stats, reason, err := validatePreparedBytes(input)
	if err != nil {
		return CheckReport{}, err
	}
	return CheckReport{Valid: reason == "", Reason: reason, Stats: stats}, nil
}

func LoadCorpus(path string) (Corpus, string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Corpus{}, "", err
	}
	check, err := Check(data)
	if err != nil {
		return Corpus{}, "", err
	}
	if !check.Valid {
		return Corpus{}, "", fmt.Errorf("canonical corpus invalid: %s", check.Reason)
	}
	corpus, err := parseCanonicalCorpus(string(data))
	if err != nil {
		return Corpus{}, "", err
	}
	sum := sha256.Sum256(data)
	return corpus, hex.EncodeToString(sum[:]), nil
}

func ReadManifest(path string) (PrepareManifest, []byte, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return PrepareManifest{}, nil, err
	}
	var manifest PrepareManifest
	if err := json.Unmarshal(data, &manifest); err != nil {
		return PrepareManifest{}, nil, err
	}
	return manifest, data, nil
}

func preparedBy(toolGitCommit string) string {
	if toolGitCommit == "" {
		return "codex_prepare"
	}
	return "codex_prepare@" + toolGitCommit
}

type normalizedStats struct {
	InputTokenCount       int
	LineCount             int
	PunctuationCount      int
	ReplacementCharCount  int
	InvalidUTF8Count      int
	ForbiddenControlCount int
}

func decodeInput(input []byte, opts Options) (string, Stats, error) {
	stats := Stats{
		CanonicalCorpusVersion:  CanonicalCorpusVersion,
		InputEncoding:           normalizeEncoding(opts.Encoding),
		OutputEncoding:          EncodingUTF8,
		CasePolicy:              normalizeCasePolicy(opts.CasePolicy),
		PunctuationPolicy:       "unicode punctuation/symbols and whitespace map to ASCII spaces; tokens are whitespace-delimited",
		WhitespacePolicy:        "tabs, Unicode Zs separators, CRLF/CR, and repeated spaces collapse to canonical ASCII spaces; hidden controls are rejected",
		LinePolicy:              normalizeLinePolicy(opts.LinePolicy),
		HighBitByteDistribution: make(map[string]int),
		SuspiciousBytes:         make(map[string]int),
	}
	encodingName := stats.InputEncoding
	if encodingName == EncodingAuto {
		if utf8.Valid(input) {
			encodingName = EncodingUTF8
		} else {
			return "", stats, fmt.Errorf("auto encoding detection is conservative; input is not valid UTF-8, so pass an explicit -encoding")
		}
		stats.InputEncoding = encodingName
	}
	switch encodingName {
	case EncodingUTF8:
		if !utf8.Valid(input) {
			stats.InvalidUTF8Count = countInvalidUTF8(input)
			return "", stats, fmt.Errorf("invalid UTF-8 input")
		}
		for _, r := range string(input) {
			if r == utf8.RuneError {
				stats.ReplacementCharCount++
			}
		}
		recordSuspiciousRawBytes(input, &stats)
		return string(input), stats, nil
	case EncodingCP1251:
		return decodeCharmap(input, charmap.Windows1251.NewDecoder(), stats)
	case EncodingKOI8R:
		return decodeCharmap(input, charmap.KOI8R.NewDecoder(), stats)
	case EncodingCP866:
		return decodeCharmap(input, charmap.CodePage866.NewDecoder(), stats)
	default:
		return "", stats, fmt.Errorf("unsupported encoding %q", opts.Encoding)
	}
}

func decodeCharmap(input []byte, decoder transform.Transformer, stats Stats) (string, Stats, error) {
	for _, b := range input {
		if b&0x80 != 0 {
			stats.HighBitByteDistribution[fmt.Sprintf("0x%02x", b)]++
		}
	}
	recordSuspiciousRawBytes(input, &stats)
	decoded, err := io.ReadAll(transform.NewReader(bytes.NewReader(input), decoder))
	if err != nil {
		return "", stats, err
	}
	text := string(decoded)
	for _, r := range text {
		if r == utf8.RuneError {
			stats.ReplacementCharCount++
		}
	}
	return text, stats, nil
}

func recordSuspiciousRawBytes(input []byte, stats *Stats) {
	for _, b := range input {
		switch b {
		case 0x00, 0x7f:
			stats.SuspiciousBytes[fmt.Sprintf("0x%02x", b)]++
		case 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x0b, 0x0c:
			stats.SuspiciousBytes[fmt.Sprintf("0x%02x", b)]++
		case 0x0e, 0x0f, 0x10, 0x11, 0x12, 0x13, 0x14, 0x15, 0x16, 0x17, 0x18, 0x19, 0x1a, 0x1b, 0x1c, 0x1d, 0x1e, 0x1f:
			stats.SuspiciousBytes[fmt.Sprintf("0x%02x", b)]++
		}
	}
}

func normalizeDecoded(decoded string, opts Options) ([]string, Corpus, normalizedStats, error) {
	resultStats := normalizedStats{}
	if opts.CasePolicy == "" {
		opts.CasePolicy = CaseLower
	}
	if opts.LinePolicy == "" {
		opts.LinePolicy = LinePreserve
	}
	if opts.CasePolicy != CasePreserve && opts.CasePolicy != CaseLower {
		return nil, Corpus{}, resultStats, fmt.Errorf("unsupported case policy %q", opts.CasePolicy)
	}
	if opts.LinePolicy != LinePreserve && opts.LinePolicy != LineReflow {
		return nil, Corpus{}, resultStats, fmt.Errorf("unsupported line policy %q", opts.LinePolicy)
	}
	if opts.DropEmptyLines {
		// Canonical output always omits empty lines; the flag is an explicit
		// request to keep the mechanical policy obvious to callers.
	}
	if opts.StripLeadingLines > 0 || opts.StripTrailingLines > 0 {
		decoded = stripLogicalLines(decoded, opts.StripLeadingLines, opts.StripTrailingLines)
	}
	if opts.CasePolicy == CaseLower {
		decoded = strings.ToLower(decoded)
	}
	logicalLines := splitLogicalLines(decoded)
	if len(logicalLines) == 0 {
		return nil, Corpus{}, resultStats, fmt.Errorf("input is empty")
	}
	corpus := Corpus{Counts: make(map[string]int)}
	var normalizedLines []string
	for _, rawLine := range logicalLines {
		inputTokens := strings.Fields(rawLine)
		resultStats.InputTokenCount += len(inputTokens)
		if strings.TrimSpace(rawLine) == "" {
			continue
		}
		tokens, linePunct, replCount, invalidCount, controlCount, err := tokenizeLine(rawLine)
		if err != nil {
			return nil, Corpus{}, resultStats, err
		}
		resultStats.PunctuationCount += linePunct
		resultStats.ReplacementCharCount += replCount
		resultStats.InvalidUTF8Count += invalidCount
		resultStats.ForbiddenControlCount += controlCount
		if len(tokens) == 0 {
			continue
		}
		normalized := strings.Join(tokens, " ")
		normalizedLines = append(normalizedLines, normalized)
		addLine(&corpus, len(normalizedLines), tokens)
	}
	if len(corpus.Lines) == 0 {
		return nil, Corpus{}, resultStats, fmt.Errorf("canonical corpus is empty after normalization")
	}
	resultStats.LineCount = len(corpus.Lines)
	if opts.LinePolicy == LineReflow {
		var stream []string
		for _, line := range normalizedLines {
			stream = append(stream, strings.Fields(line)...)
		}
		normalizedLines = []string{strings.Join(stream, " ")}
		corpus = Corpus{Counts: make(map[string]int)}
		if len(stream) == 0 {
			return nil, Corpus{}, resultStats, fmt.Errorf("canonical corpus is empty after reflow")
		}
		addLine(&corpus, 1, stream)
		resultStats.LineCount = 1
	}
	return normalizedLines, corpus, resultStats, nil
}

func tokenizeLine(line string) ([]string, int, int, int, int, error) {
	var tokens []string
	var current strings.Builder
	punctuationCount := 0
	replacementCount := 0
	invalidCount := 0
	controlCount := 0
	flush := func() {
		if current.Len() > 0 {
			tokens = append(tokens, current.String())
			current.Reset()
		}
	}
	for _, r := range line {
		switch {
		case r == utf8.RuneError:
			invalidCount++
			return nil, 0, replacementCount, invalidCount, controlCount, fmt.Errorf("invalid UTF-8 replacement rune encountered")
		case r == '\uFFFD':
			replacementCount++
			return nil, 0, replacementCount, invalidCount, controlCount, fmt.Errorf("U+FFFD is not permitted")
		case r == '\uFEFF':
			return nil, 0, replacementCount, invalidCount, controlCount, fmt.Errorf("BOM character is not permitted")
		case r == '\t' || r == ' ' || unicode.In(r, unicode.Zs):
			flush()
		case unicode.IsControl(r):
			controlCount++
			return nil, 0, replacementCount, invalidCount, controlCount, fmt.Errorf("forbidden control character U+%04X", r)
		case unicode.IsPunct(r) || unicode.IsSymbol(r):
			punctuationCount++
			flush()
		default:
			current.WriteRune(r)
		}
	}
	flush()
	return tokens, punctuationCount, replacementCount, invalidCount, controlCount, nil
}

func validatePreparedBytes(input []byte) (Stats, string, error) {
	if !utf8.Valid(input) {
		return Stats{InvalidUTF8Count: countInvalidUTF8(input)}, "invalid UTF-8", nil
	}
	var reason string
	stats := Stats{
		CanonicalCorpusVersion: CanonicalCorpusVersion,
		CasePolicy:             CaseLower,
		PunctuationPolicy:      "unicode punctuation/symbols and whitespace map to ASCII spaces; tokens are whitespace-delimited",
		WhitespacePolicy:       "tabs, Unicode Zs separators, CRLF/CR, and repeated spaces collapse to canonical ASCII spaces; hidden controls are rejected",
		LinePolicy:             LinePreserve,
		OutputEncoding:         EncodingUTF8,
		UTF8Valid:              true,
	}
	lines := splitLogicalLines(string(input))
	if len(lines) > 0 && lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}
	if len(lines) == 0 {
		return stats, "empty corpus", nil
	}
	corpusTokens := 0
	unique := make(map[string]struct{})
	lineCount := 0
	for _, line := range lines {
		if line == "" {
			lineCount++
			continue
		}
		if strings.ContainsRune(line, '\uFFFD') {
			return stats, "U+FFFD is present", nil
		}
		if strings.ContainsRune(line, '\t') || strings.ContainsRune(line, '\r') || strings.ContainsRune(line, '\u00A0') {
			reason = "non-canonical whitespace"
			break
		}
		if !isCanonicalLine(line) {
			reason = "punctuation, control characters, or repeated spaces are present"
			break
		}
		fields := strings.Fields(line)
		if len(fields) == 0 || strings.Join(fields, " ") != line {
			reason = "whitespace/tokenization invariant failed"
			break
		}
		for _, token := range fields {
			corpusTokens++
			unique[token] = struct{}{}
		}
		lineCount++
	}
	if reason != "" {
		return stats, reason, nil
	}
	if lineCount == 0 || corpusTokens == 0 {
		return stats, "empty corpus", nil
	}
	stats.LineCount = lineCount
	stats.OutputTokenCount = corpusTokens
	stats.OutputUniqueTokenCount = len(unique)
	return stats, "", nil
}

func isCanonicalLine(line string) bool {
	for _, r := range line {
		switch {
		case r == ' ':
			continue
		case r == '\uFFFD':
			return false
		case r == '\uFEFF':
			return false
		case unicode.IsControl(r):
			return false
		case unicode.IsPunct(r) || unicode.IsSymbol(r):
			return false
		}
	}
	return true
}

func parseCanonicalCorpus(text string) (Corpus, error) {
	corpus := Corpus{Counts: make(map[string]int)}
	lines := splitLogicalLines(text)
	if len(lines) > 0 && lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}
	for _, line := range lines {
		if line == "" {
			addLine(&corpus, len(corpus.Lines)+1, nil)
			continue
		}
		fields := strings.Fields(line)
		if len(fields) == 0 || strings.Join(fields, " ") != line {
			return Corpus{}, fmt.Errorf("canonical corpus line is not normalized")
		}
		addLine(&corpus, len(corpus.Lines)+1, fields)
	}
	if corpus.Occurrences == 0 {
		return Corpus{}, fmt.Errorf("canonical corpus is empty")
	}
	return corpus, nil
}

func addLine(corpus *Corpus, id int, tokens []string) {
	line := append([]string(nil), tokens...)
	corpus.Lines = append(corpus.Lines, line)
	if len(line) == 0 {
		return
	}
	corpus.NonEmpty++
	corpus.Occurrences += len(line)
	corpus.Transitions += len(line) - 1
	for _, token := range line {
		corpus.Counts[token]++
	}
}

func serialize(lines []string, linePolicy string) []byte {
	if linePolicy == LineReflow {
		if len(lines) == 0 {
			return []byte{}
		}
		return append([]byte(strings.Join(strings.Fields(strings.Join(lines, " ")), " ")), '\n')
	}
	if len(lines) == 0 {
		return []byte{}
	}
	var b strings.Builder
	for _, line := range lines {
		b.WriteString(line)
		b.WriteByte('\n')
	}
	return []byte(b.String())
}

func splitLogicalLines(text string) []string {
	var lines []string
	var current strings.Builder
	for i := 0; i < len(text); {
		r, size := utf8.DecodeRuneInString(text[i:])
		switch r {
		case '\r':
			lines = append(lines, current.String())
			current.Reset()
			i += size
			if i < len(text) {
				next, nextSize := utf8.DecodeRuneInString(text[i:])
				if next == '\n' {
					i += nextSize
				}
			}
		case '\n', '\u2028', '\u2029':
			lines = append(lines, current.String())
			current.Reset()
			i += size
		default:
			current.WriteRune(r)
			i += size
		}
	}
	lines = append(lines, current.String())
	return lines
}

func stripLogicalLines(text string, leading, trailing int) string {
	lines := splitLogicalLines(text)
	start := leading
	if start > len(lines) {
		start = len(lines)
	}
	end := len(lines) - trailing
	if end < start {
		end = start
	}
	return strings.Join(lines[start:end], "\n")
}

func normalizeEncoding(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	switch value {
	case "", EncodingUTF8:
		return EncodingUTF8
	case "windows-1251", "cp1251", "cp-1251":
		return EncodingCP1251
	case "koi8-r", "koi8r":
		return EncodingKOI8R
	case "cp866", "ibm866", "windows-866":
		return EncodingCP866
	case EncodingAuto:
		return EncodingAuto
	default:
		return value
	}
}

func normalizeCasePolicy(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	if value == "" {
		return CaseLower
	}
	return value
}

func normalizeLinePolicy(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	if value == "" {
		return LinePreserve
	}
	return value
}

func inspectNotes(input []byte, stats Stats) []string {
	var notes []string
	if stats.InputEncoding == EncodingUTF8 && !utf8.Valid(input) {
		notes = append(notes, "input is not valid UTF-8")
	}
	if stats.InvalidUTF8Count > 0 {
		notes = append(notes, fmt.Sprintf("%d invalid UTF-8 sequence(s) detected", stats.InvalidUTF8Count))
	}
	if stats.ReplacementCharCount > 0 {
		notes = append(notes, fmt.Sprintf("%d U+FFFD rune(s) detected", stats.ReplacementCharCount))
	}
	if stats.ForbiddenControlCount > 0 {
		notes = append(notes, fmt.Sprintf("%d forbidden control rune(s) detected", stats.ForbiddenControlCount))
	}
	if len(stats.HighBitByteDistribution) > 0 {
		keys := make([]string, 0, len(stats.HighBitByteDistribution))
		for k := range stats.HighBitByteDistribution {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		var parts []string
		for _, k := range keys {
			parts = append(parts, fmt.Sprintf("%s=%d", k, stats.HighBitByteDistribution[k]))
		}
		notes = append(notes, "high-bit byte distribution: "+strings.Join(parts, ", "))
	}
	return notes
}

func countInvalidUTF8(input []byte) int {
	count := 0
	for len(input) > 0 {
		_, size := utf8.DecodeRune(input)
		if size == 1 && input[0] >= utf8.RuneSelf {
			count++
		}
		input = input[size:]
	}
	return count
}
