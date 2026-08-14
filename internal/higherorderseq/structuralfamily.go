package higherorderseq

// structuralFamilyRows implements task22 Part L sections 59-63: purely
// descriptive checks, using only already-frozen structural-normalize
// classes (never a new morphological assumption), of whether the
// conditional-dependence sign survives substituting A or C for a frozen
// structural relative.
//
// For an A-relative A', it asks whether P(C|A',B) > P(C|B) still holds. For
// a C-relative C', it asks whether P(C'|A,B) > P(C'|B) still holds. Both
// reuse the same pooled left/right context tables Parts F, G and O already
// compute; this is not a new significance search (section 63).
func structuralFamilyRows(cand Candidate, blocks []Block, relatives map[string][]string) []StructuralFamilyRow {
	leftB, leftBC, totalB, totalBC := leftContextCounts(cand, blocks)
	rightB, rightAB, totalRightB, totalAB := rightContextCounts(cand, blocks)
	baselinePCGivenB := 0.0
	if totalB > 0 {
		baselinePCGivenB = float64(totalBC) / float64(totalB)
	}

	var rows []StructuralFamilyRow
	for _, rel := range relatives[cand.A()] {
		row := StructuralFamilyRow{Sequence: cand.Sequence, TokenRole: "A", Token: cand.A(), Relative: rel, FrozenP: baselinePCGivenB}
		row.Sufficient = leftB[rel] >= contextAltMinCount
		if leftB[rel] > 0 {
			row.RelativeP = float64(leftBC[rel]) / float64(leftB[rel])
		}
		row.SignHolds = row.Sufficient && row.RelativeP > baselinePCGivenB
		rows = append(rows, row)
	}
	for _, rel := range relatives[cand.C()] {
		baseline := 0.0
		if totalRightB > 0 {
			baseline = float64(rightB[rel]) / float64(totalRightB)
		}
		row := StructuralFamilyRow{Sequence: cand.Sequence, TokenRole: "C", Token: cand.C(), Relative: rel, FrozenP: baseline}
		row.Sufficient = totalAB >= contextAltMinCount
		if totalAB > 0 {
			row.RelativeP = float64(rightAB[rel]) / float64(totalAB)
		}
		row.SignHolds = row.Sufficient && row.RelativeP > baseline
		rows = append(rows, row)
	}
	return rows
}
