// Command inverse-transposition-search is task54's controlled, structural
// inverse-transposition experiment. It is intentionally outside the main
// scientific pipeline and never accepts an oracle during search.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"zcore.dev/voinich/internal/corpustransform"
	"zcore.dev/voinich/internal/inversetransposition"
)

type manifest struct {
	Schema      int                                    `json:"schema"`
	Objective   string                                 `json:"objective"`
	InputSHA256 string                                 `json:"input_sha256"`
	Candidates  []inversetransposition.ScoredCandidate `json:"candidates"`
}

func main() { os.Exit(run(os.Args[1:])) }
func run(args []string) int {
	if len(args) == 0 || args[0] == "-h" || args[0] == "--help" {
		usage()
		return 0
	}
	if args[0] == "validate" {
		return validate(args[1:])
	}
	fs := flag.NewFlagSet("inverse-transposition-search", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	input := fs.String("input", "", "transformed canonical corpus (the only search input)")
	outdir := fs.String("output-dir", "", "directory for ranked results and candidates")
	widths := fs.String("widths", "2..16", "fixed search widths, e.g. 2,4,8 or 2..16")
	orders := fs.String("orders", "natural,keyed", "pre-registered read orders")
	rounds := fs.Int("rounds", 1, "Task46 transposition rounds")
	seed := fs.Int64("seed", 1, "Task46 seed for keyed order")
	top := fs.Int("top", 10, "number of candidate corpora to write")
	if err := fs.Parse(args); err != nil {
		return 2
	}
	if *input == "" || *outdir == "" {
		fmt.Fprintln(os.Stderr, "-input and -output-dir are required")
		return 2
	}
	ws, err := parseWidths(*widths)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 2
	}
	os_, err := parseOrders(*orders)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 2
	}
	c, err := corpustransform.ReadCorpus(*input)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	var cs []inversetransposition.Candidate
	for _, w := range ws {
		for _, o := range os_ {
			cs = append(cs, inversetransposition.Candidate{Width: w, Order: o, Rounds: *rounds, Seed: *seed})
		}
	}
	ranked, err := inversetransposition.Rank(c.Tokens, cs)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	if err := os.MkdirAll(*outdir, 0755); err != nil {
		return 1
	}
	if *top < 0 {
		*top = 0
	}
	if *top > len(ranked) {
		*top = len(ranked)
	}
	for _, r := range ranked[:*top] {
		tokens, _ := r.Apply(c.Tokens)
		b, _ := corpustransform.WriteCorpus(tokens, c.LineLengths, corpustransform.LinePolicyPreserve)
		if err := os.WriteFile(filepath.Join(*outdir, r.ID()+".txt"), b, 0644); err != nil {
			return 1
		}
	}
	m := manifest{Schema: 1, Objective: inversetransposition.ObjectiveVersion, InputSHA256: c.InputSHA256Hex, Candidates: ranked}
	b, _ := json.MarshalIndent(m, "", "  ")
	b = append(b, '\n')
	if err := os.WriteFile(filepath.Join(*outdir, "search-manifest.json"), b, 0644); err != nil {
		return 1
	}
	if err := writeReport(*outdir, ranked); err != nil {
		return 1
	}
	fmt.Printf("ranked %d candidates; wrote top %d to %s\n", len(ranked), *top, *outdir)
	return 0
}

func validate(args []string) int {
	fs := flag.NewFlagSet("validate", flag.ContinueOnError)
	input := fs.String("input", "", "transformed corpus")
	oracle := fs.String("oracle", "", "original corpus, used only after search")
	candidate := fs.String("candidate", "", "candidate reconstruction")
	if fs.Parse(args) != nil {
		return 2
	}
	if *input == "" || *oracle == "" || *candidate == "" {
		fmt.Fprintln(os.Stderr, "validate requires -input, -oracle and -candidate")
		return 2
	}
	if _, err := corpustransform.ReadCorpus(*input); err != nil {
		fmt.Fprintln(os.Stderr, "reading transformed corpus:", err)
		return 1
	}
	a, e := corpustransform.ReadCorpus(*oracle)
	if e != nil {
		fmt.Fprintln(os.Stderr, e)
		return 1
	}
	b, e := corpustransform.ReadCorpus(*candidate)
	if e != nil {
		fmt.Fprintln(os.Stderr, e)
		return 1
	}
	if len(a.Tokens) == len(b.Tokens) && equal(a.Tokens, b.Tokens) {
		fmt.Println("oracle exact recovery: true")
		return 0
	}
	fmt.Println("oracle exact recovery: false")
	return 1
}
func equal(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
func parseOrders(s string) ([]string, error) {
	var r []string
	for x := range strings.SplitSeq(s, ",") {
		x = strings.TrimSpace(x)
		if x != corpustransform.OrderNatural && x != corpustransform.OrderKeyed {
			return nil, fmt.Errorf("unsupported order %q", x)
		}
		r = append(r, x)
	}
	return r, nil
}
func parseWidths(s string) ([]int, error) {
	s = strings.TrimSpace(s)
	if strings.Contains(s, "..") {
		p := strings.Split(s, "..")
		if len(p) != 2 {
			return nil, fmt.Errorf("invalid widths %q", s)
		}
		a, e := strconv.Atoi(p[0])
		if e != nil {
			return nil, e
		}
		z, e := strconv.Atoi(p[1])
		if e != nil {
			return nil, e
		}
		var r []int
		for i := a; i <= z; i++ {
			r = append(r, i)
		}
		return r, nil
	}
	var r []int
	for x := range strings.SplitSeq(s, ",") {
		i, e := strconv.Atoi(strings.TrimSpace(x))
		if e != nil || i < 1 {
			return nil, fmt.Errorf("invalid width %q", x)
		}
		r = append(r, i)
	}
	return r, nil
}
func writeReport(dir string, r []inversetransposition.ScoredCandidate) error {
	var b strings.Builder
	b.WriteString("# Inverse-transposition search\n\nObjective: `structural-v2`; candidate-set min-max family balancing; no lexical metrics and no oracle.\n\n| rank | candidate | score | transition | relation | sequence-2 | sequence-3 |\n|---:|---|---:|---:|---:|---:|---:|\n")
	for _, x := range r {
		fmt.Fprintf(&b, "| %d | %s | %.6f | %.6f | %.6f | %.6f | %.6f |\n", x.Rank, x.ID(), x.Score, x.Metrics.TransitionConcentration, x.Metrics.RelationSignificance, x.Metrics.SequenceRepetition, x.Metrics.HigherOrderRepetition)
	}
	return os.WriteFile(filepath.Join(dir, "search-report.md"), []byte(b.String()), 0644)
}
func usage() {
	fmt.Fprintln(os.Stderr, "inverse-transposition-search -input TRANSFORMED -output-dir DIR [-widths 2..16] [-orders natural,keyed]\n  validate -input TRANSFORMED -oracle ORIGINAL -candidate CANDIDATE")
}
