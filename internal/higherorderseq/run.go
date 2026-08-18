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

	// Task44: the unit of independent distributed work is a whole candidate
	// (see executor.go's CandidateExecutor doc comment for why CMI's
	// permutations can never be split across jobs), so the six previously
	// separately-checkpointed parts now run together per candidate. A
	// candidate already fully resolved by every one of its six part keys
	// (e.g. from a checkpoint saved by a pre-task44 build, or a resumed
	// distributed run) is skipped entirely; a partially-done candidate is
	// simply recomputed whole, which is idempotent and produces the exact
	// same values, just without reusing whatever partial progress existed.
	partNames := []string{"occ_cond", "cmi", "lobo", "context_meta", "jackknife", "position_family"}
	candidateDone := func(sequence string) bool {
		for _, name := range partNames {
			if !cp.PartsDone[partKey(sequence, name)] {
				return false
			}
		}
		return true
	}
	var pendingSeqs []string
	for _, cand := range candidates {
		if !candidateDone(cand.Sequence) {
			pendingSeqs = append(pendingSeqs, cand.Sequence)
		}
	}

	p.begin(2, "Part A-L: per-candidate occurrences, conditional probabilities, CMI permutation null, LOBO, context/continuation/cross-block/meta-analysis, jackknife, position and structural-family controls")
	if len(pendingSeqs) > 0 {
		executor := candidateExecutorFor(c, candidates, blocks, lineLength, relatives)
		done := total - len(pendingSeqs)
		err := runCandidateBattery(executorContext(c), executor, pendingSeqs, executorWorkers(c), func(_ int, sequence string, result CandidateResult) error {
			*cp.resultFor(sequence) = result
			for _, name := range partNames {
				cp.PartsDone[partKey(sequence, name)] = true
			}
			if err := checkpoint(); err != nil {
				return fmt.Errorf("save checkpoint: %w", err)
			}
			done++
			p.update(done, total, "Part A-L")
			return nil
		})
		if err != nil {
			return err
		}
	}
	p.update(total, total, "Part A-L")

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
