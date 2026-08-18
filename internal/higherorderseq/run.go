package higherorderseq

import (
	"crypto/sha256"
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"zcore.dev/voinich/internal/workdir"
)

func defaults(c Config) Config {
	if c.CorpusPath == "" {
		c.CorpusPath = "data_work/ZL3b-x7.txt"
	}
	if c.TokenMetadataMap == "" {
		c.TokenMetadataMap = "workdir/metadata-validation/token_metadata_map.tsv"
	}
	if c.AuditDir == "" {
		c.AuditDir = workdir.Path("replicated-local-structure")
	}
	if c.DiscoveryDir == "" {
		c.DiscoveryDir = workdir.Dir
	}
	if c.OutputDir == "" {
		c.OutputDir = workdir.Path("higher-order-sequences")
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

// auditFingerprint hashes the exact frozen previous-audit files this program
// reads, so both the checkpoint fingerprint and the reproducibility record
// in the final YAML change the moment any upstream frozen input changes.
func auditFingerprint(auditDir string) (string, error) {
	h := sha256.New()
	for _, name := range []string{
		"universal_sequence_inventory.tsv",
		"sequence_null_validation.tsv",
		"strict_replicated_sequences.tsv",
		"sequence_replication_status.tsv",
		"replicated_local_structure.yaml",
	} {
		b, err := os.ReadFile(filepath.Join(auditDir, name))
		if err != nil {
			return "", err
		}
		h.Write(b)
	}
	return fmt.Sprintf("%x", h.Sum(nil)), nil
}

// candidateSeed derives a distinct, deterministic RNG seed per candidate
// from the run seed and the candidate's own sequence text, so permutation
// streams never collide across candidates yet stay fully reproducible.
func candidateSeed(base int64, sequence string) int64 {
	h := sha256.Sum256([]byte(sequence))
	var mix int64
	for i := 0; i < 8; i++ {
		mix = mix<<8 | int64(h[i])
	}
	if mix < 0 {
		mix = -mix
	}
	return base ^ mix
}

func permutationsFor(cand Candidate, primary int) int {
	if cand.Family == "primary" {
		return primary
	}
	return secondaryPermutations(primary)
}

// RunAndWrite runs the complete higher-order-sequence-validate pipeline
// (task22 Parts A-U) and writes its outputs.
func RunAndWrite(c Config) error {
	c = defaults(c)
	if c.ProgressWriter == nil && !c.Quiet {
		c.ProgressWriter = os.Stderr
	}
	p := newProgress(c.ProgressWriter)

	p.begin(1, "Loading corpus, metadata, frozen candidates and structural classes")
	tokens, blocks, lineLength, corpusSHA, metaSHA, err := loadCorpusAndBlocks(c.CorpusPath, c.TokenMetadataMap, c.Generic)
	if err != nil {
		return fmt.Errorf("load corpus: %w", err)
	}
	_ = tokens
	candidates, err := loadFrozenCandidates(c.AuditDir)
	if err != nil {
		return fmt.Errorf("load frozen candidates: %w", err)
	}
	relatives, err := structuralRelatives(c.DiscoveryDir)
	if err != nil {
		return fmt.Errorf("load structural classes: %w", err)
	}
	auditSHA, err := auditFingerprint(c.AuditDir)
	if err != nil {
		return fmt.Errorf("hash frozen audit inputs: %w", err)
	}
	p.update(1, 1, "Loading corpus, metadata, frozen candidates and structural classes")

	fingerprint := computeFingerprint(c, corpusSHA, metaSHA, auditSHA)
	cp, resumed := loadCheckpoint(c.CheckpointPath, fingerprint)
	if !resumed {
		cp = newCheckpoint(fingerprint)
	} else if c.ProgressWriter != nil {
		fmt.Fprintf(c.ProgressWriter, "Resuming from checkpoint %s: %d parts already done\n", c.CheckpointPath, len(cp.PartsDone))
	}
	checkpoint := func() error { return saveCheckpoint(c.CheckpointPath, cp) }

	seqs := make([]string, len(candidates))
	byID := map[string]Candidate{}
	for i, cand := range candidates {
		seqs[i] = cand.Sequence
		byID[cand.Sequence] = cand
	}
	total := len(candidates)

	runPart := func(stage int, label, partName string, fn func(cand Candidate, r *CandidateResult)) error {
		p.begin(stage, label)
		for i, cand := range candidates {
			key := partKey(cand.Sequence, partName)
			if !cp.PartsDone[key] {
				fn(cand, cp.resultFor(cand.Sequence))
				cp.PartsDone[key] = true
				if err := checkpoint(); err != nil {
					return fmt.Errorf("save checkpoint: %w", err)
				}
			}
			p.update(i+1, total, label)
		}
		return nil
	}

	if err := runPart(2, "Part A-C: occurrences and conditional probabilities", "occ_cond", func(cand Candidate, r *CandidateResult) {
		r.Candidate = cand
		r.Occurrences = findOccurrences(cand, blocks, lineLength)
		r.ConditionalRows = conditionalRowsForCandidate(cand, blocks)
	}); err != nil {
		return err
	}

	if err := runPart(3, "Part D: conditional-neighbor permutation (CMI)", "cmi", func(cand Candidate, r *CandidateResult) {
		r.CMI = runCMI(cand, blocks, permutationsFor(cand, c.Permutations), candidateSeed(c.Seed, cand.Sequence))
	}); err != nil {
		return err
	}

	if err := runPart(4, "Part E: leave-one-block-out model comparison", "lobo", func(cand Candidate, r *CandidateResult) {
		r.LOBO = runLOBO(cand, blocks)
	}); err != nil {
		return err
	}

	if err := runPart(5, "Part F-I: context, continuation, cross-block, meta-analysis", "context_meta", func(cand Candidate, r *CandidateResult) {
		r.ContextControls = contextControlRows(cand, blocks)
		r.ContextRank = contextRankRow(cand, blocks)
		r.Continuations = continuationDistributions(cand, blocks)
		r.ContinuationEnt = continuationEntropy(cand, blocks)
		r.CrossBlock = crossBlockRow(r.ConditionalRows)
		r.Meta = metaAnalysisRow(r.ConditionalRows)
	}); err != nil {
		return err
	}

	if err := runPart(6, "Part J: jackknife", "jackknife", func(cand Candidate, r *CandidateResult) {
		r.Jackknife = jackknifeRow(cand, blocks, primaryEligible(r.ConditionalRows))
	}); err != nil {
		return err
	}

	if err := runPart(7, "Part K-L: position and structural-family controls", "position_family", func(cand Candidate, r *CandidateResult) {
		r.Position = positionRows(cand, r.Occurrences, blocks, lineLength)
		r.BlockPosTVD = positionTVD(r.Position, "block_position_bin")
		r.LinePosTVD = positionTVD(r.Position, "line_position")
		r.StructuralFamily = structuralFamilyRows(cand, blocks, relatives)
	}); err != nil {
		return err
	}

	p.begin(8, "Part N, P, Q, R, S: multiple comparisons, classification, and writing outputs")
	results := make([]*CandidateResult, len(seqs))
	for i, seq := range seqs {
		results[i] = cp.Results[seq]
	}

	// Part N: BH FDR applied only within the primary confirmatory family.
	var primaryIdx []int
	var primaryP []float64
	for i, r := range results {
		if r.Candidate.Family == "primary" {
			primaryIdx = append(primaryIdx, i)
			primaryP = append(primaryP, r.CMI.EmpiricalP)
		}
	}
	q := bh(primaryP)
	dependence := make([]DependenceRow, len(results))
	for i, r := range results {
		dependence[i] = DependenceRow{Sequence: r.Candidate.Sequence, Family: r.Candidate.Family, Permutations: r.CMI.Permutations, EmpiricalP: r.CMI.EmpiricalP}
	}
	for k, i := range primaryIdx {
		dependence[i].FDRQ = q[k]
		dependence[i].Significant = q[k] <= 0.05
	}
	for i, r := range results {
		if r.Candidate.Family != "primary" {
			dependence[i].FDRQ = r.CMI.EmpiricalP
			dependence[i].Significant = r.CMI.EmpiricalP <= 0.05
		}
	}

	validation := make([]ValidationRow, len(results))
	for i, r := range results {
		validation[i] = classify(classificationInput{
			Candidate: r.Candidate, Dependence: dependence[i], CrossBlock: r.CrossBlock,
			LOBO: r.LOBO, Jackknife: r.Jackknife, BlockPosTVD: r.BlockPosTVD, LinePosTVD: r.LinePosTVD,
			Generic: c.Generic,
		})
	}

	meta := runMeta{
		CorpusSHA256: corpusSHA, MetadataSHA256: metaSHA, AuditSHA256: auditSHA,
		TokenCount: len(tokens), Permutations: c.Permutations, Seed: c.Seed,
	}
	sort.SliceStable(results, func(i, j int) bool { return results[i].Candidate.Sequence < results[j].Candidate.Sequence })
	if err := writeAll(c, meta, candidates, results, dependence, validation); err != nil {
		return err
	}
	removeCheckpoint(c.CheckpointPath)
	p.update(1, 1, "Part N, P, Q, R, S: multiple comparisons, classification, and writing outputs")
	fmt.Printf("higher-order-sequence-validate completed for %d frozen candidates; results written to %s\n", len(candidates), c.OutputDir)
	return nil
}

type runMeta struct {
	CorpusSHA256   string
	MetadataSHA256 string
	AuditSHA256    string
	TokenCount     int
	Permutations   int
	Seed           int64
}
