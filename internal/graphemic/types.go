package graphemic

type Config struct {
	InputPath               string
	OutputDir               string
	MinStructuralSimilarity float64
	MinReliability          float64
	MinGraphemicDistance    float64
	MinCloseSimilarity      float64
	TopN                    int
}

type Pair struct {
	TokenA, TokenB                                         string
	CountA, CountB                                         int
	PositionSimilarity, LeftSimilarity, RightSimilarity    float64
	StructuralSimilarity, Reliability                      float64
	PositionReliability, LeftReliability, RightReliability float64
	TotalEvidenceWeight                                    float64
	DiagnosticWeightedSimilarity                           string
	GraphemeDistance                                       int
	NormalizedGraphemeDistance, GraphemeSimilarity         float64
	CommonPrefix, CommonSuffix, LengthDifference           int
	DiscoveryScore                                         float64
}

type Bin struct {
	Range                  string
	PairCount              int
	Mean, Median, P90, P95 float64
}

type FrequencyResult struct {
	MinimumCount      int
	PairCount         int
	Pearson, Spearman float64
	DistantCandidates int
}

type FamilyEdge struct {
	TokenA                     string  `yaml:"token_a"`
	TokenB                     string  `yaml:"token_b"`
	StructuralSimilarity       float64 `yaml:"structural_similarity"`
	Reliability                float64 `yaml:"reliability"`
	NormalizedGraphemeDistance float64 `yaml:"normalized_grapheme_distance"`
}

type Family struct {
	ID     int          `yaml:"id"`
	Tokens []string     `yaml:"tokens"`
	Edges  []FamilyEdge `yaml:"edges"`
}

type FamiliesOutput struct {
	Term     string             `yaml:"term"`
	Criteria map[string]float64 `yaml:"criteria"`
	Families []Family           `yaml:"families"`
}

type Result struct {
	Pairs                          []Pair
	TokenCount                     int
	Pearson, Spearman              float64
	Bins                           []Bin
	Frequency                      []FrequencyResult
	Distant, Close                 []Pair
	PercentileDistant              []Pair
	CloseFamilies, DistantFamilies []Family
	DistantPercentileCutoffs       map[string]float64
}
