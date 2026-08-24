package fingerprintv2

import (
	"bytes"
	"encoding/json"
	"math/rand"
	"strconv"
	"testing"
)

func task79Fixture(boundary bool) corpus {
	var records []tokenRecord
	line := 0
	for folio := 0; folio < 4; folio++ {
		for local := 0; local < 6; local++ {
			n := 3
			if folio >= 2 {
				n = 8
			}
			for pos := 0; pos < n; pos++ {
				tok := "mid"
				if boundary && pos == 0 {
					tok = "start"
				}
				if boundary && pos == n-1 {
					tok = "end"
				}
				records = append(records, tokenRecord{Token: tok, Glyph: []string{tok}, Line: line, LineID: "line" + strconv.Itoa(line), IndexInLine: pos, LineLength: n, Page: "f" + strconv.Itoa(folio) + "r", FolioSide: "r", LocusID: "l" + strconv.Itoa(line), LocusType: "P", Section: map[bool]string{true: "A", false: "B"}[folio < 2], Currier: "1", Hand: "1", IVTFFI: "A"})
			}
			line++
		}
	}
	return corpus{info: CorpusInfo{ID: "fixture", SHA256: "fixture", MetadataAlignment: "strict IVTFF aligned"}, records: records}
}

func TestTask79HierarchyPositiveAndFlatNegativeFixtures(t *testing.T) {
	positive := buildLineProfiles(task79Fixture(false))
	if got := varianceShare(positive, "section"); got < 0.5 {
		t.Fatalf("hierarchy-positive fixture variance share=%g", got)
	}
	flat := append([]LineProfile(nil), positive...)
	for i := range flat {
		flat[i].TokenCount = 5
	}
	if got := varianceShare(flat, "section"); got != 0 {
		t.Fatalf("flat negative fixture variance share=%g", got)
	}
}

func TestTask79LineProfilesPreserveEncounterOrder(t *testing.T) {
	c := corpus{records: []tokenRecord{
		{Token: "a", Glyph: []string{"a"}, LineID: "f1.2", Page: "f1r", LineLength: 1},
		{Token: "b", Glyph: []string{"b"}, LineID: "f1.10", Page: "f1r", LineLength: 1},
		{Token: "c", Glyph: []string{"c"}, LineID: "f1.3", Page: "f1r", LineLength: 1},
	}}
	profiles := buildLineProfiles(c)
	if profiles[0].LineID != "f1.2" || profiles[1].LineID != "f1.10" || profiles[2].LineID != "f1.3" {
		t.Fatalf("profiles were reordered lexically: %+v", profiles)
	}
}

func TestTask79BoundaryStructureFixture(t *testing.T) {
	c := task79Fixture(true)
	x, y, g := metricVectors(c, "BP1_BOUNDARY_TOKEN_NMI")
	test := nmiPermutationTest("boundary", "HN5", x, y, g, 99, rand.New(rand.NewSource(7)))
	if test.PValue > 0.05 || test.Observed <= test.NullMean {
		t.Fatalf("boundary fixture not detected: %+v", test)
	}
}

func TestTask79ConfoundedHierarchyFixture(t *testing.T) {
	x := []string{"a", "a", "b", "b"}
	section := []string{"A", "A", "B", "B"}
	if normalizedMI(x, section) == 0 {
		t.Fatal("expected pooled confounding signal")
	}
	for _, s := range []string{"A", "B"} {
		var xx, yy []string
		for i := range x {
			if section[i] == s {
				xx = append(xx, x[i])
				yy = append(yy, "constant")
			}
		}
		if normalizedMI(xx, yy) != 0 {
			t.Fatal("within-section association should be zero")
		}
	}
}

func TestTask79MetadataCorruptionFixture(t *testing.T) {
	c := task79Fixture(false)
	c.records[1].Page = "corrupt-page"
	a := auditMetadata(c, "patch-v1")
	if a.Status != "REQUIRES_REPAIR" || a.NestingViolations == 0 {
		t.Fatalf("corruption not rejected: %+v", a)
	}
}

func TestTask79SerializationRoundTripAndManifest(t *testing.T) {
	c := task79Fixture(true)
	cfg := testConfig("unused")
	cfg.Task79 = &Task79Config{Enabled: true, Permutations: 20, BootstrapReplicates: 20}
	base := CorpusResult{Corpus: c.info, Grammar: &GrammarSummary{Validation: "SUPPORTED"}, EditGraph: &EditFamilyValidation{}, CrossScale: &CrossScaleResult{}}
	r := runTask79(c, cfg, base, 123)
	if r.FreezeManifest.CandidateID != "FINGERPRINT_V2_CANDIDATE" || r.FreezeManifest.Status != "TASK79_B_REQUIRED" {
		t.Fatalf("bad freeze manifest: %+v", r.FreezeManifest)
	}
	b, err := json.Marshal(r)
	if err != nil {
		t.Fatal(err)
	}
	var decoded Task79Result
	if err := json.Unmarshal(b, &decoded); err != nil {
		t.Fatal(err)
	}
	b2, err := json.Marshal(decoded)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(b, b2) {
		t.Fatal("task79 result changed during JSON round trip")
	}
}
