package pairdecomposition

type Config struct {
	DictionaryPath, PairsPath, DistantPath, FamiliesPath, OutputDir string
	TopN, ContextLimit, Controls, FamilyID                          int
	Pair                                                            string
	// SkipSVG omits derived presentation plots. Scientific YAML/TSV outputs are
	// always written; keeping this separate makes large runs bounded by the
	// scientific result rather than by thousands of small files.
	SkipSVG bool
}

type PairSource struct {
	TokenA, TokenB                                         string
	CountA, CountB                                         int
	Structural, Reliability, Graphemic                     float64
	Position, Left, Right                                  float64
	PositionReliability, LeftReliability, RightReliability float64
}

type PositionRow struct {
	Position int     `yaml:"position"`
	A        float64 `yaml:"probability_a"`
	B        float64 `yaml:"probability_b"`
}

type PositionSummary struct {
	LineStartProbability float64 `yaml:"line_start_probability"`
	LineEndProbability   float64 `yaml:"line_end_probability"`
	Mean                 float64 `yaml:"mean_position"`
	Median               float64 `yaml:"median_position"`
}

type ContextRow struct {
	Token        string  `yaml:"token"`
	ProbabilityA float64 `yaml:"probability_a"`
	ProbabilityB float64 `yaml:"probability_b"`
	Difference   float64 `yaml:"difference_a_minus_b"`
	AssociationA float64 `yaml:"lift_a,omitempty"`
	AssociationB float64 `yaml:"lift_b,omitempty"`
}

type ProbabilityRow struct {
	Token       string  `yaml:"token"`
	Count       int     `yaml:"count"`
	Probability float64 `yaml:"probability"`
}

type ContextProfile struct {
	ObservationsA            int              `yaml:"observations_a"`
	ObservationsB            int              `yaml:"observations_b"`
	EntropyA                 float64          `yaml:"entropy_a_nats"`
	EntropyB                 float64          `yaml:"entropy_b_nats"`
	EffectiveVocabularyA     float64          `yaml:"effective_vocabulary_a"`
	EffectiveVocabularyB     float64          `yaml:"effective_vocabulary_b"`
	EntropyDifference        float64          `yaml:"entropy_difference_a_minus_b"`
	Jaccard                  float64          `yaml:"allowed_token_jaccard"`
	JensenShannonSimilarity  float64          `yaml:"jensen_shannon_similarity"`
	ExistingCosineSimilarity float64          `yaml:"existing_cosine_similarity"`
	DistributionA            []ProbabilityRow `yaml:"full_distribution_a"`
	DistributionB            []ProbabilityRow `yaml:"full_distribution_b"`
	Common                   []ContextRow     `yaml:"common,omitempty"`
	AssociatedBoth           []ContextRow     `yaml:"strongly_associated_with_both,omitempty"`
	SpecificA                []ContextRow     `yaml:"specific_a,omitempty"`
	SpecificB                []ContextRow     `yaml:"specific_b,omitempty"`
	Differential             []ContextRow     `yaml:"largest_differences,omitempty"`
	SharedRare               []ContextRow     `yaml:"shared_rare,omitempty"`
	SharedAbsent             []string         `yaml:"shared_absent_high_frequency_contexts,omitempty"`
}

type PairResult struct {
	TokenA                      string          `yaml:"token_a"`
	TokenB                      string          `yaml:"token_b"`
	CountA                      int             `yaml:"count_a"`
	CountB                      int             `yaml:"count_b"`
	StructuralSimilarity        float64         `yaml:"structural_similarity"`
	Reliability                 float64         `yaml:"reliability"`
	GraphemicDistance           float64         `yaml:"normalized_graphemic_distance"`
	PositionSimilarity          float64         `yaml:"position_similarity"`
	LeftSimilarity              float64         `yaml:"left_context_similarity"`
	RightSimilarity             float64         `yaml:"right_context_similarity"`
	PositionReliability         float64         `yaml:"position_reliability"`
	LeftReliability             float64         `yaml:"left_reliability"`
	RightReliability            float64         `yaml:"right_reliability"`
	PositionA                   PositionSummary `yaml:"position_summary_a"`
	PositionB                   PositionSummary `yaml:"position_summary_b"`
	PositionDistribution        []PositionRow   `yaml:"position_distribution"`
	PositionJSSimilarity        float64         `yaml:"position_js_similarity"`
	Left                        ContextProfile  `yaml:"left_context"`
	Right                       ContextProfile  `yaml:"right_context"`
	SharedContextStrength       float64         `yaml:"shared_context_strength"`
	DifferentialContextStrength float64         `yaml:"differential_context_strength"`
	PositionalAgreement         float64         `yaml:"positional_agreement"`
	EntropyAgreement            float64         `yaml:"entropy_agreement"`
	Explanation                 []string        `yaml:"formal_explanation"`
}

type Control struct {
	TargetA       string     `yaml:"target_a"`
	TargetB       string     `yaml:"target_b"`
	Rank          int        `yaml:"rank"`
	MatchCost     float64    `yaml:"match_cost"`
	Decomposition PairResult `yaml:"decomposition"`
}

type Matrix struct {
	Tokens []string    `yaml:"tokens"`
	Values [][]float64 `yaml:"values"`
}

type FamilyEdge struct {
	TokenA               string  `yaml:"token_a"`
	TokenB               string  `yaml:"token_b"`
	StructuralSimilarity float64 `yaml:"structural_similarity"`
	Reliability          float64 `yaml:"reliability"`
	GraphemicDistance    float64 `yaml:"normalized_graphemic_distance"`
}
type FamilyInput struct {
	ID     int      `yaml:"id"`
	Tokens []string `yaml:"tokens"`
	Edges  []struct {
		TokenA               string  `yaml:"token_a"`
		TokenB               string  `yaml:"token_b"`
		StructuralSimilarity float64 `yaml:"structural_similarity"`
		Reliability          float64 `yaml:"reliability"`
		GraphemicDistance    float64 `yaml:"normalized_grapheme_distance"`
	} `yaml:"edges"`
}
type FamilyResult struct {
	ID           int                `yaml:"id"`
	Tokens       []string           `yaml:"tokens"`
	Edges        []FamilyEdge       `yaml:"edges"`
	Structural   Matrix             `yaml:"structural_similarity_matrix"`
	Reliability  Matrix             `yaml:"reliability_matrix"`
	Graphemic    Matrix             `yaml:"normalized_graphemic_distance_matrix"`
	Position     Matrix             `yaml:"position_similarity_matrix"`
	Left         Matrix             `yaml:"left_context_similarity_matrix"`
	Right        Matrix             `yaml:"right_context_similarity_matrix"`
	Medoid       string             `yaml:"structural_medoid"`
	MeanDistance map[string]float64 `yaml:"mean_structural_distance"`
	Peripheral   []string           `yaml:"peripheral_tokens"`
}

type Output struct {
	Methodology map[string]string `yaml:"methodology"`
	Pairs       []PairResult      `yaml:"pairs"`
	Controls    []Control         `yaml:"negative_controls"`
}
