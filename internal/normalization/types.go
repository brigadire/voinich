package normalization

type Candidate struct {
	TokenA                 string  `yaml:"token_a"`
	TokenB                 string  `yaml:"token_b"`
	CountA                 int     `yaml:"count_a"`
	CountB                 int     `yaml:"count_b"`
	Similarity             float64 `yaml:"similarity"`
	PositionSimilarity     float64 `yaml:"position_similarity"`
	LeftContextSimilarity  float64 `yaml:"left_context_similarity"`
	RightContextSimilarity float64 `yaml:"right_context_similarity"`
}

type StructuralInput struct {
	Parameters struct {
		MinTokenCountForRanking int `yaml:"min_token_count_for_ranking"`
	} `yaml:"parameters"`
	EquivalenceCandidates []Candidate `yaml:"equivalence_candidates"`
}

type PairMetrics struct {
	Similarity             float64
	PositionSimilarity     float64
	LeftContextSimilarity  float64
	RightContextSimilarity float64
}

type Config struct {
	Thresholds                []float64
	MinPositionSimilarity     float64
	MinLeftContextSimilarity  float64
	MinRightContextSimilarity float64
	MinTokenCount             int
	SingletonMode             string
	RandomBaselines           int
	RandomSeed                int64
}

type Corpus struct {
	Lines       [][]string
	Counts      map[string]int
	Occurrences int
	NonEmpty    int
	Transitions int
}

type Member struct {
	Token string `yaml:"token"`
	Count int    `yaml:"count"`
}

type Class struct {
	ID                        string   `yaml:"id"`
	Members                   []Member `yaml:"members"`
	Size                      int      `yaml:"size"`
	MinSimilarity             float64  `yaml:"min_similarity"`
	MeanSimilarity            float64  `yaml:"mean_similarity"`
	MinPositionSimilarity     float64  `yaml:"min_position_similarity"`
	MinLeftContextSimilarity  float64  `yaml:"min_left_context_similarity"`
	MinRightContextSimilarity float64  `yaml:"min_right_context_similarity"`
}

type ModelStats struct {
	RawUniqueTokens         int     `yaml:"raw_unique_tokens"`
	NormalizedUniqueSymbols int     `yaml:"normalized_unique_symbols"`
	ClassifiedTokens        int     `yaml:"classified_tokens"`
	SingletonTokens         int     `yaml:"singleton_tokens"`
	Classes                 int     `yaml:"classes"`
	MultiMemberClasses      int     `yaml:"multi_member_classes"`
	TokensInMultiClasses    int     `yaml:"tokens_in_multi_member_classes"`
	LargestClass            int     `yaml:"largest_class"`
	TokenOccurrenceCoverage float64 `yaml:"token_occurrence_coverage"`
	CompressionRatio        float64 `yaml:"compression_ratio"`
}

type Model struct {
	Threshold float64    `yaml:"threshold"`
	Label     string     `yaml:"label"`
	Stats     ModelStats `yaml:"stats"`
	Classes   []Class    `yaml:"classes"`
}

type ClassesMeta struct {
	InputCorpus               string  `yaml:"input_corpus"`
	StructuralAnalysis        string  `yaml:"structural_analysis"`
	SingletonMode             string  `yaml:"singleton_mode"`
	MinTokenCount             int     `yaml:"min_token_count"`
	MinPositionSimilarity     float64 `yaml:"min_position_similarity"`
	MinLeftContextSimilarity  float64 `yaml:"min_left_context_similarity"`
	MinRightContextSimilarity float64 `yaml:"min_right_context_similarity"`
	RandomBaselines           int     `yaml:"random_baselines"`
	RandomSeed                int64   `yaml:"random_seed"`
	Clustering                string  `yaml:"clustering"`
	RandomMatching            string  `yaml:"random_matching"`
}

type ClassesOutput struct {
	Meta       ClassesMeta `yaml:"meta"`
	Thresholds []float64   `yaml:"thresholds"`
	Models     []Model     `yaml:"models"`
}
