package main

//recovered version
import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"
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

func readFileToken(fileName string) ([]Tokens, error) {
	f, err := os.Open(fileName)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	r := bufio.NewScanner(f)
	lines := make([]string, 0)
	for r.Scan() {
		lines = append(lines, r.Text())
	}
	if err := r.Err(); err != nil {
		return nil, err
	}

	freqMap := make(map[string]int)
	positionMap := make(map[string]map[int]int)
	wordBeforeMap := make(map[string]map[string]int)
	wordAfterMap := make(map[string]map[string]int)
	frequencyList := false
	startCountMap := make(map[string]int)
	endCountMap := make(map[string]int)
	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}
		_, _, frequencyList = parseFrequencyLine(line)
		break
	}

	if frequencyList {
		return nil, fmt.Errorf("input contains aggregated token counts; positions and neighboring words require the original text")
	}

	for _, line := range lines {
		lineTokens := strings.Fields(line)
		for position, token := range lineTokens {
			freqMap[token]++
			incrementPosition(positionMap, token, position)
			if position > 0 {
				incrementRelatedToken(wordBeforeMap, token, lineTokens[position-1])
			}
			if position+1 < len(lineTokens) {
				incrementRelatedToken(wordAfterMap, token, lineTokens[position+1])
			}
			if position == 0 {
				startCountMap[token]++
			}
			if position == len(lineTokens)-1 {
				endCountMap[token]++
			}

		}
	}

	result := make([]Tokens, 0, len(freqMap))
	for token, count := range freqMap {
		result = append(result, Tokens{
			Token:            token,
			Count:            count,
			PositionInString: topPositions(positionMap[token], 0),
			WordBefore:       topTokens(wordBeforeMap[token], 0),
			WordAfter:        topTokens(wordAfterMap[token], 0),
			LineStartCount:   startCountMap[token],
			LineEndCount:     endCountMap[token],
		})
	}

	sort.Slice(result, func(i, j int) bool {
		if result[i].Count == result[j].Count {
			return result[i].Token < result[j].Token
		}
		return result[i].Count > result[j].Count
	})
	return result, nil
}

func incrementPosition(positions map[string]map[int]int, token string, position int) {
	if positions[token] == nil {
		positions[token] = make(map[int]int)
	}
	positions[token][position]++
}

func topPositions(counts map[int]int, limit int) []Position {
	positions := make([]Position, 0, len(counts))
	for position, count := range counts {
		positions = append(positions, Position{Position: position, Count: count})
	}

	sort.Slice(positions, func(i, j int) bool {
		if positions[i].Count == positions[j].Count {
			return positions[i].Position < positions[j].Position
		}
		return positions[i].Count > positions[j].Count
	})

	if limit > 0 && len(positions) > limit {
		positions = positions[:limit]
	}
	return positions
}

func incrementRelatedToken(related map[string]map[string]int, token, neighbor string) {
	if related[token] == nil {
		related[token] = make(map[string]int)
	}
	related[token][neighbor]++
}

func topTokens(counts map[string]int, limit int) []Token {
	tokens := make([]Token, 0, len(counts))
	for token, count := range counts {
		tokens = append(tokens, Token{Token: token, Count: count})
	}

	sort.Slice(tokens, func(i, j int) bool {
		if tokens[i].Count == tokens[j].Count {
			return tokens[i].Token < tokens[j].Token
		}
		return tokens[i].Count > tokens[j].Count
	})
	if limit == 0 {
		limit = len(tokens)
	}
	if len(tokens) > limit {
		tokens = tokens[:limit]
	}
	return tokens
}

// parseFrequencyLine parses a line in the form "token: count". The last colon
// is used so that a colon can also be part of the token itself.
func parseFrequencyLine(line string) (string, int, bool) {
	separator := strings.LastIndexByte(line, ':')
	if separator < 0 {
		return "", 0, false
	}

	token := strings.TrimSpace(line[:separator])
	count, err := strconv.Atoi(strings.TrimSpace(line[separator+1:]))
	if token == "" || err != nil || count < 0 {
		return "", 0, false
	}

	return token, count, true
}

func writeTokensYAML(fileName string, tokens []Tokens) error {
	data, err := yaml.Marshal(tokens)
	if err != nil {
		return fmt.Errorf("marshal tokens to YAML: %w", err)
	}

	if err := os.WriteFile(fileName, data, 0o644); err != nil {
		return fmt.Errorf("write YAML file: %w", err)
	}

	return nil
}

func main() {
	flag.Parse()
	if flag.NArg() < 1 {
		fmt.Fprintln(os.Stderr, "Usage: go run . <input_file> [output.yaml]")
		os.Exit(1)
	}

	fileName := flag.Arg(0)
	outputFileName := "output.yaml"
	if flag.NArg() >= 2 {
		outputFileName = flag.Arg(1)
	}

	tokens, err := readFileToken(fileName)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading file: %v\n", err)
		os.Exit(1)
	}

	if err := writeTokensYAML(outputFileName, tokens); err != nil {
		fmt.Fprintf(os.Stderr, "Error writing output: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("YAML file written to %s\n", outputFileName)
}
