// mechanism-space-analyze is Task66's target-blind mechanism-space
// runner: it transforms the plaintext corpora through the frozen
// internal/mechanismspace grid, compares the result against the
// authoritative Task58-65 Voynich fingerprint, and writes every artifact
// under experiments/mechanism-space-v1/. Workers (Execute, in
// internal/mechanismspace) never select models or read a held-out
// target; every selection/aggregation decision lives in this file.
package main

import (
	"flag"
	"fmt"
	"os"

	"zcore.dev/voinich/internal/mechanismspace"
)

const outDir = "experiments/mechanism-space-v1"

// Replicate counts are scoped down from task66 section 40's suggested
// minimums (30 development / 100 final) because DEVELOPMENT's Full
// fingerprint includes the giant-component and topology passes, which are
// non-trivial on ~39k-token corpora; task66 section 72 explicitly allows
// right-sizing compute rather than rerunning the full Task58-65 pipeline
// per grid point. The reduction is recorded in TASK66_DESIGN.md and
// REPORT.md rather than left implicit.
const (
	screeningReplicates   = 5
	developmentReplicates = 12
	finalReplicates       = 40
	ablationReplicates    = 8
	nullReplicates        = 8
)

func main() { os.Exit(run(os.Args[1:])) }

func run(args []string) int {
	if len(args) == 0 {
		usage()
		return 2
	}
	switch args[0] {
	case "transform":
		return cmdTransform(args[1:])
	case "screen":
		return cmdScreen(args[1:])
	case "evaluate":
		return cmdEvaluate(args[1:])
	case "run":
		return cmdRun(args[1:])
	case "report":
		return cmdReport(args[1:])
	default:
		usage()
		return 2
	}
}

func usage() {
	fmt.Fprintln(os.Stderr, "mechanism-space-analyze: transform|screen|evaluate|run|report")
}

func fatal(err error) int {
	fmt.Fprintln(os.Stderr, "error:", err)
	return 1
}

func ensureDir() error { return os.MkdirAll(outDir, 0755) }

// cmdTransform runs exactly one immutable job (task66 section 74): a
// single mechanism against a single corpus file, writing the transformed
// tokens in the one-token-per-line format the rest of this repository's
// independent analyzers already read (loadGeneratedTokens).
func cmdTransform(args []string) int {
	f, in, out, family, mode, seed := transformFlags(args)
	if *in == "" || *out == "" {
		f.Usage()
		return 2
	}
	c, err := mechanismspace.LoadNatural("input", *in)
	if err != nil {
		return fatal(err)
	}
	cfg := mechanismspace.Config{Family: *family, InputMode: mechanismspace.InputMode(*mode), Seed: *seed, StateCount: 4, MacroStates: 2, DriftScale: 20, GroupLen: 4, Grammar: grammarFor(*family), Grouping: groupingFor(*family)}
	o := mechanismspace.Transform(cfg, c)
	var sb []byte
	for _, w := range o.Tokens {
		line := ""
		for i, g := range w {
			if i > 0 {
				line += " "
			}
			line += g
		}
		sb = append(sb, []byte(line+"\n")...)
	}
	if err := os.WriteFile(*out, sb, 0644); err != nil {
		return fatal(err)
	}
	fmt.Printf("%s\t%d\t%d\t%s\n", *out, o.InputUnits, o.OutputGlyphs, cfg.Hash())
	return 0
}

func transformFlags(args []string) (f *flag.FlagSet, in, out, family, mode *string, seed *int64) {
	f = flag.NewFlagSet("transform", flag.ContinueOnError)
	in = f.String("input", "", "plaintext corpus file")
	out = f.String("output", "", "output token file (one token per line, glyphs space-separated)")
	family = f.String("family", "M0", "M0..M11")
	mode = f.String("mode", "WORD_PRESERVING", "WORD_PRESERVING or STREAM")
	seed = f.Int64("seed", 1, "immutable job seed")
	_ = f.Parse(args)
	return
}

func grammarFor(f string) mechanismspace.GrammarLevel {
	switch f {
	case "M3", "M9", "M10", "M11":
		return mechanismspace.GrammarMedium
	}
	return ""
}
func groupingFor(f string) mechanismspace.Grouping {
	switch f {
	case "M8", "M9", "M10", "M11":
		return mechanismspace.StateGrouping
	}
	return ""
}

// cmdScreen runs the SCREENING evaluation set (task66 section 72): the
// full frozen grid, cheap fingerprint, few replicates.
func cmdScreen(args []string) int {
	if err := ensureDir(); err != nil {
		return fatal(err)
	}
	corpora, err := LoadCorpora()
	if err != nil {
		return fatal(err)
	}
	grid := BuildGrid()
	results := RunGrid(grid, corpora, screeningReplicates, mechanismspace.DefaultScreeningOptions(1), 100000)
	grouped := GroupByMechanismCorpus(results)
	if err := WriteResultsTSV(outDir+"/SCREENING_RESULTS.tsv", grouped); err != nil {
		return fatal(err)
	}
	fmt.Println("screening complete:", len(results), "jobs")
	return 0
}

// cmdEvaluate runs the DEVELOPMENT evaluation set: full fingerprint, more
// replicates, plus every DEVELOPMENT-only downstream artifact (Pareto,
// ablation, nulls, sensitivity, topology). It freezes the Pareto frontier
// but does not open HELDOUT_RESULTS.tsv (that only happens in `run`,
// which is explicit about opening held-out exactly once).
func cmdEvaluate(args []string) int {
	if err := ensureDir(); err != nil {
		return fatal(err)
	}
	if _, err := evaluateDevelopment(); err != nil {
		return fatal(err)
	}
	return 0
}

// cmdRun is the full pipeline: design freeze, screening, development,
// Pareto freeze, held-out (opened exactly once), error robustness, and
// the final report.
func cmdRun(args []string) int {
	if err := ensureDir(); err != nil {
		return fatal(err)
	}
	if err := WriteDesignFrozen(outDir); err != nil {
		return fatal(err)
	}

	corpora, err := LoadCorpora()
	if err != nil {
		return fatal(err)
	}
	targets, err := LoadVoynichTargets()
	if err != nil {
		return fatal(err)
	}
	if err := WriteTargetManifest(outDir+"/VOYNICH_TARGET_MANIFEST.tsv", targets); err != nil {
		return fatal(err)
	}
	grid := BuildGrid()
	if err := WriteGridTSV(outDir+"/MECHANISM_GRID.tsv", grid, developmentReplicates, finalReplicates); err != nil {
		return fatal(err)
	}
	if err := WriteComplexityTSV(outDir+"/MECHANISM_COMPLEXITY.tsv", grid); err != nil {
		return fatal(err)
	}

	fmt.Println("screening...")
	screenResults := RunGrid(grid, corpora, screeningReplicates, mechanismspace.DefaultScreeningOptions(1), 100000)
	if err := WriteResultsTSV(outDir+"/SCREENING_RESULTS.tsv", GroupByMechanismCorpus(screenResults)); err != nil {
		return fatal(err)
	}

	fmt.Println("baselines...")
	baselines := computeBaselines(corpora)

	fmt.Println("development...")
	devResults := RunGrid(grid, corpora, developmentReplicates, mechanismspace.DefaultFullOptions(1), 200000)
	devGrouped := GroupByMechanismCorpus(devResults)
	if err := WriteResultsTSV(outDir+"/DEVELOPMENT_RESULTS.tsv", devGrouped); err != nil {
		return fatal(err)
	}
	devRows := ComputeFamilyMetrics(devGrouped, baselines, targets, "DEVELOPMENT", StageDevelopment)
	if err := WriteFamilyMetricsTSV(outDir+"/FAMILY_METRICS.tsv", devRows); err != nil {
		return fatal(err)
	}
	WriteTopologyResultsTSV(outDir+"/TOPOLOGY_RESULTS.tsv", devGrouped, targets)

	fmt.Println("pareto...")
	mechs, scores := MeanFamilyScoresAcrossCorpora(devRows)
	frontier, err := WriteParetoTSV(outDir+"/MECHANISM_PARETO.tsv", mechs, scores)
	if err != nil {
		return fatal(err)
	}
	frontier = withoutIdentity(frontier) // M0 is the baseline control, never a candidate
	if len(frontier) > 6 {
		frontier = frontier[:6] // keep the frozen candidate set small and reportable
	}
	if err := WriteCorpusTransferTSV(outDir+"/CORPUS_TRANSFER.tsv", frontier, devRows); err != nil {
		return fatal(err)
	}
	if err := FreezeCandidates(outDir, frontier); err != nil {
		return fatal(err)
	}

	fmt.Println("ablation...")
	ablationRows := RunAblation(corpora, baselines, targets, ablationReplicates, mechanismspace.DefaultFullOptions(2), outDir+"/ARCHITECTURE_ABLATION.tsv")

	fmt.Println("nulls...")
	RunStateNulls(corpora, nullReplicates, mechanismspace.DefaultFullOptions(3), outDir+"/STATE_NULLS.tsv")
	RunBoundaryNulls(corpora, nullReplicates, mechanismspace.DefaultFullOptions(3), outDir+"/BOUNDARY_NULLS.tsv")

	fmt.Println("plaintext sensitivity...")
	RunPlaintextSensitivity(grid, corpora, mechanismspace.DefaultScreeningOptions(4), outDir+"/PLAINTEXT_SENSITIVITY.tsv")
	RunInformationRetention(grid, corpora, outDir+"/INFORMATION_RETENTION.tsv")
	sensClasses := readSensitivityClasses(outDir + "/PLAINTEXT_SENSITIVITY.tsv")

	fmt.Println("held-out...")
	heldoutRows, overfit, err := RunHeldout(outDir, grid, corpora, baselines, targets, devRows, finalReplicates, mechanismspace.DefaultFullOptions(5))
	if err != nil {
		return fatal(err)
	}
	if err := WriteHeldoutTSV(outDir+"/HELDOUT_RESULTS.tsv", heldoutRows, overfit); err != nil {
		return fatal(err)
	}

	fmt.Println("error robustness...")
	RunErrorRobustness(frontier, grid, corpora, mechanismspace.DefaultScreeningOptions(6), outDir+"/ERROR_ROBUSTNESS.tsv")

	fmt.Println("final architecture + report...")
	verdicts := DeriveFinalArchitecture(devRows, ablationRows, sensClasses)
	if err := WriteFinalArchitectureTSV(outDir+"/FINAL_ARCHITECTURE.tsv", verdicts); err != nil {
		return fatal(err)
	}
	if err := WriteManifest(outDir, frontier, overfit); err != nil {
		return fatal(err)
	}
	if err := WriteReport(outDir, targets, devRows, ablationRows, verdicts, frontier, overfit, sensClasses); err != nil {
		return fatal(err)
	}

	fmt.Println("done")
	return 0
}

// cmdReport regenerates REPORT.md and FINAL_ARCHITECTURE.tsv purely by
// reading back the TSVs a prior `run`/`evaluate` already wrote, without
// recomputing anything - so a report edit or a REPORT.md prose fix never
// silently reruns (and never has a chance to touch) the frozen candidate
// search.
func cmdReport(args []string) int {
	dir := outDir
	if len(args) > 0 {
		dir = args[0]
	}
	targets, err := LoadVoynichTargets()
	if err != nil {
		return fatal(err)
	}
	devRows := readFamilyMetricsTSV(dir+"/FAMILY_METRICS.tsv", "mechanism", "corpus", "family", "progress", "overall_status")
	ablationRows := readFamilyMetricsTSV(dir+"/ARCHITECTURE_ABLATION.tsv", "operation_combination", "corpus", "family", "progress", "overall_status")
	if len(devRows) == 0 {
		return fatal(fmt.Errorf("no FAMILY_METRICS.tsv found under %s; run `run` or `evaluate` first", dir))
	}
	verdicts := readFinalArchitectureTSV(dir + "/FINAL_ARCHITECTURE.tsv")
	if len(verdicts) == 0 {
		sensClasses := readSensitivityClasses(dir + "/PLAINTEXT_SENSITIVITY.tsv")
		verdicts = DeriveFinalArchitecture(devRows, ablationRows, sensClasses)
		if err := WriteFinalArchitectureTSV(dir+"/FINAL_ARCHITECTURE.tsv", verdicts); err != nil {
			return fatal(err)
		}
	}
	frontier := readFrontierFromParetoTSV(dir + "/MECHANISM_PARETO.tsv")
	overfit := readOverfitFromHeldoutTSV(dir + "/HELDOUT_RESULTS.tsv")
	sensClasses := readSensitivityClasses(dir + "/PLAINTEXT_SENSITIVITY.tsv")
	if err := WriteReport(dir, targets, devRows, ablationRows, verdicts, frontier, overfit, sensClasses); err != nil {
		return fatal(err)
	}
	fmt.Println("report regenerated from existing artifacts in", dir)
	return 0
}

func evaluateDevelopment() ([]FamilyMetricsRow, error) {
	corpora, err := LoadCorpora()
	if err != nil {
		return nil, err
	}
	targets, err := LoadVoynichTargets()
	if err != nil {
		return nil, err
	}
	grid := BuildGrid()
	baselines := computeBaselines(corpora)
	devResults := RunGrid(grid, corpora, developmentReplicates, mechanismspace.DefaultFullOptions(1), 200000)
	devGrouped := GroupByMechanismCorpus(devResults)
	if err := WriteResultsTSV(outDir+"/DEVELOPMENT_RESULTS.tsv", devGrouped); err != nil {
		return nil, err
	}
	rows := ComputeFamilyMetrics(devGrouped, baselines, targets, "DEVELOPMENT", StageDevelopment)
	return rows, WriteFamilyMetricsTSV(outDir+"/FAMILY_METRICS.tsv", rows)
}

func withoutIdentity(frontier []string) []string {
	out := make([]string, 0, len(frontier))
	for _, m := range frontier {
		if m != "M0_IDENTITY" {
			out = append(out, m)
		}
	}
	return out
}

func computeBaselines(corpora map[string]mechanismspace.Corpus) map[string]mechanismspace.Fingerprint {
	out := map[string]mechanismspace.Fingerprint{}
	for name, c := range corpora {
		o := mechanismspace.Transform(mechanismspace.Config{Family: "M0"}, c)
		out[name] = mechanismspace.ComputeFingerprint(o.Tokens, o.Lines, mechanismspace.DefaultFullOptions(1))
	}
	return out
}
