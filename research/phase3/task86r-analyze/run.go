package main

import (
	"fmt"
	"os"
	"time"

	"zcore.dev/voinich/internal/evaglyph"
)

const outDir = "research/phase3/task86r"

// run is the full Task86R pipeline entry point.
func run() error {
	if err := os.MkdirAll(outDir, 0o755); err != nil {
		return err
	}
	workDir, err := os.MkdirTemp("", "task86r-run-")
	if err != nil {
		return err
	}
	defer os.RemoveAll(workDir)

	logf("preflight: loading candidate grid")
	candidates, err := loadCandidateGrid("research/phase3/task85a/G1_HYPERPARAMETER_GRID.tsv")
	if err != nil {
		return err
	}
	if len(candidates) != 84 {
		return fmt.Errorf("candidate grid has %d rows, expected 84", len(candidates))
	}

	logf("preflight: verifying authoritative hashes")
	if err := verifyAuthoritativeHashes(); err != nil {
		return fmt.Errorf("authoritative hash verification failed: %w", err)
	}

	logf("preflight: loading corpora")
	split, err := loadSplitPartitions("research/phase3/task85/GRAMMAR_CORPUS_SPLIT.tsv")
	if err != nil {
		return err
	}
	zl3b, err := loadTranscription(corpusSources[0], split)
	if err != nil {
		return err
	}
	it2a, err := loadTranscription(corpusSources[1], split)
	if err != nil {
		return err
	}
	transcriptions := map[string]TranscriptionCorpus{"ZL3b": zl3b, "IT2a": it2a}

	// ---------------- Stage A: MFC calibration ----------------
	logf("stage A: MFC calibration starting (3 generators x 16 populations x 84 candidates)")
	startA := time.Now()
	stageA := runStageA(namespace, candidates, workDir)
	logf("stage A: done in %s (%d jobs, %d failed)", time.Since(startA), stageA.TotalJobs, stageA.FailedJobs)
	idx := NewThresholdIndex(stageA.Thresholds)
	if err := writeCalibrationTables(stageA); err != nil {
		return err
	}
	calibFreezeHash, err := writeCalibrationFreeze(stageA)
	if err != nil {
		return err
	}

	// ---------------- Stage B+C: DEVELOPMENT fit + VALIDATION selection ----------------
	devByT := map[string][]TokenOccurrence{}
	valByT := map[string][]TokenOccurrence{}
	heldByT := map[string][]TokenOccurrence{}
	bitsRealByT := map[string]float64{}
	var allDevFits []DevFitResult
	selByT := map[string]StageCSelection{}
	for _, tname := range []string{"ZL3b", "IT2a"} {
		tc := transcriptions[tname]
		dev := tc.Partition("DEVELOPMENT")
		val := tc.Partition("VALIDATION")
		held := tc.Partition("HELDOUT")
		devByT[tname], valByT[tname], heldByT[tname] = dev, val, held
		bitsReal := bitsPerRealParameter(len(dev))
		bitsRealByT[tname] = bitsReal

		logf("stage B: %s DEVELOPMENT fit (84 candidates)", tname)
		fits := runStageB(dev, tname, candidates)
		allDevFits = append(allDevFits, fits...)

		logf("stage C: %s VALIDATION selection", tname)
		sel := runStageC(dev, val, tname, candidates, bitsReal)
		selByT[tname] = sel
	}
	if err := writeDevFits(allDevFits); err != nil {
		return err
	}
	if err := writeModelSelection(selByT, bitsRealByT, devByT, valByT); err != nil {
		return err
	}
	selectionFreezeHash, err := writeSelectionFreeze(calibFreezeHash, selByT)
	if err != nil {
		return err
	}
	logf("selection freeze: %s", selectionFreezeHash)

	// ---------------- Stage D: HELDOUT confirmatory ----------------
	classes := []string{"M0", "M1", "M2", "M3", "M4", "M5"}
	stageD := map[string]map[string]*StageDResult{} // transcription -> class -> result
	heldF2ByT := map[string]map[string]float64{}
	heldF2ValidByT := map[string]bool{}
	aliasByT := map[string]*GlyphAlias{}
	rawToGlyphsByT := map[string]func(string) []string{}
	alphabetSizeByT := map[string]int{}

	for _, tname := range []string{"ZL3b", "IT2a"} {
		dev, val, held := devByT[tname], valByT[tname], heldByT[tname]
		bitsReal := bitsRealByT[tname]
		sel := selByT[tname]

		fullAlphabet := glyphAlphabet(append(append(append([]TokenOccurrence{}, dev...), val...), held...))
		alias := NewGlyphAlias(fullAlphabet)
		aliasByT[tname] = alias
		alphabetSizeByT[tname] = len(glyphAlphabet(dev))
		rawToGlyphsByT[tname] = evaGlyphsOf

		logf("stage D: %s HELDOUT F2 baseline", tname)
		var heldPops [][]string
		for _, o := range held {
			heldPops = append(heldPops, o.Glyphs)
		}
		heldF2, heldF2Valid, _ := StructuralMetrics(alias, heldPops, int64(SeedFields{Namespace: namespace, ModelClass: "HELDOUT", CandidateID: "REAL", CorpusID: tname, Transcription: tname, Partition: "HELDOUT", Scale: 1, Replicate: 0}.Seed()), workDir)
		heldF2ByT[tname] = heldF2
		heldF2ValidByT[tname] = heldF2Valid

		base := computeTranscriptionBaselines(namespace, tname, sel, dev, val, held)
		stageD[tname] = map[string]*StageDResult{}
		for _, class := range classes {
			logf("stage D: %s %s confirmatory evaluation", tname, class)
			r := runStageDClass(namespace, tname, class, sel, base, dev, val, held, bitsReal, idx, alias, evaGlyphsOf, len(glyphAlphabet(dev)), heldF2, heldF2Valid, workDir)
			stageD[tname][class] = r
		}
	}

	logf("synthesizing cross-transcription verdicts")
	synth := synthesize(stageD, idx)

	if err := writeStageDTables(stageD, synth, idx); err != nil {
		return err
	}
	if err := writeExecutionLedger(stageA, stageD); err != nil {
		return err
	}
	if err := writeResultsManifest(calibFreezeHash, selectionFreezeHash, synth); err != nil {
		return err
	}
	if err := writeReports(stageA, allDevFits, selByT, stageD, synth); err != nil {
		return err
	}
	marker, err := writeFinalFreeze(synth)
	if err != nil {
		return err
	}
	logf("done; final marker=%s", marker)
	return nil
}

func evaGlyphsOf(raw string) []string { return evaglyph.CollapseEVA(raw) }

func logf(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, "[%s] "+format+"\n", append([]interface{}{time.Now().Format("15:04:05")}, args...)...)
}
