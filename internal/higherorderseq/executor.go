package higherorderseq

import (
	"context"
	"fmt"
	"sync"
)

// CandidateExecutor runs one frozen candidate's entire Part A-L computation
// (task44): the unit of independent work here is a whole candidate, never a
// single CMI permutation, because runCMI's permutations for one candidate
// share a single sequentially-advanced *rand.Rand stream
// (cmiWorkspace.permute) - splitting that stream across two jobs would
// change RNG semantics, which task44 forbids. There are only as many jobs
// as frozen candidates (a handful), so this stage's distribution ceiling is
// low by design, not a bug.
type CandidateExecutor interface {
	Run(ctx context.Context, sequence string) (CandidateResult, error)
}

// ComputeCandidate is the one function every backend (in-process default,
// subprocess, or remote worker) calls: Parts A-L's math, copied verbatim
// from RunAndWrite's per-part closures, for exactly one candidate identified
// by its frozen sequence text.
func ComputeCandidate(candidates []Candidate, blocks []Block, lineLength map[string]int, relatives map[string][]string, primaryPermutations int, seed int64, sequence string) (CandidateResult, error) {
	cand, ok := candidateBySequence(candidates, sequence)
	if !ok {
		return CandidateResult{}, fmt.Errorf("higher-order-sequence-validate: unknown candidate sequence %q", sequence)
	}
	var r CandidateResult
	r.Candidate = cand
	r.Occurrences = findOccurrences(cand, blocks, lineLength)
	r.ConditionalRows = conditionalRowsForCandidate(cand, blocks)
	r.CMI = runCMI(cand, blocks, permutationsFor(cand, primaryPermutations), candidateSeed(seed, cand.Sequence))
	r.LOBO = runLOBO(cand, blocks)
	r.ContextControls = contextControlRows(cand, blocks)
	r.ContextRank = contextRankRow(cand, blocks)
	r.Continuations = continuationDistributions(cand, blocks)
	r.ContinuationEnt = continuationEntropy(cand, blocks)
	r.CrossBlock = crossBlockRow(r.ConditionalRows)
	r.Meta = metaAnalysisRow(r.ConditionalRows)
	r.Jackknife = jackknifeRow(cand, blocks, primaryEligible(r.ConditionalRows))
	r.Position = positionRows(cand, r.Occurrences, blocks, lineLength)
	r.BlockPosTVD = positionTVD(r.Position, "block_position_bin")
	r.LinePosTVD = positionTVD(r.Position, "line_position")
	r.StructuralFamily = structuralFamilyRows(cand, blocks, relatives)
	return r, nil
}

func candidateBySequence(candidates []Candidate, sequence string) (Candidate, bool) {
	for _, c := range candidates {
		if c.Sequence == sequence {
			return c, true
		}
	}
	return Candidate{}, false
}

// defaultCandidateExecutor is the in-process CandidateExecutor used whenever
// Config.CandidateExecutor is nil (Config.Executor == "" or "goroutine").
// Every field below is read-only after construction: runCMI allocates its
// own cmiWorkspace fresh per call (never a shared package-level scratch
// buffer), so unlike transitionnetwork/tokenrelationvalidation's profile
// paths, concurrent calls for distinct candidates need no mutex here.
type defaultCandidateExecutor struct {
	candidates          []Candidate
	blocks              []Block
	lineLength          map[string]int
	relatives           map[string][]string
	primaryPermutations int
	seed                int64
}

func newDefaultCandidateExecutor(candidates []Candidate, blocks []Block, lineLength map[string]int, relatives map[string][]string, primaryPermutations int, seed int64) *defaultCandidateExecutor {
	return &defaultCandidateExecutor{candidates: candidates, blocks: blocks, lineLength: lineLength, relatives: relatives, primaryPermutations: primaryPermutations, seed: seed}
}

func (e *defaultCandidateExecutor) Run(_ context.Context, sequence string) (CandidateResult, error) {
	return ComputeCandidate(e.candidates, e.blocks, e.lineLength, e.relatives, e.primaryPermutations, e.seed, sequence)
}

// runCandidateBattery dispatches n independent candidate jobs (identified by
// items[i]) through executor with a bounded worker pool, restoring the
// original ascending items order before invoking onReady for each -
// regardless of the completion order any goroutine/process/remote backend
// delivers results in. No result here depends positionally on another
// candidate's result (each candidate's CandidateResult is self-contained,
// keyed by its own sequence in the checkpoint), so restoring canonical
// order is only about keeping checkpoint writes deterministic and
// resumable, not about any reduction correctness requirement - the same
// single mechanism task44 already reuses in every other distributed stage.
func runCandidateBattery(ctx context.Context, executor CandidateExecutor, items []string, workers int, onReady func(i int, sequence string, r CandidateResult) error) error {
	n := len(items)
	if n == 0 {
		return nil
	}
	if workers < 1 {
		workers = 1
	}
	if workers > n {
		workers = n
	}
	if ctx == nil {
		ctx = context.Background()
	}
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	type done struct {
		i      int
		result CandidateResult
		err    error
	}
	jobs := make(chan int)
	results := make(chan done, workers)
	go func() {
		defer close(jobs)
		for i := 0; i < n; i++ {
			select {
			case jobs <- i:
			case <-ctx.Done():
				return
			}
		}
	}()
	var wg sync.WaitGroup
	for range workers {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := range jobs {
				r, err := executor.Run(ctx, items[i])
				select {
				case results <- done{i, r, err}:
				case <-ctx.Done():
					return
				}
				if err != nil {
					return
				}
			}
		}()
	}
	go func() { wg.Wait(); close(results) }()

	pending := make(map[int]CandidateResult, n)
	next := 0
	drain := func() {
		for range results {
		}
	}
	for d := range results {
		if d.err != nil {
			cancel()
			drain()
			return fmt.Errorf("higher-order-sequence candidate %q: %w", items[d.i], d.err)
		}
		pending[d.i] = d.result
		for {
			r, ok := pending[next]
			if !ok {
				break
			}
			delete(pending, next)
			if err := onReady(next, items[next], r); err != nil {
				cancel()
				drain()
				return err
			}
			next++
		}
	}
	if next != n {
		if err := ctx.Err(); err != nil {
			return err
		}
		return fmt.Errorf("higher-order-sequence-validate: expected %d candidate jobs, got %d", n, next)
	}
	return nil
}

func candidateExecutorFor(c Config, candidates []Candidate, blocks []Block, lineLength map[string]int, relatives map[string][]string) CandidateExecutor {
	if c.CandidateExecutor != nil {
		return c.CandidateExecutor
	}
	return newDefaultCandidateExecutor(candidates, blocks, lineLength, relatives, c.Permutations, c.Seed)
}

func executorContext(c Config) context.Context {
	if c.Context != nil {
		return c.Context
	}
	return context.Background()
}

func executorWorkers(c Config) int {
	if c.Workers < 1 {
		return 1
	}
	return c.Workers
}
