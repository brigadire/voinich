package g1v2

import "sort"

// NewEngineeringManifest creates a fully known, non-confirmatory M0-M5
// execution fixture. Its labels exercise routing only and are not scientific
// implementations or recovery evidence.
func NewEngineeringManifest(codeHash, configHash string, perModel, iterations int) Manifest {
	jobs := []JobBundle{}
	structural := []string{}
	makeJob := func(model, stage string, rep int, deps []string) JobBundle {
		j := JobBundle{JobID: "", Experiment: Experiment, ProtocolVersion: ProtocolVersion, Stage: stage, CorpusID: "open-engineering-fixture", Model: model, Candidate: model + "-engineering", Scale: "engineering-small", Replicate: rep, Seed: uint64(86000 + rep), InputHashes: []string{HashBytes([]byte("synthetic-nonconfirmatory-input-v1"))}, DependencyHashes: []string{}, CodeHash: codeHash, ConfigHash: configHash, OutputSchema: "engineering-output-v1", DependsOn: sortedCopy(deps), Work: WorkSpec{Kind: "sha256-chain-v1", Payload: model + "/" + stage, Iterations: iterations}}
		id, _ := j.ComputedID()
		j.JobID = id
		return j
	}
	for _, model := range []string{"M0", "M1", "M2", "M3", "M4", "M5"} {
		for rep := 0; rep < perModel; rep++ {
			f := makeJob(model, "FIT", rep, nil)
			p := makeJob(model, "PREDICTIVE", rep, []string{f.JobID})
			g := makeJob(model, "GENERATION", rep, []string{p.JobID})
			s := makeJob(model, "STRUCTURAL", rep, []string{g.JobID})
			jobs = append(jobs, f, p, g, s)
			structural = append(structural, s.JobID)
		}
	}
	sort.Strings(structural)
	jobs = append(jobs, makeJob("stage-neutral", "AGGREGATION", 0, structural))
	return Manifest{SchemaVersion, CanonicalVersion, jobs}
}
