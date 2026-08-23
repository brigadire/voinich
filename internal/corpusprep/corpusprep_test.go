package corpusprep

import (
	"bytes"
	"io"
	"reflect"
	"strings"
	"testing"
	"unicode/utf8"

	"golang.org/x/text/encoding/charmap"
	"golang.org/x/text/transform"
)

func TestPrepareSupportsUTF8AndLegacyEncodings(t *testing.T) {
	cases := []struct {
		name     string
		encoding string
		input    []byte
		wantText string
	}{
		{
			name:     "utf8",
			encoding: EncodingUTF8,
			input:    []byte("Привет, мир!\n"),
			wantText: "привет мир\n",
		},
		{
			name:     "cp1251",
			encoding: EncodingCP1251,
			input:    mustEncode(t, charmap.Windows1251.NewEncoder(), "Привет, мир!\n"),
			wantText: "привет мир\n",
		},
		{
			name:     "koi8r",
			encoding: EncodingKOI8R,
			input:    mustEncode(t, charmap.KOI8R.NewEncoder(), "Привет, мир!\n"),
			wantText: "привет мир\n",
		},
		{
			name:     "cp866",
			encoding: EncodingCP866,
			input:    mustEncode(t, charmap.CodePage866.NewEncoder(), "Привет, мир!\n"),
			wantText: "привет мир\n",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			res, manifest, err := Prepare(tc.input, Options{Encoding: tc.encoding, CasePolicy: CaseLower, LinePolicy: LinePreserve}, "deadbeef", "input.txt", "prepared.txt")
			if err != nil {
				t.Fatal(err)
			}
			if string(res.Text) != tc.wantText {
				t.Fatalf("text = %q, want %q", res.Text, tc.wantText)
			}
			if !utf8.Valid(res.Text) {
				t.Fatal("output is not utf-8")
			}
			check, err := Check(res.Text)
			if err != nil {
				t.Fatal(err)
			}
			if !check.Valid {
				t.Fatalf("prepared output rejected: %+v", check)
			}
			if manifest.ReplacementCharCount != 0 || manifest.InvalidUTF8Count != 0 || manifest.ForbiddenControlCount != 0 {
				t.Fatalf("unexpected manifest safety counts: %+v", manifest)
			}
			if manifest.OutputEncoding != EncodingUTF8 || manifest.CanonicalCorpusVersion != CanonicalCorpusVersion {
				t.Fatalf("manifest canonical metadata mismatch: %+v", manifest)
			}
		})
	}
}

func TestPreparePunctuationAndWhitespaceNormalization(t *testing.T) {
	input := []byte("Word,word 2—3 well-known \"text\" (abc) 1/2\r\nA\u00A0B\tC…D\r\n")
	res, _, err := Prepare(input, Options{Encoding: EncodingUTF8, CasePolicy: CaseLower, LinePolicy: LinePreserve}, "deadbeef", "input.txt", "prepared.txt")
	if err != nil {
		t.Fatal(err)
	}
	want := "word word 2 3 well known text abc 1 2\na b c d\n"
	if string(res.Text) != want {
		t.Fatalf("normalize = %q, want %q", res.Text, want)
	}
	if res.Corpus.Occurrences != 14 || res.Corpus.NonEmpty != 2 || res.Corpus.Transitions != 12 {
		t.Fatalf("unexpected corpus stats: %+v", res.Corpus)
	}
}

func TestPrepareLinePolicyReflow(t *testing.T) {
	input := []byte("Alpha beta\nGamma delta\n")
	res, _, err := Prepare(input, Options{Encoding: EncodingUTF8, CasePolicy: CaseLower, LinePolicy: LineReflow}, "deadbeef", "input.txt", "prepared.txt")
	if err != nil {
		t.Fatal(err)
	}
	if string(res.Text) != "alpha beta gamma delta\n" {
		t.Fatalf("reflow output = %q", res.Text)
	}
	if res.Corpus.NonEmpty != 1 || len(res.Corpus.Lines) != 1 {
		t.Fatalf("reflow corpus shape changed: %+v", res.Corpus)
	}
}

func TestPrepareRejectsUnsafeInput(t *testing.T) {
	cases := map[string][]byte{
		"utf8-invalid": []byte{0xff, 0xfe, 0xfd},
		"nul":          []byte("a\x00b\n"),
		"control":      []byte("a\x1fb\n"),
		"fffd":         []byte("a�b\n"),
	}
	for name, input := range cases {
		t.Run(name, func(t *testing.T) {
			if _, _, err := Prepare(input, Options{Encoding: EncodingUTF8, CasePolicy: CaseLower, LinePolicy: LinePreserve}, "deadbeef", "input.txt", "prepared.txt"); err == nil {
				t.Fatal("unsafe input was accepted")
			}
		})
	}
}

func TestPrepareDeterministic(t *testing.T) {
	input := []byte("alpha, beta\ngamma\n")
	first, manifest1, err := Prepare(input, Options{Encoding: EncodingUTF8, CasePolicy: CaseLower, LinePolicy: LinePreserve}, "deadbeef", "input.txt", "prepared.txt")
	if err != nil {
		t.Fatal(err)
	}
	second, manifest2, err := Prepare(input, Options{Encoding: EncodingUTF8, CasePolicy: CaseLower, LinePolicy: LinePreserve}, "deadbeef", "input.txt", "prepared.txt")
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(first.Text, second.Text) {
		t.Fatal("output bytes changed across identical runs")
	}
	if !reflect.DeepEqual(manifest1, manifest2) {
		t.Fatalf("manifest changed across identical runs:\n%+v\n%+v", manifest1, manifest2)
	}
}

func TestLegacyCP1251NeedsExplicitEncoding(t *testing.T) {
	rawPath := "fixture.cp1251"
	raw := mustEncode(t, charmap.Windows1251.NewEncoder(), "Пример текста\n")
	if _, _, err := Prepare(raw, Options{Encoding: EncodingUTF8, CasePolicy: CasePreserve, LinePolicy: LinePreserve}, "deadbeef", rawPath, "prepared.txt"); err == nil {
		t.Fatal("raw CP1251 fixture was accepted as UTF-8")
	}
	prepared, manifest, err := Prepare(raw, Options{Encoding: EncodingCP1251, CasePolicy: CasePreserve, LinePolicy: LinePreserve}, "deadbeef", rawPath, "prepared.txt")
	if err != nil {
		t.Fatalf("cp1251 prepare failed: %v", err)
	}
	if !utf8.Valid(prepared.Text) {
		t.Fatal("prepared CP1251 output is not utf-8")
	}
	check, err := Check(prepared.Text)
	if err != nil {
		t.Fatal(err)
	}
	if !check.Valid {
		t.Fatalf("prepared CP1251 fixture failed check: %+v", check)
	}
	if manifest.ReplacementCharCount != 0 || manifest.InvalidUTF8Count != 0 {
		t.Fatalf("CP1251 manifest indicates encoding loss: %+v", manifest)
	}
	if _, _, err := Prepare(raw, Options{Encoding: EncodingAuto, CasePolicy: CasePreserve, LinePolicy: LinePreserve}, "deadbeef", rawPath, "prepared.txt"); err == nil {
		t.Fatal("auto detection accepted a non-utf8 legacy corpus")
	}
}

func mustEncode(t *testing.T, encoder transform.Transformer, text string) []byte {
	t.Helper()
	data, err := io.ReadAll(transform.NewReader(strings.NewReader(text), encoder))
	if err != nil {
		t.Fatal(err)
	}
	return data
}
