package softstructural

import (
	"bytes"
	"math"
	"os"
	"path/filepath"
	"testing"

	"zcore.dev/voinich/internal/profilestability"
	"zcore.dev/voinich/internal/structuralreliability"
)

func closeEnough(a, b float64) bool { return math.Abs(a-b) < 1e-12 }
func curves(value float64) structuralreliability.ReliabilityCurves {
	return structuralreliability.ReliabilityCurves{Position: map[int]float64{10: value}, LeftContext: map[int]float64{10: value}, RightContext: map[int]float64{10: value}}
}
func profile(count int, pos int, left, right string) profilestability.Profile {
	return profilestability.Profile{Count: count, Positions: map[int]int{pos: count}, Left: map[string]int{left: count}, Right: map[string]int{right: count}}
}

func TestPairUsesExistingSimilarityAndIsSymmetric(t *testing.T) {
	a := profile(10, 0, "x", "y")
	b := profile(20, 1, "x", "z")
	expected := profilestability.Compare(a, b)
	ab := MakePair("a", "b", 10, 20, a, b, curves(.5))
	ba := MakePair("b", "a", 20, 10, b, a, curves(.5))
	if ab.TokenA != "a" || ab.TokenB != "b" || ab.CountA != ba.CountA || ab.CountB != ba.CountB || !closeEnough(ab.RawSimilarity, ba.RawSimilarity) || !closeEnough(*ab.DiagnosticWeightedSimilarity, *ba.DiagnosticWeightedSimilarity) {
		t.Fatalf("pair must be canonical and symmetric: %#v %#v", ab, ba)
	}
	if !closeEnough(ab.RawSimilarity, expected.Similarity) || !closeEnough(ab.PositionSimilarity, expected.PositionSimilarity) || !closeEnough(ab.LeftSimilarity, expected.LeftSimilarity) || !closeEnough(ab.RightSimilarity, expected.RightSimilarity) {
		t.Fatal("pair does not exactly reuse profilestability.Compare")
	}
}
func TestReliabilityEvidenceAndDiagnostic(t *testing.T) {
	c := structuralreliability.ReliabilityCurves{Position: map[int]float64{10: .25, 20: 1}, LeftContext: map[int]float64{10: .25, 20: 1}, RightContext: map[int]float64{10: .25, 20: 1}}
	rp, rl, rr := PairReliabilities(c, 10, 20)
	if !closeEnough(rp, .5) || rp != rl || rl != rr {
		t.Fatalf("geometric mean: %v %v %v", rp, rl, rr)
	}
	if !closeEnough(EvidenceStrength(.3, .6, .9), .6) {
		t.Fatal("bad evidence strength")
	}
	v := DiagnosticWeightedMean([3]float64{1, .5, 0}, [3]float64{.2, .3, .5})
	if v == nil || !closeEnough(*v, .35) {
		t.Fatalf("bad weighted diagnostic: %v", v)
	}
	if DiagnosticWeightedMean([3]float64{1, 1, 1}, [3]float64{}) != nil {
		t.Fatal("zero reliability must produce null")
	}
}
func TestLowReliabilityDoesNotChangeRawSimilarity(t *testing.T) {
	p := profile(10, 0, "x", "y")
	low := MakePair("a", "b", 10, 10, p, p, curves(.1))
	high := MakePair("c", "d", 10, 10, profile(10, 0, "x", "y"), profile(10, 0, "x", "z"), curves(.9))
	if low.RawSimilarity != 1 || low.EvidenceStrength >= high.EvidenceStrength || high.RawSimilarity >= low.RawSimilarity {
		t.Fatalf("similarity and evidence got conflated: low=%#v high=%#v", low, high)
	}
}
func TestRankMutualFilterBucketsAndReference(t *testing.T) {
	one := 1.
	pairs := []Pair{{TokenA: "a", TokenB: "b", RawSimilarity: .9, EvidenceStrength: .2, DiagnosticWeightedSimilarity: &one}, {TokenA: "a", TokenB: "c", RawSimilarity: .8, EvidenceStrength: .9, DiagnosticWeightedSimilarity: &one}, {TokenA: "b", TokenB: "c", RawSimilarity: .7, EvidenceStrength: .9, DiagnosticWeightedSimilarity: &one}}
	ranked := rank(pairs[:2], "a", "raw", .7)
	if len(ranked) != 1 || ranked[0].Token != "c" {
		t.Fatalf("high-evidence filter/ranking: %#v", ranked)
	}
	sets := map[string]map[string]bool{"a": {"b": true}, "b": {"a": true}, "c": {}}
	m := mutual(pairs, sets)
	if len(m) != 1 || m[0].TokenA != "a" || m[0].TokenB != "b" {
		t.Fatalf("mutual: %#v", m)
	}
	rows := buckets(pairs)
	if rows[5].EvidenceLT030 != 1 || rows[4].EvidenceGE090 != 1 || rows[3].EvidenceGE090 != 1 {
		t.Fatalf("buckets: %#v", rows)
	}
	r := reference(pairs[0])
	if r.TokenA != "a" || r.RawStructuralSimilarity != .9 {
		t.Fatalf("reference: %#v", r)
	}
}
func TestTSVDeterministicAndCanonical(t *testing.T) {
	dir := t.TempDir()
	path1 := filepath.Join(dir, "a.tsv")
	path2 := filepath.Join(dir, "b.tsv")
	p := MakePair("z", "a", 10, 10, profile(10, 0, "x", "y"), profile(10, 0, "x", "y"), curves(.5))
	pairs := []Pair{p}
	if err := WriteTSV(path1, pairs); err != nil {
		t.Fatal(err)
	}
	if err := WriteTSV(path2, pairs); err != nil {
		t.Fatal(err)
	}
	a, _ := os.ReadFile(path1)
	b, _ := os.ReadFile(path2)
	if !bytes.Equal(a, b) {
		t.Fatal("TSV is not deterministic")
	}
	if !bytes.Contains(a, []byte("a\tz\t")) {
		t.Fatalf("non-canonical TSV: %s", a)
	}
}
