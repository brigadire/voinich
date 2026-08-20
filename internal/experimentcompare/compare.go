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
	// Version 2 is the first Task52 schema. The former Task45 v1 output was a
	// prototype and is intentionally not treated as an immutable contract.
	FingerprintSchemaVersion = 2
	FormulasVersion          = 2
)

type Experiment struct {
	ID, Path, InputMode, CorpusPath, CorpusSHA256, GitCommit string
	TokenCount, VocabularySize                               int64
	Applicable, Unavailable                                  []string
	Raw, Normalized                                          map[string]float64
	Status                                                   map[string]string
	ManifestSHA256, ArtifactSHA256                           string
	Transformation                                           string
	ManifestGitDirty                                         bool
	Frozen, Verified                                         bool
	VerificationStatus                                       string
	ArtifactHashes                                           map[string]string
	Trajectory                                               map[int]float64
	NullEffects                                              map[int]float64
	SegmentBetas, SegmentRates                               []float64
}
type Options struct {
	Experiments   []string
	OutputDir     string
	AllowUnfrozen bool
	Args          []string
	GitCommit     string
	GitDirty      bool
}
type manifest struct {
	ExperimentID string `json:"experiment_id"`
	GitCommit    string `json:"git_commit"`
	GitDirty     bool   `json:"git_dirty"`
	CorpusPath   string `json:"corpus_path"`
	CorpusSHA256 string `json:"corpus_sha256"`
	InputMode    string `json:"input_mode"`
	IVTFFPath    string `json:"ivtff_path"`
	IVTFFSHA256  string `json:"ivtff_sha256"`
	Corpus       *struct {
		Path   string `json:"path"`
		SHA256 string `json:"sha256"`
	} `json:"corpus"`
	Transformation *struct {
		Type string `json:"type"`
		Mode string `json:"mode"`
	} `json:"transformation"`
	Stages []struct {
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
	SchemaVersion      int                          `json:"fingerprint_schema_version"`
	FormulasVersion    int                          `json:"formulas_version"`
	Experiments        []string                     `json:"experiment_identities"`
	ManifestHashes     map[string]string            `json:"experiment_manifest_sha256"`
	ArtifactHashes     map[string]string            `json:"experiment_artifact_sha256"`
	ArtifactFileHashes map[string]map[string]string `json:"experiment_artifact_file_hashes"`
	GitCommit          string                       `json:"comparison_program_git_commit"`
	GitDirty           bool                         `json:"comparison_program_git_dirty"`
	Args               []string                     `json:"cli_arguments"`
	Warnings           []string                     `json:"warnings,omitempty"`
	ComparisonID       string                       `json:"comparison_id"`
	NormalizationScope string                       `json:"normalization_scope"`
	ExperimentCount    int                          `json:"experiment_count"`
	Verification       map[string]string            `json:"verification_status"`
	Frozen             map[string]bool              `json:"frozen"`
	CommonCoreFeatures []string                     `json:"common_core_features"`
	FeatureDefinitions map[string]map[string]any    `json:"feature_definitions"`
}

var formulas = map[string]string{
	"corpus.eligible_token_rate":              "eligible_tokens / token_count",
	"sequence.significant_rate":               "significant_candidates / frozen_candidates",
	"sequence.replication_rate":               "replicated_candidates / significant_candidates",
	"relation.significant_rate":               "significant_relations / tested_relations",
	"relation.replication_rate":               "replicated_relations / significant_relations",
	"transition.preferred_backbone_retention": "backbone_preferred / fdr_significant_preferred",
	"transition.depleted_backbone_retention":  "backbone_depleted / fdr_significant_depleted",
	"transition.backbone_retention":           "(backbone_preferred + backbone_depleted) / (fdr_significant_preferred + fdr_significant_depleted)",
	"transition.outgoing_profile_rate":        "replicated_outgoing_profiles / eligible_tokens",
	"transition.incoming_profile_rate":        "replicated_incoming_profiles / eligible_tokens",
	"higher_order.replication_rate":           "higher_order_replicated / candidate_count",
	"vocabulary_growth.V_1000_per_token":      "V_1000 / 1000",
	"vocabulary_growth.V_2000_per_token":      "V_2000 / 2000",
	"vocabulary_growth.V_4000_per_token":      "V_4000 / 4000",
	"vocabulary_growth.V_8000_per_token":      "V_8000 / 8000",
	"vocabulary_growth.V_16000_per_token":     "V_16000 / 16000",
	"vocabulary_growth.V_32000_per_token":     "V_32000 / 32000",
}

func addVocabularyFeatures() {
	for _, k := range []string{"total_tokens", "unique_tokens", "heaps_K", "heaps_beta", "heaps_R2", "final_hapax_count", "final_hapax_fraction_of_types", "final_hapax_fraction_of_tokens", "final_dis_legomena_count", "singleton_to_doubleton_ratio", "final_new_type_rate", "order_effect_mean", "order_effect_max", "order_effect_final", "segment_beta_mean", "segment_beta_sd", "segment_new_type_rate_sd"} {
		formulas["vocabulary_growth."+k] = "raw Task49 artifact field"
	}
	for _, n := range []int{1000, 2000, 4000, 8000, 16000, 32000} {
		formulas["vocabulary_growth.V_"+strconv.Itoa(n)] = "V(" + strconv.Itoa(n) + ")"
	}
}

var featureDefinitions = map[string]map[string]any{}

func init() {
	addVocabularyFeatures()
	for k, f := range formulas {
		featureDefinitions[k] = map[string]any{"formula": f, "version": FormulasVersion, "use_for_distance": true}
	}
	for _, k := range []string{"corpus.eligible_token_rate", "sequence.significant_rate", "sequence.replication_rate", "relation.significant_rate", "relation.replication_rate", "transition.preferred_backbone_retention", "transition.depleted_backbone_retention", "transition.outgoing_profile_rate", "transition.incoming_profile_rate", "higher_order.replication_rate"} {
		if d := featureDefinitions[k]; d != nil {
			d["family"] = strings.Split(k, ".")[0]
			d["kind"] = "proportion"
		}
	}
	for _, k := range []string{"vocabulary_growth.total_tokens", "vocabulary_growth.unique_tokens", "vocabulary_growth.final_hapax_count", "vocabulary_growth.final_dis_legomena_count"} {
		if d := featureDefinitions[k]; d != nil {
			d["include_in_fingerprint"] = false
			d["use_for_distance"] = false
		}
	}
	for _, n := range []int{1000, 2000, 4000, 8000, 16000, 32000} {
		if d := featureDefinitions["vocabulary_growth.V_"+strconv.Itoa(n)]; d != nil {
			d["include_in_fingerprint"] = false
			d["use_for_distance"] = false
		}
	}
	for k, d := range featureDefinitions {
		if _, ok := d["family"]; !ok {
			d["family"] = strings.Split(k, ".")[0]
		}
		if _, ok := d["include_in_fingerprint"]; !ok {
			d["include_in_fingerprint"] = true
		}
	}
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
func checksumEntries(dir string) (map[string]string, error) {
	b, err := os.Open(filepath.Join(dir, "checksums.sha256"))
	if err != nil {
		return nil, err
	}
	defer b.Close()
	out := map[string]string{}
	s := bufio.NewScanner(b)
	for s.Scan() {
		f := strings.Fields(strings.TrimSpace(s.Text()))
		if len(f) == 2 {
			out[f[1]] = f[0]
		}
	}
	return out, s.Err()
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
func putIfNumber(e *Experiment, key, s string) {
	if s != "" {
		if v, err := strconv.ParseFloat(s, 64); err == nil {
			put(e, key, v)
		}
	}
}
func loadTask49(e *Experiment, root string) {
	dir := filepath.Join(root, "vocabulary-growth")
	growth, err := tsv(filepath.Join(dir, "vocabulary_growth.tsv"))
	if err != nil {
		return
	}
	e.Trajectory = map[int]float64{}
	for _, r := range growth {
		n := int(parse(r["checkpoint_n"]))
		v := parse(r["vocabulary_size"])
		e.Trajectory[n] = v
		put(e, "vocabulary_growth.V_"+strconv.Itoa(n), v)
	}
	for _, n := range []int{1000, 2000, 4000, 8000, 16000, 32000} {
		if _, ok := e.Trajectory[n]; ok {
			put(e, "vocabulary_growth.V_"+strconv.Itoa(n)+"_per_token", e.Trajectory[n]/float64(n))
		} else {
			e.Status["vocabulary_growth.V_"+strconv.Itoa(n)] = "NOT_APPLICABLE"
			e.Status["vocabulary_growth.V_"+strconv.Itoa(n)+"_per_token"] = "NOT_APPLICABLE"
		}
	}
	if sm, er := readMap(filepath.Join(dir, "summary.yaml")); er == nil {
		if fp, ok := sm["final_profile"].(map[string]any); ok {
			for _, x := range []struct{ src, dst string }{{"total_tokens", "vocabulary_growth.total_tokens"}, {"unique_tokens", "vocabulary_growth.unique_tokens"}, {"heaps_K", "vocabulary_growth.heaps_K"}, {"heaps_beta", "vocabulary_growth.heaps_beta"}, {"heaps_R2", "vocabulary_growth.heaps_R2"}, {"hapax", "vocabulary_growth.final_hapax_count"}, {"hapax_fraction_of_types", "vocabulary_growth.final_hapax_fraction_of_types"}, {"hapax_fraction_of_tokens", "vocabulary_growth.final_hapax_fraction_of_tokens"}, {"dis_legomena", "vocabulary_growth.final_dis_legomena_count"}, {"singleton_to_doubleton_ratio", "vocabulary_growth.singleton_to_doubleton_ratio"}} {
				if v, yes := fp[x.src]; yes {
					put(e, x.dst, number(v))
				}
			}
		}
	}
	if rows, er := tsv(filepath.Join(dir, "new_type_rate.tsv")); er == nil {
		best := ""
		maxStart := -1
		for _, r := range rows {
			if r["window_end"] != "" && parse(r["window_end"]) > float64(maxStart) {
				maxStart = int(parse(r["window_end"]))
				best = r["new_type_rate"]
			}
		}
		putIfNumber(e, "vocabulary_growth.final_new_type_rate", best)
	}
	if rows, er := tsv(filepath.Join(dir, "vocabulary_growth_null.tsv")); er == nil {
		e.NullEffects = map[int]float64{}
		for _, r := range rows {
			n := int(parse(r["checkpoint_n"]))
			eff := parse(r["effect"])
			e.NullEffects[n] = eff
			put(e, "vocabulary_growth.order_effect_"+strconv.Itoa(n), eff)
		}
		if len(e.NullEffects) > 0 {
			sum, max := 0., 0.
			for _, v := range e.NullEffects {
				sum += math.Abs(v)
				if math.Abs(v) > max {
					max = math.Abs(v)
				}
			}
			put(e, "vocabulary_growth.order_effect_mean", sum/float64(len(e.NullEffects)))
			put(e, "vocabulary_growth.order_effect_max", max)
			if v, ok := e.NullEffects[int(e.TokenCount)]; ok {
				put(e, "vocabulary_growth.order_effect_final", v)
			}
		}
	}
	if rows, er := tsv(filepath.Join(dir, "segment_vocabulary_growth.tsv")); er == nil {
		e.SegmentBetas, e.SegmentRates = []float64{}, []float64{}
		for _, r := range rows {
			if r["checkpoint_n"] == r["vocabulary_size"] {
				continue
			}
			if v, er := strconv.ParseFloat(r["heaps_beta"], 64); er == nil && v != 0 {
				e.SegmentBetas = append(e.SegmentBetas, v)
			}
			if v, er := strconv.ParseFloat(r["new_type_rate"], 64); er == nil {
				e.SegmentRates = append(e.SegmentRates, v)
			}
		}
		if len(e.SegmentBetas) > 0 {
			m := meanFloat(e.SegmentBetas)
			sd := stdFloat(e.SegmentBetas)
			put(e, "vocabulary_growth.segment_beta_mean", m)
			put(e, "vocabulary_growth.segment_beta_sd", sd)
		}
		if len(e.SegmentRates) > 0 {
			put(e, "vocabulary_growth.segment_new_type_rate_sd", stdFloat(e.SegmentRates))
		}
	}
}
func promoteTask49(e *Experiment) {
	for _, k := range []string{"vocabulary_growth.heaps_K", "vocabulary_growth.heaps_beta", "vocabulary_growth.heaps_R2", "vocabulary_growth.final_hapax_fraction_of_types", "vocabulary_growth.final_hapax_fraction_of_tokens", "vocabulary_growth.singleton_to_doubleton_ratio", "vocabulary_growth.final_new_type_rate", "vocabulary_growth.order_effect_mean", "vocabulary_growth.order_effect_max", "vocabulary_growth.order_effect_final", "vocabulary_growth.segment_beta_mean", "vocabulary_growth.segment_beta_sd", "vocabulary_growth.segment_new_type_rate_sd"} {
		if v, ok := e.Raw[k]; ok {
			e.Normalized[k] = v
		}
	}
	for _, n := range []int{1000, 2000, 4000, 8000, 16000, 32000} {
		k := "vocabulary_growth.V_" + strconv.Itoa(n) + "_per_token"
		if v, ok := e.Raw[k]; ok {
			e.Normalized[k] = v
		}
	}
}
func meanFloat(xs []float64) float64 {
	if len(xs) == 0 {
		return 0
	}
	s := 0.
	for _, v := range xs {
		s += v
	}
	return s / float64(len(xs))
}
func stdFloat(xs []float64) float64 {
	if len(xs) == 0 {
		return 0
	}
	m := meanFloat(xs)
	s := 0.
	for _, v := range xs {
		s += (v - m) * (v - m)
	}
	return math.Sqrt(s / float64(len(xs)))
}
func loadExperiment(dir string, allow bool) (Experiment, error) {
	var e Experiment
	e.Path = dir
	e.Raw = map[string]float64{}
	e.Normalized = map[string]float64{}
	e.Status = map[string]string{}
	_, markerErr := os.Stat(filepath.Join(dir, "FROZEN"))
	e.Frozen = markerErr == nil
	if !allow || e.Frozen {
		checksumDigest, err := verify(dir)
		if err != nil {
			if !allow {
				return e, err
			}
			e.VerificationStatus = "FAILED"
		} else {
			e.Verified = true
			e.VerificationStatus = "PASS"
			e.ArtifactSHA256 = checksumDigest
		}
	} else {
		e.VerificationStatus = "NOT_RUN"
		e.Status["warning"] = "NON_REPRODUCIBLE_DEVELOPMENT_COMPARISON"
	}
	if hashes, err := checksumEntries(dir); err == nil {
		e.ArtifactHashes = hashes
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
	e.ManifestGitDirty = m.GitDirty
	if m.Transformation != nil {
		e.Transformation = m.Transformation.Type + ":" + m.Transformation.Mode
	}
	if e.InputMode == "" {
		if m.IVTFFPath != "" || m.IVTFFSHA256 != "" {
			e.InputMode = "legacy_ivtff"
		} else if m.Corpus != nil {
			e.InputMode = "generic"
		} else {
			e.InputMode = "legacy_unknown"
		}
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
	loadTask49(&e, root)
	promoteTask49(&e)
	if e.TokenCount == 0 {
		if v, ok := e.Raw["vocabulary_growth.total_tokens"]; ok {
			e.TokenCount = int64(v)
		}
	}
	if e.VocabularySize == 0 {
		if v, ok := e.Raw["vocabulary_growth.unique_tokens"]; ok {
			e.VocabularySize = int64(v)
		}
	}
	if e.TokenCount > 0 {
		if _, ok := e.Raw["transition.eligible_tokens"]; ok {
			norm(&e, "corpus.eligible_token_rate", e.Raw["transition.eligible_tokens"], float64(e.TokenCount))
		}
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
			norm(&e, "transition.preferred_backbone_retention", v, e.Raw["transition.fdr_significant_preferred"])
		case "transition.backbone_depleted":
			norm(&e, "transition.depleted_backbone_retention", v, e.Raw["transition.fdr_significant_depleted"])
		case "transition.replicated_outgoing_profiles":
			norm(&e, "transition.outgoing_profile_rate", v, e.Raw["transition.eligible_tokens"])
		case "transition.replicated_incoming_profiles":
			norm(&e, "transition.incoming_profile_rate", v, e.Raw["transition.eligible_tokens"])
		case "higher_order.higher_order_replicated":
			norm(&e, "higher_order.replication_rate", v, e.Raw["higher_order.candidate_count"])
		}
	}
	if bp, ok := e.Raw["transition.backbone_preferred"]; ok {
		if bd, ok2 := e.Raw["transition.backbone_depleted"]; ok2 {
			if sp, ok3 := e.Raw["transition.fdr_significant_preferred"]; ok3 {
				if sd, ok4 := e.Raw["transition.fdr_significant_depleted"]; ok4 {
					put(&e, "transition.significant_edges", sp+sd)
					put(&e, "transition.strict_backbone", bp+bd)
					norm(&e, "transition.backbone_retention", bp+bd, sp+sd)
				}
			}
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
		return strconv.FormatFloat(v, 'g', 17, 64), "VALUE"
	}
	if v, ok := e.Raw[k]; ok {
		return strconv.FormatFloat(v, 'g', 17, 64), "VALUE"
	}
	return "", "MISSING_ARTIFACT"
}
func fingerprintFeatures() []string {
	out := []string{}
	for k, d := range featureDefinitions {
		if v, ok := d["include_in_fingerprint"].(bool); ok && !v {
			continue
		}
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}
func distanceFeatureNames() []string {
	out := []string{}
	for k, d := range featureDefinitions {
		if v, ok := d["use_for_distance"].(bool); v {
			if v && (!ok || d["include_in_fingerprint"] != false) {
				out = append(out, k)
			}
		}
	}
	sort.Strings(out)
	return out
}
func cohortScales(es []Experiment, keys []string) map[string]float64 {
	out := map[string]float64{}
	for _, k := range keys {
		xs := []float64{}
		for _, e := range es {
			if v, ok := e.Normalized[k]; ok {
				xs = append(xs, v)
			}
		}
		sd := stdFloat(xs)
		if sd == 0 {
			sd = 1
		}
		out[k] = sd
	}
	return out
}
func commonKeys(es []Experiment, keys []string) []string {
	out := []string{}
	for _, k := range keys {
		ok := true
		for _, e := range es {
			if _, yes := e.Normalized[k]; !yes {
				ok = false
				break
			}
		}
		if ok {
			out = append(out, k)
		}
	}
	return out
}
func vectorDistance(kind string, a, b []float64, scales []float64) float64 {
	return distance(kind, a, b, scales)
}
func distanceRow(a, b Experiment, kind string, keys []string, scales map[string]float64) []string {
	av, bv, ss := []float64{}, []float64{}, []float64{}
	for _, k := range keys {
		if x, ok := a.Normalized[k]; ok {
			if y, ok2 := b.Normalized[k]; ok2 {
				av = append(av, x)
				bv = append(bv, y)
				ss = append(ss, scales[k])
			}
		}
	}
	base := "standardized_euclidean"
	if strings.Contains(kind, "cosine") {
		base = "cosine"
	} else if strings.Contains(kind, "manhattan") {
		base = "manhattan"
	}
	d := vectorDistance(base, av, bv, ss)
	return []string{a.ID, b.ID, kind, strconv.FormatFloat(d, 'g', 17, 64), strconv.Itoa(len(av)), strconv.Itoa(len(keys)), strconv.FormatFloat(float64(len(av))/float64(maxInt(1, len(keys))), 'g', 17, 64)}
}
func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}
func task49PairRows(a, b Experiment) [][]string {
	if len(a.Trajectory) == 0 || len(b.Trajectory) == 0 {
		return nil
	}
	common := []int{}
	for n := range a.Trajectory {
		if _, ok := b.Trajectory[n]; ok {
			common = append(common, n)
		}
	}
	sort.Ints(common)
	if len(common) == 0 {
		return nil
	}
	n := common[len(common)-1]
	va, vb := a.Trajectory[n], b.Trajectory[n]
	rows := [][]string{{"vocabulary_growth.common_n", a.ID, b.ID, strconv.Itoa(n), "", "VALUE"}, {"vocabulary_growth.V_at_common_n_A", a.ID, b.ID, fmt.Sprintf("%.17g", va), "", "VALUE"}, {"vocabulary_growth.V_at_common_n_B", a.ID, b.ID, fmt.Sprintf("%.17g", vb), "", "VALUE"}, {"vocabulary_growth.common_n_absolute_difference", a.ID, b.ID, fmt.Sprintf("%.17g", math.Abs(va-vb)), "", "VALUE"}}
	sum := 0.
	for _, x := range common {
		d := (a.Trajectory[x] - b.Trajectory[x]) / float64(x)
		sum += d * d
	}
	rows = append(rows, []string{"vocabulary_growth.trajectory_rms_normalized", a.ID, b.ID, fmt.Sprintf("%.17g", math.Sqrt(sum/float64(len(common)))), "", "VALUE"})
	return rows
}

func Run(o Options) error { return runV2(o) }

func runV2(o Options) error {
	if len(o.Experiments) < 2 {
		return errors.New("at least two -experiment values are required")
	}
	if o.OutputDir == "" {
		return errors.New("-output-dir is required")
	}
	if !o.AllowUnfrozen && (o.GitCommit == "" || o.GitCommit == "unknown") {
		return errors.New("production comparison requires a resolved comparison program git commit")
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
	features := fingerprintFeatures()
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
	support := [][]string{}
	for _, e := range es {
		for _, k := range []string{"transition.eligible_tokens", "transition.testable_edges", "transition.fdr_significant_preferred", "transition.fdr_significant_depleted", "sequence.frozen_candidates", "sequence.significant_candidates", "relation.tested_relations", "relation.significant_relations", "higher_order.candidate_count", "vocabulary_growth.total_tokens"} {
			if v, ok := e.Raw[k]; ok {
				support = append(support, []string{e.ID, k, fmt.Sprintf("%.17g", v), "VALUE"})
			}
		}
	}
	if err := writeTSV(filepath.Join(o.OutputDir, "metric_support.tsv"), []string{"experiment_id", "support_metric", "support_value", "status"}, support); err != nil {
		return err
	}
	derived := [][]string{}
	for _, e := range es {
		for k, f := range formulas {
			if _, ok := e.Normalized[k]; ok {
				derived = append(derived, []string{e.ID, k, f, "", "", strconv.FormatFloat(e.Normalized[k], 'g', 17, 64), "VALUE"})
			} else if s := e.Status[k]; s != "" {
				derived = append(derived, []string{e.ID, k, f, "", "", "", s})
			}
		}
	}
	if err := writeTSV(filepath.Join(o.OutputDir, "derived_metrics.tsv"), []string{"experiment_id", "metric", "formula", "numerator", "denominator", "value", "status"}, derived); err != nil {
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
	yamlFingerprints := []map[string]any{}
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
	yb, _ := yaml.Marshal(map[string]any{"fingerprint_schema_version": FingerprintSchemaVersion, "formulas_version": FormulasVersion, "features": features, "formulas": formulas, "feature_definitions": featureDefinitions, "experiments": yamlFingerprints})
	if err := os.WriteFile(filepath.Join(o.OutputDir, "structural_fingerprint.yaml"), yb, 0644); err != nil {
		return err
	}
	available := distanceFeatureNames()
	scales, common := cohortScales(es, available), commonKeys(es, available)
	commonScales := cohortScales(es, common)
	prow, comparisonRows := [][]string{}, [][]string{}
	for i := 0; i < len(es); i++ {
		for j := i + 1; j < len(es); j++ {
			for _, kind := range []string{"standardized_euclidean", "manhattan", "cosine"} {
				if len(available) > 0 {
					prow = append(prow, distanceRow(es[i], es[j], "pairwise_available_"+kind, available, scales))
				}
			}
			for _, kind := range []string{"standardized_euclidean_common_core", "cosine_common_core"} {
				if len(common) > 0 {
					row := distanceRow(es[i], es[j], kind, common, commonScales)
					row[5] = strconv.Itoa(len(common))
					row[6] = "1"
					prow = append(prow, row)
				}
			}
			comparisonRows = append(comparisonRows, task49PairRows(es[i], es[j])...)
			for _, k := range features {
				av, ao := es[i].Normalized[k]
				bv, bo := es[j].Normalized[k]
				if !ao || !bo {
					comparisonRows = append(comparisonRows, []string{k, es[i].ID, es[j].ID, "", "", "MISSING_ARTIFACT"})
					continue
				}
				sd := cohortScales(es, []string{k})[k]
				z := ""
				if sd > 0 {
					z = fmt.Sprintf("%.17g", (av-bv)/sd)
				}
				comparisonRows = append(comparisonRows, []string{k, es[i].ID, es[j].ID, fmt.Sprintf("%.17g", math.Abs(av-bv)), z, "VALUE"})
			}
		}
	}
	if err := writeTSV(filepath.Join(o.OutputDir, "pairwise_distances.tsv"), []string{"experiment_a", "experiment_b", "metric", "distance", "dimensions_used", "dimensions_total", "coverage_fraction"}, prow); err != nil {
		return err
	}
	if err := writeTSV(filepath.Join(o.OutputDir, "pairwise_comparisons.tsv"), []string{"metric", "experiment_a", "experiment_b", "absolute_difference", "standardized_difference", "status"}, comparisonRows); err != nil {
		return err
	}
	familyRows := [][]string{}
	families := map[string][]string{}
	for _, k := range available {
		family := strings.Split(k, ".")[0]
		families[family] = append(families[family], k)
	}
	for family, keys := range families {
		ck := commonKeys(es, keys)
		if len(ck) == 0 {
			continue
		}
		for i := 0; i < len(es); i++ {
			for j := i + 1; j < len(es); j++ {
				familyRows = append(familyRows, distanceRow(es[i], es[j], family+"_common_core_standardized_euclidean", ck, cohortScales(es, ck)))
			}
		}
	}
	if err := writeTSV(filepath.Join(o.OutputDir, "family_distances.tsv"), []string{"experiment_a", "experiment_b", "metric", "distance", "dimensions_used", "dimensions_total", "coverage_fraction"}, familyRows); err != nil {
		return err
	}
	verification, frozen, artifacts := map[string]string{}, map[string]bool{}, map[string]map[string]string{}
	for _, e := range es {
		verification[e.ID] = e.VerificationStatus
		frozen[e.ID] = e.Frozen
		artifacts[e.ID] = e.ArtifactHashes
	}
	identityPayload := struct {
		Schema    int
		IDs       []string
		Manifests map[string]string
		Artifacts map[string]map[string]string
		Features  []string
		Commit    string
	}{FingerprintSchemaVersion, []string{}, map[string]string{}, artifacts, features, o.GitCommit}
	for _, e := range es {
		identityPayload.IDs = append(identityPayload.IDs, e.ID)
		identityPayload.Manifests[e.ID] = e.ManifestSHA256
	}
	sort.Strings(identityPayload.IDs)
	ib, _ := json.Marshal(identityPayload)
	comparisonID := shaBytes(ib)
	comp := comparisonManifest{SchemaVersion: FingerprintSchemaVersion, FormulasVersion: FormulasVersion, GitCommit: o.GitCommit, GitDirty: o.GitDirty, Args: o.Args, ManifestHashes: map[string]string{}, ArtifactHashes: map[string]string{}, ArtifactFileHashes: artifacts, ComparisonID: comparisonID, NormalizationScope: "current comparison cohort", ExperimentCount: len(es), Verification: verification, Frozen: frozen, CommonCoreFeatures: common, FeatureDefinitions: featureDefinitions}
	for _, e := range es {
		comp.Experiments = append(comp.Experiments, e.ID)
		comp.ManifestHashes[e.ID] = e.ManifestSHA256
		comp.ArtifactHashes[e.ID] = e.ArtifactSHA256
		if e.Status["warning"] != "" {
			comp.Warnings = append(comp.Warnings, e.ID+": "+e.Status["warning"])
		}
	}
	sort.Strings(comp.Experiments)
	if len(es) < 5 {
		comp.Warnings = append(comp.Warnings, fmt.Sprintf("small comparison cohort: N=%d; cohort standardization is exploratory", len(es)))
	}
	jb, _ := json.MarshalIndent(comp, "", "  ")
	if err := os.WriteFile(filepath.Join(o.OutputDir, "comparison_manifest.json"), append(jb, '\n'), 0644); err != nil {
		return err
	}
	return writeReportV2(o.OutputDir, es, features, common, prow, artifacts, verification)
}

func writeReportV2(dir string, es []Experiment, features, common []string, dist [][]string, artifacts map[string]map[string]string, verification map[string]string) error {
	var b strings.Builder
	b.WriteString("# Comparative experiment report\n\nSimilarity is not classification, causal explanation, or proof that the nearest corpus has the same document type. No semantic interpretation is computed.\n\n## Experiment inventory and provenance\n\n")
	for _, e := range es {
		fmt.Fprintf(&b, "- `%s`: mode=%s, frozen=%t, verification=%s, corpus=`%s`, tokens=%d, vocabulary=%d, manifest_sha256=`%s`, transformation=%s.\n", e.ID, e.InputMode, e.Frozen, e.VerificationStatus, e.CorpusSHA256, e.TokenCount, e.VocabularySize, e.ManifestSHA256, nonempty(e.Transformation, "none"))
	}
	b.WriteString("\n## Fingerprint families\n\nFeatures are grouped by their public semantic prefix: lexical/corpus, sequence, relation, higher_order, transition, and vocabulary_growth. Raw metrics and support values are retained separately; raw counts and support-only fields are excluded from distance unless explicitly defined.\n\n## Common core and missingness\n\n")
	fmt.Fprintf(&b, "- comparison schema: v%d\n- experiments: %d\n- common-core dimensions: %d\n- common-core features: `%s`\n- normalization scope: current comparison cohort; standardized distances are exploratory for small N.\n\n", FingerprintSchemaVersion, len(es), len(common), strings.Join(common, ", "))
	b.WriteString("Statuses distinguish VALUE, COMPLETED_EMPTY, NOT_APPLICABLE, NOT_COMPUTED, MISSING_ARTIFACT, and FAILED/INVALID.\n\n## Vocabulary Growth Comparison\n\n")
	has := 0
	for _, e := range es {
		if len(e.Trajectory) > 0 {
			has++
		}
	}
	if has < 2 {
		b.WriteString("Task49 vocabulary-growth artifacts are MISSING_ARTIFACT for enough experiments to form a trajectory comparison. Legacy experiments remain valid.\n\n")
	} else {
		for _, e := range es {
			fmt.Fprintf(&b, "- `%s`: N=%d, V=%d, Heaps beta=%s, R²=%s, hapax fraction=%s.\n", e.ID, e.TokenCount, e.VocabularySize, metricText(e, "vocabulary_growth.heaps_beta"), metricText(e, "vocabulary_growth.heaps_R2"), metricText(e, "vocabulary_growth.final_hapax_fraction_of_types"))
		}
		b.WriteString("\nCommon-N and trajectory rows are in `pairwise_comparisons.tsv`; no corpus is physically truncated.\n\n")
	}
	b.WriteString("## Distances\n\n`pairwise_available_*` distances use only valid dimensions for each pair and report coverage. `*_common_core` distances use the same dimensions for every experiment. `family_distances.tsv` reports family-level common-core distances. All features have equal weight; no weights were fitted to these corpora.\n\n## Support and limitations\n\nSupport denominators remain in `raw_metrics.tsv` and `metric_support.tsv` where available. Missing Task49 is not zero. Categorical metadata labels are excluded as incompatible semantics. Transformation provenance is audit metadata only and never a feature.\n\nSimilarity != classification. Similarity != causal explanation. Nearest corpus != same document type.\n")
	return os.WriteFile(filepath.Join(dir, "COMPARATIVE_REPORT.md"), []byte(b.String()), 0644)
}
func metricText(e Experiment, k string) string {
	if v, ok := e.Raw[k]; ok {
		return fmt.Sprintf("%.8g", v)
	}
	return e.Status[k]
}
func nonempty(v, fallback string) string {
	if v == "" {
		return fallback
	}
	return v
}

// legacyRunTail is retained only as source-compatible historical code while
// the Task52 path above is the sole production path. It is not invoked.
func legacyRunTail(o Options, es []Experiment, features []string, rows [][]string) error {
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
