package notation

import (
	"math"
	"strconv"
	"testing"
)

// dryRunSyntheticCorpus builds a small, deterministic corpus that is
// neither a calibration control nor a fixture reused elsewhere, satisfying
// section 41's requirement for an independent synthetic dry-run source.
func dryRunSyntheticCorpus(id string, lines, perLine int, seed int64) []Record {
	rng := newRNG(seed)
	alphabet := []string{"p", "q", "r", "s"}
	var out []Record
	idx := 0
	for li := 0; li < lines; li++ {
		for ti := 0; ti < perLine; ti++ {
			n := 2 + rng.Intn(3)
			syms := make([]string, n)
			tok := ""
			for j := range n {
				syms[j] = alphabet[rng.Intn(len(alphabet))]
				tok += syms[j]
			}
			out = append(out, Record{
				SchemaVersion: SchemaVersion, CorpusID: id, Representation: "DRYRUN-R1",
				Document:     ObservedLevel{Value: "doc-a", Observed: true},
				PhysicalLine: ObservedLevel{Value: strconv.Itoa(li), Observed: true},
				TokenID:      id + "-" + strconv.Itoa(idx), TokenIndex: ti, Token: tok, Symbols: syms,
			})
			idx++
		}
	}
	return out
}

// TestEndToEndSyntheticDryRun exercises the full B01-B04 pipeline —
// USC -> validation -> structural analyzer -> rarefaction ->
// distributions/bootstrap -> frozen "VM" reference -> frozen calibration
// scales -> comparator -> result — end to end on synthetic-only data,
// proving executable completeness without interpreting similarity
// (task section 41-42).
func TestEndToEndSyntheticDryRun(t *testing.T) {
	source := dryRunSyntheticCorpus("DRYRUN-SOURCE", 40, 6, 777)
	reference := dryRunSyntheticCorpus("DRYRUN-REFERENCE", 40, 6, 888)

	if err := Validate(source); err != nil {
		t.Fatalf("USC validation failed: %v", err)
	}
	fp, err := Analyze(source)
	if err != nil {
		t.Fatalf("structural analyzer failed: %v", err)
	}
	refFP, err := Analyze(reference)
	if err != nil {
		t.Fatal(err)
	}

	checkpoints := []int{50, 500}
	rareRows, rareSummary, err := RunRarefaction(source, source[0].CorpusID, source[0].Representation, checkpoints, 5, BaseSeed)
	if err != nil {
		t.Fatalf("rarefaction failed: %v", err)
	}
	if len(rareRows) == 0 || len(rareSummary) == 0 {
		t.Fatal("rarefaction produced no rows")
	}
	foundNotComparableCheckpoint := false
	for _, r := range rareRows {
		if r.CheckpointRequested > len(source) {
			foundNotComparableCheckpoint = true
			if r.Comparable {
				t.Fatal("a checkpoint above corpus size must never be comparable")
			}
		}
	}
	_ = foundNotComparableCheckpoint

	bootRows, err := RunBootstrap(source, source[0].CorpusID, source[0].Representation, 20, BaseSeed)
	if err != nil {
		t.Fatalf("bootstrap failed: %v", err)
	}
	for _, r := range bootRows {
		if math.IsNaN(r.Estimate) || math.IsInf(r.Estimate, 0) || math.IsNaN(r.BootstrapSD) || math.IsInf(r.BootstrapSD, 0) {
			t.Fatalf("NaN/Inf in bootstrap row: %+v", r)
		}
	}

	dist := BuildDistributions(source, source[0].CorpusID, source[0].Representation)
	for _, d := range dist {
		if math.IsNaN(d.Probability) || math.IsInf(d.Probability, 0) {
			t.Fatalf("NaN/Inf in distribution row: %+v", d)
		}
	}

	// Frozen calibration scale, built the same way production scales are:
	// pooled MAD/IQR over independently generated calibration corpora, never
	// from source or reference.
	calRuns := RunCalibrationPanel(500, 8, BaseSeed)
	calMetrics, err := AnalyzeCalibrationRuns(calRuns)
	if err != nil {
		t.Fatal(err)
	}
	calScales := BuildCalibrationScales(calMetrics, 500)
	var scales []Scale
	for _, s := range calScales {
		if s.Status == ScaleStatusOK {
			scales = append(scales, Scale{MetricID: s.MetricID, Regime: s.SupportRegime, Center: s.Median, Spread: s.Scale})
		}
	}
	if len(scales) == 0 {
		t.Fatal("dry run produced no usable calibration scale at all")
	}

	rows, fam, err := Compare(fp, refFP, scales)
	if err != nil {
		t.Fatalf("comparator failed: %v", err)
	}
	if len(rows) == 0 {
		t.Fatal("expected comparable metric rows")
	}
	for _, r := range rows {
		if math.IsNaN(r.Distance) || math.IsInf(r.Distance, 0) {
			t.Fatalf("NaN/Inf distance: %+v", r)
		}
	}
	sawL := false
	for _, r := range rows {
		if r.Family == "L" {
			sawL = true
		}
	}
	if !sawL {
		t.Fatal("expected L family rows (comparable or NOT_COMPARABLE) in the comparison output")
	}
	seenFamilies := map[string]bool{}
	for _, f := range fam {
		if f.Family == "TOTAL" {
			t.Fatal("d_TOTAL must never be emitted")
		}
		seenFamilies[f.Family] = true
	}
	for _, want := range []string{"G", "T", "S", "L", "D"} {
		if !seenFamilies[want] {
			t.Fatalf("missing family distance for %s", want)
		}
	}

	// Repeated run must be byte-identical (acceptance criterion 10),
	// excluding nothing here since this dry run has no timestamp field.
	rows2, fam2, err := Compare(fp, refFP, scales)
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != len(rows2) || len(fam) != len(fam2) {
		t.Fatal("repeated comparator run produced a different row count")
	}
	for i := range rows {
		if rows[i] != rows2[i] {
			t.Fatalf("repeated run diverged at row %d: %+v vs %+v", i, rows[i], rows2[i])
		}
	}
}
