package beginendanalyze

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
)

func WriteReports(directory string, output Output) error {
	data, err := yaml.Marshal(output)
	if err != nil {
		return fmt.Errorf("marshal YAML: %w", err)
	}
	if err := os.WriteFile(filepath.Join(directory, "begin_end_candidates.yaml"), data, 0o644); err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(directory, "begin_end_top.tsv"), []byte(topTSV(output.Candidates, 100)), 0o644); err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(directory, "begin_end_report.md"), []byte(markdownReport(output)), 0o644); err != nil {
		return err
	}
	return nil
}

func topTSV(candidates []Candidate, limit int) string {
	var b strings.Builder
	b.WriteString("rank\topening_candidate\tclosing_candidate\tscore\treliability\tranking_scope\tline_probability\tpage_probability\tdirectionality\tranking_z_score\tpermutation_p\tmean_distance\tadjacent_share\tpage_balance_stddev\tAABB\tABAB\n")
	if len(candidates) < limit {
		limit = len(candidates)
	}
	for i, c := range candidates[:limit] {
		significance := rankingSignificance(c)
		fmt.Fprintf(&b, "%d\t%s\t%s\t%.6f\t%.6f\t%s\t%.6f\t%.6f\t%.6f\t%.6f\t%.6g\t%.6f\t%.6f\t%.6f\t%d\t%d\n", i+1, c.BeginCandidate, c.EndCandidate, c.Score, c.Reliability, c.Directionality.Scope, c.WithinLine.Probability, c.WithinPage.Probability, c.Directionality.Score, significance.ZScore, significance.PermutationP, c.WithinPage.Mean, c.LocalCompatibility.AdjacentShare, c.PageBalance.StddevDifference, c.Nesting.AABB, c.Nesting.ABAB)
	}
	return b.String()
}

func markdownReport(output Output) string {
	var b bytes.Buffer
	b.WriteString("# Directed paired-token candidate report\n\n")
	fmt.Fprintf(&b, "The analysis covers %d token occurrences, %d lines, and %d inferred pages. %d tokens met the minimum frequency. The main ranking contains %d non-local candidate pairs; %d likely local pairs are separated from it.\n\n", output.Meta.TokenOccurrences, output.Meta.Lines, output.Meta.Pages, output.Meta.EligibleTokens, len(output.Candidates), len(output.LikelyLocalPairs))
	if !output.Meta.PageBoundariesKnown {
		b.WriteString("> Page boundaries were not present in the corpus. Page-scope results therefore use the entire corpus as one page and should be treated as provisional. Supply a corpus with blank-line, form-feed, or supported page markers for page-level inference.\n\n")
	}
	b.WriteString("Tokens containing `?` are excluded from the main ranking by default because they contain uncertain signs. `@NNN;` forms are preserved as ordinary complete tokens. The labels opening and closing candidate describe direction only and do not assign semantics.\n\n")
	b.WriteString("## Metrics\n\n")
	b.WriteString("For every occurrence of the first token, the analyzer finds the nearest later occurrence of the second token within the same line or page. It reports fixed windows, full-scope coverage, distance summaries, reverse direction, page balance relative to pairs in similar frequency bins, four neutral four-event orders, immediate adjacency, and boundary-preserving permutation significance. Low-count pairs receive the explicit reliability factor `min(count)/(min(count)+20)`. The final score is only a documented sorting aid; all component metrics remain in the YAML output.\n\n")
	writeSection(&b, "Best overall non-local pairs", output.Candidates, 10, func(c Candidate) float64 { return c.Score })
	writeSection(&b, "Best pairs within a line", output.Candidates, 10, func(c Candidate) float64 { return c.WithinLine.Probability * (1 - c.LocalCompatibility.AdjacentShare) })
	writeSection(&b, "Best pairs within a page", output.Candidates, 10, func(c Candidate) float64 { return c.WithinPage.Probability * (1 - c.LocalCompatibility.AdjacentShare) })
	writeSection(&b, "Strongest directionality", output.Candidates, 10, func(c Candidate) float64 { return c.Directionality.Score })
	writeSection(&b, "Most expressed page balance", output.Candidates, 10, func(c Candidate) float64 { return c.PageBalance.RelativeScore })
	writeSection(&b, "Pairs with nesting-like order contrast", output.Candidates, 10, func(c Candidate) float64 {
		total := c.Nesting.AABB + c.Nesting.ABAB + c.Nesting.ABBA + c.Nesting.BAAB
		if total == 0 {
			return 0
		}
		return float64(c.Nesting.AABB-c.Nesting.ABAB) / float64(total)
	})
	if len(output.LikelyLocalPairs) > 0 {
		writeSection(&b, "Likely local pairs (reported separately)", output.LikelyLocalPairs, 10, func(c Candidate) float64 { return c.LocalCompatibility.AdjacentShare })
	}
	b.WriteString("## Interpretation limits\n\n")
	b.WriteString("A small permutation p-value indicates stability against the selected constrained shuffle, not a grammatical construction. Frequent-token effects, transcription uncertainty, multiple testing, page-boundary quality, and corpus heterogeneity remain possible explanations. Candidates are therefore targets for follow-up inspection rather than identified operators.\n")
	return b.String()
}

func writeSection(b *bytes.Buffer, title string, candidates []Candidate, limit int, metric func(Candidate) float64) {
	b.WriteString("## " + title + "\n\n")
	items := append([]Candidate(nil), candidates...)
	sort.Slice(items, func(i, j int) bool {
		left, right := metric(items[i]), metric(items[j])
		if left != right {
			return left > right
		}
		if items[i].BeginCandidate != items[j].BeginCandidate {
			return items[i].BeginCandidate < items[j].BeginCandidate
		}
		return items[i].EndCandidate < items[j].EndCandidate
	})
	if len(items) < limit {
		limit = len(items)
	}
	if limit == 0 {
		b.WriteString("No qualifying pairs.\n\n")
		return
	}
	b.WriteString("| opening candidate | closing candidate | section metric | score | reliability | scope | P(line) | P(page) | directionality | ranking z | p | balance sd | AABB/ABAB |\n|---|---|---:|---:|---:|---|---:|---:|---:|---:|---:|---:|---:|\n")
	for _, c := range items[:limit] {
		significance := rankingSignificance(c)
		fmt.Fprintf(b, "| `%s` | `%s` | %.4f | %.4f | %.4f | %s | %.4f | %.4f | %.4f | %.3f | %.4g | %.3f | %d/%d |\n", c.BeginCandidate, c.EndCandidate, metric(c), c.Score, c.Reliability, c.Directionality.Scope, c.WithinLine.Probability, c.WithinPage.Probability, c.Directionality.Score, significance.ZScore, significance.PermutationP, c.PageBalance.StddevDifference, c.Nesting.AABB, c.Nesting.ABAB)
	}
	b.WriteString("\n")
}

func rankingSignificance(candidate Candidate) SignificanceResult {
	if candidate.Directionality.Scope == "line" {
		return candidate.SignificanceLine
	}
	return candidate.SignificancePage
}
