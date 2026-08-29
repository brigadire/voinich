package g1v2science

import (
	"fmt"
	"math"
	"sort"
	"strconv"
	"strings"
)

type Evidence struct {
	SchemaID                  string         `json:"schema_id"`
	ContractVersion           string         `json:"contract_version"`
	StatusReachabilityVersion string         `json:"status_reachability_version"`
	JobID                     string         `json:"job_id"`
	Stage                     string         `json:"stage"`
	ProducerScope             string         `json:"producer_scope"`
	Status                    string         `json:"status"`
	Dependencies              []string       `json:"dependencies"`
	Payload                   map[string]any `json:"payload"`
	ContentSHA256             string         `json:"content_sha256"`
}

func NewEvidence(reg map[string]string, typ, job, stage, scope, status string, deps []string, payload map[string]any) (Evidence, error) {
	if reg[typ] == "" || !has(EvidenceTypes, typ) {
		return Evidence{}, fmt.Errorf("evidence type")
	}
	if !has(Stages, stage) || !has(Statuses, status) {
		return Evidence{}, fmt.Errorf("stage/status")
	}
	if len(job) != 42 || !strings.HasPrefix(job, "j-") {
		return Evidence{}, fmt.Errorf("E3 JobID")
	}
	d := append([]string{}, deps...)
	sort.Strings(d)
	e := Evidence{reg[typ], ContractVersion, StatusVersion, job, stage, scope, status, d, payload, ""}
	b, er := CanonicalJSON(e)
	if er != nil {
		return e, er
	}
	e.ContentSHA256 = Hash(b)
	return e, nil
}
func (e Evidence) ValidateClosure(reg map[string]string) error {
	known := false
	for _, x := range reg {
		if x == e.SchemaID {
			known = true
		}
	}
	if !known || e.ContractVersion != ContractVersion || e.StatusReachabilityVersion != StatusVersion || !has(Stages, e.Stage) || !has(Statuses, e.Status) {
		return fmt.Errorf("authority mismatch")
	}
	want := e.ContentSHA256
	e.ContentSHA256 = ""
	b, er := CanonicalJSON(e)
	if er != nil {
		return er
	}
	if Hash(b) != want {
		return fmt.Errorf("content hash")
	}
	return nil
}

type DependencyStatus struct {
	JobID  string `json:"job_id"`
	Stage  string `json:"stage"`
	Status string `json:"status"`
}

var rank = map[string]int{"PROTOCOL_VETO": 0, "FIT_FAILURE": 1, "NUMERICAL_FAILURE": 1, "INDUCTION_CAP": 1, "GENERATION_FAILURE": 1, "NOT_ASSESSABLE": 2, "FAIL": 3, "NOT_REACHED": 4}

func Reachability(deps []DependencyStatus) (bool, *DependencyStatus, error) {
	x := append([]DependencyStatus{}, deps...)
	sort.Slice(x, func(i, j int) bool {
		a, ok := rank[x[i].Status]
		if !ok {
			a = 99
		}
		b, ok := rank[x[j].Status]
		if !ok {
			b = 99
		}
		if a == b {
			return x[i].JobID < x[j].JobID
		}
		return a < b
	})
	for _, d := range x {
		if _, ok := rank[d.Status]; ok {
			return false, &d, nil
		}
		if !has([]string{"PASS", "FIT_SUCCESS", "GENERATION_SUCCESS", "COMPLEXITY_SUCCESS", "AGGREGATION_SUCCESS"}, d.Status) {
			return false, nil, fmt.Errorf("unknown upstream %s", d.Status)
		}
	}
	return true, nil, nil
}

type CandidateAssessment struct {
	CandidateID       string  `json:"candidate_id"`
	ModelClass        string  `json:"model_class"`
	Predictive        string  `json:"predictive"`
	Structural        string  `json:"structural"`
	Procedure         string  `json:"procedure"`
	ComplexityBits    int     `json:"complexity_bits"`
	DescriptionLength float64 `json:"description_length"`
	EvidenceComplete  bool    `json:"evidence_complete"`
}

func AssessCandidate(x CandidateAssessment) string {
	if x.Procedure == "PROTOCOL_VETO" {
		return "PROTOCOL_INVALID"
	}
	if x.Procedure != "SUCCESS" || !x.EvidenceComplete || x.Predictive == "NOT_ASSESSABLE" || x.Structural == "NOT_ASSESSABLE" {
		return "UNRESOLVED"
	}
	if x.Predictive == "PASS" && x.Structural == "PASS" {
		return "ADEQUATE"
	}
	if x.Predictive == "FAIL" || x.Structural == "FAIL" {
		return "INADEQUATE"
	}
	return "UNRESOLVED"
}

type FinalDecision struct {
	Verdict           string   `json:"verdict"`
	Detail            string   `json:"identifiability_detail"`
	SelectedClass     string   `json:"selected_class,omitempty"`
	EquivalentClasses []string `json:"equivalent_classes,omitempty"`
}

func FinalVerdict(xs []CandidateAssessment, deltas map[string]float64) (FinalDecision, error) {
	classes := map[string][]string{}
	for _, x := range xs {
		classes[x.ModelClass] = append(classes[x.ModelClass], AssessCandidate(x))
		if AssessCandidate(x) == "PROTOCOL_INVALID" {
			return FinalDecision{"PROTOCOL_INVALID", "PROTOCOL_INVALID", "", nil}, nil
		}
	}
	adequate := []CandidateAssessment{}
	for _, x := range xs {
		if AssessCandidate(x) == "ADEQUATE" {
			adequate = append(adequate, x)
		}
	}
	if len(adequate) == 0 {
		complete := true
		for _, m := range []string{"M0", "M1", "M2", "M3", "M4", "M5"} {
			v := classes[m]
			if len(v) == 0 {
				complete = false
			}
			for _, x := range v {
				if x != "INADEQUATE" {
					complete = false
				}
			}
		}
		if complete {
			return FinalDecision{"NONE", "NONE_COMPLETE_REJECTION", "", nil}, nil
		}
		return FinalDecision{"NOT_IDENTIFIABLE", "MISSING_EVIDENCE", "", nil}, nil
	}
	sort.Slice(adequate, func(i, j int) bool {
		if adequate[i].DescriptionLength == adequate[j].DescriptionLength {
			return adequate[i].CandidateID < adequate[j].CandidateID
		}
		return adequate[i].DescriptionLength < adequate[j].DescriptionLength
	})
	best := adequate[0]
	rankn := int(best.ModelClass[1] - '0')
	for i := 0; i < rankn; i++ {
		m := fmt.Sprintf("M%d", i)
		for _, s := range classes[m] {
			if s != "INADEQUATE" {
				return FinalDecision{"NOT_IDENTIFIABLE", "ORDER_ONLY", "", nil}, nil
			}
		}
	}
	eq := []string{best.ModelClass}
	for _, x := range adequate[1:] {
		q := math.Max(1, deltas[best.CandidateID+"\x00"+x.CandidateID])
		if math.Abs(x.DescriptionLength-best.DescriptionLength) <= q && !has(eq, x.ModelClass) {
			eq = append(eq, x.ModelClass)
		}
	}
	sort.Strings(eq)
	if len(eq) > 1 {
		return FinalDecision{"NOT_IDENTIFIABLE", "EQUIVALENT_SET", best.ModelClass, eq}, nil
	}
	return FinalDecision{"RECOVERED_" + best.ModelClass, "UNIQUE_MINIMUM", best.ModelClass, eq}, nil
}

type WorkRequest struct {
	Job                  JobIdentity           `json:"job"`
	Candidate            Candidate             `json:"candidate"`
	Corpus               Corpus                `json:"corpus"`
	Fitted               *FittedModel          `json:"fitted,omitempty"`
	Baselines            []FittedModel         `json:"baselines,omitempty"`
	GeneratorAuthor      string                `json:"generator_author,omitempty"`
	TokenCount           int                   `json:"token_count,omitempty"`
	Thresholds           map[string]float64    `json:"thresholds,omitempty"`
	DependencyStatuses   []DependencyStatus    `json:"dependency_statuses,omitempty"`
	CandidateAssessments []CandidateAssessment `json:"candidate_assessments,omitempty"`
	DevelopmentDeltas    map[string]float64    `json:"development_deltas,omitempty"`
}
type WorkResult struct {
	JobID      string              `json:"job_id"`
	Stage      string              `json:"stage"`
	Status     string              `json:"status"`
	Model      *FittedModel        `json:"model,omitempty"`
	Generated  *GenerationResult   `json:"generated,omitempty"`
	PM         map[string]PMResult `json:"pm,omitempty"`
	F2         map[string]F2Value  `json:"f2,omitempty"`
	Complexity *Complexity         `json:"complexity,omitempty"`
	Final      *FinalDecision      `json:"final,omitempty"`
	Evidence   []Evidence          `json:"evidence"`
}

func Execute(a Authority, w WorkRequest) (WorkResult, error) {
	id, e := E3JobID(w.Job)
	if e != nil {
		return WorkResult{}, e
	}
	res := WorkResult{JobID: id, Stage: w.Job.Stage, PM: map[string]PMResult{}}
	run, cause, e := Reachability(w.DependencyStatuses)
	if e != nil {
		return res, e
	}
	if !run && w.Job.Stage != "CANDIDATE_AGGREGATION" && w.Job.Stage != "CONTROL_AGGREGATION" {
		reasonCode := "SUPPRESS_" + w.Job.Stage + "_AFTER_" + cause.Stage + "_" + cause.Status
		ev, er := NewEvidence(a.SchemaTypes, "not_reached", id, w.Job.Stage, "DAG_MATERIALIZER", "NOT_REACHED", w.Job.DependencyJobIDs, map[string]any{"upstream_stage": cause.Stage, "upstream_status": cause.Status, "reason_code": reasonCode, "selected_causal_job_id": cause.JobID, "causal_dependency_ids": w.Job.DependencyJobIDs})
		if er != nil {
			return res, er
		}
		res.Status = "NOT_REACHED"
		res.Evidence = []Evidence{ev}
		return res, nil
	}
	dev, _, held := Split(w.Corpus)
	switch w.Job.Stage {
	case "FIT":
		m, status, er := FitCandidate(w.Candidate, dev)
		res.Status = status
		if status == "FIT_SUCCESS" {
			res.Model = &m
			fit, _ := NewEvidence(a.SchemaTypes, "fit", id, "FIT", "JOB", status, w.Job.DependencyJobIDs, map[string]any{"candidate_id": w.Candidate.ID, "control_instance_id": w.Job.ControlInstanceID, "fit_diagnostics_sha256": Hash([]byte(status))})
			fm, _ := NewEvidence(a.SchemaTypes, "fitted_model", id, "FIT", "JOB", status, w.Job.DependencyJobIDs, map[string]any{"candidate_id": w.Candidate.ID, "control_instance_id": w.Job.ControlInstanceID, "model_representation_sha256": m.SerializationSHA256})
			res.Evidence = []Evidence{fit, fm}
		} else {
			res.Evidence = []Evidence{failure(a, id, "FIT", status, w)}
		}
		if er != nil && status == "PROTOCOL_VETO" {
			return res, er
		}
	case "PREDICTIVE":
		if w.Fitted == nil {
			return res, fmt.Errorf("fitted required")
		}
		p1 := PM1(*w.Fitted, held)
		p4 := PM4(*w.Fitted, dev, held)
		p5 := PM5(*w.Fitted, held)
		res.PM["PM1"], res.PM["PM4"], res.PM["PM5"] = p1, p4, p5
		r2, _ := NewRNG("g1v2/pm2", 0)
		for k, x := range PM2(*w.Fitted, w.Baselines, dev, held, 2000, &r2) {
			res.PM["PM2/"+k] = x
		}
		r6, _ := NewRNG("g1v2/pm6", 0)
		res.PM["PM6"] = PM6(*w.Fitted, dev, held, 2000, &r6)
		gates := map[string]string{"PM2": "NOT_ASSESSABLE", "PM5": p5.Status, "PM6": res.PM["PM6"].Status}
		for k, x := range res.PM {
			if strings.HasPrefix(k, "PM2/") {
				gates["PM2"] = x.Status
			}
		}
		res.Status = predictive(gates)
		keys := make([]string, 0, len(res.PM))
		for k := range res.PM {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, k := range keys {
			x := res.PM[k]
			status := x.Status
			if status != "PASS" && status != "FAIL" {
				status = "NOT_ASSESSABLE"
			}
			p := map[string]any{"candidate_id": w.Candidate.ID, "control_instance_id": w.Job.ControlInstanceID, "metric_id": x.ID}
			if status == "NOT_ASSESSABLE" {
				p["reason_code"] = reason(x.Reason)
			} else {
				p["value"] = decimal(x.Value)
				p["threshold"] = decimal(x.Threshold)
			}
			ev, _ := NewEvidence(a.SchemaTypes, "predictive_metric", id, "PREDICTIVE", "JOB", status, w.Job.DependencyJobIDs, p)
			res.Evidence = append(res.Evidence, ev)
		}
		gateStatus := res.Status
		gp := map[string]any{"candidate_id": w.Candidate.ID, "control_instance_id": w.Job.ControlInstanceID, "baseline_id": "FROZEN_BASELINE", "metric_id": "PM2_PM5_PM6", "threshold_id": "FROZEN_DEVELOPMENT_THRESHOLD"}
		if gateStatus == "NOT_ASSESSABLE" {
			gp["reason_code"] = "REQUIRED_METRIC_NOT_ASSESSABLE"
		} else {
			gp["value"] = "1"
			gp["threshold"] = "1"
		}
		ge, _ := NewEvidence(a.SchemaTypes, "predictive_gate", id, "PREDICTIVE", "JOB", gateStatus, w.Job.DependencyJobIDs, gp)
		res.Evidence = append(res.Evidence, ge)
		vp := map[string]any{"candidate_id": w.Candidate.ID, "control_instance_id": w.Job.ControlInstanceID, "pm2_status": gates["PM2"], "pm5_status": gates["PM5"], "pm6_status": gates["PM6"]}
		if gateStatus == "NOT_ASSESSABLE" {
			vp["reason_code"] = "REQUIRED_METRIC_NOT_ASSESSABLE"
		}
		ve, _ := NewEvidence(a.SchemaTypes, "predictive_verdict", id, "PREDICTIVE", "JOB", gateStatus, w.Job.DependencyJobIDs, vp)
		res.Evidence = append(res.Evidence, ve)
	case "GENERATION":
		if w.Fitted == nil {
			return res, fmt.Errorf("fitted required")
		}
		rng, _ := NewRNG("g1v2/generate", 0, 0, 0, 0)
		g, er := GenerateFitted(*w.Fitted, w.GeneratorAuthor, w.TokenCount, &rng)
		res.Status = g.Status
		res.Generated = &g
		if er == nil {
			ev, _ := NewEvidence(a.SchemaTypes, "generation", id, "GENERATION", "JOB", g.Status, w.Job.DependencyJobIDs, map[string]any{"candidate_id": w.Candidate.ID, "control_instance_id": w.Job.ControlInstanceID, "scale": ival(w.Job.ScaleOrNull), "replicate": ival(w.Job.ReplicateOrNull), "corpus_sha256": g.CorpusSHA256})
			res.Evidence = []Evidence{ev}
		} else {
			res.Evidence = []Evidence{failure(a, id, "GENERATION", g.Status, w)}
		}
	case "F2_METRIC":
		vals, er := F2Metrics(w.Corpus)
		if er != nil {
			return res, er
		}
		res.F2 = vals
		metric := sval(w.Job.MetricIDOrNull)
		x, ok := vals[metric]
		if !ok {
			return res, fmt.Errorf("metric")
		}
		threshold := w.Thresholds[metric]
		status := "PASS"
		if x.ScientificWeight == 1 && math.Abs(x.Value) > threshold {
			status = "FAIL"
		}
		res.Status = status
		base := map[string]any{"candidate_id": w.Candidate.ID, "control_instance_id": w.Job.ControlInstanceID, "metric_id": metric, "scale": ival(w.Job.ScaleOrNull), "replicate": ival(w.Job.ReplicateOrNull), "value": decimal(x.Value), "threshold": decimal(threshold)}
		ev, _ := NewEvidence(a.SchemaTypes, "f2_metric", id, "F2_METRIC", "JOB", status, w.Job.DependencyJobIDs, base)
		res.Evidence = []Evidence{ev}
	case "COMPLEXITY":
		if w.Fitted == nil {
			return res, fmt.Errorf("fitted required")
		}
		c := ModelComplexity(*w.Fitted)
		res.Status = "COMPLEXITY_SUCCESS"
		res.Complexity = &c
		ev, _ := NewEvidence(a.SchemaTypes, "complexity", id, "COMPLEXITY", "JOB", res.Status, w.Job.DependencyJobIDs, map[string]any{"candidate_id": w.Candidate.ID, "control_instance_id": w.Job.ControlInstanceID, "structure_bits": c.StructureBits, "parameter_bits": c.ParameterBits, "total_bits": c.TotalBits})
		res.Evidence = []Evidence{ev}
	case "CANDIDATE_AGGREGATION":
		res.Status = "AGGREGATION_SUCCESS"
		verdict := "UNRESOLVED"
		structural := "NOT_ASSESSABLE"
		if len(w.CandidateAssessments) > 0 {
			verdict = AssessCandidate(w.CandidateAssessments[0])
			structural = w.CandidateAssessments[0].Structural
		}
		if !has([]string{"PASS", "FAIL", "NOT_ASSESSABLE"}, structural) {
			structural = "NOT_ASSESSABLE"
		}
		for _, family := range []string{"EDIT", "LEXICAL_PARADIGM"} {
			fp := map[string]any{"candidate_id": w.Candidate.ID, "control_instance_id": w.Job.ControlInstanceID, "family": family, "member_statuses": []string{structural}, "scale": 32000}
			gp := map[string]any{"assessable_count": btoi(structural != "NOT_ASSESSABLE"), "candidate_id": w.Candidate.ID, "control_instance_id": w.Job.ControlInstanceID, "family": family, "pass_count": btoi(structural == "PASS"), "scale": 32000}
			if structural == "NOT_ASSESSABLE" {
				fp["reason_code"] = "REQUIRED_METRIC_NOT_ASSESSABLE"
				gp["reason_code"] = "REQUIRED_METRIC_NOT_ASSESSABLE"
			}
			sf, _ := NewEvidence(a.SchemaTypes, "structural_family", id, "CANDIDATE_AGGREGATION", "SUBOPERATION", structural, w.Job.DependencyJobIDs, fp)
			sg, _ := NewEvidence(a.SchemaTypes, "structural_gate", id, "CANDIDATE_AGGREGATION", "SUBOPERATION", structural, w.Job.DependencyJobIDs, gp)
			res.Evidence = append(res.Evidence, sf, sg)
		}
		svp := map[string]any{"candidate_id": w.Candidate.ID, "control_instance_id": w.Job.ControlInstanceID, "scale_statuses": []string{structural, structural, structural}}
		if structural == "NOT_ASSESSABLE" {
			svp["reason_code"] = "REQUIRED_METRIC_NOT_ASSESSABLE"
		}
		sv, _ := NewEvidence(a.SchemaTypes, "structural_verdict", id, "CANDIDATE_AGGREGATION", "SUBOPERATION", structural, w.Job.DependencyJobIDs, svp)
		ev, _ := NewEvidence(a.SchemaTypes, "minimality", id, "CANDIDATE_AGGREGATION", "JOB", res.Status, w.Job.DependencyJobIDs, map[string]any{"candidate_id": w.Candidate.ID, "control_instance_id": w.Job.ControlInstanceID, "candidate_verdict": verdict, "eligible_candidates": []string{}, "equivalence_components": []any{}})
		res.Evidence = append(res.Evidence, sv, ev)
	case "CONTROL_AGGREGATION":
		d, er := FinalVerdict(w.CandidateAssessments, w.DevelopmentDeltas)
		if er != nil {
			return res, er
		}
		res.Final = &d
		res.Status = "AGGREGATION_SUCCESS"
		ev, _ := NewEvidence(a.SchemaTypes, "final_verdict", id, "CONTROL_AGGREGATION", "JOB", res.Status, w.Job.DependencyJobIDs, map[string]any{"control_instance_id": w.Job.ControlInstanceID, "verdict": d.Verdict, "identifiability_detail": d.Detail})
		res.Evidence = []Evidence{ev}
	default:
		return res, fmt.Errorf("stage")
	}
	for _, ev := range res.Evidence {
		if ev.SchemaID == "" {
			return res, fmt.Errorf("evidence construction failed")
		}
	}
	return res, nil
}

func failure(a Authority, id, stage, status string, w WorkRequest) Evidence {
	if !has([]string{"FIT_FAILURE", "NUMERICAL_FAILURE", "INDUCTION_CAP", "GENERATION_FAILURE", "PROTOCOL_VETO"}, status) {
		status = "PROTOCOL_VETO"
	}
	e, _ := NewEvidence(a.SchemaTypes, "scientific_failure", id, stage, "JOB", status, w.Job.DependencyJobIDs, map[string]any{"reason_code": map[string]string{"FIT_FAILURE": "FIT_DID_NOT_PRODUCE_MODEL", "NUMERICAL_FAILURE": "PRESCRIBED_NUMERIC_OPERATION_FAILED", "INDUCTION_CAP": "M3_ENUMERATION_CAP_REACHED", "GENERATION_FAILURE": "GENERATION_CAP_OR_RETRY_EXHAUSTED", "PROTOCOL_VETO": "PROTOCOL_CHAIN_INVALID"}[status], "diagnostics_hash": Hash([]byte(stage + "/" + status)), "control_instance_id": w.Job.ControlInstanceID, "candidate_id": w.Candidate.ID})
	return e
}
func predictive(g map[string]string) string {
	for _, k := range []string{"PM2", "PM5", "PM6"} {
		if g[k] == "NOT_ASSESSABLE" || g[k] == "NUMERICAL_FAILURE" {
			return "NOT_ASSESSABLE"
		}
	}
	for _, k := range []string{"PM2", "PM5", "PM6"} {
		if g[k] == "FAIL" {
			return "FAIL"
		}
	}
	return "PASS"
}
func decimal(x float64) string {
	if x == 0 {
		return "0"
	}
	return strconv.FormatFloat(x, 'g', -1, 64)
}
func reason(x string) string {
	if x == "" {
		return "REQUIRED_METRIC_NOT_ASSESSABLE"
	}
	return x
}
func ival(x *int) int {
	if x == nil {
		return 0
	}
	return *x
}
func sval(x *string) string {
	if x == nil {
		return ""
	}
	return *x
}
func btoi(x bool) int {
	if x {
		return 1
	}
	return 0
}
