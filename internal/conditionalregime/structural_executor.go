package conditionalregime

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"

	"zcore.dev/voinich/internal/structuralprojection"
)

type structuralExecutorAdapter struct {
	pool        jobExecutor
	fingerprint string
}

func structuralInit(c structuralprojection.Config, fingerprint string) protocolMessage {
	return protocolMessage{Workload: "structural_projection", Fingerprint: fingerprint, CorpusPath: c.CorpusPath, StructuralPairsPath: c.StructuralPairsPath,
		DistancePairsPath: c.DistancePairsPath, FamiliesPath: c.FamiliesPath, MinStructuralSimilarity: c.MinStructuralSimilarity, MinReliability: c.MinReliability,
		ProjectionK: c.ProjectionK, RandomProjections: c.RandomProjections, MaxDistance: c.MaxDistance, MinObservations: c.MinObservations, TopN: c.TopN,
		FamilyID: c.FamilyID, ProjectionMode: c.ProjectionMode, Pair: c.Pair, Seed: c.Seed}
}

// NewStructuralProcessExecutor reuses the existing persistent subprocess
// protocol/pool; only the immutable workload descriptor and result payload differ.
func NewStructuralProcessExecutor(c structuralprojection.Config) (structuralprojection.TrialExecutor, error) {
	self, err := os.Executable()
	if err != nil {
		return nil, err
	}
	fp, err := structuralprojection.Fingerprint(c)
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
	p, err := newProcessPool(c.Workers, newCmd, structuralInit(c, fp))
	if err != nil {
		return nil, err
	}
	return &structuralExecutorAdapter{pool: p, fingerprint: fp}, nil
}

func NewStructuralRemoteExecutor(c structuralprojection.Config) (structuralprojection.TrialExecutor, error) {
	fp, err := structuralprojection.Fingerprint(c)
	if err != nil {
		return nil, err
	}
	p, err := newStructuralRemotePool(c, fp)
	if err != nil {
		return nil, err
	}
	return &structuralExecutorAdapter{pool: p, fingerprint: fp}, nil
}

func (e *structuralExecutorAdapter) Run(ctx context.Context, trial int) (structuralprojection.TrialResult, error) {
	id := JobID{Stage: "structural_projection_trial", Combination: e.fingerprint, ReplicateIndex: trial}
	b, err := e.pool.RunBlob(ctx, id)
	if err != nil {
		return structuralprojection.TrialResult{}, err
	}
	var r structuralprojection.TrialResult
	if err := json.Unmarshal(b, &r); err != nil {
		return r, fmt.Errorf("decode structural trial %d: %w", trial, err)
	}
	return r, nil
}
func (e *structuralExecutorAdapter) Close() error { return e.pool.Close() }
func (e *structuralExecutorAdapter) TrialStats() (active, retries int) {
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
