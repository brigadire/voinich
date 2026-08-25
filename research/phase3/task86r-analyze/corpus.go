package main

import (
	"bufio"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"regexp"
	"sort"
	"strings"

	"zcore.dev/voinich/internal/evaglyph"
	"zcore.dev/voinich/internal/genericsegmentation"
	"zcore.dev/voinich/internal/metadatavalidation"
)

// TokenOccurrence is one whitespace-delimited TOKEN occurrence, aligned to
// its frozen Task85 partition. Alignment reuses Task85's own machinery
// (metadatavalidation.ParseIVTFF + Align, internal/evaglyph.CollapseEVA),
// per task86r.txt section: "Использовать Task85 alignment machinery. Не
// выполнять naive line-index alignment." The split itself is read from the
// frozen GRAMMAR_CORPUS_SPLIT.tsv, never regenerated.
type TokenOccurrence struct {
	Raw       string
	Glyphs    []string
	Partition string // DEVELOPMENT, VALIDATION, HELDOUT
	Leaf      string // originating physical leaf (folio_id in GRAMMAR_CORPUS_SPLIT.tsv), for leaf-ordered nested prefixes
}

type TranscriptionCorpus struct {
	Name        string
	CanonicalSHA256 string
	Occurrences []TokenOccurrence
}

func (c TranscriptionCorpus) Partition(p string) []TokenOccurrence {
	var out []TokenOccurrence
	for _, o := range c.Occurrences {
		if o.Partition == p {
			out = append(out, o)
		}
	}
	return out
}

var leafRe = regexp.MustCompile(`^(f[0-9]+)`)

func leafOf(pageID string) string {
	m := leafRe.FindStringSubmatch(pageID)
	if m == nil {
		return pageID
	}
	return m[1]
}

// leafNumber extracts the numeric leaf index (e.g. "f52" -> 52) for
// deterministic numeric leaf ordering; a non-numeric leaf (e.g. "fRos")
// sorts after all numbered leaves via a large sentinel.
func leafNumber(leaf string) int {
	n := 0
	found := false
	for _, r := range leaf {
		if r >= '0' && r <= '9' {
			n = n*10 + int(r-'0')
			found = true
		}
	}
	if !found {
		return 1 << 30
	}
	return n
}

func sha256File(path string) (string, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	h := sha256.Sum256(b)
	return hex.EncodeToString(h[:]), nil
}

func loadSplitPartitions(path string) (map[string]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	out := map[string]string{}
	sc := bufio.NewScanner(f)
	sc.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	header := true
	var cols []string
	for sc.Scan() {
		line := sc.Text()
		if line == "" {
			continue
		}
		fields := strings.Split(line, "\t")
		if header {
			cols = fields
			header = false
			continue
		}
		row := map[string]string{}
		for i, c := range cols {
			if i < len(fields) {
				row[c] = fields[i]
			}
		}
		out[row["folio_id"]] = row["partition"]
	}
	if err := sc.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

type corpusSource struct {
	name      string
	canonical string
	ivtff     string
}

var corpusSources = []corpusSource{
	{"ZL3b", "data_work/ZL3b-x7.canonical.txt", "data/ZL3b-n.txt"},
	{"IT2a", "data_work/IT2a-x7.canonical.txt", "data/IT2a-n.txt"},
}

func loadTranscription(s corpusSource, splitByFolio map[string]string) (TranscriptionCorpus, error) {
	rawTokens, _, sha, err := genericsegmentation.ReadCorpus(s.canonical)
	if err != nil {
		return TranscriptionCorpus{}, err
	}
	doc, err := metadatavalidation.ParseIVTFF(s.ivtff)
	if err != nil {
		return TranscriptionCorpus{}, err
	}
	aligned, err := metadatavalidation.Align(doc, rawTokens, sha)
	if err != nil {
		return TranscriptionCorpus{}, fmt.Errorf("%s: strict IVTFF alignment: %w", s.name, err)
	}
	if len(aligned.Records) != len(rawTokens) {
		return TranscriptionCorpus{}, fmt.Errorf("%s: aligned %d records for %d tokens", s.name, len(aligned.Records), len(rawTokens))
	}
	occ := make([]TokenOccurrence, len(rawTokens))
	for i, raw := range rawTokens {
		leaf := leafOf(aligned.Records[i].Folio)
		part, ok := splitByFolio[leaf]
		if !ok {
			return TranscriptionCorpus{}, fmt.Errorf("%s: token %d folio %q (leaf %q) missing from frozen split", s.name, i, aligned.Records[i].Folio, leaf)
		}
		occ[i] = TokenOccurrence{
			Raw:       raw,
			Glyphs:    evaglyph.CollapseEVA(raw),
			Partition: part,
			Leaf:      leaf,
		}
	}
	return TranscriptionCorpus{Name: s.name, CanonicalSHA256: sha, Occurrences: occ}, nil
}

// glyphAlphabet returns the sorted distinct glyph inventory over a set of
// token occurrences (used to fix a model's DEVELOPMENT alphabet).
func glyphAlphabet(occ []TokenOccurrence) []string {
	set := map[string]bool{}
	for _, o := range occ {
		for _, g := range o.Glyphs {
			set[g] = true
		}
	}
	out := make([]string, 0, len(set))
	for g := range set {
		out = append(out, g)
	}
	sort.Strings(out)
	return out
}
