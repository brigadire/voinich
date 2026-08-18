package tokenrelationvalidation

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

func TestLoadGenericMetadataDeterministicAndPartitions(t *testing.T) {
	path := writeGenericCorpus(t, 300, 4)
	corpus, _, err := loadCorpus(path)
	if err != nil {
		t.Fatal(err)
	}
	tokensA, blocksA, unknownA, shaA, err := loadGenericMetadata(path, corpus)
	if err != nil {
		t.Fatal(err)
	}
	if unknownA != 0 {
		t.Fatalf("generic segmentation must never produce unknown tokens, got %d", unknownA)
	}
	if len(tokensA) != len(corpus) {
		t.Fatalf("token count %d != corpus token count %d", len(tokensA), len(corpus))
	}
	if len(blocksA) < 2 {
		t.Fatalf("expected >=2 physical blocks, got %d", len(blocksA))
	}
	for _, tok := range tokensA {
		if tok.Hand != "generic" {
			t.Fatalf("Hand must always be the generic sentinel, got %q", tok.Hand)
		}
	}

	tokensB, blocksB, unknownB, shaB, err := loadGenericMetadata(path, corpus)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(tokensA, tokensB) || !reflect.DeepEqual(blocksA, blocksB) || unknownA != unknownB || shaA != shaB {
		t.Fatal("loadGenericMetadata is not deterministic across identical calls")
	}
}

func TestClassifyGenericNeverEmitsMetadataVocabulary(t *testing.T) {
	s := RelationSummary{EligibleBlocks: 5, JointClasses: 3, SignConsistency: .9, TransferSuccess: .8}
	for _, tc := range []struct{ withinGroup, crossGroup bool }{
		{true, true}, {true, false}, {false, false},
	} {
		got := ClassifyGeneric(s, tc.withinGroup, tc.crossGroup)
		for _, forbidden := range []string{"CURRIER", "HAND", "UNIVERSAL"} {
			if strings.Contains(got, forbidden) {
				t.Fatalf("ClassifyGeneric(%v) = %q must never contain %q", tc, got, forbidden)
			}
		}
	}
}
