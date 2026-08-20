// Command corpus-transform is task46's deterministic historical cipher
// corpus transformer. It is a standalone experiment-input generator: it is
// never invoked by pipeline-orchestrate, is not itself a scientific
// pipeline stage, performs no statistical analysis, and never compares its
// output to the Voynich manuscript (see corpus-transform/README.md).
package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"zcore.dev/voinich/internal/corpustransform"
)

func newFlagSet(name string) *flag.FlagSet {
	fs := flag.NewFlagSet(name, flag.ContinueOnError)
	fs.Usage = usage
	return fs
}

func main() {
	os.Exit(run(os.Args[1:]))
}

func run(args []string) int {
	if len(args) == 0 {
		usage()
		return 2
	}
	switch args[0] {
	case "batch":
		return runBatch(args[1:])
	case "-h", "-help", "--help", "help":
		usage()
		return 0
	default:
		return runSingle(args)
	}
}

func usage() {
	fmt.Fprint(os.Stderr, `corpus-transform: deterministic historical-cipher corpus transformer (task46)

Usage:
  corpus-transform -corpus FILE -method transposition -output FILE -seed N \
      -transposition-width N [-transposition-order natural|keyed] [-rounds N] [-line-policy preserve|reflow]

  corpus-transform -corpus FILE -method homophonic -output FILE -seed N \
      -homophones N [-homophone-selection uniform|weighted] [-homophone-model fixed] [-line-policy preserve|reflow]

  corpus-transform batch -corpus FILE -output-dir DIR [-label NAME] \
      [-transposition-widths 2,4,8,16,32] [-transposition-order natural|keyed] [-rounds N] \
      [-homophone-counts 2,4,8] [-homophone-selection uniform|weighted] [-homophone-model fixed] \
      [-seeds 1,2,3] [-line-policy preserve|reflow]

Every output <path> is written alongside <path>.transform.json; homophonic
runs additionally write <path>.mapping.tsv. See TRANSFORMATION_METHODS.md.
`)
}

// gitCommit shells out to git rather than relying on runtime/debug build
// info: "go run", the invocation shown in every task46 example, does not
// embed VCS stamps (only "go build"/"go install" do), so ReadBuildInfo
// would always report "unknown" for the tool's primary usage.
func gitCommit() string {
	out, err := exec.Command("git", "rev-parse", "HEAD").Output()
	if err != nil {
		return "unknown"
	}
	return strings.TrimSpace(string(out))
}

func runSingle(args []string) int {
	fs := newFlagSet("corpus-transform")
	corpus := fs.String("corpus", "", "input corpus (whitespace-tokenized text; not modified)")
	method := fs.String("method", "", "transposition or homophonic")
	output := fs.String("output", "", "output corpus path")
	seed := fs.Int64("seed", 1, "deterministic random seed")
	linePolicy := fs.String("line-policy", corpustransform.LinePolicyPreserve, "preserve or reflow (see TRANSFORMATION_METHODS.md)")

	width := fs.Int("transposition-width", 0, "rectangular columnar transposition width")
	order := fs.String("transposition-order", corpustransform.OrderNatural, "natural or keyed")
	rounds := fs.Int("rounds", 1, "repeated transposition rounds")

	homophones := fs.Int("homophones", 4, "homophones per plaintext token (H)")
	selection := fs.String("homophone-selection", corpustransform.SelectionUniform, "uniform or weighted")
	model := fs.String("homophone-model", corpustransform.HomophoneModelFixed, "fixed (frequency is backlog, not implemented)")

	if err := fs.Parse(args); err != nil {
		return 2
	}
	if *corpus == "" || *method == "" || *output == "" {
		fmt.Fprintln(os.Stderr, "Error: -corpus, -method, and -output are required")
		return 2
	}

	switch *method {
	case corpustransform.MethodTransposition:
		if *width < 1 {
			fmt.Fprintln(os.Stderr, "Error: -transposition-width is required and must be >= 1")
			return 2
		}
		res, err := corpustransform.RunTransposition(corpustransform.TranspositionRequest{
			CorpusPath: *corpus,
			OutputPath: *output,
			GitCommit:  gitCommit(),
			LinePolicy: *linePolicy,
			Params:     corpustransform.TranspositionParams{Width: *width, Order: *order, Round: *rounds, Seed: *seed},
		})
		if err != nil {
			fmt.Fprintln(os.Stderr, "Error:", err)
			return 1
		}
		fmt.Printf("Wrote %s\nWrote %s.transform.json\n%s", *output, *output, res.TranspositionSanity.String())
		return 0
	case corpustransform.MethodHomophonic:
		res, err := corpustransform.RunHomophonic(corpustransform.HomophonicRequest{
			CorpusPath: *corpus,
			OutputPath: *output,
			GitCommit:  gitCommit(),
			LinePolicy: *linePolicy,
			Params:     corpustransform.HomophonicParams{Model: *model, Homophones: *homophones, Selection: *selection, Seed: *seed},
		})
		if err != nil {
			fmt.Fprintln(os.Stderr, "Error:", err)
			return 1
		}
		fmt.Printf("Wrote %s\nWrote %s.transform.json\nWrote %s.mapping.tsv\n%s", *output, *output, *output, res.HomophonicSanity.String())
		return 0
	default:
		fmt.Fprintf(os.Stderr, "Error: unsupported -method %q (want transposition or homophonic)\n", *method)
		return 2
	}
}

func runBatch(args []string) int {
	fs := newFlagSet("corpus-transform batch")
	corpus := fs.String("corpus", "", "input corpus (whitespace-tokenized text; not modified)")
	outputDir := fs.String("output-dir", "", "directory to write generated corpora + manifests into")
	label := fs.String("label", "", "experiment-id label prefix (default: corpus base filename)")
	linePolicy := fs.String("line-policy", corpustransform.LinePolicyPreserve, "preserve or reflow, applied to every generated corpus")

	transpositionWidths := fs.String("transposition-widths", "", "comma-separated widths, e.g. 2,4,8,16,32 (empty skips transposition)")
	order := fs.String("transposition-order", corpustransform.OrderNatural, "natural or keyed")
	rounds := fs.Int("rounds", 1, "repeated transposition rounds")

	homophoneCounts := fs.String("homophone-counts", "", "comma-separated H values, e.g. 2,4,8 (empty skips homophonic)")
	selection := fs.String("homophone-selection", corpustransform.SelectionUniform, "uniform or weighted")
	model := fs.String("homophone-model", corpustransform.HomophoneModelFixed, "fixed (frequency is backlog, not implemented)")

	seeds := fs.String("seeds", "1", "comma-separated seeds, e.g. 1,2,3")

	if err := fs.Parse(args); err != nil {
		return 2
	}
	if *corpus == "" || *outputDir == "" {
		fmt.Fprintln(os.Stderr, "Error: -corpus and -output-dir are required")
		return 2
	}
	widths, err := parseInts(*transpositionWidths)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error: -transposition-widths:", err)
		return 2
	}
	counts, err := parseInts(*homophoneCounts)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error: -homophone-counts:", err)
		return 2
	}
	seedList, err := parseInt64s(*seeds)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error: -seeds:", err)
		return 2
	}
	if len(widths) == 0 && len(counts) == 0 {
		fmt.Fprintln(os.Stderr, "Error: at least one of -transposition-widths or -homophone-counts is required")
		return 2
	}

	lbl := *label
	if lbl == "" {
		lbl = corpustransform.Label(*corpus)
	}
	commit := gitCommit()
	written := 0
	for _, seed := range seedList {
		for _, w := range widths {
			id := corpustransform.TranspositionExperimentID(lbl, w, *order, seed)
			outputPath := *outputDir + "/" + id + ".txt"
			res, err := corpustransform.RunTransposition(corpustransform.TranspositionRequest{
				CorpusPath: *corpus,
				OutputPath: outputPath,
				GitCommit:  commit,
				LinePolicy: *linePolicy,
				Params:     corpustransform.TranspositionParams{Width: w, Order: *order, Round: *rounds, Seed: seed},
			})
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: %s: %v\n", id, err)
				return 1
			}
			fmt.Printf("Wrote %s (multiset preserved: %v)\n", outputPath, res.TranspositionSanity.MultisetPreserved)
			written++
		}
		for _, h := range counts {
			id := corpustransform.HomophonicExperimentID(lbl, h, *selection, seed)
			outputPath := *outputDir + "/" + id + ".txt"
			res, err := corpustransform.RunHomophonic(corpustransform.HomophonicRequest{
				CorpusPath: *corpus,
				OutputPath: outputPath,
				GitCommit:  commit,
				LinePolicy: *linePolicy,
				Params:     corpustransform.HomophonicParams{Model: *model, Homophones: h, Selection: *selection, Seed: seed},
			})
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: %s: %v\n", id, err)
				return 1
			}
			fmt.Printf("Wrote %s (mapping collisions: %d)\n", outputPath, res.HomophonicSanity.MappingCollisions)
			written++
		}
	}
	fmt.Printf("Wrote %d transformed corpora to %s\n", written, *outputDir)
	return 0
}

func parseInts(csv string) ([]int, error) {
	csv = strings.TrimSpace(csv)
	if csv == "" {
		return nil, nil
	}
	var out []int
	for part := range strings.SplitSeq(csv, ",") {
		v, err := strconv.Atoi(strings.TrimSpace(part))
		if err != nil {
			return nil, fmt.Errorf("invalid integer %q: %w", part, err)
		}
		out = append(out, v)
	}
	return out, nil
}

func parseInt64s(csv string) ([]int64, error) {
	csv = strings.TrimSpace(csv)
	if csv == "" {
		return nil, nil
	}
	var out []int64
	for part := range strings.SplitSeq(csv, ",") {
		v, err := strconv.ParseInt(strings.TrimSpace(part), 10, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid integer %q: %w", part, err)
		}
		out = append(out, v)
	}
	return out, nil
}
