package notation

import (
	"crypto/sha256"
	"encoding/binary"
	"fmt"
)

// SeedFor derives a stable, non-negative int64 PRNG seed from the frozen
// seed schedule: base_seed, corpus_id, representation_id, family_group,
// checkpoint, replicate_index. No runtime-random seed is ever used anywhere
// in the preparation pipeline (B03 section 9).
//
// family_group is a documented reduction of "metric_family" to the two
// boundary-preserving unit classes that actually exist (see
// RAREFACTION_PROTOCOL.md "Shared draw"): FamilyGroupStructural for G/T/S/D,
// which all sample the deepest source-observed hierarchy block, and
// FamilyGroupLine for L, which requires source-observed physical lines and
// never falls back to a coarser unit.
func SeedFor(baseSeed int64, corpusID, representationID, familyGroup string, checkpoint, replicate int) int64 {
	key := fmt.Sprintf("%d\x1f%s\x1f%s\x1f%s\x1f%d\x1f%d", baseSeed, corpusID, representationID, familyGroup, checkpoint, replicate)
	h := sha256.Sum256([]byte(key))
	v := binary.BigEndian.Uint64(h[:8])
	return int64(v &^ (1 << 63))
}

const (
	FamilyGroupStructural = "STRUCTURAL"
	FamilyGroupLine       = "LINE"
	FamilyGroupBootstrap  = "BOOTSTRAP"
)

// BaseSeed is the frozen base seed for the entire comparative notation study.
const BaseSeed int64 = 20260830
