// Command goldencmd emits deterministic OPEN scientific vectors.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"zcore.dev/voinich/internal/g1v2science"
)

type candidateGolden struct {
	CandidateID      string                          `json:"candidate_id"`
	ModelClass       string                          `json:"model_class"`
	Route            string                          `json:"route"`
	Status           string                          `json:"status"`
	ModelSHA256      string                          `json:"model_sha256,omitempty"`
	Complexity       *g1v2science.Complexity         `json:"complexity,omitempty"`
	Predictive       map[string]g1v2science.PMResult `json:"predictive,omitempty"`
	GenerationSHA256 map[string]string               `json:"generation_sha256,omitempty"`
}

func main() {
	phase3 := flag.String("phase3", "research/phase3", "phase3 artifact directory")
	flag.Parse()
	a, err := g1v2science.LoadAuthority(
		filepath.Join(*phase3, "task85c-c/registries/G1V2_CANDIDATE_REGISTRY.tsv"),
		filepath.Join(*phase3, "task85c-g/G1V2_GENERATION_SEMANTICS_V1.json"),
		filepath.Join(*phase3, "task85c-c/registries/G1V2_STATUS_REACHABILITY_CONTRACT_V2.json"),
		filepath.Join(*phase3, "task85c-j/G1V2_V1_2_1_EVIDENCE_SCHEMA_REGISTRY.json"),
	)
	check(err)
	corpus, err := g1v2science.NewCorpus([]string{"ab", "a", "ba", "aba", "bab", "aa", "bb", "abba", "baba", "aab", "bba", "abab", "baab", "aaa", "bbb", "abb", "baa", "ababa", "babab", "aaba", "bbab", "abaa", "babb", "aaab", "bbba", "abaaba", "babbab", "aabb", "bbaa", "ababab"})
	check(err)
	dev, _, held := g1v2science.Split(corpus)
	out := struct {
		Schema              string                               `json:"schema"`
		ContractVersion     string                               `json:"contract_version"`
		FixtureClass        string                               `json:"fixture_class"`
		Candidates          []candidateGolden                    `json:"candidates"`
		F2                  map[string]g1v2science.F2Value       `json:"f2"`
		GenerationRoutes    map[string]map[string]any            `json:"generation_routes"`
		AggregationVerdicts map[string]g1v2science.FinalDecision `json:"aggregation_verdicts"`
	}{"g1v2-task85c-h-open-goldens-v1", g1v2science.ContractVersion, "NON_PRODUCTION_TEST_FIXTURE", nil, nil, map[string]map[string]any{}, map[string]g1v2science.FinalDecision{}}
	for i, c := range a.Candidates {
		m, status, fitErr := g1v2science.FitCandidate(c, corpus)
		g := candidateGolden{CandidateID: c.ID, ModelClass: c.Model, Route: c.Route, Status: status}
		if status == "FIT_SUCCESS" {
			cx := g1v2science.ModelComplexity(m)
			g.ModelSHA256, g.Complexity = m.SerializationSHA256, &cx
			g.Predictive = map[string]g1v2science.PMResult{"PM1": g1v2science.PM1(m, held), "PM4": g1v2science.PM4(m, dev, held), "PM5": g1v2science.PM5(m, held)}
			r6, _ := g1v2science.NewRNG("g1v2/pm6", uint64(i))
			g.Predictive["PM6"] = g1v2science.PM6(m, dev, held, 2000, &r6)
			g.GenerationSHA256 = map[string]string{}
			for ai, author := range []string{"A", "B"} {
				r, _ := g1v2science.NewRNG("g1v2/generate", uint64(i), uint64(ai), 0, 0)
				x, e := g1v2science.GenerateFitted(m, author, 4, &r)
				if e == nil {
					g.GenerationSHA256[author] = x.CorpusSHA256
				} else {
					g.GenerationSHA256[author] = x.Status
				}
			}
		} else if fitErr == nil {
			check(fmt.Errorf("%s returned %s without error", c.ID, status))
		}
		out.Candidates = append(out.Candidates, g)
	}
	out.F2, err = g1v2science.F2Metrics(corpus)
	check(err)
	for _, route := range a.Routes {
		r, _ := g1v2science.NewRNG("g1v2/control/generate", uint64(route.Index), 0, 0)
		g, e := g1v2science.GenerateSynthetic(route, 8, &r)
		check(e)
		out.GenerationRoutes[route.ID] = map[string]any{"status": g.Status, "corpus_sha256": g.CorpusSHA256, "draws": g.Draws}
	}
	fixtures := map[string][]g1v2science.CandidateAssessment{
		"singleton":         {{CandidateID: "M0-x", ModelClass: "M0", Predictive: "PASS", Structural: "PASS", Procedure: "SUCCESS", DescriptionLength: 10, EvidenceComplete: true}},
		"equivalent":        {{CandidateID: "M0-x", ModelClass: "M0", Predictive: "PASS", Structural: "PASS", Procedure: "SUCCESS", DescriptionLength: 10, EvidenceComplete: true}, {CandidateID: "M1-x", ModelClass: "M1", Predictive: "PASS", Structural: "PASS", Procedure: "SUCCESS", DescriptionLength: 10, EvidenceComplete: true}},
		"none":              {{CandidateID: "M0-x", ModelClass: "M0", Predictive: "FAIL", Structural: "PASS", Procedure: "SUCCESS", EvidenceComplete: true}, {CandidateID: "M1-x", ModelClass: "M1", Predictive: "FAIL", Structural: "PASS", Procedure: "SUCCESS", EvidenceComplete: true}, {CandidateID: "M2-x", ModelClass: "M2", Predictive: "FAIL", Structural: "PASS", Procedure: "SUCCESS", EvidenceComplete: true}, {CandidateID: "M3-x", ModelClass: "M3", Predictive: "FAIL", Structural: "PASS", Procedure: "SUCCESS", EvidenceComplete: true}, {CandidateID: "M4-x", ModelClass: "M4", Predictive: "FAIL", Structural: "PASS", Procedure: "SUCCESS", EvidenceComplete: true}, {CandidateID: "M5-x", ModelClass: "M5", Predictive: "FAIL", Structural: "PASS", Procedure: "SUCCESS", EvidenceComplete: true}},
		"missing":           {{CandidateID: "M0-x", ModelClass: "M0", Predictive: "NOT_ASSESSABLE", Structural: "PASS", Procedure: "SUCCESS", EvidenceComplete: false}},
		"not_assessable":    {{CandidateID: "M0-x", ModelClass: "M0", Predictive: "NOT_ASSESSABLE", Structural: "PASS", Procedure: "SUCCESS", EvidenceComplete: true}},
		"procedure_failure": {{CandidateID: "M0-x", ModelClass: "M0", Predictive: "PASS", Structural: "PASS", Procedure: "FIT_FAILURE", EvidenceComplete: true}},
		"veto":              {{CandidateID: "M0-x", ModelClass: "M0", Procedure: "PROTOCOL_VETO"}},
	}
	for name, xs := range fixtures {
		d, e := g1v2science.FinalVerdict(xs, map[string]float64{})
		check(e)
		out.AggregationVerdicts[name] = d
	}
	enc := json.NewEncoder(os.Stdout)
	enc.SetEscapeHTML(false)
	check(enc.Encode(out))
}

func check(err error) {
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
