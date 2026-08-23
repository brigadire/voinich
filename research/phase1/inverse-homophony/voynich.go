package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"zcore.dev/voinich/internal/inversehomophony"
)

type frozenManifestStub struct {
	GatePass bool `json:"gate_pass"`
	Config   struct {
		Threshold float64 `json:"Threshold"`
	} `json:"config"`
}

// checkMethodFrozen enforces task57 section 21/24/36-test-10: Voynich may
// only be analyzed once METHOD_FROZEN exists AND the manifest it was
// written alongside records a passed validation gate. Returns the frozen
// threshold on success.
func checkMethodFrozen(syntheticValidationDir string) (float64, error) {
	if _, err := os.Stat(filepath.Join(syntheticValidationDir, "METHOD_FROZEN")); err != nil {
		return 0, fmt.Errorf("METHOD_FROZEN not found in %s: Phase A validation gate has not passed - Voynich must not be analyzed (task57 section 21)", syntheticValidationDir)
	}
	b, err := os.ReadFile(filepath.Join(syntheticValidationDir, "manifest.json"))
	if err != nil {
		return 0, fmt.Errorf("reading manifest.json: %w", err)
	}
	var m frozenManifestStub
	if err := json.Unmarshal(b, &m); err != nil {
		return 0, fmt.Errorf("parsing manifest.json: %w", err)
	}
	if !m.GatePass {
		return 0, fmt.Errorf("manifest.json records gate_pass=false: Voynich must not be analyzed (task57 section 21)")
	}
	return m.Config.Threshold, nil
}

func runVoynich(args []string) int {
	fs := newFlagSet("voynich")
	corpus := fs.String("corpus", "data_work/ZL3b-x7.canonical.txt", "canonical Voynich corpus path (task57 section 23)")
	outDir := fs.String("out-dir", "experiments/inverse-homophony/voynich-v1", "output directory (task57 section 34)")
	svDir := fs.String("synthetic-validation-dir", "experiments/inverse-homophony/synthetic-validation", "directory written by 'validate', containing METHOD_FROZEN + manifest.json")
	if err := fs.Parse(args); err != nil {
		return 2
	}

	tau, err := checkMethodFrozen(*svDir)
	if err != nil {
		fmt.Fprintln(os.Stderr, "voynich:", err)
		return 1
	}

	if err := os.MkdirAll(*outDir, 0o755); err != nil {
		fmt.Fprintln(os.Stderr, "voynich:", err)
		return 1
	}

	loaded, err := inversehomophony.LoadCorpus(*corpus)
	if err != nil {
		fmt.Fprintln(os.Stderr, "load corpus:", err)
		return 1
	}
	cfg := inversehomophony.FrozenConfig()
	cfg.Threshold = tau

	freq := make(map[string]int, len(loaded.Relabel.ToOpaque))
	for _, t := range loaded.Relabel.Tokens {
		freq[t]++
	}
	features := inversehomophony.BuildFeatures(loaded.Relabel.Tokens, loaded.LineOfToken, cfg)
	pairs := inversehomophony.CandidatePairs(features, cfg)
	partition, events := inversehomophony.Recover(freq, pairs, cfg)

	if err := writeLatentTSV(filepath.Join(*outDir, "voy_to_latent.tsv"), partition, loaded.Relabel.ToOriginal); err != nil {
		fmt.Fprintln(os.Stderr, "write voy_to_latent.tsv:", err)
		return 1
	}
	collapsed := inversehomophony.Collapse(loaded.Relabel.Tokens, partition)
	collapsedText := ""
	for i, t := range collapsed {
		if i > 0 {
			collapsedText += " "
		}
		collapsedText += t
	}
	if err := os.WriteFile(filepath.Join(*outDir, "voynich_collapsed.txt"), []byte(collapsedText+"\n"), 0o644); err != nil {
		fmt.Fprintln(os.Stderr, "write voynich_collapsed.txt:", err)
		return 1
	}
	if err := writeMergeAudit(filepath.Join(*outDir, "merge_audit.tsv"), events); err != nil {
		fmt.Fprintln(os.Stderr, "write merge_audit.tsv:", err)
		return 1
	}

	classes := map[string]bool{}
	for _, c := range partition {
		classes[c] = true
	}
	fmt.Printf("Voynich corpus SHA256: %s\n", loaded.SHA256)
	fmt.Printf("cipher types: %d, recovered classes: %d, tau=%.4f\n", len(features), len(classes), cfg.Threshold)
	fmt.Println("This is the one preregistered recovery run (task57 section 24) - no re-search after inspection.")
	fmt.Println("Before/after structural comparison (task57 section 25) and the matched-null sweep (section 27) are a separate,")
	fmt.Println("follow-up step using the real Stage 23-28 CLIs via pipeline-orchestrate -generic-corpus on voynich_collapsed.txt,")
	fmt.Println("per INVERSE_HOMOPHONY_DESIGN.md section 0 - not run automatically here given their multi-hour runtime.")
	return 0
}
