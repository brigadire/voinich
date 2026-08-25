package main

// TranscriptionBaselines holds B1/B2's own HELDOUT metrics, computed once
// per transcription and reused by every class's gate evaluation.
type TranscriptionBaselines struct {
	B1Held         PredictiveMetrics
	B1PM6          float64
	B1PM6OK        bool
	B2Held         PredictiveMetrics
	B2PM6          float64
	B2PM6OK        bool
	B2Applicable   bool
}

func computeTranscriptionBaselines(namespace, transcription string, sel StageCSelection, dev, val, heldout []TokenOccurrence) TranscriptionBaselines {
	out := TranscriptionBaselines{B2Applicable: sel.B2Applicable}
	devVocab := devVocabulary(dev)
	if !sel.B1.TrainingFailed {
		out.B1Held = ComputePM1PM2PM3PM5(sel.B1.Model, heldout)
		out.B1Held.PM4 = ComputePM4(sel.B1.Model, heldout, devVocab)
		if pairs, ok := BuildNegativePairs(namespace, transcription, sel.B1.Candidate.CandidateID, "M0", dev, val, heldout, nil); ok {
			auc, valid, _, _ := ComputePM6(sel.B1.Model, pairs, namespace, transcription, sel.B1.Candidate.CandidateID, "M0")
			out.B1PM6, out.B1PM6OK = auc, valid
		}
	}
	if out.B2Applicable {
		out.B2Held = ComputePM1PM2PM3PM5(sel.B2.Model, heldout)
		out.B2Held.PM4 = ComputePM4(sel.B2.Model, heldout, devVocab)
		if pairs, ok := BuildNegativePairs(namespace, transcription, sel.B2.Candidate.CandidateID, "M1", dev, val, heldout, nil); ok {
			auc, valid, _, _ := ComputePM6(sel.B2.Model, pairs, namespace, transcription, sel.B2.Candidate.CandidateID, "M1")
			out.B2PM6, out.B2PM6OK = auc, valid
		}
	}
	return out
}

// StageDResult is one class's confirmatory HELDOUT evaluation for one
// transcription.
type StageDResult struct {
	Transcription, ModelClass, CandidateID string
	Candidate                              Candidate
	Model                                  FittedModel
	Failed                                 bool
	FailureClasses                         []string

	DevPM2   float64
	HeldPM   PredictiveMetrics
	PM6      float64
	PM6Valid bool

	MetricGates    map[string]MetricGateResult
	PredictivePass bool

	Complexity       ComplexityBreakdown
	Memorization     MemorizationResult
	ComplexityGrowth ComplexityGrowthResult

	HeldF2      map[string]float64
	HeldF2Valid bool
	Generation  map[float64]GenerationResult
}

var scales = []float64{0.5, 1.0, 2.0}

func runStageDClass(namespace, transcription, class string, sel StageCSelection, base TranscriptionBaselines, dev, val, heldout []TokenOccurrence, bitsReal float64, idx *ThresholdIndex, alias *GlyphAlias, rawToGlyphs func(string) []string, alphabetSize int, heldF2 map[string]float64, heldF2Valid bool, workDir string) *StageDResult {
	sr := sel.ByClass[class]
	res := &StageDResult{Transcription: transcription, ModelClass: class, CandidateID: sr.Candidate.CandidateID, Candidate: sr.Candidate}
	if sr.TrainingFailed {
		res.Failed = true
		res.FailureClasses = []string{"TRAINING_FAILED"}
		return res
	}
	res.Model = sr.Model
	devVocab := devVocabulary(dev)
	res.DevPM2 = ComputePM2Only(sr.Model, dev)
	res.HeldPM = ComputePM1PM2PM3PM5(sr.Model, heldout)
	res.HeldPM.PM4 = ComputePM4(sr.Model, heldout, devVocab)

	var componentCountOf func([]string) (int, bool)
	if m5, ok := sr.Model.(*M5Model); ok {
		componentCountOf = m5.ComponentCount
	}
	if pairs, ok := BuildNegativePairs(namespace, transcription, sr.Candidate.CandidateID, class, dev, val, heldout, componentCountOf); ok {
		auc, valid, _, _ := ComputePM6(sr.Model, pairs, namespace, transcription, sr.Candidate.CandidateID, class)
		res.PM6, res.PM6Valid = auc, valid
	} else {
		res.FailureClasses = append(res.FailureClasses, "NEGATIVE_EXHAUSTED")
	}

	res.MetricGates = map[string]MetricGateResult{}
	allPass := true
	// B2 (M1 order=2, GLYPH-scoring-unit) is only a comparable baseline
	// for other GLYPH-unit model classes; M0 is TOKEN-unit and is
	// compared only against B1 (which M0 itself instantiates).
	b2Applicable := base.B2Applicable && class != "M0"
	for _, m := range predictiveMetricNames {
		cv, _ := metricValue(res.HeldPM, m, res.PM6, res.PM6Valid)
		b1v, _ := metricValue(base.B1Held, m, base.B1PM6, base.B1PM6OK)
		b2v, _ := metricValue(base.B2Held, m, base.B2PM6, base.B2PM6OK)
		g := EvalPredictiveMetric(m, class, sr.Candidate.CandidateID, cv, b1v, b2Applicable, b2v, idx)
		res.MetricGates[m] = g
		if !g.Pass {
			allPass = false
		}
	}
	res.PredictivePass = allPass

	res.Complexity = sr.Model.Complexity()
	res.Memorization = EvalMemorization(class, sr.Candidate.CandidateID, res.DevPM2, res.HeldPM.PM2, res.Complexity.Total(), len(dev), alphabetSize, idx)
	if res.Memorization.Dominated {
		res.FailureClasses = append(res.FailureClasses, "MEMORIZATION_DOMINATED")
	}

	res.ComplexityGrowth = computeComplexityGrowth(namespace, transcription, sr.Candidate.CandidateID, class, dev, sr.Candidate)
	if res.ComplexityGrowth.Unbounded {
		res.FailureClasses = append(res.FailureClasses, "COMPLEXITY_UNBOUNDED")
	}

	res.HeldF2, res.HeldF2Valid = heldF2, heldF2Valid
	res.Generation = map[float64]GenerationResult{}
	for _, sc := range scales {
		g := runGeneration(namespace, transcription, class, sr.Candidate.CandidateID, sr.Model, sc, len(heldout), rawToGlyphs, alias, workDir)
		res.Generation[sc] = g
		if !g.Converged {
			addFailureOnce(res, "NUMERICALLY_UNSTABLE")
		}
		if g.ExcessiveCV {
			addFailureOnce(res, "NUMERICALLY_UNSTABLE")
		}
	}
	res.Failed = len(res.FailureClasses) > 0
	return res
}

func addFailureOnce(res *StageDResult, class string) {
	for _, c := range res.FailureClasses {
		if c == class {
			return
		}
	}
	res.FailureClasses = append(res.FailureClasses, class)
}
