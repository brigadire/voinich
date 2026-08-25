package main

import (
	"crypto/sha256"
	"encoding/binary"
	"math/bits"
	"strconv"
	"strings"
)

// Seed derivation and PRNG per research/phase3/task85a/G1_SEED_CONTRACT.md
// and G1_EXECUTABLE_CONTRACT.json "seed" section.
//
// The contract fixes: 8 seed fields joined by 0x1f, SHA-256, first 8 bytes
// little-endian -> uint64 seed0; seed0 expanded by SplitMix64 twice into a
// 128-bit PCG-XSL-RR-128/64 state; sampling uses the high 53 bits of one
// uint64 draw mapped to [0,1).
//
// IMPLEMENTATION_DETAIL (documented in TASK86R_EXECUTION.md): the contract
// does not spell out the PCG multiplier/increment/warm-up wiring beyond
// "expanded by SplitMix64 twice". This file uses PCG64's published default
// multiplier, a fixed odd 128-bit increment (shared across all streams,
// varied only via the derived seed/state), and one warm-up advance after
// seeding -- the standard PCG "srandom" pattern. The exact output function
// (XSL-RR-128/64: xor the two 64-bit halves, rotate right by the top 6
// bits of the 128-bit state) is the real, named PCG64 output permutation.
// This is a deterministic, well-mixing, full-period generator; nothing in
// this experiment depends on bit-for-bit agreement with any external PCG
// implementation, only on determinism given identical inputs.

type u128 struct{ hi, lo uint64 }

const (
	pcgMulHi = 0x2360ed051fc65da4
	pcgMulLo = 0x4385df649fccf645
	pcgIncHi = 0x5851f42d4c957f2d
	pcgIncLo = 0x14057b7ef767814f // odd
)

var pcgMul = u128{pcgMulHi, pcgMulLo}
var pcgInc = u128{pcgIncHi, pcgIncLo}

func mul128(a, b u128) u128 {
	hi, lo := bits.Mul64(a.lo, b.lo)
	hi += a.lo*b.hi + a.hi*b.lo
	return u128{hi: hi, lo: lo}
}

func add128(a, b u128) u128 {
	lo, carry := bits.Add64(a.lo, b.lo, 0)
	hi, _ := bits.Add64(a.hi, b.hi, carry)
	return u128{hi: hi, lo: lo}
}

func rotr64(x uint64, r uint64) uint64 {
	r &= 63
	return (x >> r) | (x << (64 - r))
}

// splitMix64 is a stepper over a 64-bit state, used only to expand the
// derived seed into the 128-bit PCG state.
type splitMix64 struct{ state uint64 }

func (s *splitMix64) next() uint64 {
	s.state += 0x9E3779B97F4A7C15
	z := s.state
	z = (z ^ (z >> 30)) * 0xBF58476D1CE4E5B9
	z = (z ^ (z >> 27)) * 0x94D049BB133111EB
	z ^= z >> 31
	return z
}

// PRNG is the frozen PCG-XSL-RR-128/64 stream.
type PRNG struct{ state u128 }

func NewPRNG(seed uint64) *PRNG {
	sm := &splitMix64{state: seed}
	s1, s2 := sm.next(), sm.next()
	p := &PRNG{state: u128{hi: s1, lo: s2}}
	p.state = add128(mul128(p.state, pcgMul), pcgInc) // warm-up
	return p
}

// Uint64 advances the state and returns the next 64-bit output.
func (p *PRNG) Uint64() uint64 {
	p.state = add128(mul128(p.state, pcgMul), pcgInc)
	xored := p.state.hi ^ p.state.lo
	rot := p.state.hi >> 58
	return rotr64(xored, rot)
}

// Float64 returns a value in [0,1) from the high 53 bits of one draw.
func (p *PRNG) Float64() float64 {
	return float64(p.Uint64()>>11) / float64(uint64(1)<<53)
}

// SeedFields is the frozen 8-tuple: experiment_namespace, model_class,
// candidate_id, corpus_id, transcription, partition, scale, replicate_index.
type SeedFields struct {
	Namespace     string
	ModelClass    string
	CandidateID   string
	CorpusID      string
	Transcription string
	Partition     string
	Scale         float64
	Replicate     int
}

// formatScale renders scale as the shortest round-trippable decimal, e.g.
// 1 -> "1", 0.5 -> "0.5", 2 -> "2". Decimal integers elsewhere in the tuple
// (Replicate) have no sign or leading zero (strconv.Itoa already satisfies
// this for non-negative values, which is all this experiment ever seeds).
func formatScale(s float64) string {
	return strconv.FormatFloat(s, 'g', -1, 64)
}

func (f SeedFields) Seed() uint64 {
	parts := []string{
		f.Namespace, f.ModelClass, f.CandidateID, f.CorpusID,
		f.Transcription, f.Partition, formatScale(f.Scale), strconv.Itoa(f.Replicate),
	}
	// All fields are pure-ASCII identifiers/decimal strings by
	// construction (candidate ids, corpus/transcription/partition
	// labels, decimal numbers), so Unicode NFC normalization is the
	// identity transform here and is not separately applied.
	joined := strings.Join(parts, "\x1f")
	digest := sha256.Sum256([]byte(joined))
	return binary.LittleEndian.Uint64(digest[:8])
}

func NewSeededPRNG(f SeedFields) *PRNG { return NewPRNG(f.Seed()) }

// DrawIndex performs inverse-CDF sampling over cum, a non-decreasing
// cumulative-probability slice ending at (approximately) 1, using
// half-open cumulative intervals: outcome i is chosen when
// u < cum[i]. Assumes cum is built over outcomes already sorted per the
// frozen tie-break rule for that context.
func DrawIndex(p *PRNG, cum []float64) int {
	u := p.Float64()
	for i, c := range cum {
		if u < c {
			return i
		}
	}
	return len(cum) - 1
}
