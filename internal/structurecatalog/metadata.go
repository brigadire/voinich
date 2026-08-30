package structurecatalog

import (
	"math"
	"path/filepath"
	"sort"
)

func (cat *Catalog) writeMetadata() error {
	if !cat.Primary.MetadataAvailable {
		return writeEmptyMetadata(cat.Config.OutputDir)
	}
	if err := cat.writeDimension("locus", cat.Primary.LocusTypes, "LOCUS_RULES.tsv", "D1"); err != nil {
		return err
	}
	if err := cat.writeDimension("folio", cat.Primary.Folios, "FOLIO_RULES.tsv", "D2"); err != nil {
		return err
	}
	if err := cat.writeDimension("section", cat.Primary.Sections, "SECTION_RULES.tsv", "D3"); err != nil {
		return err
	}
	return nil
}

func writeEmptyMetadata(out string) error {
	h := []string{"entity_type", "entity", "family_id", "dimension", "value", "entity_total", "observed_count", "expected_count", "effect_size", "p_raw", "q_value", "observed_status", "corpus_rule", "inferred_status"}
	for _, f := range []string{"LOCUS_RULES.tsv", "FOLIO_RULES.tsv", "SECTION_RULES.tsv"} {
		if err := writeTSV(filepath.Join(out, f), h, nil); err != nil {
			return err
		}
	}
	return nil
}

func (cat *Catalog) writeDimension(dim string, values []string, file, level string) error {
	c := cat.Primary
	global := map[string]int{}
	tok := map[string]map[string]int{}
	fam := map[int]map[string]int{}
	famTotal := map[int]int{}
	value := func(o Occurrence) string {
		switch dim {
		case "locus":
			return o.Meta.LocusType
		case "folio":
			return o.Meta.Folio
		default:
			return o.Meta.Section
		}
	}
	for _, o := range c.Occurrences {
		v := value(o)
		if v == "" {
			continue
		}
		global[v]++
		if tok[o.Token] == nil {
			tok[o.Token] = map[string]int{}
		}
		tok[o.Token][v]++
		id := cat.TokenFamily[o.Token]
		if fam[id] == nil {
			fam[id] = map[string]int{}
		}
		fam[id][v]++
		famTotal[id]++
	}
	type entity struct {
		kind, name string
		id, total  int
		counts     map[string]int
	}
	es := []entity{}
	for _, t := range sortedTokens(c.Counts) {
		if c.Counts[t] >= cat.Config.MinFrequency {
			total := 0
			for _, n := range tok[t] {
				total += n
			}
			es = append(es, entity{"TOKEN", t, cat.TokenFamily[t], total, tok[t]})
		}
	}
	for _, f := range cat.Families {
		if famTotal[f.ID] >= cat.Config.MinFrequency {
			es = append(es, entity{"FAMILY", si(f.ID), f.ID, famTotal[f.ID], fam[f.ID]})
		}
	}
	rules := []Rule{}
	refs := [][2]int{} // entity index, value index
	for ei, e := range es {
		for vi, v := range values {
			n := e.counts[v]
			exp := float64(e.total) * float64(global[v]) / float64(len(c.Occurrences))
			typ := e.kind + "_BY_" + dim
			r := Rule{Level: level, RuleType: typ, LHS: e.name, RHS: v, Context: dim, ObservedCount: n, OpportunityCount: e.total, ObservedProbability: ratio(n, e.total), ExpectedCount: exp, EffectSize: effect(float64(n), exp), PRaw: poissonTail(n, exp, float64(n) >= exp), ObservedStatus: "OBSERVED", CorpusRule: "OBSERVED", Stability: "NOT_COMPARABLE", Provenance: provenance(c)}
			if n == 0 {
				r.ObservedStatus = "UNOBSERVED"
				r.CorpusRule = "NEVER_IN_" + dim
			}
			rules = append(rules, r)
			refs = append(refs, [2]int{ei, vi})
		}
	}
	applyBH(rules)
	rows := [][]string{}
	exclusiveFamilies := map[int]bool{}
	for i, r := range rules {
		e := es[refs[i][0]]
		v := values[refs[i][1]]
		corpusRule := r.CorpusRule
		if r.ObservedCount == e.total && e.total > 0 {
			corpusRule = "ONLY_IN_" + dim
			if e.kind == "FAMILY" {
				exclusiveFamilies[e.id] = true
			}
		}
		rows = append(rows, []string{e.kind, e.name, si(e.id), dim, v, si(e.total), si(r.ObservedCount), sf(r.ExpectedCount), sf(r.EffectSize), sf(r.PRaw), sf(r.QValue), r.ObservedStatus, corpusRule, r.InferredStatus})
		r.CorpusRule = corpusRule
		cat.Rules = append(cat.Rules, r)
	}
	sort.Slice(rows, func(i, j int) bool {
		for k := 0; k < 5; k++ {
			if rows[i][k] != rows[j][k] {
				return rows[i][k] < rows[j][k]
			}
		}
		return false
	})
	if err := writeTSV(filepath.Join(cat.Config.OutputDir, file), []string{"entity_type", "entity", "family_id", "dimension", "value", "entity_total", "observed_count", "expected_count", "effect_size", "p_raw", "q_value", "observed_status", "corpus_rule", "inferred_status"}, rows); err != nil {
		return err
	}
	cat.Summary[dim+"_values"] = si(len(values))
	cat.Summary[dim+"_exclusive_families"] = si(len(exclusiveFamilies))
	cat.Summary[dim+"_metadata_available"] = "true"
	_ = math.NaN()
	return nil
}
