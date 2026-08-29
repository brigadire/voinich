package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"zcore.dev/voinich/internal/g1v2science"
)

func main() {
	root := "."
	if len(os.Args) == 2 {
		root = os.Args[1]
	}
	p3 := filepath.Join(root, "research", "phase3")
	a, err := g1v2science.LoadAuthority(
		filepath.Join(p3, "task85c-c", "registries", "G1V2_CANDIDATE_REGISTRY.tsv"),
		filepath.Join(p3, "task85c-g", "G1V2_GENERATION_SEMANTICS_V1.json"),
		filepath.Join(p3, "task85c-c", "registries", "G1V2_STATUS_REACHABILITY_CONTRACT_V2.json"),
		filepath.Join(p3, "task85c-j", "G1V2_V1_2_1_EVIDENCE_SCHEMA_REGISTRY.json"),
	)
	check(err)
	var route g1v2science.GenerationRoute
	for _, r := range a.Routes {
		if r.ID == "M1_GEN_A" {
			route = r
		}
	}
	mutated := route
	mutated.Parameters = map[string]any{"deliberately": "scientifically different"}
	r1, _ := g1v2science.NewRNG("g1v2/control/generate", uint64(route.Index), 0, 0)
	r2, _ := g1v2science.NewRNG("g1v2/control/generate", uint64(route.Index), 0, 0)
	g1, e1 := g1v2science.GenerateSynthetic(route, 64, &r1)
	g2, e2 := g1v2science.GenerateSynthetic(mutated, 64, &r2)
	check(e1)
	check(e2)
	if g1.CorpusSHA256 != g2.CorpusSHA256 {
		panic("parameter mutation unexpectedly changed output")
	}
	c, _ := g1v2science.NewCorpus([]string{"ab", "a", "ba", "aba"})
	f2, err := g1v2science.F2Metrics(c)
	check(err)
	result := map[string]any{
		"schema":                     "task86c-v2-v1_2_1-implementation-mismatch-v1",
		"status":                     "IMPLEMENTATION_VALIDATION_FAILURE",
		"implementation_root_sha256": "5687f219c049f6e38b2a9048c7799965948e34b90fa2fe37a6a3679427ff7a0b",
		"findings": []any{
			map[string]any{"id": "PF-IMPL-01-GENERATION-ROUTE-PARAMETERS-IGNORED", "route": route.ID, "original_parameters": route.Parameters, "mutated_parameters": mutated.Parameters, "original_corpus_sha256": g1.CorpusSHA256, "mutated_corpus_sha256": g2.CorpusSHA256, "equal": true},
			map[string]any{"id": "PF-IMPL-02-F2-SCIENTIFIC-WEIGHT-MISMATCH", "EF3_expected_weight": 1, "EF3_actual_weight": f2["EF3_DEGREE_FREQUENCY_SPEARMAN"].ScientificWeight, "HR1_expected_weight": 0, "HR1_actual_weight": f2["HR1_FOLIO_VARIANCE_SHARE"].ScientificWeight},
		},
		"materialization_allowed": false,
	}
	b, _ := json.MarshalIndent(result, "", "  ")
	fmt.Println(string(b))
}

func check(err error) {
	if err != nil {
		panic(err)
	}
}
