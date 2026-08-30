package notation

import (
	"strconv"
	"testing"
)

// syntheticLinedCorpus builds a deterministic multi-line synthetic corpus of
// exactly n tokens laid out as physical lines of varying length, used to
// exercise the rarefaction sampler without touching any real corpus.
func syntheticLinedCorpus(nLines, tokensPerLine int) []Record {
	var out []Record
	alphabet := []string{"a", "b", "c", "d", "e"}
	id := 0
	for li := 0; li < nLines; li++ {
		for ti := 0; ti < tokensPerLine; ti++ {
			sym := alphabet[(li*7+ti*3)%len(alphabet)]
			sym2 := alphabet[(li*3+ti*5+1)%len(alphabet)]
			out = append(out, Record{
				SchemaVersion:  SchemaVersion,
				CorpusID:       "SYN-RAREFY",
				Representation: "SYN-R1",
				Document:       ObservedLevel{Value: "doc1", Observed: true},
				PhysicalLine:   ObservedLevel{Value: strconv.Itoa(li), Observed: true},
				TokenID:        "SYN-" + strconv.Itoa(id),
				TokenIndex:     ti,
				Token:          sym + sym2,
				Symbols:        []string{sym, sym2},
			})
			id++
		}
	}
	return out
}

func TestR1Determinism(t *testing.T) {
	rs := syntheticLinedCorpus(200, 10)
	a, err := Rarefy(rs, 500, SeedFor(BaseSeed, "SYN-RAREFY", "SYN-R1", FamilyGroupStructural, 500, 3))
	if err != nil {
		t.Fatal(err)
	}
	b, err := Rarefy(rs, 500, SeedFor(BaseSeed, "SYN-RAREFY", "SYN-R1", FamilyGroupStructural, 500, 3))
	if err != nil {
		t.Fatal(err)
	}
	if a.ActualN != b.ActualN || len(a.Records) != len(b.Records) {
		t.Fatalf("non-deterministic draw: %d vs %d", a.ActualN, b.ActualN)
	}
	for i := range a.Records {
		if a.Records[i].TokenID != b.Records[i].TokenID {
			t.Fatalf("record order differs at %d", i)
		}
	}
}

func TestR2BoundaryPreservation(t *testing.T) {
	rs := syntheticLinedCorpus(200, 10)
	draw, err := Rarefy(rs, 537, SeedFor(BaseSeed, "SYN-RAREFY", "SYN-R1", FamilyGroupStructural, 537, 1))
	if err != nil {
		t.Fatal(err)
	}
	counts := map[string]int{}
	for _, r := range draw.Records {
		counts[lineKey(r)]++
	}
	full := map[string]int{}
	for _, r := range rs {
		full[lineKey(r)]++
	}
	for k, n := range counts {
		if n != full[k] {
			t.Fatalf("line %s partially sampled: got %d want %d", k, n, full[k])
		}
	}
}

func TestR3NoSyntheticTransitions(t *testing.T) {
	rs := syntheticLinedCorpus(200, 10)
	draw, err := Rarefy(rs, 733, SeedFor(BaseSeed, "SYN-RAREFY", "SYN-R1", FamilyGroupStructural, 733, 7))
	if err != nil {
		t.Fatal(err)
	}
	for i := 1; i < len(draw.Records); i++ {
		a, b := draw.Records[i-1], draw.Records[i]
		if lineKey(a) != lineKey(b) {
			continue
		}
		if b.TokenIndex != a.TokenIndex+1 {
			t.Fatalf("synthetic mid-line transition at %d: %+v -> %+v", i, a, b)
		}
	}
}

func TestR4SizeAccounting(t *testing.T) {
	rs := syntheticLinedCorpus(200, 10)
	draw, err := Rarefy(rs, 1234, SeedFor(BaseSeed, "SYN-RAREFY", "SYN-R1", FamilyGroupStructural, 1234, 2))
	if err != nil {
		t.Fatal(err)
	}
	if draw.ActualN != len(draw.Records) {
		t.Fatalf("actual_N=%d but len(records)=%d", draw.ActualN, len(draw.Records))
	}
	// The frozen overshoot policy picks whichever of {last-block-excluded,
	// last-block-included} is closer to requestedN, so actualN may undershoot
	// by up to one block's size (10 tokens here) rather than always
	// overshoot.
	dev := draw.ActualN - draw.RequestedN
	if dev < -10 || dev > 10 {
		t.Fatalf("actual_N=%d too far from requested %d", draw.ActualN, draw.RequestedN)
	}
}

func TestR5OrderPreservation(t *testing.T) {
	rs := syntheticLinedCorpus(200, 10)
	draw, err := Rarefy(rs, 900, SeedFor(BaseSeed, "SYN-RAREFY", "SYN-R1", FamilyGroupStructural, 900, 5))
	if err != nil {
		t.Fatal(err)
	}
	byLine := map[string][]int{}
	for _, r := range draw.Records {
		byLine[lineKey(r)] = append(byLine[lineKey(r)], r.TokenIndex)
	}
	for k, idxs := range byLine {
		for i := 1; i < len(idxs); i++ {
			if idxs[i] <= idxs[i-1] {
				t.Fatalf("line %s token order not preserved: %v", k, idxs)
			}
		}
	}
}

func TestR6VMCandidateSymmetry(t *testing.T) {
	a := syntheticLinedCorpus(200, 10)
	b := syntheticLinedCorpus(150, 8)
	for _, rs := range [][]Record{a, b} {
		_, summary, err := RunRarefaction(rs, rs[0].CorpusID, rs[0].Representation, []int{500}, 5, BaseSeed)
		if err != nil {
			t.Fatal(err)
		}
		if len(summary) == 0 {
			t.Fatal("expected summary rows from identical algorithm")
		}
	}
}

func TestRarefyRejectsUndersizedCorpus(t *testing.T) {
	rs := syntheticLinedCorpus(5, 10)
	if _, err := Rarefy(rs, 1000, 1); err == nil {
		t.Fatal("expected error for checkpoint above corpus size")
	}
}

func TestPercentileMonotone(t *testing.T) {
	v := []float64{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	if percentile(v, 0) != 1 || percentile(v, 1) != 10 {
		t.Fatalf("percentile endpoints wrong: %v %v", percentile(v, 0), percentile(v, 1))
	}
	if percentile(v, 0.5) < 5 || percentile(v, 0.5) > 6 {
		t.Fatalf("median out of range: %v", percentile(v, 0.5))
	}
}
