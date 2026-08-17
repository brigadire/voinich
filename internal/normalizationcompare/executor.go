package normalizationcompare

import (
	"context"
	"fmt"
	"sort"
	"sync"

	"zcore.dev/voinich/internal/normalization"
	"zcore.dev/voinich/internal/sequenceanalyze"
)

// defaultExecutor is the in-process BaselineExecutor used whenever
// Config.BaselineExecutor is nil (Config.Executor == "" or "goroutine"):
// the exact same computation normalization-compare ran sequentially before
// Task42, just reachable through the same interface the process/remote
// backends implement.
type defaultExecutor struct {
	corpus        normalization.Corpus
	minTokenCount int
	singletonMode string
	seed          int64
	params        sequenceanalyze.Parameters
	models        map[string]normalization.Model
}

func newDefaultExecutor(classes normalization.ClassesOutput, corpus normalization.Corpus, seed int64, params sequenceanalyze.Parameters) *defaultExecutor {
	models := make(map[string]normalization.Model, len(classes.Models))
	for _, m := range classes.Models {
		models[m.Label] = m
	}
	return &defaultExecutor{
		corpus: corpus, minTokenCount: classes.Meta.MinTokenCount, singletonMode: classes.Meta.SingletonMode,
		seed: seed, params: params, models: models,
	}
}

func (e *defaultExecutor) Run(_ context.Context, threshold string, run int) (BaselineResult, error) {
	model, ok := e.models[threshold]
	if !ok {
		return BaselineResult{}, fmt.Errorf("unknown threshold %q", threshold)
	}
	return RunRandomTrial(model, e.corpus, e.minTokenCount, e.singletonMode, e.seed, run, e.params)
}

func (e *defaultExecutor) Close() error { return nil }

// runBaselines dispatches n independent (threshold, run) trials through
// executor with a bounded worker pool, restoring canonical run-ascending
// order before returning - regardless of the completion order any
// goroutine/process/remote backend delivers results in. This is the single
// reduction path every backend goes through: "network arrival order never
// determines summation order" (Task42 section 3).
func runBaselines(ctx context.Context, executor BaselineExecutor, threshold string, n, workers int) ([]BaselineResult, error) {
	if n == 0 {
		return nil, nil
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
		run    int
		result BaselineResult
		err    error
	}
	jobs := make(chan int)
	results := make(chan done, workers)
	var wg sync.WaitGroup
	for range workers {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for run := range jobs {
				r, err := executor.Run(ctx, threshold, run)
				select {
				case results <- done{run, r, err}:
				case <-ctx.Done():
					return
				}
				if err != nil {
					return
				}
			}
		}()
	}
	go func() {
		defer close(jobs)
		for run := 0; run < n; run++ {
			select {
			case jobs <- run:
			case <-ctx.Done():
				return
			}
		}
	}()
	go func() { wg.Wait(); close(results) }()

	byRun := make(map[int]BaselineResult, n)
	for d := range results {
		if d.err != nil {
			cancel()
			for range results {
			}
			return nil, fmt.Errorf("threshold %s baseline run %d: %w", threshold, d.run, d.err)
		}
		byRun[d.run] = d.result
	}
	if len(byRun) != n {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		return nil, fmt.Errorf("threshold %s: expected %d baseline results, got %d", threshold, n, len(byRun))
	}
	runs := make([]int, 0, n)
	for r := range byRun {
		runs = append(runs, r)
	}
	sort.Ints(runs)
	out := make([]BaselineResult, n)
	for _, r := range runs {
		out[r] = byRun[r]
	}
	return out, nil
}
