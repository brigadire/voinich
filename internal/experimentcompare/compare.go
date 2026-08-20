// Package experimentcompare extracts a fixed, deliberately small set of
// comparable measurements from completed pipeline experiments.  It does not
// run pipeline stages and it never infers a document class.
package experimentcompare

import (
	"bufio"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"
)

const (
	FingerprintSchemaVersion = 1
	FormulasVersion          = 1
)

type Experiment struct {
	ID, Path, InputMode, CorpusPath, CorpusSHA256, GitCommit string
	TokenCount, VocabularySize                               int64
	Applicable, Unavailable                                  []string
	Raw, Normalized                                          map[string]float64
	Status                                                   map[string]string
	ManifestSHA256, ArtifactSHA256                           string
}
type Options struct {
	Experiments   []string
	OutputDir     string
	AllowUnfrozen bool
	Args          []string
	GitCommit     string
}
type manifest struct {
	ExperimentID string `json:"experiment_id"`
	GitCommit    string `json:"git_commit"`
	CorpusPath   string `json:"corpus_path"`
	CorpusSHA256 string `json:"corpus_sha256"`
	InputMode    string `json:"input_mode"`
	Stages       []struct {
		Name   string `json:"name"`
		Status string `json:"status"`
	} `json:"stages"`
}
type runState struct {
	Stages []struct {
		Name   string `json:"name"`
		Status string `json:"status"`
	} `json:"stages"`
}
type comparisonManifest struct {
	SchemaVersion   int               `json:"fingerprint_schema_version"`
	FormulasVersion int               `json:"formulas_version"`
	Experiments     []string          `json:"experiment_identities"`
	ManifestHashes  map[string]string `json:"experiment_manifest_sha256"`
	ArtifactHashes  map[string]string `json:"experiment_artifact_sha256"`
	GitCommit       string            `json:"comparison_program_git_commit"`
	Args            []string          `json:"cli_arguments"`
	Warnings        []string          `json:"warnings,omitempty"`
}

var formulas = map[string]string{
	"corpus.eligible_token_rate":       "eligible_tokens / token_count",
	"sequence.significant_rate":        "significant_candidates / frozen_candidates",
	"sequence.replication_rate":        "replicated_candidates / significant_candidates",
	"relation.significant_rate":        "significant_relations / tested_relations",
	"relation.replication_rate":        "replicated_relations / significant_relations",
	"transition.preferred_rate":        "preferred_significant / significant_preferred",
	"transition.depleted_rate":         "depleted_significant / significant_depleted",
	"transition.backbone_retention":    "strict_backbone / significant_backbone",
	"transition.outgoing_profile_rate": "replicated_outgoing_profiles / eligible_tokens",
	"transition.incoming_profile_rate": "replicated_incoming_profiles / eligible_tokens",
	"higher_order.replication_rate":    "higher_order_replicated / candidate_count",
}

func ratio(a, b float64) (float64, string) {
	if b == 0 {
		return 0, "NOT_COMPUTED"
	}
	return a / b, ""
}
func put(e *Experiment, k string, v float64) { e.Raw[k] = v }
func norm(e *Experiment, k string, v, d float64) {
	if d == 0 {
		e.Status[k] = "NOT_COMPUTED"
		return
	}
	e.Normalized[k] = v / d
}

func shaFile(path string) (string, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	h := sha256.Sum256(b)
	return hex.EncodeToString(h[:]), nil
}
func verify(dir string) (string, error) {
	marker := filepath.Join(dir, "FROZEN")
	if _, err := os.Stat(marker); err != nil {
		return "", fmt.Errorf("experiment %s is not frozen (FROZEN marker missing)", dir)
	}
	b, err := os.Open(filepath.Join(dir, "checksums.sha256"))
	if err != nil {
		return "", fmt.Errorf("%s: checksums.sha256: %w", dir, err)
	}
	defer b.Close()
	h := sha256.New()
	s := bufio.NewScanner(b)
	n := 0
	for s.Scan() {
		line := strings.TrimSpace(s.Text())
		if line == "" {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) != 2 {
			return "", fmt.Errorf("invalid checksum line in %s", dir)
		}
		expected := fields[0]
		rel := fields[1]
		p := filepath.Join(dir, "outputs", rel)
		got, er := shaFile(p)
		if er != nil {
			return "", fmt.Errorf("checksum artifact %s: %w", rel, er)
		}
		if got != expected {
			return "", fmt.Errorf("checksum mismatch for %s", rel)
		}
		h.Write([]byte(line))
		h.Write([]byte("\n"))
		n++
	}
	if err := s.Err(); err != nil {
		return "", err
	}
	if n == 0 {
		return "", errors.New("checksums.sha256 is empty")
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}
func readMap(path string) (map[string]any, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var v map[string]any
	if err = yaml.Unmarshal(b, &v); err != nil {
		return nil, err
	}
	return v, nil
}
func number(v any) float64 {
	switch x := v.(type) {
	case int:
		return float64(x)
	case int64:
		return float64(x)
	case uint64:
		return float64(x)
	case float64:
		return x
	case string:
		f, _ := strconv.ParseFloat(x, 64)
		return f
	}
	return 0
}
func tsv(path string) ([]map[string]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	sc := bufio.NewScanner(f)
	sc.Buffer(make([]byte, 4096), 16<<20)
	if !sc.Scan() {
		return nil, sc.Err()
	}
	hdr := strings.Split(sc.Text(), "\t")
	var out []map[string]string
	for sc.Scan() {
		c := strings.Split(sc.Text(), "\t")
		m := map[string]string{}
		for i, k := range hdr {
			if i < len(c) {
				m[k] = c[i]
			}
		}
		out = append(out, m)
	}
	return out, sc.Err()
}
func colCount(rows []map[string]string, col string, pred func(string) bool) float64 {
	n := 0
	for _, r := range rows {
		if pred(r[col]) {
			n++
		}
	}
	return float64(n)
}
func loadExperiment(dir string, allow bool) (Experiment, error) {
	var e Experiment
	e.Path = dir
	e.Raw = map[string]float64{}
	e.Normalized = map[string]float64{}
	e.Status = map[string]string{}
	if !allow {
		checksumDigest, err := verify(dir)
		if err != nil {
			return e, err
		}
		e.ArtifactSHA256 = checksumDigest
	} else if _, err := os.Stat(filepath.Join(dir, "FROZEN")); err != nil {
		e.Status["warning"] = "ALLOW_UNFROZEN"
	}
	mb, err := os.ReadFile(filepath.Join(dir, "manifest.json"))
	if err != nil {
		return e, err
	}
	var m manifest
	if err = json.Unmarshal(mb, &m); err != nil {
		return e, err
	}
	e.ManifestSHA256 = shaBytes(mb)
	e.ID = m.ExperimentID
	e.GitCommit = m.GitCommit
	e.InputMode = m.InputMode
	e.CorpusPath = m.CorpusPath
	e.CorpusSHA256 = m.CorpusSHA256
	if e.InputMode == "" {
		e.InputMode = "unknown"
	}
	for _, s := range m.Stages {
		if strings.EqualFold(s.Status, "NOT_APPLICABLE") {
			e.Unavailable = append(e.Unavailable, s.Name)
		} else {
			e.Applicable = append(e.Applicable, s.Name)
		}
	}
	if rsb, er := os.ReadFile(filepath.Join(dir, "run-state.json")); er == nil {
		var rs runState
		if json.Unmarshal(rsb, &rs) == nil {
			e.Applicable, e.Unavailable = nil, nil
			for _, s := range rs.Stages {
				if strings.EqualFold(s.Status, "NOT_APPLICABLE") {
					e.Unavailable = append(e.Unavailable, s.Name)
				} else {
					e.Applicable = append(e.Applicable, s.Name)
				}
			}
			sort.Strings(e.Applicable)
			sort.Strings(e.Unavailable)
		}
	}
	sort.Strings(e.Applicable)
	sort.Strings(e.Unavailable)
	root := filepath.Join(dir, "workspace", "workdir")
	if _, er := os.Stat(filepath.Join(dir, "outputs")); er == nil {
		root = filepath.Join(dir, "outputs")
	}
	if sm, er := readMap(filepath.Join(root, "transition-network", "transition_network_summary.yaml")); er == nil {
		e.TokenCount = int64(number(sm["token_count"]))
		e.VocabularySize = int64(number(sm["unique_tokens"]))
		for _, k := range []string{"eligible_tokens", "testable_edges", "fdr_significant_preferred", "fdr_significant_depleted", "backbone_preferred", "backbone_depleted", "replicated_outgoing_profiles", "replicated_incoming_profiles", "mean_m0_minus_m1_log_loss", "mean_m1_minus_m2_log_loss"} {
			if v, ok := sm[k]; ok {
				put(&e, "transition."+k, number(v))
			}
		}
	}
	if e.TokenCount == 0 {
		if sm, er := readMap(filepath.Join(root, "sequence_analysis.yaml")); er == nil {
			e.TokenCount = int64(number(sm["token_occurrences"]))
		}
	}
	if e.VocabularySize == 0 {
		if rows, er := tsv(filepath.Join(root, "token-relation-validation", "relation_classification.tsv")); er == nil {
			_ = rows
		}
	}
	if rows, er := tsv(filepath.Join(root, "replicated-local-structure", "replicated_local_structure_summary.tsv")); er == nil {
		for _, r := range rows {
			if r["family"] == "sequence" {
				put(&e, "sequence.frozen_candidates", parse(r["frozen_candidates"]))
				put(&e, "sequence.significant_candidates", parse(r["fdr_significant"]))
				put(&e, "sequence.replicated_candidates", parse(r["robust_cross_block"]))
			}
		}
	}
	if rows, er := tsv(filepath.Join(root, "token-relation-validation", "relation_classification.tsv")); er == nil {
		put(&e, "relation.tested_relations", colCount(rows, "classification", func(s string) bool { return s != "" }))
		put(&e, "relation.significant_relations", colCount(rows, "classification", func(s string) bool { return s == "GROUP_CONSISTENT" || s == "UNIVERSAL" }))
		put(&e, "relation.replicated_relations", colCount(rows, "classification", func(s string) bool { return s == "GROUP_CONSISTENT" || s == "UNIVERSAL" }))
	}
	if rows, er := tsv(filepath.Join(root, "higher-order-sequences", "higher_order_validation.tsv")); er == nil {
		put(&e, "higher_order.candidate_count", float64(len(rows)))
		put(&e, "higher_order.higher_order_replicated", colCount(rows, "final_status", func(s string) bool { return s == "HIGHER_ORDER_REPLICATED" }))
	}
	for k, v := range e.Raw {
		switch k {
		case "transition.eligible_tokens":
			norm(&e, "corpus.eligible_token_rate", v, float64(e.TokenCount))
		case "sequence.significant_candidates":
			norm(&e, "sequence.significant_rate", v, e.Raw["sequence.frozen_candidates"])
		case "sequence.replicated_candidates":
			norm(&e, "sequence.replication_rate", v, e.Raw["sequence.significant_candidates"])
		case "relation.significant_relations":
			norm(&e, "relation.significant_rate", v, e.Raw["relation.tested_relations"])
		case "relation.replicated_relations":
			norm(&e, "relation.replication_rate", v, e.Raw["relation.significant_relations"])
		case "transition.backbone_preferred":
			norm(&e, "transition.preferred_rate", v, e.Raw["transition.fdr_significant_preferred"])
		case "transition.backbone_depleted":
			norm(&e, "transition.depleted_rate", v, e.Raw["transition.fdr_significant_depleted"])
		case "transition.replicated_outgoing_profiles":
			norm(&e, "transition.outgoing_profile_rate", v, e.Raw["transition.eligible_tokens"])
		case "transition.replicated_incoming_profiles":
			norm(&e, "transition.incoming_profile_rate", v, e.Raw["transition.eligible_tokens"])
		case "higher_order.higher_order_replicated":
			norm(&e, "higher_order.replication_rate", v, e.Raw["higher_order.candidate_count"])
		}
	}
	return e, nil
}
func shaBytes(b []byte) string { h := sha256.Sum256(b); return hex.EncodeToString(h[:]) }
func parse(s string) float64   { v, _ := strconv.ParseFloat(s, 64); return v }

func writeTSV(path string, header []string, rows [][]string) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	fmt.Fprintln(f, strings.Join(header, "\t"))
	for _, r := range rows {
		fmt.Fprintln(f, strings.Join(r, "\t"))
	}
	return nil
}
func val(e Experiment, k string) (string, string) {
	if s, ok := e.Status[k]; ok {
		return "", s
	}
	if v, ok := e.Normalized[k]; ok {
		return strconv.FormatFloat(v, 'g', 17, 64), ""
	}
	if v, ok := e.Raw[k]; ok {
		return strconv.FormatFloat(v, 'g', 17, 64), ""
	}
	return "", "MISSING_ARTIFACT"
}

func Run(o Options) error {
	if len(o.Experiments) < 2 {
		return errors.New("at least two -experiment values are required")
	}
	if o.OutputDir == "" {
		return errors.New("-output-dir is required")
	}
	es := make([]Experiment, 0, len(o.Experiments))
	for _, p := range o.Experiments {
		e, err := loadExperiment(p, o.AllowUnfrozen)
		if err != nil {
			return err
		}
		es = append(es, e)
	}
	sort.Slice(es, func(i, j int) bool { return es[i].ID < es[j].ID })
	if err := os.MkdirAll(o.OutputDir, 0755); err != nil {
		return err
	}
	features := make([]string, 0, len(formulas))
	for k := range formulas {
		features = append(features, k)
	}
	sort.Strings(features)
	rawKeys := map[string]bool{}
	for _, e := range es {
		for k := range e.Raw {
			rawKeys[k] = true
		}
	}
	raw := make([]string, 0, len(rawKeys))
	for k := range rawKeys {
		raw = append(raw, k)
	}
	sort.Strings(raw)
	rows := [][]string{}
	for _, e := range es {
		for _, k := range raw {
			v, s := val(e, k)
			rows = append(rows, []string{e.ID, k, v, s})
		}
	}
	if err := writeTSV(filepath.Join(o.OutputDir, "raw_metrics.tsv"), []string{"experiment_id", "metric", "value", "status"}, rows); err != nil {
		return err
	}
	rows = nil
	for _, e := range es {
		for _, k := range features {
			v, s := val(e, k)
			rows = append(rows, []string{e.ID, k, v, s})
		}
	}
	if err := writeTSV(filepath.Join(o.OutputDir, "normalized_metrics.tsv"), []string{"experiment_id", "metric", "value", "status"}, rows); err != nil {
		return err
	}
	if err := writeTSV(filepath.Join(o.OutputDir, "structural_fingerprint.tsv"), []string{"experiment_id", "metric", "value", "status"}, rows); err != nil {
		return err
	}
	yamlFingerprints := make([]map[string]any, 0, len(es))
	for _, e := range es {
		values, statuses := map[string]any{}, map[string]string{}
		for _, k := range features {
			if v, ok := e.Normalized[k]; ok {
				values[k] = v
			} else if s, ok := e.Status[k]; ok {
				statuses[k] = s
			} else {
				statuses[k] = "MISSING_ARTIFACT"
			}
		}
		yamlFingerprints = append(yamlFingerprints, map[string]any{"experiment_id": e.ID, "values": values, "statuses": statuses})
	}
	yamlRows := map[string]any{"fingerprint_schema_version": FingerprintSchemaVersion, "formulas_version": FormulasVersion, "features": features, "formulas": formulas, "experiments": yamlFingerprints}
	yb, _ := yaml.Marshal(yamlRows)
	if err := os.WriteFile(filepath.Join(o.OutputDir, "structural_fingerprint.yaml"), yb, 0644); err != nil {
		return err
	}
	prow := [][]string{}
	comparisonRows := [][]string{}
	for i := 0; i < len(es); i++ {
		for j := i + 1; j < len(es); j++ {
			for _, metric := range []string{"standardized_euclidean", "manhattan", "cosine"} {
				a, b := []float64{}, []float64{}
				scales := []float64{}
				used := 0
				for _, k := range features {
					av, ao := es[i].Normalized[k]
					bv, bo := es[j].Normalized[k]
					if ao && bo {
						a = append(a, av)
						b = append(b, bv)
						mean := 0.0
						count := 0
						for _, e := range es {
							if v, ok := e.Normalized[k]; ok {
								mean += v
								count++
							}
						}
						if count > 0 {
							mean /= float64(count)
						}
						variance := 0.0
						for _, e := range es {
							if v, ok := e.Normalized[k]; ok {
								variance += (v - mean) * (v - mean)
							}
						}
						if count > 1 {
							variance /= float64(count)
						} else {
							variance = 1
						}
						scales = append(scales, math.Sqrt(variance))
						used++
					}
				}
				if used == 0 {
					continue
				}
				d := distance(metric, a, b, scales)
				prow = append(prow, []string{es[i].ID, es[j].ID, metric, strconv.FormatFloat(d, 'g', 17, 64), strconv.Itoa(used), strconv.Itoa(len(features)), strconv.Itoa(len(features) - used)})
			}
			for _, k := range features {
				av, ao := es[i].Normalized[k]
				bv, bo := es[j].Normalized[k]
				if !ao || !bo {
					comparisonRows = append(comparisonRows, []string{k, es[i].ID, es[j].ID, "", "", "MISSING"})
					continue
				}
				mean := 0.
				n := 0
				for _, e := range es {
					if v, ok := e.Normalized[k]; ok {
						mean += v
						n++
					}
				}
				if n > 0 {
					mean /= float64(n)
				}
				variance := 0.
				for _, e := range es {
					if v, ok := e.Normalized[k]; ok {
						variance += (v - mean) * (v - mean)
					}
				}
				if n > 1 {
					variance /= float64(n)
				}
				sd := math.Sqrt(variance)
				z := ""
				if sd > 0 {
					z = strconv.FormatFloat((av-bv)/sd, 'g', 17, 64)
				}
				comparisonRows = append(comparisonRows, []string{k, es[i].ID, es[j].ID, strconv.FormatFloat(math.Abs(av-bv), 'g', 17, 64), z, ""})
			}
		}
	}
	if err := writeTSV(filepath.Join(o.OutputDir, "pairwise_distances.tsv"), []string{"experiment_a", "experiment_b", "metric", "distance", "dimensions_used", "dimensions_total", "dimensions_missing"}, prow); err != nil {
		return err
	}
	if err := writeTSV(filepath.Join(o.OutputDir, "pairwise_comparisons.tsv"), []string{"metric", "experiment_a", "experiment_b", "absolute_difference", "standardized_difference", "status"}, comparisonRows); err != nil {
		return err
	}
	comp := comparisonManifest{SchemaVersion: FingerprintSchemaVersion, FormulasVersion: FormulasVersion, GitCommit: o.GitCommit, Args: o.Args, ManifestHashes: map[string]string{}, ArtifactHashes: map[string]string{}}
	for _, e := range es {
		if warning := e.Status["warning"]; warning != "" {
			comp.Warnings = append(comp.Warnings, e.ID+": "+warning)
		}
	}
	for _, e := range es {
		comp.Experiments = append(comp.Experiments, e.ID)
		comp.ManifestHashes[e.ID] = e.ManifestSHA256
		comp.ArtifactHashes[e.ID] = e.ArtifactSHA256
	}
	sort.Strings(comp.Experiments)
	jb, _ := json.MarshalIndent(comp, "", "  ")
	if err := os.WriteFile(filepath.Join(o.OutputDir, "comparison_manifest.json"), append(jb, '\n'), 0644); err != nil {
		return err
	}
	return writeReport(o.OutputDir, es, features, prow)
}
func distance(kind string, a, b, scales []float64) float64 {
	switch kind {
	case "manhattan":
		s := 0.
		for i := range a {
			s += math.Abs(a[i] - b[i])
		}
		return s
	case "cosine":
		dot, na, nb := 0., 0., 0.
		for i := range a {
			dot += a[i] * b[i]
			na += a[i] * a[i]
			nb += b[i] * b[i]
		}
		if na == 0 || nb == 0 {
			return 1
		}
		return 1 - dot/math.Sqrt(na*nb)
	default:
		s := 0.
		for i := range a {
			sd := scales[i]
			if sd == 0 {
				sd = 1
			}
			s += math.Pow((a[i]-b[i])/sd, 2)
		}
		return math.Sqrt(s)
	}
}
func writeReport(dir string, es []Experiment, features []string, dist [][]string) error {
	var b strings.Builder
	b.WriteString("# Comparative experiment report\n\n")
	b.WriteString("This report is descriptive structural comparison only; no document-class or semantic conclusion is computed.\n\n## Experiment inventory\n\n| experiment | input mode | corpus SHA256 | tokens | vocabulary |\n|---|---|---|---:|---:|\n")
	for _, e := range es {
		fmt.Fprintf(&b, "| %s | %s | `%s` | %d | %d |\n", e.ID, e.InputMode, e.CorpusSHA256, e.TokenCount, e.VocabularySize)
	}
	b.WriteString("\n## Applicability and missingness\n\n")
	for _, e := range es {
		fmt.Fprintf(&b, "- `%s`: applicable stages=%s; unavailable stages=%s.\n", e.ID, strings.Join(e.Applicable, ", "), strings.Join(e.Unavailable, ", "))
	}
	b.WriteString("\nThe machine-readable tables retain explicit status values. Missing artifacts remain missing (never numeric zero); pairwise distances use only dimensions present in both experiments and report used/missing counts.\n\n## Raw metrics and normalized structural fingerprint\n\nSee `raw_metrics.tsv`, `normalized_metrics.tsv`, and `structural_fingerprint.yaml`. The schema and formula registry are fixed at version 1.\n\n## Distances and pairwise comparisons\n\nSee `pairwise_distances.tsv` and `pairwise_comparisons.tsv`. With N=4, any distance or comparison is exploratory/illustrative and is not a class assignment.\n\n## PCA and clustering\n\nOmitted for this small sample; no attempt is made to force a projection or cluster solution.\n\n## Methodological limitations\n\nRaw counts depend on corpus size and are retained for audit. Normalized rates use the formulas documented in `experiment-compare/METRIC_COMPATIBILITY.md`. Categorical labels from Voynich-specific metadata and generic GROUP labels are not compared. No semantic interpretation, classifier, or post-hoc feature selection is performed.\n")
	return os.WriteFile(filepath.Join(dir, "COMPARATIVE_REPORT.md"), []byte(b.String()), 0644)
}
