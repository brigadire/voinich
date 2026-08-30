package notation

import (
	"fmt"
	"math/rand"
	"sort"
)

func RenameSymbols(rs []Record, mapping map[string]string) ([]Record, error) {
	out := cloneRecords(rs)
	for i := range out {
		for j, s := range out[i].Symbols {
			x, ok := mapping[s]
			if !ok {
				return nil, fmt.Errorf("missing rename for %q", s)
			}
			out[i].Symbols[j] = x
		}
		out[i].Token = joinSymbols(out[i].Symbols)
	}
	return out, nil
}
func Duplicate(rs []Record) []Record {
	out := cloneRecords(rs)
	base := len(out)
	for i, r := range cloneRecords(rs) {
		r.TokenID = fmt.Sprintf("%s-dup", r.TokenID)
		r.Document.Value += "-dup"
		r.TokenIndex = rs[i].TokenIndex
		out = append(out, r)
		_ = base
	}
	return out
}
func ShuffleTokenOrder(rs []Record, seed int64) []Record {
	out := cloneRecords(rs)
	rng := rand.New(rand.NewSource(seed))
	groups := groupIndices(out, func(r Record) string { return lineKey(r) })
	keys := make([]string, 0, len(groups))
	for k := range groups {
		keys = append(keys, k)
	}
	sort.Strings(keys) // Go randomizes map iteration order per process; a shared rng consumed in map order would make this non-deterministic across otherwise-identical runs.
	for _, k := range keys {
		idx := groups[k]
		rng.Shuffle(len(idx), func(i, j int) { out[idx[i]], out[idx[j]] = out[idx[j]], out[idx[i]] })
		for i, recIdx := range idx {
			out[recIdx].TokenIndex = i
		}
	}
	return out
}
func ShuffleWithinTokens(rs []Record, seed int64) []Record {
	out := cloneRecords(rs)
	rng := rand.New(rand.NewSource(seed))
	for i := range out {
		rng.Shuffle(len(out[i].Symbols), func(a, b int) { out[i].Symbols[a], out[i].Symbols[b] = out[i].Symbols[b], out[i].Symbols[a] })
		out[i].Token = joinSymbols(out[i].Symbols)
	}
	return out
}
func ShuffleLines(rs []Record, seed int64) []Record {
	return shuffleBlocks(rs, seed, func(r Record) string { return lineKey(r) })
}
func ShufflePages(rs []Record, seed int64) []Record {
	return shuffleBlocks(rs, seed, func(r Record) string { return r.Document.Value + "\x1f" + r.Page.Value })
}
func shuffleBlocks(rs []Record, seed int64, key func(Record) string) []Record {
	blocks := map[string][]Record{}
	var order []string
	seen := map[string]bool{}
	for _, r := range cloneRecords(rs) {
		k := key(r)
		if !seen[k] {
			seen[k] = true
			order = append(order, k)
		}
		blocks[k] = append(blocks[k], r)
	}
	rand.New(rand.NewSource(seed)).Shuffle(len(order), func(i, j int) { order[i], order[j] = order[j], order[i] })
	var out []Record
	for _, k := range order {
		out = append(out, blocks[k]...)
	}
	return out
}
func groupIndices(rs []Record, key func(Record) string) map[string][]int {
	m := map[string][]int{}
	for i, r := range rs {
		m[key(r)] = append(m[key(r)], i)
	}
	return m
}
func cloneRecords(rs []Record) []Record {
	out := make([]Record, len(rs))
	for i, r := range rs {
		out[i] = r
		out[i].Symbols = append([]string(nil), r.Symbols...)
		if r.Attributes != nil {
			out[i].Attributes = map[string]string{}
			for k, v := range r.Attributes {
				out[i].Attributes[k] = v
			}
		}
	}
	return out
}
func joinSymbols(s []string) string {
	out := ""
	for _, x := range s {
		out += x
	}
	return out
}
