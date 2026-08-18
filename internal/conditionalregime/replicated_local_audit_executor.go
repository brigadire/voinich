package conditionalregime

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"

	"zcore.dev/voinich/internal/replicatedlocalaudit"
)

// replicatedLocalAuditInit builds the Init handshake for the
// replicated_local_null workload (Task44): RelationDir/DiscoveryDir name
// the reconstructed directories holding every staged relation-dir/
// discovery-dir file (see remote.go's newReplicatedLocalAuditRemotePool).
func replicatedLocalAuditInit(c replicatedlocalaudit.Config, fingerprint string) protocolMessage {
	return protocolMessage{
		Workload: "replicated_local_null", Fingerprint: fingerprint,
		CorpusPath: c.CorpusPath, TokenMetadataMap: c.MetadataPath,
		RelationDir: c.RelationDir, DiscoveryDir: c.DiscoveryDir, Generic: c.Generic,
		Permutations: c.Permutations, Seed: c.Seed,
	}
}

type replicatedLocalAuditExecutorAdapter struct {
	pool jobExecutor
}

// NewReplicatedLocalAuditProcessExecutor reuses the existing persistent
// subprocess protocol/pool (Task32); only the immutable workload
// descriptor and result payload differ from the other workloads it
// already serves.
func NewReplicatedLocalAuditProcessExecutor(c replicatedlocalaudit.Config) (replicatedlocalaudit.PermutationExecutor, error) {
	fp, err := replicatedlocalaudit.Fingerprint(c)
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
	p, err := newProcessPool(c.Workers, newCmd, replicatedLocalAuditInit(c, fp))
	if err != nil {
		return nil, err
	}
	return &replicatedLocalAuditExecutorAdapter{pool: p}, nil
}

func NewReplicatedLocalAuditRemoteExecutor(c replicatedlocalaudit.Config) (replicatedlocalaudit.PermutationExecutor, error) {
	fp, err := replicatedlocalaudit.Fingerprint(c)
	if err != nil {
		return nil, err
	}
	p, err := newReplicatedLocalAuditRemotePool(c, fp)
	if err != nil {
		return nil, err
	}
	return &replicatedLocalAuditExecutorAdapter{pool: p}, nil
}

func (e *replicatedLocalAuditExecutorAdapter) Run(ctx context.Context, phase string, run int) (replicatedlocalaudit.ReplicateResult, error) {
	id := JobID{Stage: "replicated_local_null", Combination: phase, ReplicateIndex: run}
	b, err := e.pool.RunBlob(ctx, id)
	if err != nil {
		return replicatedlocalaudit.ReplicateResult{}, err
	}
	var r replicatedlocalaudit.ReplicateResult
	if err := json.Unmarshal(b, &r); err != nil {
		return replicatedlocalaudit.ReplicateResult{}, fmt.Errorf("decode replicated-local-audit replicate %s/%d: %w", phase, run, err)
	}
	return r, nil
}
