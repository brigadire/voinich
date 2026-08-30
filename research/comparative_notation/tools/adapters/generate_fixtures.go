//go:build ignore

// This one-shot deterministic generator creates inspectable adapter fixtures.
// It is not a production corpus generator.
package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"zcore.dev/voinich/internal/notation"
)

type fixture struct{ class, rep string }

func main() {
	fixtures := []fixture{{"C01", "LATIN-DIPLOMATIC"}, {"C01", "LATIN-EXPANDED"}, {"C02", "LATIN-DIPLOMATIC"}, {"C02", "LATIN-EXPANDED"}, {"C03", "SHORTHAND-DIPLOMATIC"}, {"C03", "SHORTHAND-EXPANDED"}, {"C04", "CIPHER-SIGNS"}, {"C04", "CIPHER-PLAINTEXT"}, {"C05", "CIPHER-SIGNS"}, {"C06", "MUSIC-R1"}, {"C06", "MUSIC-R2"}, {"C06", "MUSIC-R3"}, {"C07", "TAB-R1"}, {"C07", "TAB-R2"}, {"C08", "NUMERIC-RECORD"}, {"C09", "TABLE-CELL"}, {"C10", "SYNTHETIC-TOKEN"}}
	root := "research/comparative_notation/tools/adapters/fixtures"
	for _, f := range fixtures {
		if err := write(root, f); err != nil {
			panic(err)
		}
	}
}

func write(root string, f fixture) error {
	dir := filepath.Join(root, f.class+"_"+strings.ReplaceAll(f.rep, "-", "_"))
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	src := notation.SourceDocument{CorpusID: f.class + "-FIXTURE", ClassID: f.class, Representation: f.rep, DocumentID: "doc-1"}
	lex := [][]string{{"a", "b"}, {"a", "c"}, {"b", "a", "d"}, {"c", "a"}, {"d", "b", "a"}, {"a", "d"}}
	for i := 0; i < 24; i++ {
		sym := append([]string(nil), lex[i%len(lex)]...)
		for extra := 0; extra < i/8; extra++ {
			sym = append(sym, "d")
		}
		attrs := map[string]string{"fixture_unit": fmt.Sprint(i)}
		if f.class == "C06" {
			attrs["voice"] = "v" + fmt.Sprint(i%2+1)
			attrs["staff"] = "s" + fmt.Sprint(i%2+1)
			attrs["system"] = "sys" + fmt.Sprint(i/6+1)
			attrs["simultaneity_group"] = "g" + fmt.Sprint(i/2)
			switch f.rep {
			case "MUSIC-R1":
				attrs["event"] = "note"
			case "MUSIC-R2":
				attrs["interval"] = fmt.Sprintf("%+d", i%5-2)
			case "MUSIC-R3":
				attrs["pitch"] = fmt.Sprint(60 + i%7)
				attrs["duration"] = []string{"brevis", "semibrevis", "minima"}[i%3]
			}
		}
		if f.class == "C07" {
			attrs["tradition"] = "French"
			attrs["simultaneity_group"] = "g" + fmt.Sprint(i/2)
			attrs["course"] = fmt.Sprint(i%6 + 1)
			attrs["fret"] = string(rune('a' + i%5))
			attrs["rhythm"] = "semiminim"
		}
		src.Units = append(src.Units, notation.SourceUnit{Section: "sec-" + fmt.Sprint(i/12+1), Page: "p-" + fmt.Sprint(i/8+1), Locus: "loc-" + fmt.Sprint(i/6+1), PhysicalLine: "line-" + fmt.Sprint(i/4+1), Token: strings.Join(sym, ""), Symbols: sym, Attributes: attrs})
	}
	var source bytes.Buffer
	e := json.NewEncoder(&source)
	e.SetIndent("", "  ")
	if err := e.Encode(src); err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(dir, "source.json"), source.Bytes(), 0644); err != nil {
		return err
	}
	recs, err := notation.NormalizeFixture(src)
	if err != nil {
		return err
	}
	var usc bytes.Buffer
	if err := notation.WriteJSONL(&usc, recs); err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(dir, "expected.usc.jsonl"), usc.Bytes(), 0644); err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(dir, "generated.usc.jsonl"), usc.Bytes(), 0644); err != nil {
		return err
	}
	manifest := map[string]any{"class_id": f.class, "representation_id": f.rep, "source_units": 24, "manual_inspection": "REQUIRED_BEFORE_PRODUCTION", "generated_by": "generate_fixtures.go"}
	var mb bytes.Buffer
	me := json.NewEncoder(&mb)
	me.SetIndent("", "  ")
	_ = me.Encode(manifest)
	return os.WriteFile(filepath.Join(dir, "FIXTURE_MANIFEST.json"), mb.Bytes(), 0644)
}
