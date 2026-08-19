package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"zcore.dev/voinich/internal/corpusprep"
)

func main() {
	os.Exit(run(os.Args[1:]))
}

func run(args []string) int {
	if len(args) == 0 {
		usage()
		return 2
	}
	switch args[0] {
	case "prepare":
		return runPrepare(args[1:])
	case "inspect":
		return runInspect(args[1:])
	case "check":
		return runCheck(args[1:])
	case "-h", "-help", "--help", "help":
		usage()
		return 0
	default:
		if strings.HasPrefix(args[0], "-") {
			return runPrepare(args)
		}
		fmt.Fprintf(os.Stderr, "Error: unknown subcommand %q\n\n", args[0])
		usage()
		return 2
	}
}

func usage() {
	fmt.Fprint(os.Stderr, `codex_prepare: canonical generic corpus preparation and validation

Usage:
  codex_prepare prepare -input FILE [-output prepared.txt] [-encoding utf-8|windows-1251|koi8-r|cp866|auto] [-case preserve|lower] [-line-policy preserve|reflow] [-strip-leading-lines N] [-strip-trailing-lines N] [-drop-empty-lines]
  codex_prepare inspect -input FILE [same options as prepare] [-json]
  codex_prepare check   -input prepared.txt [-json]

prepare writes the canonical corpus and a sibling manifest:
  prepared.txt.prepare.json
`)
}

func runPrepare(args []string) int {
	fs := flag.NewFlagSet("prepare", flag.ContinueOnError)
	input := fs.String("input", "", "raw input text")
	output := fs.String("output", "prepared.txt", "canonical corpus output")
	encoding := fs.String("encoding", corpusprep.EncodingUTF8, "input encoding")
	casePolicy := fs.String("case", corpusprep.CaseLower, "case policy")
	linePolicy := fs.String("line-policy", corpusprep.LinePreserve, "line policy")
	stripLeading := fs.Int("strip-leading-lines", 0, "strip leading raw lines")
	stripTrailing := fs.Int("strip-trailing-lines", 0, "strip trailing raw lines")
	dropEmpty := fs.Bool("drop-empty-lines", true, "drop empty lines")
	if err := fs.Parse(args); err != nil {
		return 2
	}
	if *input == "" {
		fmt.Fprintln(os.Stderr, "Error: -input is required")
		return 2
	}
	inputBytes, err := os.ReadFile(*input)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		return 1
	}
	commit, _ := gitCommit(".")
	result, manifest, err := corpusprep.Prepare(inputBytes, corpusprep.Options{
		Encoding:           *encoding,
		CasePolicy:         *casePolicy,
		LinePolicy:         *linePolicy,
		DropEmptyLines:     *dropEmpty,
		StripLeadingLines:  *stripLeading,
		StripTrailingLines: *stripTrailing,
	}, commit, *input, *output)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		return 1
	}
	if err := os.WriteFile(*output, result.Text, 0644); err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		return 1
	}
	manifest.OutputPath = *output
	manifestPath := *output + ".prepare.json"
	data, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		return 1
	}
	data = append(data, '\n')
	if err := os.WriteFile(manifestPath, data, 0644); err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		return 1
	}
	fmt.Printf("Wrote %s\nWrote %s\nInput encoding: %s\nOutput tokens: %d\nLines: %d\n", filepath.Clean(*output), filepath.Clean(manifestPath), manifest.InputEncoding, manifest.OutputTokenCount, manifest.LineCount)
	return 0
}

func runInspect(args []string) int {
	fs := flag.NewFlagSet("inspect", flag.ContinueOnError)
	input := fs.String("input", "", "raw input text")
	encoding := fs.String("encoding", corpusprep.EncodingUTF8, "input encoding")
	casePolicy := fs.String("case", corpusprep.CaseLower, "case policy")
	linePolicy := fs.String("line-policy", corpusprep.LinePreserve, "line policy")
	stripLeading := fs.Int("strip-leading-lines", 0, "strip leading raw lines")
	stripTrailing := fs.Int("strip-trailing-lines", 0, "strip trailing raw lines")
	dropEmpty := fs.Bool("drop-empty-lines", true, "drop empty lines")
	jsonOut := fs.Bool("json", false, "machine-readable JSON output")
	if err := fs.Parse(args); err != nil {
		return 2
	}
	if *input == "" {
		fmt.Fprintln(os.Stderr, "Error: -input is required")
		return 2
	}
	inputBytes, err := os.ReadFile(*input)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		return 1
	}
	commit, _ := gitCommit(".")
	report, err := corpusprep.Inspect(inputBytes, corpusprep.Options{
		Encoding:           *encoding,
		CasePolicy:         *casePolicy,
		LinePolicy:         *linePolicy,
		DropEmptyLines:     *dropEmpty,
		StripLeadingLines:  *stripLeading,
		StripTrailingLines: *stripTrailing,
	}, commit, *input)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		return 1
	}
	if *jsonOut {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		if err := enc.Encode(report); err != nil {
			fmt.Fprintln(os.Stderr, "Error:", err)
			return 1
		}
		return 0
	}
	fmt.Printf("Encoding: %s\nUTF-8 valid: %v\nTokens in: %d\nTokens out: %d\nLines: %d\nPunctuation: %d\nReplacement chars: %d\nForbidden controls: %d\n", report.Stats.InputEncoding, report.Stats.UTF8Valid, report.Stats.InputTokenCount, report.Stats.OutputTokenCount, report.Stats.LineCount, report.Stats.PunctuationCount, report.Stats.ReplacementCharCount, report.Stats.ForbiddenControlCount)
	for _, note := range report.Notes {
		fmt.Println(note)
	}
	return 0
}

func runCheck(args []string) int {
	fs := flag.NewFlagSet("check", flag.ContinueOnError)
	input := fs.String("input", "", "prepared corpus")
	jsonOut := fs.Bool("json", false, "machine-readable JSON output")
	if err := fs.Parse(args); err != nil {
		return 2
	}
	if *input == "" {
		fmt.Fprintln(os.Stderr, "Error: -input is required")
		return 2
	}
	inputBytes, err := os.ReadFile(*input)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		return 1
	}
	report, err := corpusprep.Check(inputBytes)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		return 1
	}
	if *jsonOut {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		if err := enc.Encode(report); err != nil {
			fmt.Fprintln(os.Stderr, "Error:", err)
			return 1
		}
	} else {
		fmt.Printf("Valid: %v\nReason: %s\nTokens: %d\nLines: %d\n", report.Valid, report.Reason, report.Stats.OutputTokenCount, report.Stats.LineCount)
	}
	if !report.Valid {
		return 1
	}
	return 0
}

func gitCommit(repoPath string) (string, bool) {
	out, err := exec.Command("git", "-C", repoPath, "rev-parse", "HEAD").Output()
	if err != nil {
		return "", false
	}
	return strings.TrimSpace(string(out)), false
}
