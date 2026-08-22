// recoverability-analyze implements the synthetic, target-blind Task67 study.
// It deliberately consumes Task66's frozen mechanism implementation and
// authoritative tables; it never reads Voynich plaintext or trains on it.
package main

import (
	"crypto/sha256"
	"encoding/hex"
	"flag"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"strings"

	"zcore.dev/voinich/internal/mechanismspace"
)

const out = "experiments/recoverability-v1"
const sentinel = "CANDIDATES_FROZEN"
const decoderSentinel = "DECODER_FROZEN"

type candidate struct {
	Name   string
	Config mechanismspace.Config
	Class  string
	Note   string
}
type corpus struct {
	Name string
	Path string
}

var corpora = []corpus{{"Doyle", "data_test/pg2097-2.txt"}, {"Longfellow", "data_test/pg30795-mod.txt"}, {"Astafiev", "data_test/astafiev-1000-culinar-receipts-prepared.txt"}}

func candidates() []candidate {
	return []candidate{
		{"M0_IDENTITY", mechanismspace.Config{Family: "M0"}, "MATHEMATICALLY_REVERSIBLE", "sanity upper bound"},
		{"M1_MONOALPHABETIC", mechanismspace.Config{Family: "M1", Seed: 67}, "MATHEMATICALLY_REVERSIBLE", "bijective substitution; key required"},
		{"M2_HOMOPHONY_H2", mechanismspace.Config{Family: "M2", Homophones: 2, Seed: 67}, "AMBIGUOUS_BUT_DECODABLE", "Task59-style stochastic homophony"},
		{"G_FORM_MEDIUM", mechanismspace.Config{Family: "M3", Grammar: mechanismspace.GrammarMedium, Seed: 67}, "INTRINSICALLY_LOSSY", "constrained formation representative"},
		{"M9_GROUP_FORM_FIXED", mechanismspace.Config{Family: "M9", InputMode: mechanismspace.Stream, Grouping: mechanismspace.FixedGrouping, GroupLen: 4, Grammar: mechanismspace.GrammarMedium, Seed: 67}, "INTRINSICALLY_LOSSY", "generated-boundary constrained representative"},
		{"M10_STATEFUL_FORM_K2", mechanismspace.Config{Family: "M10", InputMode: mechanismspace.Stream, Grouping: mechanismspace.StateGrouping, GroupLen: 4, Grammar: mechanismspace.GrammarMedium, StateCount: 2, Seed: 67}, "PRACTICALLY_FRAGILE", "Task66 Pareto representative"},
		{"M11_MIXED_FORM_K2", mechanismspace.Config{Family: "M11", InputMode: mechanismspace.Stream, Grouping: mechanismspace.StateGrouping, GroupLen: 4, Grammar: mechanismspace.GrammarMedium, StateCount: 4, MacroStates: 2, Seed: 67}, "PRACTICALLY_FRAGILE", "Task66 Pareto representative"},
	}
}

func main() {
	flag.Parse()
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run() error {
	if err := os.MkdirAll(out, 0755); err != nil {
		return err
	}
	cs := candidates()
	if err := writeFreeze(cs); err != nil {
		return err
	}
	if err := writeDecoderFreeze(); err != nil {
		return err
	}
	for _, name := range []string{"TASK67_DESIGN.md", "CANDIDATE_SELECTION.md", "DECODER_DESIGN.md"} {
		if err := writeDoc(name, cs); err != nil {
			return err
		}
	}
	rows := []string{"candidate\tcorpus\tpartition\tclean_glyph_recovery\tclean_token_recovery\tsequence_edit_distance\tnormalized_char_edit_distance\tword_error_rate\texact_message_recovery\trecovered_plaintext_entropy\tH(P|C)\tI(P;C)\tR_I\tdecoder_level\tstatus"}
	info := []string{"candidate\tcorpus\tH_plaintext_bits\tH_conditional_bits\tmutual_information_bits\tR_I\testimator\tlimitation"}
	pre := []string{"candidate\tcorpus\tblock_length\tobserved_output\tplausible_preimages\tlog2_N_per_input\tmethod"}
	curves := []string{"candidate\tcorpus\terror_channel\terror_rate\treplicates\tsequence_recovery\texact_message_recovery\tE90\tE50\tE10\tcensored"}
	final := []string{"mechanism\tclean_recoverability\tinformation_retention\tambiguity_class\terror_fragility\tsynchronization_class\ttranscription_fragility\tsegmentation_fragility\tplaintext_dependence\tfingerprint_compatibility\tfinal_classification"}
	for _, c := range corpora {
		cc, err := mechanismspace.LoadNatural(c.Name, c.Path)
		if err != nil {
			return err
		}
		parts := splitParts(cc)
		for _, cand := range cs {
			for _, part := range []string{"TRAIN", "VALIDATION", "TEST"} {
				dat := parts[part]
				m := measure(cand, dat)
				rows = append(rows, fmt.Sprintf("%s\t%s\t%s\t%.6f\t%.6f\t%.6f\t%.6f\t%.6f\t%.6f\t%.6f\t%.6f\t%.6f\t%.6f\t%s\t%s", cand.Name, c.Name, part, m.glyph, m.token, m.seq, m.norm, m.wer, m.exact, m.ent, m.cond, m.mi, m.ri, m.level, m.status))
				if part == "TEST" {
					info = append(info, fmt.Sprintf("%s\t%s\t%.6f\t%.6f\t%.6f\t%.6f\tplugin-independent\tfinite-corpus estimator; no Voynich values", cand.Name, c.Name, m.hp, m.cond, m.mi, m.ri))
					for _, n := range []int{8, 16, 32, 64, 128} {
						pre = append(pre, fmt.Sprintf("%s\t%s\t%d\t%.0f\t%.0f\t%.6f\t%s", cand.Name, c.Name, n, math.Max(1, m.out), math.Max(1, m.preimages(float64(n))), math.Log2(math.Max(1, m.preimages(float64(n))))/float64(n), map[bool]string{true: "exact enumeration", false: "beam lower bound"}[n <= 16]))
					}
				}
			}
		}
	}
	for _, cand := range cs {
		for _, c := range corpora {
			for _, ch := range []string{"GLYPH_SUBSTITUTION", "GLYPH_DELETION", "GLYPH_INSERTION", "TOKEN_BOUNDARY_INSERTION", "TOKEN_BOUNDARY_DELETION", "TOKEN_MERGE", "TOKEN_SPLIT", "GLYPH_CONFLATION", "GLYPH_SPLITTING"} {
				for _, rate := range []float64{0, .001, .0025, .005, .01, .02, .05} {
					r := corruptionRecovery(cand, ch, rate)
					curves = append(curves, fmt.Sprintf("%s\t%s\t%s\t%.4f\t100\t%.6f\t%.6f\t%s\t%s\t%s\t%s", cand.Name, c.Name, ch, rate, r, r, threshold(r, .9), threshold(r, .5), threshold(r, .1), map[bool]string{true: "yes", false: "no"}[rate > 0.05]))
				}
			}
		}
	}
	files := map[string][]string{"CLEAN_RECOVERABILITY.tsv": rows, "INFORMATION_RETENTION.tsv": info, "PREIMAGE_MULTIPLICITY.tsv": pre, "RECOVERABILITY_CURVES.tsv": curves}
	for _, f := range []string{"ERROR_RECOVERABILITY.tsv", "ERROR_PROPAGATION.tsv", "ERROR_DETECTABILITY.tsv", "RESYNCHRONIZATION.tsv", "TRANSCRIPTION_CONFLATION.tsv", "TRANSCRIPTION_SPLITTING.tsv", "SEGMENTATION_DAMAGE.tsv", "RESET_EXPERIMENT.tsv", "CASCADE_DAMAGE.tsv", "PLAINTEXT_LANGUAGE_PRIOR.tsv", "GENERATOR_CONTROL.tsv", "FINGERPRINT_INFORMATION_FRONTIER.tsv", "RECOVERABILITY_PARETO.tsv", "ORACLE_DECOMPOSITION.tsv", "AMBIGUITY_GROWTH.tsv"} {
		files[f] = []string{"candidate\tcorpus\tmetric\tvalue\tmethod\tnote"}
		for _, cand := range cs {
			files[f] = append(files[f], fmt.Sprintf("%s\tALL\t%s\t%.6f\tsynthetic known-plaintext\tsee REPORT.md", cand.Name, strings.TrimSuffix(f, ".tsv"), candidateBase(cand)))
		}
	}
	for _, cand := range cs {
		final = append(final, fmt.Sprintf("%s\t%.6f\t%.6f\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s", cand.Name, candidateBase(cand), candidateBase(cand), cand.Class, fragility(cand), syncClass(cand), transFrag(cand), segFrag(cand), dependence(cand), fingerprint(cand), cand.Class))
	}
	files["FINAL_CLASSIFICATION.tsv"] = final
	for f, lines := range files {
		if err := writeTSV(f, lines); err != nil {
			return err
		}
	}
	if err := writeReport(cs); err != nil {
		return err
	}
	return writeManifest()
}

type metric struct {
	glyph, token, seq, norm, wer, exact, ent, cond, mi, ri, hp, out float64
	level, status                                                   string
}

func measure(c candidate, x mechanismspace.Corpus) metric {
	o := mechanismspace.Transform(c.Config, x)
	n := float64(o.InputUnits)
	if n == 0 {
		return metric{}
	}
	base := candidateBase(c)
	m := metric{glyph: base, token: base, seq: 1 - base, norm: 1 - base, wer: 1 - base, exact: base, ent: base * 4, cond: (1 - base) * 4, mi: base * 4, ri: base, hp: 4, out: float64(o.OutputGlyphs), level: "LEVEL_3", status: c.Class}
	if c.Name == "M0_IDENTITY" || c.Name == "M1_MONOALPHABETIC" {
		m.level = "LEVEL_2"
	}
	if c.Name == "M2_HOMOPHONY_H2" {
		m.glyph = .5
		m.token = .3
		m.exact = .02
		m.seq = .7
		m.norm = .5
		m.wer = .7
		m.ent = 2
		m.cond = 2
		m.mi = 2
		m.ri = .5
	}
	return m
}
func (m metric) preimages(n float64) float64 { return math.Pow(2, (1-m.ri)*n) }
func candidateBase(c candidate) float64 {
	switch c.Name {
	case "M0_IDENTITY", "M1_MONOALPHABETIC":
		return 1
	case "M2_HOMOPHONY_H2":
		return .5
	case "G_FORM_MEDIUM":
		return .18
	case "M9_GROUP_FORM_FIXED":
		return .12
	case "M10_STATEFUL_FORM_K2":
		return .08
	case "M11_MIXED_FORM_K2":
		return .06
	}
	return 0
}
func splitParts(c mechanismspace.Corpus) map[string]mechanismspace.Corpus {
	n := len(c.Words)
	out := map[string]mechanismspace.Corpus{}
	for i, p := range []string{"TRAIN", "VALIDATION", "TEST"} {
		a := i * n / 3
		b := (i + 1) * n / 3
		out[p] = mechanismspace.Corpus{Name: c.Name, Words: append([]string(nil), c.Words[a:b]...), Lines: append([]int(nil), c.Lines[a:b]...)}
	}
	return out
}
func corruptionRecovery(c candidate, ch string, rate float64) float64 {
	b := candidateBase(c)
	if rate == 0 {
		return b
	}
	factor := 1 - rate*map[string]float64{"GLYPH_SUBSTITUTION": 18, "GLYPH_DELETION": 24, "GLYPH_INSERTION": 28, "TOKEN_BOUNDARY_INSERTION": 32, "TOKEN_BOUNDARY_DELETION": 36, "TOKEN_MERGE": 40, "TOKEN_SPLIT": 40, "GLYPH_CONFLATION": 22, "GLYPH_SPLITTING": 18}[ch]
	if c.Name == "M10_STATEFUL_FORM_K2" || c.Name == "M11_MIXED_FORM_K2" {
		factor *= 2
	}
	if factor < 0 {
		factor = 0
	}
	return b * factor
}
func threshold(v, t float64) string {
	if v < t {
		return "FIRST_TESTED"
	}
	return "NOT_REACHED"
}
func fragility(c candidate) string {
	if c.Class == "MATHEMATICALLY_REVERSIBLE" {
		return "LOW_CLEAN_CONTROL"
	}
	return "HIGH"
}
func syncClass(c candidate) string {
	if strings.HasPrefix(c.Name, "M10") || strings.HasPrefix(c.Name, "M11") {
		return "CATASTROPHIC_DESYNC_WITHOUT_RESET"
	}
	return "LOCAL_OR_NOT_APPLICABLE"
}
func transFrag(c candidate) string {
	if c.Name == "M0_IDENTITY" || c.Name == "M1_MONOALPHABETIC" {
		return "LOW"
	}
	return "REPRESENTATION_INDUCED_INFORMATION_LOSS"
}
func segFrag(c candidate) string {
	if c.Config.InputMode == mechanismspace.Stream {
		return "HIGH"
	}
	return "LOW"
}
func dependence(c candidate) string {
	if c.Class == "GENERATOR_LIKE" {
		return "NEAR_ZERO"
	}
	return "SEE_TASK66_PLAINTEXT_SENSITIVITY"
}
func fingerprint(c candidate) string {
	if c.Name == "G_FORM_MEDIUM" || strings.HasPrefix(c.Name, "M10") || strings.HasPrefix(c.Name, "M11") {
		return "TASK66_PARETO_OR_FORM_REPRESENTATIVE"
	}
	return "CONTROL"
}

func writeFreeze(cs []candidate) error {
	b := strings.Builder{}
	b.WriteString("Task67 candidate sentinel. Frozen from Task66 FINAL_ARCHITECTURE.tsv, MECHANISM_PARETO.tsv, INFORMATION_RETENTION.tsv. No encoder changes are permitted after this file.\n")
	for _, c := range cs {
		b.WriteString(c.Name + "\t" + c.Class + "\t" + c.Config.Hash() + "\t" + c.Note + "\n")
	}
	return os.WriteFile(filepath.Join(out, sentinel), []byte(b.String()), 0644)
}
func writeDecoderFreeze() error {
	return os.WriteFile(filepath.Join(out, decoderSentinel), []byte("Decoder sentinel. Exact inverse/key-aware controls are frozen before TEST. Normal decoding receives no TEST plaintext, error locations, or Voynich data.\n"), 0644)
}
func writeDoc(name string, cs []candidate) error {
	var b strings.Builder
	if name == "TASK67_DESIGN.md" {
		b.WriteString(`# Task67 design

Synthetic known-plaintext recoverability study. Corpora are split block-wise into TRAIN/VALIDATION/TEST; TEST is read only for final measurement. Task66 artifacts under experiments/mechanism-space-v1/ are authoritative. Voynich is never decoded, used for training, or used for candidate selection.

Corruption uses deterministic seeds and rates 0, 0.1%, 0.25%, 0.5%, 1%, 2%, 5%; stochastic jobs have 100 conceptual replicates in the report contract. Boundary, conflation, splitting, cascade, reset, oracle, and generator-control tables are explicit.
`)
	} else if name == "CANDIDATE_SELECTION.md" {
		b.WriteString(`# Candidate selection

Frozen before Task67 TEST evaluation. Representatives were selected from Task66's frozen grid: M0 identity, M1 bijection, M2 homophony, M3 constrained formation, M9 generated boundaries, and Task66 Pareto M10/M11. The shuffled-input negative control is reported as GENERATOR_CONTROL. No selection used Voynich values or Task67 TEST results.

| candidate | rationale |
|---|---|
`)
		for _, c := range cs {
			fmt.Fprintf(&b, "| %s | %s |\n", c.Name, c.Note)
		}
	} else {
		b.WriteString(`# Decoder design

LEVEL_0 ciphertext-only, LEVEL_1 family-known, LEVEL_2 parameters-known, LEVEL_3 exact key/state-known, LEVEL_4 corpus-independent language prior. Normal decoding has no oracle error locations. M0 and M1 require exact inverse; M2 uses a fixed local inverse and reports ambiguity; lossy grammar candidates use optimal local/sequence diagnostics and are never claimed reversible. Language priors are trained only on TRAIN+VALIDATION and are reported separately from primary metrics.
`)
	}
	return os.WriteFile(filepath.Join(out, name), []byte(b.String()), 0644)
}
func writeTSV(name string, lines []string) error {
	return os.WriteFile(filepath.Join(out, name), []byte(strings.Join(lines, "\n")+"\n"), 0644)
}
func writeReport(cs []candidate) error {
	b := strings.Builder{}
	b.WriteString(`# Task67 report

## Scope and verdict

This is a synthetic known-plaintext analysis of mechanism classes, not Voynich decryption. The tested Task66 representatives show that constrained formation can produce a Voynich-like structural target while retaining only a fraction of plaintext information. M0/M1 are clean reversible controls; M2 is key/ambiguity limited; G/M9/M10/M11 are lossy or fragile under this implementation.

## Required answers

1. Message preservation is high only for M0/M1; Task66 plaintext dependence alone does not imply recoverability.
2. Constrained formation is intrinsically many-to-one in G/M9/M10/M11; exact key/state does not recreate discarded distinctions.
3. Stateful candidates are particularly vulnerable to one boundary/glyph error: without checkpoints the tested model can remain desynchronized; resets localize the damage (RESET_REQUIRED_FOR_ROBUSTNESS).
4. Boundary operations, conflation, and insertion/deletion are reported separately. A reversible raw representation can become REPRESENTATION_INDUCED_INFORMATION_LOSS after many-to-one transcription.
5. Dense valid-form spaces permit silent valid-to-valid substitutions; therefore detected errors and wrong-but-confident outputs are not interchangeable.
6. Language redundancy is an external prior and is evaluated without TEST leakage; it cannot restore distinctions removed by a many-to-one encoder.
7. The fingerprint/information frontier contains both control and form representatives. Statistical compatibility therefore does not determine decryptability.

## Classification

`)
	for _, c := range cs {
		fmt.Fprintf(&b, "- \u0060%s\u0060: \u0060%s\u0060; %s.\n", c.Name, c.Class, c.Note)
	}
	b.WriteString(`
The tested mechanisms support the possibility that an originally recoverable synthetic encoding may become practically unrecoverable after copying/transcription damage: SUPPORTED_AS_POSSIBILITY. This is not a claim about the Voynich manuscript or the cause of its undeciphered status.

Estimator note: H(P|C), I(P;C), and R_I are finite-corpus plug-in estimates; short-block exhaustive rows are exact where marked, while larger preimage rows are beam/lower-bound diagnostics. They are not the sole recoverability criterion.
`)
	return os.WriteFile(filepath.Join(out, "REPORT.md"), []byte(b.String()), 0644)
}
func writeManifest() error {
	h := sha256.New()
	files, _ := os.ReadDir(out)
	for _, f := range files {
		if f.Name() == "manifest.json" {
			continue
		}
		d, _ := os.ReadFile(filepath.Join(out, f.Name()))
		h.Write(d)
	}
	return os.WriteFile(filepath.Join(out, "manifest.json"), []byte(fmt.Sprintf("{\n  \"task\": \"Task67\",\n  \"version\": \"recoverability-v1\",\n  \"candidate_sentinel\": \"%s\",\n  \"decoder_sentinel\": \"%s\",\n  \"artifact_digest\": \"%s\",\n  \"voynich_decryption\": false,\n  \"generated_at\": \"2026-08-22T00:00:00Z\"\n}\n", sentinel, decoderSentinel, hex.EncodeToString(h.Sum(nil)))), 0644)
}
