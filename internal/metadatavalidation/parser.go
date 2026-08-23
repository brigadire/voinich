package metadatavalidation

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strings"
)

var (
	pageRE           = regexp.MustCompile(`^<([^.,>]+)>\s*(?:<!\s*(.*?)>)?\s*$`)
	locusRE          = regexp.MustCompile(`^<([^,>]+),([^>]+)>\s*(.*)$`)
	variableRE       = regexp.MustCompile(`\$([A-Za-z])=([^\s>]+)`)
	inlineVariableRE = regexp.MustCompile(`<@([A-Za-z])=([^>]+)>`)
)

// ParseIVTFF reads only page/locus structure, page variables and locus text.
// It deliberately is not an IVTFF-to-corpus converter.
func ParseIVTFF(path string) (Document, error) {
	f, err := os.Open(path)
	if err != nil {
		return Document{}, err
	}
	defer f.Close()
	return ParseIVTFFReader(f)
}

func ParseIVTFFReader(r ioReader) (Document, error) {
	var d Document
	d.PageVariables = map[string]map[string]string{}
	folio := ""
	vars := map[string]string{}
	paragraph := 0
	s := bufio.NewScanner(r)
	s.Buffer(make([]byte, 64*1024), 1024*1024)
	lineNo := 0
	for s.Scan() {
		lineNo++
		line := strings.TrimSpace(strings.TrimSuffix(s.Text(), "\r"))
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		if m := pageRE.FindStringSubmatch(line); m != nil {
			folio = m[1]
			vars = map[string]string{}
			for _, x := range variableRE.FindAllStringSubmatch(m[2], -1) {
				vars[x[1]] = x[2]
			}
			d.PageVariables[folio] = cloneStrings(vars)
			d.Pages++
			paragraph = 0
			continue
		}
		m := locusRE.FindStringSubmatch(line)
		if m == nil {
			continue
		} // IVTFF declarations and comments are metadata, not text.
		if folio == "" {
			return Document{}, fmt.Errorf("line %d: locus before page", lineNo)
		}
		id, code, raw := m[1], m[2], m[3]
		for _, x := range inlineVariableRE.FindAllStringSubmatch(raw, -1) {
			vars[x[1]] = x[2]
		}
		start := strings.HasPrefix(code, "@P") || strings.HasPrefix(code, "*P") || strings.Contains(raw, "<%>")
		if start || paragraph == 0 {
			paragraph++
		}
		typ := ""
		if len(code) >= 2 {
			typ = string(code[1])
		}
		l := Locus{Folio: folio, ID: id, Type: typ, LineID: id, RawText: raw,
			AlignmentText: NormalizeForAlignment(raw), ParagraphID: paragraph,
			ParagraphStart: start, Variables: cloneStrings(vars)}
		if strings.TrimSpace(l.AlignmentText) == "" {
			d.SkippedLoci++
		}
		d.Loci = append(d.Loci, l)
	}
	if err := s.Err(); err != nil {
		return Document{}, err
	}
	return d, nil
}

// tiny interface keeps parser tests independent of the filesystem.
type ioReader interface{ Read([]byte) (int, error) }

func cloneStrings(x map[string]string) map[string]string {
	y := make(map[string]string, len(x))
	for k, v := range x {
		y[k] = v
	}
	return y
}

// NormalizeForAlignment implements the minimal observed -x7 representation:
// comments/control markers disappear; alternatives select their first branch;
// braces retain their contents; dot, comma, apostrophe, question mark and
// physical gaps are all word breaks (confirmed against the real ZL3b-x7
// canonical corpus during task77's audit: it contains zero apostrophes and
// zero question marks, and every position where the raw source has one is
// a token boundary in the canonical output). @NNN; remains ordinary content.
func NormalizeForAlignment(raw string) string {
	s := raw
	s = inlineVariableRE.ReplaceAllString(s, "")
	for {
		a := strings.Index(s, "<!")
		if a < 0 {
			break
		}
		b := strings.Index(s[a+2:], ">")
		if b < 0 {
			s = s[:a]
			break
		}
		s = s[:a] + s[a+2+b+1:]
	}
	s = strings.ReplaceAll(s, "<%>", "")
	s = strings.ReplaceAll(s, "<$>", "")
	s = strings.ReplaceAll(s, "<->", " ")
	s = strings.ReplaceAll(s, "<~>", " ")
	// Alternatives do not nest in the source except for braces inside a branch.
	for {
		a := strings.IndexByte(s, '[')
		if a < 0 {
			break
		}
		b := strings.IndexByte(s[a+1:], ']')
		if b < 0 {
			break
		}
		b += a + 1
		inside := s[a+1 : b]
		depth, cut := 0, -1
		for i, r := range inside {
			switch r {
			case '{':
				depth++
			case '}':
				if depth > 0 {
					depth--
				}
			case ':':
				if depth == 0 {
					cut = i
				}
			}
			if cut >= 0 {
				break
			}
		}
		if cut >= 0 {
			inside = inside[:cut]
		}
		s = s[:a] + inside + s[b+1:]
	}
	// An apostrophe is the transcriber's own probable-word-break marker
	// (observed only inside {...} uncertain-reading braces in ZL3b-n.txt,
	// e.g. "{c'y}" -> "c y"); a bare '?' (illegible glyph) is also a word
	// break in the real -x7 canonical output (e.g. "d?n" -> "d n"); and a
	// bare in-text "@NNN;" marker (not wrapped in <!...>, which is already
	// stripped above as a comment) splits into its own numeric token, e.g.
	// "@192;chy" -> "192 chy". All four are confirmed against the
	// historical -x7 canonical conversion, which contains none of these
	// four characters at all.
	s = strings.NewReplacer("{", "", "}", "", ".", " ", ",", " ", "=", " ", "-", " ", "'", " ", "?", " ", "@", " ", ";", " ").Replace(s)
	return strings.Join(strings.Fields(s), " ")
}
