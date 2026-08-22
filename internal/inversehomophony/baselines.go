package inversehomophony

import (
	"math"
	"math/rand/v2"
	"sort"
)

// NoCollapsePartition is the NO_COLLAPSE baseline (task57 section 14A):
// every cipher type is its own class.
func NoCollapsePartition(freq map[string]int) Partition {
	p := make(Partition, len(freq))
	for t := range freq {
		p[t] = t
	}
	return p
}

// FrequencyOnlyPartition is the FREQUENCY-ONLY baseline (section 14B): a
// deterministic class from frequency alone, no context features at all -
// class = floor(log2(freq+1)), so it tests whether the recovery method's
// context features add anything beyond raw frequency grouping.
func FrequencyOnlyPartition(freq map[string]int) Partition {
	p := make(Partition, len(freq))
	for t, f := range freq {
		bucket := int(math.Floor(math.Log2(float64(f) + 1)))
		p[t] = "freq_bucket_" + itoa2(bucket)
	}
	return p
}

func itoa2(i int) string {
	if i == 0 {
		return "0"
	}
	neg := i < 0
	if neg {
		i = -i
	}
	var buf [20]byte
	pos := len(buf)
	for i > 0 {
		pos--
		buf[pos] = byte('0' + i%10)
		i /= 10
	}
	if neg {
		pos--
		buf[pos] = '-'
	}
	return string(buf[pos:])
}

// RandomPartition is the RANDOM_PARTITION baseline (section 14C): a
// uniform-random assignment of tokens into classes reproducing exactly the
// occurrence-weighted class-size multiset of matchSizes (typically the
// recovered partition's sizes), seeded deterministically. Because classes
// are matched on occurrence-weighted size (not token count), tokens are
// shuffled and greedily packed into class "slots" until each slot's
// accumulated frequency reaches its target size.
func RandomPartition(freq map[string]int, matchSizes []int, seed int64) Partition {
	tokens := make([]string, 0, len(freq))
	for t := range freq {
		tokens = append(tokens, t)
	}
	sort.Strings(tokens)
	r := rand.New(rand.NewPCG(uint64(seed), 0x5A4D))
	r.Shuffle(len(tokens), func(i, j int) { tokens[i], tokens[j] = tokens[j], tokens[i] })

	targets := append([]int{}, matchSizes...)
	sort.Sort(sort.Reverse(sort.IntSlice(targets)))
	if len(targets) == 0 {
		targets = []int{0}
	}

	p := make(Partition, len(tokens))
	slot := 0
	accum := 0
	for _, t := range tokens {
		for slot < len(targets)-1 && accum >= targets[slot] {
			slot++
			accum = 0
		}
		p[t] = "rand_class_" + itoa2(slot)
		accum += freq[t]
	}
	return p
}
