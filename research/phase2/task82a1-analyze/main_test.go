package main

import "testing"

func TestGenericMetricsPreserveFrozenSingleTokenLineSemantics(t *testing.T) {
	got := genericMetrics([][]string{{"ab"}, {"abc"}})
	for _, id := range []string{
		"2DL1_LAYOUT_POSITION_MI",
		"BP1_BOUNDARY_TOKEN_NMI",
		"LS1_LINE_LENGTH_CV",
		"LS2_POSITIONAL_LEXICON_NMI",
		"LS3_BOUNDARY_LENGTH_ASYMMETRY",
		"LS4_WITHIN_LINE_EXACT_REPETITION",
	} {
		if got[id] != 0 {
			t.Fatalf("%s = %v, want frozen-formula zero for singleton assembler lines", id, got[id])
		}
	}
}

func BenchmarkGenericMetrics256AssemblerLines(b *testing.B) {
	lines := make([][]string, 256)
	for i := range lines {
		lines[i] = []string{"a", "bb", "ccc", "a"}
	}
	b.ReportAllocs()
	for b.Loop() {
		genericMetrics(lines)
	}
}

func TestFrozenAuditAndExtensionCardinality(t *testing.T) {
	if len(metrics) != 33 {
		t.Fatalf("audited metrics = %d, want 33", len(metrics))
	}
	statuses := map[string]int{}
	extended := 0
	seen := map[string]bool{}
	for _, m := range metrics {
		if seen[m.ID] {
			t.Fatalf("duplicate metric %s", m.ID)
		}
		seen[m.ID] = true
		statuses[m.Status]++
		if isNewGenericMetric(m.ID) {
			extended++
		}
	}
	want := map[string]int{"DIRECTLY_APPLICABLE": 8, "ASSEMBLER_APPLICABLE": 9, "NOT_APPLICABLE_STRUCTURE": 5, "NOT_APPLICABLE_METADATA": 11}
	for status, n := range want {
		if statuses[status] != n {
			t.Fatalf("%s = %d, want %d", status, statuses[status], n)
		}
	}
	if extended != 6 {
		t.Fatalf("new generic metrics = %d, want 6", extended)
	}
}
