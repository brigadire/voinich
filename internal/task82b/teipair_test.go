package task82b

import (
	"path/filepath"
	"testing"
)

func TestExtractTEIPairsBDD(t *testing.T) {
	paths, err := filepath.Glob("../../data_test/bdd-tei/koeln-edd-c-119/*.xml")
	if err != nil {
		t.Fatal(err)
	}
	if len(paths) == 0 {
		t.Skip("BDD TEI XML not present locally")
	}
	res, err := ExtractTEIPairs(paths)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("pairs=%d abbrLines=%d expanLines=%d glyphrefs=%d", len(res.Pairs), len(res.AbbrGroups), len(res.ExpanGroups), len(res.GlyphPlaceholders))
	if len(res.Pairs) < 100 {
		t.Fatalf("expected >=100 pairs from 5 chapters, got %d", len(res.Pairs))
	}
	nMark, nCombining := 0, 0
	for _, p := range res.Pairs {
		if p.HasMark {
			nMark++
		}
		if p.MarkIsCombining {
			nCombining++
		}
	}
	t.Logf("nMark=%d nCombining=%d", nMark, nCombining)
	for i := 0; i < 10 && i < len(res.Pairs); i++ {
		p := res.Pairs[i]
		t.Logf("pair[%d] file=%s abbr=%q expan=%q mark=%v combining=%v", i, filepath.Base(p.File), p.AbbrText, p.ExpanText, p.HasMark, p.MarkIsCombining)
	}
	abbrTok, _ := FlattenTokens(res.AbbrGroups)
	expanTok, _ := FlattenTokens(res.ExpanGroups)
	t.Logf("abbr tokens=%d expan tokens=%d", len(abbrTok), len(expanTok))
}
