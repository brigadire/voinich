package vocabularygrowth

import (
	"fmt"
	"math"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"
)

func Write(result Result, outputDir string) error {
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return err
	}
	if err := tsv(filepath.Join(outputDir, "vocabulary_growth.tsv"), []string{"checkpoint_n", "vocabulary_size", "type_token_ratio", "hapax_count", "dis_count", "tri_count", "beta_effective"}, func(y func(...any)) {
		for _, p := range result.Growth {
			y(p.N, p.Vocabulary, p.TTR, p.Hapax, p.Dis, p.Tri, p.BetaEffective)
		}
	}); err != nil {
		return err
	}
	if err := tsv(filepath.Join(outputDir, "new_type_rate.tsv"), []string{"window_start", "window_end", "tokens", "new_types", "new_type_rate"}, func(y func(...any)) {
		for _, p := range result.Windows {
			y(p.Start, p.End, p.Tokens, p.NewTypes, p.NewTypeRate)
		}
	}); err != nil {
		return err
	}
	if err := tsv(filepath.Join(outputDir, "frequency_of_frequencies.tsv"), []string{"checkpoint_n", "vocabulary_size", "hapax_count", "dis_legomena_count", "tris_legomena_count", "hapax_fraction_of_types", "hapax_fraction_of_tokens", "dis_fraction_of_types", "dis_fraction_of_tokens"}, func(y func(...any)) {
		for _, p := range result.Growth {
			types, tokens := float64(p.Vocabulary), float64(p.N)
			y(p.N, p.Vocabulary, p.Hapax, p.Dis, p.Tri, float64(p.Hapax)/types, float64(p.Hapax)/tokens, 2*float64(p.Dis)/types, 2*float64(p.Dis)/tokens)
		}
	}); err != nil {
		return err
	}
	if err := tsv(filepath.Join(outputDir, "vocabulary_growth_null.tsv"), []string{"checkpoint_n", "observed", "null_mean", "null_sd", "effect", "empirical_p"}, func(y func(...any)) {
		for _, p := range result.Null {
			y(p.N, p.Observed, p.NullMean, p.NullSD, p.Effect, p.EmpiricalP)
		}
	}); err != nil {
		return err
	}
	if err := tsv(filepath.Join(outputDir, "segment_vocabulary_growth.tsv"), []string{"segment_count", "segment", "checkpoint_n", "vocabulary_size", "heaps_beta", "beta_effective", "new_type_rate"}, func(y func(...any)) {
		for _, p := range result.Segments {
			y(p.Segments, p.Segment, p.CheckpointN, p.Vocabulary, p.HeapsBeta, p.BetaEffective, p.NewTypeRate)
		}
	}); err != nil {
		return err
	}
	profile := map[string]any{"final_profile": FinalProfileFromPoint(result), "heaps_fit": result.Fit, "parameters": result.Parameters, "total_tokens": result.TotalTokens}
	b, err := yaml.Marshal(profile)
	if err != nil {
		return err
	}
	if err = os.WriteFile(filepath.Join(outputDir, "summary.yaml"), b, 0644); err != nil {
		return err
	}
	b, err = yaml.Marshal(map[string]any{"heaps_fit": result.Fit, "fit_model": "log(V(n)) = log(K) + beta*log(n)", "descriptive_only": true})
	if err != nil {
		return err
	}
	if err = os.WriteFile(filepath.Join(outputDir, "heaps_fit.yaml"), b, 0644); err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(outputDir, "REPORT.md"), []byte(report(result)), 0644)
}

func FinalProfileFromPoint(r Result) map[string]float64 {
	m := map[string]float64{"total_tokens": float64(r.TotalTokens), "unique_tokens": float64(r.Final.Vocabulary), "type_token_ratio": r.Final.TTR, "hapax": float64(r.Final.Hapax), "dis_legomena": float64(r.Final.Dis), "tris_legomena": float64(r.Final.Tri), "hapax_fraction_of_types": float64(r.Final.Hapax) / float64(r.Final.Vocabulary), "hapax_fraction_of_tokens": float64(r.Final.Hapax) / float64(r.TotalTokens), "dis_fraction_of_types": 2 * float64(r.Final.Dis) / float64(r.Final.Vocabulary), "dis_fraction_of_tokens": 2 * float64(r.Final.Dis) / float64(r.TotalTokens)}
	if r.Final.Dis > 0 {
		m["singleton_to_doubleton_ratio"] = float64(r.Final.Hapax) / float64(r.Final.Dis)
	} else {
		m["singleton_to_doubleton_ratio"] = math.Inf(1)
	}
	m["heaps_K"], m["heaps_beta"], m["heaps_R2"] = r.Fit.K, r.Fit.Beta, r.Fit.R2
	return m
}
func tsv(path string, h []string, fill func(func(...any))) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	if _, err = fmt.Fprintln(f, strings.Join(h, "\t")); err != nil {
		return err
	}
	fill(func(xs ...any) {
		out := make([]string, len(xs))
		for i, x := range xs {
			if v, ok := x.(float64); ok {
				if math.IsInf(v, 0) {
					out[i] = "Inf"
				} else {
					out[i] = strconv.FormatFloat(v, 'g', 17, 64)
				}
			} else {
				out[i] = fmt.Sprint(x)
			}
		}
		_, _ = fmt.Fprintln(f, strings.Join(out, "\t"))
	})
	return nil
}
func report(r Result) string {
	var b strings.Builder
	b.WriteString("# Vocabulary growth analysis\n\nThis is a language-agnostic descriptive corpus analysis. Heaps-law fit, hapax counts, and null effects do not establish language, cipher, generator, or semantic class.\n\n## Fixed parameters\n\n")
	fmt.Fprintf(&b, "- tokens: %d\n- checkpoints: %v\n- windows: %v\n- segments: %v\n- null permutations: %d\n- seed: %d\n\n", r.TotalTokens, r.Checkpoints, r.Parameters.WindowSizes, r.Parameters.SegmentCounts, r.Parameters.NullPermutations, r.Parameters.Seed)
	fmt.Fprintf(&b, "## Heaps fit\n\n`V(n) = K * n^beta`, K=%.10g, beta=%.10g, R²=%.10g, fitting range=[%d,%d], points=%d.\n\n", r.Fit.K, r.Fit.Beta, r.Fit.R2, r.Fit.NMin, r.Fit.NMax, r.Fit.Points)
	b.WriteString("## Outputs\n\n- `vocabulary_growth.tsv`: observed trajectory and effective beta.\n- `frequency_of_frequencies.tsv`: hapax/dis/tris-legomena trajectory.\n- `new_type_rate.tsv`: fixed-window productivity.\n- `vocabulary_growth_null.tsv`: deterministic shuffled-token null ensemble.\n- `segment_vocabulary_growth.tsv`: positional segment analysis.\n\n## Limitations\n\nThe observed trajectory depends on token order and the canonical tokenization contract. The shuffled null preserves the token multiset but destroys order. Segment analysis is positional only and uses no manuscript metadata. Corpus-size comparisons must use a common checkpoint; unavailable checkpoints are not zero.\n")
	return b.String()
}
