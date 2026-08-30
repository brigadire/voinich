package main

import (
	"encoding/csv"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"zcore.dev/voinich/internal/structurecatalog"
)

func main() { os.Exit(run(os.Args[1:], os.Stdout, os.Stderr)) }
func run(args []string, out, errOut io.Writer) int {
	if len(args) == 0 {
		fmt.Fprintln(errOut, "usage: vm-structure generate [flags] | query [--catalog DIR] TYPE VALUE")
		return 2
	}
	switch args[0] {
	case "generate":
		return generate(args[1:], out, errOut)
	case "query":
		return query(args[1:], out, errOut)
	default:
		fmt.Fprintln(errOut, "unknown command:", args[0])
		return 2
	}
}
func generate(args []string, out, errOut io.Writer) int {
	fs := flag.NewFlagSet("generate", flag.ContinueOnError)
	fs.SetOutput(errOut)
	c := structurecatalog.Config{}
	fs.StringVar(&c.CorpusPath, "corpus", "data_work/ZL3b-x7.canonical.txt", "primary canonical corpus")
	fs.StringVar(&c.IVTFFPath, "ivtff", "data/ZL3b-n.txt", "primary IVTFF metadata source")
	fs.StringVar(&c.IT2aPath, "it2a", "data_work/IT2a-x7.canonical.txt", "independent replication corpus")
	fs.StringVar(&c.IT2aIVTFFPath, "it2a-ivtff", "", "optional independently aligned IT2a metadata source")
	fs.StringVar(&c.OutputDir, "output", "research/structure_catalog", "catalog output directory")
	fs.IntVar(&c.MinFrequency, "min-frequency", structurecatalog.DefaultMinFreq, "fixed frequent-token threshold")
	if e := fs.Parse(args); e != nil {
		return 2
	}
	cat, e := structurecatalog.Run(c)
	if e != nil {
		fmt.Fprintln(errOut, "Error:", e)
		return 1
	}
	fmt.Fprintf(out, "wrote %d rules to %s\n", len(cat.Rules), c.OutputDir)
	return 0
}

func query(args []string, out, errOut io.Writer) int {
	fs := flag.NewFlagSet("query", flag.ContinueOnError)
	fs.SetOutput(errOut)
	dir := fs.String("catalog", "research/structure_catalog", "frozen catalog directory")
	if e := fs.Parse(args); e != nil {
		return 2
	}
	a := fs.Args()
	if len(a) != 2 {
		fmt.Fprintln(errOut, "usage: vm-structure query [--catalog DIR] glyph|token|follows|precedes|cooccurs|absent-with|position|section VALUE")
		return 2
	}
	kind, value := a[0], a[1]
	type qspec struct {
		file  string
		cols  []string
		match string
	}
	specs := map[string][]qspec{
		"glyph":       {{"GLYPH_INVENTORY.tsv", nil, "glyph"}, {"GLYPH_POSITION_RULES.tsv", nil, "glyph"}, {"GLYPH_BIGRAM_RULES.tsv", nil, "left_glyph"}, {"GLYPH_BIGRAM_RULES.tsv", nil, "right_glyph"}},
		"token":       {{"TOKEN_CATALOG.tsv", nil, "token"}, {"TOKEN_POSITION_RULES.tsv", nil, "token"}},
		"follows":     {{"TOKEN_TRANSITIONS_OBSERVED.tsv", nil, "token_A"}, {"TOKEN_TRANSITIONS_UNOBSERVED.tsv", nil, "token_A"}},
		"precedes":    {{"TOKEN_TRANSITIONS_OBSERVED.tsv", nil, "token_B"}, {"TOKEN_TRANSITIONS_UNOBSERVED.tsv", nil, "token_B"}},
		"cooccurs":    {{"TOKEN_LINE_COOCCURRENCE.tsv", nil, "token_A"}, {"TOKEN_LINE_COOCCURRENCE.tsv", nil, "token_B"}},
		"absent-with": {{"TOKEN_LINE_NONCOOCCURRENCE.tsv", nil, "token_A"}, {"TOKEN_LINE_NONCOOCCURRENCE.tsv", nil, "token_B"}},
		"position":    {{"TOKEN_POSITION_RULES.tsv", nil, "token"}}, "section": {{"SECTION_RULES.tsv", nil, "entity"}},
	}
	qs, ok := specs[kind]
	if !ok {
		fmt.Fprintln(errOut, "unknown query type:", kind)
		return 2
	}
	found := 0
	seen := map[string]bool{}
	for _, q := range qs {
		path := filepath.Join(*dir, q.file)
		f, e := os.Open(path)
		if e != nil {
			fmt.Fprintln(errOut, "Error:", e)
			return 1
		}
		rd := csv.NewReader(f)
		rd.Comma = '\t'
		rd.FieldsPerRecord = -1
		header, e := rd.Read()
		if e != nil {
			f.Close()
			fmt.Fprintln(errOut, "Error:", e)
			return 1
		}
		mi := -1
		for i, h := range header {
			if h == q.match {
				mi = i
			}
		}
		if mi < 0 {
			f.Close()
			continue
		}
		for {
			row, e := rd.Read()
			if e == io.EOF {
				break
			}
			if e != nil {
				f.Close()
				fmt.Fprintln(errOut, "Error:", e)
				return 1
			}
			if mi < len(row) && row[mi] == value {
				k := q.file + "\x00" + strings.Join(row, "\x00")
				if seen[k] {
					continue
				}
				seen[k] = true
				if found == 0 || !seen["#"+q.file] {
					fmt.Fprintf(out, "# %s\n%s\n", q.file, strings.Join(header, "\t"))
					seen["#"+q.file] = true
				}
				fmt.Fprintln(out, strings.Join(row, "\t"))
				found++
			}
		}
		f.Close()
	}
	if found == 0 {
		fmt.Fprintf(out, "No frozen catalog rows for %q in query %s.\n", value, kind)
	}
	return 0
}
