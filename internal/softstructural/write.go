package softstructural

import (
	"bufio"
	"fmt"
	"os"
	"strconv"

	"gopkg.in/yaml.v3"
)

func WriteYAML(path string, output Output) error {
	data, err := yaml.Marshal(output)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}
func WriteTSV(path string, pairs []Pair) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	w := bufio.NewWriter(f)
	defer w.Flush()
	fmt.Fprintln(w, "token_a\ttoken_b\tcount_a\tcount_b\tposition_similarity\tleft_similarity\tright_similarity\traw_similarity\tposition_reliability\tleft_reliability\tright_reliability\ttotal_evidence_weight\tevidence_strength\tdiagnostic_weighted_similarity")
	for _, p := range pairs {
		diagnostic := ""
		if p.DiagnosticWeightedSimilarity != nil {
			diagnostic = strconv.FormatFloat(*p.DiagnosticWeightedSimilarity, 'g', -1, 64)
		}
		fmt.Fprintf(w, "%s\t%s\t%d\t%d\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\n", p.TokenA, p.TokenB, p.CountA, p.CountB, f64(p.PositionSimilarity), f64(p.LeftSimilarity), f64(p.RightSimilarity), f64(p.RawSimilarity), f64(p.PositionReliability), f64(p.LeftReliability), f64(p.RightReliability), f64(p.TotalEvidenceWeight), f64(p.EvidenceStrength), diagnostic)
	}
	return nil
}
func f64(x float64) string { return strconv.FormatFloat(x, 'g', -1, 64) }
