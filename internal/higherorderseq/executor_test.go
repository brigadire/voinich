package higherorderseq

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"testing"
)

// executorFixture builds two independent frozen candidates ("a b c" and
// "x y z") each with their own vocabulary and blocks, mirroring
// TestDeterministicSeed/TestJackknifeRemovesExactlyOneBlock's fixture shape
// (repeatABC-style padding above primaryMinCountB), so ComputeCandidate has
// real per-candidate work (CMI, LOBO, jackknife, position) to do for each.
func executorFixture() ([]Candidate, []Block, map[string]int, map[string][]string) {
	repeat := func(a, b, c string, n int) []Token {
		var out []Token
		for range n {
			out = append(out, tok(a, "L", 0), tok(b, "L", 0), tok(c, "L", 0))
		}
		for range primaryMinCountB {
			out = append(out, tok("q", "L", 0), tok(b, "L", 0), tok("q", "L", 0))
		}
		return out
	}
	candidates := []Candidate{
		{Sequence: "a b c", Tokens: []string{"a", "b", "c"}, Family: "primary"},
		{Sequence: "x y z", Tokens: []string{"x", "y", "z"}, Family: "primary"},
	}
	blocks := []Block{
		mkBlock("B1", repeat("a", "b", "c", 5)...),
		mkBlock("B2", repeat("a", "b", "c", 5)...),
		mkBlock("B3", repeat("a", "b", "c", 5)...),
		mkBlock("B4", repeat("x", "y", "z", 5)...),
		mkBlock("B5", repeat("x", "y", "z", 5)...),
		mkBlock("B6", repeat("x", "y", "z", 5)...),
	}
	lineLength := map[string]int{"L": 3}
	relatives := map[string][]string{}
	return candidates, blocks, lineLength, relatives
}

func TestComputeCandidateDeterministicPerSequence(t *testing.T) {
	candidates, blocks, lineLength, relatives := executorFixture()
	r1, err := ComputeCandidate(candidates, blocks, lineLength, relatives, 200, 1, "a b c")
	if err != nil {
		t.Fatal(err)
	}
	r2, err := ComputeCandidate(candidates, blocks, lineLength, relatives, 200, 1, "a b c")
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(r1, r2) {
		t.Fatalf("ComputeCandidate is not deterministic for the same sequence\nr1=%#v\nr2=%#v", r1, r2)
	}
	if r1.Candidate.Sequence != "a b c" {
		t.Fatalf("wrong candidate embedded: %+v", r1.Candidate)
	}

	r3, err := ComputeCandidate(candidates, blocks, lineLength, relatives, 200, 1, "x y z")
	if err != nil {
		t.Fatal(err)
	}
	if r3.CMI.Sequence != "x y z" || reflect.DeepEqual(r1.CMI, r3.CMI) {
		t.Fatalf("two distinct candidates must not produce identical CMI results: r1=%+v r3=%+v", r1.CMI, r3.CMI)
	}
}

func TestComputeCandidateUnknownSequence(t *testing.T) {
	candidates, blocks, lineLength, relatives := executorFixture()
	if _, err := ComputeCandidate(candidates, blocks, lineLength, relatives, 50, 1, "no such sequence"); err == nil {
		t.Fatal("expected an error for an unknown candidate sequence")
	}
}

func TestDefaultCandidateExecutorMatchesComputeCandidate(t *testing.T) {
	candidates, blocks, lineLength, relatives := executorFixture()
	ex := newDefaultCandidateExecutor(candidates, blocks, lineLength, relatives, 200, 1)
	got, err := ex.Run(context.Background(), "a b c")
	if err != nil {
		t.Fatal(err)
	}
	want, err := ComputeCandidate(candidates, blocks, lineLength, relatives, 200, 1, "a b c")
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("default executor diverged from ComputeCandidate\ngot=%#v\nwant=%#v", got, want)
	}
}

func TestRunCandidateBatteryRestoresCanonicalOrder(t *testing.T) {
	candidates, blocks, lineLength, relatives := executorFixture()
	items := []string{"a b c", "x y z", "a b c", "x y z"}
	ex := newDefaultCandidateExecutor(candidates, blocks, lineLength, relatives, 200, 1)

	var orderSeen []int
	err := runCandidateBattery(context.Background(), ex, items, 4, func(i int, sequence string, r CandidateResult) error {
		orderSeen = append(orderSeen, i)
		if r.Candidate.Sequence != sequence {
			return fmt.Errorf("onReady got mismatched sequence/result: %q vs %+v", sequence, r.Candidate)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	for i, v := range orderSeen {
		if v != i {
			t.Fatalf("onReady was not called in ascending canonical order: %v", orderSeen)
		}
	}
}

func TestRunCandidateBatteryPropagatesWorkerError(t *testing.T) {
	candidates, blocks, lineLength, relatives := executorFixture()
	_, _, _, _ = candidates, blocks, lineLength, relatives
	stub := stubCandidateExecutor{fail: map[string]bool{"bad": true}}
	err := runCandidateBattery(context.Background(), stub, []string{"a b c", "bad"}, 2, func(int, string, CandidateResult) error { return nil })
	if err == nil {
		t.Fatal("expected an error to propagate from a failing candidate job")
	}
}

func TestRunCandidateBatteryPropagatesOnReadyError(t *testing.T) {
	stub := stubCandidateExecutor{}
	boom := errors.New("boom")
	err := runCandidateBattery(context.Background(), stub, []string{"a b c"}, 1, func(int, string, CandidateResult) error { return boom })
	if !errors.Is(err, boom) {
		t.Fatalf("expected onReady's error to propagate, got %v", err)
	}
}

// stubCandidateExecutor is a minimal CandidateExecutor for testing
// runCandidateBattery's dispatch/ordering/error-propagation in isolation
// from the real per-candidate computation.
type stubCandidateExecutor struct {
	fail map[string]bool
}

func (s stubCandidateExecutor) Run(_ context.Context, sequence string) (CandidateResult, error) {
	if s.fail[sequence] {
		return CandidateResult{}, fmt.Errorf("stub failure for %q", sequence)
	}
	return CandidateResult{Candidate: Candidate{Sequence: sequence}}, nil
}
