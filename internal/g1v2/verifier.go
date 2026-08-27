package g1v2

import (
	"encoding/json"
	"fmt"
	"math"
	"sort"
	"strings"
)

type PMRecord struct {
	ID              string   `json:"id"`
	Value           *float64 `json:"value"`
	Baseline        float64  `json:"baseline"`
	Threshold       float64  `json:"threshold"`
	FrozenThreshold float64  `json:"frozen_threshold"`
	Comparison      string   `json:"comparison"`
	Available       bool     `json:"available"`
	Finite          bool     `json:"finite"`
	Gate            string   `json:"gate"`
	ArtifactHash    string   `json:"artifact_hash"`
}

type F2Record struct {
	ID           string      `json:"id"`
	Metric       string      `json:"metric"`
	Family       string      `json:"family"`
	Scale        string      `json:"scale"`
	Replicate    int         `json:"replicate"`
	Distance     *float64    `json:"distance"`
	Threshold    float64     `json:"threshold"`
	Applicable   bool        `json:"applicable"`
	Gate         string      `json:"gate"`
	NotReached   *NotReached `json:"not_reached,omitempty"`
	ArtifactHash string      `json:"artifact_hash"`
}

type NotReached struct {
	Reason         string `json:"reason"`
	UpstreamJob    string `json:"upstream_job"`
	DependencyHash string `json:"dependency_hash"`
}

type Decision struct {
	PredictiveStatus string `json:"predictive_status"`
	StructuralStatus string `json:"structural_status"`
	ModelStatus      string `json:"model_status"`
	MinimalityStatus string `json:"minimality_status"`
}

type DecisionEvidence struct {
	SchemaVersion    string              `json:"schema_version"`
	JobID            string              `json:"job_id"`
	CodeHash         string              `json:"code_hash"`
	ConfigHash       string              `json:"config_hash"`
	Seed             uint64              `json:"seed"`
	DependencyHashes []string            `json:"dependency_hashes"`
	FitStatus        string              `json:"fit_status"`
	PMs              map[string]PMRecord `json:"pms"`
	ExpectedF2       []string            `json:"expected_f2"`
	F2               []F2Record          `json:"f2"`
	ComplexityRank   *int                `json:"complexity_rank"`
	Recorded         Decision            `json:"recorded"`
}

func recordHash(v any) string {
	b, _ := canonicalJSON(v)
	return HashBytes(b)
}

func PMHash(r PMRecord) string { r.ArtifactHash = ""; return recordHash(r) }
func F2Hash(r F2Record) string { r.ArtifactHash = ""; return recordHash(r) }

func verifyPM(r PMRecord) (string, error) {
	if r.ID == "" || r.Threshold != r.FrozenThreshold {
		return "", fmt.Errorf("PM %s threshold differs from frozen value", r.ID)
	}
	if r.ArtifactHash != PMHash(r) {
		return "", fmt.Errorf("PM %s artifact hash mismatch", r.ID)
	}
	gate := "NOT_ASSESSABLE"
	if r.Available && r.Finite && r.Value != nil && !math.IsNaN(*r.Value) && !math.IsInf(*r.Value, 0) {
		switch r.Comparison {
		case "GTE":
			if *r.Value >= r.Threshold {
				gate = "PASS"
			} else {
				gate = "FAIL"
			}
		case "LTE":
			if *r.Value <= r.Threshold {
				gate = "PASS"
			} else {
				gate = "FAIL"
			}
		default:
			return "", fmt.Errorf("PM %s bad comparison", r.ID)
		}
	}
	if gate != r.Gate {
		return "", fmt.Errorf("PM %s gate mismatch", r.ID)
	}
	return gate, nil
}

func Predictive(g map[string]string) (string, error) {
	for _, id := range []string{"PM2", "PM5", "PM6"} {
		if g[id] == "" {
			return "", fmt.Errorf("missing %s gate", id)
		}
		if g[id] != "PASS" && g[id] != "FAIL" && g[id] != "NOT_ASSESSABLE" {
			return "", fmt.Errorf("invalid %s gate", id)
		}
	}
	if g["PM2"] == "NOT_ASSESSABLE" || g["PM5"] == "NOT_ASSESSABLE" || g["PM6"] == "NOT_ASSESSABLE" {
		return "PREDICTIVE_NOT_ASSESSABLE", nil
	}
	if g["PM2"] == "FAIL" || g["PM5"] == "FAIL" || g["PM6"] == "FAIL" {
		return "PREDICTIVE_FAIL", nil
	}
	return "PREDICTIVE_PASS", nil
}

func VerifyDecision(e DecisionEvidence) (Decision, error) {
	if e.SchemaVersion != SchemaVersion || !ValidHash(e.JobID) || !ValidHash(e.CodeHash) || !ValidHash(e.ConfigHash) || !sortedUnique(e.DependencyHashes) {
		return Decision{}, fmt.Errorf("invalid decision evidence closure")
	}
	if e.FitStatus != "FITTED" {
		if !map[string]bool{"TRAINING_FAILED": true, "INDUCTION_LIMIT_REACHED": true, "CONVERGENCE_FAILED": true, "NUMERICALLY_UNSTABLE": true}[e.FitStatus] {
			return Decision{}, fmt.Errorf("invalid fit status")
		}
		// A scientific induction failure must carry explicit reachability for
		// every planned downstream cell; absence is not NOT_REACHED.
		if len(e.ExpectedF2) == 0 || len(e.F2) != len(e.ExpectedF2) {
			return Decision{}, fmt.Errorf("incomplete NOT_REACHED evidence")
		}
		for _, r := range e.F2 {
			if r.NotReached == nil || r.Gate != "NOT_REACHED" || r.NotReached.Reason == "" || !ValidHash(r.NotReached.DependencyHash) || r.ArtifactHash != F2Hash(r) {
				return Decision{}, fmt.Errorf("invalid NOT_REACHED evidence")
			}
		}
		d := Decision{"PREDICTIVE_NOT_ASSESSABLE", "STRUCTURAL_NOT_ASSESSABLE", "MODEL_NOT_IDENTIFIABLE", "NOT_IDENTIFIABLE"}
		if d != e.Recorded {
			return Decision{}, fmt.Errorf("recorded induction-failure decision mismatch")
		}
		return d, nil
	}
	requiredPM := []string{"PM1", "PM2", "PM4", "PM5", "PM6"}
	gates := map[string]string{}
	if len(e.PMs) != len(requiredPM) {
		return Decision{}, fmt.Errorf("PM evidence set incomplete")
	}
	for _, id := range requiredPM {
		r, ok := e.PMs[id]
		if !ok || r.ID != id {
			return Decision{}, fmt.Errorf("missing %s", id)
		}
		g, err := verifyPM(r)
		if err != nil {
			return Decision{}, err
		}
		gates[id] = g
	}
	pred, err := Predictive(gates)
	if err != nil {
		return Decision{}, err
	}
	expected := sortedCopy(e.ExpectedF2)
	if !sortedUnique(expected) {
		return Decision{}, fmt.Errorf("duplicate expected F2")
	}
	actual := make([]string, 0, len(e.F2))
	family := map[string][]string{}
	for _, r := range e.F2 {
		if r.ArtifactHash != F2Hash(r) {
			return Decision{}, fmt.Errorf("F2 %s artifact hash mismatch", r.ID)
		}
		actual = append(actual, r.ID)
		gate := "NOT_ASSESSABLE"
		if r.NotReached != nil {
			if r.NotReached.Reason == "" || r.NotReached.UpstreamJob == "" || !ValidHash(r.NotReached.DependencyHash) {
				return Decision{}, fmt.Errorf("bad NOT_REACHED record")
			}
			gate = "NOT_REACHED"
		} else if r.Applicable && r.Distance != nil && !math.IsNaN(*r.Distance) && !math.IsInf(*r.Distance, 0) {
			if *r.Distance <= r.Threshold {
				gate = "PASS"
			} else {
				gate = "FAIL"
			}
		}
		if gate != r.Gate {
			return Decision{}, fmt.Errorf("F2 %s gate mismatch", r.ID)
		}
		family[r.Family] = append(family[r.Family], gate)
	}
	sort.Strings(actual)
	if strings.Join(expected, "\x00") != strings.Join(actual, "\x00") {
		return Decision{}, fmt.Errorf("missing or extra F2 evidence")
	}
	structural := "STRUCTURAL_PASS"
	if len(family) == 0 {
		structural = "STRUCTURAL_NOT_ASSESSABLE"
	}
	for _, gs := range family {
		for _, g := range gs {
			if g == "FAIL" {
				structural = "STRUCTURAL_FAIL"
				break
			}
			if (g == "NOT_ASSESSABLE" || g == "NOT_REACHED") && structural != "STRUCTURAL_FAIL" {
				structural = "STRUCTURAL_NOT_ASSESSABLE"
			}
		}
	}
	model := "MODEL_ADEQUATE"
	if pred == "PREDICTIVE_FAIL" || structural == "STRUCTURAL_FAIL" {
		model = "MODEL_INADEQUATE"
	} else if pred != "PREDICTIVE_PASS" || structural != "STRUCTURAL_PASS" {
		model = "MODEL_NOT_IDENTIFIABLE"
	}
	minimal := "NOT_IDENTIFIABLE"
	if model == "MODEL_ADEQUATE" && e.ComplexityRank != nil {
		minimal = "ORDER_ONLY"
	}
	d := Decision{pred, structural, model, minimal}
	if d != e.Recorded {
		return Decision{}, fmt.Errorf("regenerated decision differs from recorded decision")
	}
	return d, nil
}

// VerifyDecisionForJob binds evidence to the canonical scientific manifest.
// This is the production entry point: self-consistent evidence with a mutated
// seed, code/config hash, or dependency closure must still fail closed.
func VerifyDecisionForJob(j JobBundle, e DecisionEvidence) (Decision, error) {
	if err := j.Validate(); err != nil {
		return Decision{}, err
	}
	if e.JobID != j.JobID || e.CodeHash != j.CodeHash || e.ConfigHash != j.ConfigHash || e.Seed != j.Seed || strings.Join(e.DependencyHashes, "\x00") != strings.Join(j.DependencyHashes, "\x00") {
		return Decision{}, fmt.Errorf("decision evidence differs from frozen manifest closure")
	}
	return VerifyDecision(e)
}

func CanonicalDecision(d Decision) ([]byte, error) { return canonicalJSON(d) }

func ParseDecisionEvidence(b []byte) (DecisionEvidence, error) {
	var e DecisionEvidence
	dec := json.NewDecoder(strings.NewReader(string(b)))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&e); err != nil {
		return e, err
	}
	return e, nil
}
