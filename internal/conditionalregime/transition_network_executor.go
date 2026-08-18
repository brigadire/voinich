package conditionalregime

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"

	"zcore.dev/voinich/internal/transitionnetwork"
)

// transitionNetworkInit builds the Init handshake for the
// transition_network_permutation workload (Task44).
func transitionNetworkInit(c transitionnetwork.Config, fingerprint string) protocolMessage {
	return protocolMessage{
		Workload: "transition_network_permutation", Fingerprint: fingerprint,
		CorpusPath: c.CorpusPath, TokenMetadataMap: c.MetadataPath, Generic: c.Generic,
		MinTokenCount: c.MinTokenCount, MinBlockTokens: c.MinBlockTokenCount,
		Permutations: c.Permutations, RefinePermutations: c.RefinePermutations, Seed: c.Seed,
	}
}

type transitionNetworkExecutorAdapter struct {
	pool jobExecutor
}

// NewTransitionNetworkProcessExecutor reuses the existing persistent
// subprocess protocol/pool (Task32); only the immutable workload
// descriptor and result payload differ from the other workloads it
// already serves.
func NewTransitionNetworkProcessExecutor(c transitionnetwork.Config) (transitionnetwork.PermutationExecutor, error) {
	fp, err := transitionnetwork.Fingerprint(c)
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
	p, err := newProcessPool(c.Workers, newCmd, transitionNetworkInit(c, fp))
	if err != nil {
		return nil, err
	}
	return &transitionNetworkExecutorAdapter{pool: p}, nil
}

func NewTransitionNetworkRemoteExecutor(c transitionnetwork.Config) (transitionnetwork.PermutationExecutor, error) {
	fp, err := transitionnetwork.Fingerprint(c)
	if err != nil {
		return nil, err
	}
	p, err := newTransitionNetworkRemotePool(c, fp)
	if err != nil {
		return nil, err
	}
	return &transitionNetworkExecutorAdapter{pool: p}, nil
}

func (e *transitionNetworkExecutorAdapter) Run(ctx context.Context, phase string, rep int) (transitionnetwork.ReplicateResult, error) {
	id := JobID{Stage: "transition_network_permutation", Combination: phase, ReplicateIndex: rep}
	b, err := e.pool.RunBlob(ctx, id)
	if err != nil {
		return transitionnetwork.ReplicateResult{}, err
	}
	var r transitionnetwork.ReplicateResult
	if err := json.Unmarshal(b, &r); err != nil {
		return transitionnetwork.ReplicateResult{}, fmt.Errorf("decode transition-network replicate %s/%d: %w", phase, rep, err)
	}
	return r, nil
}
