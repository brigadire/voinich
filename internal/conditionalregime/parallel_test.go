package conditionalregime

import (
	"context"
	"errors"
	"reflect"
	"sync/atomic"
	"testing"
	"time"
)

func deterministicTestReplicate(_ context.Context, i int) (float64, error) {
	// Force a different completion order from replicate order when several
	// workers are active, without introducing any nondeterminism in values.
	time.Sleep(time.Duration((11-i)%4) * time.Millisecond)
	return float64(i*i+3*i) / 7, nil
}

func TestJobIDDeterministicAndWorkerIndependent(t *testing.T) {
	a := JobID{Stage: "part_a_significance", Combination: "joint|A/1|50|k_medoids", ReplicateIndex: 17}
	b := JobID{Stage: "part_a_significance", Combination: "joint|A/1|50|k_medoids", ReplicateIndex: 17}
	if a != b {
		t.Fatalf("identical scientific coordinates produced different job IDs: %+v != %+v", a, b)
	}

	one, err := runIndexedReplicates(context.Background(), 1, a.Stage, a.Combination, 12, nil, deterministicTestReplicate, nil)
	if err != nil {
		t.Fatal(err)
	}
	four, err := runIndexedReplicates(context.Background(), 4, a.Stage, a.Combination, 12, nil, deterministicTestReplicate, nil)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(one, four) {
		t.Fatalf("worker count changed indexed results:\n1=%v\n4=%v", one, four)
	}
	if buildEmpiricalStats(9, one) != buildEmpiricalStats(9, four) {
		t.Fatal("worker count changed the order-sensitive deterministic reduction")
	}
}

func TestCompletionOrderNeverBecomesReductionOrder(t *testing.T) {
	var completion []int
	values, err := runIndexedReplicates(context.Background(), 4, "stage", "combo", 12, nil, deterministicTestReplicate, func(prefix []float64) {
		completion = append(completion, len(prefix))
	})
	if err != nil {
		t.Fatal(err)
	}
	for i, got := range values {
		want, _ := deterministicTestReplicate(context.Background(), i)
		if got != want {
			t.Fatalf("slot %d = %v, want %v", i, got, want)
		}
	}
	if len(completion) == 0 || completion[len(completion)-1] != len(values) {
		t.Fatalf("checkpoint prefixes did not finish at the canonical length: %v", completion)
	}
}

func TestIndexedResumeWithChangedWorkerCount(t *testing.T) {
	full, err := runIndexedReplicates(context.Background(), 1, "stage", "combo", 14, nil, deterministicTestReplicate, nil)
	if err != nil {
		t.Fatal(err)
	}
	partial := append([]float64(nil), full[:5]...)
	var lastPrefix []float64
	resumed, err := runIndexedReplicates(context.Background(), 4, "stage", "combo", 14, partial, deterministicTestReplicate, func(prefix []float64) {
		lastPrefix = append([]float64(nil), prefix...)
	})
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(full, resumed) || !reflect.DeepEqual(full, lastPrefix) {
		t.Fatalf("changed-worker resume differed from uninterrupted run:\nfull=%v\nresumed=%v\ncheckpoint=%v", full, resumed, lastPrefix)
	}
}

func TestOutOfOrderCheckpointJobsResumeWithoutDuplicateExecution(t *testing.T) {
	completed := map[int]float64{1: 4, 3: 8}
	var executed atomic.Int64
	values, err := runIndexedReplicatesState(context.Background(), 3, nil, "stage", "combo", 5, []float64{2}, completed, func(_ context.Context, i int) (float64, error) {
		executed.Add(1)
		return float64(2 * (i + 1)), nil
	}, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	want := []float64{2, 4, 6, 8, 10}
	if !reflect.DeepEqual(values, want) {
		t.Fatalf("out-of-order checkpoint restore = %v, want %v", values, want)
	}
	if executed.Load() != 2 { // only missing JobIDs 2 and 4
		t.Fatalf("resume re-executed completed JobIDs; ran %d jobs", executed.Load())
	}
}

func TestWorkerErrorCancelsPool(t *testing.T) {
	boom := errors.New("boom")
	var started atomic.Int64
	jobs := make([]permutationJob, 100)
	for i := range jobs {
		jobs[i] = permutationJob{JobID{Stage: "stage", Combination: "combo", ReplicateIndex: i}}
	}
	_, err := executePermutationJobs(context.Background(), 4, jobs, func(ctx context.Context, id JobID) (float64, error) {
		started.Add(1)
		if id.ReplicateIndex == 0 {
			return 0, boom
		}
		select {
		case <-ctx.Done():
			return 0, ctx.Err()
		case <-time.After(100 * time.Millisecond):
			return float64(id.ReplicateIndex), nil
		}
	}, nil)
	if !errors.Is(err, boom) {
		t.Fatalf("expected fatal job error, got %v", err)
	}
	if got := started.Load(); got >= int64(len(jobs)) {
		t.Fatalf("fatal error failed to cancel queued work; started %d jobs", got)
	}
}

func TestDuplicateJobIDRejected(t *testing.T) {
	id := JobID{Stage: "stage", Combination: "combo", ReplicateIndex: 1}
	jobs := []permutationJob{{id}, {id}}
	_, err := executePermutationJobs(context.Background(), 2, jobs, func(context.Context, JobID) (float64, error) { return 0, nil }, nil)
	if err == nil {
		t.Fatal("duplicate job identity must not affect final results silently")
	}
}
