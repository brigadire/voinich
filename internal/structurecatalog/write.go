package structurecatalog

import (
	"bufio"
	"compress/gzip"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

func writeTSV(path string, header []string, rows [][]string) error {
	f, e := os.Create(path)
	if e != nil {
		return e
	}
	defer f.Close()
	w := bufio.NewWriter(f)
	defer w.Flush()
	if _, e = w.WriteString(strings.Join(header, "\t") + "\n"); e != nil {
		return e
	}
	for _, r := range rows {
		for i := range r {
			r[i] = clean(r[i])
		}
		if _, e = w.WriteString(strings.Join(r, "\t") + "\n"); e != nil {
			return e
		}
	}
	return nil
}
func clean(s string) string { return strings.NewReplacer("\t", " ", "\r", " ", "\n", " ").Replace(s) }
func si(x int) string       { return strconv.Itoa(x) }
func sf(x float64) string {
	if math.IsNaN(x) {
		return "NA"
	}
	if math.IsInf(x, 1) {
		return "Inf"
	}
	if math.IsInf(x, -1) {
		return "-Inf"
	}
	return strconv.FormatFloat(x, 'g', 10, 64)
}
func ss(xs []string) string { return strings.Join(xs, "|") }

func writeRules(path string, rules []Rule) error {
	h := []string{"rule_id", "level", "rule_type", "lhs", "rhs", "context", "observed_count", "opportunity_count", "observed_probability", "expected_count", "effect_size", "p_raw", "q_value", "observed_status", "corpus_rule", "inferred_status", "transcription_stability", "provenance"}
	rows := make([][]string, 0, len(rules))
	for _, r := range rules {
		rows = append(rows, []string{r.RuleID, r.Level, r.RuleType, r.LHS, r.RHS, r.Context, si(r.ObservedCount), si(r.OpportunityCount), sf(r.ObservedProbability), sf(r.ExpectedCount), sf(r.EffectSize), sf(r.PRaw), sf(r.QValue), r.ObservedStatus, r.CorpusRule, r.InferredStatus, r.Stability, r.Provenance})
	}
	return writeTSV(path, h, rows)
}

func writeComplement(path string, tokens []string, observed map[string]map[string]int) error {
	f, e := os.Create(path)
	if e != nil {
		return e
	}
	defer f.Close()
	gz := gzip.NewWriter(f)
	defer gz.Close()
	enc := json.NewEncoder(gz)
	if _, e = gz.Write([]byte("{\"schema_version\":\"" + SchemaVersion + "\",\"representation\":\"sorted token universe plus inclusive missing-rhs index ranges\",\"tokens\":")); e != nil {
		return e
	}
	tokenJSON, e := json.Marshal(tokens)
	if e != nil {
		return e
	}
	if _, e = gz.Write(tokenJSON); e != nil {
		return e
	}
	if _, e = gz.Write([]byte(",\"rows\":[\n")); e != nil {
		return e
	}
	for i, a := range tokens {
		ranges := [][2]int{}
		start := -1
		for j, b := range tokens {
			missing := observed[a][b] == 0
			if missing && start < 0 {
				start = j
			}
			if !missing && start >= 0 {
				ranges = append(ranges, [2]int{start, j - 1})
				start = -1
			}
		}
		if start >= 0 {
			ranges = append(ranges, [2]int{start, len(tokens) - 1})
		}
		if i > 0 {
			if _, e = gz.Write([]byte(",\n")); e != nil {
				return e
			}
		}
		if e = enc.Encode(struct {
			LHSIndex      int      `json:"lhs_index"`
			MissingRanges [][2]int `json:"missing_rhs_index_ranges"`
		}{i, ranges}); e != nil {
			return e
		}
	}
	_, e = gz.Write([]byte("]}\n"))
	return e
}

func ensureDir(path string) error { return os.MkdirAll(filepath.Clean(path), 0755) }
func provenance(c Corpus) string {
	return fmt.Sprintf("%s;sha256=%s;transcription=%s;schema=%s", c.Path, c.SHA, c.Transcription, SchemaVersion)
}

func (cat *Catalog) writeManifest() error {
	entries, err := os.ReadDir(cat.Config.OutputDir)
	if err != nil {
		return err
	}
	rows := [][]string{}
	for _, entry := range entries {
		if entry.IsDir() || entry.Name() == "CATALOG_MANIFEST.tsv" {
			continue
		}
		path := filepath.Join(cat.Config.OutputDir, entry.Name())
		f, err := os.Open(path)
		if err != nil {
			return err
		}
		h := sha256.New()
		n, err := io.Copy(h, f)
		f.Close()
		if err != nil {
			return err
		}
		rows = append(rows, []string{entry.Name(), fmt.Sprintf("%x", h.Sum(nil)), fmt.Sprintf("%d", n)})
	}
	return writeTSV(filepath.Join(cat.Config.OutputDir, "CATALOG_MANIFEST.tsv"), []string{"path", "sha256", "size_bytes"}, rows)
}
