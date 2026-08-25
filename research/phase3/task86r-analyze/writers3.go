package main

import (
	"encoding/json"
	"os"
	"path/filepath"
)

func writeStageDTables(stageD map[string]map[string]*StageDResult, synth SynthResult, idx *ThresholdIndex) error {
	classes := []string{"M0", "M1", "M2", "M3", "M4", "M5"}
	transcriptions := []string{"ZL3b", "IT2a"}

	pred, err := NewTSVWriter(filepath.Join(outDir, "G1_PREDICTIVE_RESULTS.tsv"),
		[]string{"transcription", "model_class", "candidate_id", "pm1", "pm2", "pm3", "pm4", "pm5", "pm6", "pm6_valid", "gate_pm1", "gate_pm2", "gate_pm4", "gate_pm5", "gate_pm6", "predictive_pass"})
	if err != nil {
		return err
	}
	comp, err := NewTSVWriter(filepath.Join(outDir, "G1_COMPLEXITY_RESULTS.tsv"),
		[]string{"transcription", "model_class", "candidate_id", "structure_cost", "lexicon_cost", "exception_cost", "complexity", "free_params", "states", "rules", "components", "dev_pm2", "held_pm2", "memorization_gap", "memorization_ratio", "memorization_dominated"})
	if err != nil {
		return err
	}
	growth, err := NewTSVWriter(filepath.Join(outDir, "G1_COMPLEXITY_GROWTH.tsv"),
		[]string{"transcription", "model_class", "candidate_id", "point_slope", "lower_ci", "points", "unbounded"})
	if err != nil {
		return err
	}
	pm5, err := NewTSVWriter(filepath.Join(outDir, "G1_PM5_RESULTS.tsv"), []string{"transcription", "model_class", "candidate_id", "pm5"})
	if err != nil {
		return err
	}
	pm6, err := NewTSVWriter(filepath.Join(outDir, "G1_PM6_RESULTS.tsv"), []string{"transcription", "model_class", "candidate_id", "pm6", "pm6_valid"})
	if err != nil {
		return err
	}
	neg, err := NewTSVWriter(filepath.Join(outDir, "G1_NEGATIVE_TOKEN_RESULTS.tsv"), []string{"transcription", "model_class", "candidate_id", "negative_exhausted"})
	if err != nil {
		return err
	}
	gen, err := NewTSVWriter(filepath.Join(outDir, "G1_GENERATION_RESULTS.tsv"),
		[]string{"transcription", "model_class", "candidate_id", "scale", "replicates", "converged", "excessive_cv"})
	if err != nil {
		return err
	}
	genStab, err := NewTSVWriter(filepath.Join(outDir, "G1_GENERATION_STABILITY.tsv"),
		[]string{"transcription", "model_class", "candidate_id", "scale", "metric", "median_generated", "cv"})
	if err != nil {
		return err
	}
	f2, err := NewTSVWriter(filepath.Join(outDir, "G1_F2_RESULTS.tsv"),
		[]string{"transcription", "model_class", "candidate_id", "metric", "median_generated_scale1", "heldout_value", "distance", "threshold", "pass"})
	if err != nil {
		return err
	}
	fam, err := NewTSVWriter(filepath.Join(outDir, "G1_FAMILY_VALIDATION.tsv"),
		[]string{"transcription", "model_class", "candidate_id", "edit_pass", "edit_count", "lexical_pass", "lexical_count", "structural_adequate"})
	if err != nil {
		return err
	}
	failLedger, err := NewTSVWriter(filepath.Join(outDir, "G1_FAILURE_LEDGER.tsv"), []string{"transcription", "model_class", "candidate_id", "failure_class"})
	if err != nil {
		return err
	}

	for _, tname := range transcriptions {
		for _, class := range classes {
			d := stageD[tname][class]
			if d.Model == nil {
				failLedger.Row(tname, class, d.CandidateID, "TRAINING_FAILED")
				continue
			}
			gz := d.MetricGates
			pred.Row(tname, class, d.CandidateID, f64(d.HeldPM.PM1), f64(d.HeldPM.PM2), f64(d.HeldPM.PM3), f64(d.HeldPM.PM4), f64(d.HeldPM.PM5), f64(d.PM6), boolStr(d.PM6Valid),
				boolStr(gz["PM1"].Pass), boolStr(gz["PM2"].Pass), boolStr(gz["PM4"].Pass), boolStr(gz["PM5"].Pass), boolStr(gz["PM6"].Pass), boolStr(d.PredictivePass))
			comp.Row(tname, class, d.CandidateID, f64(d.Complexity.StructureCost), f64(d.Complexity.LexiconCost), f64(d.Complexity.ExceptionCost), f64(d.Complexity.Total()),
				i64(d.Complexity.FreeParams), i64(d.Complexity.States), i64(d.Complexity.Rules), i64(d.Complexity.Components),
				f64(d.DevPM2), f64(d.HeldPM.PM2), f64(d.Memorization.Gap), f64(d.Memorization.Ratio), boolStr(d.Memorization.Dominated))
			growth.Row(tname, class, d.CandidateID, f64(d.ComplexityGrowth.PointSlope), f64(d.ComplexityGrowth.LowerCI), i64(d.ComplexityGrowth.Points), boolStr(d.ComplexityGrowth.Unbounded))
			pm5.Row(tname, class, d.CandidateID, f64(d.HeldPM.PM5))
			pm6.Row(tname, class, d.CandidateID, f64(d.PM6), boolStr(d.PM6Valid))
			neg.Row(tname, class, d.CandidateID, boolStr(!d.PM6Valid))
			for _, sc := range scales {
				g := d.Generation[sc]
				gen.Row(tname, class, d.CandidateID, f64(sc), i64(g.Replicates), boolStr(g.Converged), boolStr(g.ExcessiveCV))
				for _, m := range StructuralMetricIDs {
					genStab.Row(tname, class, d.CandidateID, f64(sc), m, f64(g.MedianAtStop[m]), f64(g.CV[m]))
				}
			}
			g1 := d.Generation[1.0]
			for _, m := range StructuralMetricIDs {
				medGen, okGen := g1.MedianAtStop[m]
				heldV, okHeld := d.HeldF2[m]
				pass, dist, thr := StructuralMetricPass(class, d.CandidateID, m, medGen, okGen, heldV, okHeld, idx)
				f2.Row(tname, class, d.CandidateID, m, f64(medGen), f64(heldV), f64(dist), f64(thr), boolStr(pass))
			}
			cs := synth.ByClass[class]
			fam.Row(tname, class, d.CandidateID, boolStr(cs.EditPassByT[tname]), "-", boolStr(cs.LexicalPassByT[tname]), "-", boolStr(cs.StructuralAdequate))
			for _, fc := range d.FailureClasses {
				failLedger.Row(tname, class, d.CandidateID, fc)
			}
		}
	}
	for _, w := range []*TSVWriter{pred, comp, growth, pm5, pm6, neg, gen, genStab, f2, fam, failLedger} {
		if err := w.Close(); err != nil {
			return err
		}
	}

	stab, err := NewTSVWriter(filepath.Join(outDir, "G1_TRANSCRIPTION_STABILITY.tsv"), []string{"model_class", "metric", "stability"})
	if err != nil {
		return err
	}
	for _, class := range classes {
		cs := synth.ByClass[class]
		for _, m := range predictiveMetricNames {
			stab.Row(class, m, string(cs.MetricStability[m]))
		}
		stab.Row(class, "EDIT_FAMILY", string(cs.EditFamilyStability))
		stab.Row(class, "LEXICAL_FAMILY", string(cs.LexicalFamilyStability))
	}
	if err := stab.Close(); err != nil {
		return err
	}

	ladder, err := NewTSVWriter(filepath.Join(outDir, "G1_MODEL_LADDER.tsv"), []string{"parent", "child", "representational_gain"})
	if err != nil {
		return err
	}
	for _, e := range synth.LadderEdges {
		ladder.Row(e.Parent, e.Child, e.Gain)
	}
	if err := ladder.Close(); err != nil {
		return err
	}

	selected := map[string]interface{}{
		"schema":                 "G1_SELECTED_MODELS_V1",
		"g1_minimal_class":       synth.G1MinimalClass,
		"token_formation_depth":  synth.TokenFormationDepth,
		"explicit_rule_required": synth.ExplicitRuleGrammarRequired,
	}
	byT := map[string]interface{}{}
	for _, tname := range transcriptions {
		w := synth.MinimalByT[tname]
		if w == nil {
			byT[tname] = nil
			continue
		}
		byT[tname] = map[string]interface{}{"model_class": w.ModelClass, "candidate_id": w.CandidateID, "complexity": w.Complexity}
	}
	selected["minimal_by_transcription"] = byT
	b, _ := json.MarshalIndent(selected, "", "  ")
	return os.WriteFile(filepath.Join(outDir, "G1_SELECTED_MODELS.json"), append(b, '\n'), 0o644)
}

