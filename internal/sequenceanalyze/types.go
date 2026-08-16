package sequenceanalyze

type Parameters struct {
	MinN                   int `yaml:"min_n"`
	MaxN                   int `yaml:"max_n"`
	MinCount               int `yaml:"min_count"`
	MaxItems               int `yaml:"max_items"`
	ContextLimit           int `yaml:"context_limit"`
	MaxContextLength       int `yaml:"max_context_length"`
	ContextMinObservations int `yaml:"context_min_observations"`
	ContextMaxItems        int `yaml:"context_max_items"`
}

// DefaultParameters returns the same defaults sequence-analyze's CLI flags
// fall back to. Callers that invoke AnalyzeFile/AnalyzeLines in-process
// (rather than shelling out to the compiled sequence-analyze binary) should
// use this instead of duplicating the literal values, so the two paths
// cannot silently drift apart.
func DefaultParameters() Parameters {
	return Parameters{
		MinN: 2, MaxN: 8, MinCount: 2, MaxItems: 200, ContextLimit: 10,
		MaxContextLength: 7, ContextMinObservations: 10, ContextMaxItems: 200,
	}
}

type Meta struct {
	TokenOccurrences int `yaml:"token_occurrences"`
	Lines            int `yaml:"lines"`
	Transitions      int `yaml:"transitions"`
}

type Methodology struct {
	SequenceBoundary   string `yaml:"sequence_boundary"`
	Tokenization       string `yaml:"tokenization"`
	Count              string `yaml:"count"`
	LineCount          string `yaml:"line_count"`
	Entropy            string `yaml:"entropy"`
	NormalizedEntropy  string `yaml:"normalized_entropy"`
	Predictability     string `yaml:"predictability"`
	MaximalRepeated    string `yaml:"maximal_repeated_sequence"`
	Coordinates        string `yaml:"coordinates"`
	OutputLimits       string `yaml:"output_limits"`
	CrossLine          string `yaml:"cross_line"`
	ConditionalEntropy string `yaml:"conditional_entropy"`
	ContextCoverage    string `yaml:"context_coverage"`
	EntropyDelta       string `yaml:"entropy_delta"`
	Interpretation     string `yaml:"interpretation"`
}

type NGramSummary struct {
	N                  int `yaml:"n"`
	TotalOccurrences   int `yaml:"total_occurrences"`
	Unique             int `yaml:"unique"`
	Repeated           int `yaml:"repeated"`
	MultiOccurrence    int `yaml:"multi_occurrence"`
	MultiLine          int `yaml:"multi_line"`
	MultiLineRepeated  int `yaml:"multi_line_repeated"`
	SingleLineRepeated int `yaml:"single_line_repeated"`
	Hapax              int `yaml:"hapax"`
	MaxCount           int `yaml:"max_count"`
}

type NGramResult struct {
	Tokens           []string `yaml:"tokens"`
	N                int      `yaml:"n"`
	Count            int      `yaml:"count"`
	LineCount        int      `yaml:"line_count"`
	CrossLine        bool     `yaml:"cross_line"`
	StartCount       int      `yaml:"start_count"`
	EndCount         int      `yaml:"end_count"`
	StartProbability float64  `yaml:"start_probability"`
	EndProbability   float64  `yaml:"end_probability"`
}

type ContextToken struct {
	Token       string  `yaml:"token"`
	Count       int     `yaml:"count"`
	Probability float64 `yaml:"probability"`
}

type ContinuationResult struct {
	Prefix                []string       `yaml:"prefix"`
	N                     int            `yaml:"n"`
	PrefixCount           int            `yaml:"prefix_count"`
	ObservedContinuations int            `yaml:"observed_continuations"`
	LineEndCount          int            `yaml:"line_end_count"`
	UniqueContinuations   int            `yaml:"unique_continuations"`
	Entropy               float64        `yaml:"entropy"`
	NormalizedEntropy     float64        `yaml:"normalized_entropy"`
	Predictability        float64        `yaml:"predictability"`
	Next                  []ContextToken `yaml:"next"`
}

type PredecessorResult struct {
	Suffix               []string       `yaml:"suffix"`
	N                    int            `yaml:"n"`
	SuffixCount          int            `yaml:"suffix_count"`
	ObservedPredecessors int            `yaml:"observed_predecessors"`
	LineStartCount       int            `yaml:"line_start_count"`
	UniquePredecessors   int            `yaml:"unique_predecessors"`
	Entropy              float64        `yaml:"entropy"`
	NormalizedEntropy    float64        `yaml:"normalized_entropy"`
	Predictability       float64        `yaml:"predictability"`
	Previous             []ContextToken `yaml:"previous"`
}

type DominantExtension struct {
	Token       string  `yaml:"token"`
	Count       int     `yaml:"count"`
	Probability float64 `yaml:"probability"`
}

type ExtensionSide struct {
	Observed int                `yaml:"observed"`
	Unique   int                `yaml:"unique"`
	Dominant *DominantExtension `yaml:"dominant,omitempty"`
}

type ExtensionResult struct {
	Sequence       []string      `yaml:"sequence"`
	N              int           `yaml:"n"`
	Count          int           `yaml:"count"`
	Left           ExtensionSide `yaml:"left"`
	Right          ExtensionSide `yaml:"right"`
	LineStartCount int           `yaml:"line_start_count"`
	LineEndCount   int           `yaml:"line_end_count"`
}

type Coordinate struct {
	Line        int `yaml:"line"`
	TokenOffset int `yaml:"token_offset"`
}

type MaximalRepeatedSequence struct {
	Tokens      []string     `yaml:"tokens"`
	N           int          `yaml:"n"`
	Count       int          `yaml:"count"`
	LineCount   int          `yaml:"line_count"`
	StartCount  int          `yaml:"start_count"`
	EndCount    int          `yaml:"end_count"`
	Occurrences []Coordinate `yaml:"occurrences"`
}

type ContextOrderResult struct {
	ContextLength                     int      `yaml:"context_length"`
	Observations                      int      `yaml:"observations"`
	UniqueContexts                    int      `yaml:"unique_contexts"`
	SingletonContexts                 int      `yaml:"singleton_contexts"`
	RepeatedContexts                  int      `yaml:"repeated_contexts"`
	SingletonFraction                 float64  `yaml:"singleton_fraction"`
	ObservationsInRepeatedContexts    int      `yaml:"observations_in_repeated_contexts"`
	RepeatedContextCoverage           float64  `yaml:"repeated_context_coverage"`
	ConditionalEntropy                float64  `yaml:"conditional_entropy"`
	Perplexity                        float64  `yaml:"perplexity"`
	RepeatedContextConditionalEntropy float64  `yaml:"repeated_context_conditional_entropy"`
	RepeatedContextPerplexity         float64  `yaml:"repeated_context_perplexity"`
	EntropyDeltaFromPrevious          *float64 `yaml:"entropy_delta_from_previous,omitempty"`
	RepeatedEntropyDeltaFromPrevious  *float64 `yaml:"repeated_entropy_delta_from_previous,omitempty"`
}

type DominantNext struct {
	Token       string  `yaml:"token"`
	Count       int     `yaml:"count"`
	Probability float64 `yaml:"probability"`
}

type ContextExtensionResult struct {
	ShortContext      []string     `yaml:"short_context"`
	LongContext       []string     `yaml:"long_context"`
	ShortCount        int          `yaml:"short_count"`
	LongCount         int          `yaml:"long_count"`
	ShortEntropy      float64      `yaml:"short_entropy"`
	LongEntropy       float64      `yaml:"long_entropy"`
	EntropyReduction  float64      `yaml:"entropy_reduction"`
	ShortUniqueNext   int          `yaml:"short_unique_next"`
	LongUniqueNext    int          `yaml:"long_unique_next"`
	ShortDominantNext DominantNext `yaml:"short_dominant_next"`
	LongDominantNext  DominantNext `yaml:"long_dominant_next"`
}

type Output struct {
	Meta                      Meta                      `yaml:"meta"`
	Parameters                Parameters                `yaml:"parameters"`
	Methodology               Methodology               `yaml:"methodology"`
	NGramSummary              []NGramSummary            `yaml:"ngram_summary"`
	RepeatedNGrams            map[int][]NGramResult     `yaml:"repeated_ngrams"`
	CrossLineRepeatedNGrams   map[int][]NGramResult     `yaml:"cross_line_repeated_ngrams"`
	Continuations             []ContinuationResult      `yaml:"continuations"`
	PredecessorContexts       []PredecessorResult       `yaml:"predecessor_contexts"`
	Extensions                []ExtensionResult         `yaml:"extensions"`
	MaximalRepeatedSequences  []MaximalRepeatedSequence `yaml:"maximal_repeated_sequences"`
	MaximalCrossLineSequences []MaximalRepeatedSequence `yaml:"maximal_cross_line_sequences"`
	ContextOrderAnalysis      []ContextOrderResult      `yaml:"context_order_analysis"`
	ContextExtensions         []ContextExtensionResult  `yaml:"context_extensions"`
}
