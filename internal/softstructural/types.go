package softstructural

import "zcore.dev/voinich/internal/profilestability"

type Config struct {
	DictionaryPath      string
	AnalysisPath        string
	ReliabilityPath     string
	OutputPath          string
	PairsPath           string
	MinTokenCount       int
	Neighbors           int
	MinEvidenceStrength float64
	GraphMinSimilarity  float64
}

type Pair struct {
	TokenA, TokenB                                                           string
	CountA, CountB                                                           int
	PositionSimilarity, LeftSimilarity, RightSimilarity, RawSimilarity       float64
	PositionReliability, LeftReliability, RightReliability, EvidenceStrength float64
	TotalEvidenceWeight                                                      float64
	DiagnosticWeightedSimilarity                                             *float64
	BootstrapProbabilityAbove070                                             *float64
}

type Neighbor struct {
	Token                        string   `yaml:"token"`
	RawSimilarity                float64  `yaml:"raw_similarity"`
	EvidenceStrength             float64  `yaml:"evidence_strength"`
	DiagnosticWeightedSimilarity *float64 `yaml:"diagnostic_weighted_similarity"`
	BootstrapProbabilityAbove070 *float64 `yaml:"bootstrap_probability_above_070,omitempty"`
}

type TokenNeighborhood struct {
	Token                    string     `yaml:"token"`
	Count                    int        `yaml:"count"`
	TopRawNeighbors          []Neighbor `yaml:"top_raw_neighbors"`
	TopSupportedNeighbors    []Neighbor `yaml:"top_supported_neighbors"`
	TopHighEvidenceNeighbors []Neighbor `yaml:"top_high_evidence_raw_neighbors"`
}

type Distribution struct {
	Mean   float64 `yaml:"mean"`
	Median float64 `yaml:"median"`
	P10    float64 `yaml:"p10,omitempty"`
	P25    float64 `yaml:"p25,omitempty"`
	P50    float64 `yaml:"p50,omitempty"`
	P75    float64 `yaml:"p75,omitempty"`
	P90    float64 `yaml:"p90,omitempty"`
	P95    float64 `yaml:"p95,omitempty"`
	P99    float64 `yaml:"p99,omitempty"`
	Max    float64 `yaml:"max,omitempty"`
}

type JointDiagnostics struct {
	RawGE070              int `yaml:"pairs_raw_similarity_ge_070"`
	RawGE070EvidenceGE050 int `yaml:"pairs_raw_similarity_ge_070_and_evidence_ge_050"`
	RawGE070EvidenceGE070 int `yaml:"pairs_raw_similarity_ge_070_and_evidence_ge_070"`
	RawGE070EvidenceGE090 int `yaml:"pairs_raw_similarity_ge_070_and_evidence_ge_090"`
}

type BucketRow struct {
	RawSimilarityBin string `yaml:"raw_similarity_bin"`
	EvidenceLT030    int    `yaml:"evidence_lt_0_3"`
	Evidence030To050 int    `yaml:"evidence_0_3_to_0_5"`
	Evidence050To070 int    `yaml:"evidence_0_5_to_0_7"`
	Evidence070To090 int    `yaml:"evidence_0_7_to_0_9"`
	EvidenceGE090    int    `yaml:"evidence_ge_0_9"`
}

type GraphEdge struct {
	TokenA           string  `yaml:"token_a"`
	TokenB           string  `yaml:"token_b"`
	RawSimilarity    float64 `yaml:"raw_similarity"`
	EvidenceStrength float64 `yaml:"evidence_strength"`
}
type MutualPair struct {
	TokenA                       string   `yaml:"token_a"`
	TokenB                       string   `yaml:"token_b"`
	RawSimilarity                float64  `yaml:"raw_similarity"`
	EvidenceStrength             float64  `yaml:"evidence_strength"`
	DiagnosticWeightedSimilarity *float64 `yaml:"diagnostic_weighted_similarity"`
}

type ComponentEvidence struct {
	Similarity  float64 `yaml:"similarity"`
	Reliability float64 `yaml:"reliability"`
}
type ReferencePair struct {
	TokenA                       string            `yaml:"token_a"`
	TokenB                       string            `yaml:"token_b"`
	CountA                       int               `yaml:"count_a"`
	CountB                       int               `yaml:"count_b"`
	Position                     ComponentEvidence `yaml:"position"`
	LeftContext                  ComponentEvidence `yaml:"left_context"`
	RightContext                 ComponentEvidence `yaml:"right_context"`
	RawStructuralSimilarity      float64           `yaml:"raw_structural_similarity"`
	TotalEvidenceWeight          float64           `yaml:"total_evidence_weight"`
	EvidenceStrength             float64           `yaml:"evidence_strength"`
	DiagnosticWeightedSimilarity *float64          `yaml:"diagnostic_weighted_similarity"`
	BootstrapProbabilityAbove070 *float64          `yaml:"bootstrap_probability_above_070,omitempty"`
}

type Output struct {
	Parameters struct {
		MinTokenCount       int     `yaml:"min_token_count"`
		Neighbors           int     `yaml:"neighbors"`
		MinEvidenceStrength float64 `yaml:"min_evidence_strength"`
		GraphMinSimilarity  float64 `yaml:"graph_min_similarity"`
		PairsFile           string  `yaml:"pairs_file"`
	} `yaml:"parameters"`
	Methodology struct {
		Similarity       string `yaml:"similarity"`
		Reliability      string `yaml:"reliability"`
		EvidenceStrength string `yaml:"evidence_strength"`
		Diagnostic       string `yaml:"diagnostic_weighted_similarity"`
		Graph            string `yaml:"graph"`
	} `yaml:"methodology"`
	EligibleTokens               int                 `yaml:"eligible_tokens"`
	PairCount                    int                 `yaml:"pair_count"`
	RawSimilarityDistribution    Distribution        `yaml:"raw_similarity_distribution"`
	EvidenceStrengthDistribution Distribution        `yaml:"evidence_strength_distribution"`
	JointDiagnostics             JointDiagnostics    `yaml:"joint_diagnostics"`
	DiagnosticBuckets            []BucketRow         `yaml:"diagnostic_buckets_2d"`
	Neighborhoods                []TokenNeighborhood `yaml:"neighborhoods"`
	GraphEdges                   []GraphEdge         `yaml:"soft_neighborhood_graph_edges"`
	MutualRaw                    []MutualPair        `yaml:"mutual_nearest_neighbors_raw"`
	MutualSupported              []MutualPair        `yaml:"mutual_nearest_neighbors_supported"`
	ReferencePairs               []ReferencePair     `yaml:"reference_pairs"`
}

type dataset struct {
	profiles map[string]profilestability.Profile
	counts   map[string]int
}
