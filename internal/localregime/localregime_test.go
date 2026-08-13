package localregime

import (
	"bytes"
	"math"
	"reflect"
	"strings"
	"testing"
	"time"
)

func tiny() corpus {
	x := []string{"a", "b", "c", "d", "e", "f", "g"}
	return corpus{Tokens: x, Lines: [][]string{x}, Counts: map[string]int{"a": 1, "b": 1, "c": 1, "d": 1, "e": 1, "f": 1, "g": 1}, LineAt: []int{0, 0, 0, 0, 0, 0, 0}}
}
func TestLocalWindowExtractionAndGap(t *testing.T) {
	p := localProfile(tiny(), 3, 3, 1, "symmetric", false)
	if p["a"] != .25 || p["b"] != .25 || p["f"] != .25 || p["g"] != .25 || p["d"] != 0 {
		t.Fatalf("bad profile %#v", p)
	}
}
func TestProfileSides(t *testing.T) {
	c := tiny()
	l := localProfile(c, 3, 3, 1, "left", false)
	r := localProfile(c, 3, 3, 1, "right", false)
	if l["b"] == 0 || l["f"] != 0 || r["f"] == 0 || r["b"] != 0 {
		t.Fatalf("left=%#v right=%#v", l, r)
	}
}
func TestJSSimilarity(t *testing.T) {
	if jsSimilarity(profile{"a": 1}, profile{"a": 1}) != 1 || jsSimilarity(profile{"a": 1}, profile{"b": 1}) != 0 {
		t.Fatal("JS endpoints wrong")
	}
}
func TestSlidingWindows(t *testing.T) {
	p, r := slidingProfiles(tiny(), 3, 2)
	if len(p) != 3 || r[1].Start != 2 || r[2].End != 7 {
		t.Fatalf("%d %#v", len(p), r)
	}
}
func TestLocalBlockShufflePreservesBlocksAndIsDeterministic(t *testing.T) {
	x := []string{"a", "b", "c", "d", "e", "f"}
	a := blockShuffle(x, 3, 7)
	b := blockShuffle(x, 3, 7)
	if !reflect.DeepEqual(a, b) {
		t.Fatal("not deterministic")
	}
	for _, v := range a[:3] {
		if !strings.Contains("abc", v) {
			t.Fatal("crossed block")
		}
	}
}
func TestFrequencyMatchedOccurrenceControls(t *testing.T) {
	c := tiny()
	cfg := defaults(Config{RegimeRadius: 2, RegimeGap: 0, RegimeControlsK: 2, MaxDistance: 2})
	pool := buildControlPool(c, cfg)
	x := matchedExpected(c, "d", []int{3}, []profile{localProfile(c, 3, 2, 0, "symmetric", false)}, pool, cfg)
	if len(x) != 2 {
		t.Fatalf("missing distances %#v", x)
	}
}
func TestResidualDependency(t *testing.T) {
	obs, exp := .9, .7
	if math.Abs(residualDependency(obs, exp)-.2) > 1e-12 {
		t.Fatal("residual formula")
	}
}
func TestCorpusShufflesDeterministic(t *testing.T) {
	c := tiny()
	a := shuffleCorpus(c, "global", 0, 42)
	b := shuffleCorpus(c, "global", 0, 42)
	if !reflect.DeepEqual(a.Tokens, b.Tokens) {
		t.Fatal("global shuffle not deterministic")
	}
}
func TestRetainedEffect(t *testing.T) {
	if math.Abs(retainedEffect(.9, .7, .5)-.5) > 1e-12 {
		t.Fatal("retained effect")
	}
}
func TestChangePoints(t *testing.T) {
	r := []WindowRow{{Size: 10}, {Size: 10, Start: 10, AdjacentJSDistance: .01}, {Size: 10, Start: 20, AdjacentJSDistance: .01}, {Size: 10, Start: 30, AdjacentJSDistance: 1}}
	if len(changePoints(r)) != 1 {
		t.Fatal("expected change point")
	}
}
func TestProgressStatusBar(t *testing.T) {
	var b bytes.Buffer
	now := time.Unix(0, 0)
	p := newProgress(&b)
	p.clock = func() time.Time { return now }
	p.begin(2, "Profiles")
	now = now.Add(5 * time.Second)
	p.update(1, 2, "Profiles")
	s := b.String()
	for _, v := range []string{"[2/7]", "50%", "elapsed 00:05", "ETA 00:05"} {
		if !strings.Contains(s, v) {
			t.Fatalf("%q lacks %q", s, v)
		}
	}
}
