package transitionnetwork

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

	tokensA, blocksA, corpusSHAA, metaSHAA, err := loadCorpusAndBlocks(path, "", true)
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

	tokensB, blocksB, corpusSHAB, metaSHAB, err := loadCorpusAndBlocks(path, "", true)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(tokensA, tokensB) || !reflect.DeepEqual(blocksA, blocksB) || corpusSHAA != corpusSHAB || metaSHAA != metaSHAB {
		t.Fatal("loadCorpusAndBlocks(generic=true) is not deterministic across identical calls")
	}
}

func TestComputeGraphDiagnosticsGenericSkipsCurrierAndHand(t *testing.T) {
	tokens, blocks, _, _, err := loadCorpusAndBlocks(writeGenericCorpus(t, 60, 4), "", true)
	if err != nil {
		t.Fatal(err)
	}
	counts, vocab, edges, data := buildData(tokens, blocks, 1)
	a := &analysis{Tokens: tokens, Blocks: blocks, Counts: counts, Vocab: vocab, Edges: edges, Data: data}
	summarizeEdges(a, 1)
	computeGraphDiagnostics(a, 1, true)
	for _, r := range a.MetadataTransfer {
		if r.Dimension == "currier" || r.Dimension == "hand" {
			t.Fatalf("generic mode must never compute a currier/hand metadata-transfer row (no real second dimension), got %+v", r)
		}
	}
}

func TestClassifyGenericRelabelsJointSpecific(t *testing.T) {
	r := &EdgeSummary{EligibleBlocks: 5, FDRQ: .01, JointClasses: 1}
	classify(r, true)
	if r.Status != "GROUP_SPECIFIC" {
		t.Fatalf("generic mode must emit GROUP_SPECIFIC, got %q", r.Status)
	}
	r2 := &EdgeSummary{EligibleBlocks: 5, FDRQ: .01, JointClasses: 1}
	classify(r2, false)
	if r2.Status != "METADATA_SPECIFIC" {
		t.Fatalf("IVTFF mode must keep emitting METADATA_SPECIFIC, got %q", r2.Status)
	}
}
