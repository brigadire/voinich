package main

import (
	"bufio"
	"crypto/sha256"
	"encoding/csv"
	"encoding/hex"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

type schema struct {
	Version     string       `json:"version"`
	Frozen      bool         `json:"frozen"`
	Descriptors []descriptor `json:"descriptors"`
}
type descriptor struct {
	ID            string   `json:"id"`
	AllowedValues []string `json:"allowed_values"`
}
type resultManifest struct {
	SchemaVersion string `json:"schema_version"`
	Counts        struct {
		Total              int `json:"TOTAL_IVTFF_UNITS"`
		FullyAnnotated     int `json:"FULLY_ANNOTATED"`
		PartiallyAnnotated int `json:"PARTIALLY_ANNOTATED"`
		UnresolvablePanel  int `json:"UNRESOLVABLE_PANEL"`
		ImageMissing       int `json:"IMAGE_MISSING"`
		PanelPending       int `json:"PANEL_MAP_PENDING_COUNT"`
	} `json:"coverage"`
	Status map[string]bool `json:"status"`
}

var panelID = regexp.MustCompile(`^f[0-9]+[rv][0-9]+$`)
var leafID = regexp.MustCompile(`^(f[0-9]+)[rv]`)

func main() {
	base := flag.String("base", "research/visual_descriptors", "bundle directory")
	flag.Parse()
	if err := validateBundle(*base); err != nil {
		fmt.Fprintln(os.Stderr, "VISUAL_DESCRIPTOR_VALIDATION_PASS=false:", err)
		os.Exit(1)
	}
	fmt.Println("VISUAL_DESCRIPTOR_VALIDATION_PASS=true")
}

func validateBundle(base string) error {
	s, allowed, err := loadSchema(filepath.Join(base, "VISUAL_FEATURE_SCHEMA.json"))
	if err != nil {
		return err
	}
	if s.Version != "1.0.0" || !s.Frozen {
		return errors.New("schema is not frozen 1.0.0")
	}
	if err := validateFreeze(base); err != nil {
		return err
	}
	units, err := loadUnits(filepath.Join(base, "VISUAL_ANNOTATION_UNIT_REGISTRY.tsv"))
	if err != nil {
		return err
	}
	if len(units) != 227 {
		return fmt.Errorf("unit registry %d/227", len(units))
	}
	panels, err := loadPanels(filepath.Join(base, "VISUAL_PANEL_CROP_REGISTRY.tsv"), units)
	if err != nil {
		return err
	}
	data, counts, err := loadData(filepath.Join(base, "VISUAL_PAGE_DESCRIPTORS.tsv"), s, allowed, units, panels)
	if err != nil {
		return err
	}
	if err := loadProvenance(filepath.Join(base, "VISUAL_PAGE_DESCRIPTOR_PROVENANCE.tsv"), s, units, panels); err != nil {
		return err
	}
	if err := loadAdjudication(filepath.Join(base, "VISUAL_DESCRIPTOR_ADJUDICATION.tsv"), s, allowed, data); err != nil {
		return err
	}
	if err := validateReliability(filepath.Join(base, "VISUAL_DESCRIPTOR_PILOT_RELIABILITY.tsv"), s); err != nil {
		return err
	}
	return validateManifest(filepath.Join(base, "VISUAL_DESCRIPTOR_RESULTS_MANIFEST.json"), s, counts)
}
func loadSchema(path string) (schema, map[string]map[string]bool, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return schema{}, nil, err
	}
	var s schema
	if err := json.Unmarshal(b, &s); err != nil {
		return s, nil, err
	}
	if len(s.Descriptors) == 0 {
		return s, nil, errors.New("empty schema")
	}
	a := map[string]map[string]bool{}
	for _, d := range s.Descriptors {
		if d.ID == "" || a[d.ID] != nil {
			return s, nil, fmt.Errorf("duplicate descriptor %q", d.ID)
		}
		a[d.ID] = map[string]bool{}
		for _, v := range d.AllowedValues {
			a[d.ID][v] = true
		}
	}
	return s, a, nil
}
func validateFreeze(base string) error {
	f, err := os.Open(filepath.Join(base, "VISUAL_DESCRIPTOR_FREEZE_SHA256SUMS"))
	if err != nil {
		return err
	}
	defer f.Close()
	seen := map[string]bool{}
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		x := strings.Fields(sc.Text())
		if len(x) != 2 || filepath.Base(x[1]) != x[1] {
			return errors.New("bad freeze checksum line")
		}
		b, err := os.ReadFile(filepath.Join(base, x[1]))
		if err != nil {
			return err
		}
		sum := sha256.Sum256(b)
		if hex.EncodeToString(sum[:]) != x[0] {
			return fmt.Errorf("freeze hash mismatch: %s", x[1])
		}
		seen[x[1]] = true
	}
	if err := sc.Err(); err != nil {
		return err
	}
	for _, n := range []string{"VISUAL_FEATURE_SCHEMA.json", "VISUAL_FEATURE_SCHEMA.md", "VISUAL_DESCRIPTOR_ANNOTATION_PROTOCOL.md", "VISUAL_PANEL_CROP_REGISTRY.tsv"} {
		if !seen[n] {
			return fmt.Errorf("missing freeze hash %s", n)
		}
	}
	return nil
}
func loadUnits(path string) (map[string]string, error) {
	r, err := readTSV(path)
	if err != nil {
		return nil, err
	}
	if len(r) < 2 || strings.Join(r[0], "\t") != "page_id\tphysical_leaf_id" {
		return nil, errors.New("bad unit registry")
	}
	out := map[string]string{}
	for _, row := range r[1:] {
		if len(row) != 2 || row[0] == "" || out[row[0]] != "" || wantLeaf(row[0]) != row[1] {
			return nil, fmt.Errorf("bad/duplicate unit %q", row[0])
		}
		out[row[0]] = row[1]
	}
	return out, nil
}

type panel struct{ status, oid, crop string }

func loadPanels(path string, units map[string]string) (map[string]panel, error) {
	r, err := readTSV(path)
	if err != nil {
		return nil, err
	}
	head := "page_id\tphysical_leaf_id\tyale_canvas_oid\tcrop_locator\tmapping_status\tmapping_confidence\tadjudication_note"
	if len(r) < 2 || strings.Join(r[0], "\t") != head {
		return nil, errors.New("bad panel registry")
	}
	out := map[string]panel{}
	for _, row := range r[1:] {
		if len(row) != 7 || !panelID.MatchString(row[0]) || out[row[0]].status != "" || units[row[0]] != row[1] {
			return nil, fmt.Errorf("bad panel %q", row[0])
		}
		if row[4] != "VERIFIED" && row[4] != "AMBIGUOUS" && row[4] != "UNRESOLVABLE" {
			return nil, fmt.Errorf("panel status %s", row[4])
		}
		if !strings.HasPrefix(row[3], "pct:") {
			return nil, fmt.Errorf("bad crop %s", row[0])
		}
		out[row[0]] = panel{row[4], row[2], row[3]}
	}
	for id := range units {
		if panelID.MatchString(id) {
			if _, ok := out[id]; !ok {
				return nil, fmt.Errorf("missing panel %s", id)
			}
		}
	}
	return out, nil
}

type coverage struct{ total, fully, partial, unresolvable, missing int }

func loadData(path string, s schema, allowed map[string]map[string]bool, units map[string]string, panels map[string]panel) (map[string]map[string]string, coverage, error) {
	r, err := readTSV(path)
	if err != nil {
		return nil, coverage{}, err
	}
	req := []string{"page_id", "physical_leaf_id", "image_source_id", "schema_version", "annotation_status"}
	if len(r) < 2 || len(r[0]) != len(req)+len(s.Descriptors) {
		return nil, coverage{}, errors.New("unexpected data columns")
	}
	for i, v := range req {
		if r[0][i] != v {
			return nil, coverage{}, fmt.Errorf("metadata column %d", i)
		}
	}
	for i, d := range s.Descriptors {
		if r[0][len(req)+i] != d.ID {
			return nil, coverage{}, fmt.Errorf("descriptor column %s", d.ID)
		}
	}
	if err := forbiddenColumns(r[0]); err != nil {
		return nil, coverage{}, err
	}
	out := map[string]map[string]string{}
	c := coverage{total: len(r) - 1}
	for _, row := range r[1:] {
		if len(row) != len(r[0]) || out[row[0]] != nil || units[row[0]] != row[1] || row[3] != s.Version {
			return nil, c, fmt.Errorf("bad data row %q", row[0])
		}
		vals := map[string]string{}
		nv, im := 0, 0
		for i, d := range s.Descriptors {
			v := row[len(req)+i]
			if !allowed[d.ID][v] {
				return nil, c, fmt.Errorf("%s/%s invalid %q", row[0], d.ID, v)
			}
			vals[d.ID] = v
			if v == "NOT_VISIBLE" {
				nv++
			}
			if v == "IMAGE_MISSING" {
				im++
			}
		}
		switch row[4] {
		case "FULLY_ANNOTATED":
			if nv+im > 0 {
				return nil, c, fmt.Errorf("%s falsely full", row[0])
			}
			c.fully++
		case "PARTIALLY_ANNOTATED":
			// An image may be present but entirely non-diagnostic; retain the
			// unit as partial rather than misclassifying it as missing.
			if nv+im == 0 || (nv+im == len(s.Descriptors) && im == len(s.Descriptors)) {
				return nil, c, fmt.Errorf("%s bad partial", row[0])
			}
			c.partial++
		case "UNRESOLVABLE_PANEL":
			if p, ok := panels[row[0]]; !ok || p.status != "UNRESOLVABLE" || nv != len(s.Descriptors) {
				return nil, c, fmt.Errorf("%s bad unresolvable", row[0])
			}
			c.unresolvable++
		case "IMAGE_MISSING":
			if im != len(s.Descriptors) {
				return nil, c, fmt.Errorf("%s bad image missing", row[0])
			}
			c.missing++
		default:
			return nil, c, fmt.Errorf("%s pending/invalid status %s", row[0], row[4])
		}
		out[row[0]] = vals
	}
	if len(out) != len(units) {
		return nil, c, fmt.Errorf("data coverage %d/%d", len(out), len(units))
	}
	return out, c, nil
}
func loadProvenance(path string, s schema, units map[string]string, panels map[string]panel) error {
	r, err := readTSV(path)
	if err != nil {
		return err
	}
	head := []string{"page_id", "physical_leaf_id", "yale_canvas_oid", "source_image_locator", "source_version", "image_checksum_if_available", "annotator_pass", "annotation_timestamp", "schema_version", "crop_locator", "adjudication_status"}
	if len(r) < 2 || strings.Join(r[0], "\x00") != strings.Join(head, "\x00") {
		return errors.New("unexpected provenance columns")
	}
	seen := map[string]bool{}
	for _, row := range r[1:] {
		if len(row) != len(head) || seen[row[0]] || units[row[0]] != row[1] {
			return fmt.Errorf("bad provenance %q", row[0])
		}
		for _, v := range row {
			if v == "" {
				return fmt.Errorf("empty provenance %s", row[0])
			}
		}
		if row[8] != s.Version {
			return fmt.Errorf("provenance version %s", row[0])
		}
		if p, ok := panels[row[0]]; ok {
			if row[2] != p.oid || row[9] != p.crop {
				return fmt.Errorf("panel provenance %s", row[0])
			}
		} else if row[9] != "full" {
			return fmt.Errorf("ordinary crop %s", row[0])
		}
		seen[row[0]] = true
	}
	if len(seen) != len(units) {
		return fmt.Errorf("provenance coverage %d", len(seen))
	}
	return nil
}
func loadAdjudication(path string, s schema, allowed map[string]map[string]bool, data map[string]map[string]string) error {
	r, err := readTSV(path)
	if err != nil {
		return err
	}
	head := "page_id\tdescriptor_id\tpass1\tpass2\tfinal_value\tadjudication_rule\treviewer_note"
	if len(r) < 2 || strings.Join(r[0], "\t") != head {
		return errors.New("bad adjudication header")
	}
	seen := map[string]bool{}
	for _, row := range r[1:] {
		if len(row) != 7 {
			return errors.New("malformed adjudication")
		}
		key := row[0] + "\x00" + row[1]
		if seen[key] || data[row[0]] == nil || allowed[row[1]] == nil || !allowed[row[1]][row[2]] || !allowed[row[1]][row[4]] {
			return fmt.Errorf("bad adjudication %s/%s", row[0], row[1])
		}
		if row[3] != "NOT_AUDITED" && !allowed[row[1]][row[3]] {
			return fmt.Errorf("bad pass2 %s/%s", row[0], row[1])
		}
		if data[row[0]][row[1]] != row[4] || row[5] == "" || row[6] == "" {
			return fmt.Errorf("adjudication mismatch %s/%s", row[0], row[1])
		}
		seen[key] = true
	}
	want := len(data) * len(s.Descriptors)
	if len(seen) != want {
		return fmt.Errorf("adjudication coverage %d/%d", len(seen), want)
	}
	return nil
}
func validateReliability(path string, s schema) error {
	r, err := readTSV(path)
	if err != nil {
		return err
	}
	if len(r) < 2 {
		return errors.New("empty reliability")
	}
	idx := map[string]int{}
	for i, h := range r[0] {
		idx[h] = i
	}
	seen := map[string]bool{}
	for _, row := range r[1:] {
		if row[idx["decision"]] == "RETAIN" {
			seen[row[idx["descriptor_id"]]] = true
		}
	}
	for _, d := range s.Descriptors {
		if !seen[d.ID] {
			return fmt.Errorf("reliability failed %s", d.ID)
		}
	}
	return nil
}
func validateManifest(path string, s schema, c coverage) error {
	b, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	var m resultManifest
	if err := json.Unmarshal(b, &m); err != nil {
		return err
	}
	if m.SchemaVersion != s.Version || m.Counts.Total != c.total || m.Counts.FullyAnnotated != c.fully || m.Counts.PartiallyAnnotated != c.partial || m.Counts.UnresolvablePanel != c.unresolvable || m.Counts.ImageMissing != c.missing || m.Counts.PanelPending != 0 {
		return errors.New("manifest version/coverage mismatch")
	}
	for _, k := range []string{"IMAGE_SOURCE_FROZEN", "ANNOTATION_UNIT_FROZEN", "VISUAL_FEATURE_SCHEMA_FROZEN", "ANNOTATION_PROTOCOL_FROZEN", "BLINDNESS_CONSTRAINT_SATISFIED", "PILOT_COMPLETED", "RELIABILITY_GATE_PASSED", "FULL_VISUAL_ANNOTATION_COMPLETED", "VISUAL_DESCRIPTOR_VALIDATION_PASS", "VISUAL_PAGE_DESCRIPTORS_READY", "LEVEL_C_VISUAL_CONTEXT_TEST_AUTHORIZED"} {
		if !m.Status[k] {
			return fmt.Errorf("manifest false %s", k)
		}
	}
	for _, k := range []string{"TEXTUAL_ASSOCIATION_ANALYSIS_PERFORMED", "TEXTUAL_RESULTS_USED_FOR_DESCRIPTOR_SELECTION", "LEVEL_C_VISUAL_CONTEXT_TEST_EXECUTED"} {
		if m.Status[k] {
			return fmt.Errorf("manifest prohibited true %s", k)
		}
	}
	return nil
}
func forbiddenColumns(head []string) error {
	for _, h := range head {
		l := strings.ToLower(h)
		for _, bad := range []string{"currier", "scribe", "quire", "token_count", "line_count", "entropy", "fingerprint", "textual_metric", "section_result"} {
			if strings.Contains(l, bad) {
				return fmt.Errorf("prohibited field %q", h)
			}
		}
	}
	return nil
}
func wantLeaf(id string) string {
	if id == "fRos" {
		return "f85-f86-foldout"
	}
	m := leafID.FindStringSubmatch(id)
	if len(m) == 2 {
		return m[1]
	}
	return ""
}
func readTSV(path string) ([][]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	r := csv.NewReader(f)
	r.Comma = '\t'
	r.FieldsPerRecord = -1
	return r.ReadAll()
}
