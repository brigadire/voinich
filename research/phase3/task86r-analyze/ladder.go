package main

import "math"

var modelRank = map[string]int{"M0": 0, "M1": 1, "M2": 2, "M3": 3, "M4": 4, "M5": 5}

// MinimalityCandidate is one class's confirmatory candidate, evaluated
// per-transcription.
type MinimalityCandidate struct {
	ModelClass         string
	CandidateID        string
	Complexity         float64 // this transcription's own Complexity(G); NaN if UNBOUNDED/unavailable
	PredictiveAdequate bool    // cross-transcription-conjoined PredictiveAdequacy(G)=PASS
	StructuralAdequate bool    // cross-transcription-conjoined StructuralAdequacy(G)=PASS
	AnyFailure         bool    // any registered frozen failure class
}

const minimalityAbsBits = 1.0
const minimalityRelative = 1e-6

// SelectMinimalPerTranscription implements G1_ADEQUACY_GATES.md
// minimality: argmin Complexity among candidates passing both adequacy
// gates with no failure; deterministic tie-break by (model rank,
// candidate id) within the equivalence set.
func SelectMinimalPerTranscription(cands []MinimalityCandidate) (winner *MinimalityCandidate, equivalenceSet []MinimalityCandidate) {
	var eligible []MinimalityCandidate
	for _, c := range cands {
		if c.PredictiveAdequate && c.StructuralAdequate && !c.AnyFailure && isFinite(c.Complexity) {
			eligible = append(eligible, c)
		}
	}
	if len(eligible) == 0 {
		return nil, nil
	}
	minC := math.Inf(1)
	for _, c := range eligible {
		if c.Complexity < minC {
			minC = c.Complexity
		}
	}
	tol := math.Max(minimalityAbsBits, minimalityRelative*minC)
	for _, c := range eligible {
		if math.Abs(c.Complexity-minC) <= tol {
			equivalenceSet = append(equivalenceSet, c)
		}
	}
	best := equivalenceSet[0]
	for _, c := range equivalenceSet[1:] {
		if modelRank[c.ModelClass] < modelRank[best.ModelClass] ||
			(modelRank[c.ModelClass] == modelRank[best.ModelClass] && c.CandidateID < best.CandidateID) {
			best = c
		}
	}
	return &best, equivalenceSet
}

// LadderEdge is one M0..M5 adjacent (or finite-state-parent) comparison.
type LadderEdge struct {
	Parent, Child string
	Gain          string // SUPPORTED / NOT_SUPPORTED / INCONCLUSIVE
}

// descriptionLength is Task85 Complexity + HELDOUT PM1 (G1_MODEL_LADDER_CONTRACT.md).
func descriptionLength(complexity, pm1 float64) float64 { return complexity + pm1 }

// RepresentationalGain evaluates one child-vs-parent edge across both
// transcriptions.
func RepresentationalGain(childAdequate map[string]bool, childDL, parentDL map[string]float64, childClass, parentClass string, idx *ThresholdIndex, childCandID string, structuralRegression map[string]bool) string {
	transcriptions := []string{"ZL3b", "IT2a"}
	var effects []float64
	for _, tr := range transcriptions {
		if !childAdequate[tr] {
			return "NOT_SUPPORTED"
		}
		if structuralRegression[tr] {
			return "NOT_SUPPORTED"
		}
	}
	for _, tr := range transcriptions {
		cdl, ok1 := childDL[tr]
		pdl, ok2 := parentDL[tr]
		if !ok1 || !ok2 || !isFinite(cdl) || !isFinite(pdl) {
			return "INCONCLUSIVE"
		}
		effects = append(effects, pdl-cdl) // positive = child lower/better
	}
	t, ok := idx.Get("predictive_gain_vs_b1", "PM1", childClass, childCandID)
	threshold := 0.0
	if ok {
		threshold = t.Threshold
	}
	for _, e := range effects {
		if e <= threshold {
			return "NOT_SUPPORTED"
		}
	}
	cls := ClassifyStability(effects[0], effects[1])
	if !atLeastDirectionStable(cls) {
		return "NOT_SUPPORTED"
	}
	return "SUPPORTED"
}

// TokenFormationDepth applies the frozen class->depth mapping.
func TokenFormationDepth(class string) string {
	switch class {
	case "M0":
		return "FREQUENCY_ONLY"
	case "M1":
		return "LOCAL_MARKOV"
	case "M2":
		return "VARIABLE_MEMORY"
	case "M3", "M4":
		return "FINITE_STATE"
	case "M5":
		return "EXPLICIT_RULE_SYSTEM"
	}
	return "NOT_IDENTIFIABLE"
}
