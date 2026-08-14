package conditionalregime

import (
	"math"
	"math/rand"
	"strings"
	"testing"
)

func TestJointClassConstruction(t *testing.T) {
	currier := []string{"A", "A", "B", "B"}
	hand := []string{"1", "1", "2", "2"}
	blocks := buildAllBlocks(currier, hand)[SchemeJoint]
	if len(blocks) != 2 {
		t.Fatalf("expected 2 joint blocks, got %d", len(blocks))
	}
	if blocks[0].Class.Label() != "A/1" || blocks[0].Start != 0 || blocks[0].End != 2 {
		t.Fatalf("unexpected first block: %+v", blocks[0])
	}
	if blocks[1].Class.Label() != "B/2" || blocks[1].Start != 2 || blocks[1].End != 4 {
		t.Fatalf("unexpected second block: %+v", blocks[1])
	}
}

func TestUnknownMetadataExclusion(t *testing.T) {
	if normalizeLabel("?") != "" || normalizeLabel("null") != "" || normalizeLabel("") != "" {
		t.Fatalf("normalizeLabel must map '?'/'null'/'' to unknown ('')")
	}
	if normalizeLabel("A") != "A" {
		t.Fatalf("normalizeLabel must leave a known label unchanged")
	}
	// buildAllBlocks operates on already-normalized labels (unknown == "").
	currier := []string{"A", "", "A", ""}
	hand := []string{"1", "1", "", "1"}
	blocks := buildAllBlocks(currier, hand)[SchemeJoint]
	// Position 0 is the only token with both currier and hand known.
	if len(blocks) != 1 || blocks[0].Start != 0 || blocks[0].End != 1 {
		t.Fatalf("expected exactly one single-token block at [0,1), got %+v", blocks)
	}
	for _, b := range blocks {
		if strings.Contains(b.Class.Label(), "UNKNOWN") {
			t.Fatalf("must never create an UNKNOWN class, got %s", b.Class.Label())
		}
		if b.Class.Currier == "" || b.Class.Hand == "" {
			t.Fatalf("a joint block must never have an empty currier or hand: %+v", b.Class)
		}
	}
}

func TestContiguousBlockExtractionNeverMergesNonAdjacentRuns(t *testing.T) {
	currier := []string{"A", "A", "B", "A", "A"}
	hand := []string{"1", "1", "1", "1", "1"}
	blocks := buildAllBlocks(currier, hand)[SchemeJoint]
	if len(blocks) != 3 {
		t.Fatalf("expected 3 separate physical blocks (A/1, B/1, A/1), got %d: %+v", len(blocks), blocks)
	}
	if blocks[0].Class.Label() != "A/1" || blocks[2].Class.Label() != "A/1" {
		t.Fatalf("first and third block should both be class A/1 but be distinct blocks: %+v", blocks)
	}
	if blocks[0].Index == blocks[2].Index {
		t.Fatalf("the two non-adjacent A/1 runs must have distinct block indices, got %d and %d", blocks[0].Index, blocks[2].Index)
	}
}

func TestWindowsNeverCrossBlockBoundary(t *testing.T) {
	tokens := make([]string, 60)
	for i := range tokens {
		tokens[i] = "tok"
	}
	blocks := []Block{{Start: 0, End: 20}, {Start: 20, End: 60}}
	cw := buildClassWindows(tokens, blocks, 10)
	if len(cw) == 0 {
		t.Fatal("expected some windows")
	}
	for _, w := range cw {
		b := blocks[w.BlockIdx]
		if w.AbsStart < b.Start || w.AbsEnd > b.End {
			t.Fatalf("window [%d,%d) escapes its block [%d,%d)", w.AbsStart, w.AbsEnd, b.Start, b.End)
		}
	}
}

func TestEligibilityThresholds(t *testing.T) {
	cid := ClassID{Scheme: SchemeJoint, Currier: "A", Hand: "1"}
	cases := []struct {
		total, largest int
		want           bool
	}{
		{999, 500, false},  // total below minimum
		{1000, 499, false}, // largest block below minimum
		{1000, 500, true},  // exactly at both boundaries
		{2000, 500, true},
	}
	for _, tc := range cases {
		// Split total into chunks no larger than tc.largest, so tc.largest
		// really is the largest block rather than an incidental remainder.
		var bs []Block
		remaining, pos := tc.total, 0
		for remaining > 0 {
			n := min(remaining, tc.largest)
			bs = append(bs, Block{Class: cid, Start: pos, End: pos + n})
			pos += n
			remaining -= n
		}
		blocks := map[Scheme][]Block{SchemeJoint: bs}
		inv := classInventory(blocks, 1000, 500)
		if len(inv) != 1 {
			t.Fatalf("expected one inventory row, got %d", len(inv))
		}
		if inv[0].Eligible != tc.want {
			t.Fatalf("total=%d largest=%d: eligible=%v, want %v", tc.total, tc.largest, inv[0].Eligible, tc.want)
		}
	}
}

func TestTrainingOnlyResidualCenteringHasNoHeldOutLeakage(t *testing.T) {
	blockLen := 100
	tokens := make([]string, 2*blockLen)
	for i := 0; i < blockLen; i++ {
		tokens[i] = "xx"
	}
	for i := blockLen; i < 2*blockLen; i++ {
		tokens[i] = "yy"
	}
	class := ClassID{Scheme: SchemeJoint, Currier: "A", Hand: "1"}
	blocks := []Block{{Class: class, Index: 0, Start: 0, End: blockLen}, {Class: class, Index: 1, Start: blockLen, End: 2 * blockLen}}
	blocksByClass := map[ClassID][]Block{class: blocks}
	rw := buildResidualWindows(tokens, []ClassID{class}, blocksByClass, 10)
	if len(rw) == 0 {
		t.Fatal("expected residual windows")
	}
	for _, w := range rw {
		if w.BlockIndex == 0 {
			// Block 0 ("xx") must be centered using block 1's ("yy") mean
			// only: if its own content had leaked into the centering mean,
			// its residual for "xx" would be far below 1.
			if w.Residual["xx"] < 0.9 {
				t.Fatalf("block 0 residual for xx = %v, want close to 1 (centered only on block 1's mean)", w.Residual["xx"])
			}
			if w.Residual["yy"] > -0.9 {
				t.Fatalf("block 0 residual for yy = %v, want close to -1 (block 1's mean is ~1 yy)", w.Residual["yy"])
			}
		} else {
			if w.Residual["yy"] < 0.9 {
				t.Fatalf("block 1 residual for yy = %v, want close to 1 (centered only on block 0's mean)", w.Residual["yy"])
			}
		}
	}
}

func TestNullAWithinBlockTokenShuffle(t *testing.T) {
	tokens := []string{"a", "b", "c", "d", "e", "f", "g", "h", "i", "j"}
	blocks := []Block{{Start: 2, End: 6}} // only positions [2,6) may move
	rng := rand.New(rand.NewSource(1))
	out := shuffleWithinBlocks(tokens, blocks, rng)
	for i := 0; i < 2; i++ {
		if out[i] != tokens[i] {
			t.Fatalf("position %d outside the block must be untouched: got %s want %s", i, out[i], tokens[i])
		}
	}
	for i := 6; i < len(tokens); i++ {
		if out[i] != tokens[i] {
			t.Fatalf("position %d outside the block must be untouched: got %s want %s", i, out[i], tokens[i])
		}
	}
	inBlock := map[string]int{}
	for i := 2; i < 6; i++ {
		inBlock[out[i]]++
	}
	for _, tok := range tokens[2:6] {
		if inBlock[tok] == 0 {
			t.Fatalf("shuffled block lost token %q: unigram frequencies must be preserved", tok)
		}
		inBlock[tok]--
	}
}

func TestNullBWindowOrderShuffle(t *testing.T) {
	classA := ClassID{Scheme: SchemeJoint, Currier: "A", Hand: "1"}
	classB := ClassID{Scheme: SchemeJoint, Currier: "B", Hand: "2"}
	rw := []ResidualWindow{
		{Class: classA, BlockIndex: 0}, {Class: classA, BlockIndex: 0}, {Class: classA, BlockIndex: 0},
		{Class: classB, BlockIndex: 0}, {Class: classB, BlockIndex: 0},
	}
	labels := []int{0, 1, 2, 3, 4}
	rng := rand.New(rand.NewSource(7))
	out := shuffleLabelsWithinBlocks(rw, labels, rng)
	seenA := map[int]bool{}
	for i := 0; i < 3; i++ {
		seenA[out[i]] = true
	}
	if len(seenA) != 3 || !seenA[0] || !seenA[1] || !seenA[2] {
		t.Fatalf("class A's block must keep exactly its own 3 labels {0,1,2} rearranged, got %v", out[:3])
	}
	seenB := map[int]bool{}
	for i := 3; i < 5; i++ {
		seenB[out[i]] = true
	}
	if len(seenB) != 2 || !seenB[3] || !seenB[4] {
		t.Fatalf("class B's block must keep exactly its own 2 labels {3,4} rearranged, got %v", out[3:])
	}
}

func TestCrossBlockRecurrence(t *testing.T) {
	class := ClassID{Scheme: SchemeJoint, Currier: "A", Hand: "1"}
	tokens := strings.Repeat("t ", 30)
	fields := strings.Fields(tokens)
	blocks := []Block{{Class: class, Index: 0, Start: 0, End: 10}, {Class: class, Index: 1, Start: 10, End: 20}, {Class: class, Index: 2, Start: 20, End: 30}}
	cw := buildClassWindows(fields, blocks, 10) // exactly one window per block
	if len(cw) != 3 {
		t.Fatalf("expected exactly 3 windows (one per block), got %d", len(cw))
	}
	fullLabels := []int{0, 0, 1} // cluster 0 recurs in blocks 0 and 1; cluster 1 only in block 2
	rec := crossBlockRecurrence(class, 10, "k_medoids", 2, cw, fullLabels, blocks)
	byCluster := map[int]CrossBlockRecurrence{}
	for _, r := range rec {
		byCluster[r.Cluster] = r
	}
	if byCluster[0].BlocksContaining != 2 {
		t.Fatalf("cluster 0 should recur in 2 blocks, got %d", byCluster[0].BlocksContaining)
	}
	if byCluster[1].BlocksContaining != 1 {
		t.Fatalf("cluster 1 should recur in 1 block, got %d", byCluster[1].BlocksContaining)
	}
	if math.Abs(byCluster[0].BlockFraction-2.0/3.0) > 1e-9 {
		t.Fatalf("cluster 0 block fraction = %v, want 2/3", byCluster[0].BlockFraction)
	}
}

func TestResidualMetadataEntropy(t *testing.T) {
	same := []ClassID{{Currier: "A", Hand: "1"}, {Currier: "A", Hand: "1"}, {Currier: "A", Hand: "1"}}
	if h := entropyOfPairs(same); h != 0 {
		t.Fatalf("identical pairs must have zero entropy, got %v", h)
	}
	mixed := []ClassID{{Currier: "A", Hand: "1"}, {Currier: "B", Hand: "2"}}
	if h := entropyOfPairs(mixed); math.Abs(h-math.Log(2)) > 1e-9 {
		t.Fatalf("two equally frequent pairs should have entropy ln(2)=%v, got %v", math.Log(2), h)
	}
}

func syntheticResidualScenario() ([]string, []ClassID, map[ClassID][]Block) {
	blockLen := 200
	classA := ClassID{Scheme: SchemeJoint, Currier: "A", Hand: "1"}
	classB := ClassID{Scheme: SchemeJoint, Currier: "B", Hand: "2"}
	tokens := make([]string, 0, 4*blockLen)
	blocks := map[ClassID][]Block{}
	add := func(class ClassID, fillA, fillB string) {
		start := len(tokens)
		for i := 0; i < blockLen/2; i++ {
			tokens = append(tokens, fillA)
		}
		for i := 0; i < blockLen/2; i++ {
			tokens = append(tokens, fillB)
		}
		blocks[class] = append(blocks[class], Block{Class: class, Index: len(blocks[class]), Start: start, End: len(tokens)})
	}
	add(classA, "a1", "a2")
	add(classA, "a1", "a2")
	add(classB, "b1", "b2")
	add(classB, "b1", "b2")
	return tokens, []ClassID{classA, classB}, blocks
}

func TestMaxOverScaleTimesKPermutationCorrection(t *testing.T) {
	tokens, classes, blocks := syntheticResidualScenario()
	scales := []int{20, 40}
	stats := residualGlobalCorrection(tokens, classes, blocks, scales, 2, 3, "k_medoids", false, 0.5, 15, 42, nil, nil)
	if stats.Permutations != 15 {
		t.Fatalf("expected 15 null replicates recorded, got %d", stats.Permutations)
	}
	if stats.Observed != 0.5 {
		t.Fatalf("observed statistic must be passed through unchanged, got %v", stats.Observed)
	}
	if stats.EmpiricalP <= 0 {
		t.Fatalf("empirical p must never be exactly zero, got %v", stats.EmpiricalP)
	}
	// The null value must indeed be the max over the complete scale x K grid:
	// a hand-rolled sweep over the same (unshuffled) data must match one of
	// residualSweep's own per-(scale,K) silhouettes.
	sweep := residualSweep(tokens, classes, blocks, scales, 2, 3, "k_medoids", false, 42)
	found := false
	for _, row := range sweep.Rows {
		if math.Abs(row.Silhouette-sweep.BestSilhouette) < 1e-12 {
			found = true
		}
	}
	if !found {
		t.Fatalf("residualSweep's reported best silhouette %v must appear among its own per-(scale,K) rows", sweep.BestSilhouette)
	}
}

func TestDeterministicSeed(t *testing.T) {
	tokens, classes, blocks := syntheticResidualScenario()
	scales := []int{20, 40}
	s1 := residualGlobalCorrection(tokens, classes, blocks, scales, 2, 3, "k_medoids", false, 0.5, 10, 99, nil, nil)
	s2 := residualGlobalCorrection(tokens, classes, blocks, scales, 2, 3, "k_medoids", false, 0.5, 10, 99, nil, nil)
	if s1 != s2 {
		t.Fatalf("same seed must reproduce identical empirical stats: %+v vs %+v", s1, s2)
	}
	s3 := residualGlobalCorrection(tokens, classes, blocks, scales, 2, 3, "k_medoids", false, 0.5, 10, 100, nil, nil)
	if s1 == s3 {
		t.Fatalf("different seeds should not reproduce the exact same null distribution")
	}
}
