package transitionnetwork

import (
	"bufio"
	"encoding/csv"
	"fmt"
	"html"
	"math"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

func f64(x float64) string { return strconv.FormatFloat(x, 'g', 10, 64) }
func writeTSV(path string, head []string, rows [][]string) error {
	f, e := os.Create(path)
	if e != nil {
		return e
	}
	w := csv.NewWriter(bufio.NewWriterSize(f, 1<<20))
	w.Comma = '\t'
	if e = w.Write(head); e == nil {
		for _, r := range rows {
			if e = w.Write(r); e != nil {
				break
			}
		}
	}
	w.Flush()
	if e == nil {
		e = w.Error()
	}
	if ce := f.Close(); e == nil {
		e = ce
	}
	return e
}
func setsString(xs []BlockStats, which string) string {
	m := map[string]bool{}
	for _, x := range xs {
		v := x.Joint
		if which == "currier" {
			v = x.Currier
		} else if which == "hand" {
			v = x.Hand
		} else if which == "block" {
			v = x.Block
		}
		m[v] = true
	}
	var s []string
	for v := range m {
		s = append(s, v)
	}
	sort.Strings(s)
	return strings.Join(s, ",")
}

func writeAll(c Config, a *analysis, corpusSHA, metaSHA string) error {
	if e := os.MkdirAll(filepath.Join(c.OutputDir, "plots"), 0755); e != nil {
		return e
	}
	var rows [][]string
	for _, x := range a.BlockEffects {
		rows = append(rows, []string{x.Block, x.Currier, x.Hand, x.Joint, x.Source, x.Target, strconv.Itoa(x.SourceCount), strconv.Itoa(x.TargetCount), strconv.Itoa(x.EdgeCount), strconv.Itoa(x.Opportunities), strconv.Itoa(x.BlockTokens), f64(x.PConditional), f64(x.PBaseline), f64(x.Enrichment), f64(x.Log2Enrichment), sign(x.Log2Enrichment)})
	}
	if e := writeTSV(filepath.Join(c.OutputDir, "transition_block_effects.tsv"), []string{"block", "currier", "hand", "joint_class", "source", "target", "source_count", "target_count", "edge_count", "opportunities", "block_tokens", "p_target_given_source", "p_target_baseline", "enrichment", "log2_enrichment", "sign"}, rows); e != nil {
		return e
	}
	rows = nil
	for _, r := range a.Summaries {
		rows = append(rows, edgeSummaryRow(r, a.ByEdge[r.EdgeKey]))
	}
	edgeHead := []string{"source", "target", "global_count", "eligible_blocks", "positive_blocks", "negative_blocks", "neutral_blocks", "joint_classes", "currier_classes", "hands", "joint_class_ids", "currier_ids", "hand_ids", "median_log2_enrichment", "mean_log2_enrichment", "between_block_sd", "sign", "sign_consistency", "tested_blocks", "successful_sign_predictions", "transfer_fraction", "empirical_p", "fdr_q", "permutations", "max_block_observation_fraction", "max_block_effect_weight_fraction", "status"}
	if e := writeTSV(filepath.Join(c.OutputDir, "transition_edge_summary.tsv"), edgeHead, rows); e != nil {
		return e
	}
	rows = nil
	for _, r := range a.Summaries {
		rows = append(rows, []string{r.Source, r.Target, r.ExpectedSign, strconv.Itoa(r.Permutations), f64(r.EmpiricalP), f64(r.FDRQ), f64(r.MaxBlockObservationFraction), f64(r.MaxBlockEffectWeightFraction), r.Status})
	}
	if e := writeTSV(filepath.Join(c.OutputDir, "transition_edge_controls.tsv"), []string{"source", "target", "sign", "permutations", "empirical_p", "fdr_q", "max_block_observation_fraction", "max_block_effect_weight_fraction", "status"}, rows); e != nil {
		return e
	}
	if e := writeProfiles(c.OutputDir, a.OutgoingRows, "transition_outgoing_profiles.tsv"); e != nil {
		return e
	}
	if e := writeProfiles(c.OutputDir, a.IncomingRows, "transition_incoming_profiles.tsv"); e != nil {
		return e
	}
	rows = nil
	for _, r := range a.Stability {
		rows = append(rows, []string{r.Token, r.Direction, strconv.Itoa(r.GlobalCount), strconv.Itoa(r.EligibleBlocks), strconv.Itoa(r.JointClasses), f64(r.PairwiseJSMean), f64(r.PairwiseJSMedian), f64(r.PairwiseJSMin), f64(r.PairwiseJSSD), f64(r.LOBOMedianCorrelation), f64(r.LOBOMeanCorrelation), f64(r.LOBOMedianSpearman), f64(r.SignAgreement), f64(r.PermutationP), f64(r.SignPermutationP), f64(r.EntropyEffect), f64(r.EntropySignConsistency), f64(r.EntropyPermutationP), r.EntropyStatus, strconv.FormatBool(r.Replicated)})
	}
	if e := writeTSV(filepath.Join(c.OutputDir, "transition_profile_stability.tsv"), []string{"token", "direction", "global_count", "eligible_blocks", "joint_classes", "pairwise_js_mean", "pairwise_js_median", "pairwise_js_min", "pairwise_js_sd", "lobo_median_correlation", "lobo_mean_correlation", "lobo_median_spearman", "sign_agreement", "permutation_p", "sign_permutation_p", "entropy_effect", "entropy_sign_consistency", "entropy_permutation_p", "entropy_status", "replicated"}, rows); e != nil {
		return e
	}
	rows = nil
	for _, r := range a.Entropies {
		rows = append(rows, []string{r.Token, r.Block, r.Direction, f64(r.ConditionalEntropy), f64(r.EffectiveCount), f64(r.BaselineEntropy), f64(r.EntropyEffect), strconv.FormatBool(r.Eligible)})
	}
	if e := writeTSV(filepath.Join(c.OutputDir, "transition_entropy.tsv"), []string{"token", "block", "direction", "conditional_entropy", "effective_successor_predecessor_count", "baseline_entropy", "entropy_effect", "eligible"}, rows); e != nil {
		return e
	}
	rows = nil
	for _, r := range a.Summaries {
		if strings.HasPrefix(r.Status, "BACKBONE_") {
			rows = append(rows, []string{r.Source, r.Target, r.ExpectedSign, strconv.Itoa(r.EligibleBlocks), strconv.Itoa(r.JointClasses), setsString(a.ByEdge[r.EdgeKey], "currier"), setsString(a.ByEdge[r.EdgeKey], "hand"), f64(r.MedianLog2), f64(r.SignConsistency), f64(r.TransferFraction), f64(r.EmpiricalP), f64(r.FDRQ), r.Status})
		}
	}
	if e := writeTSV(filepath.Join(c.OutputDir, "transition_network_backbone.tsv"), []string{"source", "target", "relation", "eligible_blocks", "joint_classes", "currier_classes", "hands", "median_log2_enrichment", "sign_consistency", "transfer_fraction", "empirical_p", "fdr_q", "status"}, rows); e != nil {
		return e
	}
	rows = nil
	byTok := map[string]map[string]ProfileStability{}
	for _, r := range a.Stability {
		if byTok[r.Token] == nil {
			byTok[r.Token] = map[string]ProfileStability{}
		}
		byTok[r.Token][r.Direction] = r
	}
	for _, t := range a.Vocab {
		o, i := byTok[t]["outgoing"], byTok[t]["incoming"]
		if o.Replicated {
			rows = append(rows, []string{t, strconv.Itoa(a.Counts[t]), strconv.Itoa(o.EligibleBlocks), strconv.Itoa(o.JointClasses), f64(o.LOBOMedianCorrelation), f64(i.LOBOMedianCorrelation), f64(o.EntropyEffect), f64(i.EntropyEffect)})
		}
	}
	if e := writeTSV(filepath.Join(c.OutputDir, "transition_profile_backbone.tsv"), []string{"token", "global_count", "eligible_blocks", "joint_classes", "outgoing_profile_stability", "incoming_profile_stability", "outgoing_entropy_effect", "incoming_entropy_effect"}, rows); e != nil {
		return e
	}
	rows = nil
	for _, r := range a.MetadataTransfer {
		rows = append(rows, []string{r.Dimension, r.GroupA, r.GroupB, strconv.Itoa(r.CommonEdges), f64(r.SignAgreement), f64(r.EffectCorrelation), f64(r.ProfileSimilarity)})
	}
	if e := writeTSV(filepath.Join(c.OutputDir, "transition_metadata_transfer.tsv"), []string{"dimension", "group_a", "group_b", "common_edges", "edge_sign_agreement", "effect_correlation", "profile_similarity"}, rows); e != nil {
		return e
	}
	rows = nil
	for _, r := range a.GraphSimilarity {
		rows = append(rows, []string{r.BlockA, r.BlockB, strconv.Itoa(r.EdgesA), strconv.Itoa(r.EdgesB), strconv.Itoa(r.Intersection), f64(r.EdgeJaccard), f64(r.DegreeRankCorrelation), f64(r.SCCOverlap)})
	}
	if e := writeTSV(filepath.Join(c.OutputDir, "transition_block_graph_similarity.tsv"), []string{"block_a", "block_b", "edges_a", "edges_b", "intersection", "edge_jaccard", "degree_rank_correlation", "scc_overlap"}, rows); e != nil {
		return e
	}
	rows = nil
	for _, r := range a.Predictions {
		rows = append(rows, []string{r.Block, r.Scope, strconv.Itoa(r.N), f64(r.LossM0), f64(r.LossM1), f64(r.Delta)})
	}
	if e := writeTSV(filepath.Join(c.OutputDir, "transition_prediction_lobo.tsv"), []string{"block", "scope", "transitions", "log_loss_m0", "log_loss_m1", "delta_m0_minus_m1"}, rows); e != nil {
		return e
	}
	rows = nil
	for _, r := range a.ModelOrder {
		rows = append(rows, []string{r.Block, strconv.Itoa(r.N), f64(r.LossM1), f64(r.LossM2), f64(r.Delta)})
	}
	if e := writeTSV(filepath.Join(c.OutputDir, "transition_model_order_lobo.tsv"), []string{"block", "transitions", "log_loss_m1", "log_loss_m2", "delta_m1_minus_m2"}, rows); e != nil {
		return e
	}
	if e := writeGraphML(filepath.Join(c.OutputDir, "preferred_backbone.graphml"), a, "BACKBONE_PREFERRED"); e != nil {
		return e
	}
	if e := writeGraphML(filepath.Join(c.OutputDir, "depleted_backbone.graphml"), a, "BACKBONE_DEPLETED"); e != nil {
		return e
	}
	acct := account(a)
	if e := writeSummaryAndReport(c, a, corpusSHA, metaSHA, acct); e != nil {
		return e
	}
	return writePlots(c.OutputDir, a)
}
func sign(x float64) string {
	if x > 0 {
		return "preferred"
	}
	if x < 0 {
		return "depleted"
	}
	return "neutral"
}
func edgeSummaryRow(r *EdgeSummary, x []BlockStats) []string {
	return []string{r.Source, r.Target, strconv.Itoa(r.GlobalCount), strconv.Itoa(r.EligibleBlocks), strconv.Itoa(r.PositiveBlocks), strconv.Itoa(r.NegativeBlocks), strconv.Itoa(r.NeutralBlocks), strconv.Itoa(r.JointClasses), strconv.Itoa(r.CurrierClasses), strconv.Itoa(r.Hands), setsString(x, "joint"), setsString(x, "currier"), setsString(x, "hand"), f64(r.MedianLog2), f64(r.MeanLog2), f64(r.BetweenBlockSD), r.ExpectedSign, f64(r.SignConsistency), strconv.Itoa(r.TestedBlocks), strconv.Itoa(r.SuccessfulSignPredictions), f64(r.TransferFraction), f64(r.EmpiricalP), f64(r.FDRQ), strconv.Itoa(r.Permutations), f64(r.MaxBlockObservationFraction), f64(r.MaxBlockEffectWeightFraction), r.Status}
}
func writeProfiles(dir string, x []profileRow, name string) error {
	rows := make([][]string, 0, len(x))
	for _, r := range x {
		rows = append(rows, []string{r.Token, r.Block, r.Direction, r.Destination, strconv.Itoa(r.Count), f64(r.Probability), f64(r.Log2Enrichment)})
	}
	return writeTSV(filepath.Join(dir, name), []string{"token", "block", "direction", "counterpart", "raw_count", "probability", "log2_enrichment"}, rows)
}

func writeGraphML(path string, a *analysis, status string) error {
	f, e := os.Create(path)
	if e != nil {
		return e
	}
	defer f.Close()
	fmt.Fprintln(f, `<?xml version="1.0" encoding="UTF-8"?><graphml xmlns="http://graphml.graphdrawing.org/xmlns"><key id="weight" for="edge" attr.name="weight" attr.type="double"/><key id="in_degree" for="node" attr.name="in_degree" attr.type="int"/><key id="out_degree" for="node" attr.name="out_degree" attr.type="int"/><key id="weighted_in_degree" for="node" attr.name="weighted_in_degree" attr.type="double"/><key id="weighted_out_degree" for="node" attr.name="weighted_out_degree" attr.type="double"/><graph edgedefault="directed">`)
	nodes := map[string]bool{}
	inD, outD := map[string]int{}, map[string]int{}
	win, wout := map[string]float64{}, map[string]float64{}
	for _, r := range a.Summaries {
		if r.Status == status {
			nodes[r.Source] = true
			nodes[r.Target] = true
			outD[r.Source]++
			inD[r.Target]++
			wout[r.Source] += r.MedianLog2
			win[r.Target] += r.MedianLog2
		}
	}
	var ns []string
	for n := range nodes {
		ns = append(ns, n)
	}
	sort.Strings(ns)
	for _, n := range ns {
		fmt.Fprintf(f, "<node id=\"%s\"><data key=\"in_degree\">%d</data><data key=\"out_degree\">%d</data><data key=\"weighted_in_degree\">%s</data><data key=\"weighted_out_degree\">%s</data></node>\n", html.EscapeString(n), inD[n], outD[n], f64(win[n]), f64(wout[n]))
	}
	for _, r := range a.Summaries {
		if r.Status == status {
			fmt.Fprintf(f, "<edge source=\"%s\" target=\"%s\"><data key=\"weight\">%s</data></edge>\n", html.EscapeString(r.Source), html.EscapeString(r.Target), f64(r.MedianLog2))
		}
	}
	fmt.Fprintln(f, "</graph></graphml>")
	return nil
}

type accounting struct{ Preferred, Depleted, BackPreferred, BackDepleted, Metadata, Concentrated, Unstable, Testable, Two, Three, CrossJoint, CrossCurrier, CrossHand, ProfileOut, ProfileIn int }

type reconstruction struct{ EdgeCoverage, TokenCoverage, TransitionCoverage, RandomTransitionCoverage float64 }
type topology struct {
	Nodes, Edges, SCC, WCC int
	Reciprocity            float64
}

func backboneTopology(a *analysis, status string) topology {
	g, nodes := map[EdgeKey]bool{}, map[string]bool{}
	for _, r := range a.Summaries {
		if r.Status == status {
			g[r.EdgeKey] = true
			nodes[r.Source] = true
			nodes[r.Target] = true
		}
	}
	z := topology{Nodes: len(nodes), Edges: len(g)}
	if len(g) > 0 {
		rec := 0
		for e := range g {
			if g[EdgeKey{e.Target, e.Source}] {
				rec++
			}
		}
		z.Reciprocity = float64(rec) / float64(len(g))
	}
	adj, rev, und := map[string][]string{}, map[string][]string{}, map[string][]string{}
	for e := range g {
		adj[e.Source] = append(adj[e.Source], e.Target)
		rev[e.Target] = append(rev[e.Target], e.Source)
		und[e.Source] = append(und[e.Source], e.Target)
		und[e.Target] = append(und[e.Target], e.Source)
	}
	seen := map[string]bool{}
	var order []string
	var dfs func(string)
	dfs = func(v string) {
		if seen[v] {
			return
		}
		seen[v] = true
		for _, w := range adj[v] {
			dfs(w)
		}
		order = append(order, v)
	}
	for v := range nodes {
		dfs(v)
	}
	seen = map[string]bool{}
	var mark func(string)
	mark = func(v string) {
		if seen[v] {
			return
		}
		seen[v] = true
		for _, w := range rev[v] {
			mark(w)
		}
	}
	for i := len(order) - 1; i >= 0; i-- {
		if !seen[order[i]] {
			z.SCC++
			mark(order[i])
		}
	}
	seen = map[string]bool{}
	var weak func(string)
	weak = func(v string) {
		if seen[v] {
			return
		}
		seen[v] = true
		for _, w := range und[v] {
			weak(w)
		}
	}
	for v := range nodes {
		if !seen[v] {
			z.WCC++
			weak(v)
		}
	}
	return z
}

func reconstructionMetrics(a *analysis) reconstruction {
	back, nodes := map[EdgeKey]bool{}, map[string]bool{}
	for _, r := range a.Summaries {
		if r.Status == "BACKBONE_PREFERRED" {
			back[r.EdgeKey] = true
			nodes[r.Source] = true
			nodes[r.Target] = true
		}
	}
	counts := map[EdgeKey]int{}
	total, hit := 0, 0
	for _, d := range a.Data {
		for e, n := range d.Edges {
			counts[e] += n
			total += n
			if back[e] {
				hit += n
			}
		}
	}
	z := reconstruction{}
	if len(a.Edges) > 0 {
		z.EdgeCoverage = float64(len(back)) / float64(len(a.Edges))
	}
	if len(a.Vocab) > 0 {
		z.TokenCoverage = float64(len(nodes)) / float64(len(a.Vocab))
	}
	if total > 0 {
		z.TransitionCoverage = float64(hit) / float64(total)
	}
	chosen := map[EdgeKey]bool{}
	for b := range back {
		best := EdgeKey{}
		bestD := math.Inf(1)
		for e, n := range counts {
			if back[e] || chosen[e] {
				continue
			}
			d := math.Abs(math.Log(float64(n)+1) - math.Log(float64(counts[b])+1))
			if d < bestD {
				best, bestD = e, d
			}
		}
		if !math.IsInf(bestD, 1) {
			chosen[best] = true
		}
	}
	randomHit := 0
	for e := range chosen {
		randomHit += counts[e]
	}
	if total > 0 {
		z.RandomTransitionCoverage = float64(randomHit) / float64(total)
	}
	return z
}

func account(a *analysis) accounting {
	var z accounting
	for _, r := range a.Summaries {
		if r.EligibleBlocks > 0 {
			z.Testable++
		}
		if r.EligibleBlocks >= 2 {
			z.Two++
		}
		if r.EligibleBlocks >= 3 {
			z.Three++
		}
		if r.JointClasses >= 2 {
			z.CrossJoint++
		}
		if r.CurrierClasses >= 2 {
			z.CrossCurrier++
		}
		if r.Hands >= 2 {
			z.CrossHand++
		}
		if r.FDRQ <= .05 {
			if r.ExpectedSign == "preferred" {
				z.Preferred++
			} else {
				z.Depleted++
			}
		}
		switch r.Status {
		case "BACKBONE_PREFERRED":
			z.BackPreferred++
		case "BACKBONE_DEPLETED":
			z.BackDepleted++
		case "METADATA_SPECIFIC":
			z.Metadata++
		case "BLOCK_CONCENTRATED":
			z.Concentrated++
		case "SIGNIFICANT_UNSTABLE":
			z.Unstable++
		}
	}
	for _, r := range a.Stability {
		if r.Replicated {
			if r.Direction == "outgoing" {
				z.ProfileOut++
			} else {
				z.ProfileIn++
			}
		}
	}
	return z
}

func writeSummaryAndReport(c Config, a *analysis, cs, ms string, z accounting) error {
	recon := reconstructionMetrics(a)
	prefTopology, deplTopology := backboneTopology(a, "BACKBONE_PREFERRED"), backboneTopology(a, "BACKBONE_DEPLETED")
	secondaryTokens := 0
	for _, n := range a.Counts {
		if n >= 5 {
			secondaryTokens++
		}
	}
	pred := []float64{}
	for _, r := range a.Predictions {
		if r.Scope == "all_eligible" {
			pred = append(pred, r.Delta)
		}
	}
	order := []float64{}
	better := 0
	for _, r := range a.ModelOrder {
		order = append(order, r.Delta)
		if r.Delta > 0 {
			better++
		}
	}
	outcome := "NO_BACKBONE"
	if z.BackPreferred+z.BackDepleted > 0 {
		outcome = "EDGE_BACKBONE_ONLY"
	}
	if z.ProfileOut > 0 && mean(pred) > 0 {
		outcome = "PROFILE_BACKBONE"
	}
	if outcome == "PROFILE_BACKBONE" && mean(order) > 0 {
		outcome = "HIGHER_ORDER_NEEDED"
	}
	yaml := fmt.Sprintf("corpus_sha256: %s\nmetadata_sha256: %s\ntoken_count: %d\nunique_tokens: %d\neligible_tokens: %d\nsecondary_eligible_tokens_count_ge_5: %d\nphysical_blocks: %d\nobserved_primary_edges: %d\ntestable_edges: %d\nfdr_significant_preferred: %d\nfdr_significant_depleted: %d\nbackbone_preferred: %d\nbackbone_depleted: %d\nreplicated_outgoing_profiles: %d\nreplicated_incoming_profiles: %d\nsequence_reconstruction:\n  edge_coverage: %s\n  token_coverage: %s\n  transition_coverage: %s\n  matched_random_transition_coverage: %s\nmean_m0_minus_m1_log_loss: %s\nmean_m1_minus_m2_log_loss: %s\noutcome: %s\nparameters:\n  min_token_count: %d\n  secondary_min_token_count: 5\n  min_block_token_count: %d\n  alpha: 0.5\n  permutations: %d\n  refine_permutations: %d\n  seed: %d\n", cs, ms, len(a.Tokens), len(a.Counts), len(a.Vocab), secondaryTokens, len(a.Blocks), len(a.Edges), z.Testable, z.Preferred, z.Depleted, z.BackPreferred, z.BackDepleted, z.ProfileOut, z.ProfileIn, f64(recon.EdgeCoverage), f64(recon.TokenCoverage), f64(recon.TransitionCoverage), f64(recon.RandomTransitionCoverage), f64(mean(pred)), f64(mean(order)), outcome, c.MinTokenCount, c.MinBlockTokenCount, c.Permutations, c.RefinePermutations, c.Seed)
	yaml += fmt.Sprintf("preferred_topology: {nodes: %d, edges: %d, strongly_connected_components: %d, weakly_connected_components: %d, reciprocity: %s}\ndepleted_topology: {nodes: %d, edges: %d, strongly_connected_components: %d, weakly_connected_components: %d, reciprocity: %s}\n", prefTopology.Nodes, prefTopology.Edges, prefTopology.SCC, prefTopology.WCC, f64(prefTopology.Reciprocity), deplTopology.Nodes, deplTopology.Edges, deplTopology.SCC, deplTopology.WCC, f64(deplTopology.Reciprocity))
	if e := os.WriteFile(filepath.Join(c.OutputDir, "transition_network_summary.yaml"), []byte(yaml), 0644); e != nil {
		return e
	}
	report := fmt.Sprintf("# Transition network validation\n\n## Global accounting\n\nUnique tokens: %d; eligible tokens: %d; observed adjacent eligible edges: %d; testable edges: %d. Edges in >=2 blocks: %d; in >=3 blocks: %d. Cross-joint: %d; cross-Currier: %d; cross-hand: %d.\n\nFDR-significant preferred: %d; depleted: %d. Strict backbone preferred: %d; depleted: %d. Metadata-specific: %d; block-concentrated: %d; significant unstable: %d.\n\nReplicated outgoing profiles: %d; incoming profiles: %d.\n\n## Predictive validation\n\nMean held-out M0-M1 log-loss delta: %.6g (positive favors knowledge of A). Mean held-out M1-M2 delta: %.6g; M2 wins in %d/%d blocks.\n\n## Required questions\n\n1. **Does A improve held-out prediction?** %s\n2. **Do edge preferences/depletions reproduce?** %s\n3. **Is there a metadata-independent backbone?** %s\n4. **Are continuation constraints reproducible?** %s\n5. **Does second order materially improve?** %s\n\n## Outcome\n\n`%s`. This is evidence only about reproducible transition constraints; it does not imply language, grammar, semantics, or decipherment.\n", len(a.Counts), len(a.Vocab), len(a.Edges), z.Testable, z.Two, z.Three, z.CrossJoint, z.CrossCurrier, z.CrossHand, z.Preferred, z.Depleted, z.BackPreferred, z.BackDepleted, z.Metadata, z.Concentrated, z.Unstable, z.ProfileOut, z.ProfileIn, mean(pred), mean(order), better, len(order), yesno(mean(pred) > 0), yesno(z.BackPreferred+z.BackDepleted > 0), yesno(z.BackPreferred+z.BackDepleted > 0), yesno(z.ProfileOut+z.ProfileIn > 0), yesno(mean(order) > 0), outcome)
	report += fmt.Sprintf("\n## Sequence reconstruction diagnostic\n\nPreferred-backbone edge coverage: %.6g; token coverage: %.6g; observed transition coverage: %.6g. A deterministic frequency-matched graph with the same number of edges covers %.6g of transitions.\n", recon.EdgeCoverage, recon.TokenCoverage, recon.TransitionCoverage, recon.RandomTransitionCoverage)
	return os.WriteFile(filepath.Join(c.OutputDir, "transition_network_report.md"), []byte(report), 0644)
}
func yesno(x bool) string {
	if x {
		return "Yes, by the pre-specified diagnostic."
	}
	return "No convincing evidence under the pre-specified criteria."
}

func writePlots(dir string, a *analysis) error {
	names := []string{"edge_sign_consistency.svg", "edge_enrichment_vs_blocks.svg", "profile_stability_distribution.svg", "outgoing_entropy_effect.svg", "incoming_entropy_effect.svg", "currier_edge_transfer.svg", "hand_edge_transfer.svg", "backbone_degree_distribution.svg", "first_order_prediction_lobo.svg", "model_order_comparison.svg"}
	for _, n := range names {
		title := strings.TrimSuffix(strings.ReplaceAll(n, "_", " "), ".svg")
		vals := plotValues(n, a)
		if e := writeBarSVG(filepath.Join(dir, "plots", n), title, vals); e != nil {
			return e
		}
	}
	return nil
}
func plotValues(name string, a *analysis) []float64 {
	var x []float64
	switch name {
	case "edge_sign_consistency.svg":
		for _, r := range a.Summaries {
			x = append(x, r.SignConsistency)
		}
	case "edge_enrichment_vs_blocks.svg":
		for _, r := range a.Summaries {
			x = append(x, r.MedianLog2)
		}
	case "profile_stability_distribution.svg":
		for _, r := range a.Stability {
			x = append(x, r.LOBOMedianCorrelation)
		}
	case "outgoing_entropy_effect.svg":
		for _, r := range a.Stability {
			if r.Direction == "outgoing" {
				x = append(x, r.EntropyEffect)
			}
		}
	case "incoming_entropy_effect.svg":
		for _, r := range a.Stability {
			if r.Direction == "incoming" {
				x = append(x, r.EntropyEffect)
			}
		}
	case "currier_edge_transfer.svg":
		for _, r := range a.MetadataTransfer {
			if r.Dimension == "currier" {
				x = append(x, r.EffectCorrelation)
			}
		}
	case "hand_edge_transfer.svg":
		for _, r := range a.MetadataTransfer {
			if r.Dimension == "hand" {
				x = append(x, r.EffectCorrelation)
			}
		}
	case "first_order_prediction_lobo.svg":
		for _, r := range a.Predictions {
			if r.Scope == "all_eligible" {
				x = append(x, r.Delta)
			}
		}
	case "model_order_comparison.svg":
		for _, r := range a.ModelOrder {
			x = append(x, r.Delta)
		}
	default:
		deg := map[string]float64{}
		for _, r := range a.Summaries {
			if strings.HasPrefix(r.Status, "BACKBONE_") {
				deg[r.Source]++
				deg[r.Target]++
			}
		}
		for _, v := range deg {
			x = append(x, v)
		}
	}
	sort.Float64s(x)
	if len(x) > 80 {
		x = x[len(x)-80:]
	}
	return x
}
func writeBarSVG(path, title string, x []float64) error {
	var b strings.Builder
	fmt.Fprintf(&b, "<svg xmlns=\"http://www.w3.org/2000/svg\" width=\"900\" height=\"360\" viewBox=\"0 0 900 360\"><rect width=\"100%%\" height=\"100%%\" fill=\"white\"/><text x=\"30\" y=\"28\" font-family=\"sans-serif\" font-size=\"18\">%s</text><line x1=\"30\" y1=\"190\" x2=\"880\" y2=\"190\" stroke=\"#888\"/>", html.EscapeString(title))
	max := 0.
	for _, v := range x {
		if math.Abs(v) > max {
			max = math.Abs(v)
		}
	}
	if max == 0 {
		max = 1
	}
	bw := math.Max(1, 840/math.Max(1, float64(len(x))))
	for i, v := range x {
		bh := v / max * 140
		y := 190 - bh
		if bh < 0 {
			y = 190
		}
		fmt.Fprintf(&b, "<rect x=\"%.2f\" y=\"%.2f\" width=\"%.2f\" height=\"%.2f\" fill=\"#3976af\" opacity=\".75\"/>", 35+float64(i)*bw, y, math.Max(.5, bw-.5), math.Abs(bh))
	}
	b.WriteString("</svg>")
	return os.WriteFile(path, []byte(b.String()), 0644)
}
