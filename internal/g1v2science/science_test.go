package g1v2science

import (
	"encoding/hex"
	"encoding/json"
	"math"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"testing"
)

func authority(t *testing.T) Authority {
	t.Helper()
	r := filepath.Join("..", "..", "research", "phase3")
	a, e := LoadAuthority(filepath.Join(r, "task85c-c", "registries", "G1V2_CANDIDATE_REGISTRY.tsv"), filepath.Join(r, "task85c-g", "G1V2_GENERATION_SEMANTICS_V1.json"), filepath.Join(r, "task85c-c", "registries", "G1V2_STATUS_REACHABILITY_CONTRACT_V2.json"), filepath.Join(r, "task85c-j", "G1V2_V1_2_1_EVIDENCE_SCHEMA_REGISTRY.json"))
	if e != nil {
		t.Fatal(e)
	}
	return a
}
func openCorpus() Corpus {
	c, _ := NewCorpus([]string{"ab", "a", "ba", "aba", "bab", "aa", "bb", "abba", "baba", "aab", "bba", "abab", "baab", "aaa", "bbb", "abb", "baa", "ababa", "babab", "aaba", "bbab", "abaa", "babb", "aaab", "bbba", "abaaba", "babbab", "aabb", "bbaa", "ababab"})
	return c
}
func TestAuthorityClosure(t *testing.T) {
	a := authority(t)
	if len(a.Candidates) != 43 || len(a.Routes) != 12 || a.Transitions != 45 || len(a.SchemaTypes) != 15 {
		t.Fatalf("%d/%d/%d/%d", len(a.Candidates), len(a.Routes), a.Transitions, len(a.SchemaTypes))
	}
	for i, r := range a.Routes {
		if r.ID == "" || r.Index != i {
			t.Fatalf("route %d", i)
		}
	}
}
func TestE3JobIDGolden(t *testing.T) {
	root := filepath.Join("..", "..", "research", "phase3", "task85c-j")
	b, e := json.Marshal(JobIdentity{ContractVersion: ContractVersion, ControlInstanceID: "OPEN-M0-REFREEZE-1", CandidateID: "M0-iid-1", Stage: "FIT", DependencyJobIDs: []string{}})
	if e != nil {
		t.Fatal(e)
	}
	var x JobIdentity
	if e = json.Unmarshal(b, &x); e != nil {
		t.Fatal(e)
	}
	id, e := E3JobID(x)
	if e != nil {
		t.Fatal(e)
	}
	raw, e := filepath.Abs(filepath.Join(root, "G1V2_E3_JOBID_REGRESSION.json"))
	if e != nil {
		t.Fatal(e)
	}
	var f struct {
		V121 struct {
			JobID string `json:"jobid"`
		} `json:"v1_2_1"`
	}
	fb, _ := osRead(raw)
	if e = json.Unmarshal(fb, &f); e != nil {
		t.Fatal(e)
	}
	if id != f.V121.JobID {
		t.Fatalf("%s != %s", id, f.V121.JobID)
	}
}
func osRead(path string) ([]byte, error) { return os.ReadFile(path) }
func TestRNGAndM0V121Goldens(t *testing.T) {
	r, _ := NewRNG("g1v2/generate", 0, 0, 0, 0)
	d := r.Digest()
	if hex.EncodeToString(d[:]) != "21b9eeee669faf5a48467b44bc4f8116351c46ad18193f4286e39a8000d81817" {
		t.Fatal("RNG")
	}
	if math.Abs(r.U53()-.13174336738904791) > 1e-17 {
		t.Fatal("U53")
	}
	c, _ := NewCorpus([]string{"ab", "a"})
	m, s, e := FitCandidate(Candidate{ID: "M0-iid-1", Model: "M0", Route: "iid", Hyper: map[string]any{"alpha": "1"}}, c)
	if e != nil || s != "FIT_SUCCESS" {
		t.Fatal(s, e)
	}
	p := m.Rows[""]
	if p["a"] != 1./3 || p["b"] != 2./9 || p["<UNK>"] != 1./9 || p["<EOS>"] != 1./3 {
		t.Fatalf("%#v", p)
	}
	if math.Abs(Neumaier([]float64{p["a"], p["b"], p["<UNK>"], p["<EOS>"]})-1) > 2e-16 {
		t.Fatal("normalization")
	}
	lp, _ := m.LogProbToken([]string{"z"})
	if math.IsInf(lp, -1) {
		t.Fatal("positive-alpha unseen")
	}
}
func TestAll43Candidates(t *testing.T) {
	a := authority(t)
	c := openCorpus()
	seen := map[string]bool{}
	for _, x := range a.Candidates {
		m, s, e := FitCandidate(x, c)
		if e != nil && s != "FIT_FAILURE" {
			t.Errorf("%s %s %v", x.ID, s, e)
		}
		if !has([]string{"FIT_SUCCESS", "INDUCTION_CAP", "FIT_FAILURE"}, s) {
			t.Errorf("%s %s", x.ID, s)
		}
		if s == "FIT_SUCCESS" && m.ModelClass != x.Model {
			t.Errorf("dispatch %s", x.ID)
		}
		seen[x.Model+"/"+x.Route] = true
	}
	for _, x := range []string{"M0/iid", "M1/markov", "M2/pst", "M3/exact", "M3/approx", "M4/hmm", "M5/grammar"} {
		if !seen[x] {
			t.Error(x)
		}
	}
}
func fittedByRoute(t *testing.T, a Authority) map[string]FittedModel {
	c := openCorpus()
	out := map[string]FittedModel{}
	for _, x := range a.Candidates {
		key := x.Model + "/" + x.Route
		if out[key].ModelClass != "" {
			continue
		}
		m, s, _ := FitCandidate(x, c)
		if s == "FIT_SUCCESS" {
			out[key] = m
		}
	}
	return out
}
func TestGenerationPaths26AndPFSC01(t *testing.T) {
	a := authority(t)
	for _, x := range a.Routes {
		r, _ := NewRNG("g1v2/control/generate", uint64(x.Index), 0, 0)
		g, e := GenerateSynthetic(x, 8, &r)
		if e != nil || g.Status != "GENERATION_SUCCESS" || len(g.Tokens) != 8 {
			t.Errorf("%s %v", x.ID, e)
		}
	}
	models := fittedByRoute(t, a)
	n := 0
	for _, key := range []string{"M0/iid", "M1/markov", "M2/pst", "M3/exact", "M3/approx", "M4/hmm", "M5/grammar"} {
		m := models[key]
		if m.ModelClass == "" {
			t.Fatalf("missing %s", key)
		}
		for _, author := range []string{"A", "B"} {
			r, _ := NewRNG("g1v2/generate", uint64(n), 0, 0, 0)
			g, e := GenerateFitted(m, author, 3, &r)
			if e != nil || g.Status != "GENERATION_SUCCESS" {
				t.Fatalf("%s/%s: %v", key, author, e)
			}
			n++
		}
	}
	if 12+n != 26 {
		t.Fatal("path count")
	}
	x, e := DirectCDF([]string{"a", "b", "c", "d", "<EOS>"}, []float64{.28, .22, .18, .12, .2}, map[string]bool{"a": true, "b": true, "c": true, "d": true}, .92848667210989588)
	if e != nil || x != "d" {
		t.Fatal("PF-SC01")
	}
	for _, tc := range []struct {
		in  []string
		hex string
	}{{[]string{}, ""}, {[]string{"a"}, "610a"}, {[]string{"a", "bb"}, "610a62620a"}, {[]string{"café"}, "636166c3a90a"}} {
		b, e := SerializeCorpus(tc.in)
		if e != nil || hex.EncodeToString(b) != tc.hex {
			t.Fatal("serialization")
		}
	}
}
func TestPMF2ComplexityAggregation(t *testing.T) {
	a := authority(t)
	m := fittedByRoute(t, a)["M0/iid"]
	c := openCorpus()
	dev, _, held := Split(c)
	if PM1(m, held).ID != "PM1" || PM4(m, dev, held).ID != "PM4" || PM5(m, held).ID != "PM5" {
		t.Fatal("PM")
	}
	r, _ := NewRNG("g1v2/pm6", 0)
	if PM6(m, dev, held, 16, &r).ID != "PM6" {
		t.Fatal("PM6")
	}
	if !Holm(map[string]float64{"a": .01, "b": .02}, .05)["a"] {
		t.Fatal("Holm")
	}
	q, _ := QuantileType7([]float64{1, 2, 3, 4}, .95)
	if q != 3.8499999999999996 {
		t.Fatalf("q %.17g", q)
	}
	f, e := F2Metrics(c)
	if e != nil || len(f) != 12 {
		t.Fatal("F2")
	}
	for _, id := range F2MetricIDs {
		if f[id].ID != id {
			t.Error(id)
		}
	}
	for _, m := range fittedByRoute(t, a) {
		x := ModelComplexity(m)
		if x.TotalBits != x.StructureBits+x.ParameterBits || x.TotalBits <= 0 {
			t.Fatal("complexity")
		}
	}
	xs := []CandidateAssessment{{"M0-x", "M0", "FAIL", "PASS", "SUCCESS", 1, 10, true}, {"M1-x", "M1", "PASS", "PASS", "SUCCESS", 2, 20, true}}
	d, e := FinalVerdict(xs, map[string]float64{})
	if e != nil || d.Verdict != "RECOVERED_M1" {
		t.Fatalf("%+v %v", d, e)
	}
}
func TestReachabilityEvidenceAndExecuteDeterminism(t *testing.T) {
	a := authority(t)
	run, c, e := Reachability([]DependencyStatus{{"j-b", "FIT", "FIT_FAILURE"}, {"j-a", "FIT", "PROTOCOL_VETO"}})
	if e != nil || run || c.Status != "PROTOCOL_VETO" {
		t.Fatal("precedence")
	}
	cand := a.Candidates[3]
	cor := openCorpus()
	j := JobIdentity{ContractVersion: ContractVersion, ControlInstanceID: "OPEN-H-FIXTURE", CandidateID: cand.ID, Stage: "FIT", DependencyJobIDs: []string{}}
	w := WorkRequest{Job: j, Candidate: cand, Corpus: cor}
	x, e := Execute(a, w)
	if e != nil || x.Status != "FIT_SUCCESS" {
		t.Fatal(e, x.Status)
	}
	y, e := Execute(a, w)
	if e != nil || !reflect.DeepEqual(x, y) {
		t.Fatal("nondeterminism")
	}
	types := map[string]bool{}
	for _, ev := range x.Evidence {
		if e = ev.ValidateClosure(a.SchemaTypes); e != nil {
			t.Fatal(e)
		}
		types[ev.SchemaID] = true
	}
	if len(types) != 2 {
		t.Fatal("FIT evidence")
	}
}
func TestIndependentGeneratorB(t *testing.T) {
	out := []string{"a", "b", "c"}
	w := []float64{.5, .3, .2}
	r1, _ := NewRNG("independent", 0)
	r2 := r1
	a, _ := DirectCDF(out, w, all(out), r1.U53())
	b, _ := ExponentialRace(out, w, all(out), &r2)
	if r1.Draw != 1 || r2.Draw != 3 {
		t.Fatal("B consumption")
	}
	_ = a
	_ = b
}
func TestCanonicalOrdering(t *testing.T) {
	b, e := CanonicalJSON(map[string]any{"b": 1, "a": "e\u0301"})
	if e != nil || string(b) != "{\"a\":\"é\",\"b\":1}\n" {
		t.Fatalf("%s %v", b, e)
	}
	x := []string{"b", "a"}
	sort.Strings(x)
	if x[0] != "a" {
		t.Fatal()
	}
}

func TestEvidenceOnlyReconstructionAndMutations(t *testing.T) {
	a := authority(t)
	job := "j-00000000000000000000000000000000000000aa"
	candidate := "M0-iid-1"
	pv, _ := NewEvidence(a.SchemaTypes, "predictive_verdict", job, "PREDICTIVE", "JOB", "PASS", nil, map[string]any{"candidate_id": candidate, "control_instance_id": "OPEN-EVIDENCE-ONLY", "pm2_status": "PASS", "pm5_status": "PASS", "pm6_status": "PASS"})
	sv, _ := NewEvidence(a.SchemaTypes, "structural_verdict", job, "CANDIDATE_AGGREGATION", "SUBOPERATION", "PASS", nil, map[string]any{"candidate_id": candidate, "control_instance_id": "OPEN-EVIDENCE-ONLY", "scale_statuses": []string{"PASS", "PASS", "PASS"}})
	cx, _ := NewEvidence(a.SchemaTypes, "complexity", job, "COMPLEXITY", "JOB", "COMPLEXITY_SUCCESS", nil, map[string]any{"candidate_id": candidate, "control_instance_id": "OPEN-EVIDENCE-ONLY", "structure_bits": 4, "parameter_bits": 8, "total_bits": 12})
	mn, _ := NewEvidence(a.SchemaTypes, "minimality", job, "CANDIDATE_AGGREGATION", "JOB", "AGGREGATION_SUCCESS", nil, map[string]any{"candidate_id": candidate, "control_instance_id": "OPEN-EVIDENCE-ONLY", "candidate_verdict": "ADEQUATE", "eligible_candidates": []string{candidate}, "equivalence_components": []any{[]string{candidate}}})
	fv, _ := NewEvidence(a.SchemaTypes, "final_verdict", job, "CONTROL_AGGREGATION", "JOB", "AGGREGATION_SUCCESS", nil, map[string]any{"control_instance_id": "OPEN-EVIDENCE-ONLY", "verdict": "RECOVERED_M0", "identifiability_detail": "UNIQUE_MINIMUM"})
	evidence := []Evidence{pv, sv, cx, mn, fv}
	r, e := ReconstructEvidenceOnly(a.SchemaTypes, evidence)
	if e != nil || r.FinalVerdict != "RECOVERED_M0" || r.ComplexityBits[candidate] != 12 {
		t.Fatalf("%+v %v", r, e)
	}
	for name, mutate := range map[string]func(*Evidence){
		"wrong contract": func(e *Evidence) { e.ContractVersion = "G1_V2_EXECUTABLE_CONTRACT_V1_2" },
		"wrong schema":   func(e *Evidence) { e.SchemaID = "g1v2.unknown.v1_2_1" },
		"illegal status": func(e *Evidence) { e.Status = "SCIENTIFIC_FAILURE" },
		"content":        func(e *Evidence) { e.Payload["pm2_status"] = "FAIL" },
	} {
		t.Run(name, func(t *testing.T) {
			bad := append([]Evidence(nil), evidence...)
			bad[0].Payload = map[string]any{}
			for k, v := range evidence[0].Payload {
				bad[0].Payload[k] = v
			}
			mutate(&bad[0])
			if _, err := ReconstructEvidenceOnly(a.SchemaTypes, bad); err == nil {
				t.Fatal("mutation accepted")
			}
		})
	}
}
