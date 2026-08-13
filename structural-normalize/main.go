package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
	"zcore.dev/voinich/internal/normalization"
	"zcore.dev/voinich/internal/workdir"
)

func main() {
	inputPath := flag.String("input", "data_work/ZL3b-x7.txt", "IVTT -x7 derived corpus")
	structuralPath := flag.String("structural", workdir.Path("dataset", "structural_analysis.yaml"), "structural analysis YAML")
	outputPattern := flag.String("output", workdir.Path("normalized.txt"), "normalized corpus base name")
	classesPath := flag.String("classes", workdir.Path("structural_classes.yaml"), "class-map YAML")
	thresholdText := flag.String("thresholds", "0.70,0.75,0.80,0.85,0.90", "comma-separated complete-link thresholds")
	minPosition := flag.Float64("min-position-similarity", 0, "minimum position similarity")
	minLeft := flag.Float64("min-left-context-similarity", 0, "minimum left-context similarity")
	minRight := flag.Float64("min-right-context-similarity", 0, "minimum right-context similarity")
	minCount := flag.Int("min-token-count", 0, "minimum token count; 0 reads structural metadata or uses 10")
	singletonMode := flag.String("singleton-mode", "preserve", "preserve or class")
	randomBaselines := flag.Int("random-baselines", 100, "matched random runs recorded in experiment metadata")
	randomSeed := flag.Int64("random-seed", 1, "base seed for matched random models")
	flag.Parse()

	if *singletonMode != "preserve" && *singletonMode != "class" {
		fatal("singleton-mode must be preserve or class")
	}
	if *randomBaselines < 0 {
		fatal("random-baselines cannot be negative")
	}
	thresholds, err := normalization.ParseThresholds(*thresholdText)
	if err != nil {
		fatal(err.Error())
	}
	structural, err := normalization.LoadStructural(*structuralPath)
	if err != nil {
		fatal(fmt.Sprintf("read structural analysis: %v", err))
	}
	effectiveMinCount := *minCount
	if effectiveMinCount == 0 {
		effectiveMinCount = structural.Parameters.MinTokenCountForRanking
		if effectiveMinCount == 0 {
			effectiveMinCount = 10
		}
	}
	corpus, err := normalization.LoadCorpus(*inputPath)
	if err != nil {
		fatal(fmt.Sprintf("read corpus: %v", err))
	}
	config := normalization.Config{
		Thresholds: thresholds, MinPositionSimilarity: *minPosition, MinLeftContextSimilarity: *minLeft,
		MinRightContextSimilarity: *minRight, MinTokenCount: effectiveMinCount, SingletonMode: *singletonMode,
		RandomBaselines: *randomBaselines, RandomSeed: *randomSeed,
	}
	models, _, err := normalization.BuildModels(corpus, structural, config)
	if err != nil {
		fatal(fmt.Sprintf("build classes: %v", err))
	}
	if err := workdir.EnsureParent(*outputPattern); err != nil {
		fatal(fmt.Sprintf("create output directory: %v", err))
	}
	if err := workdir.EnsureParent(*classesPath); err != nil {
		fatal(fmt.Sprintf("create classes directory: %v", err))
	}
	for _, model := range models {
		path := thresholdPath(*outputPattern, model.Label, len(models))
		if err := normalization.WriteNormalized(path, corpus, normalization.Mapping(model, *singletonMode)); err != nil {
			fatal(fmt.Sprintf("write %s: %v", path, err))
		}
		fmt.Printf("Normalized corpus written to %s\n", path)
	}
	result := normalization.ClassesOutput{
		Meta: normalization.ClassesMeta{
			InputCorpus: *inputPath, StructuralAnalysis: *structuralPath, SingletonMode: *singletonMode,
			MinTokenCount: effectiveMinCount, MinPositionSimilarity: *minPosition,
			MinLeftContextSimilarity: *minLeft, MinRightContextSimilarity: *minRight,
			RandomBaselines: *randomBaselines, RandomSeed: *randomSeed,
			Clustering:     "deterministic agglomerative complete-link; every member pair must be present and satisfy all thresholds",
			RandomMatching: "same multi-member class sizes; members sampled without replacement from logarithmic base-2 frequency bins with nearest-bin fallback",
		},
		Thresholds: thresholds,
		Models:     models,
	}
	data, err := yaml.Marshal(result)
	if err != nil {
		fatal(fmt.Sprintf("encode classes: %v", err))
	}
	if err := os.WriteFile(*classesPath, data, 0o644); err != nil {
		fatal(fmt.Sprintf("write classes: %v", err))
	}
	fmt.Printf("Class map written to %s\n", *classesPath)
}

func thresholdPath(pattern, label string, modelCount int) string {
	if strings.Contains(pattern, "%s") {
		return fmt.Sprintf(pattern, label)
	}
	if modelCount == 1 {
		return pattern
	}
	extension := filepath.Ext(pattern)
	base := strings.TrimSuffix(pattern, extension)
	return base + "_" + label + extension
}

func fatal(message string) {
	fmt.Fprintln(os.Stderr, "Error:", message)
	os.Exit(1)
}
