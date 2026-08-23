package main

import (
	"fmt"

	"zcore.dev/voinich/internal/mechanismspace"
)

// CorpusSpec is one of task66 section 6's three minimum plaintext
// corpora, plus the file it lives at in this repository.
type CorpusSpec struct{ Name, Path string }

// PlaintextCorpora is task66 section 6's minimum set. Paths match the
// already-cleaned (lowercased, punctuation-stripped) files used
// throughout this repository's other independent Task58-65 analyzers
// (e.g. independent/local-regime-topology-analyze), so tokenisation is
// consistent with the authoritative target values.
var PlaintextCorpora = []CorpusSpec{
	{"Doyle", "data_test/pg2097-2.txt"},
	{"Longfellow", "data_test/pg30795-mod.txt"},
	{"Astafiev", "data_test/astafiev-1000-culinar-receipts-prepared.txt"},
}

// VoynichMatchedSize is the authoritative Voynich token count (task58-65
// artifacts, e.g. experiments/rozanova-temerev-v1/comparison.tsv's
// Voynich row), used for task66 section 7's length normalization.
const VoynichMatchedSize = 39380

// LoadCorpora loads and length-normalizes every plaintext corpus (task66
// section 7): corpora larger than Voynich are cut to a deterministic
// matched block; Longfellow, smaller than Voynich, is used in full and
// flagged in the report rather than silently padded.
func LoadCorpora() (map[string]mechanismspace.Corpus, error) {
	out := map[string]mechanismspace.Corpus{}
	for _, spec := range PlaintextCorpora {
		c, err := mechanismspace.LoadNatural(spec.Name, spec.Path)
		if err != nil {
			return nil, fmt.Errorf("loading %s: %w", spec.Name, err)
		}
		out[spec.Name] = c.MatchedSample(VoynichMatchedSize, 1)
	}
	return out, nil
}
