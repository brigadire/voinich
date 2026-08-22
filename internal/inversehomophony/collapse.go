package inversehomophony

// Collapse rewrites tokens through partition, producing the collapsed
// corpus R(H(P)) (task57 section 6): every occurrence of a cipher token is
// replaced by its class label. Token count and line structure are
// preserved exactly - collapse never merges or splits occurrences, only
// relabels them.
func Collapse(tokens []string, partition Partition) []string {
	out := make([]string, len(tokens))
	for i, t := range tokens {
		out[i] = partition[t]
	}
	return out
}

// CollapseLines applies Collapse line-by-line, preserving line structure.
func CollapseLines(lines [][]string, partition Partition) [][]string {
	out := make([][]string, len(lines))
	for i, line := range lines {
		out[i] = Collapse(line, partition)
	}
	return out
}
