// Package tokentransition contains Task63's canonical adjacent-form analysis.
// Edit semantics are delegated to the authoritative Task60 implementation.
package tokentransition

import "zcore.dev/voinich/internal/tokenrepetition"

type Pair struct {
	A, B          []string
	Distance      int
	Operation     string
	Position      int
	PositionClass string
}

func Analyze(a, b []string) Pair {
	d := tokenrepetition.LevenshteinGlyphs(a, b)
	p := Pair{A: a, B: b, Distance: d, Operation: "", Position: -1}
	if d == 1 {
		op, pos, _, _, ok := tokenrepetition.ClassifyDistanceOne(a, b)
		if ok {
			p.Operation = op
			p.Position = pos
			n := len(a)
			if len(b) > n {
				n = len(b)
			}
			p.PositionClass = tokenrepetition.PositionClass(pos, n)
		}
	}
	return p
}
func IsNear(p Pair) bool             { return p.Distance <= 1 }
func EditDistance(a, b []string) int { return tokenrepetition.LevenshteinGlyphs(a, b) }
