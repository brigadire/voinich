package main

import (
	"path/filepath"
	"sort"
	"strings"

	"zcore.dev/voinich/internal/task82b"
)

type shorthandIndex struct {
	// scale -> variant -> records (1 for deterministic variants, 3 for NULL_* seeds)
	byScaleVariant map[string]map[string][]Rec
}

func indexShorthand(recs []Rec) shorthandIndex {
	idx := shorthandIndex{byScaleVariant: map[string]map[string][]Rec{}}
	for _, r := range recs {
		if r.Kind != "shorthand_variant" {
			continue
		}
		if idx.byScaleVariant[r.ShorthandScale] == nil {
			idx.byScaleVariant[r.ShorthandScale] = map[string][]Rec{}
		}
		idx.byScaleVariant[r.ShorthandScale][r.ShorthandVariant] = append(idx.byScaleVariant[r.ShorthandScale][r.ShorthandVariant], r)
	}
	return idx
}

func writeShorthandOutputs(out string, recs []Rec, pairs []task82b.PairUnit) error {
	if err := writeShorthandProvenance(out, pairs); err != nil {
		return err
	}
	if err := writeAbbreviationRegistry(out, pairs); err != nil {
		return err
	}
	if err := writeAlignmentStats(out, pairs); err != nil {
		return err
	}
	if err := writeShorthandRecovery(out, pairs); err != nil {
		return err
	}
	if err := writeSXOutputs(out, pairs); err != nil {
		return err
	}

	idx := indexShorthand(recs)
	if len(idx.byScaleVariant) == 0 {
		return nil
	}

	beforeAfter := newTSV("scale", "metric_id", "classification", "f2_expanded", "expanded_available", "f2_abbreviated", "abbreviated_available", "tokens_expanded", "tokens_abbreviated")
	trajectories := newTSV("scale", "metric_id", "classification", "delta", "both_available")
	nullCmp := newTSV("scale", "null_kind", "metric_id", "observed_delta", "null_mean_delta", "null_sd_delta", "effect_size", "p_value", "n_replicates")

	var scales []string
	for s := range idx.byScaleVariant {
		scales = append(scales, s)
	}
	sort.Strings(scales)

	// stability[metric][scale] = delta (available only if both sides present)
	stability := map[string]map[string]struct {
		delta     float64
		available bool
	}{}

	for _, scale := range scales {
		byVariant := idx.byScaleVariant[scale]
		expRecs := byVariant["EXPANDED"]
		abbrRecs := byVariant["ABBREVIATED"]
		if len(expRecs) == 0 || len(abbrRecs) == 0 {
			continue
		}
		expRec, abbrRec := expRecs[0], abbrRecs[0]
		for _, metricID := range task82b.AllMetricIDs() {
			before, beforeOK := expRec.metric(metricID)
			after, afterOK := abbrRec.metric(metricID)
			classification := metricClass(metricID)
			beforeAfter.row(scale, metricID, classification, fOr(before, beforeOK), bstr(beforeOK), fOr(after, afterOK), bstr(afterOK), istr(expRec.ChosenCount), istr(abbrRec.ChosenCount))
			bothAvail := beforeOK && afterOK
			delta := 0.0
			if bothAvail {
				delta = after - before
			}
			trajectories.row(scale, metricID, classification, fOr(delta, bothAvail), bstr(bothAvail))
			if stability[metricID] == nil {
				stability[metricID] = map[string]struct {
					delta     float64
					available bool
				}{}
			}
			stability[metricID][scale] = struct {
				delta     float64
				available bool
			}{delta, bothAvail}

			for _, nullKind := range []string{"NULL_RANDOM_DELETION_MATCHED", "NULL_FREQUENCY_MATCHED_DELETION", "NULL_POSITION_MATCHED"} {
				nullRecs := byVariant[nullKind]
				vals := valuesFor(nullRecs, metricID)
				if len(vals) == 0 || !bothAvail {
					continue
				}
				cmp := task82b.CompareToNull(after, vals)
				nullCmp.row(scale, nullKind, metricID, fstr(delta), fstr(cmp.NullMean-before), fstr(cmp.NullSD), fstr(cmp.EffectSize), fstr(cmp.PValue), istr(cmp.N))
			}
		}
	}
	if err := beforeAfter.write(out, "SHORTHAND_F2_BEFORE_AFTER.tsv"); err != nil {
		return err
	}
	if err := trajectories.write(out, "SHORTHAND_F2_TRAJECTORIES.tsv"); err != nil {
		return err
	}
	if err := nullCmp.write(out, "SHORTHAND_NULL_COMPARISON.tsv"); err != nil {
		return err
	}

	return writeShorthandStability(out, stability, scales)
}

// writeShorthandStability checks cross-document (chapter) consistency and
// convergence toward the combined scale (task82b.txt sec.24/25/62). No
// cross-tradition data was obtained (TASK82B_DESIGN.md sec.4), so
// SHORTHAND_GENERAL_EFFECT is never assigned here -- at most
// SYSTEM_SPECIFIC_EFFECT (stable within this one manuscript/tradition).
func writeShorthandStability(out string, stability map[string]map[string]struct {
	delta     float64
	available bool
}, scales []string) error {
	w := newTSV("metric_id", "classification", "n_chapters_available", "n_chapters_consistent_sign", "combined_delta", "convergence")
	var chapters []string
	for _, s := range scales {
		if s != "combined" {
			chapters = append(chapters, s)
		}
	}
	for _, metricID := range task82b.AllMetricIDs() {
		byScale := stability[metricID]
		nAvail, nPos, nNeg := 0, 0, 0
		for _, ch := range chapters {
			c, ok := byScale[ch]
			if !ok || !c.available {
				continue
			}
			nAvail++
			if c.delta > 0 {
				nPos++
			} else if c.delta < 0 {
				nNeg++
			}
		}
		consistentSign := nAvail > 0 && (nPos == 0 || nNeg == 0)
		// A metric with delta==0 in every chapter (2DL1/BP1/LS1-4/cs6 are
		// all trivially 0 for the PAIR_DEFINED one-word-per-line
		// convention: no within-line multi-token structure ever exists
		// for either EXPANDED or ABBREVIATED) vacuously satisfies
		// consistentSign but carries no evidence at all; it must not be
		// reported as a detected effect (TASK82B_DESIGN.md sec.5).
		allZero := nAvail > 0 && nPos == 0 && nNeg == 0
		combined, combinedOK := byScale["combined"]
		classification := "NOT_STABLE"
		switch {
		case allZero:
			classification = "STRUCTURALLY_DEGENERATE_NO_VARIATION"
		case nAvail >= 3 && consistentSign:
			classification = "SYSTEM_SPECIFIC_EFFECT"
		}
		convergence := "INCONCLUSIVE"
		if combinedOK && nAvail > 0 {
			combinedSign := signOf(combined.delta)
			chapterSign := 0
			if nPos > nNeg {
				chapterSign = 1
			} else if nNeg > nPos {
				chapterSign = -1
			}
			switch {
			case combinedSign == chapterSign && combinedSign != 0:
				convergence = "CONVERGED"
			case combinedSign != chapterSign:
				convergence = "NOT_CONVERGED"
			}
		}
		w.row(metricID, classification, istr(nAvail), istr(boolToInt(consistentSign)), fOr(combined.delta, combinedOK), convergence)
	}
	return w.write(out, "SHORTHAND_STABILITY.tsv")
}

func signOf(v float64) int {
	switch {
	case v > 0:
		return 1
	case v < 0:
		return -1
	default:
		return 0
	}
}

func writeShorthandProvenance(out string, pairs []task82b.PairUnit) error {
	w := newTSV("corpus_id", "manuscript_source", "date", "language", "notation_tradition", "abbreviated_representation", "expanded_representation", "alignment_type", "source_reference", "license", "checksum_note", "preprocessing", "known_limitations")
	w.row(
		"BDD_koeln-edd-c-119",
		"Burchard of Worms, Decretum; Koeln Erzbischoefliche Dioezesan- und Dombibliothek Cod. 119, books 6/7/11/12/13",
		"manuscript ~11th c.; digital edition published 2024-01-29",
		"Latin",
		"scribal abbreviation (TEI <choice><abbr>/<expan>)",
		"TEI <abbr> branch, <g> marks -> reserved Glagolitic placeholders (internal/task82b/teipair.go)",
		"TEI <expan> branch, lowercased/trimmed",
		"n:1/1:n/n:m per <choice> (internal/task82b.PairUnit)",
		"github.com/burchards-dekret-digital/website, data/mss/koeln-edd-c-119/Tei/v1/*.xml, repo commit 29f9cb1c34cc9ee3c50e75a6e3e99cfa4a2bc362",
		"CC BY 4.0 (stated in each file's own TEI teiHeader)",
		"see research/phase2/fingerprint/CONTROL_PROVENANCE.tsv for the original per-file XML SHA-256 chain (reused unmodified for Task82b)",
		"cmd/tei-abbr-extract-equivalent walk (internal/task82b/teipair.go), independent of Task79c's frozen abbr-only tool",
		"single manuscript/scribe/notation tradition only; no second independent paired historical corpus was obtainable within this task's scope (TASK82B_DESIGN.md sec.4)",
	)
	return w.write(out, "SHORTHAND_CORPUS_PROVENANCE.tsv")
}

func writeAbbreviationRegistry(out string, pairs []task82b.PairUnit) error {
	w := newTSV("abbr_text", "expan_text", "count", "class", "uses_mark", "mark_combining", "ambiguous_expan", "info_class")
	if pairs == nil {
		return w.write(out, "ABBREVIATION_OPERATION_REGISTRY.tsv")
	}
	rows := task82b.BuildOperationRegistry(pairs)
	sort.Slice(rows, func(i, j int) bool {
		if rows[i].Count != rows[j].Count {
			return rows[i].Count > rows[j].Count
		}
		return rows[i].AbbrText < rows[j].AbbrText
	})
	for _, row := range rows {
		w.row(row.AbbrText, row.ExpanText, istr(row.Count), row.Class, bstr(row.UsesMark), bstr(row.MarkCombining), bstr(row.AmbiguousExpan), row.InfoClass)
	}
	return w.write(out, "ABBREVIATION_OPERATION_REGISTRY.tsv")
}

func writeAlignmentStats(out string, pairs []task82b.PairUnit) error {
	w := newTSV("scope", "n_pairs", "n_distinct_abbr_types", "n_distinct_expan_types", "mean_deletion_rate", "n_with_mark", "n_mark_combining")
	if pairs == nil {
		return w.write(out, "SHORTHAND_ALIGNMENT_STATS.tsv")
	}
	writeScope := func(name string, ps []task82b.PairUnit) {
		abbrTypes, expanTypes := map[string]bool{}, map[string]bool{}
		nMark, nCombining := 0, 0
		totalDel, totalLen, n := 0, 0, 0
		for _, p := range ps {
			abbrTypes[p.AbbrText] = true
			expanTypes[p.ExpanText] = true
			if p.HasMark {
				nMark++
			}
			if p.MarkIsCombining {
				nCombining++
			}
			if p.ExpanText != "" {
				totalDel += task82b.DeletionCount(p)
				totalLen += len([]rune(p.ExpanText))
				n++
			}
		}
		rate := 0.0
		if totalLen > 0 {
			rate = float64(totalDel) / float64(totalLen)
		}
		w.row(name, istr(len(ps)), istr(len(abbrTypes)), istr(len(expanTypes)), fstr(rate), istr(nMark), istr(nCombining))
	}
	writeScope("combined", pairs)
	byFile := map[string][]task82b.PairUnit{}
	for _, p := range pairs {
		byFile[filepath.Base(p.File)] = append(byFile[filepath.Base(p.File)], p)
	}
	var files []string
	for f := range byFile {
		files = append(files, f)
	}
	sort.Strings(files)
	for _, fl := range files {
		writeScope(fl, byFile[fl])
	}
	return w.write(out, "SHORTHAND_ALIGNMENT_STATS.tsv")
}

func writeShorthandRecovery(out string, pairs []task82b.PairUnit) error {
	w := newTSV("abbr_text", "n_observed_expansions", "expansions", "ambiguous", "info_class")
	if pairs == nil {
		return w.write(out, "SHORTHAND_RECOVERY.tsv")
	}
	expansOf := map[string]map[string]bool{}
	for _, p := range pairs {
		if expansOf[p.AbbrText] == nil {
			expansOf[p.AbbrText] = map[string]bool{}
		}
		expansOf[p.AbbrText][p.ExpanText] = true
	}
	var abbrs []string
	for a := range expansOf {
		abbrs = append(abbrs, a)
	}
	sort.Strings(abbrs)
	for _, a := range abbrs {
		exps := expansOf[a]
		var list []string
		for e := range exps {
			list = append(list, e)
		}
		sort.Strings(list)
		ambiguous := len(list) > 1
		info := "SELF_SUFFICIENT"
		if ambiguous {
			info = "AMBIGUOUS_WITHOUT_CONTEXT"
		}
		w.row(a, istr(len(list)), joinComma(list), bstr(ambiguous), info)
	}
	return w.write(out, "SHORTHAND_RECOVERY.tsv")
}

func joinComma(xs []string) string {
	return strings.Join(xs, "; ")
}
