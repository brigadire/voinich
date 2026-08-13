package clustermetadataglobal

import "math/rand"

// labelBlock is one contiguous run of identical token-level metadata labels,
// including runs of "" (unknown/uncertain metadata).
type labelBlock struct {
	value  string
	length int
}

func blocksOf(labels []string) []labelBlock {
	var bs []labelBlock
	for _, v := range labels {
		if n := len(bs); n > 0 && bs[n-1].value == v {
			bs[n-1].length++
		} else {
			bs = append(bs, labelBlock{v, 1})
		}
	}
	return bs
}

// permuteKnownBlocks reassigns labels among contiguous known-label blocks
// only, preserving every block's position and length. Unknown ("") blocks are
// left untouched in place. The known/unknown token mask is therefore
// identical between the real corpus and every permutation, which guarantees
// an identical valid-window denominator for observed and null statistics
// (task requirement: never change the denominator between observed and
// null). One realization produced here is reused, unchanged, across the
// entire frozen window_size x method x K search space for that replicate.
func permuteKnownBlocks(bs []labelBlock, rng *rand.Rand) []string {
	vals := make([]string, 0, len(bs))
	for _, b := range bs {
		if b.value != "" {
			vals = append(vals, b.value)
		}
	}
	rng.Shuffle(len(vals), func(i, j int) { vals[i], vals[j] = vals[j], vals[i] })
	total := 0
	for _, b := range bs {
		total += b.length
	}
	out := make([]string, 0, total)
	vi := 0
	for _, b := range bs {
		v := b.value
		if v != "" {
			v = vals[vi]
			vi++
		}
		for j := 0; j < b.length; j++ {
			out = append(out, v)
		}
	}
	return out
}
