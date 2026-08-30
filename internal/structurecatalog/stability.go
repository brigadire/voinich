package structurecatalog

import (
	"path/filepath"
	"sort"
	"strings"
)

func (cat *Catalog) applyStability() {
	for i := range cat.Rules {
		cat.Rules[i].RuleID = cat.Rules[i].Level + "-" + leftPad(i+1, 7)
	}
	if len(cat.Replication.Occurrences) == 0 {
		return
	}
	r := cat.Replication
	glyphPos := map[string]int{}
	big := map[string]int{}
	trans := map[string]int{}
	position := map[string]int{}
	lineSets := map[string]map[int]bool{}
	for _, o := range r.Occurrences {
		rr := []rune(o.Token)
		if len(rr) > 0 {
			glyphPos["INITIAL\x00"+string(rr[0])]++
			glyphPos["FINAL\x00"+string(rr[len(rr)-1])]++
			for i, g := range rr {
				if i > 0 && i < len(rr)-1 {
					glyphPos["INTERNAL\x00"+string(g)]++
				}
				if i+1 < len(rr) {
					big[string(rr[i:i+2])]++
				}
			}
		}
		if o.Index == 0 {
			position[o.Token+"\x00FIRST"]++
		}
		if o.Index == len(r.Lines[o.Line])-1 {
			position[o.Token+"\x00LAST"]++
		}
		if lineSets[o.Token] == nil {
			lineSets[o.Token] = map[int]bool{}
		}
		lineSets[o.Token][o.Line] = true
	}
	for _, line := range r.Lines {
		for i := 0; i+1 < len(line); i++ {
			trans[line[i]+"\x00"+line[i+1]]++
		}
	}
	rows := [][]string{}
	for i := range cat.Rules {
		x := &cat.Rules[i]
		n, comparable := 0, true
		switch x.RuleType {
		case "GLYPH_POSITION":
			n = glyphPos[x.RHS+"\x00"+x.LHS]
		case "GLYPH_BIGRAM":
			n = big[x.LHS+x.RHS]
		case "TOKEN_TRANSITION":
			if r.Counts[x.LHS] == 0 || r.Counts[x.RHS] == 0 {
				comparable = false
			} else {
				n = trans[x.LHS+"\x00"+x.RHS]
			}
		case "TOKEN_POSITION":
			if r.Counts[x.LHS] == 0 {
				comparable = false
			} else {
				n = position[x.LHS+"\x00"+x.RHS]
			}
		case "TOKEN_LINE_COOCCURRENCE":
			if r.Counts[x.LHS] == 0 || r.Counts[x.RHS] == 0 {
				comparable = false
			} else {
				n = intersection(lineSets[x.LHS], lineSets[x.RHS])
			}
		default:
			comparable = false
		}
		if !comparable {
			x.Stability = "NOT_COMPARABLE"
			continue
		}
		repStatus := "OBSERVED"
		if n == 0 {
			repStatus = "UNOBSERVED"
		}
		if repStatus == x.ObservedStatus {
			x.Stability = "SAME"
		} else {
			x.Stability = "DIFFERENT"
		}
		rows = append(rows, []string{x.RuleID, x.RuleType, x.LHS, x.RHS, si(x.ObservedCount), x.ObservedStatus, si(n), repStatus, x.Stability, cat.Primary.Transcription, cat.Replication.Transcription})
	}
	sort.Slice(rows, func(i, j int) bool { return rows[i][0] < rows[j][0] })
	_ = writeTSV(filepath.Join(cat.Config.OutputDir, "TRANSCRIPTION_STABILITY.tsv"), []string{"rule_id", "rule_type", "lhs", "rhs", "primary_count", "primary_observed_status", "replication_count", "replication_observed_status", "stability", "primary_transcription", "replication_transcription"}, rows)
}
func leftPad(n, w int) string { s := si(n); return strings.Repeat("0", max(0, w-len(s))) + s }
