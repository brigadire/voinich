package main

import (
	"zcore.dev/voinich/internal/task82b"
)

// cell is one (metric,operator,carrier) trajectory observation used by
// the stability classifier.
type cell struct {
	delta     float64
	available bool
	effect    float64
	hasEffect bool
}

type extractionIndex struct {
	baseline   map[string]Rec            // corpus -> carrier_baseline record
	output     map[string]map[string]Rec // corpus -> operator -> operator_output record
	nullRandom map[string]map[string][]Rec
	nullStrat  map[string]map[string][]Rec
	nullPeriod map[string]map[string][]Rec
}

func indexExtraction(recs []Rec) extractionIndex {
	idx := extractionIndex{
		baseline:   map[string]Rec{},
		output:     map[string]map[string]Rec{},
		nullRandom: map[string]map[string][]Rec{},
		nullStrat:  map[string]map[string][]Rec{},
		nullPeriod: map[string]map[string][]Rec{},
	}
	ensure := func(m map[string]map[string][]Rec, corpus string) {
		if m[corpus] == nil {
			m[corpus] = map[string][]Rec{}
		}
	}
	for _, r := range recs {
		switch r.Kind {
		case "carrier_baseline":
			idx.baseline[r.CorpusID] = r
		case "operator_output":
			if idx.output[r.CorpusID] == nil {
				idx.output[r.CorpusID] = map[string]Rec{}
			}
			idx.output[r.CorpusID][r.OperatorID] = r
		case "operator_null_random":
			ensure(idx.nullRandom, r.CorpusID)
			idx.nullRandom[r.CorpusID][r.OperatorID] = append(idx.nullRandom[r.CorpusID][r.OperatorID], r)
		case "operator_null_stratified":
			ensure(idx.nullStrat, r.CorpusID)
			idx.nullStrat[r.CorpusID][r.OperatorID] = append(idx.nullStrat[r.CorpusID][r.OperatorID], r)
		case "operator_null_periodic":
			ensure(idx.nullPeriod, r.CorpusID)
			idx.nullPeriod[r.CorpusID][r.OperatorID] = append(idx.nullPeriod[r.CorpusID][r.OperatorID], r)
		}
	}
	return idx
}

func writeExtractionOutputs(out string, recs []Rec) error {
	idx := indexExtraction(recs)
	ops := task82b.Registry()

	// EXTRACTION_OPERATOR_REGISTRY.tsv
	reg := newTSV("operator_id", "structural_class", "extraction_class", "provenance", "null_class", "param", "window_size")
	for _, op := range ops {
		reg.row(op.ID, op.StructuralClass, op.ExtractionClass, op.Provenance, op.NullClass, istr(op.Param), istr(op.WindowSize))
	}
	if err := reg.write(out, "EXTRACTION_OPERATOR_REGISTRY.tsv"); err != nil {
		return err
	}

	beforeAfter := newTSV("corpus", "operator_id", "metric_id", "classification", "f2_before", "before_available", "f2_after", "after_available", "chosen_count", "candidate_pool", "degenerate")
	trajectories := newTSV("corpus", "operator_id", "metric_id", "classification", "delta", "both_available")
	nullCmp := newTSV("corpus", "operator_id", "null_kind", "metric_id", "observed_delta", "null_mean_delta", "null_sd_delta", "effect_size", "p_value", "n_replicates")

	// stability[metric][operator][corpus] = cell
	stability := map[string]map[string]map[string]cell{}

	for _, corpus := range task82b.CarrierOrder {
		base, ok := idx.baseline[corpus]
		if !ok {
			continue
		}
		for _, op := range ops {
			outRec, ok := idx.output[corpus][op.ID]
			if !ok {
				continue
			}
			for _, metricID := range task82b.AllMetricIDs() {
				before, beforeOK := base.metric(metricID)
				after, afterOK := outRec.metric(metricID)
				classification := metricClass(metricID)
				beforeAfter.row(corpus, op.ID, metricID, classification, fOr(before, beforeOK), bstr(beforeOK), fOr(after, afterOK), bstr(afterOK), istr(outRec.ChosenCount), istr(outRec.CandidatePool), bstr(outRec.Degenerate))
				bothAvail := beforeOK && afterOK
				delta := 0.0
				if bothAvail {
					delta = after - before
				}
				trajectories.row(corpus, op.ID, metricID, classification, fOr(delta, bothAvail), bstr(bothAvail))

				if stability[metricID] == nil {
					stability[metricID] = map[string]map[string]cell{}
				}
				if stability[metricID][op.ID] == nil {
					stability[metricID][op.ID] = map[string]cell{}
				}
				c := cell{delta: delta, available: bothAvail}

				// null comparisons
				randReps := valuesFor(idx.nullRandom[corpus][op.ID], metricID)
				if len(randReps) > 0 && bothAvail {
					cmp := task82b.CompareToNull(after, randReps)
					nullCmp.row(corpus, op.ID, "RANDOM_SUBSEQUENCE_MATCHED", metricID, fstr(delta), fstr(cmp.NullMean-before), fstr(cmp.NullSD), fstr(cmp.EffectSize), fstr(cmp.PValue), istr(cmp.N))
					c.effect = cmp.EffectSize
					c.hasEffect = true
				}
				if strat := valuesFor(idx.nullStrat[corpus][op.ID], metricID); len(strat) > 0 && bothAvail {
					cmp := task82b.CompareToNull(after, strat)
					nullCmp.row(corpus, op.ID, "POSITION_STRATIFIED_RANDOM", metricID, fstr(delta), fstr(cmp.NullMean-before), fstr(cmp.NullSD), fstr(cmp.EffectSize), fstr(cmp.PValue), istr(cmp.N))
				}
				if periodic := valuesFor(idx.nullPeriod[corpus][op.ID], metricID); len(periodic) > 0 && bothAvail {
					cmp := task82b.CompareToNull(after, periodic)
					nullCmp.row(corpus, op.ID, "PERIODIC_PHASE", metricID, fstr(delta), fstr(cmp.NullMean-before), fstr(cmp.NullSD), fstr(cmp.EffectSize), fstr(cmp.PValue), istr(cmp.N))
				}
				stability[metricID][op.ID][corpus] = c
			}
		}
	}
	if err := beforeAfter.write(out, "EXTRACTION_F2_BEFORE_AFTER.tsv"); err != nil {
		return err
	}
	if err := trajectories.write(out, "EXTRACTION_F2_TRAJECTORIES.tsv"); err != nil {
		return err
	}
	if err := nullCmp.write(out, "EXTRACTION_NULL_COMPARISON.tsv"); err != nil {
		return err
	}

	return writeExtractionStability(out, ops, stability)
}

func valuesFor(recs []Rec, metricID string) []float64 {
	var out []float64
	for _, r := range recs {
		if v, ok := r.metric(metricID); ok {
			out = append(out, v)
		}
	}
	return out
}

func metricClass(id string) string {
	for _, c := range task82b.CoreMetricIDs {
		if c == id {
			return "CORE"
		}
	}
	return "SUPPORTING"
}

// writeExtractionStability classifies each (operator,metric) pair
// (task82b.txt sec.39): cross-carrier stability first, then whether a
// stable effect is shared by most other operators (EXTRACTION_GENERAL)
// or unique to this one (OPERATOR_SPECIFIC).
func writeExtractionStability(out string, ops []task82b.Operator, stability map[string]map[string]map[string]cell) error {
	w := newTSV("metric_id", "operator_id", "extraction_class", "n_carriers_significant", "n_carriers_consistent_sign", "cross_carrier_stable", "n_operators_sharing_direction", "classification")
	for _, metricID := range task82b.AllMetricIDs() {
		perOp := stability[metricID]
		// determine, for each operator, whether it is cross-carrier stable and its dominant sign
		type opStat struct {
			stable bool
			sign   int
		}
		opStats := map[string]opStat{}
		for _, op := range ops {
			cells := perOp[op.ID]
			nSig, nPos, nNeg := 0, 0, 0
			for _, c := range cells {
				if !c.available {
					continue
				}
				if c.hasEffect && absF(c.effect) >= 2 {
					nSig++
				}
				if c.delta > 0 {
					nPos++
				} else if c.delta < 0 {
					nNeg++
				}
			}
			consistentSign := nPos == 0 || nNeg == 0
			stable := nSig >= 2 && consistentSign
			sign := 0
			if nPos > nNeg {
				sign = 1
			} else if nNeg > nPos {
				sign = -1
			}
			opStats[op.ID] = opStat{stable: stable, sign: sign}
		}
		for _, op := range ops {
			st := opStats[op.ID]
			cells := perOp[op.ID]
			nSig, nPos, nNeg := 0, 0, 0
			for _, c := range cells {
				if !c.available {
					continue
				}
				if c.hasEffect && absF(c.effect) >= 2 {
					nSig++
				}
				if c.delta > 0 {
					nPos++
				} else if c.delta < 0 {
					nNeg++
				}
			}
			consistentSign := nPos == 0 || nNeg == 0
			sharing := 0
			for _, other := range ops {
				os := opStats[other.ID]
				if os.stable && os.sign == st.sign && st.sign != 0 {
					sharing++
				}
			}
			classification := "NOT_STABLE"
			switch {
			case st.stable && sharing >= len(ops)/2:
				classification = "EXTRACTION_GENERAL"
			case st.stable:
				classification = "OPERATOR_SPECIFIC"
			case nSig >= 1 && !consistentSign:
				classification = "PLAINTEXT_DRIVEN"
			}
			w.row(metricID, op.ID, extractionClassOf(ops, op.ID), istr(nSig), istr(boolToInt(consistentSign)), bstr(st.stable), istr(sharing), classification)
		}
	}
	return w.write(out, "EXTRACTION_STABILITY.tsv")
}

func extractionClassOf(ops []task82b.Operator, id string) string {
	for _, op := range ops {
		if op.ID == id {
			return op.ExtractionClass
		}
	}
	return ""
}

func absF(v float64) float64 {
	if v < 0 {
		return -v
	}
	return v
}

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}
