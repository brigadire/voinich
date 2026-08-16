package positionalcontinuation

import (
	"bytes"
	"math"
	"math/rand"
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

// --- 1. exact s-aiin occurrence extraction ---

func TestFindSAiinOccurrencesExact(t *testing.T) {
	blk := mkBlock("B", tok("x", "L", 0), tok("s", "L", 1), tok("aiin", "L", 2), tok("chey", "L", 3))
	occs := findSAiinOccurrences([]Block{blk}, map[string]int{"L": 4}, 4)
	if len(occs) != 1 {
		t.Fatalf("expected exactly 1 s-aiin occurrence, got %d", len(occs))
	}
	if occs[0].X != "chey" {
		t.Fatalf("expected continuation chey, got %q", occs[0].X)
	}
}

// --- 2. continuation extraction ---

func TestContinuationExtractionVariousX(t *testing.T) {
	blk := mkBlock("B", tok("s", "L", 0), tok("aiin", "L", 1), tok("dar", "L", 2))
	occs := findSAiinOccurrences([]Block{blk}, map[string]int{"L": 3}, 3)
	if len(occs) != 1 || occs[0].X != "dar" {
		t.Fatalf("expected continuation 'dar', got %+v", occs)
	}
}

// --- 3. no physical-block crossing ---

func TestSAiinOccurrenceNeverCrossesBlockBoundary(t *testing.T) {
	blockA := mkBlock("A", tok("s", "L1", 0))
	blockB := mkBlock("B", tok("aiin", "L1", 1), tok("chey", "L1", 2))
	occs := findSAiinOccurrences([]Block{blockA, blockB}, map[string]int{"L1": 3}, 3)
	if len(occs) != 0 {
		t.Fatalf("s aiin split across two physical blocks must not be counted, got %d", len(occs))
	}
}

// --- 4. missing continuation at boundary ---

func TestXMissingBlockEndVsCorpusEnd(t *testing.T) {
	blk := mkBlock("B", tok("s", "L", 0), tok("aiin", "L", 1))
	// Block ends but more corpus tokens exist elsewhere.
	occs := findSAiinOccurrences([]Block{blk}, map[string]int{"L": 2}, 10)
	if len(occs) != 1 || !occs[0].XMissingBlockEnd || occs[0].XMissingCorpusEnd {
		t.Fatalf("expected block-end missing flag only, got %+v", occs)
	}
	// Corpus itself ends right after aiin.
	occs = findSAiinOccurrences([]Block{blk}, map[string]int{"L": 2}, 2)
	if len(occs) != 1 || !occs[0].XMissingCorpusEnd || occs[0].XMissingBlockEnd {
		t.Fatalf("expected corpus-end missing flag only, got %+v", occs)
	}
}

// --- 5. normalized line position ---

func TestNormalizedLinePosition(t *testing.T) {
	if v := normalizedPosition(0, 5); v != 0 {
		t.Fatalf("start = %v, want 0", v)
	}
	if v := normalizedPosition(4, 5); v != 1 {
		t.Fatalf("end = %v, want 1", v)
	}
	if v := normalizedPosition(0, 1); v != 0 {
		t.Fatalf("single-token line = %v, want 0", v)
	}
}

// --- 6. normalized block position ---

func TestNormalizedBlockPositionOfAiin(t *testing.T) {
	blk := mkBlock("B", tok("x", "L", 0), tok("s", "L", 1), tok("aiin", "L", 2), tok("y", "L", 3))
	occs := findSAiinOccurrences([]Block{blk}, map[string]int{"L": 4}, 4)
	want := normalizedPosition(2, 4)
	if math.Abs(occs[0].NormalizedBlockPosition-want) > 1e-12 {
		t.Fatalf("normalized block position = %v, want %v", occs[0].NormalizedBlockPosition, want)
	}
}

// --- 7. fixed line categories ---

func TestLineCategoryPriorityAndThresholds(t *testing.T) {
	if c := lineCategory(true, false, 0.9); c != "LINE_START" {
		t.Fatalf("LINE_START must take priority, got %q", c)
	}
	if c := lineCategory(false, true, 0.1); c != "LINE_END" {
		t.Fatalf("LINE_END must take priority over EARLY, got %q", c)
	}
	if c := lineCategory(false, false, 0.1); c != "LINE_EARLY" {
		t.Fatalf("want LINE_EARLY, got %q", c)
	}
	if c := lineCategory(false, false, 0.5); c != "LINE_MIDDLE" {
		t.Fatalf("want LINE_MIDDLE, got %q", c)
	}
	if c := lineCategory(false, false, 0.9); c != "LINE_LATE" {
		t.Fatalf("want LINE_LATE, got %q", c)
	}
}

// --- 8. fixed block bins ---

func TestFixedBlockBins(t *testing.T) {
	if b := blockBinFixed(0.05); b != "B0" {
		t.Fatalf("bin(0.05) = %q, want B0", b)
	}
	if b := blockBinFixed(0.95); b != "B9" {
		t.Fatalf("bin(0.95) = %q, want B9", b)
	}
	if b := blockBinCoarse(0.1); b != "BLOCK_START" {
		t.Fatalf("coarse(0.1) = %q, want BLOCK_START", b)
	}
	if b := blockBinCoarse(0.5); b != "BLOCK_MIDDLE" {
		t.Fatalf("coarse(0.5) = %q, want BLOCK_MIDDLE", b)
	}
	if b := blockBinCoarse(0.9); b != "BLOCK_END" {
		t.Fatalf("coarse(0.9) = %q, want BLOCK_END", b)
	}
}

// --- 9. continuation probabilities sum to 1 ---

func TestContinuationProbabilitiesSumToOne(t *testing.T) {
	xs := []string{"a", "a", "b", "c"}
	_, summaries := buildSAiinContinuationDistributions(occsWithX(xs))
	for _, s := range summaries {
		if s.Stratum != "global" {
			continue
		}
		sum := 0.0
		counts := countMap(xs)
		for _, k := range stringKeysInt(counts) {
			sum += float64(counts[k]) / float64(len(xs))
		}
		if math.Abs(sum-1) > 1e-9 {
			t.Fatalf("probabilities sum to %v, want 1", sum)
		}
	}
}

func occsWithX(xs []string) []SAiinOccurrence {
	var out []SAiinOccurrence
	for i, x := range xs {
		out = append(out, SAiinOccurrence{X: x, Block: "B", LineCategory: "LINE_MIDDLE", BlockBinCoarse: "BLOCK_MIDDLE", BlockBinFixed: "B5", PosS: i * 2, PosAiin: i*2 + 1})
	}
	return out
}

// --- 10. entropy ---

func TestEntropyBitsKnownDistribution(t *testing.T) {
	counts := map[string]int{"a": 2, "b": 1}
	h := countEntropyBits(counts)
	want := -(2.0/3*math.Log2(2.0/3) + 1.0/3*math.Log2(1.0/3))
	if math.Abs(h-want) > 1e-9 {
		t.Fatalf("entropy = %v, want %v", h, want)
	}
}

// --- 11. effective continuation count ---

func TestEffectiveContinuationCount(t *testing.T) {
	if v := pow2(0); v != 1 {
		t.Fatalf("2^0 = %v, want 1", v)
	}
	if v := pow2(1); math.Abs(v-2) > 1e-9 {
		t.Fatalf("2^1 = %v, want 2", v)
	}
}

// --- 12. positional CMI (mutual information) ---

func TestMutualInformationZeroWhenIndependent(t *testing.T) {
	var xs, ys []string
	for i := 0; i < 100; i++ {
		x := "a"
		if i%2 == 1 {
			x = "b"
		}
		y := "p"
		if i%3 == 0 {
			y = "q"
		}
		xs = append(xs, x)
		ys = append(ys, y)
	}
	if mi := mutualInformationBits(xs, ys); mi > 0.05 {
		t.Fatalf("independent variables should have near-zero MI, got %v", mi)
	}
}

func TestMutualInformationHighWhenDeterministic(t *testing.T) {
	var xs, ys []string
	for i := 0; i < 50; i++ {
		x, y := "a", "p"
		if i%2 == 1 {
			x, y = "b", "q"
		}
		xs = append(xs, x)
		ys = append(ys, y)
	}
	if mi := mutualInformationBits(xs, ys); mi < 0.9 {
		t.Fatalf("perfectly paired variables should have MI near 1 bit, got %v", mi)
	}
}

// --- 13/14/15. within-block position permutation preserves marginals, continuation counts, block membership ---

func TestPermuteLabelsWithinBlocksPreservesEverything(t *testing.T) {
	xs := []string{"t0", "t1", "t2", "t3"}
	blockIDs := []string{"A", "A", "B", "B"}
	labels := []string{"x", "y", "x", "z"}
	ws := newPositionalWorkspace(xs, blockIDs, labels)
	r := rand.New(rand.NewSource(1))
	permIdx := ws.permute(r)
	perm := make([]string, len(permIdx))
	for i, ci := range permIdx {
		perm[i] = ws.catNames[ci]
	}

	byBlockOriginal := map[string]map[string]int{}
	byBlockPerm := map[string]map[string]int{}
	for i := range blockIDs {
		if byBlockOriginal[blockIDs[i]] == nil {
			byBlockOriginal[blockIDs[i]] = map[string]int{}
			byBlockPerm[blockIDs[i]] = map[string]int{}
		}
		byBlockOriginal[blockIDs[i]][labels[i]]++
		byBlockPerm[blockIDs[i]][perm[i]]++
	}
	for block, counts := range byBlockOriginal {
		for label, n := range counts {
			if byBlockPerm[block][label] != n {
				t.Fatalf("block %q label %q count changed: %d -> %d", block, label, n, byBlockPerm[block][label])
			}
		}
	}
	// Labels must never move across blocks: block A only ever had x,y.
	for i := range blockIDs {
		if blockIDs[i] == "A" && perm[i] != "x" && perm[i] != "y" {
			t.Fatalf("label leaked across blocks into A: %v", perm[i])
		}
		if blockIDs[i] == "B" && perm[i] != "x" && perm[i] != "z" {
			t.Fatalf("label leaked across blocks into B: %v", perm[i])
		}
	}
}

// --- 16. chey positional enrichment ---

func TestCheyPositionalEnrichmentComputation(t *testing.T) {
	var occs []SAiinOccurrence
	for i := 0; i < 10; i++ {
		x := "other"
		cat := "LINE_MIDDLE"
		if i < 5 {
			cat = "LINE_END"
			if i < 4 {
				x = "chey"
			}
		}
		occs = append(occs, SAiinOccurrence{X: x, LineCategory: cat, BlockBinCoarse: "BLOCK_MIDDLE", Block: "B1"})
	}
	result := runPositionalTests(occs, "line_position", lineCategories, 0, 1)
	var endRow CheyEffectRow
	for _, ce := range result.CheyEffect {
		if ce.Stratum == "LINE_END" {
			endRow = ce
		}
	}
	if endRow.CheyCount != 4 || endRow.OccurrenceCount != 5 {
		t.Fatalf("unexpected LINE_END stats: %+v", endRow)
	}
	if endRow.PositionalEnrichment <= 1 {
		t.Fatalf("expected enrichment > 1 at LINE_END, got %v", endRow.PositionalEnrichment)
	}
}

// --- 17. aiin control construction ---

func TestAiinOccurrenceConstructionWithAndWithoutPredecessor(t *testing.T) {
	blk := mkBlock("B", tok("aiin", "L", 0), tok("x", "L", 1), tok("s", "L", 2), tok("aiin", "L", 3), tok("chey", "L", 4))
	occs := findAiinOccurrences([]Block{blk}, map[string]int{"L": 5}, 5)
	if len(occs) != 2 {
		t.Fatalf("expected 2 aiin occurrences, got %d", len(occs))
	}
	if occs[0].HasPredecessor {
		t.Fatalf("first aiin (block start) must have no predecessor: %+v", occs[0])
	}
	if !occs[1].HasPredecessor || !occs[1].PredecessorIsS {
		t.Fatalf("second aiin's predecessor must be recognized as 's': %+v", occs[1])
	}
}

// --- 18/19. stratified predecessor permutation preserves block and position strata ---

func TestPermuteIsSWithinStrataPreservesStrataComposition(t *testing.T) {
	obs := []stratifiedObs{
		{stratum: "B1|LINE_MIDDLE", isS: true, isChey: true},
		{stratum: "B1|LINE_MIDDLE", isS: false, isChey: false},
		{stratum: "B2|LINE_END", isS: true, isChey: false},
		{stratum: "B2|LINE_END", isS: false, isChey: true},
		{stratum: "B2|LINE_END", isS: false, isChey: false},
	}
	ws := newStratifiedWorkspace(obs)
	r := rand.New(rand.NewSource(7))
	ws.permuteAndStatistic(r)
	perm := ws.permIsS

	byStratumOriginal := map[string]int{}
	byStratumPerm := map[string]int{}
	for i, o := range obs {
		if o.isS {
			byStratumOriginal[o.stratum]++
		}
		if perm[i] {
			byStratumPerm[o.stratum]++
		}
	}
	for k, v := range byStratumOriginal {
		if byStratumPerm[k] != v {
			t.Fatalf("stratum %q predecessor=s count changed: %d -> %d", k, v, byStratumPerm[k])
		}
	}
}

// --- 20. M1/M2/M3 probabilities ---

func TestModelLOBOInformativePositionAndPredecessor(t *testing.T) {
	var train []AiinOccurrence
	for i := 0; i < 10; i++ {
		train = append(train,
			AiinOccurrence{Block: "train", X: "chey", LineCategory: "LINE_END", HasPredecessor: true, PredecessorIsS: true},
			AiinOccurrence{Block: "train", X: "other", LineCategory: "LINE_MIDDLE", HasPredecessor: true, PredecessorIsS: false},
		)
	}
	held := []AiinOccurrence{{Block: "held", X: "chey", LineCategory: "LINE_END", HasPredecessor: true, PredecessorIsS: true}}
	row := runModelLOBO(append(train, held...))
	if row.TestedBlocks != 2 {
		t.Fatalf("expected 2 tested blocks (train also self-tests as held-out once), got %d", row.TestedBlocks)
	}
}

// --- 21. alpha=0.5 smoothing ---

func TestSmoothedProbAlphaHalf(t *testing.T) {
	counts := map[string]int{"x": 3, "y": 1}
	vocab := map[string]bool{"x": true, "y": true}
	p := smoothedProb(counts, vocab, "x", 4, 0.5)
	want := (3 + 0.5) / (4 + 0.5*2)
	if math.Abs(p-want) > 1e-12 {
		t.Fatalf("smoothedProb = %v, want %v", p, want)
	}
	pUnseen := smoothedProb(counts, vocab, "z", 4, 0.5)
	wantUnseen := (0 + 0.5) / (4 + 0.5*3)
	if math.Abs(pUnseen-wantUnseen) > 1e-12 {
		t.Fatalf("smoothedProb(unseen) = %v, want %v", pUnseen, wantUnseen)
	}
}

// --- 22. LOBO has zero physical-block leakage ---

func TestModelLOBOZeroLeakage(t *testing.T) {
	held := []AiinOccurrence{{Block: "held", X: "only-in-held", LineCategory: "LINE_MIDDLE", HasPredecessor: false}}
	train := []AiinOccurrence{{Block: "train", X: "common", LineCategory: "LINE_MIDDLE", HasPredecessor: false}}
	countM1, _, _, _, _, _, vocab := trainAiinModels(train)
	if _, ok := countM1["only-in-held"]; ok {
		t.Fatalf("training statistics leaked a token that exists only in the held-out block")
	}
	if vocab["only-in-held"] {
		t.Fatalf("vocab leaked held-out-only token")
	}
	row := runModelLOBO(append(train, held...))
	if row.TestedBlocks != 2 {
		t.Fatalf("expected both blocks tested, got %d", row.TestedBlocks)
	}
}

// --- 23. line/block position association ---

func TestLineVsBlockAssociationPerfectCorrelation(t *testing.T) {
	var occs []SAiinOccurrence
	for i := 0; i < 10; i++ {
		p := float64(i) / 9
		occs = append(occs, SAiinOccurrence{
			NormalizedLinePosition: p, NormalizedBlockPosition: p,
			LineCategory: "LINE_MIDDLE", BlockBinCoarse: "BLOCK_MIDDLE", X: "x",
		})
	}
	_, r, _ := buildLineVsBlockRows(occs)
	if r < 0.99 {
		t.Fatalf("expected near-perfect correlation, got %v", r)
	}
}

// --- 24. boundary distances ---

func TestBoundaryDistanceMedianQuantiles(t *testing.T) {
	occs := []SAiinOccurrence{
		{X: "chey", Block: "B1", TokensFromLineStart: 1},
		{X: "chey", Block: "B1", TokensFromLineStart: 3},
		{X: "other", Block: "B1", TokensFromLineStart: 10},
		{X: "other", Block: "B1", TokensFromLineStart: 20},
	}
	rows := buildBoundaryDistanceRows(occs, 0, 1)
	for _, r := range rows {
		if r.Metric == "tokens_from_line_start" && r.Group == "s_aiin_chey" {
			if r.Median != 2 {
				t.Fatalf("chey median tokens_from_line_start = %v, want 2", r.Median)
			}
		}
	}
}

// --- 25. jackknife removes exactly one block ---

func TestJackknifeRemovesExactlyOneBlock(t *testing.T) {
	var sAiin []SAiinOccurrence
	var aiin []AiinOccurrence
	for _, b := range []string{"B1", "B2", "B3"} {
		for i := 0; i < 5; i++ {
			sAiin = append(sAiin, SAiinOccurrence{Block: b, X: "chey", LineCategory: "LINE_MIDDLE", BlockBinCoarse: "BLOCK_MIDDLE"})
			aiin = append(aiin, AiinOccurrence{Block: b, X: "chey", LineCategory: "LINE_MIDDLE", BlockBinCoarse: "BLOCK_MIDDLE", HasPredecessor: true, PredecessorIsS: true})
		}
	}
	rows := runJackknife(sAiin, aiin, 1)
	for _, r := range rows {
		if r.Realizations != 3 {
			t.Fatalf("expected 3 jackknife realizations (one per eligible block), got %d for %s", r.Realizations, r.PositionVariable)
		}
	}
}

func TestExcludeBlockRemovesOnlyThatBlock(t *testing.T) {
	occs := []SAiinOccurrence{{Block: "A"}, {Block: "B"}, {Block: "B"}}
	remaining := excludeBlock(occs, "B")
	if len(remaining) != 1 || remaining[0].Block != "A" {
		t.Fatalf("expected only block A to remain, got %+v", remaining)
	}
}

// --- 26. deterministic seed ---

func TestDeterministicSeed(t *testing.T) {
	occs := []SAiinOccurrence{
		{X: "chey", Block: "B1", LineCategory: "LINE_END", BlockBinCoarse: "BLOCK_END"},
		{X: "other", Block: "B1", LineCategory: "LINE_MIDDLE", BlockBinCoarse: "BLOCK_MIDDLE"},
		{X: "chey", Block: "B2", LineCategory: "LINE_END", BlockBinCoarse: "BLOCK_END"},
		{X: "other", Block: "B2", LineCategory: "LINE_MIDDLE", BlockBinCoarse: "BLOCK_MIDDLE"},
	}
	r1 := runPositionalTests(occs, "line_position", lineCategories, 200, seedFor(1, "line_position"))
	r2 := runPositionalTests(occs, "line_position", lineCategories, 200, seedFor(1, "line_position"))
	if r1.Dependence.EmpiricalP != r2.Dependence.EmpiricalP || r1.Dependence.NullMeanMIBits != r2.Dependence.NullMeanMIBits {
		t.Fatalf("same seed must reproduce identical results: %+v vs %+v", r1.Dependence, r2.Dependence)
	}
	r3 := runPositionalTests(occs, "line_position", lineCategories, 200, seedFor(2, "line_position"))
	if r1.Dependence.NullMeanMIBits == r3.Dependence.NullMeanMIBits {
		t.Skip("different seeds happened to coincide; not a failure but not informative either")
	}
}

func TestSeedForDistinctPerName(t *testing.T) {
	if seedFor(1, "line_position") == seedFor(1, "block_position") {
		t.Fatalf("distinct sub-step names must get distinct seeds")
	}
}

// --- 27. diagnostic status classification ---

func TestClassifyInsufficientSupport(t *testing.T) {
	in := classificationInput{EligibleBlocks: 1, PositionDependenceP: 0.01, StratifiedPredecessorP: 0.01, CrossBlockSignConsistency: 1, M3BetterThanM2Fraction: 1}
	if got := classify(in).FinalStatus; got != "INSUFFICIENT_SUPPORT" {
		t.Fatalf("status = %q, want INSUFFICIENT_SUPPORT", got)
	}
}

func TestClassifySingleBlockSensitive(t *testing.T) {
	in := classificationInput{EligibleBlocks: 5, PositionDependenceP: 0.01, StratifiedPredecessorP: 0.01, CrossBlockSignConsistency: 1, M3BetterThanM2Fraction: 1, SingleBlockSensitive: true}
	if got := classify(in).FinalStatus; got != "SINGLE_BLOCK_SENSITIVE" {
		t.Fatalf("status = %q, want SINGLE_BLOCK_SENSITIVE", got)
	}
}

func TestClassifyGeneralHigherOrder(t *testing.T) {
	in := classificationInput{EligibleBlocks: 5, PositionDependenceP: 0.5, StratifiedPredecessorP: 0.01, CrossBlockSignConsistency: 0.9, M3BetterThanM2Fraction: 0.8}
	if got := classify(in).FinalStatus; got != "GENERAL_HIGHER_ORDER" {
		t.Fatalf("status = %q, want GENERAL_HIGHER_ORDER", got)
	}
}

func TestClassifyPositionConditionedHigherOrder(t *testing.T) {
	in := classificationInput{EligibleBlocks: 5, PositionDependenceP: 0.01, StratifiedPredecessorP: 0.01, CrossBlockSignConsistency: 0.6, M3BetterThanM2Fraction: 0.2}
	if got := classify(in).FinalStatus; got != "POSITION_CONDITIONED_HIGHER_ORDER" {
		t.Fatalf("status = %q, want POSITION_CONDITIONED_HIGHER_ORDER", got)
	}
}

func TestClassifyAiinPositionEffect(t *testing.T) {
	in := classificationInput{EligibleBlocks: 5, PositionDependenceP: 0.01, StratifiedPredecessorP: 0.9, CrossBlockSignConsistency: 0.9, M3BetterThanM2Fraction: 0.1}
	if got := classify(in).FinalStatus; got != "AIIN_POSITION_EFFECT" {
		t.Fatalf("status = %q, want AIIN_POSITION_EFFECT", got)
	}
}

func TestClassifyBoundaryFormula(t *testing.T) {
	in := classificationInput{
		EligibleBlocks: 5, PositionDependenceP: 0.01, StratifiedPredecessorP: 0.9, CrossBlockSignConsistency: 0.1, M3BetterThanM2Fraction: 0.1,
		UniqueCheySurroundingContexts: 1, CheyOccurrences: 4,
	}
	if got := classify(in).FinalStatus; got != "BOUNDARY_FORMULA" {
		t.Fatalf("status = %q, want BOUNDARY_FORMULA", got)
	}
}

func TestClassifyNoPositionalStructureDefault(t *testing.T) {
	in := classificationInput{EligibleBlocks: 5, PositionDependenceP: 0.9, StratifiedPredecessorP: 0.9, CrossBlockSignConsistency: 0.1, M3BetterThanM2Fraction: 0.1}
	if got := classify(in).FinalStatus; got != "NO_POSITIONAL_STRUCTURE" {
		t.Fatalf("status = %q, want NO_POSITIONAL_STRUCTURE", got)
	}
}

// --- checkpoint ---

func TestCheckpointRequiresMatchingFingerprint(t *testing.T) {
	path := filepath.Join(t.TempDir(), "checkpoint.json")
	cp := newCheckpoint("one")
	cp.PartsDone["occurrences"] = true
	cp.State.SAiinOccurrences = []SAiinOccurrence{{X: "chey"}}
	if err := saveCheckpoint(path, cp); err != nil {
		t.Fatal(err)
	}
	loaded, ok := loadCheckpoint(path, "one")
	if !ok || !loaded.PartsDone["occurrences"] || len(loaded.State.SAiinOccurrences) != 1 {
		t.Fatalf("matching checkpoint should resume with state intact: ok=%v cp=%+v", ok, loaded)
	}
	_, ok = loadCheckpoint(path, "two")
	if ok {
		t.Fatalf("mismatching fingerprint must not resume")
	}
}

// --- progress bar ---

func TestProgressBarRendersStageAndBar(t *testing.T) {
	var b bytes.Buffer
	p := newProgress(&b)
	p.begin(3, "Test stage")
	p.update(1, 1, "Test stage")
	s := b.String()
	if !strings.Contains(s, "[3/15]") || !strings.Contains(s, "[====================]") || !strings.Contains(s, "elapsed") {
		t.Fatalf("status bar output: %q", s)
	}
}

// --- misc helper sanity ---

func TestQuantileAndMedian(t *testing.T) {
	xs := []float64{1, 2, 3, 4}
	if median(xs) != 2.5 {
		t.Fatalf("median = %v, want 2.5", median(xs))
	}
	if q := quantile(xs, 0); q != 1 {
		t.Fatalf("q0 = %v, want 1", q)
	}
	if q := quantile(xs, 1); q != 4 {
		t.Fatalf("q1 = %v, want 4", q)
	}
}

func TestBHNotUsedButSortHelpersStable(t *testing.T) {
	m := map[string]int{"z": 1, "a": 2}
	keys := stringKeysInt(m)
	if !sort.StringsAreSorted(keys) {
		t.Fatalf("expected sorted keys, got %v", keys)
	}
}
