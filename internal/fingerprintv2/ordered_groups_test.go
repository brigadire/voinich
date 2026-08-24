package fingerprintv2

import (
	"math"
	"testing"
)

func TestOrderedGroupMetricsEqualFrozenTask79Estimators(t *testing.T) {
	groups := [][]string{{"a", "bb", "a", "ccc"}, {"dd", "dd"}, {"eee"}}
	var c corpus
	for line, group := range groups {
		for i, tok := range group {
			glyph := make([]string, 0, len([]rune(tok)))
			for _, r := range []rune(tok) {
				glyph = append(glyph, string(r))
			}
			c.records = append(c.records, tokenRecord{Token: tok, Glyph: glyph, Line: line, LineID: string(rune('A' + line)), IndexInLine: i, LineLength: len(group)})
		}
	}
	got := OrderedGroupMetrics(groups)
	for _, id := range []string{"2DL1_LAYOUT_POSITION_MI", "BP1_BOUNDARY_TOKEN_NMI", "LS2_POSITIONAL_LEXICON_NMI"} {
		x, y, _ := metricVectors(c, id)
		if want := normalizedMI(x, y); math.Abs(got[id]-want) > 1e-15 {
			t.Fatalf("%s generic=%g frozen=%g", id, got[id], want)
		}
	}
	if want := boundaryLengthAsymmetry(c); math.Abs(got["LS3_BOUNDARY_LENGTH_ASYMMETRY"]-want) > 1e-15 {
		t.Fatalf("LS3 generic=%g frozen=%g", got["LS3_BOUNDARY_LENGTH_ASYMMETRY"], want)
	}
	if want := adjacentRepeat(c); math.Abs(got["LS4_WITHIN_LINE_EXACT_REPETITION"]-want) > 1e-15 {
		t.Fatalf("LS4 generic=%g frozen=%g", got["LS4_WITHIN_LINE_EXACT_REPETITION"], want)
	}
	lengths := []float64{4, 2, 1}
	wantCV := sd(lengths, mean(lengths)) / mean(lengths)
	if math.Abs(got["LS1_LINE_LENGTH_CV"]-wantCV) > 1e-15 {
		t.Fatalf("LS1 generic=%g frozen=%g", got["LS1_LINE_LENGTH_CV"], wantCV)
	}
}
