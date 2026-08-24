package fingerprintv2

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	"zcore.dev/voinich/internal/evaglyph"
	"zcore.dev/voinich/internal/genericsegmentation"
	"zcore.dev/voinich/internal/metadatavalidation"
)

type tokenRecord struct {
	Token string
	Glyph []string
	Line  int
	Page  string

	// The fields below are populated only under strict IVTFF alignment
	// (task77 cross-scale block); they are zero-valued otherwise and
	// callers must check corpus.info.MetadataAlignment before relying on
	// them.
	LineID         string
	LocusID        string
	IndexInLine    int
	IndexInLocus   int
	IndexInFolio   int
	LineLength     int
	LocusType      string
	ParagraphID    int
	ParagraphStart bool
	FolioSide      string
	Section        string
	Currier        string
	Hand           string
	Quire          string
	IVTFFI         string
	IVTFFX         string
}

type corpus struct {
	info    CorpusInfo
	records []tokenRecord
}

func loadCorpus(c CorpusConfig) (corpus, error) {
	mode, err := validateGlyphMode(c.GlyphMode)
	if err != nil {
		return corpus{}, err
	}
	tokens, lines, sha, err := genericsegmentation.ReadCorpus(c.Path)
	if err != nil {
		return corpus{}, fmt.Errorf("read corpus %q: %w", c.ID, err)
	}
	if len(tokens) == 0 {
		return corpus{}, fmt.Errorf("corpus %q has no whitespace-delimited tokens", c.ID)
	}
	out := corpus{records: make([]tokenRecord, len(tokens))}
	alignment := "not requested"
	pages := make([]string, len(tokens))
	meta := make([]metadatavalidation.TokenMetadata, len(tokens))
	haveMeta := false
	if c.IVTFFPath != "" {
		doc, e := metadatavalidation.ParseIVTFF(c.IVTFFPath)
		if e != nil {
			return corpus{}, fmt.Errorf("parse IVTFF metadata for %q: %w", c.ID, e)
		}
		aligned, e := metadatavalidation.Align(doc, tokens, sha)
		if e != nil {
			return corpus{}, fmt.Errorf("strict IVTFF alignment for %q: %w", c.ID, e)
		}
		if len(aligned.Records) != len(tokens) {
			return corpus{}, fmt.Errorf("strict IVTFF alignment for %q returned %d records for %d tokens", c.ID, len(aligned.Records), len(tokens))
		}
		lineIndex := map[string]int{}
		for i, r := range aligned.Records {
			key := r.Folio + "\x00" + r.LineID
			v, ok := lineIndex[key]
			if !ok {
				v = len(lineIndex)
				lineIndex[key] = v
			}
			lines[i] = v
			pages[i] = r.Folio
		}
		meta = aligned.Records
		haveMeta = true
		alignment = "strict IVTFF aligned"
	}
	lineLength := map[int]int{}
	linePosition := map[int]int{}
	for i := range tokens {
		lineLength[lines[i]]++
	}
	for i, token := range tokens {
		glyphs := glyphsFor(token, mode)
		if len(glyphs) == 0 {
			return corpus{}, fmt.Errorf("corpus %q token %d has no glyphs after %s preprocessing", c.ID, i, mode)
		}
		rec := tokenRecord{Token: glyphKey(glyphs), Glyph: glyphs, Line: lines[i], Page: pages[i],
			LineID: "line:" + strconv.Itoa(lines[i]), IndexInLine: linePosition[lines[i]], LineLength: lineLength[lines[i]]}
		linePosition[lines[i]]++
		if haveMeta {
			m := meta[i]
			rec.LineID = m.Folio + "\x00" + m.LineID
			rec.LocusID = m.Folio + "\x00" + m.LocusID
			rec.IndexInLine = m.IndexInLine
			rec.IndexInLocus = m.IndexInLocus
			rec.IndexInFolio = m.IndexInFolio
			rec.LineLength = lineLength[lines[i]]
			rec.LocusType = m.LocusType
			rec.ParagraphID = m.ParagraphID
			rec.ParagraphStart = m.ParagraphStart
			rec.FolioSide = folioSide(m.Folio)
			rec.Section = m.Section
			rec.Currier = m.Currier
			rec.Hand = m.Hand
			rec.Quire = m.Quire
			rec.IVTFFI = m.Variables["I"]
			rec.IVTFFX = m.Variables["X"]
		}
		out.records[i] = rec
	}
	out.info = corpusInfo(c, sha, mode, alignment, out.records)
	return out, nil
}

// folioSide returns the trailing recto/verso letter of an IVTFF folio
// identifier (e.g. "f1r" -> "r"), or "" if the identifier does not end in
// one of the two documented side letters.
func folioSide(folio string) string {
	if folio == "" {
		return ""
	}
	last := folio[len(folio)-1]
	if last == 'r' || last == 'v' {
		return string(last)
	}
	return ""
}

func glyphsFor(token, mode string) []string {
	if mode == "natural" {
		return evaglyph.NaturalGlyphs(token)
	}
	return evaglyph.CollapseEVA(token)
}

// glyphKey is an unambiguous internal type identity. It is intentionally not
// re-parsed as EVA; generated corpora retain atomic composite glyphs.
func glyphKey(glyphs []string) string { return strings.Join(glyphs, "\x1f") }

func corpusInfo(c CorpusConfig, sha, mode, alignment string, records []tokenRecord) CorpusInfo {
	types := map[string]bool{}
	line, page := false, false
	for _, r := range records {
		types[r.Token] = true
		line = true
		if r.Page != "" {
			page = true
		}
	}
	profile := "generic whitespace tokens with natural line boundaries"
	if mode == "eva" {
		profile += "; EVA composites collapsed by internal/evaglyph"
	} else {
		profile += "; lowercase Unicode letters/digits by internal/evaglyph"
	}
	if c.IVTFFPath != "" {
		profile += "; strict IVTFF token alignment"
	}
	return CorpusInfo{
		ID: c.ID, Path: c.Path, SHA256: sha, TokenCount: len(records), VocabularySize: len(types),
		GlyphMode: mode, MetadataAlignment: alignment, LineMetadata: line, PageMetadata: page,
		Preprocessing: profile,
	}
}

func generatedCorpus(source corpus, glyphs [][]string) corpus {
	out := corpus{records: make([]tokenRecord, len(glyphs)), info: source.info}
	for i, g := range glyphs {
		src := source.records[i]
		out.records[i] = tokenRecord{
			Token: glyphKey(g), Glyph: append([]string(nil), g...),
			Line: src.Line, Page: src.Page,
			LineID: src.LineID, IndexInLine: src.IndexInLine, LineLength: src.LineLength,
			LocusID: src.LocusID, IndexInLocus: src.IndexInLocus, IndexInFolio: src.IndexInFolio,
			LocusType: src.LocusType, ParagraphID: src.ParagraphID, ParagraphStart: src.ParagraphStart, FolioSide: src.FolioSide,
			Section: src.Section, Currier: src.Currier, Hand: src.Hand, Quire: src.Quire, IVTFFI: src.IVTFFI, IVTFFX: src.IVTFFX,
		}
	}
	out.info.ID = source.info.ID + ":c-grammar"
	out.info.Path = ""
	out.info.SHA256 = ""
	out.info.TokenCount = len(out.records)
	out.info.VocabularySize = len(vocabulary(out))
	out.info.MetadataAlignment = "synthetic positions inherited; no lexical alignment"
	out.info.Preprocessing = source.info.Preprocessing + "; C-GRAMMAR synthetic tokens"
	return out
}

func vocabulary(c corpus) []string {
	m := map[string]bool{}
	for _, r := range c.records {
		m[r.Token] = true
	}
	out := make([]string, 0, len(m))
	for t := range m {
		out = append(out, t)
	}
	sort.Strings(out)
	return out
}

func glyphByToken(c corpus) map[string][]string {
	out := map[string][]string{}
	for _, r := range c.records {
		if _, ok := out[r.Token]; !ok {
			out[r.Token] = r.Glyph
		}
	}
	return out
}

func frequencies(c corpus) map[string]int {
	out := map[string]int{}
	for _, r := range c.records {
		out[r.Token]++
	}
	return out
}
