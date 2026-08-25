package main

import (
	"bufio"
	"encoding/json"
	"os"
	"strings"
)

// Candidate is one row of the frozen 84-row G1_HYPERPARAMETER_GRID.tsv,
// parsed directly from the frozen artifact (never re-derived) so the
// candidate set is byte-identical to what Task85a validated.
type Candidate struct {
	ModelClass  string
	CandidateID string
	Params      map[string]float64
}

func (c Candidate) Int(name string, def int) int {
	if v, ok := c.Params[name]; ok {
		return int(v)
	}
	return def
}
func (c Candidate) Float(name string, def float64) float64 {
	if v, ok := c.Params[name]; ok {
		return v
	}
	return def
}

func loadCandidateGrid(path string) ([]Candidate, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	sc := bufio.NewScanner(f)
	sc.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	var cols []string
	header := true
	var out []Candidate
	for sc.Scan() {
		line := sc.Text()
		if line == "" {
			continue
		}
		fields := parseTSVLine(line)
		if header {
			cols = fields
			header = false
			continue
		}
		row := map[string]string{}
		for i, c := range cols {
			if i < len(fields) {
				row[c] = fields[i]
			}
		}
		var params map[string]float64
		if err := json.Unmarshal([]byte(row["parameters_json"]), &params); err != nil {
			return nil, err
		}
		out = append(out, Candidate{ModelClass: row["model_class"], CandidateID: row["candidate_id"], Params: params})
	}
	if err := sc.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

// parseTSVLine handles the csv.DictWriter-quoted parameters_json column
// (Python csv writer quotes fields containing the delimiter or quotes,
// doubling embedded quotes) without pulling in encoding/csv's own
// delimiter assumptions.
func parseTSVLine(line string) []string {
	var fields []string
	var cur strings.Builder
	inQuotes := false
	runes := []rune(line)
	for i := 0; i < len(runes); i++ {
		ch := runes[i]
		switch {
		case inQuotes:
			if ch == '"' {
				if i+1 < len(runes) && runes[i+1] == '"' {
					cur.WriteRune('"')
					i++
				} else {
					inQuotes = false
				}
			} else {
				cur.WriteRune(ch)
			}
		case ch == '"':
			inQuotes = true
		case ch == '\t':
			fields = append(fields, cur.String())
			cur.Reset()
		default:
			cur.WriteRune(ch)
		}
	}
	fields = append(fields, cur.String())
	return fields
}

func candidatesByClass(all []Candidate, class string) []Candidate {
	var out []Candidate
	for _, c := range all {
		if c.ModelClass == class {
			out = append(out, c)
		}
	}
	return out
}
