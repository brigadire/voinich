package g1v2

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

var ScientificStatuses = map[string]bool{
	"FITTED": true, "TRAINING_FAILED": true, "INDUCTION_LIMIT_REACHED": true,
	"CONVERGENCE_FAILED": true, "NUMERICALLY_UNSTABLE": true,
	"METRIC_AVAILABLE": true, "METRIC_UNAVAILABLE": true, "METRIC_NONFINITE": true,
	"GENERATION_SUCCESS": true, "GENERATION_FAILED": true,
	"GENERATION_NOT_REACHED": true, "STRUCTURAL_PASS": true,
	"STRUCTURAL_FAIL": true, "STRUCTURAL_NOT_ASSESSABLE": true,
	"EVIDENCE_COMPLETE": true,
}

type Artifact struct {
	Name string          `json:"name"`
	Hash string          `json:"hash"`
	Data json.RawMessage `json:"data"`
}

// ScientificResult is the immutable, content-addressed scientific payload.
// It intentionally has no worker, host, lease, attempt, or time field.
type ScientificResult struct {
	SchemaVersion    string     `json:"schema_version"`
	ProducingJobID   string     `json:"producing_job_id"`
	InputHashes      []string   `json:"input_hashes"`
	DependencyHashes []string   `json:"dependency_hashes"`
	CodeHash         string     `json:"code_hash"`
	ConfigHash       string     `json:"config_hash"`
	Seed             uint64     `json:"seed"`
	ScientificStatus string     `json:"scientific_status"`
	Artifacts        []Artifact `json:"artifacts"`
}

type Telemetry struct {
	Worker               string   `json:"worker"`
	Host                 string   `json:"host"`
	LeaseHistory         []string `json:"lease_history"`
	StartUTC             string   `json:"start_utc"`
	EndUTC               string   `json:"end_utc"`
	WallSeconds          float64  `json:"wall_seconds"`
	CPUSeconds           float64  `json:"cpu_seconds"`
	PeakRSSBytes         int64    `json:"peak_rss_bytes"`
	TransferBytes        int64    `json:"transfer_bytes"`
	RetryCount           int      `json:"retry_count"`
	InfrastructureStatus string   `json:"infrastructure_status"`
}

type CopyRecord struct {
	ResultHash    string `json:"result_hash"`
	TelemetryHash string `json:"telemetry_hash"`
	Worker        string `json:"worker"`
	PublishedUTC  string `json:"published_utc"`
}

type IndexRecord struct {
	JobID      string       `json:"job_id"`
	Status     string       `json:"status"` // VERIFIED or CONFLICT
	ResultHash string       `json:"result_hash,omitempty"`
	Copies     []CopyRecord `json:"copies"`
}

func (r ScientificResult) canonical() ([]byte, error) { return canonicalJSON(r) }

func (r ScientificResult) Validate(j JobBundle) error {
	if err := j.Validate(); err != nil {
		return err
	}
	if r.SchemaVersion != SchemaVersion || r.ProducingJobID != j.JobID || r.CodeHash != j.CodeHash || r.ConfigHash != j.ConfigHash || r.Seed != j.Seed {
		return fmt.Errorf("result identity/closure mismatch")
	}
	if strings.Join(r.InputHashes, ",") != strings.Join(j.InputHashes, ",") || strings.Join(r.DependencyHashes, ",") != strings.Join(j.DependencyHashes, ",") {
		return fmt.Errorf("result input/dependency closure mismatch")
	}
	if !ScientificStatuses[r.ScientificStatus] {
		return fmt.Errorf("invalid scientific status %q", r.ScientificStatus)
	}
	if len(r.Artifacts) == 0 {
		return errors.New("missing artifact")
	}
	names := map[string]bool{}
	for _, a := range r.Artifacts {
		if a.Name == "" || names[a.Name] {
			return fmt.Errorf("empty/duplicate artifact name")
		}
		names[a.Name] = true
		if HashBytes(a.Data) != a.Hash {
			return fmt.Errorf("artifact %s content hash mismatch", a.Name)
		}
	}
	return nil
}

type Store struct{ Root string }

func (s Store) ensure() error {
	for _, d := range []string{"objects", "telemetry", "index", "quarantine", "tmp"} {
		if err := os.MkdirAll(filepath.Join(s.Root, d), 0755); err != nil {
			return err
		}
	}
	return nil
}

func atomicWrite(path string, data []byte) error {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	f, err := os.CreateTemp(filepath.Dir(path), ".publish-*")
	if err != nil {
		return err
	}
	tmp := f.Name()
	ok := false
	defer func() {
		_ = f.Close()
		if !ok {
			_ = os.Remove(tmp)
		}
	}()
	if _, err = f.Write(data); err != nil {
		return err
	}
	if err = f.Sync(); err != nil {
		return err
	}
	if err = f.Close(); err != nil {
		return err
	}
	if err = os.Rename(tmp, path); err != nil {
		return err
	}
	ok = true
	d, err := os.Open(filepath.Dir(path))
	if err == nil {
		_ = d.Sync()
		_ = d.Close()
	}
	return nil
}

func (s Store) indexPath(id string) string { return filepath.Join(s.Root, "index", id+".json") }
func (s Store) objectPath(h string) string { return filepath.Join(s.Root, "objects", h[:2], h+".json") }

func (s Store) ReadIndex(id string) (IndexRecord, error) {
	b, err := os.ReadFile(s.indexPath(id))
	if err != nil {
		return IndexRecord{}, err
	}
	var x IndexRecord
	if err := json.Unmarshal(b, &x); err != nil {
		return x, err
	}
	return x, nil
}

func (s Store) ReadResult(hash string) (ScientificResult, error) {
	b, err := os.ReadFile(s.objectPath(hash))
	if err != nil {
		return ScientificResult{}, err
	}
	if HashBytes(b) != hash {
		return ScientificResult{}, fmt.Errorf("stored result content hash mismatch")
	}
	var r ScientificResult
	if err := json.Unmarshal(b, &r); err != nil {
		return r, err
	}
	return r, nil
}

// Publish validates before its atomic object publication and atomic JobID
// index update. Identical duplicates retain every provenance copy. A conflict
// atomically changes the index to CONFLICT and preserves all copies.
func (s Store) Publish(j JobBundle, r ScientificResult, t Telemetry) (IndexRecord, error) {
	if err := s.ensure(); err != nil {
		return IndexRecord{}, err
	}
	if err := r.Validate(j); err != nil {
		return IndexRecord{}, err
	}
	rb, err := r.canonical()
	if err != nil {
		return IndexRecord{}, err
	}
	rh := HashBytes(rb)
	tb, err := canonicalJSON(t)
	if err != nil {
		return IndexRecord{}, err
	}
	th := HashBytes(tb)
	if err := atomicWrite(s.objectPath(rh), rb); err != nil {
		return IndexRecord{}, err
	}
	if err := atomicWrite(filepath.Join(s.Root, "telemetry", th+".json"), tb); err != nil {
		return IndexRecord{}, err
	}
	x := IndexRecord{JobID: j.JobID, Status: "VERIFIED", ResultHash: rh}
	if old, e := s.ReadIndex(j.JobID); e == nil {
		x = old
	}
	copy := CopyRecord{rh, th, t.Worker, time.Now().UTC().Format(time.RFC3339Nano)}
	for _, c := range x.Copies {
		if c.TelemetryHash == th && c.ResultHash == rh {
			return x, nil
		}
	}
	x.Copies = append(x.Copies, copy)
	sort.Slice(x.Copies, func(i, k int) bool {
		if x.Copies[i].Worker == x.Copies[k].Worker {
			return x.Copies[i].TelemetryHash < x.Copies[k].TelemetryHash
		}
		return x.Copies[i].Worker < x.Copies[k].Worker
	})
	if x.ResultHash != "" && x.ResultHash != rh {
		x.Status = "CONFLICT"
		x.ResultHash = ""
	}
	if x.Status == "CONFLICT" {
		qb, _ := canonicalJSON(x)
		if err := atomicWrite(filepath.Join(s.Root, "quarantine", j.JobID+".json"), qb); err != nil {
			return x, err
		}
	}
	ib, _ := canonicalJSON(x)
	if err := atomicWrite(s.indexPath(j.JobID), ib); err != nil {
		return x, err
	}
	if x.Status == "CONFLICT" {
		return x, fmt.Errorf("conflicting duplicate for %s", j.JobID)
	}
	return x, nil
}

func (s Store) Completed(id string) bool {
	x, err := s.ReadIndex(id)
	return err == nil && x.Status == "VERIFIED" && ValidHash(x.ResultHash)
}
