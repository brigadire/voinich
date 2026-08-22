package main

import (
	"bufio"
	"os"
	"strconv"
	"strings"
)

func readTSV(path string) (header []string, rows [][]string, err error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, nil, err
	}
	defer f.Close()
	sc := bufio.NewScanner(f)
	sc.Buffer(make([]byte, 64*1024), 1024*1024)
	first := true
	for sc.Scan() {
		cols := strings.Split(sc.Text(), "\t")
		if first {
			header = cols
			first = false
			continue
		}
		rows = append(rows, cols)
	}
	return header, rows, sc.Err()
}

func colIndex(header []string, name string) int {
	for i, h := range header {
		if h == name {
			return i
		}
	}
	return -1
}

// task60CorpusAliases maps this task's corpus names onto the (differing)
// names Task58/59 used for the same underlying corpus, since the two
// pre-existing tools named things independently (e.g. "Doyle-H4" vs
// "Doyle-H4-uniform").
func normalizeCorpusName(n string) string {
	n = strings.TrimSuffix(n, "-uniform")
	return n
}

// joinTask58 implements task60 section 35: pull the already-computed
// Task58 token-order/glyph-edge MI for the same corpora, rather than
// recomputing the estimand.
func joinTask58(rep *report, outDir string) error {
	header, rows, err := readTSV("experiments/rozanova-temerev-v1/comparison.tsv")
	if err != nil {
		rep.note("Task58 comparison.tsv not found (%v); TASK58_COMPARISON.tsv not written.", err)
		return nil
	}
	corpusIdx, tokenObsIdx, tokenShareIdx, edgeObsIdx := colIndex(header, "corpus"), colIndex(header, "token_observed_bits"), colIndex(header, "token_share"), colIndex(header, "edge_observed_bits")
	byName := map[string][]string{}
	for _, r := range rows {
		if corpusIdx < len(r) {
			byName[normalizeCorpusName(r[corpusIdx])] = r
		}
	}
	out := newTSV(outDir, "TASK58_COMPARISON.tsv", "Corpus", "TokenOrderMI", "TokenShare", "GlyphEdgeMI")
	for _, name := range rep.order {
		r, ok := byName[normalizeCorpusName(name)]
		if !ok {
			continue
		}
		mi, _ := strconv.ParseFloat(getCol(r, tokenObsIdx), 64)
		share, _ := strconv.ParseFloat(getCol(r, tokenShareIdx), 64)
		edge, _ := strconv.ParseFloat(getCol(r, edgeObsIdx), 64)
		out.row(name, f8(mi), f8(share), f8(edge))
		s := rep.summary(name)
		s.HasTask58, s.TokenOrderMI, s.TokenShare, s.GlyphEdgeMI = true, mi, share, edge
	}
	out.close()
	return nil
}

// joinTask59 implements task60 section 37: pull the already-computed
// Task59 positional-specialization comparison for the same corpora.
func joinTask59(rep *report, outDir string) error {
	header, rows, err := readTSV("experiments/glyph-position-v1/POSITIONAL_SPECIALIZATION_COMPARISON.tsv")
	if err != nil {
		rep.note("Task59 comparison TSV not found (%v); TASK59_COMPARISON.tsv not written.", err)
		return nil
	}
	corpusIdx, hfIdx, weIdx := colIndex(header, "Corpus"), colIndex(header, "HighFreqSpecialists"), colIndex(header, "WeightedEntropy")
	byName := map[string][]string{}
	for _, r := range rows {
		if corpusIdx < len(r) {
			byName[normalizeCorpusName(r[corpusIdx])] = r
		}
	}
	out := newTSV(outDir, "TASK59_COMPARISON.tsv", "Corpus", "HighFreqSpecialists", "WeightedEntropy")
	for _, name := range rep.order {
		r, ok := byName[normalizeCorpusName(name)]
		if !ok {
			continue
		}
		hf, _ := strconv.Atoi(getCol(r, hfIdx))
		we, _ := strconv.ParseFloat(getCol(r, weIdx), 64)
		out.row(name, i(hf), f8(we))
		s := rep.summary(name)
		s.HasTask59, s.HighFreqSpecialists, s.WeightedEntropy = true, hf, we
	}
	out.close()
	return nil
}

func getCol(r []string, idx int) string {
	if idx < 0 || idx >= len(r) {
		return ""
	}
	return r[idx]
}
