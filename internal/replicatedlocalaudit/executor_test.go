package replicatedlocalaudit

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"testing"
	"time"
)

func synthDistributionState() *DistributionState {
	blocks := randomMarkovCorpus(101, 6, 30, 6)
	tokens := make([]token, 0)
	for _, b := range blocks {
		tokens = append(tokens, b.Tokens...)
	}
	dc := []distanceCandidate{{ID: "d1", A: "t00", B: "t01", Q: .01}}
	sc := []sequenceCandidate{{ID: "s1", Sequence: "t00 t01", Tokens: []string{"t00", "t01"}}}
	profiles := map[string]profile{}
	counts := map[string]map[string]int{}
	for _, b := range blocks {
		profiles[b.ID] = buildProfile(b)
		counts[b.ID] = map[string]int{}
		for _, t := range b.Tokens {
			counts[b.ID][t.Text]++
		}
	}
	refs := map[string]profile{}
	for _, b := range blocks {
		refs[b.ID] = mergeProfiles(profiles, b.ID)
	}
	eligible := map[string][]string{}
	for _, d := range dc {
		for _, b := range blocks {
			if counts[b.ID][d.A] >= 1 && counts[b.ID][d.B] >= 1 {
				eligible[d.ID] = append(eligible[d.ID], b.ID)
			}
		}
	}
	choices := map[string]matchedVocab{"d1": {a: []string{"t02", "t03"}, b: []string{"t04", "t05"}}}
	frozenPairs := map[string]bool{"t00\x00t01": true}
	return &DistributionState{
		tokens: tokens, blocks: blocks, dc: dc, sc: sc,
		profiles: profiles, refs: refs, eligible: eligible, choices: choices,
		frozenPairs: frozenPairs, markovTraining: buildMarkovTraining(blocks),
	}
}

func TestComputeReplicateDeterministicPerPhase(t *testing.T) {
	state := synthDistributionState()
	for _, phase := range []string{"distance", "shuffle", "markov"} {
		a, err := ComputeReplicate(state, 5, phase, 2)
		if err != nil {
			t.Fatal(err)
		}
		b, err := ComputeReplicate(state, 5, phase, 2)
		if err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(a, b) {
			t.Fatalf("phase=%s: ComputeReplicate not deterministic across identical calls", phase)
		}
	}
}

func TestComputeReplicateUnknownPhase(t *testing.T) {
	if _, err := ComputeReplicate(synthDistributionState(), 1, "nonsense", 0); err == nil {
		t.Fatal("expected an error for an unknown phase")
	}
}

// TestDistanceOrderRestorationMatchesSequentialAppend is the specific
// Task44 correctness requirement for this stage: cp.Distance[key] is read
// positionally by run index later (buildDistanceResults' jackknife), so
// appending replicate results as runBattery delivers them (always in
// ascending run order, regardless of completion order) must produce the
// exact same slice a purely sequential loop would have.
func TestDistanceOrderRestorationMatchesSequentialAppend(t *testing.T) {
	state := synthDistributionState()
	n := 10
	seq := map[string][]float64{}
	for run := range n {
		r, err := ComputeReplicate(state, 5, "distance", run)
		if err != nil {
			t.Fatal(err)
		}
		for k, v := range r.Distance {
			seq[k] = append(seq[k], v)
		}
	}

	exec := &stubDelayExecutor{state: state, seed: 5, delay: func(run int) time.Duration { return time.Duration(n-run) * time.Millisecond }}
	got := map[string][]float64{}
	err := runBattery(context.Background(), exec, "distance", 0, n, 6, func(run int, r ReplicateResult) error {
		for k, v := range r.Distance {
			got[k] = append(got[k], v)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(got, seq) {
		t.Fatalf("out-of-order completion produced a different Distance append order\ngot=%v\nwant=%v", got, seq)
	}
}

type stubDelayExecutor struct {
	state *DistributionState
	seed  int64
	delay func(run int) time.Duration
}

func (s *stubDelayExecutor) Run(ctx context.Context, phase string, run int) (ReplicateResult, error) {
	if s.delay != nil {
		select {
		case <-time.After(s.delay(run)):
		case <-ctx.Done():
			return ReplicateResult{}, ctx.Err()
		}
	}
	return ComputeReplicate(s.state, s.seed, phase, run)
}

type stubExecutor struct {
	fail map[int]bool
}

func (s *stubExecutor) Run(ctx context.Context, phase string, run int) (ReplicateResult, error) {
	if s.fail[run] {
		return ReplicateResult{}, fmt.Errorf("synthetic failure at run %d", run)
	}
	return ReplicateResult{ShuffleTotal: map[string]int{"k": run}}, nil
}

func TestRunBatteryRestoresCanonicalOrder(t *testing.T) {
	n := 12
	exec := &stubDelayExecutor{state: synthDistributionState(), seed: 1, delay: func(run int) time.Duration { return time.Duration(n-run) * time.Millisecond }}
	var got []int
	err := runBattery(context.Background(), exec, "shuffle", 0, n, 6, func(run int, r ReplicateResult) error {
		got = append(got, run)
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

func TestRunBatteryResumesFromCheckpoint(t *testing.T) {
	n, resumeFrom := 10, 4
	exec := &stubExecutor{}
	var got []int
	if err := runBattery(context.Background(), exec, "shuffle", resumeFrom, n, 3, func(run int, r ReplicateResult) error {
		got = append(got, run)
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	want := []int{4, 5, 6, 7, 8, 9}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("resumed runs = %v, want %v", got, want)
	}
}

func TestRunBatteryPropagatesWorkerError(t *testing.T) {
	exec := &stubExecutor{fail: map[int]bool{3: true}}
	if err := runBattery(context.Background(), exec, "shuffle", 0, 8, 4, func(int, ReplicateResult) error { return nil }); err == nil {
		t.Fatal("expected the synthetic failure to propagate")
	}
}

func TestRunBatteryPropagatesOnReadyError(t *testing.T) {
	sentinel := errors.New("checkpoint write failed")
	exec := &stubExecutor{}
	err := runBattery(context.Background(), exec, "shuffle", 0, 8, 4, func(run int, r ReplicateResult) error {
		if run == 2 {
			return sentinel
		}
		return nil
	})
	if !errors.Is(err, sentinel) {
		t.Fatalf("got %v, want sentinel error propagated", err)
	}
}

func TestDefaultPermutationExecutorMatchesComputeReplicate(t *testing.T) {
	state := synthDistributionState()
	exec := newDefaultPermutationExecutor(state, 9)
	for _, phase := range []string{"distance", "shuffle", "markov"} {
		for _, run := range []int{0, 3} {
			got, err := exec.Run(context.Background(), phase, run)
			if err != nil {
				t.Fatal(err)
			}
			want, err := ComputeReplicate(state, 9, phase, run)
			if err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(got, want) {
				t.Fatalf("phase=%s run=%d: default executor diverged from ComputeReplicate", phase, run)
			}
		}
	}
}
