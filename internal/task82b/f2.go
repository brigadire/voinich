package task82b

import (
	"strings"

	"zcore.dev/voinich/internal/fingerprintv2"
)

// F2CorpusVersion is the frozen config version internal/fingerprintv2
// accepts (task79-v1/task82a lineage). Task82b introduces no new version.
const F2CorpusVersion = "fingerprint-v2-lexical-paradigms-v1"

// F2Repetitions is Task82b's own preregistered, cost-driven, target-blind
// Repetitions value (task82b.txt sec.60, "не только p-values"). A pilot
// measurement (research/phase2/task82b/TASK82B_COST_MODEL.tsv) found the
// point estimates this package actually uses (EF1/EF2/EF3/LP1/LP4 and the
// cross-scale ObservedStatistic values) IDENTICAL at Repetitions in
// {1,5,10,30} for Doyle/Longfellow/Astafiev; only the internal
// null-significance precision of cross-scale SUPPORTED/NOT_SUPPORTED
// verdicts (all SUPPORTING, never CORE, in this package's registry) and
// warning diagnostics depend on it. Task82b's grid needs several hundred
// F2 calls (vs Task82a's 468 over much smaller synthetic corpora), so 5
// is chosen over Task82a's 30 for cost, exactly as Task82a itself reduced
// the canonical Task79 control's 1000 to 30 for the same documented
// reason.
const F2Repetitions = 5

// buildF2Config constructs the frozen-feature-extraction-only config for
// one corpus text file already written to disk.
func buildF2Config(corpusPath, corpusID string, seed int64, outDir string) (fingerprintv2.Config, error) {
	if err := assertNoVoynichPath(corpusPath); err != nil {
		return fingerprintv2.Config{}, err
	}
	return fingerprintv2.Config{
		Version:             F2CorpusVersion,
		OutputDir:           outDir,
		Primary:             fingerprintv2.CorpusConfig{ID: corpusID, Path: corpusPath, GlyphMode: "natural"},
		Seed:                seed,
		Repetitions:         F2Repetitions,
		MinRuleSupport:      3,
		Alpha:               0.05,
		GraphSwaps:          10,
		DiagnosticTolerance: 0.20,
		Grammar:             fingerprintv2.GrammarConfig{Modes: []string{"structure-preserving", "frequency-aware"}},
	}, nil
}

// ExtractF2 runs the frozen extractor plus the free generic ordered-group
// projection (Task82a.1's fingerprintv2.OrderedGroupMetrics) over one
// already-written corpus text file, returning the full task82b.AllMetricIDs
// vector. groups is the same token-per-line partition used to write
// corpusPath, passed separately so OrderedGroupMetrics never re-parses the
// file with different tokenization rules.
func ExtractF2(corpusPath, jobID, corpusID string, seed int64, outDir string, groups [][]string) (F2Vector, error) {
	cfg, err := buildF2Config(corpusPath, corpusID, seed, outDir)
	if err != nil {
		return F2Vector{}, err
	}
	fp, err := fingerprintv2.Run(cfg)
	if err != nil {
		return F2Vector{}, err
	}
	m := fp.Primary.Metrics
	direct := map[string]float64{
		"EF1_GIANT_COMPONENT_SHARE":     m.EF1.GiantComponentShare,
		"EF1_ISOLATE_SHARE":             m.EF1.IsolateShare,
		"EF2_GLOBAL_CLUSTERING":         m.EF2.GlobalClustering,
		"EF3_DEGREE_FREQUENCY_SPEARMAN": m.EF3.SpearmanDegreeLogFrequency,
		"LP1_RULE_SUPPORT_GINI":         m.LP1.SupportGini,
		"LP4_PREFIX_ATTACHMENT_NMI":     m.LP4.Prefix.NormalizedMI,
		"LP4_SUFFIX_ATTACHMENT_NMI":     m.LP4.Suffix.NormalizedMI,
	}
	insufficient := fp.Primary.Metrics.LP2.ProductivityState == "INSUFFICIENT_SUPPORT"

	csByID := map[string]fingerprintv2.CrossScaleMetric{}
	if fp.Primary.CrossScale != nil {
		for _, c := range fp.Primary.CrossScale.Metrics {
			csByID[c.MetricID] = c
		}
	}

	generic := fingerprintv2.OrderedGroupMetrics(groups)

	classOf := map[string]string{}
	for _, id := range CoreMetricIDs {
		classOf[id] = "CORE"
	}
	for _, id := range SupportingMetricIDs {
		classOf[id] = "SUPPORTING"
	}

	var metrics []F2Metric
	for _, id := range AllMetricIDs() {
		row := F2Metric{MetricID: id, Family: family(id), Classification: classOf[id]}
		switch {
		case strings.HasPrefix(id, "cs"):
			row.Source = "ASSEMBLER_PROJECTION"
			cm, ok := csByID[id]
			switch {
			case !ok:
				row.MissingReason = "cross-scale metric not returned by frozen extractor"
			case cm.Status == "SUPPORTED" || cm.Status == "NOT_SUPPORTED":
				row.Available = true
				row.Value = cm.ObservedStatistic
			default:
				row.MissingReason = cm.Status + ": " + cm.Limitations
			}
		case isOrderedGroupMetric(id):
			row.Source = "ASSEMBLER_PROJECTION"
			v, ok := generic[id]
			if !ok {
				row.MissingReason = "OrderedGroupMetrics did not return this id"
				break
			}
			row.Available = true
			row.Value = v
		default:
			row.Source = "DIRECT"
			if insufficient {
				row.MissingReason = "INSUFFICIENT_SUPPORT: fewer than two vocabulary types in edit graph"
				break
			}
			row.Available = true
			row.Value = direct[id]
		}
		metrics = append(metrics, row)
	}

	types := map[string]bool{}
	tokens := 0
	for _, g := range groups {
		tokens += len(g)
		for _, t := range g {
			types[t] = true
		}
	}
	return F2Vector{
		JobID:    jobID,
		CorpusID: corpusID,
		SHA256:   fp.Primary.Corpus.SHA256,
		Tokens:   tokens,
		Types:    len(types),
		Lines:    len(groups),
		Warnings: fp.Warnings,
		Metrics:  metrics,
	}, nil
}

func isOrderedGroupMetric(id string) bool {
	switch id {
	case "2DL1_LAYOUT_POSITION_MI", "BP1_BOUNDARY_TOKEN_NMI", "LS1_LINE_LENGTH_CV", "LS2_POSITIONAL_LEXICON_NMI", "LS3_BOUNDARY_LENGTH_ASYMMETRY", "LS4_WITHIN_LINE_EXACT_REPETITION":
		return true
	default:
		return false
	}
}

func family(id string) string {
	switch {
	case strings.HasPrefix(id, "EF"):
		return "edit family"
	case strings.HasPrefix(id, "LP"):
		return "lexical paradigm"
	case strings.HasPrefix(id, "cs"):
		return "cross-scale"
	case strings.HasPrefix(id, "2DL"):
		return "2D-LITE"
	case strings.HasPrefix(id, "BP"):
		return "boundary"
	case strings.HasPrefix(id, "LS"):
		return "line"
	default:
		return "unknown"
	}
}
