// Package notation implements the corpus-neutral preparation pipeline for the
// comparative notation study. It deliberately contains no Voynich- or
// corpus-specific branches.
package notation

import (
	"bufio"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"sort"
	"strings"
)

const SchemaVersion = "usc-1.0"

type ObservedLevel struct {
	Value    string `json:"value,omitempty"`
	Observed bool   `json:"observed"`
}

// Record is one ordered token in the Universal Structural Corpus (USC).
// An unobserved physical level is represented by Observed=false and an empty
// value; adapters must never invent such boundaries.
type Record struct {
	SchemaVersion  string            `json:"schema_version"`
	CorpusID       string            `json:"corpus_id"`
	Representation string            `json:"representation_id"`
	Document       ObservedLevel     `json:"document"`
	Section        ObservedLevel     `json:"section"`
	Page           ObservedLevel     `json:"page"`
	Locus          ObservedLevel     `json:"locus"`
	PhysicalLine   ObservedLevel     `json:"physical_line"`
	TokenID        string            `json:"token_id"`
	TokenIndex     int               `json:"token_index"`
	Token          string            `json:"token"`
	Symbols        []string          `json:"symbols"`
	Attributes     map[string]string `json:"attributes,omitempty"`
}

type Status string

const (
	Comparable          Status = "COMPARABLE"
	PartiallyComparable Status = "PARTIALLY_COMPARABLE"
	NotComparable       Status = "NOT_COMPARABLE"
)

type Metric struct {
	MetricID string  `json:"metric_id"`
	Family   string  `json:"family"`
	Regime   string  `json:"regime,omitempty"`
	Value    float64 `json:"value,omitempty"`
	Status   Status  `json:"status"`
	Reason   string  `json:"reason,omitempty"`
}

type Fingerprint struct {
	SchemaVersion  string            `json:"schema_version"`
	CorpusID       string            `json:"corpus_id"`
	Representation string            `json:"representation_id"`
	InputSHA256    string            `json:"input_sha256"`
	RecordCount    int               `json:"record_count"`
	Metrics        []Metric          `json:"metrics"`
	Curves         []CurvePoint      `json:"curves"`
	Metadata       map[string]string `json:"metadata,omitempty"`
}

type CurvePoint struct {
	CurveID    string  `json:"curve_id"`
	Checkpoint int     `json:"checkpoint"`
	Value      float64 `json:"value,omitempty"`
	Status     Status  `json:"status"`
	Reason     string  `json:"reason,omitempty"`
}

func ReadJSONL(r io.Reader) ([]Record, error) {
	s := bufio.NewScanner(r)
	s.Buffer(make([]byte, 64*1024), 16*1024*1024)
	var out []Record
	line := 0
	for s.Scan() {
		line++
		if strings.TrimSpace(s.Text()) == "" {
			continue
		}
		var rec Record
		if err := json.Unmarshal(s.Bytes(), &rec); err != nil {
			return nil, fmt.Errorf("USC line %d: %w", line, err)
		}
		out = append(out, rec)
	}
	if err := s.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

func WriteJSONL(w io.Writer, records []Record) error {
	enc := json.NewEncoder(w)
	enc.SetEscapeHTML(false)
	for _, rec := range records {
		if err := enc.Encode(rec); err != nil {
			return err
		}
	}
	return nil
}

func CanonicalSHA256(records []Record) (string, error) {
	var b strings.Builder
	if err := WriteJSONL(&b, records); err != nil {
		return "", err
	}
	sum := sha256.Sum256([]byte(b.String()))
	return hex.EncodeToString(sum[:]), nil
}

// Validate enforces identity, hierarchy, ordering, and symbol-boundary rules.
func Validate(records []Record) error {
	if len(records) == 0 {
		return fmt.Errorf("USC corpus is empty")
	}
	ids := make(map[string]bool, len(records))
	corpus, representation := records[0].CorpusID, records[0].Representation
	lastIndex := map[string]int{}
	for i, rec := range records {
		if rec.SchemaVersion != SchemaVersion {
			return fmt.Errorf("record %d: schema_version=%q", i, rec.SchemaVersion)
		}
		if rec.CorpusID == "" || rec.CorpusID != corpus {
			return fmt.Errorf("record %d: inconsistent corpus_id", i)
		}
		if rec.Representation == "" || rec.Representation != representation {
			return fmt.Errorf("record %d: inconsistent representation_id", i)
		}
		if !rec.Document.Observed || rec.Document.Value == "" {
			return fmt.Errorf("record %d: document must be observed", i)
		}
		for name, level := range map[string]ObservedLevel{"section": rec.Section, "page": rec.Page, "locus": rec.Locus, "physical_line": rec.PhysicalLine} {
			if level.Observed != (level.Value != "") {
				return fmt.Errorf("record %d: %s observed/value mismatch", i, name)
			}
		}
		if rec.TokenID == "" || ids[rec.TokenID] {
			return fmt.Errorf("record %d: missing or duplicate token_id %q", i, rec.TokenID)
		}
		ids[rec.TokenID] = true
		if rec.TokenIndex < 0 || rec.Token == "" || len(rec.Symbols) == 0 {
			return fmt.Errorf("record %d: invalid token/index/symbols", i)
		}
		key := rec.Document.Value + "\x1f" + rec.Section.Value + "\x1f" + rec.Page.Value + "\x1f" + rec.Locus.Value + "\x1f" + rec.PhysicalLine.Value
		if prev, ok := lastIndex[key]; ok && rec.TokenIndex != prev+1 {
			return fmt.Errorf("record %d: non-contiguous token_index in hierarchy unit", i)
		}
		if _, ok := lastIndex[key]; !ok && rec.TokenIndex != 0 {
			return fmt.Errorf("record %d: first token_index must be zero", i)
		}
		lastIndex[key] = rec.TokenIndex
	}
	return nil
}

func SortedMetrics(fp Fingerprint) []Metric {
	out := append([]Metric(nil), fp.Metrics...)
	sort.Slice(out, func(i, j int) bool {
		if out[i].Family != out[j].Family {
			return out[i].Family < out[j].Family
		}
		if out[i].MetricID != out[j].MetricID {
			return out[i].MetricID < out[j].MetricID
		}
		return out[i].Regime < out[j].Regime
	})
	return out
}
