package conditionalregime

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"strconv"
	"strings"
	"sync"
)

func writeMessage(w *bufio.Writer, m protocolMessage) error {
	b, err := json.Marshal(m)
	if err != nil {
		return err
	}
	if _, err := w.Write(b); err != nil {
		return err
	}
	if err := w.WriteByte('\n'); err != nil {
		return err
	}
	return w.Flush()
}

// readMessage returns (msg, true, nil) for a well-formed line, (zero, false,
// nil) on a clean EOF (the peer closed its side) and (zero, true, err) for a
// malformed line: callers must treat all three cases differently rather
// than folding EOF and malformed input into one "no message" outcome.
func readMessage(s *bufio.Scanner) (protocolMessage, bool, error) {
	if !s.Scan() {
		return protocolMessage{}, false, s.Err()
	}
	var m protocolMessage
	if err := json.Unmarshal(s.Bytes(), &m); err != nil {
		return protocolMessage{}, true, fmt.Errorf("malformed protocol message: %w", err)
	}
	return m, true, nil
}

// workerSweep caches the deterministic within-class discovery sweep for one
// (scheme, class label, window size): both part_a_significance and
// part_a_refinement jobs for every replicate of every method at that combo
// need only the resulting best-K per method, not a fresh sweep per
// replicate. The sweep is a pure function of (class, blocks, windowSize,
// KMin, KMaxWithin, Seed), so caching it changes nothing about the result -
// it only avoids recomputing an expensive fit thousands of times per worker.
type workerSweep struct {
	best   map[string]WithinClassRegime
	blocks []Block
}

// workerState is the read-only-after-init context one worker process builds
// once from the coordinator's Init message, then reuses across every Job it
// serves. It is deliberately built by calling the exact same unexported
// functions run.go itself calls (buildAllBlocks, classInventory,
// withinClassSweep, bestByMethod, nullSilhouetteAtK, residualNullMax): a
// subprocess worker is not a separate scientific implementation, it is the
// same package's code running in a different OS process.
type workerState struct {
	tokens         []string
	init           protocolMessage
	blocksByScheme map[Scheme]map[ClassID][]Block
	jointEligible  []ClassID
	jointBlocks    map[ClassID][]Block
	sweepCache     map[string]workerSweep
	sweepMu        sync.Mutex
}

func classIDFromLabel(scheme Scheme, label string) ClassID {
	switch scheme {
	case SchemeCurrierOnly:
		return ClassID{Scheme: scheme, Currier: label}
	case SchemeHandOnly:
		return ClassID{Scheme: scheme, Hand: label}
	default:
		currier, hand, _ := strings.Cut(label, "/")
		return ClassID{Scheme: scheme, Currier: currier, Hand: hand}
	}
}

// computePartA serves both part_a_significance and part_a_refinement: they
// share the same combination shape (scheme|classLabel|windowSize|method) and
// the same observed best-K, differing only in the seed/salt offset already
// fixed by nullmodels.go's withinClassSignificanceParallel/
// refineTopCandidatesParallel.
func (w *workerState) computePartA(id JobID) (float64, error) {
	parts := strings.SplitN(id.Combination, "|", 4)
	if len(parts) != 4 {
		return 0, fmt.Errorf("malformed part_a combination %q", id.Combination)
	}
	scheme := Scheme(parts[0])
	class := classIDFromLabel(scheme, parts[1])
	windowSize, err := strconv.Atoi(parts[2])
	if err != nil {
		return 0, fmt.Errorf("malformed part_a window size in %q: %w", id.Combination, err)
	}
	method := parts[3]

	sweepKey := parts[0] + "|" + parts[1] + "|" + parts[2]
	w.sweepMu.Lock()
	sweep, ok := w.sweepCache[sweepKey]
	if !ok {
		blocks := w.blocksByScheme[scheme][class]
		rows, _ := withinClassSweep(w.tokens, class, blocks, windowSize, w.init.KMin, w.init.KMaxWithin, w.init.Seed)
		sweep = workerSweep{best: bestByMethod(rows), blocks: blocks}
		w.sweepCache[sweepKey] = sweep
	}
	w.sweepMu.Unlock()
	row, ok := sweep.best[method]
	if !ok {
		return 0, fmt.Errorf("no observed row for method %q at combination %q", method, id.Combination)
	}

	seed, salt := w.init.Seed, methodSalt(method)
	switch id.Stage {
	case "part_a_significance":
	case "part_a_refinement":
		seed += 999983
		salt += 100
	default:
		return 0, fmt.Errorf("computePartA invoked for stage %q", id.Stage)
	}
	rng := rand.New(rand.NewSource(replicateSeed(seed, salt, id.ReplicateIndex)))
	return nullSilhouetteAtK(w.tokens, sweep.blocks, windowSize, method, row.K, rng), nil
}

// partBCorrectionIndex must match run.go's fixed correctionTargets order
// (k_medoids first, hierarchical second): that index, not the method name,
// is what run.go adds to the base seed before deriving each replicate's rng.
func partBCorrectionIndex(method string) (int64, error) {
	switch method {
	case "k_medoids":
		return 0, nil
	case "hierarchical":
		return 1, nil
	default:
		return 0, fmt.Errorf("unknown part_b method %q", method)
	}
}

func (w *workerState) computePartB(id JobID) (float64, error) {
	parts := strings.SplitN(id.Combination, "|", 2)
	if len(parts) != 2 {
		return 0, fmt.Errorf("malformed part_b combination %q", id.Combination)
	}
	method, standardized := parts[0], parts[1] == "standardized"
	correctionIndex, err := partBCorrectionIndex(method)
	if err != nil {
		return 0, err
	}
	seed := w.init.Seed + correctionIndex
	rng := rand.New(rand.NewSource(replicateSeed(seed, methodSalt(method), id.ReplicateIndex)))
	return residualNullMax(w.tokens, w.jointEligible, w.jointBlocks, w.init.ResidualWindowSizes, w.init.KMin, w.init.KMaxResidual, method, standardized, rng), nil
}

func (w *workerState) compute(id JobID) (float64, error) {
	switch id.Stage {
	case "part_a_significance", "part_a_refinement":
		return w.computePartA(id)
	case "part_b_global_correction":
		return w.computePartB(id)
	default:
		return 0, fmt.Errorf("unknown job stage %q", id.Stage)
	}
}

func newWorkerState(initMsg protocolMessage) (*workerState, error) {
	tokens, corpusHash, err := readCorpus(initMsg.CorpusPath)
	if err != nil {
		return nil, fmt.Errorf("read corpus: %w", err)
	}
	currier, hand, metaHash, err := loadTokenLabels(initMsg.TokenMetadataMap)
	if err != nil {
		return nil, fmt.Errorf("load token metadata map: %w", err)
	}
	if len(currier) != len(tokens) {
		return nil, fmt.Errorf("token metadata map has %d tokens but corpus has %d", len(currier), len(tokens))
	}
	if fp := computeFingerprint(initMsg.scientificConfig(), corpusHash, metaHash); fp != initMsg.Fingerprint {
		return nil, fmt.Errorf("input/config fingerprint mismatch: worker loaded different input or parameters")
	}
	allBlocks := buildAllBlocks(currier, hand)
	w := &workerState{
		tokens: tokens, init: initMsg,
		blocksByScheme: map[Scheme]map[ClassID][]Block{
			SchemeJoint: blocksByClass(allBlocks[SchemeJoint]), SchemeCurrierOnly: blocksByClass(allBlocks[SchemeCurrierOnly]), SchemeHandOnly: blocksByClass(allBlocks[SchemeHandOnly]),
		},
		sweepCache: map[string]workerSweep{},
	}
	inventory := classInventory(allBlocks, initMsg.MinClassTokens, initMsg.MinBlockTokens)
	w.jointEligible = eligibleSorted(inventory, SchemeJoint)
	w.jointBlocks = map[ClassID][]Block{}
	for _, cid := range w.jointEligible {
		w.jointBlocks[cid] = w.blocksByScheme[SchemeJoint][cid]
	}
	return w, nil
}

// RunWorker runs the Task32 subprocess worker protocol loop on in/out: an
// Init handshake with an explicit protocol version and scientific
// fingerprint (the same computeFingerprint that guards checkpoint resume),
// then a Job/Result loop until the coordinator closes its side or sends
// Shutdown. Every write to out is a single JSON line: nothing else may ever
// be written there, so the caller must route any incidental logging
// elsewhere (the process pool wires a worker's stderr straight to the
// coordinator's own stderr, never to this stream).
func RunWorker(ctx context.Context, in io.Reader, out io.Writer) error {
	scanner := bufio.NewScanner(in)
	scanner.Buffer(make([]byte, 4096), 1<<20)
	writer := bufio.NewWriter(out)

	fail := func(format string, args ...any) error {
		msg := fmt.Sprintf(format, args...)
		_ = writeMessage(writer, protocolMessage{Kind: "ready", OK: false, Error: msg})
		return fmt.Errorf("worker: %s", msg)
	}

	initMsg, ok, err := readMessage(scanner)
	if err != nil {
		return fmt.Errorf("worker: reading init: %w", err)
	}
	if !ok {
		return fmt.Errorf("worker: stdin closed before init handshake")
	}
	if initMsg.Kind != "init" {
		return fail("expected init message, got %q", initMsg.Kind)
	}
	if initMsg.Version != workerProtocolVersion {
		return fail("protocol version mismatch: worker speaks %d, coordinator sent %d", workerProtocolVersion, initMsg.Version)
	}

	w, err := newWorkerState(initMsg)
	if err != nil {
		return fail("%v", err)
	}

	if err := writeMessage(writer, protocolMessage{Kind: "ready", OK: true}); err != nil {
		return fmt.Errorf("worker: writing ready: %w", err)
	}

	for {
		if err := ctx.Err(); err != nil {
			return err
		}
		msg, ok, err := readMessage(scanner)
		if err != nil {
			return fmt.Errorf("worker: reading message: %w", err)
		}
		if !ok {
			return nil // coordinator closed stdin: clean shutdown
		}
		switch msg.Kind {
		case "shutdown":
			return nil
		case "job":
			if msg.JobID == nil {
				return fmt.Errorf("worker: job message missing job_id")
			}
			value, jobErr := w.compute(*msg.JobID)
			result := protocolMessage{Kind: "result", JobID: msg.JobID, Value: value}
			if jobErr != nil {
				result.Error = jobErr.Error()
			}
			if err := writeMessage(writer, result); err != nil {
				return fmt.Errorf("worker: writing result for %+v: %w", *msg.JobID, err)
			}
		default:
			return fmt.Errorf("worker: unexpected message kind %q", msg.Kind)
		}
	}
}
