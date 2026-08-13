package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

type corpusLine struct {
	Tokens []string
	Page   int
}

type corpus struct {
	Lines          []corpusLine
	Pages          [][]int
	Counts         map[string]int
	ExplicitBreaks int
}

func loadDictionary(path string) ([]DictionaryToken, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	decoder := yaml.NewDecoder(f)
	decoder.KnownFields(true)
	var result []DictionaryToken
	if err := decoder.Decode(&result); err != nil {
		return nil, err
	}
	var extra any
	if err := decoder.Decode(&extra); err != io.EOF {
		if err != nil {
			return nil, err
		}
		return nil, fmt.Errorf("dictionary: expected one YAML document")
	}
	seen := make(map[string]bool)
	for i, token := range result {
		if token.Token == "" || token.Count < 0 || token.LineStartCount < 0 || token.LineEndCount < 0 || token.LineStartCount > token.Count || token.LineEndCount > token.Count {
			return nil, fmt.Errorf("dictionary entry %d has invalid counts", i)
		}
		if seen[token.Token] {
			return nil, fmt.Errorf("dictionary contains duplicate token %q", token.Token)
		}
		seen[token.Token] = true
		positionTotal := 0
		for _, p := range token.PositionInString {
			if p.Position < 0 || p.Count < 0 {
				return nil, fmt.Errorf("token %q has invalid position", token.Token)
			}
			positionTotal += p.Count
		}
		if positionTotal != token.Count {
			return nil, fmt.Errorf("token %q position count %d differs from count %d", token.Token, positionTotal, token.Count)
		}
	}
	return result, nil
}

func isPageMarker(trimmed string) bool {
	lower := strings.ToLower(trimmed)
	return strings.HasPrefix(lower, "# page:") || strings.HasPrefix(lower, "#page:") || (strings.HasPrefix(lower, "===") && strings.Contains(lower, "page") && strings.HasSuffix(lower, "==="))
}

func loadCorpus(path string) (*corpus, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	c := &corpus{Counts: make(map[string]int)}
	page := 0
	pendingBreak := false
	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 64*1024), 16*1024*1024)
	for scanner.Scan() {
		raw := scanner.Text()
		parts := strings.Split(raw, "\f")
		for partIndex, part := range parts {
			trimmed := strings.TrimSpace(part)
			marker := isPageMarker(trimmed)
			if partIndex > 0 && len(c.Lines) > 0 {
				pendingBreak = true
			}
			if marker || trimmed == "" {
				if len(c.Lines) > 0 {
					pendingBreak = true
				}
				continue
			}
			if pendingBreak {
				page++
				c.ExplicitBreaks++
				pendingBreak = false
			}
			tokens := strings.Fields(part)
			lineIndex := len(c.Lines)
			c.Lines = append(c.Lines, corpusLine{Tokens: tokens, Page: page})
			for _, token := range tokens {
				c.Counts[token]++
			}
			for len(c.Pages) <= page {
				c.Pages = append(c.Pages, nil)
			}
			c.Pages[page] = append(c.Pages[page], lineIndex)
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	if len(c.Lines) == 0 {
		return nil, fmt.Errorf("corpus contains no non-empty lines")
	}
	return c, nil
}

func validateCorpusDictionary(c *corpus, dictionary []DictionaryToken) error {
	if len(c.Counts) != len(dictionary) {
		return fmt.Errorf("corpus has %d unique tokens, dictionary has %d", len(c.Counts), len(dictionary))
	}
	for _, token := range dictionary {
		if c.Counts[token.Token] != token.Count {
			return fmt.Errorf("corpus count for %q is %d, dictionary count is %d", token.Token, c.Counts[token.Token], token.Count)
		}
	}
	return nil
}
