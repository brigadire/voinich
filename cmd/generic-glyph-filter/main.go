// Command generic-glyph-filter is a Task79c control-preparation step. It
// drops whitespace-delimited tokens that contain no Unicode letter or
// digit at all (e.g. a bare "_", "-", ":", or a punctuation-only run) from
// a plain-text corpus, preserving line structure otherwise (a line left
// empty after filtering is also dropped). This is required, not optional:
// internal/fingerprintv2's loadCorpus (glyph_mode=natural) hard-fails on
// any token with zero glyphs after internal/evaglyph.NaturalGlyphs
// preprocessing (unicode.IsLetter/IsNumber only), and mechanically
// symbol-only tokens are common in non-prose sources such as assembly
// source (e.g. a bare "_" placeholder identifier). The filter is applied
// uniformly, is decided before any metric is computed on the filtered
// corpus, and is recorded (dropped-token/line counts) in the sidecar
// manifest so it is auditable, not silent.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"
	"unicode"
)

type manifest struct {
	InputPath           string   `json:"input_path"`
	OutputPath          string   `json:"output_path"`
	InputLineCount      int      `json:"input_line_count"`
	OutputLineCount     int      `json:"output_line_count"`
	InputTokenCount     int      `json:"input_token_count"`
	OutputTokenCount    int      `json:"output_token_count"`
	DroppedTokenCount   int      `json:"dropped_token_count"`
	DroppedLineCount    int      `json:"dropped_line_count"`
	SampleDroppedTokens []string `json:"sample_dropped_tokens"`
}

func hasLetterOrDigit(s string) bool {
	for _, r := range s {
		if unicode.IsLetter(r) || unicode.IsNumber(r) {
			return true
		}
	}
	return false
}

func main() {
	os.Exit(run())
}

func run() int {
	input := flag.String("input", "", "raw plain-text corpus (required)")
	output := flag.String("output", "", "filtered output path (required)")
	manifestPath := flag.String("manifest", "", "manifest JSON path (default: <output>.filter.json)")
	flag.Parse()
	if *input == "" || *output == "" {
		fmt.Fprintln(os.Stderr, "usage: generic-glyph-filter -input FILE -output FILE [-manifest FILE]")
		return 2
	}
	if *manifestPath == "" {
		*manifestPath = *output + ".filter.json"
	}
	raw, err := os.ReadFile(*input)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		return 1
	}
	lines := strings.Split(string(raw), "\n")
	if len(lines) > 0 && lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}
	m := manifest{InputPath: *input, OutputPath: *output, InputLineCount: len(lines)}
	var outLines []string
	for _, line := range lines {
		fields := strings.Fields(line)
		var kept []string
		for _, tok := range fields {
			m.InputTokenCount++
			if hasLetterOrDigit(tok) {
				kept = append(kept, tok)
				m.OutputTokenCount++
			} else {
				m.DroppedTokenCount++
				if len(m.SampleDroppedTokens) < 20 {
					m.SampleDroppedTokens = append(m.SampleDroppedTokens, tok)
				}
			}
		}
		if len(kept) == 0 {
			if len(fields) > 0 {
				m.DroppedLineCount++
			}
			continue
		}
		outLines = append(outLines, strings.Join(kept, " "))
	}
	m.OutputLineCount = len(outLines)
	if err := os.WriteFile(*output, []byte(strings.Join(outLines, "\n")+"\n"), 0644); err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		return 1
	}
	data, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		return 1
	}
	data = append(data, '\n')
	if err := os.WriteFile(*manifestPath, data, 0644); err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		return 1
	}
	fmt.Printf("Wrote %s and %s: %d/%d tokens dropped, %d/%d lines dropped\n",
		*output, *manifestPath, m.DroppedTokenCount, m.InputTokenCount, m.DroppedLineCount, m.InputLineCount)
	return 0
}
