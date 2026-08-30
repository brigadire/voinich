package main

import (
	"archive/tar"
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"encoding/xml"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"zcore.dev/voinich/internal/notation"
)

type musicXMLScore struct {
	Parts []musicXMLPart `xml:"part"`
}

type musicXMLPart struct {
	ID       string            `xml:"id,attr"`
	Measures []musicXMLMeasure `xml:"measure"`
}

type musicXMLMeasure struct {
	Number string
	Items  []musicXMLItem
}

type musicXMLItem struct {
	Kind      string
	Divisions int
	Duration  int
	Print     musicXMLPrint
	Note      musicXMLNote
}

type musicXMLPrint struct {
	NewPage   string `xml:"new-page,attr"`
	NewSystem string `xml:"new-system,attr"`
}

type musicXMLNote struct {
	Chord      *struct{}  `xml:"chord"`
	Rest       *struct{}  `xml:"rest"`
	Duration   int        `xml:"duration"`
	Voice      string     `xml:"voice"`
	Type       string     `xml:"type"`
	Dots       []struct{} `xml:"dot"`
	Staff      string     `xml:"staff"`
	Accidental string     `xml:"accidental"`
	Ties       []struct {
		Type string `xml:"type,attr"`
	} `xml:"tie"`
	Pitch struct {
		Step   string `xml:"step"`
		Alter  string `xml:"alter"`
		Octave int    `xml:"octave"`
	} `xml:"pitch"`
}

func (m *musicXMLMeasure) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	for _, a := range start.Attr {
		if a.Name.Local == "number" {
			m.Number = a.Value
		}
	}
	for {
		tok, err := d.Token()
		if err != nil {
			return err
		}
		switch t := tok.(type) {
		case xml.StartElement:
			switch t.Name.Local {
			case "print":
				var v musicXMLPrint
				if err := d.DecodeElement(&v, &t); err != nil {
					return err
				}
				m.Items = append(m.Items, musicXMLItem{Kind: "print", Print: v})
			case "attributes":
				var v struct {
					Divisions int `xml:"divisions"`
				}
				if err := d.DecodeElement(&v, &t); err != nil {
					return err
				}
				m.Items = append(m.Items, musicXMLItem{Kind: "attributes", Divisions: v.Divisions})
			case "note":
				var v musicXMLNote
				if err := d.DecodeElement(&v, &t); err != nil {
					return err
				}
				m.Items = append(m.Items, musicXMLItem{Kind: "note", Note: v})
			case "backup", "forward":
				var v struct {
					Duration int `xml:"duration"`
				}
				if err := d.DecodeElement(&v, &t); err != nil {
					return err
				}
				m.Items = append(m.Items, musicXMLItem{Kind: t.Name.Local, Duration: v.Duration})
			default:
				if err := d.Skip(); err != nil {
					return err
				}
			}
		case xml.EndElement:
			if t.Name == start.Name {
				return nil
			}
		}
	}
}

type musicEvent struct {
	Document, Part, Voice, Staff, Measure, Page, System string
	MeasureIndex, OnsetNum, OnsetDen, SourceOrdinal     int
	Rest, Chord                                         bool
	Step, Alter, Accidental                             string
	Octave                                              int
	DurationNum, DurationDen                            int
	NoteType                                            string
	Dots                                                int
	Ties                                                []string
}

func musicXMLUSCCmd(args []string) error {
	fs := flag.NewFlagSet("musicxml-usc", flag.ContinueOnError)
	inName := fs.String("input", "", "deterministic tar.gz containing MusicXML")
	outDir := fs.String("output-dir", "", "output directory for MUSIC-R1/R2/R3 USC")
	corpusID := fs.String("corpus-id", "C06-JLSDD-SECURE", "frozen corpus identifier")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *inName == "" || *outDir == "" {
		return fmt.Errorf("musicxml-usc requires --input and --output-dir")
	}
	files, err := readMusicXMLArchive(*inName)
	if err != nil {
		return err
	}
	var events []musicEvent
	for _, name := range sortedKeys(files) {
		var score musicXMLScore
		if err := xml.Unmarshal(files[name], &score); err != nil {
			return fmt.Errorf("%s: %w", name, err)
		}
		docEvents, err := extractMusicEvents(filepath.Base(name), score)
		if err != nil {
			return fmt.Errorf("%s: %w", name, err)
		}
		events = append(events, docEvents...)
	}
	if len(files) == 0 || len(events) == 0 {
		return fmt.Errorf("archive contains no usable MusicXML")
	}
	for _, rep := range []string{"MUSIC-R1", "MUSIC-R2", "MUSIC-R3"} {
		records, err := musicRecords(*corpusID, rep, events)
		if err != nil {
			return err
		}
		path := filepath.Join(*outDir, rep, "corpus.usc.jsonl")
		if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
			return err
		}
		f, err := os.Create(path)
		if err != nil {
			return err
		}
		writeErr := notation.WriteJSONL(f, records)
		closeErr := f.Close()
		if writeErr != nil {
			return writeErr
		}
		if closeErr != nil {
			return closeErr
		}
		fmt.Printf("%s records=%d documents=%d\n", rep, len(records), len(files))
	}
	return nil
}

func readMusicXMLArchive(path string) (map[string][]byte, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	gz, err := gzip.NewReader(f)
	if err != nil {
		return nil, err
	}
	defer gz.Close()
	tr := tar.NewReader(gz)
	out := map[string][]byte{}
	for {
		h, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		if h.Typeflag != tar.TypeReg || !strings.HasSuffix(strings.ToLower(h.Name), ".xml") {
			continue
		}
		if h.Size > 16<<20 {
			return nil, fmt.Errorf("oversized MusicXML entry %s", h.Name)
		}
		b, err := io.ReadAll(io.LimitReader(tr, h.Size+1))
		if err != nil {
			return nil, err
		}
		if int64(len(b)) != h.Size {
			return nil, fmt.Errorf("short archive entry %s", h.Name)
		}
		out[h.Name] = b
	}
	return out, nil
}

func sortedKeys[V any](m map[string]V) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

func extractMusicEvents(document string, score musicXMLScore) ([]musicEvent, error) {
	if len(score.Parts) == 0 {
		return nil, fmt.Errorf("score has no parts")
	}
	var out []musicEvent
	for partIndex, part := range score.Parts {
		if part.ID == "" {
			return nil, fmt.Errorf("part %d has no id", partIndex)
		}
		divisions, page, system, ordinal := 0, 0, 0, 0
		for measureIndex, measure := range part.Measures {
			cursor, priorOnset := 0, 0
			for _, item := range measure.Items {
				switch item.Kind {
				case "attributes":
					if item.Divisions > 0 {
						divisions = item.Divisions
					}
				case "print":
					if item.Print.NewPage == "yes" {
						page++
						system++
					} else if item.Print.NewSystem == "yes" {
						system++
					}
				case "backup":
					cursor -= item.Duration
				case "forward":
					cursor += item.Duration
				case "note":
					if divisions <= 0 {
						return nil, fmt.Errorf("measure %s has note before divisions", measure.Number)
					}
					if page == 0 || system == 0 {
						return nil, fmt.Errorf("measure %s lacks encoded initial page/system break", measure.Number)
					}
					n := item.Note
					onset := cursor
					if n.Chord != nil {
						onset = priorOnset
					}
					voice, staff := n.Voice, n.Staff
					if voice == "" {
						voice = "1"
					}
					if staff == "" {
						staff = "1"
					}
					e := musicEvent{Document: document, Part: part.ID, Voice: voice, Staff: staff, Measure: measure.Number,
						Page: fmt.Sprintf("page-%04d", page), System: fmt.Sprintf("system-%04d", system), MeasureIndex: measureIndex,
						OnsetNum: onset, OnsetDen: divisions, SourceOrdinal: ordinal, Rest: n.Rest != nil, Chord: n.Chord != nil,
						Step: n.Pitch.Step, Alter: n.Pitch.Alter, Octave: n.Pitch.Octave, DurationNum: n.Duration,
						DurationDen: divisions, NoteType: n.Type, Dots: len(n.Dots), Accidental: n.Accidental}
					for _, tie := range n.Ties {
						e.Ties = append(e.Ties, tie.Type)
					}
					e.OnsetNum, e.OnsetDen = reduce(e.OnsetNum, e.OnsetDen)
					e.DurationNum, e.DurationDen = reduce(e.DurationNum, e.DurationDen)
					if e.DurationNum <= 0 {
						return nil, fmt.Errorf("measure %s has non-positive note duration", measure.Number)
					}
					if !e.Rest && (e.Step == "" || e.Octave == 0) {
						return nil, fmt.Errorf("measure %s has incomplete pitch", measure.Number)
					}
					if !e.Rest {
						if _, ok := map[string]bool{"A": true, "B": true, "C": true, "D": true, "E": true, "F": true, "G": true}[e.Step]; !ok {
							return nil, fmt.Errorf("measure %s has unsupported pitch step %q", measure.Number, e.Step)
						}
						if e.Alter != "" {
							if _, err := strconv.Atoi(e.Alter); err != nil {
								return nil, fmt.Errorf("measure %s has non-integral pitch alteration %q", measure.Number, e.Alter)
							}
						}
					}
					out = append(out, e)
					ordinal++
					priorOnset = onset
					if n.Chord == nil {
						cursor += n.Duration
					}
				}
				if cursor < 0 {
					return nil, fmt.Errorf("measure %s has negative cursor", measure.Number)
				}
			}
		}
	}
	sort.SliceStable(out, func(i, j int) bool {
		a, b := out[i], out[j]
		if a.MeasureIndex != b.MeasureIndex {
			return a.MeasureIndex < b.MeasureIndex
		}
		left, right := int64(a.OnsetNum)*int64(b.OnsetDen), int64(b.OnsetNum)*int64(a.OnsetDen)
		if left != right {
			return left < right
		}
		if a.Part != b.Part {
			return a.Part < b.Part
		}
		if a.Voice != b.Voice {
			return a.Voice < b.Voice
		}
		return a.SourceOrdinal < b.SourceOrdinal
	})
	return out, nil
}

func musicRecords(corpusID, rep string, events []musicEvent) ([]notation.Record, error) {
	var selected []musicEvent
	if rep == "MUSIC-R2" {
		byVoice := map[string][]musicEvent{}
		for _, e := range events {
			key := e.Document + "\x1f" + e.Part + "\x1f" + e.Voice + "\x1f" + e.Staff
			byVoice[key] = append(byVoice[key], e)
		}
		for _, key := range sortedKeys(byVoice) {
			var prev *musicEvent
			for _, e := range byVoice[key] {
				if e.Rest {
					prev = nil
					continue
				}
				current := e
				if prev != nil {
					e.Alter = strconv.Itoa(midiPitch(current) - midiPitch(*prev))
					selected = append(selected, e)
				}
				prev = &current
			}
		}
	} else {
		for _, e := range events {
			if rep == "MUSIC-R3" && e.Rest {
				continue
			}
			selected = append(selected, e)
		}
	}
	lineIndex := map[string]int{}
	records := make([]notation.Record, 0, len(selected))
	for i, e := range selected {
		var symbols []string
		attrs := map[string]string{
			"voice": e.Part + "/" + e.Voice, "staff": e.Part + "/" + e.Staff, "system": e.System,
			"simultaneity_group": fmt.Sprintf("m%04d@%d/%d", e.MeasureIndex+1, e.OnsetNum, e.OnsetDen),
			"source_measure":     e.Measure, "source_event_ordinal": strconv.Itoa(e.SourceOrdinal),
		}
		duration := fmt.Sprintf("%d/%d", e.DurationNum, e.DurationDen)
		pitch := pitchName(e)
		switch rep {
		case "MUSIC-R1":
			kind := "note"
			if e.Rest {
				kind = "rest"
			}
			symbols = append(symbols, "kind:"+kind)
			if !e.Rest {
				symbols = append(symbols, "pitch-class:"+e.Step+alterSuffix(e.Alter), "register:"+strconv.Itoa(e.Octave))
			}
			symbols = append(symbols, "duration:"+duration)
			if e.Accidental != "" {
				symbols = append(symbols, "accidental:"+e.Accidental)
			}
			for _, tie := range e.Ties {
				symbols = append(symbols, "tie:"+tie)
			}
			attrs["event"], attrs["duration"] = kind, duration
			if !e.Rest {
				attrs["pitch"] = pitch
			}
		case "MUSIC-R2":
			interval := signed(e.Alter)
			symbols = []string{"interval:" + interval}
			attrs["interval"] = interval
		case "MUSIC-R3":
			symbols = []string{"pitch:" + pitch, "duration:" + duration}
			attrs["pitch"], attrs["duration"] = pitch, duration
		default:
			return nil, fmt.Errorf("unsupported representation %s", rep)
		}
		if e.NoteType != "" {
			attrs["source_note_type"] = e.NoteType
		}
		if e.Dots > 0 {
			attrs["source_dots"] = strconv.Itoa(e.Dots)
		}
		if e.Chord {
			attrs["source_chord_member"] = "true"
		}
		section := notation.ObservedLevel{}
		if rep == "MUSIC-R2" {
			section = observed("voice:" + e.Part + "/" + e.Voice + "/staff:" + e.Staff)
		}
		key := e.Document + "\x1f" + section.Value + "\x1f" + e.Page + "\x1f" + e.System
		idx := lineIndex[key]
		lineIndex[key]++
		h := sha256.Sum256([]byte(corpusID + "\x00" + rep + "\x00" + e.Document + "\x00" + strconv.Itoa(e.SourceOrdinal) + "\x00" + strconv.Itoa(i)))
		records = append(records, notation.Record{SchemaVersion: notation.SchemaVersion, CorpusID: corpusID, Representation: rep,
			Document: observed(e.Document), Section: section, Page: observed(e.Page), Locus: notation.ObservedLevel{}, PhysicalLine: observed(e.System),
			TokenID: corpusID + "-" + hex.EncodeToString(h[:6]), TokenIndex: idx, Token: strings.Join(symbols, "|"), Symbols: symbols, Attributes: attrs})
	}
	if err := notation.Validate(records); err != nil {
		return nil, err
	}
	return records, nil
}

func observed(v string) notation.ObservedLevel {
	return notation.ObservedLevel{Value: v, Observed: true}
}

func midiPitch(e musicEvent) int {
	base := map[string]int{"C": 0, "D": 2, "E": 4, "F": 5, "G": 7, "A": 9, "B": 11}[e.Step]
	alter, _ := strconv.Atoi(e.Alter)
	return (e.Octave+1)*12 + base + alter
}

func pitchName(e musicEvent) string {
	if e.Rest {
		return ""
	}
	return e.Step + alterSuffix(e.Alter) + strconv.Itoa(e.Octave)
}
func alterSuffix(v string) string {
	switch v {
	case "", "0":
		return ""
	case "1":
		return "#"
	case "-1":
		return "b"
	default:
		return "(" + v + ")"
	}
}
func signed(v string) string {
	if strings.HasPrefix(v, "-") {
		return v
	}
	return "+" + v
}
func reduce(a, b int) (int, int) {
	if b == 0 {
		return a, b
	}
	x, y := a, b
	if x < 0 {
		x = -x
	}
	for y != 0 {
		x, y = y, x%y
	}
	if x == 0 {
		return 0, 1
	}
	return a / x, b / x
}
