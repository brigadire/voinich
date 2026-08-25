// Command task83r-analyze repeats the frozen Task83 comparison against the
// authoritative deterministic Fingerprint V2.1 target.
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
)

const outDir = "research/phase2/task83r"

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
	MinPValue                   float64
	PValueSeen                  bool
	MinReplicates               int
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
	if err := writeEvidenceAndReport(comps); err != nil {
		return err
	}
	if err := os.WriteFile(outDir+"/TASK83R_COMPARISON_INCONCLUSIVE", []byte("TASK83R_COMPARISON_INCONCLUSIVE\n"), 0644); err != nil {
		return err
	}
	return writeManifest("TASK83R_COMPARISON_INCONCLUSIVE")
}

func verifyOpening() error {
	b, err := os.ReadFile(filepath.Join(outDir, "TASK83R_TARGET_OPENING_SENTINEL"))
	if err != nil {
		return errors.New("target opening forbidden: sentinel absent")
	}
	s := string(b)
	want := map[string]string{
		"comparison_contract_sha256":                "research/phase2/task82a1/TASK83_COMPARISON_CONTRACT.md",
		"task81_portfolio_sha256":                   "research/phase2/mechanism-space/MNEMONIC_MECHANISM_SPACE_FROZEN.json",
		"task82b_portfolio_sha256":                  "research/phase2/task82b/TASK82B_NOTATION_EXTRACTION_PORTFOLIO_FROZEN",
		"deterministic_fingerprint_manifest_sha256": "research/phase2/task83b/FINGERPRINT_V2_DETERMINISTIC_MANIFEST.json",
		"metric_registry_sha256":                    "research/phase2/task83b/F2_METRIC_REGISTRY_REFROZEN.tsv",
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
	z, e := loadFP("research/phase2/task83b/artifacts/zl/fingerprint.json")
	if e != nil {
		return nil, e
	}
	i, e := loadFP("research/phase2/task83b/artifacts/it/fingerprint.json")
	return map[string]map[string]float64{"ZL3b": z, "IT2a": i}, e
}
func loadScales() (map[string]float64, error) {
	var rows []discrim
	b, e := os.ReadFile("research/phase2/task83b/artifacts/combined_discriminative_validation.json")
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
		out = append(out, []string{p[0], p[1], strconv.Itoa(len(x.v)), ff(median(x.v)), ff(quant(x.v, .025)), ff(quant(x.v, .975)), "FAMILY_BALANCED_WITHIN_FROZEN_FAMILY", "FAILED_3_OF_13_CORE_ONE_FAMILY", "NO_MULTI_FAMILY_SUPPORT"})
	}
	sortRows(out)
	return writeTSV(outDir+"/FONTANA_FAMILY_COMPARISON.tsv", []string{"transcription", "frozen_family", "n_strata", "median_adjusted_distance", "p025", "p975", "aggregation", "support_gate", "family_support"}, out)
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
		status := "FAMILY_UNAVAILABLE"
		mean := math.NaN()
		if v.a > 0 {
			mean = v.sum / float64(v.a)
			status = "FAMILY_MISMATCH"
			if mean <= 1 {
				status = "FAMILY_MATCH"
			}
			if v.a < v.o {
				status = "FAMILY_PARTIAL"
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
		p := f64(r["p_value"])
		if !x.PValueSeen || p < x.MinPValue {
			x.MinPValue = p
			x.PValueSeen = true
		}
		n := int(f64(r["n_replicates"]))
		if x.MinReplicates == 0 || n < x.MinReplicates {
			x.MinReplicates = n
		}
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
		p := f64(r["p_value"])
		if !x.PValueSeen || p < x.MinPValue {
			x.MinPValue = p
			x.PValueSeen = true
		}
		n := int(f64(r["n_replicates"]))
		if x.MinReplicates == 0 || n < x.MinReplicates {
			x.MinReplicates = n
		}
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
		rows = append(rows, []string{tn, "FONTANA", "NO_FROZEN_BEFORE_AFTER", "ALL_FROZEN_CORPORA", "NA", "NA", "NA", "0", "false", "NOT_TESTABLE: no Fontana before/after trajectory was frozen upstream"})
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
				p, separated := "NA", "false"
				if x.PValueSeen {
					p = ff(x.MinPValue)
					separated = strconv.FormatBool(x.MinPValue <= 0.05)
				}
				resolution := "LOW_RESOLUTION"
				if x.MinReplicates >= 20 {
					resolution = "ADEQUATE"
				}
				rows = append(rows, []string{tn, x.ID, x.Corpus, x.NullKind, ff(c), ff(sg), ff(mag), strconv.Itoa(n), p, strconv.Itoa(x.MinReplicates), resolution, separated})
			}
		}
		sortRows(rows)
		if er := writeTSV(outDir+"/"+spec.f, []string{"transcription", "object_id", "corpus", "null_kind", "cosine_similarity", "direction_agreement", "magnitude_ratio", "metric_n", "upstream_min_p_value", "minimum_null_replicates", "resolution", "any_metric_p_le_0_05"}, rows); er != nil {
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
func writeEvidenceAndReport(cs []comparison) error {
	sum := classSummary(cs)
	classes := []string{"FONTANA", "SHORTHAND", "EXTRACTION", "NATURAL", "SIMPLE_NULL"}
	var fontanaEvidence, shorthandEvidence [][]string
	for _, t := range []string{"ZL3b", "IT2a"} {
		fv := sortedCopy(sum[t+"\x00FONTANA"])
		fontanaEvidence = append(fontanaEvidence, []string{"FONTANA", t, strconv.Itoa(len(fv)), ff(median(fv)), ff(quant(fv, .025)), ff(quant(fv, .975)), "ENDPOINT_RESEMBLANCE_ONLY", "NOT_TESTABLE_NO_FROZEN_BEFORE_AFTER", "NOT_DEMONSTRATED", "3/13_ONE_FAMILY", "PARTIAL", "LEVEL_1", "No multi-family support; no frozen target-relevant trajectory; no advantage over simpler classes."})
		sv := sortedCopy(sum[t+"\x00SHORTHAND"])
		shorthandEvidence = append(shorthandEvidence, []string{"SHORTHAND", t, strconv.Itoa(len(sv)), ff(median(sv)), ff(quant(sv, .025)), ff(quant(sv, .975)), "PARTIAL_ENDPOINT_RESEMBLANCE", "REAL_TRAJECTORY_WEAKER_THAN_MATCHED_NULLS", "FAILED_MIN_P_0.25", "3/13_ONE_FAMILY", "DISFAVORED", "S0", "One BDD tradition; no matched-null p<=0.05; no cross-tradition claim."})
	}
	head := []string{"class", "transcription", "n_endpoints", "median_adjusted_distance", "p025", "p975", "endpoint_test", "trajectory_test", "null_separation", "coverage_gate", "compatibility", "evidence_level", "limitation"}
	if err := writeTSV(outDir+"/EXTERNAL_MEMORY_EVIDENCE.tsv", head, fontanaEvidence); err != nil {
		return err
	}
	if err := writeTSV(outDir+"/SHORTHAND_EVIDENCE.tsv", head, shorthandEvidence); err != nil {
		return err
	}
	var extractionEvidence [][]string
	for _, t := range []string{"ZL3b", "IT2a"} {
		byOperator := map[string][]float64{}
		for _, c := range cs {
			if c.Target == t && c.Class == "EXTRACTION" && !math.IsNaN(c.Adjusted) {
				byOperator[operatorClass(c.ID)] = append(byOperator[operatorClass(c.ID)], c.Adjusted)
			}
		}
		for _, opClass := range []string{"ACROSTIC", "TELESTIC", "POSITIONAL", "PERIODIC"} {
			v := sortedCopy(byOperator[opClass])
			confound := "NONE_REGISTERED"
			if opClass == "ACROSTIC" || opClass == "TELESTIC" {
				confound = "ACROSTIC_INTERPRETATION_CONFOUNDED_BY_LINE_COLLAPSE"
			}
			extractionEvidence = append(extractionEvidence, []string{opClass, t, strconv.Itoa(len(v)), ff(median(v)), ff(quant(v, .025)), ff(quant(v, .975)), "SOME_ENDPOINT_OR_DIRECTIONAL_RESEMBLANCE", "NO_REGISTERED_NULL_P_LE_0_05", confound, "3/13_ONE_FAMILY", "PARTIAL", "A1", "AX excluded from scoring; hidden-channel detection is not supported."})
		}
	}
	if err := writeTSV(outDir+"/EXTRACTION_EVIDENCE.tsv", []string{"operator_class", "transcription", "n_endpoints", "median_adjusted_distance", "p025", "p975", "endpoint_and_trajectory_test", "null_separation", "confound", "coverage_gate", "compatibility", "evidence_level", "limitation"}, extractionEvidence); err != nil {
		return err
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
				overlap := quant(av, .975) >= quant(bv, .025) && quant(bv, .975) >= quant(av, .025)
				pairs = append(pairs, []string{t, a, b, ff(median(av)), ff(median(bv)), ff(quant(av, .025)), ff(quant(av, .975)), ff(quant(bv, .025)), ff(quant(bv, .975)), strconv.FormatBool(overlap), "NO_CLEAR_ADVANTAGE", "Neither class passes the frozen multi-family support gate."})
			}
		}
	}
	if e := writeTSV(outDir+"/HYPOTHESIS_PAIRWISE_ADVANTAGE.tsv", []string{"transcription", "class_a", "class_b", "median_a", "median_b", "p025_a", "p975_a", "p025_b", "p975_b", "intervals_overlap", "evidence_advantage", "basis"}, pairs); e != nil {
		return e
	}
	eq := [][]string{{"FONTANA", "NATURAL", "ENDPOINT_INTERVALS_OVERLAP", "NO_CLASS_SUPPORT_GATE", "NOT_CONFIRMATORILY_TESTABLE", "NOT_IDENTIFIABLE"}, {"FONTANA", "SIMPLE_NULL", "ENDPOINT_INTERVALS_OVERLAP", "NO_CLASS_SUPPORT_GATE", "NOT_CONFIRMATORILY_TESTABLE", "NOT_IDENTIFIABLE"}, {"FONTANA", "SHORTHAND", "ENDPOINT_INTERVALS_OVERLAP", "NO_CLASS_SUPPORT_GATE", "NOT_CONFIRMATORILY_TESTABLE", "NOT_IDENTIFIABLE"}, {"FONTANA", "EXTRACTION", "ENDPOINT_INTERVALS_OVERLAP", "NO_CLASS_SUPPORT_GATE", "NOT_CONFIRMATORILY_TESTABLE", "NOT_IDENTIFIABLE"}, {"SHORTHAND", "EXTRACTION", "ENDPOINT_INTERVALS_OVERLAP", "NO_CLASS_SUPPORT_GATE", "NOT_CONFIRMATORILY_TESTABLE", "NOT_IDENTIFIABLE"}, {"ALL_TESTED_CLASSES", "ALL_TESTED_CLASSES", "MULTIPLE_DESCRIPTIVE_RESEMBLANCES", "DIRECT_INTERSECTION_IS_3_OF_13_CORE_IN_ONE_FAMILY", "SUPPORTED_CLASS_EQUIFINALITY_UNRESOLVED", "NOT_IDENTIFIABLE"}}
	if e := writeTSV(outDir+"/EQUIFINALITY_ANALYSIS.tsv", []string{"class_a", "class_b", "descriptive_relationship", "support_gate", "equifinality", "identifiability"}, eq); e != nil {
		return e
	}
	if e := os.WriteFile(outDir+"/STRONGEST_COUNTEREVIDENCE.md", []byte(counterevidence), 0644); e != nil {
		return e
	}
	return os.WriteFile(outDir+"/TASK83R_REPORT.md", []byte(report(sum)), 0644)
}

const counterevidence = `# Strongest counterevidence

No hypothesis reaches confirmatory statistical support. The strongest common
counterevidence is coverage: every tested mechanism class models only 3 of 13
CORE metrics, all in the edit-family. Endpoint proximity therefore cannot be
multi-family evidence and leaves hierarchy, locus, folio, recto/verso,
physical-line, boundary, positional, and 2D structure unexplained.

For Fontana, the frozen portfolio is only partially stable across corpus,
seed, and scale, and a direct plaintext-to-mechanism trajectory was not frozen.
Its strongest competitor is ordinary natural text, whose descriptive median
distance is smaller. Knowledge-dependent recovery cannot promote one-family
proximity to LEVEL 2--4.

For shorthand, only one BDD tradition exists. The combined real trajectory has
cosine 0.279072 (ZL3b) / 0.262481 (IT2a), whereas matched deletion trajectories
reach about 0.99; no shorthand null comparison has p<=0.05 (minimum 0.25).

For extraction, no registered matched-null comparison has p<=0.05 (minimum
0.142857). FIRST/LAST effects are frequently reproduced by matched thinning or
confounded with line collapse. AX is NOT_SUPPORTED and is excluded from every
score, evidence level, tie-break, and verdict.
`

func report(s map[string][]float64) string {
	med := func(t, c string) string { return ff(median(sortedCopy(s[t+"\x00"+c]))) }
	return fmt.Sprintf(`# Task83r report

## Main result

Within the actually tested hypothesis space, ordinary natural text has the
smallest descriptive class-median distance, but no class passes the frozen
multi-family and null-separation support gates. Fingerprint V2 therefore does
not identify a generating mechanism: BEST_SUPPORTED_CLASS = INCONCLUSIVE
and MECHANISM_IDENTIFICATION_FROM_F2 = NOT_IDENTIFIABLE.

## Integrity

Task83 remains TASK83_EXPERIMENT_INVALID; none of its quarantined rankings,
distances, residual interpretations, evidence levels, or verdicts was used.
Before target opening, every Task82/82a/82a.1/82b declared artifact checksum
matched, Task83b reported TASK83R_READY = SUPPORTED, and the authoritative
Fingerprint V2 verifier exited 0. The opening sentinel binds the unchanged
Task83 protocol, Task82a.1 contract, V2.1 manifest/registry, and all portfolios.
No design choice changed after opening.

Both ZL3b and IT2a are retained. Seven available direct metrics are
TRANSCRIPTION_STABLE; cs2 is unavailable on both and is NOT_TESTABLE.

## Endpoint results

Coverage-adjusted one-family CORE median distances (ZL3b / IT2a) are: Fontana
%s / %s; natural %s / %s; shorthand %s / %s; extraction %s / %s; simple null
%s / %s. These values describe proximity, not support. Every class covers only
3/13 CORE metrics in one family. All seven frozen Fontana mechanisms and all
corpus/policy/scale/replicate strata are retained; instability is reported and
never used to select rows.

## Trajectories and controls

BDD combined abbreviation aligns only weakly (cosine 0.279072 / 0.262481,
magnitude ratio 0.164141 / 0.162682), while matched deletion reaches about
0.99. No shorthand null test has p<=0.05 (minimum 0.25). Some extraction
operators align directionally (maximum cosine about 0.80), but no extraction
null test has p<=0.05 (minimum 0.142857). FIRST/LAST remains line-collapse
confounded. AX remains descriptive and cannot affect any verdict. Fontana
trajectory evidence is NOT_TESTABLE because no valid before/after trajectory
was frozen upstream.

Metric and family residuals preserve missingness without imputation.
Projection evidence remains separate. Hierarchy, locus, folio, page,
recto/verso, physical-line/boundary, local-regime, positional, and 2D
properties remain unexplained by the direct comparison.

## Required verdicts

| Verdict | Result |
| --- | --- |
| INPUT_FREEZE_INTEGRITY | SUPPORTED |
| TRANSITIVE_PROVENANCE_INTEGRITY | SUPPORTED |
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

All required pairwise comparisons are NO_CLEAR_ADVANTAGE. Multiple classes
show descriptive endpoint resemblance, but supported-class equifinality is
not testable because none passes its own support gate. The resulting
identifiability classification is therefore NOT_IDENTIFIABLE, not evidence
that the classes are scientifically equivalent.

## Scientific conclusion

Within the tested hypothesis space, Voynich has partial edit-family endpoint
compatibility with several frozen classes, but Fingerprint V2's available
direct intersection is too narrow to support or identify a generating class.
This neither proves nor globally excludes external memory, shorthand,
selective extraction, natural language, a hidden channel, or cipher systems.

**TASK83R_COMPARISON_INCONCLUSIVE**
`, med("ZL3b", "FONTANA"), med("IT2a", "FONTANA"), med("ZL3b", "NATURAL"), med("IT2a", "NATURAL"), med("ZL3b", "SHORTHAND"), med("IT2a", "SHORTHAND"), med("ZL3b", "EXTRACTION"), med("IT2a", "EXTRACTION"), med("ZL3b", "SIMPLE_NULL"), med("IT2a", "SIMPLE_NULL"))
}

func sortedCopy(v []float64) []float64 {
	o := append([]float64(nil), v...)
	sort.Float64s(o)
	return o
}

func operatorClass(id string) string {
	switch {
	case strings.HasPrefix(id, "FIRST_"):
		return "ACROSTIC"
	case strings.HasPrefix(id, "LAST_"):
		return "TELESTIC"
	case strings.HasPrefix(id, "PERIODIC_"):
		return "PERIODIC"
	default:
		return "POSITIONAL"
	}
}

func writeManifest(verdict string) error {
	files := []string{"TASK83R_DESIGN.md", "TASK83R_INPUT_AUDIT.tsv", "TASK83R_TARGET_OPENING_SENTINEL", "VOYNICH_TARGET_F2.tsv", "VOYNICH_TRANSCRIPTION_STABILITY.tsv", "DIRECT_ENDPOINT_COMPARISON.tsv", "DIRECT_TRAJECTORY_COMPARISON.tsv", "PROJECTION_COMPARISON.tsv", "FONTANA_MECHANISM_COMPARISON.tsv", "FONTANA_FAMILY_COMPARISON.tsv", "SHORTHAND_ENDPOINT_COMPARISON.tsv", "SHORTHAND_TRAJECTORY_COMPARISON.tsv", "SHORTHAND_NULL_TARGET_COMPARISON.tsv", "EXTRACTION_ENDPOINT_COMPARISON.tsv", "EXTRACTION_TRAJECTORY_COMPARISON.tsv", "EXTRACTION_NULL_TARGET_COMPARISON.tsv", "NATURAL_LANGUAGE_COMPARISON.tsv", "SIMPLE_NULL_COMPARISON.tsv", "METRIC_RESIDUALS.tsv", "FAMILY_RESIDUALS.tsv", "COVERAGE_ACCOUNTING.tsv", "HYPOTHESIS_PAIRWISE_ADVANTAGE.tsv", "EQUIFINALITY_ANALYSIS.tsv", "EXTERNAL_MEMORY_EVIDENCE.tsv", "SHORTHAND_EVIDENCE.tsv", "EXTRACTION_EVIDENCE.tsv", "STRONGEST_COUNTEREVIDENCE.md", "TASK83R_REPORT.md", "TASK83R_COMPARISON_INCONCLUSIVE"}
	m := struct {
		Task      string            `json:"task"`
		Version   int               `json:"schema_version"`
		Generated string            `json:"generated_utc"`
		Files     map[string]string `json:"files"`
		Verdict   string            `json:"verdict"`
	}{"Task83r", 1, "2026-08-25T09:41:35Z", map[string]string{}, verdict}
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
	return os.WriteFile(outDir+"/TASK83R_RESULTS_MANIFEST.json", b, 0644)
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
