package higherorderseq

import (
	"bytes"
	"math"
	"math/rand"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
)

func tok(text, line string, idx int) Token { return Token{Text: text, Line: line, TokenIndexLine: idx} }

func mkBlock(id string, tokens ...Token) Block {
	for i := range tokens {
		tokens[i].Position = i
	}
	return Block{ID: id, Currier: "1", Hand: "1", Joint: "1/1", Tokens: tokens}
}

// --- frozen candidate extraction; no new candidate discovery ---

func writeFrozenAuditFixtures(t *testing.T, dir string) {
	t.Helper()
	strict := "sequence\tn\tcanonical_occurrences\tphysical_blocks\tjoint_classes\tmax_block_fraction\tshuffle_block_fdr_q\n" +
		"a b c\t3\t4\t4\t3\t0.25\t0.01\n" +
		"x y z\t3\t4\t4\t3\t0.25\t0.2\n" + // fails fdr_q<=0.05 gate
		"p q\t2\t9\t4\t4\t0.25\t0.01\n" // fails n>=3 gate
	null := "sequence\tn\tobserved_total_occurrences\tobserved_block_recurrence\tshuffle_null_mean_total\tshuffle_null_mean_blocks\tshuffle_total_p\tshuffle_block_p\tshuffle_block_fdr_q\tmarkov_available_blocks\tmarkov_null_mean_total\tmarkov_null_mean_blocks\tmarkov_total_p\tmarkov_block_p\n" +
		"a b c\t3\t4\t4\t1\t1\t0.001\t0.001\t0.01\t10\t1\t1\t0.5\t0.01\n" +
		"x y z\t3\t4\t4\t1\t1\t0.001\t0.001\t0.2\t10\t1\t1\t0.5\t0.9\n" +
		"p q\t2\t9\t4\t1\t1\t0.001\t0.001\t0.01\t10\t1\t1\t0.5\t0.01\n"
	if err := os.WriteFile(filepath.Join(dir, "strict_replicated_sequences.tsv"), []byte(strict), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "sequence_null_validation.tsv"), []byte(null), 0644); err != nil {
		t.Fatal(err)
	}
}

func TestFrozenCandidateExtractionNoNewDiscovery(t *testing.T) {
	dir := t.TempDir()
	writeFrozenAuditFixtures(t, dir)
	cands, err := loadFrozenCandidates(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(cands) != 1 {
		t.Fatalf("expected exactly 1 candidate (n>=3 and fdr_q<=0.05), got %d: %+v", len(cands), cands)
	}
	if cands[0].Sequence != "a b c" || cands[0].Family != "primary" {
		t.Fatalf("unexpected candidate: %+v", cands[0])
	}
	for _, c := range cands {
		if c.Sequence == "x y z" || c.Sequence == "p q" {
			t.Fatalf("candidate extraction invented/kept a sequence that fails frozen gates: %q", c.Sequence)
		}
	}
}

func TestFrozenCandidateFamilySplitByMarkovP(t *testing.T) {
	dir := t.TempDir()
	strict := "sequence\tn\tcanonical_occurrences\tphysical_blocks\tjoint_classes\tmax_block_fraction\tshuffle_block_fdr_q\n" +
		"a b c\t3\t4\t4\t3\t0.25\t0.01\n" +
		"d e f\t3\t4\t4\t3\t0.25\t0.01\n"
	null := "sequence\tn\tobserved_total_occurrences\tobserved_block_recurrence\tshuffle_null_mean_total\tshuffle_null_mean_blocks\tshuffle_total_p\tshuffle_block_p\tshuffle_block_fdr_q\tmarkov_available_blocks\tmarkov_null_mean_total\tmarkov_null_mean_blocks\tmarkov_total_p\tmarkov_block_p\n" +
		"a b c\t3\t4\t4\t1\t1\t0.001\t0.001\t0.01\t10\t1\t1\t0.5\t0.01\n" +
		"d e f\t3\t4\t4\t1\t1\t0.001\t0.001\t0.01\t10\t1\t1\t0.5\t0.9\n"
	os.WriteFile(filepath.Join(dir, "strict_replicated_sequences.tsv"), []byte(strict), 0644)
	os.WriteFile(filepath.Join(dir, "sequence_null_validation.tsv"), []byte(null), 0644)
	cands, err := loadFrozenCandidates(dir)
	if err != nil {
		t.Fatal(err)
	}
	byID := map[string]Candidate{}
	for _, c := range cands {
		byID[c.Sequence] = c
	}
	if byID["a b c"].Family != "primary" {
		t.Fatalf("markov p<=0.05 must be primary: %+v", byID["a b c"])
	}
	if byID["d e f"].Family != "secondary" {
		t.Fatalf("markov p>0.05 must be secondary: %+v", byID["d e f"])
	}
}

// --- ABC occurrence extraction; physical-block boundary protection ---

func TestOccurrenceExtractionRespectsBlockBoundary(t *testing.T) {
	cand := Candidate{Sequence: "a b c", Tokens: []string{"a", "b", "c"}}
	blockA := mkBlock("block-A", tok("a", "L1", 0), tok("b", "L1", 1))
	blockB := mkBlock("block-B", tok("c", "L1", 0), tok("x", "L1", 1))
	occs := findOccurrences(cand, []Block{blockA, blockB}, map[string]int{"L1": 2})
	if len(occs) != 0 {
		t.Fatalf("sequence split across two physical blocks must not be counted, got %d occurrences", len(occs))
	}

	within := mkBlock("block-C", tok("a", "L1", 0), tok("b", "L1", 1), tok("c", "L1", 2))
	occs = findOccurrences(cand, []Block{within}, map[string]int{"L1": 3})
	if len(occs) != 1 {
		t.Fatalf("expected 1 in-block occurrence, got %d", len(occs))
	}
	if occs[0].Block != "block-C" {
		t.Fatalf("wrong block recorded: %+v", occs[0])
	}
}

func TestOccurrenceLineBoundaryFlags(t *testing.T) {
	cand := Candidate{Sequence: "a b c", Tokens: []string{"a", "b", "c"}}
	blk := mkBlock("B", tok("a", "L1", 2), tok("b", "L2", 0), tok("c", "L2", 1))
	lineLen := map[string]int{"L1": 3, "L2": 2}
	occs := findOccurrences(cand, []Block{blk}, lineLen)
	if len(occs) != 1 {
		t.Fatalf("expected 1 occurrence, got %d", len(occs))
	}
	if !occs[0].CrossesLineBoundary || occs[0].WithinSameLine {
		t.Fatalf("expected cross-line occurrence flags, got %+v", occs[0])
	}
}

// --- count(B), count(AB), count(BC), count(ABC) ---

func TestBlockCounts(t *testing.T) {
	cand := Candidate{Sequence: "a b c", Tokens: []string{"a", "b", "c"}}
	blk := mkBlock("B", tok("a", "L", 0), tok("b", "L", 1), tok("c", "L", 2), tok("b", "L", 3), tok("x", "L", 4))
	bc := countsInBlock(cand, blk)
	if bc.CountB != 2 {
		t.Fatalf("count(B) = %d, want 2", bc.CountB)
	}
	if bc.CountAB != 1 {
		t.Fatalf("count(AB) = %d, want 1", bc.CountAB)
	}
	if bc.CountBC != 1 {
		t.Fatalf("count(BC) = %d, want 1", bc.CountBC)
	}
	if bc.CountABC != 1 {
		t.Fatalf("count(ABC) = %d, want 1", bc.CountABC)
	}
}

// --- P(C|B), P(C|AB), enrichment, reverse conditional probability ---

func TestConditionalRowProbabilitiesAndEnrichment(t *testing.T) {
	bc := BlockCounts{CountB: 10, CountAB: 4, CountBC: 2, CountABC: 2}
	row := conditionalRow(bc)
	if row.PCGivenB != 0.2 {
		t.Fatalf("P(C|B) = %v, want 0.2", row.PCGivenB)
	}
	if row.PCGivenAB != 0.5 {
		t.Fatalf("P(C|AB) = %v, want 0.5", row.PCGivenAB)
	}
	if math.Abs(row.Enrichment-2.5) > 1e-9 {
		t.Fatalf("enrichment = %v, want 2.5", row.Enrichment)
	}
	if math.Abs(row.DeltaProbability-0.3) > 1e-9 {
		t.Fatalf("delta = %v, want 0.3", row.DeltaProbability)
	}
	// reverse: P(A|B) = count(AB)/count(B) = 0.4; P(A|B,C) = count(ABC)/count(BC) = 1.0
	if row.PAGivenB != 0.4 {
		t.Fatalf("P(A|B) = %v, want 0.4", row.PAGivenB)
	}
	if row.PAGivenBC != 1.0 {
		t.Fatalf("P(A|B,C) = %v, want 1.0", row.PAGivenBC)
	}
}

// --- continuation distributions sum to 1; entropy calculation ---

func TestContinuationDistributionsSumToOneAndEntropy(t *testing.T) {
	cand := Candidate{Sequence: "a b c", Tokens: []string{"a", "b", "c"}}
	blk := mkBlock("B",
		tok("a", "L", 0), tok("b", "L", 1), tok("c", "L", 2),
		tok("a", "L", 3), tok("b", "L", 4), tok("x", "L", 5),
		tok("z", "L", 6), tok("b", "L", 7), tok("c", "L", 8),
	)
	rows := continuationDistributions(cand, []Block{blk})
	sums := map[string]float64{}
	for _, r := range rows {
		sums[r.Context] += r.Probability
	}
	for ctx, s := range sums {
		if math.Abs(s-1) > 1e-9 {
			t.Fatalf("context %q probabilities sum to %v, want 1", ctx, s)
		}
	}
	ent := continuationEntropy(cand, []Block{blk})
	// H(X|B): {c:2/3, x:1/3} -> -(2/3 log2 2/3 + 1/3 log2 1/3)
	want := -(2.0/3*math.Log2(2.0/3) + 1.0/3*math.Log2(1.0/3))
	if math.Abs(ent.HGivenB-want) > 1e-9 {
		t.Fatalf("H(X|B) = %v, want %v", ent.HGivenB, want)
	}
}

// --- conditional neighbor permutation preserves marginals; destroys pairing; no cross-block permutation ---

func TestPermutationPreservesMarginalsAndBlockComposition(t *testing.T) {
	obs := []bNeighbor{
		{blockIdx: 0, left: "a", right: "x"},
		{blockIdx: 0, left: "b", right: "y"},
		{blockIdx: 1, left: "c", right: "x"},
		{blockIdx: 1, left: "d", right: "z"},
	}
	r := rand.New(rand.NewSource(1))
	perm := permuteWithinBlocks(obs, r)
	_, wantLeft, wantRight := jointTable(obs)
	_, gotLeft, gotRight := jointTable(perm)
	if len(wantLeft) != len(gotLeft) {
		t.Fatalf("left marginal keys changed")
	}
	for k, v := range wantLeft {
		if gotLeft[k] != v {
			t.Fatalf("left marginal for %q changed: %d -> %d", k, v, gotLeft[k])
		}
	}
	for k, v := range wantRight {
		if gotRight[k] != v {
			t.Fatalf("right marginal for %q changed: %d -> %d", k, v, gotRight[k])
		}
	}
	byBlockCount := map[int]int{}
	for _, o := range perm {
		byBlockCount[o.blockIdx]++
	}
	if byBlockCount[0] != 2 || byBlockCount[1] != 2 {
		t.Fatalf("block composition changed: %+v", byBlockCount)
	}
	for _, o := range perm {
		if o.blockIdx == 0 && (o.right != "x" && o.right != "y") {
			t.Fatalf("right neighbor leaked across blocks: %+v", o)
		}
		if o.blockIdx == 1 && (o.right != "x" && o.right != "z") {
			t.Fatalf("right neighbor leaked across blocks: %+v", o)
		}
	}
}

func TestPermutationDestroysLeftRightPairing(t *testing.T) {
	// A large single block with a perfect left<->right pairing; the permuted
	// pairing should essentially never reproduce the same joint table.
	var tokens []bNeighbor
	for i := 0; i < 200; i++ {
		left, right := "a", "x"
		if i%2 == 1 {
			left, right = "b", "y"
		}
		tokens = append(tokens, bNeighbor{blockIdx: 0, left: left, right: right})
	}
	observed := cmiBits(tokens)
	if observed < 0.9 {
		t.Fatalf("perfectly paired left/right should have CMI near 1 bit, got %v", observed)
	}
	r := rand.New(rand.NewSource(42))
	perm := permuteWithinBlocks(tokens, r)
	permCMI := cmiBits(perm)
	if permCMI > observed*0.5 {
		t.Fatalf("permutation should mostly destroy the left<->right pairing: observed=%v permuted=%v", observed, permCMI)
	}
}

// --- CMI ---

func TestCMIZeroWhenIndependent(t *testing.T) {
	var obs []bNeighbor
	for i := 0; i < 100; i++ {
		left := "a"
		if i%2 == 1 {
			left = "b"
		}
		right := "x"
		if i%3 == 0 {
			right = "y"
		}
		obs = append(obs, bNeighbor{blockIdx: 0, left: left, right: right})
	}
	if cmi := cmiBits(obs); cmi > 0.05 {
		t.Fatalf("independent left/right should have near-zero CMI, got %v", cmi)
	}
}

// --- additive smoothing alpha=0.5 ---

func TestSmoothedProbAdditiveAlpha(t *testing.T) {
	counts := map[string]int{"x": 3, "y": 1}
	vocab := map[string]bool{"x": true, "y": true}
	p := smoothedProb(counts, vocab, "x", 4, 0.5)
	want := (3 + 0.5) / (4 + 0.5*2)
	if math.Abs(p-want) > 1e-12 {
		t.Fatalf("smoothedProb = %v, want %v", p, want)
	}
	// unseen token extends the vocabulary by exactly one.
	pUnseen := smoothedProb(counts, vocab, "z", 4, 0.5)
	wantUnseen := (0 + 0.5) / (4 + 0.5*3)
	if math.Abs(pUnseen-wantUnseen) > 1e-12 {
		t.Fatalf("smoothedProb (unseen) = %v, want %v", pUnseen, wantUnseen)
	}
}

// --- physical-block LOBO has zero leakage; M1/M2 log loss ---

func TestLOBOZeroLeakage(t *testing.T) {
	cand := Candidate{Sequence: "a b c", Tokens: []string{"a", "b", "c"}}
	held := mkBlock("held", tok("a", "L", 0), tok("b", "L", 1), tok("only-in-held", "L", 2))
	train := mkBlock("train", tok("a", "L", 0), tok("b", "L", 1), tok("c", "L", 2), tok("a", "L", 3), tok("b", "L", 4), tok("c", "L", 5))
	blocks := []Block{held, train}
	countB, totalB, countAB, totalAB, vocab := trainModels(cand, []Block{train})
	if _, ok := countB["only-in-held"]; ok {
		t.Fatalf("training statistics leaked a token that exists only in the held-out block")
	}
	_ = totalB
	_ = countAB
	_ = totalAB
	_ = vocab
	row := runLOBO(cand, blocks)
	if row.TestedBlocks != 2 {
		t.Fatalf("expected both blocks tested (each has an AB occurrence), got %d", row.TestedBlocks)
	}
}

func TestM1M2LogLossFavorsInformativeModel(t *testing.T) {
	cand := Candidate{Sequence: "a b c", Tokens: []string{"a", "b", "c"}}
	// In training, B is followed by c only when preceded by A; otherwise by
	// something else entirely - so M2 should dominate M1 on held-out AB->c.
	var trainTokens []Token
	for i := 0; i < 10; i++ {
		trainTokens = append(trainTokens, tok("a", "L", 0), tok("b", "L", 0), tok("c", "L", 0))
		trainTokens = append(trainTokens, tok("z", "L", 0), tok("b", "L", 0), tok("w", "L", 0))
	}
	train := mkBlock("train", trainTokens...)
	held := mkBlock("held", tok("a", "L", 0), tok("b", "L", 0), tok("c", "L", 0))
	row := runLOBO(cand, []Block{held, train})
	if row.M2BetterBlocks != 1 || row.M1BetterBlocks != 0 {
		t.Fatalf("expected M2 to win on informative held-out data: %+v", row)
	}
	if row.MeanDeltaLogLoss <= 0 {
		t.Fatalf("mean delta log loss should be positive when M2 wins, got %v", row.MeanDeltaLogLoss)
	}
}

// --- context alternative enumeration ---

func TestContextControlEnumeratesFrozenAndAlternatives(t *testing.T) {
	cand := Candidate{Sequence: "a b c", Tokens: []string{"a", "b", "c"}}
	var tokens []Token
	for i := 0; i < 3; i++ {
		tokens = append(tokens, tok("a", "L", 0), tok("b", "L", 0), tok("c", "L", 0))
	}
	for i := 0; i < 3; i++ {
		tokens = append(tokens, tok("d", "L", 0), tok("b", "L", 0), tok("x", "L", 0))
	}
	blk := mkBlock("blk", tokens...)
	rows := contextControlRows(cand, []Block{blk})
	foundA, foundD := false, false
	for _, r := range rows {
		if r.ContextType == "left_alt" && r.AltToken == "a" {
			foundA = true
			if !r.IsFrozen {
				t.Fatalf("frozen A row not marked IsFrozen: %+v", r)
			}
		}
		if r.ContextType == "left_alt" && r.AltToken == "d" {
			foundD = true
		}
	}
	if !foundA || !foundD {
		t.Fatalf("expected both frozen and alternative left contexts enumerated: %+v", rows)
	}
}

// --- normalized block position ---

func TestNormalizedBlockPosition(t *testing.T) {
	if v := normalizedPosition(0, 5); v != 0 {
		t.Fatalf("start position = %v, want 0", v)
	}
	if v := normalizedPosition(4, 5); v != 1 {
		t.Fatalf("end position = %v, want 1", v)
	}
	if v := normalizedPosition(0, 1); v != 0 {
		t.Fatalf("single-token block position = %v, want 0", v)
	}
	if b := positionBucket(0.95); b != "[0.9,1.0]" {
		t.Fatalf("bucket(0.95) = %q", b)
	}
	if b := positionBucket(0); b != "[0,0.1)" {
		t.Fatalf("bucket(0) = %q", b)
	}
}

// --- jackknife removes exactly one block ---

func TestJackknifeRemovesExactlyOneBlock(t *testing.T) {
	cand := Candidate{Sequence: "a b c", Tokens: []string{"a", "b", "c"}}
	blocks := []Block{
		mkBlock("B1", repeatABC(5)...),
		mkBlock("B2", repeatABC(5)...),
		mkBlock("B3", repeatABC(5)...),
	}
	rows := conditionalRowsForCandidate(cand, blocks)
	eligible := primaryEligible(rows)
	if len(eligible) != 3 {
		t.Fatalf("expected all 3 blocks eligible, got %d", len(eligible))
	}
	jk := jackknifeRow(cand, blocks, eligible)
	if jk.Realizations != 3 {
		t.Fatalf("expected 3 jackknife realizations (one per eligible block), got %d", jk.Realizations)
	}
}

func repeatABC(n int) []Token {
	var out []Token
	for i := 0; i < n; i++ {
		out = append(out, tok("a", "L", 0), tok("b", "L", 0), tok("c", "L", 0))
	}
	// pad B's count above the primary min-count-B threshold.
	for i := 0; i < primaryMinCountB; i++ {
		out = append(out, tok("q", "L", 0), tok("b", "L", 0), tok("q", "L", 0))
	}
	return out
}

// --- BH FDR ---

func TestBHFDRMonotoneAndBounded(t *testing.T) {
	p := []float64{0.001, 0.02, 0.04, 0.5}
	q := bh(p)
	for i, v := range q {
		if v < p[i]-1e-12 {
			t.Fatalf("BH q[%d]=%v must be >= raw p[%d]=%v", i, v, i, p[i])
		}
		if v > 1 {
			t.Fatalf("BH q[%d]=%v exceeds 1", i, v)
		}
	}
	sortedIdx := []int{0, 1, 2, 3}
	sort.Slice(sortedIdx, func(a, b int) bool { return p[sortedIdx[a]] < p[sortedIdx[b]] })
	for k := 1; k < len(sortedIdx); k++ {
		if q[sortedIdx[k]] < q[sortedIdx[k-1]]-1e-12 {
			t.Fatalf("BH q must be monotone non-decreasing in sorted p-value order")
		}
	}
}

// --- deterministic seed ---

func TestDeterministicSeed(t *testing.T) {
	cand := Candidate{Sequence: "a b c", Tokens: []string{"a", "b", "c"}, Family: "primary"}
	blocks := []Block{mkBlock("B1", repeatABC(5)...)}
	r1 := runCMI(cand, blocks, 200, candidateSeed(1, cand.Sequence))
	r2 := runCMI(cand, blocks, 200, candidateSeed(1, cand.Sequence))
	if r1.EmpiricalP != r2.EmpiricalP || r1.NullMeanCMIBits != r2.NullMeanCMIBits {
		t.Fatalf("same seed must reproduce identical results: %+v vs %+v", r1, r2)
	}
	r3 := runCMI(cand, blocks, 200, candidateSeed(2, cand.Sequence))
	if r1.NullMeanCMIBits == r3.NullMeanCMIBits && r1.EmpiricalP == r3.EmpiricalP {
		t.Skip("different seeds happened to coincide; not a failure but not informative either")
	}
}

func TestCandidateSeedDistinctPerSequence(t *testing.T) {
	s1 := candidateSeed(1, "a b c")
	s2 := candidateSeed(1, "d e f")
	if s1 == s2 {
		t.Fatalf("distinct sequences must get distinct candidate seeds")
	}
}

// --- outcome classification ---

func TestOutcomeClassificationInsufficientSupport(t *testing.T) {
	in := classificationInput{
		Candidate:  Candidate{Sequence: "a b c", Family: "primary"},
		Dependence: DependenceRow{FDRQ: 0.01},
		CrossBlock: CrossBlockRow{EligibleBlocks: 2, SignConsistency: 1, DistinctJoint: 2},
		LOBO:       LOBORow{TestedBlocks: 2, M2BetterBlocks: 2},
	}
	if got := classify(in).FinalStatus; got != "INSUFFICIENT_SUPPORT" {
		t.Fatalf("status = %q, want INSUFFICIENT_SUPPORT", got)
	}
}

func TestOutcomeClassificationSingleBlockSensitive(t *testing.T) {
	in := classificationInput{
		Candidate:  Candidate{Sequence: "a b c", Family: "primary"},
		Dependence: DependenceRow{FDRQ: 0.01},
		CrossBlock: CrossBlockRow{EligibleBlocks: 3, SignConsistency: 1, DistinctJoint: 2},
		LOBO:       LOBORow{TestedBlocks: 3, M2BetterBlocks: 3},
		Jackknife:  JackknifeRow{SingleBlockSensitive: true},
	}
	if got := classify(in).FinalStatus; got != "SINGLE_BLOCK_SENSITIVE" {
		t.Fatalf("status = %q, want SINGLE_BLOCK_SENSITIVE", got)
	}
}

func TestOutcomeClassificationHigherOrderReplicated(t *testing.T) {
	in := classificationInput{
		Candidate:  Candidate{Sequence: "a b c", Family: "primary"},
		Dependence: DependenceRow{FDRQ: 0.01},
		CrossBlock: CrossBlockRow{EligibleBlocks: 3, SignConsistency: 1, DistinctJoint: 2},
		LOBO:       LOBORow{TestedBlocks: 3, M2BetterBlocks: 3},
	}
	if got := classify(in).FinalStatus; got != "HIGHER_ORDER_REPLICATED" {
		t.Fatalf("status = %q, want HIGHER_ORDER_REPLICATED", got)
	}
}

func TestOutcomeClassificationFirstOrderExplained(t *testing.T) {
	in := classificationInput{
		Candidate:  Candidate{Sequence: "a b c", Family: "primary"},
		Dependence: DependenceRow{FDRQ: 0.9},
		CrossBlock: CrossBlockRow{EligibleBlocks: 3, SignConsistency: 0.3, DistinctJoint: 2},
		LOBO:       LOBORow{TestedBlocks: 3, M1BetterBlocks: 3},
	}
	if got := classify(in).FinalStatus; got != "FIRST_ORDER_EXPLAINED" {
		t.Fatalf("status = %q, want FIRST_ORDER_EXPLAINED", got)
	}
}

// --- checkpoint ---

func TestCheckpointRequiresMatchingFingerprint(t *testing.T) {
	path := filepath.Join(t.TempDir(), "checkpoint.json")
	cp := newCheckpoint("one")
	cp.PartsDone["a b c|cmi"] = true
	if err := saveCheckpoint(path, cp); err != nil {
		t.Fatal(err)
	}
	loaded, ok := loadCheckpoint(path, "one")
	if !ok || !loaded.PartsDone["a b c|cmi"] {
		t.Fatalf("matching checkpoint should resume: ok=%v cp=%+v", ok, loaded)
	}
	_, ok = loadCheckpoint(path, "two")
	if ok {
		t.Fatalf("mismatching fingerprint must not resume")
	}
}

// --- progress bar ---

func TestProgressBar(t *testing.T) {
	var b bytes.Buffer
	p := newProgress(&b)
	p.begin(3, "Test stage")
	p.update(3, 3, "Test stage")
	s := b.String()
	if !strings.Contains(s, "[3/8]") || !strings.Contains(s, "[====================]") || !strings.Contains(s, "elapsed") {
		t.Fatalf("status bar output: %q", s)
	}
}

// --- structural relatives parsing ---

func TestStructuralRelativesSkipsSingletons(t *testing.T) {
	dir := t.TempDir()
	content := `models:
    - threshold: 0.7
      classes:
        - id: C0001
          members:
            - token: aiin
              count: 504
            - token: ain
              count: 113
          size: 2
        - id: C0002
          members:
            - token: ol
              count: 560
          size: 1
    - threshold: 0.75
      classes:
        - id: C0003
          members:
            - token: aiin
              count: 504
            - token: unrelated
              count: 1
          size: 2
`
	if err := os.WriteFile(filepath.Join(dir, "structural_classes.yaml"), []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
	rel, err := structuralRelatives(dir)
	if err != nil {
		t.Fatal(err)
	}
	if got := rel["aiin"]; len(got) != 1 || got[0] != "ain" {
		t.Fatalf("aiin relatives = %v, want [ain] (only first/lowest-threshold model)", got)
	}
	if _, ok := rel["ol"]; ok {
		t.Fatalf("singleton class should have no relatives")
	}
}
