package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"zcore.dev/voinich/internal/notation"
)

// familyDistancesFromComparisonRows replicates notation.Compare()'s own
// family-aggregation formula (mean distance among Comparable metrics per
// family) so it can be applied identically to freshly computed rows and to
// rows re-read from a written VM_COMPARISON.tsv (task section 16
// "aggregation integrity": recomputed aggregates must match).
func familyDistancesFromComparisonRows(rows []notation.ComparisonRow) []notation.FamilyDistance {
	by := map[string][]float64{}
	for _, r := range rows {
		if r.Status == notation.Comparable {
			by[r.Family] = append(by[r.Family], r.Distance)
		}
	}
	var fam []notation.FamilyDistance
	for _, f := range []string{"G", "T", "S", "L", "D"} {
		x := by[f]
		d := notation.FamilyDistance{Family: f, Status: notation.NotComparable, Reason: "no mutually comparable scaled metrics"}
		if len(x) > 0 {
			var sum float64
			for _, v := range x {
				sum += v
			}
			d.Status, d.Reason, d.Distance, d.ComparableMetrics = notation.Comparable, "", sum/float64(len(x)), len(x)
		}
		fam = append(fam, d)
	}
	return fam
}

// aggregateRow is one row of AGGREGATE_SUMMARY.tsv: derived data only,
// fully recomputable from the raw/vm_comparison outputs (task section 12
// and 16 "aggregation integrity").
type aggregateRow struct {
	CandidateID      string
	RepresentationID string
	Checkpoint       int
	Family           string
	Status           notation.Status
	ComparableCount  int
	MeanDistance     float64
	HasDistance      bool
	Reason           string
}

type aggregateResult struct {
	Rows                   []aggregateRow
	RarefactionRunsTotal   int
	BootstrapRunsTotal     int
	MeasurementsTotal      int
	WithinClassApplicable  bool
	CrossClassOutOfScope   bool
}

// computeAggregate derives every aggregate figure directly from the raw
// per-representation results — it is never an independently stored
// measurement (task section 12).
func computeAggregate(results []*perRepresentationResult) (aggregateResult, error) {
	var agg aggregateResult
	for _, r := range results {
		var checkpoints []int
		for ck := range r.Comparisons {
			checkpoints = append(checkpoints, ck)
		}
		sort.Ints(checkpoints)
		for _, ck := range checkpoints {
			c := r.Comparisons[ck]
			for _, f := range familyDistancesFromComparisonRows(c.Rows) {
				row := aggregateRow{CandidateID: r.CandidateID, RepresentationID: r.RepresentationID, Checkpoint: ck, Family: f.Family, Status: f.Status, ComparableCount: f.ComparableMetrics, Reason: f.Reason}
				if f.Status == notation.Comparable {
					row.MeanDistance, row.HasDistance = f.Distance, true
				}
				agg.Rows = append(agg.Rows, row)
			}
		}
		for _, row := range r.RarefactionRows {
			if row.Replicate >= 0 {
				agg.RarefactionRunsTotal++
			}
		}
		agg.BootstrapRunsTotal += len(r.BootstrapRows)
		agg.MeasurementsTotal += len(r.Fingerprint.Metrics)
	}
	// Within-class distribution requires >=3 independent corpora in one
	// class (COMPARISON_PROTOCOL.md line 28); every class in this subset
	// (C01, C02, C06) has exactly one, so it is NOT_APPLICABLE_FOR_CURRENT_PANEL
	// exactly as production-run-preflight already determined (task section 11).
	agg.WithinClassApplicable = false
	agg.CrossClassOutOfScope = true
	return agg, nil
}

func writeAggregateOutputs(bundleRoot string, agg aggregateResult) error {
	path := filepath.Join(bundleRoot, "aggregate", "AGGREGATE_SUMMARY.tsv")
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	fmt.Fprintln(f, "candidate_id\trepresentation_id\tcheckpoint\tfamily\tstatus\tcomparable_metric_count\tmean_distance\treason")
	for _, row := range agg.Rows {
		dist := ""
		if row.HasDistance {
			dist = fmt.Sprintf("%.12g", row.MeanDistance)
		}
		fmt.Fprintf(f, "%s\t%s\t%d\t%s\t%s\t%d\t%s\t%s\n", row.CandidateID, row.RepresentationID, row.Checkpoint, row.Family, row.Status, row.ComparableCount, dist, row.Reason)
	}
	applicability := map[string]any{
		"schema_version": "production-statistical-applicability-1.0",
		"within_class_pair_distances_and_variance": map[string]any{
			"status": "NOT_APPLICABLE_FOR_CURRENT_PANEL",
			"reason": "every class (C01, C02, C06) has exactly one independent corpus; frozen protocol requires >=3 independent corpora in one class and states this as conditional (\"may\"), never mandatory",
		},
		"cross_class_ranking_pca_umap_nearest_neighbour": map[string]any{
			"status": "OUT_OF_SCOPE_REPOSITORY_LOCKED",
			"reason": "frozen protocol places these outside preparation and repository-locked unconditionally",
		},
		"measurements_total":       agg.MeasurementsTotal,
		"rarefaction_runs_total":   agg.RarefactionRunsTotal,
		"bootstrap_runs_total":     agg.BootstrapRunsTotal,
	}
	return writeJSONFile(filepath.Join(bundleRoot, "aggregate", "STATISTICAL_APPLICABILITY.json"), applicability)
}

// recomputeAggregateFromRaw re-derives an aggregateResult purely by
// re-reading the written bundle's raw/vm_comparison files (not from the
// in-memory results), used by post-run aggregation-integrity validation
// (task section 16 "Пересчитать aggregate summaries из raw outputs").
func recomputeAggregateFromRaw(bundleRoot string, results []*perRepresentationResult) (aggregateResult, error) {
	var recomputed []*perRepresentationResult
	for _, r := range results {
		reread := &perRepresentationResult{CandidateID: r.CandidateID, RepresentationID: r.RepresentationID, Dir: r.Dir, Records: r.Records, Comparisons: map[int]vmComparisonAtCheckpoint{}, Fingerprint: r.Fingerprint}
		for ck := range r.Comparisons {
			path := filepath.Join(bundleRoot, "vm_comparison", r.Dir, fmt.Sprintf("checkpoint_%d", ck), "VM_COMPARISON.tsv")
			rows, err := readComparisonTSV(path)
			if err != nil {
				return aggregateResult{}, err
			}
			reread.Comparisons[ck] = vmComparisonAtCheckpoint{Checkpoint: ck, Reachable: r.Comparisons[ck].Reachable, Rows: rows}
		}
		reread.RarefactionRows = r.RarefactionRows
		reread.BootstrapRows = r.BootstrapRows
		recomputed = append(recomputed, reread)
	}
	return computeAggregate(recomputed)
}

// readComparisonTSV parses a written VM_COMPARISON.tsv back into its
// metric-level ComparisonRow set, skipping the trailing d_<family> summary
// lines (those are recomputed independently via
// familyDistancesFromComparisonRows, since WriteComparisonTSV does not
// serialize ComparableMetrics).
func readComparisonTSV(path string) ([]notation.ComparisonRow, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	var rows []notation.ComparisonRow
	s := bufio.NewScanner(f)
	first := true
	for s.Scan() {
		if first {
			first = false
			continue
		}
		cols := strings.Split(s.Text(), "\t")
		if len(cols) != 8 {
			continue
		}
		id := cols[0]
		if strings.HasPrefix(id, "d_") {
			continue // family summary line, recomputed independently
		}
		metricID, regime := id, ""
		if i := strings.IndexByte(id, '['); i >= 0 && strings.HasSuffix(id, "]") {
			metricID, regime = id[:i], id[i+1:len(id)-1]
		}
		row := notation.ComparisonRow{MetricID: metricID, Family: cols[1], Regime: regime, Status: notation.Status(cols[6]), Reason: cols[7]}
		if row.Status == notation.Comparable {
			row.Reference, _ = strconv.ParseFloat(cols[2], 64)
			row.Candidate, _ = strconv.ParseFloat(cols[3], 64)
			row.Distance, _ = strconv.ParseFloat(cols[5], 64)
		}
		rows = append(rows, row)
	}
	return rows, s.Err()
}

func writeCalibrationReferenceRecord(base, bundleRoot string, scales []notation.CalibrationScale) error {
	h, err := notation.FileSHA256(filepath.Join(base, "CALIBRATION_SCALES.tsv"))
	if err != nil {
		return err
	}
	record := map[string]any{
		"schema_version": "production-calibration-reference-1.0",
		"calibration_scales_path": "research/comparative_notation/CALIBRATION_SCALES.tsv",
		"calibration_scales_sha256": h,
		"rows": len(scales),
		"note": "No calibration panel was recomputed during this production run; every calibrated distance references this frozen file by SHA-256.",
	}
	return writeJSONFile(filepath.Join(bundleRoot, "calibration", "CALIBRATION_REFERENCE.json"), record)
}
