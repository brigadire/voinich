package main

import (
	"fmt"
	"io"
	"math"
	"os"

	"gopkg.in/yaml.v3"
)

type Dataset struct {
	Tokens          []DictionaryToken
	Analysis        map[string]TokenAnalysisInput
	ByToken         map[string]DictionaryToken
	Positions       map[string]map[int]int
	Left            map[string]map[string]int
	Right           map[string]map[string]int
	Outgoing        map[string]int
	Incoming        map[string]int
	CorpusPositions map[int]int
	Meta            Meta
}

func loadDataset(dictionaryPath, analysisPath string) (*Dataset, error) {
	var dictionary []DictionaryToken
	if err := readYAML(dictionaryPath, &dictionary); err != nil {
		return nil, fmt.Errorf("read dictionary: %w", err)
	}
	var analyses []TokenAnalysisInput
	if err := readYAML(analysisPath, &analyses); err != nil {
		return nil, fmt.Errorf("read token analysis: %w", err)
	}

	dataset := &Dataset{
		Tokens:          dictionary,
		Analysis:        make(map[string]TokenAnalysisInput, len(analyses)),
		ByToken:         make(map[string]DictionaryToken, len(dictionary)),
		Positions:       make(map[string]map[int]int, len(dictionary)),
		Left:            make(map[string]map[string]int, len(dictionary)),
		Right:           make(map[string]map[string]int, len(dictionary)),
		Outgoing:        make(map[string]int, len(dictionary)),
		Incoming:        make(map[string]int, len(dictionary)),
		CorpusPositions: make(map[int]int),
	}

	for i, token := range dictionary {
		if err := validateDictionaryToken(token); err != nil {
			return nil, fmt.Errorf("dictionary entry %d: %w", i, err)
		}
		if _, exists := dataset.ByToken[token.Token]; exists {
			return nil, fmt.Errorf("dictionary entry %d: duplicate token %q", i, token.Token)
		}
		dataset.ByToken[token.Token] = token
		dataset.Positions[token.Token] = positionCounts(token.PositionInString)
		dataset.Left[token.Token] = neighborCounts(token.WordBefore)
		dataset.Right[token.Token] = neighborCounts(token.WordAfter)

		outgoing := sumStringCounts(dataset.Right[token.Token])
		incoming := sumStringCounts(dataset.Left[token.Token])
		if outgoing != token.Count-token.LineEndCount {
			return nil, fmt.Errorf("token %q: outgoing transitions %d, want count-end %d", token.Token, outgoing, token.Count-token.LineEndCount)
		}
		if incoming != token.Count-token.LineStartCount {
			return nil, fmt.Errorf("token %q: incoming transitions %d, want count-start %d", token.Token, incoming, token.Count-token.LineStartCount)
		}
		dataset.Outgoing[token.Token] = outgoing
		dataset.Incoming[token.Token] = incoming

		dataset.Meta.TokenOccurrences += token.Count
		dataset.Meta.Lines += token.LineStartCount
		dataset.Meta.Transitions += outgoing
		for position, count := range dataset.Positions[token.Token] {
			dataset.CorpusPositions[position] += count
			dataset.Meta.PositionObservations += count
		}
	}

	endLines := 0
	for _, token := range dictionary {
		endLines += token.LineEndCount
	}
	if endLines != dataset.Meta.Lines {
		return nil, fmt.Errorf("corpus invariant failed: starts=%d, ends=%d", dataset.Meta.Lines, endLines)
	}
	if dataset.Meta.TokenOccurrences-dataset.Meta.Lines != dataset.Meta.Transitions {
		return nil, fmt.Errorf("corpus invariant failed: occurrences-lines=%d, transitions=%d", dataset.Meta.TokenOccurrences-dataset.Meta.Lines, dataset.Meta.Transitions)
	}
	dataset.Meta.DatasetVersion = 1
	dataset.Meta.UniqueTokens = len(dictionary)
	dataset.Meta.PositionCoverage = ratio(dataset.Meta.PositionObservations, dataset.Meta.TokenOccurrences)

	for i, analysis := range analyses {
		if analysis.Token == "" {
			return nil, fmt.Errorf("analysis entry %d: empty token", i)
		}
		if _, exists := dataset.Analysis[analysis.Token]; exists {
			return nil, fmt.Errorf("analysis entry %d: duplicate token %q", i, analysis.Token)
		}
		dataset.Analysis[analysis.Token] = analysis
	}
	if len(dataset.Analysis) != len(dataset.ByToken) {
		return nil, fmt.Errorf("dataset mismatch: dictionary has %d tokens, analysis has %d", len(dataset.ByToken), len(dataset.Analysis))
	}
	if err := validateAnalysis(dataset); err != nil {
		return nil, err
	}
	if err := validateTransitionSymmetry(dataset); err != nil {
		return nil, err
	}
	return dataset, nil
}

func readYAML(path string, destination any) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	decoder := yaml.NewDecoder(file)
	decoder.KnownFields(true)
	if err := decoder.Decode(destination); err != nil {
		return err
	}
	var extra any
	if err := decoder.Decode(&extra); err != io.EOF {
		if err != nil {
			return err
		}
		return fmt.Errorf("expected one YAML document")
	}
	return nil
}

func validateDictionaryToken(token DictionaryToken) error {
	if token.Token == "" {
		return fmt.Errorf("empty token")
	}
	if token.Count < 0 || token.LineStartCount < 0 || token.LineStartCount > token.Count || token.LineEndCount < 0 || token.LineEndCount > token.Count {
		return fmt.Errorf("token %q has invalid occurrence counts", token.Token)
	}
	for _, position := range token.PositionInString {
		if position.Position < 0 || position.Count < 0 {
			return fmt.Errorf("token %q has invalid position", token.Token)
		}
	}
	for _, neighbor := range append(append([]Neighbor(nil), token.WordBefore...), token.WordAfter...) {
		if neighbor.Token == "" || neighbor.Count < 0 {
			return fmt.Errorf("token %q has invalid neighbor", token.Token)
		}
	}
	return nil
}

func validateAnalysis(dataset *Dataset) error {
	for token, dictionary := range dataset.ByToken {
		analysis, exists := dataset.Analysis[token]
		if !exists {
			return fmt.Errorf("dataset mismatch: token %q is absent from analysis", token)
		}
		if analysis.Count != dictionary.Count {
			return fmt.Errorf("dataset mismatch for %q: dictionary count=%d, analysis count=%d", token, dictionary.Count, analysis.Count)
		}
		startProbability := ratio(dictionary.LineStartCount, dictionary.Count)
		endProbability := ratio(dictionary.LineEndCount, dictionary.Count)
		if !almostEqual(analysis.StartProbability, startProbability) || !almostEqual(analysis.EndProbability, endProbability) {
			return fmt.Errorf("dataset mismatch for %q: boundary probabilities are stale", token)
		}
		leftUnique, leftEntropy := entropy(dataset.Left[token])
		rightUnique, rightEntropy := entropy(dataset.Right[token])
		if analysis.Left.Unique != leftUnique || analysis.Right.Unique != rightUnique || !almostEqual(analysis.Left.Entropy, leftEntropy) || !almostEqual(analysis.Right.Entropy, rightEntropy) {
			return fmt.Errorf("dataset mismatch for %q: context statistics are stale", token)
		}
	}
	return nil
}

func validateTransitionSymmetry(dataset *Dataset) error {
	for from, transitions := range dataset.Right {
		for to, count := range transitions {
			if dataset.Left[to][from] != count {
				return fmt.Errorf("transition mismatch %q -> %q: word_after=%d, word_before=%d", from, to, count, dataset.Left[to][from])
			}
		}
	}
	return nil
}

func neighborCounts(neighbors []Neighbor) map[string]int {
	counts := make(map[string]int, len(neighbors))
	for _, neighbor := range neighbors {
		counts[neighbor.Token] += neighbor.Count
	}
	return counts
}

func positionCounts(positions []Position) map[int]int {
	counts := make(map[int]int, len(positions))
	for _, position := range positions {
		counts[position.Position] += position.Count
	}
	return counts
}

func sumStringCounts(counts map[string]int) int {
	total := 0
	for _, count := range counts {
		total += count
	}
	return total
}

func sumPositionCounts(counts map[int]int) int {
	total := 0
	for _, count := range counts {
		total += count
	}
	return total
}

func ratio(numerator, denominator int) float64 {
	if numerator <= 0 || denominator <= 0 {
		return 0
	}
	return float64(numerator) / float64(denominator)
}

func entropy(counts map[string]int) (int, float64) {
	total := sumStringCounts(counts)
	if total == 0 {
		return 0, 0
	}
	unique := 0
	value := 0.0
	for _, count := range counts {
		if count == 0 {
			continue
		}
		unique++
		probability := float64(count) / float64(total)
		value -= probability * math.Log2(probability)
	}
	return unique, value
}

func almostEqual(left, right float64) bool {
	return math.Abs(left-right) <= 1e-9*math.Max(1, math.Max(math.Abs(left), math.Abs(right)))
}
