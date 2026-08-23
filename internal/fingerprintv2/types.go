// Package fingerprintv2 implements the deterministic lexical-paradigm
// block of the proposed (not frozen) Fingerprint v2.
package fingerprintv2

import (
	"fmt"
	"sort"
)

const (
	Version          = "fingerprint-v2-lexical-paradigms-v1"
	MetricVersion    = "lp-ef-v1"
	NullModelVersion = "c-grammar-v1"
)

// Config is the declarative YAML configuration accepted by the CLI.
// Paths are supplied by callers and are never inferred from an installation.
type Config struct {
	Version             string        `yaml:"version" json:"version"`
	OutputDir           string        `yaml:"output_dir" json:"output_dir"`
	Primary             CorpusConfig  `yaml:"primary" json:"primary"`
	Controls            []NamedCorpus `yaml:"controls" json:"controls"`
	Seed                int64         `yaml:"seed" json:"seed"`
	Repetitions         int           `yaml:"repetitions" json:"repetitions"`
	MinRuleSupport      int           `yaml:"min_rule_support" json:"min_rule_support"`
	Alpha               float64       `yaml:"alpha" json:"alpha"`
	GraphSwaps          int           `yaml:"graph_swaps" json:"graph_swaps"`
	DiagnosticTolerance float64       `yaml:"diagnostic_tolerance" json:"diagnostic_tolerance"`
	Grammar             GrammarConfig `yaml:"grammar" json:"grammar"`
}

type CorpusConfig struct {
	ID        string `yaml:"id" json:"id"`
	Path      string `yaml:"path" json:"path"`
	GlyphMode string `yaml:"glyph_mode" json:"glyph_mode"` // eva or natural
	IVTFFPath string `yaml:"ivtff_path,omitempty" json:"ivtff_path,omitempty"`
}

type NamedCorpus struct {
	Name   string       `yaml:"name" json:"name"`
	Corpus CorpusConfig `yaml:"corpus" json:"corpus"`
}

type GrammarConfig struct {
	Modes []string `yaml:"modes" json:"modes"` // structure-preserving, frequency-aware
}

func (c Config) normalized() (Config, error) {
	if c.Version == "" {
		c.Version = Version
	}
	if c.Version != Version {
		return c, fmt.Errorf("unsupported config version %q (want %q)", c.Version, Version)
	}
	if c.OutputDir == "" {
		return c, fmt.Errorf("output_dir is required")
	}
	if c.Primary.ID == "" {
		c.Primary.ID = "primary"
	}
	if c.Primary.Path == "" {
		return c, fmt.Errorf("primary.path is required")
	}
	if c.Repetitions <= 0 {
		c.Repetitions = 100
	}
	if c.MinRuleSupport <= 0 {
		c.MinRuleSupport = 3
	}
	if c.Alpha <= 0 || c.Alpha >= 1 {
		c.Alpha = 0.05
	}
	if c.GraphSwaps <= 0 {
		c.GraphSwaps = 10
	}
	if c.DiagnosticTolerance <= 0 {
		c.DiagnosticTolerance = 0.20
	}
	if len(c.Grammar.Modes) == 0 {
		c.Grammar.Modes = []string{"structure-preserving", "frequency-aware"}
	}
	seenModes := map[string]bool{}
	for _, mode := range c.Grammar.Modes {
		if mode != "structure-preserving" && mode != "frequency-aware" {
			return c, fmt.Errorf("unsupported grammar mode %q", mode)
		}
		if seenModes[mode] {
			return c, fmt.Errorf("grammar mode %q is repeated", mode)
		}
		seenModes[mode] = true
	}
	for i := range c.Controls {
		if c.Controls[i].Name == "" {
			return c, fmt.Errorf("controls[%d].name is required", i)
		}
		if c.Controls[i].Corpus.ID == "" {
			c.Controls[i].Corpus.ID = c.Controls[i].Name
		}
		if c.Controls[i].Corpus.Path == "" {
			return c, fmt.Errorf("controls[%d].corpus.path is required", i)
		}
	}
	return c, nil
}

func validateGlyphMode(v string) (string, error) {
	if v == "" {
		return "eva", nil
	}
	if v != "eva" && v != "natural" {
		return "", fmt.Errorf("glyph_mode must be eva or natural, got %q", v)
	}
	return v, nil
}

type Provenance struct {
	ImplementationCommit string `json:"implementation_commit"`
	MetricVersion        string `json:"metric_version"`
	NullModelVersion     string `json:"null_model_version"`
	Seed                 int64  `json:"seed"`
	Repetitions          int    `json:"repetitions"`
	PreprocessingProfile string `json:"preprocessing_profile"`
	GeneratorSettings    string `json:"generator_settings"`
}

type CorpusInfo struct {
	ID                string `json:"id"`
	Path              string `json:"path"`
	SHA256            string `json:"sha256"`
	TokenCount        int    `json:"token_count"`
	VocabularySize    int    `json:"vocabulary_size"`
	GlyphMode         string `json:"glyph_mode"`
	MetadataAlignment string `json:"metadata_alignment"`
	LineMetadata      bool   `json:"line_metadata"`
	PageMetadata      bool   `json:"page_metadata"`
	Preprocessing     string `json:"preprocessing"`
}

type RuleSupport struct {
	Rule    string  `json:"rule"`
	Support int     `json:"support"`
	Share   float64 `json:"share"`
}

type LP1Result struct {
	DirectedPairCount int           `json:"directed_pair_count"`
	RuleCount         int           `json:"rule_count"`
	SupportGini       float64       `json:"support_gini"`
	TopRuleShare      float64       `json:"top_rule_share"`
	SupportThreshold  int           `json:"support_threshold"`
	ProductiveSupport int           `json:"productive_support"`
	Rules             []RuleSupport `json:"rules"`
}

type NullTest struct {
	ID            string  `json:"id"`
	NullModel     string  `json:"null_model"`
	Observed      float64 `json:"observed"`
	NullMean      float64 `json:"null_mean"`
	NullSD        float64 `json:"null_sd"`
	EffectSize    float64 `json:"effect_size"`
	EffectDefined bool    `json:"effect_defined"`
	PValue        float64 `json:"p_value"`
	QValue        float64 `json:"q_value"`
	Replicates    int     `json:"replicates"`
	Alternative   string  `json:"alternative"`
}

type LP2Result struct {
	Statistic         string     `json:"statistic"`
	Tests             []NullTest `json:"tests"`
	ProductiveRules   []string   `json:"productive_rules"`
	ProductivityState string     `json:"productivity_state"`
}

type FamilySummary struct {
	Size        int     `json:"size"`
	Branching   float64 `json:"branching"`
	Depth       int     `json:"depth"`
	Overlap     float64 `json:"overlap"`
	DepthMethod string  `json:"depth_method"`
}

type LocalityResult struct {
	Available    bool      `json:"available"`
	SameLineRate float64   `json:"same_line_rate"`
	SamePageRate *float64  `json:"same_page_rate,omitempty"`
	GlobalNull   *NullTest `json:"global_null,omitempty"`
	FamilyCount  int       `json:"family_count"`
}

type LP3Result struct {
	ProductiveRuleCount int             `json:"productive_rule_count"`
	Families            []FamilySummary `json:"families"`
	SmallFamilyCount    int             `json:"small_family_count"`
	Locality            LocalityResult  `json:"locality"`
}

type AttachmentResult struct {
	NormalizedMI float64    `json:"normalized_mutual_information"`
	Eligible     int        `json:"eligible_tokens"`
	Excluded     int        `json:"excluded_tokens"`
	Permutation  NullTest   `json:"permutation"`
	GrammarTests []NullTest `json:"grammar_tests"`
}

type LP4Result struct {
	ZoneConvention string           `json:"zone_convention"`
	Prefix         AttachmentResult `json:"prefix"`
	Suffix         AttachmentResult `json:"suffix"`
}

type Count struct {
	Value int `json:"value"`
	Count int `json:"count"`
}

type EF1Result struct {
	VertexCount         int     `json:"vertex_count"`
	EdgeCount           int     `json:"edge_count"`
	IsolateCount        int     `json:"isolate_count"`
	IsolateShare        float64 `json:"isolate_share"`
	ComponentCount      int     `json:"component_count"`
	GiantComponent      int     `json:"giant_component"`
	GiantComponentShare float64 `json:"giant_component_share"`
	DegreeDistribution  []Count `json:"degree_distribution"`
	ComponentSizes      []Count `json:"component_sizes"`
}

type EF2Result struct {
	GlobalClustering   float64  `json:"global_clustering"`
	Triangles          int      `json:"triangles"`
	Paths3             int      `json:"paths3"`
	Cycles4            int      `json:"cycles4"`
	ConfigurationTest  NullTest `json:"configuration_test"`
	ControlDescription string   `json:"control_description"`
}

type EF3Result struct {
	SpearmanDegreeLogFrequency float64  `json:"spearman_degree_log_frequency"`
	FrequencyControl           NullTest `json:"frequency_control"`
}

type EF4Result struct {
	Verdict string     `json:"verdict"`
	Tests   []NullTest `json:"grammar_tests"`
	Reason  string     `json:"reason"`
}

type Metrics struct {
	LP1 LP1Result `json:"lp1"`
	LP2 LP2Result `json:"lp2"`
	LP3 LP3Result `json:"lp3"`
	LP4 LP4Result `json:"lp4"`
	EF1 EF1Result `json:"ef1"`
	EF2 EF2Result `json:"ef2"`
	EF3 EF3Result `json:"ef3"`
	EF4 EF4Result `json:"ef4"`
}

type DistributionDiagnostic struct {
	Name      string  `json:"name"`
	Observed  float64 `json:"observed"`
	Generated float64 `json:"generated"`
	Distance  float64 `json:"distance"`
}

type GrammarDiagnostic struct {
	TokenCountExact         bool                   `json:"token_count_exact"`
	LengthDistributionExact bool                   `json:"length_distribution_exact"`
	AlphabetExact           bool                   `json:"alphabet_exact"`
	PositionalGlyphTV       float64                `json:"positional_glyph_tv"`
	InitialGlyphTV          float64                `json:"initial_glyph_tv"`
	FinalGlyphTV            float64                `json:"final_glyph_tv"`
	BigramTV                float64                `json:"bigram_tv"`
	VocabularySize          DistributionDiagnostic `json:"vocabulary_size"`
	SingletonShare          DistributionDiagnostic `json:"singleton_share"`
	RareShare               DistributionDiagnostic `json:"rare_share"`
	TokenFrequencyTV        float64                `json:"token_frequency_distribution_tv"`
}

type GrammarRun struct {
	Mode       string            `json:"mode"`
	Replicate  int               `json:"replicate"`
	Seed       int64             `json:"seed"`
	Diagnostic GrammarDiagnostic `json:"diagnostic"`
	LP1Gini    float64           `json:"lp1_support_gini"`
	PrefixNMI  float64           `json:"lp4_prefix_nmi"`
	SuffixNMI  float64           `json:"lp4_suffix_nmi"`
	EF1        EF1Result         `json:"ef1"`
	EF2        EF2Result         `json:"ef2"`
	EF3        EF3Result         `json:"ef3"`
}

type GrammarSummary struct {
	Runs       []GrammarRun `json:"runs"`
	Validation string       `json:"validation"`
	Reason     string       `json:"reason"`
}

type CorpusResult struct {
	Corpus  CorpusInfo      `json:"corpus"`
	Metrics Metrics         `json:"metrics"`
	Grammar *GrammarSummary `json:"grammar,omitempty"`
}

type Verdict struct {
	ID          string `json:"id"`
	Value       string `json:"value"`
	Basis       string `json:"basis"`
	Limitations string `json:"limitations"`
}

type Fingerprint struct {
	Version    string         `json:"version"`
	Provenance Provenance     `json:"provenance"`
	Primary    CorpusResult   `json:"primary"`
	Controls   []CorpusResult `json:"controls"`
	Verdicts   []Verdict      `json:"verdicts"`
	Warnings   []string       `json:"warnings"`
	Errors     []string       `json:"errors"`
}

func stableTests(tests []NullTest) {
	sort.Slice(tests, func(i, j int) bool { return tests[i].ID < tests[j].ID })
}
