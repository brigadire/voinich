package beginendanalyze

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"testing"
	"time"
)

func testWorkspace(t *testing.T) *Workspace {
	t.Helper()
	dictionary, corpusPath := testInputs(t, strings.Repeat("a b c d\nb c d a\nc d a b\n\n", 4))
	ws, err := LoadForDistribution(Config{DictionaryPath: dictionary, CorpusPath: corpusPath, MaxWindow: 3, Permutations: 5, MinTokenCount: 1, RandomSeed: 1, PermutationMode: "page", MaxCandidates: 20})
	if err != nil {
		t.Fatal(err)
	}
	return ws
}

func TestComputeBatchMatchesDirectCandidateAt(t *testing.T) {
	ws := testWorkspace(t)
	total := ws.TotalPairs()
	got := ComputeBatch(ws, 0, total)
	want := make([]Candidate, 0)
	for idx := 0; idx < total; idx++ {
		ai, bi := idx/ws.k, idx%ws.k
		if ai == bi {
			continue
		}
		want = append(want, ws.candidateAt(ai, bi))
	}
	if !reflect.DeepEqual(got.Candidates, want) {
		t.Fatalf("single-batch ComputeBatch diverged from per-pair candidateAt")
	}
}

func TestComputeBatchPartitioningMatchesSingleBatch(t *testing.T) {
	ws := testWorkspace(t)
	total := ws.TotalPairs()
	whole := ComputeBatch(ws, 0, total)

	for _, batchSize := range []int{1, 2, 3, 5} {
		numBatches := (total + batchSize - 1) / batchSize
		var got []Candidate
		for b := 0; b < numBatches; b++ {
			got = append(got, ComputeBatch(ws, b, batchSize).Candidates...)
		}
		if !reflect.DeepEqual(got, whole.Candidates) {
			t.Fatalf("batchSize=%d: concatenated batches diverged from single whole-space batch", batchSize)
		}
	}
}

func TestComputeBatchOutOfRangeIsEmpty(t *testing.T) {
	ws := testWorkspace(t)
	total := ws.TotalPairs()
	numBatches := (total + 9) / 10
	r := ComputeBatch(ws, numBatches+5, 10)
	if len(r.Candidates) != 0 {
		t.Fatalf("out-of-range batch should be empty, got %d candidates", len(r.Candidates))
	}
}

type stubBatchExecutor struct {
	delay func(batch int) time.Duration
	fail  map[int]bool
}

func (s *stubBatchExecutor) Run(ctx context.Context, batch int) (BatchResult, error) {
	if s.delay != nil {
		select {
		case <-time.After(s.delay(batch)):
		case <-ctx.Done():
			return BatchResult{}, ctx.Err()
		}
	}
	if s.fail[batch] {
		return BatchResult{}, fmt.Errorf("synthetic failure at batch %d", batch)
	}
	return BatchResult{Candidates: []Candidate{{BeginCandidate: fmt.Sprintf("batch-%d", batch)}}}, nil
}

func TestRunCandidateBatchesRestoresCanonicalOrderRegardlessOfCompletionOrder(t *testing.T) {
	n := 12
	exec := &stubBatchExecutor{delay: func(batch int) time.Duration { return time.Duration(n-batch) * time.Millisecond }}
	var got []int
	err := runCandidateBatches(context.Background(), exec, n, 6, func(batch int, r BatchResult) error {
		got = append(got, batch)
		if r.Candidates[0].BeginCandidate != fmt.Sprintf("batch-%d", batch) {
			t.Fatalf("batch %d: got %v", batch, r.Candidates)
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

func TestRunCandidateBatchesWorkerCountDoesNotChangeResult(t *testing.T) {
	n := 30
	collect := func(workers int) []string {
		exec := &stubBatchExecutor{}
		var names []string
		if err := runCandidateBatches(context.Background(), exec, n, workers, func(_ int, r BatchResult) error {
			names = append(names, r.Candidates[0].BeginCandidate)
			return nil
		}); err != nil {
			t.Fatal(err)
		}
		return names
	}
	w1, w4, w16 := collect(1), collect(4), collect(16)
	if !reflect.DeepEqual(w1, w4) || !reflect.DeepEqual(w4, w16) {
		t.Fatalf("result depends on worker count: 1->%v 4->%v 16->%v", w1, w4, w16)
	}
}

func TestRunCandidateBatchesPropagatesWorkerError(t *testing.T) {
	exec := &stubBatchExecutor{fail: map[int]bool{3: true}}
	if err := runCandidateBatches(context.Background(), exec, 8, 4, func(int, BatchResult) error { return nil }); err == nil {
		t.Fatal("expected the synthetic failure to propagate")
	}
}

func TestRunCandidateBatchesPropagatesOnReadyError(t *testing.T) {
	sentinel := errors.New("write failed")
	exec := &stubBatchExecutor{}
	err := runCandidateBatches(context.Background(), exec, 8, 4, func(batch int, r BatchResult) error {
		if batch == 2 {
			return sentinel
		}
		return nil
	})
	if !errors.Is(err, sentinel) {
		t.Fatalf("got %v, want sentinel error propagated", err)
	}
}

func TestDefaultCandidateBatchExecutorMatchesComputeBatch(t *testing.T) {
	ws := testWorkspace(t)
	exec := &defaultCandidateBatchExecutor{ws: ws, batchSize: 2}
	for _, batch := range []int{0, 1, 3} {
		got, err := exec.Run(context.Background(), batch)
		if err != nil {
			t.Fatal(err)
		}
		want := ComputeBatch(ws, batch, 2)
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("batch=%d: default executor diverged from ComputeBatch", batch)
		}
	}
}

func TestDefaultCandidateBatchExecutorConcurrencySafe(t *testing.T) {
	ws := testWorkspace(t)
	total := ws.TotalPairs()
	batchSize := 3
	numBatches := (total + batchSize - 1) / batchSize

	seq := &defaultCandidateBatchExecutor{ws: ws, batchSize: batchSize}
	var seqCandidates []Candidate
	for b := 0; b < numBatches; b++ {
		r, err := seq.Run(context.Background(), b)
		if err != nil {
			t.Fatal(err)
		}
		seqCandidates = append(seqCandidates, r.Candidates...)
	}

	par := &defaultCandidateBatchExecutor{ws: ws, batchSize: batchSize}
	var parCandidates []Candidate
	err := runCandidateBatches(context.Background(), par, numBatches, 8, func(_ int, r BatchResult) error {
		parCandidates = append(parCandidates, r.Candidates...)
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(seqCandidates, parCandidates) {
		t.Fatal("concurrent batch execution over a shared Workspace diverged from sequential execution")
	}
}
