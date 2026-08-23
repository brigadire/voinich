package main

import (
	"bufio"
	"os"
	"strconv"
	"strings"
)

// readFamilyMetricsTSV parses a FAMILY_METRICS.tsv/ARCHITECTURE_ABLATION.tsv
// -shaped file back into []FamilyMetricsRow, so `report` can regenerate
// REPORT.md from already-computed artifacts without rerunning the grid.
func readFamilyMetricsTSV(path, mechCol, corpusCol, familyCol, progressCol, statusCol string) []FamilyMetricsRow {
	f, err := os.Open(path)
	if err != nil {
		return nil
	}
	defer f.Close()
	sc := bufio.NewScanner(f)
	sc.Buffer(make([]byte, 4096), 1<<20)
	var headers []string
	byKey := map[string]*FamilyMetricsRow{}
	var order []string
	for sc.Scan() {
		fields := strings.Split(sc.Text(), "\t")
		if headers == nil {
			headers = fields
			continue
		}
		row := map[string]string{}
		for i, h := range headers {
			if i < len(fields) {
				row[h] = fields[i]
			}
		}
		key := row[mechCol] + "|" + row[corpusCol]
		r, ok := byKey[key]
		if !ok {
			r = &FamilyMetricsRow{Mechanism: row[mechCol], Corpus: row[corpusCol], FamilyScores: map[string]float64{}, OverallStatus: row[statusCol]}
			byKey[key] = r
			order = append(order, key)
		}
		v, _ := strconv.ParseFloat(row[progressCol], 64)
		r.FamilyScores[row[familyCol]] = v
	}
	var out []FamilyMetricsRow
	for _, k := range order {
		out = append(out, *byKey[k])
	}
	return out
}

func readFinalArchitectureTSV(path string) []FinalVerdict {
	f, err := os.Open(path)
	if err != nil {
		return nil
	}
	defer f.Close()
	sc := bufio.NewScanner(f)
	first := true
	var out []FinalVerdict
	for sc.Scan() {
		if first {
			first = false
			continue
		}
		fields := strings.SplitN(sc.Text(), "\t", 4)
		if len(fields) < 3 {
			continue
		}
		v := FinalVerdict{Operation: fields[0], Verdict: fields[1], Evidence: fields[2]}
		if len(fields) > 3 {
			v.Caveat = fields[3]
		}
		out = append(out, v)
	}
	return out
}

func readFrontierFromParetoTSV(path string) []string {
	f, err := os.Open(path)
	if err != nil {
		return nil
	}
	defer f.Close()
	sc := bufio.NewScanner(f)
	first := true
	var out []string
	for sc.Scan() {
		if first {
			first = false
			continue
		}
		fields := strings.Split(sc.Text(), "\t")
		if len(fields) >= 2 && fields[1] == "true" {
			out = append(out, fields[0])
		}
	}
	return out
}

func readOverfitFromHeldoutTSV(path string) map[string]string {
	out := map[string]string{}
	f, err := os.Open(path)
	if err != nil {
		return out
	}
	defer f.Close()
	sc := bufio.NewScanner(f)
	first := true
	for sc.Scan() {
		if first {
			first = false
			continue
		}
		fields := strings.Split(sc.Text(), "\t")
		if len(fields) >= 5 {
			out[fields[0]] = fields[4]
		}
	}
	return out
}
