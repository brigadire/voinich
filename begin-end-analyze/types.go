package main

type PositionInput struct {
	Position int `yaml:"position"`
	Count    int `yaml:"count"`
}

type NeighborInput struct {
	Token string `yaml:"token"`
	Count int    `yaml:"count"`
}

type DictionaryToken struct {
	Token            string          `yaml:"token"`
	Count            int             `yaml:"count"`
	PositionInString []PositionInput `yaml:"position_in_string"`
	WordBefore       []NeighborInput `yaml:"word_before"`
	WordAfter        []NeighborInput `yaml:"word_after"`
	LineStartCount   int             `yaml:"line_start_count"`
	LineEndCount     int             `yaml:"line_end_count"`
}

type Parameters struct {
	MaxWindow       int    `yaml:"max_window"`
	Windows         []int  `yaml:"windows"`
	Permutations    int    `yaml:"permutations"`
	MinTokenCount   int    `yaml:"min_token_count"`
	RandomSeed      int64  `yaml:"random_seed"`
	PermutationMode string `yaml:"permutation_mode"`
	IncludeUnclear  bool   `yaml:"include_unclear_tokens"`
	MaxCandidates   int    `yaml:"max_candidates"`
}

type Meta struct {
	TokenOccurrences    int  `yaml:"token_occurrences"`
	UniqueTokens        int  `yaml:"unique_tokens"`
	EligibleTokens      int  `yaml:"eligible_tokens"`
	Lines               int  `yaml:"lines"`
	Pages               int  `yaml:"pages"`
	ExplicitPageBreaks  int  `yaml:"explicit_page_breaks"`
	PageBoundariesKnown bool `yaml:"page_boundaries_known"`
	CandidatePairs      int  `yaml:"candidate_pairs"`
	UnclearExcluded     int  `yaml:"unclear_tokens_excluded"`
}

type Methodology struct {
	Tokenization       string `yaml:"tokenization"`
	Pages              string `yaml:"pages"`
	DirectedDependency string `yaml:"directed_dependency"`
	Permutation        string `yaml:"permutation"`
	Nesting            string `yaml:"nesting"`
	Score              string `yaml:"score"`
	Interpretation     string `yaml:"interpretation"`
}

type CountsResult struct {
	Begin int `yaml:"begin"`
	End   int `yaml:"end"`
}

type PositionSide struct {
	StartProbability       float64 `yaml:"start_probability"`
	EndProbability         float64 `yaml:"end_probability"`
	MeanPosition           float64 `yaml:"mean_position"`
	MeanNormalizedPosition float64 `yaml:"mean_normalized_position"`
}

type PositionResult struct {
	Begin           PositionSide `yaml:"begin"`
	End             PositionSide `yaml:"end"`
	Complementarity float64      `yaml:"complementarity"`
}

type WindowResult struct {
	Window       string  `yaml:"window"`
	Observations int     `yaml:"observations"`
	Probability  float64 `yaml:"probability"`
}

type DistanceResult struct {
	Scope                string         `yaml:"scope"`
	BeginOccurrences     int            `yaml:"begin_occurrences"`
	Observations         int            `yaml:"observations"`
	Probability          float64        `yaml:"probability"`
	Mean                 float64        `yaml:"mean"`
	Median               float64        `yaml:"median"`
	Histogram            map[int]int    `yaml:"histogram"`
	WithoutEnd           int            `yaml:"without_end"`
	EndWithoutPriorBegin int            `yaml:"end_without_prior_begin"`
	Windows              []WindowResult `yaml:"windows"`
}

type DirectionalityResult struct {
	Scope    string  `yaml:"scope"`
	AToB     float64 `yaml:"a_to_b"`
	BToA     float64 `yaml:"b_to_a"`
	Score    float64 `yaml:"score"`
	LogRatio float64 `yaml:"log_ratio"`
}

type PageBalanceResult struct {
	MeanDifference         float64 `yaml:"mean_difference"`
	StddevDifference       float64 `yaml:"stddev_difference"`
	MeanAbsoluteDifference float64 `yaml:"mean_absolute_difference"`
	NearZeroFraction       float64 `yaml:"near_zero_fraction"`
	ComparableStddev       float64 `yaml:"comparable_pair_stddev"`
	StddevRatio            float64 `yaml:"stddev_ratio"`
	RelativeScore          float64 `yaml:"relative_score"`
}

type NestingResult struct {
	AABB int `yaml:"AABB"`
	ABAB int `yaml:"ABAB"`
	ABBA int `yaml:"ABBA"`
	BAAB int `yaml:"BAAB"`
}

type SignificanceResult struct {
	PermutationP   float64 `yaml:"permutation_p"`
	ZScore         float64 `yaml:"z_score"`
	ExpectedMean   float64 `yaml:"expected_mean"`
	ExpectedStddev float64 `yaml:"expected_stddev"`
	Permutations   int     `yaml:"permutations"`
}

type LocalCompatibility struct {
	AdjacentCount int     `yaml:"adjacent_count"`
	AdjacentShare float64 `yaml:"adjacent_share"`
	LikelyLocal   bool    `yaml:"likely_local"`
}

type Candidate struct {
	BeginCandidate     string               `yaml:"begin_candidate"`
	EndCandidate       string               `yaml:"end_candidate"`
	ContainsUnclear    bool                 `yaml:"contains_unclear"`
	Counts             CountsResult         `yaml:"counts"`
	Position           PositionResult       `yaml:"position"`
	WithinLine         DistanceResult       `yaml:"within_line"`
	WithinPage         DistanceResult       `yaml:"within_page"`
	Directionality     DirectionalityResult `yaml:"directionality"`
	PageBalance        PageBalanceResult    `yaml:"page_balance"`
	Nesting            NestingResult        `yaml:"nesting"`
	SignificanceLine   SignificanceResult   `yaml:"significance_line"`
	SignificancePage   SignificanceResult   `yaml:"significance_page"`
	LocalCompatibility LocalCompatibility   `yaml:"local_compatibility"`
	Reliability        float64              `yaml:"reliability"`
	Score              float64              `yaml:"score"`
}

type Output struct {
	Meta             Meta        `yaml:"meta"`
	Parameters       Parameters  `yaml:"parameters"`
	Methodology      Methodology `yaml:"methodology"`
	Candidates       []Candidate `yaml:"candidates"`
	LikelyLocalPairs []Candidate `yaml:"likely_local_pairs"`
}
