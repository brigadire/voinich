package higherorderseq

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func writeGenericCorpus(t *testing.T, lines, tokensPerLine int) string {
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

func TestLoadCorpusAndBlocksGenericDeterministic(t *testing.T) {
	path := writeGenericCorpus(t, 300, 4)

	tokensA, blocksA, lineLenA, corpusSHAA, metaSHAA, err := loadCorpusAndBlocks(path, "", true)
	if err != nil {
		t.Fatal(err)
	}
	if len(blocksA) < 2 {
		t.Fatalf("expected >=2 physical blocks, got %d", len(blocksA))
	}
	for _, tok := range tokensA {
		if tok.Hand != "generic" {
			t.Fatalf("Hand must always be the generic sentinel, got %q", tok.Hand)
		}
	}

	tokensB, blocksB, lineLenB, corpusSHAB, metaSHAB, err := loadCorpusAndBlocks(path, "", true)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(tokensA, tokensB) || !reflect.DeepEqual(blocksA, blocksB) ||
		!reflect.DeepEqual(lineLenA, lineLenB) || corpusSHAA != corpusSHAB || metaSHAA != metaSHAB {
		t.Fatal("loadCorpusAndBlocks(generic=true) is not deterministic across identical calls")
	}
}

func TestMetadataLimitedGenericIgnoresConstantHandDimension(t *testing.T) {
	// DistinctHand is always 1 in generic mode (constant sentinel); the
	// MetadataLimited computation must not fall back to the IVTFF-mode
	// formula, which would trivially mislabel any candidate with
	// DistinctJoint>=2 as limited purely because DistinctHand<=1.
	in := classificationInput{
		CrossBlock: CrossBlockRow{EligibleBlocks: 5, DistinctCurrier: 3, DistinctHand: 1, DistinctJoint: 3, SignConsistency: .9},
		Generic:    true,
	}
	row := classify(in)
	if row.MetadataLimited {
		t.Fatalf("generic MetadataLimited must ignore the constant Hand dimension: %+v", row)
	}
	if row.FinalStatus == "METADATA_LIMITED" {
		t.Fatalf("generic mode must never emit METADATA_LIMITED, got %q", row.FinalStatus)
	}
}
