package transitionnetwork

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"sync"
	"testing"
	"time"
)

func TestComputeReplicateMatchesRawWorkspaceRun(t *testing.T) {
	a := buildTestAnalysis(11, 6, 40, 8, 2)
	ws := newPermWorkspace(a, 2)
	seed := int64(5)
	for _, tc := range []struct {
		phase           string
		rep             int
		computeProfiles bool
	}{
		{"primary", 0, true},
		{"primary", 3, true},
		{"refine", 10, false},
	} {
		got, err := ComputeReplicate(ws, seed, tc.phase, tc.rep)
		if err != nil {
			t.Fatal(err)
		}
		es, outs, ins := ws.run(seed, tc.rep, tc.computeProfiles)
		want := map[string]float64{}
		for k, v := range es {
			want[k.String()] = v
		}
		if !reflect.DeepEqual(got.Edges, want) {
			t.Fatalf("phase=%s rep=%d: Edges diverged\ngot=%v\nwant=%v", tc.phase, tc.rep, got.Edges, want)
		}
		if !reflect.DeepEqual(got.Out, outs) || !reflect.DeepEqual(got.In, ins) {
			t.Fatalf("phase=%s rep=%d: Out/In diverged", tc.phase, tc.rep)
		}
	}
}

func TestComputeReplicateUnknownPhase(t *testing.T) {
	a := buildTestAnalysis(1, 2, 10, 4, 1)
	ws := newPermWorkspace(a, 1)
	if _, err := ComputeReplicate(ws, 1, "nonsense", 0); err == nil {
		t.Fatal("expected an error for an unknown phase")
	}
}

type stubExecutor struct {
	delay func(rep int) time.Duration
	fail  map[int]bool
}

func (s *stubExecutor) Run(ctx context.Context, phase string, rep int) (ReplicateResult, error) {
	if s.delay != nil {
		select {
		case <-time.After(s.delay(rep)):
		case <-ctx.Done():
			return ReplicateResult{}, ctx.Err()
		}
	}
	if s.fail[rep] {
		return ReplicateResult{}, fmt.Errorf("synthetic failure at rep %d", rep)
	}
	return ReplicateResult{Edges: map[string]float64{"e": float64(rep)}}, nil
}

func TestRunBatteryRestoresCanonicalOrderRegardlessOfCompletionOrder(t *testing.T) {
	n := 12
	exec := &stubExecutor{delay: func(rep int) time.Duration { return time.Duration(n-rep) * time.Millisecond }}
	var got []int
	err := runBattery(context.Background(), exec, "primary", 0, n, 6, func(rep int, r ReplicateResult) error {
		got = append(got, rep)
		if r.Edges["e"] != float64(rep) {
			t.Fatalf("rep %d: got %v", rep, r.Edges["e"])
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
		if err := runBattery(context.Background(), exec, "primary", 0, n, workers, func(rep int, r ReplicateResult) error {
			sum += r.Edges["e"]
			return nil
		}); err != nil {
			t.Fatal(err)
		}
		return sum
	}
	s1, s4, s16 := sumFor(1), sumFor(4), sumFor(16)
	if s1 != s4 || s4 != s16 {
		t.Fatalf("accumulation depends on worker count: 1->%v 4->%v 16->%v", s1, s4, s16)
	}
}

func TestRunBatteryResumesFromCheckpoint(t *testing.T) {
	n, resumeFrom := 10, 4
	exec := &stubExecutor{}
	var got []int
	if err := runBattery(context.Background(), exec, "primary", resumeFrom, n, 3, func(rep int, r ReplicateResult) error {
		got = append(got, rep)
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	want := []int{4, 5, 6, 7, 8, 9}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("resumed reps = %v, want %v", got, want)
	}
}

func TestRunBatteryPropagatesWorkerError(t *testing.T) {
	exec := &stubExecutor{fail: map[int]bool{3: true}}
	if err := runBattery(context.Background(), exec, "primary", 0, 8, 4, func(int, ReplicateResult) error { return nil }); err == nil {
		t.Fatal("expected the synthetic failure to propagate")
	}
}

func TestRunBatteryPropagatesOnReadyError(t *testing.T) {
	sentinel := errors.New("checkpoint write failed")
	exec := &stubExecutor{}
	err := runBattery(context.Background(), exec, "primary", 0, 8, 4, func(rep int, r ReplicateResult) error {
		if rep == 2 {
			return sentinel
		}
		return nil
	})
	if !errors.Is(err, sentinel) {
		t.Fatalf("got %v, want sentinel error propagated", err)
	}
}

func TestDefaultPermutationExecutorMatchesComputeReplicate(t *testing.T) {
	a := buildTestAnalysis(21, 5, 30, 6, 1)
	ws := newPermWorkspace(a, 1)
	exec := newDefaultPermutationExecutor(ws, 9)
	for _, rep := range []int{0, 2, 7} {
		got, err := exec.Run(context.Background(), "primary", rep)
		if err != nil {
			t.Fatal(err)
		}
		want, err := ComputeReplicate(ws, 9, "primary", rep)
		if err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("rep=%d: default executor diverged from ComputeReplicate", rep)
		}
	}
}

func TestDefaultPermutationExecutorConcurrencySafe(t *testing.T) {
	// Accumulates per-edge-key across reps (cp.EdgeExceed[key]+=... in
	// run.go's own shape) rather than collapsing every rep's Edges map into
	// one scalar by ranging it - that ordering is not float-associative and
	// would make this test flaky by its own construction, independent of
	// whether defaultPermutationExecutor is concurrency-safe.
	a := buildTestAnalysis(33, 6, 25, 7, 1)
	ws := newPermWorkspace(a, 1)
	n := 20
	seq := newDefaultPermutationExecutor(ws, 3)
	seqSum := map[string]float64{}
	for rep := range n {
		r, err := seq.Run(context.Background(), "primary", rep)
		if err != nil {
			t.Fatal(err)
		}
		for k, v := range r.Edges {
			seqSum[k] += v
		}
	}

	par := newDefaultPermutationExecutor(ws, 3)
	var parMu sync.Mutex
	parSum := map[string]float64{}
	err := runBattery(context.Background(), par, "primary", 0, n, 8, func(rep int, r ReplicateResult) error {
		parMu.Lock()
		for k, v := range r.Edges {
			parSum[k] += v
		}
		parMu.Unlock()
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(seqSum, parSum) {
		t.Fatalf("concurrent execution diverged: sequential=%v concurrent=%v", seqSum, parSum)
	}
}
