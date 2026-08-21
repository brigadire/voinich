package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"zcore.dev/voinich/internal/corpustransform"
	"zcore.dev/voinich/internal/inversetransposition"
	"zcore.dev/voinich/internal/voynichvalidation"
)

type manifest struct {
	Schema                int                            `json:"schema"`
	SourceCorpus          string                         `json:"source_corpus"`
	SourceSHA256          string                         `json:"source_sha256"`
	DiscoveryArtifact     string                         `json:"discovery_artifact"`
	DiscoverySHA256       string                         `json:"discovery_sha256"`
	GitCommit             string                         `json:"git_commit"`
	FixedCandidate        inversetransposition.Candidate `json:"fixed_candidate"`
	Objective             string                         `json:"objective_version"`
	SplitDefinition       string                         `json:"split_definition"`
	DiscoverySHA256Corpus string                         `json:"discovery_corpus_sha256"`
	HoldoutSHA256         string                         `json:"holdout_corpus_sha256"`
	NullConfiguration     string                         `json:"null_configuration"`
	CalibrationProvenance string                         `json:"fixed_calibration_provenance"`
}

func main() {
	fs := flag.NewFlagSet("voynich-validation", flag.ExitOnError)
	input := fs.String("input", "data_work/ZL3b-x7.canonical.txt", "canonical Voynich corpus")
	discovery := fs.String("discovery", "experiments/inverse-transposition/voynich-discovery", "saved discovery directory")
	out := fs.String("output-dir", "experiments/inverse-transposition/voynich-validation", "validation output")
	nullN := fs.Int("null-replicates", 30, "fixed null budget")
	seed := fs.Int64("null-seed", 5401, "fixed null seed")
	fs.Parse(os.Args[1:])
	if err := run(*input, *discovery, *out, *nullN, *seed); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(input, discovery, out string, nullN int, seed int64) error {
	tokens, lengths, sourceBytes, err := voynichvalidation.Read(input)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(out, 0755); err != nil {
		return err
	}
	discoveryTokens, holdoutTokens := voynichvalidation.SplitByLines(tokens, lengths)
	writeCorpus := func(path string, ts []string) error {
		b, _ := corpustransform.WriteCorpus(ts, nil, corpustransform.LinePolicyReflow)
		return os.WriteFile(path, b, 0644)
	}
	if err := writeCorpus(filepath.Join(out, "discovery.corpus.txt"), discoveryTokens); err != nil {
		return err
	}
	if err := writeCorpus(filepath.Join(out, "holdout.corpus.txt"), holdoutTokens); err != nil {
		return err
	}
	original := inversetransposition.Measure(tokens)
	holdBase := inversetransposition.Measure(holdoutTokens)
	fixed := inversetransposition.Candidate{Width: 2, Order: "natural", Rounds: 1, Seed: 1}
	fixedTokens, err := fixed.Apply(holdoutTokens)
	if err != nil {
		return err
	}
	holdFixed := inversetransposition.Measure(fixedTokens)
	entries, err := os.ReadDir(discovery)
	if err != nil {
		return err
	}
	rows := make([]voynichvalidation.CandidateRow, 0)
	for _, e := range entries {
		if !strings.HasPrefix(e.Name(), "w") || !strings.HasSuffix(e.Name(), ".txt") {
			continue
		}
		b, er := os.ReadFile(filepath.Join(discovery, e.Name()))
		if er != nil {
			return er
		}
		ts := strings.Fields(string(b))
		m := inversetransposition.Measure(ts)
		name := strings.TrimSuffix(e.Name(), ".txt")
		parts := strings.Split(name, "-")
		if len(parts) < 3 {
			return fmt.Errorf("invalid candidate filename %q", e.Name())
		}
		var width int
		fmt.Sscanf(parts[0], "w%d", &width)
		rows = append(rows, voynichvalidation.CandidateRow{ID: name, Width: width, Order: parts[1], Raw: m})
	} // score/rank values are copied only from the frozen manifest, never recomputed as a new search.
	var saved struct {
		Candidates []struct {
			Candidate inversetransposition.Candidate `json:"candidate"`
			Score     float64                        `json:"score"`
		}
	}
	b, er := os.ReadFile(filepath.Join(discovery, "search-manifest.json"))
	if er != nil {
		return er
	}
	if json.Unmarshal(b, &saved) != nil {
		return fmt.Errorf("invalid discovery manifest")
	}
	for i := range rows {
		for _, x := range saved.Candidates {
			if x.Candidate.ID() == rows[i].ID {
				rows[i].Score = x.Score
			}
		}
	}
	voynichvalidation.CandidateRows(original, rows)
	nulls := voynichvalidation.NullDistribution(holdoutTokens, nullN, seed, holdBase)
	if err := writeTSV(out, original, rows, holdBase, holdFixed, nulls); err != nil {
		return err
	}
	commit := "unknown"
	if x, e := exec.Command("git", "rev-parse", "HEAD").Output(); e == nil {
		commit = strings.TrimSpace(string(x))
	}
	dm, _ := os.ReadFile(filepath.Join(discovery, "search-manifest.json"))
	sm, _ := os.ReadFile(filepath.Join(out, "discovery.corpus.txt"))
	hm, _ := os.ReadFile(filepath.Join(out, "holdout.corpus.txt"))
	mf := manifest{1, input, voynichvalidation.SHA256(sourceBytes), filepath.Join(discovery, "search-manifest.json"), voynichvalidation.SHA256(dm), commit, fixed, inversetransposition.ObjectiveVersion, "first 80% of existing logical lines for discovery, last 20% for holdout; split was not pre-registered in Task54 (limitation)", voynichvalidation.SHA256(sm), voynichvalidation.SHA256(hm), fmt.Sprintf("%d deterministic full-token Fisher-Yates permutations; seed=%d", nullN, seed), "INVERSE_TRANSPOSITION_TASK54_1_REPORT.md, frozen Doyle/T2/T4/T8 ranges; no Voynich values used"}
	mb, _ := json.MarshalIndent(mf, "", "  ")
	if err := os.WriteFile(filepath.Join(out, "manifest.json"), append(mb, '\n'), 0644); err != nil {
		return err
	}
	if err := writeReport(out, original, holdBase, holdFixed, rows, nulls); err != nil {
		return err
	}
	if err := writeControlAudit(out); err != nil {
		return err
	}
	return nil
}

func writeControlAudit(out string) error {
	controls := []struct{ name, original, transformed string }{
		{"Doyle", "data_test/pg2097-2.txt", "data_test/transformed/doyle__transposition__w002__natural__seed001.txt"},
		{"Longfellow", "data_test/pg30795-mod.txt", "data_test/transformed/longfellow__transposition__w002__natural__seed001.txt"},
	}
	var s strings.Builder
	s.WriteString("corpus\tstate\ttransition\trelation\tsequence2\tsequence3\tdelta_sequence3\n")
	for _, c := range controls {
		ot, _, _, err := voynichvalidation.Read(c.original)
		if err != nil {
			return err
		}
		tt, _, _, err := voynichvalidation.Read(c.transformed)
		if err != nil {
			return err
		}
		o := inversetransposition.Measure(ot)
		x := inversetransposition.Measure(tt)
		d := voynichvalidation.Delta(x, o)
		fmt.Fprintf(&s, "%s\toriginal\t%s\t0\n", c.name, voynichvalidation.FormatMetric(o))
		fmt.Fprintf(&s, "%s\tinverse_w2_control\t%s\t%.12g\n", c.name, voynichvalidation.FormatMetric(x), d.HigherOrderRepetition)
	}
	return os.WriteFile(filepath.Join(out, "width2_natural_controls.tsv"), []byte(s.String()), 0644)
}

func writeTSV(out string, original inversetransposition.Metrics, rows []voynichvalidation.CandidateRow, hb, hf inversetransposition.Metrics, nulls []voynichvalidation.NullRow) error {
	var s strings.Builder
	s.WriteString("metric\ttransition\trelation\tsequence2\tsequence3\noriginal\t" + voynichvalidation.FormatMetric(original) + "\n")
	if err := os.WriteFile(filepath.Join(out, "original_metrics.tsv"), []byte(s.String()), 0644); err != nil {
		return err
	}
	var d strings.Builder
	d.WriteString("candidate\twidth\torder\tsearch_score\ttransition_raw\ttransition_delta\ttransition_relative_delta\trelation_raw\trelation_delta\trelation_relative_delta\tsequence2_raw\tsequence2_delta\tsequence2_relative_delta\tsequence3_raw\tsequence3_delta\tsequence3_relative_delta\n")
	for _, r := range rows {
		fmt.Fprintf(&d, "%s\t%d\t%s\t%.12g\t%s\t%s\t%s\n", r.ID, r.Width, r.Order, r.Score, voynichvalidation.FormatMetric(r.Raw), voynichvalidation.FormatMetric(r.Delta), voynichvalidation.FormatMetric(r.Relative))
	}
	if err := os.WriteFile(filepath.Join(out, "discovery_effects.tsv"), []byte(d.String()), 0644); err != nil {
		return err
	}
	var landscape strings.Builder
	landscape.WriteString("width\torder\tsearch_score\ttransition_raw\trelation_raw\tsequence2_raw\tsequence3_raw\n")
	for _, r := range rows {
		fmt.Fprintf(&landscape, "%d\t%s\t%.12g\t%s\n", r.Width, r.Order, r.Score, voynichvalidation.FormatMetric(r.Raw))
	}
	if err := os.WriteFile(filepath.Join(out, "parameter_landscape.tsv"), []byte(landscape.String()), 0644); err != nil {
		return err
	}
	var h strings.Builder
	h.WriteString("corpus\ttransition\trelation\tsequence2\tsequence3\noriginal_holdout\t" + voynichvalidation.FormatMetric(hb) + "\nw2_natural_holdout\t" + voynichvalidation.FormatMetric(hf) + "\nholdout_delta\t" + voynichvalidation.FormatMetric(voynichvalidation.Delta(hf, hb)) + "\n")
	if err := os.WriteFile(filepath.Join(out, "holdout_metrics.tsv"), []byte(h.String()), 0644); err != nil {
		return err
	}
	var n strings.Builder
	n.WriteString("replicate\ttransition_delta\trelation_delta\tsequence2_delta\tsequence3_delta\tfixed_calibration_effect_score\n")
	for _, x := range nulls {
		fmt.Fprintf(&n, "%d\t%s\t%.12g\n", x.Replicate, voynichvalidation.FormatMetric(x.Delta), x.Score)
	}
	if err := os.WriteFile(filepath.Join(out, "null_distribution.tsv"), []byte(n.String()), 0644); err != nil {
		return err
	}
	var summary strings.Builder
	summary.WriteString("metric\tmean\tmedian\tq025\tq975\tw2_empirical_percentile\n")
	deltas := [4][]float64{{}, {}, {}, {}}
	for _, x := range nulls {
		v := []float64{x.Delta.TransitionConcentration, x.Delta.RelationSignificance, x.Delta.SequenceRepetition, x.Delta.HigherOrderRepetition}
		for i := range v {
			deltas[i] = append(deltas[i], v[i])
		}
	}
	w2p := voynichvalidation.Percentile(nulls, voynichvalidation.Delta(hf, hb))
	p := []float64{w2p.TransitionConcentration, w2p.RelationSignificance, w2p.SequenceRepetition, w2p.HigherOrderRepetition}
	for i, name := range []string{"transition", "relation", "sequence2", "sequence3"} {
		var sum float64
		for _, x := range deltas[i] {
			sum += x
		}
		fmt.Fprintf(&summary, "%s\t%.12g\t%.12g\t%.12g\t%.12g\t%.12g\n", name, sum/float64(len(deltas[i])), voynichvalidation.Quantile(append([]float64(nil), deltas[i]...), .5), voynichvalidation.Quantile(append([]float64(nil), deltas[i]...), .025), voynichvalidation.Quantile(append([]float64(nil), deltas[i]...), .975), p[i])
	}
	return os.WriteFile(filepath.Join(out, "null_summary.tsv"), []byte(summary.String()), 0644)
}

func writeReport(out string, original, hb, hf inversetransposition.Metrics, rows []voynichvalidation.CandidateRow, nulls []voynichvalidation.NullRow) error {
	var s strings.Builder
	s.WriteString("# Voynich inverse-transposition validation\n\nThis is post-search validation. `structural-v2` search scores are ranking values only; they are not effect sizes or percentages. Raw effect is `candidate - original`.\n\n## Fixed holdout\n\n| corpus | transition | relation | sequence-2 | sequence-3 |\n|---|---:|---:|---:|---:|\n")
	hd := voynichvalidation.Delta(hf, hb)
	fmt.Fprintf(&s, "| original | %.12g | %.12g | %.12g | %.12g |\n| w2 natural | %.12g | %.12g | %.12g | %.12g |\n| delta | %.12g | %.12g | %.12g | %.12g |\n", hb.TransitionConcentration, hb.RelationSignificance, hb.SequenceRepetition, hb.HigherOrderRepetition, hf.TransitionConcentration, hf.RelationSignificance, hf.SequenceRepetition, hf.HigherOrderRepetition, hd.TransitionConcentration, hd.RelationSignificance, hd.SequenceRepetition, hd.HigherOrderRepetition)
	var w2 voynichvalidation.CandidateRow
	for _, r := range rows {
		if r.ID == "w002-natural-r01" {
			w2 = r
			break
		}
	}
	s.WriteString("\n## Discovery versus holdout\n\n| metric | discovery delta | holdout delta | direction |\n|---|---:|---:|---|\n")
	fmt.Fprintf(&s, "| transition | %.12g | %.12g | %s |\n| relation | %.12g | %.12g | %s |\n| sequence-2 | %.12g | %.12g | %s |\n| sequence-3 | %.12g | %.12g | %s |\n", w2.Delta.TransitionConcentration, hd.TransitionConcentration, direction(w2.Delta.TransitionConcentration, hd.TransitionConcentration), w2.Delta.RelationSignificance, hd.RelationSignificance, direction(w2.Delta.RelationSignificance, hd.RelationSignificance), w2.Delta.SequenceRepetition, hd.SequenceRepetition, direction(w2.Delta.SequenceRepetition, hd.SequenceRepetition), w2.Delta.HigherOrderRepetition, hd.HigherOrderRepetition, direction(w2.Delta.HigherOrderRepetition, hd.HigherOrderRepetition))
	s.WriteString("\nThe complete candidate table is `discovery_effects.tsv`; `parameter_landscape.tsv` contains the same frozen candidates with raw metrics. The w2 score is a candidate-local min-max ranking score and is not interpreted as improvement.\n\n## Controls and calibration\n\nDoyle/T2/T4/T8 ranges are taken unchanged from `INVERSE_TRANSPOSITION_TASK54_1_REPORT.md`; no Voynich candidate participates. `fixed_calibration_effect_score` is post-hoc. The width-2 natural-text audit is in `width2_natural_controls.tsv` for Doyle and Longfellow.\n\n## Null\n\nHoldout nulls use the fixed budget and deterministic seed in `manifest.json`. Signed raw-delta percentiles are descriptive; the full distribution and fixed-calibration composite are in `null_distribution.tsv`.\n\n## Classification\n\nNo pre-registered meaningful-effect threshold exists, so there is no significance claim. The split was not pre-registered and is documented as a limitation. No conclusion about decryption, decipherment, or plaintext recovery is made.\n")
	return os.WriteFile(filepath.Join(out, "VALIDATION_REPORT.md"), []byte(s.String()), 0644)
}

func direction(a, b float64) string {
	if a == 0 || b == 0 {
		return "NO_MEANINGFUL_CHANGE"
	}
	if (a > 0) == (b > 0) {
		return "SAME_DIRECTION"
	}
	return "OPPOSITE_DIRECTION"
}
