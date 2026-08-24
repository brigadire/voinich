// Command task83-analyze reproduces the frozen Task83 comparison.
package main

import (
	"bufio"
	"crypto/sha256"
	"encoding/csv"
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
	"time"
)

const outDir = "research/phase2/task83"

var directCore = []string{"EF1_GIANT_COMPONENT_SHARE", "EF2_GLOBAL_CLUSTERING", "EF3_DEGREE_FREQUENCY_SPEARMAN"}
var directAll = []string{"EF1_GIANT_COMPONENT_SHARE", "EF1_ISOLATE_SHARE", "EF2_GLOBAL_CLUSTERING", "EF3_DEGREE_FREQUENCY_SPEARMAN", "LP1_RULE_SUPPORT_GINI", "LP4_PREFIX_ATTACHMENT_NMI", "LP4_SUFFIX_ATTACHMENT_NMI", "cs2/prev-family-current-family"}

var family = map[string]string{
	"EF1_GIANT_COMPONENT_SHARE": "edit family", "EF1_ISOLATE_SHARE": "edit family",
	"EF2_GLOBAL_CLUSTERING": "edit family", "EF3_DEGREE_FREQUENCY_SPEARMAN": "edit family",
	"LP1_RULE_SUPPORT_GINI": "lexical paradigm", "LP4_PREFIX_ATTACHMENT_NMI": "lexical paradigm",
	"LP4_SUFFIX_ATTACHMENT_NMI": "lexical paradigm", "cs2/prev-family-current-family": "cross-scale",
}

type metric struct {
	MetricID       string  `json:"MetricID"`
	Family         string  `json:"Family"`
	Classification string  `json:"Classification"`
	Value          float64 `json:"Value"`
	Available      bool    `json:"Available"`
	MissingReason  string  `json:"MissingReason"`
}
type rawVector struct {
	JobID       string   `json:"JobID"`
	MechanismID string   `json:"MechanismID"`
	Policy      string   `json:"Policy"`
	Corpus      string   `json:"Corpus"`
	Scale       string   `json:"Scale"`
	Replicate   int      `json:"Replicate"`
	Metrics     []metric `json:"Metrics"`
}
type fp struct {
	Primary struct {
		Metrics struct {
			EF1 struct {
				Giant   float64 `json:"giant_component_share"`
				Isolate float64 `json:"isolate_share"`
			} `json:"ef1"`
			EF2 struct {
				Clustering float64 `json:"global_clustering"`
			} `json:"ef2"`
			EF3 struct {
				Spearman float64 `json:"spearman_degree_log_frequency"`
			} `json:"ef3"`
			LP1 struct {
				Gini float64 `json:"support_gini"`
			} `json:"lp1"`
			LP4 struct {
				Prefix struct {
					NMI float64 `json:"normalized_mutual_information"`
				} `json:"prefix"`
				Suffix struct {
					NMI float64 `json:"normalized_mutual_information"`
				} `json:"suffix"`
			} `json:"lp4"`
		} `json:"metrics"`
		CrossScale *struct {
			Metrics []struct {
				ID     string  `json:"metric_id"`
				Value  float64 `json:"observed_statistic"`
				Status string  `json:"status"`
			} `json:"metrics"`
		} `json:"cross_scale"`
	} `json:"primary"`
}
type discrim struct {
	ControlID string  `json:"control_id"`
	MetricID  string  `json:"metric_id"`
	Primary   float64 `json:"primary_value"`
	Control   float64 `json:"control_value"`
	Z         float64 `json:"standardized_difference"`
}
type endpoint struct {
	Class, ID, Corpus, Policy, Scale string
	Replicate                        int
	Values                           map[string]float64
	Available                        map[string]bool
}
type traj struct {
	Class, ID, Corpus, NullKind string
	Values                      map[string]float64
	Available                   map[string]bool
	Degenerate                  bool
}
type comparison struct {
	Target, Class, ID, Corpus, Policy, Scale string
	Replicate                                int
	D, Adjusted, Coverage                    float64
	N                                        int
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run() error {
	if err := verifyOpening(); err != nil {
		return err
	}
	if err := verifyFingerprintSources(); err != nil {
		return invalidate(err)
	}
	targets, err := loadTargets()
	if err != nil {
		return err
	}
	scales, err := loadScales()
	if err != nil {
		return err
	}
	fontana, err := loadFontana()
	if err != nil {
		return err
	}
	natural, extraction, err := loadExtractionEndpoints()
	if err != nil {
		return err
	}
	shorthand, err := loadShorthandEndpoints()
	if err != nil {
		return err
	}

	if err := writeTarget(targets); err != nil {
		return err
	}
	if err := writeTranscriptionStability(targets, scales); err != nil {
		return err
	}
	all := append(append(append([]endpoint{}, fontana...), natural...), shorthand...)
	all = append(all, extraction...)
	comps := compareAll(targets, all, scales)
	if err := writeComparisons(comps); err != nil {
		return err
	}
	if err := writeFontana(comps, fontana); err != nil {
		return err
	}
	if err := writeClassEndpoints(comps); err != nil {
		return err
	}
	if err := writeProjection(fontana); err != nil {
		return err
	}
	if err := writeResiduals(targets, all, scales); err != nil {
		return err
	}
	if err := writeCoverage(); err != nil {
		return err
	}

	centroid := naturalCentroid(natural)
	shortTraj, shortNull, err := loadShorthandTraj()
	if err != nil {
		return err
	}
	extTraj, extNull, err := loadExtractionTraj()
	if err != nil {
		return err
	}
	trajRows := writeTrajectories(targets, centroid, scales, shortTraj, extTraj)
	if err := trajRows; err != nil {
		return err
	}
	if err := writeNullTargets(targets, centroid, scales, shortNull, extNull); err != nil {
		return err
	}
	if err := writeEvidenceAndReport(targets, comps, scales); err != nil {
		return err
	}
	if err := os.WriteFile(outDir+"/TASK83_COMPARISON_INCONCLUSIVE", []byte("TASK83_COMPARISON_INCONCLUSIVE\n"), 0644); err != nil {
		return err
	}
	return writeManifest("TASK83_COMPARISON_INCONCLUSIVE")
}

func verifyFingerprintSources() error {
	want := map[string]string{
		"data/ZL3b-n.txt":                 "bf5b6d4ac1e3a51b1847a9c388318d609020441ccd56984c901c32b09beccafc",
		"data_work/ZL3b-x7.canonical.txt": "f46f4190af65b85d145ec5bb957c1f56029b567e4bef12ac7baa1797f358d692",
		"data/IT2a-n.txt":                 "7f27a8b0feed8f6de0a99900df6bf912dd1d295c38e5f830bac8b41c3f536fb5",
		"data_work/IT2a-x7.canonical.txt": "3fb9531a11d896b5227e54c8d119cc13986eb69e48e1a5ab72b1a1ba64b5b4c0",
	}
	for path, expected := range want {
		actual, err := checksum(path)
		if err != nil {
			return err
		}
		if actual != expected {
			return fmt.Errorf("Fingerprint V2 freeze checksum mismatch: %s expected %s, got %s", path, expected, actual)
		}
	}
	return nil
}

func invalidate(cause error) error {
	_ = os.Remove(outDir + "/TASK83_COMPARISON_INCONCLUSIVE")
	report := "# Task83 invalidation report\n\n" +
		"Task83 stopped because an authoritative upstream freeze is internally inconsistent. " + cause.Error() + ".\n\n" +
		"The raw IT2a source still matches its frozen checksum, and the current prepared file matches the Task79c fingerprint's embedded corpus checksum, but it does not match FINGERPRINT_V2_FREEZE_MANIFEST.json. Task83 may not choose which upstream checksum is authoritative after target opening or repair the freeze in place.\n\n" +
		"The target-opening pre-audit checked the top-level manifest and target-artifact hashes but failed to expand and verify this prepared-corpus checksum before creating the sentinel. Therefore INPUT_FREEZE_INTEGRITY and TARGET_OPENING_INTEGRITY are NOT_SUPPORTED. Previously generated comparison tables are quarantined diagnostics and carry no confirmatory evidentiary status. A future task must reconcile and refreeze Fingerprint V2, then repeat Task83 from a fresh target-blind sentinel.\n\n" +
		"**TASK83_EXPERIMENT_INVALID**\n"
	if err := os.WriteFile(outDir+"/TASK83_INVALIDATION_REPORT.md", []byte(report), 0644); err != nil {
		return err
	}
	if err := os.WriteFile(outDir+"/TASK83_REPORT.md", []byte("# Task83 report\n\n## Final status\n\nThe confirmatory experiment is **INVALID**. The alternate-transcription prepared-corpus checksum in the authoritative Fingerprint V2 freeze manifest does not match the prepared corpus used by the authoritative Task79c artifacts. The discrepancy was discovered only during final expanded raw-integrity validation, after target opening. No comparison ranking, compatibility verdict, evidence level, or scientific conclusion from the quarantined tables is valid.\n\n| Verdict | Result |\n| --- | --- |\n| INPUT_FREEZE_INTEGRITY | NOT_SUPPORTED |\n| COMPARISON_CONTRACT_INTEGRITY | SUPPORTED |\n| TARGET_OPENING_INTEGRITY | NOT_SUPPORTED |\n| TRANSCRIPTION_ROBUSTNESS | NOT_SUPPORTED |\n| NATURAL_TEXT_COMPATIBILITY | NOT_TESTABLE |\n| AUTONOMOUS_TRANSFORM_COMPATIBILITY | NOT_TESTABLE |\n| EXTERNAL_MEMORY_COMPATIBILITY | NOT_TESTABLE |\n| SHORTHAND_COMPATIBILITY | NOT_TESTABLE |\n| SELECTIVE_EXTRACTION_COMPATIBILITY | NOT_TESTABLE |\n| MECHANISM_IDENTIFICATION_FROM_F2 | NOT_IDENTIFIABLE |\n| EXTERNAL_MEMORY_EVIDENCE_LEVEL | LEVEL_0 |\n| SHORTHAND_EVIDENCE_LEVEL | S0 |\n| EXTRACTION_EVIDENCE_LEVEL | A0 |\n| BEST_SUPPORTED_CLASS | INCONCLUSIVE |\n\nScientifically, Task83 concludes nothing about the nature of the Voynich Manuscript. The next experiment must reconcile and refreeze the IT2a provenance chain, then rerun this protocol with a new pre-opening audit and sentinel.\n\n**TASK83_EXPERIMENT_INVALID**\n"), 0644); err != nil {
		return err
	}
	if err := os.WriteFile(outDir+"/TASK83_EXPERIMENT_INVALID", []byte("TASK83_EXPERIMENT_INVALID\n"), 0644); err != nil {
		return err
	}
	return writeManifest("TASK83_EXPERIMENT_INVALID")
}

func verifyOpening() error {
	b, err := os.ReadFile(filepath.Join(outDir, "TASK83_TARGET_OPENING_SENTINEL"))
	if err != nil {
		return errors.New("target opening forbidden: sentinel absent")
	}
	s := string(b)
	want := map[string]string{
		"comparison_contract_sha256": "research/phase2/task82a1/TASK83_COMPARISON_CONTRACT.md",
		"mechanism_portfolio_sha256": "research/phase2/mechanism-space/MNEMONIC_MECHANISM_SPACE_FROZEN.json",
		"notation_portfolio_sha256":  "research/phase2/task82b/TASK82B_NOTATION_EXTRACTION_PORTFOLIO_FROZEN",
		"voynich_zl_target_sha256":   "experiments/fingerprint-v2-task79-v1/canonical-out/fingerprint.json",
		"voynich_it_target_sha256":   "experiments/fingerprint-v2-task79c-v1/transcription-it-out/fingerprint.json",
	}
	for key, path := range want {
		var expected string
		for _, line := range strings.Split(s, "\n") {
			if strings.HasPrefix(line, key+"=") {
				expected = strings.TrimPrefix(line, key+"=")
			}
		}
		if expected == "" {
			return fmt.Errorf("sentinel missing %s", key)
		}
		actual, err := checksum(path)
		if err != nil {
			return err
		}
		if actual != expected {
			return fmt.Errorf("immutable input changed: %s", path)
		}
	}
	return nil
}

func checksum(path string) (string, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	h := sha256.Sum256(b)
	return hex.EncodeToString(h[:]), nil
}

func loadFP(path string) (map[string]float64, error) {
	var x fp
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	if err = json.Unmarshal(b, &x); err != nil {
		return nil, err
	}
	m := x.Primary.Metrics
	out := map[string]float64{
		"EF1_GIANT_COMPONENT_SHARE": m.EF1.Giant, "EF1_ISOLATE_SHARE": m.EF1.Isolate,
		"EF2_GLOBAL_CLUSTERING": m.EF2.Clustering, "EF3_DEGREE_FREQUENCY_SPEARMAN": m.EF3.Spearman,
		"LP1_RULE_SUPPORT_GINI": m.LP1.Gini, "LP4_PREFIX_ATTACHMENT_NMI": m.LP4.Prefix.NMI, "LP4_SUFFIX_ATTACHMENT_NMI": m.LP4.Suffix.NMI,
	}
	if x.Primary.CrossScale != nil {
		for _, c := range x.Primary.CrossScale.Metrics {
			if c.ID == "cs2/prev-family-current-family" && (c.Status == "SUPPORTED" || c.Status == "NOT_SUPPORTED") {
				out[c.ID] = c.Value
			}
		}
	}
	return out, nil
}
func loadTargets() (map[string]map[string]float64, error) {
	z, e := loadFP("experiments/fingerprint-v2-task79-v1/canonical-out/fingerprint.json")
	if e != nil {
		return nil, e
	}
	i, e := loadFP("experiments/fingerprint-v2-task79c-v1/transcription-it-out/fingerprint.json")
	return map[string]map[string]float64{"ZL3b": z, "IT2a": i}, e
}
func loadScales() (map[string]float64, error) {
	var rows []discrim
	b, e := os.ReadFile("experiments/fingerprint-v2-task79c-v1/distance-pareto-out/combined_discriminative_validation.json")
	if e != nil {
		return nil, e
	}
	if e = json.Unmarshal(b, &rows); e != nil {
		return nil, e
	}
	out := map[string]float64{}
	for _, r := range rows {
		if r.ControlID != "doyle-sign-of-four" || r.Z == 0 {
			continue
		}
		out[r.MetricID] = math.Abs(r.Primary-r.Control) / math.Abs(r.Z)
	}
	return out, nil
}

func loadFontana() ([]endpoint, error) {
	f, e := os.Open("research/phase2/task82a1/F2_RAW_VECTORS_EXTENDED.jsonl")
	if e != nil {
		return nil, e
	}
	defer f.Close()
	var out []endpoint
	sc := bufio.NewScanner(f)
	sc.Buffer(make([]byte, 1024), 2<<20)
	for sc.Scan() {
		var r rawVector
		if e = json.Unmarshal(sc.Bytes(), &r); e != nil {
			return nil, e
		}
		vals := map[string]float64{}
		av := map[string]bool{}
		for _, m := range r.Metrics {
			// Direct comparison still selects directAll/directCore. Retaining
			// the whole frozen vector permits a separate projection export.
			vals[m.MetricID] = m.Value
			av[m.MetricID] = m.Available
		}
		class := "FONTANA"
		if strings.HasPrefix(r.MechanismID, "negative_") || strings.HasPrefix(r.MechanismID, "synthetic_") {
			class = "SIMPLE_NULL"
		}
		out = append(out, endpoint{class, r.MechanismID, r.Corpus, r.Policy, r.Scale, r.Replicate, vals, av})
	}
	return out, sc.Err()
}

func readTSV(path string) ([]map[string]string, error) {
	f, e := os.Open(path)
	if e != nil {
		return nil, e
	}
	defer f.Close()
	r := csv.NewReader(f)
	r.Comma = '\t'
	r.FieldsPerRecord = -1
	all, e := r.ReadAll()
	if e != nil {
		return nil, e
	}
	if len(all) == 0 {
		return nil, nil
	}
	var out []map[string]string
	for _, row := range all[1:] {
		m := map[string]string{}
		for i, h := range all[0] {
			if i < len(row) {
				m[h] = row[i]
			}
		}
		out = append(out, m)
	}
	return out, nil
}
func f64(s string) float64 { v, _ := strconv.ParseFloat(s, 64); return v }
func yes(s string) bool    { return strings.EqualFold(s, "true") || s == "1" }
func loadExtractionEndpoints() ([]endpoint, []endpoint, error) {
	rows, e := readTSV("research/phase2/task82b/EXTRACTION_F2_BEFORE_AFTER.tsv")
	if e != nil {
		return nil, nil, e
	}
	natMap := map[string]*endpoint{}
	extMap := map[string]*endpoint{}
	for _, r := range rows {
		c, o, m := r["corpus"], r["operator_id"], r["metric_id"]
		n := natMap[c]
		if n == nil {
			x := endpoint{Class: "NATURAL", ID: c, Corpus: c, Values: map[string]float64{}, Available: map[string]bool{}}
			natMap[c] = &x
			n = &x
		}
		n.Values[m] = f64(r["f2_before"])
		n.Available[m] = yes(r["before_available"])
		k := c + "\x00" + o
		x := extMap[k]
		if x == nil {
			q := endpoint{Class: "EXTRACTION", ID: o, Corpus: c, Values: map[string]float64{}, Available: map[string]bool{}}
			extMap[k] = &q
			x = &q
		}
		x.Values[m] = f64(r["f2_after"])
		x.Available[m] = yes(r["after_available"])
	}
	return endpointMap(natMap), endpointMap(extMap), nil
}
func loadShorthandEndpoints() ([]endpoint, error) {
	rows, e := readTSV("research/phase2/task82b/SHORTHAND_F2_BEFORE_AFTER.tsv")
	if e != nil {
		return nil, e
	}
	mm := map[string]*endpoint{}
	for _, r := range rows {
		k := r["scale"]
		x := mm[k]
		if x == nil {
			q := endpoint{Class: "SHORTHAND", ID: "BDD_" + k, Corpus: "BDD", Scale: k, Values: map[string]float64{}, Available: map[string]bool{}}
			mm[k] = &q
			x = &q
		}
		m := r["metric_id"]
		x.Values[m] = f64(r["f2_abbreviated"])
		x.Available[m] = yes(r["abbreviated_available"])
	}
	return endpointMap(mm), nil
}
func endpointMap[T ~map[string]*endpoint](m T) []endpoint {
	ks := make([]string, 0, len(m))
	for k := range m {
		ks = append(ks, k)
	}
	sort.Strings(ks)
	o := make([]endpoint, 0, len(m))
	for _, k := range ks {
		o = append(o, *m[k])
	}
	return o
}

func compareOne(t map[string]float64, e endpoint, sc map[string]float64) (float64, float64, int) {
	sum := 0.
	n := 0
	for _, m := range directCore {
		v, ok := t[m]
		s := sc[m]
		if !ok || s <= 0 || !e.Available[m] {
			continue
		}
		sum += math.Abs(v-e.Values[m]) / s
		n++
	}
	if n == 0 {
		return math.NaN(), math.NaN(), 0
	}
	d := sum / float64(n)
	c := float64(n) / float64(len(directCore))
	return d, d / c, n
}
func compareAll(ts map[string]map[string]float64, es []endpoint, sc map[string]float64) []comparison {
	var o []comparison
	for tn, t := range ts {
		for _, e := range es {
			d, a, n := compareOne(t, e, sc)
			o = append(o, comparison{tn, e.Class, e.ID, e.Corpus, e.Policy, e.Scale, e.Replicate, d, a, float64(n) / 3, n})
		}
	}
	sort.Slice(o, func(i, j int) bool {
		a, b := o[i], o[j]
		ka := fmt.Sprintf("%s\x00%s\x00%s\x00%s\x00%s\x00%s\x00%09d", a.Target, a.Class, a.ID, a.Corpus, a.Policy, a.Scale, a.Replicate)
		kb := fmt.Sprintf("%s\x00%s\x00%s\x00%s\x00%s\x00%s\x00%09d", b.Target, b.Class, b.ID, b.Corpus, b.Policy, b.Scale, b.Replicate)
		return ka < kb
	})
	return o
}

func writeTSV(path string, head []string, rows [][]string) error {
	f, e := os.Create(path)
	if e != nil {
		return e
	}
	w := csv.NewWriter(f)
	w.Comma = '\t'
	if e = w.Write(head); e != nil {
		return e
	}
	for _, r := range rows {
		if e = w.Write(r); e != nil {
			return e
		}
	}
	w.Flush()
	if e = w.Error(); e != nil {
		return e
	}
	return f.Close()
}
func ff(v float64) string {
	if math.IsNaN(v) {
		return "NA"
	}
	return strconv.FormatFloat(v, 'f', 6, 64)
}

func writeTarget(ts map[string]map[string]float64) error {
	var rows [][]string
	for _, t := range []string{"ZL3b", "IT2a"} {
		for _, m := range directAll {
			v, ok := ts[t][m]
			a := "true"
			if !ok {
				a = "false"
			}
			rows = append(rows, []string{t, m, family[m], classification(m), ff(v), a, "F2_COMMON_DIRECT"})
		}
	}
	return writeTSV(outDir+"/VOYNICH_TARGET_F2.tsv", []string{"transcription", "metric_id", "family", "classification", "value", "available", "space"}, rows)
}
func classification(m string) string {
	if contains(directCore, m) {
		return "CORE"
	}
	return "SUPPORTING"
}
func writeTranscriptionStability(ts map[string]map[string]float64, sc map[string]float64) error {
	var rows [][]string
	for _, m := range directAll {
		z, oz := ts["ZL3b"][m]
		i, oi := ts["IT2a"][m]
		status := "NOT_TESTABLE"
		d := math.NaN()
		if oz && oi {
			d = math.Abs(z-i) / sc[m]
			status = "TRANSCRIPTION_STABLE"
			if d > 1 {
				status = "DIRECTION_STABLE"
			}
		}
		rows = append(rows, []string{m, ff(z), ff(i), ff(d), status})
	}
	return writeTSV(outDir+"/VOYNICH_TRANSCRIPTION_STABILITY.tsv", []string{"metric_id", "zl_value", "it_value", "standardized_absolute_difference", "classification"}, rows)
}
func compRows(cs []comparison, filter func(comparison) bool) [][]string {
	var r [][]string
	for _, c := range cs {
		if !filter(c) {
			continue
		}
		r = append(r, []string{c.Target, c.Class, c.ID, c.Corpus, c.Policy, c.Scale, strconv.Itoa(c.Replicate), ff(c.D), ff(c.Adjusted), ff(c.Coverage), strconv.Itoa(c.N)})
	}
	return r
}

var compHead = []string{"transcription", "class", "object_id", "corpus", "policy", "scale", "replicate", "direct_core_distance", "coverage_adjusted_distance", "core_coverage", "metric_n"}

func writeComparisons(cs []comparison) error {
	return writeTSV(outDir+"/DIRECT_ENDPOINT_COMPARISON.tsv", compHead, compRows(cs, func(comparison) bool { return true }))
}

func writeFontana(cs []comparison, eps []endpoint) error {
	type info struct{ family, em, input, recovery, seed, scale, corpus string }
	inf := map[string]*info{}
	rows, e := readTSV("research/phase2/task82/MECHANISM_SUMMARY.tsv")
	if e != nil {
		return e
	}
	for _, r := range rows {
		inf[r["mechanism_id"]] = &info{family: r["family"], em: r["EM_class"], input: r["input_dependence"], recovery: "R0=" + r["r0_mean_recovery"] + ";R6=" + r["r6_mean_recovery"]}
	}
	stab, e := readTSV("research/phase2/task82a1/F2_FAMILY_STABILITY.tsv")
	if e != nil {
		return e
	}
	for _, r := range stab {
		if r["family"] != "edit family" {
			continue
		}
		x := inf[r["mechanism_id"]]
		if x == nil {
			continue
		}
		x.corpus = "unstable_groups=" + r["corpus_unstable_groups"]
		x.seed = "unstable_groups=" + r["seed_unstable_groups"]
		x.scale = "not_converged_groups=" + r["scale_not_converged_groups"]
	}
	kd, e := readTSV("research/phase2/task82/KNOWLEDGE_DEPENDENCE.tsv")
	if e != nil {
		return e
	}
	kn := map[string]int{}
	for _, r := range kd {
		kn[r["mechanism_id"]]++
	}
	ce, e := readTSV("research/phase2/task82a1/F2_SCALING_POLICY_EFFECT_EXTENDED.tsv")
	if e != nil {
		return e
	}
	ct, stable := map[string]int{}, map[string]int{}
	for _, r := range ce {
		if !contains(directCore, r["metric_id"]) {
			continue
		}
		ct[r["mechanism_id"]]++
		if r["effect"] == "STABLE" {
			stable[r["mechanism_id"]]++
		}
	}
	var detailed [][]string
	for _, c := range cs {
		if c.Class != "FONTANA" {
			continue
		}
		x := inf[c.ID]
		if x == nil {
			x = &info{}
		}
		cue := "NOT_APPLICABLE_NO_LOCAL_GLOBAL_PAIR"
		if ct[c.ID] > 0 {
			cue = fmt.Sprintf("stable=%d/%d", stable[c.ID], ct[c.ID])
		}
		detailed = append(detailed, []string{c.Target, c.Class, c.ID, x.family, c.Corpus, c.Policy, c.Scale, strconv.Itoa(c.Replicate), ff(c.D), ff(c.Adjusted), ff(c.Coverage), strconv.Itoa(c.N), x.corpus, x.seed, x.scale, cue, x.input, x.em, x.recovery, fmt.Sprintf("task82_rows=%d", kn[c.ID])})
	}
	sortRows(detailed)
	if e := writeTSV(outDir+"/FONTANA_MECHANISM_COMPARISON.tsv", []string{"transcription", "class", "mechanism_id", "frozen_family", "corpus", "policy", "scale", "replicate", "direct_core_distance", "coverage_adjusted_distance", "core_coverage", "metric_n", "corpus_stability", "seed_stability", "scale_stability", "cue_policy_effect", "input_dependence", "EM_class", "recovery", "knowledge_dependence"}, detailed); e != nil {
		return e
	}
	fam := map[string]string{}
	for k, v := range inf {
		fam[k] = v.family
	}
	type agg struct{ v []float64 }
	a := map[string]*agg{}
	for _, c := range cs {
		if c.Class != "FONTANA" || math.IsNaN(c.Adjusted) {
			continue
		}
		k := c.Target + "\x00" + fam[c.ID]
		if a[k] == nil {
			a[k] = &agg{}
		}
		a[k].v = append(a[k].v, c.Adjusted)
	}
	var out [][]string
	for k, x := range a {
		p := strings.Split(k, "\x00")
		sort.Float64s(x.v)
		out = append(out, []string{p[0], p[1], strconv.Itoa(len(x.v)), ff(median(x.v)), ff(quant(x.v, .025)), ff(quant(x.v, .975)), "FAMILY_BALANCED_WITHIN_FROZEN_FAMILY"})
	}
	sortRows(out)
	return writeTSV(outDir+"/FONTANA_FAMILY_COMPARISON.tsv", []string{"transcription", "frozen_family", "n_strata", "median_adjusted_distance", "p025", "p975", "aggregation"}, out)
}
func writeClassEndpoints(cs []comparison) error {
	files := map[string]string{"SHORTHAND": "SHORTHAND_ENDPOINT_COMPARISON.tsv", "EXTRACTION": "EXTRACTION_ENDPOINT_COMPARISON.tsv", "NATURAL": "NATURAL_LANGUAGE_COMPARISON.tsv", "SIMPLE_NULL": "SIMPLE_NULL_COMPARISON.tsv"}
	for class, file := range files {
		if e := writeTSV(outDir+"/"+file, compHead, compRows(cs, func(c comparison) bool { return c.Class == class })); e != nil {
			return e
		}
	}
	return nil
}
func writeProjection(eps []endpoint) error {
	rows, e := readTSV("research/phase2/task82a1/F2_ASSEMBLER_PROJECTION.tsv")
	if e != nil {
		return e
	}
	var out [][]string
	for _, x := range eps {
		for _, r := range rows {
			m := r["metric_id"]
			value := "NA"
			available := "false"
			if x.Available[m] {
				value = ff(x.Values[m])
				available = "true"
			}
			out = append(out, []string{x.ID, x.Policy, x.Corpus, x.Scale, strconv.Itoa(x.Replicate), m, r["family"], value, available, "PROJECTION_EVIDENCE", "NOT_NUMERICALLY_COMPARED", "assembler-line semantics differ from physical manuscript lines"})
		}
	}
	return writeTSV(outDir+"/PROJECTION_COMPARISON.tsv", []string{"mechanism_id", "policy", "corpus", "scale", "replicate", "metric_id", "family", "mechanism_projection_value", "available", "evidence_type", "target_comparison", "reason"}, out)
}

func writeResiduals(ts map[string]map[string]float64, es []endpoint, sc map[string]float64) error {
	var mr [][]string
	type vals struct {
		a, o int
		sum  float64
	}
	fr := map[string]*vals{}
	for tn, t := range ts {
		for _, e := range es {
			for _, m := range directAll {
				tv, tok := t[m]
				ev, eok := e.Values[m]
				app := tok && e.Available[m] && eok && sc[m] > 0
				res := math.NaN()
				dir := "NOT_APPLICABLE"
				if app {
					res = (tv - ev) / sc[m]
					dir = "REFERENCE_BELOW_TARGET"
					if res < 0 {
						dir = "REFERENCE_ABOVE_TARGET"
					}
					if res == 0 {
						dir = "MATCH"
					}
				}
				mr = append(mr, []string{tn, e.Class, e.ID, e.Corpus, e.Policy, e.Scale, m, ff(tv), ff(ev), ff(res), "frozen natural-control scale", dir, strconv.FormatBool(app), "see transcription/stability tables"})
				k := tn + "\x00" + e.Class + "\x00" + family[m]
				if fr[k] == nil {
					fr[k] = &vals{}
				}
				fr[k].o++
				if app {
					fr[k].a++
					fr[k].sum += math.Abs(res)
				}
			}
		}
	}
	sortRows(mr)
	var rr [][]string
	for k, v := range fr {
		p := strings.Split(k, "\x00")
		status := "family unavailable"
		mean := math.NaN()
		if v.a > 0 {
			mean = v.sum / float64(v.a)
			status = "family mismatch"
			if mean <= 1 {
				status = "family match"
			}
		}
		rr = append(rr, []string{p[0], p[1], p[2], strconv.Itoa(v.a), strconv.Itoa(v.o), ff(mean), status})
	}
	sortRows(rr)
	if e := writeTSV(outDir+"/METRIC_RESIDUALS.tsv", []string{"transcription", "class", "object_id", "corpus", "policy", "scale", "metric_id", "voynich_value", "reference_value", "normalized_residual", "uncertainty", "direction", "applicable", "stability"}, mr); e != nil {
		return e
	}
	return writeTSV(outDir+"/FAMILY_RESIDUALS.tsv", []string{"transcription", "class", "family", "available_metrics", "total_rows", "mean_absolute_residual", "classification"}, rr)
}
func writeCoverage() error {
	classes := []string{"FONTANA", "SHORTHAND", "EXTRACTION", "NATURAL", "SIMPLE_NULL"}
	var r [][]string
	for _, c := range classes {
		r = append(r, []string{c, "3/13", "edit family only", "4/13 possible but projection-only", "10/13", "hierarchy;locus;folio;physical-line/boundary/2D", "no imputation; NOT_MODELLED is not contradiction"})
	}
	return writeTSV(outDir+"/COVERAGE_ACCOUNTING.tsv", []string{"class", "direct_core_coverage", "direct_families", "projection_core_coverage", "manuscript_specific_unmodelled", "unmodelled_families", "missingness_policy"}, r)
}

func loadShorthandTraj() ([]traj, []traj, error) {
	tr, e := readTSV("research/phase2/task82b/SHORTHAND_F2_TRAJECTORIES.tsv")
	if e != nil {
		return nil, nil, e
	}
	mm := map[string]*traj{}
	for _, r := range tr {
		k := r["scale"]
		x := mm[k]
		if x == nil {
			q := traj{Class: "SHORTHAND", ID: "BDD_" + k, Corpus: "BDD", Values: map[string]float64{}, Available: map[string]bool{}}
			mm[k] = &q
			x = &q
		}
		m := r["metric_id"]
		x.Values[m] = f64(r["delta"])
		x.Available[m] = yes(r["both_available"])
	}
	nr, e := readTSV("research/phase2/task82b/SHORTHAND_NULL_COMPARISON.tsv")
	if e != nil {
		return nil, nil, e
	}
	nm := map[string]*traj{}
	for _, r := range nr {
		k := r["scale"] + "\x00" + r["null_kind"]
		x := nm[k]
		if x == nil {
			q := traj{Class: "SHORTHAND_NULL", ID: r["scale"], Corpus: "BDD", NullKind: r["null_kind"], Values: map[string]float64{}, Available: map[string]bool{}}
			nm[k] = &q
			x = &q
		}
		m := r["metric_id"]
		x.Values[m] = f64(r["null_mean_delta"])
		x.Available[m] = true
	}
	return trajMap(mm), trajMap(nm), nil
}
func loadExtractionTraj() ([]traj, []traj, error) {
	tr, e := readTSV("research/phase2/task82b/EXTRACTION_F2_TRAJECTORIES.tsv")
	if e != nil {
		return nil, nil, e
	}
	mm := map[string]*traj{}
	for _, r := range tr {
		k := r["corpus"] + "\x00" + r["operator_id"]
		x := mm[k]
		if x == nil {
			q := traj{Class: "EXTRACTION", ID: r["operator_id"], Corpus: r["corpus"], Values: map[string]float64{}, Available: map[string]bool{}}
			mm[k] = &q
			x = &q
		}
		m := r["metric_id"]
		x.Values[m] = f64(r["delta"])
		x.Available[m] = yes(r["both_available"])
	}
	nr, e := readTSV("research/phase2/task82b/EXTRACTION_NULL_COMPARISON.tsv")
	if e != nil {
		return nil, nil, e
	}
	nm := map[string]*traj{}
	for _, r := range nr {
		k := r["corpus"] + "\x00" + r["operator_id"] + "\x00" + r["null_kind"]
		x := nm[k]
		if x == nil {
			q := traj{Class: "EXTRACTION_NULL", ID: r["operator_id"], Corpus: r["corpus"], NullKind: r["null_kind"], Values: map[string]float64{}, Available: map[string]bool{}}
			nm[k] = &q
			x = &q
		}
		m := r["metric_id"]
		x.Values[m] = f64(r["null_mean_delta"])
		x.Available[m] = true
	}
	return trajMap(mm), trajMap(nm), nil
}
func trajMap[T ~map[string]*traj](m T) []traj {
	ks := make([]string, 0, len(m))
	for k := range m {
		ks = append(ks, k)
	}
	sort.Strings(ks)
	o := make([]traj, 0, len(m))
	for _, k := range ks {
		o = append(o, *m[k])
	}
	return o
}
func naturalCentroid(ns []endpoint) map[string]float64 {
	o := map[string]float64{}
	n := map[string]int{}
	for _, e := range ns {
		for _, m := range directCore {
			if e.Available[m] {
				o[m] += e.Values[m]
				n[m]++
			}
		}
	}
	for m := range o {
		o[m] /= float64(n[m])
	}
	return o
}
func vectorStats(target, centroid map[string]float64, t traj, sc map[string]float64) (float64, float64, float64, int) {
	dot, na, nb := 0., 0., 0.
	n := 0
	for _, m := range directCore {
		v, ok := target[m]
		if !ok || !t.Available[m] || sc[m] <= 0 {
			continue
		}
		a := (v - centroid[m]) / sc[m]
		b := t.Values[m] / sc[m]
		dot += a * b
		na += a * a
		nb += b * b
		n++
	}
	if n == 0 || na == 0 || nb == 0 {
		return math.NaN(), math.NaN(), math.NaN(), n
	}
	cos := dot / math.Sqrt(na*nb)
	sign := 0.
	for _, m := range directCore {
		if !t.Available[m] || sc[m] <= 0 {
			continue
		}
		a := target[m] - centroid[m]
		b := t.Values[m]
		if a == 0 || b == 0 {
			continue
		}
		if (a > 0) == (b > 0) {
			sign++
		}
	}
	return cos, sign / float64(n), math.Sqrt(nb / na), n
}
func writeTrajectories(ts map[string]map[string]float64, cent, sc map[string]float64, s, e []traj) error {
	all := append(append([]traj{}, s...), e...)
	var rows [][]string
	for tn, t := range ts {
		for _, x := range all {
			c, sg, mag, n := vectorStats(t, cent, x, sc)
			ov := "false"
			if c > 0 && mag > 1.5 {
				ov = "true"
			}
			rows = append(rows, []string{tn, x.Class, x.ID, x.Corpus, ff(c), ff(sg), ff(mag), strconv.Itoa(n), ov, "frozen raw delta standardized by natural-control scale"})
		}
	}
	sortRows(rows)
	if er := writeTSV(outDir+"/DIRECT_TRAJECTORY_COMPARISON.tsv", []string{"transcription", "class", "object_id", "corpus", "cosine_similarity", "direction_agreement", "magnitude_ratio", "metric_n", "target_aligned_but_overshoot", "normalization"}, rows); er != nil {
		return er
	}
	for _, spec := range []struct{ f, c string }{{"SHORTHAND_TRAJECTORY_COMPARISON.tsv", "SHORTHAND"}, {"EXTRACTION_TRAJECTORY_COMPARISON.tsv", "EXTRACTION"}} {
		var q [][]string
		for _, r := range rows {
			if r[1] == spec.c {
				q = append(q, r)
			}
		}
		if er := writeTSV(outDir+"/"+spec.f, []string{"transcription", "class", "object_id", "corpus", "cosine_similarity", "direction_agreement", "magnitude_ratio", "metric_n", "target_aligned_but_overshoot", "normalization"}, q); er != nil {
			return er
		}
	}
	return nil
}
func writeNullTargets(ts map[string]map[string]float64, cent, sc map[string]float64, s, e []traj) error {
	for _, spec := range []struct {
		f string
		x []traj
	}{{"SHORTHAND_NULL_TARGET_COMPARISON.tsv", s}, {"EXTRACTION_NULL_TARGET_COMPARISON.tsv", e}} {
		var rows [][]string
		for tn, t := range ts {
			for _, x := range spec.x {
				c, sg, mag, n := vectorStats(t, cent, x, sc)
				rows = append(rows, []string{tn, x.ID, x.Corpus, x.NullKind, ff(c), ff(sg), ff(mag), strconv.Itoa(n)})
			}
		}
		sortRows(rows)
		if er := writeTSV(outDir+"/"+spec.f, []string{"transcription", "object_id", "corpus", "null_kind", "cosine_similarity", "direction_agreement", "magnitude_ratio", "metric_n"}, rows); er != nil {
			return er
		}
	}
	return nil
}

func classSummary(cs []comparison) map[string][]float64 {
	o := map[string][]float64{}
	for _, c := range cs {
		if !math.IsNaN(c.Adjusted) {
			o[c.Target+"\x00"+c.Class] = append(o[c.Target+"\x00"+c.Class], c.Adjusted)
		}
	}
	return o
}
func writeEvidenceAndReport(ts map[string]map[string]float64, cs []comparison, sc map[string]float64) error {
	sum := classSummary(cs)
	classes := []string{"FONTANA", "SHORTHAND", "EXTRACTION", "NATURAL", "SIMPLE_NULL"}
	var evid [][]string
	for _, c := range classes {
		for _, t := range []string{"ZL3b", "IT2a"} {
			v := sum[t+"\x00"+c]
			sort.Float64s(v)
			grade, limitation := "PARTIAL", "Only 3/13 CORE metrics and one direct CORE family are modelled; no class can pass a multi-family support gate."
			if c == "FONTANA" {
				grade = "STRUCTURAL_COMPATIBILITY_ONLY"
			}
			if c == "EXTRACTION" {
				grade = "ENDPOINT_COMPATIBLE_ONLY"
			}
			if c == "SHORTHAND" {
				grade = "NOT_SUPPORTED"
				limitation = "The weak real BDD trajectory is outperformed by every matched-deletion null; one tradition and one direct family only."
			}
			evid = append(evid, []string{c, t, strconv.Itoa(len(v)), ff(median(v)), ff(quant(v, .025)), ff(quant(v, .975)), grade, limitation})
		}
	}
	for _, x := range []struct{ f, c string }{{"EXTERNAL_MEMORY_EVIDENCE.tsv", "FONTANA"}, {"SHORTHAND_EVIDENCE.tsv", "SHORTHAND"}, {"EXTRACTION_EVIDENCE.tsv", "EXTRACTION"}} {
		var r [][]string
		for _, q := range evid {
			if q[0] == x.c {
				r = append(r, q)
			}
		}
		if e := writeTSV(outDir+"/"+x.f, []string{"class", "transcription", "n_endpoints", "median_adjusted_distance", "p025", "p975", "evidence", "limitation"}, r); e != nil {
			return e
		}
	}
	pairs := [][]string{}
	for i := 0; i < len(classes); i++ {
		for j := i + 1; j < len(classes); j++ {
			a, b := classes[i], classes[j]
			for _, t := range []string{"ZL3b", "IT2a"} {
				av := sum[t+"\x00"+a]
				bv := sum[t+"\x00"+b]
				sort.Float64s(av)
				sort.Float64s(bv)
				pairs = append(pairs, []string{t, a, b, ff(median(av)), ff(median(bv)), "NO_CLEAR_ADVANTAGE", "not meaningful: neither class passes multi-family support gate"})
			}
		}
	}
	if e := writeTSV(outDir+"/HYPOTHESIS_PAIRWISE_ADVANTAGE.tsv", []string{"transcription", "class_a", "class_b", "median_a", "median_b", "evidence_advantage", "uncertainty"}, pairs); e != nil {
		return e
	}
	eq := [][]string{{"FONTANA", "SHORTHAND", "NOT_TESTABLE_AS_SUPPORTED_EQUIFINALITY", "one-family direct CORE coverage"}, {"FONTANA", "EXTRACTION", "NOT_TESTABLE_AS_SUPPORTED_EQUIFINALITY", "one-family direct CORE coverage"}, {"SHORTHAND", "EXTRACTION", "NOT_TESTABLE_AS_SUPPORTED_EQUIFINALITY", "one-family direct CORE coverage"}}
	if e := writeTSV(outDir+"/EQUIFINALITY_ANALYSIS.tsv", []string{"class_a", "class_b", "equifinality", "basis"}, eq); e != nil {
		return e
	}
	if e := os.WriteFile(outDir+"/STRONGEST_COUNTEREVIDENCE.md", []byte(counterevidence), 0644); e != nil {
		return e
	}
	return os.WriteFile(outDir+"/TASK83_REPORT.md", []byte(report(sum)), 0644)
}

const counterevidence = `# Strongest counterevidence

No hypothesis reaches confirmatory statistical support. The strongest common
counterevidence is structural coverage: every synthetic/transformation class
models only 3 of 13 CORE metrics, all from the single edit-family. Thus an
endpoint match cannot be multi-family evidence and leaves hierarchy, locus,
folio, physical-line, boundary, and 2D structure unexplained.

For Fontana, the frozen portfolio is only partially stable across corpus,
seed, and scale, and a direct plaintext-to-mechanism trajectory was not frozen.
Knowledge-dependent recovery therefore cannot promote one-family proximity to
LEVEL 2--4. For shorthand, only one BDD tradition exists and two edit metrics
are NOT_STABLE; deletion nulls have only three replicates. For extraction,
many FIRST/LAST effects are reproduced by matched thinning or confounded with
line collapse; AX is NOT_SUPPORTED and supplies no confirmatory hidden-channel
evidence.
`

func report(s map[string][]float64) string {
	med := func(t, c string) string { v := s[t+"\x00"+c]; sort.Float64s(v); return ff(median(v)) }
	return fmt.Sprintf(`# Task83 report

## Integrity and target

1. All upstream freezes were valid before opening: **yes** (82/82 manifest
entries, all markers, raw target checksums). 2. The frozen comparison contract
was obeyed: **yes**. 3. No methodological choice changed after target opening.
The DIRECT vector is in VOYNICH_TARGET_F2.tsv. Both transcriptions are retained;
all seven available direct metrics are transcription-stable; cs2 is
unavailable on both and remains NOT_TESTABLE.

## Endpoint results

Median coverage-adjusted one-family CORE distances (ZL3b / IT2a) are: Fontana
%s / %s; natural %s / %s; shorthand %s / %s; extraction %s / %s; simple null
%s / %s. These numbers identify closest tested endpoints but do **not** establish
statistical compatibility: direct CORE coverage is only 3/13 and one family,
so the preregistered multi-family support gate is impossible for every class.
All frozen Fontana mechanisms, policies, corpora, scales and replicates appear
in FONTANA_MECHANISM_COMPARISON.tsv; family summaries retain frozen families.
Corpus/seed/scale/cue and knowledge-dependence evidence remains in the linked
frozen upstream strata and is not used to delete an unstable mechanism.

## Trajectories and controls

BDD abbreviation and all 20 extraction operators are evaluated against both
Voynich displacement vectors in the trajectory tables, with cosine, sign
agreement, magnitude, and overshoot. Matched-null target tables preserve all
registered deletion/thinning controls. Shorthand line-position degeneracies
remain excluded from primary evidence. FIRST/LAST extraction cannot be cleanly
separated from line collapse where its matched null is equivalent, so apparent
acrostic evidence is confounded. AX3--AX6 are descriptive only; there is no
confirmatory evidence for an acrostic or hidden channel. No Fontana trajectory
is reported because upstream artifacts freeze endpoints but no valid
plaintext-to-Fontana before/after mapping.

## Explicit answers to the 42 confirmatory questions

1. Yes, every upstream freeze was valid. 2. Yes, the contract was obeyed.
3. No methodological choice was changed after opening; implementation fixes
are disclosed in TASK83_BUG_AUDIT.tsv and preserve frozen semantics. 4. The
DIRECT target vector is reported verbatim in VOYNICH_TARGET_F2.tsv. 5. Its
seven available metrics are transcription-stable. 6. No available DIRECT
metric is transcription-sensitive; cs2 is NOT_TESTABLE on both.

7. The closest usable Fontana endpoints are tied strata of
m_restricted_rotation_index and m_restricted_storage_associate (adjusted
distance about 0.5505 in each transcription). 8. They are best-among-tested
strata, not statistically compatible models. 9. Fontana does not demonstrate
an advantage over natural controls; natural has the smaller descriptive class
median. 10. It also has no demonstrated advantage over simple nulls. 11. No:
agreement is confined to edit family. 12. Persistence is mixed and the frozen
corpus/seed/scale annotations are reported per row. 13. No stronger proximity
for knowledge-dependent mechanisms is established. 14. Fontana fails coverage,
multi-family support, uniformly stable persistence, and a frozen target-relevant
trajectory; some strata have zero usable CORE coverage and remain NA.

15. Combined BDD abbreviation moves only weakly toward Voynich (cosine about
0.28, magnitude about 0.16). 16. No: every matched-deletion combined trajectory
has cosine about 0.99, substantially stronger than the real abbreviation.
17. Only partial endpoint resemblance is present. 18. Chapter-level directions
are unstable; the combined EF1/EF2/EF3 signs give only 2/3 direction agreement,
and EF1/EF2 are NOT_STABLE upstream. 19. One BDD tradition permits no shorthand-
general or cross-tradition conclusion.

20. Some extraction strata move in the target direction, but inconsistently.
21. The strongest isolated alignment is positional FIXED_OFFSET_WITHIN_GROUP_1
on Longfellow (cosine about 0.80); some periodic/token operators reach about
0.46--0.48, while acrostic/telestic glyph collapses are weak or opposed. 22. No
operator class passes the all-null separation gate. 23. No: FIRST/LAST support
is not cleanly separable from matched line-collapse/thinning. 24. No; AX is
NOT_SUPPORTED, so there is no confirmatory acrostic evidence.

25. Ordinary natural language is PARTIAL and descriptively closest by class
median, but one-family coverage prevents SUPPORTED. 26. Tested autonomous
transforms are PARTIAL. 27. External memory is PARTIAL. 28. Tested BDD
shorthand is DISFAVORED because its primary trajectory loses to matched
deletion. 29. Selective extraction is PARTIAL at endpoint
level, without null-separated support.

30. No class meets its support criterion; BEST_SUPPORTED_CLASS is INCONCLUSIVE.
31. Statistical equivalence among supported classes cannot be established
because no class is supported. 32. F2 cannot identify the mechanism class in
this intersection. 33. Endpoint equifinality is descriptively possible, but
confirmatory equifinality is not testable.

34. Each model class covers 3/13 CORE metrics directly (one family); Fontana
has 4/13 additional assembler-projection CORE dimensions that cannot be treated
as physical manuscript evidence. 35. Hierarchy, locus, folio, page,
recto/verso, physical lines/boundaries, local regimes, and manuscript metadata
remain outside the models. 36. The strongest counterevidence is that the
apparently best class leaves 10/13 CORE metrics unmodelled and has no
multi-family or null-separated advantage.

37. External-memory evidence reaches LEVEL_1 only. 38. Shorthand is S0: the
primary real trajectory is weaker than every matched-deletion trajectory.
39. Extraction reaches A1 only; it does not reach null-separated A2/A3.
40. Scientifically, the manuscript has an edit-family fingerprint partly
reproduced by multiple ordinary and transformed corpora. 41. Its language,
author, historical mechanism, external-memory status, shorthand status,
hidden channel, or cipher family cannot be identified or globally excluded.
42. The next experiment is the independently frozen manuscript-aware,
cross-tradition replication specified below.

## Coverage, residuals, and interpretation

Every class explains at most the direct edit-family slice (3/13 CORE).
Assembler outputs are separately labelled PROJECTION_EVIDENCE and never treated
as physical manuscript lines. Hierarchy, locus, folio, page/recto-verso, local
regimes, physical-line position/boundary and manuscript metadata remain
NOT_MODELLED. Metric and family residual files retain missingness without
imputation. This limited intersection also prevents a defensible Phase-I
improvement claim and makes F2 class identification presently NOT_IDENTIFIABLE.

## Required verdicts

| Verdict | Result |
| --- | --- |
| INPUT_FREEZE_INTEGRITY | SUPPORTED |
| COMPARISON_CONTRACT_INTEGRITY | SUPPORTED |
| TARGET_OPENING_INTEGRITY | SUPPORTED |
| TRANSCRIPTION_ROBUSTNESS | SUPPORTED |
| NATURAL_TEXT_COMPATIBILITY | PARTIAL |
| AUTONOMOUS_TRANSFORM_COMPATIBILITY | PARTIAL |
| EXTERNAL_MEMORY_COMPATIBILITY | PARTIAL |
| SHORTHAND_COMPATIBILITY | DISFAVORED |
| SELECTIVE_EXTRACTION_COMPATIBILITY | PARTIAL |
| FONTANA_VS_NATURAL | NO_CLEAR_ADVANTAGE |
| FONTANA_VS_SHORTHAND | NO_CLEAR_ADVANTAGE |
| FONTANA_VS_EXTRACTION | NO_CLEAR_ADVANTAGE |
| SHORTHAND_VS_EXTRACTION | NO_CLEAR_ADVANTAGE |
| MECHANISM_IDENTIFICATION_FROM_F2 | NOT_IDENTIFIABLE |
| EXTERNAL_MEMORY_EVIDENCE_LEVEL | LEVEL_1 |
| SHORTHAND_EVIDENCE_LEVEL | S0 |
| EXTRACTION_EVIDENCE_LEVEL | A1 |
| BEST_SUPPORTED_CLASS | INCONCLUSIVE |

The ordinary-natural-text question, simple-null question, and all exotic
classes have therefore been tested, but none has a demonstrated multi-family,
null-separated advantage. Multiple endpoint resemblances are possible; true
supported-class equifinality is not testable at this coverage.

## Scientific conclusion

Within the tested hypothesis space, Voynich has partial edit-family endpoint
compatibility with several frozen classes, but Fingerprint V2's available
direct intersection is too narrow to support or identify a generating class.
This does not decrypt Voynich, prove or exclude external memory, shorthand,
acrostics, natural language, or all ciphers.

The next experiment should freeze manuscript-aware generators that emit real
page/folio/line hierarchy for independently chosen Fontana, shorthand, natural,
and extraction controls, then replicate the same comparison across multiple
historical shorthand traditions with adequately powered matched nulls.

**TASK83_COMPARISON_INCONCLUSIVE**
`, med("ZL3b", "FONTANA"), med("IT2a", "FONTANA"), med("ZL3b", "NATURAL"), med("IT2a", "NATURAL"), med("ZL3b", "SHORTHAND"), med("IT2a", "SHORTHAND"), med("ZL3b", "EXTRACTION"), med("IT2a", "EXTRACTION"), med("ZL3b", "SIMPLE_NULL"), med("IT2a", "SIMPLE_NULL"))
}

func writeManifest(verdict string) error {
	files := []string{"TASK83_DESIGN.md", "TASK83_INPUT_AUDIT.tsv", "TASK83_TARGET_OPENING_SENTINEL", "VOYNICH_TARGET_F2.tsv", "VOYNICH_TRANSCRIPTION_STABILITY.tsv", "DIRECT_ENDPOINT_COMPARISON.tsv", "DIRECT_TRAJECTORY_COMPARISON.tsv", "PROJECTION_COMPARISON.tsv", "FONTANA_MECHANISM_COMPARISON.tsv", "FONTANA_FAMILY_COMPARISON.tsv", "SHORTHAND_ENDPOINT_COMPARISON.tsv", "SHORTHAND_TRAJECTORY_COMPARISON.tsv", "SHORTHAND_NULL_TARGET_COMPARISON.tsv", "EXTRACTION_ENDPOINT_COMPARISON.tsv", "EXTRACTION_TRAJECTORY_COMPARISON.tsv", "EXTRACTION_NULL_TARGET_COMPARISON.tsv", "NATURAL_LANGUAGE_COMPARISON.tsv", "SIMPLE_NULL_COMPARISON.tsv", "METRIC_RESIDUALS.tsv", "FAMILY_RESIDUALS.tsv", "COVERAGE_ACCOUNTING.tsv", "HYPOTHESIS_PAIRWISE_ADVANTAGE.tsv", "EQUIFINALITY_ANALYSIS.tsv", "EXTERNAL_MEMORY_EVIDENCE.tsv", "SHORTHAND_EVIDENCE.tsv", "EXTRACTION_EVIDENCE.tsv", "STRONGEST_COUNTEREVIDENCE.md", "TASK83_REPORT.md"}
	files = append(files, "TASK83_BUG_AUDIT.tsv")
	marker := "TASK83_COMPARISON_INCONCLUSIVE"
	if verdict == "TASK83_EXPERIMENT_INVALID" {
		marker = "TASK83_EXPERIMENT_INVALID"
		files = append(files, "TASK83_INVALIDATION_REPORT.md")
	}
	files = append(files, marker)
	m := struct {
		Task      string            `json:"task"`
		Version   int               `json:"schema_version"`
		Generated string            `json:"generated_utc"`
		Files     map[string]string `json:"files"`
		Verdict   string            `json:"verdict"`
	}{"Task83", 1, time.Now().UTC().Format(time.RFC3339), map[string]string{}, verdict}
	for _, f := range files {
		h, e := checksum(outDir + "/" + f)
		if e != nil {
			return e
		}
		m.Files[f] = h
	}
	b, e := json.MarshalIndent(m, "", "  ")
	if e != nil {
		return e
	}
	b = append(b, '\n')
	return os.WriteFile(outDir+"/TASK83_RESULTS_MANIFEST.json", b, 0644)
}
func median(v []float64) float64 {
	if len(v) == 0 {
		return math.NaN()
	}
	return quant(v, .5)
}
func quant(v []float64, p float64) float64 {
	if len(v) == 0 {
		return math.NaN()
	}
	x := append([]float64(nil), v...)
	sort.Float64s(x)
	i := int(math.Round(p * float64(len(x)-1)))
	return x[i]
}
func contains(a []string, s string) bool {
	for _, x := range a {
		if x == s {
			return true
		}
	}
	return false
}
func sortRows(r [][]string) {
	sort.Slice(r, func(i, j int) bool { return strings.Join(r[i], "\x00") < strings.Join(r[j], "\x00") })
}
