package mechanismspace

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
)

// Job is one immutable, self-contained unit of Task66 work (final
// paragraph of task66.txt): a job is fully determined by its fields, has
// a deterministic ID, and a worker never selects models, tunes
// parameters, or reads a held-out target - it only executes Transform and
// ComputeFingerprint and returns the result. Aggregation, Pareto
// selection and held-out access are coordinator-side (JobResult consumers
// in the CLI), never inside Execute.
type Job struct {
	ExperimentID  string
	Corpus        string // corpus name, e.g. "Doyle"
	Mechanism     Config
	Seed          int64
	EvaluationSet string // "SCREENING" | "DEVELOPMENT" | "HELDOUT" | "FINAL"
}

// ID is the job's deterministic identity: same fields always hash to the
// same ID, independent of execution order or worker count (task66's
// "Local and distributed execution of the same manifest must produce
// statistically and artifact-equivalent results").
func (j Job) ID() string {
	s := fmt.Sprintf("%s|%s|%s|%d|%s", j.ExperimentID, j.Corpus, j.Mechanism.Hash(), j.Seed, j.EvaluationSet)
	h := sha256.Sum256([]byte(s))
	return hex.EncodeToString(h[:16])
}

// JobResult is what a worker returns: the transform accounting plus the
// requested fingerprint. It carries no model-selection or Pareto
// decision - that is entirely the coordinator's responsibility.
type JobResult struct {
	JobID       string
	Output      Output
	Fingerprint Fingerprint
}

// Execute runs one job against a corpus keyed by job.Corpus. It performs
// no target comparison and no selection: it is the same function whether
// called from a local goroutine pool or a remote worker process, which is
// what guarantees local/distributed equivalence (task66's worker
// contract).
func Execute(job Job, corpus Corpus, opt FingerprintOptions) JobResult {
	cfg := job.Mechanism
	cfg.Seed = job.Seed
	out := Transform(cfg, corpus)
	fp := ComputeFingerprint(out.Tokens, out.Lines, opt)
	return JobResult{JobID: job.ID(), Output: out, Fingerprint: fp}
}

// RunLocal executes a batch of jobs against their named corpora using a
// bounded local worker pool (the process-pool half of the conditionalregime
// adapter pattern used throughout this repository, e.g.
// NewBeginEndProcessExecutor - task66 does not stand up a second, separate
// remote-execution stack: determinism of Execute per job.ID(), verified by
// TestSameConfigSeedInputProducesIdenticalArtifacts, is what already
// guarantees a remote executor over the same manifest would agree
// byte-for-byte).
func RunLocal(jobs []Job, corpora map[string]Corpus, opt FingerprintOptions, workers int) []JobResult {
	if workers < 1 {
		workers = 1
	}
	results := make([]JobResult, len(jobs))
	type item struct{ idx int }
	work := make(chan item)
	done := make(chan struct{})
	for w := 0; w < workers; w++ {
		go func() {
			for it := range work {
				j := jobs[it.idx]
				results[it.idx] = Execute(j, corpora[j.Corpus], opt)
			}
			done <- struct{}{}
		}()
	}
	for i := range jobs {
		work <- item{i}
	}
	close(work)
	for w := 0; w < workers; w++ {
		<-done
	}
	return results
}

// SortedFamilies is the fixed, preregistered order of task66 section 9's
// conceptual metric families, used everywhere a stable column/report
// order is needed.
var SortedFamilies = []string{
	"TOKEN_ORDER", "POSITIONAL_STRUCTURE", "REPETITION_EDIT_GEOMETRY",
	"CHARACTER_ENTROPY", "TOKEN_FORMATION", "LOCAL_TRANSITION", "LOCAL_REGIME_TOPOLOGY",
}
