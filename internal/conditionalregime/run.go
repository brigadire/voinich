package conditionalregime

import (
	"bufio"
	"context"
	"crypto/sha256"
	"fmt"
	"os"
	"sort"
	"strings"

	"zcore.dev/voinich/internal/workdir"
)

func defaults(c Config) Config {
	if c.CorpusPath == "" {
		c.CorpusPath = "data_work/ZL3b-x7.txt"
	}
	if c.TokenMetadataMap == "" {
		c.TokenMetadataMap = "workdir/metadata-validation/token_metadata_map.tsv"
	}
	if c.OutputDir == "" {
		c.OutputDir = workdir.Path("conditional-regimes")
	}
	if len(c.WindowSizes) == 0 {
		c.WindowSizes = []int{50, 100, 200, 500}
	}
	if len(c.ResidualWindowSizes) == 0 {
		c.ResidualWindowSizes = []int{50, 100, 200, 500, 1000}
	}
	if c.MinClassTokens <= 0 {
		c.MinClassTokens = 1000
	}
	if c.MinBlockTokens <= 0 {
		c.MinBlockTokens = 500
	}
	if c.KMin <= 0 {
		c.KMin = 2
	}
	if c.KMaxWithin <= 0 {
		c.KMaxWithin = 10
	}
	if c.KMaxResidual <= 0 {
		c.KMaxResidual = 15
	}
	if c.Permutations <= 0 {
		c.Permutations = 1000
	}
	if c.Seed == 0 {
		c.Seed = 1
	}
	if c.Workers <= 0 {
		c.Workers = 1
	}
	if c.Executor == "" {
		c.Executor = "goroutine"
	}
	if c.Context == nil {
		c.Context = context.Background()
	}
	if c.CheckpointPath == "" {
		c.CheckpointPath = c.OutputDir + "/checkpoint.json"
	}
	if c.CheckpointPath == "-" {
		c.CheckpointPath = ""
	}
	return c
}

func readCorpus(path string) ([]string, string, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, "", err
	}
	hash := fmt.Sprintf("%x", sha256.Sum256(b))
	s := bufio.NewScanner(strings.NewReader(string(b)))
	s.Buffer(make([]byte, 64*1024), 16*1024*1024)
	var tokens []string
	for s.Scan() {
		tokens = append(tokens, strings.Fields(s.Text())...)
	}
	if err := s.Err(); err != nil {
		return nil, "", err
	}
	return tokens, hash, nil
}

// runResult collects everything written to output files.
type runResult struct {
	TokenCount         int
	CorpusSHA256       string
	MetadataSHA256     string
	Inventory          []ClassInfo
	WithinRegimes      []WithinClassRegime
	WithinStability    []WithinClassStability
	CrossBlock         []CrossBlockRecurrence
	Candidates         []WithinClassCandidate
	ResidualSummary    []ResidualClusterSummary
	ResidualCorrection map[string]EmpiricalStats
	ResidualClasses    []ClassID
	ResidualWindows    []ResidualWindow
	ResidualLabels     []int
	ResidualK          int
	ResidualScale      int
	ResidualAssoc      []ResidualMetadataAssociation
	ResidualCandRows   []ResidualCandidate
	Boundaries         []ConditionalStableBoundary
	RecurringTypes     []RecurringBoundaryType
	Transitions        []ResidualTransitionCell
	ExcludedTokens     int
}

// eligibleSorted returns the eligible classes of one scheme, in a fixed
// deterministic order (by label), from the inventory.
func eligibleSorted(inv []ClassInfo, scheme Scheme) []ClassID {
	var out []ClassID
	for _, ci := range inv {
		if ci.Class.Scheme == scheme && ci.Eligible {
			out = append(out, ci.Class)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Label() < out[j].Label() })
	return out
}

func classInfoOf(inv []ClassInfo, class ClassID) ClassInfo {
	for _, ci := range inv {
		if ci.Class == class {
			return ci
		}
	}
	return ClassInfo{Class: class}
}

// windowSizesFor returns the primary within-class scales plus the 1000
// secondary diagnostic scale when the class's largest block clears the
// higher bar task19 section 10 sets for it.
func windowSizesFor(base []int, largestBlock int) []int {
	sizes := append([]int(nil), base...)
	if largestBlock >= 3000 {
		has1000 := false
		for _, s := range sizes {
			if s == 1000 {
				has1000 = true
			}
		}
		if !has1000 {
			sizes = append(sizes, 1000)
		}
	}
	return sizes
}

// RunAndWrite runs the complete conditional-regime-analyze pipeline and
// writes its outputs. Discovery never sees an "improved" corpus: Part A
// clusters within metadata-controlled blocks, Part B removes the expected
// per-class signature before pooled clustering, and Part C detects change
// points within controlled blocks - all reusing global-regime-analyze's own
// distributional representation and clustering/change-point primitives.
//
// Progress is checkpointed to Config.CheckpointPath after every completed
// class x window_size combo (Parts A/B's sweeps) and, for Part B's global
// permutation correction specifically - by far the most expensive loop in
// the pipeline - after every single permutation replicate. If a matching
// checkpoint (same corpus, metadata, and every parameter) already exists
// when a run starts, already-completed work is loaded from it instead of
// being recomputed, so a run interrupted by a crash, kill, or power loss can
// resume close to where it left off rather than starting over.
func RunAndWrite(c Config) error {
	c = defaults(c)
	if c.ProgressWriter == nil && !c.Quiet {
		c.ProgressWriter = os.Stderr
	}
	p := newProgress(c.ProgressWriter)

	p.begin(1, "Loading corpus, metadata and physical blocks")
	tokens, corpusHash, err := readCorpus(c.CorpusPath)
	if err != nil {
		return fmt.Errorf("read corpus: %w", err)
	}
	currier, hand, metaHash, err := loadTokenLabels(c.TokenMetadataMap)
	if err != nil {
		return fmt.Errorf("load token metadata map: %w", err)
	}
	if len(currier) != len(tokens) {
		return fmt.Errorf("token metadata map has %d tokens but corpus has %d", len(currier), len(tokens))
	}
	allBlocks := buildAllBlocks(currier, hand)
	inventory := classInventory(allBlocks, c.MinClassTokens, c.MinBlockTokens)
	excluded := 0
	for i := range currier {
		if currier[i] == "" || hand[i] == "" {
			excluded++
		}
	}
	r := &runResult{TokenCount: len(tokens), CorpusSHA256: corpusHash, MetadataSHA256: metaHash, Inventory: inventory, ExcludedTokens: excluded, ResidualCorrection: map[string]EmpiricalStats{}}
	p.update(1, 1, "Loading corpus, metadata and physical blocks")

	fingerprint := computeFingerprint(c, corpusHash, metaHash)
	cp, resumed := loadCheckpoint(c.CheckpointPath, fingerprint)
	if !resumed {
		cp = newCheckpoint(fingerprint)
	} else {
		r.WithinRegimes = append(r.WithinRegimes, cp.WithinRegimes...)
		r.WithinStability = append(r.WithinStability, cp.WithinStability...)
		r.CrossBlock = append(r.CrossBlock, cp.CrossBlock...)
		r.Candidates = append(r.Candidates, cp.Candidates...)
		r.ResidualSummary = append(r.ResidualSummary, cp.ResidualSummary...)
		r.Boundaries = append(r.Boundaries, cp.Boundaries...)
		r.RecurringTypes = append(r.RecurringTypes, cp.RecurringTypes...)
		r.Transitions = append(r.Transitions, cp.Transitions...)
		for k, v := range cp.ResidualCorrection {
			r.ResidualCorrection[k] = v
		}
		if c.ProgressWriter != nil {
			fmt.Fprintf(c.ProgressWriter, "Resuming from checkpoint %s: %d within-class combos, %d significance combos, %d residual sweep combos, %d/2 residual corrections already done\n",
				c.CheckpointPath, len(cp.WithinCombosDone), len(cp.SignificanceCombosDone), len(cp.ResidualSweepCombosDone), len(cp.ResidualCorrectionDone))
		}
	}
	checkpoint := func() error { return saveCheckpoint(c.CheckpointPath, cp) }

	pool, err := newExecutorPool(c, fingerprint)
	if err != nil {
		return fmt.Errorf("start executor: %w", err)
	}
	if pool != nil {
		defer pool.Close()
	}

	blocksByScheme := map[Scheme]map[ClassID][]Block{
		SchemeJoint:       blocksByClass(allBlocks[SchemeJoint]),
		SchemeCurrierOnly: blocksByClass(allBlocks[SchemeCurrierOnly]),
		SchemeHandOnly:    blocksByClass(allBlocks[SchemeHandOnly]),
	}
	schemes := []Scheme{SchemeJoint, SchemeCurrierOnly, SchemeHandOnly}
	var allEligible []ClassID
	byClassBlocks := map[ClassID][]Block{}
	for _, scheme := range schemes {
		for _, cid := range eligibleSorted(inventory, scheme) {
			allEligible = append(allEligible, cid)
			byClassBlocks[cid] = blocksByScheme[scheme][cid]
		}
	}

	p.begin(2, "Part A: within-class discovery")
	total := 0
	for _, class := range allEligible {
		total += len(windowSizesFor(c.WindowSizes, classInfoOf(inventory, class).LargestBlock))
	}
	done := 0
	for _, class := range allEligible {
		blocks := byClassBlocks[class]
		info := classInfoOf(inventory, class)
		for _, ws := range windowSizesFor(c.WindowSizes, info.LargestBlock) {
			key := withinComboKey(class, ws)
			if !cp.WithinCombosDone[key] {
				rows, cw := withinClassSweep(tokens, class, blocks, ws, c.KMin, c.KMaxWithin, c.Seed)
				r.WithinRegimes = append(r.WithinRegimes, rows...)
				best := bestByMethod(rows)
				for method, row := range best {
					stab := stabilityForClass(tokens, class, blocks, ws, row.K, method, c.Seed)
					r.WithinStability = append(r.WithinStability, stab)
					r.CrossBlock = append(r.CrossBlock, crossBlockRecurrence(class, ws, method, row.K, cw, row.fullLabels, blocks)...)
				}
				cp.WithinCombosDone[key] = true
				cp.WithinRegimes, cp.WithinStability, cp.CrossBlock = r.WithinRegimes, r.WithinStability, r.CrossBlock
				if err := checkpoint(); err != nil {
					return fmt.Errorf("save checkpoint: %w", err)
				}
			}
			done++
			p.update(done, total, "Part A: within-class discovery")
		}
	}
	p.update(total, total, "Part A: within-class discovery")

	p.begin(3, "Part A: block-aware permutation controls")
	done = 0
	byMethodCache := map[string]map[string]WithinClassRegime{}
	for _, class := range allEligible {
		info := classInfoOf(inventory, class)
		for _, ws := range windowSizesFor(c.WindowSizes, info.LargestBlock) {
			key := withinComboKey(class, ws)
			if cp.SignificanceCombosDone[key] {
				done++
				p.update(done, total, "Part A: block-aware permutation controls")
				continue
			}
			best := byMethodCache[key]
			if best == nil {
				var rows []WithinClassRegime
				for _, row := range r.WithinRegimes {
					if row.Scheme == class.Scheme && row.Class == class.Label() && row.WindowSize == ws {
						rows = append(rows, row)
					}
				}
				best = bestByMethod(rows)
				byMethodCache[key] = best
			}
			if len(best) > 0 && ws < 1000 {
				// window_size=1000 is a secondary diagnostic scale only
				// (task19 section 10): it is reported observed-only and is
				// skipped here, both because it is not part of the primary
				// within-class test and because it is by far the most
				// expensive scale to refit thousands of times.
				cands, err := withinClassSignificanceParallel(c.Context, c.Workers, pool, tokens, class, byClassBlocks[class], ws, best, c.Permutations, c.Seed)
				if err != nil {
					return fmt.Errorf("within-class permutation jobs: %w", err)
				}
				r.Candidates = append(r.Candidates, cands...)
				cp.Candidates = r.Candidates
			}
			cp.SignificanceCombosDone[key] = true
			if err := checkpoint(); err != nil {
				return fmt.Errorf("save checkpoint: %w", err)
			}
			done++
			p.update(done, total, "Part A: block-aware permutation controls")
		}
	}
	if !cp.RefinementDone {
		var err error
		r.Candidates, err = refineTopCandidatesParallel(c.Context, c.Workers, pool, tokens, byClassBlocks, r.Candidates, c.Seed)
		if err != nil {
			return fmt.Errorf("candidate refinement jobs: %w", err)
		}
		cp.Candidates, cp.RefinementDone = r.Candidates, true
		if err := checkpoint(); err != nil {
			return fmt.Errorf("save checkpoint: %w", err)
		}
	} else {
		r.Candidates = cp.Candidates
	}

	p.begin(4, "Part B: residualized feature-space clustering")
	jointEligible := eligibleSorted(inventory, SchemeJoint)
	jointBlocks := map[ClassID][]Block{}
	for _, cid := range jointEligible {
		jointBlocks[cid] = blocksByScheme[SchemeJoint][cid]
	}
	repSteps := 0
	for range c.ResidualWindowSizes {
		repSteps += c.KMaxResidual - c.KMin + 1
	}
	repTotal := repSteps * 2 * 2 // 2 methods x {raw,standardized}
	repDone := 0
	var bestRaw, bestRawHier residualSweepResult
	if b, ok := cp.ResidualSweepCombosDone[residualSweepComboKey("k_medoids", false)]; ok && b {
		bestRaw = cp.BestRaw
	}
	if b, ok := cp.ResidualSweepCombosDone[residualSweepComboKey("hierarchical", false)]; ok && b {
		bestRawHier = cp.BestRawHier
	}
	for _, method := range []string{"k_medoids", "hierarchical"} {
		for _, standardized := range []bool{false, true} {
			comboKey := residualSweepComboKey(method, standardized)
			if cp.ResidualSweepCombosDone[comboKey] {
				repDone += repSteps
				p.update(repDone, repTotal, "Part B: residualized feature-space clustering")
				continue
			}
			res := residualSweepProgress(tokens, jointEligible, jointBlocks, c.ResidualWindowSizes, c.KMin, c.KMaxResidual, method, standardized, c.Seed, func(n int) {
				repDone += n
				p.update(repDone, repTotal, "Part B: residualized feature-space clustering")
			})
			r.ResidualSummary = append(r.ResidualSummary, res.Rows...)
			cp.ResidualSummary = r.ResidualSummary
			cp.ResidualSweepCombosDone[comboKey] = true
			if !standardized && method == "k_medoids" {
				bestRaw, cp.BestRaw = res, res
			}
			if !standardized && method == "hierarchical" {
				bestRawHier, cp.BestRawHier = res, res
			}
			if err := checkpoint(); err != nil {
				return fmt.Errorf("save checkpoint: %w", err)
			}
		}
	}
	p.update(repTotal, repTotal, "Part B: residualized feature-space clustering")

	p.begin(5, "Part B: metadata independence and residual candidates")
	r.ResidualClasses = jointEligible
	r.ResidualWindows = bestRaw.BestWindows
	r.ResidualLabels = bestRaw.BestFullLabels
	r.ResidualK = bestRaw.BestK
	r.ResidualScale = bestRaw.BestWindowSize
	origCurrier, origHand, e := readOriginalGlobalNMI(workdir.Path("metadata-validation", "cluster_metadata_global_summary.tsv"))
	if e != nil {
		origCurrier, origHand = 0, 0
	}
	if len(r.ResidualWindows) > 0 {
		r.ResidualAssoc = residualMetadataAssociations(r.ResidualScale, r.ResidualK, "k_medoids", "raw", r.ResidualWindows, r.ResidualLabels, origCurrier, origHand)
		totalCurriers, totalHands, totalJoints, totalBlocks := countTotals(jointEligible, byClassBlocks)
		r.ResidualCandRows = residualCandidates(r.ResidualScale, r.ResidualK, "k_medoids", "raw", r.ResidualWindows, r.ResidualLabels, totalCurriers, totalHands, totalJoints, totalBlocks)
	}
	correctionTargets := []struct {
		method   string
		observed float64
	}{{"k_medoids", bestRaw.BestSilhouette}, {"hierarchical", bestRawHier.BestSilhouette}}
	for i, target := range correctionTargets {
		key := target.method + "|raw"
		if cp.ResidualCorrectionDone[key] {
			r.ResidualCorrection[key] = cp.ResidualCorrection[key]
			continue
		}
		resumeNull := cp.ResidualCorrectionNull[key]
		stage, combination := "part_b_global_correction", target.method+"|raw"
		completed := checkpointJobsFor(cp.PermutationJobs, stage, combination, c.Permutations)
		stats, err := residualGlobalCorrectionParallelState(c.Context, c.Workers, pool, tokens, jointEligible, jointBlocks, c.ResidualWindowSizes, c.KMin, c.KMaxResidual, target.method, false, target.observed, c.Permutations, c.Seed+int64(i), resumeNull, completed, nil, func(result JobResult) {
			cp.PermutationJobs[checkpointJobKey(result.JobID)] = result.Value
			_ = checkpoint() // best-effort per-job save; the final save remains fatal
		})
		if err != nil {
			return fmt.Errorf("residual correction jobs: %w", err)
		}
		r.ResidualCorrection[key] = stats
		cp.ResidualCorrection[key] = stats
		cp.ResidualCorrectionDone[key] = true
		delete(cp.ResidualCorrectionNull, key)
		prefix := checkpointJobPrefix(stage, combination)
		for jobKey := range cp.PermutationJobs {
			if strings.HasPrefix(jobKey, prefix) {
				delete(cp.PermutationJobs, jobKey)
			}
		}
		if err := checkpoint(); err != nil {
			return fmt.Errorf("save checkpoint: %w", err)
		}
	}
	p.update(1, 1, "Part B: metadata independence and residual candidates")

	p.begin(6, "Part C: conditional boundaries and transition matrix")
	if cp.PartCDone {
		r.Boundaries, r.RecurringTypes, r.Transitions = cp.Boundaries, cp.RecurringTypes, cp.Transitions
	} else {
		r.Boundaries = conditionalBoundaries(tokens, jointEligible, jointBlocks, c.WindowSizes)
		r.RecurringTypes = recurringBoundaryTypes(r.Boundaries)
		if len(r.ResidualWindows) > 0 && r.ResidualK > 0 {
			r.Transitions = residualTransitionMatrix(r.ResidualWindows, r.ResidualLabels, r.ResidualK, c.Permutations, c.Seed+2)
		}
		cp.Boundaries, cp.RecurringTypes, cp.Transitions, cp.PartCDone = r.Boundaries, r.RecurringTypes, r.Transitions, true
		if err := checkpoint(); err != nil {
			return fmt.Errorf("save checkpoint: %w", err)
		}
	}
	p.update(1, 1, "Part C: conditional boundaries and transition matrix")

	p.begin(7, "Writing results")
	if err := writeAll(c, r); err != nil {
		return err
	}
	removeCheckpoint(c.CheckpointPath)
	p.update(1, 1, "Writing results")
	fmt.Printf("Conditional-regime analysis completed for %d tokens (%d eligible classes); results written to %s\n", r.TokenCount, len(jointEligible), c.OutputDir)
	return nil
}

func countTotals(classes []ClassID, byClassBlocks map[ClassID][]Block) (currierN, handN, jointN, blocksN int) {
	curriers, hands := map[string]bool{}, map[string]bool{}
	for _, c := range classes {
		curriers[c.Currier] = true
		hands[c.Hand] = true
		blocksN += len(byClassBlocks[c])
	}
	return len(curriers), len(hands), len(classes), blocksN
}
