package structurecatalog

import (
	"bufio"
	"crypto/sha256"
	"fmt"
	"os"
	"sort"
	"strings"

	"zcore.dev/voinich/internal/metadatavalidation"
)

func LoadCorpus(path, transcription, ivtff string) (Corpus, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return Corpus{}, err
	}
	c := Corpus{Path: path, SHA: fmt.Sprintf("%x", sha256.Sum256(b)), Transcription: transcription, Counts: map[string]int{}}
	s := bufio.NewScanner(strings.NewReader(string(b)))
	for s.Scan() {
		fields := strings.Fields(s.Text())
		if len(fields) == 0 {
			continue
		}
		line := append([]string(nil), fields...)
		li := len(c.Lines)
		c.Lines = append(c.Lines, line)
		for i, t := range line {
			c.Counts[t]++
			c.Occurrences = append(c.Occurrences, Occurrence{Token: t, Line: li, Index: i})
		}
	}
	if err := s.Err(); err != nil {
		return Corpus{}, err
	}
	if len(c.Occurrences) == 0 {
		return Corpus{}, fmt.Errorf("empty corpus: %s", path)
	}
	if ivtff != "" {
		d, e := metadatavalidation.ParseIVTFF(ivtff)
		if e != nil {
			return Corpus{}, fmt.Errorf("parse metadata: %w", e)
		}
		tokens, hash, e := metadatavalidation.ReadFrozenCorpus(path)
		if e != nil {
			return Corpus{}, e
		}
		a, e := metadatavalidation.Align(d, tokens, hash)
		if e != nil {
			return Corpus{}, fmt.Errorf("strict metadata alignment: %w", e)
		}
		if len(a.Records) != len(c.Occurrences) {
			return Corpus{}, fmt.Errorf("metadata records %d != occurrences %d", len(a.Records), len(c.Occurrences))
		}
		folios, sections, loci := map[string]bool{}, map[string]bool{}, map[string]bool{}
		for i := range c.Occurrences {
			c.Occurrences[i].Meta = a.Records[i]
			folios[a.Records[i].Folio] = true
			if a.Records[i].Section != "" {
				sections[a.Records[i].Section] = true
			}
			if a.Records[i].LocusType != "" {
				loci[a.Records[i].LocusType] = true
			}
		}
		c.Folios = keys(folios)
		c.Sections = keys(sections)
		c.LocusTypes = keys(loci)
		c.MetadataAvailable = true
	}
	glyphs := map[rune]bool{}
	for t := range c.Counts {
		for _, g := range []rune(t) {
			glyphs[g] = true
		}
	}
	for g := range glyphs {
		c.Inventory = append(c.Inventory, g)
	}
	sort.Slice(c.Inventory, func(i, j int) bool { return c.Inventory[i] < c.Inventory[j] })
	return c, nil
}

func keys(m map[string]bool) []string {
	r := make([]string, 0, len(m))
	for x := range m {
		r = append(r, x)
	}
	sort.Strings(r)
	return r
}

func groupMeta(c Corpus, field string) map[int]string {
	r := map[int]string{}
	for i, o := range c.Occurrences {
		switch field {
		case "folio":
			r[i] = o.Meta.Folio
		case "section":
			r[i] = o.Meta.Section
		case "locus":
			r[i] = o.Meta.LocusType
		}
	}
	return r
}
