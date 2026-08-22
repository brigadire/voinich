package tokenrepetition

import (
	"strings"

	"zcore.dev/voinich/internal/metadatavalidation"
)

// LabelInstance is one IVTFF label locus (task60 section 3): the raw
// label text and its whitespace-split tokens (dot/comma/gap-as-word-break,
// via metadatavalidation.NormalizeForAlignment - the same convention used
// to build the canonical running-text corpus).
type LabelInstance struct {
	Locus     string
	RawLabel  string
	Tokens    []string
}

// ExtractLabels reads the raw IVTFF source and returns every label-type
// locus (IVTFF locus-type letter "L"), never inferred heuristically from
// the whitespace canonical corpus (task60 section 3).
func ExtractLabels(ivtffPath string) ([]LabelInstance, error) {
	doc, err := metadatavalidation.ParseIVTFF(ivtffPath)
	if err != nil {
		return nil, err
	}
	var out []LabelInstance
	for _, l := range doc.Loci {
		if l.Type != "L" {
			continue
		}
		toks := strings.Fields(l.AlignmentText)
		if len(toks) == 0 {
			continue
		}
		out = append(out, LabelInstance{Locus: l.Folio + "." + l.ID, RawLabel: l.RawText, Tokens: toks})
	}
	return out, nil
}

// LabelCorpus builds a Corpus-shaped view of the labels where each label
// instance is its own "line" (task60 section 4/5: this lets every
// existing adjacent-repetition/run/near-repetition function operate on
// labels directly, with within-label adjacency only - never treating two
// different labels as textually adjacent).
func LabelCorpus(labels []LabelInstance) Corpus {
	c := Corpus{Name: "Voynich-labels"}
	for i, l := range labels {
		for _, t := range l.Tokens {
			c.Tokens = append(c.Tokens, t)
			c.LineOfToken = append(c.LineOfToken, i)
		}
	}
	return c
}

// LabelRepetitionStats is task60 section 29's required summary.
type LabelRepetitionStats struct {
	Instances            int
	UniqueCompleteLabels int
	RepeatedCompleteLabels int
	HapaxLabelFraction   float64
	TokenN               int
	TokenV               int
	TokenHapaxFraction   float64
	RepeatedTokenTypes   int
}

func SummarizeLabelRepetition(labels []LabelInstance) LabelRepetitionStats {
	var s LabelRepetitionStats
	s.Instances = len(labels)
	complete := map[string]int{}
	tokenFreq := map[string]int{}
	for _, l := range labels {
		complete[strings.Join(l.Tokens, " ")]++
		for _, t := range l.Tokens {
			tokenFreq[t]++
			s.TokenN++
		}
	}
	hapaxLabels := 0
	for _, n := range complete {
		if n == 1 {
			hapaxLabels++
		} else {
			s.RepeatedCompleteLabels++
		}
	}
	s.UniqueCompleteLabels = len(complete)
	if s.UniqueCompleteLabels > 0 {
		s.HapaxLabelFraction = float64(hapaxLabels) / float64(s.UniqueCompleteLabels)
	}
	s.TokenV = len(tokenFreq)
	hapaxTokens := 0
	for _, n := range tokenFreq {
		if n == 1 {
			hapaxTokens++
		} else {
			s.RepeatedTokenTypes++
		}
	}
	if s.TokenV > 0 {
		s.TokenHapaxFraction = float64(hapaxTokens) / float64(s.TokenV)
	}
	return s
}

// VocabularyOverlap is task60 section 32.
type VocabularyOverlap struct {
	LabelTypes, RunningTypes, SharedTypes int
	LabelTypeCoverage    float64 // fraction of label token TYPES also found in running text
	LabelOccurrenceCoverage float64 // fraction of label token OCCURRENCES using running-text types
	Jaccard              float64
}

func ComputeVocabularyOverlap(labelTokens, runningTokens []string) VocabularyOverlap {
	labelFreq, runFreq := map[string]int{}, map[string]int{}
	for _, t := range labelTokens {
		labelFreq[t]++
	}
	for _, t := range runningTokens {
		runFreq[t]++
	}
	var ov VocabularyOverlap
	ov.LabelTypes = len(labelFreq)
	ov.RunningTypes = len(runFreq)
	sharedOcc := 0
	for t, n := range labelFreq {
		if runFreq[t] > 0 {
			ov.SharedTypes++
			sharedOcc += n
		}
	}
	if ov.LabelTypes > 0 {
		ov.LabelTypeCoverage = float64(ov.SharedTypes) / float64(ov.LabelTypes)
	}
	if len(labelTokens) > 0 {
		ov.LabelOccurrenceCoverage = float64(sharedOcc) / float64(len(labelTokens))
	}
	union := ov.LabelTypes + ov.RunningTypes - ov.SharedTypes
	if union > 0 {
		ov.Jaccard = float64(ov.SharedTypes) / float64(union)
	}
	return ov
}
