package tokenrelationvalidation

import "zcore.dev/voinich/internal/profilestability"

// profileWorkspace caches, per block ID, the shape of buildLocalProfiles'
// P and D structures - a profilestability.Profile plus a
// [2][maxD]map[string]int - for every distinct token in that block, across
// repeated calls. A block's distinct-token SET never changes under
// PermuteWithinBlocks (which only ever reorders which position holds which
// token text, never adds, removes, or moves a token across blocks): only
// the position/neighbor *counts* inside those maps need refilling each
// call. Reusing (via clear(), not reallocating) that skeleton instead of
// rebuilding it from scratch on every one of the ~11000
// profilePermutationScores replicate calls removed the vast majority of
// this analyzer's allocation pressure - profiling a full production run
// showed buildLocalProfiles' per-token map construction (the getD closure)
// alone was 93% of all allocated objects.
type profileWorkspace struct {
	cache map[string]localProfiles
}

func newProfileWorkspace() *profileWorkspace {
	return &profileWorkspace{cache: map[string]localProfiles{}}
}

// buildLocalProfiles mirrors the package-level buildLocalProfiles exactly,
// but reuses ws's cached per-block skeleton (cleared, not reallocated)
// across repeated calls for the same block ID instead of allocating one
// fresh.
func (ws *profileWorkspace) buildLocalProfiles(b Block, maxD int) localProfiles {
	x, ok := ws.cache[b.ID]
	if !ok {
		x = localProfiles{P: map[string]profilestability.Profile{}, D: map[string][][]map[string]int{}}
		for _, t := range b.Tokens {
			if x.D[t.Text] != nil {
				continue
			}
			z := make([][]map[string]int, 2)
			for side := range z {
				z[side] = make([]map[string]int, maxD)
				for d := range z[side] {
					z[side][d] = map[string]int{}
				}
			}
			x.D[t.Text] = z
			x.P[t.Text] = profilestability.Profile{Positions: map[int]int{}, Left: map[string]int{}, Right: map[string]int{}}
		}
		ws.cache[b.ID] = x
	} else {
		for tok, p := range x.P {
			clear(p.Positions)
			clear(p.Left)
			clear(p.Right)
			p.Count = 0
			x.P[tok] = p
		}
		for _, dmaps := range x.D {
			for side := range dmaps {
				for d := range dmaps[side] {
					clear(dmaps[side][d])
				}
			}
		}
	}
	for _, line := range splitLines(b) {
		for i, t := range line {
			p := x.P[t.Text]
			p.Count++
			p.Positions[t.LineIndex]++
			if i > 0 {
				p.Left[line[i-1].Text]++
			}
			if i+1 < len(line) {
				p.Right[line[i+1].Text]++
			}
			x.P[t.Text] = p
			dmap := x.D[t.Text]
			for d := 1; d <= maxD; d++ {
				if i-d >= 0 {
					dmap[0][d-1][line[i-d].Text]++
				}
				if i+d < len(line) {
					dmap[1][d-1][line[i+d].Text]++
				}
			}
		}
	}
	return x
}
