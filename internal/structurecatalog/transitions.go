package structurecatalog

import (
	"math"
	"path/filepath"
	"sort"
	"strings"
)

func (cat *Catalog) writeTransitions() error {
	c := cat.Primary
	obs := map[string]map[string]int{}
	pred, succ := map[string]int{}, map[string]int{}
	opps := 0
	pairs := map[string]int{}
	triples := map[string]int{}
	ctx2 := map[string]int{}
	for _, line := range c.Lines {
		for i := 0; i+1 < len(line); i++ {
			a, b := line[i], line[i+1]
			if obs[a] == nil {
				obs[a] = map[string]int{}
			}
			obs[a][b]++
			pred[a]++
			succ[b]++
			opps++
			pairs[a+"\x00"+b]++
		}
		for i := 0; i+2 < len(line); i++ {
			ctx := line[i] + "\x00" + line[i+1]
			triples[ctx+"\x00"+line[i+2]]++
			ctx2[ctx]++
		}
	}
	all := sortedTokens(c.Counts)
	frequent := []string{}
	for _, t := range all {
		if c.Counts[t] >= cat.Config.MinFrequency {
			frequent = append(frequent, t)
		}
	}
	observedRules := []Rule{}
	orows := [][]string{}
	for _, a := range all {
		bs := make([]string, 0, len(obs[a]))
		for b := range obs[a] {
			bs = append(bs, b)
		}
		sort.Strings(bs)
		for _, b := range bs {
			n := obs[a][b]
			exp := float64(pred[a]) * float64(succ[b]) / float64(max(1, opps))
			r := transitionRule(c, a, b, n, pred[a], exp)
			observedRules = append(observedRules, r)
		}
	}
	applyBH(observedRules)
	preferred, avoided := 0, 0
	for _, r := range observedRules {
		if c.Counts[r.LHS] < cat.Config.MinFrequency || c.Counts[r.RHS] < cat.Config.MinFrequency {
			r.InferredStatus = "INSUFFICIENT_SUPPORT"
		}
		orows = append(orows, []string{r.LHS, r.RHS, si(c.Counts[r.LHS]), si(c.Counts[r.RHS]), si(r.ObservedCount), sf(r.ObservedProbability), sf(ratio(r.ObservedCount, succ[r.RHS])), sf(ratio(succ[r.RHS], opps)), sf(r.EffectSize), sf(1 / r.EffectSize), sf(r.ExpectedCount), sf(r.PRaw), sf(r.QValue), r.InferredStatus})
		if strings.Contains(r.InferredStatus, "PREFERRED") {
			preferred++
		}
		if r.InferredStatus == "DEPLETED" {
			avoided++
		}
		cat.Rules = append(cat.Rules, r)
	}
	if err := writeTSV(filepath.Join(cat.Config.OutputDir, "TOKEN_TRANSITIONS_OBSERVED.tsv"), []string{"token_A", "token_B", "count_A", "count_B", "count_AB", "p_B_given_A", "p_A_given_B_as_successor", "baseline_p_B", "enrichment", "depletion_inverse", "expected_AB", "p_raw", "q_value", "inferred_status"}, orows); err != nil {
		return err
	}
	unRules := []Rule{}
	for _, a := range frequent {
		for _, b := range frequent {
			if obs[a][b] > 0 {
				continue
			}
			exp := float64(pred[a]) * float64(succ[b]) / float64(max(1, opps))
			unRules = append(unRules, transitionRule(c, a, b, 0, pred[a], exp))
		}
	}
	applyBH(unRules)
	urows := [][]string{}
	for _, r := range unRules {
		if r.OpportunityCount < cat.Config.MinFrequency {
			r.InferredStatus = "INSUFFICIENT_SUPPORT"
		}
		urows = append(urows, []string{r.LHS, r.RHS, si(c.Counts[r.LHS]), si(c.Counts[r.RHS]), "0", sf(r.ExpectedCount), sf(math.Exp(-r.ExpectedCount)), sf(r.PRaw), sf(r.QValue), "NEVER_OBSERVED", r.InferredStatus})
		if strings.Contains(r.InferredStatus, "AVOIDED") {
			avoided++
		}
		cat.Rules = append(cat.Rules, r)
	}
	if err := writeTSV(filepath.Join(cat.Config.OutputDir, "TOKEN_TRANSITIONS_UNOBSERVED.tsv"), []string{"token_A", "token_B", "count_A", "count_B", "observed_AB", "expected_AB", "null_probability_zero", "p_raw", "q_value", "corpus_rule", "inferred_status"}, urows); err != nil {
		return err
	}
	if err := writeComplement(filepath.Join(cat.Config.OutputDir, "TOKEN_TRANSITION_COMPLEMENT.json.gz"), all, obs); err != nil {
		return err
	}
	if err := cat.writeHigherOrder(triples, ctx2, pairs, pred); err != nil {
		return err
	}
	cat.Summary["frequent_token_threshold"] = si(cat.Config.MinFrequency)
	cat.Summary["frequent_tokens"] = si(len(frequent))
	cat.Summary["possible_frequent_token_transitions"] = si(len(frequent) * len(frequent))
	freqObserved := 0
	for _, a := range frequent {
		for _, b := range frequent {
			if obs[a][b] > 0 {
				freqObserved++
			}
		}
	}
	cat.Summary["observed_frequent_token_transitions"] = si(freqObserved)
	cat.Summary["unobserved_frequent_token_transitions"] = si(len(frequent)*len(frequent) - freqObserved)
	cat.Summary["all_observed_token_transitions"] = si(len(observedRules))
	cat.Summary["statistically_preferred_transitions"] = si(preferred)
	cat.Summary["statistically_avoided_or_depleted_transitions"] = si(avoided)
	cat.Summary["frequent_transition_constraint_density"] = sf(1 - float64(freqObserved)/float64(len(frequent)*len(frequent)))
	return nil
}

func transitionRule(c Corpus, a, b string, n, opp int, exp float64) Rule {
	r := Rule{Level: "T3", RuleType: "TOKEN_TRANSITION", LHS: a, RHS: b, Context: "within_physical_line", ObservedCount: n, OpportunityCount: opp, ObservedProbability: ratio(n, opp), ExpectedCount: exp, EffectSize: effect(float64(n), exp), PRaw: poissonTail(n, exp, float64(n) >= exp), ObservedStatus: "OBSERVED", CorpusRule: "OBSERVED", Stability: "NOT_COMPARABLE", Provenance: provenance(c)}
	if n == 0 {
		r.ObservedStatus = "UNOBSERVED"
		r.CorpusRule = "NEVER_OBSERVED"
	}
	return r
}

func (cat *Catalog) writeHigherOrder(triples, ctx2, pairs map[string]int, pred map[string]int) error {
	keys := make([]string, 0, len(triples))
	for k := range triples {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	rows := [][]string{}
	for _, k := range keys {
		p := strings.Split(k, "\x00")
		a, b, d := p[0], p[1], p[2]
		n := triples[k]
		ctx := a + "\x00" + b
		baseN := pairs[b+"\x00"+d]
		rows = append(rows, []string{a + " " + b, d, si(n), si(ctx2[ctx]), sf(ratio(n, ctx2[ctx])), si(baseN), si(pred[b]), sf(ratio(baseN, pred[b])), sf(ratio(n, ctx2[ctx]) - ratio(baseN, pred[b])), "OBSERVED"})
		cat.addRule("T4", "TOKEN_HIGHER_ORDER", a+" "+b, d, "within_physical_line", n, ctx2[ctx], 0)
	}
	return writeTSV(filepath.Join(cat.Config.OutputDir, "TOKEN_HIGHER_ORDER_RULES.tsv"), []string{"context_AB", "continuation_C", "count_ABC", "count_AB", "p_C_given_AB", "count_BC", "count_B_as_predecessor", "p_C_given_B", "probability_difference", "observed_status"}, rows)
}
