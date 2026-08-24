// ivtff-x7-extract derives a plain, whitespace-delimited, one-line-per-locus
// token stream from a raw IVTFF file using the same
// metadatavalidation.NormalizeForAlignment logic already used (and audited
// against the real ivtt -x7 output) to strictly align ZL3b-n.txt in
// internal/fingerprintv2. It exists so a second, independently produced
// IVTFF transcription (e.g. data/IT2a-n.txt) can be turned into a
// Fingerprint v2 primary-corpus token stream without requiring the external
// ivtt binary, for Task79c's cross-transcription stability gate.
package main

import (
	"bufio"
	"crypto/sha256"
	"flag"
	"fmt"
	"os"
	"strings"

	"zcore.dev/voinich/internal/metadatavalidation"
)

func main() {
	input := flag.String("input", "", "raw IVTFF input (required)")
	output := flag.String("output", "", "plain token-stream output (required)")
	flag.Parse()
	if *input == "" || *output == "" {
		fmt.Fprintln(os.Stderr, "usage: ivtff-x7-extract -input data/IT2a-n.txt -output data_work/IT2a-x7.txt")
		os.Exit(2)
	}
	if err := run(*input, *output); err != nil {
		fmt.Fprintln(os.Stderr, "ivtff-x7-extract:", err)
		os.Exit(1)
	}
}

func run(input, output string) error {
	doc, err := metadatavalidation.ParseIVTFF(input)
	if err != nil {
		return fmt.Errorf("parse IVTFF %q: %w", input, err)
	}
	f, err := os.Create(output)
	if err != nil {
		return fmt.Errorf("create %q: %w", output, err)
	}
	defer f.Close()
	w := bufio.NewWriter(f)
	lines, tokens := 0, 0
	for _, l := range doc.Loci {
		text := strings.TrimSpace(l.AlignmentText)
		if text == "" {
			continue
		}
		fields := strings.Fields(text)
		tokens += len(fields)
		lines++
		if _, err := fmt.Fprintln(w, strings.Join(fields, " ")); err != nil {
			return err
		}
	}
	if err := w.Flush(); err != nil {
		return err
	}
	raw, err := os.ReadFile(input)
	if err != nil {
		return err
	}
	rawSHA := fmt.Sprintf("%x", sha256.Sum256(raw))
	out, err := os.ReadFile(output)
	if err != nil {
		return err
	}
	outSHA := fmt.Sprintf("%x", sha256.Sum256(out))
	fmt.Printf("input=%s input_sha256=%s pages=%d loci=%d skipped_loci=%d\n", input, rawSHA, doc.Pages, len(doc.Loci), doc.SkippedLoci)
	fmt.Printf("output=%s output_sha256=%s lines=%d tokens=%d\n", output, outSHA, lines, tokens)
	return nil
}
