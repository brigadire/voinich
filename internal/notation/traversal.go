package notation

// TraversalSummary reports structural traversal facts about a USC record
// set — record count, boundary-preserving block count, and whether physical
// lines are source-observed — without computing any structural, rarefaction,
// bootstrap, or comparative metric. It exists so a technical pre-run can
// prove that loading and traversal of a frozen production corpus is
// deterministic and complete without producing a scientific result.
func TraversalSummary(rs []Record) (blocks int, tokens int, lines bool) {
	return len(buildStructuralBlocks(rs)), len(rs), linesObserved(rs)
}
