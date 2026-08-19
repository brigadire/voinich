package beginendanalyze

import (
	"context"
	"fmt"
	"sync"
)

// DefaultCandidateBatchSize is Task47 section 13's granularity-study
// result: large enough that mTLS/HTTP round-trip overhead stays small next
// to a batch's own compute time (each pair costs roughly 0.7ms on the
// profiled Astafiev production workload - see DISTRIBUTED_EXECUTION_AUDIT.md
// section on begin-end-analyze), small enough that ten remote workers still
// see hundreds of batches to load-balance across.
const DefaultCandidateBatchSize = 2048

// BatchResult is one candidate-pair batch's output: every Candidate for
// flat pair indexes in [lo, hi) with ai != bi, in ascending index order -
// exactly the slice the pre-Task47 inline double loop would have appended
// for that same index range.
type BatchResult struct {
	Candidates []Candidate
}

// CandidateBatchExecutor runs one independent batch of the candidate-pair
// loop. Task47 mirrors every other internal/conditionalregime-distributed
// package (transitionnetwork.PermutationExecutor, and so on): the interface
// abstracts only *where* a batch executes (goroutine/process/remote), never
// what it computes or in what order results are folded into the resulting
// candidate slice - see runCandidateBatches below for the single reduction
// path every backend goes through. Unlike every other stage's executor,
// this one is never RNG-dependent: LoadForDistribution's permutation
// moments are computed once, sequentially, before any batch dispatch (see
// analyze.go), so a batch's own computation reads only already-frozen,
// read-only inputs.
type CandidateBatchExecutor interface {
	Run(ctx context.Context, batchIndex int) (BatchResult, error)
}

// defaultCandidateBatchExecutor is the in-process CandidateBatchExecutor
// used whenever Config.BatchExecutor is nil (Config.Executor == "" or
// "goroutine"). ws is read-only once built, so Run is safe for concurrent
// use by multiple goroutines - unlike the shared-scratch-buffer executors
// several other stages need to serialize.
type defaultCandidateBatchExecutor struct {
	ws        *Workspace
	batchSize int
}

func (e *defaultCandidateBatchExecutor) Run(_ context.Context, batchIndex int) (BatchResult, error) {
	return ComputeBatch(e.ws, batchIndex, e.batchSize), nil
}

// ComputeBatch is the one function every backend (in-process default,
// subprocess, or remote worker) calls: Workspace.candidateAt is completely
// unchanged from the pre-Task47 inline loop body, this only bounds it to
// one batch's flat-index range and skips the ai==bi diagonal exactly as
// the original loop's `if ai == bi { continue }` did.
func ComputeBatch(ws *Workspace, batchIndex, batchSize int) BatchResult {
	total := ws.TotalPairs()
	lo := batchIndex * batchSize
	hi := lo + batchSize
	if hi > total {
		hi = total
	}
	if lo >= hi {
		return BatchResult{}
	}
	k := ws.k
	candidates := make([]Candidate, 0, hi-lo)
	for idx := lo; idx < hi; idx++ {
		ai, bi := idx/k, idx%k
		if ai == bi {
			continue
		}
		candidates = append(candidates, ws.candidateAt(ai, bi))
	}
	return BatchResult{Candidates: candidates}
}

// runCandidateBatches dispatches n independent batches through executor
// with a bounded worker pool, restoring canonical batch-index-ascending
// order before invoking onReady for each - regardless of the completion
// order any goroutine/process/remote backend delivers results in. Because
// batches partition the flat pair-index space contiguously and
// batch-index order is pair-index order, this reduction produces exactly
// the same candidate order the pre-Task47 single-threaded double loop
// produced, regardless of worker count or completion order.
func runCandidateBatches(ctx context.Context, executor CandidateBatchExecutor, n, workers int, onReady func(batchIndex int, r BatchResult) error) error {
	if n <= 0 {
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
		batch  int
		result BatchResult
		err    error
	}
	jobs := make(chan int)
	results := make(chan done, workers)
	go func() {
		defer close(jobs)
		for batch := 0; batch < n; batch++ {
			select {
			case jobs <- batch:
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
			for batch := range jobs {
				r, err := executor.Run(ctx, batch)
				select {
				case results <- done{batch, r, err}:
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

	pending := make(map[int]BatchResult, n)
	next := 0
	drain := func() {
		for range results {
		}
	}
	for d := range results {
		if d.err != nil {
			cancel()
			drain()
			return fmt.Errorf("candidate batch %d: %w", d.batch, d.err)
		}
		pending[d.batch] = d.result
		for {
			r, ok := pending[next]
			if !ok {
				break
			}
			delete(pending, next)
			if err := onReady(next, r); err != nil {
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
		return fmt.Errorf("candidate batches: expected %d, got %d", n, next)
	}
	return nil
}

// collectCandidates dispatches every candidate-pair batch and concatenates
// results in ascending batch order, producing exactly the same []Candidate
// (same length, same order, same values) the pre-Task47 inline double loop
// built directly.
func collectCandidates(c Config, ws *Workspace) ([]Candidate, error) {
	batchSize := c.CandidateBatchSize
	if batchSize <= 0 {
		batchSize = DefaultCandidateBatchSize
	}
	total := ws.TotalPairs()
	numBatches := (total + batchSize - 1) / batchSize
	candidates := make([]Candidate, 0, ws.k*(ws.k-1))
	executor := batchExecutorFor(c, ws, batchSize)
	err := runCandidateBatches(executorContext(c), executor, numBatches, executorWorkers(c), func(_ int, r BatchResult) error {
		candidates = append(candidates, r.Candidates...)
		return nil
	})
	if err != nil {
		return nil, err
	}
	return candidates, nil
}

func batchExecutorFor(c Config, ws *Workspace, batchSize int) CandidateBatchExecutor {
	if c.BatchExecutor != nil {
		return c.BatchExecutor
	}
	return &defaultCandidateBatchExecutor{ws: ws, batchSize: batchSize}
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
