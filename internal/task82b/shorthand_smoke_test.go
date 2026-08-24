package task82b

import (
	"math/rand"
	"path/filepath"
	"sort"
	"testing"
)

func TestShorthandOpsAndSXSmoke(t *testing.T) {
	paths, _ := filepath.Glob("../../data_test/bdd-tei/koeln-edd-c-119/*.xml")
	if len(paths) == 0 {
		t.Skip("BDD TEI XML not present locally")
	}
	res, err := ExtractTEIPairs(paths)
	if err != nil {
		t.Fatal(err)
	}
	reg := BuildOperationRegistry(res.Pairs)
	classCounts := map[string]int{}
	for _, row := range reg {
		classCounts[row.Class] += row.Count
	}
	keys := make([]string, 0, len(classCounts))
	for k := range classCounts {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		t.Logf("class=%-28s occurrences=%d", k, classCounts[k])
	}
	t.Logf("distinct (abbr,expan) types=%d", len(reg))

	sx := ComputeSX(res.Pairs)
	for _, m := range sx {
		t.Logf("%-40s = %.6f  (%s)", m.ID, m.Value, m.Note)
	}

	st := BuildCharDeletionStats(res.Pairs)
	r := rand.New(rand.NewSource(1))
	for i := 0; i < 5 && i < len(res.Pairs); i++ {
		p := res.Pairs[i]
		d := DeletionCount(p)
		nullR := NullWord("RANDOM_DELETION_MATCHED", p.ExpanText, d, st, r)
		nullF := NullWord("FREQUENCY_MATCHED_DELETION", p.ExpanText, d, st, r)
		nullP := NullWord("POSITION_MATCHED", p.ExpanText, d, st, r)
		t.Logf("expan=%q real_abbr=%q d=%d null_random=%q null_freq=%q null_pos=%q", p.ExpanText, p.AbbrText, d, nullR, nullF, nullP)
	}
}
