package task82b

import (
	"math/rand"
	"strings"
)

// alignDeletions greedily aligns strippedAbbr (assumed to be an ordered
// subsequence of expanLower -- true by construction for SUSPENSION and
// CONTRACTION, the two classes that actually remove letters; task82b.txt
// sec.13/18) against expanLower, returning which expanLower runes were
// kept versus deleted by the real abbreviation. This is a documented
// heuristic (greedy longest-common-subsequence-style two-pointer scan),
// not a claim of paleographic ground truth; it only drives the null
// models' deletion-count and deletion-position statistics.
func alignDeletions(strippedAbbr, expanLower string) (kept []bool) {
	a := []rune(strippedAbbr)
	e := []rune(expanLower)
	kept = make([]bool, len(e))
	ai := 0
	for ei := 0; ei < len(e); ei++ {
		if ai < len(a) && e[ei] == a[ai] {
			kept[ei] = true
			ai++
		}
	}
	return kept
}

// CharDeletionStats is the empirical, corpus-wide profile of which
// characters (and which relative within-word position) the real
// abbreviation actually deleted, estimated once from all real pairs
// (task82b.txt sec.19/20).
type CharDeletionStats struct {
	DeleteFreq map[rune]int
	TotalFreq  map[rune]int
	PosBuckets [5]int // relative position of deleted chars, 5 equal-width bins over [0,1)
	PosTotal   int
}

// BuildCharDeletionStats scans every real pair once.
func BuildCharDeletionStats(pairs []PairUnit) CharDeletionStats {
	st := CharDeletionStats{DeleteFreq: map[rune]int{}, TotalFreq: map[rune]int{}}
	for _, p := range pairs {
		stripped := strings.ToLower(stripMarks(p.AbbrText))
		expan := strings.ToLower(strings.TrimSpace(p.ExpanText))
		if expan == "" {
			continue
		}
		kept := alignDeletions(stripped, expan)
		e := []rune(expan)
		for i, r := range e {
			st.TotalFreq[r]++
			if !kept[i] {
				st.DeleteFreq[r]++
				bucket := i * 5 / len(e)
				if bucket > 4 {
					bucket = 4
				}
				st.PosBuckets[bucket]++
				st.PosTotal++
			}
		}
	}
	return st
}

// deletionRate returns the empirical per-character deletion probability,
// falling back to the corpus-wide average when a character was never
// observed in the real paired data (matched-frequency null needs a
// weight for every character it might encounter).
func (st CharDeletionStats) deletionRate(r rune) float64 {
	total := st.TotalFreq[r]
	if total == 0 {
		return st.averageRate()
	}
	return float64(st.DeleteFreq[r]) / float64(total)
}

func (st CharDeletionStats) averageRate() float64 {
	var d, t int
	for r := range st.TotalFreq {
		d += st.DeleteFreq[r]
		t += st.TotalFreq[r]
	}
	if t == 0 {
		return 0
	}
	return float64(d) / float64(t)
}

// NullWord produces one null "abbreviated-like" surface form for a real
// pair's expansion, deleting exactly deletionCount runes (matching the
// real abbreviation's own retained-letter count, task82b.txt sec.18's
// "same output length / deletion rate") under one of three rules:
//
//	RANDOM_DELETION_MATCHED:    uniformly random positions (sec.18)
//	FREQUENCY_MATCHED_DELETION: positions weighted by BuildCharDeletionStats'
//	                            empirical per-character deletion rate (sec.19)
//	POSITION_MATCHED:           positions drawn from the empirical relative-
//	                            position histogram (sec.20)
func NullWord(kind, expanLower string, deletionCount int, st CharDeletionStats, r *rand.Rand) string {
	e := []rune(expanLower)
	if deletionCount <= 0 || len(e) == 0 {
		return expanLower
	}
	deletionCount = min(deletionCount, len(e))
	var weight func(i int) float64
	switch kind {
	case "RANDOM_DELETION_MATCHED":
		weight = func(i int) float64 { return 1 }
	case "FREQUENCY_MATCHED_DELETION":
		weight = func(i int) float64 { return st.deletionRate(e[i]) + 1e-6 }
	case "POSITION_MATCHED":
		weight = func(i int) float64 {
			bucket := min(i*5/len(e), 4)
			if st.PosTotal == 0 {
				return 1
			}
			return float64(st.PosBuckets[bucket])/float64(st.PosTotal) + 1e-6
		}
	default:
		weight = func(i int) float64 { return 1 }
	}
	deleted := weightedSampleWithoutReplacement(r, len(e), deletionCount, weight)
	del := make([]bool, len(e))
	for _, i := range deleted {
		del[i] = true
	}
	var b strings.Builder
	for i, ch := range e {
		if !del[i] {
			b.WriteRune(ch)
		}
	}
	return b.String()
}

// weightedSampleWithoutReplacement draws k distinct indices from [0,n)
// without replacement, probability proportional to weight(i), by
// repeated weighted draws over a shrinking pool (simple, adequate at the
// word lengths this package ever sees, a few dozen runes at most).
func weightedSampleWithoutReplacement(r *rand.Rand, n, k int, weight func(int) float64) []int {
	pool := make([]int, n)
	for i := range pool {
		pool[i] = i
	}
	var chosen []int
	for len(chosen) < k && len(pool) > 0 {
		total := 0.0
		w := make([]float64, len(pool))
		for i, idx := range pool {
			w[i] = weight(idx)
			total += w[i]
		}
		x := r.Float64() * total
		pick := len(pool) - 1
		acc := 0.0
		for i, wi := range w {
			acc += wi
			if x <= acc {
				pick = i
				break
			}
		}
		chosen = append(chosen, pool[pick])
		pool = append(pool[:pick], pool[pick+1:]...)
	}
	return chosen
}

// DeletionCount returns how many literal expanded letters the real
// abbreviation removed for one pair, from the same alignment heuristic
// used to build CharDeletionStats.
func DeletionCount(p PairUnit) int {
	stripped := strings.ToLower(stripMarks(p.AbbrText))
	expan := strings.ToLower(strings.TrimSpace(p.ExpanText))
	if expan == "" {
		return 0
	}
	kept := alignDeletions(stripped, expan)
	n := 0
	for _, k := range kept {
		if !k {
			n++
		}
	}
	return n
}
