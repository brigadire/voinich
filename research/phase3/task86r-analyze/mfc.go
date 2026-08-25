package main

import (
	"fmt"
	"math"
)

// mfcAlphabet is the frozen 12-letter calibration alphabet
// (G1_EXECUTABLE_CONTRACT.json calibration.alphabet).
var mfcAlphabet = []string{"a", "b", "c", "d", "e", "f", "g", "h", "i", "j", "k", "l"}

func mfcGlyph(idx int) string { return mfcAlphabet[idx%12] }

// mfcGeometric draws Geometric(p) with support {0,1,2,...} via inverse
// transform: k = floor(log(1-U)/log(1-p)).
func mfcGeometric(p *PRNG, prob float64) int {
	u := p.Float64()
	if u >= 1 {
		u = 0.9999999999
	}
	k := math.Floor(math.Log(1-u) / math.Log(1-prob))
	if k < 0 {
		k = 0
	}
	return int(k)
}

func genMFC0Token(p *PRNG) []string {
	L := mfcGeometric(p, 0.25) + 1
	if L > 12 {
		L = 12
	}
	out := make([]string, L)
	for i := 0; i < L; i++ {
		out[i] = mfcGlyph(int(p.Float64() * 12))
	}
	return out
}

func genMFC1Token(p *PRNG) []string {
	first := int(p.Float64() * 12)
	out := []string{mfcGlyph(first)}
	prev := first
	for {
		if len(out) >= 12 {
			break
		}
		if p.Float64() < 0.20 {
			break
		}
		u := p.Float64()
		var next int
		switch {
		case u < 0.55:
			next = (prev + 1) % 12
		case u < 0.75:
			next = prev
		default:
			// remaining ten letters (excluding prev and prev+1), each 0.025,
			// enumerated in ascending index order skipping those two.
			rank := int((u - 0.75) / 0.025)
			if rank > 9 {
				rank = 9
			}
			next = nthRemaining(prev, (prev+1)%12, rank)
		}
		out = append(out, mfcGlyph(next))
		prev = next
	}
	return out
}

func nthRemaining(exclude1, exclude2, rank int) int {
	count := 0
	for i := 0; i < 12; i++ {
		if i == exclude1 || i == exclude2 {
			continue
		}
		if count == rank {
			return i
		}
		count++
	}
	return exclude1
}

func genMFC2Token(p *PRNG) []string {
	state := 0
	var out []string
	for {
		if len(out) >= 12 {
			break
		}
		accepting := state == 2 || state == 4 || state == 5
		if accepting {
			if p.Float64() < 0.25 {
				break
			}
		}
		u := p.Float64()
		var target, label int
		switch {
		case u < 0.50:
			target, label = (state+1)%6, (2*state)%12
		case u < 0.80:
			target, label = (state+2)%6, (2*state+1)%12
		default:
			target, label = (2*state+1)%6, (2*state+2)%12
		}
		out = append(out, mfcGlyph(label))
		state = target
	}
	return out
}

// MFCPopulation is one of a generator's 16 independent
// DEVELOPMENT/VALIDATION/HELDOUT-analogue populations.
type MFCPopulation struct {
	Generator string // MFC0, MFC1, MFC2
	Index     int    // 1..16
	Dev, Val, Heldout []TokenOccurrence
}

func generateMFCPopulation(namespace, generator string, index int) MFCPopulation {
	gen := map[string]func(*PRNG) []string{
		"MFC0": genMFC0Token, "MFC1": genMFC1Token, "MFC2": genMFC2Token,
	}[generator]
	popID := fmt.Sprintf("P%02d", index)
	draw := func(partition string, n int) []TokenOccurrence {
		seed := SeedFields{
			Namespace: namespace, ModelClass: generator, CandidateID: "GENERATOR",
			CorpusID: generator, Transcription: popID, Partition: partition, Scale: 1.0, Replicate: 0,
		}
		prng := NewSeededPRNG(seed)
		out := make([]TokenOccurrence, n)
		for i := 0; i < n; i++ {
			g := gen(prng)
			out[i] = TokenOccurrence{Raw: joinGlyphs(g), Glyphs: g, Partition: partition}
		}
		return out
	}
	return MFCPopulation{
		Generator: generator, Index: index,
		Dev:     draw("DEVELOPMENT", 20000),
		Val:     draw("VALIDATION", 5000),
		Heldout: draw("HELDOUT", 5000),
	}
}
