package positionalcontinuation

import (
	"context"
	"fmt"
	"sync"
)

// batteryNames is the fixed set of task44-distributable batteries: the
// audit found exactly these five permutation-null tests independent of
// each other (each builds its own local workspace/rand.Rand - see
// permutation.go/stratified.go/boundary.go - never a shared package-level
// scratch buffer), while every other Part in run.go either has no
// permutation loop or (Part L's jackknife) explicitly reuses the
// permutations<=0 fast path rather than looping, so it stays local.
var batteryNames = []string{"postest_line", "postest_block", "stratified_line", "stratified_block", "boundary"}

// BatteryExecutor runs one independent battery of task44's 5 distributable
// batteries. The unit of independent work here is a whole named battery,
// never a single permutation within it: each battery already builds its
// own workspace and *rand.Rand once and loops internally (permutation.go's
// runPositionalTests, stratified.go's runStratifiedPredecessorTest,
// boundary.go's buildBoundaryDistanceRows), so splitting inside one
// battery would mean re-deriving the same "one sequential RNG stream per
// battery" constraint task22/task23's other candidate/battery-level stages
// already have.
type BatteryExecutor interface {
	Run(ctx context.Context, battery string) (BatteryResult, error)
}

// BatteryResult carries only the fields the requested battery populates -
// exactly the values the pre-Task44 inline runPart body for that battery
// computed before folding them into RunState.
type BatteryResult struct {
	// "postest_line"/"postest_block".
	Dependence PositionDependenceRow
	Entropy    []PositionalEntropyRow
	CheyEffect []CheyEffectRow

	// "stratified_line"/"stratified_block".
	Stratified StratifiedPredecessorRow

	// "boundary".
	Boundary []BoundaryDistanceRow
}

// ComputeBattery is the one function every backend (in-process default,
// subprocess, or remote worker) calls: each battery's math, copied
// verbatim from RunAndWrite's per-part closures, including the exact same
// seedFor(...) sub-seed derivation every one of them already used.
func ComputeBattery(sAiinOccs []SAiinOccurrence, aiinOccs []AiinOccurrence, battery string, permutations int, seed int64) (BatteryResult, error) {
	switch battery {
	case "postest_line":
		r := runPositionalTests(sAiinOccs, "line_position", lineCategories, permutations, seedFor(seed, "line_position"))
		return BatteryResult{Dependence: r.Dependence, Entropy: r.Entropy, CheyEffect: r.CheyEffect}, nil
	case "postest_block":
		r := runPositionalTests(sAiinOccs, "block_position_coarse", blockCoarseCategories, permutations, seedFor(seed, "block_position"))
		return BatteryResult{Dependence: r.Dependence, Entropy: r.Entropy, CheyEffect: r.CheyEffect}, nil
	case "stratified_line":
		row := runStratifiedPredecessorTest(aiinOccs, "line_position", permutations, seedFor(seed, "stratified_line"))
		return BatteryResult{Stratified: row}, nil
	case "stratified_block":
		row := runStratifiedPredecessorTest(aiinOccs, "block_position_coarse", permutations, seedFor(seed, "stratified_block"))
		return BatteryResult{Stratified: row}, nil
	case "boundary":
		rows := buildBoundaryDistanceRows(sAiinOccs, permutations, seedFor(seed, "boundary"))
		return BatteryResult{Boundary: rows}, nil
	default:
		return BatteryResult{}, fmt.Errorf("positional-continuation-validate: unknown battery %q", battery)
	}
}

// defaultBatteryExecutor is the in-process BatteryExecutor used whenever
// Config.BatteryExecutor is nil (Config.Executor == "" or "goroutine").
// Every field below is read-only after construction: each battery builds
// its own workspace/*rand.Rand fresh per call, so concurrent calls for
// distinct batteries need no mutex here.
type defaultBatteryExecutor struct {
	sAiinOccs    []SAiinOccurrence
	aiinOccs     []AiinOccurrence
	permutations int
	seed         int64
}

func newDefaultBatteryExecutor(sAiinOccs []SAiinOccurrence, aiinOccs []AiinOccurrence, permutations int, seed int64) *defaultBatteryExecutor {
	return &defaultBatteryExecutor{sAiinOccs: sAiinOccs, aiinOccs: aiinOccs, permutations: permutations, seed: seed}
}

func (e *defaultBatteryExecutor) Run(_ context.Context, battery string) (BatteryResult, error) {
	return ComputeBattery(e.sAiinOccs, e.aiinOccs, battery, e.permutations, e.seed)
}

// runBatteryDispatch dispatches n independent battery jobs (identified by
// items[i]) through executor with a bounded worker pool, restoring the
// original ascending items order before invoking onReady for each -
// regardless of the completion order any goroutine/process/remote backend
// delivers results in. Mirrors higherorderseq.runCandidateBattery's exact
// shape, retyped for BatteryResult.
func runBatteryDispatch(ctx context.Context, executor BatteryExecutor, items []string, workers int, onReady func(i int, battery string, r BatteryResult) error) error {
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
		result BatteryResult
		err    error
	}
	jobs := make(chan int)
	results := make(chan done, workers)
	go func() {
		defer close(jobs)
		for i := range n {
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

	pending := make(map[int]BatteryResult, n)
	next := 0
	drain := func() {
		for range results {
		}
	}
	for d := range results {
		if d.err != nil {
			cancel()
			drain()
			return fmt.Errorf("positional-continuation battery %q: %w", items[d.i], d.err)
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
		return fmt.Errorf("positional-continuation-validate: expected %d battery jobs, got %d", n, next)
	}
	return nil
}

func batteryExecutorFor(c Config, sAiinOccs []SAiinOccurrence, aiinOccs []AiinOccurrence) BatteryExecutor {
	if c.BatteryExecutor != nil {
		return c.BatteryExecutor
	}
	return newDefaultBatteryExecutor(sAiinOccs, aiinOccs, c.Permutations, c.Seed)
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
