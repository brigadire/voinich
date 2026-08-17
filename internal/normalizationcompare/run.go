package normalizationcompare

import (
	"context"
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
	"zcore.dev/voinich/internal/normalization"
	"zcore.dev/voinich/internal/sequenceanalyze"
	"zcore.dev/voinich/internal/workdir"
)

// Config carries every file path and operational (non-scientific) choice
// normalization-compare's CLI exposes. As with structuralprojection.Config
// and conditionalregime.Config, no field here changes a scientific result -
// only where/how the independent random-baseline trials execute.
type Config struct {
	ClassesPath, InputPath, RawAnalysisPath                 string
	NormalizedPattern, AnalysisPattern, SequenceAnalyzerPath string
	OutputPath                                               string
	RandomRuns                                               int
	RandomSeed                                               int64

	Executor         string
	Workers          int
	BaselineExecutor BaselineExecutor
	Context          context.Context

	RemoteListen, TLSCert, TLSKey, ClientCA, RemoteDenyList string
	RemoteTimeout                                           time.Duration
	RemoteRetries                                           int
}

// RunAndWrite reproduces exactly the sequence of scientific computations,
// fatal-error checks and artifact writes normalization-compare's original
// single-file main() performed, differing only in how the up-to-100
// per-threshold random-baseline trials execute (Config.BaselineExecutor).
func RunAndWrite(c Config) error {
	ctx := c.Context
	if ctx == nil {
		ctx = context.Background()
	}
	if c.RandomRuns < 1 {
		return fmt.Errorf("random-baselines must be at least 1")
	}
	classes, err := LoadClasses(c.ClassesPath)
	if err != nil {
		return fmt.Errorf("read classes: %w", err)
	}
	corpus, err := normalization.LoadCorpus(c.InputPath)
	if err != nil {
		return fmt.Errorf("read corpus: %w", err)
	}
	raw, err := LoadSequence(c.RawAnalysisPath)
	if err != nil {
		return fmt.Errorf("read raw analysis: %w", err)
	}

	executor := c.BaselineExecutor
	params := sequenceanalyze.DefaultParameters()
	if executor == nil {
		executor = newDefaultExecutor(classes, corpus, c.RandomSeed, params)
	}
	defer executor.Close()
	workers := c.Workers
	if workers < 1 {
		workers = 1
	}

	output := ComparisonOutput{}
	output.Meta.RandomBaselines = c.RandomRuns
	output.Meta.RandomSeed = c.RandomSeed
	output.Meta.SequenceAnalyzer = c.SequenceAnalyzerPath
	output.Meta.RandomMatching = classes.Meta.RandomMatching
	output.Meta.EmpiricalTests = "upper tail for repeat counts, maximum length, and coverage; lower tail for both entropy metrics; +1 correction"

	for _, model := range classes.Models {
		normalizedPath := fmt.Sprintf(c.NormalizedPattern, model.Label)
		analysisPath := fmt.Sprintf(c.AnalysisPattern, model.Label)
		structuralOutput, err := sequenceanalyze.AnalyzeFile(normalizedPath, params)
		if err != nil {
			return fmt.Errorf("analyze %s: %w", normalizedPath, err)
		}
		if err := WriteAnalysisYAML(analysisPath, structuralOutput); err != nil {
			return err
		}
		structural := FromAnalyzerOutput(structuralOutput)
		if structural.Meta != raw.Meta {
			return fmt.Errorf("corpus invariants changed for threshold %s", model.Label)
		}

		structuralMetrics := ExtractMetrics(structural)
		var randomMetrics []Metrics
		if model.Stats.MultiMemberClasses == 0 {
			// With no merges every matched model is exactly the raw model.
			// Preserve the requested run count without executing identical
			// analyses (and without ever dispatching a trial for it).
			randomMetrics = make([]Metrics, c.RandomRuns)
			for run := range randomMetrics {
				randomMetrics[run] = structuralMetrics
			}
		} else {
			results, err := runBaselines(ctx, executor, model.Label, c.RandomRuns, workers)
			if err != nil {
				return err
			}
			randomMetrics = make([]Metrics, len(results))
			for run, r := range results {
				if r.Meta != raw.Meta {
					return fmt.Errorf("random corpus invariants changed for threshold %s run %d", model.Label, run)
				}
				randomMetrics[run] = r.Metrics
			}
		}
		output.Models = append(output.Models, CompareModel(model, ExtractMetrics(raw), structuralMetrics, randomMetrics))
		fmt.Printf("Compared threshold %s (%d random baselines)\n", model.Label, c.RandomRuns)
	}

	data, err := yaml.Marshal(output)
	if err != nil {
		return err
	}
	if err := workdir.EnsureParent(c.OutputPath); err != nil {
		return err
	}
	if err := os.WriteFile(c.OutputPath, data, 0o644); err != nil {
		return err
	}
	fmt.Printf("Comparison written to %s\n", c.OutputPath)
	return nil
}
