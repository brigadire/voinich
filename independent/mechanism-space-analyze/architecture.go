package main

import (
	"fmt"
	"strings"
)

// FinalVerdict is one row of FINAL_ARCHITECTURE.tsv (task66 section 70).
type FinalVerdict struct {
	Operation, Verdict, Evidence, Caveat string
}

func avgFamilyScore(rows []FamilyMetricsRow, mechanism, family string) (float64, bool) {
	sum, n := 0.0, 0
	for _, r := range rows {
		if r.Mechanism != mechanism {
			continue
		}
		if v, ok := r.FamilyScores[family]; ok {
			sum += v
			n++
		}
	}
	if n == 0 {
		return 0, false
	}
	return sum / float64(n), true
}

func avgFamilies(rows []FamilyMetricsRow, mechanism string, families ...string) (float64, bool) {
	sum, n := 0.0, 0
	for _, f := range families {
		if v, ok := avgFamilyScore(rows, mechanism, f); ok {
			sum += v
			n++
		}
	}
	if n == 0 {
		return 0, false
	}
	return sum / float64(n), true
}

// verdictFromDelta turns a "with vs without" family-score delta into
// task66 section 70's five-way verdict.
func verdictFromDelta(with, without float64, ok bool) (string, string) {
	if !ok {
		return "UNRESOLVED", "insufficient data to compare"
	}
	delta := with - without
	ev := fmt.Sprintf("family score with=%.3g without=%.3g delta=%.3g", with, without, delta)
	switch {
	case delta > 0.3:
		return "REQUIRED", ev
	case delta > 0.1:
		return "SUPPORTED", ev
	case delta > -0.1:
		return "NOT_REQUIRED", ev
	default:
		return "DISFAVORED", ev
	}
}

// DeriveFinalArchitecture computes task66 section 70's eight operation-
// level verdicts from the already-computed development and ablation
// family-score tables plus the plaintext-sensitivity classes, following
// the critical-test decompositions of sections 59-63.
func DeriveFinalArchitecture(devRows, ablationRows []FamilyMetricsRow, sensitivityClasses map[string]int) []FinalVerdict {
	var out []FinalVerdict
	add := func(op, verdict, evidence, caveat string) {
		out = append(out, FinalVerdict{Operation: op, Verdict: verdict, Evidence: evidence, Caveat: caveat})
	}

	// MEMORY_REQUIRED (sections 59-61): S_ONLY vs G_ONLY on the topology
	// family, since topology is the property state is meant to explain.
	sOnly, sOK := avgFamilyScore(ablationRows, "S_ONLY", "LOCAL_REGIME_TOPOLOGY")
	gOnly, gOK := avgFamilyScore(ablationRows, "G_ONLY", "LOCAL_REGIME_TOPOLOGY")
	v, ev := verdictFromDelta(sOnly, gOnly, sOK && gOK)
	add("MEMORY_REQUIRED", v, ev, "no Voynich target tuning; compares ablation S_ONLY vs G_ONLY on LOCAL_REGIME_TOPOLOGY")

	// SLOW_STATE_REQUIRED (section 18): M5 (slow drift) vs M4 (per-unit
	// state) on the topology family - does slowing the state down help
	// correlation length specifically.
	m5, m5OK := avgFamilyScore(devRows, "M5_DRIFT_N20", "LOCAL_REGIME_TOPOLOGY")
	m4, m4OK := avgFamilyScore(devRows, "M4_STATE_K4_A", "LOCAL_REGIME_TOPOLOGY")
	v, ev = verdictFromDelta(m5, m4, m5OK && m4OK)
	add("SLOW_STATE_REQUIRED", v, ev, "compares M5 (drift scale 20) vs M4 (per-unit state) on LOCAL_REGIME_TOPOLOGY")

	// MACRO_STATE_REQUIRED (section 62): M+S vs S_ONLY.
	mPlusS, mpsOK := avgFamilyScore(ablationRows, "M_PLUS_S", "LOCAL_REGIME_TOPOLOGY")
	v, ev = verdictFromDelta(mPlusS, sOnly, mpsOK && sOK)
	add("MACRO_STATE_REQUIRED", v, ev, "compares ablation M_PLUS_S vs S_ONLY on LOCAL_REGIME_TOPOLOGY")

	// CONSTRAINED_FORMATION_REQUIRED (sections 59-61): G_ONLY vs S_ONLY
	// on the families constrained formation specifically targets.
	gComposite, gcOK := avgFamilies(ablationRows, "G_ONLY", "POSITIONAL_STRUCTURE", "CHARACTER_ENTROPY", "TOKEN_FORMATION")
	sComposite, scOK := avgFamilies(ablationRows, "S_ONLY", "POSITIONAL_STRUCTURE", "CHARACTER_ENTROPY", "TOKEN_FORMATION")
	v, ev = verdictFromDelta(gComposite, sComposite, gcOK && scOK)
	add("CONSTRAINED_FORMATION_REQUIRED", v, ev, "compares ablation G_ONLY vs S_ONLY on POSITIONAL_STRUCTURE+CHARACTER_ENTROPY+TOKEN_FORMATION")

	// GENERATED_BOUNDARIES_REQUIRED (section 63): STREAM+generated
	// boundaries (M9 state grouping) vs WORD_PRESERVING form-only (M3).
	m9, m9OK := avgFamilies(devRows, "M9_GROUP_FORM_STATE", "POSITIONAL_STRUCTURE", "CHARACTER_ENTROPY")
	m3, m3OK := avgFamilies(devRows, "M3_FORM_MEDIUM", "POSITIONAL_STRUCTURE", "CHARACTER_ENTROPY")
	v, ev = verdictFromDelta(m9, m3, m9OK && m3OK)
	add("GENERATED_BOUNDARIES_REQUIRED", v, ev, "compares STREAM+generated-boundary M9 vs WORD_PRESERVING form-only M3 on POSITIONAL_STRUCTURE+CHARACTER_ENTROPY")

	// HOMOPHONY_HELPFUL: M2 vs M1 (both memoryless substitution).
	m2, m2OK := avgFamilies(devRows, "M2_HOMOPHONY_H4", "TOKEN_ORDER", "POSITIONAL_STRUCTURE", "CHARACTER_ENTROPY", "REPETITION_EDIT_GEOMETRY")
	m1, m1OK := avgFamilies(devRows, "M1_MONOALPHABETIC", "TOKEN_ORDER", "POSITIONAL_STRUCTURE", "CHARACTER_ENTROPY", "REPETITION_EDIT_GEOMETRY")
	v, ev = verdictFromDelta(m2, m1, m2OK && m1OK)
	add("HOMOPHONY_HELPFUL", v, ev, "compares M2 (H=4 homophony) vs M1 (monoalphabetic) across TOKEN_ORDER/POSITIONAL_STRUCTURE/CHARACTER_ENTROPY/REPETITION_EDIT_GEOMETRY")

	// STOCHASTIC_OUTPUT_REQUIRED: M2 (stochastic) vs M1 (deterministic)
	// again, and M8 RANDOM vs M8 FIXED grouping.
	m8r, m8rOK := avgFamilyScore(devRows, "M8_BOUNDARY_RANDOM", "REPETITION_EDIT_GEOMETRY")
	m8f, m8fOK := avgFamilyScore(devRows, "M8_BOUNDARY_FIXED", "REPETITION_EDIT_GEOMETRY")
	stochDelta1 := 0.0
	if m2OK && m1OK {
		stochDelta1 = m2 - m1
	}
	stochDelta2 := 0.0
	if m8rOK && m8fOK {
		stochDelta2 = m8r - m8f
	}
	combinedOK := (m2OK && m1OK) || (m8rOK && m8fOK)
	v, ev = verdictFromDelta(stochDelta1+stochDelta2, 0, combinedOK)
	add("STOCHASTIC_OUTPUT_REQUIRED", v, ev, "compares M2 vs M1 and M8-RANDOM vs M8-FIXED; a positive combined delta favors stochastic output")

	// PLAINTEXT_DEPENDENCE_PRESERVED (sections 65-66): majority
	// sensitivity class across representative mechanisms.
	strong := sensitivityClasses["STRONG_INPUT_DEPENDENCE"] + sensitivityClasses["PARTIAL_INPUT_DEPENDENCE"]
	weak := sensitivityClasses["WEAK_INPUT_DEPENDENCE"] + sensitivityClasses["INPUT_INDEPENDENT"]
	switch {
	case strong == 0 && weak == 0:
		add("PLAINTEXT_DEPENDENCE_PRESERVED", "UNRESOLVED", "no plaintext-sensitivity rows available", "")
	case strong >= weak:
		add("PLAINTEXT_DEPENDENCE_PRESERVED", "SUPPORTED", fmt.Sprintf("%d/%d representative mechanisms show strong/partial input dependence", strong, strong+weak), "")
	default:
		add("PLAINTEXT_DEPENDENCE_PRESERVED", "DISFAVORED", fmt.Sprintf("%d/%d representative mechanisms show weak/no input dependence", weak, strong+weak), "at least one mechanism risks being an independent generator rather than a transformation")
	}

	return out
}

// WriteFinalArchitectureTSV writes FINAL_ARCHITECTURE.tsv.
func WriteFinalArchitectureTSV(path string, verdicts []FinalVerdict) error {
	var b strings.Builder
	b.WriteString("operation\tverdict\tevidence\tcaveat\n")
	for _, v := range verdicts {
		b.WriteString(fmt.Sprintf("%s\t%s\t%s\t%s\n", v.Operation, v.Verdict, v.Evidence, v.Caveat))
	}
	return writeFile(path, b.String())
}
