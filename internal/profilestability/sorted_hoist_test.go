package profilestability

import (
	"fmt"
	"math"
	"math/rand"
	"reflect"
	"sort"
	"testing"
)

// referencePositionJSD and referenceCosine are the pre-hoist implementations
// of positionJSD/cosine, preserved verbatim (task27 item 10): each call
// re-sorted its map's keys from scratch instead of reusing a SortedProfile's
// cached sort. Compare/CompareSorted's algorithm must produce byte-identical
// output to these across every fixture below.

func referencePositionJSD(left, right map[int]int) float64 {
	leftTotal, rightTotal := sumCounts(left), sumCounts(right)
	if leftTotal == 0 || rightTotal == 0 {
		return 1
	}
	positions := make(map[int]bool, len(left)+len(right))
	for position := range left {
		positions[position] = true
	}
	for position := range right {
		positions[position] = true
	}
	ordered := make([]int, 0, len(positions))
	for position := range positions {
		ordered = append(ordered, position)
	}
	sort.Ints(ordered)
	value := 0.0
	for _, position := range ordered {
		p := float64(left[position]) / float64(leftTotal)
		q := float64(right[position]) / float64(rightTotal)
		middle := (p + q) / 2
		if p > 0 {
			value += .5 * p * math.Log2(p/middle)
		}
		if q > 0 {
			value += .5 * q * math.Log2(q/middle)
		}
	}
	return clamp(value)
}

func referenceCosine(left, right map[string]int) float64 {
	if len(left) == 0 || len(right) == 0 {
		return 0
	}
	keys := make([]string, 0, len(left))
	for key := range left {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	dot, leftNorm := 0.0, 0.0
	for _, key := range keys {
		count := left[key]
		dot += float64(count * right[key])
		leftNorm += float64(count * count)
	}
	keys = keys[:0]
	for key := range right {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	rightNorm := 0.0
	for _, key := range keys {
		rightNorm += float64(right[key] * right[key])
	}
	if leftNorm == 0 || rightNorm == 0 {
		return 0
	}
	return clamp(dot / (math.Sqrt(leftNorm) * math.Sqrt(rightNorm)))
}

func referenceCompare(left, right Profile) Components {
	result := Components{
		PositionSimilarity: 1 - referencePositionJSD(left.Positions, right.Positions),
		LeftSimilarity:     referenceCosine(left.Left, right.Left),
		RightSimilarity:    referenceCosine(left.Right, right.Right),
	}
	result.Similarity = (result.PositionSimilarity + result.LeftSimilarity + result.RightSimilarity) / 3
	return result
}

func fixtureProfile(rng *rand.Rand, maxKeys int) Profile {
	p := Profile{Positions: map[int]int{}, Left: map[string]int{}, Right: map[string]int{}}
	nPos := rng.Intn(maxKeys + 1)
	for i := 0; i < nPos; i++ {
		p.Positions[rng.Intn(maxKeys*3)] += 1 + rng.Intn(5)
	}
	nLeft := rng.Intn(maxKeys + 1)
	for i := 0; i < nLeft; i++ {
		p.Left[fmt.Sprintf("tok%02d", rng.Intn(maxKeys*2))] += 1 + rng.Intn(5)
	}
	nRight := rng.Intn(maxKeys + 1)
	for i := 0; i < nRight; i++ {
		p.Right[fmt.Sprintf("tok%02d", rng.Intn(maxKeys*2))] += 1 + rng.Intn(5)
	}
	for _, v := range p.Positions {
		p.Count += v
	}
	return p
}

func TestCompareSortedMatchesReference(t *testing.T) {
	rng := rand.New(rand.NewSource(42))
	sizes := []int{0, 1, 2, 5, 20, 60}
	for _, sizeA := range sizes {
		for _, sizeB := range sizes {
			for trial := 0; trial < 5; trial++ {
				a := fixtureProfile(rng, sizeA)
				b := fixtureProfile(rng, sizeB)
				want := referenceCompare(a, b)
				got := Compare(a, b)
				if !reflect.DeepEqual(want, got) {
					t.Fatalf("sizeA=%d sizeB=%d trial=%d: want %+v got %+v", sizeA, sizeB, trial, want, got)
				}
				gotSorted := CompareSorted(Precompute(a), Precompute(b))
				if !reflect.DeepEqual(want, gotSorted) {
					t.Fatalf("CompareSorted sizeA=%d sizeB=%d trial=%d: want %+v got %+v", sizeA, sizeB, trial, want, gotSorted)
				}
			}
		}
	}
}

func TestPrecomputeAllMatchesReference(t *testing.T) {
	rng := rand.New(rand.NewSource(7))
	profiles := map[string]Profile{}
	tokens := make([]string, 0, 30)
	for i := 0; i < 30; i++ {
		tok := fmt.Sprintf("t%02d", i)
		tokens = append(tokens, tok)
		profiles[tok] = fixtureProfile(rng, 15)
	}
	ws := PrecomputeAll(profiles)
	for i, a := range tokens {
		for _, b := range tokens[i+1:] {
			want := referenceCompare(profiles[a], profiles[b])
			got := CompareSorted(ws[a], ws[b])
			if !reflect.DeepEqual(want, got) {
				t.Fatalf("%s vs %s: want %+v got %+v", a, b, want, got)
			}
		}
	}
}

// TestCompareSortedMissingWorkspaceEntry mirrors what CompareTokens sees when
// one side's token isn't present in that workspace's source profiles map
// (e.g. a token absent from a bootstrap-resampled corpus): looking it up in
// the ws.profiles map returns the zero-value SortedProfile, matching the
// original map[string]Profile[token] zero-value fallback exactly.
func TestCompareSortedMissingWorkspaceEntry(t *testing.T) {
	ws := PrecomputeAll(map[string]Profile{"a": {Positions: map[int]int{0: 1}, Left: map[string]int{"x": 1}, Right: map[string]int{"y": 1}}})
	want := referenceCompare(Profile{}, Profile{Positions: map[int]int{0: 1}, Left: map[string]int{"x": 1}, Right: map[string]int{"y": 1}})
	got := CompareSorted(ws["missing"], ws["a"])
	if !reflect.DeepEqual(want, got) {
		t.Fatalf("want %+v got %+v", want, got)
	}
}
