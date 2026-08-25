package main

import "math"

// SelectionResult is one class-wise VALIDATION selection outcome.
type SelectionResult struct {
	Candidate      Candidate
	Model          FittedModel
	ValidationPM2  float64
	TrainingFailed bool
	FailureReason  string
}

// SelectByValidation fits every candidate in class (subject to filter) on
// dev and scores VALIDATION-partition PM2, returning the minimizer. This
// is the frozen "class-wise selection" rule (task86r.txt section 18):
// argmin VALIDATION PM2 -- the primary predictive criterion already
// established by PM1/PM2's PRIMARY evidence tier -- with ties broken by
// the candidate grid's own deterministic enumeration order (the first
// strictly-better candidate scanned wins, since ties never replace it).
func SelectByValidation(all []Candidate, class string, dev, val []TokenOccurrence, bitsReal float64, filter func(Candidate) bool) SelectionResult {
	best := SelectionResult{ValidationPM2: math.Inf(1)}
	haveBest := false
	for _, cand := range all {
		if cand.ModelClass != class {
			continue
		}
		if filter != nil && !filter(cand) {
			continue
		}
		model := FitCandidate(dev, cand, bitsReal)
		if failed, why := model.TrainingFailed(); failed {
			if !haveBest {
				best = SelectionResult{Candidate: cand, Model: model, TrainingFailed: true, FailureReason: why, ValidationPM2: math.Inf(1)}
			}
			continue
		}
		pm := ComputePM1PM2PM3PM5(model, val)
		if !haveBest || pm.PM2 < best.ValidationPM2 {
			best = SelectionResult{Candidate: cand, Model: model, ValidationPM2: pm.PM2}
			haveBest = true
		}
	}
	return best
}
