package main

import (
	"encoding/csv"
	"encoding/json"
	"flag"
	"fmt"
	"math"
	"os"
	"sort"
	"strconv"
)

type schema struct {
	Descriptors []descriptor `json:"descriptors"`
}
type descriptor struct {
	ID, Type string
	Allowed  []string `json:"allowed_values"`
}
type result struct {
	ID                                string
	N                                 int
	Agreement, Kappa, Weighted, Alpha float64
	Decision, Notes                   string
}

func main() {
	schemaPath := flag.String("schema", "research/visual_descriptors/VISUAL_FEATURE_SCHEMA.json", "schema")
	aPath := flag.String("a", "", "pass A TSV")
	bPath := flag.String("b", "", "pass B TSV")
	outPath := flag.String("out", "", "output TSV")
	flag.Parse()
	if *aPath == "" || *bPath == "" || *outPath == "" {
		fmt.Fprintln(os.Stderr, "-a, -b and -out are required")
		os.Exit(2)
	}
	if err := run(*schemaPath, *aPath, *bPath, *outPath); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(schemaPath, aPath, bPath, outPath string) error {
	b, err := os.ReadFile(schemaPath)
	if err != nil {
		return err
	}
	var s schema
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}
	a, err := read(aPath)
	if err != nil {
		return err
	}
	bb, err := read(bPath)
	if err != nil {
		return err
	}
	results := make([]result, 0, len(s.Descriptors))
	for _, d := range s.Descriptors {
		pairs := [][2]string{}
		for id, ar := range a {
			br, ok := bb[id]
			if !ok {
				continue
			}
			av, aok := ar[d.ID]
			bv, bok := br[d.ID]
			if !aok || !bok || missing(av) || missing(bv) {
				continue
			}
			pairs = append(pairs, [2]string{av, bv})
		}
		sort.Slice(pairs, func(i, j int) bool { return pairs[i][0]+"\x00"+pairs[i][1] < pairs[j][0]+"\x00"+pairs[j][1] })
		r := compute(d, pairs)
		results = append(results, r)
	}
	f, err := os.Create(outPath)
	if err != nil {
		return err
	}
	defer f.Close()
	w := csv.NewWriter(f)
	w.Comma = '\t'
	w.Write([]string{"descriptor_id", "n_compared", "agreement", "kappa", "weighted_kappa", "alpha", "decision", "notes"})
	for _, r := range results {
		w.Write([]string{r.ID, strconv.Itoa(r.N), number(r.Agreement), number(r.Kappa), number(r.Weighted), number(r.Alpha), r.Decision, r.Notes})
	}
	w.Flush()
	return w.Error()
}

func read(path string) (map[string]map[string]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	r := csv.NewReader(f)
	r.Comma = '\t'
	r.FieldsPerRecord = -1
	rows, err := r.ReadAll()
	if err != nil {
		return nil, err
	}
	if len(rows) < 2 {
		return nil, fmt.Errorf("%s: empty", path)
	}
	h := rows[0]
	out := map[string]map[string]string{}
	for _, row := range rows[1:] {
		if len(row) != len(h) {
			return nil, fmt.Errorf("%s: malformed row", path)
		}
		m := map[string]string{}
		for i, v := range row {
			m[h[i]] = v
		}
		id := m["page_id"]
		if id == "" || out[id] != nil {
			return nil, fmt.Errorf("%s: duplicate/empty page", path)
		}
		out[id] = m
	}
	return out, nil
}

func missing(v string) bool {
	switch v {
	case "UNCERTAIN", "NOT_VISIBLE", "IMAGE_MISSING":
		return true
	}
	return false
}
func compute(d descriptor, pairs [][2]string) result {
	r := result{ID: d.ID, N: len(pairs), Kappa: math.NaN(), Weighted: math.NaN(), Alpha: math.NaN()}
	if len(pairs) == 0 {
		r.Decision = "REMOVE_OR_COARSEN"
		r.Notes = "NO_COMPARABLE_UNITS"
		return r
	}
	match := 0
	ca := map[string]int{}
	cb := map[string]int{}
	pooled := map[string]int{}
	for _, p := range pairs {
		if p[0] == p[1] {
			match++
		}
		ca[p[0]]++
		cb[p[1]]++
		pooled[p[0]]++
		pooled[p[1]]++
	}
	r.Agreement = float64(match) / float64(len(pairs))
	pe := 0.0
	for k, n := range ca {
		pe += float64(n*cb[k]) / float64(len(pairs)*len(pairs))
	}
	if pe < 1 {
		r.Kappa = (r.Agreement - pe) / (1 - pe)
	}
	do := 1 - r.Agreement
	total := 2 * len(pairs)
	deNum := float64(total * total)
	for _, n := range pooled {
		deNum -= float64(n * n)
	}
	de := deNum / float64(total*(total-1))
	if de > 0 {
		r.Alpha = 1 - do/de
	}
	if d.Type == "ordinal" {
		r.Weighted = weightedKappa(d.Allowed, pairs)
	}
	stat := r.Alpha
	if d.Type == "ordinal" {
		stat = r.Weighted
	}
	switch {
	case !math.IsNaN(stat) && stat >= .80 && r.Agreement >= .85:
		r.Decision = "RETAIN"
		r.Notes = "FIXED_GATE_PASSED"
	case (math.IsNaN(stat) && r.Agreement >= .90) || (!math.IsNaN(stat) && stat >= .60 && stat < .80):
		r.Decision = "CLARIFY_AND_REPILOT"
		r.Notes = "MODERATE_OR_PREVALENCE_CASE"
	case (!math.IsNaN(stat) && stat < .60) || r.Agreement < .75:
		r.Decision = "REMOVE_OR_COARSEN"
		r.Notes = "FIXED_GATE_FAILED"
	default:
		r.Decision = "ADJUDICATE"
		r.Notes = "BORDERLINE_FIXED_GATE"
	}
	return r
}

func weightedKappa(order []string, pairs [][2]string) float64 {
	idx := map[string]int{}
	k := 0
	for _, v := range order {
		if !missing(v) {
			idx[v] = k
			k++
		}
	}
	if k < 2 {
		return math.NaN()
	}
	ca := make([]int, k)
	cb := make([]int, k)
	obs := 0.0
	weight := func(i, j int) float64 { return 1 - math.Abs(float64(i-j))/float64(k-1) }
	for _, p := range pairs {
		i, ok := idx[p[0]]
		if !ok {
			continue
		}
		j, ok := idx[p[1]]
		if !ok {
			continue
		}
		ca[i]++
		cb[j]++
		obs += weight(i, j)
	}
	n := len(pairs)
	obs /= float64(n)
	exp := 0.0
	for i := 0; i < k; i++ {
		for j := 0; j < k; j++ {
			exp += weight(i, j) * float64(ca[i]*cb[j]) / float64(n*n)
		}
	}
	if exp >= 1 {
		return math.NaN()
	}
	return (obs - exp) / (1 - exp)
}
func number(v float64) string {
	if math.IsNaN(v) {
		return "NA"
	}
	return strconv.FormatFloat(v, 'f', 6, 64)
}
