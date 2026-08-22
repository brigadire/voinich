package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

// Target is one preregistered (family, metric, corpus-baseline, Voynich
// target) triple, sourced from a frozen Task58-65 artifact (task66
// section 8): the value is read directly from that file, never
// recomputed by a second implementation of the metric.
type Target struct {
	Family, Metric, SourceTask, Artifact, Field string
	Voynich                                     float64
	Status                                      string
}

// tsvRow reads path as a header+rows TSV and returns rows keyed by
// header name, plus the raw rows in file order.
func tsvRow(path string) (headers []string, rows []map[string]string, err error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, nil, err
	}
	defer f.Close()
	sc := bufio.NewScanner(f)
	sc.Buffer(make([]byte, 4096), 1<<20)
	first := true
	for sc.Scan() {
		fields := strings.Split(sc.Text(), "\t")
		if first {
			headers = fields
			first = false
			continue
		}
		row := map[string]string{}
		for i, h := range headers {
			if i < len(fields) {
				row[h] = fields[i]
			}
		}
		rows = append(rows, row)
	}
	return headers, rows, sc.Err()
}

func findRow(rows []map[string]string, key, value string) map[string]string {
	for _, r := range rows {
		if r[key] == value {
			return r
		}
	}
	return nil
}

func f64(s string) float64 {
	v, _ := strconv.ParseFloat(strings.TrimSpace(s), 64)
	return v
}

// LoadVoynichTargets reads the seven authoritative Task58-65 artifacts
// (task66 section 8/44-51) and returns one Target per preregistered
// metric. It errors loudly (MISSING_ARTIFACT) rather than silently
// defaulting to zero if a frozen artifact is absent (task66 section 10).
func LoadVoynichTargets() ([]Target, error) {
	var out []Target
	add := func(family, metric, task, artifact, field string, value float64, status string) {
		out = append(out, Target{Family: family, Metric: metric, SourceTask: task, Artifact: artifact, Field: field, Voynich: value, Status: status})
	}
	missing := func(family, metric, task, artifact, field string) {
		add(family, metric, task, artifact, field, 0, "MISSING_ARTIFACT")
	}

	// A. TOKEN_ORDER (Task58)
	art := "experiments/rozanova-temerev-v1/comparison.tsv"
	if _, rows, err := tsvRow(art); err == nil {
		if v := findRow(rows, "corpus", "Voynich"); v != nil {
			add("TOKEN_ORDER", "token_order_bits", "Task58", art, "token_corrected_bits", f64(v["token_corrected_bits"]), "VALUE")
			add("TOKEN_ORDER", "edge_order_bits", "Task58", art, "edge_corrected_bits", f64(v["edge_corrected_bits"]), "VALUE")
		} else {
			missing("TOKEN_ORDER", "token_order_bits", "Task58", art, "token_corrected_bits")
			missing("TOKEN_ORDER", "edge_order_bits", "Task58", art, "edge_corrected_bits")
		}
	} else {
		missing("TOKEN_ORDER", "token_order_bits", "Task58", art, "token_corrected_bits")
		missing("TOKEN_ORDER", "edge_order_bits", "Task58", art, "edge_corrected_bits")
	}

	// B. POSITIONAL_STRUCTURE (Task59)
	art = "experiments/glyph-position-v1/POSITIONAL_SPECIALIZATION_COMPARISON.tsv"
	if _, rows, err := tsvRow(art); err == nil {
		if v := findRow(rows, "Corpus", "Voynich"); v != nil {
			add("POSITIONAL_STRUCTURE", "weighted_entropy", "Task59", art, "WeightedEntropy", f64(v["WeightedEntropy"]), "VALUE")
			add("POSITIONAL_STRUCTURE", "high_freq_specialists", "Task59", art, "HighFreqSpecialists", f64(v["HighFreqSpecialists"]), "VALUE")
		} else {
			missing("POSITIONAL_STRUCTURE", "weighted_entropy", "Task59", art, "WeightedEntropy")
			missing("POSITIONAL_STRUCTURE", "high_freq_specialists", "Task59", art, "HighFreqSpecialists")
		}
	} else {
		missing("POSITIONAL_STRUCTURE", "weighted_entropy", "Task59", art, "WeightedEntropy")
		missing("POSITIONAL_STRUCTURE", "high_freq_specialists", "Task59", art, "HighFreqSpecialists")
	}

	// C. REPETITION_EDIT_GEOMETRY (Task60): giant-component fraction is
	// LargestComponent (EDIT_FAMILIES.tsv) over vocabulary size (types,
	// comparison.tsv); exact-adjacent-repeat rate is ObservedR2
	// (NULL_EXACT_REPETITION.tsv).
	familiesArt := "experiments/token-repetition-v1/EDIT_FAMILIES.tsv"
	nullArt := "experiments/token-repetition-v1/NULL_EXACT_REPETITION.tsv"
	if _, rows, err := tsvRow(familiesArt); err == nil {
		if v := findRow(rows, "Corpus", "Voynich"); v != nil {
			if _, crows, err2 := tsvRow("experiments/rozanova-temerev-v1/comparison.tsv"); err2 == nil {
				if cv := findRow(crows, "corpus", "Voynich"); cv != nil && f64(cv["types"]) > 0 {
					add("REPETITION_EDIT_GEOMETRY", "giant_component_fraction", "Task60", familiesArt, "LargestComponent/types", f64(v["LargestComponent"])/f64(cv["types"]), "VALUE")
				} else {
					missing("REPETITION_EDIT_GEOMETRY", "giant_component_fraction", "Task60", familiesArt, "LargestComponent/types")
				}
			}
		} else {
			missing("REPETITION_EDIT_GEOMETRY", "giant_component_fraction", "Task60", familiesArt, "LargestComponent/types")
		}
	} else {
		missing("REPETITION_EDIT_GEOMETRY", "giant_component_fraction", "Task60", familiesArt, "LargestComponent/types")
	}
	if _, rows, err := tsvRow(nullArt); err == nil {
		if v := findRow(rows, "Corpus", "Voynich"); v != nil {
			add("REPETITION_EDIT_GEOMETRY", "exact_adjacent_repeat_rate", "Task60", nullArt, "ObservedR2", f64(v["ObservedR2"]), "VALUE")
		} else {
			missing("REPETITION_EDIT_GEOMETRY", "exact_adjacent_repeat_rate", "Task60", nullArt, "ObservedR2")
		}
	} else {
		missing("REPETITION_EDIT_GEOMETRY", "exact_adjacent_repeat_rate", "Task60", nullArt, "ObservedR2")
	}

	// D. CHARACTER_ENTROPY (Task61): TOKEN_BOUNDARY orders 0-3 => h1-h4.
	art = "experiments/character-entropy-v1/ENTROPY_BY_ORDER.tsv"
	if _, rows, err := tsvRow(art); err == nil {
		for i, name := range []string{"h1", "h2", "h3", "h4"} {
			found := false
			for _, r := range rows {
				if r["Corpus"] == "Voynich" && r["Mode"] == "TOKEN_BOUNDARY" && r["Order"] == strconv.Itoa(i) {
					add("CHARACTER_ENTROPY", name, "Task61", art, "EntropyBits@Order="+strconv.Itoa(i), f64(r["EntropyBits"]), "VALUE")
					found = true
					break
				}
			}
			if !found {
				missing("CHARACTER_ENTROPY", name, "Task61", art, "EntropyBits@Order="+strconv.Itoa(i))
			}
		}
	} else {
		for _, name := range []string{"h1", "h2", "h3", "h4"} {
			missing("CHARACTER_ENTROPY", name, "Task61", art, "EntropyBits")
		}
	}

	// E. TOKEN_FORMATION (Task62): position/order cross-entropy gain over IID.
	art = "experiments/token-formation-v1/MODEL_HELDOUT_FIT.tsv"
	if _, rows, err := tsvRow(art); err == nil {
		iid := findRow(rows, "Model", "IID")
		pos := findRow(rows, "Model", "POSITION_IID")
		mk1 := findRow(rows, "Model", "MARKOV_1")
		if iid != nil && pos != nil {
			add("TOKEN_FORMATION", "position_gain_bits", "Task62", art, "TestCrossEntropy(IID)-TestCrossEntropy(POSITION_IID)", f64(iid["TestCrossEntropy"])-f64(pos["TestCrossEntropy"]), "VALUE")
		} else {
			missing("TOKEN_FORMATION", "position_gain_bits", "Task62", art, "TestCrossEntropy")
		}
		if iid != nil && mk1 != nil {
			add("TOKEN_FORMATION", "order_gain_bits", "Task62", art, "TestCrossEntropy(IID)-TestCrossEntropy(MARKOV_1)", f64(iid["TestCrossEntropy"])-f64(mk1["TestCrossEntropy"]), "VALUE")
		} else {
			missing("TOKEN_FORMATION", "order_gain_bits", "Task62", art, "TestCrossEntropy")
		}
	} else {
		missing("TOKEN_FORMATION", "position_gain_bits", "Task62", art, "TestCrossEntropy")
		missing("TOKEN_FORMATION", "order_gain_bits", "Task62", art, "TestCrossEntropy")
	}

	// F. LOCAL_TRANSITION (Task63): adjacent near-rate and its residual
	// over separation 10.
	art = "experiments/token-transition-v1/DISTANCE_BY_SEPARATION.tsv"
	if _, rows, err := tsvRow(art); err == nil {
		sep1 := findRow(rows, "Separation", "1")
		sep10 := findRow(rows, "Separation", "10")
		if sep1 != nil {
			add("LOCAL_TRANSITION", "adjacent_near_rate", "Task63", art, "NearRate@Separation=1", f64(sep1["NearRate"]), "VALUE")
		} else {
			missing("LOCAL_TRANSITION", "adjacent_near_rate", "Task63", art, "NearRate@Separation=1")
		}
		if sep1 != nil && sep10 != nil {
			add("LOCAL_TRANSITION", "residual_adjacency", "Task63", art, "NearRate@1-NearRate@10", f64(sep1["NearRate"])-f64(sep10["NearRate"]), "VALUE")
		} else {
			missing("LOCAL_TRANSITION", "residual_adjacency", "Task63", art, "NearRate@1-NearRate@10")
		}
	} else {
		missing("LOCAL_TRANSITION", "adjacent_near_rate", "Task63", art, "NearRate")
		missing("LOCAL_TRANSITION", "residual_adjacency", "Task63", art, "NearRate")
	}

	// G. LOCAL_REGIME_TOPOLOGY (Task64-65)
	clArt := "experiments/local-regime-topology-v1/CORRELATION_LENGTH.tsv"
	if _, rows, err := tsvRow(clArt); err == nil {
		found := false
		for _, r := range rows {
			if r["Unit"] == "TOKEN" && r["Threshold"] == "50pct" {
				add("LOCAL_REGIME_TOPOLOGY", "correlation_length_tokens", "Task65", clArt, "Value@Unit=TOKEN,Threshold=50pct", f64(r["Value"]), "VALUE")
				found = true
				break
			}
		}
		if !found {
			missing("LOCAL_REGIME_TOPOLOGY", "correlation_length_tokens", "Task65", clArt, "Value")
		}
	} else {
		missing("LOCAL_REGIME_TOPOLOGY", "correlation_length_tokens", "Task65", clArt, "Value")
	}
	manifestArt := "experiments/local-regime-topology-v1/manifest.json"
	if b, err := os.ReadFile(manifestArt); err == nil {
		s := string(b)
		if strings.Contains(s, `"topology": "MIXED_DRIFT_AND_STATES"`) {
			add("LOCAL_REGIME_TOPOLOGY", "topology_class_is_mixed", "Task65", manifestArt, "topology", 1, "VALUE")
		} else {
			add("LOCAL_REGIME_TOPOLOGY", "topology_class_is_mixed", "Task65", manifestArt, "topology", 0, "VALUE")
		}
	} else {
		missing("LOCAL_REGIME_TOPOLOGY", "topology_class_is_mixed", "Task65", manifestArt, "topology")
	}
	bdArt := "experiments/local-regime-topology-v1/BOUNDARY_DISCONTINUITY.tsv"
	if _, rows, err := tsvRow(bdArt); err == nil {
		if v := findRow(rows, "Boundary", "LINE_BOUNDARY"); v != nil {
			add("LOCAL_REGIME_TOPOLOGY", "line_boundary_delta", "Task65", bdArt, "Delta@Boundary=LINE_BOUNDARY", f64(v["Delta"]), "VALUE")
		} else {
			missing("LOCAL_REGIME_TOPOLOGY", "line_boundary_delta", "Task65", bdArt, "Delta")
		}
	} else {
		missing("LOCAL_REGIME_TOPOLOGY", "line_boundary_delta", "Task65", bdArt, "Delta")
	}

	return out, nil
}

// WriteTargetManifest writes VOYNICH_TARGET_MANIFEST.tsv (task66 section
// 8): source task, artifact, field, value, status for every preregistered
// target metric.
func WriteTargetManifest(path string, targets []Target) error {
	var b strings.Builder
	b.WriteString("family\tmetric\tsource_task\tartifact\tfield\tvalue\tstatus\n")
	for _, t := range targets {
		val := ""
		if t.Status == "VALUE" {
			val = fmt.Sprintf("%.9g", t.Voynich)
		}
		b.WriteString(fmt.Sprintf("%s\t%s\t%s\t%s\t%s\t%s\t%s\n", t.Family, t.Metric, t.SourceTask, t.Artifact, t.Field, val, t.Status))
	}
	return os.WriteFile(path, []byte(b.String()), 0644)
}
