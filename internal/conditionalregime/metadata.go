package conditionalregime

import (
	"bufio"
	"crypto/sha256"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"
)

// loadTokenLabels reads only the columns needed from the already frozen,
// aligned token_metadata_map.tsv (produced by metadata-validate). It never
// re-runs IVTFF parsing or alignment.
func loadTokenLabels(path string) (currier, hand []string, sha256Hex string, err error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, nil, "", err
	}
	sha256Hex = fmt.Sprintf("%x", sha256.Sum256(b))
	s := bufio.NewScanner(strings.NewReader(string(b)))
	s.Buffer(make([]byte, 64*1024), 16*1024*1024)
	if !s.Scan() {
		return nil, nil, "", fmt.Errorf("empty token metadata map: %s", path)
	}
	header := strings.Split(strings.TrimSuffix(s.Text(), "\r"), "\t")
	col := map[string]int{}
	for i, h := range header {
		col[h] = i
	}
	for _, need := range []string{"token_position", "currier", "hand"} {
		if _, ok := col[need]; !ok {
			return nil, nil, "", fmt.Errorf("token metadata map %s is missing required column %q", path, need)
		}
	}
	expect := 0
	for s.Scan() {
		row := strings.Split(strings.TrimSuffix(s.Text(), "\r"), "\t")
		pos, e := strconv.Atoi(row[col["token_position"]])
		if e != nil {
			return nil, nil, "", fmt.Errorf("token metadata map %s: non-integer token_position %q", path, row[col["token_position"]])
		}
		if pos != expect {
			return nil, nil, "", fmt.Errorf("token metadata map %s: non-contiguous token_position %d, expected %d", path, pos, expect)
		}
		expect++
		currier = append(currier, normalizeLabel(row[col["currier"]]))
		hand = append(hand, normalizeLabel(row[col["hand"]]))
	}
	if e := s.Err(); e != nil {
		return nil, nil, "", e
	}
	return currier, hand, sha256Hex, nil
}

func normalizeLabel(v string) string {
	if v == "" || v == "?" || v == "null" {
		return ""
	}
	return v
}

// buildBlocks extracts maximal contiguous runs of one class from per-token
// currier/hand labels. A token with unknown currier or hand (per the active
// scheme) never belongs to any class: known material on either side of an
// unknown gap always starts a new physical block, even if the class is
// identical (task19 sections 5-6). classOf returns ("", false) for a token
// that does not belong to any class under this scheme.
func buildBlocks(n int, classOf func(i int) (ClassID, bool)) []Block {
	var blocks []Block
	counts := map[ClassID]int{}
	var cur ClassID
	open := false
	start := 0
	flush := func(end int) {
		if !open {
			return
		}
		idx := counts[cur]
		counts[cur] = idx + 1
		blocks = append(blocks, Block{Class: cur, Index: idx, Start: start, End: end})
		open = false
	}
	for i := 0; i < n; i++ {
		cls, ok := classOf(i)
		if ok && open && cls == cur {
			continue
		}
		flush(i)
		if ok {
			cur, open, start = cls, true, i
		}
	}
	flush(n)
	return blocks
}

// buildAllBlocks builds the primary joint Currier x hand blocks and the two
// secondary single-dimension block sets, from the same per-token labels.
func buildAllBlocks(currier, hand []string) map[Scheme][]Block {
	n := len(currier)
	joint := buildBlocks(n, func(i int) (ClassID, bool) {
		if currier[i] == "" || hand[i] == "" {
			return ClassID{}, false
		}
		return ClassID{Scheme: SchemeJoint, Currier: currier[i], Hand: hand[i]}, true
	})
	currierOnly := buildBlocks(n, func(i int) (ClassID, bool) {
		if currier[i] == "" {
			return ClassID{}, false
		}
		return ClassID{Scheme: SchemeCurrierOnly, Currier: currier[i]}, true
	})
	handOnly := buildBlocks(n, func(i int) (ClassID, bool) {
		if hand[i] == "" {
			return ClassID{}, false
		}
		return ClassID{Scheme: SchemeHandOnly, Hand: hand[i]}, true
	})
	return map[Scheme][]Block{SchemeJoint: joint, SchemeCurrierOnly: currierOnly, SchemeHandOnly: handOnly}
}

func blocksByClass(blocks []Block) map[ClassID][]Block {
	out := map[ClassID][]Block{}
	for _, b := range blocks {
		out[b.Class] = append(out[b.Class], b)
	}
	return out
}

func medianLen(blocks []Block) float64 {
	if len(blocks) == 0 {
		return 0
	}
	lens := make([]int, len(blocks))
	for i, b := range blocks {
		lens[i] = b.Len()
	}
	sort.Ints(lens)
	n := len(lens)
	if n%2 == 1 {
		return float64(lens[n/2])
	}
	return float64(lens[n/2-1]+lens[n/2]) / 2
}

// classInventory builds one ClassInfo per class, in every scheme, applying
// the fixed eligibility thresholds (task19 section 8).
func classInventory(allBlocks map[Scheme][]Block, minClassTokens, minBlockTokens int) []ClassInfo {
	var out []ClassInfo
	for _, scheme := range []Scheme{SchemeJoint, SchemeCurrierOnly, SchemeHandOnly} {
		byClass := blocksByClass(allBlocks[scheme])
		classes := make([]ClassID, 0, len(byClass))
		for c := range byClass {
			classes = append(classes, c)
		}
		sort.Slice(classes, func(i, j int) bool { return classes[i].Label() < classes[j].Label() })
		for _, c := range classes {
			bs := byClass[c]
			total, largest := 0, 0
			for _, b := range bs {
				total += b.Len()
				if b.Len() > largest {
					largest = b.Len()
				}
			}
			out = append(out, ClassInfo{
				Class: c, TotalTokens: total, BlockCount: len(bs), LargestBlock: largest,
				MedianBlock: medianLen(bs),
				Eligible:    total >= minClassTokens && largest >= minBlockTokens,
			})
		}
	}
	return out
}
