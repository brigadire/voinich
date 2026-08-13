package localregime

import (
	"bufio"
	"fmt"
	"gopkg.in/yaml.v3"
	"os"
	"strings"
)

func readCorpus(path string) (corpus, error) {
	f, e := os.Open(path)
	if e != nil {
		return corpus{}, e
	}
	defer f.Close()
	c := corpus{Counts: map[string]int{}}
	s := bufio.NewScanner(f)
	s.Buffer(make([]byte, 64<<10), 4<<20)
	for s.Scan() {
		x := strings.Fields(s.Text())
		if len(x) == 0 {
			continue
		}
		li := len(c.Lines)
		c.Lines = append(c.Lines, x)
		for _, t := range x {
			c.Tokens = append(c.Tokens, t)
			c.LineAt = append(c.LineAt, li)
			c.Counts[t]++
		}
	}
	if e = s.Err(); e != nil {
		return c, e
	}
	if len(c.Tokens) == 0 {
		return c, fmt.Errorf("corpus is empty")
	}
	return c, nil
}
func readPairs(path string, limit int) ([]pair, error) {
	b, e := os.ReadFile(path)
	if e != nil {
		return nil, e
	}
	var x struct {
		Pairs []struct {
			A string `yaml:"token_a"`
			B string `yaml:"token_b"`
		} `yaml:"pairs"`
	}
	if e = yaml.Unmarshal(b, &x); e != nil {
		return nil, e
	}
	out := make([]pair, 0)
	for _, q := range x.Pairs {
		out = append(out, pair{q.A, q.B})
		if limit > 0 && len(out) >= limit {
			break
		}
	}
	return out, nil
}

func readControlPairs(path string) ([]controlRow, error) {
	f, e := os.Open(path)
	if e != nil {
		return nil, e
	}
	defer f.Close()
	s := bufio.NewScanner(f)
	if !s.Scan() {
		return nil, fmt.Errorf("empty controls TSV")
	}
	h := strings.Split(s.Text(), "\t")
	col := map[string]int{}
	for i, v := range h {
		col[v] = i
	}
	for _, name := range []string{"target_a", "target_b", "control_rank", "control_a", "control_b"} {
		if _, ok := col[name]; !ok {
			return nil, fmt.Errorf("controls TSV lacks %s", name)
		}
	}
	seen := map[string]bool{}
	var out []controlRow
	for s.Scan() {
		r := strings.Split(s.Text(), "\t")
		if len(r) < len(h) {
			continue
		}
		key := r[col["target_a"]] + "\x00" + r[col["target_b"]] + "\x00" + r[col["control_rank"]]
		if seen[key] {
			continue
		}
		seen[key] = true
		rank := 0
		fmt.Sscan(r[col["control_rank"]], &rank)
		out = append(out, controlRow{Target: pair{r[col["target_a"]], r[col["target_b"]]}, Control: pair{r[col["control_a"]], r[col["control_b"]]}, Rank: rank})
	}
	return out, s.Err()
}
