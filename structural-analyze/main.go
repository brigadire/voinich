package main

import (
	"flag"
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

func main() {
	dictionaryPath := flag.String("dictionary", "dataset/dictionary.yaml", "path to dictionary.yaml")
	analysisPath := flag.String("analysis", "dataset/tokens_analysis.yaml", "path to tokens_analysis.yaml")
	outputPath := flag.String("output", "structural_analysis.yaml", "path to the result YAML")
	minTokenCount := flag.Int("min-token-count", 10, "minimum token frequency for ranked sections")
	minTransitionCount := flag.Int("min-transition-count", 3, "minimum transition frequency")
	minContextObservations := flag.Int("min-context-observations", 10, "minimum observed predecessors or successors for predictability rankings")
	minSelfTransitionCount := flag.Int("min-self-transition-count", 3, "minimum self-transition frequency")
	reliabilityPrior := flag.Float64("reliability-prior", 10, "pseudo-count used for ranking reliability")
	minSimilarity := flag.Float64("min-similarity", 0.7, "minimum raw equivalence similarity")
	maxItems := flag.Int("max-items", 100, "maximum entries per section; 0 means unlimited")
	maxEquivalenceCandidates := flag.Int("max-equivalence-candidates", 0, "maximum equivalence candidates; 0 means unlimited")
	dominantLimit := flag.Int("dominant-context-limit", 5, "maximum dominant neighbors per constrained token")
	flag.Parse()

	parameters := Parameters{
		MinTokenCountForRanking:  *minTokenCount,
		MinTransitionCount:       *minTransitionCount,
		MinContextObservations:   *minContextObservations,
		MinSelfTransitionCount:   *minSelfTransitionCount,
		ReliabilityPriorCount:    *reliabilityPrior,
		MinEquivalenceSimilarity: *minSimilarity,
		MaxItemsPerSection:       *maxItems,
		MaxEquivalenceCandidates: *maxEquivalenceCandidates,
		DominantContextLimit:     *dominantLimit,
	}
	if err := validateParameters(parameters); err != nil {
		fmt.Fprintf(os.Stderr, "Invalid parameters: %v\n", err)
		os.Exit(2)
	}

	dataset, err := loadDataset(*dictionaryPath, *analysisPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading dataset: %v\n", err)
		os.Exit(1)
	}
	result := buildOutput(dataset, parameters)
	if err := writeOutput(*outputPath, result); err != nil {
		fmt.Fprintf(os.Stderr, "Error writing output: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("YAML file written to %s\n", *outputPath)
}

func validateParameters(parameters Parameters) error {
	if parameters.MinTokenCountForRanking < 1 {
		return fmt.Errorf("min-token-count must be at least 1")
	}
	if parameters.MinTransitionCount < 1 {
		return fmt.Errorf("min-transition-count must be at least 1")
	}
	if parameters.MinContextObservations < 1 {
		return fmt.Errorf("min-context-observations must be at least 1")
	}
	if parameters.MinSelfTransitionCount < 1 {
		return fmt.Errorf("min-self-transition-count must be at least 1")
	}
	if parameters.ReliabilityPriorCount < 0 {
		return fmt.Errorf("reliability-prior cannot be negative")
	}
	if parameters.MinEquivalenceSimilarity < 0 || parameters.MinEquivalenceSimilarity > 1 {
		return fmt.Errorf("min-similarity must be in [0,1]")
	}
	if parameters.MaxItemsPerSection < 0 {
		return fmt.Errorf("max-items cannot be negative")
	}
	if parameters.MaxEquivalenceCandidates < 0 {
		return fmt.Errorf("max-equivalence-candidates cannot be negative")
	}
	if parameters.DominantContextLimit < 0 {
		return fmt.Errorf("dominant-context-limit cannot be negative")
	}
	return nil
}

func writeOutput(path string, output Output) error {
	data, err := yaml.Marshal(output)
	if err != nil {
		return fmt.Errorf("marshal YAML: %w", err)
	}
	if err := os.WriteFile(path, data, 0o644); err != nil {
		return err
	}
	return nil
}
