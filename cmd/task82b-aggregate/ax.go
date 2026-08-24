package main

import "zcore.dev/voinich/internal/task82b"

func writeAXOutputs(out string, recs []Rec) error {
	reg := newTSV("ax_id", "description", "implemented", "redundant_with_f2_metric", "note")
	reg.row("AX1", "information concentration at structural beginnings/endings", "false", "BP1_BOUNDARY_TOKEN_NMI, LS2_POSITIONAL_LEXICON_NMI", "audited redundant with existing F2 subset, not reimplemented (sec.61)")
	reg.row("AX2", "first-position vs internal-position distribution divergence", "false", "BP1_BOUNDARY_TOKEN_NMI, LS2_POSITIONAL_LEXICON_NMI", "audited redundant with existing F2 subset, not reimplemented (sec.61)")
	reg.row("AX3", "extracted-stream entropy relative to matched random extraction", "true", "", "Shannon entropy of the rendered output stream")
	reg.row("AX4", "extracted-stream lexical coherence proxy without language-specific decoding", "true", "", "type/token ratio")
	reg.row("AX5", "periodic-position information excess", "true", "", "max NMI(stream identity, index mod k) over k in {2,3,5,7}")
	reg.row("AX6", "cross-line positional persistence", "true", "", "adjacent-line first-atom match rate / (1/vocab size)")
	reg.row("AX7", "mutual information between structural position and glyph/token class", "false", "BP1_BOUNDARY_TOKEN_NMI, 2DL1_LAYOUT_POSITION_MI", "audited redundant with existing F2 subset, not reimplemented (sec.61)")
	if err := reg.write(out, "AX_REGISTRY.tsv"); err != nil {
		return err
	}

	if err := writeAXValidation(out); err != nil {
		return err
	}

	res := newTSV("job_id", "kind", "corpus_id", "operator_id", "null_class", "shorthand_variant", "shorthand_scale", "ax3_entropy", "ax4_ttr", "ax5_nmi_max", "ax5_best_period", "ax6_line_persistence", "n")
	for _, r := range recs {
		if r.AX == nil {
			continue
		}
		res.row(r.JobID, r.Kind, r.CorpusID, r.OperatorID, r.NullClass, r.ShorthandVariant, r.ShorthandScale,
			fstr(r.AX.AX3StreamEntropy), fstr(r.AX.AX4TypeTokenRatio), fstr(r.AX.AX5PeriodicNMIMax), istr(r.AX.AX5BestPeriod), fstr(r.AX.AX6LinePersistence), istr(r.AX.N))
	}
	return res.write(out, "AX_RESULTS.tsv")
}

// writeAXValidation recomputes the sec.48-50 validation gate: a
// SYNTHETIC period-2 positive control and a deterministic-shuffle
// negative control, both evaluated with the exact same ComputeAX/AX5
// code path used on real data (mirrors internal/task82b/
// ax_synthetic_test.go, kept here too so the validation result is a
// pipeline artifact, not only a test log).
func writeAXValidation(out string) error {
	pattern := []string{"a", "b"}
	periodic := make([]string, 400)
	for i := range periodic {
		periodic[i] = pattern[i%2]
	}
	positive := task82b.ComputeAX([][]string{periodic})

	shuffled := append([]string{}, periodic...)
	seed := uint64(12345)
	next := func() uint64 { seed = seed*6364136223846793005 + 1; return seed }
	for i := len(shuffled) - 1; i > 0; i-- {
		j := int(next() % uint64(i+1))
		shuffled[i], shuffled[j] = shuffled[j], shuffled[i]
	}
	negative := task82b.ComputeAX([][]string{shuffled})

	sensitive := positive.AX5PeriodicNMIMax >= 0.3
	separated := positive.AX5PeriodicNMIMax > negative.AX5PeriodicNMIMax

	w := newTSV("ax_id", "control", "description", "ax5_value", "ax5_period", "gate_component", "result")
	w.row("AX5", "SYNTHETIC_POSITIVE", "period-2 alternating stream (a b a b ...), n=400", fstr(positive.AX5PeriodicNMIMax), istr(positive.AX5BestPeriod), "positive_control_sensitivity", bstr(sensitive))
	w.row("AX5", "NEGATIVE", "deterministic shuffle of the same 400 symbols", fstr(negative.AX5PeriodicNMIMax), istr(negative.AX5BestPeriod), "null_calibration", bstr(negative.AX5PeriodicNMIMax < 0.1))
	w.row("AX5", "SUMMARY", "positive vs negative separation", fstr(positive.AX5PeriodicNMIMax-negative.AX5PeriodicNMIMax), "", "positive_beats_negative", bstr(separated))
	w.row("AX5", "SUMMARY", "cross-corpus robustness", "", "", "cross_corpus_robustness", "false")
	w.row("AX5", "GATE", "sec.50: positive-control sensitivity AND null calibration AND cross-corpus robustness", "", "", "overall", "PARTIAL: no cross-corpus replication of the positive control was attempted (no openly available historical-acrostic corpus was located)")
	w.row("AX3", "NOT_VALIDATED", "no dedicated positive/negative control constructed", "", "", "gate", "NOT_VALIDATED")
	w.row("AX4", "NOT_VALIDATED", "no dedicated positive/negative control constructed", "", "", "gate", "NOT_VALIDATED")
	w.row("AX6", "NOT_VALIDATED", "no dedicated positive/negative control constructed", "", "", "gate", "NOT_VALIDATED")
	return w.write(out, "AX_VALIDATION.tsv")
}
