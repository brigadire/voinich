package main

// FitCandidate dispatches to the model class's own Fit routine. bitsReal is
// the transcription/population's own DEVELOPMENT-count-derived
// ceil(log2(N_dev))/2 constant (GRAMMAR_COMPLEXITY_CONTRACT.md).
func FitCandidate(dev []TokenOccurrence, cand Candidate, bitsReal float64) FittedModel {
	switch cand.ModelClass {
	case "M0":
		return FitM0(dev, cand, bitsReal)
	case "M1":
		return FitM1(dev, cand, bitsReal)
	case "M2":
		return FitM2(dev, cand, bitsReal)
	case "M3":
		return FitM3(dev, cand)
	case "M4":
		return FitM4(dev, cand, bitsReal)
	case "M5":
		return FitM5(dev, cand, bitsReal)
	}
	panic("unknown model class " + cand.ModelClass)
}
