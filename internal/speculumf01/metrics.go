package speculumf01

// CorruptionMetrics is the per-scenario, per-message measurement set
// required by task76 Block 5.
type CorruptionMetrics struct {
	Scenario                string
	Message                 string
	Decoded                 string
	ExactRecovery           bool
	CharacterErrorRate      float64 // edit distance / max(len(gt), len(decoded))
	FractionAfterFirstError float64 // of the characters strictly after the first mismatch, fraction still correct
	ErrorClass              string  // "none" | "local" | "synchronization" | "cascading" | "global"
	Detectable              bool    // decoded string is not a lexicon word (only meaningful for natural-language messages)
	CorrectableWithoutM     bool    // a unique nearest lexicon word exists at edit distance <= 1
}

func Measure(scenario, groundTruth, decoded string, lexicon map[string]bool, naturalLanguage bool) CorruptionMetrics {
	gt := []rune(groundTruth)
	dec := []rune(decoded)
	m := CorruptionMetrics{Scenario: scenario, Message: groundTruth, Decoded: decoded}
	m.ExactRecovery = groundTruth == decoded

	dist := levenshtein(gt, dec)
	maxLen := max(len(gt), len(dec))
	if maxLen > 0 {
		m.CharacterErrorRate = float64(dist) / float64(maxLen)
	}

	firstErr := -1
	for i := 0; i < min(len(gt), len(dec)); i++ {
		if gt[i] != dec[i] {
			firstErr = i
			break
		}
	}
	if len(gt) != len(dec) && firstErr == -1 {
		firstErr = min(len(gt), len(dec))
	}
	if firstErr == -1 {
		m.FractionAfterFirstError = 1.0
		m.ErrorClass = "none"
	} else {
		remaining := 0
		correct := 0
		for i := firstErr + 1; i < len(gt); i++ {
			remaining++
			if i < len(dec) && dec[i] == gt[i] {
				correct++
			}
		}
		if remaining == 0 {
			m.FractionAfterFirstError = 1.0
		} else {
			m.FractionAfterFirstError = float64(correct) / float64(remaining)
		}
		m.ErrorClass = classifyError(m.FractionAfterFirstError, dist, len(gt), len(dec))
	}

	if naturalLanguage && lexicon != nil {
		m.Detectable = !lexicon[decoded]
		m.CorrectableWithoutM = nearestLexiconMatchUnique(decoded, lexicon)
	}
	return m
}

// classifyError turns the post-first-error positional-correctness fraction
// and total edit distance into the qualitative class task76 Block 5 asks
// for. The key discriminator between "synchronization" and "cascading" is
// NOT how many character positions look wrong after the first mismatch —
// a one-ring collapse makes every later position look wrong under a
// naive per-position comparison — but whether a single small edit (one
// indel) accounts for all of it. That is precisely the classic
// bit-slip/frame-sync failure signature: low edit distance, high apparent
// positional error. A genuinely "cascading"/scrambled corruption has no
// such short realignment and shows a large edit distance too.
//   - "local": everything after the first mismatch still matches.
//   - "synchronization": post-error positions mismatch, but the whole
//     decoded string is within edit distance 2 of ground truth (a
//     single-indel realignment explains it — e.g. one ring physically
//     collapsing the stack).
//   - "global": every position differs (dist equals the full length, with
//     no length change) -- the signature of a single-cause uniform additive
//     shift (e.g. orientation-mark loss), which is a single degree of
//     freedom, not len(M) independent errors, even though it touches every
//     character.
//   - "cascading": post-error positions mismatch AND the edit distance is
//     large but less than the full length -- several genuinely independent
//     errors, no simple realignment recovers the message.
func classifyError(fractionAfter float64, dist, gtLen, decLen int) string {
	if fractionAfter >= 0.999 {
		return "local"
	}
	if gtLen == decLen && dist == gtLen {
		return "global"
	}
	if dist <= 2 {
		return "synchronization"
	}
	return "cascading"
}

func nearestLexiconMatchUnique(decoded string, lexicon map[string]bool) bool {
	dec := []rune(decoded)
	bestDist := -1
	matches := 0
	for _, word := range sortedLexiconKeys(lexicon) {
		d := levenshtein(dec, []rune(word))
		if bestDist == -1 || d < bestDist {
			bestDist = d
			matches = 1
		} else if d == bestDist {
			matches++
		}
	}
	return bestDist <= 1 && matches == 1
}

func sortedLexiconKeys(m map[string]bool) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	// simple insertion sort keeps this deterministic without importing sort
	// twice across the package; n is tiny (a few dozen lexicon entries).
	for i := 1; i < len(out); i++ {
		for j := i; j > 0 && out[j] < out[j-1]; j-- {
			out[j], out[j-1] = out[j-1], out[j]
		}
	}
	return out
}

func levenshtein(a, b []rune) int {
	na, nb := len(a), len(b)
	prev := make([]int, nb+1)
	cur := make([]int, nb+1)
	for j := 0; j <= nb; j++ {
		prev[j] = j
	}
	for i := 1; i <= na; i++ {
		cur[0] = i
		for j := 1; j <= nb; j++ {
			cost := 1
			if a[i-1] == b[j-1] {
				cost = 0
			}
			del := prev[j] + 1
			ins := cur[j-1] + 1
			sub := prev[j-1] + cost
			cur[j] = min(del, ins, sub)
		}
		prev, cur = cur, prev
	}
	return prev[nb]
}
