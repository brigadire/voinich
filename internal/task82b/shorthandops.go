package task82b

import "strings"

// OperationRow is one row of ABBREVIATION_OPERATION_REGISTRY.tsv
// (task82b.txt sec.13): one distinct (abbreviated, expanded) surface-form
// pair, classified functionally from the data itself.
type OperationRow struct {
	AbbrText       string
	ExpanText      string
	Count          int
	Class          string // deletion/contraction/suspension/superscript/special_sign/mark_only/other
	UsesMark       bool
	MarkCombining  bool
	AmbiguousExpan bool   // this AbbrText also maps to >=1 other ExpanText elsewhere in the corpus
	InfoClass      string // sec.23: SELF_SUFFICIENT / CONVENTION_DEPENDENT / CONTEXT_DEPENDENT / AMBIGUOUS_WITHOUT_CONTEXT
}

// stripMarks removes the reserved Glagolitic placeholder runes (U+2C00-
// U+2C5F) that stand for a <g> combining mark or special sign, so the
// remaining literal letters can be compared against the expansion.
func stripMarks(s string) string {
	var b strings.Builder
	for _, r := range s {
		if r >= 0x2C00 && r <= 0x2C5F {
			continue
		}
		b.WriteRune(r)
	}
	return b.String()
}

// classifyPair assigns one operation class to a single (abbr, expan)
// occurrence, from the surface forms alone (task82b.txt sec.13: names
// follow the data, not an imposed a-priori taxonomy).
func classifyPair(abbrText, expanText string, usesMark bool) string {
	stripped := strings.ToLower(stripMarks(abbrText))
	expanLower := strings.ToLower(strings.TrimSpace(expanText))
	stripped = strings.TrimSpace(stripped)
	switch {
	case stripped == "" && usesMark:
		return "SPECIAL_SIGN_WHOLE_WORD"
	case stripped == expanLower:
		if usesMark {
			return "MARK_ONLY_ABBREVIATION"
		}
		return "NO_VISIBLE_CHANGE"
	case len(stripped) < len(expanLower) && strings.HasPrefix(expanLower, stripped):
		return "SUSPENSION"
	case len(stripped) < len(expanLower):
		return "CONTRACTION"
	default:
		return "OTHER_SUBSTITUTION"
	}
}

// BuildOperationRegistry aggregates a PairUnit list into distinct
// (abbr,expan) surface pairs with counts, classification, and the
// sec.23 functional information-dependence class.
func BuildOperationRegistry(pairs []PairUnit) []OperationRow {
	type key struct{ a, e string }
	counts := map[key]int{}
	marks := map[key]bool{}
	combining := map[key]bool{}
	expansOf := map[string]map[string]bool{}
	var order []key
	for _, p := range pairs {
		k := key{p.AbbrText, p.ExpanText}
		if counts[k] == 0 {
			order = append(order, k)
		}
		counts[k]++
		if p.HasMark {
			marks[k] = true
		}
		if p.MarkIsCombining {
			combining[k] = true
		}
		if expansOf[p.AbbrText] == nil {
			expansOf[p.AbbrText] = map[string]bool{}
		}
		expansOf[p.AbbrText][p.ExpanText] = true
	}
	rows := make([]OperationRow, 0, len(order))
	for _, k := range order {
		ambiguous := len(expansOf[k.a]) > 1
		class := classifyPair(k.a, k.e, marks[k])
		info := "SELF_SUFFICIENT"
		switch {
		case ambiguous:
			info = "AMBIGUOUS_WITHOUT_CONTEXT"
		case class == "SPECIAL_SIGN_WHOLE_WORD" || class == "MARK_ONLY_ABBREVIATION":
			info = "CONVENTION_DEPENDENT"
		case class == "SUSPENSION" || class == "CONTRACTION":
			info = "CONVENTION_DEPENDENT"
		}
		rows = append(rows, OperationRow{
			AbbrText: k.a, ExpanText: k.e, Count: counts[k], Class: class,
			UsesMark: marks[k], MarkCombining: combining[k],
			AmbiguousExpan: ambiguous, InfoClass: info,
		})
	}
	return rows
}
