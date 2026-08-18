package tokenrelationvalidation

import (
	"context"
	"fmt"
	"sort"
	"sync"
)

// PermutationExecutor runs one independent replicate of one of the six
// frozen permutation batteries ("direction", "refine_direction",
// "sequence", "refine_sequence", "profile", "refine_profile"). Task44
// mirrors normalizationcompare.BaselineExecutor: the interface abstracts
// only *where* a replicate executes (goroutine/process/remote), never what
// it computes or in what order results are folded into the running sums -
// see runBattery below for the single reduction path every backend goes
// through. ComputeReplicate is the one function every implementation
// (including the in-process default) ultimately calls.
type PermutationExecutor interface {
	Run(ctx context.Context, family string, run int) (map[string]float64, error)
}

// candidatesByFamily returns the (unordered) subset of candidates whose
// Family is one of families. Callers only ever read the result by
// candidate ID, so map order never matters.
func candidatesByFamily(candidates []Candidate, families ...string) map[string]Candidate {
	want := make(map[string]bool, len(families))
	for _, f := range families {
		want[f] = true
	}
	out := map[string]Candidate{}
	for _, c := range candidates {
		if want[c.Family] {
			out[c.ID] = c
		}
	}
	return out
}

// ComputeReplicate runs exactly the scoring logic every in-process battery
// loop already used before Task44: PermuteWithinBlocks with the frozen
// c.Seed+run*1000003 formula (unchanged), then that family's own scorer.
// It reads the *full* family-typed candidate set (all "directional"
// candidates for direction/refine_direction, all "sequence" candidates for
// sequence/refine_sequence, all "structural"/"distance-profile" candidates
// for profile/refine_profile) rather than whichever eligibility-filtered
// subset a base vs. refine phase actually accumulates over: every family's
// scorer computes each candidate ID's score independently of which other
// candidates are also being scored in the same call (buildDirectionEdges'
// globalMax widening from extra candidates only affects the per-edge scan
// depth, gated per-candidate by its own maxD; makeSequenceTrie tracks match
// counts per leaf ID independently; profilePermutationScores' values[id]
// loop is a plain per-ID range), so this is exactly the same score a
// narrower call would have produced for any ID both calls have in common -
// letting one job type serve both a battery and its refinement without
// needing to ship a separate eligible-ID list per job.
func ComputeReplicate(blocks []Block, candidates []Candidate, maxD int, seed int64, family string, run int) (map[string]float64, error) {
	permuted := PermuteWithinBlocks(blocks, seed+int64(run)*1000003)
	switch family {
	case "direction", "refine_direction":
		lookup := candidatesByFamily(candidates, "directional")
		de := buildDirectionEdges(lookup, maxD)
		return directionScoresAll(permuted, lookup, de), nil
	case "sequence", "refine_sequence":
		return sequenceScores(permuted, makeSequenceTrie(candidates)), nil
	case "profile", "refine_profile":
		lookup := candidatesByFamily(candidates, "structural", "distance-profile")
		return profilePermutationScores(permuted, lookup, maxD, newProfileWorkspace()), nil
	default:
		return nil, fmt.Errorf("token-relation-validate: unknown permutation family %q", family)
	}
}

// defaultPermutationExecutor is the in-process PermutationExecutor used
// whenever Config.PermutationExecutor is nil (Config.Executor == "" or
// "goroutine"): the exact same computation every battery ran sequentially
// before Task44, reachable through the same interface process/remote
// executors implement. profileWS is reused across every Run call (mirrors
// profilePermutations'/refineProfilePermutations' pre-existing
// per-battery `ws := newProfileWorkspace()` scratch-reuse optimization;
// sharing it across the profile and refine_profile phases too is safe -
// buildLocalProfiles' cache is keyed by block ID and content, not by which
// phase asked for it - and only saves more allocation).
type defaultPermutationExecutor struct {
	blocks     []Block
	candidates []Candidate
	maxD       int
	seed       int64
	mu         sync.Mutex
	profileWS  *profileWorkspace
}

func newDefaultPermutationExecutor(blocks []Block, candidates []Candidate, maxD int, seed int64) *defaultPermutationExecutor {
	return &defaultPermutationExecutor{blocks: blocks, candidates: candidates, maxD: maxD, seed: seed, profileWS: newProfileWorkspace()}
}

func (e *defaultPermutationExecutor) Run(_ context.Context, family string, run int) (map[string]float64, error) {
	if family == "profile" || family == "refine_profile" {
		// profileWorkspace's cache is not safe for concurrent use by
		// multiple goroutines; the shared instance is reused for its
		// allocation savings, so serialize access to it specifically.
		e.mu.Lock()
		defer e.mu.Unlock()
		permuted := PermuteWithinBlocks(e.blocks, e.seed+int64(run)*1000003)
		lookup := candidatesByFamily(e.candidates, "structural", "distance-profile")
		return profilePermutationScores(permuted, lookup, e.maxD, e.profileWS), nil
	}
	return ComputeReplicate(e.blocks, e.candidates, e.maxD, e.seed, family, run)
}

// runBattery dispatches n independent (family, run) replicates through
// executor with a bounded worker pool, restoring canonical run-ascending
// order before invoking onReady for each - regardless of the completion
// order any goroutine/process/remote backend delivers results in. This is
// the single reduction path every backend goes through: arrival order
// never reaches a running sum or a checkpoint write. resumeFrom is the
// checkpoint's *Completed value: onReady is invoked starting at
// run=resumeFrom, exactly reproducing the original for-loop's resumption
// behavior, so a resumed run's checkpoint writes land at the same points
// they always did.
func runBattery(ctx context.Context, executor PermutationExecutor, family string, resumeFrom, n, workers int, onReady func(run int, scores map[string]float64) error) error {
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
		scores map[string]float64
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
				scores, err := executor.Run(ctx, family, run)
				select {
				case results <- done{run, scores, err}:
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

	pending := make(map[int]map[string]float64, remaining)
	next := resumeFrom
	drain := func() {
		for range results {
		}
	}
	for d := range results {
		if d.err != nil {
			cancel()
			drain()
			return fmt.Errorf("%s permutation run %d: %w", family, d.run, d.err)
		}
		pending[d.run] = d.scores
		for {
			scores, ok := pending[next]
			if !ok {
				break
			}
			delete(pending, next)
			if err := onReady(next, scores); err != nil {
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
		return fmt.Errorf("%s: expected %d permutation runs, got %d", family, n, next)
	}
	return nil
}

// permutationExecutorFor returns c's configured PermutationExecutor, or the
// in-process default built from a's own blocks/candidates/maxD when none
// was set (Config.Executor == "" or "goroutine").
func permutationExecutorFor(c Config, a *Analysis, maxD int) PermutationExecutor {
	if c.PermutationExecutor != nil {
		return c.PermutationExecutor
	}
	return newDefaultPermutationExecutor(a.Blocks, a.Candidates, maxD, c.Seed)
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

// sortedIDs is a small helper the six battery functions use to iterate a
// candidate-ID-keyed map in a fixed order when building deterministic
// output (e.g. Controls) - never for anything that affects a computed
// value, only for byte-stable output ordering.
func sortedIDs(m map[string]float64) []string {
	ids := make([]string, 0, len(m))
	for id := range m {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	return ids
}
