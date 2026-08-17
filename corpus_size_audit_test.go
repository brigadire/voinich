package main

import (
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// TestNoVoynichCorpusSizeInAlgorithmicSources prevents historical corpus
// sizes from becoming executable constants. Frozen experiment artifacts,
// documentation, task specifications, and tests may legitimately describe
// concrete corpus sizes, so this guard deliberately scans production source
// and orchestration files only.
func TestNoVoynichCorpusSizeInAlgorithmicSources(t *testing.T) {
	historicalSize := regexp.MustCompile(`\b(?:39026|39_026|38887|38_887)\b`)
	extensions := map[string]bool{
		".c": true, ".go": true, ".json": true, ".sh": true,
		".yaml": true, ".yml": true, ".j2": true,
	}
	skipDirs := map[string]bool{
		".git": true, "data": true, "data_work": true, "experiments": true,
		"profiles": true, "tasks": true, "workdir": true,
	}

	err := filepath.WalkDir(".", func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() {
			if path != "." && (skipDirs[entry.Name()] || strings.HasPrefix(entry.Name(), "workdir.")) {
				return filepath.SkipDir
			}
			return nil
		}
		if strings.HasSuffix(path, "_test.go") || !extensions[filepath.Ext(path)] {
			return nil
		}
		contents, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		if match := historicalSize.Find(contents); match != nil {
			t.Errorf("%s contains historical corpus size %q in algorithmic source", path, match)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}
