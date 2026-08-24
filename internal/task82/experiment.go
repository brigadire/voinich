// Package task82 executes the target-blind Task82 manifest and builds its
// aggregate portfolio. It deliberately imports only the Task81 mechanism
// package and ordinary control corpora.
package task82

import (
	"bufio"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"unicode"

	"zcore.dev/voinich/internal/mnemonicspace"
)

const (
	Version       = "V1.1"
	FreezeVersion = "V1.1"
)

type Options struct {
	Root       string
	ShardIndex int
	ShardCount int
	Resume     bool
	VerifyOnly bool
}

type Manifest struct {
	Version            string              `json:"version"`
	MasterSeed         uint64              `json:"master_seed"`
	EvaluationCorpora  []string            `json:"evaluation_corpora"`
	RecoveryConditions []string            `json:"recovery_conditions"`
	FrozenAblations    map[string][]string `json:"frozen_ablations"`
	Jobs               []ManifestJob       `json:"jobs"`
}

type ManifestJob struct {
	JobID             string `json:"job_id"`
	MechanismID       string `json:"mechanism_id"`
	ParameterSetID    string `json:"parameter_set_id"`
	InputCorpusID     string `json:"input_corpus_id"`
	Seed              uint64 `json:"seed"`
	Replicate         int    `json:"replicate"`
	RecoveryCondition string `json:"recovery_condition"`
}

type Corpus struct {
	ID             string                        `json:"id"`
	Path           string                        `json:"path"`
	SHA256         string                        `json:"sha256"`
	Bytes          int                           `json:"bytes"`
	UnicodeLetters int                           `json:"unicode_letters"`
	Preprocessing  string                        `json:"preprocessing"`
	Sequence       []mnemonicspace.Symbol        `json:"-"`
	Items          []mnemonicspace.RetrievedItem `json:"-"`
}

type Metrics struct {
	InputCount           int     `json:"input_count"`
	ObservableCount      int     `json:"observable_symbol_count"`
	ObservableUnits      int     `json:"observable_unit_count"`
	ExpansionRatio       float64 `json:"expansion_ratio"`
	DistinctSymbols      int     `json:"distinct_observable_symbols"`
	DistinctUnits        int     `json:"distinct_observable_units"`
	SymbolEntropy        float64 `json:"symbol_entropy_plugin_bits"`
	ConditionalEntropy   float64 `json:"conditional_symbol_entropy_plugin_bits"`
	RepetitionRate       float64 `json:"adjacent_repetition_rate"`
	RetainedProxy        float64 `json:"information_retained_count_proxy"`
	LostProxy            float64 `json:"information_lost_count_proxy"`
	ExactMatch           bool    `json:"exact_match"`
	RecoveryFraction     float64 `json:"recovery_fraction"`
	AmbiguityCardinality int     `json:"ambiguity_cardinality"`
	CandidateEntropy     float64 `json:"candidate_set_entropy_bits"`
	HMGivenXProxy        float64 `json:"h_m_given_x_ambiguity_proxy_bits"`
}

type Artifact struct {
	Schema           string                           `json:"schema"`
	Implementation   string                           `json:"implementation_version"`
	FreezeVersion    string                           `json:"freeze_version"`
	Job              ManifestJob                      `json:"job"`
	MechanismVersion string                           `json:"mechanism_version"`
	Family           mnemonicspace.Family             `json:"family"`
	HistoricalStatus mnemonicspace.HistoricalStatus   `json:"historical_status"`
	AblationStatus   string                           `json:"ablation_control_status"`
	Input            InputSummary                     `json:"input_M"`
	Carriers         CarrierAccounting                `json:"carrier_accounting"`
	Observable       mnemonicspace.ObservableDocument `json:"observable_document_X"`
	InformationTrace []mnemonicspace.InformationEvent `json:"information_trace"`
	Recovery         mnemonicspace.RecoveryResult     `json:"retrieval_result"`
	Metrics          Metrics                          `json:"metrics"`
	InputSHA256      string                           `json:"input_checksum"`
	OutputSHA256     string                           `json:"observable_checksum"`
	StateSHA256      string                           `json:"external_state_checksum"`
	Warnings         []string                         `json:"warnings"`
	RuntimeNS        int64                            `json:"runtime_ns"`
	SoftwareVersion  string                           `json:"software_version"`
}

type InputSummary struct {
	CorpusID string   `json:"corpus_id"`
	ItemIDs  []string `json:"item_ids"`
	Symbols  []string `json:"symbols"`
	Count    int      `json:"count"`
}

type CarrierAccounting struct {
	E string `json:"E_external_state"`
	G string `json:"G_geometry"`
	H string `json:"H_history"`
	K string `json:"K_convention"`
	I string `json:"I_internal_memory"`
	C string `json:"C_context"`
}

var bindings = map[string]string{
	"MNEMONIC_MECHANISM_REGISTRY.json": "2b1d53038a790863d356892a9a147acf8e551fd6e721632111c11276a1d9016e",
	"MNEMONIC_PARAMETER_REGISTRY.tsv":  "fb88dcf9a5ca9c1dc52dd8c3fd45be2d0cc2e3da3245127fc58d53947dc3fcdb",
	"MNEMONIC_RECOVERY_CONTRACT.md":    "d0e6fe3756dd5e65c8c108db794fbb6404a9a3aee739d4fe76fdf77b2a333fcb",
	"OBSERVABLE_DOCUMENT_CONTRACT.md":  "f6b57179e4169ff58b741ded52eb06cb306d0a9f91ba8c8d51fd645389a15f51",
	"TASK82_BLIND_MANIFEST.json":       "2ce3d0d9508518ceda0203a9eebce234e667fa5f0cd4e12a190fac7fa2861883",
}

func Execute(o Options) error {
	if o.Root == "" {
		o.Root = "."
	}
	if o.ShardCount <= 0 {
		o.ShardCount = 1
	}
	if o.ShardIndex < 0 || o.ShardIndex >= o.ShardCount {
		return fmt.Errorf("invalid shard %d/%d", o.ShardIndex, o.ShardCount)
	}
	authority, manifest, specs, corpora, err := verifyFreeze(o.Root)
	if err != nil {
		return fmt.Errorf("FREEZE_MISMATCH: %w", err)
	}
	out := filepath.Join(o.Root, "research", "phase2", "task82")
	raw := filepath.Join(out, "raw")
	if err := os.MkdirAll(raw, 0o755); err != nil {
		return err
	}
	if o.VerifyOnly {
		return verifyRaw(manifest, raw, o.ShardIndex, o.ShardCount)
	}
	for n, job := range manifest.Jobs {
		if n%o.ShardCount != o.ShardIndex {
			continue
		}
		path := filepath.Join(raw, job.JobID+".json")
		if o.Resume {
			if data, e := os.ReadFile(path); e == nil {
				var a Artifact
				if json.Unmarshal(data, &a) == nil && a.Job == job && a.FreezeVersion == FreezeVersion {
					continue
				}
			}
		}
		spec, ok := specs[job.MechanismID]
		if !ok {
			return fmt.Errorf("manifest mechanism %s is not frozen", job.MechanismID)
		}
		param, ok := spec.ParameterSet(job.ParameterSetID)
		if !ok {
			return fmt.Errorf("manifest parameter %s is not frozen for %s", job.ParameterSetID, spec.ID)
		}
		c, ok := corpora[job.InputCorpusID]
		if !ok {
			return fmt.Errorf("manifest corpus %s is unavailable", job.InputCorpusID)
		}
		a, err := runJob(authority, manifest, job, spec, param, c)
		if err != nil {
			return fmt.Errorf("job %s: %w", job.JobID, err)
		}
		data, err := json.MarshalIndent(a, "", "  ")
		if err != nil {
			return err
		}
		data = append(data, '\n')
		if err := os.WriteFile(path, data, 0o644); err != nil {
			return err
		}
	}
	if o.ShardCount == 1 {
		return Aggregate(o.Root, manifest, specs, corpora)
	}
	return nil
}

func verifyFreeze(root string) (mnemonicspace.Authority, Manifest, map[string]mnemonicspace.MechanismSpec, map[string]Corpus, error) {
	base := filepath.Join(root, "research", "phase2", "mechanism-space")
	for name, want := range bindings {
		got, err := fileHash(filepath.Join(base, name))
		if err != nil {
			return mnemonicspace.Authority{}, Manifest{}, nil, nil, err
		}
		if got != want {
			return mnemonicspace.Authority{}, Manifest{}, nil, nil, fmt.Errorf("%s checksum %s, want %s", name, got, want)
		}
	}
	var frozen struct {
		Version       string `json:"version"`
		Task80Algebra string `json:"task80_algebra_checksum"`
		Task80Models  string `json:"task80_model_freeze_checksum"`
	}
	if err := readJSON(filepath.Join(base, "MNEMONIC_MECHANISM_SPACE_FROZEN.json"), &frozen); err != nil {
		return mnemonicspace.Authority{}, Manifest{}, nil, nil, err
	}
	if frozen.Version != FreezeVersion {
		return mnemonicspace.Authority{}, Manifest{}, nil, nil, fmt.Errorf("freeze version %s", frozen.Version)
	}
	authority, err := mnemonicspace.LoadTask80Authority(filepath.Join(root, "research", "phase2", "fontana", "task80"))
	if err != nil {
		return mnemonicspace.Authority{}, Manifest{}, nil, nil, err
	}
	if frozen.Task80Algebra != authority.AlgebraSHA256 || frozen.Task80Models != authority.FrozenSHA256 {
		return mnemonicspace.Authority{}, Manifest{}, nil, nil, errors.New("Task80 binding mismatch")
	}
	var m Manifest
	if err := readJSON(filepath.Join(base, "TASK82_BLIND_MANIFEST.json"), &m); err != nil {
		return mnemonicspace.Authority{}, Manifest{}, nil, nil, err
	}
	if m.Version != FreezeVersion || len(m.Jobs) != 672 {
		return mnemonicspace.Authority{}, Manifest{}, nil, nil, fmt.Errorf("manifest version/count %s/%d", m.Version, len(m.Jobs))
	}
	reg := mnemonicspace.FrozenRegistry()
	if err := mnemonicspace.ValidateRegistry(authority, reg); err != nil {
		return mnemonicspace.Authority{}, Manifest{}, nil, nil, err
	}
	specs := map[string]mnemonicspace.MechanismSpec{}
	for _, s := range reg {
		specs[s.ID] = s
	}
	seen := map[string]bool{}
	for _, j := range m.Jobs {
		if seen[j.JobID] {
			return mnemonicspace.Authority{}, Manifest{}, nil, nil, fmt.Errorf("duplicate job %s", j.JobID)
		}
		seen[j.JobID] = true
		s, ok := specs[j.MechanismID]
		if !ok {
			return mnemonicspace.Authority{}, Manifest{}, nil, nil, fmt.Errorf("unknown mechanism %s", j.MechanismID)
		}
		mj := mnemonicspace.Job{MechanismID: j.MechanismID, ParameterSetID: j.ParameterSetID, InputID: j.InputCorpusID, RecoveryCondition: mnemonicspace.RecoveryCondition(j.RecoveryCondition), Replicate: j.Replicate, MasterSeed: m.MasterSeed}
		if mj.DerivedSeed() != j.Seed || mj.ID(s, authority) != j.JobID {
			return mnemonicspace.Authority{}, Manifest{}, nil, nil, fmt.Errorf("job identity mismatch %s", j.JobID)
		}
	}
	corpora, err := loadCorpora(root)
	if err != nil {
		return mnemonicspace.Authority{}, Manifest{}, nil, nil, err
	}
	return authority, m, specs, corpora, nil
}

func loadCorpora(root string) (map[string]Corpus, error) {
	paths := map[string]string{"Doyle": "data_test/pg2097-2.txt", "Longfellow": "data_test/pg30795-mod.txt", "Astafiev": "data_test/astafiev-1000-culinar-receipts-utf8.txt"}
	out := map[string]Corpus{}
	for id, rel := range paths {
		b, err := os.ReadFile(filepath.Join(root, rel))
		if err != nil {
			return nil, err
		}
		c := Corpus{ID: id, Path: rel, SHA256: sum(b), Bytes: len(b), Preprocessing: "first Unicode letters -> upper(codepoint mod Latin23); first Unicode whitespace-delimited words -> SHA-256 item IDs"}
		const alphabet = "ABCDEFGHIKLMNOPQRSTVXYZ"
		words := bufio.NewScanner(strings.NewReader(string(b)))
		words.Split(bufio.ScanWords)
		words.Buffer(make([]byte, 1024), 1024*1024)
		for words.Scan() {
			if len(c.Items) >= 4 {
				break
			}
			token := strings.TrimFunc(words.Text(), func(r rune) bool { return !unicode.IsLetter(r) })
			if token == "" {
				continue
			}
			h := sha256.Sum256([]byte(strings.ToLower(token)))
			c.Items = append(c.Items, mnemonicspace.RetrievedItem("I"+hex.EncodeToString(h[:8])))
		}
		for _, r := range string(b) {
			if unicode.IsLetter(r) {
				c.UnicodeLetters++
				if len(c.Sequence) < 8 {
					c.Sequence = append(c.Sequence, mnemonicspace.Symbol(string(alphabet[int(r)%len(alphabet)])))
				}
			}
		}
		if len(c.Sequence) != 8 || len(c.Items) != 4 {
			return nil, fmt.Errorf("corpus %s cannot produce bounded adapter", id)
		}
		out[id] = c
	}
	return out, nil
}

func runJob(authority mnemonicspace.Authority, manifest Manifest, j ManifestJob, spec mnemonicspace.MechanismSpec, param mnemonicspace.ParameterSet, c Corpus) (Artifact, error) {
	input := mnemonicspace.InputModel{ID: c.ID, Sequence: append([]mnemonicspace.Symbol(nil), c.Sequence...), IndexedCues: map[mnemonicspace.Index]mnemonicspace.Cue{}, PositionedCues: map[mnemonicspace.Position]mnemonicspace.Cue{}, TimedCues: map[int]mnemonicspace.Cue{}}
	items := append([]mnemonicspace.RetrievedItem(nil), c.Items...)
	for i := 0; i < 4; i++ {
		cue := mnemonicspace.Cue(fmt.Sprintf("C%d", i))
		input.IndexedCues[mnemonicspace.Index(i)] = cue
		input.PositionedCues[mnemonicspace.Position(i)] = cue
		input.TimedCues[(i+1)%4] = cue
	}
	idx := mnemonicspace.Index(0)
	if spec.ID == "m_restricted_rotation_index" {
		// The frozen +1 rotation places C0 at visible index 1.
		idx = 1
	}
	pos := mnemonicspace.Position(0)
	request := mnemonicspace.RetrievalRequest{Condition: mnemonicspace.RecoveryCondition(j.RecoveryCondition), TargetIndex: &idx, TargetPosition: &pos}
	candidates := []mnemonicspace.RetrievedItem{items[0], mnemonicspace.RetrievedItem("ALT_" + string(items[1]))}
	env := mnemonicspace.RecoveryEnvironment{Geometry: &mnemonicspace.GeometryKnowledge{Path: []mnemonicspace.Position{0, 1, 2, 3, 4, 5, 6, 7}}, History: &mnemonicspace.HistoryKnowledge{Steps: 1}, Convention: &mnemonicspace.ConventionKnowledge{MessageLength: len(input.Sequence), CueMeanings: map[mnemonicspace.Cue][]mnemonicspace.RetrievedItem{"C0": candidates}}, InternalMemory: &mnemonicspace.InternalMemoryState{Associations: map[mnemonicspace.Cue][]mnemonicspace.RetrievedItem{"C0": candidates}}, Context: &mnemonicspace.ContextKnowledge{Allowed: []mnemonicspace.RetrievedItem{items[0]}}}
	// Full knowledge contains exactly the carriers declared by the frozen
	// mechanism, never helpful undeclared state.
	if !hasCarrier(spec.Carriers.Retrieve, mnemonicspace.CarrierGeometry) {
		env.Geometry = nil
	}
	if !hasCarrier(spec.Carriers.Retrieve, mnemonicspace.CarrierHistory) {
		env.History = nil
	}
	if !hasCarrier(spec.Carriers.Retrieve, mnemonicspace.CarrierConvention) {
		env.Convention = nil
	}
	if !hasCarrier(spec.Carriers.Retrieve, mnemonicspace.CarrierInternal) {
		env.InternalMemory = nil
	}
	if !hasCarrier(spec.Carriers.Retrieve, mnemonicspace.CarrierContext) {
		env.Context = nil
	}
	switch spec.ID {
	case "negative_randomized_convention":
		env.Convention.MessageLength = len(input.Sequence) - 1
	case "negative_randomized_path":
		env.Geometry = nil
	case "negative_randomized_cue_association":
		env.InternalMemory.Associations = map[mnemonicspace.Cue][]mnemonicspace.RetrievedItem{"WRONG": {items[0]}}
	case "negative_randomized_index_mapping":
		env.Convention.CueMeanings = map[mnemonicspace.Cue][]mnemonicspace.RetrievedItem{"WRONG": {items[0]}}
	}
	exec, err := (mnemonicspace.Runner{}).Run(spec, param, input, int64(j.Seed), request, env)
	if err != nil {
		return Artifact{}, err
	}
	if err := validateDocument(spec, exec.Prepared.Document); err != nil {
		return Artifact{}, fmt.Errorf("LEAKAGE_FAILURE: %w", err)
	}
	stateHash, err := hashJSON(exec.Prepared.State)
	if err != nil {
		return Artifact{}, err
	}
	seq := make([]string, len(c.Sequence))
	for i, s := range c.Sequence {
		seq[i] = string(s)
	}
	itemStrings := make([]string, len(items))
	for i, v := range items {
		itemStrings[i] = string(v)
	}
	inputCount := len(seq)
	target := strings.Join(seq, "")
	if spec.SurfaceRole == mnemonicspace.SurfaceOpaqueCue {
		inputCount = len(items)
		target = string(items[0])
	}
	met := measure(exec.Prepared.Document.Symbols, inputCount, exec.Recovery, target)
	warnings := []string{}
	if strings.HasPrefix(spec.ID, "f01_") || spec.ID == "negative_randomized_convention" {
		warnings = append(warnings, "FROZEN_CONDITION_SPECIFIC_SEED: R0-R6 contrast is not a paired generated state")
	}
	a := Artifact{Schema: "TASK82_RAW_JOB_V1", Implementation: Version, FreezeVersion: FreezeVersion, Job: j, MechanismVersion: spec.Version, Family: spec.Family, HistoricalStatus: spec.Status, AblationStatus: ablationStatus(manifest, spec.ID), Input: InputSummary{CorpusID: c.ID, ItemIDs: itemStrings, Symbols: seq, Count: inputCount}, Carriers: CarrierAccounting{E: "present; checksum only", G: presence(env.Geometry), H: presence(env.History), K: presence(env.Convention), I: presence(env.InternalMemory), C: presence(env.Context)}, Observable: exec.Prepared.Document, InformationTrace: exec.Prepared.Trace, Recovery: exec.Recovery, Metrics: met, InputSHA256: c.SHA256, OutputSHA256: exec.Prepared.Document.Checksum(), StateSHA256: stateHash, Warnings: warnings, RuntimeNS: 0, SoftwareVersion: "task82-blind/" + Version}
	_ = authority
	return a, nil
}

func measure(symbols []string, inputCount int, r mnemonicspace.RecoveryResult, target string) Metrics {
	m := Metrics{InputCount: inputCount, ObservableCount: len(symbols), ObservableUnits: len(symbols)}
	if inputCount > 0 {
		m.ExpansionRatio = float64(len(symbols)) / float64(inputCount)
		m.RetainedProxy = math.Min(float64(inputCount), float64(len(symbols)))
		m.LostProxy = float64(inputCount) - m.RetainedProxy
	}
	counts := map[string]int{}
	pairs := map[string]int{}
	repeats := 0
	for i, s := range symbols {
		counts[s]++
		if i > 0 {
			pairs[symbols[i-1]+"\x00"+s]++
			if s == symbols[i-1] {
				repeats++
			}
		}
	}
	m.DistinctSymbols = len(counts)
	m.DistinctUnits = len(counts)
	m.SymbolEntropy = entropy(counts, len(symbols))
	m.ConditionalEntropy = conditionalEntropy(symbols)
	if len(symbols) > 1 {
		m.RepetitionRate = float64(repeats) / float64(len(symbols)-1)
	}
	m.ExactMatch = r.Class == mnemonicspace.ResultExact && string(r.Value) == target
	switch r.Class {
	case mnemonicspace.ResultExact:
		if m.ExactMatch {
			m.RecoveryFraction = 1
		}
	case mnemonicspace.ResultPartial:
		m.RecoveryFraction = 0.5
	case mnemonicspace.ResultAmbiguitySet:
		m.AmbiguityCardinality = len(r.Candidates)
		if len(r.Candidates) > 0 {
			m.RecoveryFraction = 1 / float64(len(r.Candidates))
			m.CandidateEntropy = math.Log2(float64(len(r.Candidates)))
			m.HMGivenXProxy = m.CandidateEntropy
		}
	}
	return m
}

func validateDocument(spec mnemonicspace.MechanismSpec, d mnemonicspace.ObservableDocument) error {
	if !spec.Serialization.ExcludesHiddenState {
		return errors.New("registry permits hidden state")
	}
	if d.TokenBoundary != spec.Serialization.TokenBoundary || d.LineBoundary != spec.Serialization.LineBoundary || d.UnitBoundary != spec.Serialization.UnitBoundary {
		return errors.New("boundary provenance mismatch")
	}
	for k := range d.Metadata {
		s := strings.ToLower(k)
		for _, bad := range []string{"plaintext", "offset", "association", "internal_memory", "message_length", "path", "history"} {
			if strings.Contains(s, bad) {
				return fmt.Errorf("hidden metadata %q", k)
			}
		}
	}
	return nil
}

func verifyRaw(m Manifest, raw string, si, sc int) error {
	for n, j := range m.Jobs {
		if n%sc != si {
			continue
		}
		b, err := os.ReadFile(filepath.Join(raw, j.JobID+".json"))
		if err != nil {
			return err
		}
		var a Artifact
		if err = json.Unmarshal(b, &a); err != nil {
			return err
		}
		if a.Job != j || a.OutputSHA256 != a.Observable.Checksum() {
			return fmt.Errorf("artifact mismatch %s", j.JobID)
		}
	}
	return nil
}

func entropy(counts map[string]int, n int) float64 {
	if n == 0 {
		return 0
	}
	h := 0.0
	keys := make([]string, 0, len(counts))
	for k := range counts {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		v := counts[k]
		p := float64(v) / float64(n)
		h -= p * math.Log2(p)
	}
	return h
}
func conditionalEntropy(s []string) float64 {
	if len(s) < 2 {
		return 0
	}
	prefix := map[string]int{}
	pair := map[string]int{}
	for i := 1; i < len(s); i++ {
		prefix[s[i-1]]++
		pair[s[i-1]+"\x00"+s[i]]++
	}
	h := 0.0
	n := float64(len(s) - 1)
	keys := make([]string, 0, len(pair))
	for k := range pair {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		v := pair[k]
		p := float64(v) / n
		pre := strings.SplitN(k, "\x00", 2)[0]
		h -= p * math.Log2(float64(v)/float64(prefix[pre]))
	}
	return h
}
func presence(v any) string {
	if v == nil {
		return "absent"
	}
	switch x := v.(type) {
	case *mnemonicspace.GeometryKnowledge:
		if x == nil {
			return "absent"
		}
	case *mnemonicspace.HistoryKnowledge:
		if x == nil {
			return "absent"
		}
	case *mnemonicspace.ConventionKnowledge:
		if x == nil {
			return "absent"
		}
	case *mnemonicspace.InternalMemoryState:
		if x == nil {
			return "absent"
		}
	case *mnemonicspace.ContextKnowledge:
		if x == nil {
			return "absent"
		}
	}
	return "present; values withheld"
}
func ablationStatus(m Manifest, id string) string {
	var matches []string
	for parent, vals := range m.FrozenAblations {
		for _, v := range vals {
			p := strings.SplitN(v, ":", 2)
			if len(p) == 2 && p[1] == id {
				matches = append(matches, "ABLATION_OF:"+parent+":"+p[0])
			}
		}
	}
	if len(matches) > 0 {
		sort.Strings(matches)
		return strings.Join(matches, ",")
	}
	if strings.HasPrefix(id, "negative_") {
		return "NEGATIVE_CONTROL"
	}
	if strings.HasPrefix(id, "synthetic_") {
		return "GENERIC_CONTROL"
	}
	return "FULL_MECHANISM"
}
func readJSON(path string, v any) error {
	b, e := os.ReadFile(path)
	if e != nil {
		return e
	}
	return json.Unmarshal(b, v)
}
func hashJSON(v any) (string, error) {
	b, e := json.Marshal(v)
	if e != nil {
		return "", e
	}
	return sum(b), nil
}
func fileHash(path string) (string, error) {
	b, e := os.ReadFile(path)
	if e != nil {
		return "", e
	}
	return sum(b), nil
}
func sum(b []byte) string { h := sha256.Sum256(b); return hex.EncodeToString(h[:]) }
func f6(v float64) string { return strconv.FormatFloat(v, 'f', 6, 64) }
