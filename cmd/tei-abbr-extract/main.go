// Command tei-abbr-extract produces a deterministic diplomatic (as-written)
// plaintext stream from one or more TEI-XML manuscript transcription files
// that mark scribal abbreviations as
// <choice><abbr>...</abbr><expan>...</expan></choice>.
//
// It is a Task79c Gate B data-preparation step: it does not compute any
// Fingerprint v2 metric. It only turns third-party TEI-XML bytes (never
// committed to this repository; see DATA.md) into a plain corpus file that
// internal/fingerprintv2 can load with glyph_mode=natural, plus a
// PrepareManifest sidecar produced by the existing internal/corpusprep
// pipeline so the output has the same provenance/checksum shape as every
// other prepared corpus in this repository.
//
// For every <choice>, only the <abbr> branch is kept (the manuscript's own
// abbreviated surface form); the <expan> branch is dropped. teiHeader,
// note, the chapter table-of-contents (<div type="toc">), running page
// headers/footers (<fw>) and marginal chapter-number labels (<label>) are
// excluded as apparatus, not scribal running text. <lb/> becomes a line
// break, matching the manuscript's own line segmentation.
//
// Every <g ref="..."/> element (a combining abbreviation mark or a
// Private-Use-Area glyph with no literal Unicode assignment) is replaced by
// one dedicated placeholder rune per distinct ref value, taken from the
// Unicode Glagolitic block. This is necessary, not cosmetic: downstream
// internal/evaglyph.NaturalGlyphs keeps only unicode.IsLetter /
// unicode.IsNumber runes, so a bare combining mark or PUA codepoint would
// silently vanish and erase exactly the abbreviation signal this control
// exists to carry. The ref->placeholder table is written to a sidecar next
// to the output for auditability.
package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"encoding/xml"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"sort"
	"strings"

	"zcore.dev/voinich/internal/corpusprep"
)

// skipSubtree names apparatus elements whose entire subtree (including any
// nested <choice>/<g>) contributes no text to the diplomatic stream.
var skipSubtree = map[string]bool{
	"teiHeader": true,
	"note":      true,
	"fw":        true,
	"label":     true,
}

// glagoliticBase is the first codepoint of the Unicode Glagolitic block
// (U+2C00), used as a reserved, never-otherwise-occurring Letter alphabet
// for abbreviation-mark placeholders. Glagolitic runes satisfy
// unicode.IsLetter, so they survive internal/evaglyph.NaturalGlyphs.
const glagoliticBase = rune(0x2C00)

// glagoliticSize is the number of assignable code points before the block
// reaches its combining/punctuation tail; ample headroom over any observed
// TEI <g> ref vocabulary.
const glagoliticSize = 90

type extractStats struct {
	Path        string `json:"path"`
	SHA256      string `json:"sha256"`
	SizeBytes   int64  `json:"size_bytes"`
	AbbrCount   int    `json:"abbr_count"`
	ExpanCount  int    `json:"expan_count"`
	GlyphMarks  int    `json:"glyph_mark_count"`
	LineBreaks  int    `json:"line_break_count"`
	SkippedTOC  int    `json:"skipped_toc_divs"`
	OutputRunes int    `json:"output_rune_count"`
}

type manifest struct {
	SchemaVersion     int               `json:"schema_version"`
	Tool              string            `json:"tool"`
	ToolGitCommit     string            `json:"tool_git_commit"`
	Inputs            []extractStats    `json:"inputs"`
	GlyphPlaceholders map[string]string `json:"glyph_placeholders"`
	OutputPath        string            `json:"diplomatic_output_path"`
	OutputSHA256      string            `json:"diplomatic_output_sha256"`
	Notes             []string          `json:"notes"`
}

func main() {
	os.Exit(run(os.Args[1:]))
}

func run(args []string) int {
	fs := flag.NewFlagSet("tei-abbr-extract", flag.ContinueOnError)
	diplomaticOut := fs.String("diplomatic-output", "", "path for the intermediate diplomatic plaintext (required)")
	preparedOut := fs.String("prepared-output", "", "path for the corpusprep-normalized corpus (required)")
	manifestOut := fs.String("manifest-output", "", "path for the extraction manifest JSON (default: <prepared-output>.extract.json)")
	if err := fs.Parse(args); err != nil {
		return 2
	}
	inputs := fs.Args()
	if len(inputs) == 0 || *diplomaticOut == "" || *preparedOut == "" {
		fmt.Fprintln(os.Stderr, "usage: tei-abbr-extract -diplomatic-output FILE -prepared-output FILE [-manifest-output FILE] INPUT.xml [INPUT.xml ...]")
		return 2
	}
	if *manifestOut == "" {
		*manifestOut = *preparedOut + ".extract.json"
	}
	sorted := append([]string(nil), inputs...)
	sort.Strings(sorted)

	refs := map[string]bool{}
	for _, p := range sorted {
		if err := collectRefs(p, refs); err != nil {
			fmt.Fprintf(os.Stderr, "Error: collect refs from %s: %v\n", p, err)
			return 1
		}
	}
	sortedRefs := make([]string, 0, len(refs))
	for r := range refs {
		sortedRefs = append(sortedRefs, r)
	}
	sort.Strings(sortedRefs)
	if len(sortedRefs) > glagoliticSize {
		fmt.Fprintf(os.Stderr, "Error: %d distinct <g ref> values exceeds reserved placeholder alphabet size %d\n", len(sortedRefs), glagoliticSize)
		return 1
	}
	placeholder := map[string]rune{}
	placeholderStr := map[string]string{}
	for i, r := range sortedRefs {
		p := glagoliticBase + rune(i)
		placeholder[r] = p
		placeholderStr[r] = string(p)
	}

	var out strings.Builder
	var stats []extractStats
	for i, p := range sorted {
		text, st, err := extractFile(p, placeholder)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: extract %s: %v\n", p, err)
			return 1
		}
		out.WriteString(text)
		if i != len(sorted)-1 {
			out.WriteString("\n")
		}
		stats = append(stats, st)
	}

	diplomatic := []byte(out.String())
	if err := os.WriteFile(*diplomaticOut, diplomatic, 0644); err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		return 1
	}

	commit, _ := gitCommit(".")
	result, prepManifest, err := corpusprep.Prepare(diplomatic, corpusprep.Options{
		Encoding:       corpusprep.EncodingUTF8,
		CasePolicy:     corpusprep.CaseLower,
		LinePolicy:     corpusprep.LinePreserve,
		DropEmptyLines: true,
	}, commit, *diplomaticOut, *preparedOut)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error: corpusprep.Prepare:", err)
		return 1
	}
	if err := os.WriteFile(*preparedOut, result.Text, 0644); err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		return 1
	}
	prepManifest.OutputPath = *preparedOut
	prepData, err := json.MarshalIndent(prepManifest, "", "  ")
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		return 1
	}
	prepData = append(prepData, '\n')
	if err := os.WriteFile(*preparedOut+".prepare.json", prepData, 0644); err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		return 1
	}

	outSHA := sha256Hex(diplomatic)
	m := manifest{
		SchemaVersion:     1,
		Tool:              "tei-abbr-extract",
		ToolGitCommit:     commit,
		Inputs:            stats,
		GlyphPlaceholders: placeholderStr,
		OutputPath:        *diplomaticOut,
		OutputSHA256:      outSHA,
		Notes: []string{
			"abbr branch of <choice> kept; expan branch dropped",
			"teiHeader, note, fw, label subtrees excluded as apparatus",
			"div[@type=toc] subtree excluded as apparatus, not running text",
			"<lb/> becomes a line break",
			"every <g> becomes one reserved Glagolitic placeholder rune per distinct ref, regardless of literal fallback content, so evaglyph.NaturalGlyphs cannot silently drop it",
		},
	}
	mData, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		return 1
	}
	mData = append(mData, '\n')
	if err := os.WriteFile(*manifestOut, mData, 0644); err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		return 1
	}
	fmt.Printf("Wrote %s\nWrote %s\nWrote %s\nWrote %s\nDistinct <g> refs: %d\nPrepared tokens: %d\n",
		*diplomaticOut, *preparedOut, *preparedOut+".prepare.json", *manifestOut, len(sortedRefs), prepManifest.OutputTokenCount)
	return 0
}

func collectRefs(path string, refs map[string]bool) error {
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
		ref := attr(se, "ref")
		if ref == "" {
			ref = "#unknown"
		}
		refs[ref] = true
	}
}

func attr(se xml.StartElement, local string) string {
	for _, a := range se.Attr {
		if a.Name.Local == local {
			return a.Value
		}
	}
	return ""
}

// frame tracks one open element for skip-subtree bookkeeping.
type frame struct {
	name string
	skip bool
}

func extractFile(path string, placeholder map[string]rune) (string, extractStats, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return "", extractStats{}, err
	}
	st := extractStats{Path: path, SHA256: sha256Hex(raw), SizeBytes: int64(len(raw))}

	dec := xml.NewDecoder(strings.NewReader(string(raw)))
	dec.Strict = false
	var out strings.Builder
	var stack []frame
	inBody := false

	skipNow := func() bool {
		return len(stack) > 0 && stack[len(stack)-1].skip
	}

	for {
		tok, err := dec.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return "", extractStats{}, err
		}
		switch t := tok.(type) {
		case xml.StartElement:
			name := t.Name.Local
			if name == "body" {
				inBody = true
			}
			skip := skipNow()
			switch {
			case skipSubtree[name]:
				skip = true
			case name == "div" && attr(t, "type") == "toc":
				skip = true
				st.SkippedTOC++
			case name == "expan":
				skip = true
				st.ExpanCount++
			case name == "abbr":
				st.AbbrCount++
			case name == "lb":
				if !skip && inBody {
					out.WriteString("\n")
					st.LineBreaks++
				}
			case name == "g":
				if !skip && inBody {
					ref := attr(t, "ref")
					if ref == "" {
						ref = "#unknown"
					}
					r, ok := placeholder[ref]
					if !ok {
						return "", extractStats{}, fmt.Errorf("unmapped <g ref=%q>", ref)
					}
					out.WriteRune(r)
					st.GlyphMarks++
				}
				skip = true // never descend into <g> children; the placeholder already stands for it
			}
			stack = append(stack, frame{name: name, skip: skip})
		case xml.EndElement:
			if len(stack) > 0 {
				stack = stack[:len(stack)-1]
			}
			if t.Name.Local == "body" {
				inBody = false
			}
		case xml.CharData:
			if inBody && !skipNow() {
				out.Write(t)
			}
		}
	}
	text := out.String()
	st.OutputRunes = len([]rune(text))
	return text, st, nil
}

func sha256Hex(b []byte) string {
	sum := sha256.Sum256(b)
	return hex.EncodeToString(sum[:])
}

func gitCommit(repoPath string) (string, bool) {
	out, err := exec.Command("git", "-C", repoPath, "rev-parse", "HEAD").Output()
	if err != nil {
		return "", false
	}
	return strings.TrimSpace(string(out)), false
}
