// Command codex_orientation creates deterministic orientation variants of a
// canonical corpus. It never runs scientific analysis.
package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"zcore.dev/voinich/internal/orientation"
)

func main() {
	os.Exit(run(os.Args[1:]))
}

func run(args []string) int {
	fs := flag.NewFlagSet("codex_orientation", flag.ContinueOnError)
	inputPath := fs.String("input", "", "canonical corpus input")
	outputPath := fs.String("output", "", "transformed canonical corpus output")
	mode := fs.String("mode", "", "TOKEN_REVERSE|GLYPH_REVERSE|FULL_REVERSE")
	force := fs.Bool("force", false, "overwrite existing output and manifest")
	if err := fs.Parse(args); err != nil {
		return 2
	}
	if *inputPath == "" || *outputPath == "" || *mode == "" {
		fmt.Fprintln(os.Stderr, "Usage: codex_orientation -input FILE -output FILE -mode TOKEN_REVERSE|GLYPH_REVERSE|FULL_REVERSE [-force]")
		return 2
	}
	if !orientation.ValidMode(*mode) {
		fmt.Fprintf(os.Stderr, "Error: unsupported orientation mode %q\n", *mode)
		return 2
	}
	inputAbs, err := filepath.Abs(*inputPath)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error: resolve input:", err)
		return 1
	}
	outputAbs, err := filepath.Abs(*outputPath)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error: resolve output:", err)
		return 1
	}
	if inputAbs == outputAbs {
		fmt.Fprintln(os.Stderr, "Error: output must not overwrite input")
		return 1
	}
	manifestPath := outputAbs + ".transform.json"
	if !*force {
		for _, path := range []string{outputAbs, manifestPath} {
			if _, err := os.Stat(path); err == nil {
				fmt.Fprintf(os.Stderr, "Error: %s already exists (use -force to overwrite)\n", path)
				return 1
			} else if !os.IsNotExist(err) {
				fmt.Fprintln(os.Stderr, "Error:", err)
				return 1
			}
		}
	}
	input, err := os.ReadFile(inputAbs)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		return 1
	}
	output, before, after, err := orientation.Transform(input, *mode)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		return 1
	}
	manifest, err := orientation.MarshalManifest(orientation.NewManifest(*mode, *inputPath, *outputPath, input, output, before, after))
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		return 1
	}
	if err := writePair(outputAbs, output, manifestPath, manifest); err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		return 1
	}
	fmt.Printf("Wrote %s\nWrote %s\nTokens: %d\nLines: %d\n", filepath.Clean(*outputPath), filepath.Clean(*outputPath)+".transform.json", after.Tokens, after.Lines)
	return 0
}

func writePair(outputPath string, output []byte, manifestPath string, manifest []byte) error {
	if err := os.MkdirAll(filepath.Dir(outputPath), 0755); err != nil {
		return err
	}
	outputTmp, err := os.CreateTemp(filepath.Dir(outputPath), "."+filepath.Base(outputPath)+".tmp-*")
	if err != nil {
		return err
	}
	defer os.Remove(outputTmp.Name())
	if _, err := outputTmp.Write(output); err != nil {
		outputTmp.Close()
		return err
	}
	if err := outputTmp.Chmod(0644); err != nil {
		outputTmp.Close()
		return err
	}
	if err := outputTmp.Close(); err != nil {
		return err
	}
	manifestTmp, err := os.CreateTemp(filepath.Dir(manifestPath), "."+filepath.Base(manifestPath)+".tmp-*")
	if err != nil {
		return err
	}
	defer os.Remove(manifestTmp.Name())
	if _, err := manifestTmp.Write(manifest); err != nil {
		manifestTmp.Close()
		return err
	}
	if err := manifestTmp.Chmod(0644); err != nil {
		manifestTmp.Close()
		return err
	}
	if err := manifestTmp.Close(); err != nil {
		return err
	}
	if err := os.Rename(outputTmp.Name(), outputPath); err != nil {
		return err
	}
	if err := os.Rename(manifestTmp.Name(), manifestPath); err != nil {
		return err
	}
	return nil
}
