package g1v2science

import (
	"fmt"
)

// EvidenceOnlyResult is reconstructed without fitting, prediction, generation,
// or model state. It is deliberately limited to frozen evidence fields.
type EvidenceOnlyResult struct {
	PredictiveByCandidate map[string]string `json:"predictive_by_candidate"`
	StructuralByCandidate map[string]string `json:"structural_by_candidate"`
	CandidateVerdicts     map[string]string `json:"candidate_verdicts"`
	ComplexityBits        map[string]int    `json:"complexity_bits"`
	FinalVerdict          string            `json:"final_verdict"`
}

func ReconstructEvidenceOnly(reg map[string]string, evidence []Evidence) (EvidenceOnlyResult, error) {
	r := EvidenceOnlyResult{map[string]string{}, map[string]string{}, map[string]string{}, map[string]int{}, ""}
	seen := map[string]string{}
	for _, e := range evidence {
		if err := e.ValidateClosure(reg); err != nil {
			return r, err
		}
		key := e.JobID + "\x00" + e.SchemaID
		if old, ok := seen[key]; ok && old != e.ContentSHA256 {
			return r, fmt.Errorf("conflicting evidence %s", key)
		}
		seen[key] = e.ContentSHA256
		candidate, _ := e.Payload["candidate_id"].(string)
		switch e.SchemaID {
		case reg["predictive_verdict"]:
			want, err := evidencePredictiveStatus(e.Payload)
			if err != nil || want != e.Status {
				return r, fmt.Errorf("predictive verdict mismatch")
			}
			r.PredictiveByCandidate[candidate] = want
		case reg["structural_verdict"]:
			want, err := evidenceStructuralStatus(e.Payload)
			if err != nil || want != e.Status {
				return r, fmt.Errorf("structural verdict mismatch")
			}
			r.StructuralByCandidate[candidate] = want
		case reg["minimality"]:
			v, ok := e.Payload["candidate_verdict"].(string)
			if !ok || !has([]string{"ADEQUATE", "INADEQUATE", "UNRESOLVED", "PROTOCOL_INVALID"}, v) {
				return r, fmt.Errorf("candidate verdict")
			}
			r.CandidateVerdicts[candidate] = v
		case reg["complexity"]:
			n, ok := integerEvidence(e.Payload["total_bits"])
			if !ok || n < 0 {
				return r, fmt.Errorf("complexity")
			}
			r.ComplexityBits[candidate] = n
		case reg["final_verdict"]:
			v, ok := e.Payload["verdict"].(string)
			if !ok || !validFinalVerdict(v) {
				return r, fmt.Errorf("final verdict")
			}
			if r.FinalVerdict != "" && r.FinalVerdict != v {
				return r, fmt.Errorf("conflicting final verdict")
			}
			r.FinalVerdict = v
		}
	}
	for c, v := range r.CandidateVerdicts {
		p, pok := r.PredictiveByCandidate[c]
		s, sok := r.StructuralByCandidate[c]
		if v == "ADEQUATE" && (!pok || !sok || p != "PASS" || s != "PASS") {
			return r, fmt.Errorf("adequacy unsupported for %s", c)
		}
		if v == "INADEQUATE" && pok && sok && p == "PASS" && s == "PASS" {
			return r, fmt.Errorf("inadequacy contradicted for %s", c)
		}
	}
	if r.FinalVerdict == "NONE" {
		if len(r.CandidateVerdicts) == 0 {
			return r, fmt.Errorf("NONE without candidate evidence")
		}
		for _, v := range r.CandidateVerdicts {
			if v != "INADEQUATE" {
				return r, fmt.Errorf("NONE with non-inadequate candidate")
			}
		}
	}
	return r, nil
}

func evidencePredictiveStatus(p map[string]any) (string, error) {
	statuses := make([]string, 0, 3)
	for _, k := range []string{"pm2_status", "pm5_status", "pm6_status"} {
		s, ok := p[k].(string)
		if !ok || !has([]string{"PASS", "FAIL", "NOT_ASSESSABLE"}, s) {
			return "", fmt.Errorf("%s", k)
		}
		statuses = append(statuses, s)
	}
	return gateFromStatuses(statuses), nil
}

func evidenceStructuralStatus(p map[string]any) (string, error) {
	var statuses []string
	switch x := p["scale_statuses"].(type) {
	case []string:
		statuses = append(statuses, x...)
	case []any:
		for _, raw := range x {
			s, ok := raw.(string)
			if !ok {
				return "", fmt.Errorf("scale status")
			}
			statuses = append(statuses, s)
		}
	default:
		return "", fmt.Errorf("scale_statuses")
	}
	for _, s := range statuses {
		if !has([]string{"PASS", "FAIL", "NOT_ASSESSABLE"}, s) {
			return "", fmt.Errorf("scale status")
		}
	}
	if len(statuses) == 0 {
		return "", fmt.Errorf("empty scales")
	}
	return gateFromStatuses(statuses), nil
}

func gateFromStatuses(statuses []string) string {
	for _, s := range statuses {
		if s == "NOT_ASSESSABLE" {
			return "NOT_ASSESSABLE"
		}
	}
	for _, s := range statuses {
		if s == "FAIL" {
			return "FAIL"
		}
	}
	return "PASS"
}

func integerEvidence(v any) (int, bool) {
	switch x := v.(type) {
	case int:
		return x, true
	case float64:
		n := int(x)
		return n, float64(n) == x
	default:
		return 0, false
	}
}

func validFinalVerdict(v string) bool {
	if has([]string{"NONE", "NOT_IDENTIFIABLE", "PROTOCOL_INVALID"}, v) {
		return true
	}
	return len(v) == len("RECOVERED_M0") && v[:len("RECOVERED_M")] == "RECOVERED_M" && v[len(v)-1] >= '0' && v[len(v)-1] <= '5'
}
