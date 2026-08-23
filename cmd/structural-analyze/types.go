package main

type Position struct {
	Position int `yaml:"position"`
	Count    int `yaml:"count"`
}

type Neighbor struct {
	Token string `yaml:"token"`
	Count int    `yaml:"count"`
}

type DictionaryToken struct {
	Token            string     `yaml:"token"`
	Count            int        `yaml:"count"`
	PositionInString []Position `yaml:"position_in_string"`
	WordBefore       []Neighbor `yaml:"word_before"`
	WordAfter        []Neighbor `yaml:"word_after"`
	LineStartCount   int        `yaml:"line_start_count"`
	LineEndCount     int        `yaml:"line_end_count"`
}

type EnvironmentAnalysis struct {
	Unique  int     `yaml:"unique"`
	Entropy float64 `yaml:"entropy"`
}

type TransitionInput struct {
	Token              string  `yaml:"token"`
	Count              int     `yaml:"count"`
	Probability        float64 `yaml:"probability"`
	ReverseCount       int     `yaml:"reverse_count"`
	ReverseProbability float64 `yaml:"reverse_probability"`
	Asymmetry          float64 `yaml:"asymmetry"`
}

type SelfTransitionInput struct {
	Count       int     `yaml:"count"`
	Probability float64 `yaml:"probability"`
}

type StructuralScoresInput struct {
	PositionalSpecialization float64 `yaml:"positional_specialization"`
	SuccessorRestriction     float64 `yaml:"successor_restriction"`
	PredecessorRestriction   float64 `yaml:"predecessor_restriction"`
}

type TokenAnalysisInput struct {
	Token            string                `yaml:"token"`
	Count            int                   `yaml:"count"`
	StartProbability float64               `yaml:"start_probability"`
	EndProbability   float64               `yaml:"end_probability"`
	Left             EnvironmentAnalysis   `yaml:"left"`
	Right            EnvironmentAnalysis   `yaml:"right"`
	Transitions      []TransitionInput     `yaml:"transitions"`
	SelfTransition   SelfTransitionInput   `yaml:"self_transition"`
	StructuralScores StructuralScoresInput `yaml:"structural_scores"`
}

type Parameters struct {
	MinTokenCountForRanking  int     `yaml:"min_token_count_for_ranking"`
	MinTransitionCount       int     `yaml:"min_transition_count"`
	MinContextObservations   int     `yaml:"min_context_observations"`
	MinSelfTransitionCount   int     `yaml:"min_self_transition_count"`
	ReliabilityPriorCount    float64 `yaml:"reliability_prior_count"`
	MinEquivalenceSimilarity float64 `yaml:"min_equivalence_similarity"`
	MaxItemsPerSection       int     `yaml:"max_items_per_section"`
	MaxEquivalenceCandidates int     `yaml:"max_equivalence_candidates"`
	DominantContextLimit     int     `yaml:"dominant_context_limit"`
}

type Meta struct {
	DatasetVersion       int     `yaml:"dataset_version"`
	TokenOccurrences     int     `yaml:"token_occurrences"`
	UniqueTokens         int     `yaml:"unique_tokens"`
	Lines                int     `yaml:"lines"`
	Transitions          int     `yaml:"transitions"`
	PositionObservations int     `yaml:"position_observations"`
	PositionCoverage     float64 `yaml:"position_coverage"`
}

type Methodology struct {
	PositionalScore string `yaml:"positional_score"`
	Predictability  string `yaml:"predictability"`
	Reliability     string `yaml:"reliability"`
	Asymmetry       string `yaml:"asymmetry"`
	PMI             string `yaml:"pmi"`
	LogLikelihood   string `yaml:"log_likelihood"`
	Equivalence     string `yaml:"equivalence"`
}

type PositionBaseline struct {
	Position    int     `yaml:"position"`
	Count       int     `yaml:"count"`
	Probability float64 `yaml:"probability"`
}

type PositionValue struct {
	Position          int     `yaml:"position"`
	Count             int     `yaml:"count"`
	Probability       float64 `yaml:"probability"`
	CorpusProbability float64 `yaml:"corpus_probability"`
}

type PositionalResult struct {
	Token                string          `yaml:"token"`
	Count                int             `yaml:"count"`
	LineStartCount       int             `yaml:"line_start_count"`
	LineEndCount         int             `yaml:"line_end_count"`
	StartProbability     float64         `yaml:"start_probability"`
	EndProbability       float64         `yaml:"end_probability"`
	PositionObservations int             `yaml:"position_observations"`
	PositionCoverage     float64         `yaml:"position_coverage"`
	Positions            []PositionValue `yaml:"positions"`
	Score                float64         `yaml:"score"`
	Reliability          float64         `yaml:"reliability"`
	RankingScore         float64         `yaml:"ranking_score"`
}

type DominantNeighbor struct {
	Token       string  `yaml:"token"`
	Count       int     `yaml:"count"`
	Probability float64 `yaml:"probability"`
}

type ConstraintResult struct {
	Token                string             `yaml:"token"`
	Count                int                `yaml:"count"`
	ObservedTransitions  int                `yaml:"observed_transitions"`
	UniqueSuccessors     int                `yaml:"unique_successors,omitempty"`
	UniquePredecessors   int                `yaml:"unique_predecessors,omitempty"`
	Entropy              float64            `yaml:"entropy"`
	MaxEntropy           float64            `yaml:"max_entropy"`
	NormalizedEntropy    float64            `yaml:"normalized_entropy"`
	Predictability       float64            `yaml:"predictability"`
	Reliability          float64            `yaml:"reliability"`
	RankingScore         float64            `yaml:"ranking_score"`
	DominantSuccessors   []DominantNeighbor `yaml:"dominant_successors,omitempty"`
	DominantPredecessors []DominantNeighbor `yaml:"dominant_predecessors,omitempty"`
}

type SignificantTransition struct {
	From               string  `yaml:"from"`
	To                 string  `yaml:"to"`
	Count              int     `yaml:"count"`
	FromTransitions    int     `yaml:"from_transitions"`
	ToIncoming         int     `yaml:"to_incoming"`
	Probability        float64 `yaml:"probability"`
	ReverseCount       int     `yaml:"reverse_count"`
	ReverseProbability float64 `yaml:"reverse_probability"`
	Asymmetry          float64 `yaml:"asymmetry"`
	Expected           float64 `yaml:"expected"`
	PMI                float64 `yaml:"pmi"`
	LogLikelihood      float64 `yaml:"log_likelihood"`
}

type SelfTransitionResult struct {
	Token        string  `yaml:"token"`
	TokenCount   int     `yaml:"token_count"`
	Count        int     `yaml:"count"`
	Outgoing     int     `yaml:"outgoing"`
	Incoming     int     `yaml:"incoming"`
	Probability  float64 `yaml:"probability"`
	Expected     float64 `yaml:"expected"`
	Enrichment   float64 `yaml:"enrichment"`
	Reliability  float64 `yaml:"reliability"`
	RankingScore float64 `yaml:"ranking_score"`
}

type EquivalenceCandidate struct {
	TokenA                 string  `yaml:"token_a"`
	TokenB                 string  `yaml:"token_b"`
	CountA                 int     `yaml:"count_a"`
	CountB                 int     `yaml:"count_b"`
	Similarity             float64 `yaml:"similarity"`
	PositionSimilarity     float64 `yaml:"position_similarity"`
	LeftContextSimilarity  float64 `yaml:"left_context_similarity"`
	RightContextSimilarity float64 `yaml:"right_context_similarity"`
	Reliability            float64 `yaml:"reliability"`
	RankingScore           float64 `yaml:"ranking_score"`
}

type Output struct {
	Meta                      Meta                    `yaml:"meta"`
	Parameters                Parameters              `yaml:"parameters"`
	Methodology               Methodology             `yaml:"methodology"`
	PositionBaseline          []PositionBaseline      `yaml:"position_baseline"`
	PositionalSpecialization  []PositionalResult      `yaml:"positional_specialization"`
	SuccessorPredictability   []ConstraintResult      `yaml:"successor_predictability"`
	PredecessorPredictability []ConstraintResult      `yaml:"predecessor_predictability"`
	SignificantTransitions    []SignificantTransition `yaml:"significant_transitions"`
	SelfTransitions           []SelfTransitionResult  `yaml:"self_transitions"`
	EquivalenceCandidates     []EquivalenceCandidate  `yaml:"equivalence_candidates"`
}
