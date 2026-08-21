package corpustransform

import (
	"fmt"
	"os"
	"path/filepath"
)

// RunResult is what one single transform invocation produced, for the CLI
// to print a sanity report from (task46 section 15).
type RunResult struct {
	Manifest            Manifest
	TranspositionSanity *TranspositionSanity
	HomophonicSanity    *HomophonicSanity
}

// TranspositionRequest is everything one deterministic transposition run
// needs.
type TranspositionRequest struct {
	CorpusPath string
	OutputPath string
	GitCommit  string
	LinePolicy string
	Params     TranspositionParams
}

// RunTransposition reads CorpusPath, transposes it, and writes OutputPath
// plus its <output>.transform.json manifest. It never touches the input
// file.
func RunTransposition(req TranspositionRequest) (RunResult, error) {
	if req.LinePolicy == "" {
		req.LinePolicy = LinePolicyPreserve
	}
	input, err := ReadCorpus(req.CorpusPath)
	if err != nil {
		return RunResult{}, fmt.Errorf("reading corpus: %w", err)
	}
	outputTokens, err := Transpose(input.Tokens, req.Params)
	if err != nil {
		return RunResult{}, err
	}
	outputBytes, err := WriteCorpus(outputTokens, input.LineLengths, req.LinePolicy)
	if err != nil {
		return RunResult{}, fmt.Errorf("serializing output: %w", err)
	}
	if err := writeNewFile(req.OutputPath, outputBytes); err != nil {
		return RunResult{}, err
	}
	manifest := NewTranspositionManifest(req.GitCommit, req.CorpusPath, input, req.OutputPath, outputBytes, outputTokens, req.Params, req.LinePolicy)
	if err := writeManifest(req.OutputPath, manifest); err != nil {
		return RunResult{}, err
	}
	sanity := NewTranspositionSanity(input.Tokens, outputTokens, manifest.OutputSHA256)
	return RunResult{Manifest: manifest, TranspositionSanity: &sanity}, nil
}

// HomophonicRequest is everything one deterministic homophonic run needs.
type HomophonicRequest struct {
	CorpusPath string
	OutputPath string
	GitCommit  string
	LinePolicy string
	Params     HomophonicParams
}

// RunHomophonic reads CorpusPath, applies homophonic substitution, and
// writes OutputPath plus its <output>.transform.json manifest and
// <output>.mapping.tsv audit file. It never touches the input file.
func RunHomophonic(req HomophonicRequest) (RunResult, error) {
	if req.LinePolicy == "" {
		req.LinePolicy = LinePolicyPreserve
	}
	input, err := ReadCorpus(req.CorpusPath)
	if err != nil {
		return RunResult{}, fmt.Errorf("reading corpus: %w", err)
	}
	mapping, err := BuildMapping(input.Tokens, req.Params)
	if err != nil {
		return RunResult{}, err
	}
	outputTokens := Encode(input.Tokens, mapping, req.Params.Seed)
	outputBytes, err := WriteCorpus(outputTokens, input.LineLengths, req.LinePolicy)
	if err != nil {
		return RunResult{}, fmt.Errorf("serializing output: %w", err)
	}
	if err := writeNewFile(req.OutputPath, outputBytes); err != nil {
		return RunResult{}, err
	}
	mappingBytes := MarshalMappingTSV(mapping)
	mappingPath := req.OutputPath + ".mapping.tsv"
	if err := writeNewFile(mappingPath, mappingBytes); err != nil {
		return RunResult{}, err
	}
	allocationBytes := MarshalAllocationTSV(mapping)
	allocationPath := req.OutputPath + ".homophone_allocation.tsv"
	if err := writeNewFile(allocationPath, allocationBytes); err != nil {
		return RunResult{}, err
	}
	mappingSHA256 := ShaBytes(mappingBytes)
	allocationSHA256 := ShaBytes(allocationBytes)
	manifest := NewHomophonicManifest(req.GitCommit, req.CorpusPath, input, req.OutputPath, outputBytes, outputTokens, req.Params, req.LinePolicy, mappingSHA256, allocationSHA256)
	if err := writeManifest(req.OutputPath, manifest); err != nil {
		return RunResult{}, err
	}
	sanity := NewHomophonicSanity(input.Tokens, outputTokens, mapping, manifest.OutputSHA256)
	return RunResult{Manifest: manifest, HomophonicSanity: &sanity}, nil
}

func writeNewFile(path string, data []byte) error {
	if dir := filepath.Dir(path); dir != "." {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return fmt.Errorf("creating output directory %s: %w", dir, err)
		}
	}
	if err := os.WriteFile(path, data, 0o644); err != nil {
		return fmt.Errorf("writing %s: %w", path, err)
	}
	return nil
}

func writeManifest(outputPath string, manifest Manifest) error {
	data, err := MarshalManifest(manifest)
	if err != nil {
		return fmt.Errorf("marshaling manifest: %w", err)
	}
	return writeNewFile(outputPath+".transform.json", data)
}
