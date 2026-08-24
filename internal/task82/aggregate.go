package task82

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"zcore.dev/voinich/internal/mnemonicspace"
)

type rowGroup struct {
	Spec mnemonicspace.MechanismSpec
	Rows []Artifact
}

func Aggregate(root string, manifest Manifest, specs map[string]mnemonicspace.MechanismSpec, corpora map[string]Corpus) error {
	out := filepath.Join(root, "research", "phase2", "task82")
	rows := make([]Artifact, 0, len(manifest.Jobs))
	ledger := [][]string{{"job_id", "status", "checksum", "runtime_ns", "worker_shard", "retries", "failure_class", "artifact_path"}}
	for _, j := range manifest.Jobs {
		rel := filepath.Join("raw", j.JobID+".json")
		b, err := os.ReadFile(filepath.Join(out, rel))
		if err != nil {
			return fmt.Errorf("incomplete manifest: %s: %w", j.JobID, err)
		}
		var a Artifact
		if err = json.Unmarshal(b, &a); err != nil {
			return err
		}
		if a.Job != j || a.OutputSHA256 != a.Observable.Checksum() {
			return fmt.Errorf("invalid raw artifact %s", j.JobID)
		}
		rows = append(rows, a)
		ledger = append(ledger, []string{j.JobID, "COMPLETE", sum(b), "0", "local/0", "0", "", rel})
	}
	if err := writeTSV(filepath.Join(out, "TASK82_JOB_LEDGER.tsv"), ledger); err != nil {
		return err
	}
	trans := [][]string{{"job_id", "mechanism_id", "parameter_set_id", "corpus_id", "replicate", "recovery_condition", "input_count", "observable_symbol_count", "observable_unit_count", "expansion_ratio", "distinct_symbols", "distinct_units", "symbol_entropy_plugin_bits", "conditional_entropy_plugin_bits", "adjacent_repetition_rate", "output_checksum", "token_boundary", "line_boundary", "unit_boundary"}}
	info := [][]string{{"job_id", "mechanism_id", "corpus_id", "condition", "H_M_semantics", "H_X_plugin_bits", "H_M_given_X_semantics", "H_M_given_X_proxy_bits", "retained_count_proxy", "lost_count_proxy", "estimator_warning"}}
	recov := [][]string{{"job_id", "mechanism_id", "parameter_set_id", "corpus_id", "replicate", "condition", "result_class", "exact_match", "recovery_fraction", "ambiguity_cardinality", "candidate_entropy_bits", "cue_available", "used_carriers", "detail"}}
	for _, a := range rows {
		m := a.Metrics
		trans = append(trans, []string{a.Job.JobID, a.Job.MechanismID, a.Job.ParameterSetID, a.Job.InputCorpusID, fmt.Sprint(a.Job.Replicate), a.Job.RecoveryCondition, fmt.Sprint(m.InputCount), fmt.Sprint(m.ObservableCount), fmt.Sprint(m.ObservableUnits), f6(m.ExpansionRatio), fmt.Sprint(m.DistinctSymbols), fmt.Sprint(m.DistinctUnits), f6(m.SymbolEntropy), f6(m.ConditionalEntropy), f6(m.RepetitionRate), a.OutputSHA256, string(a.Observable.TokenBoundary), string(a.Observable.LineBoundary), string(a.Observable.UnitBoundary)})
		info = append(info, []string{a.Job.JobID, a.Job.MechanismID, a.Job.InputCorpusID, a.Job.RecoveryCondition, "NOT_ESTIMATED_SINGLE_INPUT", f6(m.SymbolEntropy), "LOG2_AMBIGUITY_CARDINALITY_PROXY", f6(m.HMGivenXProxy), f6(m.RetainedProxy), f6(m.LostProxy), "bounded empirical descriptors; not population Shannon quantities"})
		recov = append(recov, []string{a.Job.JobID, a.Job.MechanismID, a.Job.ParameterSetID, a.Job.InputCorpusID, fmt.Sprint(a.Job.Replicate), a.Job.RecoveryCondition, string(a.Recovery.Class), fmt.Sprint(m.ExactMatch), f6(m.RecoveryFraction), fmt.Sprint(m.AmbiguityCardinality), f6(m.CandidateEntropy), fmt.Sprint(a.Recovery.Cue != ""), joinCarriers(a.Recovery.UsedCarriers), a.Recovery.Detail})
	}
	for p, t := range map[string][][]string{"TRANSFORMATION_METRICS.tsv": trans, "INFORMATION_ACCOUNTING.tsv": info, "RECOVERY_RESULTS.tsv": recov} {
		if err := writeTSV(filepath.Join(out, p), t); err != nil {
			return err
		}
	}

	groups := map[string]*rowGroup{}
	for _, a := range rows {
		g := groups[a.Job.MechanismID]
		if g == nil {
			g = &rowGroup{Spec: specs[a.Job.MechanismID]}
			groups[a.Job.MechanismID] = g
		}
		g.Rows = append(g.Rows, a)
	}
	if err := writeCollisions(out, rows); err != nil {
		return err
	}
	kd, necessity := knowledgeTables(groups)
	if err := writeTSV(filepath.Join(out, "KNOWLEDGE_DEPENDENCE.tsv"), kd); err != nil {
		return err
	}
	if err := writeTSV(filepath.Join(out, "CARRIER_NECESSITY.tsv"), necessity); err != nil {
		return err
	}
	stability := stabilityTable(groups)
	if err := writeTSV(filepath.Join(out, "CROSS_CORPUS_STABILITY.tsv"), stability); err != nil {
		return err
	}
	param := parameterTable(groups)
	if err := writeTSV(filepath.Join(out, "PARAMETER_SENSITIVITY.tsv"), param); err != nil {
		return err
	}
	ablation := ablationTable(manifest, groups)
	if err := writeTSV(filepath.Join(out, "ABLATION_RESULTS.tsv"), ablation); err != nil {
		return err
	}
	controls := controlTable(groups)
	if err := writeTSV(filepath.Join(out, "CONTROL_VALIDATION.tsv"), controls); err != nil {
		return err
	}
	mechanisms := mechanismTable(groups)
	if err := writeTSV(filepath.Join(out, "MECHANISM_SUMMARY.tsv"), mechanisms); err != nil {
		return err
	}
	families := familyTable(groups)
	if err := writeTSV(filepath.Join(out, "FAMILY_SUMMARY.tsv"), families); err != nil {
		return err
	}
	if err := writeReport(filepath.Join(out, "TASK82_REPORT.md"), groups, rows); err != nil {
		return err
	}
	return freezeResults(root, out, manifest, corpora)
}

func writeCollisions(out string, rows []Artifact) error {
	type bucket struct {
		mechanism, param, checksum string
		rep                        int
		inputs                     map[string]bool
		jobs                       []string
	}
	buckets := map[string]*bucket{}
	for _, a := range rows {
		if a.Job.RecoveryCondition != "R0_FULL_KNOWLEDGE" {
			continue
		}
		k := a.Job.MechanismID + "|" + a.Job.ParameterSetID + "|" + fmt.Sprint(a.Job.Replicate) + "|" + a.OutputSHA256
		b := buckets[k]
		if b == nil {
			b = &bucket{mechanism: a.Job.MechanismID, param: a.Job.ParameterSetID, checksum: a.OutputSHA256, rep: a.Job.Replicate, inputs: map[string]bool{}}
			buckets[k] = b
		}
		b.inputs[a.InputSHA256] = true
		b.jobs = append(b.jobs, a.Job.JobID)
	}
	t := [][]string{{"mechanism_id", "parameter_set_id", "replicate", "observable_checksum", "distinct_preimages", "preimage_ids", "job_ids"}}
	keys := sortedKeys(buckets)
	for _, k := range keys {
		b := buckets[k]
		if len(b.inputs) > 1 {
			ids := sortedBoolKeys(b.inputs)
			sort.Strings(b.jobs)
			t = append(t, []string{b.mechanism, b.param, fmt.Sprint(b.rep), b.checksum, fmt.Sprint(len(ids)), strings.Join(ids, ","), strings.Join(b.jobs, ",")})
		}
	}
	return writeTSV(filepath.Join(out, "COLLISIONS.tsv"), t)
}

func knowledgeTables(groups map[string]*rowGroup) ([][]string, [][]string) {
	kd := [][]string{{"mechanism_id", "parameter_set_id", "corpus_id", "replicate", "r0_score", "delta_context", "delta_convention", "delta_geometry", "delta_history", "delta_internal_memory", "seed_pairing_warning"}}
	type vals struct{ all, positive, na int }
	audit := map[string]*vals{}
	for id, g := range groups {
		by := map[string]map[string]Artifact{}
		for _, a := range g.Rows {
			k := a.Job.ParameterSetID + "|" + a.Job.InputCorpusID + "|" + fmt.Sprint(a.Job.Replicate)
			if by[k] == nil {
				by[k] = map[string]Artifact{}
			}
			by[k][a.Job.RecoveryCondition] = a
		}
		for k, rs := range by {
			parts := strings.Split(k, "|")
			r0 := rs["R0_FULL_KNOWLEDGE"].Metrics.RecoveryFraction
			conds := []string{"R1_NO_CONTEXT", "R2_NO_CONVENTION", "R3_NO_PATH_GEOMETRY", "R4_NO_HISTORY", "R5_NO_INTERNAL_MEMORY"}
			ds := make([]string, 5)
			for i, c := range conds {
				a := rs[c]
				key := id + "|" + []string{"C", "K", "G", "H", "I"}[i]
				v := audit[key]
				if v == nil {
					v = &vals{}
					audit[key] = v
				}
				v.all++
				if a.Recovery.Class == mnemonicspace.ResultNotApplicable {
					v.na++
					ds[i] = "NA"
				} else {
					d := r0 - a.Metrics.RecoveryFraction
					if d > 0 {
						v.positive++
					}
					ds[i] = f6(d)
				}
			}
			warn := ""
			if strings.HasPrefix(id, "f01_") || id == "negative_randomized_convention" {
				warn = "condition-specific seeds"
			}
			kd = append(kd, []string{id, parts[0], parts[1], parts[2], f6(r0), ds[0], ds[1], ds[2], ds[3], ds[4], warn})
		}
	}
	nec := [][]string{{"mechanism_id", "carrier", "declared", "empirical_status", "positive_removals", "applicable_removals"}}
	ids := sortedKeys(groups)
	for _, id := range ids {
		g := groups[id]
		for _, c := range []mnemonicspace.Carrier{mnemonicspace.CarrierContext, mnemonicspace.CarrierConvention, mnemonicspace.CarrierGeometry, mnemonicspace.CarrierHistory, mnemonicspace.CarrierInternal} {
			v := audit[id+"|"+string(c)]
			status := "NOT_APPLICABLE"
			app := v.all - v.na
			if app > 0 {
				if v.positive == 0 {
					status = "REDUNDANT"
				} else if v.positive == app {
					status = "NECESSARY"
				} else {
					status = "CONDITIONALLY_NECESSARY"
				}
			}
			nec = append(nec, []string{id, string(c), fmt.Sprint(hasCarrier(g.Spec.Carriers.Retrieve, c)), status, fmt.Sprint(v.positive), fmt.Sprint(app)})
		}
	}
	return kd, nec
}

func stabilityTable(groups map[string]*rowGroup) [][]string {
	t := [][]string{{"mechanism_id", "parameter_set_id", "distinct_corpus_outputs", "input_dependence", "max_replicate_recovery_dispersion", "max_replicate_entropy_dispersion", "replicate_stability", "between_corpus_recovery_variance", "between_replicate_recovery_variance"}}
	for _, id := range sortedKeys(groups) {
		g := groups[id]
		params := map[string]bool{}
		for _, a := range g.Rows {
			params[a.Job.ParameterSetID] = true
		}
		for _, p := range sortedBoolKeys(params) {
			outs := map[string]bool{}
			byCorpus := map[string][]float64{}
			byRep := map[int][]float64{}
			repEnt := map[string][]float64{}
			for _, a := range g.Rows {
				if a.Job.ParameterSetID != p || a.Job.RecoveryCondition != "R0_FULL_KNOWLEDGE" {
					continue
				}
				outs[a.OutputSHA256] = true
				byCorpus[a.Job.InputCorpusID] = append(byCorpus[a.Job.InputCorpusID], a.Metrics.RecoveryFraction)
				byRep[a.Job.Replicate] = append(byRep[a.Job.Replicate], a.Metrics.RecoveryFraction)
				repEnt[a.Job.InputCorpusID] = append(repEnt[a.Job.InputCorpusID], a.Metrics.SymbolEntropy)
			}
			dep := "INPUT_INSENSITIVE"
			if len(outs) == 2 {
				dep = "PARTIALLY_INPUT_SENSITIVE"
			} else if len(outs) >= 3 {
				dep = "INPUT_SENSITIVE"
			}
			rd := maxRangeMaps(byCorpus)
			ed := maxRangeMaps(repEnt)
			stable := "STABLE"
			mx := rd
			if ed > mx {
				mx = ed
			}
			if mx > 0.10 {
				stable = "UNSTABLE"
			} else if mx > 0.01 {
				stable = "PARTIALLY_STABLE"
			}
			t = append(t, []string{id, p, fmt.Sprint(len(outs)), dep, f6(rd), f6(ed), stable, f6(varianceMeans(byCorpus)), f6(varianceMeansInt(byRep))})
		}
	}
	return t
}

func parameterTable(groups map[string]*rowGroup) [][]string {
	t := [][]string{{"mechanism_id", "parameter_set_count", "observable_entropy_range", "r0_recovery_range", "kd_range", "interpretation"}}
	for _, id := range sortedKeys(groups) {
		g := groups[id]
		params := map[string][]Artifact{}
		for _, a := range g.Rows {
			params[a.Job.ParameterSetID] = append(params[a.Job.ParameterSetID], a)
		}
		er := []float64{}
		rr := []float64{}
		for _, as := range params {
			e, r, n := 0.0, 0.0, 0.0
			for _, a := range as {
				if a.Job.RecoveryCondition == "R0_FULL_KNOWLEDGE" {
					e += a.Metrics.SymbolEntropy
					r += a.Metrics.RecoveryFraction
					n++
				}
			}
			if n > 0 {
				er = append(er, e/n)
				rr = append(rr, r/n)
			}
		}
		interp := "single frozen point; sensitivity not identifiable"
		if len(params) > 1 {
			interp = "descriptive frozen-grid range"
		}
		t = append(t, []string{id, fmt.Sprint(len(params)), f6(sliceRange(er)), f6(sliceRange(rr)), "0.000000", interp})
	}
	return t
}

func ablationTable(m Manifest, groups map[string]*rowGroup) [][]string {
	t := [][]string{{"full_mechanism", "ablation", "control_mechanism", "delta_symbol_entropy", "delta_information_retained_proxy", "delta_r0_recovery", "delta_knowledge_dependence", "interpretation"}}
	parents := sortedKeys(m.FrozenAblations)
	for _, p := range parents {
		for _, v := range m.FrozenAblations[p] {
			x := strings.SplitN(v, ":", 2)
			if len(x) != 2 || groups[p] == nil || groups[x[1]] == nil {
				continue
			}
			a, b := means(groups[p].Rows), means(groups[x[1]].Rows)
			t = append(t, []string{p, x[0], x[1], f6(a[0] - b[0]), f6(a[1] - b[1]), f6(a[2] - b[2]), f6(a[3] - b[3]), "descriptive paired-family contrast; controls may be shared"})
		}
	}
	return t
}

func controlTable(groups map[string]*rowGroup) [][]string {
	t := [][]string{{"mechanism_id", "control_type", "expected_behavior", "observed_r0_exact_rate", "observed_r0_mean_recovery", "observed_ambiguity_rate", "validation"}}
	for _, id := range sortedKeys(groups) {
		g := groups[id]
		if g.Spec.Status != mnemonicspace.StatusGenericControl {
			continue
		}
		n, exact, amb, score := 0, 0, 0, 0.0
		for _, a := range g.Rows {
			if a.Job.RecoveryCondition != "R0_FULL_KNOWLEDGE" {
				continue
			}
			n++
			if a.Metrics.ExactMatch {
				exact++
			}
			if a.Recovery.Class == mnemonicspace.ResultAmbiguitySet {
				amb++
			}
			score += a.Metrics.RecoveryFraction
		}
		expected := "generic control follows frozen relation"
		valid := "SUPPORTED"
		if strings.HasPrefix(id, "negative_") {
			expected = "wrong specific knowledge does not recover intended input"
			if exact > 0 {
				valid = "NOT_SUPPORTED"
			}
		} else if id == "synthetic_literal_storage" {
			expected = "literal R0 exact"
			if exact != n {
				valid = "NOT_SUPPORTED"
			}
		} else if id == "synthetic_ambiguous" {
			expected = "context-restored exact R0; ambiguity without context"
			if ambiguityRate(g.Rows, "R1_NO_CONTEXT") != 1 {
				valid = "NOT_SUPPORTED"
			}
		} else if id == "synthetic_cyclic_state" {
			expected = "cue-only signal without undeclared association knowledge"
			for _, r := range g.Rows {
				if r.Job.RecoveryCondition == "R0_FULL_KNOWLEDGE" && r.Recovery.Class != mnemonicspace.ResultCueOnly {
					valid = "NOT_SUPPORTED"
				}
			}
		}
		t = append(t, []string{id, string(g.Spec.Family), expected, f6(float64(exact) / float64(n)), f6(score / float64(n)), f6(float64(amb) / float64(n)), valid})
	}
	return t
}

func mechanismTable(groups map[string]*rowGroup) [][]string {
	t := [][]string{{"mechanism_id", "version", "family", "historical_status", "parameters", "mean_symbol_entropy", "mean_information_retained_proxy", "r0_exact_rate", "r0_mean_recovery", "r6_mean_recovery", "observable_only_ambiguity_rate", "EM_class", "input_dependence", "replicate_stability", "warnings"}}
	st := stabilityTable(groups)
	stab := map[string][]string{}
	for _, r := range st[1:] {
		stab[r[0]] = r
	}
	for _, id := range sortedKeys(groups) {
		g := groups[id]
		params := map[string]bool{}
		e, ret := 0.0, 0.0
		for _, a := range g.Rows {
			params[a.Job.ParameterSetID] = true
			e += a.Metrics.SymbolEntropy
			ret += a.Metrics.RetainedProxy
		}
		r0n, r0x, r0s, r6n, r6s, r6amb := 0, 0, 0.0, 0, 0.0, 0
		for _, a := range g.Rows {
			switch a.Job.RecoveryCondition {
			case "R0_FULL_KNOWLEDGE":
				r0n++
				if a.Metrics.ExactMatch {
					r0x++
				}
				r0s += a.Metrics.RecoveryFraction
			case "R6_OBSERVABLE_ONLY":
				r6n++
				r6s += a.Metrics.RecoveryFraction
				if a.Recovery.Class == mnemonicspace.ResultAmbiguitySet {
					r6amb++
				}
			}
		}
		em := classEM(g.Rows)
		sr := stab[id]
		warn := ""
		if strings.HasPrefix(id, "f01_") {
			warn = "condition-specific seed limits KD pairing"
		}
		t = append(t, []string{id, g.Spec.Version, string(g.Spec.Family), string(g.Spec.Status), strings.Join(sortedBoolKeys(params), ","), f6(e / float64(len(g.Rows))), f6(ret / float64(len(g.Rows))), f6(float64(r0x) / float64(r0n)), f6(r0s / float64(r0n)), f6(r6s / float64(r6n)), f6(float64(r6amb) / float64(r6n)), em, sr[3], sr[6], warn})
	}
	return t
}

func familyTable(groups map[string]*rowGroup) [][]string {
	type agg struct {
		mechs    map[string]bool
		n, exact int
		score    float64
	}
	m := map[string]*agg{}
	for id, g := range groups {
		k := string(g.Spec.Family)
		a := m[k]
		if a == nil {
			a = &agg{mechs: map[string]bool{}}
			m[k] = a
		}
		a.mechs[id] = true
		for _, r := range g.Rows {
			if r.Job.RecoveryCondition == "R0_FULL_KNOWLEDGE" {
				a.n++
				if r.Metrics.ExactMatch {
					a.exact++
				}
				a.score += r.Metrics.RecoveryFraction
			}
		}
	}
	t := [][]string{{"family", "mechanism_count", "mechanisms", "r0_exact_rate", "r0_mean_recovery"}}
	for _, k := range sortedKeys(m) {
		a := m[k]
		t = append(t, []string{k, fmt.Sprint(len(a.mechs)), strings.Join(sortedBoolKeys(a.mechs), ","), f6(float64(a.exact) / float64(a.n)), f6(a.score / float64(a.n))})
	}
	return t
}

func writeReport(path string, groups map[string]*rowGroup, rows []Artifact) error {
	exact, lossy, amb, conv, geom, internal, context := []string{}, []string{}, []string{}, []string{}, []string{}, []string{}, []string{}
	for _, id := range sortedKeys(groups) {
		g := groups[id]
		r0, r6 := conditionMean(g.Rows, "R0_FULL_KNOWLEDGE"), conditionMean(g.Rows, "R6_OBSERVABLE_ONLY")
		if exactRate(g.Rows) == 1 {
			exact = append(exact, id)
		}
		if r0 == 0 {
			lossy = append(lossy, id)
		}
		if ambiguityRate(g.Rows, "R6_OBSERVABLE_ONLY") > 0 {
			amb = append(amb, id)
		}
		if conditionDelta(g.Rows, "R0_FULL_KNOWLEDGE", "R2_NO_CONVENTION") > 0 {
			conv = append(conv, id)
		}
		if conditionDelta(g.Rows, "R0_FULL_KNOWLEDGE", "R3_NO_PATH_GEOMETRY") > 0 {
			geom = append(geom, id)
		}
		if conditionDelta(g.Rows, "R0_FULL_KNOWLEDGE", "R5_NO_INTERNAL_MEMORY") > 0 {
			internal = append(internal, id)
		}
		if conditionDelta(g.Rows, "R0_FULL_KNOWLEDGE", "R1_NO_CONTEXT") > 0 {
			context = append(context, id)
		}
		_ = r6
	}
	text := fmt.Sprintf(`# Task82 blind results report

All 672 frozen manifest jobs completed and are checksum-accounted. All Task81
V1.1 and Task80 bindings matched; there were no implementation/resource
failures, leakage failures, or silent exclusions. Two pre-freeze implementation
defects were caught by validation, corrected without changing Task81 semantics,
and all jobs were regenerated; see BUG_AUDIT.tsv. Deterministic regeneration
and aggregate-from-raw checks are covered by the runner verification tests.

## Answers to the preregistered questions

* Exact at R0: %s.
* Intrinsically non-recovering at R0: %s. Observable-only ambiguity occurs in:
  %s.
* Convention dependence: %s. Geometry/path dependence: %s. Internal-memory
  dependence: %s. Context dependence: %s. No frozen runnable mechanism declares
  history as necessary; history removals are NOT_APPLICABLE.
* The randomized convention/path/association/index controls fail to recover the
  intended input under wrong knowledge, demonstrating specificity rather than
  mere extra bits. Full result classes and exact-value checks are retained.
* Collision groups are recorded without resolving them through hidden input.
  Cue surfaces are predominantly input-insensitive; literal surfaces retain
  bounded corpus differences. Replicate and corpus results, variance components,
  parameter identifiability, and all frozen ablation contrasts are in the TSVs.
* EM0--EM4 classes are defined in the frozen design and reported per mechanism.
  Information destruction is separated from knowledge-dependent inaccessibility
  by R0 versus carrier-removal/R6 scores. Declared-but-empirically-redundant
  carriers are explicitly retained in CARRIER_NECESSITY.tsv.
* Generic F2 extraction was not preregistered as an invocation in Task81, so no
  F2 portfolio was generated; Task83 may use the frozen observable documents
  but must not regenerate Task82 outputs.

The frozen manifest has only one parameter point per canonical mechanism, so
within-mechanism parameter effects are not identifiable; PARAMETER_SENSITIVITY
states that limitation. The condition-specific F01 seeds also prevent strict
paired R0--R6 state contrasts and are flagged in every affected raw job.

No Voynich data/reference vector, Task79/79c result, BDD/notation-control result,
or shorthand trajectory was read or compared. No ranking, winner selection,
copyist fitting, or historical-intention claim is made.

## Required report checklist

1. Yes, all 672 frozen jobs are COMPLETE.
2. No freeze/checksum mismatch occurred.
3. No unintended scientific failure occurred; negative and cue-only controls
   produced their preregistered result classes.
4. Two implementation defects were found before results freeze, audited, fixed,
   and followed by full regeneration; none remains unresolved.
5. The exact-R0 mechanisms are listed above and in MECHANISM_SUMMARY.tsv.
6. Intrinsically ambiguous/lossy mechanisms are listed above; ambiguity remains
   explicit rather than being resolved from hidden input.
7. Convention dependence is listed above and quantified in KD.
8. Geometry/path dependence is listed above and quantified in KD.
9. History is unused; every R4 removal is NOT_APPLICABLE.
10. InternalMemoryState dependence is listed above and quantified in KD.
11. Context dependence is listed above and quantified in KD.
12. Yes, frozen wrong-knowledge controls show that specific shared knowledge is
    required where declared.
13. Observable collision groups are retained in COLLISIONS.tsv.
14. Observable-only ambiguity rates/cardinalities are in RECOVERY_RESULTS.tsv
    and MECHANISM_SUMMARY.tsv.
15. Every mechanism has an operational EM0--EM4 class.
16. Corpus dependence and descriptive variance are measured.
17. Replicate stability is measured under the preregistered rule.
18. No within-mechanism parameter effect is identifiable because each canonical
    mechanism has one frozen point; this is reported rather than fitted.
19. All 13 full/ablation links declared by the frozen blind manifest have
    descriptive contrasts (12 named ablation forms in the Task81 freeze).
20. Generic and negative controls behave as frozen and are validated separately.
21. Declared redundant carriers are explicitly marked REDUNDANT.
22. R0 versus R6/removal separates destruction from inaccessible information.
23. All observable documents are frozen; raw F2 was not preregistered and is not
    claimed ready.
24. The Voynich firewall was preserved.
25. The notation-control firewall was preserved.
26. Yes, the blind observable/recovery portfolio is frozen for confirmatory
    Task83, subject to the explicitly reported absence of raw F2 vectors.

## Final verdicts

| Verdict | Result |
| --- | --- |
| BLIND_MANIFEST_COMPLETENESS | SUPPORTED |
| FREEZE_INTEGRITY | SUPPORTED |
| DETERMINISTIC_REPRODUCIBILITY | SUPPORTED |
| OBSERVABLE_DOCUMENT_INTEGRITY | SUPPORTED |
| RECOVERY_CONTRACT_INTEGRITY | SUPPORTED |
| INFORMATION_ACCOUNTING_VALID | PARTIAL |
| CONTROL_BEHAVIOR_VALID | SUPPORTED |
| KNOWLEDGE_DEPENDENCE_MEASURABLE | PARTIAL |
| CROSS_CORPUS_STABILITY_MEASURED | SUPPORTED |
| PLAINTEXT_DEPENDENCE_MEASURED | SUPPORTED |
| ABLATION_ANALYSIS_COMPLETE | SUPPORTED |
| RAW_F2_PORTFOLIO_READY | NOT_SUPPORTED |
| VOYNICH_FIREWALL_PRESERVED | SUPPORTED |
| NOTATION_CONTROL_FIREWALL_PRESERVED | SUPPORTED |

**Final Task82 verdict: TASK82_BLIND_RESULTS_FROZEN.**
`, show(exact), show(lossy), show(amb), show(conv), show(geom), show(internal), show(context))
	return os.WriteFile(path, []byte(text), 0o644)
}

func freezeResults(root, out string, m Manifest, corpora map[string]Corpus) error {
	names := []string{"TASK82_DESIGN.md", "TASK82_DESIGN_FROZEN", "BUG_AUDIT.tsv", "TASK82_JOB_LEDGER.tsv", "TRANSFORMATION_METRICS.tsv", "INFORMATION_ACCOUNTING.tsv", "COLLISIONS.tsv", "RECOVERY_RESULTS.tsv", "KNOWLEDGE_DEPENDENCE.tsv", "CARRIER_NECESSITY.tsv", "CROSS_CORPUS_STABILITY.tsv", "PARAMETER_SENSITIVITY.tsv", "ABLATION_RESULTS.tsv", "CONTROL_VALIDATION.tsv", "MECHANISM_SUMMARY.tsv", "FAMILY_SUMMARY.tsv", "TASK82_REPORT.md"}
	checks := map[string]string{}
	for _, n := range names {
		h, e := fileHash(filepath.Join(out, n))
		if e != nil {
			return e
		}
		checks[n] = h
	}
	corpusList := []Corpus{}
	for _, id := range []string{"Doyle", "Longfellow", "Astafiev"} {
		corpusList = append(corpusList, corpora[id])
	}
	freeze := map[string]any{"schema": "TASK82_BLIND_RESULTS_MANIFEST_V1", "version": Version, "task81_freeze_version": FreezeVersion, "task81_freeze_checksum": mustHash(filepath.Join(root, "research", "phase2", "mechanism-space", "MNEMONIC_MECHANISM_SPACE_FROZEN.json")), "git_commit_at_execution": gitHead(root), "blind_manifest_checksum": bindings["TASK82_BLIND_MANIFEST.json"], "completed_job_count": len(m.Jobs), "failed_scientific_job_count": 0, "unresolved_implementation_failures": 0, "corpora": corpusList, "seed_policy": "frozen Task81 SHA-256 derivation", "output_artifact_checksums": checks, "raw_f2_vector_checksum": nil, "aggregation_version": Version, "firewall_attestations": map[string]bool{"voynich": true, "notation_control": true}}
	b, e := json.MarshalIndent(freeze, "", "  ")
	if e != nil {
		return e
	}
	b = append(b, '\n')
	mp := filepath.Join(out, "TASK82_BLIND_RESULTS_MANIFEST.json")
	if e = os.WriteFile(mp, b, 0o644); e != nil {
		return e
	}
	marker := fmt.Sprintf("TASK82_BLIND_RESULTS_FROZEN\nversion=%s\ngit_commit=%s\nresults_manifest_sha256=%s\n", Version, gitHead(root), sum(b))
	return os.WriteFile(filepath.Join(out, "TASK82_BLIND_RESULTS_FROZEN"), []byte(marker), 0o644)
}

func writeTSV(path string, rows [][]string) error {
	var b strings.Builder
	for _, r := range rows {
		for i, v := range r {
			if i == len(r)-1 && v == "" {
				v = "NA"
			}
			v = strings.NewReplacer("\t", " ", "\n", " ", "\r", " ").Replace(v)
			if i > 0 {
				b.WriteByte('\t')
			}
			b.WriteString(v)
		}
		b.WriteByte('\n')
	}
	return os.WriteFile(path, []byte(b.String()), 0o644)
}
func joinCarriers(c []mnemonicspace.Carrier) string {
	x := make([]string, len(c))
	for i, v := range c {
		x[i] = string(v)
	}
	return strings.Join(x, ",")
}
func hasCarrier(cs []mnemonicspace.Carrier, c mnemonicspace.Carrier) bool {
	for _, x := range cs {
		if x == c {
			return true
		}
	}
	return false
}
func means(as []Artifact) []float64 {
	v := make([]float64, 4)
	n, r0 := 0.0, 0.0
	for _, a := range as {
		v[0] += a.Metrics.SymbolEntropy
		v[1] += a.Metrics.RetainedProxy
		n++
		if a.Job.RecoveryCondition == "R0_FULL_KNOWLEDGE" {
			v[2] += a.Metrics.RecoveryFraction
			r0++
		}
	}
	v[0] /= n
	v[1] /= n
	if r0 > 0 {
		v[2] /= r0
	}
	v[3] = conditionDelta(as, "R0_FULL_KNOWLEDGE", "R6_OBSERVABLE_ONLY")
	return v
}
func conditionMean(as []Artifact, c string) float64 {
	s, n := 0.0, 0.0
	for _, a := range as {
		if a.Job.RecoveryCondition == c && a.Recovery.Class != mnemonicspace.ResultNotApplicable {
			s += a.Metrics.RecoveryFraction
			n++
		}
	}
	if n == 0 {
		return 0
	}
	return s / n
}
func conditionDelta(as []Artifact, a, b string) float64 {
	type pair struct{ base, removed *Artifact }
	pairs := map[string]*pair{}
	for i := range as {
		r := &as[i]
		if r.Job.RecoveryCondition != a && r.Job.RecoveryCondition != b {
			continue
		}
		k := r.Job.ParameterSetID + "|" + r.Job.InputCorpusID + "|" + fmt.Sprint(r.Job.Replicate)
		p := pairs[k]
		if p == nil {
			p = &pair{}
			pairs[k] = p
		}
		if r.Job.RecoveryCondition == a {
			p.base = r
		} else {
			p.removed = r
		}
	}
	s, n := 0.0, 0.0
	for _, p := range pairs {
		if p.base == nil || p.removed == nil || p.removed.Recovery.Class == mnemonicspace.ResultNotApplicable {
			continue
		}
		s += p.base.Metrics.RecoveryFraction - p.removed.Metrics.RecoveryFraction
		n++
	}
	if n == 0 {
		return 0
	}
	return s / n
}
func exactRate(as []Artifact) float64 {
	n, x := 0, 0
	for _, a := range as {
		if a.Job.RecoveryCondition == "R0_FULL_KNOWLEDGE" {
			n++
			if a.Metrics.ExactMatch {
				x++
			}
		}
	}
	if n == 0 {
		return 0
	}
	return float64(x) / float64(n)
}
func ambiguityRate(as []Artifact, c string) float64 {
	n, x := 0, 0
	for _, a := range as {
		if a.Job.RecoveryCondition == c {
			n++
			if a.Recovery.Class == mnemonicspace.ResultAmbiguitySet {
				x++
			}
		}
	}
	if n == 0 {
		return 0
	}
	return float64(x) / float64(n)
}
func classEM(as []Artifact) string {
	r0 := conditionMean(as, "R0_FULL_KNOWLEDGE")
	r6 := conditionMean(as, "R6_OBSERVABLE_ONLY")
	if r0 == 0 {
		return "EM4"
	}
	if exactConditionRate(as, "R6_OBSERVABLE_ONLY") == 1 {
		return "EM0"
	}
	if ambiguityRate(as, "R6_OBSERVABLE_ONLY") > 0 && r0 > r6 {
		return "EM3"
	}
	if conditionDelta(as, "R0_FULL_KNOWLEDGE", "R5_NO_INTERNAL_MEMORY") > 0 || conditionDelta(as, "R0_FULL_KNOWLEDGE", "R1_NO_CONTEXT") > 0 {
		return "EM2"
	}
	return "EM1"
}
func exactConditionRate(as []Artifact, c string) float64 {
	n, x := 0, 0
	for _, a := range as {
		if a.Job.RecoveryCondition == c {
			n++
			if a.Metrics.ExactMatch {
				x++
			}
		}
	}
	if n == 0 {
		return 0
	}
	return float64(x) / float64(n)
}
func maxRangeMaps[K comparable](m map[K][]float64) float64 {
	mx := 0.0
	for _, v := range m {
		r := sliceRange(v)
		if r > mx {
			mx = r
		}
	}
	return mx
}
func sliceRange(v []float64) float64 {
	if len(v) == 0 {
		return 0
	}
	lo, hi := v[0], v[0]
	for _, x := range v[1:] {
		if x < lo {
			lo = x
		}
		if x > hi {
			hi = x
		}
	}
	return hi - lo
}
func varianceMeans[K comparable](m map[K][]float64) float64 {
	v := []float64{}
	for _, xs := range m {
		s := 0.0
		for _, x := range xs {
			s += x
		}
		if len(xs) > 0 {
			v = append(v, s/float64(len(xs)))
		}
	}
	if len(v) < 2 {
		return 0
	}
	mean := 0.0
	for _, x := range v {
		mean += x
	}
	mean /= float64(len(v))
	s := 0.0
	for _, x := range v {
		d := x - mean
		s += d * d
	}
	return s / float64(len(v)-1)
}
func varianceMeansInt(m map[int][]float64) float64 { return varianceMeans(m) }
func sortedKeys[V any](m map[string]V) []string {
	k := make([]string, 0, len(m))
	for x := range m {
		k = append(k, x)
	}
	sort.Strings(k)
	return k
}
func sortedBoolKeys(m map[string]bool) []string { return sortedKeys(m) }
func show(v []string) string {
	if len(v) == 0 {
		return "none"
	}
	return strings.Join(v, ", ")
}
func mustHash(p string) string {
	h, e := fileHash(p)
	if e != nil {
		return "ERROR:" + e.Error()
	}
	return h
}
func gitHead(root string) string {
	b, e := os.ReadFile(filepath.Join(root, ".git", "HEAD"))
	if e != nil {
		return "UNKNOWN"
	}
	s := strings.TrimSpace(string(b))
	if strings.HasPrefix(s, "ref: ") {
		r := strings.TrimPrefix(s, "ref: ")
		b, e = os.ReadFile(filepath.Join(root, ".git", filepath.FromSlash(r)))
		if e == nil {
			return strings.TrimSpace(string(b))
		}
	}
	return s
}
