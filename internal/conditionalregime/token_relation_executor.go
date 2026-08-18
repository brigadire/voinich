package conditionalregime

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"

	"zcore.dev/voinich/internal/tokenrelationvalidation"
)

// tokenRelationInit builds the Init handshake for the
// token_relation_permutation workload (Task44): DiscoveryDir/Generic are
// carried explicitly so a worker never has to guess either from CorpusPath/
// TokenMetadataMap alone.
func tokenRelationInit(c tokenrelationvalidation.Config, fingerprint string) protocolMessage {
	return protocolMessage{
		Workload: "token_relation_permutation", Fingerprint: fingerprint,
		CorpusPath: c.CorpusPath, TokenMetadataMap: c.MetadataPath, DiscoveryDir: c.DiscoveryDir,
		Generic: c.Generic, Permutations: c.Permutations, RefinePermutations: c.RefinePermutations, Seed: c.Seed,
	}
}

type tokenRelationExecutorAdapter struct {
	pool jobExecutor
}

// NewTokenRelationProcessExecutor reuses the existing persistent subprocess
// protocol/pool (Task32); only the immutable workload descriptor and result
// payload differ from the other workloads it already serves.
func NewTokenRelationProcessExecutor(c tokenrelationvalidation.Config) (tokenrelationvalidation.PermutationExecutor, error) {
	fp, err := tokenrelationvalidation.Fingerprint(c)
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
	p, err := newProcessPool(c.Workers, newCmd, tokenRelationInit(c, fp))
	if err != nil {
		return nil, err
	}
	return &tokenRelationExecutorAdapter{pool: p}, nil
}

func NewTokenRelationRemoteExecutor(c tokenrelationvalidation.Config) (tokenrelationvalidation.PermutationExecutor, error) {
	fp, err := tokenrelationvalidation.Fingerprint(c)
	if err != nil {
		return nil, err
	}
	p, err := newTokenRelationRemotePool(c, fp)
	if err != nil {
		return nil, err
	}
	return &tokenRelationExecutorAdapter{pool: p}, nil
}

func (e *tokenRelationExecutorAdapter) Run(ctx context.Context, family string, run int) (map[string]float64, error) {
	id := JobID{Stage: "token_relation_permutation", Combination: family, ReplicateIndex: run}
	b, err := e.pool.RunBlob(ctx, id)
	if err != nil {
		return nil, err
	}
	var scores map[string]float64
	if err := json.Unmarshal(b, &scores); err != nil {
		return nil, fmt.Errorf("decode token-relation replicate %s/%d: %w", family, run, err)
	}
	return scores, nil
}
