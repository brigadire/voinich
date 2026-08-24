package fingerprintv2

import (
	"math"
	"strconv"
)

// OrderedGroupMetrics is the generic, metadata-free projection of the
// task79-v1 estimators whose observed statistics require only ordered token
// groups. Group is deliberately neutral: callers must retain their own
// ASSEMBLER_LINE versus PHYSICAL_LINE provenance.
//
// The 2DL implementation intentionally preserves task79-v1's implemented
// three-class boundary projection. Although the registry's prose calls this
// a normalized line-position class, changing the frozen implementation here
// would change the metric and therefore requires a new fingerprint version.
func OrderedGroupMetrics(groups [][]string) map[string]float64 {
	var token, pos, boundary, lengthClass []string
	lengths := make([]float64, 0, len(groups))
	first, last := []float64{}, []float64{}
	repeats, transitions := 0, 0
	for _, group := range groups {
		if len(group) == 0 {
			continue
		}
		lengths = append(lengths, float64(len(group)))
		for i, tok := range group {
			p := "interior"
			if i == 0 {
				p = "first"
			}
			if i == 1 {
				p = "second"
			}
			if len(group) > 1 && i == len(group)-2 {
				p = "penultimate"
			}
			if i == len(group)-1 {
				p = "final"
			}
			b := "interior"
			if i == 0 {
				b = "initial"
			}
			if i == len(group)-1 {
				b = "final"
			}
			token = append(token, tok)
			pos = append(pos, p)
			boundary = append(boundary, b)
			lengthClass = append(lengthClass, strconv.Itoa(min(8, len([]rune(tok)))))
			if i > 0 {
				transitions++
				if group[i-1] == tok {
					repeats++
				}
			}
		}
		// task79-v1 excludes singleton groups from LS3.
		if len(group) >= 2 {
			first = append(first, float64(len([]rune(group[0]))))
			last = append(last, float64(len([]rune(group[len(group)-1]))))
		}
	}
	cv := 0.0
	if len(lengths) > 0 && mean(lengths) != 0 {
		cv = sd(lengths, mean(lengths)) / mean(lengths)
	}
	return map[string]float64{
		"2DL1_LAYOUT_POSITION_MI":          normalizedMI(lengthClass, boundary),
		"BP1_BOUNDARY_TOKEN_NMI":           normalizedMI(token, boundary),
		"LS1_LINE_LENGTH_CV":               cv,
		"LS2_POSITIONAL_LEXICON_NMI":       normalizedMI(token, pos),
		"LS3_BOUNDARY_LENGTH_ASYMMETRY":    math.Abs(mean(first) - mean(last)),
		"LS4_WITHIN_LINE_EXACT_REPETITION": safeDiv(float64(repeats), float64(transitions)),
	}
}
