package main

import (
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"

	"zcore.dev/voinich/internal/inversehomophony"
)

func runRecover(args []string) int {
	fs := newFlagSet("recover")
	corpus := fs.String("corpus", "", "ciphertext corpus path (required)")
	output := fs.String("output", "", "output prefix (required): writes <output>.latent.tsv, <output>.collapsed.txt, <output>.merge_audit.tsv")
	tau := fs.Float64("tau", -1, "override frozen threshold (for experimentation only; validate/voynich never use this)")
	if err := fs.Parse(args); err != nil {
		return 2
	}
	if *corpus == "" || *output == "" {
		fmt.Fprintln(os.Stderr, "recover: -corpus and -output are required")
		return 2
	}

	loaded, err := inversehomophony.LoadCorpus(*corpus)
	if err != nil {
		fmt.Fprintln(os.Stderr, "load corpus:", err)
		return 1
	}
	cfg := inversehomophony.FrozenConfig()
	if *tau >= 0 {
		cfg.Threshold = *tau
	} else {
		fmt.Fprintln(os.Stderr, "recover: -tau not given and this standalone command has no frozen threshold source; pass -tau explicitly (validate/voynich set it from the frozen manifest)")
		return 2
	}

	freq := make(map[string]int, len(loaded.Relabel.ToOpaque))
	for _, t := range loaded.Relabel.Tokens {
		freq[t]++
	}
	features := inversehomophony.BuildFeatures(loaded.Relabel.Tokens, loaded.LineOfToken, cfg)
	pairs := inversehomophony.CandidatePairs(features, cfg)
	partition, events := inversehomophony.Recover(freq, pairs, cfg)

	if err := writeLatentTSV(*output+".latent.tsv", partition, loaded.Relabel.ToOriginal); err != nil {
		fmt.Fprintln(os.Stderr, "write latent tsv:", err)
		return 1
	}
	collapsed := inversehomophony.Collapse(loaded.Relabel.Tokens, partition)
	if err := os.WriteFile(*output+".collapsed.txt", []byte(strings.Join(collapsed, " ")+"\n"), 0o644); err != nil {
		fmt.Fprintln(os.Stderr, "write collapsed corpus:", err)
		return 1
	}
	if err := writeMergeAudit(*output+".merge_audit.tsv", events); err != nil {
		fmt.Fprintln(os.Stderr, "write merge audit:", err)
		return 1
	}

	classes := map[string]bool{}
	for _, c := range partition {
		classes[c] = true
	}
	fmt.Printf("cipher types: %d, recovered classes: %d, tau=%.4f\n", len(features), len(classes), cfg.Threshold)
	return 0
}

func writeLatentTSV(path string, partition inversehomophony.Partition, toOriginal map[string]string) error {
	tokens := make([]string, 0, len(partition))
	for t := range partition {
		tokens = append(tokens, t)
	}
	sort.Strings(tokens)
	var b strings.Builder
	b.WriteString("ciphertext_token\toriginal_token\tlatent_class\n")
	for _, t := range tokens {
		fmt.Fprintf(&b, "%s\t%s\t%s\n", t, toOriginal[t], partition[t])
	}
	return os.WriteFile(path, []byte(b.String()), 0o644)
}

func writeMergeAudit(path string, events []inversehomophony.MergeEvent) error {
	var b strings.Builder
	b.WriteString("a\tb\tscore\taccepted\treason\tclass_size_after\tclass_fraction_after\n")
	for _, e := range events {
		b.WriteString(e.A)
		b.WriteByte('\t')
		b.WriteString(e.B)
		b.WriteByte('\t')
		b.WriteString(strconv.FormatFloat(e.Score, 'g', -1, 64))
		b.WriteByte('\t')
		b.WriteString(strconv.FormatBool(e.Accepted))
		b.WriteByte('\t')
		b.WriteString(e.Reason)
		b.WriteByte('\t')
		b.WriteString(strconv.Itoa(e.ClassSizeAfter))
		b.WriteByte('\t')
		b.WriteString(strconv.FormatFloat(e.ClassFractionAfter, 'g', -1, 64))
		b.WriteByte('\n')
	}
	return os.WriteFile(path, []byte(b.String()), 0o644)
}
