package conditionalregime

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"sync"
)

// JobID is the stable identity of one independently reproducible
// permutation replicate. It deliberately contains no worker or scheduling
// information: the same scientific job has the same identity in every run.
type JobID struct {
	Stage          string
	Combination    string
	ReplicateIndex int
}

// JobResult is the existing per-replicate scientific result tagged with the
// identity needed to restore canonical replicate-index order.
type JobResult struct {
	JobID JobID
	Value float64
}

func checkpointJobKey(id JobID) string {
	return id.Stage + "\x1f" + id.Combination + "\x1f" + strconv.Itoa(id.ReplicateIndex)
}

func checkpointJobPrefix(stage, combination string) string {
	return stage + "\x1f" + combination + "\x1f"
}

func checkpointJobsFor(ids map[string]float64, stage, combination string, permutations int) map[int]float64 {
	out := map[int]float64{}
	prefix := checkpointJobPrefix(stage, combination)
	for key, value := range ids {
		if !strings.HasPrefix(key, prefix) {
			continue
		}
		i, err := strconv.Atoi(strings.TrimPrefix(key, prefix))
		if err == nil && i >= 0 && i < permutations {
			out[i] = value
		}
	}
	return out
}

type permutationJob struct{ JobID JobID }

// executePermutationJobs runs a bounded worker pool. Results may reach
// onComplete in any order, but the returned slice is always in jobs order.
// Only the coordinator goroutine invokes onComplete.
func executePermutationJobs(
	ctx context.Context,
	workers int,
	jobs []permutationJob,
	work func(context.Context, JobID) (float64, error),
	onComplete func(JobResult) error,
) ([]JobResult, error) {
	if workers < 1 {
		return nil, fmt.Errorf("workers must be positive")
	}
	if ctx == nil {
		ctx = context.Background()
	}
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	slots := make(map[JobID]int, len(jobs))
	for i, job := range jobs {
		if _, exists := slots[job.JobID]; exists {
			return nil, fmt.Errorf("duplicate permutation job ID: %+v", job.JobID)
		}
		slots[job.JobID] = i
	}
	if len(jobs) == 0 {
		return []JobResult{}, nil
	}
	if workers > len(jobs) {
		workers = len(jobs)
	}

	type completed struct {
		result JobResult
		err    error
	}
	jobCh := make(chan permutationJob)
	resultCh := make(chan completed, workers)
	var wg sync.WaitGroup
	wg.Add(workers)
	for i := 0; i < workers; i++ {
		go func() {
			defer wg.Done()
			for {
				select {
				case <-ctx.Done():
					return
				case job, ok := <-jobCh:
					if !ok {
						return
					}
					value, err := work(ctx, job.JobID)
					c := completed{result: JobResult{JobID: job.JobID, Value: value}, err: err}
					select {
					case resultCh <- c:
					case <-ctx.Done():
						return
					}
					if err != nil {
						return
					}
				}
			}
		}()
	}
	go func() {
		defer close(jobCh)
		for _, job := range jobs {
			select {
			case jobCh <- job:
			case <-ctx.Done():
				return
			}
		}
	}()
	go func() {
		wg.Wait()
		close(resultCh)
	}()

	results := make([]JobResult, len(jobs))
	seen := make([]bool, len(jobs))
	completedCount := 0
	for c := range resultCh {
		if c.err != nil {
			cancel()
			for range resultCh {
			}
			return nil, fmt.Errorf("job %+v: %w", c.result.JobID, c.err)
		}
		slot, ok := slots[c.result.JobID]
		if !ok {
			cancel()
			for range resultCh {
			}
			return nil, fmt.Errorf("unknown completed job ID: %+v", c.result.JobID)
		}
		if seen[slot] { // duplicate delivery is idempotent
			continue
		}
		seen[slot] = true
		results[slot] = c.result
		completedCount++
		if onComplete != nil {
			if err := onComplete(c.result); err != nil {
				cancel()
				for range resultCh {
				}
				return nil, err
			}
		}
	}
	if completedCount != len(jobs) {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		return nil, fmt.Errorf("worker pool stopped after %d of %d jobs", completedCount, len(jobs))
	}
	return results, nil
}

// runIndexedReplicates restores out-of-order completions to exact replicate
// order. onPrefix only sees the longest contiguous prefix, which makes an
// interrupted checkpoint safe to resume with any worker count.
func runIndexedReplicates(
	ctx context.Context,
	workers int,
	stage, combination string,
	permutations int,
	resume []float64,
	work func(context.Context, int) (float64, error),
	onPrefix func([]float64),
) ([]float64, error) {
	return runIndexedReplicatesState(ctx, workers, stage, combination, permutations, resume, nil, work, onPrefix, nil)
}

func runIndexedReplicatesState(
	ctx context.Context,
	workers int,
	stage, combination string,
	permutations int,
	resume []float64,
	completed map[int]float64,
	work func(context.Context, int) (float64, error),
	onPrefix func([]float64),
	onComplete func(JobResult),
) ([]float64, error) {
	if len(resume) > permutations {
		resume = resume[:permutations]
	}
	values := make([]float64, permutations)
	copy(values, resume)
	next := len(resume)
	ready := make([]bool, permutations)
	for i := 0; i < next; i++ {
		ready[i] = true
	}
	for i, value := range completed {
		if i >= next && i < permutations {
			values[i] = value
			ready[i] = true
		}
	}
	for next < permutations && ready[next] {
		next++
	}
	jobs := make([]permutationJob, 0, permutations-next)
	for i := len(resume); i < permutations; i++ {
		if !ready[i] {
			jobs = append(jobs, permutationJob{JobID{Stage: stage, Combination: combination, ReplicateIndex: i}})
		}
	}
	_, err := executePermutationJobs(ctx, workers, jobs, func(ctx context.Context, id JobID) (float64, error) {
		return work(ctx, id.ReplicateIndex)
	}, func(result JobResult) error {
		i := result.JobID.ReplicateIndex
		values[i] = result.Value
		ready[i] = true
		if onComplete != nil {
			onComplete(result)
		}
		oldNext := next
		for next < permutations && ready[next] {
			next++
		}
		if next > oldNext && onPrefix != nil {
			onPrefix(values[:next])
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return values, nil
}
