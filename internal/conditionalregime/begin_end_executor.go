package conditionalregime

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"

	"zcore.dev/voinich/internal/beginendanalyze"
)

// resolvedBeginEndBatchSize returns c.CandidateBatchSize, or
// beginendanalyze.DefaultCandidateBatchSize if unset - the coordinator and
// every worker must agree on this exact value, since JobID.ReplicateIndex
// (the batch index) only determines a [lo,hi) pair range together with the
// batch size.
func resolvedBeginEndBatchSize(c beginendanalyze.Config) int {
	if c.CandidateBatchSize > 0 {
		return c.CandidateBatchSize
	}
	return beginendanalyze.DefaultCandidateBatchSize
}

// beginEndInit builds the Init handshake for the begin_end_candidate_batch
// workload (Task47).
func beginEndInit(c beginendanalyze.Config, fingerprint string) protocolMessage {
	return protocolMessage{
		Workload: "begin_end_candidate_batch", Fingerprint: fingerprint,
		CorpusPath: c.CorpusPath, DictionaryPath: c.DictionaryPath,
		MaxWindow: c.MaxWindow, Permutations: c.Permutations, MinTokenCount: c.MinTokenCount, Seed: c.RandomSeed,
		PermutationMode: c.PermutationMode, IncludeUnclear: c.IncludeUnclear, MaxCandidates: c.MaxCandidates,
		CandidateBatchSize: resolvedBeginEndBatchSize(c),
	}
}

type beginEndExecutorAdapter struct {
	pool jobExecutor
}

// NewBeginEndProcessExecutor reuses the existing persistent subprocess
// protocol/pool (Task32); only the immutable workload descriptor and result
// payload differ from the other workloads it already serves.
func NewBeginEndProcessExecutor(c beginendanalyze.Config) (beginendanalyze.CandidateBatchExecutor, error) {
	fp, err := beginendanalyze.Fingerprint(c)
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
	p, err := newProcessPool(c.Workers, newCmd, beginEndInit(c, fp))
	if err != nil {
		return nil, err
	}
	return &beginEndExecutorAdapter{pool: p}, nil
}

func NewBeginEndRemoteExecutor(c beginendanalyze.Config) (beginendanalyze.CandidateBatchExecutor, error) {
	fp, err := beginendanalyze.Fingerprint(c)
	if err != nil {
		return nil, err
	}
	p, err := newBeginEndRemotePool(c, fp)
	if err != nil {
		return nil, err
	}
	return &beginEndExecutorAdapter{pool: p}, nil
}

func (e *beginEndExecutorAdapter) Run(ctx context.Context, batchIndex int) (beginendanalyze.BatchResult, error) {
	id := JobID{Stage: "begin_end_candidate_batch", Combination: "candidates", ReplicateIndex: batchIndex}
	b, err := e.pool.RunBlob(ctx, id)
	if err != nil {
		return beginendanalyze.BatchResult{}, err
	}
	var wire wireBeginEndBatchResult
	if err := json.Unmarshal(b, &wire); err != nil {
		return beginendanalyze.BatchResult{}, fmt.Errorf("decode begin-end-analyze candidate batch %d: %w", batchIndex, err)
	}
	return wire.decode(), nil
}

func (e *beginEndExecutorAdapter) Close() error { return e.pool.Close() }

// begin-end-analyze token text is verbatim corpus/dictionary content and is
// not guaranteed to be valid UTF-8 (the Astafiev/Voynich corpora both
// contain non-UTF-8 byte sequences in some tokens). encoding/json silently
// replaces invalid UTF-8 in a Go string with U+FFFD when marshaling -
// harmless for every other workload's wire payloads (edge keys, numeric
// aggregates), but would silently corrupt BeginCandidate/EndCandidate for
// this one. wireCandidate/wireBeginEndBatchResult carry those two fields
// as []byte instead (json encodes []byte as base64, which round-trips any
// byte sequence exactly) - a pure transport-layer fix: beginendanalyze's
// own Candidate/BatchResult types (used for YAML output and the
// goroutine-backend in-process path, where no serialization ever happens)
// are completely unchanged.
type wireBeginEndBatchResult struct {
	Candidates []wireCandidate `json:"candidates"`
}

type wireCandidate struct {
	beginendanalyze.Candidate
	BeginCandidateRaw []byte `json:"begin_candidate_raw"`
	EndCandidateRaw   []byte `json:"end_candidate_raw"`
}

func encodeBeginEndBatchResult(r beginendanalyze.BatchResult) wireBeginEndBatchResult {
	out := wireBeginEndBatchResult{Candidates: make([]wireCandidate, len(r.Candidates))}
	for i, c := range r.Candidates {
		beginRaw, endRaw := []byte(c.BeginCandidate), []byte(c.EndCandidate)
		c.BeginCandidate, c.EndCandidate = "", ""
		out.Candidates[i] = wireCandidate{Candidate: c, BeginCandidateRaw: beginRaw, EndCandidateRaw: endRaw}
	}
	return out
}

func (w wireBeginEndBatchResult) decode() beginendanalyze.BatchResult {
	out := beginendanalyze.BatchResult{Candidates: make([]beginendanalyze.Candidate, len(w.Candidates))}
	for i, wc := range w.Candidates {
		c := wc.Candidate
		c.BeginCandidate, c.EndCandidate = string(wc.BeginCandidateRaw), string(wc.EndCandidateRaw)
		out.Candidates[i] = c
	}
	return out
}
