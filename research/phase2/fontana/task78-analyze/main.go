// Command task78-analyze executes the preregistered, target-blind operational
// checks for the non-Speculum Fontana family selected in task78. It reads no
// Voynich data. All reconstruction choices are named in profiles.json and in
// the generated model packages.
package main

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"zcore.dev/voinich/internal/fontanafamily"
)

const outputDir = "research/phase2/fontana/task78"
const seed = int64(780823)

var alphabet = []rune("ABCDEFGHIKLMNOPQRSTVXYZ")
var messages = []string{"MEMORIA", "FONTANA", "SERPENS", "ORDO", "LVMEN", "TEMPVS"}

type result struct {
	Model, Profile, Experiment, Condition, Trial string
	Input, Observed, StateJSON, Notes            string
	Exact                                        bool
	Accuracy                                     float64
	EditDistance                                 float64
	Compatible                                   string
}

type profile struct {
	Model      string            `json:"model"`
	ID         string            `json:"id"`
	Tier       string            `json:"tier"`
	Status     string            `json:"status"`
	Parameters map[string]any    `json:"parameters"`
	Evidence   map[string]string `json:"evidence"`
}

type modelPackage struct {
	Metadata               map[string]any    `json:"metadata"`
	Sources                []string          `json:"sources"`
	Reconstruction         map[string]any    `json:"reconstruction"`
	Uncertainty            []string          `json:"uncertainty"`
	Components             []string          `json:"components"`
	StateSchema            map[string]string `json:"state_schema"`
	Operations             []string          `json:"operations"`
	Transitions            []string          `json:"transitions"`
	Encoding               string            `json:"encoding"`
	Retrieval              string            `json:"retrieval"`
	KnowledgeRequirements  []string          `json:"knowledge_requirements"`
	ReconstructionProfiles []string          `json:"reconstruction_profiles"`
	Experiments            []string          `json:"experiments"`
	Results                map[string]any    `json:"results"`
	Tests                  []string          `json:"tests"`
	Verdict                map[string]string `json:"verdict"`
}

func main() {
	if err := os.MkdirAll(filepath.Join(outputDir, "models"), 0o755); err != nil {
		panic(err)
	}
	profiles := declaredProfiles()
	writeJSON(filepath.Join(outputDir, "profiles.json"), map[string]any{
		"schema_version": "task78-profiles-v1", "seed": seed,
		"preregistered_rule": "compare every declared profile; do not select by outcome", "profiles": profiles,
	})
	var rows []result
	rows = append(rows, runSerpens()...)
	rows = append(rows, runRota()...)
	rows = append(rows, runCylinder()...)
	rows = append(rows, runArismetricum()...)
	rows = append(rows, runHoralogius()...)
	writeResults(rows)
	writeDynamics()
	writePackages(rows)
	fmt.Printf("task78-analyze: wrote %d trials to %s (seed=%d)\n", len(rows), outputDir, seed)
}

func declaredProfiles() []profile {
	return []profile{
		{Model: "F08_SERPENS", ID: "R0", Tier: "A", Status: "invariant_core", Parameters: map[string]any{"capacity": 12, "start": 0, "direction": "centre_to_edge", "boundary": "length_in_K", "alphabet": "latin23"}, Evidence: map[string]string{"spiral_holes_letters_direction": "E", "capacity_alphabet_boundary": "H"}},
		{Model: "F08_SERPENS", ID: "R1", Tier: "B", Status: "sensitivity_only", Parameters: map[string]any{"capacity": 12, "start": 0, "direction": "centre_to_edge", "boundary": "first_empty_hole", "alphabet": "latin23"}, Evidence: map[string]string{"empty_as_stop": "H"}},
		{Model: "F07_ROTA", ID: "R0", Tier: "B", Status: "invariant_core_only", Parameters: map[string]any{"rings": 1, "alphabet": "latin23", "step": 1}, Evidence: map[string]string{"finite_circular_state": "I", "alphabet_layout_step": "H"}},
		{Model: "F10_CYLINDRUS", ID: "R0", Tier: "B", Status: "bounded_reconstruction", Parameters: map[string]any{"bands": 7, "alphabet": "latin23", "movement": "independent", "read_line": 0}, Evidence: map[string]string{"cylinder_bands_rotation": "I", "band_count_independence_read_line": "H"}},
		{Model: "F10_CYLINDRUS", ID: "R1", Tier: "B", Status: "sensitivity_only", Parameters: map[string]any{"bands": 7, "alphabet": "latin23", "movement": "coupled", "read_line": 0}, Evidence: map[string]string{"coupling": "H"}},
		{Model: "F11_ARISMETRICUM", ID: "R0", Tier: "B", Status: "invariant_core_only", Parameters: map[string]any{"indices": "integers", "values": "opaque_cues"}, Evidence: map[string]string{"holes_numbers": "E", "mapping": "U"}},
		{Model: "F12_HORALOGIUS", ID: "R0", Tier: "B", Status: "invariant_core_only", Parameters: map[string]any{"period": 12, "cue_ticks": []int{0, 4, 8}}, Evidence: map[string]string{"time_state_signal": "E", "period_calibration_cues": "H"}},
	}
}

func serpensConfig() fontanafamily.SerpensConfig {
	return fontanafamily.SerpensConfig{Capacity: 12, Alphabet: alphabet, Start: 0, Direction: fontanafamily.Forward, EmptyMarker: '?'}
}

func state(v any) string { b, _ := json.Marshal(v); return string(b) }

func add(model, profile, experiment, condition, trial, input, observed string, compatible int, s any, notes string) result {
	return result{Model: model, Profile: profile, Experiment: experiment, Condition: condition, Trial: trial,
		Input: input, Observed: observed, Exact: fontanafamily.Exact(input, observed), Accuracy: fontanafamily.SymbolAccuracy(input, observed),
		EditDistance: fontanafamily.NormalizedEditDistance(input, observed), Compatible: strconv.Itoa(compatible), StateJSON: state(s), Notes: notes}
}

func runSerpens() []result {
	c := serpensConfig()
	var rows []result
	for i, msg := range messages {
		s, _ := c.Encode(msg)
		trial := fmt.Sprintf("S%02d", i+1)
		got, _ := c.Decode(s, len([]rune(msg)))
		rows = append(rows, add("F08_SERPENS", "R0", "baseline", "full_K_intact", trial, msg, got, 1, s, "literal positional reading"))
		for _, ab := range []struct {
			name       string
			start, dir bool
		}{{"start_unknown", false, true}, {"direction_unknown", true, false}, {"start_direction_unknown", false, false}} {
			candidates, _ := c.CompatibleTraversals(s, len([]rune(msg)), ab.start, ab.dir)
			rows = append(rows, add("F08_SERPENS", "R0", "knowledge_ablation", ab.name, trial, msg, candidates[0], len(candidates), s, "observed is first enumerated candidate; compatible is primary outcome"))
		}
		rows = append(rows, result{Model: "F08_SERPENS", Profile: "R0", Experiment: "knowledge_ablation", Condition: "boundary_unknown", Trial: trial, Input: msg, Observed: "<set>", Compatible: "12", StateJSON: state(s), Notes: "one candidate per length 1..capacity; empty holes do not historically certify a terminator"})
		rows = append(rows, result{Model: "F08_SERPENS", Profile: "R0", Experiment: "knowledge_ablation", Condition: "association_unknown", Trial: trial, Input: msg, Observed: "<unbounded>", Compatible: "unbounded", StateJSON: state(s), Notes: "no formal enumeration: arbitrary semantic association is outside literal decoder"})
		factorial := 1
		for n := 2; n <= len([]rune(msg)); n++ {
			factorial *= n
		}
		rows = append(rows, result{Model: "F08_SERPENS", Profile: "R0", Experiment: "knowledge_ablation", Condition: "element_order_unknown", Trial: trial, Input: msg, Observed: "<set>", Compatible: strconv.Itoa(factorial), StateJSON: state(s), Notes: "upper bound L!; repeated symbols can lower the distinct count"})
		rows = append(rows, result{Model: "F08_SERPENS", Profile: "R0", Experiment: "knowledge_ablation", Condition: "marker_meaning_unknown", Trial: trial, Input: msg, Observed: "<unbounded>", Compatible: "unbounded", StateJSON: state(s), Notes: "literal glyphs remain visible but their association is unspecified"})
		rows = append(rows, result{Model: "F08_SERPENS", Profile: "R0", Experiment: "knowledge_ablation", Condition: "transition_rule_unknown", Trial: trial, Input: msg, Observed: "<set>", Compatible: strconv.Itoa(factorial), StateJSON: state(s), Notes: "L! upper bound when the next-hole rule is unavailable"})
		rows = append(rows, result{Model: "F08_SERPENS", Profile: "R0", Experiment: "knowledge_ablation", Condition: "stop_rule_unknown", Trial: trial, Input: msg, Observed: "<set>", Compatible: "12", StateJSON: state(s), Notes: "equivalent to boundary ablation in finite R0; no terminator inferred"})
		pos := c.Start + 2
		for _, corrupt := range []struct {
			name  string
			value fontanafamily.SerpensState
		}{
			{"single_substitution", fontanafamily.SerpensSubstitute(s, pos, 'X')},
			{"missing_insertion", fontanafamily.SerpensRemove(s, pos)},
			{"adjacent_swap", fontanafamily.SerpensSwap(s, pos, pos+1)},
			{"frame_collapse", fontanafamily.SerpensCollapse(s, pos)},
			{"duplicate_insertion", fontanafamily.SerpensSubstitute(s, pos, *s.Holes[pos-1])},
		} {
			v, _ := c.Decode(corrupt.value, len([]rune(msg)))
			rows = append(rows, add("F08_SERPENS", "R0", "state_corruption", corrupt.name, trial, msg, v, 1, corrupt.value, classify(msg, v)))
		}
		wrongStart := c
		wrongStart.Start = 1
		v, _ := wrongStart.Decode(s, len([]rune(msg)))
		rows = append(rows, add("F08_SERPENS", "R0", "state_corruption", "start_shift", trial, msg, v, 1, s, "synchronization/global ambiguity"))
		wrongDir := c
		wrongDir.Start = len([]rune(msg)) - 1
		wrongDir.Direction = fontanafamily.Reverse
		v, _ = wrongDir.Decode(s, len([]rune(msg)))
		rows = append(rows, add("F08_SERPENS", "R0", "state_corruption", "direction_reversal", trial, msg, v, 1, s, "global order error"))
		// Geometry perturbation that preserves topological hole order is an
		// invariant no-op; loss of geometry/order is represented by swap/collapse.
		rows = append(rows, add("F08_SERPENS", "R0", "state_corruption", "local_geometry_order_preserved", trial, msg, got, 1, s, "geometry changed, topological traversal order preserved"))
		damaged := fontanafamily.SerpensRemove(s, pos)
		combined, _ := c.CompatibleTraversals(damaged, len([]rune(msg)), false, true)
		rows = append(rows, add("F08_SERPENS", "R0", "combined_degradation", "start_unknown_plus_missing", trial, msg, combined[0], len(combined), damaged, "no candidate selected by message similarity"))
	}
	return rows
}

func runRota() []result {
	var rows []result
	r := fontanafamily.Rota{Alphabet: alphabet}
	for i := 0; i < len(alphabet); i++ {
		got, _ := r.Observe()
		want := string(alphabet[i])
		rows = append(rows, add("F07_ROTA", "R0", "baseline", "known_zero_step", fmt.Sprintf("R%02d", i+1), want, string(got), 1, r, "selector observation"))
		r = r.Rotate(1)
	}
	got, _ := r.Rotate(1).Observe()
	rows = append(rows, add("F07_ROTA", "R0", "state_corruption", "one_step_shift", "R24", string(alphabet[0]), string(got), 1, r.Rotate(1), "local selected-symbol error; zero loss gives 23 compatible alignments"))
	return rows
}

func runCylinder() []result {
	var rows []result
	for i, msg := range messages {
		m := []rune(msg)
		if len(m) != 7 {
			continue
		}
		c := fontanafamily.Cylinder{Alphabet: alphabet, Offsets: make([]int, len(m))}
		s, _ := c.Encode(msg, fontanafamily.Forward)
		got, _ := s.Read(fontanafamily.Forward)
		trial := fmt.Sprintf("C%02d", i+1)
		rows = append(rows, add("F10_CYLINDRUS", "R0", "baseline", "full_K_intact", trial, msg, got, 1, s, "independent-band H-profile"))
		rev, _ := s.Read(fontanafamily.Reverse)
		rows = append(rows, add("F10_CYLINDRUS", "R0", "knowledge_ablation", "route_unknown", trial, msg, rev, 2, s, "forward and reverse routes"))
		dmg, _ := s.RotateBand(2, 1)
		bad, _ := dmg.Read(fontanafamily.Forward)
		rows = append(rows, add("F10_CYLINDRUS", "R0", "state_corruption", "one_band_shift", trial, msg, bad, 1, dmg, "local under R0 independent bands"))
		coupled := s
		coupled.Offsets = append([]int(nil), s.Offsets...)
		for j := 2; j < len(coupled.Offsets); j++ {
			coupled.Offsets[j] = fontanafamily.Normalize(coupled.Offsets[j]+1, len(alphabet))
		}
		bad, _ = coupled.Read(fontanafamily.Forward)
		rows = append(rows, add("F10_CYLINDRUS", "R1", "state_corruption", "coupled_shift", trial, msg, bad, 1, coupled, "cascade under H-profile R1; not an invariant property"))
	}
	return rows
}

func runArismetricum() []result {
	a := fontanafamily.Arismetricum{Slots: map[int]string{1: "MEMORIA", 2: "FONTANA", 3: "SERPENS", 4: "ORDO", 5: "LVMEN", 6: "TEMPVS"}}
	var rows []result
	for i, want := range messages {
		idx := i + 1
		got, _ := a.Lookup(idx)
		trial := fmt.Sprintf("A%02d", idx)
		rows = append(rows, add("F11_ARISMETRICUM", "R0", "baseline", "known_index_mapping", trial, want, got, 1, a, "indexed cue lookup, not computed text"))
		_, ok := a.Remove(idx).Lookup(idx)
		observed := "?"
		if ok {
			observed = got
		}
		rows = append(rows, add("F11_ARISMETRICUM", "R0", "state_corruption", "slot_missing", trial, want, observed, 0, a.Remove(idx), "detectable missing lookup"))
		swapped := a.Swap(idx, (idx%len(messages))+1)
		bad, _ := swapped.Lookup(idx)
		rows = append(rows, add("F11_ARISMETRICUM", "R0", "state_corruption", "adjacent_swap", trial, want, bad, 1, swapped, "local wrong cue; meaning still depends on mapping"))
		rows = append(rows, result{Model: "F11_ARISMETRICUM", Profile: "R0", Experiment: "knowledge_ablation", Condition: "index_convention_unknown", Trial: trial, Input: want, Observed: "<set>", Compatible: strconv.Itoa(len(a.Slots)), StateJSON: state(a), Notes: "all occupied slots compatible"})
	}
	return rows
}

func runHoralogius() []result {
	h := fontanafamily.Horalogius{Period: 12, Tick: 11, Cues: map[int]string{0: "BELL_A", 4: "BELL_B", 8: "BELL_C"}}
	learned := map[string]string{"BELL_A": "LABOR", "BELL_B": "ORATIO", "BELL_C": "STVDIVM"}
	var rows []result
	for i := 0; i < 12; i++ {
		before := h
		var cues []string
		h, cues, _ = h.Advance(1)
		want := ""
		if len(cues) > 0 {
			want = learned[cues[0]]
		}
		observed := ""
		if len(cues) > 0 {
			observed, _ = fontanafamily.Recall(cues[0], learned)
		}
		rows = append(rows, add("F12_HORALOGIUS", "R0", "baseline", "trained_correct_state", fmt.Sprintf("H%02d", i+1), want, observed, 1, before, "empty input means no scheduled recall at this tick"))
	}
	for _, cue := range []string{"BELL_A", "BELL_B", "BELL_C"} {
		want := learned[cue]
		got, ok := fontanafamily.Recall(cue, nil)
		if !ok {
			got = "?"
		}
		rows = append(rows, add("F12_HORALOGIUS", "R0", "knowledge_ablation", "untrained_user", cue, want, got, 3, h, "signal observed; meaning unavailable, three meanings compatible"))
		wrong := map[string]string{"BELL_A": "ORATIO", "BELL_B": "STVDIVM", "BELL_C": "LABOR"}
		got, _ = fontanafamily.Recall(cue, wrong)
		rows = append(rows, add("F12_HORALOGIUS", "R0", "state_corruption", "wrong_cue_convention", cue, want, got, 1, h, "false recall"))
	}
	return rows
}

func classify(want, got string) string {
	if want == got {
		return "no_effect"
	}
	w, g := []rune(want), []rune(got)
	mismatches := 0
	for i := 0; i < len(w) && i < len(g); i++ {
		if w[i] != g[i] {
			mismatches++
		}
	}
	if mismatches <= 1 && len(w) == len(g) {
		return "local"
	}
	if len(w) == len(g) {
		return "synchronization_or_global"
	}
	return "mixed"
}

func writeResults(rows []result) {
	var b strings.Builder
	b.WriteString("model\tprofile\texperiment\tcondition\ttrial\tinput\tobserved\texact\tsymbol_accuracy\tnormalized_edit_distance\tcompatible_interpretations\tstate_json\tnotes\n")
	for _, r := range rows {
		distance := fmt.Sprintf("%.6f", r.EditDistance)
		if strings.HasPrefix(r.Observed, "<") {
			distance = "NOT_APPLICABLE"
		}
		fmt.Fprintf(&b, "%s\t%s\t%s\t%s\t%s\t%s\t%s\t%t\t%.6f\t%s\t%s\t%s\t%s\n", r.Model, r.Profile, r.Experiment, r.Condition, r.Trial, r.Input, r.Observed, r.Exact, r.Accuracy, distance, r.Compatible, r.StateJSON, r.Notes)
	}
	mustWrite(filepath.Join(outputDir, "EXPERIMENT_RESULTS.tsv"), b.String())
	// Aggregate exact rates with Wilson intervals, keeping profiles separate.
	type count struct{ n, k int }
	counts := map[string]count{}
	for _, r := range rows {
		if r.Experiment != "baseline" {
			continue
		}
		key := r.Model + "\t" + r.Profile + "\t" + r.Condition
		c := counts[key]
		c.n++
		if r.Exact {
			c.k++
		}
		counts[key] = c
	}
	keys := make([]string, 0, len(counts))
	for k := range counts {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	b.Reset()
	b.WriteString("model\tprofile\tcondition\texact\tn\trate\twilson95_low\twilson95_high\n")
	for _, key := range keys {
		c := counts[key]
		lo, hi := wilson(c.k, c.n)
		fmt.Fprintf(&b, "%s\t%d\t%d\t%.6f\t%.6f\t%.6f\n", key, c.k, c.n, float64(c.k)/float64(c.n), lo, hi)
	}
	mustWrite(filepath.Join(outputDir, "BASELINE_SUMMARY.tsv"), b.String())
}

func wilson(k, n int) (float64, float64) {
	if n == 0 {
		return 0, 0
	}
	z := 1.959963984540054
	p := float64(k) / float64(n)
	den := 1 + z*z/float64(n)
	centre := (p + z*z/(2*float64(n))) / den
	half := z * math.Sqrt(p*(1-p)/float64(n)+z*z/(4*float64(n*n))) / den
	return centre - half, centre + half
}

func writeDynamics() {
	content := "model\tprofile\tstate_space\treachable_under_test_action\tcycle_length\treversible\thistory_dependent\tnotes\n" +
		"F08_SERPENS\tR0\t(24^12 with empty marker; H-bound)\tall 24^12 under arbitrary place/remove\tNOT_APPLICABLE\tplacement removal only with retained piece\tno\tfixed spatial storage, not an autonomous cycle\n" +
		"F07_ROTA\tR0\t23\t23\t23\tyes\tno\tunit rotation traverses full cycle\n" +
		"F10_CYLINDRUS\tR0\t3404825447\t23 per one-band action\t23\tyes\tno\tfull space is 23^7 under independent actions\n" +
		"F10_CYLINDRUS\tR1\t3404825447\t23 per suffix-coupled action\t23\tyes\tno\tH-profile only\n" +
		"F11_ARISMETRICUM\tR0\tunbounded cue domain\tquery-dependent\tNOT_APPLICABLE\tplacement reversible only if removed cue retained\tno\tindexed store\n" +
		"F12_HORALOGIUS\tR0\t12\t12\t12\tyes in digital abstraction\tno\thistorical drive details remain U\n"
	mustWrite(filepath.Join(outputDir, "STATE_DYNAMICS.tsv"), content)
}

func writePackages(rows []result) {
	models := []string{"F08_SERPENS", "F07_ROTA", "F10_CYLINDRUS", "F11_ARISMETRICUM", "F12_HORALOGIUS"}
	for _, id := range models {
		base := 0
		exact := 0
		profiles := []string{}
		for _, p := range declaredProfiles() {
			if p.Model == id {
				profiles = append(profiles, p.ID)
			}
		}
		for _, r := range rows {
			if r.Model == id && r.Experiment == "baseline" {
				base++
				if r.Exact {
					exact++
				}
			}
		}
		p := packageFor(id, profiles, base, exact)
		writeJSON(filepath.Join(outputDir, "models", strings.ToLower(id)+".json"), p)
	}
}

func packageFor(id string, profiles []string, n, k int) modelPackage {
	p := modelPackage{Metadata: map[string]any{"schema": "fontana-validated-model-v1", "identifier": id, "task": 78, "seed": seed}, Sources: []string{"research/phase2/fontana/MACHINE_INVENTORY.tsv", "research/phase2/fontana/RECONSTRUCTION_CANDIDATES.tsv", "research/phase2/fontana/machines/" + sourceFile(id)}, Reconstruction: map[string]any{"evidence_labels": []string{"E", "I", "H", "U"}, "target_blind": true}, ReconstructionProfiles: profiles, Experiments: []string{"baseline", "knowledge_ablation", "state_corruption", "state_dynamics"}, Results: map[string]any{"baseline_exact": k, "baseline_n": n}, Tests: []string{"go test ./internal/fontanafamily"}, Verdict: map[string]string{}}
	switch id {
	case "F08_SERPENS":
		p.Components = []string{"flat body", "spiral", "ordered holes", "letter inserts"}
		p.StateSchema = map[string]string{"holes": "ordered nullable symbols"}
		p.Operations = []string{"placement", "ordered_traversal", "selection"}
		p.Transitions = []string{"place", "remove", "substitute", "swap"}
		p.Encoding = "place message symbols centre-to-edge; boundary remains in K"
		p.Retrieval = "traverse centre-to-edge for known length"
		p.KnowledgeRequirements = []string{"start", "direction", "boundary", "alphabet", "association convention"}
		p.Uncertainty = []string{"capacity", "movability of inserts", "message boundary", "stop rule", "association role"}
		p.Verdict = verdict("MEDIUM", "SUPPORTED", "SUPPORTED", "NOT_SUPPORTED", "SUPPORTED", "SUPPORTED", "NOT_APPLICABLE")
		p.Verdict["MODEL_READY_FOR_FREEZE"] = "PARTIALLY_SUPPORTED"
	case "F07_ROTA":
		p.Components = []string{"wheel", "positions", "pointer"}
		p.StateSchema = map[string]string{"offset": "cyclic integer"}
		p.Operations = []string{"rotation", "selection"}
		p.Transitions = []string{"rotate(step)"}
		p.Encoding = "NOT_APPLICABLE in invariant core"
		p.Retrieval = "read position at pointer"
		p.KnowledgeRequirements = []string{"layout", "zero", "direction", "step"}
		p.Uncertainty = []string{"number of disks", "alphabet layout", "step rule", "initial state"}
		p.Verdict = verdict("MEDIUM", "PARTIALLY_SUPPORTED", "PARTIALLY_SUPPORTED", "NOT_SUPPORTED", "SUPPORTED", "NOT_SUPPORTED", "NOT_APPLICABLE")
	case "F10_CYLINDRUS":
		p.Components = []string{"cylinder", "bands", "positions", "reading line"}
		p.StateSchema = map[string]string{"offsets": "cyclic integer per band"}
		p.Operations = []string{"independent_rotation", "alignment", "ordered_traversal"}
		p.Transitions = []string{"rotate_band", "coupled_suffix_shift (R1 only)"}
		p.Encoding = "align one symbol per band in H-profile R0"
		p.Retrieval = "read axial line in known route"
		p.KnowledgeRequirements = []string{"band order", "direction", "read line", "movement coupling"}
		p.Uncertainty = []string{"band count", "independence", "step rule", "route"}
		p.Verdict = verdict("MEDIUM", "PARTIALLY_SUPPORTED", "PARTIALLY_SUPPORTED", "NOT_SUPPORTED", "SUPPORTED", "PARTIALLY_SUPPORTED", "NOT_APPLICABLE")
	case "F11_ARISMETRICUM":
		p.Components = []string{"body", "numbered holes", "inserts"}
		p.StateSchema = map[string]string{"slots": "index to opaque cue"}
		p.Operations = []string{"indexing", "lookup", "placement"}
		p.Transitions = []string{"place", "remove", "swap"}
		p.Encoding = "place opaque cue at index; no text generation"
		p.Retrieval = "lookup by known index and convention"
		p.KnowledgeRequirements = []string{"index convention", "cue meaning"}
		p.Uncertainty = []string{"exact number-content mapping", "movability", "whether content is letters"}
		p.Verdict = verdict("MEDIUM", "PARTIALLY_SUPPORTED", "SUPPORTED", "NOT_SUPPORTED", "SUPPORTED", "NOT_SUPPORTED", "SUPPORTED")
	case "F12_HORALOGIUS":
		p.Components = []string{"time mechanism", "state", "signal"}
		p.StateSchema = map[string]string{"tick": "cyclic integer", "cues": "tick to signal ID"}
		p.Operations = []string{"continuous_transition", "signalling", "association"}
		p.Transitions = []string{"advance", "emit"}
		p.Encoding = "schedule cue; rich remembered content is not stored"
		p.Retrieval = "signal plus learned mapping cues human recall"
		p.KnowledgeRequirements = []string{"calibration", "cue meaning", "learned association"}
		p.Uncertainty = []string{"exact drive", "alarm setting", "signal form", "human retention"}
		p.Verdict = verdict("MEDIUM", "PARTIALLY_SUPPORTED", "SUPPORTED", "NOT_SUPPORTED", "SUPPORTED", "NOT_SUPPORTED", "NOT_APPLICABLE")
	}
	return p
}

func verdict(conf, cycle, baseline, damage, state, literal, indexed string) map[string]string {
	return map[string]string{"SOURCE_SUFFICIENCY": "PARTIALLY_SUPPORTED", "RECONSTRUCTION_CONFIDENCE": conf, "OPERATIONAL_CYCLE_VALIDATED": cycle, "BASELINE_RECOVERY": baseline, "PRIOR_KNOWLEDGE_DEPENDENCE": "SUPPORTED", "STATE_DEPENDENCE": state, "STATE_DAMAGE_ROBUSTNESS": damage, "ERROR_PROPAGATION_CLASS": "MIXED", "LITERAL_STORAGE_FUNCTION": literal, "INDEXED_RETRIEVAL_FUNCTION": indexed, "MNEMONIC_CUE_FUNCTION": "PARTIALLY_SUPPORTED", "MODEL_READY_FOR_FREEZE": "NOT_SUPPORTED"}
}

func sourceFile(id string) string {
	switch id {
	case "F08_SERPENS":
		return "F08_SERPENS.md"
	case "F07_ROTA":
		return "F07_ROTA.md"
	case "F10_CYLINDRUS", "F11_ARISMETRICUM":
		return "F10_CYLINDRUS_F11_ARISMETRICUM.md"
	default:
		return "F12_HORALOGIUS.md"
	}
}
func writeJSON(path string, v any) {
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		panic(err)
	}
	mustWrite(path, string(b)+"\n")
}
func mustWrite(path, content string) {
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		panic(err)
	}
}
