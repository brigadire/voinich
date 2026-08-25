package main

import "math"

// StabilityClass is the frozen G1_TRANSCRIPTION_STABILITY_CONTRACT.md
// classification.
type StabilityClass string

const (
	TranscriptionStable      StabilityClass = "TRANSCRIPTION_STABLE"
	DirectionStable          StabilityClass = "DIRECTION_STABLE"
	TranscriptionSensitive   StabilityClass = "TRANSCRIPTION_SENSITIVE"
	StabilityInconclusive    StabilityClass = "INCONCLUSIVE"
)

func sign(x float64) float64 {
	if x > 0 {
		return 1
	}
	if x < 0 {
		return -1
	}
	return 0
}

// ClassifyStability applies G1_TRANSCRIPTION_STABILITY_CONTRACT.md to a
// pair of effect values (one per transcription), relative to whatever
// baseline each already encodes (signed improvement, or
// threshold-minus-distance for a structural metric).
func ClassifyStability(eZL3b, eIT2a float64) StabilityClass {
	if !isFinite(eZL3b) || !isFinite(eIT2a) {
		return TranscriptionSensitive
	}
	sZ, sI := sign(eZL3b), sign(eIT2a)
	if (sZ == 0) != (sI == 0) {
		return TranscriptionSensitive
	}
	if sZ != sI {
		return TranscriptionSensitive
	}
	denom := math.Max(math.Abs(eZL3b), math.Max(math.Abs(eIT2a), 1e-12))
	d := math.Abs(eZL3b-eIT2a) / denom
	if sZ == 0 && sI == 0 {
		// Exact zero-zero agreement: direction agrees (both zero); use
		// discrepancy alone (0) -> stable.
		return TranscriptionStable
	}
	if d <= 0.20 {
		return TranscriptionStable
	}
	return DirectionStable
}

func atLeastDirectionStable(c StabilityClass) bool {
	return c == TranscriptionStable || c == DirectionStable
}
