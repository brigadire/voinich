package notation

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestBuildBDDUSCDeterministicAndPaired(t *testing.T) {
	d := t.TempDir()
	p := filepath.Join(d, "book.xml")
	xml := `<TEI><teiHeader><p>apparatus</p></teiHeader><text><body><div type="book" n="06"><pb n="1r"/><cb n="a"/><lb n="1" xml:id="l1"/>In <choice><abbr>dn<g ref="#bar">x</g>o</abbr><expan>domino</expan></choice><pc>.</pc> est<lb n="2" xml:id="l2"/>finis<note>omit</note></div></body></text></TEI>`
	if err := os.WriteFile(p, []byte(xml), 0644); err != nil {
		t.Fatal(err)
	}
	dip1, ds, err := BuildBDDUSC([]string{p}, "C02-BDD-DIP", "LATIN-DIPLOMATIC")
	if err != nil {
		t.Fatal(err)
	}
	dip2, _, err := BuildBDDUSC([]string{p}, "C02-BDD-DIP", "LATIN-DIPLOMATIC")
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(dip1, dip2) {
		t.Fatal("repeated conversion differs")
	}
	if ds.Pages != 1 || ds.PhysicalLines != 2 || ds.Choices != 1 || ds.AbbreviationSigns != 1 {
		t.Fatalf("stats=%+v", ds)
	}
	exp, _, err := BuildBDDUSC([]string{p}, "C01-BDD-EXP", "LATIN-EXPANDED")
	if err != nil {
		t.Fatal(err)
	}
	if dip1[1].Token == exp[1].Token || exp[1].Token != "domino" {
		t.Fatalf("choice branches were not separated: dip=%q exp=%q", dip1[1].Token, exp[1].Token)
	}
	for _, r := range dip1 {
		if !r.Page.Observed || !r.PhysicalLine.Observed || !r.Section.Observed || !r.Locus.Observed {
			t.Fatalf("source hierarchy lost: %+v", r)
		}
	}
}
