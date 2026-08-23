// Package fontanafamily implements deliberately small operational models for
// the non-Speculum devices selected in task78.  The package models only the
// invariant or explicitly named reconstruction profiles documented under
// research/phase2/fontana/task78; it is not a collection of conjectural text
// generators.
package fontanafamily

import "fmt"

// Direction is an ordered traversal convention.  Zero is always the source-
// supported direction for the relevant model; Reverse is an ablation/profile
// alternative, not an additional historical claim.
type Direction int

const (
	Forward Direction = iota
	Reverse
)

// Normalize returns v in [0,n).  It is shared by the cyclic models.
func Normalize(v, n int) int {
	if n <= 0 {
		panic("fontanafamily: non-positive modulus")
	}
	return ((v % n) + n) % n
}

// Exact compares rune sequences without importing semantic or lexicon-based
// guesses into the formal decoder.
func Exact(want, got string) bool { return want == got }

// SymbolAccuracy is positional accuracy. Missing positions count as errors.
func SymbolAccuracy(want, got string) float64 {
	w, g := []rune(want), []rune(got)
	den := len(w)
	if len(g) > den {
		den = len(g)
	}
	if den == 0 {
		return 1
	}
	correct := 0
	for i := 0; i < len(w) && i < len(g); i++ {
		if w[i] == g[i] {
			correct++
		}
	}
	return float64(correct) / float64(den)
}

// NormalizedEditDistance is Levenshtein distance divided by the longer rune
// sequence. It distinguishes a frame loss from positional substitutions.
func NormalizedEditDistance(a, b string) float64 {
	x, y := []rune(a), []rune(b)
	if len(x) == 0 && len(y) == 0 {
		return 0
	}
	prev := make([]int, len(y)+1)
	for j := range prev {
		prev[j] = j
	}
	for i := 1; i <= len(x); i++ {
		cur := make([]int, len(y)+1)
		cur[0] = i
		for j := 1; j <= len(y); j++ {
			cost := 1
			if x[i-1] == y[j-1] {
				cost = 0
			}
			cur[j] = min(cur[j-1]+1, prev[j]+1, prev[j-1]+cost)
		}
		prev = cur
	}
	den := max(len(x), len(y))
	return float64(prev[len(y)]) / float64(den)
}

// ValidateAlphabet rejects duplicate/empty symbol sets, which otherwise make
// lookup and rotation results silently ambiguous.
func ValidateAlphabet(alphabet []rune) error {
	if len(alphabet) == 0 {
		return fmt.Errorf("empty alphabet")
	}
	seen := make(map[rune]bool, len(alphabet))
	for _, r := range alphabet {
		if seen[r] {
			return fmt.Errorf("duplicate alphabet symbol %q", r)
		}
		seen[r] = true
	}
	return nil
}
