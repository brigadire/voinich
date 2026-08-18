package genericsegmentation

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func writeCorpus(t *testing.T, lines int, tokensPerLine int) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "corpus.txt")
	var b strings.Builder
	for l := range lines {
		for i := range tokensPerLine {
			fmt.Fprintf(&b, "w%d_%d ", l, i)
		}
		b.WriteByte('\n')
	}
	if err := os.WriteFile(path, []byte(b.String()), 0644); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestBuildErrNotEnoughData(t *testing.T) {
	for _, lines := range []int{0, 1} {
		path := writeCorpus(t, lines, 5)
		tokens, lineOfToken, _, err := ReadCorpus(path)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := Build(lineOfToken); err != ErrNotEnoughData {
			t.Fatalf("lines=%d tokens=%d: got err=%v, want ErrNotEnoughData", lines, len(tokens), err)
		}
	}
}

func TestBuildNeverSplitsANaturalLine(t *testing.T) {
	path := writeCorpus(t, 500, 6)
	tokens, lineOfToken, _, err := ReadCorpus(path)
	if err != nil {
		t.Fatal(err)
	}
	infos, err := Build(lineOfToken)
	if err != nil {
		t.Fatal(err)
	}
	if len(infos) != len(tokens) {
		t.Fatalf("got %d infos, want %d", len(infos), len(tokens))
	}
	// Every token on the same natural line must carry the same Group.
	groupOfLine := map[string]string{}
	for i, info := range infos {
		line := info.LineID
		if g, ok := groupOfLine[line]; ok {
			if g != info.Group {
				t.Fatalf("token %d: line %s split across groups %s and %s", i, line, g, info.Group)
			}
		} else {
			groupOfLine[line] = info.Group
		}
	}
}

func TestBuildIndexInLineResetsPerLine(t *testing.T) {
	path := writeCorpus(t, 20, 4)
	_, lineOfToken, _, err := ReadCorpus(path)
	if err != nil {
		t.Fatal(err)
	}
	infos, err := Build(lineOfToken)
	if err != nil {
		t.Fatal(err)
	}
	want := 0
	for i, info := range infos {
		if info.IndexInLine != want {
			t.Fatalf("token %d: IndexInLine=%d, want %d", i, info.IndexInLine, want)
		}
		want++
		if i+1 < len(infos) && lineOfToken[i+1] != lineOfToken[i] {
			want = 0
		}
	}
}

func TestBuildDeterministic(t *testing.T) {
	path := writeCorpus(t, 300, 5)
	_, lineOfToken, _, err := ReadCorpus(path)
	if err != nil {
		t.Fatal(err)
	}
	a, err := Build(lineOfToken)
	if err != nil {
		t.Fatal(err)
	}
	b, err := Build(lineOfToken)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(a, b) {
		t.Fatal("Build is not deterministic for identical input")
	}
}

func TestBuildDistinctGroupsAtEveryCorpusScale(t *testing.T) {
	// Phase 10: N ~ 1,000 / 8,000 / 43,713 (Doyle scale) / 60,000+.
	scales := []struct {
		name          string
		lines, tokens int
	}{
		{"n~1000", 200, 5},
		{"n~8000", 1600, 5},
		{"n~43713", 8743, 5},
		{"n~60000+", 12000, 5},
	}
	for _, s := range scales {
		t.Run(s.name, func(t *testing.T) {
			path := writeCorpus(t, s.lines, s.tokens)
			tokens, lineOfToken, _, err := ReadCorpus(path)
			if err != nil {
				t.Fatal(err)
			}
			infos, err := Build(lineOfToken)
			if err != nil {
				t.Fatalf("N=%d: %v", len(tokens), err)
			}
			groups := map[string]bool{}
			blocks := map[string]bool{}
			for _, info := range infos {
				groups[info.Group] = true
			}
			for i := range infos {
				blocks[fmt.Sprintf("%s|%s", infos[i].LineID, infos[i].Group)] = false
			}
			if len(groups) < 2 {
				t.Fatalf("N=%d: only %d distinct groups, want >=2", len(tokens), len(groups))
			}
			if len(groups) > coarseGroups {
				t.Fatalf("N=%d: %d distinct groups exceeds coarseGroups=%d", len(tokens), len(groups), coarseGroups)
			}
		})
	}
}

func TestBuildAdjacentFineBlocksLandInDifferentGroups(t *testing.T) {
	path := writeCorpus(t, 400, 3)
	_, lineOfToken, _, err := ReadCorpus(path)
	if err != nil {
		t.Fatal(err)
	}
	infos, err := Build(lineOfToken)
	if err != nil {
		t.Fatal(err)
	}
	// A "physical block" downstream is a maximal contiguous run of equal
	// Group. Assert such runs are much smaller than the whole line count,
	// i.e. Group actually varies along the corpus rather than collapsing
	// to one constant value.
	distinctRuns := 1
	for i := 1; i < len(infos); i++ {
		if infos[i].Group != infos[i-1].Group {
			distinctRuns++
		}
	}
	if distinctRuns < 2 {
		t.Fatalf("Group never changes across the corpus; got %d contiguous runs", distinctRuns)
	}
}
