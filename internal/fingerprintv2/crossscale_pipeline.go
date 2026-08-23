package fingerprintv2

import (
	"math/rand"
)

// runEditGraphValidation implements task77 stage 2: it audits the
// productive family graph already computed by LP3/EF1-EF4 for transitive
// merging, hub dependence and partition stability, then only builds a
// consensus family catalog when that battery shows enough stability.
func runEditGraphValidation(c corpus, familyGraph editGraph, families [][]string, glyphs map[string][]string, cfg Config, freq map[string]int, seed int64) EditFamilyValidation {
	merge := transitiveMergeAudit(familyGraph, families, glyphs, c, freq, rand.New(rand.NewSource(seed+1)))
	stability, replicates := runFamilyStabilityBattery(c, cfg, familyGraph, families, seed+2)
	totalTokens, totalTypes := len(c.records), len(vocabulary(c))
	validation := buildConsensusFamilies(familyGraph, families, stability, replicates, freq, totalTokens, totalTypes)
	validation.TransitiveMerge = merge
	return validation
}

// runCrossScale implements task77 stages 4-10 over a metadata-bearing
// corpus and its already-computed productive family graph.
func runCrossScale(c corpus, cfg Config, familyGraph editGraph, families [][]string, glyphs map[string][]string, freq map[string]int, grammarRuns []GrammarRun, seed int64) CrossScaleResult {
	csCfg := cfg.CrossScale
	if csCfg == nil {
		normalized := (CrossScaleConfig{}).normalized()
		csCfg = &normalized
	}
	familyOf, roleOf := assignFamiliesAndRoles(familyGraph, families)
	zoneOf := zoneProfileLabels(familyGraph, glyphs)
	variables := crossScaleVariables(c)

	metrics := make([]CrossScaleMetric, 0, 12)
	toAdjust := make([]int, 0, 12) // indices into metrics that carry a real NullTest and need FDR
	tests := make([]NullTest, 0, 12)

	addTest := func(idx int, t NullTest) {
		toAdjust = append(toAdjust, idx)
		tests = append(tests, t)
	}

	// CS1
	rng := rand.New(rand.NewSource(seed + 100))
	famTest, roleTest, n1, diag := cs1Test(c, familyOf, roleOf, cfg.Repetitions, rng)
	if n1 >= 20 {
		m := csMetric("cs1/family-line-position", "P(edit family or family role | line position) != P(edit family or family role)", "occurrence (family-bearing, line length>=2)",
			[]string{"edit_family", "family_role", "position_class"}, nil, []string{"line length", "token frequency"},
			famTest, n1, "A significant result means family membership is not uniformly distributed across line-initial/interior/final slots.",
			"Tests family and role membership jointly with line position; does not decompose by individual transformation direction (deferred, see report).", "see partition_stability")
		metrics = append(metrics, m)
		addTest(len(metrics)-1, famTest)
		mr := csMetric("cs1/role-line-position", "P(family role | line position) != P(family role)", "occurrence (family-bearing, line length>=2)",
			[]string{"family_role", "position_class"}, nil, []string{"line length", "token frequency"},
			roleTest, n1, "A significant result means CORE vs PERIPHERY status is not uniformly distributed across line position.",
			"Role is a k-core-based structural label, not a linguistic annotation.", "see partition_stability")
		metrics = append(metrics, mr)
		addTest(len(metrics)-1, roleTest)
		heldOut := groupedKFoldLogLoss(c.records,
			func(r tokenRecord) string { return r.Page },
			func(r tokenRecord) (string, bool) {
				if _, ok := familyOf[r.Token]; !ok {
					return "", false
				}
				return familyLabelOf(familyOf, r.Token), true
			},
			func(r tokenRecord) (string, bool) { return positionClass(r) },
			csCfg.Folds, seed+101)
		metrics[len(metrics)-2].HeldOutPerformance = &heldOut
		metrics[len(metrics)-2].PartitionStability = cs8Test(c, familyOf, cfg.Repetitions, rand.New(rand.NewSource(seed+102)))
		metrics[len(metrics)-2].AdditionalNulls = []NullTest{cs1FamilyLabelPermutationNull(c, familyOf, cfg.Repetitions, rand.New(rand.NewSource(seed+103)))}
		_ = diag
	} else {
		metrics = append(metrics, csInconclusive("cs1/family-line-position", "P(edit family | line position) != P(edit family)", "insufficient family-bearing occurrences with line-position metadata (need line metadata from ivtff_path)", n1))
	}

	// CS2
	famAdjTest, zoneTest, n2 := cs2Test(c, familyOf, zoneOf, cfg.Repetitions, rand.New(rand.NewSource(seed+200)))
	if n2 >= 20 {
		m := csMetric("cs2/prev-family-current-family", "P(current family | previous-token family) != P(current family)", "adjacent occurrence pair (current family-bearing)",
			[]string{"previous_token_family", "current_token_family"}, nil, []string{"corpus-wide family-size distribution"},
			famAdjTest, n2, "A significant result means adjacent tokens cluster by family more than a within-line reordering predicts.",
			"Operationalizes 'local context predicts variant/family member' as adjacent-occurrence family co-membership; does not separately test transformation direction.", "see cs5 stratified rate")
		metrics = append(metrics, m)
		addTest(len(metrics)-1, famAdjTest)
		mz := csMetric("cs2/prev-family-current-zone", "P(current transformation zone profile | previous-token family) != P(zone profile)", "adjacent occurrence pair (current family-bearing)",
			[]string{"previous_token_family", "transformation_zone_profile"}, nil, []string{"corpus-wide zone-profile distribution"},
			zoneTest, n2, "A significant result means local context predicts whether the current token carries a prefix/suffix/internal edit relationship.",
			"Zone profile is derived from the productive family graph only, not from a full generative cipher model.", "not separately assessed")
		metrics = append(metrics, mz)
		addTest(len(metrics)-1, zoneTest)
	} else {
		metrics = append(metrics, csInconclusive("cs2/prev-family-current-family", "P(current family | previous-token family) != P(current family)", "fewer than 20 family-bearing adjacent occurrence pairs", n2))
	}

	// CS3
	cs3, n3, ok3 := cs3Test(c, familyOf, cfg.Repetitions, rand.New(rand.NewSource(seed+300)))
	if ok3 {
		m := csMetric("cs3/family-locus-type", "P(edit family | locus type) != P(edit family)", "occurrence (locus-type metadata available)",
			[]string{"edit_family", "locus_type"}, nil, []string{"per-folio locus-type composition"},
			cs3, n3, "A significant result separates label loci from continuous-text loci in family composition.",
			"Locus-type codes beyond TEXT/LABEL are pooled into SPECIAL for statistical power; see report for the individual code counts.", "not separately assessed")
		metrics = append(metrics, m)
		addTest(len(metrics)-1, cs3)
	} else {
		metrics = append(metrics, csNotApplicable("cs3/family-locus-type", "P(edit family | locus type) != P(edit family)", "requires ivtff_path locus-type metadata; unavailable or below the 20-occurrence floor for this corpus"))
	}

	// CS4: EF5 (same-folio/regime concentration, computed once in Metrics.EF5) + Currier/Section family association
	csCurrier, nCurrier, okCurrier := cs4MetadataTest(c, familyOf, func(r tokenRecord) (string, bool) { return r.Currier, r.Currier != "" }, "cs4/family-currier", "folio-level Currier-label permutation", cfg.Repetitions, rand.New(rand.NewSource(seed+400)))
	if okCurrier {
		m := csMetric("cs4/family-currier", "P(edit family | Currier language) != P(edit family)", "occurrence (family-bearing, Currier recorded)",
			[]string{"edit_family", "currier"}, nil, []string{"folio count per Currier partition"},
			csCurrier, nCurrier, "A significant result means specific families concentrate in Currier A or B rather than spreading evenly.",
			"Currier B and later hands are a much smaller partition (task65), limiting power; restricted to family-bearing occurrences.", "see stage 9 Currier-stratified runs")
		metrics = append(metrics, m)
		addTest(len(metrics)-1, csCurrier)
	} else {
		metrics = append(metrics, csNotApplicable("cs4/family-currier", "P(edit family | Currier language) != P(edit family)", "requires ivtff_path Currier metadata and >=4 folios; unavailable or below the 20-occurrence floor"))
	}
	csSection, nSection, okSection := cs4MetadataTest(c, familyOf, func(r tokenRecord) (string, bool) { return r.Section, r.Section != "" }, "cs4/family-section", "folio-level section-label permutation", cfg.Repetitions, rand.New(rand.NewSource(seed+410)))
	if okSection {
		m := csMetric("cs4/family-section", "P(edit family | section/$I illustration code) != P(edit family)", "occurrence (family-bearing, section recorded)",
			[]string{"edit_family", "section"}, nil, []string{"folio count per section"},
			csSection, nSection, "A significant result means specific families concentrate in a specific illustration-type section.",
			"Section sizes are very uneven (task65 §25-28 precedent); restricted to family-bearing occurrences.", "see stage 9 per-section runs")
		metrics = append(metrics, m)
		addTest(len(metrics)-1, csSection)
	} else {
		metrics = append(metrics, csNotApplicable("cs4/family-section", "P(edit family | section) != P(edit family)", "requires ivtff_path section metadata and >=4 folios; unavailable or below the 20-occurrence floor"))
	}

	// CS5
	cs5, n5, rates5, ok5 := cs5Test(c, familyOf, cfg.Repetitions, rand.New(rand.NewSource(seed+500)))
	if ok5 {
		_ = rates5
		m := csMetric("cs5/local-adjacency-x-regime", "the same-family adjacency rate (CS2) differs across Currier x section regimes more than a folio-level regime-label permutation predicts", "adjacent occurrence pair, stratified by folio regime",
			[]string{"same_family_adjacency", "currier", "section"}, nil, []string{"regime-stratum sample size imbalance"},
			cs5, n5, "A significant result is a direct interaction: local sequence structure itself changes across regimes, not just regime base rates.",
			"Uses the max-minus-min rate across strata as the interaction statistic; only strata with >=5 pairs are compared.", "not separately assessed")
		metrics = append(metrics, m)
		addTest(len(metrics)-1, cs5)
	} else {
		metrics = append(metrics, csNotApplicable("cs5/local-adjacency-x-regime", "same-family adjacency rate differs across regimes", "requires >=2 sufficiently sampled Currier x section strata"))
	}

	// CS6
	cs6, n6, ok6 := cs6Test(c, familyOf, cfg.Repetitions, rand.New(rand.NewSource(seed+600)))
	if ok6 {
		m := csMetric("cs6/family-diversity-x-line-length", "line-level family-composition entropy correlates with line length beyond a line-length-preserving global token shuffle", "line",
			[]string{"family_diversity_entropy", "line_length"}, nil, []string{"corpus-wide family-size distribution"},
			cs6, n6, "A significant result means longer lines are not simply proportionally more diverse than the global shuffle predicts.",
			"Family composition is measured over {NONE plus each family}, so most of a typical line's entropy budget reflects the large NONE category.", "not separately assessed")
		metrics = append(metrics, m)
		addTest(len(metrics)-1, cs6)
	} else {
		metrics = append(metrics, csNotApplicable("cs6/family-diversity-x-line-length", "family-composition entropy correlates with line length", "requires ivtff_path line metadata and >=10 lines"))
	}

	// CS7
	cs7, n7, ok7 := cs7Test(c, glyphs, freq, csCfg.StructuralSample, cfg.Repetitions, rand.New(rand.NewSource(seed+700)))
	if ok7 {
		m := csMetric("cs7/edit-distance-x-structural-distance", "raw glyph edit distance between two vocabulary types correlates with the minimum distance (in lines) between their occurrences, beyond a frequency-bin-matched permutation", "sampled vocabulary type pair",
			[]string{"edit_distance", "min_line_distance"}, nil, []string{"token frequency (via frequency-bin matching, log2 bins of the summed frequency rank)"},
			cs7, n7, "A significant result means edit-adjacent (or edit-close) tokens tend to occur closer together in the manuscript than frequency alone predicts.",
			"Sampled, not exhaustive, over vocabulary pairs; occurrence lists are capped at 50 per token for cost.", "not separately assessed")
		metrics = append(metrics, m)
		addTest(len(metrics)-1, cs7)
	} else {
		metrics = append(metrics, csInconclusive("cs7/edit-distance-x-structural-distance", "edit distance correlates with structural distance", "fewer than 20 sampled vocabulary pairs with sufficient per-bin support", n7))
	}

	adjusted := fdr(tests)
	adjustedByID := map[string]NullTest{}
	for _, t := range adjusted {
		adjustedByID[t.ID] = t
	}
	minN := 20
	for _, idx := range toAdjust {
		id := metrics[idx].MetricID
		q := adjustedByID[id].QValue
		metrics[idx].MultipleTestingAdjustment = q
		switch {
		case metrics[idx].N < minN:
			metrics[idx].Status = "INCONCLUSIVE"
		case q <= cfg.Alpha:
			metrics[idx].Status = "SUPPORTED"
		default:
			metrics[idx].Status = "NOT_SUPPORTED"
		}
	}

	redundancyRows, classifications := redundancyAnalysis(grammarRuns)
	confirmatory, exploratory := confirmatoryAndExploratoryFindings(metrics, c)

	return CrossScaleResult{
		VariablesAvailable:    variables,
		Metrics:               metrics,
		NullRegistry:          nullModelRegistry(),
		RedundancyMatrix:      redundancyRows,
		MetricClassifications: classifications,
		ConfirmatoryFindings:  confirmatory,
		ExploratoryFindings:   exploratory,
	}
}

func confirmatoryAndExploratoryFindings(metrics []CrossScaleMetric, c corpus) ([]string, []string) {
	confirmatory := make([]string, 0, len(metrics))
	for _, m := range metrics {
		confirmatory = append(confirmatory, m.MetricID+": "+m.Status+" ("+m.Hypothesis+")")
	}
	exploratory := []string{
		"Connected-components-vs-label-propagation disagreement (see edit_graph_validation.transitive_merge.community_vs_components): any large ARI/NMI gap suggests the productive-rule graph's components merge sub-communities that a local, degree-aware method would keep separate. This is reported as a finding, not folded into any SUPPORTED verdict, and should be re-run as its own preregistered confirmatory test (e.g. a fixed, seed-independent community method) before task79.",
		"Hub-removal sensitivity (edit_graph_validation.transitive_merge.hub_dependence): a large giant-component share drop under top-5%-degree removal is exploratory evidence that family connectivity concentrates on a few hub tokens; a dedicated hub-identity-stability test (which specific tokens are hubs across partitions) is deferred.",
	}
	if c.info.PageMetadata {
		exploratory = append(exploratory, "Locus-type SPECIAL bucket (comments/running/radial text, pooled for power in CS3) may hide heterogeneous sub-effects; a per-code confirmatory follow-up needs more data than most individual special-locus codes currently provide.")
	}
	return confirmatory, exploratory
}
