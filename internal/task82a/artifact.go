package task82a

import (
	"math"
	"sort"
	"strings"

	"zcore.dev/voinich/internal/mnemonicspace"
)

type CorpusInfo struct {
	ID     string `json:"id"`
	Path   string `json:"path"`
	SHA256 string `json:"sha256"`
}

type BoundaryProvenance struct {
	Token string `json:"token"`
	Line  string `json:"line"`
	Page  string `json:"page"`
}

type F2Result struct {
	CorpusFileChecksum string        `json:"corpus_file_sha256"`
	CorpusFilePath     string        `json:"corpus_file_path"`
	Metrics            []F2RawMetric `json:"metrics"`
}

// Collision is a within-job group of >=2 chunks whose observable document
// checksum coincides despite distinct intended content (task82a.txt
// sec.39).
type Collision struct {
	Checksum     string   `json:"checksum"`
	ChunkIndices []int    `json:"chunk_indices"`
	IntendedIDs  []string `json:"intended_ids"`
}

// CorpusMetrics extends internal/task82's frozen entropy/repetition
// estimators (unchanged definitions, task82a.txt sec.41) to the full
// assembled symbol stream and to the assembler's token stream.
type CorpusMetrics struct {
	SymbolCount            int     `json:"observable_symbol_count"`
	TokenCount             int     `json:"observable_token_count"`
	DistinctSymbols        int     `json:"distinct_observable_symbols"`
	DistinctTokens         int     `json:"distinct_observable_tokens"`
	SymbolEntropyBits      float64 `json:"symbol_entropy_plugin_bits"`
	ConditionalEntropyBits float64 `json:"conditional_symbol_entropy_plugin_bits"`
	TokenEntropyBits       float64 `json:"token_entropy_plugin_bits"`
	RepetitionRate         float64 `json:"adjacent_symbol_repetition_rate"`
	TokenRepetitionRate    float64 `json:"adjacent_token_repetition_rate"`
}

// Artifact is one frozen Task82a raw job (task82a.txt sec.33, 69).
type Artifact struct {
	Schema             string                           `json:"schema"`
	Implementation     string                           `json:"implementation_version"`
	FreezeVersion      string                           `json:"freeze_version"`
	Job                ManifestJob                      `json:"job"`
	Family             string                           `json:"family"`
	HistoricalStatus   string                           `json:"historical_status"`
	Corpus             CorpusInfo                       `json:"input_corpus"`
	LocalRunCount      int                              `json:"local_run_count"`
	StatePolicy        string                           `json:"state_policy"`
	ScalingPolicyID    string                           `json:"scaling_policy_id"`
	BoundaryProvenance BoundaryProvenance               `json:"boundary_provenance"`
	Chunks             []ChunkSummary                   `json:"chunks"`
	Document           mnemonicspace.ObservableDocument `json:"observable_document"`
	LocalRecoveries    []SampledRecovery                `json:"local_recoveries"`
	LocalCollisions    []Collision                      `json:"local_collisions"`
	F2                 F2Result                         `json:"f2"`
	Metrics            CorpusMetrics                    `json:"metrics"`
	DocumentSHA256     string                           `json:"observable_checksum"`
	Warnings           []string                         `json:"warnings"`
	RuntimeNS          int64                            `json:"runtime_ns"`
	SoftwareVersion    string                           `json:"software_version"`
}

func computeCorpusMetrics(symbols, tokens []string) CorpusMetrics {
	m := CorpusMetrics{SymbolCount: len(symbols), TokenCount: len(tokens)}
	sc := map[string]int{}
	for _, s := range symbols {
		sc[s]++
	}
	m.DistinctSymbols = len(sc)
	m.SymbolEntropyBits = entropyOf(sc, len(symbols))
	m.ConditionalEntropyBits = conditionalEntropyOf(symbols)
	if len(symbols) > 1 {
		repeats := 0
		for i := 1; i < len(symbols); i++ {
			if symbols[i] == symbols[i-1] {
				repeats++
			}
		}
		m.RepetitionRate = float64(repeats) / float64(len(symbols)-1)
	}
	tc := map[string]int{}
	for _, t := range tokens {
		tc[t]++
	}
	m.DistinctTokens = len(tc)
	m.TokenEntropyBits = entropyOf(tc, len(tokens))
	if len(tokens) > 1 {
		repeats := 0
		for i := 1; i < len(tokens); i++ {
			if tokens[i] == tokens[i-1] {
				repeats++
			}
		}
		m.TokenRepetitionRate = float64(repeats) / float64(len(tokens)-1)
	}
	return m
}

func entropyOf(counts map[string]int, n int) float64 {
	if n == 0 {
		return 0
	}
	keys := make([]string, 0, len(counts))
	for k := range counts {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	h := 0.0
	for _, k := range keys {
		p := float64(counts[k]) / float64(n)
		h -= p * math.Log2(p)
	}
	return h
}

func conditionalEntropyOf(s []string) float64 {
	if len(s) < 2 {
		return 0
	}
	prefix := map[string]int{}
	pair := map[string]int{}
	for i := 1; i < len(s); i++ {
		prefix[s[i-1]]++
		pair[s[i-1]+"\x00"+s[i]]++
	}
	keys := make([]string, 0, len(pair))
	for k := range pair {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	h := 0.0
	n := float64(len(s) - 1)
	for _, k := range keys {
		v := pair[k]
		p := float64(v) / n
		pre := strings.SplitN(k, "\x00", 2)[0]
		h -= p * math.Log2(float64(v)/float64(prefix[pre]))
	}
	return h
}
