package conditionalregime

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"

	"zcore.dev/voinich/internal/higherorderseq"
)

// higherOrderInit builds the Init handshake for the higher_order_candidate
// workload (Task44): AuditDir/DiscoveryDir name the reconstructed
// directories holding every staged replicated-local-structure-audit output
// file / structural_classes.yaml (see remote.go's newHigherOrderRemotePool).
func higherOrderInit(c higherorderseq.Config, fingerprint string) protocolMessage {
	return protocolMessage{
		Workload: "higher_order_candidate", Fingerprint: fingerprint,
		CorpusPath: c.CorpusPath, TokenMetadataMap: c.TokenMetadataMap,
		AuditDir: c.AuditDir, DiscoveryDir: c.DiscoveryDir, Generic: c.Generic,
		Permutations: c.Permutations, Seed: c.Seed,
	}
}

type higherOrderExecutorAdapter struct {
	pool jobExecutor
}

// NewHigherOrderProcessExecutor reuses the existing persistent subprocess
// protocol/pool (Task32); only the immutable workload descriptor and result
// payload differ from the other workloads it already serves.
func NewHigherOrderProcessExecutor(c higherorderseq.Config) (higherorderseq.CandidateExecutor, error) {
	fp, err := higherorderseq.Fingerprint(c)
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
	p, err := newProcessPool(c.Workers, newCmd, higherOrderInit(c, fp))
	if err != nil {
		return nil, err
	}
	return &higherOrderExecutorAdapter{pool: p}, nil
}

func NewHigherOrderRemoteExecutor(c higherorderseq.Config) (higherorderseq.CandidateExecutor, error) {
	fp, err := higherorderseq.Fingerprint(c)
	if err != nil {
		return nil, err
	}
	p, err := newHigherOrderRemotePool(c, fp)
	if err != nil {
		return nil, err
	}
	return &higherOrderExecutorAdapter{pool: p}, nil
}

func (e *higherOrderExecutorAdapter) Run(ctx context.Context, sequence string) (higherorderseq.CandidateResult, error) {
	id := JobID{Stage: "higher_order_candidate", Combination: sequence, ReplicateIndex: 0}
	b, err := e.pool.RunBlob(ctx, id)
	if err != nil {
		return higherorderseq.CandidateResult{}, err
	}
	var r higherorderseq.CandidateResult
	if err := json.Unmarshal(b, &r); err != nil {
		return higherorderseq.CandidateResult{}, fmt.Errorf("decode higher-order-sequence candidate %q: %w", sequence, err)
	}
	return r, nil
}
