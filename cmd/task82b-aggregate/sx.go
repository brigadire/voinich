package main

import (
	"path/filepath"
	"sort"

	"zcore.dev/voinich/internal/task82b"
)

func writeSXOutputs(out string, pairs []task82b.PairUnit) error {
	reg := newTSV("sx_id", "description")
	reg.row("SX1_CONTRACTION_RATE", "mean fraction of expanded letters removed by the real abbreviation")
	reg.row("SX2_EXPANSION_AMBIGUITY", "mean number of distinct expansions per distinct abbreviated surface form")
	reg.row("SX3_ABBREVIATION_FAMILY_REUSE", "Gini coefficient of (abbr,expan) pair occurrence counts")
	reg.row("SX4_POSITIONAL_ABBREVIATION_PREFERENCE", "line-initial minus non-initial mean contraction rate")
	reg.row("SX5_CONTEXT_DEPENDENCE", "fraction of abbreviated-form types with >=2 observed distinct expansions")
	reg.row("SX6_MANY_TO_ONE_MAPPING", "distinct abbreviated-form types / distinct expansion types")
	reg.row("SX7_ABBREVIATION_GRAPH_DENSITY", "bipartite (abbr type, expan type) edge density")
	if err := reg.write(out, "SX_REGISTRY.tsv"); err != nil {
		return err
	}

	if err := writeSXValidation(out); err != nil {
		return err
	}

	res := newTSV("scope", "sx_id", "value", "note")
	if pairs == nil {
		return res.write(out, "SX_RESULTS.tsv")
	}
	writeScope := func(name string, ps []task82b.PairUnit) {
		for _, m := range task82b.ComputeSX(ps) {
			res.row(name, m.ID, fstr(m.Value), m.Note)
		}
	}
	writeScope("combined", pairs)
	byFile := map[string][]task82b.PairUnit{}
	for _, p := range pairs {
		byFile[filepath.Base(p.File)] = append(byFile[filepath.Base(p.File)], p)
	}
	var files []string
	for fl := range byFile {
		files = append(files, fl)
	}
	sort.Strings(files)
	for _, fl := range files {
		writeScope(fl, byFile[fl])
	}
	return res.write(out, "SX_RESULTS.tsv")
}

// writeSXValidation is a self-consistency check (TASK82B_DESIGN.md sec.8):
// SX1 (contraction rate) must be 0 for an identity (no-deletion) pair and
// 1 for a whole-word special-sign pair, using the real SX/DeletionCount
// code path on two minimal synthetic pair sets.
func writeSXValidation(out string) error {
	identity := []task82b.PairUnit{
		{AbbrText: "roma", ExpanText: "roma"},
		{AbbrText: "papa", ExpanText: "papa"},
	}
	wholeWord := []task82b.PairUnit{
		{AbbrText: "Ⰰ", ExpanText: "per", HasMark: true},
		{AbbrText: "Ⰱ", ExpanText: "non", HasMark: true},
	}
	sx1 := func(ps []task82b.PairUnit) float64 {
		for _, m := range task82b.ComputeSX(ps) {
			if m.ID == "SX1_CONTRACTION_RATE" {
				return m.Value
			}
		}
		return -1
	}
	idVal := sx1(identity)
	wwVal := sx1(wholeWord)
	w := newTSV("control", "description", "sx1_value", "expected", "pass")
	w.row("IDENTITY", "abbr==expan, no deletion", fstr(idVal), "0", bstr(idVal == 0))
	w.row("WHOLE_WORD_MARK", "abbr is a single mark standing for the whole word", fstr(wwVal), "1", bstr(wwVal == 1))
	return w.write(out, "SX_VALIDATION.tsv")
}
