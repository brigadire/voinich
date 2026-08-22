// Command token-repetition-analyze implements task60's independent study
// of exact adjacent repetition, exact runs, near-repetition (glyph-level
// edit-distance-1 adjacency), and illustration-label repetition in the
// canonical Voynich corpus, plus natural-language and existing
// homophonic controls. It is deliberately not a pipeline stage.
package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"zcore.dev/voinich/internal/tokenrepetition"
)

const (
	nullPermutations = 1000
	matchedNullDraws = 200
	rankTolerance    = 25
	labelSubsamples  = 500
	minChainLength   = 3
	maxLociPerToken  = 20
	baseSeed         = int64(20260823)
	ivtffPath        = "data/ZL3b-n.txt"
)

func main() {
	out := flag.String("output-dir", "experiments/token-repetition-v1", "output directory")
	voynichPath := flag.String("corpus", "data_work/ZL3b-x7.canonical.txt", "canonical Voynich corpus")
	flag.Parse()

	if err := run(*out, *voynichPath); err != nil {
		fmt.Fprintln(os.Stderr, "token-repetition-analyze:", err)
		os.Exit(1)
	}
}

func run(outDir, voynichPath string) error {
	if err := os.MkdirAll(outDir, 0o755); err != nil {
		return err
	}
	rng := newRand(baseSeed)

	voynich, err := tokenrepetition.LoadCorpus(voynichPath, "Voynich")
	if err != nil {
		return fmt.Errorf("load voynich: %w", err)
	}
	naturalSpecs := []struct{ name, path string }{
		{"Doyle", "data_test/pg2097-2.txt"},
		{"Longfellow", "data_test/pg30795-mod.txt"},
		{"Astafiev", "data_test/astafiev-1000-culinar-receipts-prepared.txt"},
	}
	corpora := []tokenrepetition.Corpus{voynich}
	for _, s := range naturalSpecs {
		c, err := tokenrepetition.LoadCorpus(s.path, s.name)
		if err != nil {
			return fmt.Errorf("load %s: %w", s.name, err)
		}
		corpora = append(corpora, c)
	}

	rep := newReport(outDir)
	w := openWriters(outDir)
	defer w.closeAll()

	for _, c := range corpora {
		mode := tokenrepetition.GlyphNatural
		if c.Name == "Voynich" {
			mode = tokenrepetition.GlyphVoynich
		}
		if err := analyzeCorpus(c, mode, rng, rep, w); err != nil {
			return fmt.Errorf("analyze %s: %w", c.Name, err)
		}
	}

	if err := runHomophonyDoseResponse(w); err != nil {
		return fmt.Errorf("homophony dose-response: %w", err)
	}
	if err := runGlyphHomophonyNearRepetition(rep, rng); err != nil {
		return fmt.Errorf("glyph homophony near-repetition: %w", err)
	}
	if err := runLabels(voynich, outDir, rng, rep); err != nil {
		return fmt.Errorf("labels: %w", err)
	}
	if err := joinTask58(rep, outDir); err != nil {
		return fmt.Errorf("task58 join: %w", err)
	}
	if err := joinTask59(rep, outDir); err != nil {
		return fmt.Errorf("task59 join: %w", err)
	}

	writeManifest(outDir, voynich)
	rep.writeReport()
	return nil
}

func gitCommit() string {
	out, err := exec.Command("git", "rev-parse", "HEAD").Output()
	if err != nil {
		return "unknown"
	}
	return strings.TrimSpace(string(out))
}
func gitDirty() bool {
	out, err := exec.Command("git", "status", "--porcelain").Output()
	if err != nil {
		return false
	}
	return len(strings.TrimSpace(string(out))) > 0
}

