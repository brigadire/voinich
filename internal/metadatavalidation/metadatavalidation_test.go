package metadatavalidation

import (
	"bytes"
	"errors"
	"math/rand"
	"reflect"
	"strings"
	"testing"
	"time"
)

func TestProgressStatusBar(t *testing.T) {
	var b bytes.Buffer
	p := newProgress(&b)
	now := time.Unix(0, 0)
	p.clock = func() time.Time { return now }
	p.begin(2, "Strict alignment")
	now = now.Add(2 * time.Second)
	p.update(1, 4, "Strict alignment")
	s := b.String()
	if !strings.Contains(s, "[2/7]") || !strings.Contains(s, "[=====...............]") || !strings.Contains(s, "elapsed 00:02") || !strings.Contains(s, "ETA 00:06") {
		t.Fatalf("status bar: %q", s)
	}
}

func TestParseIVTFFMetadataAndInheritance(t *testing.T) {
	src := `#=IVTFF Eva- 2.0
<f1r> <! $Q=A $C=1 $H=2 $X=V>
# comment words are not text
<f1r.1,@P0> <%>foo.@135;.b?r<!note>
<f1r.2,+P0> baz
<f1r.3,@L0> <@H=3>label
<f1v> <! $H=4>
<f1v.1,+P0> qux
`
	d, err := ParseIVTFFReader(strings.NewReader(src))
	if err != nil {
		t.Fatal(err)
	}
	if d.Pages != 2 || len(d.Loci) != 4 {
		t.Fatalf("pages/loci = %d/%d", d.Pages, len(d.Loci))
	}
	a := d.Loci[0]
	if a.Folio != "f1r" || a.ID != "f1r.1" || a.Type != "P" || !a.ParagraphStart || a.ParagraphID != 1 {
		t.Fatalf("bad first locus: %+v", a)
	}
	if a.Variables["C"] != "1" || a.Variables["H"] != "2" || a.Variables["Q"] != "A" {
		t.Fatalf("inheritance: %#v", a.Variables)
	}
	if d.Loci[1].ParagraphID != 1 || d.Loci[2].Variables["H"] != "3" {
		t.Fatal("paragraph or inline variable inheritance failed")
	}
	if d.Loci[3].Variables["C"] != "" || d.Loci[3].Variables["Q"] != "" || d.Loci[3].Variables["H"] != "4" {
		t.Fatalf("unknown page metadata: %#v", d.Loci[3].Variables)
	}
}

func TestNormalizeForAlignment(t *testing.T) {
	in := `<%>[cth:oto]res.{c'y}.@135;.d?n,?<!ignored><->x<$>`
	want := "cthres c'y @135; d?n ? x"
	if got := NormalizeForAlignment(in); got != want {
		t.Fatalf("got %q want %q", got, want)
	}
}

func TestExactAlignmentAndInvariant(t *testing.T) {
	d := Document{Loci: []Locus{{Folio: "f1r", ID: "f1r.1", Type: "P", LineID: "f1r.1", AlignmentText: "a @135;", Variables: map[string]string{"C": "A"}}, {Folio: "f1v", ID: "f1v.1", Type: "P", LineID: "f1v.1", AlignmentText: "b? c", Variables: map[string]string{}}}}
	r, e := Align(d, []string{"a", "@135;", "b?", "c"}, "hash")
	if e != nil {
		t.Fatal(e)
	}
	if len(r.Records) != 4 || r.Records[0].Position != 0 || r.Records[3].Position != 3 || r.Records[2].Folio != "f1v" {
		t.Fatalf("bad map: %+v", r.Records)
	}
	_, e = Align(d, []string{"a", "@135;", "WRONG", "c"}, "")
	var ae *AlignmentError
	if !errors.As(e, &ae) || ae.Position != 2 || !strings.Contains(e.Error(), "context before") || !strings.Contains(e.Error(), "f1v.1") {
		t.Fatalf("bad diagnostic: %v", e)
	}
	_, e = Align(d, []string{"a", "@135;"}, "")
	if e == nil {
		t.Fatal("token count mismatch accepted")
	}
}

func sampleRecords() []TokenMetadata {
	return []TokenMetadata{{Position: 0, Folio: "f1r", LineID: "l1", ParagraphID: 1, Currier: "A", Hand: "1", Quire: "Q1"}, {Position: 1, Folio: "f1r", LineID: "l1", ParagraphID: 1, Currier: "A", Hand: "1", Quire: "Q1"}, {Position: 2, Folio: "f1r", LineID: "l2", ParagraphID: 2, Currier: "B", Hand: "", Quire: "Q1"}, {Position: 3, Folio: "f1v", LineID: "l3", ParagraphID: 1, Currier: "B", Hand: "2", Quire: "Q2"}}
}
func TestMetadataTransitionsAndNearest(t *testing.T) {
	b := ExtractBoundaries(sampleRecords())
	if len(b["line"]) != 2 || len(b["paragraph"]) != 2 || len(b["folio"]) != 1 || len(b["currier"]) != 1 || len(b["hand"]) != 0 || len(b["quire"]) != 1 {
		t.Fatalf("boundaries: %#v", b)
	}
	p, d := NearestBoundary(4, []int{2, 8})
	if p != 2 || d != 2 {
		t.Fatalf("nearest %d/%d", p, d)
	}
	if MatchWithin([]int{1, 5, 10}, []int{2, 8}, 2) != 2 {
		t.Fatal("fixed tolerance")
	}
}

func TestPermutationControlsDeterministic(t *testing.T) {
	a := UniformBoundaries(20, 4, rand.New(rand.NewSource(7)), make([]int, 19))
	b := UniformBoundaries(20, 4, rand.New(rand.NewSource(7)), make([]int, 19))
	if !reflect.DeepEqual(a, b) {
		t.Fatal("uniform non-deterministic")
	}
	shift := CircularShiftBoundaries([]int{2, 5, 9}, 10, 3)
	if !reflect.DeepEqual(shift, []int{3, 5, 8}) {
		t.Fatalf("shift %v", shift)
	}
	x := []string{"A", "A", "B", "B", "B", "A"}
	p := PermuteBlockLabels(x, rand.New(rand.NewSource(3)))
	if len(p) != len(x) {
		t.Fatal("block permutation length")
	}
	if !reflect.DeepEqual(p, PermuteBlockLabels(x, rand.New(rand.NewSource(3)))) {
		t.Fatal("block permutation non-deterministic")
	}
}

func TestMetadataPurityAndAssociationMetrics(t *testing.T) {
	w := MetadataComposition(sampleRecords(), 0, 3, "currier")
	if w.Label != "A" || w.Purity != 2.0/3 {
		t.Fatalf("composition %+v", w)
	}
	m := AssociationMetrics([]string{"A", "A", "B", "B"}, []string{"0", "0", "1", "1"})
	if m.NMI < .999999 || m.ARI < .999999 || m.Homogeneity < .999999 || m.Completeness < .999999 {
		t.Fatalf("metrics %+v", m)
	}
	zero := AssociationMetrics([]string{"A", "A", "B", "B"}, []string{"0", "1", "0", "1"})
	if zero.MI > 1e-12 {
		t.Fatalf("MI independence %g", zero.MI)
	}
}

func TestAnalyzeAssignmentsKSweepDeterministic(t *testing.T) {
	a := []Assignment{{2, "hierarchical", 2, 0, 0, 2, 0}, {2, "hierarchical", 2, 1, 2, 4, 1}, {2, "hierarchical", 3, 0, 0, 2, 0}, {2, "hierarchical", 3, 1, 2, 4, 2}}
	x := AnalyzeAssignments(a, sampleRecords())
	y := AnalyzeAssignments(a, sampleRecords())
	if len(x) != 12 || !reflect.DeepEqual(x, y) {
		t.Fatalf("association output len=%d deterministic=%t", len(x), reflect.DeepEqual(x, y))
	}
}
