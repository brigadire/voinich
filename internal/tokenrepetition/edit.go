package tokenrepetition

// LevenshteinGlyphs computes the classic edit distance between two glyph
// sequences (task60 section 15): standard O(len(a)*len(b)) dynamic
// program. Tokens are short, and this is only ever called on adjacent
// pairs (O(number of transitions) calls total, task60 section 48), never
// all-pairs across the vocabulary.
func LevenshteinGlyphs(a, b []string) int {
	la, lb := len(a), len(b)
	if la == 0 {
		return lb
	}
	if lb == 0 {
		return la
	}
	prev := make([]int, lb+1)
	curr := make([]int, lb+1)
	for j := 0; j <= lb; j++ {
		prev[j] = j
	}
	for i := 1; i <= la; i++ {
		curr[0] = i
		for j := 1; j <= lb; j++ {
			cost := 1
			if a[i-1] == b[j-1] {
				cost = 0
			}
			del := prev[j] + 1
			ins := curr[j-1] + 1
			sub := prev[j-1] + cost
			m := del
			if ins < m {
				m = ins
			}
			if sub < m {
				m = sub
			}
			curr[j] = m
		}
		prev, curr = curr, prev
	}
	return prev[lb]
}

// NormalizedDistance is d(a,b)/max(len(a),len(b)) (task60 section 17).
func NormalizedDistance(a, b []string, d int) float64 {
	m := len(a)
	if len(b) > m {
		m = len(b)
	}
	if m == 0 {
		return 0
	}
	return float64(d) / float64(m)
}

// ClassifyDistanceOne determines the SUBSTITUTION/INSERTION/DELETION
// operation that turns a into b, given that their glyph-sequence edit
// distance is exactly 1 (task60 section 16), and the 0-indexed position
// of the changed/inserted/deleted glyph within the longer of the two
// tokens (or a's own length for a substitution, where both are equal).
// ok is false only if the precondition (distance exactly 1) is violated.
func ClassifyDistanceOne(a, b []string) (op string, position int, sourceGlyph, targetGlyph string, ok bool) {
	switch len(b) - len(a) {
	case 0:
		diff := -1
		mismatches := 0
		for i := range a {
			if a[i] != b[i] {
				mismatches++
				diff = i
			}
		}
		if mismatches != 1 {
			return "", 0, "", "", false
		}
		return "SUBSTITUTION", diff, a[diff], b[diff], true
	case 1: // b is one glyph longer than a: an INSERTION into a produces b
		i := 0
		for i < len(a) && a[i] == b[i] {
			i++
		}
		if !glyphsEqual(a[i:], b[i+1:]) {
			return "", 0, "", "", false
		}
		return "INSERTION", i, "", b[i], true
	case -1: // a is one glyph longer than b: a DELETION from a produces b
		i := 0
		for i < len(b) && a[i] == b[i] {
			i++
		}
		if !glyphsEqual(a[i+1:], b[i:]) {
			return "", 0, "", "", false
		}
		return "DELETION", i, a[i], "", true
	default:
		return "", 0, "", "", false
	}
}

// ClassifyEditDistanceOne exposes the shared Task60 edit-operation
// classification to analyses which enumerate vocabulary-wide edit edges.
// Its result and preconditions are identical to ClassifyDistanceOne.
func ClassifyEditDistanceOne(a, b []string) (op string, position int, sourceGlyph, targetGlyph string, ok bool) {
	return ClassifyDistanceOne(a, b)
}

func glyphsEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// PositionClass buckets a 0-indexed position within a token of the given
// length into BEGIN/MIDDLE/END thirds (task60 section 21), fixed here
// before any corpus is scored. For an INSERTION, length is the longer
// token's length (position indexes into it); for a DELETION, length is
// the longer (source) token's length; for a SUBSTITUTION, both tokens
// have the same length.
func PositionClass(position, length int) string {
	if length <= 1 {
		return "MIDDLE"
	}
	frac := float64(position) / float64(length-1)
	switch {
	case frac < 1.0/3:
		return "BEGIN"
	case frac >= 2.0/3:
		return "END"
	default:
		return "MIDDLE"
	}
}
