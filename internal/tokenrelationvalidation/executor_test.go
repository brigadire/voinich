package tokenrelationvalidation

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"reflect"
	"sync"
	"testing"
	"time"
)

func candidateSlice(m map[string]Candidate) []Candidate {
	out := make([]Candidate, 0, len(m))
	for _, c := range m {
		out = append(out, c)
	}
	return out
}

// TestComputeReplicateDirectionMatchesInlineComputation proves ComputeReplicate's
// "direction"/"refine_direction" branch (which always scores the full
// "directional"-family candidate set) produces the exact same score for
// any candidate ID a narrower, eligibility-filtered call would have -
// the equivalence executor.go's doc comment claims and the six battery
// functions depend on for correctness.
func TestComputeReplicateDirectionMatchesInlineComputation(t *testing.T) {
	full, blocks := fixtureDirectionCandidatesAndBlocks(20, 15, 4242)
	// A narrower eligibility-filtered subset, as a base/refine battery would use.
	narrow := map[string]Candidate{}
	i := 0
	for id, c := range full {
		if i%2 == 0 {
			narrow[id] = c
		}
		i++
	}
	maxD := 3
	seed := int64(7)
	for _, run := range []int{0, 1, 5} {
		got, err := ComputeReplicate(blocks, candidateSlice(full), maxD, seed, "direction", run)
		if err != nil {
			t.Fatal(err)
		}
		permuted := PermuteWithinBlocks(blocks, seed+int64(run)*1000003)
		de := buildDirectionEdges(narrow, maxD)
		want := directionScoresAll(permuted, narrow, de)
		for id := range narrow {
			if got[id] != want[id] {
				t.Fatalf("run=%d id=%s: full-set score %v != narrow-set score %v", run, id, got[id], want[id])
			}
		}
	}
}

// TestComputeReplicateUnknownFamily proves an unrecognized family is an
// explicit error, never a silent no-op or a panic.
func TestComputeReplicateUnknownFamily(t *testing.T) {
	if _, err := ComputeReplicate(nil, nil, 1, 1, "nonsense", 0); err == nil {
		t.Fatal("expected an error for an unknown family")
	}
}

// stubExecutor lets a test control exactly which run completes when,
// including deliberately finishing out of ascending order, to prove
// runBattery restores canonical order regardless.
type stubExecutor struct {
	delay func(run int) time.Duration
	fail  map[int]bool
	mu    sync.Mutex
	seen  []int
}

func (s *stubExecutor) Run(ctx context.Context, family string, run int) (map[string]float64, error) {
	if s.delay != nil {
		select {
		case <-time.After(s.delay(run)):
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}
	s.mu.Lock()
	s.seen = append(s.seen, run)
	s.mu.Unlock()
	if s.fail[run] {
		return nil, fmt.Errorf("synthetic failure at run %d", run)
	}
	return map[string]float64{"c": float64(run)}, nil
}

func TestRunBatteryRestoresCanonicalOrderRegardlessOfCompletionOrder(t *testing.T) {
	// Runs complete in reverse order (highest run finishes first) purely
	// via an inverted artificial delay - onReady must still fire 0..n-1.
	n := 12
	exec := &stubExecutor{delay: func(run int) time.Duration { return time.Duration(n-run) * time.Millisecond }}
	var got []int
	err := runBattery(context.Background(), exec, "direction", 0, n, 6, func(run int, scores map[string]float64) error {
		got = append(got, run)
		if scores["c"] != float64(run) {
			t.Fatalf("run %d: got score %v", run, scores["c"])
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	want := make([]int, n)
	for i := range want {
		want[i] = i
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("onReady order = %v, want ascending %v", got, want)
	}
}

func TestRunBatteryWorkerCountDoesNotChangeAccumulation(t *testing.T) {
	n := 30
	sumFor := func(workers int) float64 {
		exec := &stubExecutor{}
		var sum float64
		if err := runBattery(context.Background(), exec, "direction", 0, n, workers, func(run int, scores map[string]float64) error {
			sum += scores["c"]
			return nil
		}); err != nil {
			t.Fatal(err)
		}
		return sum
	}
	s1, s4, s16 := sumFor(1), sumFor(4), sumFor(16)
	if s1 != s4 || s4 != s16 {
		t.Fatalf("accumulation depends on worker count: workers=1 -> %v, workers=4 -> %v, workers=16 -> %v", s1, s4, s16)
	}
}

func TestRunBatteryResumesFromCheckpoint(t *testing.T) {
	n, resumeFrom := 10, 4
	exec := &stubExecutor{}
	var got []int
	err := runBattery(context.Background(), exec, "direction", resumeFrom, n, 3, func(run int, scores map[string]float64) error {
		got = append(got, run)
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	want := []int{4, 5, 6, 7, 8, 9}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("resumed runs = %v, want %v", got, want)
	}
	// Resuming at n (already fully complete) dispatches nothing.
	exec2 := &stubExecutor{}
	called := false
	if err := runBattery(context.Background(), exec2, "direction", n, n, 3, func(int, map[string]float64) error {
		called = true
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	if called {
		t.Fatal("onReady must not be invoked when resumeFrom == n")
	}
}

func TestRunBatteryPropagatesWorkerError(t *testing.T) {
	exec := &stubExecutor{fail: map[int]bool{3: true}}
	err := runBattery(context.Background(), exec, "direction", 0, 8, 4, func(int, map[string]float64) error { return nil })
	if err == nil {
		t.Fatal("expected the synthetic failure to propagate")
	}
}

func TestRunBatteryPropagatesOnReadyError(t *testing.T) {
	sentinel := errors.New("checkpoint write failed")
	exec := &stubExecutor{}
	err := runBattery(context.Background(), exec, "direction", 0, 8, 4, func(run int, scores map[string]float64) error {
		if run == 2 {
			return sentinel
		}
		return nil
	})
	if !errors.Is(err, sentinel) {
		t.Fatalf("got %v, want sentinel error propagated", err)
	}
}

// TestDefaultPermutationExecutorMatchesComputeReplicate proves the
// in-process default executor (used whenever Config.PermutationExecutor
// is nil) computes exactly what ComputeReplicate computes - the same
// function every process/remote backend will call.
func TestDefaultPermutationExecutorMatchesComputeReplicate(t *testing.T) {
	full, blocks := fixtureDirectionCandidatesAndBlocks(15, 10, 99)
	candidates := candidateSlice(full)
	maxD := 2
	seed := int64(3)
	exec := newDefaultPermutationExecutor(blocks, candidates, maxD, seed)
	for _, run := range []int{0, 2, 9} {
		got, err := exec.Run(context.Background(), "direction", run)
		if err != nil {
			t.Fatal(err)
		}
		want, err := ComputeReplicate(blocks, candidates, maxD, seed, "direction", run)
		if err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("run=%d: default executor diverged from ComputeReplicate", run)
		}
	}
}

// TestDefaultPermutationExecutorProfileConcurrencySafe exercises the
// profile family (the one case defaultPermutationExecutor serializes
// itself, since profileWorkspace's cache is not safe for concurrent use)
// through runBattery with several workers, checking for data races (run
// with -race) and for a result identical to a strictly sequential run.
func TestDefaultPermutationExecutorProfileConcurrencySafe(t *testing.T) {
	blocksSrc := []Block{}
	r := rand.New(rand.NewSource(55))
	vocab := []string{"aiin", "chey", "shey", "ol", "or"}
	for bi := 0; bi < 8; bi++ {
		n := 15 + r.Intn(10)
		toks := make([]Token, n)
		for i := range toks {
			toks[i] = Token{Text: vocab[r.Intn(len(vocab))], Line: fmt.Sprintf("l%d", i/6)}
		}
		blocksSrc = append(blocksSrc, Block{ID: fmt.Sprintf("b%d", bi), Tokens: toks})
	}
	candidates := []Candidate{
		{ID: "p1", A: "aiin", B: "chey", Family: "structural"},
		{ID: "p2", A: "ol", B: "or", Family: "distance-profile"},
	}
	maxD := 2
	seed := int64(11)
	n := 20

	seq := newDefaultPermutationExecutor(blocksSrc, candidates, maxD, seed)
	seqSums := map[string]float64{}
	for run := 0; run < n; run++ {
		scores, err := seq.Run(context.Background(), "profile", run)
		if err != nil {
			t.Fatal(err)
		}
		for id, v := range scores {
			seqSums[id] += v
		}
	}

	par := newDefaultPermutationExecutor(blocksSrc, candidates, maxD, seed)
	parSums := map[string]float64{}
	err := runBattery(context.Background(), par, "profile", 0, n, 8, func(run int, scores map[string]float64) error {
		for id, v := range scores {
			parSums[id] += v
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(seqSums, parSums) {
		t.Fatalf("concurrent profile execution diverged: sequential=%v concurrent=%v", seqSums, parSums)
	}
}
