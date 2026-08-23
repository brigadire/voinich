package main

import (
	"os"
	"path/filepath"
	"testing"

	"zcore.dev/voinich/internal/orientation"
)

func TestGenericManifestLoadsOrientationProvenance(t *testing.T) {
	repo, err := filepath.Abs("../..")
	if err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	source := []byte("a b c\n")
	transformed, before, after, err := orientation.Transform(source, orientation.TokenReverse)
	if err != nil {
		t.Fatal(err)
	}
	corpus := filepath.Join(dir, "corpus.token-reverse.txt")
	if err := os.WriteFile(corpus, transformed, 0644); err != nil {
		t.Fatal(err)
	}
	manifest := orientation.NewManifest(orientation.TokenReverse, filepath.Join(dir, "source.txt"), corpus, source, transformed, before, after)
	data, err := orientation.MarshalManifest(manifest)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(corpus+".transform.json", data, 0644); err != nil {
		t.Fatal(err)
	}
	m, err := buildManifest(repo, "generic", "", corpus, orchestratorOptions{Executor: "process", LocalWorkers: 2}, nil)
	if err != nil {
		t.Fatal(err)
	}
	if m.Transformation == nil {
		t.Fatal("orientation provenance missing")
	}
	if m.Transformation.Type != orientation.Transformation || m.Transformation.Mode != orientation.TokenReverse || m.Transformation.SourceSHA256 != manifest.InputSHA256 || m.Transformation.OutputSHA256 != m.CorpusSHA256 || m.Transformation.ManifestSHA256 == "" {
		t.Fatalf("unexpected transformation provenance: %+v", m.Transformation)
	}
}
