package notation

import (
	"fmt"
	"sort"
	"strconv"
)

// OutputType is the frozen B04 comparison semantics for one metric_id
// (DISTRIBUTION_OUTPUT_CONTRACT.md / METRIC_OUTPUT_TYPES.tsv).
type OutputType string

const (
	OutputScalar      OutputType = "SCALAR"
	OutputCategorical OutputType = "CATEGORICAL_DISTRIBUTION"
	OutputOrdered     OutputType = "ORDERED_DISTRIBUTION"
	OutputCurve       OutputType = "CURVE"
	OutputDescriptive OutputType = "DESCRIPTIVE_ONLY"
)

// MetricOutputTypeRow is one row of METRIC_OUTPUT_TYPES.tsv.
type MetricOutputTypeRow struct {
	MetricID   string
	OutputType OutputType
	Note       string
}

// descriptiveOnly are metrics that are raw counts or cardinalities rather
// than scale-normalizable proportions; they are reported but never enter a
// calibrated scalar distance.
var descriptiveOnly = map[string]bool{
	"G01_ALPHABET_SIZE":         true,
	"S04_REPEATED_BIGRAM_TYPES": true,
	"S05_REPEATED_TRIGRAM_TYPES": true,
}

// MetricOutputTypes enumerates the frozen output type of every metric_id
// this analyzer can emit, plus the curve ids and the distribution-only ids
// that are serialized in DISTRIBUTIONS.tsv but never in the scalar Metrics
// list (B04 section 15). Every metric has exactly one output type.
func MetricOutputTypes() []MetricOutputTypeRow {
	var out []MetricOutputTypeRow
	for _, d := range MetricRegistry() {
		t := OutputScalar
		note := ""
		switch {
		case descriptiveOnly[d.ID]:
			t = OutputDescriptive
			note = "raw count/cardinality; not scale-normalized"
		case d.ID == "T01_MEAN_TOKEN_LENGTH":
			note = "paired ordered distribution: T01_TOKEN_LENGTH_DISTRIBUTION"
		case d.ID == "T11_POSITIONAL_RESTRICTION_DENSITY":
			note = "paired categorical distribution: T11_POSITIONAL_RESTRICTION_DISTRIBUTION"
		}
		out = append(out, MetricOutputTypeRow{MetricID: d.ID, OutputType: t, Note: note})
	}
	out = append(out,
		MetricOutputTypeRow{"T01_TOKEN_LENGTH_DISTRIBUTION", OutputOrdered, "integer token-length support, no arbitrary bins"},
		MetricOutputTypeRow{"T11_POSITIONAL_RESTRICTION_DISTRIBUTION", OutputCategorical, "category universe {INITIAL_RESTRICTED,INTERNAL_RESTRICTED,FINAL_RESTRICTED}"},
		MetricOutputTypeRow{"A2_BIGRAM_TYPES", OutputCurve, "accumulation curve, checkpoints 5k/10k/20k/39380"},
		MetricOutputTypeRow{"A3_TRIGRAM_TYPES", OutputCurve, "accumulation curve, checkpoints 5k/10k/20k/39380"},
		MetricOutputTypeRow{"AT_TRANSITION_TYPES", OutputCurve, "accumulation curve, checkpoints 5k/10k/20k/39380"},
	)
	sort.Slice(out, func(i, j int) bool { return out[i].MetricID < out[j].MetricID })
	return out
}

// DistributionPoint is one row of DISTRIBUTIONS.tsv.
type DistributionPoint struct {
	CorpusID, RepresentationID string
	MetricID, SupportID        string
	BinOrCategory              string
	Value                      float64
	Probability                float64
	Comparable                 bool
	Reason                     string
}

// PositionalRestrictionCategories is the frozen category universe for
// T11_POSITIONAL_RESTRICTION_DISTRIBUTION (B04 section 16). It never grows
// or shrinks per corpus.
var PositionalRestrictionCategories = []string{"INITIAL_RESTRICTED", "INTERNAL_RESTRICTED", "FINAL_RESTRICTED"}

// TokenLengthDistribution builds the ORDERED_DISTRIBUTION for token length
// directly over the integer support, with no arbitrary bins (B04 section
// 17). It is label-invariant: renaming symbols never changes token length.
func TokenLengthDistribution(rs []Record, corpusID, representationID string) []DistributionPoint {
	counts := map[int]int{}
	maxLen := 0
	for _, r := range rs {
		n := len(r.Symbols)
		counts[n]++
		if n > maxLen {
			maxLen = n
		}
	}
	var out []DistributionPoint
	if maxLen == 0 {
		return out
	}
	total := len(rs)
	for l := 1; l <= maxLen; l++ {
		out = append(out, DistributionPoint{
			CorpusID: corpusID, RepresentationID: representationID,
			MetricID: "T01_TOKEN_LENGTH_DISTRIBUTION", SupportID: "TOKEN_LENGTH_INTEGER",
			BinOrCategory: strconv.Itoa(l), Value: float64(l),
			Probability: safe(float64(counts[l]), float64(total)), Comparable: true,
		})
	}
	return out
}

// PositionalRestrictionDistribution builds the CATEGORICAL_DISTRIBUTION over
// the frozen {INITIAL,INTERNAL,FINAL}_RESTRICTED category universe: the
// share of each restriction type among all observed restricted-symbol
// slots. If the alphabet has no restricted symbol at all, the distribution
// is explicitly marked not comparable rather than silently emitting NaN
// (B04 D5).
func PositionalRestrictionDistribution(rs []Record, corpusID, representationID string) []DistributionPoint {
	alphabet, initial, internal, final := map[string]bool{}, map[string]bool{}, map[string]bool{}, map[string]bool{}
	for _, r := range rs {
		for i, s := range r.Symbols {
			alphabet[s] = true
			switch {
			case i == 0:
				initial[s] = true
			case i == len(r.Symbols)-1:
				final[s] = true
			default:
				internal[s] = true
			}
		}
	}
	counts := map[string]int{
		"INITIAL_RESTRICTED":  len(alphabet) - len(initial),
		"INTERNAL_RESTRICTED": len(alphabet) - len(internal),
		"FINAL_RESTRICTED":    len(alphabet) - len(final),
	}
	total := 0
	for _, c := range PositionalRestrictionCategories {
		total += counts[c]
	}
	var out []DistributionPoint
	for _, c := range PositionalRestrictionCategories {
		p := DistributionPoint{CorpusID: corpusID, RepresentationID: representationID, MetricID: "T11_POSITIONAL_RESTRICTION_DISTRIBUTION", SupportID: "POSITION_CATEGORY", BinOrCategory: c, Value: float64(counts[c])}
		if total == 0 {
			p.Comparable = false
			p.Reason = "no restricted symbol observed at any position"
		} else {
			p.Comparable = true
			p.Probability = float64(counts[c]) / float64(total)
		}
		out = append(out, p)
	}
	return out
}

// BuildDistributions returns every frozen DISTRIBUTIONS.tsv row for rs.
func BuildDistributions(rs []Record, corpusID, representationID string) []DistributionPoint {
	var out []DistributionPoint
	out = append(out, TokenLengthDistribution(rs, corpusID, representationID)...)
	out = append(out, PositionalRestrictionDistribution(rs, corpusID, representationID)...)
	return out
}

// CategoricalJensenShannon computes JS divergence between two categorical
// distributions strictly over the given frozen common support (B04 section
// 16 / D6). A category present in p or q but absent from commonSupport is a
// candidate-specific support and is rejected rather than silently included.
func CategoricalJensenShannon(commonSupport []string, p, q []DistributionPoint) (float64, error) {
	pm, qm := map[string]float64{}, map[string]float64{}
	for _, x := range p {
		pm[x.BinOrCategory] = x.Probability
	}
	for _, x := range q {
		qm[x.BinOrCategory] = x.Probability
	}
	allowed := map[string]bool{}
	for _, c := range commonSupport {
		allowed[c] = true
	}
	for cat := range pm {
		if !allowed[cat] {
			return 0, fmt.Errorf("category %q is outside the frozen common support", cat)
		}
	}
	for cat := range qm {
		if !allowed[cat] {
			return 0, fmt.Errorf("category %q is outside the frozen common support", cat)
		}
	}
	pv := make([]float64, len(commonSupport))
	qv := make([]float64, len(commonSupport))
	for i, c := range commonSupport {
		pv[i] = pm[c] // missing observed category = probability 0 on the frozen common support
		qv[i] = qm[c]
	}
	return JensenShannon(pv, qv)
}

// OrderedWasserstein computes Wasserstein-1 distance between two ordered
// distributions built by TokenLengthDistribution or an equivalent frozen
// integer-support builder.
func OrderedWasserstein(p, q []DistributionPoint) (float64, error) {
	xp, pv := make([]float64, len(p)), make([]float64, len(p))
	for i, x := range p {
		xp[i], pv[i] = x.Value, x.Probability
	}
	xq, qv := make([]float64, len(q)), make([]float64, len(q))
	for i, x := range q {
		xq[i], qv[i] = x.Value, x.Probability
	}
	return Wasserstein1(xp, pv, xq, qv)
}
