package structurecatalog

import (
	"fmt"
	"math"
	"path/filepath"
	"sort"
	"strings"
)

func Run(c Config) (*Catalog, error) {
	if c.MinFrequency <= 0 {
		c.MinFrequency = DefaultMinFreq
	}
	if c.OutputDir == "" {
		return nil, fmt.Errorf("output directory is required")
	}
	if err := ensureDir(c.OutputDir); err != nil {
		return nil, err
	}
	p, err := LoadCorpus(c.CorpusPath, "ZL3b-x7 canonical", c.IVTFFPath)
	if err != nil {
		return nil, err
	}
	var rep Corpus
	if c.IT2aPath != "" {
		rep, err = LoadCorpus(c.IT2aPath, "IT2a-x7 canonical", c.IT2aIVTFFPath)
		if err != nil {
			return nil, err
		}
	}
	fams, tf, edges, degree := buildFamilies(p.Counts)
	cat := &Catalog{Config: c, Primary: p, Replication: rep, Families: fams, TokenFamily: tf, Summary: map[string]string{}}
	if err = cat.writeGlyph(); err != nil {
		return nil, err
	}
	if err = cat.writeTokens(edges, degree); err != nil {
		return nil, err
	}
	if err = cat.writeTransitions(); err != nil {
		return nil, err
	}
	if err = cat.writeLines(); err != nil {
		return nil, err
	}
	if err = cat.writeMetadata(); err != nil {
		return nil, err
	}
	cat.applyStability()
	if err = writeRules(filepath.Join(c.OutputDir, "VM_STRUCTURAL_RULES.tsv"), cat.Rules); err != nil {
		return nil, err
	}
	if err = cat.writeSummaryAndDocs(); err != nil {
		return nil, err
	}
	if err = cat.writeManifest(); err != nil {
		return nil, err
	}
	return cat, nil
}

func (cat *Catalog) addRule(level, typ, lhs, rhs, ctx string, obs, opp int, exp float64) {
	r := Rule{Level: level, RuleType: typ, LHS: lhs, RHS: rhs, Context: ctx, ObservedCount: obs, OpportunityCount: opp, ObservedProbability: ratio(obs, opp), ExpectedCount: exp, EffectSize: effect(float64(obs), exp), ObservedStatus: "OBSERVED", CorpusRule: "OBSERVED", InferredStatus: "NOT_TESTED", Stability: "NOT_COMPARABLE", Provenance: provenance(cat.Primary), PRaw: math.NaN(), QValue: math.NaN()}
	if obs == 0 {
		r.ObservedStatus = "UNOBSERVED"
		r.CorpusRule = "NEVER_OBSERVED"
	}
	cat.Rules = append(cat.Rules, r)
}

func sortedTokens(m map[string]int) []string {
	r := make([]string, 0, len(m))
	for t := range m {
		r = append(r, t)
	}
	sort.Strings(r)
	return r
}
func setStrings(m map[string]bool) []string {
	r := make([]string, 0, len(m))
	for x := range m {
		r = append(r, x)
	}
	sort.Strings(r)
	return r
}

func (cat *Catalog) writeGlyph() error {
	c := cat.Primary
	type gs struct {
		total, tokens, types, lines                         int
		folios, sections                                    map[string]bool
		initial, internal, final, singleton, second, penult int
	}
	stats := map[rune]*gs{}
	for _, g := range c.Inventory {
		stats[g] = &gs{folios: map[string]bool{}, sections: map[string]bool{}}
	}
	lineSeen := map[rune]map[int]bool{}
	for _, g := range c.Inventory {
		lineSeen[g] = map[int]bool{}
	}
	for _, o := range c.Occurrences {
		r := []rune(o.Token)
		seen := map[rune]bool{}
		for i, g := range r {
			s := stats[g]
			s.total++
			lineSeen[g][o.Line] = true
			if c.MetadataAvailable {
				s.folios[o.Meta.Folio] = true
				if o.Meta.Section != "" {
					s.sections[o.Meta.Section] = true
				}
			}
			if i == 0 {
				s.initial++
			}
			if i == len(r)-1 {
				s.final++
			}
			if i > 0 && i < len(r)-1 {
				s.internal++
			}
			if len(r) == 1 {
				s.singleton++
			}
			if i == 1 {
				s.second++
			}
			if i == len(r)-2 {
				s.penult++
			}
			seen[g] = true
		}
		for g := range seen {
			stats[g].tokens++
		}
	}
	for t := range c.Counts {
		seen := map[rune]bool{}
		for _, g := range []rune(t) {
			seen[g] = true
		}
		for g := range seen {
			stats[g].types++
		}
	}
	invRows, posRows := [][]string{}, [][]string{}
	neverI, neverF, neverN := 0, 0, 0
	for _, g := range c.Inventory {
		s := stats[g]
		s.lines = len(lineSeen[g])
		invRows = append(invRows, []string{string(g), si(s.total), si(s.tokens), si(s.types), si(s.lines), si(len(s.folios)), si(len(s.sections))})
		cat.addRule("G1", "GLYPH_INVENTORY", string(g), "PRESENT", "corpus", s.total, s.total, 0)
		classes := []string{}
		if s.initial == 0 {
			classes = append(classes, "NEVER_INITIAL")
			neverI++
		}
		if s.final == 0 {
			classes = append(classes, "NEVER_FINAL")
			neverF++
		}
		if s.internal == 0 {
			classes = append(classes, "NEVER_INTERNAL")
			neverN++
		}
		if s.total == s.initial && s.singleton == 0 {
			classes = append(classes, "INITIAL_ONLY")
		}
		if s.total == s.final && s.singleton == 0 {
			classes = append(classes, "FINAL_ONLY")
		}
		if s.total == s.internal {
			classes = append(classes, "INTERNAL_ONLY")
		}
		posRows = append(posRows, []string{string(g), si(s.total), si(s.initial), si(s.internal), si(s.final), si(s.singleton), si(s.second), si(s.penult), sf(ratio(s.initial, s.total)), sf(ratio(s.final, s.total)), ss(classes)})
		for _, x := range classes {
			if strings.HasPrefix(x, "NEVER_") {
				cat.addRule("G2", "GLYPH_POSITION", string(g), strings.TrimPrefix(x, "NEVER_"), "token", 0, s.total, 0)
			}
		}
	}
	if err := writeTSV(filepath.Join(cat.Config.OutputDir, "GLYPH_INVENTORY.tsv"), []string{"glyph", "total_count", "token_occurrences_containing", "unique_token_types_containing", "physical_lines_containing", "folio_coverage", "section_coverage"}, invRows); err != nil {
		return err
	}
	if err := writeTSV(filepath.Join(cat.Config.OutputDir, "GLYPH_POSITION_RULES.tsv"), []string{"glyph", "total_count", "initial_count", "internal_count", "final_count", "singleton_count", "second_count", "penultimate_count", "p_initial_given_glyph", "p_final_given_glyph", "corpus_rules"}, posRows); err != nil {
		return err
	}
	cat.Summary["number_of_glyphs"] = si(len(c.Inventory))
	cat.Summary["glyphs_never_initial"] = si(neverI)
	cat.Summary["glyphs_never_final"] = si(neverF)
	cat.Summary["glyphs_never_internal"] = si(neverN)
	return cat.writeGlyphNgrams()
}

func (cat *Catalog) writeGlyphNgrams() error {
	c := cat.Primary
	big := map[string]int{}
	left, right := map[rune]int{}, map[rune]int{}
	opp := 0
	ngrams := map[int]map[string]int{2: {}, 3: {}, 4: {}}
	contexts := map[int]map[string]int{2: {}, 3: {}, 4: {}}
	for _, o := range c.Occurrences {
		r := []rune(o.Token)
		for i := 0; i+1 < len(r); i++ {
			k := string(r[i : i+2])
			big[k]++
			left[r[i]]++
			right[r[i+1]]++
			opp++
		}
		for n := 2; n <= 4; n++ {
			for i := 0; i+n <= len(r); i++ {
				ngrams[n][string(r[i:i+n])]++
				contexts[n][string(r[i:i+n-1])]++
			}
		}
	}
	rules := make([]Rule, 0, len(c.Inventory)*len(c.Inventory))
	rows := [][]string{}
	observed := 0
	for _, a := range c.Inventory {
		for _, b := range c.Inventory {
			k := string([]rune{a, b})
			n := big[k]
			exp := float64(left[a]) * float64(right[b]) / float64(max(1, opp))
			r := Rule{Level: "G3", RuleType: "GLYPH_BIGRAM", LHS: string(a), RHS: string(b), Context: "within_token", ObservedCount: n, OpportunityCount: left[a], ObservedProbability: ratio(n, left[a]), ExpectedCount: exp, EffectSize: effect(float64(n), exp), PRaw: poissonTail(n, exp, float64(n) >= exp), ObservedStatus: "OBSERVED", CorpusRule: "OBSERVED", Stability: "NOT_COMPARABLE", Provenance: provenance(c)}
			if n == 0 {
				r.ObservedStatus = "UNOBSERVED"
				r.CorpusRule = "NEVER_OBSERVED"
			} else {
				observed++
			}
			rules = append(rules, r)
		}
	}
	applyBH(rules)
	for _, r := range rules {
		rows = append(rows, []string{r.LHS, r.RHS, si(r.ObservedCount), si(r.OpportunityCount), sf(r.ObservedProbability), sf(r.ExpectedCount), sf(r.EffectSize), sf(r.PRaw), sf(r.QValue), r.ObservedStatus, r.CorpusRule, r.InferredStatus})
		cat.Rules = append(cat.Rules, r)
	}
	if err := writeTSV(filepath.Join(cat.Config.OutputDir, "GLYPH_BIGRAM_RULES.tsv"), []string{"left_glyph", "right_glyph", "observed_count", "left_opportunities", "p_right_given_left", "expected_count", "effect_size", "p_raw", "q_value", "observed_status", "corpus_rule", "inferred_status"}, rows); err != nil {
		return err
	}
	nrows := [][]string{}
	for n := 2; n <= 4; n++ {
		seqs := sortedTokens(ngrams[n])
		for _, seq := range seqs {
			rr := []rune(seq)
			ctx := string(rr[:len(rr)-1])
			count := ngrams[n][seq]
			nrows = append(nrows, []string{si(n), ctx, string(rr[len(rr)-1]), seq, si(count), si(contexts[n][ctx]), sf(ratio(count, contexts[n][ctx])), "OBSERVED", "OBSERVED"})
		}
		if n <= 3 {
			for ctx, cn := range contexts[n] {
				for _, g := range c.Inventory {
					seq := ctx + string(g)
					if ngrams[n][seq] == 0 {
						nrows = append(nrows, []string{si(n), ctx, string(g), seq, "0", si(cn), "0", "UNOBSERVED", "NEVER_OBSERVED"})
					}
				}
			}
		}
	}
	sort.Slice(nrows, func(i, j int) bool { return strings.Join(nrows[i], "\t") < strings.Join(nrows[j], "\t") })
	for _, row := range nrows {
		count := parseInt(row[4])
		ctxCount := parseInt(row[5])
		cat.addRule("G4", "GLYPH_NGRAM", row[1], row[2], "n="+row[0], count, ctxCount, 0)
	}
	if err := writeTSV(filepath.Join(cat.Config.OutputDir, "GLYPH_NGRAM_RULES.tsv"), []string{"n", "context", "continuation", "sequence", "count", "context_count", "conditional_probability", "observed_status", "corpus_rule"}, nrows); err != nil {
		return err
	}
	possible := len(c.Inventory) * len(c.Inventory)
	cat.Summary["possible_glyph_bigrams"] = si(possible)
	cat.Summary["observed_glyph_bigrams"] = si(observed)
	cat.Summary["unobserved_glyph_bigrams"] = si(possible - observed)
	cat.Summary["glyph_bigram_constraint_density"] = sf(1 - float64(observed)/float64(possible))
	cat.Summary["glyph_trigram_observed"] = si(len(ngrams[3]))
	cat.Summary["glyph_trigram_possible"] = si(len(c.Inventory) * len(c.Inventory) * len(c.Inventory))
	cat.Summary["glyph_trigram_constraint_density"] = sf(1 - float64(len(ngrams[3]))/math.Pow(float64(len(c.Inventory)), 3))
	return nil
}
