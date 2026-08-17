package conditionalregime

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"

	"zcore.dev/voinich/internal/normalizationcompare"
)

// normalizationInit builds the Init handshake for the normalization_compare
// workload (Task42): MinTokenCount/SingletonMode are read from classes.yaml
// itself, never assumed by the worker, exactly as classInventory reads its
// own eligibility thresholds from Config rather than guessing them.
func normalizationInit(c normalizationcompare.Config, classesMinTokenCount int, classesSingletonMode, fingerprint string) protocolMessage {
	return protocolMessage{
		Workload: "normalization_compare", Fingerprint: fingerprint,
		CorpusPath: c.InputPath, ClassesPath: c.ClassesPath,
		MinTokenCount: classesMinTokenCount, SingletonMode: classesSingletonMode, Seed: c.RandomSeed, RandomRuns: c.RandomRuns,
	}
}

type normalizationExecutorAdapter struct {
	pool        jobExecutor
	fingerprint string
}

// NewNormalizationProcessExecutor reuses the existing persistent subprocess
// protocol/pool (Task32); only the immutable workload descriptor and result
// payload differ from the conditional-regime and structural-projection
// workloads it already serves.
func NewNormalizationProcessExecutor(c normalizationcompare.Config) (normalizationcompare.BaselineExecutor, error) {
	classes, err := normalizationcompare.LoadClasses(c.ClassesPath)
	if err != nil {
		return nil, err
	}
	fp, err := normalizationcompare.Fingerprint(c.InputPath, c.ClassesPath, classes.Meta.MinTokenCount, classes.Meta.SingletonMode, c.RandomSeed, c.RandomRuns)
	if err != nil {
		return nil, err
	}
	self, err := os.Executable()
	if err != nil {
		return nil, err
	}
	ctx := c.Context
	if ctx == nil {
		ctx = context.Background()
	}
	newCmd := func() *exec.Cmd {
		cmd := exec.CommandContext(ctx, self, "-internal-worker")
		cmd.Stderr = os.Stderr
		return cmd
	}
	p, err := newProcessPool(c.Workers, newCmd, normalizationInit(c, classes.Meta.MinTokenCount, classes.Meta.SingletonMode, fp))
	if err != nil {
		return nil, err
	}
	return &normalizationExecutorAdapter{pool: p, fingerprint: fp}, nil
}

func NewNormalizationRemoteExecutor(c normalizationcompare.Config) (normalizationcompare.BaselineExecutor, error) {
	classes, err := normalizationcompare.LoadClasses(c.ClassesPath)
	if err != nil {
		return nil, err
	}
	fp, err := normalizationcompare.Fingerprint(c.InputPath, c.ClassesPath, classes.Meta.MinTokenCount, classes.Meta.SingletonMode, c.RandomSeed, c.RandomRuns)
	if err != nil {
		return nil, err
	}
	p, err := newNormalizationRemotePool(c, classes, fp)
	if err != nil {
		return nil, err
	}
	return &normalizationExecutorAdapter{pool: p, fingerprint: fp}, nil
}

func (e *normalizationExecutorAdapter) Run(ctx context.Context, threshold string, run int) (normalizationcompare.BaselineResult, error) {
	id := JobID{Stage: "normalization_compare_baseline", Combination: threshold, ReplicateIndex: run}
	b, err := e.pool.RunBlob(ctx, id)
	if err != nil {
		return normalizationcompare.BaselineResult{}, err
	}
	var r normalizationcompare.BaselineResult
	if err := json.Unmarshal(b, &r); err != nil {
		return r, fmt.Errorf("decode normalization baseline %s/%d: %w", threshold, run, err)
	}
	return r, nil
}
func (e *normalizationExecutorAdapter) Close() error { return e.pool.Close() }
func (e *normalizationExecutorAdapter) BaselineStats() (active, retries int) {
	if p, ok := e.pool.(*remotePool); ok {
		p.mu.Lock()
		active = len(p.leases)
		p.mu.Unlock()
		retries = int(p.leasesReclaimed.Load())
	} else if p, ok := e.pool.(*processPool); ok {
		active = len(p.workers) - len(p.free)
	}
	return
}
