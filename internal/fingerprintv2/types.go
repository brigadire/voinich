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

	// CrossScale is task77's additional configuration for the edit-graph
	// validation and cross-scale block. A caller who omits this section
	// gets task77's declared defaults, not a skipped block, so the new
	// artifacts are always present for any corpus with >=2 vocabulary
	// types; task75's own LP/EF fields and existing configs are otherwise
	// untouched.
	CrossScale *CrossScaleConfig `yaml:"cross_scale,omitempty" json:"cross_scale,omitempty"`
	Task79     *Task79Config     `yaml:"task79,omitempty" json:"task79,omitempty"`
}

// Task79Config enables the page/hierarchy/freeze-readiness extension.  It is
// opt-in so task75/task77 configurations retain byte-for-byte semantics.
type Task79Config struct {
	Enabled             bool     `yaml:"enabled" json:"enabled"`
	Permutations        int      `yaml:"permutations" json:"permutations"`
	BootstrapReplicates int      `yaml:"bootstrap_replicates" json:"bootstrap_replicates"`
	MinGroupSize        int      `yaml:"min_group_size" json:"min_group_size"`
	ChangePointPenalty  float64  `yaml:"change_point_penalty" json:"change_point_penalty"`
	AuditArtifacts      []string `yaml:"audit_artifacts,omitempty" json:"audit_artifacts,omitempty"`
	CorrectionLayer     string   `yaml:"correction_layer,omitempty" json:"correction_layer,omitempty"`
}

func (c Task79Config) normalized() Task79Config {
	if c.Permutations <= 0 {
		c.Permutations = 200
	}
	if c.BootstrapReplicates <= 0 {
		c.BootstrapReplicates = 200
	}
	if c.MinGroupSize <= 0 {
		c.MinGroupSize = 5
	}
	if c.ChangePointPenalty <= 0 {
		c.ChangePointPenalty = 2.0
	}
	return c
}

type CrossScaleConfig struct {
	Folds            int `yaml:"folds" json:"folds"`
	HubTopPercent    int `yaml:"hub_top_percent" json:"hub_top_percent"`
	MinFamilySize    int `yaml:"min_family_size" json:"min_family_size"`
	StructuralSample int `yaml:"structural_sample" json:"structural_sample"`
}

func (c CrossScaleConfig) normalized() CrossScaleConfig {
	if c.Folds <= 0 {
		c.Folds = 5
	}
	if c.HubTopPercent <= 0 {
		c.HubTopPercent = 5
	}
	if c.MinFamilySize <= 0 {
		c.MinFamilySize = 3
	}
	if c.StructuralSample <= 0 {
		c.StructuralSample = 2000
	}
	return c
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
	if c.CrossScale == nil {
		c.CrossScale = &CrossScaleConfig{}
	}
	normalizedCrossScale := c.CrossScale.normalized()
	c.CrossScale = &normalizedCrossScale
	if c.Task79 != nil {
		n := c.Task79.normalized()
		c.Task79 = &n
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
	Available      bool      `json:"available"`
	SameLineRate   float64   `json:"same_line_rate"`
	SamePageRate   *float64  `json:"same_page_rate,omitempty"`
	SameRegimeRate *float64  `json:"same_regime_rate,omitempty"`
	GlobalNull     *NullTest `json:"global_null,omitempty"`
	RegimeNull     *NullTest `json:"regime_null,omitempty"`
	FamilyCount    int       `json:"family_count"`
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
	// EF5 is the same locality computation as LP3.Locality (task73 §5:
	// "identical to LP3's locality computation and therefore implemented
	// once, not twice"), extended here with the C-REGIME comparison.
	// It is copied by value, not recomputed.
	EF5 LocalityResult `json:"ef5"`
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
	Corpus     CorpusInfo            `json:"corpus"`
	Metrics    Metrics               `json:"metrics"`
	Grammar    *GrammarSummary       `json:"grammar,omitempty"`
	EditGraph  *EditFamilyValidation `json:"edit_graph_validation,omitempty"`
	CrossScale *CrossScaleResult     `json:"cross_scale,omitempty"`
	Task79     *Task79Result         `json:"task79,omitempty"`
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

// ---- Task77: edit-graph validation (stage 2) ----

type GraphRepresentation struct {
	Kind        string `json:"kind"`
	Status      string `json:"status"` // IMPLEMENTED or DEFERRED
	Description string `json:"description"`
	Reason      string `json:"reason,omitempty"`
}

type FamilyStructuralDiagnostic struct {
	FamilyIndex              int     `json:"family_index"`
	Size                     int     `json:"size"`
	Diameter                 int     `json:"diameter"`
	AverageShortestPath      float64 `json:"average_shortest_path"`
	IndirectPairShare        float64 `json:"indirect_pair_share"`
	MeanInternalEditDistance float64 `json:"mean_internal_edit_distance"`
	ArticulationPoints       int     `json:"articulation_points"`
	BridgeEdges              int     `json:"bridge_edges"`
	CoreSize                 int     `json:"core_size"`
	PeripherySize            int     `json:"periphery_size"`
}

type HubDependence struct {
	HubFraction      float64 `json:"hub_fraction"`
	RemovedNodes     int     `json:"removed_nodes"`
	GiantShareBefore float64 `json:"giant_share_before"`
	GiantShareAfter  float64 `json:"giant_share_after"`
	GiantShareDrop   float64 `json:"giant_share_drop"`
}

type PathRestriction struct {
	MaxHops        int     `json:"max_hops"`
	ComponentCount int     `json:"component_count"`
	GiantShare     float64 `json:"giant_share"`
}

type CommunityComparison struct {
	Method string  `json:"method"`
	ARI    float64 `json:"adjusted_rand_index"`
	NMI    float64 `json:"normalized_mutual_information"`
	VI     float64 `json:"variation_of_information"`
	Seed   int64   `json:"seed"`
}

type TransitiveMergeAudit struct {
	Families                  []FamilyStructuralDiagnostic `json:"families"`
	HubDependence             HubDependence                `json:"hub_dependence"`
	PathRestrictions          []PathRestriction            `json:"path_restrictions"`
	CommunityVsComponents     CommunityComparison          `json:"community_vs_components"`
	FrequencyWeightedHubShare float64                      `json:"frequency_weighted_hub_share"`
	ContextWeightedHubShare   float64                      `json:"context_weighted_hub_share"`
}

type StabilityRun struct {
	Perturbation    string  `json:"perturbation"`
	Value           string  `json:"value"`
	ARI             float64 `json:"adjusted_rand_index"`
	NMI             float64 `json:"normalized_mutual_information"`
	ComparableNodes int     `json:"comparable_nodes"`
	Status          string  `json:"status"` // GLOBAL/PARTITION_SPECIFIC/UNSTABLE/INSUFFICIENT_DATA/NOT_TESTABLE
	Note            string  `json:"note,omitempty"`
}

type FamilyMember struct {
	Token      string  `json:"token"`
	Role       string  `json:"role"` // CORE or PERIPHERY
	Confidence float64 `json:"confidence"`
}

type ConsensusFamily struct {
	Index               int            `json:"index"`
	Members             []FamilyMember `json:"members"`
	TransformationTypes []string       `json:"transformation_types"`
	CorpusCoverage      float64        `json:"corpus_coverage"`
	OccurrenceCoverage  float64        `json:"occurrence_coverage"`
	DominantHub         string         `json:"dominant_hub,omitempty"`
	StabilityScore      float64        `json:"stability_score"`
}

type EditFamilyValidation struct {
	GraphRepresentations []GraphRepresentation `json:"graph_representations"`
	TransitiveMerge      TransitiveMergeAudit  `json:"transitive_merge"`
	StabilityRuns        []StabilityRun        `json:"stability_runs"`
	ConsensusStatus      string                `json:"consensus_status"`
	ConsensusFamilies    []ConsensusFamily     `json:"consensus_families,omitempty"`
}

// ---- Task77: cross-scale block (stages 4-10) ----

type CrossScaleVariable struct {
	Scale         string `json:"scale"`
	Name          string `json:"name"`
	Origin        string `json:"origin"`
	Domain        string `json:"domain"`
	MissingPolicy string `json:"missing_data_policy"`
	Available     bool   `json:"available"`
}

type HeldOutResult struct {
	Scheme          string  `json:"scheme"`
	Folds           int     `json:"folds"`
	BaselineLogLoss float64 `json:"baseline_log_loss"`
	ModelLogLoss    float64 `json:"model_log_loss"`
	Improvement     float64 `json:"improvement"`
	ImprovementSD   float64 `json:"improvement_sd"`
	N               int     `json:"n"`
	Note            string  `json:"note,omitempty"`
}

type CrossScaleMetric struct {
	MetricID                  string         `json:"metric_id"`
	MetricVersion             string         `json:"metric_version"`
	Status                    string         `json:"status"`
	Hypothesis                string         `json:"hypothesis"`
	UnitOfAnalysis            string         `json:"unit_of_analysis"`
	Variables                 []string       `json:"variables"`
	ConditioningVariables     []string       `json:"conditioning_variables,omitempty"`
	Confounders               []string       `json:"confounders,omitempty"`
	ObservedStatistic         float64        `json:"observed_statistic"`
	EffectSize                float64        `json:"effect_size"`
	EffectDefined             bool           `json:"effect_defined"`
	Uncertainty               string         `json:"uncertainty"`
	NullModel                 string         `json:"null_model"`
	NullMean                  float64        `json:"null_mean"`
	NullSD                    float64        `json:"null_sd"`
	EmpiricalP                float64        `json:"empirical_p"`
	MultipleTestingAdjustment float64        `json:"multiple_testing_adjustment"`
	PartitionStability        []StabilityRun `json:"partition_stability,omitempty"`
	HeldOutPerformance        *HeldOutResult `json:"held_out_performance,omitempty"`
	Sensitivity               string         `json:"sensitivity"`
	RedundancyClass           string         `json:"redundancy_class"`
	Interpretation            string         `json:"interpretation"`
	Limitations               string         `json:"limitations"`
	N                         int            `json:"n"`
	AdditionalNulls           []NullTest     `json:"additional_nulls,omitempty"`
	AnalysisType              string         `json:"analysis_type"` // CONFIRMATORY or EXPLORATORY
}

type NullModelRegistryEntry struct {
	ID                   string   `json:"id"`
	Name                 string   `json:"name"`
	Preserves            string   `json:"preserves"`
	Destroys             string   `json:"destroys"`
	TestsHypotheses      []string `json:"tests_hypotheses"`
	RemainingConfounders string   `json:"remaining_confounders"`
	Justification        string   `json:"justification"`
}

type RedundancyRow struct {
	MetricA     string  `json:"metric_a"`
	MetricB     string  `json:"metric_b"`
	Correlation float64 `json:"correlation"`
	N           int     `json:"n"`
}

type MetricClassification struct {
	MetricID string `json:"metric_id"`
	Class    string `json:"class"` // CORE/SUPPORTING/DIAGNOSTIC/REDUNDANT/UNSTABLE/DEFERRED
	Reason   string `json:"reason"`
}

type CrossScaleResult struct {
	VariablesAvailable    []CrossScaleVariable     `json:"variables_available"`
	Metrics               []CrossScaleMetric       `json:"metrics"`
	NullRegistry          []NullModelRegistryEntry `json:"null_registry"`
	RedundancyMatrix      []RedundancyRow          `json:"redundancy_matrix"`
	MetricClassifications []MetricClassification   `json:"metric_classifications"`
	ConfirmatoryFindings  []string                 `json:"confirmatory_findings"`
	ExploratoryFindings   []string                 `json:"exploratory_findings"`
	Verdicts              []CrossScaleVerdict      `json:"verdicts"`
}

// CrossScaleVerdict is task77's final-verdict record shape: every field
// task77 requires ("primary statistic; effect size; null comparison;
// held-out result; partition stability; sensitivity; ограничения") is
// present, even when its value must be a documented "not applicable"
// rather than a number.
type CrossScaleVerdict struct {
	ID                 string  `json:"id"`
	Value              string  `json:"value"`
	PrimaryStatistic   string  `json:"primary_statistic"`
	EffectSize         float64 `json:"effect_size"`
	EffectDefined      bool    `json:"effect_defined"`
	NullComparison     string  `json:"null_comparison"`
	HeldOutResult      string  `json:"held_out_result"`
	PartitionStability string  `json:"partition_stability"`
	Sensitivity        string  `json:"sensitivity"`
	Limitations        string  `json:"limitations"`
}
