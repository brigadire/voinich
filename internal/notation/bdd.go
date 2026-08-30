package notation

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/xml"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"unicode"

	"golang.org/x/text/unicode/norm"
)

// BDDConversionStats records source-to-USC reconciliation counts. It is
// deliberately structural only and contains no comparative measurements.
type BDDConversionStats struct {
	SourceFiles       int `json:"source_files"`
	Pages             int `json:"pages"`
	PhysicalLines     int `json:"physical_lines"`
	Tokens            int `json:"tokens"`
	AbbreviationSigns int `json:"abbreviation_signs"`
	Choices           int `json:"choices"`
	SkippedApparatus  int `json:"skipped_apparatus_subtrees"`
}

type bddFrame struct {
	name string
	skip bool
}

type bddBuilder struct {
	corpusID, representation, document, section, page, locus, line string
	text                                                           strings.Builder
	records                                                        []Record
	tokenCounters                                                  map[string]int
	ordinal                                                        int
	stats                                                          BDDConversionStats
	placeholders                                                   map[string]string
}

// BuildBDDUSC converts the frozen Burchards Dekret Digital TEI snapshot to
// canonical USC. representation must be LATIN-DIPLOMATIC (the abbr branch)
// or LATIN-EXPANDED (the expan branch). Files and glyph refs are sorted before
// conversion, so caller argument order and Go map iteration cannot affect it.
func BuildBDDUSC(paths []string, corpusID, representation string) ([]Record, BDDConversionStats, error) {
	if representation != "LATIN-DIPLOMATIC" && representation != "LATIN-EXPANDED" {
		return nil, BDDConversionStats{}, fmt.Errorf("unsupported BDD representation %q", representation)
	}
	if len(paths) == 0 || corpusID == "" {
		return nil, BDDConversionStats{}, fmt.Errorf("BDD paths and corpus ID are required")
	}
	paths = append([]string(nil), paths...)
	sort.Strings(paths)
	refs := map[string]bool{}
	for _, path := range paths {
		if err := collectBDDGlyphRefs(path, refs); err != nil {
			return nil, BDDConversionStats{}, err
		}
	}
	names := make([]string, 0, len(refs))
	for ref := range refs {
		names = append(names, ref)
	}
	sort.Strings(names)
	placeholders := map[string]string{}
	for i, ref := range names {
		// Lowercase Glagolitic letters are stable under the frozen lowercase
		// policy and remain letters for generic symbol processing.
		placeholders[ref] = string(rune(0x2C30 + i))
	}
	b := &bddBuilder{corpusID: corpusID, representation: representation, placeholders: placeholders, tokenCounters: map[string]int{}}
	for _, path := range paths {
		if err := b.convertFile(path); err != nil {
			return nil, BDDConversionStats{}, err
		}
	}
	b.flush()
	b.stats.SourceFiles = len(paths)
	b.stats.Tokens = len(b.records)
	if err := Validate(b.records); err != nil {
		return nil, BDDConversionStats{}, fmt.Errorf("generated BDD USC: %w", err)
	}
	return b.records, b.stats, nil
}

func collectBDDGlyphRefs(path string, refs map[string]bool) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()
	d := xml.NewDecoder(f)
	d.Strict = false
	for {
		t, err := d.Token()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return fmt.Errorf("parse %s: %w", path, err)
		}
		if e, ok := t.(xml.StartElement); ok && e.Name.Local == "g" {
			ref := xmlAttr(e, "ref")
			if ref == "" {
				ref = "#unknown"
			}
			refs[ref] = true
		}
	}
}

func xmlAttr(e xml.StartElement, name string) string {
	for _, a := range e.Attr {
		if a.Name.Local == name {
			return a.Value
		}
	}
	return ""
}

func (b *bddBuilder) convertFile(path string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()
	d := xml.NewDecoder(f)
	d.Strict = false
	b.document = strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))
	b.section, b.page, b.locus, b.line = "", "", "", ""
	var stack []bddFrame
	inBody := false
	skipNow := func() bool { return len(stack) != 0 && stack[len(stack)-1].skip }
	for {
		t, err := d.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("parse %s: %w", path, err)
		}
		switch x := t.(type) {
		case xml.StartElement:
			name := x.Name.Local
			if name == "body" {
				inBody = true
			}
			skip := skipNow()
			apparatus := name == "teiHeader" || name == "note" || name == "fw" || name == "label" || name == "del" || name == "pc"
			if name == "div" && xmlAttr(x, "type") == "toc" {
				apparatus = true
			}
			if apparatus {
				skip = true
				if inBody {
					b.separator()
					b.stats.SkippedApparatus++
				}
			}
			if name == "choice" && !skip {
				b.stats.Choices++
			}
			if (name == "abbr" && b.representation == "LATIN-EXPANDED") || (name == "expan" && b.representation == "LATIN-DIPLOMATIC") {
				skip = true
			}
			if !skip && inBody {
				switch name {
				case "div":
					if xmlAttr(x, "type") == "book" && xmlAttr(x, "n") != "" {
						b.flush()
						b.section = xmlAttr(x, "n")
					}
				case "pb":
					b.flush()
					b.page, b.locus, b.line = xmlAttr(x, "n"), "", ""
					b.stats.Pages++
				case "cb":
					b.flush()
					b.locus, b.line = xmlAttr(x, "n"), ""
				case "lb":
					b.flush()
					b.line = xmlAttr(x, "id")
					if b.line == "" {
						b.line = xmlAttr(x, "n")
					}
					b.stats.PhysicalLines++
				case "g":
					ref := xmlAttr(x, "ref")
					if ref == "" {
						ref = "#unknown"
					}
					p, ok := b.placeholders[ref]
					if !ok {
						return fmt.Errorf("unmapped BDD glyph ref %q", ref)
					}
					b.text.WriteString(p)
					b.stats.AbbreviationSigns++
					skip = true // placeholder replaces the literal PUA child
				}
			}
			stack = append(stack, bddFrame{name: name, skip: skip})
		case xml.EndElement:
			if len(stack) != 0 {
				stack = stack[:len(stack)-1]
			}
			if x.Name.Local == "body" {
				inBody = false
			}
		case xml.CharData:
			if inBody && !skipNow() {
				b.text.Write(x)
			}
		}
	}
	b.flush()
	return nil
}

func (b *bddBuilder) separator() {
	if b.text.Len() != 0 {
		b.text.WriteByte(' ')
	}
}

func (b *bddBuilder) flush() {
	text := norm.NFC.String(strings.ToLower(b.text.String()))
	b.text.Reset()
	for _, token := range strings.FieldsFunc(text, func(r rune) bool {
		return unicode.IsSpace(r) || unicode.IsPunct(r) || unicode.IsSymbol(r)
	}) {
		if token == "" {
			continue
		}
		syms := make([]string, 0, len([]rune(token)))
		for _, r := range token {
			syms = append(syms, string(r))
		}
		identity := fmt.Sprintf("%s\x00%s\x00%s\x00%d", b.corpusID, b.representation, b.document, b.ordinal)
		h := sha256.Sum256([]byte(identity))
		hierarchyKey := strings.Join([]string{b.document, b.section, b.page, b.locus, b.line}, "\x1f")
		tokenIndex := b.tokenCounters[hierarchyKey]
		b.tokenCounters[hierarchyKey]++
		b.records = append(b.records, Record{
			SchemaVersion: SchemaVersion, CorpusID: b.corpusID, Representation: b.representation,
			Document: level(b.document), Section: level(b.section), Page: level(b.page), Locus: level(b.locus), PhysicalLine: level(b.line),
			TokenID: b.corpusID + "-" + hex.EncodeToString(h[:8]), TokenIndex: tokenIndex, Token: token, Symbols: syms,
		})
		b.ordinal++
	}
}
