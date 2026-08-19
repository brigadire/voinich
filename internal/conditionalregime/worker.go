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

	"zcore.dev/voinich/internal/beginendanalyze"
	"zcore.dev/voinich/internal/higherorderseq"
	"zcore.dev/voinich/internal/normalization"
	"zcore.dev/voinich/internal/normalizationcompare"
	"zcore.dev/voinich/internal/positionalcontinuation"
	"zcore.dev/voinich/internal/replicatedlocalaudit"
	"zcore.dev/voinich/internal/sequenceanalyze"
	"zcore.dev/voinich/internal/structuralprojection"
	"zcore.dev/voinich/internal/tokenrelationvalidation"
	"zcore.dev/voinich/internal/transitionnetwork"
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

type protocolComputer interface {
	compute(JobID) (float64, json.RawMessage, error)
}
type conditionalComputer struct{ state *workerState }

func (w conditionalComputer) compute(id JobID) (float64, json.RawMessage, error) {
	v, e := w.state.compute(id)
	return v, nil, e
}

type structuralComputer struct {
	state *structuralprojection.TrialWorker
}

func (w structuralComputer) compute(id JobID) (float64, json.RawMessage, error) {
	if id.Stage != "structural_projection_trial" {
		return 0, nil, fmt.Errorf("unknown structural job stage %q", id.Stage)
	}
	r, err := w.state.Run(context.Background(), id.ReplicateIndex)
	if err != nil {
		return 0, nil, err
	}
	b, err := json.Marshal(r)
	return 0, b, err
}

// normalizationComputer serves the Task42 normalization_compare_baseline
// job type: one random-baseline trial per (threshold label, run index),
// identical to the computation normalization-compare's own default
// in-process executor runs, just dispatched by JobID instead of a for loop.
type normalizationComputer struct {
	classes normalization.ClassesOutput
	corpus  normalization.Corpus
	seed    int64
	params  sequenceanalyze.Parameters
}

func (w normalizationComputer) compute(id JobID) (float64, json.RawMessage, error) {
	if id.Stage != "normalization_compare_baseline" {
		return 0, nil, fmt.Errorf("unknown normalization job stage %q", id.Stage)
	}
	var model normalization.Model
	found := false
	for _, m := range w.classes.Models {
		if m.Label == id.Combination {
			model, found = m, true
			break
		}
	}
	if !found {
		return 0, nil, fmt.Errorf("unknown threshold %q", id.Combination)
	}
	result, err := normalizationcompare.RunRandomTrial(model, w.corpus, w.classes.Meta.MinTokenCount, w.classes.Meta.SingletonMode, w.seed, id.ReplicateIndex, w.params)
	if err != nil {
		return 0, nil, err
	}
	b, err := json.Marshal(result)
	return 0, b, err
}

func newNormalizationComputer(initMsg protocolMessage) (normalizationComputer, error) {
	corpus, err := normalization.LoadCorpus(initMsg.CorpusPath)
	if err != nil {
		return normalizationComputer{}, fmt.Errorf("read corpus: %w", err)
	}
	classes, err := normalizationcompare.LoadClasses(initMsg.ClassesPath)
	if err != nil {
		return normalizationComputer{}, fmt.Errorf("read classes: %w", err)
	}
	fp, err := normalizationcompare.Fingerprint(initMsg.CorpusPath, initMsg.ClassesPath, initMsg.MinTokenCount, initMsg.SingletonMode, initMsg.Seed, initMsg.RandomRuns)
	if err != nil || fp != initMsg.Fingerprint {
		return normalizationComputer{}, fmt.Errorf("input/config fingerprint mismatch: worker loaded different input or parameters")
	}
	if classes.Meta.MinTokenCount != initMsg.MinTokenCount || classes.Meta.SingletonMode != initMsg.SingletonMode {
		return normalizationComputer{}, fmt.Errorf("classes.yaml meta does not match coordinator's declared parameters")
	}
	return normalizationComputer{classes: classes, corpus: corpus, seed: initMsg.Seed, params: sequenceanalyze.DefaultParameters()}, nil
}

// tokenRelationComputer serves the Task44 token_relation_permutation job
// type: one permutation replicate per (family, run index), identical to
// the computation token-relation-validate's own default in-process
// executor runs, just dispatched by JobID instead of runBattery's pool.
type tokenRelationComputer struct {
	blocks     []tokenrelationvalidation.Block
	candidates []tokenrelationvalidation.Candidate
	maxD       int
	seed       int64
}

func (w tokenRelationComputer) compute(id JobID) (float64, json.RawMessage, error) {
	if id.Stage != "token_relation_permutation" {
		return 0, nil, fmt.Errorf("unknown token-relation job stage %q", id.Stage)
	}
	scores, err := tokenrelationvalidation.ComputeReplicate(w.blocks, w.candidates, w.maxD, w.seed, id.Combination, id.ReplicateIndex)
	if err != nil {
		return 0, nil, err
	}
	b, err := json.Marshal(scores)
	return 0, b, err
}

// newTokenRelationComputer reconstructs the exact Blocks/Candidates/MaxD a
// local analyze() run would have from initMsg's (already locally-resolved,
// for remote mode) CorpusPath/TokenMetadataMap/DiscoveryDir, verifying the
// coordinator's declared Fingerprint before trusting any of it.
func newTokenRelationComputer(initMsg protocolMessage) (tokenRelationComputer, error) {
	cfg := tokenrelationvalidation.Config{
		CorpusPath: initMsg.CorpusPath, MetadataPath: initMsg.TokenMetadataMap, DiscoveryDir: initMsg.DiscoveryDir,
		Generic: initMsg.Generic, Permutations: initMsg.Permutations, RefinePermutations: initMsg.RefinePermutations, Seed: initMsg.Seed,
	}
	fp, err := tokenrelationvalidation.Fingerprint(cfg)
	if err != nil || fp != initMsg.Fingerprint {
		return tokenRelationComputer{}, fmt.Errorf("input/config fingerprint mismatch: worker loaded different input or parameters")
	}
	blocks, candidates, maxD, err := tokenrelationvalidation.LoadForDistribution(cfg)
	if err != nil {
		return tokenRelationComputer{}, fmt.Errorf("reconstruct blocks/candidates: %w", err)
	}
	return tokenRelationComputer{blocks: blocks, candidates: candidates, maxD: maxD, seed: initMsg.Seed}, nil
}

// transitionNetworkComputer serves the Task44
// transition_network_permutation job type: one permutation replicate per
// (phase, rep index), identical to the computation
// transition-network-validate's own default in-process executor runs.
// PermWorkspace.run is not safe for concurrent use (it reuses scratch
// buffers - see permworkspace.go), so compute serializes access to it.
type transitionNetworkComputer struct {
	ws   *transitionnetwork.PermWorkspace
	seed int64
	mu   *sync.Mutex
}

func (w transitionNetworkComputer) compute(id JobID) (float64, json.RawMessage, error) {
	if id.Stage != "transition_network_permutation" {
		return 0, nil, fmt.Errorf("unknown transition-network job stage %q", id.Stage)
	}
	w.mu.Lock()
	defer w.mu.Unlock()
	r, err := transitionnetwork.ComputeReplicate(w.ws, w.seed, id.Combination, id.ReplicateIndex)
	if err != nil {
		return 0, nil, err
	}
	b, err := json.Marshal(r)
	return 0, b, err
}

// newTransitionNetworkComputer reconstructs the exact PermWorkspace a
// local RunAndWrite run would have from initMsg's (already
// locally-resolved, for remote mode) CorpusPath/TokenMetadataMap,
// verifying the coordinator's declared Fingerprint before trusting any of
// it.
func newTransitionNetworkComputer(initMsg protocolMessage) (transitionNetworkComputer, error) {
	cfg := transitionnetwork.Config{
		CorpusPath: initMsg.CorpusPath, MetadataPath: initMsg.TokenMetadataMap, Generic: initMsg.Generic,
		MinTokenCount: initMsg.MinTokenCount, MinBlockTokenCount: initMsg.MinBlockTokens,
		Permutations: initMsg.Permutations, RefinePermutations: initMsg.RefinePermutations, Seed: initMsg.Seed,
	}
	fp, err := transitionnetwork.Fingerprint(cfg)
	if err != nil || fp != initMsg.Fingerprint {
		return transitionNetworkComputer{}, fmt.Errorf("input/config fingerprint mismatch: worker loaded different input or parameters")
	}
	ws, _, _, err := transitionnetwork.LoadForDistribution(cfg)
	if err != nil {
		return transitionNetworkComputer{}, fmt.Errorf("reconstruct permutation workspace: %w", err)
	}
	return transitionNetworkComputer{ws: ws, seed: initMsg.Seed, mu: &sync.Mutex{}}, nil
}

// beginEndComputer serves the Task47 begin_end_candidate_batch job type:
// one candidate-pair batch per JobID.ReplicateIndex, identical to the
// computation begin-end-analyze's own default in-process executor runs.
// Unlike transitionNetworkComputer's PermWorkspace, beginendanalyze.
// Workspace is read-only once built (LoadForDistribution never mutates it
// afterward), so compute needs no mutex.
type beginEndComputer struct {
	ws        *beginendanalyze.Workspace
	batchSize int
}

func (w beginEndComputer) compute(id JobID) (float64, json.RawMessage, error) {
	if id.Stage != "begin_end_candidate_batch" {
		return 0, nil, fmt.Errorf("unknown begin-end-analyze job stage %q", id.Stage)
	}
	r := beginendanalyze.ComputeBatch(w.ws, id.ReplicateIndex, w.batchSize)
	b, err := json.Marshal(encodeBeginEndBatchResult(r))
	return 0, b, err
}

// newBeginEndComputer reconstructs the exact Workspace a local RunAndWrite
// run would have from initMsg's (already locally-resolved, for remote
// mode) CorpusPath/DictionaryPath, verifying the coordinator's declared
// Fingerprint before trusting any of it.
func newBeginEndComputer(initMsg protocolMessage) (beginEndComputer, error) {
	cfg := beginendanalyze.Config{
		DictionaryPath: initMsg.DictionaryPath, CorpusPath: initMsg.CorpusPath,
		MaxWindow: initMsg.MaxWindow, Permutations: initMsg.Permutations, MinTokenCount: initMsg.MinTokenCount,
		RandomSeed: initMsg.Seed, PermutationMode: initMsg.PermutationMode, IncludeUnclear: initMsg.IncludeUnclear,
		MaxCandidates: initMsg.MaxCandidates, CandidateBatchSize: initMsg.CandidateBatchSize,
	}
	fp, err := beginendanalyze.Fingerprint(cfg)
	if err != nil || fp != initMsg.Fingerprint {
		return beginEndComputer{}, fmt.Errorf("input/config fingerprint mismatch: worker loaded different input or parameters")
	}
	ws, err := beginendanalyze.LoadForDistribution(cfg)
	if err != nil {
		return beginEndComputer{}, fmt.Errorf("reconstruct candidate workspace: %w", err)
	}
	return beginEndComputer{ws: ws, batchSize: initMsg.CandidateBatchSize}, nil
}

// replicatedLocalAuditComputer serves the Task44 replicated_local_null job
// type: one permutation replicate per (phase, run index) across the
// distance/shuffle/markov null batteries, identical to the computation
// replicated-local-structure-audit's own default in-process executor runs.
type replicatedLocalAuditComputer struct {
	state *replicatedlocalaudit.DistributionState
	seed  int64
}

func (w replicatedLocalAuditComputer) compute(id JobID) (float64, json.RawMessage, error) {
	if id.Stage != "replicated_local_null" {
		return 0, nil, fmt.Errorf("unknown replicated-local-audit job stage %q", id.Stage)
	}
	r, err := replicatedlocalaudit.ComputeReplicate(w.state, w.seed, id.Combination, id.ReplicateIndex)
	if err != nil {
		return 0, nil, err
	}
	b, err := json.Marshal(r)
	return 0, b, err
}

// newReplicatedLocalAuditComputer reconstructs the exact distributionState
// a local RunAndWrite run would have from initMsg's (already
// locally-resolved, for remote mode) CorpusPath/TokenMetadataMap/
// RelationDir/DiscoveryDir, verifying the coordinator's declared
// Fingerprint before trusting any of it.
func newReplicatedLocalAuditComputer(initMsg protocolMessage) (replicatedLocalAuditComputer, error) {
	cfg := replicatedlocalaudit.Config{
		CorpusPath: initMsg.CorpusPath, MetadataPath: initMsg.TokenMetadataMap,
		RelationDir: initMsg.RelationDir, DiscoveryDir: initMsg.DiscoveryDir,
		Generic: initMsg.Generic, Permutations: initMsg.Permutations, Seed: initMsg.Seed,
	}
	fp, err := replicatedlocalaudit.Fingerprint(cfg)
	if err != nil || fp != initMsg.Fingerprint {
		return replicatedLocalAuditComputer{}, fmt.Errorf("input/config fingerprint mismatch: worker loaded different input or parameters")
	}
	state, _, err := replicatedlocalaudit.LoadForDistribution(cfg)
	if err != nil {
		return replicatedLocalAuditComputer{}, fmt.Errorf("reconstruct distribution state: %w", err)
	}
	return replicatedLocalAuditComputer{state: state, seed: initMsg.Seed}, nil
}

// higherOrderComputer serves the Task44 higher_order_candidate job type:
// one job per whole frozen candidate (never one CMI permutation - see
// higherorderseq.CandidateExecutor's doc comment for why that RNG stream
// can never be split), identical to the computation
// higher-order-sequence-validate's own default in-process executor runs.
type higherOrderComputer struct {
	candidates   []higherorderseq.Candidate
	blocks       []higherorderseq.Block
	lineLength   map[string]int
	relatives    map[string][]string
	permutations int
	seed         int64
}

func (w higherOrderComputer) compute(id JobID) (float64, json.RawMessage, error) {
	if id.Stage != "higher_order_candidate" {
		return 0, nil, fmt.Errorf("unknown higher-order-sequence job stage %q", id.Stage)
	}
	r, err := higherorderseq.ComputeCandidate(w.candidates, w.blocks, w.lineLength, w.relatives, w.permutations, w.seed, id.Combination)
	if err != nil {
		return 0, nil, err
	}
	b, err := json.Marshal(r)
	return 0, b, err
}

// newHigherOrderComputer reconstructs the exact candidates/blocks/
// lineLength/relatives a local RunAndWrite run would have from initMsg's
// (already locally-resolved, for remote mode) CorpusPath/TokenMetadataMap/
// AuditDir/DiscoveryDir, verifying the coordinator's declared Fingerprint
// before trusting any of it.
func newHigherOrderComputer(initMsg protocolMessage) (higherOrderComputer, error) {
	cfg := higherorderseq.Config{
		CorpusPath: initMsg.CorpusPath, TokenMetadataMap: initMsg.TokenMetadataMap,
		AuditDir: initMsg.AuditDir, DiscoveryDir: initMsg.DiscoveryDir,
		Generic: initMsg.Generic, Permutations: initMsg.Permutations, Seed: initMsg.Seed,
	}
	fp, err := higherorderseq.Fingerprint(cfg)
	if err != nil || fp != initMsg.Fingerprint {
		return higherOrderComputer{}, fmt.Errorf("input/config fingerprint mismatch: worker loaded different input or parameters")
	}
	candidates, blocks, lineLength, relatives, err := higherorderseq.LoadForDistribution(cfg)
	if err != nil {
		return higherOrderComputer{}, fmt.Errorf("reconstruct candidates/blocks: %w", err)
	}
	return higherOrderComputer{candidates: candidates, blocks: blocks, lineLength: lineLength, relatives: relatives, permutations: initMsg.Permutations, seed: initMsg.Seed}, nil
}

// positionalContinuationComputer serves the Task44
// positional_continuation_battery job type: one job per whole named
// battery (never one permutation within it - see
// positionalcontinuation.BatteryExecutor's doc comment for why each
// battery's RNG stream can never be split), identical to the computation
// positional-continuation-validate's own default in-process executor runs.
type positionalContinuationComputer struct {
	sAiinOccs    []positionalcontinuation.SAiinOccurrence
	aiinOccs     []positionalcontinuation.AiinOccurrence
	permutations int
	seed         int64
}

func (w positionalContinuationComputer) compute(id JobID) (float64, json.RawMessage, error) {
	if id.Stage != "positional_continuation_battery" {
		return 0, nil, fmt.Errorf("unknown positional-continuation job stage %q", id.Stage)
	}
	r, err := positionalcontinuation.ComputeBattery(w.sAiinOccs, w.aiinOccs, id.Combination, w.permutations, w.seed)
	if err != nil {
		return 0, nil, err
	}
	b, err := json.Marshal(r)
	return 0, b, err
}

// newPositionalContinuationComputer reconstructs the exact SAiinOccurrences/
// AiinOccurrences a local RunAndWrite run would have from initMsg's
// (already locally-resolved, for remote mode) CorpusPath/TokenMetadataMap/
// HigherOrderDir, verifying the coordinator's declared Fingerprint before
// trusting any of it.
func newPositionalContinuationComputer(initMsg protocolMessage) (positionalContinuationComputer, error) {
	cfg := positionalcontinuation.Config{
		CorpusPath: initMsg.CorpusPath, TokenMetadataMap: initMsg.TokenMetadataMap,
		HigherOrderDir: initMsg.HigherOrderDir, Generic: initMsg.Generic,
		Permutations: initMsg.Permutations, Seed: initMsg.Seed,
	}
	fp, err := positionalcontinuation.Fingerprint(cfg)
	if err != nil || fp != initMsg.Fingerprint {
		return positionalContinuationComputer{}, fmt.Errorf("input/config fingerprint mismatch: worker loaded different input or parameters")
	}
	sAiinOccs, aiinOccs, err := positionalcontinuation.LoadForDistribution(cfg)
	if err != nil {
		return positionalContinuationComputer{}, fmt.Errorf("reconstruct occurrences: %w", err)
	}
	return positionalContinuationComputer{sAiinOccs: sAiinOccs, aiinOccs: aiinOccs, permutations: initMsg.Permutations, seed: initMsg.Seed}, nil
}

func structuralConfigFromMessage(m protocolMessage) structuralprojection.Config {
	return structuralprojection.Config{CorpusPath: m.CorpusPath, StructuralPairsPath: m.StructuralPairsPath, DistancePairsPath: m.DistancePairsPath, FamiliesPath: m.FamiliesPath,
		MinStructuralSimilarity: m.MinStructuralSimilarity, MinReliability: m.MinReliability, ProjectionK: m.ProjectionK, RandomProjections: m.RandomProjections,
		MaxDistance: m.MaxDistance, MinObservations: m.MinObservations, TopN: m.TopN, FamilyID: m.FamilyID, ProjectionMode: m.ProjectionMode, Pair: m.Pair, Seed: m.Seed}
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

	var computer protocolComputer
	switch initMsg.Workload {
	case "structural_projection":
		cfg := structuralConfigFromMessage(initMsg)
		fp, e := structuralprojection.Fingerprint(cfg)
		if e != nil || fp != initMsg.Fingerprint {
			return fail("input/config fingerprint mismatch: worker loaded different input or parameters")
		}
		w, e := structuralprojection.NewTrialWorker(cfg)
		if e != nil {
			return fail("%v", e)
		}
		computer = structuralComputer{w}
	case "normalization_compare":
		w, e := newNormalizationComputer(initMsg)
		if e != nil {
			return fail("%v", e)
		}
		computer = w
	case "token_relation_permutation":
		w, e := newTokenRelationComputer(initMsg)
		if e != nil {
			return fail("%v", e)
		}
		computer = w
	case "transition_network_permutation":
		w, e := newTransitionNetworkComputer(initMsg)
		if e != nil {
			return fail("%v", e)
		}
		computer = w
	case "begin_end_candidate_batch":
		w, e := newBeginEndComputer(initMsg)
		if e != nil {
			return fail("%v", e)
		}
		computer = w
	case "replicated_local_null":
		w, e := newReplicatedLocalAuditComputer(initMsg)
		if e != nil {
			return fail("%v", e)
		}
		computer = w
	case "higher_order_candidate":
		w, e := newHigherOrderComputer(initMsg)
		if e != nil {
			return fail("%v", e)
		}
		computer = w
	case "positional_continuation_battery":
		w, e := newPositionalContinuationComputer(initMsg)
		if e != nil {
			return fail("%v", e)
		}
		computer = w
	default:
		w, e := newWorkerState(initMsg)
		if e != nil {
			return fail("%v", e)
		}
		computer = conditionalComputer{w}
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
			value, blob, jobErr := computer.compute(*msg.JobID)
			result := protocolMessage{Kind: "result", JobID: msg.JobID, Value: value, Blob: blob}
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
