package task82a

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"zcore.dev/voinich/internal/fingerprintv2"
)

// F2 is used strictly as a frozen feature extractor (task82a.txt sec.1,
// 47). This guard rejects any path that names a known Voynich transcription
// or IVTFF artifact, and is the only place Task82a is allowed to construct
// a fingerprintv2.CorpusConfig.
func assertNoVoynichPath(path string) error {
	lower := strings.ToLower(path)
	for _, bad := range []string{"voynich", "zl3b", "it2a", "eva.txt", "data/ivtff", "data_work/"} {
		if strings.Contains(lower, bad) {
			return fmt.Errorf("VOYNICH_FIREWALL: refusing to build an F2 config over %q", path)
		}
	}
	return nil
}

// F2RawMetric is one row of the frozen-raw F2 vector (task82a.txt sec.48).
type F2RawMetric struct {
	MetricID           string  `json:"metric_id"`
	MetricVersion      string  `json:"metric_version"`
	Family             string  `json:"family"`
	Classification     string  `json:"final_classification"` // CORE or SUPPORTING, from F2_METRIC_REGISTRY_FINAL.tsv
	Value              float64 `json:"value"`
	Available          bool    `json:"available"`
	Verdict            string  `json:"verdict,omitempty"` // cross-scale metrics only: SUPPORTED/NOT_SUPPORTED significance verdict from the frozen extractor
	MissingReason      string  `json:"missing_reason,omitempty"`
	BoundaryProvenance string  `json:"boundary_provenance"`
}

// f2Registry is the subset of F2_METRIC_REGISTRY_FINAL.tsv that Task82a
// attempts: the base edit-family/lexical-paradigm metrics (computable from
// vocabulary alone, no manuscript hierarchy) and task77's cross-scale
// metrics (computable from token adjacency/line index alone). The
// hierarchy/folio/locus/line-profile families gated behind
// fingerprintv2's Task79Config pipeline (2DL, BP, HR, LC, LS, PF) are
// deliberately NOT attempted here -- see TASK82A_DESIGN.md sec. "F2 scope"
// for the cost/scope justification recorded before generation.
var f2Registry = []struct {
	id, family, class string
}{
	{"EF1_GIANT_COMPONENT_SHARE", "edit family", "CORE"},
	{"EF1_ISOLATE_SHARE", "edit family", "SUPPORTING"},
	{"EF2_GLOBAL_CLUSTERING", "edit family", "CORE"},
	{"EF3_DEGREE_FREQUENCY_SPEARMAN", "edit family", "CORE"},
	{"LP1_RULE_SUPPORT_GINI", "lexical paradigm", "SUPPORTING"},
	{"LP4_PREFIX_ATTACHMENT_NMI", "lexical paradigm", "SUPPORTING"},
	{"LP4_SUFFIX_ATTACHMENT_NMI", "lexical paradigm", "SUPPORTING"},
	{"cs1/family-line-position", "cross-scale", "SUPPORTING"},
	{"cs2/prev-family-current-family", "cross-scale", "SUPPORTING"},
	{"cs3/family-locus-type", "cross-scale", "SUPPORTING"},
	{"cs4/family-currier", "cross-scale", "SUPPORTING"},
	{"cs4/family-section", "cross-scale", "SUPPORTING"},
	{"cs5/local-adjacency-x-regime", "cross-scale", "SUPPORTING"},
	{"cs6/family-diversity-x-line-length", "cross-scale", "SUPPORTING"},
	{"cs7/edit-distance-x-structural-distance", "cross-scale", "SUPPORTING"},
}

// NotAttemptedFamilies are the F2_METRIC_REGISTRY_FINAL.tsv families
// Task82a never invokes (task77/79's Task79Config-gated pipeline), recorded
// with a distinct status from fingerprintv2's own NOT_APPLICABLE so the
// coverage report never confuses "we didn't ask" with "the frozen tool
// determined it is inapplicable".
var NotAttemptedFamilies = []string{"2DL", "BP", "HR", "LC", "LS", "PF"}

const f2CorpusVersion = "fingerprint-v2-lexical-paradigms-v1"

// buildF2Config constructs the frozen-feature-extraction-only config for
// one job's assembled corpus text file. Repetitions is reduced from the
// canonical Task79 control's 1000 to a documented, target-blind,
// cost-driven value (TASK82A_COST_MODEL.tsv); no metric definition,
// weight, or null-model choice is altered.
func buildF2Config(corpusPath, corpusID string, seed int64, outDir string) (fingerprintv2.Config, error) {
	if err := assertNoVoynichPath(corpusPath); err != nil {
		return fingerprintv2.Config{}, err
	}
	return fingerprintv2.Config{
		Version:             f2CorpusVersion,
		OutputDir:           outDir,
		Primary:             fingerprintv2.CorpusConfig{ID: corpusID, Path: corpusPath, GlyphMode: "natural"},
		Seed:                seed,
		Repetitions:         f2Repetitions,
		MinRuleSupport:      3,
		Alpha:               0.05,
		GraphSwaps:          10,
		DiagnosticTolerance: 0.20,
		Grammar:             fingerprintv2.GrammarConfig{Modes: []string{"structure-preserving", "frequency-aware"}},
	}, nil
}

// f2Repetitions is a Task82a-specific, preregistered, cost-driven
// reduction of the canonical Task79 control's Repetitions=1000. It affects
// only null-distribution precision (wider confidence intervals), never a
// metric definition, weight, or inclusion decision (task82a.txt sec.52).
const f2Repetitions = 30

// extractF2 runs the frozen extractor over one assembled corpus text file
// and returns the raw vector for the metric IDs in f2Registry.
func extractF2(corpusPath, corpusID string, seed int64, outDir string) ([]F2RawMetric, []string, error) {
	cfg, err := buildF2Config(corpusPath, corpusID, seed, outDir)
	if err != nil {
		return nil, nil, err
	}
	fp, err := fingerprintv2.Run(cfg)
	if err != nil {
		return nil, nil, err
	}
	m := fp.Primary.Metrics
	values := map[string]float64{
		"EF1_GIANT_COMPONENT_SHARE":     m.EF1.GiantComponentShare,
		"EF1_ISOLATE_SHARE":             m.EF1.IsolateShare,
		"EF2_GLOBAL_CLUSTERING":         m.EF2.GlobalClustering,
		"EF3_DEGREE_FREQUENCY_SPEARMAN": m.EF3.SpearmanDegreeLogFrequency,
		"LP1_RULE_SUPPORT_GINI":         m.LP1.SupportGini,
		"LP4_PREFIX_ATTACHMENT_NMI":     m.LP4.Prefix.NormalizedMI,
		"LP4_SUFFIX_ATTACHMENT_NMI":     m.LP4.Suffix.NormalizedMI,
	}
	insufficient := len(fp.Primary.Metrics.LP2.ProductivityState) > 0 && fp.Primary.Metrics.LP2.ProductivityState == "INSUFFICIENT_SUPPORT"
	var cs []fingerprintv2.CrossScaleMetric
	if fp.Primary.CrossScale != nil {
		cs = fp.Primary.CrossScale.Metrics
	}
	csByID := map[string]fingerprintv2.CrossScaleMetric{}
	for _, c := range cs {
		csByID[c.MetricID] = c
	}
	out := make([]F2RawMetric, 0, len(f2Registry))
	for _, reg := range f2Registry {
		row := F2RawMetric{MetricID: reg.id, MetricVersion: fp.Version, Family: reg.family, Classification: reg.class, BoundaryProvenance: "ASSEMBLER_DEFINED"}
		if strings.HasPrefix(reg.id, "cs") {
			cm, ok := csByID[reg.id]
			switch {
			case !ok:
				row.Available = false
				row.MissingReason = "cross-scale metric not returned by frozen extractor"
			case cm.Status == "SUPPORTED" || cm.Status == "NOT_SUPPORTED":
				// The frozen extractor computed a real observed statistic
				// with sufficient N; Status here is its significance
				// verdict against the registered null model, not an
				// availability flag (internal/fingerprintv2/
				// crossscale_pipeline.go's post-FDR classification loop).
				row.Available = true
				row.Value = cm.ObservedStatistic
				row.Verdict = cm.Status
			default:
				// NOT_APPLICABLE (no metadata) or INCONCLUSIVE (below the
				// frozen sample-size floor): genuinely unavailable.
				row.Available = false
				row.MissingReason = cm.Status + ": " + cm.Limitations
			}
			out = append(out, row)
			continue
		}
		if insufficient {
			row.Available = false
			row.MissingReason = "INSUFFICIENT_SUPPORT: fewer than two vocabulary types in edit graph"
			out = append(out, row)
			continue
		}
		row.Available = true
		row.Value = values[reg.id]
		out = append(out, row)
	}
	return out, fp.Warnings, nil
}

// writeAssembledCorpusFile writes one physical text line per chunk
// (task82a's ASSEMBLER_DEFINED line convention), each line holding that
// chunk's whitespace-separated tokens, so
// genericsegmentation.ReadCorpus recovers exactly the chunk boundaries F2
// will treat as "lines".
func writeAssembledCorpusFile(path string, lines [][]string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	var b strings.Builder
	for _, line := range lines {
		b.WriteString(strings.Join(line, " "))
		b.WriteByte('\n')
	}
	return os.WriteFile(path, []byte(b.String()), 0o644)
}
