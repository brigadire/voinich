package main

import "math"

// Decision is one categorical scoring/generation decision point, shared by
// every model class. Outcomes must already be sorted per the model's frozen
// tie-break rule (code-point order, or a documented extension of it).
// Probs is parallel to Outcomes and should sum to ~1. ObservedIndex is the
// index of the actually-observed outcome, or -1 if the observed outcome is
// not a member of Outcomes (probability 0).
type Decision struct {
	Outcomes      []string
	Probs         []float64
	ObservedIndex int
}

// ScoreEvent is PM5's per-decision-point (confidence, correctness) pair
// plus the decision's own contribution to PM1 (-log2 observed prob).
type ScoreEvent struct {
	Confidence  float64
	Correct     bool
	NegLog2Prob float64
}

func (d Decision) Event() ScoreEvent {
	maxP := -1.0
	predicted := -1
	for i, p := range d.Probs {
		if p > maxP {
			maxP = p
			predicted = i
		}
	}
	correct := predicted >= 0 && predicted == d.ObservedIndex
	var neg float64
	if d.ObservedIndex < 0 || d.ObservedIndex >= len(d.Probs) || d.Probs[d.ObservedIndex] <= 0 {
		neg = math.Inf(1)
	} else {
		neg = -math.Log2(d.Probs[d.ObservedIndex])
	}
	return ScoreEvent{Confidence: maxP, Correct: correct, NegLog2Prob: neg}
}

// Cumulative returns the half-open cumulative-interval boundaries over
// Probs, in Outcomes order (already sorted), renormalized to sum to
// exactly 1 by construction of the final cumulative value.
func Cumulative(probs []float64) []float64 {
	cum := make([]float64, len(probs))
	total := 0.0
	for i, p := range probs {
		total += p
		cum[i] = total
	}
	if len(cum) > 0 {
		cum[len(cum)-1] = 1.0
	}
	return cum
}

const (
	eosSymbol       = "<EOS>"
	unkGlyphSymbol  = "<UNK_GLYPH>"
	unkTokenLiteral = "<UNK>"
	maxTokenGlyphs  = 64
)
