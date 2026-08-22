package main

import (
	"encoding/json"
	"os"
	"path/filepath"

	"zcore.dev/voinich/internal/tokenrepetition"
)

func writeManifest(outDir string, voynich tokenrepetition.Corpus) {
	m := map[string]any{
		"task":            "Task60",
		"schema_version":  1,
		"git_commit":      gitCommit(),
		"dirty":           gitDirty(),
		"corpus_path":     voynich.Path,
		"corpus_sha256":   voynich.SHA256,
		"tokens":          len(voynich.Tokens),
		"ivtff_path":      ivtffPath,
		"seed":            baseSeed,
		"null_permutations_exact": nullPermutations,
		"null_permutations_near":  nearNullPermutations,
		"matched_null_draws":      matchedNullDraws,
		"rank_tolerance":           rankTolerance,
		"label_subsamples":         labelSubsamples,
		"min_chain_length":         minChainLength,
		"parser":          "internal/evaglyph (shared with Task58/59)",
		"edit_distance_definition": "glyph-level Levenshtein, adjacent pairs only (O(number of transitions))",
		"position_class_definition": "changed/inserted/deleted glyph position, divided by the longer token's length, bucketed at thirds",
		"adjacency_definition": "pairs crossing a natural line boundary are excluded, matching Task58's own adjacency convention",
		"task58_source": "experiments/rozanova-temerev-v1/comparison.tsv",
		"task59_source": "experiments/glyph-position-v1/POSITIONAL_SPECIALIZATION_COMPARISON.tsv",
	}
	b, _ := json.MarshalIndent(m, "", "  ")
	os.WriteFile(filepath.Join(outDir, "manifest.json"), append(b, '\n'), 0o644)
}
