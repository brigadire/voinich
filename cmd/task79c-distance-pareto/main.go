// Command task79c-distance-pareto exercises the Fingerprint v2 distance
// interface (TASK79C_DESIGN.md section 6) and Pareto interface (section 7)
// on a discriminative_validation.json produced by cmd/fingerprint-v2-analyze
// against one or more held-out controls. It computes no new distance
// method: per-metric standardized differences are exactly what
// discriminative_validation.json already reports (unchanged pipeline
// code); this command only aggregates them to family-level and
// family-balanced distances (FINGERPRINT_V2_DISTANCE.md section 2) and
// runs the existing frozen Task66 Pareto dominance rule
// (internal/mechanismspace.ParetoFront) over the result.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"sort"

	"zcore.dev/voinich/internal/mechanismspace"
	"zcore.dev/voinich/internal/workdir"
)

type contrast struct {
	ControlID              string  `json:"control_id"`
	MetricID               string  `json:"metric_id"`
	PrimaryValue           float64 `json:"primary_value"`
	ControlValue           float64 `json:"control_value"`
	StandardizedDifference float64 `json:"standardized_difference"`
	Status                 string  `json:"status"`
	Limitation             string  `json:"limitation"`
}

type registryEntry struct {
	MetricID string `json:"metric_id"`
	Family   string `json:"family"`
}

type familyResult struct {
	Family               string  `json:"family"`
	MetricCount          int     `json:"metric_count"`
	MeanStandardizedDist float64 `json:"mean_standardized_distance"`
	Closeness            float64 `json:"closeness"`
}

type controlResult struct {
	ControlID          string         `json:"control_id"`
	FamilyResults      []familyResult `json:"family_results"`
	FamilyBalancedDist float64        `json:"family_balanced_distance"`
	CommonCoreDist     float64        `json:"common_core_distance"`
	CommonCoreFamilies []string       `json:"common_core_families"`
	AvailableFamilies  int            `json:"available_families"`
	OnParetoFront      bool           `json:"on_pareto_front"`
}

// commonCoreFamilies is fixed in TASK79C_DESIGN.md section 6 item 4: the
// families every corpus in the portfolio can support without IVTFF
// metadata (every plain-text corpus supplies line-count metadata).
var commonCoreFamilies = map[string]bool{
	"lexical paradigm": true,
	"edit family":      true,
	"line":             true,
	"boundary":         true,
	"2D-LITE":          true,
}

func main() {
	os.Exit(run())
}

func run() int {
	discriminative := flag.String("discriminative", "", "path to discriminative_validation.json (required)")
	registry := flag.String("registry", "", "path to metric_registry.json for family grouping (required)")
	out := flag.String("output", "", fmt.Sprintf("path for the combined distance/Pareto result JSON (required; output_dir is explicit, not implicit %s)", workdir.Dir))
	flag.Parse()
	if *discriminative == "" || *registry == "" || *out == "" {
		fmt.Fprintln(os.Stderr, "usage: task79c-distance-pareto -discriminative PATH -registry PATH -output PATH")
		return 2
	}

	var contrasts []contrast
	if err := readJSON(*discriminative, &contrasts); err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		return 1
	}
	var entries []registryEntry
	if err := readJSON(*registry, &entries); err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		return 1
	}
	familyOf := map[string]string{}
	for _, e := range entries {
		familyOf[e.MetricID] = e.Family
	}

	byControl := map[string][]contrast{}
	var controlOrder []string
	for _, c := range contrasts {
		if _, ok := byControl[c.ControlID]; !ok {
			controlOrder = append(controlOrder, c.ControlID)
		}
		byControl[c.ControlID] = append(byControl[c.ControlID], c)
	}
	sort.Strings(controlOrder)

	results := map[string]*controlResult{}
	scoresForPareto := make([]map[string]float64, len(controlOrder))
	for idx, cid := range controlOrder {
		cs := byControl[cid]
		byFamily := map[string][]float64{}
		for _, c := range cs {
			fam, ok := familyOf[c.MetricID]
			if !ok {
				fam = "UNKNOWN"
			}
			byFamily[fam] = append(byFamily[fam], c.StandardizedDifference)
		}
		var families []string
		for f := range byFamily {
			families = append(families, f)
		}
		sort.Strings(families)
		var famResults []familyResult
		closenessMap := map[string]float64{}
		var famBalancedSum float64
		var commonCoreSum float64
		var commonCoreCount int
		var ccFamilies []string
		for _, f := range families {
			vals := byFamily[f]
			mean := 0.0
			for _, v := range vals {
				mean += v
			}
			mean /= float64(len(vals))
			closeness := 1.0 / (1.0 + mean)
			famResults = append(famResults, familyResult{Family: f, MetricCount: len(vals), MeanStandardizedDist: mean, Closeness: closeness})
			closenessMap[f] = closeness
			famBalancedSum += mean
			if commonCoreFamilies[f] {
				commonCoreSum += mean
				commonCoreCount++
				ccFamilies = append(ccFamilies, f)
			}
		}
		famBalanced := famBalancedSum / float64(len(families))
		commonCore := 0.0
		if commonCoreCount > 0 {
			commonCore = commonCoreSum / float64(commonCoreCount)
		}
		results[cid] = &controlResult{
			ControlID: cid, FamilyResults: famResults, FamilyBalancedDist: famBalanced,
			CommonCoreDist: commonCore, CommonCoreFamilies: ccFamilies, AvailableFamilies: len(families),
		}
		scoresForPareto[idx] = closenessMap
	}

	front := mechanismspace.ParetoFront(scoresForPareto)
	onFront := map[int]bool{}
	for _, i := range front {
		onFront[i] = true
	}
	ordered := make([]*controlResult, len(controlOrder))
	for i, cid := range controlOrder {
		results[cid].OnParetoFront = onFront[i]
		ordered[i] = results[cid]
	}

	data, err := json.MarshalIndent(ordered, "", "  ")
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		return 1
	}
	data = append(data, '\n')
	if err := os.WriteFile(*out, data, 0644); err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		return 1
	}
	fmt.Printf("Wrote %s (%d controls)\n", *out, len(ordered))
	for _, r := range ordered {
		fmt.Printf("  %-30s family_balanced=%.4f common_core=%.4f families=%d pareto_front=%v\n",
			r.ControlID, r.FamilyBalancedDist, r.CommonCoreDist, r.AvailableFamilies, r.OnParetoFront)
	}
	return 0
}

func readJSON(path string, v any) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, v)
}
