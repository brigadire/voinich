package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"zcore.dev/voinich/internal/orientation"
)

func TestRunWritesCorpusAndManifestWithoutChangingInput(t *testing.T) {
	dir := t.TempDir()
	input := filepath.Join(dir, "source.txt")
	output := filepath.Join(dir, "reversed.txt")
	source := []byte("a b c\nde fg h\n")
	if err := os.WriteFile(input, source, 0644); err != nil {
		t.Fatal(err)
	}
	if got := run([]string{"-input", input, "-output", output, "-mode", orientation.TokenReverse}); got != 0 {
		t.Fatalf("exit code %d", got)
	}
	if got, err := os.ReadFile(input); err != nil || string(got) != string(source) {
		t.Fatalf("input changed: %q, %v", got, err)
	}
	if got, err := os.ReadFile(output); err != nil || string(got) != "c b a\nh fg de\n" {
		t.Fatalf("output = %q, %v", got, err)
	}
	data, err := os.ReadFile(output + ".transform.json")
	if err != nil {
		t.Fatal(err)
	}
	var manifest orientation.Manifest
	if err := json.Unmarshal(data, &manifest); err != nil {
		t.Fatal(err)
	}
	if manifest.Mode != orientation.TokenReverse || manifest.InputTokens != manifest.OutputTokens || !manifest.TokenOrderReversed || manifest.LineOrderReversed {
		t.Fatalf("unexpected manifest: %+v", manifest)
	}
	if got := run([]string{"-input", input, "-output", output, "-mode", orientation.TokenReverse}); got != 1 {
		t.Fatalf("existing output exit code %d, want 1", got)
	}
}
