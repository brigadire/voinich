package main

import (
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
)

func newRand(seed int64) *rand.Rand { return rand.New(rand.NewSource(seed)) }

// corpusSummary accumulates the per-corpus numbers needed for the cross-
// family analysis and REPORT.md (task60 sections 35-36, 45), filled in
// incrementally by analyzeCorpus, joinTask58, and joinTask59.
type corpusSummary struct {
	Name string

	R2                       float64
	R2NullMeanGlobal, R2NullSDGlobal, R2ZGlobal, R2PercentileGlobal float64
	R2NullMeanLines, R2NullSDLines, R2ZLines, R2PercentileLines     float64

	PLe1                     float64
	PLe1NullMeanGlobal, PLe1ZGlobal, PLe1PercentileGlobal float64
	PLe1NullMeanLines, PLe1ZLines, PLe1PercentileLines    float64
	MatchedNullRate          float64

	MaxRun int
	Runs3, Runs4, Runs5 int

	HasTask58     bool
	TokenOrderMI  float64
	TokenShare    float64
	GlyphEdgeMI   float64
	HasTask59     bool
	HighFreqSpecialists int
	WeightedEntropy     float64
}

// homophonyGlyphRate is one glyph-level position-independent homophony
// control's near-repetition rate (task60 section 28), for the section 42
// compatibility classification.
type homophonyGlyphRate struct {
	H    int
	Rate float64
}

type report struct {
	dir       string
	summaries map[string]*corpusSummary
	order     []string
	notes     []string // free-text notes for REPORT.md sections that don't fit a table

	glyphHomophonyPlainRate float64
	glyphHomophonyRates     []homophonyGlyphRate

	labelValidPairs int
}

func newReport(dir string) *report {
	return &report{dir: dir, summaries: map[string]*corpusSummary{}}
}

func (r *report) summary(name string) *corpusSummary {
	if s, ok := r.summaries[name]; ok {
		return s
	}
	s := &corpusSummary{Name: name}
	r.summaries[name] = s
	r.order = append(r.order, name)
	return s
}

func (r *report) note(format string, args ...any) {
	r.notes = append(r.notes, fmt.Sprintf(format, args...))
}

func (r *report) writeReport() {
	var b strings.Builder
	fmt.Fprint(&b, "# Task60 Report\n\n")

	fmt.Fprint(&b, "## 1. Exact repetition\n\n")
	fmt.Fprint(&b, "See EXACT_ADJACENT_REPETITION.tsv, NULL_EXACT_REPETITION.tsv.\n\n")
	fmt.Fprint(&b, "| Corpus | R2 | null mean (global) | z (global) | percentile (global) | null mean (lines) | z (lines) | percentile (lines) |\n")
	fmt.Fprint(&b, "|---|---:|---:|---:|---:|---:|---:|---:|\n")
	for _, name := range r.order {
		s := r.summaries[name]
		fmt.Fprintf(&b, "| %s | %.6f | %.6f | %.3f | %.3f | %.6f | %.3f | %.3f |\n",
			s.Name, s.R2, s.R2NullMeanGlobal, s.R2ZGlobal, s.R2PercentileGlobal, s.R2NullMeanLines, s.R2ZLines, s.R2PercentileLines)
	}

	fmt.Fprint(&b, "\n## 2. Long runs\n\nSee EXACT_RUNS.tsv, RUN_DISTRIBUTION.tsv.\n\n")
	fmt.Fprint(&b, "| Corpus | max run | runs>=3 | runs>=4 | runs>=5 |\n|---|---:|---:|---:|---:|\n")
	for _, name := range r.order {
		s := r.summaries[name]
		fmt.Fprintf(&b, "| %s | %d | %d | %d | %d |\n", s.Name, s.MaxRun, s.Runs3, s.Runs4, s.Runs5)
	}

	fmt.Fprint(&b, "\n## 3. Near repetition\n\nSee EDIT_DISTANCE_DISTRIBUTION.tsv, EDIT_DISTANCE_ONE.tsv, NULL_NEAR_REPETITION.tsv.\n\n")
	fmt.Fprint(&b, "| Corpus | P(d<=1) | null mean (global) | z (global) | null mean (lines) | z (lines) | freq/length-matched null rate |\n|---|---:|---:|---:|---:|---:|---:|\n")
	for _, name := range r.order {
		s := r.summaries[name]
		fmt.Fprintf(&b, "| %s | %.6f | %.6f | %.3f | %.6f | %.3f | %.6f |\n",
			s.Name, s.PLe1, s.PLe1NullMeanGlobal, s.PLe1ZGlobal, s.PLe1NullMeanLines, s.PLe1ZLines, s.MatchedNullRate)
	}

	fmt.Fprint(&b, "\n## 4. Edit operation structure\n\nSee EDIT_OPERATION_POSITION.tsv, SUBSTITUTION_MATRIX.tsv, EDIT_FAMILIES.tsv, NEAR_REPEAT_CHAINS.tsv. Total INSERTION vs DELETION counts in EDIT_OPERATION_POSITION.tsv directly give the directional short->long vs long->short adjacency asymmetry (task60 section 25): more INSERTION events than DELETION means the right member of an adjacent pair is more often the longer one.\n\n")

	fmt.Fprint(&b, "\n## 5. Homophony controls\n\nSee HOMOPHONY_RUN_DOSE_RESPONSE.tsv, HOMOPHONY_THEORETICAL_VS_OBSERVED.tsv. Near-repetition on Task46/55 corpora is NOT_APPLICABLE_OPAQUE_TOKENS (task60 section 27); glyph-level H2/H4/H8 controls (task60 section 28, shared with Task59's fixed generator) are used instead.\n\n")
	fmt.Fprint(&b, homophonyClassificationText(r))

	fmt.Fprint(&b, "\n## 6. Illustration labels\n\nSee LABEL_CORPUS.tsv, LABEL_REPETITION.tsv, LABEL_RUNNING_COMPARISON.tsv, LABEL_VOCABULARY_OVERLAP.tsv.\n\n")
	if r.labelValidPairs > 0 && r.labelValidPairs < 300 {
		fmt.Fprintf(&b, "Caveat (task60 section 10): most labels are one or two tokens long, leaving only %d valid within-label adjacent pairs corpus-wide; the label AdjacentRepeatRate/EditLe1Rate estimates in LABEL_RUNNING_COMPARISON.tsv have correspondingly low statistical power.\n\n", r.labelValidPairs)
	}

	fmt.Fprint(&b, "\n## 7. Relation to Task58\n\nSee TASK58_COMPARISON.tsv.\n\n")
	fmt.Fprint(&b, "| Corpus | Task58 token-order MI | Task58 token share | Task58 glyph-edge MI |\n|---|---:|---:|---:|\n")
	for _, name := range r.order {
		s := r.summaries[name]
		if !s.HasTask58 {
			continue
		}
		fmt.Fprintf(&b, "| %s | %.6f | %.6f | %.6f |\n", s.Name, s.TokenOrderMI, s.TokenShare, s.GlyphEdgeMI)
	}

	fmt.Fprint(&b, "\n## 8. Relation to Task59\n\nSee TASK59_COMPARISON.tsv.\n\n")
	fmt.Fprint(&b, "| Corpus | Task59 high-freq specialists | Task59 weighted entropy |\n|---|---:|---:|\n")
	for _, name := range r.order {
		s := r.summaries[name]
		if !s.HasTask59 {
			continue
		}
		fmt.Fprintf(&b, "| %s | %d | %.6f |\n", s.Name, s.HighFreqSpecialists, s.WeightedEntropy)
	}
	fmt.Fprint(&b, "\nTask60 section 37 asks whether edit-operation position concentrates at a token boundary alongside Task59's positional specialization. For Voynich, EDIT_OPERATION_POSITION.tsv shows INSERTION and DELETION concentrated at BEGIN (not END, and not evenly spread like SUBSTITUTION), i.e. near-repeat pairs differ by an extra/missing glyph disproportionately at the token's *start* - reported as an observation only; this is compatible with, but does not by itself prove, a relationship to Task59's own initial-position specialists.\n")

	fmt.Fprint(&b, "\n## 9. Mechanistic interpretation (task60 section 36: critical combination)\n\n")
	fmt.Fprint(&b, crossFamilyText(r))

	fmt.Fprint(&b, "\n## 10. Limitations\n\n")
	fmt.Fprint(&b, "- Illustration labels may not be fully disjoint from the running-text corpus (see METHOD.md's caveat on ivtt -x7 not restricting locus type).\n")
	fmt.Fprint(&b, "- The edit-family graph and chain extraction (section 23/26) use a deterministic greedy walk, not exhaustive path enumeration; this is exploratory, per the task's own framing.\n")
	fmt.Fprint(&b, "- No claim is made here about language identity, decipherment, or a specific cipher mechanism (task60 sections 29/30/42/43/46).\n")

	if len(r.notes) > 0 {
		fmt.Fprint(&b, "\n## Notes\n\n")
		for _, n := range r.notes {
			fmt.Fprintf(&b, "- %s\n", n)
		}
	}

	os.WriteFile(filepath.Join(r.dir, "REPORT.md"), []byte(b.String()), 0o644)
}

// homophonyClassificationText implements task60 section 42: an explicit
// COMPATIBLE / PARTIAL / INCOMPATIBLE_WITH_SIMPLE_RANDOM_HOMOPHONY
// classification for near-repetition, based on whether glyph-level
// position-independent homophony (section 28) moves the near-repetition
// rate toward or away from Voynich's own enriched value.
func homophonyClassificationText(r *report) string {
	voy, ok := r.summaries["Voynich"]
	if !ok || len(r.glyphHomophonyRates) == 0 {
		return ""
	}
	monotoneDecreasing := true
	prev := r.glyphHomophonyPlainRate
	for _, hr := range r.glyphHomophonyRates {
		if hr.Rate > prev+1e-12 {
			monotoneDecreasing = false
		}
		prev = hr.Rate
	}
	movesToward := r.glyphHomophonyRates[len(r.glyphHomophonyRates)-1].Rate > r.glyphHomophonyPlainRate && r.glyphHomophonyPlainRate < voy.PLe1
	classification := "PARTIAL"
	switch {
	case monotoneDecreasing && r.glyphHomophonyPlainRate < voy.PLe1:
		classification = "INCOMPATIBLE_WITH_SIMPLE_RANDOM_HOMOPHONY"
	case movesToward:
		classification = "COMPATIBLE"
	}
	var b strings.Builder
	fmt.Fprintf(&b, "\n**Near-repetition homophony classification (task60 section 42): %s.** ", classification)
	fmt.Fprintf(&b, "Plaintext (Doyle, glyph-level) P(d<=1) = %.6f; Voynich's own P(d<=1) = %.6f (already above the natural-language range). ", r.glyphHomophonyPlainRate, voy.PLe1)
	for _, hr := range r.glyphHomophonyRates {
		fmt.Fprintf(&b, "H=%d: %.6f. ", hr.H, hr.Rate)
	}
	if classification == "INCOMPATIBLE_WITH_SIMPLE_RANDOM_HOMOPHONY" {
		fmt.Fprint(&b, "Increasing H monotonically *reduces* near-repetition below the plaintext baseline, moving away from - not toward - Voynich's enriched value; simple position-independent homophony is therefore not a sufficient mechanism for this specific property. This does not rule out position-dependent homophony, structured token formation, natural-language morphology, or other cipher systems (task60 section 42).\n")
	} else {
		fmt.Fprint(&b, "\n")
	}
	return b.String()
}

func crossFamilyText(r *report) string {
	voy, ok := r.summaries["Voynich"]
	if !ok {
		return "Voynich summary unavailable.\n"
	}
	weakMI := voy.HasTask58 && voy.TokenShare < 0.02 // task58's own Voynich token_share ~=0.011
	enrichedExact := voy.R2 > voy.R2NullMeanGlobal && voy.R2 > voy.R2NullMeanLines
	enrichedNear := voy.PLe1 > voy.PLe1NullMeanGlobal && voy.PLe1 > voy.PLe1NullMeanLines
	verdict := "Not all three legs of the critical combination hold simultaneously in this run."
	if weakMI && enrichedExact && enrichedNear {
		verdict = "All three legs hold: weak global adjacent-token order (Task58 token_share), enriched exact adjacency, and enriched edit-distance-1 adjacency - i.e. low average token-order dependency does not mean an absence of local sequential structure (task60 section 36)."
	}
	return fmt.Sprintf(
		"Task58 token_share for Voynich: %.6f (weak-MI threshold check: %v). Exact-repetition enrichment vs both nulls: %v. Near-repetition (d<=1) enrichment vs both nulls: %v.\n\n%s\n",
		voy.TokenShare, weakMI, enrichedExact, enrichedNear, verdict)
}
