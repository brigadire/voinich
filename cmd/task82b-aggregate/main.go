// Command task82b-aggregate reads every raw JSON job record written by
// cmd/task82b-run and regenerates every Task82b TSV/JSON/Markdown
// deliverable from those observed values (tasks_ph2/task82b.txt sec.68).
// It performs no F2/AX computation of its own except the cheap,
// deterministic BDD pair/registry/SX recomputation (ExtractTEIPairs is
// pure and fast; there is no reason to also persist it as a "raw job").
package main

import (
	"encoding/json"
	"flag"
	"log"
	"os"
	"path/filepath"

	"zcore.dev/voinich/internal/task82b"
)

// Rec mirrors cmd/task82b-run's JobRecord JSON schema exactly (kept as an
// independent type so this command has no import edge to a `package
// main` it cannot import anyway).
type Rec struct {
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

func (r Rec) metric(id string) (float64, bool) {
	for _, m := range r.F2.Metrics {
		if m.MetricID == id {
			return m.Value, m.Available
		}
	}
	return 0, false
}

func main() {
	root := flag.String("root", ".", "repository root")
	out := flag.String("out", "research/phase2/task82b", "output directory")
	flag.Parse()
	if err := run(*root, *out); err != nil {
		log.Fatal(err)
	}
}

func run(root, out string) error {
	recs, err := loadRecords(filepath.Join(out, "raw"))
	if err != nil {
		return err
	}
	log.Printf("loaded %d raw job records", len(recs))

	pairs, _, err := loadBDD(root)
	if err != nil {
		log.Printf("WARNING: BDD pairs unavailable for aggregation-time registry rebuild: %v", err)
	}

	if err := writeShorthandOutputs(out, recs, pairs); err != nil {
		return err
	}
	if err := writeExtractionOutputs(out, recs); err != nil {
		return err
	}
	if err := writeAXOutputs(out, recs); err != nil {
		return err
	}
	if err := writeCrossBranchOutputs(out, recs); err != nil {
		return err
	}
	if err := writeManifestsAndReport(out, recs, pairs); err != nil {
		return err
	}
	log.Println("task82b-aggregate complete")
	return nil
}

func loadRecords(rawDir string) ([]Rec, error) {
	entries, err := os.ReadDir(rawDir)
	if err != nil {
		return nil, err
	}
	var recs []Rec
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		b, err := os.ReadFile(filepath.Join(rawDir, e.Name()))
		if err != nil {
			return nil, err
		}
		var r Rec
		if err := json.Unmarshal(b, &r); err != nil {
			return nil, err
		}
		recs = append(recs, r)
	}
	return recs, nil
}

func loadBDD(root string) ([]task82b.PairUnit, task82b.CharDeletionStats, error) {
	paths, err := filepath.Glob(filepath.Join(root, "data_test/bdd-tei/koeln-edd-c-119/*.xml"))
	if err != nil {
		return nil, task82b.CharDeletionStats{}, err
	}
	if len(paths) == 0 {
		return nil, task82b.CharDeletionStats{}, os.ErrNotExist
	}
	res, err := task82b.ExtractTEIPairs(paths)
	if err != nil {
		return nil, task82b.CharDeletionStats{}, err
	}
	return res.Pairs, task82b.BuildCharDeletionStats(res.Pairs), nil
}
