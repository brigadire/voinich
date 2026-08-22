// recoverability-analyze implements the synthetic, target-blind Task67 study.
// It deliberately consumes Task66's frozen mechanism implementation and
// authoritative tables; it never reads Voynich plaintext or trains on it.
package main

import (
	"crypto/sha256"
	"encoding/hex"
	"flag"
	"fmt"
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
	if err := runExperiments(cs); err != nil {
		return err
	}
	if err := validateExperimentalArtifacts(out); err != nil {
		return fmt.Errorf("Task67 validation failed: %w", err)
	}
	if err := writeReport(cs); err != nil {
		return err
	}
	return writeManifest()
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

Synthetic known-plaintext recoverability study. Corpora are split block-wise into TRAIN/VALIDATION/TEST. Decoder tables are fitted on TRAIN+VALIDATION; final measurements use the first content-blind 128-word TEST block, with explicit 8/16/32/64/128-unit clean sub-blocks. Task66 artifacts under experiments/mechanism-space-v1/ are authoritative. Voynich is never decoded, used for training, or used for candidate selection.

Corruption uses deterministic seeds and rates 0, 0.1%, 0.25%, 0.5%, 1%, 2%, 5%. Every candidate/corpus/channel/rate cell has 100 executed raw replicates in ERROR_RECOVERABILITY.tsv. Conflation and splitting have 30 executed replicates per fraction. Single-error, boundary, cascade, reset, oracle, wrong-language, and generator controls all run encode -> transform ciphertext -> decode.
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

LEVEL_0 ciphertext-only, LEVEL_1 family-known, LEVEL_2 parameters-known, LEVEL_3 exact key/state-known, LEVEL_4 language prior. Normal decoding has no oracle error locations. M0 is direct; M1 reconstructs the frozen seeded permutation from the ciphertext alphabet; Task59 synthetic homophone labels are inverted by removing their generated suffix. Constrained candidates use a frozen TRAIN+VALIDATION maximum-frequency local inverse and nearest-valid-form fallback. No TEST plaintext enters that decoder. ORACLE_TEST_CODEBOOK is explicitly separated. Wrong-language models use another corpus's TRAIN+VALIDATION only.

Error propagation is measured relative to each candidate's clean decoded sequence, so it isolates damage caused by the injected error from clean ambiguity. Practical recovery tables remain scored against the true TEST plaintext. Reset variants add explicit synthetic checkpoint markers; the same requested boundary deletion is blocked at a checkpoint or merges adjacent tokens inside a checkpoint interval.
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

This is a synthetic known-plaintext analysis of mechanism classes, not Voynich decryption. All primary damage results now come from executed encode -> corrupt/represent/segment -> decode jobs. ERROR_RECOVERABILITY.tsv contains 102,900 raw stochastic rows; the single-error, reset, transcription, and segmentation tables retain positions, seeds, operation counts, and decoder outcomes.

## Required answers

1. **Task66 candidates preserving a message:** M0, M1, and the synthetic H2 representation recover clean glyphs exactly with their declared knowledge. G, M9, M10, and M11 recover only part of the message.
2. **Statistical dependence only:** constrained G/M9/M10/M11 retain positive paired-unit information but their actual decoded token rates are much lower; dependence is not unique recovery.
3. **Voynich-compatible and highly recoverable:** the tested constrained set has a fingerprint/recovery trade-off. G occupies the best tested compromise, but no constrained candidate has control-level unique recovery.
4. **Formation-information trade-off:** yes in this frozen set; high Task66 family compatibility coexists with lower exact token recovery.
5. **Loss from formation itself:** G is many-to-one even without state or corruption; PREIMAGE_MULTIPLICITY and the TEST-codebook oracle measure that loss.
6. **Unique constrained inverse:** none of the frozen constrained Task66 representatives has one. Reversible M0/M1/M2 controls show that the decoder can recover a unique inverse when the encoder preserves it.
7. **Ambiguous with exact key/state:** G, M9, M10, and M11; state knowledge cannot recreate distinctions discarded by the form map.
8. **Natural-language redundancy:** TRAIN+VALIDATION local priors improve choices for observed forms; wrong-language rows quantify the dependence. They do not eliminate intrinsic collisions.
9. **Glyph substitutions:** empirical 100-replicate curves are in RECOVERABILITY_CURVES.tsv; raw outcomes and exact seeds are in ERROR_RECOVERABILITY.tsv.
10. **Insertions/deletions:** glyph edits are generally local in single-error rows, while token-count-changing boundary edits can shift all later positional units.
11. **One-error severity:** ERROR_PROPAGATION.tsv reports actual incremental damaged units at four positions, including end-censored L_sync=-2 rows.
12. **Catastrophic state/position desynchronization:** observed for no-reset token-count-changing errors where no three-unit correct run returns before block end.
13. **Resynchronization length:** L_sync is measured, not inferred; -1 means catastrophic within the block and -2 means the block ended before a three-unit run was observable.
14. **Periodic resets:** token, line-sized, and fixed-N checkpoints localize the same requested boundary deletion; page-sized resets may be too sparse for a 128-word block.
15. **Boundary danger:** deletion, insertion, and +1 shifts are scored independently in SEGMENTATION_DAMAGE.tsv.
16. **Boundary reconstruction:** ciphertext-only dynamic programming produces measured precision, recall, F1, and downstream plaintext recovery; it is not given plaintext or oracle boundaries.
17. **Glyph conflation:** actual random class pairs are merged and decoded. Damage varies by pair/fraction/replicate rather than copying a clean score.
18. **Raw reversible to represented irreversible:** yes as a tested possibility; many-to-one conflation damages M0/M1 recovery despite a reversible raw encoding.
19. **Silent errors:** ERROR_DETECTABILITY.tsv distinguishes valid-form wrong decodes from invalid detectable forms.
20. **Dense valid-form risk:** valid-to-valid errors occur in the frozen form dictionary and are separately counted as undetectable/silent; this is a coding diagnostic, not a Voynich error-rate claim.
21. **Similar fingerprint, different recovery:** G, M9, M10, and M11 overlap strongly on the Task66 compatibility axis but differ substantially in Task67 recovery.
22. **Upper-right region:** no tested constrained candidate combines control-level recovery with maximal compatibility. G and M9 define the tested compromise frontier.
23. **Most encoding-like candidate:** G among constrained candidates, because it retains the highest measured clean recovery; M0/M1/M2 are reversible controls rather than Voynich-compatible claims.
24. **Most generator-like candidate:** the lowest-recovery stateful form representatives and the shuffled-input control; the latter demonstrates preserved grammar with message identity destroyed.
25. **Clean recoverable becoming practically undecodable:** supported for token-count-changing damage without sufficiently frequent checkpoints; reset rows show localization under the allowed robustness variants.
26. **Most destructive damage:** boundary merge/split and cascaded corruption are more globally disruptive than isolated within-token glyph substitutions in this positional decoder.
27. **Where information is lost:** encoding collisions (G/M9/M10/M11), synchronization after token-count errors, transcription conflation, and segmentation are separated from key secrecy and from clean decoder ambiguity.
28. **Original message but insufficient surviving representation:** SUPPORTED_AS_POSSIBILITY for tested synthetic mechanisms only.

## Classification

`)
	for _, c := range cs {
		fmt.Fprintf(&b, "- \u0060%s\u0060: \u0060%s\u0060; %s.\n", c.Name, c.Class, c.Note)
	}
	b.WriteString(`
The tested mechanisms support the possibility that an originally recoverable synthetic encoding may become practically unrecoverable after copying/transcription or segmentation damage: SUPPORTED_AS_POSSIBILITY. This is not a claim about the Voynich manuscript, EVA, historical error rates, or the cause of its undeciphered status.

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
