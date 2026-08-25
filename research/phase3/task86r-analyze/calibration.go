package main

import "sort"

// CalibJobResult is one (generator, population, candidate) calibration
// pipeline fit.
type CalibJobResult struct {
	Generator    string
	Population   int
	ModelClass   string
	CandidateID  string
	Failed       bool
	FailureWhy   string
	DevPM2       float64
	ValPM        PredictiveMetrics
	HeldPM       PredictiveMetrics
	PM6          float64
	PM6Valid     bool
	F2Generated  map[string]float64
	F2Valid      bool
}

// CalibPopulationBaselines caches one population's B1/B2 baseline fits and
// its HELDOUT-analogue structural fingerprint (computed once, reused by
// every candidate's structural-distance comparison).
type CalibPopulationBaselines struct {
	B1, B2       SelectionResult
	B1Held, B2Held PredictiveMetrics
	B1PM6V, B2PM6V float64
	B1PM6OKv, B2PM6OKv bool
	B2Applicable bool
	HeldF2       map[string]float64
	HeldF2Valid  bool
}

func (b CalibPopulationBaselines) B1BasePM() PredictiveMetrics { return b.B1Held }
func (b CalibPopulationBaselines) B2BasePM() PredictiveMetrics { return b.B2Held }
func (b CalibPopulationBaselines) B1PM6() float64               { return b.B1PM6V }
func (b CalibPopulationBaselines) B2PM6() float64               { return b.B2PM6V }
func (b CalibPopulationBaselines) B1PM6OK() bool                { return b.B1PM6OKv }
func (b CalibPopulationBaselines) B2PM6OK() bool                { return b.B2PM6OKv }

func computeCalibBaselines(namespace string, all []Candidate, dev, val, heldout []TokenOccurrence, bitsReal float64, alias *GlyphAlias, generator string, popIdx int, workDir string) CalibPopulationBaselines {
	b1 := SelectByValidation(all, "M0", dev, val, bitsReal, nil)
	b2 := SelectByValidation(all, "M1", dev, val, bitsReal, func(c Candidate) bool { return c.Int("order", 0) == 2 })
	var pops [][]string
	for _, o := range heldout {
		pops = append(pops, o.Glyphs)
	}
	seed := SeedFields{Namespace: namespace, ModelClass: "BASELINE", CandidateID: "HELD", CorpusID: generator, Transcription: "P" + itoa(popIdx), Partition: "CALIBRATION", Scale: 1.0, Replicate: 0}.Seed()
	f2, ok, _ := StructuralMetrics(alias, pops, int64(seed), workDir)
	out := CalibPopulationBaselines{B1: b1, B2: b2, B2Applicable: !b2.TrainingFailed, HeldF2: f2, HeldF2Valid: ok}
	devVocab := devVocabulary(dev)
	if !b1.TrainingFailed {
		out.B1Held = ComputePM1PM2PM3PM5(b1.Model, heldout)
		out.B1Held.PM4 = ComputePM4(b1.Model, heldout, devVocab)
		if pairs, ok := BuildNegativePairs(namespace, "P"+itoa(popIdx)+"-"+generator, b1.Candidate.CandidateID, "M0", dev, val, heldout, nil); ok {
			auc, valid, _, _ := ComputePM6(b1.Model, pairs, namespace, "P"+itoa(popIdx)+"-"+generator, b1.Candidate.CandidateID, "M0")
			out.B1PM6V, out.B1PM6OKv = auc, valid
		}
	}
	if out.B2Applicable {
		out.B2Held = ComputePM1PM2PM3PM5(b2.Model, heldout)
		out.B2Held.PM4 = ComputePM4(b2.Model, heldout, devVocab)
		if pairs, ok := BuildNegativePairs(namespace, "P"+itoa(popIdx)+"-"+generator, b2.Candidate.CandidateID, "M1", dev, val, heldout, nil); ok {
			auc, valid, _, _ := ComputePM6(b2.Model, pairs, namespace, "P"+itoa(popIdx)+"-"+generator, b2.Candidate.CandidateID, "M1")
			out.B2PM6V, out.B2PM6OKv = auc, valid
		}
	}
	return out
}

func runCalibJob(namespace string, cand Candidate, dev, val, heldout []TokenOccurrence, bitsReal float64, alias *GlyphAlias, generator string, popIdx int, workDir string) CalibJobResult {
	res := CalibJobResult{Generator: generator, Population: popIdx, ModelClass: cand.ModelClass, CandidateID: cand.CandidateID}
	model := FitCandidate(dev, cand, bitsReal)
	if failed, why := model.TrainingFailed(); failed {
		res.Failed = true
		res.FailureWhy = why
		return res
	}
	res.DevPM2 = ComputePM2Only(model, dev)
	res.ValPM = ComputePM1PM2PM3PM5(model, val)
	res.HeldPM = ComputePM1PM2PM3PM5(model, heldout)
	devVocab := devVocabulary(dev)
	res.HeldPM.PM4 = ComputePM4(model, heldout, devVocab)
	res.ValPM.PM4 = ComputePM4(model, val, devVocab)

	var componentCountOf func([]string) (int, bool)
	if m5, ok := model.(*M5Model); ok {
		componentCountOf = m5.ComponentCount
	}
	if pairs, ok := BuildNegativePairs(namespace, "P"+itoa(popIdx)+"-"+generator, cand.CandidateID, cand.ModelClass, dev, val, heldout, componentCountOf); ok {
		auc, valid, _, _ := ComputePM6(model, pairs, namespace, "P"+itoa(popIdx)+"-"+generator, cand.CandidateID, cand.ModelClass)
		res.PM6, res.PM6Valid = auc, valid
	}

	seedFields := SeedFields{Namespace: namespace, ModelClass: cand.ModelClass, CandidateID: cand.CandidateID, CorpusID: generator, Transcription: "P" + itoa(popIdx), Partition: "CALIBRATION", Scale: 1.0, Replicate: 0}
	prng := NewSeededPRNG(seedFields)
	n := len(heldout)
	pops := make([][]string, 0, n)
	for i := 0; i < n; i++ {
		g := model.Generate(prng)
		if g.NonGenerative {
			continue
		}
		pops = append(pops, glyphsForGenerated(model, g, splitGlyphs))
	}
	f2, ok, _ := StructuralMetrics(alias, pops, int64(seedFields.Seed()), workDir)
	res.F2Generated, res.F2Valid = f2, ok
	return res
}

func itoa(i int) string {
	if i == 0 {
		return "0"
	}
	neg := i < 0
	if neg {
		i = -i
	}
	var buf [20]byte
	pos := len(buf)
	for i > 0 {
		pos--
		buf[pos] = byte('0' + i%10)
		i /= 10
	}
	if neg {
		pos--
		buf[pos] = '-'
	}
	return string(buf[pos:])
}

// CalibThreshold is one materialized q0.95 dispersion for a (quantity,
// metric, candidate) triple, maximized across MFC0/1/2.
type CalibThreshold struct {
	Quantity    string // predictive_gain_vs_b1, predictive_gain_vs_b2, overfitting_gap, structural_distance, seed_variation
	Metric      string
	ModelClass  string
	CandidateID string
	Threshold   float64
	PerGenerator map[string]float64
}

func mfcDispersionOrNaN(values []float64) float64 {
	d, ok := mfcDispersion(values)
	if !ok {
		return posInf() // nonfinite input invalidates calibration -> treat as unusable (infinite/blocking) threshold
	}
	return d
}

// materializeThresholds implements G1_CALIBRATION_CONTRACT.md's dispersion
// formula per (quantity, metric, candidate), maximized across generators.
func materializeThresholds(byGenerator map[string]map[string][]float64) []CalibThreshold {
	// key format: "quantity|metric|modelClass|candidateID"
	var keys []string
	seen := map[string]bool{}
	for _, byKey := range byGenerator {
		for k := range byKey {
			if !seen[k] {
				seen[k] = true
				keys = append(keys, k)
			}
		}
	}
	sort.Strings(keys)
	var out []CalibThreshold
	for _, k := range keys {
		q, metric, class, cid := splitCalibKey(k)
		perGen := map[string]float64{}
		maxV := negInf()
		for _, gen := range []string{"MFC0", "MFC1", "MFC2"} {
			vals := byGenerator[gen][k]
			if len(vals) == 0 {
				continue
			}
			d := mfcDispersionOrNaN(vals)
			perGen[gen] = d
			if d > maxV {
				maxV = d
			}
		}
		out = append(out, CalibThreshold{Quantity: q, Metric: metric, ModelClass: class, CandidateID: cid, Threshold: maxV, PerGenerator: perGen})
	}
	return out
}

func calibKey(q, metric, class, cid string) string { return q + "|" + metric + "|" + class + "|" + cid }

func splitCalibKey(k string) (q, metric, class, cid string) {
	parts := splitN(k, '|', 4)
	return parts[0], parts[1], parts[2], parts[3]
}

func splitN(s string, sep byte, n int) []string {
	out := make([]string, 0, n)
	start := 0
	for i := 0; i < len(s) && len(out) < n-1; i++ {
		if s[i] == sep {
			out = append(out, s[start:i])
			start = i + 1
		}
	}
	out = append(out, s[start:])
	return out
}

func negInf() float64 { return -posInf() }
