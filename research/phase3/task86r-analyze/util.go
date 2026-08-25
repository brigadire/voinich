package main

import (
	"math"
	"sort"
)

func ceilLog2(n float64) float64 {
	if n <= 1 {
		return 0
	}
	return math.Ceil(math.Log2(n))
}

func log2(x float64) float64 {
	if x <= 0 {
		return math.Inf(-1)
	}
	return math.Log2(x)
}

// sortedIntKeys returns sorted keys of an int-count map, for deterministic
// accumulation order.
func sortedIntKeys(m map[string]int) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

func sumIntCounts(m map[string]int) int {
	total := 0
	for _, k := range sortedIntKeys(m) {
		total += m[k]
	}
	return total
}
