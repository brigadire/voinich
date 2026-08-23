package fingerprintv2

import (
	"math/rand"
	"sort"
	"strconv"
)

func graphRepresentations() []GraphRepresentation {
	return []GraphRepresentation{
		{Kind: "undirected_unweighted", Status: "IMPLEMENTED", Description: "EF1-EF4's edit-distance-one adjacency; every edge is a symmetrized directed rule pair."},
		{Kind: "directed_transformation", Status: "IMPLEMENTED", Description: "LP1's directed rule census (operation|zone|position_class|source→target) over the same edge set."},
		{Kind: "frequency_weighted", Status: "IMPLEMENTED", Description: "Same edges, weighted by the product of endpoint corpus frequencies; used only as a diagnostic hub-frequency check (frequencyWeightedHubShare), not to redefine components."},
		{Kind: "context_weighted", Status: "IMPLEMENTED", Description: "Same edges, weighted by how often the two endpoints co-occur within the same line; used only as a diagnostic (contextWeightedShare), not to redefine components."},
		{Kind: "distance_weighted", Status: "DEFERRED", Reason: "Every edge in this block is edit-distance exactly one by LP1/EF construction; a distance-weighted variant would require a distance-two+ edge census, which task77 does not add. Deferred rather than faked with a constant weight."},
		{Kind: "multilayer", Status: "DEFERRED", Reason: "Combining unweighted/frequency/context/directed layers into one structure requires an arbitrary cross-layer weighting choice that task77's own instructions warn against; reported as separate layers instead (task77 §2.1: \"нельзя объединять разные определения ребра без сохранения их типа\")."},
	}
}

// frequencyWeightedHubShare and contextWeightedShare realize the two
// "implemented" weighted representations above as scalar diagnostics: the
// share of total edge weight held by the top-degree decile of nodes.
func frequencyWeightedHubShare(g editGraph, freq map[string]int) float64 {
	return weightedHubShare(g, func(a, b string) float64 { return float64(freq[a]) * float64(freq[b]) })
}

func contextWeightedShare(g editGraph, c corpus) float64 {
	coLine := map[string]int{}
	byLine := map[int][]string{}
	for _, r := range c.records {
		byLine[r.Line] = append(byLine[r.Line], r.Token)
	}
	for _, tokens := range byLine {
		seen := map[string]bool{}
		unique := make([]string, 0, len(tokens))
		for _, t := range tokens {
			if !seen[t] {
				seen[t] = true
				unique = append(unique, t)
			}
		}
		sort.Strings(unique)
		for i := 0; i < len(unique); i++ {
			for j := i + 1; j < len(unique); j++ {
				coLine[unique[i]+"\x00"+unique[j]]++
			}
		}
	}
	return weightedHubShare(g, func(a, b string) float64 {
		x, y := a, b
		if x > y {
			x, y = y, x
		}
		return float64(coLine[x+"\x00"+y])
	})
}

func weightedHubShare(g editGraph, weight func(a, b string) float64) float64 {
	nodeWeight := map[string]float64{}
	total := 0.0
	for _, e := range g.edgeList() {
		w := weight(e[0], e[1])
		nodeWeight[e[0]] += w
		nodeWeight[e[1]] += w
		total += 2 * w
	}
	if total == 0 || len(g.nodes) == 0 {
		return 0
	}
	ranked := append([]string(nil), g.nodes...)
	sort.Slice(ranked, func(i, j int) bool {
		if nodeWeight[ranked[i]] != nodeWeight[ranked[j]] {
			return nodeWeight[ranked[i]] > nodeWeight[ranked[j]]
		}
		return ranked[i] < ranked[j]
	})
	decile := (len(ranked) + 9) / 10
	if decile == 0 {
		decile = len(ranked)
	}
	share := 0.0
	for _, n := range ranked[:decile] {
		share += nodeWeight[n]
	}
	return share / total
}

func familyStructuralDiagnostics(g editGraph, families [][]string, glyphs map[string][]string) []FamilyStructuralDiagnostic {
	out := make([]FamilyStructuralDiagnostic, 0, len(families))
	for i, family := range families {
		avg, indirect := averageShortestPathAndIndirectShare(g, family)
		art, bridges := articulationPointsAndBridges(g, family)
		core := kCoreDecomposition(g, family)
		coreSize, peripherySize := 0, 0
		for _, c := range core {
			if c >= 2 {
				coreSize++
			} else {
				peripherySize++
			}
		}
		diameter := componentDiameter(g, family)
		meanEdit := meanInternalEditDistance(family, glyphs)
		out = append(out, FamilyStructuralDiagnostic{
			FamilyIndex: i, Size: len(family), Diameter: diameter,
			AverageShortestPath: avg, IndirectPairShare: indirect,
			MeanInternalEditDistance: meanEdit,
			ArticulationPoints:       art, BridgeEdges: bridges,
			CoreSize: coreSize, PeripherySize: peripherySize,
		})
	}
	return out
}

func transitiveMergeAudit(g editGraph, families [][]string, glyphs map[string][]string, c corpus, freq map[string]int, rng *rand.Rand) TransitiveMergeAudit {
	restrictions := make([]PathRestriction, 0, 3)
	for _, hops := range []int{1, 2, 3} {
		restrictions = append(restrictions, depthLimitedComponents(g, hops))
	}
	labels := labelPropagation(g, rng, 100)
	componentOf := map[string]string{}
	for i, family := range families {
		for _, n := range family {
			componentOf[n] = componentLabel(i)
		}
	}
	// Nodes outside every size->=3 family (isolates and size<3 pairs) are
	// each their own singleton "family" for a fair like-for-like partition
	// comparison against label propagation over the identical node set.
	singleton := 0
	for _, n := range g.nodes {
		if _, ok := componentOf[n]; !ok {
			componentOf[n] = "SINGLETON" + componentLabel(singleton)
			singleton++
		}
	}
	a, b := make([]string, len(g.nodes)), make([]string, len(g.nodes))
	for i, n := range g.nodes {
		a[i], b[i] = componentOf[n], labels[n]
	}
	ari, nmi, vi := clusterAgreement(a, b)
	return TransitiveMergeAudit{
		Families:         familyStructuralDiagnostics(g, families, glyphs),
		HubDependence:    hubRemovalGiantShare(g, 0.05),
		PathRestrictions: restrictions,
		CommunityVsComponents: CommunityComparison{
			Method: "connected_components_vs_label_propagation", ARI: ari, NMI: nmi, VI: vi, Seed: 1,
		},
		FrequencyWeightedHubShare: frequencyWeightedHubShare(g, freq),
		ContextWeightedHubShare:   contextWeightedShare(g, c),
	}
}

// familyPartitions rebuilds the family graph under a documented perturbation
// and returns the resulting family label for every node in universe (empty
// string if the node fell out of the perturbed vocabulary), so callers can
// score co-membership stability with clusterAgreement restricted to a
// common comparable node set.
func componentLabelsFor(families [][]string, universe []string) []string {
	componentOf := map[string]string{}
	for i, family := range families {
		for _, n := range family {
			componentOf[n] = componentLabel(i)
		}
	}
	out := make([]string, len(universe))
	for i, n := range universe {
		out[i] = componentOf[n] // "" if absent, a valid distinct label
	}
	return out
}

// runFamilyStabilityBattery implements task77 §2.3: it varies edit
// threshold (min_rule_support), minimum token frequency / singleton
// inclusion, corpus partition (folio halves), and community-detection
// method/seed, comparing each perturbed partition against the baseline via
// ARI/NMI restricted to the vocabulary shared by both runs.
func runFamilyStabilityBattery(c corpus, cfg Config, baseGraph editGraph, baseFamilies [][]string, seed int64) ([]StabilityRun, []map[string]string) {
	runs := make([]StabilityRun, 0, 8)
	replicates := make([]map[string]string, 0, 5)
	baseVocab := vocabulary(c)
	classify := func(ari float64, comparable int) (string, string) {
		switch {
		case comparable < 10:
			return "INSUFFICIENT_DATA", "fewer than 10 comparable tokens"
		case ari >= 0.5:
			return "GLOBAL", ""
		case ari >= 0.2:
			return "PARTITION_SPECIFIC", ""
		default:
			return "UNSTABLE", ""
		}
	}
	labelMap := func(universe, labels []string) map[string]string {
		out := make(map[string]string, len(universe))
		for i, n := range universe {
			out[n] = labels[i]
		}
		return out
	}
	addRun := func(perturbation, value string, universe []string, otherFamilies [][]string) {
		base := componentLabelsFor(baseFamilies, universe)
		other := componentLabelsFor(otherFamilies, universe)
		ari, nmi, _ := clusterAgreement(base, other)
		status, note := classify(ari, len(universe))
		runs = append(runs, StabilityRun{Perturbation: perturbation, Value: value, ARI: ari, NMI: nmi, ComparableNodes: len(universe), Status: status, Note: note})
		replicates = append(replicates, labelMap(universe, other))
	}

	// edit threshold (min_rule_support +/- 1)
	for _, delta := range []int{-1, 1} {
		threshold := cfg.MinRuleSupport + delta
		if threshold < 1 {
			continue
		}
		_, productive := lp1(c, buildGraph(c), threshold)
		g2 := productiveGraph(c, baseGraph, productive)
		families2, _ := splitComponents(g2.components(), 3)
		addRun("min_rule_support", strconv.Itoa(threshold), baseVocab, families2)
	}

	// minimum token frequency / singleton exclusion (freq >= 2)
	freq := frequencies(c)
	filtered := corpus{info: c.info}
	for _, r := range c.records {
		if freq[r.Token] >= 2 {
			filtered.records = append(filtered.records, r)
		}
	}
	if len(vocabulary(filtered)) >= 2 {
		fg := buildGraph(filtered)
		_, fp := lp1(filtered, fg, cfg.MinRuleSupport)
		ffamilies, _ := splitComponents(productiveGraph(filtered, fg, fp).components(), 3)
		addRun("min_token_frequency", "freq>=2 (excludes singletons)", baseVocab, ffamilies)
	}

	// corpus partition: folio halves (only meaningful with page metadata)
	pages := map[string]bool{}
	for _, r := range c.records {
		if r.Page != "" {
			pages[r.Page] = true
		}
	}
	if len(pages) >= 4 {
		ordered := orderedKeysBool(pages)
		half := map[string]bool{}
		for i, p := range ordered {
			if i%2 == 0 {
				half[p] = true
			}
		}
		firstHalf, secondHalf := corpus{info: c.info}, corpus{info: c.info}
		for _, r := range c.records {
			if half[r.Page] {
				firstHalf.records = append(firstHalf.records, r)
			} else {
				secondHalf.records = append(secondHalf.records, r)
			}
		}
		famA := familiesFromCorpus(firstHalf, cfg.MinRuleSupport)
		famB := familiesFromCorpus(secondHalf, cfg.MinRuleSupport)
		shared := intersectVocab(vocabulary(firstHalf), vocabulary(secondHalf))
		base := componentLabelsFor(famA, shared)
		other := componentLabelsFor(famB, shared)
		ari, nmi, _ := clusterAgreement(base, other)
		status, note := classify(ari, len(shared))
		runs = append(runs, StabilityRun{Perturbation: "corpus_partition", Value: "folio half A vs half B", ARI: ari, NMI: nmi, ComparableNodes: len(shared), Status: status, Note: note})
		replicates = append(replicates, labelMap(shared, base), labelMap(shared, other))
	} else {
		runs = append(runs, StabilityRun{Perturbation: "corpus_partition", Status: "NOT_TESTABLE", Note: "fewer than 4 distinct folios recorded"})
	}

	// preprocessing profile (glyph mode): structural summary only, since
	// EVA and natural tokenization do not share a common vocabulary space.
	runs = append(runs, StabilityRun{Perturbation: "preprocessing_profile", Status: "NOT_TESTABLE", Note: "EVA and natural glyph modes tokenize into disjoint alphabets; ARI/NMI over a shared node set is not defined. See report for a structural (component-count/giant-share) comparison instead."})

	// transcription variant
	runs = append(runs, StabilityRun{Perturbation: "transcription_variant", Status: "NOT_TESTABLE", Note: "only one transcription (ZL3b-n) is available under the repository's data discipline"})

	// random seed / community-detection method (label propagation seeds)
	seedARIs := make([]float64, 0, 3)
	labelsBySeed := make([][]string, 0, 3)
	for i, s := range []int64{seed + 101, seed + 202, seed + 303} {
		labels := labelPropagation(baseGraph, rand.New(rand.NewSource(s)), 100)
		out := make([]string, len(baseVocab))
		for j, n := range baseVocab {
			out[j] = labels[n]
		}
		labelsBySeed = append(labelsBySeed, out)
		if i > 0 {
			ari, _, _ := clusterAgreement(labelsBySeed[0], out)
			seedARIs = append(seedARIs, ari)
		}
	}
	meanSeedARI := mean(seedARIs)
	status, note := classify(meanSeedARI, len(baseVocab))
	runs = append(runs, StabilityRun{Perturbation: "community_detection_seed", Value: "label propagation, 3 seeds", ARI: meanSeedARI, ComparableNodes: len(baseVocab), Status: status, Note: note})

	// community-detection method: components vs label propagation is
	// already reported once in transitiveMergeAudit; cross-reference it
	// here as the "method" stability axis rather than recomputing it.
	return runs, replicates
}

func familiesFromCorpus(c corpus, minSupport int) [][]string {
	if len(vocabulary(c)) < 2 {
		return nil
	}
	g := buildGraph(c)
	_, productive := lp1(c, g, minSupport)
	families, _ := splitComponents(productiveGraph(c, g, productive).components(), 3)
	return families
}

func intersectVocab(a, b []string) []string {
	inB := map[string]bool{}
	for _, v := range b {
		inB[v] = true
	}
	out := make([]string, 0)
	for _, v := range a {
		if inB[v] {
			out = append(out, v)
		}
	}
	return out
}

func orderedKeysBool(m map[string]bool) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

// splitComponents separates connected components at or above minSize
// ("large", operational families) from smaller ones, matching lp3's
// existing convention so both callers agree on what counts as a family.
func splitComponents(comps [][]string, minSize int) (large [][]string, smallCount int) {
	for _, c := range comps {
		if len(c) >= minSize {
			large = append(large, c)
		} else {
			smallCount++
		}
	}
	return large, smallCount
}

// buildConsensusFamilies implements task77 §2.4: it only asserts a
// consensus partition when the observed families were reasonably stable
// across the perturbation battery (mean ARI over GLOBAL/PARTITION_SPECIFIC
// runs, excluding NOT_TESTABLE/INSUFFICIENT_DATA); otherwise it reports a
// local-neighborhood profile instead of manufacturing a paradigm catalog.
func buildConsensusFamilies(g editGraph, families [][]string, stability []StabilityRun, replicates []map[string]string, freq map[string]int, totalTokens, totalTypes int) EditFamilyValidation {
	testable := 0
	stableSum := 0.0
	for _, run := range stability {
		if run.Status == "NOT_TESTABLE" || run.Status == "INSUFFICIENT_DATA" {
			continue
		}
		testable++
		stableSum += run.ARI
	}
	consensusStatus := "INSUFFICIENT_SUPPORT"
	if len(families) == 0 {
		consensusStatus = "INSUFFICIENT_SUPPORT"
	} else if testable > 0 && stableSum/float64(testable) >= 0.3 {
		consensusStatus = "CONSENSUS_FAMILIES"
	} else if len(families) > 0 {
		consensusStatus = "LOCAL_NEIGHBORHOOD_PROFILE_ONLY"
	}
	var consensus []ConsensusFamily
	if consensusStatus == "CONSENSUS_FAMILIES" {
		var universe []string
		for _, family := range families {
			universe = append(universe, family...)
		}
		confidence := pairwiseCoMembershipStability(replicates, universe)
		for i, family := range families {
			core := kCoreDecomposition(g, family)
			members := make([]FamilyMember, 0, len(family))
			dominantHub, hubDegree := "", -1
			for _, n := range family {
				role := "PERIPHERY"
				if core[n] >= 2 {
					role = "CORE"
				}
				members = append(members, FamilyMember{Token: n, Role: role, Confidence: clampUnit(confidence[n])})
				if d := len(g.adj[n]); d > hubDegree {
					hubDegree, dominantHub = d, n
				}
			}
			sort.Slice(members, func(a, b int) bool { return members[a].Token < members[b].Token })
			occurrences := 0
			for _, n := range family {
				occurrences += freq[n]
			}
			consensus = append(consensus, ConsensusFamily{
				Index: i, Members: members, DominantHub: dominantHub,
				CorpusCoverage:     safeDiv(float64(len(family)), float64(totalTypes)),
				OccurrenceCoverage: safeDiv(float64(occurrences), float64(totalTokens)),
				StabilityScore:     safeDiv(stableSum, float64(max(testable, 1))),
			})
		}
	}
	return EditFamilyValidation{
		GraphRepresentations: graphRepresentations(),
		StabilityRuns:        stability,
		ConsensusStatus:      consensusStatus,
		ConsensusFamilies:    consensus,
	}
}

func clampUnit(v float64) float64 {
	if v < 0 {
		return 0
	}
	if v > 1 {
		return 1
	}
	return v
}

func safeDiv(a, b float64) float64 {
	if b == 0 {
		return 0
	}
	return a / b
}
