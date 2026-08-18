package positionalcontinuation

import (
	"fmt"
	"os"
	"path/filepath"

	"zcore.dev/voinich/internal/workdir"
)

func defaults(c Config) Config {
	if c.CorpusPath == "" {
		c.CorpusPath = "data_work/ZL3b-x7.txt"
	}
	if c.TokenMetadataMap == "" {
		c.TokenMetadataMap = "workdir/metadata-validation/token_metadata_map.tsv"
	}
	if c.HigherOrderDir == "" {
		c.HigherOrderDir = workdir.Path("higher-order-sequences")
	}
	if c.OutputDir == "" {
		c.OutputDir = workdir.Path("positional-continuation")
	}
	if c.Permutations <= 0 {
		c.Permutations = 10000
	}
	if c.Seed == 0 {
		c.Seed = 1
	}
	if c.CheckpointPath == "" {
		c.CheckpointPath = filepath.Join(c.OutputDir, "checkpoint.json")
	}
	if c.CheckpointPath == "-" {
		c.CheckpointPath = ""
	}
	return c
}

type runMeta struct {
	CorpusSHA256      string
	MetadataSHA256    string
	HigherOrderSHA256 string
	Permutations      int
	Seed              int64
}

// seedFor derives a distinct, deterministic RNG seed per named sub-step from
// the run seed, so permutation streams never collide across steps yet stay
// fully reproducible for a fixed -seed.
func seedFor(base int64, name string) int64 {
	h := 0
	for _, c := range name {
		h = h*131 + int(c)
	}
	return base ^ int64(h)
}

// RunAndWrite runs the complete positional-continuation-validate pipeline
// (task23 Parts A-Q) and writes its outputs.
func RunAndWrite(c Config) error {
	c = defaults(c)
	if c.ProgressWriter == nil && !c.Quiet {
		c.ProgressWriter = os.Stderr
	}
	p := newProgress(c.ProgressWriter)

	if c.Generic {
		s, aiin, chey, e := resolveGenericTarget(c.HigherOrderDir)
		if e != nil {
			return fmt.Errorf("resolve generic target triple: %w", e)
		}
		FrozenS, FrozenAiin, FrozenChey = s, aiin, chey
		FrozenSAiin = FrozenS + " " + FrozenAiin
	}
	p.begin(1, "Loading corpus, metadata and frozen higher-order-sequences inputs")
	tokens, blocks, lineLength, corpusSHA, metaSHA, err := loadCorpusAndBlocks(c.CorpusPath, c.TokenMetadataMap, c.Generic)
	if err != nil {
		return fmt.Errorf("load corpus: %w", err)
	}
	if !priorResultExists(c.HigherOrderDir) {
		return fmt.Errorf("higher-order-dir %s missing higher_order_sequence_analysis.yaml", c.HigherOrderDir)
	}
	higherOrderSHA, err := higherOrderFingerprint(c.HigherOrderDir)
	if err != nil {
		return fmt.Errorf("hash frozen higher-order-sequences inputs: %w", err)
	}
	p.update(1, 1, "Loading corpus, metadata and frozen higher-order-sequences inputs")

	fingerprint := computeFingerprint(c, corpusSHA, metaSHA, higherOrderSHA)
	cp, resumed := loadCheckpoint(c.CheckpointPath, fingerprint)
	if !resumed {
		cp = newCheckpoint(fingerprint)
	} else if c.ProgressWriter != nil {
		fmt.Fprintf(c.ProgressWriter, "Resuming from checkpoint %s: %d parts already done\n", c.CheckpointPath, len(cp.PartsDone))
	}
	checkpoint := func() error { return saveCheckpoint(c.CheckpointPath, cp) }
	state := cp.State

	runPart := func(stage int, label, partName string, fn func()) error {
		p.begin(stage, label)
		if !cp.PartsDone[partName] {
			fn()
			cp.PartsDone[partName] = true
			if err := checkpoint(); err != nil {
				return fmt.Errorf("save checkpoint: %w", err)
			}
		}
		p.update(1, 1, label)
		return nil
	}

	if err := runPart(2, "Part A: s-aiin and aiin occurrence extraction", "occurrences", func() {
		state.SAiinOccurrences = findSAiinOccurrences(blocks, lineLength, len(tokens))
		state.AiinOccurrences = findAiinOccurrences(blocks, lineLength, len(tokens))
	}); err != nil {
		return err
	}

	if err := runPart(3, "Part D: continuation distributions", "continuation", func() {
		state.ContinuationRows, state.DistSummaryRows = buildSAiinContinuationDistributions(state.SAiinOccurrences)
	}); err != nil {
		return err
	}

	var lineResult, blockResult positionalTestResult
	if err := runPart(4, "Part E-G: line-position permutation test (I(X;position), entropy, chey effect)", "postest_line", func() {
		lineResult = runPositionalTests(state.SAiinOccurrences, "line_position", lineCategories, c.Permutations, seedFor(c.Seed, "line_position"))
		state.PositionDependence = append(nonVariable(state.PositionDependence, "line_position"), lineResult.Dependence)
		state.PositionalEntropy = append(nonVariableEntropy(state.PositionalEntropy, "line_position"), lineResult.Entropy...)
		state.CheyEffect = append(nonVariableChey(state.CheyEffect, "line_position"), lineResult.CheyEffect...)
	}); err != nil {
		return err
	}
	if err := runPart(5, "Part E-G: block-position permutation test (I(X;position), entropy, chey effect)", "postest_block", func() {
		blockResult = runPositionalTests(state.SAiinOccurrences, "block_position_coarse", blockCoarseCategories, c.Permutations, seedFor(c.Seed, "block_position"))
		state.PositionDependence = append(nonVariable(state.PositionDependence, "block_position_coarse"), blockResult.Dependence)
		state.PositionalEntropy = append(nonVariableEntropy(state.PositionalEntropy, "block_position_coarse"), blockResult.Entropy...)
		state.CheyEffect = append(nonVariableChey(state.CheyEffect, "block_position_coarse"), blockResult.CheyEffect...)
	}); err != nil {
		return err
	}

	if err := runPart(6, "Part H: aiin-only positional control", "aiin_control", func() {
		byVar := map[string][]CheyEffectRow{}
		for _, ce := range state.CheyEffect {
			byVar[ce.PositionVariable] = append(byVar[ce.PositionVariable], ce)
		}
		state.AiinControl = buildAiinControlRows(state.AiinOccurrences, byVar)
	}); err != nil {
		return err
	}

	if err := runPart(7, "Part I: stratified predecessor test (line position)", "stratified_line", func() {
		row := runStratifiedPredecessorTest(state.AiinOccurrences, "line_position", c.Permutations, seedFor(c.Seed, "stratified_line"))
		state.StratifiedPredecessor = append(nonVariableStrat(state.StratifiedPredecessor, "line_position"), row)
	}); err != nil {
		return err
	}
	if err := runPart(8, "Part I: stratified predecessor test (block position)", "stratified_block", func() {
		row := runStratifiedPredecessorTest(state.AiinOccurrences, "block_position_coarse", c.Permutations, seedFor(c.Seed, "stratified_block"))
		state.StratifiedPredecessor = append(nonVariableStrat(state.StratifiedPredecessor, "block_position_coarse"), row)
	}); err != nil {
		return err
	}

	if err := runPart(9, "Part J: M1/M2/M3 leave-one-block-out model comparison", "model_lobo", func() {
		state.ModelLOBO = runModelLOBO(state.AiinOccurrences)
	}); err != nil {
		return err
	}

	if err := runPart(10, "Part K: cross-block replication", "cross_block", func() {
		state.CrossBlock = buildCrossBlockPositional(blocks, state.AiinOccurrences)
	}); err != nil {
		return err
	}

	if err := runPart(11, "Part L: leave-one-block-out jackknife", "jackknife", func() {
		state.Jackknife = runJackknife(state.SAiinOccurrences, state.AiinOccurrences, seedFor(c.Seed, "jackknife"))
	}); err != nil {
		return err
	}

	if err := runPart(12, "Part M: line vs block position", "line_vs_block", func() {
		state.LineVsBlock, state.LineVsBlockCorrelation, state.LineVsBlockSource = buildLineVsBlockRows(state.SAiinOccurrences)
	}); err != nil {
		return err
	}

	if err := runPart(13, "Part N: boundary distance", "boundary", func() {
		state.BoundaryDistance = buildBoundaryDistanceRows(state.SAiinOccurrences, c.Permutations, seedFor(c.Seed, "boundary"))
	}); err != nil {
		return err
	}

	if err := runPart(14, "Part O: surrounding context", "surrounding", func() {
		state.SurroundingContext = buildSurroundingContextRows(state.SAiinOccurrences)
	}); err != nil {
		return err
	}

	if err := runPart(15, "Part P-Q: reverse position, classification, and writing outputs", "classify_write", func() {
		state.ReversePosition = buildReversePositionRows(state.SAiinOccurrences, state.AiinOccurrences)

		depByVar := map[string]PositionDependenceRow{}
		for _, d := range state.PositionDependence {
			depByVar[d.PositionVariable] = d
		}
		stratByVar := map[string]StratifiedPredecessorRow{}
		for _, s := range state.StratifiedPredecessor {
			stratByVar[s.PositionVariable] = s
		}
		eligible, _, _, _, consistency := crossBlockSignConsistency(state.CrossBlock)
		m3Frac := 0.0
		if state.ModelLOBO.TestedBlocks > 0 {
			m3Frac = float64(state.ModelLOBO.BlocksM3BetterM2) / float64(state.ModelLOBO.TestedBlocks)
		}
		sensitive := false
		for _, j := range state.Jackknife {
			if j.SingleBlockSensitive {
				sensitive = true
			}
		}
		cheyN := 0
		uniqueCheyCtx := 0
		for _, sc := range state.SurroundingContext {
			if sc.Group == "chey" {
				cheyN = sc.OccurrenceCount
				uniqueCheyCtx = sc.UniqueSurroundingContexts
			}
		}
		state.Validation = classify(classificationInput{
			PositionDependenceP:           depByVar["line_position"].EmpiricalP,
			StratifiedPredecessorP:        stratByVar["line_position"].EmpiricalP,
			M3BetterThanM2Fraction:        m3Frac,
			CrossBlockSignConsistency:     consistency,
			EligibleBlocks:                eligible,
			SingleBlockSensitive:          sensitive,
			UniqueCheySurroundingContexts: uniqueCheyCtx,
			CheyOccurrences:               cheyN,
		})
	}); err != nil {
		return err
	}

	meta := runMeta{CorpusSHA256: corpusSHA, MetadataSHA256: metaSHA, HigherOrderSHA256: higherOrderSHA, Permutations: c.Permutations, Seed: c.Seed}
	if err := writeAll(c, meta, state); err != nil {
		return err
	}
	removeCheckpoint(c.CheckpointPath)
	fmt.Printf("positional-continuation-validate completed; results written to %s\n", c.OutputDir)
	return nil
}

func nonVariable(rows []PositionDependenceRow, variable string) []PositionDependenceRow {
	var out []PositionDependenceRow
	for _, r := range rows {
		if r.PositionVariable != variable {
			out = append(out, r)
		}
	}
	return out
}
func nonVariableEntropy(rows []PositionalEntropyRow, variable string) []PositionalEntropyRow {
	var out []PositionalEntropyRow
	for _, r := range rows {
		if r.PositionVariable != variable {
			out = append(out, r)
		}
	}
	return out
}
func nonVariableChey(rows []CheyEffectRow, variable string) []CheyEffectRow {
	var out []CheyEffectRow
	for _, r := range rows {
		if r.PositionVariable != variable {
			out = append(out, r)
		}
	}
	return out
}
func nonVariableStrat(rows []StratifiedPredecessorRow, variable string) []StratifiedPredecessorRow {
	var out []StratifiedPredecessorRow
	for _, r := range rows {
		if r.PositionVariable != variable {
			out = append(out, r)
		}
	}
	return out
}
