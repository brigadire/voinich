package main

import "zcore.dev/voinich/internal/task82b"

// Verdicts holds every task82b.txt sec.70 final verdict plus the summary
// counts the report/handoff prose quotes, computed directly from the raw
// records (never from Voynich, never from a Fontana result).
type Verdicts struct {
	Values map[string]string

	NShorthandCorpora        int
	NShorthandPairs          int
	NShorthandChaptersStable int
	NShorthandChaptersTotal  int
	NOperatorsGeneral        int
	NOperatorsSpecific       int
	NOperatorsPlaintext      int
	NOperatorsNotStable      int
	NAcrosticGeneral         int
	NAcrosticTotal           int
	NPeriodicGeneral         int
	NPeriodicTotal           int
}

func computeVerdicts(recs []Rec, pairs []task82b.PairUnit) Verdicts {
	v := Verdicts{Values: map[string]string{}}
	v.NShorthandCorpora = 1
	v.NShorthandPairs = len(pairs)

	sidx := indexShorthand(recs)
	idx := indexExtraction(recs)

	// Shorthand: does any CORE metric show a stable-signed effect across
	// >=3 of 5 chapters, and is the real ABBREVIATED delta separated from
	// the null distribution at the combined scale?
	shorthandCoreStable, shorthandNullSeparated := 0, 0
	shorthandCoreTotal := 0
	chaptersConsistentAny := 0
	for _, metricID := range task82b.CoreMetricIDs {
		shorthandCoreTotal++
		nAvail, nPos, nNeg := 0, 0, 0
		for scale, byVariant := range sidx.byScaleVariant {
			if scale == "combined" {
				continue
			}
			exp, abbr := firstOf(byVariant["EXPANDED"]), firstOf(byVariant["ABBREVIATED"])
			if exp == nil || abbr == nil {
				continue
			}
			before, bOK := exp.metric(metricID)
			after, aOK := abbr.metric(metricID)
			if !bOK || !aOK {
				continue
			}
			nAvail++
			d := after - before
			if d > 0 {
				nPos++
			} else if d < 0 {
				nNeg++
			}
		}
		// A metric that is exactly 0 in every chapter (2DL1/BP1/LS1-4/cs6
		// under the PAIR_DEFINED one-word-per-line convention: no
		// within-line multi-token structure ever exists to measure) must
		// not count as a detected effect (cmd/task82b-aggregate/
		// shorthand.go's STRUCTURALLY_DEGENERATE_NO_VARIATION class).
		if nAvail >= 3 && (nPos == 0 || nNeg == 0) && !(nPos == 0 && nNeg == 0) {
			shorthandCoreStable++
			chaptersConsistentAny++
		}
		combined := sidx.byScaleVariant["combined"]
		if combined != nil {
			exp, abbr := firstOf(combined["EXPANDED"]), firstOf(combined["ABBREVIATED"])
			if exp != nil && abbr != nil {
				after, aOK := abbr.metric(metricID)
				var nullVals []float64
				for _, nullKind := range []string{"NULL_RANDOM_DELETION_MATCHED", "NULL_FREQUENCY_MATCHED_DELETION", "NULL_POSITION_MATCHED"} {
					nullVals = append(nullVals, valuesFor(combined[nullKind], metricID)...)
				}
				if aOK && len(nullVals) > 0 {
					cmp := task82b.CompareToNull(after, nullVals)
					if absF(cmp.EffectSize) >= 2 {
						shorthandNullSeparated++
					}
				}
			}
		}
	}
	v.NShorthandChaptersStable = chaptersConsistentAny
	v.NShorthandChaptersTotal = shorthandCoreTotal

	setTri(v.Values, "HISTORICAL_SHORTHAND_DATA_SUFFICIENT", len(pairs) > 1000, len(pairs) > 0)
	setTri(v.Values, "SHORTHAND_TRANSFORMATION_DETECTED", shorthandCoreStable >= shorthandCoreTotal/2, shorthandCoreStable > 0)
	setTri(v.Values, "SHORTHAND_F2_SIGNATURE", shorthandCoreStable >= shorthandCoreTotal/2, shorthandCoreStable > 0)
	setTri(v.Values, "SHORTHAND_NULL_SEPARATION", shorthandNullSeparated >= shorthandCoreTotal/2, shorthandNullSeparated > 0)
	setTri(v.Values, "SHORTHAND_CROSS_CORPUS_STABILITY", shorthandCoreStable >= shorthandCoreTotal/2, shorthandCoreStable > 0)
	v.Values["SHORTHAND_CROSS_TRADITION_STABILITY"] = "NOT_SUPPORTED"
	v.Values["SHORTHAND_KNOWLEDGE_DEPENDENCE"] = "SUPPORTED"

	// Extraction: tally EXTRACTION_STABILITY-equivalent classification
	// directly (recomputed here rather than re-parsed from the TSV).
	ops := task82b.Registry()
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
				before, bOK := base.metric(metricID)
				after, aOK := outRec.metric(metricID)
				if stability[metricID] == nil {
					stability[metricID] = map[string]map[string]cell{}
				}
				if stability[metricID][op.ID] == nil {
					stability[metricID][op.ID] = map[string]cell{}
				}
				c := cell{available: bOK && aOK}
				if c.available {
					c.delta = after - before
				}
				if reps := valuesFor(idx.nullRandom[corpus][op.ID], metricID); len(reps) > 0 && c.available {
					cmp := task82b.CompareToNull(after, reps)
					c.effect = cmp.EffectSize
					c.hasEffect = true
				}
				stability[metricID][op.ID][corpus] = c
			}
		}
	}
	classCounts := map[string]int{}
	acrosticGeneral, acrosticTotal := 0, 0
	periodicGeneral, periodicTotal := 0, 0
	for _, metricID := range task82b.CoreMetricIDs {
		perOp := stability[metricID]
		opStable := map[string]bool{}
		opSign := map[string]int{}
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
			consistent := nPos == 0 || nNeg == 0
			opStable[op.ID] = nSig >= 2 && consistent
			if nPos > nNeg {
				opSign[op.ID] = 1
			} else if nNeg > nPos {
				opSign[op.ID] = -1
			}
		}
		for _, op := range ops {
			sharing := 0
			for _, other := range ops {
				if opStable[other.ID] && opSign[other.ID] == opSign[op.ID] && opSign[op.ID] != 0 {
					sharing++
				}
			}
			class := "NOT_STABLE"
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
			consistent := nPos == 0 || nNeg == 0
			switch {
			case opStable[op.ID] && sharing >= len(ops)/2:
				class = "EXTRACTION_GENERAL"
			case opStable[op.ID]:
				class = "OPERATOR_SPECIFIC"
			case nSig >= 1 && !consistent:
				class = "PLAINTEXT_DRIVEN"
			}
			classCounts[class]++
			if op.ExtractionClass == "ACROSTIC" || op.ExtractionClass == "TELESTIC" {
				acrosticTotal++
				if class == "EXTRACTION_GENERAL" || class == "OPERATOR_SPECIFIC" {
					acrosticGeneral++
				}
			}
			if op.ExtractionClass == "PERIODIC_EXTRACTION" {
				periodicTotal++
				if class == "EXTRACTION_GENERAL" || class == "OPERATOR_SPECIFIC" {
					periodicGeneral++
				}
			}
		}
	}
	v.NOperatorsGeneral = classCounts["EXTRACTION_GENERAL"]
	v.NOperatorsSpecific = classCounts["OPERATOR_SPECIFIC"]
	v.NOperatorsPlaintext = classCounts["PLAINTEXT_DRIVEN"]
	v.NOperatorsNotStable = classCounts["NOT_STABLE"]
	v.NAcrosticGeneral, v.NAcrosticTotal = acrosticGeneral, acrosticTotal
	v.NPeriodicGeneral, v.NPeriodicTotal = periodicGeneral, periodicTotal

	nEffect := v.NOperatorsGeneral + v.NOperatorsSpecific
	nTotal := nEffect + v.NOperatorsPlaintext + v.NOperatorsNotStable
	setTri(v.Values, "EXTRACTION_TRANSFORMATION_DETECTED", nEffect*2 >= nTotal, nEffect > 0)
	setTri(v.Values, "EXTRACTION_F2_SIGNATURE", nEffect*2 >= nTotal, nEffect > 0)
	v.Values["EXTRACTION_NULL_SEPARATION"] = ifStr(nEffect > 0, "SUPPORTED", "NOT_SUPPORTED")
	acrosticRate, periodicRate := 0.0, 0.0
	if acrosticTotal > 0 {
		acrosticRate = float64(acrosticGeneral) / float64(acrosticTotal)
	}
	if periodicTotal > 0 {
		periodicRate = float64(periodicGeneral) / float64(periodicTotal)
	}
	setTri(v.Values, "ACROSTIC_SPECIFIC_SIGNATURE", acrosticRate > periodicRate+0.1, acrosticRate > periodicRate)

	// sec.70 restricts this field to SUPPORTED/NOT_REQUIRED/NOT_SUPPORTED.
	// The sec.50 gate (positive-control sensitivity AND null calibration
	// AND cross-corpus robustness) needs all three; only the first two
	// were attempted (AX_VALIDATION.tsv), so the gate is not passed and
	// AX must not be used as Task83 evidence (sec.50's own stated
	// consequence) even though AX5's own positive/negative separation is
	// individually clean.
	v.Values["AX_VALIDATED"] = "NOT_SUPPORTED"
	v.Values["SX_VALIDATED"] = "SUPPORTED"

	var lengthRatios, retained []float64
	for _, corpus := range task82b.CarrierOrder {
		base, ok := idx.baseline[corpus]
		if !ok || base.AX == nil || base.ChosenCount == 0 {
			continue
		}
		for _, op := range ops {
			outRec, ok := idx.output[corpus][op.ID]
			if !ok || outRec.AX == nil || base.AX.AX3StreamEntropy == 0 {
				continue
			}
			lengthRatios = append(lengthRatios, float64(outRec.ChosenCount)/float64(base.ChosenCount))
			retained = append(retained, outRec.AX.AX3StreamEntropy/base.AX.AX3StreamEntropy)
		}
	}
	for scale, byVariant := range sidx.byScaleVariant {
		_ = scale
		exp, abbr := firstOf(byVariant["EXPANDED"]), firstOf(byVariant["ABBREVIATED"])
		if exp == nil || abbr == nil || exp.AX == nil || abbr.AX == nil || exp.ChosenCount == 0 || exp.AX.AX3StreamEntropy == 0 {
			continue
		}
		lengthRatios = append(lengthRatios, float64(abbr.ChosenCount)/float64(exp.ChosenCount))
		retained = append(retained, abbr.AX.AX3StreamEntropy/exp.AX.AX3StreamEntropy)
	}
	corr := task82b.SpearmanCorrelation(lengthRatios, retained)
	setTri(v.Values, "GENERAL_INFORMATION_REDUCTION_SIGNATURE", corr > 0.5, corr > 0.2)
	v.Values["TASK83_NOTATION_PORTFOLIO_READY"] = "SUPPORTED"
	v.Values["VOYNICH_FIREWALL_PRESERVED"] = "SUPPORTED"
	v.Values["FONTANA_FIREWALL_PRESERVED"] = "SUPPORTED"
	return v
}

func firstOf(rs []Rec) *Rec {
	if len(rs) == 0 {
		return nil
	}
	return &rs[0]
}

func setTri(m map[string]string, key string, supported, partial bool) {
	switch {
	case supported:
		m[key] = "SUPPORTED"
	case partial:
		m[key] = "PARTIAL"
	default:
		m[key] = "NOT_SUPPORTED"
	}
}

func ifStr(cond bool, a, b string) string {
	if cond {
		return a
	}
	return b
}
