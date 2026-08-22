package mechanismspace

import "testing"

// test 25 (batch form): running the same manifest of jobs through the
// local worker pool with different worker counts produces identical,
// order-independent results per job - the property that is supposed to
// guarantee local/distributed equivalence (task66's worker contract).
func TestRunLocalIsWorkerCountIndependent(t *testing.T) {
	c := bigCorpus(150)
	corpora := map[string]Corpus{"big": c}
	var jobs []Job
	for _, fam := range []string{"M0", "M1", "M2", "M4", "M8"} {
		jobs = append(jobs, Job{ExperimentID: "e", Corpus: "big", Mechanism: Config{Family: fam, StateCount: 4, Homophones: 4, InputMode: inputModeFor(fam), Grouping: groupingFor(fam), Seed: 1}, Seed: 1, EvaluationSet: "SCREENING"})
	}
	opt := DefaultScreeningOptions(1)
	r1 := RunLocal(jobs, corpora, opt, 1)
	r4 := RunLocal(jobs, corpora, opt, 4)
	for i := range jobs {
		if r1[i].JobID != r4[i].JobID {
			t.Fatalf("job %d ID differs across worker counts", i)
		}
		if r1[i].Fingerprint.H1 != r4[i].Fingerprint.H1 {
			t.Fatalf("job %d fingerprint differs across worker counts: %v vs %v", i, r1[i].Fingerprint.H1, r4[i].Fingerprint.H1)
		}
	}
}

func inputModeFor(fam string) InputMode {
	if fam == "M8" || fam == "M9" || fam == "M10" || fam == "M11" {
		return Stream
	}
	return WordPreserving
}
func groupingFor(fam string) Grouping {
	if fam == "M8" {
		return FixedGrouping
	}
	return NoGrouping
}

// test: each job's deterministic ID depends on every distinguishing
// field (corpus, mechanism config, seed, evaluation set).
func TestJobIDDistinguishesFields(t *testing.T) {
	base := Job{ExperimentID: "e", Corpus: "Doyle", Mechanism: Config{Family: "M1", Seed: 1}, Seed: 1, EvaluationSet: "SCREENING"}
	variants := []Job{base}
	v := base
	v.Corpus = "Longfellow"
	variants = append(variants, v)
	v = base
	v.Seed = 2
	variants = append(variants, v)
	v = base
	v.EvaluationSet = "DEVELOPMENT"
	variants = append(variants, v)
	v = base
	v.Mechanism.Family = "M2"
	variants = append(variants, v)
	seen := map[string]bool{}
	for _, j := range variants {
		id := j.ID()
		if seen[id] {
			t.Fatalf("two distinct jobs collided on ID %s", id)
		}
		seen[id] = true
	}
}
