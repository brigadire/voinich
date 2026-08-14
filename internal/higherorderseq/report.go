package higherorderseq

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// writeReport implements task22 Part S: the frozen inventory first, a
// compact table per primary sequence, then a per-candidate discussion that
// explicitly answers "does C depend on A given B?" without ever saying
// "ABC is a rule" for a HIGHER_ORDER_REPLICATED outcome (section 88).
func writeReport(c Config, meta runMeta, candidates []Candidate, results []*CandidateResult, dependence []DependenceRow, validation []ValidationRow) error {
	byID := map[string]*CandidateResult{}
	for _, r := range results {
		byID[r.Candidate.Sequence] = r
	}
	depByID := map[string]DependenceRow{}
	for _, d := range dependence {
		depByID[d.Sequence] = d
	}
	valByID := map[string]ValidationRow{}
	for _, v := range validation {
		valByID[v.Sequence] = v
	}

	var b strings.Builder
	fmt.Fprintf(&b, "# Higher-order sequential dependence validation\n\nConfirmatory test of whether the first token of a frozen n>=3 replicated sequence A B C carries information about the third token C beyond what the second token B alone predicts, i.e. P(C|A,B) vs P(C|B). Corpus SHA256: `%s`. No new bigrams, trigrams or n-grams were discovered; every candidate below is read programmatically from the previous audit.\n\n## Frozen inventory\n\n", meta.CorpusSHA256)
	for _, cand := range candidates {
		fmt.Fprintf(&b, "- `%s` (%s): shuffle FDR q=%s, Markov block p=%s, %d canonical occurrences across %d physical blocks and %d joint classes.\n",
			cand.Sequence, cand.Family, f(cand.ShuffleFDRQ), f(cand.MarkovBlockP), cand.CanonicalOccurrences, cand.PhysicalBlocks, cand.JointClasses)
	}

	fmt.Fprintf(&b, "\n## Summary table\n\n| sequence | family | occurrences | eligible blocks | joint classes | P(C\\|B) pooled | P(C\\|A,B) pooled | enrichment | conditional p | conditional q | CMI (bits) | LOBO M2 advantage | sign consistency | jackknife | status |\n|---|---|---|---|---|---|---|---|---|---|---|---|---|---|---|\n")
	for _, cand := range candidates {
		r := byID[cand.Sequence]
		d := depByID[cand.Sequence]
		v := valByID[cand.Sequence]
		pB, pAB, enrich := pooledEnrichment(primaryEligible(r.ConditionalRows))
		jk := "stable"
		if r.Jackknife.SingleBlockSensitive {
			jk = "SINGLE_BLOCK_SENSITIVE"
		}
		fmt.Fprintf(&b, "| %s | %s | %d | %d | %d | %s | %s | %s | %s | %s | %s | %.2f | %s | %s | %s |\n",
			cand.Sequence, cand.Family, len(r.Occurrences), r.CrossBlock.EligibleBlocks, r.CrossBlock.DistinctJoint,
			f(pB), f(pAB), f(enrich), f(d.EmpiricalP), f(d.FDRQ), f(r.CMI.ObservedCMIBits), v.LOBOAdvantageFraction, f(r.CrossBlock.SignConsistency), jk, v.FinalStatus)
	}

	fmt.Fprintf(&b, "\n## Per-candidate discussion\n\n")
	for _, cand := range candidates {
		r := byID[cand.Sequence]
		d := depByID[cand.Sequence]
		v := valByID[cand.Sequence]
		pB, pAB, enrich := pooledEnrichment(primaryEligible(r.ConditionalRows))
		fmt.Fprintf(&b, "### `%s`\n\nQuestion: does %s depend on %s once %s is already known?\n\n", cand.Sequence, cand.C(), cand.A(), cand.B())
		fmt.Fprintf(&b, "Pooled across %d eligible physical blocks (count(A,B)>=%d, count(B)>=%d): P(%s|%s)=%s, P(%s|%s,%s)=%s, enrichment=%s. The conditional-neighbor permutation null around %s (%d permutations) gives empirical p=%s (family-corrected q=%s). Leave-one-physical-block-out prediction favors the second-order model in %d/%d tested blocks (mean delta log loss=%s bits). Sign of the effect is consistent in %.0f%% of eligible blocks, spanning %d distinct joint metadata classes (cross-Currier=%v, cross-hand=%v). Jackknife removing each eligible block one at a time: %s.\n\n",
			r.CrossBlock.EligibleBlocks, primaryMinCountAB, primaryMinCountB,
			cand.C(), cand.B(), f(pB), cand.C(), cand.A(), cand.B(), f(pAB), f(enrich),
			cand.B(), r.CMI.Permutations, f(d.EmpiricalP), f(d.FDRQ),
			r.LOBO.M2BetterBlocks, r.LOBO.TestedBlocks, f(r.LOBO.MeanDeltaLogLoss),
			r.CrossBlock.SignConsistency*100, r.CrossBlock.DistinctJoint, r.CrossBlock.CrossCurrier, r.CrossBlock.CrossHand,
			jackknifeSummary(r.Jackknife))
		switch v.FinalStatus {
		case "HIGHER_ORDER_REPLICATED":
			fmt.Fprintf(&b, "**Diagnostic status: HIGHER_ORDER_REPLICATED.** `%s` exhibits replicated higher-order conditional dependence: %s after %s is not adequately predicted by %s alone.\n\n", cand.Sequence, cand.C(), cand.B(), cand.B())
		case "FIRST_ORDER_EXPLAINED":
			fmt.Fprintf(&b, "**Diagnostic status: FIRST_ORDER_EXPLAINED.** The data do not support a second-order effect beyond what P(%s|%s) already explains.\n\n", cand.C(), cand.B())
		case "POSITION_DEPENDENT":
			fmt.Fprintf(&b, "**Diagnostic status: POSITION_DEPENDENT.** The apparent effect concentrates near block or line boundaries (block-position TVD=%.3f, line-position TVD=%.3f) rather than holding as a general second-order transition.\n\n", r.BlockPosTVD, r.LinePosTVD)
		case "METADATA_LIMITED":
			fmt.Fprintf(&b, "**Diagnostic status: METADATA_LIMITED.** The effect reproduces across eligible blocks but those blocks do not span independent Currier/hand/joint classes.\n\n")
		case "SINGLE_BLOCK_SENSITIVE":
			fmt.Fprintf(&b, "**Diagnostic status: SINGLE_BLOCK_SENSITIVE.** The effect collapses or flips sign when a single eligible physical block is removed, so it cannot be treated as a stable replicated finding.\n\n")
		default:
			fmt.Fprintf(&b, "**Diagnostic status: INSUFFICIENT_SUPPORT.** Fewer than 3 eligible physical blocks are available for `%s`.\n\n", cand.Sequence)
		}
		cr := r.ContextRank
		fmt.Fprintf(&b, "Context-substitution control (Part O): among %d sufficiently frequent left contexts of %s, `%s %s` ranks %d (percentile %.1f, frozen P(%s|%s,%s)=%s vs baseline P(%s|%s)=%s) - i.e. this asks whether `%s` is unusual among all `X %s` contexts, not merely unusual relative to the whole corpus.\n\n",
			cr.NumAlternatives, cand.B(), cand.A(), cand.B(), cr.Rank, cr.Percentile, cand.C(), cand.A(), cand.B(), f(cr.FrozenP), cand.C(), cand.B(), f(cr.BaselineP), cand.A(), cand.B())
	}

	fmt.Fprintf(&b, "\n## Interpretation guardrails\n\nA HIGHER_ORDER_REPLICATED status means the sequence exhibits replicated higher-order conditional dependence, not that \"%s is a rule\". This audit performs no new sequence discovery, tests only the frozen candidates listed above, and establishes nothing about natural language, grammar, operator/operand structure, or decipherment.\n", strings.Join(seqTexts(candidates), ", "))
	return os.WriteFile(filepath.Join(c.OutputDir, "higher_order_sequence_report.md"), []byte(b.String()), 0644)
}

func jackknifeSummary(j JackknifeRow) string {
	if j.Realizations == 0 {
		return "not applicable (fewer than 1 eligible block)"
	}
	status := "stable across all removals"
	if j.SingleBlockSensitive {
		status = "SINGLE_BLOCK_SENSITIVE - sign flips when at least one block is removed"
	}
	return fmt.Sprintf("enrichment %s-%s (median %s), CMI %s-%s bits (median %s) - %s", f(j.EnrichmentMin), f(j.EnrichmentMax), f(j.EnrichmentMedian), f(j.CMIMin), f(j.CMIMax), f(j.CMIMedian), status)
}

func seqTexts(candidates []Candidate) []string {
	out := make([]string, len(candidates))
	for i, c := range candidates {
		out[i] = c.Sequence
	}
	return out
}
