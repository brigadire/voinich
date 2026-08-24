// Package task82b implements the Task82b historical-shorthand and
// selective-extraction fingerprint experiment (tasks_ph2/task82b.txt).
//
// It is a frozen-feature-extraction consumer of internal/fingerprintv2
// only: it never redefines a Fingerprint V2 metric, null model, or
// threshold. Two independent branches share the same F2 subset (the
// F2_COMMON_DIRECT + F2_ASSEMBLER_PROJECTION union already validated by
// Task82a.1 for non-Voynich, non-manuscript-hierarchy corpora):
//
//   - Branch S (shorthand/abbreviation): real paired abbreviated/expanded
//     historical text (Burchards Dekret Digital), plus matched nulls.
//   - Branch A (acrostic/selective extraction): a small, frozen,
//     non-combinatorial grid of deterministic extraction operators
//     applied to natural-language carriers (Doyle, Longfellow, Astafiev),
//     plus matched nulls.
//
// Every Voynich-transcription path is refused by assertNoVoynichPath
// (task82b.txt sec.4/71): this package must never construct an F2 config
// over Voynich data before Task83.
package task82b

// F2Metric is one row of the frozen F2 vector for one corpus/variant.
type F2Metric struct {
	MetricID       string  `json:"metric_id"`
	Family         string  `json:"family"`
	Classification string  `json:"classification"` // CORE or SUPPORTING (F2_COMMON_DIRECT.tsv / F2_ASSEMBLER_PROJECTION.tsv)
	Source         string  `json:"source"`         // DIRECT or ASSEMBLER_PROJECTION
	Value          float64 `json:"value"`
	Available      bool    `json:"available"`
	MissingReason  string  `json:"missing_reason,omitempty"`
}

// F2Vector is the full frozen-subset F2 measurement for one text.
type F2Vector struct {
	JobID    string     `json:"job_id"`
	CorpusID string     `json:"corpus_id"`
	SHA256   string     `json:"sha256"`
	Tokens   int        `json:"tokens"`
	Types    int        `json:"types"`
	Lines    int        `json:"lines"`
	Warnings []string   `json:"warnings,omitempty"`
	Metrics  []F2Metric `json:"metrics"`
}

// Get returns the value and availability of one metric by ID.
func (v F2Vector) Get(id string) (float64, bool) {
	for _, m := range v.Metrics {
		if m.MetricID == id {
			return m.Value, m.Available
		}
	}
	return 0, false
}

// CoreMetricIDs and SupportingMetricIDs are the frozen task82b F2 subset,
// verbatim from research/phase2/task82a1/F2_COMMON_DIRECT.tsv and
// F2_ASSEMBLER_PROJECTION.tsv (task82b.txt sec.3). Order is fixed so every
// F2Vector lists metrics identically.
var CoreMetricIDs = []string{
	"EF1_GIANT_COMPONENT_SHARE",
	"EF2_GLOBAL_CLUSTERING",
	"EF3_DEGREE_FREQUENCY_SPEARMAN",
	"2DL1_LAYOUT_POSITION_MI",
	"BP1_BOUNDARY_TOKEN_NMI",
	"LS2_POSITIONAL_LEXICON_NMI",
	"LS3_BOUNDARY_LENGTH_ASYMMETRY",
}

var SupportingMetricIDs = []string{
	"EF1_ISOLATE_SHARE",
	"LP1_RULE_SUPPORT_GINI",
	"LP4_PREFIX_ATTACHMENT_NMI",
	"LP4_SUFFIX_ATTACHMENT_NMI",
	"cs2/prev-family-current-family",
	"LS1_LINE_LENGTH_CV",
	"LS4_WITHIN_LINE_EXACT_REPETITION",
	"cs1/family-line-position",
	"cs6/family-diversity-x-line-length",
	"cs7/edit-distance-x-structural-distance",
}

// AllMetricIDs is CoreMetricIDs followed by SupportingMetricIDs.
func AllMetricIDs() []string {
	out := make([]string, 0, len(CoreMetricIDs)+len(SupportingMetricIDs))
	out = append(out, CoreMetricIDs...)
	out = append(out, SupportingMetricIDs...)
	return out
}
