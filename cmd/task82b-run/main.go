// Command task82b-run generates every raw Task82b job (shorthand branch +
// extraction branch): F2 vectors, AX diagnostics, and provenance, one
// small JSON file per job under -out/raw/. It is resumable: a job whose
// raw file already exists is skipped, so an interrupted run can be
// restarted with the same command. Aggregation into the TSV/report
// deliverables is a separate step (cmd/task82b-aggregate), since
// aggregation is cheap and generation is not.
package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"flag"
	"log"
	"os"
	"path/filepath"
	"sort"

	"zcore.dev/voinich/internal/task82b"
)

func main() {
	root := flag.String("root", ".", "repository root")
	out := flag.String("out", "research/phase2/task82b", "output directory")
	baseSeed := flag.Int64("seed", 82002, "base seed")
	onlyKind := flag.String("only-kind", "", "if set, run only this job kind")
	flag.Parse()

	rawDir := filepath.Join(*out, "raw")
	if err := os.MkdirAll(rawDir, 0o755); err != nil {
		log.Fatal(err)
	}

	jobs := buildExtractionJobs(*baseSeed)
	jobs = append(jobs, buildShorthandJobs(*root, *baseSeed)...)
	sort.Slice(jobs, func(i, j int) bool { return jobs[i].JobID < jobs[j].JobID })

	log.Printf("task82b-run: %d jobs total", len(jobs))
	done, skipped, failed := 0, 0, 0
	for _, j := range jobs {
		if *onlyKind != "" && j.Kind != *onlyKind {
			continue
		}
		rawPath := filepath.Join(rawDir, safeName(j.JobID)+".json")
		if _, err := os.Stat(rawPath); err == nil {
			skipped++
			continue
		}
		rec, err := j.Run(*root, filepath.Join(*out, "f2work", safeName(j.JobID)))
		if err != nil {
			log.Printf("FAILED %s: %v", j.JobID, err)
			failed++
			continue
		}
		data, err := json.MarshalIndent(rec, "", "  ")
		if err != nil {
			log.Fatal(err)
		}
		if err := os.WriteFile(rawPath, append(data, '\n'), 0o644); err != nil {
			log.Fatal(err)
		}
		done++
		if done%10 == 0 {
			log.Printf("progress: %d done, %d skipped, %d failed (of %d)", done, skipped, failed, len(jobs))
		}
	}
	log.Printf("task82b-run complete: %d done, %d skipped, %d failed", done, skipped, failed)
}

func safeName(id string) string {
	h := sha256.Sum256([]byte(id))
	return hex.EncodeToString(h[:8])
}

// Job is one unit of F2/AX work. Exactly one of the two Run paths applies
// depending on Kind.
type Job struct {
	JobID           string
	Kind            string // carrier_baseline | operator_output | operator_null_random | operator_null_stratified | operator_null_periodic | shorthand_variant
	CorpusID        string
	OperatorID      string
	StructuralClass string
	ExtractionClass string
	Provenance      string
	NullClass       string
	Seed            int64
	Phase           int
	Period          int
	// shorthand-only:
	ShorthandVariant string
	ShorthandScale   string
}

// JobRecord is the raw persisted result of one Job.
type JobRecord struct {
	JobID            string            `json:"job_id"`
	Kind             string            `json:"kind"`
	CorpusID         string            `json:"corpus_id"`
	OperatorID       string            `json:"operator_id,omitempty"`
	StructuralClass  string            `json:"structural_class,omitempty"`
	ExtractionClass  string            `json:"extraction_class,omitempty"`
	Provenance       string            `json:"provenance,omitempty"`
	NullClass        string            `json:"null_class,omitempty"`
	Seed             int64             `json:"seed"`
	Phase            int               `json:"phase,omitempty"`
	Period           int               `json:"period,omitempty"`
	ShorthandVariant string            `json:"shorthand_variant,omitempty"`
	ShorthandScale   string            `json:"shorthand_scale,omitempty"`
	ChosenCount      int               `json:"chosen_count"`
	CandidatePool    int               `json:"candidate_pool"`
	SkippedGroups    int               `json:"skipped_groups"`
	Degenerate       bool              `json:"degenerate"`
	DegenerateNote   string            `json:"degenerate_note,omitempty"`
	F2               task82b.F2Vector  `json:"f2"`
	AX               *task82b.AXResult `json:"ax,omitempty"`
}

// safeExtractF2 runs ExtractF2 and, if the frozen extractor itself
// errors (observed cause: PERIODIC_GLYPH/other single-glyph-alphabet
// outputs can exhaust the frequency-aware C-GRAMMAR null generator's
// unique-form budget -- an intrinsic degeneracy of a near-alphabet-size
// vocabulary, not a task82b bug), returns a placeholder F2Vector with
// every metric marked unavailable and the error recorded, so the job
// still gets a definitive raw record instead of retrying forever on
// every resume (task82b.txt sec.63: mark DEGENERATE_OUTPUT, don't drop
// it).
func safeExtractF2(path, jobID, corpusID string, seed int64, outDir string, groups [][]string) (task82b.F2Vector, bool) {
	f2, err := task82b.ExtractF2(path, jobID, corpusID, seed, outDir, groups)
	if err == nil {
		return f2, true
	}
	log.Printf("F2 extractor error for %s (recorded as unavailable, not retried): %v", jobID, err)
	tokens, types := 0, map[string]bool{}
	for _, g := range groups {
		tokens += len(g)
		for _, t := range g {
			types[t] = true
		}
	}
	metrics := make([]task82b.F2Metric, 0, len(task82b.AllMetricIDs()))
	for _, id := range task82b.AllMetricIDs() {
		metrics = append(metrics, task82b.F2Metric{MetricID: id, MissingReason: "F2_EXTRACTOR_ERROR: " + err.Error()})
	}
	return task82b.F2Vector{JobID: jobID, CorpusID: corpusID, Tokens: tokens, Types: len(types), Lines: len(groups), Metrics: metrics}, false
}

// Run executes one job and returns its persisted record.
func (j Job) Run(root, workDir string) (JobRecord, error) {
	if j.Kind == "shorthand_variant" {
		return j.runShorthand(root, workDir)
	}
	return j.runExtraction(root, workDir)
}
