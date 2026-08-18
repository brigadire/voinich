package conditionalregime

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"

	"zcore.dev/voinich/internal/positionalcontinuation"
)

// positionalContinuationInit builds the Init handshake for the
// positional_continuation_battery workload (Task44): HigherOrderDir names
// the reconstructed directory holding every staged
// higher-order-sequence-validate output file (see remote.go's
// newPositionalContinuationRemotePool).
func positionalContinuationInit(c positionalcontinuation.Config, fingerprint string) protocolMessage {
	return protocolMessage{
		Workload: "positional_continuation_battery", Fingerprint: fingerprint,
		CorpusPath: c.CorpusPath, TokenMetadataMap: c.TokenMetadataMap,
		HigherOrderDir: c.HigherOrderDir, Generic: c.Generic,
		Permutations: c.Permutations, Seed: c.Seed,
	}
}

type positionalContinuationExecutorAdapter struct {
	pool jobExecutor
}

// NewPositionalContinuationProcessExecutor reuses the existing persistent
// subprocess protocol/pool (Task32); only the immutable workload
// descriptor and result payload differ from the other workloads it
// already serves.
func NewPositionalContinuationProcessExecutor(c positionalcontinuation.Config) (positionalcontinuation.BatteryExecutor, error) {
	fp, err := positionalcontinuation.Fingerprint(c)
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
	p, err := newProcessPool(c.Workers, newCmd, positionalContinuationInit(c, fp))
	if err != nil {
		return nil, err
	}
	return &positionalContinuationExecutorAdapter{pool: p}, nil
}

func NewPositionalContinuationRemoteExecutor(c positionalcontinuation.Config) (positionalcontinuation.BatteryExecutor, error) {
	fp, err := positionalcontinuation.Fingerprint(c)
	if err != nil {
		return nil, err
	}
	p, err := newPositionalContinuationRemotePool(c, fp)
	if err != nil {
		return nil, err
	}
	return &positionalContinuationExecutorAdapter{pool: p}, nil
}

func (e *positionalContinuationExecutorAdapter) Run(ctx context.Context, battery string) (positionalcontinuation.BatteryResult, error) {
	id := JobID{Stage: "positional_continuation_battery", Combination: battery, ReplicateIndex: 0}
	b, err := e.pool.RunBlob(ctx, id)
	if err != nil {
		return positionalcontinuation.BatteryResult{}, err
	}
	var r positionalcontinuation.BatteryResult
	if err := json.Unmarshal(b, &r); err != nil {
		return positionalcontinuation.BatteryResult{}, fmt.Errorf("decode positional-continuation battery %q: %w", battery, err)
	}
	return r, nil
}
