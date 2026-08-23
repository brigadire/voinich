package main

import (
	"flag"
	"fmt"
	"io"
	"math"
	"os"
	"sort"

	"gopkg.in/yaml.v3"
	"zcore.dev/voinich/internal/workdir"
)

type Position struct {
	Position int `yaml:"position"`
	Count    int `yaml:"count"`
}

type Token struct {
	Token string `yaml:"token"`
	Count int    `yaml:"count"`
}

type Tokens struct {
	Token            string     `yaml:"token"`
	Count            int        `yaml:"count"`
	PositionInString []Position `yaml:"position_in_string"`
	WordBefore       []Token    `yaml:"word_before"`
	WordAfter        []Token    `yaml:"word_after"`
	LineStartCount   int        `yaml:"line_start_count"`
	LineEndCount     int        `yaml:"line_end_count"`
}

type EnvironmentAnalysis struct {
	Unique  int     `yaml:"unique"`
	Entropy float64 `yaml:"entropy"`
}

type StructuralScores struct {
	PositionalSpecialization float64 `yaml:"positional_specialization"`
	SuccessorRestriction     float64 `yaml:"successor_restriction"`
	PredecessorRestriction   float64 `yaml:"predecessor_restriction"`
}

type TransitionAnalysis struct {
	Token              string  `yaml:"token"`
	Count              int     `yaml:"count"`
	Probability        float64 `yaml:"probability"`
	ReverseCount       int     `yaml:"reverse_count"`
	ReverseProbability float64 `yaml:"reverse_probability"`
	Asymmetry          float64 `yaml:"asymmetry"`
}

type SelfTransitionAnalysis struct {
	Count       int     `yaml:"count"`
	Probability float64 `yaml:"probability"`
}

type TokenAnalysis struct {
	Token            string                 `yaml:"token"`
	Count            int                    `yaml:"count"`
	StartProbability float64                `yaml:"start_probability"`
	EndProbability   float64                `yaml:"end_probability"`
	Left             EnvironmentAnalysis    `yaml:"left"`
	Right            EnvironmentAnalysis    `yaml:"right"`
	Transitions      []TransitionAnalysis   `yaml:"transitions"`
	SelfTransition   SelfTransitionAnalysis `yaml:"self_transition"`
	StructuralScores StructuralScores       `yaml:"structural_scores"`
}

func readFileToken(fileName string) ([]Tokens, map[string]map[string]int, map[string]map[string]int, error) {
	f, err := os.Open(fileName)
	if err != nil {
		return nil, nil, nil, err
	}
	defer f.Close()

	decoder := yaml.NewDecoder(f)
	decoder.KnownFields(true)

	var tokens []Tokens
	if err := decoder.Decode(&tokens); err != nil {
		return nil, nil, nil, fmt.Errorf("decode YAML: %w", err)
	}

	var extra any
	if err := decoder.Decode(&extra); err != io.EOF {
		if err != nil {
			return nil, nil, nil, fmt.Errorf("decode trailing YAML: %w", err)
		}
		return nil, nil, nil, fmt.Errorf("input must contain exactly one YAML document")
	}

	wordBeforeMap := make(map[string]map[string]int, len(tokens))
	wordAfterMap := make(map[string]map[string]int, len(tokens))
	seen := make(map[string]struct{}, len(tokens))

	for i, token := range tokens {
		if err := validateToken(token); err != nil {
			return nil, nil, nil, fmt.Errorf("token entry %d: %w", i, err)
		}
		if _, exists := seen[token.Token]; exists {
			return nil, nil, nil, fmt.Errorf("token entry %d: duplicate token %q", i, token.Token)
		}
		seen[token.Token] = struct{}{}

		wordBeforeMap[token.Token] = neighborCounts(token.WordBefore)
		wordAfterMap[token.Token] = neighborCounts(token.WordAfter)
	}

	sort.Slice(tokens, func(i, j int) bool {
		if tokens[i].Count == tokens[j].Count {
			return tokens[i].Token < tokens[j].Token
		}
		return tokens[i].Count > tokens[j].Count
	})

	return tokens, wordBeforeMap, wordAfterMap, nil
}

func validateToken(token Tokens) error {
	if token.Token == "" {
		return fmt.Errorf("token is empty")
	}
	if token.Count < 0 {
		return fmt.Errorf("token %q has negative count", token.Token)
	}
	if token.LineStartCount < 0 || token.LineStartCount > token.Count {
		return fmt.Errorf("token %q has invalid line_start_count", token.Token)
	}
	if token.LineEndCount < 0 || token.LineEndCount > token.Count {
		return fmt.Errorf("token %q has invalid line_end_count", token.Token)
	}
	for _, position := range token.PositionInString {
		if position.Position < 0 || position.Count < 0 {
			return fmt.Errorf("token %q has invalid position data", token.Token)
		}
	}
	for _, neighbor := range append(append([]Token(nil), token.WordBefore...), token.WordAfter...) {
		if neighbor.Token == "" || neighbor.Count < 0 {
			return fmt.Errorf("token %q has invalid neighbor data", token.Token)
		}
	}
	return nil
}

func neighborCounts(neighbors []Token) map[string]int {
	counts := make(map[string]int, len(neighbors))
	for _, neighbor := range neighbors {
		counts[neighbor.Token] += neighbor.Count
	}
	return counts
}

func countUniqueAndEntropy(counts map[string]int) (int, float64) {
	total := sumCounts(counts)
	if total == 0 {
		return 0, 0
	}

	entropy := 0.0
	unique := 0
	keys := make([]string, 0, len(counts))
	for key := range counts {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		count := counts[key]
		if count == 0 {
			continue
		}
		unique++
		p := float64(count) / float64(total)
		entropy -= p * math.Log2(p)
	}
	return unique, entropy
}

func sumCounts(counts map[string]int) int {
	total := 0
	for _, count := range counts {
		total += count
	}
	return total
}

// calculateRestriction returns 1 for a deterministic environment and tends
// to 0 as every observed transition goes to a different neighbor.
func calculateRestriction(entropy float64, uniqueOutcomes int) float64 {
	if uniqueOutcomes == 0 {
		return 0
	}
	if uniqueOutcomes == 1 {
		return 1
	}
	return clamp01(1 - entropy/math.Log2(float64(uniqueOutcomes)))
}

func clamp01(value float64) float64 {
	if value < 0 {
		return 0
	}
	if value > 1 {
		return 1
	}
	return value
}

func analyzeTokens(tokens []Tokens, wordBeforeMap, wordAfterMap map[string]map[string]int) []TokenAnalysis {
	corpusPositions := aggregateCorpusPositions(tokens)
	outgoingTotals := make(map[string]int, len(tokens))
	for token, transitions := range wordAfterMap {
		outgoingTotals[token] = sumCounts(transitions)
	}

	analyses := make([]TokenAnalysis, len(tokens))
	for i, token := range tokens {
		leftTokens := wordBeforeMap[token.Token]
		rightTokens := wordAfterMap[token.Token]
		leftUnique, leftEntropy := countUniqueAndEntropy(leftTokens)
		rightUnique, rightEntropy := countUniqueAndEntropy(rightTokens)
		rightTotal := outgoingTotals[token.Token]

		selfCount := rightTokens[token.Token]
		analyses[i] = TokenAnalysis{
			Token:            token.Token,
			Count:            token.Count,
			StartProbability: probability(token.LineStartCount, token.Count),
			EndProbability:   probability(token.LineEndCount, token.Count),
			Left:             EnvironmentAnalysis{Unique: leftUnique, Entropy: leftEntropy},
			Right:            EnvironmentAnalysis{Unique: rightUnique, Entropy: rightEntropy},
			Transitions:      analyzeTransitions(token.Token, rightTokens, wordAfterMap, outgoingTotals),
			SelfTransition: SelfTransitionAnalysis{
				Count:       selfCount,
				Probability: probability(selfCount, rightTotal),
			},
			StructuralScores: StructuralScores{
				PositionalSpecialization: positionalSpecialization(token.PositionInString, corpusPositions),
				SuccessorRestriction:     calculateRestriction(rightEntropy, rightUnique),
				PredecessorRestriction:   calculateRestriction(leftEntropy, leftUnique),
			},
		}
	}
	return analyses
}

func probability(count, total int) float64 {
	if count <= 0 || total <= 0 {
		return 0
	}
	return float64(count) / float64(total)
}

func analyzeTransitions(from string, transitions map[string]int, allTransitions map[string]map[string]int, outgoingTotals map[string]int) []TransitionAnalysis {
	result := make([]TransitionAnalysis, 0, len(transitions))
	for to, count := range transitions {
		forwardProbability := probability(count, outgoingTotals[from])
		reverseCount := allTransitions[to][from]
		reverseProbability := probability(reverseCount, outgoingTotals[to])
		asymmetry := 0.0
		if denominator := forwardProbability + reverseProbability; denominator > 0 {
			asymmetry = (forwardProbability - reverseProbability) / denominator
		}
		result = append(result, TransitionAnalysis{
			Token:              to,
			Count:              count,
			Probability:        forwardProbability,
			ReverseCount:       reverseCount,
			ReverseProbability: reverseProbability,
			Asymmetry:          asymmetry,
		})
	}

	sort.Slice(result, func(i, j int) bool {
		if result[i].Count == result[j].Count {
			return result[i].Token < result[j].Token
		}
		return result[i].Count > result[j].Count
	})
	return result
}

func aggregateCorpusPositions(tokens []Tokens) map[int]int {
	positions := make(map[int]int)
	for _, token := range tokens {
		for _, position := range token.PositionInString {
			positions[position.Position] += position.Count
		}
	}
	return positions
}

// positionalSpecialization is the Jensen-Shannon divergence, in bits,
// between a token's position distribution and the corpus distribution.
// With base-2 logarithms its value is in the range [0, 1].
func positionalSpecialization(positions []Position, corpus map[int]int) float64 {
	tokenCounts := make(map[int]int, len(positions))
	for _, position := range positions {
		tokenCounts[position.Position] += position.Count
	}
	tokenTotal := sumPositionCounts(tokenCounts)
	corpusTotal := sumPositionCounts(corpus)
	if tokenTotal == 0 || corpusTotal == 0 {
		return 0
	}

	divergence := 0.0
	allPositions := make(map[int]struct{}, len(tokenCounts)+len(corpus))
	for position := range tokenCounts {
		allPositions[position] = struct{}{}
	}
	for position := range corpus {
		allPositions[position] = struct{}{}
	}
	orderedPositions := make([]int, 0, len(allPositions))
	for position := range allPositions {
		orderedPositions = append(orderedPositions, position)
	}
	sort.Ints(orderedPositions)
	for _, position := range orderedPositions {
		p := float64(tokenCounts[position]) / float64(tokenTotal)
		q := float64(corpus[position]) / float64(corpusTotal)
		middle := (p + q) / 2
		if p > 0 {
			divergence += 0.5 * p * math.Log2(p/middle)
		}
		if q > 0 {
			divergence += 0.5 * q * math.Log2(q/middle)
		}
	}
	return clamp01(divergence)
}

func sumPositionCounts(counts map[int]int) int {
	total := 0
	for _, count := range counts {
		total += count
	}
	return total
}

func writeTokensAnalysisYAML(fileName string, analyses []TokenAnalysis) error {
	data, err := yaml.Marshal(analyses)
	if err != nil {
		return fmt.Errorf("marshal analysis to YAML: %w", err)
	}
	if err := os.WriteFile(fileName, data, 0o644); err != nil {
		return fmt.Errorf("write YAML file: %w", err)
	}
	return nil
}

func main() {
	flag.Parse()
	if flag.NArg() < 1 {
		fmt.Fprintln(os.Stderr, "Usage: dict-analyze <dictionary.yaml> [workdir/dataset/tokens_analysis.yaml]")
		os.Exit(1)
	}

	outputFileName := workdir.Path("dataset", "tokens_analysis.yaml")
	if flag.NArg() >= 2 {
		outputFileName = flag.Arg(1)
	}

	tokens, wordBeforeMap, wordAfterMap, err := readFileToken(flag.Arg(0))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading file: %v\n", err)
		os.Exit(1)
	}

	analyses := analyzeTokens(tokens, wordBeforeMap, wordAfterMap)
	if err := workdir.EnsureParent(outputFileName); err != nil {
		fmt.Fprintf(os.Stderr, "Error creating output directory: %v\n", err)
		os.Exit(1)
	}
	if err := writeTokensAnalysisYAML(outputFileName, analyses); err != nil {
		fmt.Fprintf(os.Stderr, "Error writing output: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("YAML file written to %s\n", outputFileName)
}
