package transitionnetwork

import (
	"context"
	"fmt"
	"sync"
)

// ReplicateResult is one permutation replicate's output: EdgeKey.String()
// keyed edge log2-enrichment values (always populated) plus outgoing/
// incoming profile-null stats (only for the "primary" phase, which passes
// computeProfiles=true to PermWorkspace.run - "refine" leaves Out/In nil,
// exactly like the pre-Task44 inline loop's `_, _ := ws.run(..., false)`).
type ReplicateResult struct {
	Edges map[string]float64
	Out   map[string]profileNullStat
	In    map[string]profileNullStat
}

// PermutationExecutor runs one independent replicate of the "primary" or
// "refine" permutation-null battery. Task44 mirrors
// normalizationcompare.BaselineExecutor: the interface abstracts only
// *where* a replicate executes (goroutine/process/remote), never what it
// computes or in what order results are folded into the running
// exceed-counts - see runBattery below for the single reduction path every
// backend goes through.
type PermutationExecutor interface {
	Run(ctx context.Context, phase string, rep int) (ReplicateResult, error)
}

// defaultPermutationExecutor is the in-process PermutationExecutor used
// whenever Config.PermutationExecutor is nil (Config.Executor == "" or
// "goroutine"): the exact same computation the pre-Task44 loop ran
// sequentially, reachable through the same interface process/remote
// executors implement. ws is built once (mirrors the pre-existing
// `ws := newPermWorkspace(a, c.MinBlockTokenCount)` reuse across every
// replicate) and is not safe for concurrent use, so Run serializes access.
type defaultPermutationExecutor struct {
	ws   *PermWorkspace
	seed int64
	mu   sync.Mutex
}

func newDefaultPermutationExecutor(ws *PermWorkspace, seed int64) *defaultPermutationExecutor {
	return &defaultPermutationExecutor{ws: ws, seed: seed}
}

func (e *defaultPermutationExecutor) Run(_ context.Context, phase string, rep int) (ReplicateResult, error) {
	e.mu.Lock()
	defer e.mu.Unlock()
	return ComputeReplicate(e.ws, e.seed, phase, rep)
}

// ComputeReplicate is the one function every backend (in-process default,
// subprocess, or remote worker) calls: PermWorkspace.run is completely
// unchanged, this only stringifies its EdgeKey-keyed map for a
// JSON-serializable result.
func ComputeReplicate(ws *PermWorkspace, seed int64, phase string, rep int) (ReplicateResult, error) {
	computeProfiles := phase == "primary"
	if phase != "primary" && phase != "refine" {
		return ReplicateResult{}, fmt.Errorf("transition-network-validate: unknown permutation phase %q", phase)
	}
	es, outs, ins := ws.run(seed, rep, computeProfiles)
	edges := make(map[string]float64, len(es))
	for k, v := range es {
		edges[k.String()] = v
	}
	return ReplicateResult{Edges: edges, Out: outs, In: ins}, nil
}

// runBattery dispatches n independent (phase, rep) replicates through
// executor with a bounded worker pool, restoring canonical rep-ascending
// order before invoking onReady for each - regardless of the completion
// order any goroutine/process/remote backend delivers results in. This is
// the single reduction path every backend goes through: arrival order
// never reaches a running exceed-count or a checkpoint write.
// startAt/resumeFrom mirror the pre-existing loop's own resume point
// (cp.Completed for "primary", cp.RefineCompleted for "refine").
func runBattery(ctx context.Context, executor PermutationExecutor, phase string, resumeFrom, n, workers int, onReady func(rep int, r ReplicateResult) error) error {
	if resumeFrom >= n {
		return nil
	}
	remaining := n - resumeFrom
	if workers < 1 {
		workers = 1
	}
	if workers > remaining {
		workers = remaining
	}
	if ctx == nil {
		ctx = context.Background()
	}
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	type done struct {
		rep    int
		result ReplicateResult
		err    error
	}
	jobs := make(chan int)
	results := make(chan done, workers)
	go func() {
		defer close(jobs)
		for rep := resumeFrom; rep < n; rep++ {
			select {
			case jobs <- rep:
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
			for rep := range jobs {
				r, err := executor.Run(ctx, phase, rep)
				select {
				case results <- done{rep, r, err}:
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

	pending := make(map[int]ReplicateResult, remaining)
	next := resumeFrom
	drain := func() {
		for range results {
		}
	}
	for d := range results {
		if d.err != nil {
			cancel()
			drain()
			return fmt.Errorf("%s permutation rep %d: %w", phase, d.rep, d.err)
		}
		pending[d.rep] = d.result
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
		return fmt.Errorf("%s: expected %d permutation reps, got %d", phase, n, next)
	}
	return nil
}

func permutationExecutorFor(c Config, ws *PermWorkspace) PermutationExecutor {
	if c.PermutationExecutor != nil {
		return c.PermutationExecutor
	}
	return newDefaultPermutationExecutor(ws, c.Seed)
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
