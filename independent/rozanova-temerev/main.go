package main

import (
	"encoding/csv"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

type row struct {
	Corpus string `json:"corpus"`
	Path   string `json:"path"`
	Metric `json:",inline"`
}

func main() {
	if len(os.Args) > 1 && os.Args[1] == "compare" {
		compare(os.Args[2:])
		return
	}
	f := flag.NewFlagSet("analyze", flag.ExitOnError)
	path := f.String("corpus", "", "input corpus")
	out := f.String("output-dir", "", "output directory")
	mode := f.String("glyph-mode", "text", "text, voynich, or opaque")
	cap := f.Int("vocabulary-cap", defaultCap, "R&T top-N cap")
	sh := f.Int("shuffles", defaultShuffles, "within-line shuffles")
	seed := f.Int64("seed", defaultSeed, "deterministic seed")
	name := f.String("name", "corpus", "row name")
	f.Parse(os.Args[1:])
	if *path == "" {
		f.Usage()
		os.Exit(2)
	}
	c, e := loadCorpus(*path, *mode)
	if e != nil {
		panic(e)
	}
	m, e := analyse(c, *cap, *sh, *seed)
	if e != nil {
		panic(e)
	}
	r := row{Corpus: *name, Path: *path, Metric: m}
	write(*out, r, c, *cap, *sh, *seed)
}
func write(dir string, r row, c Corpus, cap, sh int, seed int64) {
	if dir == "" {
		b, _ := json.MarshalIndent(r, "", "  ")
		fmt.Println(string(b))
		return
	}
	os.MkdirAll(dir, 0755)
	b, _ := json.MarshalIndent(map[string]any{"provenance": map[string]any{"path": r.Path, "sha256": c.SHA256, "bytes": c.Bytes, "lines": r.Lines, "tokens": r.Tokens, "vocabulary_cap": cap, "shuffles": sh, "seed": seed, "glyph_mode": c.GlyphMode}, "result": r.Metric}, "", "  ")
	os.WriteFile(filepath.Join(dir, "result.json"), append(b, '\n'), 0644)
	writeTSV(filepath.Join(dir, "results.tsv"), []row{r})
}
func writeTSV(path string, rows []row) {
	f, _ := os.Create(path)
	defer f.Close()
	w := csv.NewWriter(f)
	w.Comma = '\t'
	w.Write([]string{"corpus", "path", "tokens", "lines", "pairs", "types", "token_observed_bits", "token_shuffle_mean_bits", "token_shuffle_sd_bits", "token_corrected_bits", "token_share", "glyph_status", "edge_observed_bits", "edge_shuffle_mean_bits", "edge_shuffle_sd_bits", "edge_corrected_bits", "edge_share"})
	for _, r := range rows {
		m := r.Metric
		w.Write([]string{r.Corpus, r.Path, strconv.Itoa(m.Tokens), strconv.Itoa(m.Lines), strconv.Itoa(m.Pairs), strconv.Itoa(m.Types), fmt.Sprint(m.TokenObserved), fmt.Sprint(m.TokenShuffleMean), fmt.Sprint(m.TokenShuffleSD), fmt.Sprint(m.TokenCorrected), fmt.Sprint(m.TokenShare), m.GlyphStatus, fmt.Sprint(m.EdgeObserved), fmt.Sprint(m.EdgeShuffleMean), fmt.Sprint(m.EdgeShuffleSD), fmt.Sprint(m.EdgeCorrected), fmt.Sprint(m.EdgeShare)})
	}
	w.Flush()
}
func compare(args []string) {
	f := flag.NewFlagSet("compare", flag.ExitOnError)
	out := f.String("output-dir", "", "output directory")
	cap := f.Int("vocabulary-cap", defaultCap, "R&T top-N cap")
	sh := f.Int("shuffles", defaultShuffles, "within-line shuffles")
	seed := f.Int64("seed", defaultSeed, "seed")
	modes := f.String("voynich", "", "comma-separated NAME=PATH entries for Voynich mode")
	f.Parse(args)
	specs := f.Args()
	rows := []row{}
	for _, spec := range specs {
		n, p, ok := strings.Cut(spec, "=")
		if !ok {
			panic("compare expects NAME=PATH")
		}
		mode := "text"
		lowerPath := strings.ToLower(p)
		if strings.Contains(strings.ToLower(n), "voynich") || strings.HasSuffix(p, "canonical.txt") {
			mode = "voynich"
		}
		if strings.Contains(lowerPath, "transformed") || strings.Contains(strings.ToLower(n), "homophonic") {
			mode = "opaque"
		}
		for _, v := range strings.Split(*modes, ",") {
			if x, y, ok := strings.Cut(v, "="); ok && x == n {
				p = y
				mode = "voynich"
			}
		}
		c, e := loadCorpus(p, mode)
		if e != nil {
			panic(e)
		}
		mm, e := analyse(c, *cap, *sh, *seed)
		if e != nil {
			panic(e)
		}
		rows = append(rows, row{n, p, mm})
	}
	if *out == "" {
		writeTSV("/dev/stdout", rows)
	} else {
		os.MkdirAll(*out, 0755)
		writeTSV(filepath.Join(*out, "comparison.tsv"), rows)
		b, _ := json.MarshalIndent(map[string]any{"vocabulary_cap": *cap, "shuffles": *sh, "seed": *seed, "rows": rows}, "", "  ")
		os.WriteFile(filepath.Join(*out, "comparison.json"), append(b, '\n'), 0644)
	}
}
