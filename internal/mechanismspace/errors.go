package mechanismspace

import "math/rand"

// InjectErrors applies task66 section 67's scribal-like error model to an
// already-transformed token corpus: independent, seeded glyph
// substitution/deletion/insertion and token-boundary insertion/deletion,
// each at rate (0-1). It is a secondary robustness probe only (section
// 68); it is never used for model selection (section 67's own
// prohibition).
func InjectErrors(tokens [][]string, rate float64, seed int64) [][]string {
	if rate <= 0 {
		return tokens
	}
	r := rand.New(rand.NewSource(seed))
	alphabet := []string{"a", "b", "c", "d", "e", "f", "g", "h", "i", "j", "k", "l", "m", "n"}
	var out [][]string
	for _, t := range tokens {
		var nt []string
		for _, g := range t {
			switch {
			case r.Float64() < rate/3: // deletion
				continue
			case r.Float64() < rate/3: // substitution
				nt = append(nt, alphabet[r.Intn(len(alphabet))])
			default:
				nt = append(nt, g)
			}
			if r.Float64() < rate/3 { // insertion
				nt = append(nt, alphabet[r.Intn(len(alphabet))])
			}
		}
		if len(nt) == 0 {
			nt = []string{alphabet[r.Intn(len(alphabet))]}
		}
		// token-boundary deletion: merge with the previous token.
		if len(out) > 0 && r.Float64() < rate/3 {
			out[len(out)-1] = append(out[len(out)-1], nt...)
			continue
		}
		out = append(out, nt)
		// token-boundary insertion: split the token just appended.
		if len(nt) > 1 && r.Float64() < rate/3 {
			split := 1 + r.Intn(len(nt)-1)
			out[len(out)-1] = nt[:split]
			out = append(out, nt[split:])
		}
	}
	return out
}
