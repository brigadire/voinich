package structurecatalog

import (
	"math"
	"path/filepath"
	"sort"
	"strings"
)

func (cat *Catalog) writeLines() error {
	c := cat.Primary
	lineRows := [][]string{}
	lineMeta := make([]string, len(c.Lines))
	if c.MetadataAvailable {
		for _, o := range c.Occurrences {
			if lineMeta[o.Line] == "" {
				lineMeta[o.Line] = strings.Join([]string{o.Meta.Folio, o.Meta.LocusID, o.Meta.LocusType, o.Meta.Section}, "|")
			}
		}
	}
	first, last := map[string]int{}, map[string]int{}
	boundary := map[string]int{}
	lineSets := map[string]map[int]bool{}
	for li, line := range c.Lines {
		glyphs := 0
		repeats := 0
		seen := map[string]int{}
		fseen := map[int]int{}
		familyRepeats := 0
		lengths, families := []string{}, []string{}
		for _, t := range line {
			glyphs += len([]rune(t))
			seen[t]++
			if seen[t] > 1 {
				repeats++
			}
			fid := cat.TokenFamily[t]
			fseen[fid]++
			if fseen[fid] > 1 {
				familyRepeats++
			}
			lengths = append(lengths, si(len([]rune(t))))
			families = append(families, si(fid))
			if lineSets[t] == nil {
				lineSets[t] = map[int]bool{}
			}
			lineSets[t][li] = true
		}
		a, b := line[0], line[len(line)-1]
		first[a]++
		last[b]++
		boundary[a+"\x00"+b]++
		meta := lineMeta[li]
		lineRows = append(lineRows, []string{si(li + 1), si(len(line)), si(glyphs), a, b, string([]rune(a)[0]), string([]rune(b)[len([]rune(b))-1]), si(repeats), si(familyRepeats), ss(lengths), ss(families), meta})
	}
	if err := writeTSV(filepath.Join(cat.Config.OutputDir, "LINE_STRUCTURE.tsv"), []string{"physical_line", "token_count", "glyph_count", "first_token", "last_token", "first_glyph", "last_glyph", "repeated_token_occurrences", "repeated_family_occurrences", "token_length_progression", "token_family_progression", "folio_locus_type_section"}, lineRows); err != nil {
		return err
	}
	br := [][]string{}
	for t, n := range first {
		br = append(br, []string{"FIRST_TOKEN", t, "", si(n), si(len(c.Lines)), sf(ratio(n, len(c.Lines))), "OBSERVED"})
	}
	for t, n := range last {
		br = append(br, []string{"LAST_TOKEN", t, "", si(n), si(len(c.Lines)), sf(ratio(n, len(c.Lines))), "OBSERVED"})
	}
	boundaryKeys := make([]string, 0, len(boundary))
	for k := range boundary {
		boundaryKeys = append(boundaryKeys, k)
	}
	sort.Strings(boundaryKeys)
	for _, k := range boundaryKeys {
		n := boundary[k]
		p := strings.Split(k, "\x00")
		br = append(br, []string{"FIRST_TO_LAST", p[0], p[1], si(n), si(first[p[0]]), sf(ratio(n, first[p[0]])), "OBSERVED"})
		cat.addRule("L3", "LINE_BOUNDARY", p[0], p[1], "first_token_to_last_token", n, first[p[0]], 0)
	}
	sort.Slice(br, func(i, j int) bool { return strings.Join(br[i], "\t") < strings.Join(br[j], "\t") })
	if err := writeTSV(filepath.Join(cat.Config.OutputDir, "LINE_BOUNDARY_RULES.tsv"), []string{"rule_type", "lhs", "rhs", "observed_count", "opportunity_count", "observed_probability", "observed_status"}, br); err != nil {
		return err
	}
	return cat.writeCooccurrence(lineSets)
}

func (cat *Catalog) writeCooccurrence(lineSets map[string]map[int]bool) error {
	c := cat.Primary
	freq := []string{}
	for _, t := range sortedTokens(c.Counts) {
		if c.Counts[t] >= cat.Config.MinFrequency {
			freq = append(freq, t)
		}
	}
	null := newSameLineNull(c, freq)
	rules := []Rule{}
	for i, a := range freq {
		for j := i + 1; j < len(freq); j++ {
			b := freq[j]
			n := intersection(lineSets[a], lineSets[b])
			exp := null.expected(c.Counts[a], c.Counts[b])
			r := Rule{Level: "L1", RuleType: "TOKEN_LINE_COOCCURRENCE", LHS: a, RHS: b, Context: "same_physical_line_unordered", ObservedCount: n, OpportunityCount: len(lineSets[a]), ObservedProbability: ratio(n, len(lineSets[a])), ExpectedCount: exp, EffectSize: effect(float64(n), exp), PRaw: poissonTail(n, exp, float64(n) >= exp), ObservedStatus: "OBSERVED", CorpusRule: "OBSERVED", Stability: "NOT_COMPARABLE", Provenance: provenance(c)}
			if n == 0 {
				r.ObservedStatus = "UNOBSERVED"
				r.CorpusRule = "NEVER_COOCURS_IN_SAME_LINE"
			}
			rules = append(rules, r)
		}
	}
	applyBH(rules)
	yes, no := [][]string{}, [][]string{}
	avoided := 0
	for _, r := range rules {
		row := []string{r.LHS, r.RHS, si(len(lineSets[r.LHS])), si(len(lineSets[r.RHS])), si(r.ObservedCount), sf(r.ObservedProbability), sf(r.ExpectedCount), sf(r.EffectSize), sf(math.Exp(-r.ExpectedCount)), sf(r.PRaw), sf(r.QValue), r.CorpusRule, r.InferredStatus}
		if r.ObservedCount == 0 {
			no = append(no, row)
		} else {
			yes = append(yes, row)
		}
		if strings.Contains(r.InferredStatus, "AVOIDED") || r.InferredStatus == "DEPLETED" {
			avoided++
		}
		cat.Rules = append(cat.Rules, r)
	}
	h := []string{"token_A", "token_B", "lines_A", "lines_B", "lines_AB", "p_B_present_given_A_present", "expected_lines_AB", "effect_size", "null_probability_zero", "p_raw", "q_value", "corpus_rule", "inferred_status"}
	if err := writeTSV(filepath.Join(cat.Config.OutputDir, "TOKEN_LINE_COOCCURRENCE.tsv"), h, yes); err != nil {
		return err
	}
	if err := writeTSV(filepath.Join(cat.Config.OutputDir, "TOKEN_LINE_NONCOOCCURRENCE.tsv"), h, no); err != nil {
		return err
	}
	possible := len(freq) * (len(freq) - 1) / 2
	cat.Summary["possible_frequent_same_line_pairs"] = si(possible)
	cat.Summary["observed_frequent_same_line_pairs"] = si(len(yes))
	cat.Summary["frequent_same_line_exclusions"] = si(len(no))
	cat.Summary["statistically_avoided_or_depleted_same_line_pairs"] = si(avoided)
	cat.Summary["same_line_constraint_density"] = sf(float64(len(no)) / float64(max(1, possible)))
	return nil
}
func intersection(a, b map[int]bool) int {
	if len(a) > len(b) {
		a, b = b, a
	}
	n := 0
	for x := range a {
		if b[x] {
			n++
		}
	}
	return n
}

// expectedSameLine is the analytic expectation of the frozen null that
// assigns each token's fixed occurrence count to the corpus's fixed physical
// line slots. Thus both token frequencies and every line length are retained.
type sameLineNull struct {
	lineLengths    map[int]int
	orderedLengths []int
	presence       map[int]map[int]float64
}

func newSameLineNull(c Corpus, tokens []string) *sameLineNull {
	m := &sameLineNull{lineLengths: map[int]int{}, presence: map[int]map[int]float64{}}
	for _, line := range c.Lines {
		m.lineLengths[len(line)]++
	}
	for l := range m.lineLengths {
		m.orderedLengths = append(m.orderedLengths, l)
	}
	sort.Ints(m.orderedLengths)
	total := len(c.Occurrences)
	for _, t := range tokens {
		k := c.Counts[t]
		if m.presence[k] != nil {
			continue
		}
		m.presence[k] = map[int]float64{}
		for _, lineLen := range m.orderedLengths {
			if k > total-lineLen {
				m.presence[k][lineLen] = 1
				continue
			}
			none := 1.0
			for i := 0; i < k; i++ {
				none *= float64(total-lineLen-i) / float64(total-i)
			}
			m.presence[k][lineLen] = 1 - none
		}
	}
	return m
}
func (m *sameLineNull) expected(a, b int) float64 {
	v := 0.0
	for _, l := range m.orderedLengths {
		v += float64(m.lineLengths[l]) * m.presence[a][l] * m.presence[b][l]
	}
	return v
}
