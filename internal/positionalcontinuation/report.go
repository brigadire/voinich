package positionalcontinuation

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func distinctSet(xs []string) int {
	m := map[string]bool{}
	for _, x := range xs {
		m[x] = true
	}
	return len(m)
}

// writeReport implements task23 Part U (sections 100-103) and Part R
// (sections 93-95): counts first, then questions A-H answered explicitly with
// support counts, then Q1/Q2 kept separate rather than collapsed to one
// p-value.
func writeReport(c Config, meta runMeta, state *RunState) error {
	sAiin := state.SAiinOccurrences
	aiin := state.AiinOccurrences
	totalSAiinX, totalCheyStrict := 0, 0
	var currier, hand, joint, blocks []string
	for _, o := range sAiin {
		if o.X != "" {
			totalSAiinX++
			if o.X == FrozenChey {
				totalCheyStrict++
			}
		}
		currier = append(currier, o.Currier)
		hand = append(hand, o.Hand)
		joint = append(joint, o.Joint)
		blocks = append(blocks, o.Block)
	}

	dep := map[string]PositionDependenceRow{}
	for _, d := range state.PositionDependence {
		dep[d.PositionVariable] = d
	}
	strat := map[string]StratifiedPredecessorRow{}
	for _, s := range state.StratifiedPredecessor {
		strat[s.PositionVariable] = s
	}
	eligible, positive, negative, neutral, consistency := crossBlockSignConsistency(state.CrossBlock)
	v := state.Validation

	var b strings.Builder
	fmt.Fprintf(&b, "# Positional continuation validation: \"s aiin\" -> \"chey\"\n\n")
	fmt.Fprintf(&b, "Confirmatory deep-dive on one frozen higher-order-sequence-validate finding: does the continuation distribution after the fixed context `%s` depend on structural position, and if so what explains it? Context and target continuation are frozen inputs (task23 section 2); no new n-gram discovery happens here. Corpus SHA256: `%s`.\n\n", FrozenSAiin, meta.CorpusSHA256)

	fmt.Fprintf(&b, "## Counts (section 100)\n\n")
	fmt.Fprintf(&b, "- total `aiin` occurrences: %d\n", len(aiin))
	fmt.Fprintf(&b, "- total `s aiin` occurrences: %d\n", len(sAiin))
	fmt.Fprintf(&b, "- total `s aiin X` occurrences (X present): %d\n", totalSAiinX)
	fmt.Fprintf(&b, "- total `s aiin chey` occurrences: %d\n", totalCheyStrict)
	fmt.Fprintf(&b, "- physical blocks containing `s aiin`: %d\n", distinctSet(blocks))
	fmt.Fprintf(&b, "- joint (Currier x hand) classes: %d\n", distinctSet(joint))
	fmt.Fprintf(&b, "- Currier classes: %d\n", distinctSet(currier))
	fmt.Fprintf(&b, "- hands: %d\n\n", distinctSet(hand))

	fmt.Fprintf(&b, "Previous experiment (documented context only, not recomputed as an input - task23 section 103): canonical `s aiin chey` occurrences were reported as approximately 4, across 3 eligible physical blocks, with block-position TVD=0.571 and line-position TVD=0.226, giving a POSITION_DEPENDENT status. This run's fresh recount above should be compared against that figure honestly rather than assumed to match.\n\n")

	fmt.Fprintf(&b, "## A. Is `s aiin` itself positionally specialized?\n\n")
	for _, rp := range state.ReversePosition {
		if rp.Stratum == blockCoarseCategories[len(blockCoarseCategories)-1] || rp.Stratum == lineCategories[len(lineCategories)-1] {
			fmt.Fprintf(&b, "- %s: total variation distance between P(position|s,aiin) and P(position|aiin) = %s (support: %d s-aiin occurrences, %d aiin occurrences).\n", rp.PositionVariable, f(rp.TotalVariation), len(sAiin), len(aiin))
		}
	}
	fmt.Fprintf(&b, "\n## B. Does position change the continuation distribution after `s aiin`?\n\n")
	fmt.Fprintf(&b, "- line_position: observed I(X;position)=%s bits, permutation empirical p=%s (%d permutations, support=%d occurrences with X present).\n", f(dep["line_position"].ObservedMIBits), f(dep["line_position"].EmpiricalP), dep["line_position"].Permutations, totalSAiinX)
	fmt.Fprintf(&b, "- block_position_coarse: observed I(X;position)=%s bits, permutation empirical p=%s (%d permutations).\n\n", f(dep["block_position_coarse"].ObservedMIBits), f(dep["block_position_coarse"].EmpiricalP), dep["block_position_coarse"].Permutations)

	fmt.Fprintf(&b, "## C. Is `chey` specifically enriched at some position?\n\n")
	for _, ce := range state.CheyEffect {
		if ce.PositionVariable != "line_position" {
			continue
		}
		fmt.Fprintf(&b, "- %s: n=%d, chey=%d, P(chey|position)=%s vs P(chey|s,aiin)=%s, enrichment=%s, p=%s.\n",
			ce.Stratum, ce.OccurrenceCount, ce.CheyCount, f(ce.PCheyGivenPosition), f(ce.PCheyGlobal), f(ce.PositionalEnrichment), f(ce.EmpiricalP))
	}

	fmt.Fprintf(&b, "\n## D. Does the same positional effect exist for `aiin` generally?\n\n")
	for _, ac := range state.AiinControl {
		if ac.PositionVariable != "line_position" {
			continue
		}
		fmt.Fprintf(&b, "- %s: aiin n=%d, P(chey|aiin,position)=%s vs P(chey|s,aiin,position)=%s, within-position enrichment E(position)=%s.\n",
			ac.Stratum, ac.AiinOccurrenceCount, f(ac.PCheyGivenAiinPosition), f(ac.PCheyGivenSAiinPosition), f(ac.WithinPositionEnrichment))
	}

	fmt.Fprintf(&b, "\n## E. After controlling position, does predecessor `s` still matter?\n\n")
	fmt.Fprintf(&b, "Stratified permutation test (chey ⟂ s | aiin, position), predecessor identity shuffled within (block, position) strata: line_position empirical p=%s (%d permutations), block_position_coarse empirical p=%s (%d permutations).\n\n",
		f(strat["line_position"].EmpiricalP), strat["line_position"].Permutations, f(strat["block_position_coarse"].EmpiricalP), strat["block_position_coarse"].Permutations)

	fmt.Fprintf(&b, "## F. Does adding position improve held-out prediction? / G. Does adding predecessor after position improve it further?\n\n")
	m := state.ModelLOBO
	fmt.Fprintf(&b, "Leave-one-physical-block-out, alpha=%.1f smoothing, %d tested blocks: M2 beats M1 in %d blocks, M1 beats M2 in %d (mean delta_21=%s bits, median=%s). M3 beats M2 in %d blocks, M2 beats M3 in %d (mean delta_32=%s bits, median=%s).\n\n",
		smoothingAlpha, m.TestedBlocks, m.BlocksM2BetterM1, m.BlocksM1BetterM2, f(m.MeanDelta21), f(m.MedianDelta21), m.BlocksM3BetterM2, m.BlocksM2BetterM3, f(m.MeanDelta32), f(m.MedianDelta32))

	fmt.Fprintf(&b, "## H. Does the result reproduce across physical blocks?\n\n")
	fmt.Fprintf(&b, "Eligible blocks (>=1 s-aiin occurrence): %d, positive-sign blocks: %d, negative-sign blocks: %d, neutral: %d, sign consistency=%s.\n\n", eligible, positive, negative, neutral, f(consistency))

	fmt.Fprintf(&b, "## Line vs block position (Part M)\n\nPearson r(normalized_line_position, normalized_block_position) = %s. Source of the positional effect: **%s**.\n\n", f(state.LineVsBlockCorrelation), state.LineVsBlockSource)

	fmt.Fprintf(&b, "## Surrounding context (Part O)\n\n")
	for _, sc := range state.SurroundingContext {
		fmt.Fprintf(&b, "- %s: n=%d, preceding entropy=%s bits, following entropy=%s bits, unique surrounding contexts=%d.\n", sc.Group, sc.OccurrenceCount, f(sc.PrecedingEntropyBits), f(sc.FollowingEntropyBits), sc.UniqueSurroundingContexts)
	}

	fmt.Fprintf(&b, "\n## Part R: two questions, not one p-value (sections 93-95)\n\n")
	fmt.Fprintf(&b, "**Q1: does position affect continuation after s aiin?** primary (line_position) empirical p=%s; %s.\n\n", f(dep["line_position"].EmpiricalP), sigWord(v.PositionDependenceSig))
	fmt.Fprintf(&b, "**Q2: does s affect continuation after aiin, once position is controlled?** stratified predecessor empirical p=%s; %s.\n\n", f(strat["line_position"].EmpiricalP), sigWord(v.StratifiedPredecessorSig))

	fmt.Fprintf(&b, "## Diagnostic status\n\n**%s**\n\n", v.FinalStatus)
	fmt.Fprintf(&b, "Support: %d eligible physical blocks, cross-block sign consistency=%s, M3-vs-M2 held-out win fraction=%s, single-block-sensitive=%v, boundary-formula support=%v. This audit performs no new sequence discovery; it tests only the frozen `%s` -> `%s` finding.\n",
		v.EligibleBlocks, f(v.CrossBlockSignConsistency), f(v.M3BetterThanM2Fraction), v.SingleBlockSensitive, v.BoundaryFormulaSupported, FrozenSAiin, FrozenChey)

	return os.WriteFile(filepath.Join(c.OutputDir, "positional_continuation_report.md"), []byte(b.String()), 0644)
}

func sigWord(sig bool) string {
	if sig {
		return "significant at alpha=0.05"
	}
	return "not significant at alpha=0.05"
}
