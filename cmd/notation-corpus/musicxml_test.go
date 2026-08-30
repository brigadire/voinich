package main

import (
	"archive/tar"
	"compress/gzip"
	"os"
	"path/filepath"
	"testing"

	"zcore.dev/voinich/internal/notation"
)

func TestMusicXMLRepresentations(t *testing.T) {
	xmlDoc := `<?xml version="1.0"?><score-partwise><part id="P1"><measure number="1"><print new-page="yes"/><attributes><divisions>4</divisions></attributes><note><pitch><step>C</step><octave>4</octave></pitch><duration>4</duration><voice>1</voice><type>quarter</type><staff>1</staff></note><note><rest/><duration>2</duration><voice>1</voice><type>eighth</type><staff>1</staff></note><note><pitch><step>D</step><octave>4</octave></pitch><duration>4</duration><voice>1</voice><type>quarter</type><staff>1</staff></note><note><pitch><step>E</step><octave>4</octave></pitch><duration>4</duration><voice>1</voice><type>quarter</type><staff>1</staff></note></measure></part></score-partwise>`
	archive := filepath.Join(t.TempDir(), "source.tar.gz")
	f, err := os.Create(archive)
	if err != nil {
		t.Fatal(err)
	}
	gz := gzip.NewWriter(f)
	tw := tar.NewWriter(gz)
	if err := tw.WriteHeader(&tar.Header{Name: "C06/test.xml", Mode: 0644, Size: int64(len(xmlDoc))}); err != nil {
		t.Fatal(err)
	}
	if _, err := tw.Write([]byte(xmlDoc)); err != nil {
		t.Fatal(err)
	}
	if err := tw.Close(); err != nil {
		t.Fatal(err)
	}
	if err := gz.Close(); err != nil {
		t.Fatal(err)
	}
	if err := f.Close(); err != nil {
		t.Fatal(err)
	}
	out := filepath.Join(t.TempDir(), "usc")
	if err := musicXMLUSCCmd([]string{"--input", archive, "--output-dir", out}); err != nil {
		t.Fatal(err)
	}
	want := map[string]int{"MUSIC-R1": 4, "MUSIC-R2": 1, "MUSIC-R3": 3}
	for rep, count := range want {
		in, err := os.Open(filepath.Join(out, rep, "corpus.usc.jsonl"))
		if err != nil {
			t.Fatal(err)
		}
		records, err := notation.ReadJSONL(in)
		in.Close()
		if err != nil || notation.Validate(records) != nil {
			t.Fatalf("%s invalid: %v", rep, err)
		}
		if len(records) != count {
			t.Fatalf("%s records=%d want=%d", rep, len(records), count)
		}
		if rep == "MUSIC-R2" && (!records[0].Section.Observed || records[0].Section.Value == "") {
			t.Fatal("MUSIC-R2 must preserve the source-observed voice as its sequence section")
		}
	}
}

func TestMusicR2KeepsOriginalAlterForNextInterval(t *testing.T) {
	events := []musicEvent{
		{Document: "d", Part: "P1", Voice: "1", Staff: "1", Page: "p", System: "s", Step: "C", Octave: 4, DurationNum: 1, DurationDen: 1},
		{Document: "d", Part: "P1", Voice: "1", Staff: "1", Page: "p", System: "s", Step: "C", Alter: "1", Octave: 4, DurationNum: 1, DurationDen: 1, SourceOrdinal: 1},
		{Document: "d", Part: "P1", Voice: "1", Staff: "1", Page: "p", System: "s", Step: "D", Octave: 4, DurationNum: 1, DurationDen: 1, SourceOrdinal: 2},
	}
	records, err := musicRecords("C06-TEST", "MUSIC-R2", events)
	if err != nil {
		t.Fatal(err)
	}
	if len(records) != 2 || records[0].Token != "interval:+1" || records[1].Token != "interval:+1" {
		t.Fatalf("intervals do not use original pitches: %+v", records)
	}
}
