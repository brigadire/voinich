package positionalcontinuation

import "strings"

// buildSurroundingContextRows implements task23 Part O sections 83-87:
// purely descriptive comparison of the -3..+3 token window around s-aiin-chey
// occurrences against every other s-aiin occurrence, to check whether the
// four chey occurrences are really one repeated longer formula.
func buildSurroundingContextRows(occs []SAiinOccurrence) []SurroundingContextRow {
	var cheyGroup, otherGroup []SAiinOccurrence
	for _, o := range occs {
		if o.X == "" {
			continue
		}
		if o.X == FrozenChey {
			cheyGroup = append(cheyGroup, o)
		} else {
			otherGroup = append(otherGroup, o)
		}
	}
	return []SurroundingContextRow{
		surroundingRow("chey", cheyGroup),
		surroundingRow("not_chey", otherGroup),
	}
}

func surroundingRow(group string, occs []SAiinOccurrence) SurroundingContextRow {
	preceding := map[string]int{}
	following := map[string]int{}
	unique := map[string]bool{}
	for _, o := range occs {
		preceding[o.TokensBefore[2]]++
		following[o.TokensAfter[0]]++
		key := strings.Join([]string{o.TokensBefore[2], o.TokensBefore[1], o.TokensBefore[0], o.TokensAfter[0], o.TokensAfter[1], o.TokensAfter[2]}, "|")
		unique[key] = true
	}
	return SurroundingContextRow{
		Group: group, OccurrenceCount: len(occs),
		PrecedingEntropyBits: countEntropyBits(preceding), FollowingEntropyBits: countEntropyBits(following),
		UniqueSurroundingContexts: len(unique),
	}
}
