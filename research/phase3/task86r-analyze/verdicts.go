package main

// ExplicitRuleGrammarRequired implements G1_MODEL_LADDER_CONTRACT.md's
// final rule: SUPPORTED iff M5 passes both gates and beats every adequate
// M0-M4 candidate under the gain rule; NOT_SUPPORTED iff some M0-M4
// candidate passes both gates and no M5 beats it; INCONCLUSIVE otherwise.
func ExplicitRuleGrammarRequired(m5Adequate bool, m5BeatsAllAdequate bool, anyM0toM4Adequate bool, m5BeatsAnyM0toM4 bool) string {
	if m5Adequate && m5BeatsAllAdequate {
		return "SUPPORTED"
	}
	if anyM0toM4Adequate && !m5BeatsAnyM0toM4 {
		return "NOT_SUPPORTED"
	}
	return "INCONCLUSIVE"
}

// UnexplainedStructure is derived strictly from the frozen
// adequacy/failure results (task86r.txt section 48): PRESENT if a
// selected/minimal candidate exists but StructuralAdequacy failed on at
// least one metric family or the minimal class differs from a
// predictively-adequate one; NOT_DETECTED if the minimal candidate passes
// every family cleanly; INCONCLUSIVE if no minimal class was identified.
func UnexplainedStructure(minimalClassExists bool, allFamiliesClean bool) string {
	if !minimalClassExists {
		return "INCONCLUSIVE"
	}
	if allFamiliesClean {
		return "NOT_DETECTED"
	}
	return "PRESENT"
}
