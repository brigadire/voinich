package notation

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"sort"
	"strings"
	"testing"
)

func fixtureRoot(t *testing.T) string {
	t.Helper()
	_, file, _, _ := runtime.Caller(0)
	return filepath.Join(filepath.Dir(file), "..", "..", "research", "comparative_notation", "tools", "adapters", "fixtures")
}

func TestAllClassAdaptersHaveExactFixtures(t *testing.T) {
	dirs, err := os.ReadDir(fixtureRoot(t))
	if err != nil {
		t.Fatal(err)
	}
	var docs []SourceDocument
	seen := map[string]bool{}
	for _, d := range dirs {
		if !d.IsDir() {
			continue
		}
		dir := filepath.Join(fixtureRoot(t), d.Name())
		b, err := os.ReadFile(filepath.Join(dir, "source.json"))
		if err != nil {
			t.Fatal(err)
		}
		var src SourceDocument
		if err := json.Unmarshal(b, &src); err != nil {
			t.Fatal(err)
		}
		docs = append(docs, src)
		rs, err := NormalizeFixture(src)
		if err != nil {
			t.Fatalf("%s: %v", d.Name(), err)
		}
		if len(rs) < 20 || len(rs) > 100 {
			t.Fatalf("%s: fixture size %d", d.Name(), len(rs))
		}
		if err := Validate(rs); err != nil {
			t.Fatal(err)
		}
		var got bytes.Buffer
		if err := WriteJSONL(&got, rs); err != nil {
			t.Fatal(err)
		}
		want, err := os.ReadFile(filepath.Join(dir, "expected.usc.jsonl"))
		if err != nil {
			t.Fatal(err)
		}
		generated, err := os.ReadFile(filepath.Join(dir, "generated.usc.jsonl"))
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(got.Bytes(), want) || !bytes.Equal(got.Bytes(), generated) {
			t.Fatalf("%s: generated USC differs", d.Name())
		}
		for _, r := range rs {
			if !r.Page.Observed || !r.PhysicalLine.Observed || !r.Section.Observed || !r.Locus.Observed {
				t.Fatalf("%s: hierarchy lost", d.Name())
			}
		}
		seen[src.ClassID] = true
	}
	if err := ValidateClassCoverage(docs); err != nil {
		t.Fatal(err)
	}
	if len(seen) != 10 {
		t.Fatalf("class coverage=%d", len(seen))
	}
}

func TestMusicAndTablatureDimensions(t *testing.T) {
	dirs, _ := os.ReadDir(fixtureRoot(t))
	var reps []string
	for _, d := range dirs {
		if !d.IsDir() {
			continue
		}
		b, _ := os.ReadFile(filepath.Join(fixtureRoot(t), d.Name(), "source.json"))
		var src SourceDocument
		_ = json.Unmarshal(b, &src)
		if src.ClassID == "C06" || src.ClassID == "C07" {
			rs, err := NormalizeFixture(src)
			if err != nil {
				t.Fatal(err)
			}
			for _, r := range rs {
				if r.Attributes["simultaneity_group"] == "" {
					t.Fatalf("%s lost simultaneity", src.Representation)
				}
			}
			reps = append(reps, src.Representation)
		}
	}
	sort.Strings(reps)
	want := []string{"MUSIC-R1", "MUSIC-R2", "MUSIC-R3", "TAB-R1", "TAB-R2"}
	if !reflect.DeepEqual(reps, want) {
		t.Fatalf("representations=%v", reps)
	}
}

func TestMissingLevelsAreNotInvented(t *testing.T) {
	src := SourceDocument{CorpusID: "C01-NULL", ClassID: "C01", Representation: "LATIN-DIPLOMATIC", DocumentID: "d", Units: []SourceUnit{{Token: "ab", Symbols: []string{"a", "b"}}}}
	rs, err := NormalizeFixture(src)
	if err != nil {
		t.Fatal(err)
	}
	r := rs[0]
	if r.Section.Observed || r.Page.Observed || r.Locus.Observed || r.PhysicalLine.Observed {
		t.Fatal("adapter invented hierarchy")
	}
}

func fingerprintMetricMap(fp Fingerprint) map[string]Metric {
	m := map[string]Metric{}
	for _, x := range fp.Metrics {
		m[x.MetricID+"/"+x.Regime] = x
	}
	return m
}
func assertFamilyEqual(t *testing.T, a, b Fingerprint, families string) {
	t.Helper()
	am, bm := fingerprintMetricMap(a), fingerprintMetricMap(b)
	for k, x := range am {
		if !strings.Contains(families, x.Family) {
			continue
		}
		y := bm[k]
		if x.Status != y.Status || abs(x.Value-y.Value) > 1e-12 {
			t.Errorf("metric %s changed: %v vs %v", k, x, y)
		}
	}
}
func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}
