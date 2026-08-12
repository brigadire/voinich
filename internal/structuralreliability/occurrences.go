package structuralreliability

import (
	"math/rand"
	"sort"

	"zcore.dev/voinich/internal/profilestability"
	"zcore.dev/voinich/internal/validation"
)

// Occurrence is a single observed placement of a token: its absolute
// position in the line plus its immediate predecessor/successor, if any.
// This is the subsampling unit required by task section 13 - an occurrence
// together with everything BuildProfiles would have derived from it, never
// an artificial text.
type Occurrence struct {
	Position int
	HasLeft  bool
	Left     string
	HasRight bool
	Right    string
}

// CollectOccurrences walks the corpus exactly like profilestability.
// BuildProfiles, but keeps every individual occurrence instead of collapsing
// them into aggregate counts, so that a profile can later be rebuilt from
// any subset of a token's real occurrences.
func CollectOccurrences(corpus validation.Corpus) map[string][]Occurrence {
	result := make(map[string][]Occurrence)
	for _, line := range corpus.Lines {
		for position, token := range line.Tokens {
			occurrence := Occurrence{Position: position}
			if position > 0 {
				occurrence.HasLeft, occurrence.Left = true, line.Tokens[position-1]
			}
			if position+1 < len(line.Tokens) {
				occurrence.HasRight, occurrence.Right = true, line.Tokens[position+1]
			}
			result[token] = append(result[token], occurrence)
		}
	}
	return result
}

// ProfileFromOccurrences rebuilds a profilestability.Profile from any subset
// of a token's occurrences, using the identical field semantics BuildProfiles
// uses for the full corpus.
func ProfileFromOccurrences(occurrences []Occurrence) profilestability.Profile {
	profile := profilestability.Profile{Positions: map[int]int{}, Left: map[string]int{}, Right: map[string]int{}}
	for _, occurrence := range occurrences {
		profile.Count++
		profile.Positions[occurrence.Position]++
		if occurrence.HasLeft {
			profile.Left[occurrence.Left]++
		}
		if occurrence.HasRight {
			profile.Right[occurrence.Right]++
		}
	}
	return profile
}

// SampleOccurrences deterministically draws n distinct occurrences out of
// occurrences using rng, via a Fisher-Yates partial shuffle of a local index
// copy so the caller's slice and its order are left untouched.
func SampleOccurrences(occurrences []Occurrence, n int, rng *rand.Rand) []Occurrence {
	if n >= len(occurrences) {
		return occurrences
	}
	indexes := make([]int, len(occurrences))
	for i := range indexes {
		indexes[i] = i
	}
	for i := 0; i < n; i++ {
		j := i + rng.Intn(len(indexes)-i)
		indexes[i], indexes[j] = indexes[j], indexes[i]
	}
	selected := indexes[:n]
	sort.Ints(selected)
	result := make([]Occurrence, n)
	for i, index := range selected {
		result[i] = occurrences[index]
	}
	return result
}
