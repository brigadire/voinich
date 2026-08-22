package inversehomophony

import "sort"

func sortStrings(x []string) { sort.Strings(x) }

func sortPairScores(pairs []PairScore) {
	sort.Slice(pairs, func(i, j int) bool {
		if pairs[i].Score != pairs[j].Score {
			return pairs[i].Score > pairs[j].Score
		}
		if pairs[i].A != pairs[j].A {
			return pairs[i].A < pairs[j].A
		}
		return pairs[i].B < pairs[j].B
	})
}
