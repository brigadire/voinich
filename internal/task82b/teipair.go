// TEI paired abbreviated/expanded extraction (task82b.txt sec.7/9/12/13).
//
// This is an independent implementation, not a modification of
// cmd/tei-abbr-extract (which is Task79c's frozen abbr-only extractor):
// Task82b must not touch Task79c's frozen artifacts, and it additionally
// needs the <expan> branch that tool deliberately drops. Both tools share
// the same apparatus-exclusion policy (teiHeader/note/fw/label/toc
// skipped, <lb/> is a line break, <g ref="..."/> becomes a reserved
// Glagolitic placeholder rune) so the two streams stay line-aligned and
// evaglyph-safe.
package task82b

import (
	"encoding/xml"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"
)

// glagoliticBase mirrors cmd/tei-abbr-extract's reserved placeholder
// alphabet (U+2C00, Unicode Glagolitic block; satisfies unicode.IsLetter
// so evaglyph/NaturalGlyphs never silently drops a combining abbreviation
// mark or PUA glyph).
const glagoliticBase = rune(0x2C00)
const glagoliticSize = 90

// PairUnit is one TEI <choice> instance: the manuscript's own abbreviated
// surface form and the edition's expansion, in document order.
type PairUnit struct {
	File            string `json:"file"`
	Order           int    `json:"order"`
	Line            int    `json:"line"` // 0-based line index in both output streams
	AbbrText        string `json:"abbr_text"`
	ExpanText       string `json:"expan_text"`
	HasMark         bool   `json:"has_mark"`          // <g> present inside <abbr>
	MarkIsCombining bool   `json:"mark_is_combining"` // ref names a combining-diacritic codepoint (U+0300-U+036F)
}

// TEIPairResult is the paired extraction of one witness (a set of TEI
// chapter files sharing one glyph-placeholder table).
type TEIPairResult struct {
	AbbrText          string
	ExpanText         string
	AbbrGroups        [][]string // per-line tokens, abbreviated stream
	ExpanGroups       [][]string // per-line tokens, expanded stream
	Pairs             []PairUnit
	GlyphPlaceholders map[string]string
	Notes             []string
}

// skipSubtreeTEI mirrors cmd/tei-abbr-extract's apparatus exclusion.
var skipSubtreeTEI = map[string]bool{"teiHeader": true, "note": true, "fw": true, "label": true}

// ExtractTEIPairs walks one or more TEI-XML files (sorted for
// determinism) and produces both the abbreviated and the expanded
// diplomatic streams, plus the ordered list of <choice> pair units.
func ExtractTEIPairs(paths []string) (TEIPairResult, error) {
	sorted := append([]string(nil), paths...)
	sort.Strings(sorted)

	refs := map[string]bool{}
	for _, p := range sorted {
		if err := collectGRefs(p, refs); err != nil {
			return TEIPairResult{}, fmt.Errorf("collect refs from %s: %w", p, err)
		}
	}
	sortedRefs := make([]string, 0, len(refs))
	for r := range refs {
		sortedRefs = append(sortedRefs, r)
	}
	sort.Strings(sortedRefs)
	if len(sortedRefs) > glagoliticSize {
		return TEIPairResult{}, fmt.Errorf("%d distinct <g ref> values exceeds reserved placeholder alphabet size %d", len(sortedRefs), glagoliticSize)
	}
	placeholder := map[string]rune{}
	placeholderStr := map[string]string{}
	for i, r := range sortedRefs {
		p := glagoliticBase + rune(i)
		placeholder[r] = p
		placeholderStr[r] = string(p)
	}

	var abbrOut, expanOut strings.Builder
	var pairs []PairUnit
	order := 0
	abbrLine, expanLine := 0, 0
	for fi, p := range sorted {
		abbrText, expanText, filePairs, newAbbrLine, newExpanLine, err := extractPairFile(p, placeholder, order, abbrLine, expanLine)
		if err != nil {
			return TEIPairResult{}, fmt.Errorf("extract %s: %w", p, err)
		}
		abbrOut.WriteString(abbrText)
		expanOut.WriteString(expanText)
		if fi != len(sorted)-1 {
			abbrOut.WriteString("\n")
			expanOut.WriteString("\n")
			newAbbrLine++
			newExpanLine++
		}
		pairs = append(pairs, filePairs...)
		order += len(filePairs)
		abbrLine, expanLine = newAbbrLine, newExpanLine
	}

	abbrGroups := tokenizeLines(abbrOut.String())
	expanGroups := tokenizeLines(expanOut.String())

	return TEIPairResult{
		AbbrText:          abbrOut.String(),
		ExpanText:         expanOut.String(),
		AbbrGroups:        abbrGroups,
		ExpanGroups:       expanGroups,
		Pairs:             pairs,
		GlyphPlaceholders: placeholderStr,
		Notes: []string{
			"abbr branch and expan branch of every <choice> both kept, line-aligned",
			"teiHeader, note, fw, label subtrees excluded as apparatus",
			"div[@type=toc] subtree excluded as apparatus, not running text",
			"<lb/> becomes a line break in both streams",
			"every <g> becomes one reserved Glagolitic placeholder rune per distinct ref, identically in both streams outside <choice>, and inside <abbr> only",
		},
	}, nil
}

func tokenizeLines(s string) [][]string {
	lines := strings.Split(s, "\n")
	out := make([][]string, len(lines))
	for i, l := range lines {
		out[i] = strings.Fields(l)
	}
	return out
}

func collectGRefs(path string, refs map[string]bool) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()
	dec := xml.NewDecoder(f)
	dec.Strict = false
	for {
		tok, err := dec.Token()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}
		se, ok := tok.(xml.StartElement)
		if !ok || se.Name.Local != "g" {
			continue
		}
		ref := attrTEI(se, "ref")
		if ref == "" {
			ref = "#unknown"
		}
		refs[ref] = true
	}
}

func attrTEI(se xml.StartElement, local string) string {
	for _, a := range se.Attr {
		if a.Name.Local == local {
			return a.Value
		}
	}
	return ""
}

type teiFrame struct {
	name string
	skip bool
}

// extractPairFile walks one TEI file and emits both streams plus this
// file's PairUnits. startOrder/startAbbrLine let multi-file callers keep
// a single running order/line count across files.
func extractPairFile(path string, placeholder map[string]rune, startOrder, startAbbrLine, startExpanLine int) (abbrText, expanText string, pairs []PairUnit, endAbbrLine, endExpanLine int, err error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return "", "", nil, 0, 0, err
	}
	dec := xml.NewDecoder(strings.NewReader(string(raw)))
	dec.Strict = false
	var abbrOut, expanOut strings.Builder
	var stack []teiFrame
	inBody := false
	inChoice := false
	choiceSkip := false // ambient skip state (e.g. inside <fw>) captured at <choice> open
	var choiceAbbr, choiceExpan strings.Builder
	choiceHasMark := false
	choiceMarkCombining := false
	order := startOrder
	abbrLine, expanLine := startAbbrLine, startExpanLine

	flushChoice := func() {
		if !choiceSkip {
			pairs = append(pairs, PairUnit{
				File: path, Order: order, Line: abbrLine,
				AbbrText: choiceAbbr.String(), ExpanText: choiceExpan.String(),
				HasMark: choiceHasMark, MarkIsCombining: choiceMarkCombining,
			})
			order++
			abbrOut.WriteString(choiceAbbr.String())
			expanOut.WriteString(choiceExpan.String())
		}
		choiceAbbr.Reset()
		choiceExpan.Reset()
		choiceHasMark = false
		choiceMarkCombining = false
		choiceSkip = false
	}

	skipNow := func() bool { return len(stack) > 0 && stack[len(stack)-1].skip }
	// depth within <choice>: 0 = not in choice; tracks which branch (abbr/expan) is open
	var branch string // "", "abbr", "expan"

	for {
		tok, terr := dec.Token()
		if terr == io.EOF {
			break
		}
		if terr != nil {
			return "", "", nil, 0, 0, terr
		}
		switch t := tok.(type) {
		case xml.StartElement:
			name := t.Name.Local
			if name == "body" {
				inBody = true
			}
			skip := skipNow()
			switch {
			case skipSubtreeTEI[name]:
				skip = true
			case name == "div" && attrTEI(t, "type") == "toc":
				skip = true
			case name == "choice":
				inChoice = true
				choiceSkip = skip || !inBody
			case name == "abbr" && inChoice:
				branch = "abbr"
			case name == "expan" && inChoice:
				branch = "expan"
			case name == "lb":
				// A <lb/> occurring inside <abbr> is a mid-word manuscript
				// line break on the diplomatic surface only; the edition's
				// <expan> branch is continuous normalized text with no
				// corresponding break, so it is not mirrored there (the two
				// streams are allowed to disagree on line count, task82b.txt
				// sec.12: n:1/1:n/n:m, not forced 1:1).
				if !skip && inBody {
					switch {
					case inChoice:
						// Whether inside <abbr>, between branches, or before
						// branch is set, a break here only affects the
						// abbreviated/diplomatic surface.
						choiceAbbr.WriteString("\n")
						abbrLine++
					default:
						abbrOut.WriteString("\n")
						expanOut.WriteString("\n")
						abbrLine++
						expanLine++
					}
				}
			case name == "g":
				if !skip && inBody {
					ref := attrTEI(t, "ref")
					if ref == "" {
						ref = "#unknown"
					}
					r, ok := placeholder[ref]
					if !ok {
						return "", "", nil, 0, 0, fmt.Errorf("unmapped <g ref=%q>", ref)
					}
					if inChoice && branch == "abbr" {
						choiceAbbr.WriteRune(r)
						choiceHasMark = true
						if isCombiningMarkRef(ref) {
							choiceMarkCombining = true
						}
					} else if inChoice && branch == "expan" {
						choiceExpan.WriteRune(r)
					} else if !inChoice {
						abbrOut.WriteRune(r)
						expanOut.WriteRune(r)
					}
				}
				skip = true
			}
			stack = append(stack, teiFrame{name: name, skip: skip})
		case xml.EndElement:
			name := t.Name.Local
			if len(stack) > 0 {
				stack = stack[:len(stack)-1]
			}
			switch name {
			case "body":
				inBody = false
			case "abbr", "expan":
				branch = ""
			case "choice":
				inChoice = false
				flushChoice()
			}
		case xml.CharData:
			if !inBody || skipNow() {
				continue
			}
			if inChoice {
				switch branch {
				case "abbr":
					choiceAbbr.Write(t)
				case "expan":
					choiceExpan.Write(t)
				}
			} else {
				abbrOut.Write(t)
				expanOut.Write(t)
			}
		}
	}
	return abbrOut.String(), expanOut.String(), pairs, abbrLine, expanLine, nil
}

// isCombiningMarkRef classifies a TEI <g ref="#char-XXXX"> value as a
// Unicode combining-diacritical-mark codepoint (U+0300-U+036F, the
// standard superscript abbreviation-bar/tilde/macron range) versus a
// dedicated precomposed special-sign letterform (task82b.txt sec.13:
// "superscript" vs "special sign", named from the data, not imposed).
func isCombiningMarkRef(ref string) bool {
	const prefix = "#char-"
	if !strings.HasPrefix(ref, prefix) {
		return false
	}
	hex := strings.TrimPrefix(ref, prefix)
	var v int64
	for _, c := range hex {
		d := hexDigit(c)
		if d < 0 {
			return false
		}
		v = v*16 + int64(d)
	}
	return v >= 0x0300 && v <= 0x036F
}

func hexDigit(c rune) int {
	switch {
	case c >= '0' && c <= '9':
		return int(c - '0')
	case c >= 'a' && c <= 'f':
		return int(c-'a') + 10
	case c >= 'A' && c <= 'F':
		return int(c-'A') + 10
	default:
		return -1
	}
}
