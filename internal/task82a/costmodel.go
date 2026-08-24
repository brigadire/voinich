package task82a

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// f2SecondsPerJob is a target-blind, empirically measured cost table
// (f2_timing_test.go's TestF2Timing, run once against synthetic
// non-Voynich token streams before any manifest job executed) keyed by the
// vocabulary-driving quantity for each scaling-policy family at each
// frozen corpus_scale. LITERAL and CUE_RESET_GLOBAL vocabulary grows with
// chunk count (each chunk is a fresh, almost-certainly-unique symbol string),
// so both follow the same measured curve; CUE_RESET_LOCAL vocabulary is
// bounded at `capacity` regardless of scale, so its cost is driven by
// total token count on a much cheaper, near-linear curve. Numbers between
// measured checkpoints are linearly interpolated in log-log space
// (power-law fit b=1.578 measured from the 256/1024-token points); this is
// a compute-cost estimate only, never a metric-definition or weight
// change (task82a.txt sec.52).
var f2SecondsPerScale = map[string]map[string]float64{
	PolicyLiteralReset:   {"SMALL": 0.44, "MEDIUM": 1.31, "LARGE": 3.30},
	PolicyCueResetLocal:  {"SMALL": 0.04, "MEDIUM": 0.09, "LARGE": 0.18},
	PolicyCueResetGlobal: {"SMALL": 3.30, "MEDIUM": 9.85, "LARGE": 29.40},
}

// writeCostModel writes TASK82A_COST_MODEL.tsv (task82a.txt sec.32) before
// any main-generation job runs.
func writeCostModel(dir string, m Manifest) error {
	type agg struct {
		jobs       int
		estSeconds float64
	}
	byPolicyScale := map[string]*agg{}
	for _, j := range m.Jobs {
		key := j.ScalingPolicyID + "|" + j.CorpusScale
		a, ok := byPolicyScale[key]
		if !ok {
			a = &agg{}
			byPolicyScale[key] = a
		}
		a.jobs++
		a.estSeconds += f2SecondsPerScale[j.ScalingPolicyID][j.CorpusScale]
	}
	var b strings.Builder
	b.WriteString("component\tjobs\test_seconds_per_job\test_total_seconds\test_total_cpu_hours\test_ram_mb_per_job\test_output_storage_kb_per_job\tnotes\n")
	totalSeconds := 0.0
	totalJobs := 0
	totalStorageKB := 0.0
	for _, policy := range []string{PolicyLiteralReset, PolicyCueResetLocal, PolicyCueResetGlobal} {
		for _, sc := range CorpusScales {
			key := policy + "|" + sc.ID
			a, ok := byPolicyScale[key]
			if !ok || a.jobs == 0 {
				continue
			}
			perJob := f2SecondsPerScale[policy][sc.ID]
			total := float64(a.jobs) * perJob
			ramMB := 50.0
			storageKB := float64(sc.Chunks) * 0.05 // small per-chunk record, ~50 bytes each in raw JSON
			b.WriteString(fmt.Sprintf("f2_extraction:%s:%s\t%d\t%.3f\t%.1f\t%.4f\t%.0f\t%.1f\tmeasured via f2_timing_test.go on synthetic non-Voynich tokens; Repetitions=%d (reduced from the canonical 1000 for cost, see TASK82A_DESIGN.md)\n",
				policy, sc.ID, a.jobs, perJob, total, total/3600, ramMB, storageKB, f2Repetitions))
			totalSeconds += total
			totalJobs += a.jobs
			totalStorageKB += storageKB * float64(a.jobs)
		}
	}
	assemblySeconds := float64(len(m.Jobs)) * 0.01 // assembly + recovery sampling is pure in-process logic, sub-10ms/job
	b.WriteString(fmt.Sprintf("assembly_and_recovery_sampling\t%d\t%.4f\t%.1f\t%.6f\t%.0f\t%.1f\tmnemonicspace.Runner.Prepare/Recover calls only, no synthetic-corpus generation\n",
		len(m.Jobs), 0.01, assemblySeconds, assemblySeconds/3600, 20.0, 0.0))
	totalSeconds += assemblySeconds
	b.WriteString(fmt.Sprintf("TOTAL\t%d\t\t%.1f\t%.2f\t\t%.1f\tsingle-threaded estimate; distributed/sharded execution (-shard-index/-shard-count) available if needed (task82a.txt sec.32)\n",
		totalJobs, totalSeconds, totalSeconds/3600, totalStorageKB))
	return os.WriteFile(filepath.Join(dir, "TASK82A_COST_MODEL.tsv"), []byte(b.String()), 0o644)
}
