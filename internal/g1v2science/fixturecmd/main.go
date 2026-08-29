// Command fixturecmd emits disposable OPEN handler evidence for
// schema and evidence-only validation. It never materializes production jobs.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"zcore.dev/voinich/internal/g1v2science"
)

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
	candidate := g1v2science.Candidate{ID: "M0-iid-1", Model: "M0", Route: "iid", Hyper: map[string]any{"alpha": "1"}}
	base := g1v2science.JobIdentity{ContractVersion: g1v2science.ContractVersion, ControlInstanceID: "OPEN-H-V121-EVIDENCE", CandidateID: candidate.ID, DependencyJobIDs: []string{}}
	all := []g1v2science.Evidence{}
	run := func(stage string, mutate func(*g1v2science.WorkRequest)) g1v2science.WorkResult {
		job := base
		job.Stage = stage
		w := g1v2science.WorkRequest{Job: job, Candidate: candidate, Corpus: corpus}
		if mutate != nil {
			mutate(&w)
		}
		r, e := g1v2science.Execute(a, w)
		if e != nil && len(r.Evidence) == 0 {
			check(fmt.Errorf("%s: %w", stage, e))
		}
		all = append(all, r.Evidence...)
		return r
	}
	fit := run("FIT", nil)
	run("PREDICTIVE", func(w *g1v2science.WorkRequest) { w.Fitted = fit.Model })
	scale, replicate := 2000, 0
	run("GENERATION", func(w *g1v2science.WorkRequest) {
		w.Fitted, w.GeneratorAuthor, w.TokenCount = fit.Model, "A", 4
		w.Job.ScaleOrNull, w.Job.ReplicateOrNull = &scale, &replicate
	})
	metric := g1v2science.F2MetricIDs[0]
	run("F2_METRIC", func(w *g1v2science.WorkRequest) {
		w.Job.ScaleOrNull, w.Job.ReplicateOrNull, w.Job.MetricIDOrNull = &scale, &replicate, &metric
		w.Thresholds = map[string]float64{metric: 1}
	})
	run("COMPLEXITY", func(w *g1v2science.WorkRequest) { w.Fitted = fit.Model })
	assessment := g1v2science.CandidateAssessment{CandidateID: candidate.ID, ModelClass: "M0", Predictive: "PASS", Structural: "PASS", Procedure: "SUCCESS", ComplexityBits: 1, DescriptionLength: 1, EvidenceComplete: true}
	run("CANDIDATE_AGGREGATION", func(w *g1v2science.WorkRequest) {
		w.CandidateAssessments = []g1v2science.CandidateAssessment{assessment}
	})
	run("CONTROL_AGGREGATION", func(w *g1v2science.WorkRequest) {
		w.CandidateAssessments = []g1v2science.CandidateAssessment{assessment}
	})
	upstream := "j-0000000000000000000000000000000000000001"
	run("GENERATION", func(w *g1v2science.WorkRequest) {
		w.Job.DependencyJobIDs = []string{upstream}
		w.DependencyStatuses = []g1v2science.DependencyStatus{{JobID: upstream, Stage: "FIT", Status: "FIT_FAILURE"}}
	})
	bad := candidate
	bad.Hyper = map[string]any{"alpha": "not-a-number"}
	run("FIT", func(w *g1v2science.WorkRequest) { w.Candidate = bad })
	enc := json.NewEncoder(os.Stdout)
	enc.SetEscapeHTML(false)
	check(enc.Encode(all))
}

func check(err error) {
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
