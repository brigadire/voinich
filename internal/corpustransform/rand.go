package corpustransform

import (
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"math/rand/v2"
)

// subRand derives an independent, fully deterministic pseudo-random stream
// from (purpose, seed, extra). It never reads current time, process ID, or
// crypto/rand: SHA256 here is used purely as a fixed, pure hash function to
// spread a user seed into two PCG state words, not as an entropy source.
// Distinct purposes/extras yield independent-looking streams so that, e.g.,
// the keyed-transposition column permutation and the homophonic occurrence
// draws never share state even when derived from the same -seed.
func subRand(purpose string, seed int64, extra uint64) *rand.Rand {
	digest := sha256.Sum256(fmt.Appendf(nil, "corpus-transform/v1/%s/seed=%d/extra=%d", purpose, seed, extra))
	s1 := binary.LittleEndian.Uint64(digest[0:8])
	s2 := binary.LittleEndian.Uint64(digest[8:16])
	return rand.New(rand.NewPCG(s1, s2))
}
