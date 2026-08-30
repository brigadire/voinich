package structurecatalog

import (
	"fmt"
	"path/filepath"
	"sort"
	"strings"
)

func (cat *Catalog) writeTokens(edges []editEdge, degree map[string]int) error {
	c := cat.Primary
	type ts struct {
		lines, folios, sections map[string]bool
		pos                     []float64
		first, last             int
	}
	st := map[string]*ts{}
	for t := range c.Counts {
		st[t] = &ts{lines: map[string]bool{}, folios: map[string]bool{}, sections: map[string]bool{}}
	}
	for _, o := range c.Occurrences {
		s := st[o.Token]
		s.lines[si(o.Line+1)] = true
		if c.MetadataAvailable {
			s.folios[o.Meta.Folio] = true
			if o.Meta.Section != "" {
				s.sections[o.Meta.Section] = true
			}
		}
		l := len(c.Lines[o.Line])
		p := .5
		if l > 1 {
			p = float64(o.Index) / float64(l-1)
		}
		s.pos = append(s.pos, p)
		if o.Index == 0 {
			s.first++
		}
		if o.Index == l-1 {
			s.last++
		}
	}
	tokens := sortedTokens(c.Counts)
	rows, posRows := [][]string{}, [][]string{}
	posRules := []Rule{}
	exclusiveTokens := map[string]bool{}
	globalBoundary := ratio(len(c.Lines), len(c.Occurrences))
	for _, t := range tokens {
		s := st[t]
		r := []rune(t)
		sum := 0.0
		bins := make([]int, 5)
		for _, p := range s.pos {
			sum += p
			switch {
			case p == 0:
				bins[0]++
			case p < .25:
				bins[1]++
			case p <= .75:
				bins[2]++
			case p < 1:
				bins[3]++
			default:
				bins[4]++
			}
		}
		mean := sum / float64(len(s.pos))
		classes := []string{}
		if s.first == 0 {
			classes = append(classes, "NEVER_FIRST")
		}
		if s.last == 0 {
			classes = append(classes, "NEVER_LAST")
		}
		if s.first == len(s.pos) {
			classes = append(classes, "FIRST_ONLY")
			exclusiveTokens[t] = true
		}
		if s.last == len(s.pos) {
			classes = append(classes, "LAST_ONLY")
			exclusiveTokens[t] = true
		}
		exp := float64(c.Counts[t]) * globalBoundary
		for _, x := range []struct {
			name string
			obs  int
		}{{"FIRST", s.first}, {"LAST", s.last}} {
			rr := Rule{Level: "L2", RuleType: "TOKEN_POSITION", LHS: t, RHS: x.name, Context: "physical_line", ObservedCount: x.obs, OpportunityCount: c.Counts[t], ObservedProbability: ratio(x.obs, c.Counts[t]), ExpectedCount: exp, EffectSize: effect(float64(x.obs), exp), PRaw: poissonTail(x.obs, exp, float64(x.obs) >= exp), ObservedStatus: "OBSERVED", CorpusRule: "OBSERVED", Stability: "NOT_COMPARABLE", Provenance: provenance(c)}
			if x.obs == 0 {
				rr.ObservedStatus = "UNOBSERVED"
				rr.CorpusRule = "NEVER_OBSERVED"
			}
			posRules = append(posRules, rr)
		}
		rows = append(rows, []string{t, si(c.Counts[t]), si(len(r)), strings.Join(glyphStrings(r), "|"), string(r[0]), string(r[len(r)-1]), si(len(s.lines)), si(len(s.folios)), si(len(s.sections)), sf(mean), sf(median(s.pos)), si(cat.TokenFamily[t])})
		cat.addRule("T1", "TOKEN_EXISTENCE", t, "PRESENT", "corpus", c.Counts[t], len(c.Occurrences), 0)
		posRows = append(posRows, []string{t, si(c.Counts[t]), si(s.first), si(bins[1]), si(bins[2]), si(bins[3]), si(s.last), sf(mean), sf(median(s.pos)), sf(entropy(bins)), ss(classes)})
	}
	applyBH(posRules)
	for i := range posRows {
		firstStatus := posRules[2*i].InferredStatus
		lastStatus := posRules[2*i+1].InferredStatus
		statistical := "POSITIONALLY_NEUTRAL"
		if strings.Contains(firstStatus, "PREFERRED") {
			statistical = "STRONGLY_INITIAL"
		}
		if strings.Contains(lastStatus, "PREFERRED") {
			statistical = "STRONGLY_FINAL"
		}
		posRows[i] = append(posRows[i], firstStatus, lastStatus, statistical)
	}
	cat.Rules = append(cat.Rules, posRules...)
	if err := writeTSV(filepath.Join(cat.Config.OutputDir, "TOKEN_CATALOG.tsv"), []string{"token", "frequency", "length_symbols", "glyph_composition", "initial_glyph", "final_glyph", "physical_line_coverage", "folio_coverage", "section_coverage", "mean_normalized_line_position", "median_normalized_line_position", "edit_family_id"}, rows); err != nil {
		return err
	}
	if err := writeTSV(filepath.Join(cat.Config.OutputDir, "TOKEN_POSITION_RULES.tsv"), []string{"token", "count", "first_count", "early_count", "middle_count", "late_count", "last_count", "mean_position", "median_position", "positional_entropy_bits", "corpus_rules", "first_inferred_status", "last_inferred_status", "statistical_class"}, posRows); err != nil {
		return err
	}
	cat.Summary["position_exclusive_tokens"] = si(len(exclusiveTokens))
	if err := cat.writeAffixes(tokens); err != nil {
		return err
	}
	return cat.writeFamilies(tokens, edges, degree)
}

func glyphStrings(r []rune) []string {
	x := make([]string, len(r))
	for i, g := range r {
		x[i] = string(g)
	}
	return x
}

func (cat *Catalog) writeAffixes(tokens []string) error {
	type as struct{ types, freq int }
	m := map[string]*as{}
	for _, t := range tokens {
		r := []rune(t)
		for n := 1; n <= 4 && n <= len(r); n++ {
			for _, x := range []struct{ k, v string }{{"PREFIX", string(r[:n])}, {"SUFFIX", string(r[len(r)-n:])}} {
				key := x.k + "\x00" + x.v
				if m[key] == nil {
					m[key] = &as{}
				}
				m[key].types++
				m[key].freq += cat.Primary.Counts[t]
			}
		}
	}
	rows := [][]string{}
	for k, s := range m {
		p := strings.SplitN(k, "\x00", 2)
		class := "RARE"
		if s.types >= 10 {
			class = "PRODUCTIVE"
		}
		rows = append(rows, []string{p[0], si(len([]rune(p[1]))), p[1], si(s.types), si(s.freq), class})
	}
	sort.Slice(rows, func(i, j int) bool {
		if rows[i][0] != rows[j][0] {
			return rows[i][0] < rows[j][0]
		}
		if rows[i][1] != rows[j][1] {
			return rows[i][1] < rows[j][1]
		}
		return rows[i][2] < rows[j][2]
	})
	return writeTSV(filepath.Join(cat.Config.OutputDir, "TOKEN_AFFIX_PATTERNS.tsv"), []string{"pattern_type", "length", "pattern", "token_type_count", "token_frequency_count", "productivity_class"}, rows)
}

func (cat *Catalog) writeFamilies(tokens []string, edges []editEdge, degree map[string]int) error {
	neigh := map[string][]string{}
	for _, e := range edges {
		neigh[e.A] = append(neigh[e.A], e.B)
		neigh[e.B] = append(neigh[e.B], e.A)
	}
	rows := [][]string{}
	for _, t := range tokens {
		sort.Strings(neigh[t])
		rows = append(rows, []string{t, si(cat.TokenFamily[t]), si(cat.Primary.Counts[t]), si(degree[t]), ss(neigh[t])})
		cat.addRule("T2", "TOKEN_EDIT_FAMILY", t, si(cat.TokenFamily[t]), "literal_edit_distance_1_component;degree="+si(degree[t]), 1, 1, 0)
	}
	if err := writeTSV(filepath.Join(cat.Config.OutputDir, "TOKEN_EDIT_FAMILIES.tsv"), []string{"token", "family_id", "frequency", "edit_distance_1_degree", "edit_distance_1_neighbours"}, rows); err != nil {
		return err
	}
	type agg struct {
		support  int
		examples []string
	}
	rules := map[string]*agg{}
	add := func(kind, from, to string, pos int, ex string) {
		k := strings.Join([]string{kind, from, to, si(pos)}, "\x00")
		if rules[k] == nil {
			rules[k] = &agg{}
		}
		rules[k].support++
		if len(rules[k].examples) < 5 {
			rules[k].examples = append(rules[k].examples, ex)
		}
	}
	for _, e := range edges {
		add(e.Kind, e.From, e.To, e.Position, e.A+"→"+e.B)
		if e.Kind == "INSERTION" {
			add("DELETION", e.To, "", e.Position, e.B+"→"+e.A)
		}
	}
	rrows := [][]string{}
	for k, a := range rules {
		p := strings.Split(k, "\x00")
		rrows = append(rrows, []string{p[0], p[1], p[2], p[3], si(a.support), ss(a.examples), "observed edit-distance-1 relation; not a morphological claim"})
	}
	sort.Slice(rrows, func(i, j int) bool {
		if rrows[i][4] != rrows[j][4] {
			return parseInt(rrows[i][4]) > parseInt(rrows[j][4])
		}
		return strings.Join(rrows[i], "\t") < strings.Join(rrows[j], "\t")
	})
	cat.Summary["edit_families"] = si(len(cat.Families))
	cat.Summary["edit_distance_1_edges"] = si(len(edges))
	return writeTSV(filepath.Join(cat.Config.OutputDir, "TOKEN_EDIT_RULES.tsv"), []string{"operation", "from_symbol", "to_symbol", "position_zero_based", "support_edges", "examples", "interpretation_limit"}, rrows)
}

func parseInt(s string) int { var x int; fmt.Sscanf(s, "%d", &x); return x }
