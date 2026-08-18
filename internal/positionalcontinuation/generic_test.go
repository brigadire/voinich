package positionalcontinuation

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

func writeHigherOrderValidation(t *testing.T, rows [][2]string) string {
	// rows: [sequence, conditional_fdr_q], all HIGHER_ORDER_REPLICATED.
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "higher_order_validation.tsv")
	var b strings.Builder
	b.WriteString("sequence\tfamily\tfinal_status\tconditional_fdr_q\teligible_blocks\tsign_consistency\tlobo_advantage_fraction\tsingle_block_sensitive\tdistinct_joint_classes\tposition_dependent\tgroup_limited\n")
	for _, r := range rows {
		fmt.Fprintf(&b, "%s\tprimary\tHIGHER_ORDER_REPLICATED\t%s\t5\t.9\t.8\tfalse\t3\tfalse\tfalse\n", r[0], r[1])
	}
	fmt.Fprintf(&b, "w1_0 w1_1 w1_2\tprimary\tFIRST_ORDER_EXPLAINED\t.9\t5\t.9\t.8\tfalse\t3\tfalse\tfalse\n")
	if err := os.WriteFile(path, []byte(b.String()), 0644); err != nil {
		t.Fatal(err)
	}
	return dir
}

func TestResolveGenericTargetPicksLowestFDRQAmongReplicated(t *testing.T) {
	dir := writeHigherOrderValidation(t, [][2]string{
		{"w0_0 w0_1 w0_2", ".04"},
		{"w2_0 w2_1 w2_2", ".01"},
		{"w3_0 w3_1 w3_2", ".02"},
	})
	a, b, c, err := resolveGenericTarget(dir)
	if err != nil {
		t.Fatal(err)
	}
	if a != "w2_0" || b != "w2_1" || c != "w2_2" {
		t.Fatalf("expected the lowest-fdr_q HIGHER_ORDER_REPLICATED row (w2_0 w2_1 w2_2), got %q %q %q", a, b, c)
	}
}

func TestResolveGenericTargetErrorsWithNoReplicatedCandidate(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "higher_order_validation.tsv")
	content := "sequence\tfamily\tfinal_status\tconditional_fdr_q\n" +
		"w0_0 w0_1 w0_2\tprimary\tFIRST_ORDER_EXPLAINED\t.9\n"
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
	if _, _, _, err := resolveGenericTarget(dir); err == nil {
		t.Fatal("expected an explicit error when no HIGHER_ORDER_REPLICATED candidate exists, got nil")
	}
}
