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
		// shared ./workdir contract.
		if base := filepath.Base(filepath.Dir(path)); base == "codex_prepare" || base == "codex_orientation" {
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
