package main

import (
	"math/rand"
	"testing"
)

func TestBuildChannelsUsesFixedPositions(t *testing.T) {
	channels := buildChannels([]string{"ab cde fgh", "i"})
	if got := channels[0].Lines[0]; len(got) != 1 || got[0] != "a" {
		t.Fatalf("line first channel = %#v", got)
	}
	if got := channels[1].Lines[0]; len(got) != 1 || got[0] != "h" {
		t.Fatalf("line last channel = %#v", got)
	}
	if got := channels[6].Lines[0]; len(got) != 1 || got[0] != "c" {
		t.Fatalf("second-token channel = %#v", got)
	}
	if got := channels[8].Lines[0]; len(got) != 1 || got[0] != "c" {
		t.Fatalf("periodic channel = %#v", got)
	}
}

func TestAnalyzeIsSeedDeterministic(t *testing.T) {
	c := channel{ID: "test", Lines: [][]string{{"a", "b", "a"}, {"b", "a", "b"}}}
	first := analyze(c, 20, rand.New(rand.NewSource(7)))
	second := analyze(c, 20, rand.New(rand.NewSource(7)))
	if first != second {
		t.Fatalf("deterministic analysis mismatch: %#v != %#v", first, second)
	}
	if first.PValue <= 0 || first.PValue > 1 || first.NMI < 0 {
		t.Fatalf("invalid result: %#v", first)
	}
}

func TestSingletonChannelUsesLineOrder(t *testing.T) {
	_, pairs, _, nmi := statistics([][]string{{"a"}, {"b"}, {"a"}}, true)
	if pairs != 2 || nmi == 0 {
		t.Fatalf("singleton channel lost line-order pairs: pairs=%d nmi=%f", pairs, nmi)
	}
}
