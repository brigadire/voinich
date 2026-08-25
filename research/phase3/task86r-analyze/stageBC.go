package main

// DevFitResult is one (transcription, candidate) DEVELOPMENT fit
// (Stage B): implementation/diagnostic only, never used as confirmatory
// evidence.
type DevFitResult struct {
	Transcription string
	ModelClass    string
	CandidateID   string
	Failed        bool
	FailureReason string
	DevPM2        float64
	Complexity    ComplexityBreakdown
}

func runStageB(dev []TokenOccurrence, transcription string, candidates []Candidate) []DevFitResult {
	bitsReal := bitsPerRealParameter(len(dev))
	out := make([]DevFitResult, len(candidates))
	parallelFor(len(candidates), func(i int) {
		cand := candidates[i]
		model := FitCandidate(dev, cand, bitsReal)
		r := DevFitResult{Transcription: transcription, ModelClass: cand.ModelClass, CandidateID: cand.CandidateID}
		if failed, why := model.TrainingFailed(); failed {
			r.Failed, r.FailureReason = true, why
			out[i] = r
			return
		}
		r.DevPM2 = ComputePM2Only(model, dev)
		r.Complexity = model.Complexity()
		out[i] = r
	})
	return out
}

// StageCSelection is one transcription's class-wise VALIDATION-selected
// candidate, plus the B1/B2 baselines used for that transcription's own
// confirmatory gating.
type StageCSelection struct {
	Transcription string
	ByClass       map[string]SelectionResult
	B1, B2        SelectionResult
	B2Applicable  bool
}

func runStageC(dev, val []TokenOccurrence, transcription string, candidates []Candidate, bitsReal float64) StageCSelection {
	out := StageCSelection{Transcription: transcription, ByClass: map[string]SelectionResult{}}
	classes := []string{"M0", "M1", "M2", "M3", "M4", "M5"}
	results := make([]SelectionResult, len(classes))
	parallelFor(len(classes), func(i int) {
		results[i] = SelectByValidation(candidates, classes[i], dev, val, bitsReal, nil)
	})
	for i, c := range classes {
		out.ByClass[c] = results[i]
	}
	out.B1 = out.ByClass["M0"]
	out.B2 = SelectByValidation(candidates, "M1", dev, val, bitsReal, func(c Candidate) bool { return c.Int("order", 0) == 2 })
	out.B2Applicable = !out.B2.TrainingFailed
	return out
}
