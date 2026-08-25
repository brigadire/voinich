package main

// ClassSynth is one model class's cross-transcription synthesis.
type ClassSynth struct {
	ModelClass            string
	PredictiveAdequate    bool
	MetricStability       map[string]StabilityClass
	StructuralAdequate    bool
	EditFamilyStability   StabilityClass
	LexicalFamilyStability StabilityClass
	EditPassByT           map[string]bool
	LexicalPassByT        map[string]bool
	MetricPassByT         map[string]map[string]bool
	MultiScaleSufficient  bool
	AnyFailure            bool
}

// SynthResult is the fully combined cross-transcription verdict set.
type SynthResult struct {
	ByClass                     map[string]ClassSynth
	MinimalByT                  map[string]*MinimalityCandidate
	EquivalenceByT              map[string][]MinimalityCandidate
	G1MinimalClass               string
	TokenFormationDepth          string
	LadderEdges                  []LadderEdge
	ExplicitRuleGrammarRequired  string
	UnexplainedStructure         string
	GrammarSufficient            string
}

func classSynthesize(class string, dZ, dI *StageDResult, idx *ThresholdIndex) ClassSynth {
	cs := ClassSynth{ModelClass: class, MetricStability: map[string]StabilityClass{}, EditPassByT: map[string]bool{}, LexicalPassByT: map[string]bool{}, MetricPassByT: map[string]map[string]bool{"ZL3b": {}, "IT2a": {}}}
	if dZ == nil || dI == nil || dZ.Model == nil || dI.Model == nil {
		cs.AnyFailure = true
		return cs
	}
	allStable := true
	for _, m := range predictiveMetricNames {
		gz, gi := dZ.MetricGates[m], dI.MetricGates[m]
		st := ClassifyStability(gz.ImprovementB1, gi.ImprovementB1)
		cs.MetricStability[m] = st
		if !atLeastDirectionStable(st) {
			allStable = false
		}
	}
	cs.PredictiveAdequate = dZ.PredictivePass && dI.PredictivePass && allStable

	editPassZ, editCountZ := FamilyPass(editFamilyMetrics, scalePassMap(dZ, 1.0, idx), 3)
	editPassI, editCountI := FamilyPass(editFamilyMetrics, scalePassMap(dI, 1.0, idx), 3)
	lexPassZ, lexCountZ := FamilyPass(lexicalFamilyMetrics, scalePassMap(dZ, 1.0, idx), 2)
	lexPassI, lexCountI := FamilyPass(lexicalFamilyMetrics, scalePassMap(dI, 1.0, idx), 2)
	cs.EditPassByT["ZL3b"], cs.EditPassByT["IT2a"] = editPassZ, editPassI
	cs.LexicalPassByT["ZL3b"], cs.LexicalPassByT["IT2a"] = lexPassZ, lexPassI
	cs.MetricPassByT["ZL3b"] = scalePassMap(dZ, 1.0, idx)
	cs.MetricPassByT["IT2a"] = scalePassMap(dI, 1.0, idx)

	editEffectZ := float64(editCountZ) / 4.0
	editEffectI := float64(editCountI) / 4.0
	lexEffectZ := float64(lexCountZ) / 3.0
	lexEffectI := float64(lexCountI) / 3.0
	cs.EditFamilyStability = ClassifyStability(editEffectZ, editEffectI)
	cs.LexicalFamilyStability = ClassifyStability(lexEffectZ, lexEffectI)

	cs.StructuralAdequate = editPassZ && editPassI && lexPassZ && lexPassI &&
		atLeastDirectionStable(cs.EditFamilyStability) && atLeastDirectionStable(cs.LexicalFamilyStability)

	msZ := multiScaleSufficient(dZ, idx)
	msI := multiScaleSufficient(dI, idx)
	cs.MultiScaleSufficient = msZ && msI

	cs.AnyFailure = len(dZ.FailureClasses) > 0 || len(dI.FailureClasses) > 0
	return cs
}

func scalePassMap(d *StageDResult, scale float64, idx *ThresholdIndex) map[string]bool {
	out := map[string]bool{}
	gen, ok := d.Generation[scale]
	if !ok || !d.HeldF2Valid {
		return out
	}
	for _, m := range StructuralMetricIDs {
		medGen, okGen := gen.MedianAtStop[m]
		heldV, okHeld := d.HeldF2[m]
		pass, _, _ := StructuralMetricPass(d.ModelClass, d.CandidateID, m, medGen, okGen, heldV, okHeld, idx)
		out[m] = pass
	}
	return out
}

func multiScaleSufficient(d *StageDResult, idx *ThresholdIndex) bool {
	if len(d.FailureClasses) > 0 {
		return false
	}
	var results []bool
	for _, sc := range scales {
		gen, ok := d.Generation[sc]
		if !ok || !gen.Converged || gen.ExcessiveCV {
			return false
		}
		passMap := map[string]bool{}
		for _, m := range StructuralMetricIDs {
			medGen, okGen := gen.MedianAtStop[m]
			heldV, okHeld := d.HeldF2[m]
			pass, _, _ := StructuralMetricPass(d.ModelClass, d.CandidateID, m, medGen, okGen, heldV, okHeld, idx)
			passMap[m] = pass
		}
		editPass, _ := FamilyPass(editFamilyMetrics, passMap, 3)
		lexPass, _ := FamilyPass(lexicalFamilyMetrics, passMap, 2)
		results = append(results, editPass && lexPass)
	}
	for i := 1; i < len(results); i++ {
		if results[i] != results[0] {
			return false
		}
	}
	return len(results) == len(scales)
}

func synthesize(stageD map[string]map[string]*StageDResult, idx *ThresholdIndex) SynthResult {
	classes := []string{"M0", "M1", "M2", "M3", "M4", "M5"}
	out := SynthResult{ByClass: map[string]ClassSynth{}, MinimalByT: map[string]*MinimalityCandidate{}, EquivalenceByT: map[string][]MinimalityCandidate{}}
	for _, c := range classes {
		out.ByClass[c] = classSynthesize(c, stageD["ZL3b"][c], stageD["IT2a"][c], idx)
	}

	for _, tname := range []string{"ZL3b", "IT2a"} {
		var cands []MinimalityCandidate
		for _, c := range classes {
			d := stageD[tname][c]
			cs := out.ByClass[c]
			mc := MinimalityCandidate{ModelClass: c, CandidateID: d.CandidateID, PredictiveAdequate: cs.PredictiveAdequate, StructuralAdequate: cs.StructuralAdequate, AnyFailure: cs.AnyFailure}
			if d.Model != nil {
				mc.Complexity = d.Complexity.Total()
			} else {
				mc.Complexity = posInf()
			}
			cands = append(cands, mc)
		}
		winner, equiv := SelectMinimalPerTranscription(cands)
		out.MinimalByT[tname] = winner
		out.EquivalenceByT[tname] = equiv
	}

	wz, wi := out.MinimalByT["ZL3b"], out.MinimalByT["IT2a"]
	switch {
	case wz == nil && wi == nil:
		out.G1MinimalClass = "NONE"
	case wz != nil && wi != nil && wz.ModelClass == wi.ModelClass:
		out.G1MinimalClass = wz.ModelClass
	default:
		out.G1MinimalClass = "INCONCLUSIVE"
	}

	// Ladder edges.
	adequateByT := func(class string) map[string]bool {
		return map[string]bool{"ZL3b": out.ByClass[class].PredictiveAdequate && out.ByClass[class].StructuralAdequate, "IT2a": out.ByClass[class].PredictiveAdequate && out.ByClass[class].StructuralAdequate}
	}
	dlByT := func(class string) map[string]float64 {
		return map[string]float64{
			"ZL3b": descriptionLength(stageD["ZL3b"][class].Complexity.Total(), stageD["ZL3b"][class].HeldPM.PM1),
			"IT2a": descriptionLength(stageD["IT2a"][class].Complexity.Total(), stageD["IT2a"][class].HeldPM.PM1),
		}
	}
	structRegression := func(childClass, parentClass string) map[string]bool {
		out := map[string]bool{}
		for _, t := range []string{"ZL3b", "IT2a"} {
			cf := scalePassMap(stageD[t][childClass], 1.0, idx)
			pf := scalePassMap(stageD[t][parentClass], 1.0, idx)
			regressed := false
			for _, m := range append(append([]string{}, editFamilyMetrics...), lexicalFamilyMetrics...) {
				if pf[m] && !cf[m] {
					regressed = true
				}
			}
			out[t] = regressed
		}
		return out
	}
	edge := func(parent, child string) LadderEdge {
		gain := RepresentationalGain(adequateByT(child), dlByT(child), dlByT(parent), child, parent, idx, stageD["ZL3b"][child].CandidateID, structRegression(child, parent))
		return LadderEdge{Parent: parent, Child: child, Gain: gain}
	}
	e01 := edge("M0", "M1")
	e12 := edge("M1", "M2")
	e23 := edge("M2", "M3")
	e24 := edge("M2", "M4")
	out.LadderEdges = []LadderEdge{e01, e12, e23, e24}

	fsParent := "M3"
	m3dl := dlByT("M3")["ZL3b"] + dlByT("M3")["IT2a"]
	m4dl := dlByT("M4")["ZL3b"] + dlByT("M4")["IT2a"]
	m3adequate := out.ByClass["M3"].PredictiveAdequate && out.ByClass["M3"].StructuralAdequate
	m4adequate := out.ByClass["M4"].PredictiveAdequate && out.ByClass["M4"].StructuralAdequate
	skippedFS := false
	switch {
	case m3adequate && m4adequate:
		if m4dl < m3dl {
			fsParent = "M4"
		}
	case m4adequate:
		fsParent = "M4"
	case m3adequate:
		fsParent = "M3"
	default:
		skippedFS = true
		fsParent = "M2"
	}
	e5 := edge(fsParent, "M5")
	out.LadderEdges = append(out.LadderEdges, e5)
	_ = skippedFS

	// TOKEN_FORMATION_DEPTH.
	out.TokenFormationDepth = "NOT_IDENTIFIABLE"
	if out.G1MinimalClass != "INCONCLUSIVE" && out.G1MinimalClass != "NONE" {
		requiredEdgeOK := true
		switch out.G1MinimalClass {
		case "M1":
			requiredEdgeOK = e01.Gain == "SUPPORTED"
		case "M2":
			requiredEdgeOK = e12.Gain == "SUPPORTED"
		case "M3", "M4":
			requiredEdgeOK = (out.G1MinimalClass == fsParent && (e23.Gain == "SUPPORTED" || e24.Gain == "SUPPORTED"))
		case "M5":
			requiredEdgeOK = e5.Gain == "SUPPORTED"
		}
		if requiredEdgeOK {
			out.TokenFormationDepth = TokenFormationDepth(out.G1MinimalClass)
		}
	}

	// EXPLICIT_RULE_GRAMMAR_REQUIRED.
	m5Adequate := out.ByClass["M5"].PredictiveAdequate && out.ByClass["M5"].StructuralAdequate
	anyM0toM4Adequate := false
	m5BeatsAllAdequate := m5Adequate
	m5BeatsAny := false
	for _, c := range []string{"M0", "M1", "M2", "M3", "M4"} {
		if out.ByClass[c].PredictiveAdequate && out.ByClass[c].StructuralAdequate {
			anyM0toM4Adequate = true
		}
	}
	if m5Adequate {
		m5BeatsAny = e5.Gain == "SUPPORTED"
		m5BeatsAllAdequate = e5.Gain == "SUPPORTED"
	} else {
		m5BeatsAllAdequate = false
	}
	out.ExplicitRuleGrammarRequired = ExplicitRuleGrammarRequired(m5Adequate, m5BeatsAllAdequate, anyM0toM4Adequate, m5BeatsAny)

	allFamiliesClean := out.G1MinimalClass != "" && out.G1MinimalClass != "NONE" && out.G1MinimalClass != "INCONCLUSIVE"
	if allFamiliesClean {
		cs := out.ByClass[out.G1MinimalClass]
		allFamiliesClean = cs.StructuralAdequate
	}
	out.UnexplainedStructure = UnexplainedStructure(out.G1MinimalClass != "" && out.G1MinimalClass != "NONE" && out.G1MinimalClass != "INCONCLUSIVE", allFamiliesClean)

	switch {
	case out.G1MinimalClass != "" && out.G1MinimalClass != "NONE" && out.G1MinimalClass != "INCONCLUSIVE":
		out.GrammarSufficient = "SUPPORTED"
	case anyM0toM4Adequate || m5Adequate:
		out.GrammarSufficient = "PARTIAL"
	default:
		out.GrammarSufficient = "NOT_SUPPORTED"
	}
	return out
}
