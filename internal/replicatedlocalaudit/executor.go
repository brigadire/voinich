package replicatedlocalaudit

import (
	"context"
	"fmt"
	"math/rand"
	"sync"
)

// ReplicateResult is one permutation replicate's output. Only the fields
// for the requested phase ("distance", "shuffle", or "markov") are
// populated - exactly the values the pre-Task44 inline loop body for that
// phase computed before folding them into the running sums/exceed-counts.
type ReplicateResult struct {
	// "distance": key = candidateID+"\x00"+blockID.
	Distance map[string]float64

	// "shuffle": key = sequence candidateID.
	ShuffleTotal, ShuffleBlocks map[string]int

	// "markov": key = sequence candidateID. ObservedTotal/ObservedBlocks
	// are sequenceStatsOne's per-run result (which blocks a Markov
	// simulation could draw from varies by replicate), exactly as the
	// pre-Task44 loop recomputed them inside the loop body every run.
	MarkovTotal, MarkovBlocks, MarkovObservedTotal, MarkovObservedBlocks map[string]int
}

// PermutationExecutor runs one independent replicate of the "distance",
// "shuffle", or "markov" null battery. Task44 mirrors
// normalizationcompare.BaselineExecutor: the interface abstracts only
// *where* a replicate executes (goroutine/process/remote), never what it
// computes or in what order results are folded into the running
// sums/exceed-counts - see runBattery below for the single reduction path
// every backend goes through, which is also what restores cp.Distance's
// per-run append order to strict run-ascending regardless of completion
// order (buildDistanceResults' jackknife reads it positionally by index).
type PermutationExecutor interface {
	Run(ctx context.Context, phase string, run int) (ReplicateResult, error)
}

// ComputeReplicate is the one function every backend (in-process default,
// subprocess, or remote worker) calls. Each phase's math is copied
// verbatim from the pre-Task44 inline loop body; only the per-run RNG
// seed formulas (already independent per run before Task44) are reused
// unchanged.
func ComputeReplicate(state *DistributionState, seed int64, phase string, run int) (ReplicateResult, error) {
	switch phase {
	case "distance":
		r := rand.New(rand.NewSource(seed + int64(run)*104729 + 11))
		dist := map[string]float64{}
		for _, d := range state.dc {
			pool := state.choices[d.ID]
			if d.Q > .05 || len(pool.a) == 0 || len(pool.b) == 0 {
				continue
			}
			for _, bid := range state.eligible[d.ID] {
				a, b := "", ""
				for tries := 0; tries < 100; tries++ {
					a, b = pool.a[r.Intn(len(pool.a))], pool.b[r.Intn(len(pool.b))]
					x, y := a, b
					if y < x {
						x, y = y, x
					}
					if a != b && !state.frozenPairs[x+"\x00"+y] {
						break
					}
					a, b = "", ""
				}
				v := 0.
				if a != "" {
					v, _ = compareProfiles(state.profiles[bid], state.refs[bid], a, b)
				}
				dist[d.ID+"\x00"+bid] = v
			}
		}
		return ReplicateResult{Distance: dist}, nil
	case "shuffle":
		sim := shuffledBlocks(state.blocks, seed+int64(run)*104729+23)
		tot, bc := sequenceStats(state.sc, sim)
		return ReplicateResult{ShuffleTotal: tot, ShuffleBlocks: bc}, nil
	case "markov":
		sim, _ := markovBlocks(state.markovTraining, seed+int64(run)*104729+37)
		tot, bc := sequenceStats(state.sc, sim)
		obsTotal, obsBlocks := map[string]int{}, map[string]int{}
		for _, s := range state.sc {
			oa := sequenceStatsOne(s, state.blocks, sim)
			obsTotal[s.ID], obsBlocks[s.ID] = oa[0], oa[1]
		}
		return ReplicateResult{MarkovTotal: tot, MarkovBlocks: bc, MarkovObservedTotal: obsTotal, MarkovObservedBlocks: obsBlocks}, nil
	default:
		return ReplicateResult{}, fmt.Errorf("replicated-local-structure-audit: unknown permutation phase %q", phase)
	}
}

// defaultPermutationExecutor is the in-process PermutationExecutor used
// whenever Config.PermutationExecutor is nil (Config.Executor == "" or
// "goroutine"): the exact same computation every pre-Task44 loop ran
// sequentially, reachable through the same interface process/remote
// executors implement. state is read-only after construction, so Run
// needs no locking of its own.
type defaultPermutationExecutor struct {
	state *DistributionState
	seed  int64
}

func newDefaultPermutationExecutor(state *DistributionState, seed int64) *defaultPermutationExecutor {
	return &defaultPermutationExecutor{state: state, seed: seed}
}

func (e *defaultPermutationExecutor) Run(_ context.Context, phase string, run int) (ReplicateResult, error) {
	return ComputeReplicate(e.state, e.seed, phase, run)
}

// runBattery dispatches n independent (phase, run) replicates through
// executor with a bounded worker pool, restoring canonical run-ascending
// order before invoking onReady for each - regardless of the completion
// order any goroutine/process/remote backend delivers results in.
func runBattery(ctx context.Context, executor PermutationExecutor, phase string, resumeFrom, n, workers int, onReady func(run int, r ReplicateResult) error) error {
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
		run    int
		result ReplicateResult
		err    error
	}
	jobs := make(chan int)
	results := make(chan done, workers)
	go func() {
		defer close(jobs)
		for run := resumeFrom; run < n; run++ {
			select {
			case jobs <- run:
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
			for run := range jobs {
				r, err := executor.Run(ctx, phase, run)
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
			return fmt.Errorf("%s permutation run %d: %w", phase, d.run, d.err)
		}
		pending[d.run] = d.result
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
		return fmt.Errorf("%s: expected %d permutation runs, got %d", phase, n, next)
	}
	return nil
}

func permutationExecutorFor(c Config, state *DistributionState) PermutationExecutor {
	if c.PermutationExecutor != nil {
		return c.PermutationExecutor
	}
	return newDefaultPermutationExecutor(state, c.Seed)
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
