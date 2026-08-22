package inversehomophony

import (
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"strconv"
	"strings"
	"testing"

	"zcore.dev/voinich/internal/corpustransform"
)

// writeLines joins tokens into lines of lineLen tokens each, one line per
// physical line - the same "whitespace + newline" layout every real
// corpus in this repo uses.
func writeLines(t *testing.T, path string, tokens []string, lineLen int) {
	t.Helper()
	var b strings.Builder
	for i := 0; i < len(tokens); i += lineLen {
		end := min(i+lineLen, len(tokens))
		b.WriteString(strings.Join(tokens[i:end], " "))
		b.WriteByte('\n')
	}
	if err := os.WriteFile(path, []byte(b.String()), 0o644); err != nil {
		t.Fatal(err)
	}
}

// syntheticCorpus builds a tiny in-memory homophonic ciphertext (task46
// mechanism) from plaintext, writes ciphertext + mapping.tsv to a temp
// dir, and returns everything a test needs.
type syntheticCorpus struct {
	cipherPath, mappingPath string
	cipherTokens            []string
	plainTokens             []string
	mapping                 corpustransform.Mapping
}

func buildSyntheticCorpus(t *testing.T, plainTokens []string, h int, selection string, seed int64) syntheticCorpus {
	t.Helper()
	mapping, err := corpustransform.BuildMapping(plainTokens, corpustransform.HomophonicParams{
		Model: corpustransform.HomophoneModelFixed, Homophones: h, Selection: selection, Seed: seed,
	})
	if err != nil {
		t.Fatal(err)
	}
	cipherTokens := corpustransform.Encode(plainTokens, mapping, seed)

	dir := t.TempDir()
	cipherPath := filepath.Join(dir, "cipher.txt")
	mappingPath := cipherPath + ".mapping.tsv"
	writeLines(t, cipherPath, cipherTokens, 7)
	if err := os.WriteFile(mappingPath, corpustransform.MarshalMappingTSV(mapping), 0o644); err != nil {
		t.Fatal(err)
	}
	return syntheticCorpus{cipherPath: cipherPath, mappingPath: mappingPath, cipherTokens: cipherTokens, plainTokens: plainTokens, mapping: mapping}
}

func repeatWords(words []string, times int) []string {
	out := make([]string, 0, len(words)*times)
	for i := 0; i < times; i++ {
		out = append(out, words...)
	}
	return out
}

var testPlaintext = repeatWords([]string{
	"the", "quick", "brown", "fox", "jumps", "over", "the", "lazy", "dog",
	"the", "cat", "sat", "on", "the", "warm", "mat", "and", "slept", "all", "day",
}, 15)

// 1. Opaque relabeling actually hides plaintext identity.
func TestRelabelHidesPlaintextIdentity(t *testing.T) {
	tokens := []string{"cat", "dog", "cat", "bird", "dog", "cat"}
	r := Relabel(tokens)

	for i, orig := range tokens {
		if r.Tokens[i] == orig {
			t.Fatalf("relabeled token at %d still equals original plaintext token %q", i, orig)
		}
		if !strings.HasPrefix(r.Tokens[i], "x") {
			t.Fatalf("relabeled token %q does not follow the opaque x%%06d scheme", r.Tokens[i])
		}
	}
	if len(r.ToOpaque) != 3 || len(r.ToOriginal) != 3 {
		t.Fatalf("expected 3 distinct tokens, got ToOpaque=%d ToOriginal=%d", len(r.ToOpaque), len(r.ToOriginal))
	}
	for i, opaque := range r.Tokens {
		if r.ToOriginal[opaque] != tokens[i] {
			t.Fatalf("round-trip broken at %d: opaque %q -> %q, want %q", i, opaque, r.ToOriginal[opaque], tokens[i])
		}
	}
}

// 2. The recovery path (BuildFeatures/CandidatePairs/Recover) reads
// nothing but the relabeled token stream: renaming every original token
// via a rank-preserving bijection must not change the recovered result at
// all, proving no original-identity or graphemic signal leaks through.
func TestRecoveryIgnoresOriginalTokenIdentity(t *testing.T) {
	sc := buildSyntheticCorpus(t, testPlaintext, 3, corpustransform.SelectionUniform, 7)
	loadedA, err := LoadCorpus(sc.cipherPath)
	if err != nil {
		t.Fatal(err)
	}

	// Rename every distinct original cipher token via a rank-preserving
	// bijection (same sorted order, completely different strings) and
	// rebuild the corpus file.
	distinct := map[string]struct{}{}
	for _, tok := range sc.cipherTokens {
		distinct[tok] = struct{}{}
	}
	sorted := make([]string, 0, len(distinct))
	for tok := range distinct {
		sorted = append(sorted, tok)
	}
	sortStrings(sorted)
	rename := make(map[string]string, len(sorted))
	for i, tok := range sorted {
		rename[tok] = "zzrenamed" + strconv.Itoa(i+100000)
	}
	renamedTokens := make([]string, len(sc.cipherTokens))
	for i, tok := range sc.cipherTokens {
		renamedTokens[i] = rename[tok]
	}
	renamedPath := filepath.Join(t.TempDir(), "renamed.txt")
	writeLines(t, renamedPath, renamedTokens, 7)
	loadedB, err := LoadCorpus(renamedPath)
	if err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(loadedA.Relabel.Tokens, loadedB.Relabel.Tokens) {
		t.Fatalf("relabeled token streams differ after a rank-preserving rename of the original tokens")
	}

	cfg := FrozenConfig()
	cfg.Threshold = 0.2
	freqA := tokenFreq(loadedA.Relabel.Tokens)
	freqB := tokenFreq(loadedB.Relabel.Tokens)
	featA := BuildFeatures(loadedA.Relabel.Tokens, loadedA.LineOfToken, cfg)
	featB := BuildFeatures(loadedB.Relabel.Tokens, loadedB.LineOfToken, cfg)
	pairsA := CandidatePairs(featA, cfg)
	pairsB := CandidatePairs(featB, cfg)
	partA, eventsA := Recover(freqA, pairsA, cfg)
	partB, eventsB := Recover(freqB, pairsB, cfg)

	if !reflect.DeepEqual(partA, partB) {
		t.Fatalf("recovered partition depends on original token identity, not just relabeled structure")
	}
	if len(eventsA) != len(eventsB) {
		t.Fatalf("merge event count differs: %d vs %d", len(eventsA), len(eventsB))
	}
}

func tokenFreq(tokens []string) map[string]int {
	f := make(map[string]int)
	for _, t := range tokens {
		f[t]++
	}
	return f
}

// 3 & 4. Same input + same config -> byte-identical output, regardless of
// GOMAXPROCS (this package has no custom executor/goroutine fan-out, so
// varying GOMAXPROCS is the applicable "executor independence" check).
func TestRecoverDeterministicAcrossGOMAXPROCS(t *testing.T) {
	sc := buildSyntheticCorpus(t, testPlaintext, 4, corpustransform.SelectionWeighted, 3)
	loaded, err := LoadCorpus(sc.cipherPath)
	if err != nil {
		t.Fatal(err)
	}
	cfg := FrozenConfig()
	cfg.Threshold = 0.2
	freq := tokenFreq(loaded.Relabel.Tokens)

	run := func() (Partition, []MergeEvent) {
		features := BuildFeatures(loaded.Relabel.Tokens, loaded.LineOfToken, cfg)
		pairs := CandidatePairs(features, cfg)
		return Recover(freq, pairs, cfg)
	}

	prevProcs := runtime.GOMAXPROCS(0)
	defer runtime.GOMAXPROCS(prevProcs)

	runtime.GOMAXPROCS(1)
	part1, events1 := run()
	runtime.GOMAXPROCS(4)
	part4, events4 := run()

	if !reflect.DeepEqual(part1, part4) {
		t.Fatalf("recovered partition is not deterministic across GOMAXPROCS")
	}
	if !reflect.DeepEqual(events1, events4) {
		t.Fatalf("merge audit trail is not deterministic across GOMAXPROCS")
	}
}

// 5. All-tokens-to-one-class must be forbidden even when every candidate
// pair scores far above threshold.
func TestAntiTrivialCollapseForbidsAllToOne(t *testing.T) {
	freq := map[string]int{}
	tokens := []string{}
	for i := 0; i < 12; i++ {
		tok := "x" + strconv.Itoa(i)
		tokens = append(tokens, tok)
		freq[tok] = 10
	}
	var pairs []PairScore
	for i := 0; i < len(tokens); i++ {
		for j := i + 1; j < len(tokens); j++ {
			pairs = append(pairs, PairScore{A: tokens[i], B: tokens[j], Score: 1.0, Support: 999})
		}
	}
	sortPairScores(pairs)

	cfg := FrozenConfig()
	cfg.Threshold = 0.1
	partition, _ := Recover(freq, pairs, cfg)

	classes := map[string]int{}
	for _, c := range partition {
		classes[c]++
	}
	if len(classes) <= 1 {
		t.Fatalf("all tokens collapsed into %d class(es) despite MaxClassFraction=%.2f", len(classes), cfg.MaxClassFraction)
	}
	total := 0
	for _, sz := range classes {
		total += sz
	}
	for c, sz := range classes {
		frac := float64(sz) / float64(total)
		if frac > cfg.MaxClassFraction+1e-9 {
			t.Fatalf("class %q has occurrence fraction %.4f > MaxClassFraction %.4f", c, frac, cfg.MaxClassFraction)
		}
	}
}

// 6. Class recovery metrics must be invariant to permuting either
// partition's class-ID labels (task57 section 10: never compare class IDs
// directly).
func TestClassRecoveryMetricsInvariantToLabelPermutation(t *testing.T) {
	predicted := Partition{"a": "P1", "b": "P1", "c": "P2", "d": "P2", "e": "P3"}
	oracle := Partition{"a": "O1", "b": "O1", "c": "O1", "d": "O2", "e": "O2"}
	base := EvaluateClassRecovery(predicted, oracle)

	renamedPredicted := Partition{}
	for k, v := range predicted {
		renamedPredicted[k] = "PRED_" + v + "_relabeled"
	}
	renamedOracle := Partition{}
	for k, v := range oracle {
		renamedOracle[k] = "ORACLE_" + v + "_relabeled"
	}
	renamed := EvaluateClassRecovery(renamedPredicted, renamedOracle)

	if base != renamed {
		t.Fatalf("class recovery metrics changed under label permutation: %+v vs %+v", base, renamed)
	}
}

// 7. Collapsing ciphertext through the true oracle mapping must exactly
// reproduce the original plaintext class sequence (plaintext tokens, in
// order) - the strongest possible sanity check on OraclePartitionForRelabel
// and Collapse.
func TestOracleCollapseRecoversPlaintextSequence(t *testing.T) {
	sc := buildSyntheticCorpus(t, testPlaintext, 5, corpustransform.SelectionUniform, 11)
	loaded, err := LoadCorpus(sc.cipherPath)
	if err != nil {
		t.Fatal(err)
	}
	oracleMapping, err := LoadOracleMapping(sc.mappingPath)
	if err != nil {
		t.Fatal(err)
	}
	oracle := oracleMapping.OraclePartitionForRelabel(loaded.Relabel)
	collapsed := Collapse(loaded.Relabel.Tokens, oracle)

	if len(collapsed) != len(sc.plainTokens) {
		t.Fatalf("collapsed length %d != plaintext length %d", len(collapsed), len(sc.plainTokens))
	}
	for i := range collapsed {
		if collapsed[i] != sc.plainTokens[i] {
			t.Fatalf("position %d: oracle-collapsed token %q != original plaintext token %q", i, collapsed[i], sc.plainTokens[i])
		}
	}
}

// 8. RANDOM_PARTITION is reproducible given the same seed.
func TestRandomPartitionReproducible(t *testing.T) {
	freq := map[string]int{"x1": 5, "x2": 5, "x3": 5, "x4": 5, "x5": 5, "x6": 5}
	sizes := []int{15, 15}
	p1 := RandomPartition(freq, sizes, 42)
	p2 := RandomPartition(freq, sizes, 42)
	if !reflect.DeepEqual(p1, p2) {
		t.Fatalf("RandomPartition not reproducible for the same seed")
	}
}

// 9. Threshold fitting depends only on the DEVELOPMENT specs actually
// passed in, never on any wider/validation corpus set.
func TestFitThresholdFromDevelopmentOnlyUsesGivenSpecs(t *testing.T) {
	scA := buildSyntheticCorpus(t, testPlaintext, 3, corpustransform.SelectionUniform, 1)
	scB := buildSyntheticCorpus(t, testPlaintext, 6, corpustransform.SelectionWeighted, 2)
	specA := SyntheticCorpusSpec{Label: "a", CipherPath: scA.cipherPath, MappingPath: scA.mappingPath, Genre: "test"}
	specB := SyntheticCorpusSpec{Label: "b", CipherPath: scB.cipherPath, MappingPath: scB.mappingPath, Genre: "test"}

	base := FrozenConfig()
	cfgA, diagA, err := FitThresholdFromDevelopment([]SyntheticCorpusSpec{specA}, base)
	if err != nil {
		t.Fatal(err)
	}
	cfgAB, diagAB, err := FitThresholdFromDevelopment([]SyntheticCorpusSpec{specA, specB}, base)
	if err != nil {
		t.Fatal(err)
	}
	if diagAB.TruePairs <= diagA.TruePairs {
		t.Fatalf("pooling a second development corpus did not add true pairs: %d vs %d", diagA.TruePairs, diagAB.TruePairs)
	}
	// Re-fitting from just [specA] again must reproduce cfgA exactly -
	// i.e. nothing about the two-corpus call leaked backward.
	cfgA2, _, err := FitThresholdFromDevelopment([]SyntheticCorpusSpec{specA}, base)
	if err != nil {
		t.Fatal(err)
	}
	if cfgA.Threshold != cfgA2.Threshold {
		t.Fatalf("threshold fit from the same single spec is not stable: %v vs %v", cfgA.Threshold, cfgA2.Threshold)
	}
	_ = cfgAB
}
