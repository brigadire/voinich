package workdir

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"testing"
)

// Every command in this repository is part of the analysis pipeline. Requiring
// the shared package makes the output-location rule visible to future commands.
func TestAllPipelineCommandsUseWorkdirContract(t *testing.T) {
	_, file, _, _ := runtime.Caller(0)
	root := filepath.Clean(filepath.Join(filepath.Dir(file), "..", ".."))
	// Task70 moved every pipeline-stage/tool main.go under cmd/ and the
	// Task36 orchestrator under pipelines/pipeline-orchestrate/; research/
	// phase1/* (corpus-transform, inverse-transposition-search,
	// inverse-homophony, voynich-validation, and the Task58-67 independent
	// experiments) is deliberately out of scope, same as before the move.
	cmdDir := filepath.Join(root, "cmd")
	entries, err := os.ReadDir(cmdDir)
	if err != nil {
		t.Fatal(err)
	}
	mainFiles := []string{filepath.Join(root, "pipelines", "pipeline-orchestrate", "main.go")}
	for _, entry := range entries {
		if entry.IsDir() {
			path := filepath.Join(cmdDir, entry.Name(), "main.go")
			if _, err := os.Stat(path); err == nil {
				mainFiles = append(mainFiles, path)
			}
		}
	}
	for _, path := range mainFiles {
		// codex_prepare and codex_orientation are experiment-input
		// generators, not pipeline stages: their outputs are new corpora
		// the caller places wherever it likes (e.g. alongside
		// data_test/*.txt), never generated analysis artifacts under the
		// shared ./workdir contract. tei-abbr-extract, generic-glyph-filter
		// and ivtff-x7-extract are the same class of tool, added for
		// Task79c: each turns third-party/raw bytes into a prepared corpus
		// file at a caller-chosen path (e.g. under data_test/), not an
		// analysis result under ./workdir. task82b-run/task82b-aggregate are
		// the same class again, added for Task82b: independent research
		// entry points with an explicit -out flag, no frozen Task49 stage
		// contract, and no relation to the numbered Stage1-28 pipeline.
		// numeric-test is likewise an exploratory research CLI; its frozen
		// deliverables are under research/numeric rather than workdir.
		if base := filepath.Base(filepath.Dir(path)); base == "codex_prepare" || base == "codex_orientation" ||
			base == "tei-abbr-extract" || base == "generic-glyph-filter" || base == "ivtff-x7-extract" ||
			base == "task82b-run" || base == "task82b-aggregate" || base == "fingerprint-v2-verify" ||
			base == "g1v2-executor" || base == "numeric-test" || base == "vm-structure" {
			continue
		}
		parsed, err := parser.ParseFile(token.NewFileSet(), path, nil, parser.ImportsOnly)
		if err != nil {
			t.Errorf("parse %s: %v", path, err)
			continue
		}
		found := false
		for _, spec := range parsed.Decls {
			decl, ok := spec.(*ast.GenDecl)
			if !ok {
				continue
			}
			for _, item := range decl.Specs {
				imp, ok := item.(*ast.ImportSpec)
				if !ok {
					continue
				}
				value, _ := strconv.Unquote(imp.Path.Value)
				if value == "zcore.dev/voinich/internal/workdir" {
					found = true
				}
			}
		}
		if !found {
			t.Errorf("pipeline command %s must use internal/workdir for generated outputs", path)
		}
	}
}
