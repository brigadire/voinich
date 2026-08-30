package structurecatalog

import "zcore.dev/voinich/internal/metadatavalidation"

const (
	SchemaVersion  = "vm-structural-catalog-v1"
	DefaultMinFreq = 10
)

type Config struct {
	CorpusPath, IVTFFPath, IT2aPath, IT2aIVTFFPath, OutputDir string
	MinFrequency                                              int
}

type Occurrence struct {
	Token       string
	Line, Index int
	Meta        metadatavalidation.TokenMetadata
}

type Corpus struct {
	Path, SHA, Transcription     string
	Lines                        [][]string
	Occurrences                  []Occurrence
	Counts                       map[string]int
	Inventory                    []rune
	MetadataAvailable            bool
	Folios, Sections, LocusTypes []string
}

type Rule struct {
	RuleID, Level, RuleType, LHS, RHS, Context                        string
	ObservedCount, OpportunityCount                                   int
	ObservedProbability, ExpectedCount, EffectSize, PRaw, QValue      float64
	ObservedStatus, CorpusRule, InferredStatus, Stability, Provenance string
}

type Family struct {
	ID     int
	Tokens []string
}

type Catalog struct {
	Config               Config
	Primary, Replication Corpus
	Families             []Family
	TokenFamily          map[string]int
	Rules                []Rule
	Summary              map[string]string
}
