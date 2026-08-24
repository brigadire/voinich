package task82b

import "fmt"

// Operator is one frozen deterministic extraction rule (task82b.txt
// sec.27-30, 40). The grid is small and fixed before any carrier is
// touched; no parameter here was chosen by looking at Doyle/Longfellow/
// Astafiev output (sec.30).
type Operator struct {
	ID              string // unique, e.g. "NTH_GLYPH_OF_TOKEN_2"
	StructuralClass string // base class name from sec.27, before parameterization
	ExtractionClass string // ACROSTIC / TELESTIC / PERIODIC_EXTRACTION / POSITIONAL_EXTRACTION / GENERIC_SELECTIVE_EXTRACTION (sec.40)
	Provenance      string // HISTORICALLY_ATTESTED / HISTORICALLY_PLAUSIBLE / FORMAL_CONTROL (sec.28)
	Param           int    // n / k / offset; 0 if not parameterized
	NullClass       string // PER_GROUP (stratified null applies) or PERIODIC (phase null applies)
	WindowSize      int    // only for FIXED_OFFSET_WITHIN_GROUP
}

// Registry is the frozen operator grid (task82b.txt sec.27-30). Grid
// sizing rationale (sec.29, "no combinatorial search"): every
// NTH_*/FIXED_OFFSET instance that would duplicate a FIRST_* class at
// offset 0 is omitted rather than included and later dropped, so the
// registry itself, not a post-hoc filter, is the small frozen set.
func Registry() []Operator {
	var ops []Operator
	ops = append(ops,
		Operator{ID: "FIRST_GLYPH_OF_TOKEN", StructuralClass: "FIRST_GLYPH_OF_TOKEN", ExtractionClass: "ACROSTIC", Provenance: "HISTORICALLY_ATTESTED", NullClass: "PER_GROUP"},
		Operator{ID: "LAST_GLYPH_OF_TOKEN", StructuralClass: "LAST_GLYPH_OF_TOKEN", ExtractionClass: "TELESTIC", Provenance: "HISTORICALLY_ATTESTED", NullClass: "PER_GROUP"},
		Operator{ID: "FIRST_TOKEN_OF_LINE", StructuralClass: "FIRST_TOKEN_OF_LINE", ExtractionClass: "ACROSTIC", Provenance: "HISTORICALLY_ATTESTED", NullClass: "PER_GROUP"},
		Operator{ID: "LAST_TOKEN_OF_LINE", StructuralClass: "LAST_TOKEN_OF_LINE", ExtractionClass: "TELESTIC", Provenance: "HISTORICALLY_ATTESTED", NullClass: "PER_GROUP"},
		Operator{ID: "FIRST_GLYPH_OF_LINE", StructuralClass: "FIRST_GLYPH_OF_LINE", ExtractionClass: "ACROSTIC", Provenance: "HISTORICALLY_ATTESTED", NullClass: "PER_GROUP"},
		Operator{ID: "LAST_GLYPH_OF_LINE", StructuralClass: "LAST_GLYPH_OF_LINE", ExtractionClass: "TELESTIC", Provenance: "HISTORICALLY_ATTESTED", NullClass: "PER_GROUP"},
	)
	for _, n := range []int{2, 3} {
		ops = append(ops, Operator{ID: fmt.Sprintf("NTH_GLYPH_OF_TOKEN_%d", n), StructuralClass: "NTH_GLYPH_OF_TOKEN", ExtractionClass: "POSITIONAL_EXTRACTION", Provenance: "HISTORICALLY_PLAUSIBLE", Param: n, NullClass: "PER_GROUP"})
	}
	for _, n := range []int{2, 3} {
		ops = append(ops, Operator{ID: fmt.Sprintf("NTH_TOKEN_OF_LINE_%d", n), StructuralClass: "NTH_TOKEN_OF_LINE", ExtractionClass: "POSITIONAL_EXTRACTION", Provenance: "HISTORICALLY_PLAUSIBLE", Param: n, NullClass: "PER_GROUP"})
	}
	for _, k := range []int{2, 3, 5, 7} {
		ops = append(ops, Operator{ID: fmt.Sprintf("PERIODIC_TOKEN_%d", k), StructuralClass: "PERIODIC_TOKEN", ExtractionClass: "PERIODIC_EXTRACTION", Provenance: "FORMAL_CONTROL", Param: k, NullClass: "PERIODIC"})
	}
	for _, k := range []int{2, 3, 5, 7} {
		ops = append(ops, Operator{ID: fmt.Sprintf("PERIODIC_GLYPH_%d", k), StructuralClass: "PERIODIC_GLYPH", ExtractionClass: "PERIODIC_EXTRACTION", Provenance: "FORMAL_CONTROL", Param: k, NullClass: "PERIODIC"})
	}
	for _, off := range []int{1, 2} {
		ops = append(ops, Operator{ID: fmt.Sprintf("FIXED_OFFSET_WITHIN_GROUP_%d", off), StructuralClass: "FIXED_OFFSET_WITHIN_GROUP", ExtractionClass: "POSITIONAL_EXTRACTION", Provenance: "FORMAL_CONTROL", Param: off, NullClass: "PER_GROUP", WindowSize: 3})
	}
	return ops
}

// Selection is the result of applying one Operator to one carrier's atom
// streams: which atoms were chosen, in corpus order, plus enough
// bookkeeping for every null model in nulls.go to be derived generically.
type Selection struct {
	Kind            string // "TOKEN" or "GLYPH"
	Chosen          []int  // indices into tokenAtoms or glyphAtoms, ascending, corpus order
	GroupOf         []int  // len(Chosen); group id per choice (PER_GROUP operators only)
	GroupCandidates map[int][]int
	NullClass       string
	Period          int
	Phase           int
	CandidatePool   int // total atoms of Kind available in the carrier
	SkippedGroups   int // groups considered but too short to yield a candidate
}

// Apply runs one operator over a carrier's atom streams.
func Apply(op Operator, tokenAtoms []TokenAtom, glyphAtoms []GlyphAtom, numLines int) Selection {
	switch op.StructuralClass {
	case "FIRST_GLYPH_OF_TOKEN", "LAST_GLYPH_OF_TOKEN", "NTH_GLYPH_OF_TOKEN":
		return perTokenGlyph(op, glyphAtoms)
	case "FIRST_TOKEN_OF_LINE", "LAST_TOKEN_OF_LINE", "NTH_TOKEN_OF_LINE":
		return perLineToken(op, tokenAtoms, numLines)
	case "FIRST_GLYPH_OF_LINE", "LAST_GLYPH_OF_LINE":
		return perLineGlyph(op, glyphAtoms, numLines)
	case "PERIODIC_TOKEN":
		return periodic(op, len(tokenAtoms), "TOKEN")
	case "PERIODIC_GLYPH":
		return periodic(op, len(glyphAtoms), "GLYPH")
	case "FIXED_OFFSET_WITHIN_GROUP":
		return windowedToken(op, tokenAtoms, numLines)
	default:
		panic("task82b: unknown operator structural class " + op.StructuralClass)
	}
}

// perTokenGlyph implements FIRST/LAST/NTH_GLYPH_OF_TOKEN: group = one
// token's own glyph run.
func perTokenGlyph(op Operator, glyphAtoms []GlyphAtom) Selection {
	groups := map[int][]int{} // group id = token's ordinal among all tokens that have >=1 glyph, in stream order
	var order []int
	cur := -1
	curKey := -1
	tokenOrdinal := -1
	for i, g := range glyphAtoms {
		key := g.Line*1_000_000 + g.TokenIdxInLine
		if key != curKey {
			curKey = key
			tokenOrdinal++
			cur = tokenOrdinal
			order = append(order, cur)
		}
		groups[cur] = append(groups[cur], i)
	}
	sel := Selection{Kind: "GLYPH", GroupCandidates: groups, NullClass: op.NullClass, CandidatePool: len(glyphAtoms)}
	for _, gid := range order {
		cand := groups[gid]
		var pick int
		switch op.StructuralClass {
		case "FIRST_GLYPH_OF_TOKEN":
			pick = 0
		case "LAST_GLYPH_OF_TOKEN":
			pick = len(cand) - 1
		case "NTH_GLYPH_OF_TOKEN":
			if len(cand) < op.Param {
				sel.SkippedGroups++
				continue
			}
			pick = op.Param - 1
		}
		sel.Chosen = append(sel.Chosen, cand[pick])
		sel.GroupOf = append(sel.GroupOf, gid)
	}
	return sel
}

// perLineToken implements FIRST/LAST/NTH_TOKEN_OF_LINE: group = one line.
func perLineToken(op Operator, tokenAtoms []TokenAtom, numLines int) Selection {
	groups := map[int][]int{}
	for i, t := range tokenAtoms {
		groups[t.Line] = append(groups[t.Line], i)
	}
	sel := Selection{Kind: "TOKEN", GroupCandidates: groups, NullClass: op.NullClass, CandidatePool: len(tokenAtoms)}
	for line := range numLines {
		cand, ok := groups[line]
		if !ok {
			continue
		}
		var pick int
		switch op.StructuralClass {
		case "FIRST_TOKEN_OF_LINE":
			pick = 0
		case "LAST_TOKEN_OF_LINE":
			pick = len(cand) - 1
		case "NTH_TOKEN_OF_LINE":
			if len(cand) < op.Param {
				sel.SkippedGroups++
				continue
			}
			pick = op.Param - 1
		}
		sel.Chosen = append(sel.Chosen, cand[pick])
		sel.GroupOf = append(sel.GroupOf, line)
	}
	return sel
}

// perLineGlyph implements FIRST/LAST_GLYPH_OF_LINE: group = one line;
// candidate = the first (resp. last) token in that line's own glyph run.
func perLineGlyph(op Operator, glyphAtoms []GlyphAtom, numLines int) Selection {
	firstTok := map[int]int{} // line -> first token idx in that line
	lastTok := map[int]int{}
	seen := map[int]bool{}
	for _, g := range glyphAtoms {
		if !seen[g.Line] {
			firstTok[g.Line] = g.TokenIdxInLine
			seen[g.Line] = true
		}
		lastTok[g.Line] = g.TokenIdxInLine
	}
	groups := map[int][]int{}
	for i, g := range glyphAtoms {
		groups[g.Line] = append(groups[g.Line], i)
	}
	sel := Selection{Kind: "GLYPH", GroupCandidates: groups, NullClass: op.NullClass, CandidatePool: len(glyphAtoms)}
	for line := range numLines {
		cand, ok := groups[line]
		if !ok {
			continue
		}
		var wantTok int
		if op.StructuralClass == "FIRST_GLYPH_OF_LINE" {
			wantTok = firstTok[line]
		} else {
			wantTok = lastTok[line]
		}
		var pick = -1
		if op.StructuralClass == "FIRST_GLYPH_OF_LINE" {
			for _, idx := range cand {
				if glyphAtoms[idx].TokenIdxInLine == wantTok && glyphAtoms[idx].GlyphIdxInToken == 0 {
					pick = idx
					break
				}
			}
		} else {
			best := -1
			for _, idx := range cand {
				if glyphAtoms[idx].TokenIdxInLine == wantTok {
					if best == -1 || glyphAtoms[idx].GlyphIdxInToken > glyphAtoms[best].GlyphIdxInToken {
						best = idx
					}
				}
			}
			pick = best
		}
		if pick == -1 {
			sel.SkippedGroups++
			continue
		}
		sel.Chosen = append(sel.Chosen, pick)
		sel.GroupOf = append(sel.GroupOf, line)
	}
	return sel
}

// periodic implements PERIODIC_TOKEN_k / PERIODIC_GLYPH_k: every k-th
// element of the global flattened stream, frozen phase 0 (task82b.txt
// sec.30/36). Other phases are computed on demand by the periodic-phase
// null (nulls.go), not here.
func periodic(op Operator, n int, kind string) Selection {
	sel := Selection{Kind: kind, NullClass: op.NullClass, Period: op.Param, Phase: 0, CandidatePool: n}
	for i := range n {
		if i%op.Param == 0 {
			sel.Chosen = append(sel.Chosen, i)
		}
	}
	return sel
}

// windowedToken implements FIXED_OFFSET_WITHIN_GROUP: group = a
// non-overlapping WindowSize-token window within one line (reset per
// line; an incomplete trailing window is dropped, task82b.txt sec.30
// design note in TASK82B_DESIGN.md).
func windowedToken(op Operator, tokenAtoms []TokenAtom, numLines int) Selection {
	byLine := map[int][]int{}
	for i, t := range tokenAtoms {
		byLine[t.Line] = append(byLine[t.Line], i)
	}
	groups := map[int][]int{}
	sel := Selection{Kind: "TOKEN", NullClass: op.NullClass, CandidatePool: len(tokenAtoms)}
	gid := 0
	for line := range numLines {
		idxs := byLine[line]
		for start := 0; start+op.WindowSize <= len(idxs); start += op.WindowSize {
			window := idxs[start : start+op.WindowSize]
			groups[gid] = append([]int{}, window...)
			sel.Chosen = append(sel.Chosen, window[op.Param])
			sel.GroupOf = append(sel.GroupOf, gid)
			gid++
		}
	}
	sel.GroupCandidates = groups
	return sel
}

// Render turns a Selection back into per-line output token groups (one
// output token per chosen atom), preserving each atom's own original
// line so corpus-level line structure survives extraction (task82b.txt
// sec.37).
func Render(sel Selection, tokenAtoms []TokenAtom, glyphAtoms []GlyphAtom, numLines int) [][]string {
	out := make([][]string, numLines)
	for _, idx := range sel.Chosen {
		var line int
		var text string
		if sel.Kind == "TOKEN" {
			line, text = tokenAtoms[idx].Line, tokenAtoms[idx].Text
		} else {
			line, text = glyphAtoms[idx].Line, glyphAtoms[idx].Ch
		}
		out[line] = append(out[line], text)
	}
	return out
}
