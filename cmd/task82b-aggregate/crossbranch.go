package main

import (
	"sort"

	"zcore.dev/voinich/internal/task82b"
)

// writeCrossBranchOutputs writes INFORMATION_REDUCTION_COMPARISON.tsv and
// INPUT_DEPENDENCE_COMPARISON.tsv (task82b.txt sec.54/56/57), comparing
// both branches on the same footing without any Voynich reference.
func writeCrossBranchOutputs(out string, recs []Rec) error {
	idx := indexExtraction(recs)
	sidx := indexShorthand(recs)

	red := newTSV("branch", "scope", "variant_or_operator", "length_ratio", "type_token_ratio_before", "type_token_ratio_after", "entropy_before", "entropy_after", "retained_information_fraction")

	for _, corpus := range task82b.CarrierOrder {
		base, ok := idx.baseline[corpus]
		if !ok || base.AX == nil {
			continue
		}
		ttrBefore := ttrOf(base)
		for _, op := range task82b.Registry() {
			outRec, ok := idx.output[corpus][op.ID]
			if !ok || outRec.AX == nil {
				continue
			}
			lengthRatio := 0.0
			if base.ChosenCount > 0 {
				lengthRatio = float64(outRec.ChosenCount) / float64(base.ChosenCount)
			}
			retained := 0.0
			if base.AX.AX3StreamEntropy > 0 {
				retained = outRec.AX.AX3StreamEntropy / base.AX.AX3StreamEntropy
			}
			red.row("EXTRACTION", corpus, op.ID, fstr(lengthRatio), fstr(ttrBefore), fstr(ttrOf(outRec)), fstr(base.AX.AX3StreamEntropy), fstr(outRec.AX.AX3StreamEntropy), fstr(retained))
		}
	}
	var scales []string
	for s := range sidx.byScaleVariant {
		scales = append(scales, s)
	}
	sort.Strings(scales)
	for _, scale := range scales {
		byVariant := sidx.byScaleVariant[scale]
		expRecs, abbrRecs := byVariant["EXPANDED"], byVariant["ABBREVIATED"]
		if len(expRecs) == 0 || len(abbrRecs) == 0 || expRecs[0].AX == nil || abbrRecs[0].AX == nil {
			continue
		}
		exp, abbr := expRecs[0], abbrRecs[0]
		lengthRatio := 0.0
		if exp.ChosenCount > 0 {
			lengthRatio = float64(abbr.ChosenCount) / float64(exp.ChosenCount)
		}
		retained := 0.0
		if exp.AX.AX3StreamEntropy > 0 {
			retained = abbr.AX.AX3StreamEntropy / exp.AX.AX3StreamEntropy
		}
		red.row("SHORTHAND", scale, "ABBREVIATED_vs_EXPANDED", fstr(lengthRatio), fstr(ttrOf(exp)), fstr(ttrOf(abbr)), fstr(exp.AX.AX3StreamEntropy), fstr(abbr.AX.AX3StreamEntropy), fstr(retained))
	}
	if err := red.write(out, "INFORMATION_REDUCTION_COMPARISON.tsv"); err != nil {
		return err
	}

	return writeInputDependence(out, idx)
}

func ttrOf(r Rec) float64 {
	if r.AX == nil {
		return 0
	}
	return r.AX.AX4TypeTokenRatio
}

// writeInputDependence classifies, per operator and per CORE metric,
// whether ΔF2 variance is dominated by which carrier (INPUT_DOMINATED),
// by the operator (MECHANISM_DOMINATED), or comparable (MIXED)
// (task82b.txt sec.57), from the same before/after values already in
// EXTRACTION_F2_TRAJECTORIES.tsv.
func writeInputDependence(out string, idx extractionIndex) error {
	w := newTSV("metric_id", "classification", "variance_across_carriers_same_operator_mean", "variance_across_operators_same_carrier_mean")
	ops := task82b.Registry()
	for _, metricID := range task82b.CoreMetricIDs {
		// variance across carriers, averaged over operators
		var perOpVar []float64
		for _, op := range ops {
			var deltas []float64
			for _, corpus := range task82b.CarrierOrder {
				base, ok1 := idx.baseline[corpus]
				outRec, ok2 := idx.output[corpus][op.ID]
				if !ok1 || !ok2 {
					continue
				}
				before, bOK := base.metric(metricID)
				after, aOK := outRec.metric(metricID)
				if bOK && aOK {
					deltas = append(deltas, after-before)
				}
			}
			if len(deltas) >= 2 {
				perOpVar = append(perOpVar, varianceOf(deltas))
			}
		}
		// variance across operators, averaged over carriers
		var perCorpusVar []float64
		for _, corpus := range task82b.CarrierOrder {
			base, ok1 := idx.baseline[corpus]
			if !ok1 {
				continue
			}
			var deltas []float64
			for _, op := range ops {
				outRec, ok2 := idx.output[corpus][op.ID]
				if !ok2 {
					continue
				}
				before, bOK := base.metric(metricID)
				after, aOK := outRec.metric(metricID)
				if bOK && aOK {
					deltas = append(deltas, after-before)
				}
			}
			if len(deltas) >= 2 {
				perCorpusVar = append(perCorpusVar, varianceOf(deltas))
			}
		}
		vCarrier := meanOf(perOpVar)
		vOperator := meanOf(perCorpusVar)
		classification := "MIXED"
		switch {
		case vCarrier > 2*vOperator:
			classification = "INPUT_DOMINATED"
		case vOperator > 2*vCarrier:
			classification = "MECHANISM_DOMINATED"
		}
		w.row(metricID, classification, fstr(vCarrier), fstr(vOperator))
	}
	return w.write(out, "INPUT_DEPENDENCE_COMPARISON.tsv")
}

func meanOf(xs []float64) float64 {
	if len(xs) == 0 {
		return 0
	}
	s := 0.0
	for _, x := range xs {
		s += x
	}
	return s / float64(len(xs))
}

func varianceOf(xs []float64) float64 {
	m := meanOf(xs)
	s := 0.0
	for _, x := range xs {
		d := x - m
		s += d * d
	}
	if len(xs) < 2 {
		return 0
	}
	return s / float64(len(xs)-1)
}
