// Package g1v2science implements the frozen G1-v2 V1.2.1 scientific protocol.
package g1v2science

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"encoding/csv"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"os"
	"sort"
	"strconv"
	"strings"
	"unicode/utf8"

	"golang.org/x/text/unicode/norm"
)

const (
	ContractVersion  = "G1_V2_EXECUTABLE_CONTRACT_V1_2_1"
	EvidenceContract = "G1V2_EVIDENCE_CONTRACT_V1_2_1"
	StatusVersion    = "G1_V2_STATUS_REACHABILITY_CONTRACT_V2"
	RNGVersion       = "G1V2-RNG-1"
	RootSeedHex      = "6f5a9c731de248b480c66b237ace215044689c5fa2f593e510b73dce18a49027"
)

var Stages = []string{"FIT", "PREDICTIVE", "GENERATION", "F2_METRIC", "COMPLEXITY", "CANDIDATE_AGGREGATION", "CONTROL_AGGREGATION"}
var Statuses = []string{"PASS", "FAIL", "NOT_ASSESSABLE", "FIT_SUCCESS", "GENERATION_SUCCESS", "COMPLEXITY_SUCCESS", "AGGREGATION_SUCCESS", "FIT_FAILURE", "NUMERICAL_FAILURE", "INDUCTION_CAP", "GENERATION_FAILURE", "PROTOCOL_VETO", "NOT_REACHED"}
var EvidenceTypes = []string{"complexity", "f2_metric", "final_verdict", "fit", "fitted_model", "generation", "minimality", "not_reached", "predictive_gate", "predictive_metric", "predictive_verdict", "scientific_failure", "structural_family", "structural_gate", "structural_verdict"}
var F2MetricIDs = []string{"EF1_GIANT_COMPONENT_SHARE", "EF1_ISOLATE_SHARE", "EF2_GLOBAL_CLUSTERING", "EF3_DEGREE_FREQUENCY_SPEARMAN", "LP1_RULE_SUPPORT_GINI", "LP4_PREFIX_ATTACHMENT_NMI", "LP4_SUFFIX_ATTACHMENT_NMI", "HR1_FOLIO_VARIANCE_SHARE", "HR1_LOCUS_VARIANCE_SHARE", "HR1_SECTION_VARIANCE_SHARE", "LS1_LINE_LENGTH_CV", "PF5_WITHIN_FOLIO_PROGRESSION"}

type Candidate struct {
	ID             string         `json:"candidate_id"`
	Model          string         `json:"model_class"`
	Route          string         `json:"route"`
	Hyper          map[string]any `json:"hyperparameters"`
	SelectionGroup string         `json:"selection_group"`
	Source         string         `json:"normative_definition"`
}
type GenerationRoute struct {
	ID, Model, Author, Primitive string
	Index                        int
	Parameters                   map[string]any
}
type Authority struct {
	Candidates  []Candidate
	Routes      []GenerationRoute
	Transitions int
	SchemaTypes map[string]string
}

func LoadAuthority(candidateTSV, generationJSON, statusJSON, schemaRegistry string) (Authority, error) {
	var a Authority
	f, err := os.Open(candidateTSV)
	if err != nil {
		return a, err
	}
	defer f.Close()
	r := csv.NewReader(f)
	r.Comma = '\t'
	r.FieldsPerRecord = -1
	head, err := r.Read()
	if err != nil {
		return a, err
	}
	cols := map[string]int{}
	for i, k := range head {
		cols[k] = i
	}
	for {
		row, e := r.Read()
		if e == io.EOF {
			break
		}
		if e != nil {
			return a, e
		}
		var h map[string]any
		if e = json.Unmarshal([]byte(row[cols["hyperparameters"]]), &h); e != nil {
			return a, e
		}
		a.Candidates = append(a.Candidates, Candidate{row[cols["candidate_id"]], row[cols["model_class"]], row[cols["route"]], h, row[cols["selection_group"]], row[cols["normative_definition"]]})
	}
	var g struct {
		ContractVersion string `json:"contract_version"`
		Routes          []struct {
			GeneratorID    string         `json:"generator_id"`
			Model          string         `json:"model"`
			Author         string         `json:"author"`
			Primitive      string         `json:"primitive"`
			GeneratorIndex int            `json:"generator_index"`
			Parameters     map[string]any `json:"parameters"`
		} `json:"routes"`
	}
	b, err := os.ReadFile(generationJSON)
	if err != nil {
		return a, err
	}
	if err = json.Unmarshal(b, &g); err != nil {
		return a, err
	}
	// Generation semantics V1 remains byte-frozen to V1.2 and is selected by I2.
	if g.ContractVersion != "G1_V2_EXECUTABLE_CONTRACT_V1_2" {
		return a, fmt.Errorf("generation authority %q", g.ContractVersion)
	}
	for _, x := range g.Routes {
		a.Routes = append(a.Routes, GenerationRoute{x.GeneratorID, x.Model, x.Author, x.Primitive, x.GeneratorIndex, x.Parameters})
	}
	var s struct {
		Version     string `json:"version"`
		Transitions []any  `json:"transitions"`
		Stages      []any  `json:"stages"`
		Statuses    []any  `json:"statuses"`
	}
	b, err = os.ReadFile(statusJSON)
	if err != nil {
		return a, err
	}
	if err = json.Unmarshal(b, &s); err != nil {
		return a, err
	}
	a.Transitions = len(s.Transitions)
	if s.Version != StatusVersion || len(s.Stages) != 7 || len(s.Statuses) != 13 {
		return a, fmt.Errorf("status authority mismatch")
	}
	var reg struct {
		ScientificContractVersion string `json:"scientific_contract_version"`
		Entries                   []struct {
			EvidenceType string `json:"evidence_type"`
			SchemaID     string `json:"schema_id"`
		} `json:"entries"`
	}
	b, err = os.ReadFile(schemaRegistry)
	if err != nil {
		return a, err
	}
	if err = json.Unmarshal(b, &reg); err != nil {
		return a, err
	}
	if reg.ScientificContractVersion != ContractVersion {
		return a, fmt.Errorf("evidence contract mismatch")
	}
	a.SchemaTypes = map[string]string{}
	for _, x := range reg.Entries {
		if x.SchemaID == "" {
			return a, fmt.Errorf("empty schema authority")
		}
		a.SchemaTypes[x.EvidenceType] = "g1v2." + x.EvidenceType + ".v1_2_1"
	}
	if len(a.Candidates) != 43 || len(a.Routes) != 12 || a.Transitions != 45 || len(a.SchemaTypes) != 15 {
		return a, fmt.Errorf("authority cardinality %d/%d/%d/%d", len(a.Candidates), len(a.Routes), a.Transitions, len(a.SchemaTypes))
	}
	return a, nil
}

func normalize(v any) any {
	switch x := v.(type) {
	case string:
		return norm.NFC.String(x)
	case []any:
		z := make([]any, len(x))
		for i := range x {
			z[i] = normalize(x[i])
		}
		return z
	case map[string]any:
		z := map[string]any{}
		for k, y := range x {
			z[norm.NFC.String(k)] = normalize(y)
		}
		return z
	default:
		return v
	}
}
func CanonicalJSON(v any) ([]byte, error) {
	raw, e := json.Marshal(v)
	if e != nil {
		return nil, e
	}
	d := json.NewDecoder(bytes.NewReader(raw))
	d.UseNumber()
	var x any
	if e = d.Decode(&x); e != nil {
		return nil, e
	}
	b, e := json.Marshal(normalize(x))
	if e != nil {
		return nil, e
	}
	return append(b, '\n'), nil
}
func Hash(b []byte) string { h := sha256.Sum256(b); return hex.EncodeToString(h[:]) }

type JobIdentity struct {
	ContractVersion   string   `json:"contract_version"`
	ControlInstanceID string   `json:"control_instance_id"`
	CandidateID       string   `json:"candidate_id"`
	Stage             string   `json:"stage"`
	ScaleOrNull       *int     `json:"scale_or_null"`
	ReplicateOrNull   *int     `json:"replicate_or_null"`
	MetricIDOrNull    *string  `json:"metric_id_or_null"`
	DependencyJobIDs  []string `json:"dependency_job_ids"`
}

func E3JobID(x JobIdentity) (string, error) {
	if x.ContractVersion != ContractVersion {
		return "", fmt.Errorf("E3 requires V1.2.1")
	}
	if x.ControlInstanceID == "" || x.CandidateID == "" || !has(Stages, x.Stage) {
		return "", fmt.Errorf("invalid identity")
	}
	if !sort.StringsAreSorted(x.DependencyJobIDs) {
		return "", fmt.Errorf("dependencies unsorted")
	}
	for i := 1; i < len(x.DependencyJobIDs); i++ {
		if x.DependencyJobIDs[i-1] == x.DependencyJobIDs[i] {
			return "", fmt.Errorf("duplicate dependency")
		}
	}
	b, e := CanonicalJSON(x)
	if e != nil {
		return "", e
	}
	h := sha256.Sum256(append([]byte("G1V2-JOB\x00"), b...))
	return "j-" + hex.EncodeToString(h[:])[:40], nil
}
func has(xs []string, x string) bool {
	for _, y := range xs {
		if x == y {
			return true
		}
	}
	return false
}

type Corpus struct {
	Tokens [][]string `json:"tokens"`
}

func NewCorpus(tokens []string) (Corpus, error) {
	c := Corpus{Tokens: make([][]string, len(tokens))}
	for i, t := range tokens {
		if t == "" || !utf8.ValidString(t) || !norm.NFC.IsNormalString(t) {
			return Corpus{}, fmt.Errorf("token %d not nonempty NFC", i)
		}
		for _, r := range t {
			c.Tokens[i] = append(c.Tokens[i], string(r))
		}
	}
	return c, nil
}
func Split(c Corpus) (Corpus, Corpus, Corpus) {
	n := len(c.Tokens)
	a, b := n*6/10, n*8/10
	cp := func(lo, hi int) Corpus {
		z := Corpus{Tokens: make([][]string, hi-lo)}
		for i := lo; i < hi; i++ {
			z.Tokens[i-lo] = append([]string(nil), c.Tokens[i]...)
		}
		return z
	}
	return cp(0, a), cp(a, b), cp(b, n)
}
func Vocabulary(c Corpus) []string {
	m := map[string]bool{}
	for _, t := range c.Tokens {
		for _, g := range t {
			m[g] = true
		}
	}
	v := make([]string, 0, len(m)+2)
	for g := range m {
		v = append(v, g)
	}
	sort.Slice(v, func(i, j int) bool { return bytes.Compare([]byte(v[i]), []byte(v[j])) < 0 })
	return append(v, "<UNK>", "<EOS>")
}
func Neumaier(xs []float64) float64 {
	s, c := 0., 0.
	for _, x := range xs {
		t := s + x
		if math.Abs(s) >= math.Abs(x) {
			c += (s - t) + x
		} else {
			c += (x - t) + s
		}
		s = t
	}
	return s + c
}
func Normalize(xs []float64) ([]float64, error) {
	for _, x := range xs {
		if x < 0 || math.IsNaN(x) || math.IsInf(x, 0) {
			return nil, fmt.Errorf("invalid mass")
		}
	}
	s := Neumaier(xs)
	if !(s > 0) {
		return nil, fmt.Errorf("zero mass")
	}
	p := make([]float64, len(xs))
	for i := range xs {
		p[i] = xs[i] / s
	}
	return p, nil
}
func QuantileType7(xs []float64, p float64) (float64, error) {
	if len(xs) == 0 || p < 0 || p > 1 {
		return 0, fmt.Errorf("quantile")
	}
	z := append([]float64(nil), xs...)
	sort.Float64s(z)
	h := float64(len(z)-1) * p
	i := int(math.Floor(h))
	if i == len(z)-1 {
		return z[i], nil
	}
	return z[i] + (h-float64(i))*(z[i+1]-z[i]), nil
}

type RNG struct {
	Root      [32]byte
	Namespace string
	Prefix    []uint64
	Draw      uint64
}

func NewRNG(ns string, prefix ...uint64) (RNG, error) {
	var r RNG
	b, e := hex.DecodeString(RootSeedHex)
	if e != nil {
		return r, e
	}
	copy(r.Root[:], b)
	r.Namespace = ns
	r.Prefix = append([]uint64(nil), prefix...)
	return r, nil
}
func (r *RNG) Digest() [32]byte {
	msg := append([]byte("G1V2-RNG\x00"), r.Root[:]...)
	ns := []byte(norm.NFC.String(r.Namespace))
	u4 := make([]byte, 4)
	binary.BigEndian.PutUint32(u4, uint32(len(ns)))
	msg = append(msg, u4...)
	msg = append(msg, ns...)
	all := append(append([]uint64(nil), r.Prefix...), r.Draw)
	binary.BigEndian.PutUint32(u4, uint32(len(all)))
	msg = append(msg, u4...)
	u8 := make([]byte, 8)
	for _, x := range all {
		binary.BigEndian.PutUint64(u8, x)
		msg = append(msg, u8...)
	}
	return sha256.Sum256(msg)
}
func (r *RNG) U53() float64 {
	d := r.Digest()
	r.Draw++
	return float64(binary.BigEndian.Uint64(d[:8])>>11) / 9007199254740992
}
func (r *RNG) Bounded(n uint64) (uint64, error) {
	if n == 0 {
		return 0, fmt.Errorf("zero bound")
	}
	limit := ^uint64(0) - (^uint64(0) % n)
	for {
		d := r.Digest()
		r.Draw++
		x := binary.BigEndian.Uint64(d[:8])
		if x < limit {
			return x % n, nil
		}
	}
}
func numeric(v any) (float64, error) {
	switch x := v.(type) {
	case string:
		return strconv.ParseFloat(x, 64)
	case float64:
		return x, nil
	default:
		return 0, fmt.Errorf("not numeric %T", v)
	}
}
func integer(v any) (int, error) {
	x, e := numeric(v)
	if e != nil || x != math.Trunc(x) {
		return 0, fmt.Errorf("not integer")
	}
	return int(x), nil
}
func glyphString(x []string) string { return strings.Join(x, "") }
