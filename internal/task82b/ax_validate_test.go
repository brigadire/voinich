package task82b

import "testing"

func TestAXSensitivityOnDoyle(t *testing.T) {
	lines, err := LoadLines("../..", "Doyle", CarrierPaths["Doyle"])
	if err != nil {
		t.Skip("Doyle carrier not present: " + err.Error())
	}
	tokenAtoms, glyphAtoms := BuildAtoms(lines.Tokens)
	numLines := len(lines.Tokens)

	op := Operator{StructuralClass: "FIRST_GLYPH_OF_TOKEN", NullClass: "PER_GROUP"}
	sel := Apply(op, tokenAtoms, glyphAtoms, numLines)
	realGroups := Render(sel, tokenAtoms, glyphAtoms, numLines)
	realAX := ComputeAX(realGroups)

	nullSel := RandomSubsequenceMatched(sel, 42)
	nullGroups := Render(nullSel, tokenAtoms, glyphAtoms, numLines)
	nullAX := ComputeAX(nullGroups)

	stratSel := StratifiedRandom(sel, 42)
	stratGroups := Render(stratSel, tokenAtoms, glyphAtoms, numLines)
	stratAX := ComputeAX(stratGroups)

	t.Logf("REAL    AX3=%.4f AX4=%.4f AX5=%.4f(k=%d) AX6=%.4f n=%d",
		realAX.AX3StreamEntropy, realAX.AX4TypeTokenRatio, realAX.AX5PeriodicNMIMax, realAX.AX5BestPeriod, realAX.AX6LinePersistence, realAX.N)
	t.Logf("RANDSUB AX3=%.4f AX4=%.4f AX5=%.4f(k=%d) AX6=%.4f n=%d",
		nullAX.AX3StreamEntropy, nullAX.AX4TypeTokenRatio, nullAX.AX5PeriodicNMIMax, nullAX.AX5BestPeriod, nullAX.AX6LinePersistence, nullAX.N)
	t.Logf("STRAT   AX3=%.4f AX4=%.4f AX5=%.4f(k=%d) AX6=%.4f n=%d",
		stratAX.AX3StreamEntropy, stratAX.AX4TypeTokenRatio, stratAX.AX5PeriodicNMIMax, stratAX.AX5BestPeriod, stratAX.AX6LinePersistence, stratAX.N)
}
