package main

import (
	"fmt"
	"sync"

	"zcore.dev/voinich/internal/task82b"
)

var (
	carrierMu    sync.Mutex
	carrierCache = map[string]carrierData{}
)

type carrierData struct {
	lines      task82b.Lines
	tokenAtoms []task82b.TokenAtom
	glyphAtoms []task82b.GlyphAtom
	numLines   int
}

func loadCarrier(root, corpusID string) (carrierData, error) {
	carrierMu.Lock()
	defer carrierMu.Unlock()
	if c, ok := carrierCache[corpusID]; ok {
		return c, nil
	}
	lines, err := task82b.LoadLines(root, corpusID, task82b.CarrierPaths[corpusID])
	if err != nil {
		return carrierData{}, err
	}
	tok, gly := task82b.BuildAtoms(lines.Tokens)
	c := carrierData{lines: lines, tokenAtoms: tok, glyphAtoms: gly, numLines: len(lines.Tokens)}
	carrierCache[corpusID] = c
	return c, nil
}

// buildExtractionJobs enumerates every extraction-branch job: 3 carrier
// baselines, then for each of the 20 frozen operators (task82b.txt
// sec.27-30) x 3 carriers: the operator's own output, its
// RANDOM_SUBSEQUENCE_MATCHED null (3 seeds, sec.34), its
// POSITION_STRATIFIED_RANDOM null when NullClass==PER_GROUP (3 seeds,
// sec.35), and its PERIODIC_PHASE null (every other phase, deterministic,
// sec.36) when NullClass==PERIODIC.
const nullSeeds = 3

func buildExtractionJobs(baseSeed int64) []Job {
	var jobs []Job
	for _, corpus := range task82b.CarrierOrder {
		jobs = append(jobs, Job{
			JobID:    fmt.Sprintf("carrier_baseline__%s", corpus),
			Kind:     "carrier_baseline",
			CorpusID: corpus,
		})
		for _, op := range task82b.Registry() {
			jobs = append(jobs, Job{
				JobID:           fmt.Sprintf("operator_output__%s__%s", corpus, op.ID),
				Kind:            "operator_output",
				CorpusID:        corpus,
				OperatorID:      op.ID,
				StructuralClass: op.StructuralClass,
				ExtractionClass: op.ExtractionClass,
				Provenance:      op.Provenance,
			})
			for s := 1; s <= nullSeeds; s++ {
				jobs = append(jobs, Job{
					JobID:      fmt.Sprintf("null_random__%s__%s__seed%d", corpus, op.ID, s),
					Kind:       "operator_null_random",
					CorpusID:   corpus,
					OperatorID: op.ID,
					NullClass:  "RANDOM_SUBSEQUENCE_MATCHED",
					Seed:       baseSeed + int64(s),
				})
			}
			if op.NullClass == "PER_GROUP" {
				for s := 1; s <= nullSeeds; s++ {
					jobs = append(jobs, Job{
						JobID:      fmt.Sprintf("null_stratified__%s__%s__seed%d", corpus, op.ID, s),
						Kind:       "operator_null_stratified",
						CorpusID:   corpus,
						OperatorID: op.ID,
						NullClass:  "POSITION_STRATIFIED_RANDOM",
						Seed:       baseSeed + 1000 + int64(s),
					})
				}
			}
			if op.NullClass == "PERIODIC" {
				for phase := 0; phase < op.Param; phase++ {
					if phase == 0 {
						continue
					}
					jobs = append(jobs, Job{
						JobID:      fmt.Sprintf("null_periodic__%s__%s__phase%d", corpus, op.ID, phase),
						Kind:       "operator_null_periodic",
						CorpusID:   corpus,
						OperatorID: op.ID,
						NullClass:  "PERIODIC_PHASE",
						Period:     op.Param,
						Phase:      phase,
					})
				}
			}
		}
	}
	return jobs
}

func findOperator(id string) (task82b.Operator, bool) {
	for _, op := range task82b.Registry() {
		if op.ID == id {
			return op, true
		}
	}
	return task82b.Operator{}, false
}

func (j Job) runExtraction(root, workDir string) (JobRecord, error) {
	c, err := loadCarrier(root, j.CorpusID)
	if err != nil {
		return JobRecord{}, err
	}
	rec := JobRecord{
		JobID: j.JobID, Kind: j.Kind, CorpusID: j.CorpusID, OperatorID: j.OperatorID,
		StructuralClass: j.StructuralClass, ExtractionClass: j.ExtractionClass, Provenance: j.Provenance,
		NullClass: j.NullClass, Seed: j.Seed, Phase: j.Phase, Period: j.Period,
	}

	if j.Kind == "carrier_baseline" {
		groups := c.lines.Tokens
		f2, ok := safeExtractF2(writeTemp(workDir, groups), rec.JobID, j.CorpusID, 1, workDir+"_f2out", groups)
		rec.Degenerate = !ok
		rec.F2 = f2
		rec.ChosenCount = f2.Tokens
		rec.CandidatePool = f2.Tokens
		ax := task82b.ComputeAX(groups)
		rec.AX = &ax
		return rec, nil
	}

	op, ok := findOperator(j.OperatorID)
	if !ok {
		return JobRecord{}, fmt.Errorf("unknown operator %q", j.OperatorID)
	}
	mainSel := task82b.Apply(op, c.tokenAtoms, c.glyphAtoms, c.numLines)

	var sel task82b.Selection
	switch j.Kind {
	case "operator_output":
		sel = mainSel
	case "operator_null_random":
		sel = task82b.RandomSubsequenceMatched(mainSel, j.Seed)
	case "operator_null_stratified":
		sel = task82b.StratifiedRandom(mainSel, j.Seed)
	case "operator_null_periodic":
		phases := task82b.PeriodicPhases(mainSel)
		s, ok := phases[j.Phase]
		if !ok {
			return JobRecord{}, fmt.Errorf("phase %d not available for operator %s (period %d)", j.Phase, j.OperatorID, mainSel.Period)
		}
		sel = s
	default:
		return JobRecord{}, fmt.Errorf("unknown job kind %q", j.Kind)
	}

	groups := task82b.Render(sel, c.tokenAtoms, c.glyphAtoms, c.numLines)
	rec.ChosenCount = len(sel.Chosen)
	rec.CandidatePool = sel.CandidatePool
	rec.SkippedGroups = mainSel.SkippedGroups
	nonEmptyLines, totalTokens := 0, 0
	typeSet := map[string]bool{}
	for _, g := range groups {
		if len(g) > 0 {
			nonEmptyLines++
		}
		totalTokens += len(g)
		for _, t := range g {
			typeSet[t] = true
		}
	}
	if totalTokens < 20 || len(typeSet) < 2 {
		rec.Degenerate = true
		rec.DegenerateNote = fmt.Sprintf("tokens=%d types=%d nonEmptyLines=%d", totalTokens, len(typeSet), nonEmptyLines)
	}

	f2, ok := safeExtractF2(writeTemp(workDir, groups), rec.JobID, j.CorpusID, 1, workDir+"_f2out", groups)
	if !ok {
		rec.Degenerate = true
		if rec.DegenerateNote == "" {
			rec.DegenerateNote = "F2 extractor error"
		}
	}
	rec.F2 = f2
	ax := task82b.ComputeAX(groups)
	rec.AX = &ax
	return rec, nil
}

func writeTemp(workDir string, groups [][]string) string {
	path := workDir + ".txt"
	if err := task82b.WriteCorpusFile(path, groups); err != nil {
		panic(err)
	}
	return path
}
