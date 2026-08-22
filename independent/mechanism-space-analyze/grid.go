package main

import (
	"fmt"
	"os"
	"strings"

	"zcore.dev/voinich/internal/mechanismspace"
)

// GridEntry names one frozen mechanism-space grid point (task66 section
// 33): Name is the identifier used throughout every TSV; Config is the
// exact, plaintext-independent parameterization.
type GridEntry struct {
	Name   string
	Config mechanismspace.Config
}

// BuildGrid is task66 section 33's small, preregistered parameter grid.
// It is fixed before any corpus is transformed and is never expanded
// after DESIGN_FROZEN (section 73).
func BuildGrid() []GridEntry {
	var g []GridEntry
	add := func(name string, c mechanismspace.Config) { g = append(g, GridEntry{Name: name, Config: c}) }

	add("M0_IDENTITY", mechanismspace.Config{Family: "M0"})
	add("M1_MONOALPHABETIC", mechanismspace.Config{Family: "M1"})

	for _, h := range []int{2, 4, 8} {
		add(fmt.Sprintf("M2_HOMOPHONY_H%d", h), mechanismspace.Config{Family: "M2", Homophones: h})
	}

	for _, lvl := range []mechanismspace.GrammarLevel{mechanismspace.GrammarLow, mechanismspace.GrammarMedium, mechanismspace.GrammarHigh} {
		add("M3_FORM_"+string(lvl), mechanismspace.Config{Family: "M3", Grammar: lvl})
	}

	for _, k := range []int{2, 4, 8} {
		for _, u := range []mechanismspace.StateUpdate{mechanismspace.UpdateA, mechanismspace.UpdateB, mechanismspace.UpdateC} {
			add(fmt.Sprintf("M4_STATE_K%d_%s", k, u), mechanismspace.Config{Family: "M4", StateCount: k, Update: u})
		}
	}

	for _, scale := range []int{5, 20, 80} {
		add(fmt.Sprintf("M5_DRIFT_N%d", scale), mechanismspace.Config{Family: "M5", StateCount: 1000, DriftScale: scale})
	}

	for _, k := range []int{2, 5} {
		add(fmt.Sprintf("M6_MACRO_K%d", k), mechanismspace.Config{Family: "M6", MacroStates: k})
	}

	for _, k := range []int{2, 5} {
		for _, scale := range []int{5, 20, 80} {
			add(fmt.Sprintf("M7_MIXED_K%d_N%d", k, scale), mechanismspace.Config{Family: "M7", MacroStates: k, StateCount: 1000, DriftScale: scale})
		}
	}

	for _, grp := range []mechanismspace.Grouping{mechanismspace.FixedGrouping, mechanismspace.RandomGrouping, mechanismspace.StateGrouping} {
		add("M8_BOUNDARY_"+string(grp), mechanismspace.Config{Family: "M8", InputMode: mechanismspace.Stream, Grouping: grp, GroupLen: 4, StateCount: 4})
	}

	for _, grp := range []mechanismspace.Grouping{mechanismspace.FixedGrouping, mechanismspace.RandomGrouping, mechanismspace.StateGrouping} {
		add("M9_GROUP_FORM_"+string(grp), mechanismspace.Config{Family: "M9", InputMode: mechanismspace.Stream, Grouping: grp, GroupLen: 4, Grammar: mechanismspace.GrammarMedium})
	}

	for _, k := range []int{2, 4, 8} {
		add(fmt.Sprintf("M10_STATEFUL_FORM_K%d", k), mechanismspace.Config{Family: "M10", InputMode: mechanismspace.Stream, Grouping: mechanismspace.StateGrouping, GroupLen: 4, Grammar: mechanismspace.GrammarMedium, StateCount: k})
	}

	for _, k := range []int{2, 5} {
		add(fmt.Sprintf("M11_MIXED_FORM_K%d", k), mechanismspace.Config{Family: "M11", InputMode: mechanismspace.Stream, Grouping: mechanismspace.StateGrouping, GroupLen: 4, Grammar: mechanismspace.GrammarMedium, StateCount: 4, MacroStates: k})
	}

	return g
}

// AblationEntries is task66 section 26's compositional ablation matrix
// for the M9-M11 STREAM architecture: it varies only whether macro-state
// (M), local state (S) and constrained formation (G) are present, holding
// everything else (grouping rule, group length, seed) fixed, so the
// marginal contribution of each operation class can be read off directly.
func AblationEntries() []GridEntry {
	base := mechanismspace.Config{InputMode: mechanismspace.Stream, Grouping: mechanismspace.StateGrouping, GroupLen: 4}
	entry := func(name string, macro, state int, grammar mechanismspace.GrammarLevel) GridEntry {
		c := base
		c.Family = "ABLATION_" + name
		c.MacroStates = macro
		c.StateCount = state
		c.Grammar = grammar
		return GridEntry{Name: name, Config: c}
	}
	return []GridEntry{
		entry("G_ONLY", 1, 1, mechanismspace.GrammarMedium),
		entry("S_ONLY", 1, 4, mechanismspace.NoGrammar),
		entry("M_ONLY", 5, 1, mechanismspace.NoGrammar),
		entry("G_PLUS_S", 1, 4, mechanismspace.GrammarMedium),
		entry("M_PLUS_S", 5, 4, mechanismspace.NoGrammar),
		entry("M_PLUS_G", 5, 1, mechanismspace.GrammarMedium),
		entry("M_PLUS_S_PLUS_G", 5, 4, mechanismspace.GrammarMedium),
	}
}

// WriteGridTSV writes MECHANISM_GRID.tsv (task66 section 76).
func WriteGridTSV(path string, grid []GridEntry, devReplicates, finalReplicates int) error {
	var b strings.Builder
	b.WriteString("mechanism\tfamily\tinput_mode\tparameters\tdevelopment_replicates\tfinal_replicates\n")
	for _, e := range grid {
		b.WriteString(fmt.Sprintf("%s\t%s\t%s\t%s\t%d\t%d\n", e.Name, e.Config.Family, e.Config.InputMode, paramString(e.Config), devReplicates, finalReplicates))
	}
	return os.WriteFile(path, []byte(b.String()), 0644)
}

func paramString(c mechanismspace.Config) string {
	var parts []string
	if c.StateCount > 1 {
		parts = append(parts, fmt.Sprintf("state=%d", c.StateCount))
	}
	if c.MacroStates > 1 {
		parts = append(parts, fmt.Sprintf("macro=%d", c.MacroStates))
	}
	if c.DriftScale > 1 {
		parts = append(parts, fmt.Sprintf("drift=%d", c.DriftScale))
	}
	if c.Update != "" {
		parts = append(parts, "update="+string(c.Update))
	}
	if c.Grammar != "" {
		parts = append(parts, "grammar="+string(c.Grammar))
	}
	if c.Grouping != "" {
		parts = append(parts, "grouping="+string(c.Grouping)+fmt.Sprintf(",len=%d", c.GroupLen))
	}
	if c.Homophones > 1 {
		parts = append(parts, fmt.Sprintf("H=%d", c.Homophones))
	}
	if len(parts) == 0 {
		return "fixed"
	}
	return strings.Join(parts, ";")
}

// WriteComplexityTSV writes MECHANISM_COMPLEXITY.tsv (task66 section 31).
func WriteComplexityTSV(path string, grid []GridEntry) error {
	var b strings.Builder
	b.WriteString("mechanism\tstates\tsymbol_classes\ttransition_parameters\toutput_rules\tstochastic_distributions\tinformation_status\n")
	for _, e := range grid {
		m := mechanismspace.BuildMetadata(e.Config)
		b.WriteString(fmt.Sprintf("%s\t%d\t%d\t%d\t%d\t%d\t%s\n", e.Name, m.StateCount, m.SymbolClasses, m.TransitionParameters, m.OutputRules, m.StochasticDistributions, m.Information))
	}
	return os.WriteFile(path, []byte(b.String()), 0644)
}
