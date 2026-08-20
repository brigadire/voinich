package corpustransform

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
)

// Manifest is the reproducibility record written alongside every output
// corpus as <output>.transform.json (task46 section 11).
type Manifest struct {
	SchemaVersion        int    `json:"schema_version"`
	Transformer          string `json:"transformer"`
	TransformerGitCommit string `json:"transformer_git_commit"`

	InputPath         string `json:"input_path"`
	InputSHA256       string `json:"input_sha256"`
	InputTokenCount   int    `json:"input_token_count"`
	InputUniqueTokens int    `json:"input_unique_tokens"`

	OutputPath         string `json:"output_path"`
	OutputSHA256       string `json:"output_sha256"`
	OutputTokenCount   int    `json:"output_token_count"`
	OutputUniqueTokens int    `json:"output_unique_tokens"`

	Method     string         `json:"method"`
	Parameters map[string]any `json:"parameters"`
	Seed       int64          `json:"seed"`

	Deterministic bool `json:"deterministic"`

	LinePolicy string `json:"line_policy"`

	// Transposition-only.
	Width           *int   `json:"width,omitempty"`
	Order           string `json:"order,omitempty"`
	Rounds          *int   `json:"rounds,omitempty"`
	RemainderPolicy string `json:"remainder_policy,omitempty"`

	// Homophonic-only.
	HomophoneModel string `json:"homophone_model,omitempty"`
	Homophones     *int   `json:"homophones,omitempty"`
	Selection      string `json:"selection,omitempty"`
	MappingSHA256  string `json:"mapping_sha256,omitempty"`
}

func intPtr(v int) *int { return &v }

// NewTranspositionManifest builds the manifest for one transposition run.
func NewTranspositionManifest(gitCommit, inputPath string, input Corpus, outputPath string, output []byte, outputTokens []string, p TranspositionParams, linePolicy string) Manifest {
	return Manifest{
		SchemaVersion:        SchemaVersion,
		Transformer:          Transformer,
		TransformerGitCommit: gitCommit,

		InputPath:         inputPath,
		InputSHA256:       input.InputSHA256Hex,
		InputTokenCount:   len(input.Tokens),
		InputUniqueTokens: len(TokenCounts(input.Tokens)),

		OutputPath:         outputPath,
		OutputSHA256:       ShaBytes(output),
		OutputTokenCount:   len(outputTokens),
		OutputUniqueTokens: len(TokenCounts(outputTokens)),

		Method: MethodTransposition,
		Parameters: map[string]any{
			"width":            p.Width,
			"order":            p.Order,
			"rounds":           p.Round,
			"remainder_policy": RemainderPolicy,
		},
		Seed:          p.Seed,
		Deterministic: true,
		LinePolicy:    linePolicy,

		Width:           intPtr(p.Width),
		Order:           p.Order,
		Rounds:          intPtr(p.Round),
		RemainderPolicy: RemainderPolicy,
	}
}

// NewHomophonicManifest builds the manifest for one homophonic run.
func NewHomophonicManifest(gitCommit, inputPath string, input Corpus, outputPath string, output []byte, outputTokens []string, p HomophonicParams, linePolicy, mappingSHA256 string) Manifest {
	return Manifest{
		SchemaVersion:        SchemaVersion,
		Transformer:          Transformer,
		TransformerGitCommit: gitCommit,

		InputPath:         inputPath,
		InputSHA256:       input.InputSHA256Hex,
		InputTokenCount:   len(input.Tokens),
		InputUniqueTokens: len(TokenCounts(input.Tokens)),

		OutputPath:         outputPath,
		OutputSHA256:       ShaBytes(output),
		OutputTokenCount:   len(outputTokens),
		OutputUniqueTokens: len(TokenCounts(outputTokens)),

		Method: MethodHomophonic,
		Parameters: map[string]any{
			"homophone_model": p.Model,
			"homophones":      p.Homophones,
			"selection":       p.Selection,
			"weight_scheme":   WeightScheme,
		},
		Seed:          p.Seed,
		Deterministic: true,
		LinePolicy:    linePolicy,

		HomophoneModel: p.Model,
		Homophones:     intPtr(p.Homophones),
		Selection:      p.Selection,
		MappingSHA256:  mappingSHA256,
	}
}

// MarshalManifest renders m as indented JSON with a trailing newline,
// matching the repository's existing manifest-writing convention.
func MarshalManifest(m Manifest) ([]byte, error) {
	data, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return nil, err
	}
	return append(data, '\n'), nil
}

// MarshalMappingTSV renders mapping as a canonical, deterministically
// ordered TSV: sorted by plaintext token (mapping.Vocabulary order), then
// by homophone index within each token (task46 section 12).
func MarshalMappingTSV(mapping Mapping) []byte {
	var b strings.Builder
	b.WriteString("plaintext_token\tcipher_token\tprobability\n")
	for _, t := range mapping.Vocabulary {
		for _, e := range mapping.Entries[t] {
			b.WriteString(t)
			b.WriteByte('\t')
			b.WriteString(e.CipherToken)
			b.WriteByte('\t')
			b.WriteString(strconv.FormatFloat(e.Probability, 'g', -1, 64))
			b.WriteByte('\n')
		}
	}
	return []byte(b.String())
}

// TranspositionSanity is the human-readable invariant report printed after
// a transposition run (task46 section 15).
type TranspositionSanity struct {
	InputTokens          int
	OutputTokens         int
	InputUniqueTokens    int
	OutputUniqueTokens   int
	MultisetPreserved    bool
	PositionsChangedFrac float64
	OutputSHA256         string
}

func NewTranspositionSanity(input, output []string, outputSHA256 string) TranspositionSanity {
	changed := 0
	n := len(input)
	for i := 0; i < n && i < len(output); i++ {
		if input[i] != output[i] {
			changed++
		}
	}
	frac := 0.0
	if n > 0 {
		frac = float64(changed) / float64(n)
	}
	return TranspositionSanity{
		InputTokens:          n,
		OutputTokens:         len(output),
		InputUniqueTokens:    len(TokenCounts(input)),
		OutputUniqueTokens:   len(TokenCounts(output)),
		MultisetPreserved:    MultisetEqual(input, output),
		PositionsChangedFrac: frac,
		OutputSHA256:         outputSHA256,
	}
}

func (s TranspositionSanity) String() string {
	yn := "NO"
	if s.MultisetPreserved {
		yn = "YES"
	}
	return fmt.Sprintf(
		"input tokens: %d\noutput tokens: %d\nunique input tokens: %d\nunique output tokens: %d\ntoken multiset preserved: %s\npositions changed fraction: %.6f\noutput SHA256: %s\n",
		s.InputTokens, s.OutputTokens, s.InputUniqueTokens, s.OutputUniqueTokens, yn, s.PositionsChangedFrac, s.OutputSHA256,
	)
}

// HomophonicSanity is the human-readable invariant report printed after a
// homophonic run (task46 section 15).
type HomophonicSanity struct {
	InputTokens        int
	OutputTokens       int
	PlaintextVocab     int
	CipherVocab        int
	AvgHomophonesToken float64
	MappingCollisions  int
	OutputSHA256       string
}

func NewHomophonicSanity(input, output []string, mapping Mapping, outputSHA256 string) HomophonicSanity {
	cipherVocab := len(TokenCounts(output))
	total := 0
	for _, t := range mapping.Vocabulary {
		total += len(mapping.Entries[t])
	}
	avg := 0.0
	if len(mapping.Vocabulary) > 0 {
		avg = float64(total) / float64(len(mapping.Vocabulary))
	}
	return HomophonicSanity{
		InputTokens:        len(input),
		OutputTokens:       len(output),
		PlaintextVocab:     len(mapping.Vocabulary),
		CipherVocab:        cipherVocab,
		AvgHomophonesToken: avg,
		MappingCollisions:  MappingCollisions(mapping),
		OutputSHA256:       outputSHA256,
	}
}

func (s HomophonicSanity) String() string {
	return fmt.Sprintf(
		"input tokens: %d\noutput tokens: %d\nplaintext vocabulary: %d\ncipher vocabulary: %d\naverage homophones/token: %.6f\nmapping collisions: %d\noutput SHA256: %s\n",
		s.InputTokens, s.OutputTokens, s.PlaintextVocab, s.CipherVocab, s.AvgHomophonesToken, s.MappingCollisions, s.OutputSHA256,
	)
}
